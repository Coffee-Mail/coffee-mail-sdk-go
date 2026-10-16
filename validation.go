package coffeemail

import (
	"context"

	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/routes"
)

// EmailValidation é o resultado da validação de um endereço.
type EmailValidation struct {
	Email        string `json:"email"`
	FormatValid  bool   `json:"formatValid"`
	MXValid      bool   `json:"mxValid"`
	IsDisposable bool   `json:"isDisposable"`
	// Reason explica por que o endereço foi considerado inválido, quando for.
	Reason string `json:"reason,omitzero"`
}

// ValidationCredits é o saldo de créditos de validação do mês após a chamada.
type ValidationCredits struct {
	// Limit é o total de créditos do mês, ou nil quando ilimitado.
	Limit *int `json:"limit"`
	Used  int  `json:"used"`
	// Remaining é o saldo do mês, ou nil quando ilimitado.
	Remaining *int `json:"remaining"`
}

// ValidateEmailsResponse traz a validação de cada endereço e o saldo de créditos.
type ValidateEmailsResponse struct {
	Items        []EmailValidation `json:"items"`
	ValidCount   int               `json:"validCount"`
	InvalidCount int               `json:"invalidCount"`
	Credits      ValidationCredits `json:"credits"`
}

type validateEmailsBody struct {
	Emails []string `json:"emails"`
}

// ValidationService valida endereços de email antes do envio.
type ValidationService struct {
	r *requester
}

// Validate valida de 1 a 100 endereços (formato, MX e domínio descartável) e
// consome um crédito por endereço. Créditos do mês esgotados geram
// [ErrRateLimited] com APICode VALIDATION_CREDITS_EXHAUSTED, que nunca é
// repetido automaticamente.
func (s *ValidationService) Validate(ctx context.Context, emails []string) (*ValidateEmailsResponse, error) {
	return call[ValidateEmailsResponse](ctx, s.r, post(routes.Validate, validateEmailsBody{Emails: emails}))
}
