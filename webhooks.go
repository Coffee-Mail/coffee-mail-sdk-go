package coffeemail

import (
	"context"
	"time"

	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/routes"
)

// WebhookEvent é um tipo de evento entregue por webhook.
type WebhookEvent string

// Eventos de webhook.
const (
	WebhookEventEmailQueued          WebhookEvent = "email.queued"
	WebhookEventEmailProcessing      WebhookEvent = "email.processing"
	WebhookEventEmailSent            WebhookEvent = "email.sent"
	WebhookEventEmailDelivered       WebhookEvent = "email.delivered"
	WebhookEventEmailDeliveryDelayed WebhookEvent = "email.delivery_delayed"
	WebhookEventEmailOpened          WebhookEvent = "email.opened"
	WebhookEventEmailClicked         WebhookEvent = "email.clicked"
	WebhookEventEmailBounced         WebhookEvent = "email.bounced"
	WebhookEventEmailComplained      WebhookEvent = "email.complained"
	WebhookEventEmailFailed          WebhookEvent = "email.failed"
	WebhookEventEmailSuppressed      WebhookEvent = "email.suppressed"
	WebhookEventEmailScheduled       WebhookEvent = "email.scheduled"
	WebhookEventEmailCancelled       WebhookEvent = "email.cancelled"
	WebhookEventEmailUnsubscribed    WebhookEvent = "email.unsubscribed"
	WebhookEventBillingPlanChanged   WebhookEvent = "billing.plan_changed"
)

// WebhookStatus é o estado de um webhook.
type WebhookStatus string

// Estados de webhook. Só Active e Paused podem ser definidos com Toggle.
const (
	WebhookStatusActive   WebhookStatus = "active"
	WebhookStatusPaused   WebhookStatus = "paused"
	WebhookStatusDisabled WebhookStatus = "disabled"
)

// WebhookDeliveryStatus é o estado de uma entrega de webhook.
type WebhookDeliveryStatus string

// Estados de entrega de webhook.
const (
	WebhookDeliveryPending   WebhookDeliveryStatus = "pending"
	WebhookDeliverySuccess   WebhookDeliveryStatus = "success"
	WebhookDeliveryFailed    WebhookDeliveryStatus = "failed"
	WebhookDeliveryExhausted WebhookDeliveryStatus = "exhausted"
)

// WebhookTestStatus é o resultado de um disparo de teste.
type WebhookTestStatus string

// Resultados de teste de webhook.
const (
	WebhookTestSuccess WebhookTestStatus = "success"
	WebhookTestFailed  WebhookTestStatus = "failed"
)

// CreateWebhookParams são os dados de um webhook novo.
type CreateWebhookParams struct {
	URL         string         `json:"url"`
	Events      []WebhookEvent `json:"events"`
	Description string         `json:"description,omitzero"`
}

// UpdateWebhookParams altera um webhook. Campos vazios não são alterados.
type UpdateWebhookParams struct {
	URL         *string        `json:"url,omitzero"`
	Events      []WebhookEvent `json:"events,omitzero"`
	Description *string        `json:"description,omitzero"`
}

// Webhook é um webhook na listagem.
type Webhook struct {
	ID          string         `json:"id"`
	URL         string         `json:"url"`
	Events      []WebhookEvent `json:"events"`
	Status      WebhookStatus  `json:"status"`
	Description *string        `json:"description"`
	HasSecret   bool           `json:"hasSecret"`
	CreatedAt   time.Time      `json:"createdAt"`
}

// WebhookDetail é o detalhe de um webhook.
type WebhookDetail struct {
	ID            string         `json:"id"`
	URL           string         `json:"url"`
	Events        []WebhookEvent `json:"events"`
	Status        WebhookStatus  `json:"status"`
	Description   *string        `json:"description"`
	CreatedAt     time.Time      `json:"createdAt"`
	Secret        *string        `json:"secret"`
	SecretPreview *string        `json:"secretPreview"`
}

// CreatedWebhook é o webhook criado. Secret só é devolvido nesta resposta:
// guarde-o para verificar as assinaturas.
type CreatedWebhook struct {
	ID          string         `json:"id"`
	URL         string         `json:"url"`
	Events      []WebhookEvent `json:"events"`
	Status      WebhookStatus  `json:"status"`
	Description *string        `json:"description"`
	CreatedAt   time.Time      `json:"createdAt"`
	Secret      string         `json:"secret"`
}

// ListWebhooksParams filtra a listagem de webhooks.
type ListWebhooksParams struct {
	Status WebhookStatus
}

// ListWebhooksResponse lista os webhooks.
type ListWebhooksResponse struct {
	Webhooks []Webhook `json:"webhooks"`
}

// ToggleWebhookResponse é o novo estado do webhook.
type ToggleWebhookResponse struct {
	ID     string        `json:"id"`
	Status WebhookStatus `json:"status"`
}

// RotateWebhookSecretResponse traz o novo segredo de assinatura.
type RotateWebhookSecretResponse struct {
	WebhookID string `json:"webhookId"`
	Secret    string `json:"secret"`
}

// WebhookDelivery é uma tentativa de entrega de webhook.
type WebhookDelivery struct {
	ID             string                `json:"id"`
	WebhookID      string                `json:"webhookId"`
	EmailEventID   *string               `json:"emailEventId"`
	EventType      WebhookEvent          `json:"eventType"`
	ResponseStatus *int                  `json:"responseStatus"`
	ResponseBody   *string               `json:"responseBody"`
	Attempts       int                   `json:"attempts"`
	LastError      *string               `json:"lastError"`
	Status         WebhookDeliveryStatus `json:"status"`
	NextRunAt      time.Time             `json:"nextRunAt"`
	CreatedAt      time.Time             `json:"createdAt"`
}

// ListWebhookDeliveriesParams filtra as entregas de um webhook.
type ListWebhookDeliveriesParams struct {
	EventType WebhookEvent
	Status    WebhookDeliveryStatus
	From      time.Time
	To        time.Time
	// Limit é a quantidade máxima de entregas (máximo 200).
	Limit int
}

// ListWebhookDeliveriesResponse lista as entregas de um webhook.
type ListWebhookDeliveriesResponse struct {
	Deliveries []WebhookDelivery `json:"deliveries"`
}

// TestWebhookResponse é o resultado de um disparo de teste.
type TestWebhookResponse struct {
	Status     WebhookTestStatus `json:"status"`
	StatusCode *int              `json:"statusCode"`
	Body       *string           `json:"body"`
	Error      *string           `json:"error"`
}

type toggleWebhookBody struct {
	Status WebhookStatus `json:"status"`
}

// WebhooksService gerencia endpoints de webhook. Para verificar assinaturas de
// entregas recebidas, use o subpacote webhook.
type WebhooksService struct {
	r *requester
}

// Create cadastra um webhook. O segredo de assinatura vem só nesta resposta.
func (s *WebhooksService) Create(ctx context.Context, params *CreateWebhookParams) (*CreatedWebhook, error) {
	return call[CreatedWebhook](ctx, s.r, post(routes.Webhooks, derefOrZero(params)))
}

// List lista os webhooks. params pode ser nil.
func (s *WebhooksService) List(ctx context.Context, params *ListWebhooksParams) (*ListWebhooksResponse, error) {
	query := newQuery().str("status", string(derefOrZero(params).Status)).values()
	return call[ListWebhooksResponse](ctx, s.r, get(routes.Webhooks).withQuery(query))
}

// Get devolve um webhook.
func (s *WebhooksService) Get(ctx context.Context, id string) (*WebhookDetail, error) {
	return call[WebhookDetail](ctx, s.r, get(routes.Expand(routes.Webhook, id)))
}

// Update altera um webhook.
func (s *WebhooksService) Update(ctx context.Context, id string, params *UpdateWebhookParams) (*WebhookDetail, error) {
	return call[WebhookDetail](ctx, s.r, put(routes.Expand(routes.Webhook, id), derefOrZero(params)))
}

// Toggle ativa ou pausa um webhook.
func (s *WebhooksService) Toggle(ctx context.Context, id string, status WebhookStatus) (*ToggleWebhookResponse, error) {
	return call[ToggleWebhookResponse](ctx, s.r, patch(routes.Expand(routes.Webhook, id), toggleWebhookBody{Status: status}))
}

// Delete exclui um webhook.
func (s *WebhooksService) Delete(ctx context.Context, id string) error {
	return s.r.do(ctx, del(routes.Expand(routes.Webhook, id)), nil)
}

// RotateSecret gera um novo segredo de assinatura e invalida o anterior.
func (s *WebhooksService) RotateSecret(ctx context.Context, id string) (*RotateWebhookSecretResponse, error) {
	return call[RotateWebhookSecretResponse](ctx, s.r, post(routes.Expand(routes.WebhookRotateSecret, id), nil))
}

// ListDeliveries lista as entregas de um webhook. params pode ser nil.
func (s *WebhooksService) ListDeliveries(ctx context.Context, id string, params *ListWebhookDeliveriesParams) (*ListWebhookDeliveriesResponse, error) {
	v := derefOrZero(params)
	query := newQuery().
		str("eventType", string(v.EventType)).
		str("status", string(v.Status)).
		time("from", v.From).
		time("to", v.To).
		num("limit", v.Limit).
		values()
	return call[ListWebhookDeliveriesResponse](ctx, s.r, get(routes.Expand(routes.WebhookDeliveries, id)).withQuery(query))
}

// Test dispara um evento de teste para o webhook.
func (s *WebhooksService) Test(ctx context.Context, id string) (*TestWebhookResponse, error) {
	return call[TestWebhookResponse](ctx, s.r, post(routes.Expand(routes.WebhookTest, id), nil))
}
