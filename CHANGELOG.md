# Changelog

Todas as mudanças relevantes deste projeto são documentadas aqui. O formato segue o [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/) e o projeto adota o [Versionamento Semântico](https://semver.org/lang/pt-BR/).

## [Não lançado]

### Adicionado

- `CreateBroadcastParams.ScheduledAt` agenda o broadcast (ele nasce `BroadcastStatusScheduled`).
- `Broadcast.Metrics` (`*BroadcastMetrics`) com entregues, abertos, cliques, bounces, falhas e suprimidos, preenchido em `Get`.
- `SendersService.ResendVerification` reenvia o email de verificação com token novo.
- `Contact.Subscribed` e `Contact.IsSubscribed()` (`false` quando o contato se descadastrou).
- `DomainsService.RotateDKIM` inicia a rotação da chave DKIM (idempotente) e devolve `*DKIMRotation` com o registro TXT do novo seletor.
- `DomainsService.SetTrackingDomain` e `DomainsService.RemoveTrackingDomain` cadastram e removem o host de tracking próprio (`*CustomHostStatus`).
- `DomainDetail.DKIMRotation`, `DomainDetail.ReturnPath`, `DomainDetail.TrackingDomain` e `DomainDetail.TrackingDomainsAvailable` em `DomainsService.Get`.
- `DomainWarmupStatus.Status` (`DomainWarmupWarming`/`DomainWarmupCompleted`), `TotalDays`, `StartedAt` e `CompletedAt`.
- `TemplatesService.ListVersions`, `GetVersion` e `RestoreVersion` listam, detalham e restauram versões de um template (`TemplateVersionSummary`, `TemplateVersion`, `ListTemplateVersionsResponse`).
- `WebhooksService.GetDelivery` (entrega com `RequestPayload`, `*WebhookDeliveryDetail`) e `WebhooksService.ReplayDelivery` (reenfileira o mesmo corpo e devolve a entrega nova).
- `Client.Validation` (`ValidationService.Validate`) valida até 100 endereços e devolve o saldo de créditos do mês (`ValidateEmailsResponse`, `EmailValidation`, `ValidationCredits`).

### Alterado

- `SendEmailParams.TemplateID` aceita o id ou o alias do template.
- 429 com `APICode` `VALIDATION_CREDITS_EXHAUSTED` não é repetido pelo retry automático, como `PLAN_QUOTA_EXCEEDED`.
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
