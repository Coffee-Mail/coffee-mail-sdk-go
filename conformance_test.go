package coffeemail

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/mail"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

const (
	uuidA = "6f1c2a7e-3b4d-4e5f-8a9b-0c1d2e3f4a5b"
	uuidB = "7a2d3b8f-4c5e-4f60-9b0c-1d2e3f4a5b6c"
)

var browserOnlyRoutes = map[string]bool{"GET /v1/product/auth/csrf": true}

var optionalBodyRoutes = map[string]string{
	"DELETE /v1/product/domains/{id}": "o spec marca o corpo de reautenticação como obrigatório, mas a API só o exige em sessões do dashboard",
}

type conformanceCase struct {
	name   string
	body   string
	invoke func(ctx context.Context, c *Client) error
}

func ignore[T any](_ T, err error) error {
	return err
}

func conformanceCases() []conformanceCase {
	scheduled := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	email := &SendEmailParams{
		From:        Address{Email: "no-reply@exemplo.com.br", Name: "Exemplo"},
		To:          []Address{{Email: "cliente@gmail.com"}},
		Subject:     "Olá",
		HTML:        "<p>Olá</p>",
		Tags:        []Tag{{Name: "fluxo", Value: "boas-vindas"}},
		ScheduledAt: scheduled,
		Attachments: []Attachment{{Filename: "a.txt", Content: "YQ==", ContentType: "text/plain", Disposition: AttachmentDispositionAttachment}},
		Sandbox:     true,
	}
	return []conformanceCase{
		{"introspect", "", func(ctx context.Context, c *Client) error { return ignore(c.Introspect(ctx)) }},
		{"emails.send", "", func(ctx context.Context, c *Client) error { return ignore(c.Emails.Send(ctx, email)) }},
		{"emails.sendBatch", "[]", func(ctx context.Context, c *Client) error {
			return ignore(c.Emails.SendBatch(ctx, &SendBatchParams{Emails: []SendEmailParams{*email}}))
		}},
		{"emails.list", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Emails.List(ctx, &ListEmailsParams{Status: EmailStatusDelivered, From: scheduled, Limit: 10}))
		}},
		{"emails.get", "", func(ctx context.Context, c *Client) error { return ignore(c.Emails.Get(ctx, uuidA)) }},
		{"emails.getEvents", "", func(ctx context.Context, c *Client) error { return ignore(c.Emails.GetEvents(ctx, uuidA)) }},
		{"emails.getTags", "", func(ctx context.Context, c *Client) error { return ignore(c.Emails.GetTags(ctx)) }},
		{"emails.cancel", "", func(ctx context.Context, c *Client) error { return c.Emails.Cancel(ctx, uuidA) }},
		{"emails.resend", "", func(ctx context.Context, c *Client) error { return ignore(c.Emails.Resend(ctx, uuidA)) }},
		{"domains.create", "", func(ctx context.Context, c *Client) error { return ignore(c.Domains.Create(ctx, "exemplo.com.br")) }},
		{"domains.list", "", func(ctx context.Context, c *Client) error { return ignore(c.Domains.List(ctx)) }},
		{"domains.get", "", func(ctx context.Context, c *Client) error { return ignore(c.Domains.Get(ctx, uuidA)) }},
		{"domains.verify", "", func(ctx context.Context, c *Client) error { return ignore(c.Domains.Verify(ctx, uuidA)) }},
		{"domains.delete", "", func(ctx context.Context, c *Client) error { return ignore(c.Domains.Delete(ctx, uuidA)) }},
		{"domains.getHealth", "", func(ctx context.Context, c *Client) error { return ignore(c.Domains.GetHealth(ctx, uuidA)) }},
		{"domains.getWarmupStatus", "", func(ctx context.Context, c *Client) error { return ignore(c.Domains.GetWarmupStatus(ctx, uuidA)) }},
		{"domains.rotateDkim", "", func(ctx context.Context, c *Client) error { return ignore(c.Domains.RotateDKIM(ctx, uuidA)) }},
		{"domains.setTrackingDomain", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Domains.SetTrackingDomain(ctx, uuidA, "links.exemplo.com.br"))
		}},
		{"domains.removeTrackingDomain", "null", func(ctx context.Context, c *Client) error {
			return ignore(c.Domains.RemoveTrackingDomain(ctx, uuidA))
		}},
		{"templates.create", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Templates.Create(ctx, &CreateTemplateParams{Name: "Boas-vindas", HTML: "<p>{{nome}}</p>", Format: TemplateFormatHTML, Variables: []TemplateVariable{{Name: "nome"}}}))
		}},
		{"templates.list", "", func(ctx context.Context, c *Client) error { return ignore(c.Templates.List(ctx)) }},
		{"templates.get", "", func(ctx context.Context, c *Client) error { return ignore(c.Templates.Get(ctx, uuidA)) }},
		{"templates.update", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Templates.Update(ctx, uuidA, &UpdateTemplateParams{Name: Ptr("Novo"), IsActive: Ptr(false)}))
		}},
		{"templates.delete", "", func(ctx context.Context, c *Client) error { return c.Templates.Delete(ctx, uuidA) }},
		{"templates.preview", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Templates.Preview(ctx, &PreviewTemplateParams{HTML: "<p>x</p>", Variables: map[string]any{"nome": "Ana"}}))
		}},
		{"templates.format", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Templates.Format(ctx, &FormatTemplateParams{HTML: "<p>x</p>"}))
		}},
		{"templates.testRender", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Templates.TestRender(ctx, &TestRenderTemplateParams{HTML: "<p>x</p>"}))
		}},
		{"templates.testSend", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Templates.TestSend(ctx, uuidA, &TestSendTemplateParams{To: "teste@gmail.com"}))
		}},
		{"templates.listVersions", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Templates.ListVersions(ctx, uuidA, &OffsetParams{Limit: 20, Offset: 20}))
		}},
		{"templates.getVersion", "", func(ctx context.Context, c *Client) error { return ignore(c.Templates.GetVersion(ctx, uuidA, 2)) }},
		{"templates.restoreVersion", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Templates.RestoreVersion(ctx, uuidA, 2))
		}},
		{"audiences.create", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Audiences.Create(ctx, &CreateAudienceParams{Name: "Clientes"}))
		}},
		{"audiences.list", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Audiences.List(ctx, &OffsetParams{Limit: 20, Offset: 40}))
		}},
		{"audiences.get", "", func(ctx context.Context, c *Client) error { return ignore(c.Audiences.Get(ctx, uuidA)) }},
		{"audiences.update", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Audiences.Update(ctx, uuidA, &UpdateAudienceParams{Active: Ptr(false)}))
		}},
		{"audiences.delete", "", func(ctx context.Context, c *Client) error { return c.Audiences.Delete(ctx, uuidA) }},
		{"contacts.create", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Audiences.Contacts.Create(ctx, uuidA, &CreateContactParams{Email: "ana@gmail.com", FirstName: "Ana"}))
		}},
		{"contacts.list", "", func(ctx context.Context, c *Client) error { return ignore(c.Audiences.Contacts.List(ctx, uuidA, nil)) }},
		{"contacts.bulkAdd", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Audiences.Contacts.BulkAdd(ctx, uuidA, []CreateContactParams{{Email: "ana@gmail.com"}}))
		}},
		{"contacts.update", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Audiences.Contacts.Update(ctx, uuidA, uuidB, &UpdateContactParams{FirstName: Ptr("Ana")}))
		}},
		{"contacts.delete", "", func(ctx context.Context, c *Client) error { return c.Audiences.Contacts.Delete(ctx, uuidA, uuidB) }},
		{"broadcasts.create", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Broadcasts.Create(ctx, &CreateBroadcastParams{AudienceID: uuidA, FromEmail: "news@exemplo.com.br", TemplateID: uuidB}))
		}},
		{"broadcasts.list", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Broadcasts.List(ctx, &ListBroadcastsParams{Status: BroadcastStatusSent}))
		}},
		{"broadcasts.get", "", func(ctx context.Context, c *Client) error { return ignore(c.Broadcasts.Get(ctx, uuidA)) }},
		{"broadcasts.send", "", func(ctx context.Context, c *Client) error { return ignore(c.Broadcasts.Send(ctx, uuidA)) }},
		{"broadcasts.testSend", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Broadcasts.TestSend(ctx, uuidA, &TestSendBroadcastParams{To: "teste@gmail.com"}))
		}},
		{"broadcasts.cancel", "", func(ctx context.Context, c *Client) error { return ignore(c.Broadcasts.Cancel(ctx, uuidA)) }},
		{"senders.create", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Senders.Create(ctx, &CreateSenderParams{Email: "ana@exemplo.com.br"}))
		}},
		{"senders.list", "", func(ctx context.Context, c *Client) error { return ignore(c.Senders.List(ctx)) }},
		{"senders.verify", "", func(ctx context.Context, c *Client) error { return ignore(c.Senders.Verify(ctx, uuidA)) }},
		{"senders.resendVerification", "", func(ctx context.Context, c *Client) error {
			return c.Senders.ResendVerification(ctx, uuidA)
		}},
		{"senders.delete", "", func(ctx context.Context, c *Client) error { return c.Senders.Delete(ctx, uuidA) }},
		{"suppressions.create", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Suppressions.Create(ctx, &CreateSuppressionParams{Email: "x@gmail.com", Reason: SuppressionReasonManual}))
		}},
		{"suppressions.list", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Suppressions.List(ctx, &ListSuppressionsParams{Category: SuppressionCategoryBounce, Limit: 50}))
		}},
		{"suppressions.get", "", func(ctx context.Context, c *Client) error { return ignore(c.Suppressions.Get(ctx, uuidA)) }},
		{"suppressions.delete", "", func(ctx context.Context, c *Client) error { return c.Suppressions.Delete(ctx, uuidA) }},
		{"suppressions.reactivate", "", func(ctx context.Context, c *Client) error { return c.Suppressions.Reactivate(ctx, uuidA) }},
		{"suppressions.bulkCreate", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Suppressions.BulkCreate(ctx, []BulkSuppressionEntry{{Email: "x@gmail.com"}}))
		}},
		{"webhooks.create", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Webhooks.Create(ctx, &CreateWebhookParams{URL: "https://exemplo.com.br/hook", Events: []WebhookEvent{WebhookEventEmailDelivered}}))
		}},
		{"webhooks.list", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Webhooks.List(ctx, &ListWebhooksParams{Status: WebhookStatusActive}))
		}},
		{"webhooks.get", "", func(ctx context.Context, c *Client) error { return ignore(c.Webhooks.Get(ctx, uuidA)) }},
		{"webhooks.update", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Webhooks.Update(ctx, uuidA, &UpdateWebhookParams{Description: Ptr("principal")}))
		}},
		{"webhooks.toggle", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Webhooks.Toggle(ctx, uuidA, WebhookStatusPaused))
		}},
		{"webhooks.delete", "", func(ctx context.Context, c *Client) error { return c.Webhooks.Delete(ctx, uuidA) }},
		{"webhooks.rotateSecret", "", func(ctx context.Context, c *Client) error { return ignore(c.Webhooks.RotateSecret(ctx, uuidA)) }},
		{"webhooks.listDeliveries", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Webhooks.ListDeliveries(ctx, uuidA, &ListWebhookDeliveriesParams{Status: WebhookDeliveryFailed, Limit: 5}))
		}},
		{"webhooks.test", "", func(ctx context.Context, c *Client) error { return ignore(c.Webhooks.Test(ctx, uuidA)) }},
		{"webhooks.getDelivery", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Webhooks.GetDelivery(ctx, uuidA, uuidB))
		}},
		{"webhooks.replayDelivery", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Webhooks.ReplayDelivery(ctx, uuidA, uuidB))
		}},
		{"validation.validate", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Validation.Validate(ctx, []string{"ana@gmail.com", "nao-e-email"}))
		}},
		{"emails.sendWithTemplateAlias", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Emails.Send(ctx, &SendEmailParams{
				From: Address{Email: "no-reply@exemplo.com.br"}, To: []Address{{Email: "cliente@gmail.com"}},
				TemplateID: "boas-vindas", Variables: map[string]any{"nome": "Ana"},
			}))
		}},
		{"stats.get", "", func(ctx context.Context, c *Client) error {
			return ignore(c.Stats.Get(ctx, &GetStatsParams{Period: StatsPeriodLast30Days, Granularity: StatsGranularityDay}))
		}},
	}
}

var defineFormats = sync.OnceFunc(func() {
	openapi3.DefineStringFormatValidator("email", openapi3.NewCallbackValidator(func(value string) error {
		_, err := mail.ParseAddress(value)
		return err
	}))
	openapi3.DefineStringFormatValidator("uuid", openapi3.NewRegexpFormatValidator(openapi3.FormatOfStringForUUIDOfRFC4122))
})

func loadSpecRouter(t *testing.T) (*openapi3.T, routers.Router) {
	t.Helper()
	defineFormats()
	raw, err := os.ReadFile("spec/product.json")
	if err != nil {
		t.Fatalf("lendo spec: %v", err)
	}
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(raw)
	if err != nil {
		t.Fatalf("carregando spec: %v", err)
	}
	doc.Servers = openapi3.Servers{{URL: "https://api.coffeemail.com.br"}}
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		t.Fatalf("roteador: %v", err)
	}
	return doc, router
}

func specOperations(doc *openapi3.T) []string {
	var ops []string
	for path, item := range doc.Paths.Map() {
		for method := range item.Operations() {
			key := method + " " + path
			if !browserOnlyRoutes[key] {
				ops = append(ops, key)
			}
		}
	}
	sort.Strings(ops)
	return ops
}

func TestConformanceEveryRequestMatchesTheSpec(t *testing.T) {
	doc, router := loadSpecRouter(t)
	covered := map[string]bool{}
	for _, tc := range conformanceCases() {
		t.Run(tc.name, func(t *testing.T) {
			body := tc.body
			if body == "" {
				body = "{}"
			}
			client, transport := newTestClient(t, []fakeResponse{respond(http.StatusOK, body)})
			if err := tc.invoke(context.Background(), client); err != nil {
				t.Fatalf("chamada falhou: %v", err)
			}
			key, err := validateAgainstSpec(router, transport.last(t))
			if err != nil {
				t.Fatal(err)
			}
			covered[key] = true
		})
	}
	var missing []string
	for _, op := range specOperations(doc) {
		if !covered[op] {
			missing = append(missing, op)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("operações do spec sem método no SDK:\n%s", strings.Join(missing, "\n"))
	}
}

func validateAgainstSpec(router routers.Router, recorded recordedRequest) (string, error) {
	req, err := http.NewRequest(recorded.method, recorded.url, bytes.NewReader(recorded.body))
	if err != nil {
		return "", err
	}
	req.Header = recorded.header
	route, params, err := router.FindRoute(req)
	if err != nil {
		return "", fmt.Errorf("%s %s não existe no spec: %w", recorded.method, recorded.path, err)
	}
	key := recorded.method + " " + route.Path
	_, bodyOptional := optionalBodyRoutes[key]
	input := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: params,
		Route:      route,
		Options: &openapi3filter.Options{
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
			ExcludeRequestBody: bodyOptional && len(recorded.body) == 0,
			SchemaValidationOptions: []openapi3.SchemaValidationOption{
				openapi3.DisablePatternValidation(),
			},
		},
	}
	if err := openapi3filter.ValidateRequest(context.Background(), input); err != nil {
		return key, fmt.Errorf("requisição fora do contrato: %w\ncorpo: %s", err, recorded.body)
	}
	return key, nil
}

func TestConformanceRejectsRequestsOutsideTheContract(t *testing.T) {
	_, router := loadSpecRouter(t)
	cases := map[string]recordedRequest{
		"campo obrigatório ausente": {method: http.MethodPost, url: "https://api.coffeemail.com.br/v1/product/emails", path: "/v1/product/emails", body: []byte(`{"from":{"email":"a@b.com"}}`)},
		"enum inválido":             {method: http.MethodPatch, url: "https://api.coffeemail.com.br/v1/product/webhooks/" + uuidA, path: "/v1/product/webhooks/" + uuidA, body: []byte(`{"status":"disabled"}`)},
		"email malformado":          {method: http.MethodPost, url: "https://api.coffeemail.com.br/v1/product/suppressions", path: "/v1/product/suppressions", body: []byte(`{"email":"nao-e-email"}`)},
		"rota inexistente":          {method: http.MethodDelete, url: "https://api.coffeemail.com.br/v1/product/broadcasts/" + uuidA, path: "/v1/product/broadcasts/" + uuidA},
	}
	for name, recorded := range cases {
		t.Run(name, func(t *testing.T) {
			recorded.header = http.Header{"Content-Type": {"application/json"}}
			if _, err := validateAgainstSpec(router, recorded); err == nil {
				t.Fatal("violação de contrato não detectada")
			}
		})
	}
}
