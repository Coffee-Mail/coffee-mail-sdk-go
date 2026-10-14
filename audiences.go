package coffeemail

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/routes"
)

// Audience é uma lista de contatos usada em broadcasts.
type Audience struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Active      bool    `json:"active"`
	// TotalContacts é o número de contatos da audiência.
	TotalContacts int       `json:"totalContacts"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// UnmarshalJSON aceita a contagem tanto em totalContacts quanto em contactsCount.
func (a *Audience) UnmarshalJSON(data []byte) error {
	type plain Audience
	var wire struct {
		plain
		ContactsCount int `json:"contactsCount"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	*a = Audience(wire.plain)
	a.TotalContacts = max(a.TotalContacts, wire.ContactsCount)
	return nil
}

// CreateAudienceParams são os dados de uma audiência nova.
type CreateAudienceParams struct {
	Name        string `json:"name"`
	Description string `json:"description,omitzero"`
}

// UpdateAudienceParams altera uma audiência. Campos nil não são alterados.
type UpdateAudienceParams struct {
	Name        *string `json:"name,omitzero"`
	Description *string `json:"description,omitzero"`
	Active      *bool   `json:"active,omitzero"`
}

// UpdateAudienceResponse identifica a audiência alterada.
type UpdateAudienceResponse struct {
	ID string `json:"id"`
}

// OffsetParams pagina listagens por deslocamento. Zero usa o padrão da API.
type OffsetParams struct {
	Limit  int
	Offset int
}

// ListAudiencesResponse é uma página de audiências.
type ListAudiencesResponse struct {
	Audiences []Audience `json:"audiences"`
	Total     int        `json:"total"`
}

// AudiencesService gerencia audiências. Os contatos ficam em Contacts.
type AudiencesService struct {
	// Contacts gerencia os contatos das audiências.
	Contacts *ContactsService

	r *requester
}

// Create cria uma audiência.
func (s *AudiencesService) Create(ctx context.Context, params *CreateAudienceParams) (*Audience, error) {
	return call[Audience](ctx, s.r, post(routes.Audiences, derefOrZero(params)))
}

// List devolve uma página de audiências. params pode ser nil.
func (s *AudiencesService) List(ctx context.Context, params *OffsetParams) (*ListAudiencesResponse, error) {
	return call[ListAudiencesResponse](ctx, s.r, get(routes.Audiences).withQuery(params.values()))
}

// Get devolve uma audiência.
func (s *AudiencesService) Get(ctx context.Context, id string) (*Audience, error) {
	return call[Audience](ctx, s.r, get(routes.Expand(routes.Audience, id)))
}

// Update altera uma audiência.
func (s *AudiencesService) Update(ctx context.Context, id string, params *UpdateAudienceParams) (*UpdateAudienceResponse, error) {
	return call[UpdateAudienceResponse](ctx, s.r, put(routes.Expand(routes.Audience, id), derefOrZero(params)))
}

// Delete exclui uma audiência.
func (s *AudiencesService) Delete(ctx context.Context, id string) error {
	return s.r.do(ctx, del(routes.Expand(routes.Audience, id)), nil)
}

func (p *OffsetParams) values() map[string][]string {
	v := derefOrZero(p)
	return newQuery().num("limit", v.Limit).num("offset", v.Offset).values()
}
