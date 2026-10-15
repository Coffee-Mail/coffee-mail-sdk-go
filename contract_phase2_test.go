package coffeemail

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestResendVerificationPostsToTheResendRoute(t *testing.T) {
	client, transport := newTestClient(t, []fakeResponse{respond(204, "")})

	if err := client.Senders.ResendVerification(context.Background(), "s_1"); err != nil {
		t.Fatalf("ResendVerification: %v", err)
	}

	request := transport.last(t)
	if request.method != "POST" || request.path != "/v1/product/senders/s_1/resend-verification" {
		t.Fatalf("requisição inesperada: %s %s", request.method, request.path)
	}
}

func TestResendVerificationMapsConflictAndNotFound(t *testing.T) {
	client, _ := newTestClient(t, []fakeResponse{
		respond(409, `{"error":{"code":"CONFLICT","message":"sender is already verified"}}`),
		respond(404, `{"error":{"code":"NOT_FOUND"}}`),
	})

	if err := client.Senders.ResendVerification(context.Background(), "s_1"); !errors.Is(err, ErrConflict) {
		t.Fatalf("esperava ErrConflict, veio %v", err)
	}
	if err := client.Senders.ResendVerification(context.Background(), "s_x"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("esperava ErrNotFound, veio %v", err)
	}
}

func TestCreateBroadcastSendsScheduledAtAndGetDecodesMetrics(t *testing.T) {
	detail := `{"id":"b1","audienceId":"a1","status":"sent","createdAt":"2026-10-15T12:00:00Z",` +
		`"updatedAt":"2026-10-15T12:00:00Z","metrics":{"delivered":9,"opened":4,"clicked":1,` +
		`"bounced":1,"failed":0,"suppressed":2}}`
	client, transport := newTestClient(t, []fakeResponse{respond(201, detail), respond(200, detail)})
	at := time.Date(2030, 1, 1, 10, 0, 0, 0, time.UTC)

	_, err := client.Broadcasts.Create(context.Background(), &CreateBroadcastParams{
		AudienceID: "a1", FromEmail: "news@acme.com", TemplateID: "tpl_1", ScheduledAt: at,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	var sent map[string]any
	if err := json.Unmarshal(transport.last(t).body, &sent); err != nil {
		t.Fatalf("corpo inválido: %v", err)
	}
	if sent["scheduledAt"] != "2030-01-01T10:00:00Z" {
		t.Fatalf("scheduledAt inesperado: %v", sent["scheduledAt"])
	}

	got, err := client.Broadcasts.Get(context.Background(), "b1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Metrics == nil || got.Metrics.Delivered != 9 {
		t.Fatalf("metrics inesperado: %+v", got.Metrics)
	}
}

func TestContactSubscriptionDefaultsToSubscribed(t *testing.T) {
	unsubscribed := false
	if (Contact{}).IsSubscribed() != true {
		t.Fatal("contato sem o campo deveria contar como inscrito")
	}
	if (Contact{Subscribed: &unsubscribed}).IsSubscribed() {
		t.Fatal("contato descadastrado deveria retornar false")
	}
}
