package coffeemail

import "time"

const (
	defaultRetryBaseDelay = 500 * time.Millisecond
	defaultRetryMaxDelay  = 30 * time.Second
	singleAttempt         = 1
)

// RetryPolicy configura novas tentativas automáticas. O retry é desligado por
// padrão e só é ativado com [WithRetry] e MaxAttempts maior que 1.
//
// Só são repetidos erros de rede, respostas 5xx e 429 que não sejam cota do plano
// esgotada. Requisições GET e DELETE são sempre elegíveis; POST só quando aceita
// chave de idempotência (envio de email), e nesse caso o SDK gera a chave se ela
// não for informada, reutilizando-a em todas as tentativas.
type RetryPolicy struct {
	// MaxAttempts é o total de tentativas, incluindo a primeira.
	MaxAttempts int
	// BaseDelay é a espera base do backoff exponencial. Padrão: 500ms.
	BaseDelay time.Duration
	// MaxDelay é o teto de espera entre tentativas, inclusive quando a API
	// informa Retry-After maior. Padrão: 30s.
	MaxDelay time.Duration
}

func (p RetryPolicy) enabled() bool {
	return p.MaxAttempts > singleAttempt
}

func (p RetryPolicy) withDefaults() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: max(p.MaxAttempts, singleAttempt),
		BaseDelay:   positiveOr(p.BaseDelay, defaultRetryBaseDelay),
		MaxDelay:    positiveOr(p.MaxDelay, defaultRetryMaxDelay),
	}
}

func positiveOr(value, fallback time.Duration) time.Duration {
	if value > 0 {
		return value
	}
	return fallback
}
