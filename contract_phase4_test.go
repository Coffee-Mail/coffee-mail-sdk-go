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
