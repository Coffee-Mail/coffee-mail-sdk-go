package coffeemail

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

const (
	dkimRotationJSON = `{"domainId":"d1","selector":"cm2","startedAt":"2026-10-15T12:00:00Z",` +
		`"record":{"type":"TXT","host":"cm2._domainkey.acme.com","value":"v=DKIM1; k=rsa; p=abc"}}`
	trackingHostJSON = `{"host":"links.acme.com","target":"track.coffeemail.com.br","verified":false,` +
		`"verifiedAt":null,"record":{"type":"CNAME","host":"links.acme.com","value":"track.coffeemail.com.br","ttl":3600}}`
)

func TestRotateDKIMPostsWithoutBodyAndDecodesTheRotation(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(200, dkimRotationJSON)})

	rotation, err := client.Domains.RotateDKIM(context.Background(), "d1")
	if err != nil {
		t.Fatalf("RotateDKIM: %v", err)
	}

	request := transport.last(t)
	if request.method != "POST" || request.path != "/v1/product/domains/d1/dkim/rotate" {
		t.Fatalf("requisição inesperada: %s %s", request.method, request.path)
	}
	if len(request.body) != 0 {
		t.Fatalf("esperava corpo vazio, veio %s", request.body)
	}
	started := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	if rotation.Selector != "cm2" || !rotation.StartedAt.Equal(started) || rotation.Record.Type != DNSRecordTypeTXT {
		t.Fatalf("rotação inesperada: %+v", rotation)
	}
}

func TestRotateDKIMMapsNotFoundAndNotVerified(t *testing.T) {
	client, _ := newTestClient(t, []fakeResponse{
		respond(404, `{"error":{"code":"DOMAIN_NOT_FOUND","message":"domain not found"}}`),
		respond(409, `{"error":{"code":"DOMAIN_NOT_VERIFIED","message":"domain is not verified"}}`),
	})

	_, err := client.Domains.RotateDKIM(context.Background(), "d_x")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperava ErrNotFound, veio %v", err)
	}
	_, err = client.Domains.RotateDKIM(context.Background(), "d1")
	var apiErr *Error
	if !errors.Is(err, ErrConflict) || !errors.As(err, &apiErr) || apiErr.APICode != "DOMAIN_NOT_VERIFIED" {
		t.Fatalf("esperava ErrConflict com DOMAIN_NOT_VERIFIED, veio %v", err)
	}
}

func TestSetTrackingDomainPutsTheHost(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(200, trackingHostJSON)})

	status, err := client.Domains.SetTrackingDomain(context.Background(), "d1", "links.acme.com")
	if err != nil {
		t.Fatalf("SetTrackingDomain: %v", err)
	}

	request := transport.last(t)
	if request.method != "PUT" || request.path != "/v1/product/domains/d1/tracking-domain" {
		t.Fatalf("requisição inesperada: %s %s", request.method, request.path)
	}
	var sent map[string]any
	if err := json.Unmarshal(request.body, &sent); err != nil {
		t.Fatalf("corpo inválido: %v", err)
	}
	if len(sent) != 1 || sent["host"] != "links.acme.com" {
		t.Fatalf("corpo inesperado: %v", sent)
	}
	if status.Verified || status.VerifiedAt != nil || status.Record.TTL == nil || *status.Record.TTL != 3600 {
		t.Fatalf("status inesperado: %+v", status)
	}
}

func TestSetTrackingDomainMapsInvalidHostDisabledAndNotFound(t *testing.T) {
	client, _ := newTestClient(t, []fakeResponse{
		respond(400, `{"error":{"code":"INVALID_TRACKING_HOST","message":"host must be a subdomain"}}`),
		respond(409, `{"error":{"code":"TRACKING_DOMAINS_DISABLED","message":"tracking domains are disabled"}}`),
		respond(404, `{"error":{"code":"DOMAIN_NOT_FOUND"}}`),
	})
	expected := []struct {
		sentinel error
		apiCode  string
	}{
		{ErrValidation, "INVALID_TRACKING_HOST"},
		{ErrConflict, "TRACKING_DOMAINS_DISABLED"},
		{ErrNotFound, "DOMAIN_NOT_FOUND"},
	}

	for _, want := range expected {
		_, err := client.Domains.SetTrackingDomain(context.Background(), "d1", "links.outro.com")
		var apiErr *Error
		if !errors.Is(err, want.sentinel) || !errors.As(err, &apiErr) || apiErr.APICode != want.apiCode {
			t.Fatalf("esperava %v com %s, veio %v", want.sentinel, want.apiCode, err)
		}
	}
}

func TestRemoveTrackingDomainDeletesAndReturnsNil(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(200, "null"), respond(404, `{"error":{"code":"DOMAIN_NOT_FOUND"}}`)})

	status, err := client.Domains.RemoveTrackingDomain(context.Background(), "d1")
	if err != nil || status != nil {
		t.Fatalf("esperava (nil, nil), veio (%+v, %v)", status, err)
	}
	request := transport.last(t)
	if request.method != "DELETE" || request.path != "/v1/product/domains/d1/tracking-domain" {
		t.Fatalf("requisição inesperada: %s %s", request.method, request.path)
	}
	if _, err := client.Domains.RemoveTrackingDomain(context.Background(), "d_x"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperava ErrNotFound, veio %v", err)
	}
}

func TestGetDomainDecodesDeliverabilityFields(t *testing.T) {
	detail := `{"id":"d1","name":"acme.com","status":"verified","dkimSelector":"cm1","createdAt":"2026-10-01T00:00:00Z",` +
		`"dkimRotation":` + dkimRotationJSON + `,` +
		`"returnPath":{"host":"bounces.acme.com","target":"bounces.coffeemail.com.br","verified":true,` +
		`"verifiedAt":"2026-10-02T00:00:00Z","record":{"type":"CNAME","host":"bounces.acme.com","value":"bounces.coffeemail.com.br"}},` +
		`"trackingDomain":` + trackingHostJSON + `,"trackingDomainsAvailable":true}`
	client, _ := newTestClient(t, []fakeResponse{respond(200, detail), respond(200, `{"id":"d2","name":"b.com","status":"pending",`+
		`"dkimSelector":"cm1","createdAt":"2026-10-01T00:00:00Z","dkimRotation":null,"returnPath":null,"trackingDomain":null,`+
		`"trackingDomainsAvailable":false}`)})

	got, err := client.Domains.Get(context.Background(), "d1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.DKIMRotation == nil || got.DKIMRotation.Selector != "cm2" {
		t.Fatalf("dkimRotation inesperado: %+v", got.DKIMRotation)
	}
	if got.ReturnPath == nil || !got.ReturnPath.Verified || got.ReturnPath.VerifiedAt == nil {
		t.Fatalf("returnPath inesperado: %+v", got.ReturnPath)
	}
	if got.TrackingDomain == nil || got.TrackingDomain.Host != "links.acme.com" || !got.TrackingDomainsAvailable {
		t.Fatalf("tracking inesperado: %+v %v", got.TrackingDomain, got.TrackingDomainsAvailable)
	}

	empty, err := client.Domains.Get(context.Background(), "d2")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if empty.DKIMRotation != nil || empty.ReturnPath != nil || empty.TrackingDomain != nil || empty.TrackingDomainsAvailable {
		t.Fatalf("esperava campos vazios: %+v", empty)
	}
}

func TestWarmupStatusDecodesStateAndRampDates(t *testing.T) {
	warming := `{"domainId":"d1","status":"warming","currentDay":3,"dailyQuota":200,"sentToday":50,"remainingToday":150,` +
		`"quotaUsedPercent":25,"lastSendDate":"2026-10-15","totalDays":30,"startedAt":"2026-10-13T00:00:00Z","completedAt":null}`
	completed := `{"domainId":"d1","status":"completed","currentDay":30,"dailyQuota":0,"sentToday":0,"remainingToday":0,` +
		`"quotaUsedPercent":0,"lastSendDate":null,"totalDays":30,"startedAt":"2026-09-01T00:00:00Z","completedAt":"2026-10-01T00:00:00Z"}`
	client, _ := newTestClient(t, []fakeResponse{respond(200, warming), respond(200, completed)})

	got, err := client.Domains.GetWarmupStatus(context.Background(), "d1")
	if err != nil {
		t.Fatalf("GetWarmupStatus: %v", err)
	}
	started := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	if got.Status != DomainWarmupWarming || got.TotalDays != 30 || !got.StartedAt.Equal(started) || got.CompletedAt != nil {
		t.Fatalf("aquecimento inesperado: %+v", got)
	}

	done, err := client.Domains.GetWarmupStatus(context.Background(), "d1")
	if err != nil {
		t.Fatalf("GetWarmupStatus: %v", err)
	}
	if done.Status != DomainWarmupCompleted || done.CompletedAt == nil {
		t.Fatalf("aquecimento concluído inesperado: %+v", done)
	}
}

const (
	templateJSON = `{"id":"t1","organisationId":"o1","name":"Boas-vindas","alias":"boas-vindas","subject":"Oi","preheader":null,` +
		`"html":"<p>v2</p>","contentJson":null,"textPayload":null,"variables":[],"isActive":true,"format":"html",` +
		`"sourceLocale":"pt-BR","createdAt":"2026-10-01T00:00:00Z","updatedAt":"2026-10-15T00:00:00Z"}`
	webhookDeliveryJSON = `{"id":"dl2","webhookId":"w1","emailEventId":null,"eventType":"email.delivered","responseStatus":null,` +
		`"responseBody":null,"attempts":0,"lastError":null,"status":"pending","nextRunAt":"2026-10-15T12:00:00Z",` +
		`"createdAt":"2026-10-15T12:00:00Z","requestPayload":"{\"type\":\"email.delivered\"}"}`
)

func TestListVersionsSendsPaginationAndDecodesThePage(t *testing.T) {
	page := `{"data":[{"version":3,"name":"Boas-vindas","subject":null,"createdBy":null,"createdAt":"2026-10-15T00:00:00Z"},` +
		`{"version":2,"name":"Boas-vindas","subject":"Oi","createdBy":"u1","createdAt":"2026-10-14T00:00:00Z"}],"total":3}`
	client, transport := newTestClient(t, []fakeResponse{respond(200, page), respond(200, `{"data":[],"total":0}`)})

	got, err := client.Templates.ListVersions(context.Background(), "t1", &OffsetParams{Limit: 2, Offset: 4})
	if err != nil {
		t.Fatalf("ListVersions: %v", err)
	}
	request := transport.last(t)
	if request.method != "GET" || request.path != "/v1/product/templates/t1/versions" || request.query != "limit=2&offset=4" {
		t.Fatalf("requisição inesperada: %s %s?%s", request.method, request.path, request.query)
	}
	if got.Total != 3 || len(got.Data) != 2 || got.Data[0].Version != 3 || got.Data[0].CreatedBy != nil || *got.Data[1].Subject != "Oi" {
		t.Fatalf("página inesperada: %+v", got)
	}

	if _, err := client.Templates.ListVersions(context.Background(), "t1", nil); err != nil {
		t.Fatalf("ListVersions sem params: %v", err)
	}
	if query := transport.last(t).query; query != "" {
		t.Fatalf("esperava query vazia, veio %q", query)
	}
}

func TestGetVersionDecodesTheFullContent(t *testing.T) {
	version := `{"version":2,"name":"Boas-vindas","subject":"Oi","createdBy":null,"createdAt":"2026-10-14T00:00:00Z",` +
		`"templateId":"t1","preheader":null,"html":"<p>v2</p>","contentJson":{"blocks":[]},"textPayload":"v2",` +
		`"variables":[{"name":"nome"}],"format":"html","sourceLocale":"pt-BR"}`
	client, transport := newTestClient(t, []fakeResponse{respond(200, version)})

	got, err := client.Templates.GetVersion(context.Background(), "t1", 2)
	if err != nil {
		t.Fatalf("GetVersion: %v", err)
	}
	if request := transport.last(t); request.method != "GET" || request.path != "/v1/product/templates/t1/versions/2" {
		t.Fatalf("requisição inesperada: %s %s", request.method, request.path)
	}
	if got.Version != 2 || got.TemplateID != "t1" || got.HTML != "<p>v2</p>" || got.Variables[0].Name != "nome" || got.Format != TemplateFormatHTML {
		t.Fatalf("versão inesperada: %+v", got)
	}
}

func TestRestoreVersionPostsWithoutBodyAndReturnsTheTemplate(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(200, templateJSON)})

	got, err := client.Templates.RestoreVersion(context.Background(), "t1", 2)
	if err != nil {
		t.Fatalf("RestoreVersion: %v", err)
	}
	request := transport.last(t)
	if request.method != "POST" || request.path != "/v1/product/templates/t1/versions/2/restore" || len(request.body) != 0 {
		t.Fatalf("requisição inesperada: %s %s %s", request.method, request.path, request.body)
	}
	if got.ID != "t1" || got.HTML != "<p>v2</p>" {
		t.Fatalf("template inesperado: %+v", got)
	}
}

func TestTemplateVersionsMapNotFoundCodes(t *testing.T) {
	client, _ := newTestClient(t, []fakeResponse{
		respond(404, `{"error":{"code":"TEMPLATE_NOT_FOUND","message":"template not found"}}`),
		respond(404, `{"error":{"code":"TEMPLATE_VERSION_NOT_FOUND","message":"version not found"}}`),
		respond(404, `{"error":{"code":"TEMPLATE_VERSION_NOT_FOUND","message":"version not found"}}`),
	})
	calls := []struct {
		apiCode string
		invoke  func() error
	}{
		{"TEMPLATE_NOT_FOUND", func() error { return ignore(client.Templates.ListVersions(context.Background(), "t_x", nil)) }},
		{"TEMPLATE_VERSION_NOT_FOUND", func() error { return ignore(client.Templates.GetVersion(context.Background(), "t1", 9)) }},
		{"TEMPLATE_VERSION_NOT_FOUND", func() error { return ignore(client.Templates.RestoreVersion(context.Background(), "t1", 9)) }},
	}

	for _, want := range calls {
		err := want.invoke()
		var apiErr *Error
		if !errors.Is(err, ErrNotFound) || !errors.As(err, &apiErr) || apiErr.APICode != want.apiCode {
			t.Fatalf("esperava ErrNotFound com %s, veio %v", want.apiCode, err)
		}
	}
}

func TestSendAcceptsATemplateAlias(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(200, `{"id":"e1"}`)})

	_, err := client.Emails.Send(context.Background(), &SendEmailParams{
		From:       Address{Email: "no-reply@acme.com"},
		To:         []Address{{Email: "ana@gmail.com"}},
		TemplateID: "boas-vindas",
		Variables:  map[string]any{"nome": "Ana"},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	var sent map[string]any
	if err := json.Unmarshal(transport.last(t).body, &sent); err != nil {
		t.Fatalf("corpo inválido: %v", err)
	}
	if sent["templateId"] != "boas-vindas" {
		t.Fatalf("templateId inesperado: %v", sent["templateId"])
	}
}

func TestValidatePostsTheEmailsAndDecodesTheCredits(t *testing.T) {
	result := `{"items":[{"email":"ana@gmail.com","formatValid":true,"mxValid":true,"isDisposable":false},` +
		`{"email":"x@mailinator.com","formatValid":true,"mxValid":true,"isDisposable":true,"reason":"disposable"}],` +
		`"validCount":1,"invalidCount":1,"credits":{"limit":1000,"used":12,"remaining":988}}`
	client, transport := newTestClient(t, []fakeResponse{
		respond(200, result),
		respond(200, `{"items":[],"validCount":0,"invalidCount":0,"credits":{"limit":null,"used":3,"remaining":null}}`),
	})

	got, err := client.Validation.Validate(context.Background(), []string{"ana@gmail.com", "x@mailinator.com"})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	request := transport.last(t)
	if request.method != "POST" || request.path != "/v1/product/validate" {
		t.Fatalf("requisição inesperada: %s %s", request.method, request.path)
	}
	if string(request.body) != `{"emails":["ana@gmail.com","x@mailinator.com"]}` {
		t.Fatalf("corpo inesperado: %s", request.body)
	}
	if got.ValidCount != 1 || got.InvalidCount != 1 || !got.Items[1].IsDisposable || got.Items[1].Reason != "disposable" || got.Items[0].Reason != "" {
		t.Fatalf("validação inesperada: %+v", got)
	}
	if got.Credits.Limit == nil || *got.Credits.Limit != 1000 || got.Credits.Remaining == nil || *got.Credits.Remaining != 988 {
		t.Fatalf("créditos inesperados: %+v", got.Credits)
	}

	unlimited, err := client.Validation.Validate(context.Background(), []string{"ana@gmail.com"})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if unlimited.Credits.Limit != nil || unlimited.Credits.Remaining != nil || unlimited.Credits.Used != 3 {
		t.Fatalf("créditos ilimitados inesperados: %+v", unlimited.Credits)
	}
}

func TestValidateMapsValidationAndExhaustedCredits(t *testing.T) {
	client, _ := newTestClient(t, []fakeResponse{
		respond(400, `{"error":{"code":"VALIDATION_ERROR","message":"emails must have at least 1 item"}}`),
		respond(429, `{"error":{"code":"VALIDATION_CREDITS_EXHAUSTED","message":"validation credits exhausted"}}`),
	})

	if _, err := client.Validation.Validate(context.Background(), nil); !errors.Is(err, ErrValidation) {
		t.Fatalf("esperava ErrValidation, veio %v", err)
	}
	_, err := client.Validation.Validate(context.Background(), []string{"ana@gmail.com"})
	var apiErr *Error
	if !errors.Is(err, ErrRateLimited) || !errors.As(err, &apiErr) || apiErr.APICode != "VALIDATION_CREDITS_EXHAUSTED" {
		t.Fatalf("esperava ErrRateLimited com VALIDATION_CREDITS_EXHAUSTED, veio %v", err)
	}
}

func TestRetryNeverRepeatsExhaustedValidationCredits(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{
		respond(429, `{"error":{"code":"VALIDATION_CREDITS_EXHAUSTED"}}`),
	}, fastRetry)

	_, err := client.Domains.List(context.Background())
	if !errors.Is(err, ErrRateLimited) || len(transport.all()) != 1 {
		t.Fatalf("err = %v, tentativas = %d", err, len(transport.all()))
	}
}

func TestGetDeliveryDecodesTheRequestPayload(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(200, webhookDeliveryJSON)})

	got, err := client.Webhooks.GetDelivery(context.Background(), "w1", "dl2")
	if err != nil {
		t.Fatalf("GetDelivery: %v", err)
	}
	if request := transport.last(t); request.method != "GET" || request.path != "/v1/product/webhooks/w1/deliveries/dl2" {
		t.Fatalf("requisição inesperada: %s %s", request.method, request.path)
	}
	if got.ID != "dl2" || got.Status != WebhookDeliveryPending || got.RequestPayload != `{"type":"email.delivered"}` {
		t.Fatalf("entrega inesperada: %+v", got)
	}
}

func TestReplayDeliveryPostsWithoutBodyAndReturnsTheNewDelivery(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(202, webhookDeliveryJSON)})

	got, err := client.Webhooks.ReplayDelivery(context.Background(), "w1", "dl1")
	if err != nil {
		t.Fatalf("ReplayDelivery: %v", err)
	}
	request := transport.last(t)
	if request.method != "POST" || request.path != "/v1/product/webhooks/w1/deliveries/dl1/replay" || len(request.body) != 0 {
		t.Fatalf("requisição inesperada: %s %s %s", request.method, request.path, request.body)
	}
	if got.ID != "dl2" || got.WebhookID != "w1" {
		t.Fatalf("entrega nova inesperada: %+v", got)
	}
}

func TestWebhookDeliveriesMapNotFoundAndConflicts(t *testing.T) {
	client, _ := newTestClient(t, []fakeResponse{
		respond(404, `{"error":{"code":"WEBHOOK_NOT_FOUND"}}`),
		respond(404, `{"error":{"code":"WEBHOOK_DELIVERY_NOT_FOUND"}}`),
		respond(409, `{"error":{"code":"WEBHOOK_DELIVERY_PENDING","message":"delivery is still pending"}}`),
		respond(409, `{"error":{"code":"WEBHOOK_NOT_ACTIVE","message":"webhook is not active"}}`),
	})
	calls := []struct {
		sentinel error
		apiCode  string
		invoke   func() error
	}{
		{ErrNotFound, "WEBHOOK_NOT_FOUND", func() error { return ignore(client.Webhooks.GetDelivery(context.Background(), "w_x", "dl1")) }},
		{ErrNotFound, "WEBHOOK_DELIVERY_NOT_FOUND", func() error { return ignore(client.Webhooks.ReplayDelivery(context.Background(), "w1", "dl_x")) }},
		{ErrConflict, "WEBHOOK_DELIVERY_PENDING", func() error { return ignore(client.Webhooks.ReplayDelivery(context.Background(), "w1", "dl1")) }},
		{ErrConflict, "WEBHOOK_NOT_ACTIVE", func() error { return ignore(client.Webhooks.ReplayDelivery(context.Background(), "w1", "dl1")) }},
	}

	for _, want := range calls {
		err := want.invoke()
		var apiErr *Error
		if !errors.Is(err, want.sentinel) || !errors.As(err, &apiErr) || apiErr.APICode != want.apiCode {
			t.Fatalf("esperava %v com %s, veio %v", want.sentinel, want.apiCode, err)
		}
	}
}
