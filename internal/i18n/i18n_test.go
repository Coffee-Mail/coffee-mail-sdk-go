package i18n

import "testing"

func TestEveryKeyExistsInEveryLocale(t *testing.T) {
	keys := []Key{MissingAPIKey, RequestTimeout, NetworkError, UnexpectedError, InvalidJSON, RateLimitExceeded, PermissionDenied}
	for locale, messages := range catalog {
		for _, key := range keys {
			if messages[key] == "" {
				t.Errorf("%s sem a chave %s", locale, key)
			}
		}
		if len(messages) != len(keys) {
			t.Errorf("%s tem chaves extras", locale)
		}
	}
}

func TestFormatReplacesPlaceholdersAndFallsBack(t *testing.T) {
	got := Format(En, RequestTimeout, map[string]string{ParamTimeoutMs: "1500"})
	if got != "The request timed out after 1500ms." {
		t.Fatalf("Format = %q", got)
	}
	if Format("fr", InvalidJSON, nil) != ptBR[InvalidJSON] {
		t.Fatal("locale desconhecido deveria cair em pt-BR")
	}
	if Supported("fr") || !Supported(PtBR) {
		t.Fatal("Supported incorreto")
	}
}
