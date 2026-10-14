// Package webhook verifica a autenticidade das entregas de webhook da CoffeeMail.
//
// A assinatura é "sha256=" seguido do HMAC-SHA256 em hexadecimal de
// timestamp + "." + idempotencyKey + "." + corpo bruto, usando o segredo do
// webhook (whsec_...). Entregas fora da janela de tolerância (padrão de 5 minutos)
// são rejeitadas para impedir replay. O pacote não depende do Client nem de API key.
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//		event, err := webhook.ParseRequest(r, os.Getenv("COFFEEMAIL_WEBHOOK_SECRET"))
//		if err != nil {
//			http.Error(w, "assinatura inválida", http.StatusUnauthorized)
//			return
//		}
//		...
//	}
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

// Nomes dos headers enviados em cada entrega.
const (
	HeaderSignature      = "X-Webhook-Signature"
	HeaderTimestamp      = "X-Webhook-Timestamp"
	HeaderIdempotencyKey = "X-Webhook-Idempotency-Key"
	HeaderTest           = "X-Webhook-Test"
)

// DefaultTolerance é a janela de tolerância padrão, igual à do servidor.
const DefaultTolerance = 5 * time.Minute

const (
	signaturePrefix = "sha256="
	separator       = "."
	maxBodyBytes    = 1 << 20
	decimalBase     = 10
	int64Bits       = 64
	testEnabled     = "true"
)

var unixSeconds = regexp.MustCompile(`^-?\d{1,18}$`)

// Erros de verificação. Não exponha o motivo para quem enviou a requisição;
// use-o só em logs.
var (
	ErrMissingHeader    = errors.New("webhook: header de assinatura ausente")
	ErrInvalidTimestamp = errors.New("webhook: timestamp inválido")
	ErrTimestampExpired = errors.New("webhook: timestamp fora da janela de tolerância")
	ErrInvalidSignature = errors.New("webhook: assinatura inválida")
	ErrInvalidPayload   = errors.New("webhook: corpo ilegível")
)

// Option configura a verificação.
type Option func(*verifier)

type verifier struct {
	tolerance time.Duration
	now       func() time.Time
}

// WithTolerance define a janela de tolerância do timestamp. Padrão: [DefaultTolerance].
func WithTolerance(tolerance time.Duration) Option {
	return func(v *verifier) {
		v.tolerance = tolerance
	}
}

// WithClock define o relógio usado na checagem da janela; útil em testes.
func WithClock(now func() time.Time) Option {
	return func(v *verifier) {
		v.now = now
	}
}

// Delivery são os dados de uma entrega necessários para verificá-la.
type Delivery struct {
	// Payload é o corpo bruto, exatamente como recebido. Nunca re-serialize o JSON.
	Payload        []byte
	Signature      string
	Timestamp      string
	IdempotencyKey string
}

// DeliveryFromHeaders monta a [Delivery] a partir do corpo bruto e dos headers HTTP.
func DeliveryFromHeaders(payload []byte, h http.Header) Delivery {
	return Delivery{
		Payload:        payload,
		Signature:      h.Get(HeaderSignature),
		Timestamp:      h.Get(HeaderTimestamp),
		IdempotencyKey: h.Get(HeaderIdempotencyKey),
	}
}

// Verify confere a assinatura e a janela de tolerância. Devolve nil quando a
// entrega é autêntica.
func Verify(delivery Delivery, secret string, opts ...Option) error {
	v := verifier{tolerance: DefaultTolerance, now: time.Now}
	for _, opt := range opts {
		opt(&v)
	}
	if delivery.Signature == "" || delivery.Timestamp == "" || delivery.IdempotencyKey == "" {
		return ErrMissingHeader
	}
	if err := v.checkTimestamp(delivery.Timestamp); err != nil {
		return err
	}
	if !hmac.Equal([]byte(Sign(delivery, secret)), []byte(delivery.Signature)) {
		return ErrInvalidSignature
	}
	return nil
}

// Sign calcula a assinatura esperada de uma entrega. Útil para testar seu handler.
func Sign(delivery Delivery, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(delivery.Timestamp + separator + delivery.IdempotencyKey + separator))
	mac.Write(delivery.Payload)
	return signaturePrefix + hex.EncodeToString(mac.Sum(nil))
}

func (v verifier) checkTimestamp(raw string) error {
	if !unixSeconds.MatchString(raw) {
		return ErrInvalidTimestamp
	}
	seconds, err := strconv.ParseInt(raw, decimalBase, int64Bits)
	if err != nil {
		return ErrInvalidTimestamp
	}
	drift := v.now().Unix() - seconds
	if max(drift, -drift) > int64(v.tolerance/time.Second) {
		return ErrTimestampExpired
	}
	return nil
}

// Event é o corpo de uma entrega de webhook. Data varia conforme EventType e
// não inclui o id do email.
type Event struct {
	EventType string         `json:"event_type"`
	WebhookID string         `json:"webhook_id"`
	Data      map[string]any `json:"data"`
	// Test indica um disparo de teste (header X-Webhook-Test).
	Test bool `json:"-"`
	// IdempotencyKey muda a cada tentativa de entrega; não serve para deduplicar
	// retentativas.
	IdempotencyKey string `json:"-"`
}

// ParseRequest lê o corpo da requisição (até 1 MiB), verifica a assinatura e
// devolve o evento.
func ParseRequest(r *http.Request, secret string, opts ...Option) (*Event, error) {
	payload, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes))
	if err != nil {
		return nil, ErrInvalidPayload
	}
	return Parse(payload, r.Header, secret, opts...)
}

// Parse verifica a assinatura de um corpo já lido e devolve o evento.
func Parse(payload []byte, h http.Header, secret string, opts ...Option) (*Event, error) {
	if err := Verify(DeliveryFromHeaders(payload, h), secret, opts...); err != nil {
		return nil, err
	}
	var event Event
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, ErrInvalidPayload
	}
	event.Test = h.Get(HeaderTest) == testEnabled
	event.IdempotencyKey = h.Get(HeaderIdempotencyKey)
	return &event, nil
}
