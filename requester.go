package coffeemail

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/backoff"
	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/headers"
	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/i18n"
	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/idempotency"
	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/routes"
)

const (
	maxResponseBytes = 10 << 20
	minSuccessStatus = http.StatusOK
	maxSuccessStatus = http.StatusMultipleChoices
	decimalBase      = 10
)

type requester struct {
	apiKey     string
	httpClient *http.Client
	locale     Locale
	timeout    time.Duration
	retry      RetryPolicy
	backoff    backoff.Policy
}

type response struct {
	status int
	header http.Header
	body   []byte
}

func newRequester(apiKey string, cfg config) *requester {
	retry := cfg.retry.withDefaults()
	return &requester{
		apiKey:     apiKey,
		httpClient: cfg.httpClient,
		locale:     cfg.locale,
		timeout:    cfg.timeout,
		retry:      retry,
		backoff: backoff.Policy{
			BaseDelay: retry.BaseDelay,
			MaxDelay:  retry.MaxDelay,
			Jitter:    fullJitter,
		},
	}
}

func fullJitter(upper time.Duration) time.Duration {
	if upper <= 0 {
		return 0
	}
	return rand.N(upper + 1)
}

func call[T any](ctx context.Context, r *requester, op operation) (*T, error) {
	var out T
	if err := r.do(ctx, op, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *requester) do(ctx context.Context, op operation, out any) error {
	payload, err := encodeBody(op.body)
	if err != nil {
		return r.invalidRequest(err)
	}
	op = r.prepareIdempotency(op)
	attempts := r.attemptsFor(op)
	for attempt := 1; ; attempt++ {
		res, apiErr := r.attempt(ctx, op, payload)
		if apiErr == nil {
			return r.decode(res, out)
		}
		if attempt >= attempts || !retryable(apiErr) || r.wait(ctx, attempt, apiErr) != nil {
			return apiErr
		}
	}
}

func encodeBody(body any) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	return json.Marshal(body)
}

func (r *requester) prepareIdempotency(op operation) operation {
	if r.retry.enabled() && op.acceptsIdempotency && op.idempotencyKey == "" {
		op.idempotencyKey = idempotency.NewKey()
	}
	return op
}

var idempotentMethods = map[string]bool{http.MethodGet: true, http.MethodDelete: true}

func (r *requester) attemptsFor(op operation) int {
	safe := idempotentMethods[op.method] || (op.acceptsIdempotency && op.idempotencyKey != "")
	if safe {
		return r.retry.MaxAttempts
	}
	return singleAttempt
}

func (r *requester) attempt(ctx context.Context, op operation, payload []byte) (*response, *Error) {
	attemptCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(attemptCtx, op.method, routes.BaseURL+op.url(), bodyReader(payload))
	if err != nil {
		return nil, r.networkError(err)
	}
	req.Header = r.headersFor(ctx, op, payload)
	res, err := r.httpClient.Do(req)
	if err != nil {
		return nil, r.networkError(err)
	}
	defer closeQuietly(res.Body)
	body, err := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes))
	if err != nil {
		return nil, r.networkError(err)
	}
	result := &response{status: res.StatusCode, header: res.Header, body: body}
	if res.StatusCode < minSuccessStatus || res.StatusCode >= maxSuccessStatus {
		return nil, r.apiError(result, op)
	}
	return result, nil
}

func closeQuietly(body io.Closer) {
	_ = body.Close()
}

func bodyReader(payload []byte) io.Reader {
	if payload == nil {
		return http.NoBody
	}
	return bytes.NewReader(payload)
}

func (r *requester) headersFor(ctx context.Context, op operation, payload []byte) http.Header {
	h := http.Header{}
	h.Set(headers.Authorization, headers.BearerPrefix+r.apiKey)
	h.Set(headers.Accept, headers.JSONMediaType)
	h.Set(headers.AcceptLanguage, string(r.locale))
	h.Set(headers.UserAgent, headers.UserAgentPrefix+Version)
	setIf(h, payload != nil, headers.ContentType, headers.JSONMediaType)
	setIf(h, op.idempotencyKey != "", headers.IdempotencyKey, op.idempotencyKey)
	setIf(h, op.sandbox, headers.Sandbox, headers.SandboxEnabled)
	requestID := requestIDFrom(ctx)
	setIf(h, requestID != "", headers.RequestID, requestID)
	return h
}

func setIf(h http.Header, condition bool, name, value string) {
	if condition {
		h.Set(name, value)
	}
}

func (r *requester) decode(res *response, out any) error {
	if out == nil || res.status == http.StatusNoContent || len(bytes.TrimSpace(res.body)) == 0 {
		return nil
	}
	if err := json.Unmarshal(res.body, out); err != nil {
		return &Error{
			Status:    res.status,
			Code:      CodeInvalidResponse,
			Message:   i18n.Format(string(r.locale), i18n.InvalidJSON, nil),
			RequestID: res.header.Get(headers.RequestID),
			cause:     err,
		}
	}
	return nil
}

func (r *requester) wait(ctx context.Context, attempt int, apiErr *Error) error {
	timer := time.NewTimer(r.backoff.Delay(attempt, apiErr.RetryAfter))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (r *requester) networkError(err error) *Error {
	return &Error{
		Code:    CodeNetworkError,
		Message: r.networkMessage(err),
		cause:   err,
	}
}

func (r *requester) networkMessage(err error) string {
	locale := string(r.locale)
	if errors.Is(err, context.DeadlineExceeded) {
		return i18n.Format(locale, i18n.RequestTimeout, map[string]string{
			i18n.ParamTimeoutMs: strconv.FormatInt(r.timeout.Milliseconds(), decimalBase),
		})
	}
	return i18n.Format(locale, i18n.NetworkError, map[string]string{i18n.ParamDetails: err.Error()})
}

func (r *requester) invalidRequest(err error) *Error {
	return &Error{
		Code:    CodeValidationError,
		Message: err.Error(),
		cause:   err,
	}
}
