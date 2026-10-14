package webhook

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

type vector struct {
	Name             string `json:"name"`
	Secret           string `json:"secret"`
	Body             string `json:"body"`
	Timestamp        string `json:"timestamp"`
	IdempotencyKey   string `json:"idempotencyKey"`
	Signature        string `json:"signature"`
	NowEpochSeconds  int64  `json:"nowEpochSeconds"`
	ToleranceSeconds int64  `json:"toleranceSeconds"`
	Expected         bool   `json:"expected"`
}

func loadVectors(t *testing.T) []vector {
	t.Helper()
	raw, err := os.ReadFile("../testdata/webhook_vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var file struct {
		Vectors []vector `json:"vectors"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	if len(file.Vectors) == 0 {
		t.Fatal("nenhum vetor carregado")
	}
	return file.Vectors
}

func TestSharedVectors(t *testing.T) {
	for _, v := range loadVectors(t) {
		t.Run(v.Name, func(t *testing.T) {
			delivery := Delivery{Payload: []byte(v.Body), Signature: v.Signature, Timestamp: v.Timestamp, IdempotencyKey: v.IdempotencyKey}
			err := Verify(delivery, v.Secret,
				WithTolerance(time.Duration(v.ToleranceSeconds)*time.Second),
				WithClock(func() time.Time { return time.Unix(v.NowEpochSeconds, 0) }))
			if (err == nil) != v.Expected {
				t.Fatalf("esperado válido=%v, erro=%v", v.Expected, err)
			}
		})
	}
}

const testSecret = "whsec_x"

func signedRequest(t *testing.T, body string) *http.Request {
	t.Helper()
	delivery := Delivery{Payload: []byte(body), Timestamp: "1760000000", IdempotencyKey: "ed4e7012eb856296"}
	req := httptest.NewRequest(http.MethodPost, "/webhooks/coffeemail", strings.NewReader(body))
	req.Header.Set(HeaderSignature, Sign(delivery, testSecret))
	req.Header.Set(HeaderTimestamp, delivery.Timestamp)
	req.Header.Set(HeaderIdempotencyKey, delivery.IdempotencyKey)
	req.Header.Set(HeaderTest, "true")
	return req
}

func TestParseRequestReturnsTheEvent(t *testing.T) {
	now := time.Unix(1760000000, 0)
	req := signedRequest(t, `{"event_type":"email.delivered","webhook_id":"w1","data":{"messageId":"m1"}}`)
	event, err := ParseRequest(req, testSecret, WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	if event.EventType != "email.delivered" || event.WebhookID != "w1" || event.Data["messageId"] != "m1" || !event.Test || event.IdempotencyKey != "ed4e7012eb856296" {
		t.Fatalf("evento = %+v", event)
	}
}

func TestParseRejectsInvalidDeliveries(t *testing.T) {
	now := time.Unix(1760000000, 0)
	clock := WithClock(func() time.Time { return now })
	cases := map[string]struct {
		request *http.Request
		secret  string
		want    error
	}{
		"segredo errado": {signedRequest(t, `{}`), "whsec_y", ErrInvalidSignature},
		"corpo não JSON": {signedRequest(t, `nao-json`), testSecret, ErrInvalidPayload},
		"sem headers":    {httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`)), testSecret, ErrMissingHeader},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseRequest(tc.request, tc.secret, clock); !errors.Is(err, tc.want) {
				t.Fatalf("erro = %v, esperado %v", err, tc.want)
			}
		})
	}
	expired := signedRequest(t, `{}`)
	if _, err := ParseRequest(expired, testSecret); !errors.Is(err, ErrTimestampExpired) {
		t.Fatalf("relógio real deveria rejeitar timestamp antigo: %v", err)
	}
	invalid := signedRequest(t, `{}`)
	invalid.Header.Set(HeaderTimestamp, "99999999999999999999")
	if _, err := ParseRequest(invalid, testSecret, clock); !errors.Is(err, ErrInvalidTimestamp) {
		t.Fatalf("timestamp fora do intervalo: %v", err)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("falha de leitura")
}

func TestParseRequestReportsUnreadableBodies(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", failingReader{})
	if _, err := ParseRequest(req, testSecret); !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf("erro = %v", err)
	}
}
