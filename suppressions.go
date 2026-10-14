package coffeemail

import (
	"context"
	"iter"
	"time"

	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/routes"
)

// SuppressionReason é o motivo de uma supressão.
type SuppressionReason string

// Motivos de supressão. Na criação manual, só Manual, Bounce e Complaint são aceitos.
const (
	SuppressionReasonManual          SuppressionReason = "manual"
	SuppressionReasonBounce          SuppressionReason = "bounce"
	SuppressionReasonISPBlock        SuppressionReason = "isp_block"
	SuppressionReasonMailboxNotFound SuppressionReason = "mailbox_not_found"
	SuppressionReasonComplaint       SuppressionReason = "complaint"
	SuppressionReasonUnsubscribe     SuppressionReason = "unsubscribe"
)

// SuppressionSource indica se a supressão foi manual ou automática.
type SuppressionSource string

// Origens de supressão.
const (
	SuppressionSourceManual SuppressionSource = "manual"
	SuppressionSourceAuto   SuppressionSource = "auto"
)

// SuppressionStatus é o estado de uma supressão.
type SuppressionStatus string

// Estados de supressão.
const (
	SuppressionStatusActive   SuppressionStatus = "active"
	SuppressionStatusExpired  SuppressionStatus = "expired"
	SuppressionStatusInactive SuppressionStatus = "inactive"
)

// SuppressionCategory agrupa supressões para filtro.
type SuppressionCategory string

// Categorias de supressão.
const (
	SuppressionCategoryManual       SuppressionCategory = "manual"
	SuppressionCategoryBounce       SuppressionCategory = "bounce"
	SuppressionCategoryBlock        SuppressionCategory = "block"
	SuppressionCategoryComplaint    SuppressionCategory = "complaint"
	SuppressionCategoryUnsubscribe  SuppressionCategory = "unsubscribe"
	SuppressionCategoryInvalidEmail SuppressionCategory = "invalid_email"
	SuppressionCategoryGlobal       SuppressionCategory = "global"
)

// Suppression é um endereço suprimido.
type Suppression struct {
	ID        string            `json:"id"`
	Email     string            `json:"email"`
	Reason    SuppressionReason `json:"reason"`
	Source    SuppressionSource `json:"source"`
	Category  string            `json:"category"`
	Status    SuppressionStatus `json:"status"`
	ExpiresAt *time.Time        `json:"expiresAt"`
	CreatedAt time.Time         `json:"createdAt"`
}

// CreateSuppressionParams suprime um endereço manualmente.
type CreateSuppressionParams struct {
	Email     string            `json:"email"`
	Reason    SuppressionReason `json:"reason,omitzero"`
	ExpiresAt time.Time         `json:"expiresAt,omitzero"`
}

// BulkSuppressionEntry é um item da criação em lote.
type BulkSuppressionEntry struct {
	Email  string            `json:"email"`
	Reason SuppressionReason `json:"reason,omitzero"`
}

// BulkCreateSuppressionsResponse informa quantas supressões foram criadas.
type BulkCreateSuppressionsResponse struct {
	InsertedCount int `json:"insertedCount"`
}

// ListSuppressionsParams filtra a listagem de supressões. Todos os campos são opcionais.
type ListSuppressionsParams struct {
	Status   SuppressionStatus
	Reason   SuppressionReason
	Source   SuppressionSource
	Category SuppressionCategory
	// After é o cursor devolvido em NextCursor da página anterior.
	After string
	// Limit é o tamanho da página (máximo 200).
	Limit int
}

// ListSuppressionsResponse é uma página de supressões.
type ListSuppressionsResponse struct {
	Suppressions []Suppression `json:"suppressions"`
	NextCursor   *string       `json:"nextCursor"`
}

type bulkSuppressionsBody struct {
	Suppressions []BulkSuppressionEntry `json:"suppressions"`
}

// SuppressionsService gerencia a lista de supressão.
type SuppressionsService struct {
	r *requester
}

// Create suprime um endereço.
func (s *SuppressionsService) Create(ctx context.Context, params *CreateSuppressionParams) (*Suppression, error) {
	return call[Suppression](ctx, s.r, post(routes.Suppressions, derefOrZero(params)))
}

// List devolve uma página de supressões. params pode ser nil.
func (s *SuppressionsService) List(ctx context.Context, params *ListSuppressionsParams) (*ListSuppressionsResponse, error) {
	return call[ListSuppressionsResponse](ctx, s.r, get(routes.Suppressions).withQuery(params.values()))
}

// ListAll percorre todas as páginas a partir de params.After, sob demanda.
func (s *SuppressionsService) ListAll(ctx context.Context, params *ListSuppressionsParams) iter.Seq2[Suppression, error] {
	base := derefOrZero(params)
	return paginate(ctx, base.After, func(ctx context.Context, cursor string) ([]Suppression, *string, error) {
		page := base
		page.After = cursor
		res, err := s.List(ctx, &page)
		if err != nil {
			return nil, nil, err
		}
		return res.Suppressions, res.NextCursor, nil
	})
}

// Get devolve uma supressão pelo id.
func (s *SuppressionsService) Get(ctx context.Context, id string) (*Suppression, error) {
	return call[Suppression](ctx, s.r, get(routes.Expand(routes.Suppression, id)))
}

// Delete remove uma supressão pelo id.
func (s *SuppressionsService) Delete(ctx context.Context, id string) error {
	return s.r.do(ctx, del(routes.Expand(routes.Suppression, id)), nil)
}

// Reactivate reativa o envio para o endereço de uma supressão.
func (s *SuppressionsService) Reactivate(ctx context.Context, id string) error {
	return s.r.do(ctx, post(routes.Expand(routes.SuppressionReactivate, id), nil), nil)
}

// BulkCreate suprime até 1000 endereços numa requisição.
func (s *SuppressionsService) BulkCreate(ctx context.Context, entries []BulkSuppressionEntry) (*BulkCreateSuppressionsResponse, error) {
	return call[BulkCreateSuppressionsResponse](ctx, s.r, post(routes.SuppressionsBulk, bulkSuppressionsBody{Suppressions: entries}))
}

func (p *ListSuppressionsParams) values() map[string][]string {
	v := derefOrZero(p)
	return newQuery().
		str("status", string(v.Status)).
		str("reason", string(v.Reason)).
		str("source", string(v.Source)).
		str("category", string(v.Category)).
		str("after", v.After).
		num("limit", v.Limit).
		values()
}
