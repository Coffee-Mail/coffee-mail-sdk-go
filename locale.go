package coffeemail

import "github.com/Coffee-Mail/coffee-mail-sdk-go/internal/i18n"

// Locale é o idioma das mensagens locais do SDK e do header Accept-Language.
type Locale string

const (
	// LocalePtBR é o português do Brasil, idioma padrão.
	LocalePtBR Locale = i18n.PtBR
	// LocaleEn é o inglês.
	LocaleEn Locale = i18n.En
)
