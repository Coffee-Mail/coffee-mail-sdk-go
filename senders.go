package coffeemail

import (
	"context"
	"time"

	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/routes"
)

// CreateSenderParams são os dados de um remetente novo.
type CreateSenderParams struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName,omitzero"`
}

// Sender é um remetente.
type Sender struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	DisplayName *string    `json:"displayName"`
	VerifiedAt  *time.Time `json:"verifiedAt"`
	Active      bool       `json:"active"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// CreatedSender é o remetente criado. O token de verificação segue por email ao
// endereço cadastrado.
type CreatedSender struct {
	Sender
	// Deprecated: a API não devolve mais o token; o campo vem sempre vazio. A
	// verificação acontece pelo link enviado ao email do remetente.
	VerificationToken string `json:"verificationToken,omitzero"`
}

// ListSendersResponse lista os remetentes.
type ListSendersResponse struct {
	Senders []Sender `json:"senders"`
}

// VerifySenderResponse identifica o remetente verificado.
type VerifySenderResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type verifySenderBody struct {
	Token string `json:"token"`
}

// SendersService gerencia remetentes verificados.
type SendersService struct {
	r *requester
}

// Create cadastra um remetente e dispara o email de verificação.
func (s *SendersService) Create(ctx context.Context, params *CreateSenderParams) (*CreatedSender, error) {
	return call[CreatedSender](ctx, s.r, post(routes.Senders, derefOrZero(params)))
}

// List lista os remetentes.
func (s *SendersService) List(ctx context.Context) (*ListSendersResponse, error) {
	return call[ListSendersResponse](ctx, s.r, get(routes.Senders))
}

// Verify confirma um remetente com o token recebido por email.
func (s *SendersService) Verify(ctx context.Context, token string) (*VerifySenderResponse, error) {
	return call[VerifySenderResponse](ctx, s.r, post(routes.SendersVerify, verifySenderBody{Token: token}))
}

// ResendVerification reenvia o email de verificação com um token novo; o anterior
// expira.
func (s *SendersService) ResendVerification(ctx context.Context, id string) error {
	return s.r.do(ctx, post(routes.Expand(routes.SenderResend, id), nil), nil)
}

// Delete remove um remetente.
func (s *SendersService) Delete(ctx context.Context, id string) error {
	return s.r.do(ctx, del(routes.Expand(routes.Sender, id)), nil)
}
