package coffeemail

import (
	"fmt"
	"time"
)

// Códigos canônicos expostos em [Error.Code]. São iguais em todos os SDKs
// oficiais da CoffeeMail.
const (
	CodeValidationError     = "VALIDATION_ERROR"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodePaymentRequired     = "PAYMENT_REQUIRED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeConflict            = "CONFLICT"
	CodeRateLimitExceeded   = "RATE_LIMIT_EXCEEDED"
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
	CodeNetworkError        = "NETWORK_ERROR"
	CodePermissionDenied    = "PERMISSION_DENIED"
	CodeAPIError            = "API_ERROR"
	CodeInvalidResponse     = "INVALID_RESPONSE"
)

const missingAPIKeyCode = "MISSING_API_KEY"

type sentinel string

func (s sentinel) Error() string {
	return "coffeemail: " + string(s)
}

// Sentinelas para uso com [errors.Is]. Um [*Error] corresponde à sentinela de
// mesmo código canônico.
var (
	ErrValidation       error = sentinel(CodeValidationError)
	ErrUnauthorized     error = sentinel(CodeUnauthorized)
	ErrPaymentRequired  error = sentinel(CodePaymentRequired)
	ErrForbidden        error = sentinel(CodeForbidden)
	ErrNotFound         error = sentinel(CodeNotFound)
	ErrConflict         error = sentinel(CodeConflict)
	ErrRateLimited      error = sentinel(CodeRateLimitExceeded)
	ErrServer           error = sentinel(CodeInternalServerError)
	ErrNetwork          error = sentinel(CodeNetworkError)
	ErrPermissionDenied error = sentinel(CodePermissionDenied)
	ErrInvalidResponse  error = sentinel(CodeInvalidResponse)

	// ErrMissingAPIKey é a causa do erro devolvido por [NewClient] quando nenhuma
	// chave é informada nem encontrada em COFFEEMAIL_API_KEY.
	ErrMissingAPIKey error = sentinel(missingAPIKeyCode)
)

// Error é o erro devolvido por toda operação do SDK.
type Error struct {
	// Status é o status HTTP da resposta, ou 0 em falha de rede e de configuração.
	Status int
	// Code é o código canônico (constantes Code*).
	Code string
	// APICode é o código original enviado pela API, como PLAN_QUOTA_EXCEEDED.
	APICode string
	// Message é a mensagem da API ou, na falta dela, a mensagem local no idioma
	// configurado.
	Message string
	// Details reúne os campos meta e details do erro da API.
	Details map[string]any
	// RequestID é o identificador da requisição informado pela API, quando houver.
	RequestID string
	// RetryAfter é a espera sugerida pela API no header Retry-After, quando houver.
	RetryAfter time.Duration

	cause error
}

// Error implementa a interface error.
func (e *Error) Error() string {
	return fmt.Sprintf("coffeemail: %s (status %d, code %s)", e.Message, e.Status, e.Code)
}

// Unwrap devolve a causa original, como um erro de rede ou [context.DeadlineExceeded].
func (e *Error) Unwrap() error {
	return e.cause
}

// Is permite comparar com as sentinelas via [errors.Is].
func (e *Error) Is(target error) bool {
	code, ok := target.(sentinel)
	return ok && string(code) == e.Code
}
