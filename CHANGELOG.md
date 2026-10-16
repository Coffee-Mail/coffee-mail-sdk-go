# Changelog

Todas as mudanças relevantes deste projeto são documentadas aqui. O formato segue o [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/) e o projeto adota o [Versionamento Semântico](https://semver.org/lang/pt-BR/).

## [Não lançado]

### Adicionado

- `CreateBroadcastParams.ScheduledAt` agenda o broadcast (ele nasce `BroadcastStatusScheduled`).
- `Broadcast.Metrics` (`*BroadcastMetrics`) com entregues, abertos, cliques, bounces, falhas e suprimidos, preenchido em `Get`.
- `SendersService.ResendVerification` reenvia o email de verificação com token novo.
- `Contact.Subscribed` e `Contact.IsSubscribed()` (`false` quando o contato se descadastrou).

### Alterado

- `BroadcastsService.Send` responde assim que o envio é enfileirado (`BroadcastStatusSending`); o progresso acompanha-se por `Get`.

### Depreciado

- `CreatedSender.VerificationToken`: a API não devolve mais o token, que segue por email ao endereço cadastrado; o campo vem vazio.

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
