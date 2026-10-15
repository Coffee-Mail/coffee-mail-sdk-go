package coffeemail

import (
	"context"
	"time"

	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/routes"
)

// Contact é um contato de uma audiência.
type Contact struct {
	ID        string         `json:"id"`
	Email     string         `json:"email"`
	FirstName *string        `json:"firstName"`
	LastName  *string        `json:"lastName"`
	Metadata  map[string]any `json:"metadata"`
	// Subscribed é false quando o contato se descadastrou dos broadcasts; nil
	// equivale a inscrito.
	Subscribed *bool     `json:"subscribed"`
	CreatedAt  time.Time `json:"createdAt"`
}

// IsSubscribed indica se o contato ainda recebe broadcasts.
func (c Contact) IsSubscribed() bool {
	return c.Subscribed == nil || *c.Subscribed
}

// CreateContactParams são os dados de um contato novo.
type CreateContactParams struct {
	Email     string         `json:"email"`
	FirstName string         `json:"firstName,omitzero"`
	LastName  string         `json:"lastName,omitzero"`
	Metadata  map[string]any `json:"metadata,omitzero"`
}

// UpdateContactParams altera um contato. Campos nil não são alterados.
type UpdateContactParams struct {
	FirstName *string        `json:"firstName,omitzero"`
	LastName  *string        `json:"lastName,omitzero"`
	Metadata  map[string]any `json:"metadata,omitzero"`
}

// UpdateContactResponse identifica o contato alterado.
type UpdateContactResponse struct {
	ID string `json:"id"`
}

// ListContactsResponse é uma página de contatos com o total da audiência.
type ListContactsResponse struct {
	Contacts []Contact `json:"contacts"`
	Total    int       `json:"total"`
}

// BulkContactError descreve um contato recusado na importação em lote.
type BulkContactError struct {
	Email  string `json:"email"`
	Reason string `json:"reason"`
}

// BulkAddContactsResponse resume uma importação em lote.
type BulkAddContactsResponse struct {
	Inserted int                `json:"inserted"`
	Skipped  int                `json:"skipped"`
	Errors   []BulkContactError `json:"errors"`
}

type bulkContactsBody struct {
	Contacts []CreateContactParams `json:"contacts"`
}

// ContactsService gerencia os contatos das audiências.
type ContactsService struct {
	r *requester
}

// Create adiciona um contato à audiência.
func (s *ContactsService) Create(ctx context.Context, audienceID string, params *CreateContactParams) (*Contact, error) {
	return call[Contact](ctx, s.r, post(routes.Expand(routes.AudienceContacts, audienceID), derefOrZero(params)))
}

// List devolve uma página de contatos da audiência. params pode ser nil.
func (s *ContactsService) List(ctx context.Context, audienceID string, params *OffsetParams) (*ListContactsResponse, error) {
	op := get(routes.Expand(routes.AudienceContacts, audienceID)).withQuery(params.values())
	return call[ListContactsResponse](ctx, s.r, op)
}

// BulkAdd importa até 10 000 contatos na audiência.
func (s *ContactsService) BulkAdd(ctx context.Context, audienceID string, contacts []CreateContactParams) (*BulkAddContactsResponse, error) {
	op := post(routes.Expand(routes.AudienceContactsBulk, audienceID), bulkContactsBody{Contacts: contacts})
	return call[BulkAddContactsResponse](ctx, s.r, op)
}

// Update altera um contato.
func (s *ContactsService) Update(ctx context.Context, audienceID, contactID string, params *UpdateContactParams) (*UpdateContactResponse, error) {
	op := put(routes.ExpandContact(routes.AudienceContact, audienceID, contactID), derefOrZero(params))
	return call[UpdateContactResponse](ctx, s.r, op)
}

// Delete remove um contato da audiência.
func (s *ContactsService) Delete(ctx context.Context, audienceID, contactID string) error {
	return s.r.do(ctx, del(routes.ExpandContact(routes.AudienceContact, audienceID, contactID)), nil)
}
