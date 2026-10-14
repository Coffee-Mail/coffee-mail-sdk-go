package i18n

import "strings"

type Key string

const (
	MissingAPIKey     Key = "MISSING_API_KEY"
	RequestTimeout    Key = "REQUEST_TIMEOUT"
	NetworkError      Key = "NETWORK_ERROR"
	UnexpectedError   Key = "UNEXPECTED_ERROR"
	InvalidJSON       Key = "INVALID_JSON"
	RateLimitExceeded Key = "RATE_LIMIT_EXCEEDED"
	PermissionDenied  Key = "PERMISSION_DENIED"

	PtBR            = "pt-BR"
	En              = "en"
	DefaultLocale   = PtBR
	ParamTimeoutMs  = "timeoutMs"
	ParamDetails    = "details"
	ParamPermission = "requiredPermission"
)

var ptBR = map[Key]string{
	MissingAPIKey:     "A chave de API não foi informada. Forneça uma chave válida (ex: cm_live_...) no NewClient ou na variável de ambiente COFFEEMAIL_API_KEY.",
	RequestTimeout:    "A requisição excedeu o tempo limite configurado de {timeoutMs}ms.",
	NetworkError:      "Falha de rede ao tentar se comunicar com os servidores do CoffeeMail: {details}",
	UnexpectedError:   "Ocorreu um erro inesperado durante a comunicação com a API do CoffeeMail.",
	InvalidJSON:       "A resposta do servidor não pôde ser interpretada como JSON válido.",
	RateLimitExceeded: "Limite de requisições excedido (Rate Limit). Tente novamente em alguns segundos.",
	PermissionDenied:  "Esta chave de API não possui a permissão necessária ({requiredPermission}) para executar esta operação. Crie uma chave com permissão 'full_access' no dashboard.",
}

var en = map[Key]string{
	MissingAPIKey:     "API key was not provided. Pass a valid API key (e.g. cm_live_...) to NewClient or set the COFFEEMAIL_API_KEY environment variable.",
	RequestTimeout:    "The request timed out after {timeoutMs}ms.",
	NetworkError:      "Network error while attempting to communicate with CoffeeMail API: {details}",
	UnexpectedError:   "An unexpected error occurred while communicating with CoffeeMail API.",
	InvalidJSON:       "The server response could not be parsed as valid JSON.",
	RateLimitExceeded: "Rate limit exceeded. Please retry in a few moments.",
	PermissionDenied:  "This API key lacks the required permission ({requiredPermission}) for this operation. Create a key with 'full_access' permission in the dashboard.",
}

var catalog = map[string]map[Key]string{PtBR: ptBR, En: en}

func Supported(locale string) bool {
	_, ok := catalog[locale]
	return ok
}

func Format(locale string, key Key, params map[string]string) string {
	messages, ok := catalog[locale]
	if !ok {
		messages = catalog[DefaultLocale]
	}
	pairs := make([]string, 0, len(params)*2)
	for name, value := range params {
		pairs = append(pairs, "{"+name+"}", value)
	}
	return strings.NewReplacer(pairs...).Replace(messages[key])
}
