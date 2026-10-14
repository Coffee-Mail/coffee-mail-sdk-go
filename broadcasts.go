package coffeemail

import (
	"context"
	"iter"
	"time"

	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/routes"
)

const (
	opBroadcastsCreate   = "broadcasts.create"
	opBroadcastsList     = "broadcasts.list"
	opBroadcastsGet      = "broadcasts.get"
	opBroadcastsSend     = "broadcasts.send"
	opBroadcastsTestSend = "broadcasts.testSend"
	opBroadcastsCancel   = "broadcasts.cancel"
)

// BroadcastStatus é o estado de um broadcast.
type BroadcastStatus string

// Estados de broadcast.
const (
	BroadcastStatusDraft     BroadcastStatus = "draft"
	BroadcastStatusScheduled BroadcastStatus = "scheduled"
	BroadcastStatusSending   BroadcastStatus = "sending"
	BroadcastStatusSent      BroadcastStatus = "sent"
	BroadcastStatusFailed    BroadcastStatus = "failed"
	BroadcastStatusCancelled BroadcastStatus = "cancelled"
)

// CreateBroadcastParams são os dados de um broadcast novo. Informe TemplateID ou
// Subject com HTML/Text.
type CreateBroadcastParams struct {
	AudienceID string         `json:"audienceId"`
	FromEmail  string         `json:"fromEmail"`
	TemplateID string         `json:"templateId,omitzero"`
	Subject    string         `json:"subject,omitzero"`
	HTML       string         `json:"html,omitzero"`
	Text       string         `json:"text,omitzero"`
	ReplyTo    string         `json:"replyTo,omitzero"`
	Variables  map[string]any `json:"variables,omitzero"`
}

// Broadcast é um envio em massa para uma audiência.
type Broadcast struct {
	ID               string          `json:"id"`
	AudienceID       string          `json:"audienceId"`
	AudienceName     *string         `json:"audienceName"`
	TemplateID       *string         `json:"templateId"`
	Subject          string          `json:"subject"`
	FromEmail        string          `json:"fromEmail"`
	ReplyTo          *string         `json:"replyTo"`
	Status           BroadcastStatus `json:"status"`
	ScheduledAt      *time.Time      `json:"scheduledAt"`
	SentAt           *time.Time      `json:"sentAt"`
	CancelledAt      *time.Time      `json:"cancelledAt"`
	RecipientsTotal  int             `json:"recipientsTotal"`
	RecipientsQueued int             `json:"recipientsQueued"`
	RecipientsFailed int             `json:"recipientsFailed"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
}

// ListBroadcastsParams filtra a listagem de broadcasts. Todos os campos são opcionais.
type ListBroadcastsParams struct {
	Status BroadcastStatus
	From   time.Time
	To     time.Time
	// After é o cursor devolvido em NextCursor da página anterior.
	After string
	// Limit é o tamanho da página (máximo 200).
	Limit int
}

// ListBroadcastsResponse é uma página de broadcasts.
type ListBroadcastsResponse struct {
	Broadcasts []Broadcast `json:"broadcasts"`
	NextCursor *string     `json:"nextCursor"`
}

// SendBroadcastResponse é o resultado do disparo de um broadcast.
type SendBroadcastResponse struct {
	ID               string          `json:"id"`
	Status           BroadcastStatus `json:"status"`
	RecipientsTotal  int             `json:"recipientsTotal"`
	RecipientsQueued int             `json:"recipientsQueued"`
}

// CancelBroadcastResponse é o broadcast cancelado.
type CancelBroadcastResponse struct {
	ID          string          `json:"id"`
	Status      BroadcastStatus `json:"status"`
	CancelledAt *time.Time      `json:"cancelledAt"`
}

// TestSendBroadcastParams envia o broadcast para um destinatário de teste.
type TestSendBroadcastParams struct {
	To        string         `json:"to"`
	Variables map[string]any `json:"variables,omitzero"`
}

// BroadcastsService gerencia broadcasts. Toda operação exige chave full_access;
// sem ela, o erro tem código [CodePermissionDenied].
type BroadcastsService struct {
	r *requester
}

// Create cria um broadcast em rascunho.
func (s *BroadcastsService) Create(ctx context.Context, params *CreateBroadcastParams) (*Broadcast, error) {
	op := post(routes.Broadcasts, derefOrZero(params)).requiringFullAccess(opBroadcastsCreate)
	return call[Broadcast](ctx, s.r, op)
}

// List devolve uma página de broadcasts. params pode ser nil.
func (s *BroadcastsService) List(ctx context.Context, params *ListBroadcastsParams) (*ListBroadcastsResponse, error) {
	op := get(routes.Broadcasts).withQuery(params.values()).requiringFullAccess(opBroadcastsList)
	return call[ListBroadcastsResponse](ctx, s.r, op)
}

// ListAll percorre todas as páginas a partir de params.After, sob demanda.
func (s *BroadcastsService) ListAll(ctx context.Context, params *ListBroadcastsParams) iter.Seq2[Broadcast, error] {
	base := derefOrZero(params)
	return paginate(ctx, base.After, func(ctx context.Context, cursor string) ([]Broadcast, *string, error) {
		page := base
		page.After = cursor
		res, err := s.List(ctx, &page)
		if err != nil {
			return nil, nil, err
		}
		return res.Broadcasts, res.NextCursor, nil
	})
}

// Get devolve um broadcast.
func (s *BroadcastsService) Get(ctx context.Context, id string) (*Broadcast, error) {
	op := get(routes.Expand(routes.Broadcast, id)).requiringFullAccess(opBroadcastsGet)
	return call[Broadcast](ctx, s.r, op)
}

// Send dispara o broadcast para a audiência.
func (s *BroadcastsService) Send(ctx context.Context, id string) (*SendBroadcastResponse, error) {
	op := post(routes.Expand(routes.BroadcastSend, id), nil).requiringFullAccess(opBroadcastsSend)
	return call[SendBroadcastResponse](ctx, s.r, op)
}

// TestSend envia o broadcast para um destinatário de teste.
func (s *BroadcastsService) TestSend(ctx context.Context, id string, params *TestSendBroadcastParams) (*TestSendResponse, error) {
	op := post(routes.Expand(routes.BroadcastTestSend, id), derefOrZero(params)).requiringFullAccess(opBroadcastsTestSend)
	return call[TestSendResponse](ctx, s.r, op)
}

// Cancel cancela um broadcast agendado ou em envio.
func (s *BroadcastsService) Cancel(ctx context.Context, id string) (*CancelBroadcastResponse, error) {
	op := post(routes.Expand(routes.BroadcastCancel, id), nil).requiringFullAccess(opBroadcastsCancel)
	return call[CancelBroadcastResponse](ctx, s.r, op)
}

func (p *ListBroadcastsParams) values() map[string][]string {
	v := derefOrZero(p)
	return newQuery().
		str("status", string(v.Status)).
		time("from", v.From).
		time("to", v.To).
		str("after", v.After).
		num("limit", v.Limit).
		values()
}
