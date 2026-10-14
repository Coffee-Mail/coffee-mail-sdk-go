package coffeemail

import (
	"context"
	"iter"
	"time"

	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/routes"
)

// EmailStatus é o estado de um email. A lista é aberta: a API pode adicionar valores.
type EmailStatus string

// Estados de email.
const (
	EmailStatusQueued     EmailStatus = "queued"
	EmailStatusProcessing EmailStatus = "processing"
	EmailStatusSent       EmailStatus = "sent"
	EmailStatusDelivered  EmailStatus = "delivered"
	EmailStatusBounced    EmailStatus = "bounced"
	EmailStatusComplained EmailStatus = "complained"
	EmailStatusFailed     EmailStatus = "failed"
	EmailStatusSkipped    EmailStatus = "skipped"
	EmailStatusScheduled  EmailStatus = "scheduled"
	EmailStatusCancelled  EmailStatus = "cancelled"
)

// EmailType classifica o envio para fins de supressão e métricas.
type EmailType string

// Tipos de email.
const (
	EmailTypeTransactional EmailType = "transactional"
	EmailTypeBroadcast     EmailType = "broadcast"
)

// AttachmentDisposition indica se o anexo é exibido como arquivo ou embutido no HTML.
type AttachmentDisposition string

// Disposições de anexo.
const (
	AttachmentDispositionAttachment AttachmentDisposition = "attachment"
	AttachmentDispositionInline     AttachmentDisposition = "inline"
)

// EmailEventType é o tipo de um evento do ciclo de vida do email.
type EmailEventType string

// Tipos de evento de email.
const (
	EmailEventQueued     EmailEventType = "queued"
	EmailEventProcessing EmailEventType = "processing"
	EmailEventSent       EmailEventType = "sent"
	EmailEventDelivered  EmailEventType = "delivered"
	EmailEventOpened     EmailEventType = "opened"
	EmailEventClicked    EmailEventType = "clicked"
	EmailEventBounced    EmailEventType = "bounced"
	EmailEventComplained EmailEventType = "complained"
	EmailEventFailed     EmailEventType = "failed"
)

// Address é um endereço de email com nome opcional.
type Address struct {
	Email string `json:"email"`
	Name  string `json:"name,omitzero"`
}

// Attachment é um anexo. Content é o arquivo em base64; CID é obrigatório quando
// Disposition é inline.
type Attachment struct {
	Filename    string                `json:"filename"`
	Content     string                `json:"content"`
	ContentType string                `json:"contentType"`
	Disposition AttachmentDisposition `json:"disposition"`
	CID         string                `json:"cid,omitzero"`
}

// Tag é um par nome/valor usado para filtrar emails e métricas.
type Tag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// SendEmailParams são os dados de um envio. Informe HTML/Text com Subject, ou
// TemplateID com Variables.
//
// Sem IdempotencyKey, a API trata envios com payload idêntico como o mesmo envio
// e devolve o id original.
type SendEmailParams struct {
	From                Address           `json:"from"`
	To                  []Address         `json:"to"`
	CC                  []Address         `json:"cc,omitzero"`
	BCC                 []Address         `json:"bcc,omitzero"`
	ReplyTo             *Address          `json:"replyTo,omitzero"`
	Subject             string            `json:"subject,omitzero"`
	HTML                string            `json:"html,omitzero"`
	Text                string            `json:"text,omitzero"`
	Headers             map[string]string `json:"headers,omitzero"`
	Attachments         []Attachment      `json:"attachments,omitzero"`
	TemplateID          string            `json:"templateId,omitzero"`
	Variables           map[string]any    `json:"variables,omitzero"`
	EmailType           EmailType         `json:"emailType,omitzero"`
	SuppressionCategory string            `json:"suppressionCategory,omitzero"`
	Tags                []Tag             `json:"tags,omitzero"`
	ScheduledAt         time.Time         `json:"scheduledAt,omitzero"`
	Metadata            map[string]any    `json:"metadata,omitzero"`

	// IdempotencyKey é enviada no header X-Idempotency-Key, não no corpo.
	IdempotencyKey string `json:"-"`
	// Sandbox envia em modo de teste (header X-Coffeemail-Sandbox).
	Sandbox bool `json:"-"`
}

// SendBatchParams são os dados de um envio em lote (até 100 emails).
type SendBatchParams struct {
	Emails []SendEmailParams
	// IdempotencyKey vale para o lote inteiro.
	IdempotencyKey string
	// Sandbox envia o lote inteiro em modo de teste.
	Sandbox bool
}

// SendEmailResponse é o resultado de um envio aceito pela API.
type SendEmailResponse struct {
	ID       string      `json:"id"`
	Status   EmailStatus `json:"status"`
	QueuedAt time.Time   `json:"queuedAt"`
}

// BatchSendResult é o resultado de um item do lote: Data quando OK, Error caso contrário.
type BatchSendResult struct {
	OK    bool               `json:"ok"`
	Data  *SendEmailResponse `json:"data"`
	Error string             `json:"error"`
}

// APIKeyRef identifica a chave de API que originou o email.
type APIKeyRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Email é o detalhe de um email enviado.
type Email struct {
	ID             string      `json:"id"`
	From           string      `json:"from"`
	To             []string    `json:"to"`
	CC             []string    `json:"cc"`
	BCC            []string    `json:"bcc"`
	Subject        *string     `json:"subject"`
	Status         EmailStatus `json:"status"`
	Attempts       int         `json:"attempts"`
	LastError      *string     `json:"lastError"`
	MessageID      *string     `json:"messageId"`
	APIKey         *APIKeyRef  `json:"apiKey"`
	CreatedAt      time.Time   `json:"createdAt"`
	ScheduledAt    *time.Time  `json:"scheduledAt"`
	SentAt         *time.Time  `json:"sentAt"`
	DeliveredAt    *time.Time  `json:"deliveredAt"`
	BouncedAt      *time.Time  `json:"bouncedAt"`
	FailedAt       *time.Time  `json:"failedAt"`
	ComplainedAt   *time.Time  `json:"complainedAt"`
	SuppressedAt   *time.Time  `json:"suppressedAt"`
	OpenedAt       *time.Time  `json:"openedAt"`
	FirstClickedAt *time.Time  `json:"firstClickedAt"`
	OpenCount      int         `json:"openCount"`
	ClickCount     int         `json:"clickCount"`
	ProxyOpenCount int         `json:"proxyOpenCount"`
	Tags           []Tag       `json:"tags"`
	HTML           *string     `json:"html"`
	Text           *string     `json:"text"`
}

// ListEmailsParams filtra a listagem de emails. Todos os campos são opcionais.
type ListEmailsParams struct {
	Status          EmailStatus
	Recipient       string
	FromEmail       string
	SubjectContains string
	APIKeyID        string
	Tag             string
	From            time.Time
	To              time.Time
	// After é o cursor devolvido em NextCursor da página anterior.
	After string
	// Limit é o tamanho da página (máximo 100, padrão 50).
	Limit int
}

// ListEmailsResponse é uma página de emails.
type ListEmailsResponse struct {
	Emails     []Email `json:"emails"`
	NextCursor *string `json:"nextCursor"`
}

// EmailEvent é um evento do ciclo de vida do email.
type EmailEvent struct {
	Type      EmailEventType `json:"type"`
	Timestamp time.Time      `json:"timestamp"`
	Metadata  map[string]any `json:"metadata"`
}

// EmailEventsResponse é a linha do tempo de um email.
type EmailEventsResponse struct {
	ID     string       `json:"id"`
	Events []EmailEvent `json:"events"`
}

// EmailTag é uma tag usada nos envios com todos os seus valores.
type EmailTag struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// ListEmailTagsResponse lista as tags usadas nos envios.
type ListEmailTagsResponse struct {
	Tags []EmailTag `json:"tags"`
}

// EmailsService envia e consulta emails.
type EmailsService struct {
	r *requester
}

// Send envia um email. A API enfileira o envio e responde imediatamente.
func (s *EmailsService) Send(ctx context.Context, params *SendEmailParams) (*SendEmailResponse, error) {
	p := derefOrZero(params)
	op := post(routes.Emails, p).withSendHeaders(p.IdempotencyKey, p.Sandbox)
	return call[SendEmailResponse](ctx, s.r, op)
}

// SendBatch envia até 100 emails numa só requisição. Cada item tem seu próprio
// resultado; um item recusado não invalida os demais.
func (s *EmailsService) SendBatch(ctx context.Context, params *SendBatchParams) ([]BatchSendResult, error) {
	p := derefOrZero(params)
	op := post(routes.EmailsBatch, p.Emails).withSendHeaders(p.IdempotencyKey, p.Sandbox)
	results, err := call[[]BatchSendResult](ctx, s.r, op)
	if err != nil {
		return nil, err
	}
	return *results, nil
}

// Get devolve o detalhe de um email.
func (s *EmailsService) Get(ctx context.Context, id string) (*Email, error) {
	return call[Email](ctx, s.r, get(routes.Expand(routes.Email, id)))
}

// List devolve uma página de emails. params pode ser nil.
func (s *EmailsService) List(ctx context.Context, params *ListEmailsParams) (*ListEmailsResponse, error) {
	return call[ListEmailsResponse](ctx, s.r, get(routes.Emails).withQuery(params.values()))
}

// ListAll percorre todas as páginas a partir de params.After, buscando a próxima
// página sob demanda. Interrompa o laço para parar a paginação.
func (s *EmailsService) ListAll(ctx context.Context, params *ListEmailsParams) iter.Seq2[Email, error] {
	base := derefOrZero(params)
	return paginate(ctx, base.After, func(ctx context.Context, cursor string) ([]Email, *string, error) {
		page := base
		page.After = cursor
		res, err := s.List(ctx, &page)
		if err != nil {
			return nil, nil, err
		}
		return res.Emails, res.NextCursor, nil
	})
}

// GetEvents devolve a linha do tempo de eventos de um email.
func (s *EmailsService) GetEvents(ctx context.Context, id string) (*EmailEventsResponse, error) {
	return call[EmailEventsResponse](ctx, s.r, get(routes.Expand(routes.EmailEvents, id)))
}

// GetTags lista as tags usadas nos envios e seus valores.
func (s *EmailsService) GetTags(ctx context.Context) (*ListEmailTagsResponse, error) {
	return call[ListEmailTagsResponse](ctx, s.r, get(routes.EmailsTags))
}

// Cancel cancela um email agendado.
func (s *EmailsService) Cancel(ctx context.Context, id string) error {
	return s.r.do(ctx, post(routes.Expand(routes.EmailCancel, id), nil), nil)
}

// Resend reenvia um email existente.
func (s *EmailsService) Resend(ctx context.Context, id string) (*SendEmailResponse, error) {
	return call[SendEmailResponse](ctx, s.r, post(routes.Expand(routes.EmailResend, id), nil))
}

func (p *ListEmailsParams) values() map[string][]string {
	v := derefOrZero(p)
	return newQuery().
		str("status", string(v.Status)).
		str("recipient", v.Recipient).
		str("fromEmail", v.FromEmail).
		str("subjectContains", v.SubjectContains).
		str("apiKeyId", v.APIKeyID).
		str("tag", v.Tag).
		time("from", v.From).
		time("to", v.To).
		str("after", v.After).
		num("limit", v.Limit).
		values()
}

func derefOrZero[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}
