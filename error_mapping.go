package coffeemail

import (
	"encoding/json"
	"maps"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/headers"
	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/i18n"
)

const (
	planQuotaExceeded        = "PLAN_QUOTA_EXCEEDED"
	fullAccessPermission     = "full_access"
	detailOperation          = "operation"
	detailRequiredPermission = "requiredPermission"
	minServerErrorStatus     = http.StatusInternalServerError
	retryAfterBitSize        = 63
)

var retryAfterSeconds = regexp.MustCompile(`^\d{1,18}$`)

var canonicalCodes = map[int]string{
	http.StatusBadRequest:          CodeValidationError,
	http.StatusUnauthorized:        CodeUnauthorized,
	http.StatusPaymentRequired:     CodePaymentRequired,
	http.StatusForbidden:           CodeForbidden,
	http.StatusNotFound:            CodeNotFound,
	http.StatusConflict:            CodeConflict,
	http.StatusTooManyRequests:     CodeRateLimitExceeded,
	http.StatusInternalServerError: CodeInternalServerError,
	http.StatusBadGateway:          CodeInternalServerError,
	http.StatusServiceUnavailable:  CodeInternalServerError,
	http.StatusGatewayTimeout:      CodeInternalServerError,
}

var fallbackMessages = map[int]i18n.Key{
	http.StatusTooManyRequests: i18n.RateLimitExceeded,
}

type errorEnvelope struct {
	Error   *errorBody `json:"error"`
	Message string     `json:"message"`
}

type errorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Meta    map[string]any `json:"meta"`
	Details map[string]any `json:"details"`
}

func (r *requester) apiError(res *response, op operation) *Error {
	envelope, parsed := parseErrorEnvelope(res.body)
	body := envelope.body()
	apiErr := &Error{
		Status:     res.status,
		Code:       canonicalCode(res.status, body.Code),
		APICode:    body.Code,
		Message:    r.errorMessage(res, envelope, parsed),
		Details:    mergeDetails(body),
		RequestID:  res.header.Get(headers.RequestID),
		RetryAfter: parseRetryAfter(res.header.Get(headers.RetryAfter)),
	}
	return r.remapPermission(apiErr, op)
}

func parseErrorEnvelope(raw []byte) (errorEnvelope, bool) {
	var envelope errorEnvelope
	err := json.Unmarshal(raw, &envelope)
	return envelope, err == nil
}

func (e errorEnvelope) body() errorBody {
	if e.Error == nil {
		return errorBody{}
	}
	return *e.Error
}

func canonicalCode(status int, apiCode string) string {
	if code, ok := canonicalCodes[status]; ok {
		return code
	}
	if apiCode != "" {
		return apiCode
	}
	return CodeAPIError
}

func (r *requester) errorMessage(res *response, envelope errorEnvelope, parsed bool) string {
	candidates := []string{envelope.body().Message, envelope.Message, rawMessage(res.body, parsed)}
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) != "" {
			return candidate
		}
	}
	key, ok := fallbackMessages[res.status]
	if !ok {
		key = i18n.UnexpectedError
	}
	return i18n.Format(string(r.locale), key, nil)
}

func rawMessage(raw []byte, parsed bool) string {
	if parsed {
		return ""
	}
	return string(raw)
}

func mergeDetails(body errorBody) map[string]any {
	details := make(map[string]any, len(body.Meta)+len(body.Details))
	maps.Copy(details, body.Meta)
	maps.Copy(details, body.Details)
	return details
}

func parseRetryAfter(value string) time.Duration {
	trimmed := strings.TrimSpace(value)
	if !retryAfterSeconds.MatchString(trimmed) {
		return 0
	}
	seconds, err := strconv.ParseInt(trimmed, decimalBase, retryAfterBitSize)
	if err != nil || seconds > int64(maxRetryAfter/time.Second) {
		return maxRetryAfter
	}
	return time.Duration(seconds) * time.Second
}

const maxRetryAfter = time.Duration(1<<63 - 1)

func (r *requester) remapPermission(apiErr *Error, op operation) *Error {
	if op.fullAccessName == "" || apiErr.Status != http.StatusForbidden {
		return apiErr
	}
	apiErr.Code = CodePermissionDenied
	apiErr.Details[detailOperation] = op.fullAccessName
	apiErr.Details[detailRequiredPermission] = fullAccessPermission
	apiErr.Message = i18n.Format(string(r.locale), i18n.PermissionDenied, map[string]string{
		i18n.ParamPermission: fullAccessPermission + " (called by " + op.fullAccessName + ")",
	})
	return apiErr
}

func retryable(apiErr *Error) bool {
	if apiErr.Code == CodeNetworkError {
		return true
	}
	if apiErr.Status == http.StatusTooManyRequests {
		return apiErr.APICode != planQuotaExceeded
	}
	return apiErr.Status >= minServerErrorStatus
}
