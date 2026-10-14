# Changelog

Todas as mudanças relevantes deste projeto são documentadas aqui. O formato segue o [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/) e o projeto adota o [Versionamento Semântico](https://semver.org/lang/pt-BR/).

## [Não lançado]

## [0.1.0] - Não publicado

### Adicionado

- `coffeemail.NewClient` com opções `WithHTTPClient`, `WithTimeout`, `WithLocale` e `WithRetry`; chave lida de `COFFEEMAIL_API_KEY` quando omitida.
- Serviços `Emails`, `Domains`, `Templates`, `Audiences` (com `Contacts`), `Broadcasts`, `Senders`, `Suppressions`, `Webhooks` e `Stats`, cobrindo as 61 operações da API de produto, com o mesmo contrato dos SDKs Node.js, Python, PHP e Java.
- `*coffeemail.Error` com código canônico (`Code`), código original da API (`APICode`), `Details`, `RequestID` e `RetryAfter`, e sentinelas para `errors.Is`.
- Novas tentativas opcionais (`RetryPolicy`), com backoff exponencial, jitter, respeito ao `Retry-After` e chave de idempotência gerada para envios.
- Iteradores `ListAll` (`iter.Seq2`) para emails, broadcasts e supressões.
- Introspecção da chave de API com cache de 5 minutos.
- Subpacote `webhook` com `Verify`, `Parse`, `ParseRequest` e `Sign`, compatível com a assinatura enviada pela API.
- Mensagens de erro em português (padrão) e inglês.
