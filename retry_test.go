package coffeemail

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

var fastRetry = WithRetry(RetryPolicy{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: 2 * time.Millisecond})

func TestRetryIsOffByDefault(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(http.StatusServiceUnavailable, "{}")})
	_, _ = client.Domains.List(context.Background())
	if n := len(transport.all()); n != 1 {
		t.Fatalf("sem WithRetry deve haver 1 tentativa, houve %d", n)
	}
}

func TestRetryRecoversFromTransientFailures(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{
		{err: errors.New("reset")},
		respond(http.StatusBadGateway, "{}"),
		respond(http.StatusOK, `{"domains":[{"id":"d1"}]}`),
	}, fastRetry)
	res, err := client.Domains.List(context.Background())
	if err != nil || res.Domains[0].ID != "d1" || len(transport.all()) != 3 {
		t.Fatalf("res = %+v, err = %v, tentativas = %d", res, err, len(transport.all()))
	}
}

func TestRetryStopsAtMaxAttempts(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(http.StatusInternalServerError, "{}")}, fastRetry)
	_, err := client.Domains.List(context.Background())
	if !errors.Is(err, ErrServer) || len(transport.all()) != 3 {
		t.Fatalf("err = %v, tentativas = %d", err, len(transport.all()))
	}
}

func TestRetrySkipsNonRetryableResponses(t *testing.T) {
	cases := map[string]fakeResponse{
		"validação":      respond(http.StatusBadRequest, "{}"),
		"não encontrado": respond(http.StatusNotFound, "{}"),
		"cota do plano":  respond(http.StatusTooManyRequests, `{"error":{"code":"PLAN_QUOTA_EXCEEDED"}}`),
	}
	for name, response := range cases {
		t.Run(name, func(t *testing.T) {
			client, transport := newTestClient(t, []fakeResponse{response}, fastRetry)
			_, _ = client.Domains.List(context.Background())
			if n := len(transport.all()); n != 1 {
				t.Fatalf("tentativas = %d", n)
			}
		})
	}
}

func TestRetryNeverRepeatsPostsWithoutIdempotency(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(http.StatusServiceUnavailable, "{}")}, fastRetry)
	_, _ = client.Domains.Create(context.Background(), "exemplo.com.br")
	if n := len(transport.all()); n != 1 {
		t.Fatalf("POST sem idempotência foi repetido %d vezes", n)
	}
}

func TestRetryGeneratesAndReusesIdempotencyKeyForSends(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{
		respond(http.StatusTooManyRequests, `{"error":{"code":"RATE_LIMIT_EXCEEDED"}}`),
		respond(http.StatusAccepted, `{"id":"e1","status":"queued"}`),
	}, fastRetry)
	if _, err := client.Emails.Send(context.Background(), &SendEmailParams{}); err != nil {
		t.Fatal(err)
	}
	requests := transport.all()
	first := requests[0].header.Get("X-Idempotency-Key")
	if len(requests) != 2 || len(first) != 36 || requests[1].header.Get("X-Idempotency-Key") != first {
		t.Fatalf("chave gerada deve ser reutilizada: %q / %q", first, requests[1].header.Get("X-Idempotency-Key"))
	}
}

func TestRetryKeepsTheCallerIdempotencyKey(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(http.StatusBadGateway, "{}"), respond(http.StatusAccepted, `{}`)}, fastRetry)
	_, _ = client.Emails.SendBatch(context.Background(), &SendBatchParams{IdempotencyKey: "lote-9"})
	for _, req := range transport.all() {
		if req.header.Get("X-Idempotency-Key") != "lote-9" {
			t.Fatalf("chave alterada: %q", req.header.Get("X-Idempotency-Key"))
		}
	}
}

func TestRetryHonorsRetryAfterUpToMaxDelay(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{
		{status: http.StatusTooManyRequests, body: "{}", header: map[string]string{"Retry-After": "60"}},
		respond(http.StatusOK, `{}`),
	}, WithRetry(RetryPolicy{MaxAttempts: 2, BaseDelay: time.Millisecond, MaxDelay: 30 * time.Millisecond}))
	start := time.Now()
	if _, err := client.Domains.List(context.Background()); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	if elapsed < 30*time.Millisecond || elapsed > time.Second || len(transport.all()) != 2 {
		t.Fatalf("espera = %v, tentativas = %d", elapsed, len(transport.all()))
	}
}

func TestRetryStopsWhenContextIsCancelledWhileWaiting(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(http.StatusServiceUnavailable, "{}")},
		WithRetry(RetryPolicy{MaxAttempts: 5, BaseDelay: time.Hour, MaxDelay: time.Hour}))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := client.Domains.List(ctx)
	if !errors.Is(err, ErrServer) || len(transport.all()) != 1 {
		t.Fatalf("err = %v, tentativas = %d", err, len(transport.all()))
	}
}

func TestRetryPolicyDefaults(t *testing.T) {
	policy := RetryPolicy{MaxAttempts: -1}.withDefaults()
	if policy.MaxAttempts != 1 || policy.BaseDelay != defaultRetryBaseDelay || policy.MaxDelay != defaultRetryMaxDelay || policy.enabled() {
		t.Fatalf("padrões = %+v", policy)
	}
}
