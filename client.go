package coffeemail

import (
	"os"
	"strings"

	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/i18n"
)

// APIKeyEnvironmentVariable é a variável de ambiente lida quando [NewClient]
// recebe uma chave vazia.
const APIKeyEnvironmentVariable = "COFFEEMAIL_API_KEY"

// Client é o ponto de entrada do SDK. É imutável e seguro para uso concorrente;
// crie um por aplicação e reutilize.
type Client struct {
	// Emails envia e consulta emails.
	Emails *EmailsService
	// Domains gerencia domínios de envio.
	Domains *DomainsService
	// Templates gerencia templates de email.
	Templates *TemplatesService
	// Audiences gerencia audiências e, em Audiences.Contacts, seus contatos.
	Audiences *AudiencesService
	// Broadcasts gerencia envios em massa para audiências.
	Broadcasts *BroadcastsService
	// Senders gerencia remetentes verificados.
	Senders *SendersService
	// Suppressions gerencia a lista de supressão.
	Suppressions *SuppressionsService
	// Webhooks gerencia endpoints de webhook e suas entregas.
	Webhooks *WebhooksService
	// Stats consulta métricas de envio.
	Stats *StatsService

	introspection *introspectionCache
}

// NewClient cria um [Client]. Se apiKey for vazia, usa a variável de ambiente
// COFFEEMAIL_API_KEY. Sem chave, devolve um [*Error] com código
// [CodeUnauthorized] cuja causa é [ErrMissingAPIKey].
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	key := resolveAPIKey(apiKey)
	if key == "" {
		return nil, &Error{
			Code:    CodeUnauthorized,
			Message: i18n.Format(string(cfg.locale), i18n.MissingAPIKey, nil),
			cause:   ErrMissingAPIKey,
		}
	}
	r := newRequester(key, cfg)
	return &Client{
		Emails:        &EmailsService{r: r},
		Domains:       &DomainsService{r: r},
		Templates:     &TemplatesService{r: r},
		Audiences:     &AudiencesService{r: r, Contacts: &ContactsService{r: r}},
		Broadcasts:    &BroadcastsService{r: r},
		Senders:       &SendersService{r: r},
		Suppressions:  &SuppressionsService{r: r},
		Webhooks:      &WebhooksService{r: r},
		Stats:         &StatsService{r: r},
		introspection: newIntrospectionCache(r, cfg.now),
	}, nil
}

func resolveAPIKey(apiKey string) string {
	key := strings.TrimSpace(apiKey)
	if key != "" {
		return key
	}
	return strings.TrimSpace(os.Getenv(APIKeyEnvironmentVariable))
}
