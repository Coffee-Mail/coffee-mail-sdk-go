package coffeemail

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNewClientWithoutKeyFailsWithLocalizedMessage(t *testing.T) {
	t.Setenv(APIKeyEnvironmentVariable, "")
	cases := map[Locale]string{LocalePtBR: "A chave de API não foi informada", LocaleEn: "API key was not provided"}
	for locale, prefix := range cases {
		_, err := NewClient("  ", WithLocale(locale))
		var apiErr *Error
		if !errors.As(err, &apiErr) || !strings.HasPrefix(apiErr.Message, prefix) {
			t.Fatalf("%s: mensagem inesperada: %v", locale, err)
		}
		if !errors.Is(err, ErrMissingAPIKey) || !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("%s: erro deveria corresponder a ErrMissingAPIKey e ErrUnauthorized", locale)
		}
	}
}

func TestNewClientReadsKeyFromEnvironment(t *testing.T) {
	t.Setenv(APIKeyEnvironmentVariable, " cm_live_env ")
	transport := &fakeTransport{}
	client, err := NewClient("", WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, _ = client.Domains.List(context.Background())
	if got := transport.last(t).header.Get("Authorization"); got != "Bearer cm_live_env" {
		t.Fatalf("Authorization = %q", got)
	}
}

func TestRequestsCarryTheStandardHeaders(t *testing.T) {
	client, transport := newTestClient(t, nil, WithLocale(LocaleEn))
	ctx := WithRequestID(context.Background(), "req_abc")
	if _, err := client.Domains.List(ctx); err != nil {
		t.Fatal(err)
	}
	got := transport.last(t)
	want := map[string]string{
		"Authorization":   "Bearer " + testAPIKey,
		"Accept":          "application/json",
		"Accept-Language": "en",
		"User-Agent":      "coffeemail-go/" + Version,
		"X-Request-Id":    "req_abc",
	}
	for name, value := range want {
		if got.header.Get(name) != value {
			t.Errorf("%s = %q, esperado %q", name, got.header.Get(name), value)
		}
	}
	if got.header.Get("Content-Type") != "" || len(got.body) != 0 {
		t.Error("GET não deve ter corpo nem Content-Type")
	}
	if !strings.HasPrefix(got.url, "https://api.coffeemail.com.br/v1/product/domains") {
		t.Errorf("URL = %q", got.url)
	}
}

func TestUnsupportedLocaleFallsBackToPtBR(t *testing.T) {
	client, transport := newTestClient(t, nil, WithLocale("fr"))
	_, _ = client.Domains.List(context.Background())
	if got := transport.last(t).header.Get("Accept-Language"); got != "pt-BR" {
		t.Fatalf("Accept-Language = %q", got)
	}
}

func TestSendMovesIdempotencyAndSandboxToHeaders(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(http.StatusAccepted, `{"id":"e1","status":"queued","queuedAt":"2026-10-14T10:00:00Z"}`)})
	res, err := client.Emails.Send(context.Background(), &SendEmailParams{
		From:           Address{Email: "a@b.com"},
		To:             []Address{{Email: "c@d.com"}},
		Subject:        "Oi",
		Text:           "Oi",
		IdempotencyKey: "pedido-42",
		Sandbox:        true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ID != "e1" || res.Status != EmailStatusQueued || res.QueuedAt.IsZero() {
		t.Fatalf("resposta = %+v", res)
	}
	got := transport.last(t)
	if got.header.Get("X-Idempotency-Key") != "pedido-42" || got.header.Get("X-Coffeemail-Sandbox") != "true" {
		t.Fatalf("headers = %v", got.header)
	}
	if got.header.Get("Content-Type") != "application/json" {
		t.Fatal("POST com corpo deve ter Content-Type JSON")
	}
	var body map[string]any
	_ = json.Unmarshal(got.body, &body)
	for _, field := range []string{"idempotencyKey", "IdempotencyKey", "sandbox", "Sandbox", "cc", "scheduledAt"} {
		if _, ok := body[field]; ok {
			t.Errorf("campo %q não deveria estar no corpo: %s", field, got.body)
		}
	}
}

func TestSendWithoutIdempotencyKeyOmitsHeaderByDefault(t *testing.T) {
	client, transport := newTestClient(t, nil)
	_, _ = client.Emails.Send(context.Background(), &SendEmailParams{})
	if transport.last(t).header.Get("X-Idempotency-Key") != "" {
		t.Fatal("sem retry o SDK não deve gerar chave de idempotência")
	}
}

func TestSendBatchReturnsPerItemResults(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(http.StatusAccepted,
		`[{"ok":true,"data":{"id":"e1","status":"queued","queuedAt":"2026-10-14T10:00:00Z"}},{"ok":false,"error":"destinatário inválido"}]`)})
	results, err := client.Emails.SendBatch(context.Background(), &SendBatchParams{
		Emails:         []SendEmailParams{{Subject: "a"}, {Subject: "b"}},
		IdempotencyKey: "lote-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 || !results[0].OK || results[0].Data.ID != "e1" || results[1].OK || results[1].Error == "" {
		t.Fatalf("resultados = %+v", results)
	}
	got := transport.last(t)
	if !strings.HasPrefix(string(got.body), "[") || got.header.Get("X-Idempotency-Key") != "lote-1" {
		t.Fatalf("corpo/headers inesperados: %s %v", got.body, got.header)
	}
}

func TestPathParametersAreEscaped(t *testing.T) {
	client, transport := newTestClient(t, nil)
	_, _ = client.Emails.Get(context.Background(), "../a b?x")
	if got := transport.last(t).path; got != "/v1/product/emails/..%2Fa%20b%3Fx" {
		t.Fatalf("path = %q", got)
	}
}

func TestListEncodesQueryParameters(t *testing.T) {
	client, transport := newTestClient(t, nil)
	from := time.Date(2026, 10, 1, 9, 30, 0, 0, time.FixedZone("BRT", -3*3600))
	_, _ = client.Emails.List(context.Background(), &ListEmailsParams{Status: EmailStatusBounced, Tag: "fluxo", From: from, Limit: 25})
	want := "from=2026-10-01T12%3A30%3A00Z&limit=25&status=bounced&tag=fluxo"
	if got := transport.last(t).query; got != want {
		t.Fatalf("query = %q, esperado %q", got, want)
	}
	_, _ = client.Emails.List(context.Background(), nil)
	if got := transport.last(t).query; got != "" {
		t.Fatalf("params nil deveria gerar query vazia, veio %q", got)
	}
}

func TestEmptyResponsesDecodeAsSuccess(t *testing.T) {
	client, _ := newTestClient(t, []fakeResponse{respond(http.StatusNoContent, "")})
	if err := client.Emails.Cancel(context.Background(), "e1"); err != nil {
		t.Fatal(err)
	}
}

func TestWarmupStatusIsNilWhenAbsent(t *testing.T) {
	client, _ := newTestClient(t, []fakeResponse{respond(http.StatusOK, "null")})
	status, err := client.Domains.GetWarmupStatus(context.Background(), "d1")
	if err != nil || status != nil {
		t.Fatalf("esperado (nil, nil), veio (%v, %v)", status, err)
	}
	client, _ = newTestClient(t, []fakeResponse{respond(http.StatusOK, `{"domainId":"d1","currentDay":3}`)})
	status, err = client.Domains.GetWarmupStatus(context.Background(), "d1")
	if err != nil || status == nil || status.CurrentDay != 3 {
		t.Fatalf("aquecimento = %+v, %v", status, err)
	}
}

func TestAudienceAcceptsBothContactCountFields(t *testing.T) {
	for _, body := range []string{`{"id":"a","contactsCount":7}`, `{"id":"a","totalContacts":7}`} {
		client, _ := newTestClient(t, []fakeResponse{respond(http.StatusOK, body)})
		audience, err := client.Audiences.Get(context.Background(), "a")
		if err != nil || audience.TotalContacts != 7 {
			t.Fatalf("%s: TotalContacts = %v, %v", body, audience, err)
		}
	}
	client, _ := newTestClient(t, []fakeResponse{respond(http.StatusOK, `{"id":`)})
	if _, err := client.Audiences.Get(context.Background(), "a"); !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("JSON inválido deveria gerar ErrInvalidResponse, veio %v", err)
	}
}

func TestPreviewByIDRendersTheStoredTemplate(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{
		respond(http.StatusOK, `{"id":"t1","html":"<p>{{nome}}</p>","format":"react"}`),
		respond(http.StatusOK, `{"html":"<p>Ana</p>","text":"Ana"}`),
	})
	preview, err := client.Templates.PreviewByID(context.Background(), "t1", map[string]any{"nome": "Ana"})
	if err != nil || preview.Text != "Ana" {
		t.Fatalf("preview = %+v, %v", preview, err)
	}
	var body PreviewTemplateParams
	_ = json.Unmarshal(transport.last(t).body, &body)
	if body.HTML != "<p>{{nome}}</p>" || body.Format != TemplateFormatReact || body.Variables["nome"] != "Ana" {
		t.Fatalf("corpo do preview = %+v", body)
	}
	client, _ = newTestClient(t, []fakeResponse{respond(http.StatusNotFound, `{"error":{"code":"TEMPLATE_NOT_FOUND"}}`)})
	if _, err := client.Templates.PreviewByID(context.Background(), "x", nil); !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperado ErrNotFound, veio %v", err)
	}
}

func TestIntrospectionIsCachedAndCanBeInvalidated(t *testing.T) {
	now := time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC)
	transport := &fakeTransport{responses: []fakeResponse{respond(http.StatusOK, `{"id":"k1","name":"prod","organisationId":"o1","scopes":["full_access"]}`)}}
	client, err := NewClient(testAPIKey, WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatal(err)
	}
	client.introspection.now = func() time.Time { return now }
	for range 3 {
		info, err := client.Introspect(context.Background())
		if err != nil || info.Scopes[0] != "full_access" {
			t.Fatalf("introspect = %+v, %v", info, err)
		}
	}
	if n := len(transport.all()); n != 1 {
		t.Fatalf("esperada 1 requisição com cache, vieram %d", n)
	}
	now = now.Add(introspectionTTL)
	_, _ = client.Introspect(context.Background())
	client.InvalidateAPIKeyCache()
	_, _ = client.Introspect(context.Background())
	if n := len(transport.all()); n != 3 {
		t.Fatalf("expiração e invalidação deveriam refazer a consulta: %d requisições", n)
	}
}

func TestIntrospectionErrorIsNotCached(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(http.StatusUnauthorized, `{"error":{"code":"UNAUTHENTICATED"}}`), respond(http.StatusOK, `{"id":"k1"}`)})
	if _, err := client.Introspect(context.Background()); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("esperado ErrUnauthorized, veio %v", err)
	}
	if info, err := client.Introspect(context.Background()); err != nil || info.ID != "k1" {
		t.Fatalf("segunda consulta = %+v, %v", info, err)
	}
	if len(transport.all()) != 2 {
		t.Fatal("erro não deveria ficar em cache")
	}
}

func TestSecretsNeverAppearInFormattedValues(t *testing.T) {
	client, _ := newTestClient(t, []fakeResponse{respond(http.StatusBadRequest, `{"error":{"code":"X","message":"m"}}`)})
	_, err := client.Domains.List(context.Background())
	for _, text := range []string{fmt.Sprintf("%v %+v %#v", client, client, client), fmt.Sprintf("%v %+v", err, err)} {
		if strings.Contains(text, testAPIKey) {
			t.Fatal("a chave de API vazou numa representação textual")
		}
	}
}
