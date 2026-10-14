package coffeemail

import (
	"net/http"
	"time"

	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/i18n"
)

const defaultTimeout = 10 * time.Second

// Option configura o [Client] em [NewClient].
type Option func(*config)

type config struct {
	httpClient *http.Client
	timeout    time.Duration
	locale     Locale
	retry      RetryPolicy
	now        func() time.Time
}

func defaultConfig() config {
	return config{
		httpClient: &http.Client{CheckRedirect: refuseRedirect},
		timeout:    defaultTimeout,
		locale:     LocalePtBR,
		now:        time.Now,
	}
}

func refuseRedirect(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

// WithHTTPClient define o [http.Client] usado nas requisições. Use para proxy
// corporativo, instrumentação (OpenTelemetry) ou testes com um RoundTripper
// próprio. O endereço da API não é configurável.
func WithHTTPClient(client *http.Client) Option {
	return func(c *config) {
		if client != nil {
			c.httpClient = client
		}
	}
}

// WithTimeout define o tempo limite de cada tentativa. Padrão: 10s. Valores
// menores ou iguais a zero são ignorados. Um deadline no context continua valendo.
func WithTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.timeout = positiveOr(timeout, c.timeout)
	}
}

// WithLocale define o idioma das mensagens locais e do header Accept-Language.
// Padrão: [LocalePtBR]. Idiomas não suportados são ignorados.
func WithLocale(locale Locale) Option {
	return func(c *config) {
		if i18n.Supported(string(locale)) {
			c.locale = locale
		}
	}
}

// WithRetry ativa novas tentativas automáticas conforme a [RetryPolicy].
func WithRetry(policy RetryPolicy) Option {
	return func(c *config) {
		c.retry = policy
	}
}
