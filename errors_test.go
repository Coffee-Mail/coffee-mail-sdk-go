package coffeemail

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestStatusMapsToCanonicalErrors(t *testing.T) {
	cases := []struct {
		status   int
		code     string
		sentinel error
	}{
		{http.StatusBadRequest, CodeValidationError, ErrValidation},
		{http.StatusUnauthorized, CodeUnauthorized, ErrUnauthorized},
		{http.StatusPaymentRequired, CodePaymentRequired, ErrPaymentRequired},
		{http.StatusForbidden, CodeForbidden, ErrForbidden},
		{http.StatusNotFound, CodeNotFound, ErrNotFound},
		{http.StatusConflict, CodeConflict, ErrConflict},
		{http.StatusTooManyRequests, CodeRateLimitExceeded, ErrRateLimited},
		{http.StatusInternalServerError, CodeInternalServerError, ErrServer},
		{http.StatusBadGateway, CodeInternalServerError, ErrServer},
		{http.StatusServiceUnavailable, CodeInternalServerError, ErrServer},
		{http.StatusGatewayTimeout, CodeInternalServerError, ErrServer},
	}
	for _, tc := range cases {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			client, _ := newTestClient(t, []fakeResponse{{
				status: tc.status,
				body:   `{"error":{"code":"API_SPECIFIC","message":"falhou","meta":{"issues":[1]},"details":{"campo":"to"}}}`,
				header: map[string]string{"X-Request-Id": "req_123"},
			}})
			_, err := client.Domains.List(context.Background())
			var apiErr *Error
			if !errors.As(err, &apiErr) {
				t.Fatalf("esperado *Error, veio %T", err)
			}
			if apiErr.Status != tc.status || apiErr.Code != tc.code || apiErr.APICode != "API_SPECIFIC" {
				t.Fatalf("erro = %+v", apiErr)
			}
			if apiErr.Message != "falhou" || apiErr.RequestID != "req_123" {
				t.Fatalf("mensagem/requestId = %q/%q", apiErr.Message, apiErr.RequestID)
			}
			if apiErr.Details["campo"] != "to" || apiErr.Details["issues"] == nil {
				t.Fatalf("details deveria unir meta e details: %v", apiErr.Details)
			}
			if !errors.Is(err, tc.sentinel) || errors.Is(err, ErrNetwork) {
				t.Fatalf("errors.Is incorreto para %s", tc.code)
			}
			if !strings.Contains(err.Error(), "falhou") {
				t.Fatalf("Error() = %q", err.Error())
			}
		})
	}
}

func TestUnknownStatusKeepsTheAPICode(t *testing.T) {
	client, _ := newTestClient(t, []fakeResponse{respond(http.StatusUnprocessableEntity, `{"error":{"code":"INVALID_FROM_DOMAIN","message":"x"}}`)})
	_, err := client.Domains.List(context.Background())
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.Code != "INVALID_FROM_DOMAIN" {
		t.Fatalf("erro = %v", err)
	}
	client, _ = newTestClient(t, []fakeResponse{respond(http.StatusTeapot, `{}`)})
	_, err = client.Domains.List(context.Background())
	if !errors.As(err, &apiErr) || apiErr.Code != CodeAPIError {
		t.Fatalf("sem código da API deveria ser API_ERROR: %v", err)
	}
}

func TestMissingMessageFallsBackToLocalCatalog(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		locale Locale
		want   string
	}{
		{"401 sem mensagem", http.StatusUnauthorized, `{"error":{"code":"UNAUTHENTICATED"}}`, LocalePtBR, "Ocorreu um erro inesperado"},
		{"429 sem mensagem", http.StatusTooManyRequests, `{"error":{"code":"RATE_LIMIT_EXCEEDED"}}`, LocaleEn, "Rate limit exceeded"},
		{"corpo não JSON", http.StatusBadGateway, "<html>Bad Gateway</html>", LocalePtBR, "<html>Bad Gateway</html>"},
		{"corpo vazio", http.StatusServiceUnavailable, "", LocaleEn, "An unexpected error"},
		{"mensagem na raiz", http.StatusBadRequest, `{"message":"raiz"}`, LocalePtBR, "raiz"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, _ := newTestClient(t, []fakeResponse{respond(tc.status, tc.body)}, WithLocale(tc.locale))
			_, err := client.Domains.List(context.Background())
			var apiErr *Error
			if !errors.As(err, &apiErr) || !strings.HasPrefix(apiErr.Message, tc.want) {
				t.Fatalf("mensagem = %v", err)
			}
		})
	}
}

func TestRetryAfterIsParsed(t *testing.T) {
	cases := []struct {
		header string
		want   time.Duration
	}{
		{"60", time.Minute},
		{"\t5 ", 5 * time.Second},
		{"abc", 0},
		{"-1", 0},
		{"", 0},
		{"999999999999999999", maxRetryAfter},
	}
	for _, tc := range cases {
		header, want := tc.header, tc.want
		client, _ := newTestClient(t, []fakeResponse{{status: http.StatusTooManyRequests, body: "{}", header: map[string]string{"Retry-After": header}}})
		_, err := client.Domains.List(context.Background())
		var apiErr *Error
		if !errors.As(err, &apiErr) || apiErr.RetryAfter != want {
			t.Fatalf("Retry-After %q = %v, esperado %v", header, apiErr.RetryAfter, want)
		}
	}
}

func TestBroadcastForbiddenBecomesPermissionDenied(t *testing.T) {
	client, _ := newTestClient(t, []fakeResponse{respond(http.StatusForbidden, `{"error":{"code":"FORBIDDEN","message":"proibido"}}`)}, WithLocale(LocaleEn))
	_, err := client.Broadcasts.Send(context.Background(), "b1")
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.Code != CodePermissionDenied || !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("erro = %v", err)
	}
	if apiErr.Details["operation"] != "broadcasts.send" || apiErr.Details["requiredPermission"] != "full_access" {
		t.Fatalf("details = %v", apiErr.Details)
	}
	if !strings.Contains(apiErr.Message, "full_access (called by broadcasts.send)") || apiErr.APICode != "FORBIDDEN" {
		t.Fatalf("mensagem = %q", apiErr.Message)
	}
	client, _ = newTestClient(t, []fakeResponse{respond(http.StatusForbidden, `{}`)})
	_, err = client.Domains.List(context.Background())
	if !errors.Is(err, ErrForbidden) {
		t.Fatal("fora de broadcasts o 403 continua FORBIDDEN")
	}
}

func TestNetworkFailuresBecomeNetworkErrors(t *testing.T) {
	cause := errors.New("connection refused")
	client, _ := newTestClient(t, []fakeResponse{{err: cause}})
	_, err := client.Domains.List(context.Background())
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.Code != CodeNetworkError || apiErr.Status != 0 || !errors.Is(err, ErrNetwork) {
		t.Fatalf("erro = %v", err)
	}
	if !errors.Is(err, cause) || !strings.Contains(apiErr.Message, "connection refused") {
		t.Fatalf("a causa deveria ser preservada: %v", err)
	}
}

func TestTimeoutBecomesNetworkErrorWithTimeoutMessage(t *testing.T) {
	client, _ := newTestClient(t, []fakeResponse{{block: true}}, WithTimeout(20*time.Millisecond))
	_, err := client.Domains.List(context.Background())
	var apiErr *Error
	if !errors.As(err, &apiErr) || !errors.Is(err, ErrNetwork) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("erro = %v", err)
	}
	if !strings.Contains(apiErr.Message, "20ms") {
		t.Fatalf("mensagem deveria citar o tempo limite: %q", apiErr.Message)
	}
}

func TestCallerCancellationIsRespected(t *testing.T) {
	client, _ := newTestClient(t, []fakeResponse{{block: true}})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.Domains.List(ctx)
	if !errors.Is(err, context.Canceled) || !errors.Is(err, ErrNetwork) {
		t.Fatalf("erro = %v", err)
	}
}

func TestUnserializableParamsFailBeforeTheRequest(t *testing.T) {
	client, transport := newTestClient(t, nil)
	_, err := client.Templates.Preview(context.Background(), &PreviewTemplateParams{HTML: "x", Variables: map[string]any{"f": func() {}}})
	if !errors.Is(err, ErrValidation) || len(transport.all()) != 0 {
		t.Fatalf("erro = %v, requisições = %d", err, len(transport.all()))
	}
}

func TestRedirectsAreNotFollowed(t *testing.T) {
	client, err := NewClient(testAPIKey)
	if err != nil {
		t.Fatal(err)
	}
	httpClient := client.Emails.r.httpClient
	if httpClient.CheckRedirect == nil || !errors.Is(httpClient.CheckRedirect(nil, nil), http.ErrUseLastResponse) {
		t.Fatal("o client padrão não deve seguir redirecionamentos com a chave de API")
	}
	client, _ = newTestClient(t, []fakeResponse{respond(http.StatusMovedPermanently, "")})
	if _, err := client.Domains.List(context.Background()); err == nil {
		t.Fatal("3xx deve virar erro")
	}
}
