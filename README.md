<div align="center">

# ☕ CoffeeMail Go SDK

**SDK oficial da CoffeeMail para Go 1.24+: envio transacional, domínios, templates, audiências, campanhas, supressões, webhooks e estatísticas, sem dependências além da biblioteca padrão.**

[![Go Reference](https://pkg.go.dev/badge/github.com/Coffee-Mail/coffee-mail-sdk-go.svg)](https://pkg.go.dev/github.com/Coffee-Mail/coffee-mail-sdk-go)
[![CI](https://github.com/Coffee-Mail/coffee-mail-sdk-go/actions/workflows/ci.yml/badge.svg)](https://github.com/Coffee-Mail/coffee-mail-sdk-go/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/go-1.24%2B-00ADD8)](#-requisitos)
[![Licença: MIT](https://img.shields.io/badge/licen%C3%A7a-MIT-blue.svg)](LICENSE)

[Instalação](#-instalação) · [Início rápido](#-início-rápido) · [Erros](#️-tratamento-de-erros) · [Referência](#-referência) · [Webhooks](#-webhooks) · [Versionamento](#-versionamento-e-releases)

</div>

---

## 📑 Sumário

- [Visão geral](#-visão-geral)
- [Status do projeto](#-status-do-projeto)
- [Requisitos](#-requisitos)
- [Instalação](#-instalação)
- [Início rápido](#-início-rápido)
- [Configuração](#️-configuração)
- [Tratamento de erros](#️-tratamento-de-erros)
- [Novas tentativas](#-novas-tentativas)
- [Paginação](#-paginação)
- [Referência](#-referência)
- [Webhooks](#-webhooks)
- [Versionamento e releases](#-versionamento-e-releases)
- [Testes e qualidade](#-testes-e-qualidade)
- [Estrutura do repositório](#️-estrutura-do-repositório)
- [Contribuição](#-contribuição)
- [Projetos relacionados](#-projetos-relacionados)
- [Licença](#-licença)

## ✨ Visão geral

- **Mesmo contrato dos SDKs oficiais** (Node.js, Python, PHP, Java, Ruby): mesmos recursos, mesmos métodos, mesmos códigos de erro e mesmas mensagens locais.
- **Idiomático**: todo método recebe `context.Context` e devolve `(T, error)`. O erro é sempre um `*coffeemail.Error`, comparável com `errors.Is` e `errors.As`. Nenhum método entra em `panic`.
- **Zero dependências de runtime**: só a biblioteca padrão. O transporte é o seu `*http.Client`.
- **Seguro para concorrência**: crie um `Client` por aplicação e compartilhe entre goroutines.
- **Retry e paginação automáticos opcionais**: novas tentativas desligadas por padrão (`WithRetry`) e iteradores `iter.Seq2` (`ListAll`) para listas com cursor.
- **Validado contra o contrato da API**: as 65 operações geram requisições validadas contra o OpenAPI de produto (`spec/product.json`) em `TestConformanceEveryRequestMatchesTheSpec`.
- **Webhooks verificados com o algoritmo da API**: o subpacote `webhook` passa nos vetores de teste gerados pelo assinador real.

## 🚦 Status do projeto

| Item | Estado |
| --- | --- |
| Versão | `0.1.0` (ainda não publicada) |
| Cobertura da API de produto | 65 operações (todas as do `product.json`, exceto `GET /v1/product/auth/csrf`, exclusiva do dashboard) |
| Idiomas das mensagens de erro | `pt-BR` (padrão) e `en` |

Limitações conhecidas:

- **Pré-1.0**: a série `0.x` segue semver, mas a política da casa já proíbe quebrar a superfície pública.
- **`Templates.PreviewByID`** não é um endpoint próprio: faz `GET /templates/{id}` e depois `POST /templates/preview` com o HTML do template (duas requisições).
- **Campos que podem ser limpos** (`description`, `alias`...): os parâmetros de atualização usam ponteiros (`nil` = não alterar). Enviar `null` explícito para limpar um campo ainda não é suportado.
- **Lacuna do `product.json`**: `DELETE /domains/{id}` declara no spec um corpo de reautenticação obrigatório, que a API só exige em sessões do dashboard. O teste de contrato trata essa rota como exceção explícita.

## 📋 Requisitos

- Go 1.24 ou superior.
- Uma chave de API (`cm_live_...` ou `cm_test_...`), criada no dashboard da CoffeeMail.

## 📦 Instalação

```bash
go get github.com/Coffee-Mail/coffee-mail-sdk-go
```

```go
import coffeemail "github.com/Coffee-Mail/coffee-mail-sdk-go"
```

## 🚀 Início rápido

```go
package main

import (
	"context"
	"log"

	coffeemail "github.com/Coffee-Mail/coffee-mail-sdk-go"
)

func main() {
	client, err := coffeemail.NewClient("") // vazio: lê COFFEEMAIL_API_KEY
	if err != nil {
		log.Fatal(err)
	}

	sent, err := client.Emails.Send(context.Background(), &coffeemail.SendEmailParams{
		From:    coffeemail.Address{Email: "no-reply@suaempresa.com.br", Name: "Sua Empresa"},
		To:      []coffeemail.Address{{Email: "cliente@gmail.com"}},
		Subject: "Bem-vindo!",
		HTML:    "<p>Que bom ter você aqui.</p>",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("email %s: %s", sent.ID, sent.Status)
}
```

Com template:

```go
sent, err := client.Emails.Send(ctx, &coffeemail.SendEmailParams{
	From:       coffeemail.Address{Email: "no-reply@suaempresa.com.br"},
	To:         []coffeemail.Address{{Email: "cliente@gmail.com"}},
	TemplateID: "3f0c...",
	Variables:  map[string]any{"nome": "Ana"},
})
```

> **Idempotência:** sem `IdempotencyKey`, a API trata envios com payload idêntico como o mesmo envio e devolve o id original. Para enviar duas vezes a mesma mensagem de propósito, use chaves diferentes; para evitar duplicidade em novas tentativas, use sempre a mesma chave.

## ⚙️ Configuração

```go
client, err := coffeemail.NewClient("cm_live_...",
	coffeemail.WithTimeout(5*time.Second),                        // padrão: 10s por tentativa
	coffeemail.WithLocale(coffeemail.LocaleEn),                   // padrão: LocalePtBR
	coffeemail.WithRetry(coffeemail.RetryPolicy{MaxAttempts: 3}), // padrão: sem retry
	coffeemail.WithHTTPClient(&http.Client{Transport: meuTransporte}),
)
```

| Opção | Efeito |
| --- | --- |
| `WithTimeout` | Tempo limite de cada tentativa. Um deadline no `context` continua valendo. |
| `WithLocale` | Idioma das mensagens locais e do header `Accept-Language`. |
| `WithRetry` | Ativa novas tentativas automáticas (ver [Novas tentativas](#-novas-tentativas)). |
| `WithHTTPClient` | `*http.Client` próprio: proxy corporativo, OpenTelemetry, testes. |

O endereço da API (`https://api.coffeemail.com.br`) é fixo e não configurável, para que a chave de API nunca seja enviada a outro host. O client padrão não segue redirecionamentos.

Para correlacionar logs, envie um id de requisição:

```go
ctx = coffeemail.WithRequestID(ctx, "pedido-1234")
```

## 🛡️ Tratamento de erros

Todo erro é um `*coffeemail.Error`:

```go
_, err := client.Domains.Get(ctx, id)
if errors.Is(err, coffeemail.ErrNotFound) {
	// domínio não existe
}

var apiErr *coffeemail.Error
if errors.As(err, &apiErr) {
	log.Printf("status=%d code=%s apiCode=%s requestId=%s: %s",
		apiErr.Status, apiErr.Code, apiErr.APICode, apiErr.RequestID, apiErr.Message)
}
```

| Status | `Code` | Sentinela |
| --- | --- | --- |
| 400 | `VALIDATION_ERROR` | `ErrValidation` |
| 401 | `UNAUTHORIZED` | `ErrUnauthorized` |
| 402 | `PAYMENT_REQUIRED` | `ErrPaymentRequired` |
| 403 | `FORBIDDEN` | `ErrForbidden` |
| 403 em broadcasts | `PERMISSION_DENIED` | `ErrPermissionDenied` |
| 404 | `NOT_FOUND` | `ErrNotFound` |
| 409 | `CONFLICT` | `ErrConflict` |
| 429 | `RATE_LIMIT_EXCEEDED` | `ErrRateLimited` |
| 500, 502, 503, 504 | `INTERNAL_SERVER_ERROR` | `ErrServer` |
| falha de rede / timeout | `NETWORK_ERROR` (status 0) | `ErrNetwork` |
| 2xx com corpo ilegível | `INVALID_RESPONSE` | `ErrInvalidResponse` |
| outro | código da API ou `API_ERROR` | — |

- `APICode` traz o código original da API. Exemplo: um `429` pode ser `RATE_LIMIT_EXCEEDED` (limite de requisições) ou `PLAN_QUOTA_EXCEEDED` (cota do plano esgotada).
- `Details` reúne `meta` e `details` do erro da API (em validação, `Details["issues"]` lista os campos inválidos).
- `RetryAfter` traz o `Retry-After` informado pela API.
- Erros de rede preservam a causa: `errors.Is(err, context.DeadlineExceeded)` funciona.
- Sem chave de API, `NewClient` devolve um erro `UNAUTHORIZED` cuja causa é `ErrMissingAPIKey`.

## 🔁 Novas tentativas

Desligadas por padrão. Com `WithRetry(coffeemail.RetryPolicy{MaxAttempts: 3})`:

- repetem só erros de rede, `5xx` e `429` que não sejam `PLAN_QUOTA_EXCEEDED`;
- valem para `GET` e `DELETE`. Em `POST`, valem só para `Emails.Send` e `Emails.SendBatch`, que aceitam chave de idempotência. Se você não informar `IdempotencyKey`, o SDK gera uma e a reutiliza em todas as tentativas, então o email nunca é duplicado;
- usam backoff exponencial com jitter (`BaseDelay` padrão 500ms, teto `MaxDelay` padrão 30s) e respeitam o `Retry-After`, limitado ao teto;
- param imediatamente quando o `context` é cancelado.

## 📄 Paginação

`List` devolve uma página e o cursor da próxima. `ListAll` percorre todas as páginas sob demanda (`iter.Seq2`); interrompa o laço para parar:

```go
for email, err := range client.Emails.ListAll(ctx, &coffeemail.ListEmailsParams{Status: coffeemail.EmailStatusBounced}) {
	if err != nil {
		return err
	}
	fmt.Println(email.ID, email.To)
}
```

`ListAll` existe em `Emails`, `Broadcasts` e `Suppressions` (cursor). Audiências e contatos paginam por deslocamento (`OffsetParams{Limit, Offset}`).

## 📚 Referência

Parâmetros opcionais podem ser omitidos. Onde o valor zero é significativo (por exemplo `IsActive: false`), o campo é ponteiro: use `coffeemail.Ptr(false)`. Enums são tipos `string` abertos: valores novos da API não quebram a desserialização.

### Introspecção da chave de API

| Método | Descrição |
| --- | --- |
| `client.Introspect(ctx)` | Dados da chave em uso (`ID`, `Name`, `OrganisationID`, `Scopes`). Cache de 5 minutos. |
| `client.InvalidateAPIKeyCache()` | Descarta o cache da introspecção. |

### E-mails — `client.Emails`

| Método | Rota |
| --- | --- |
| `Send(ctx, *SendEmailParams)` | `POST /emails` |
| `SendBatch(ctx, *SendBatchParams)` | `POST /emails/batch` (até 100) |
| `Get(ctx, id)` | `GET /emails/{id}` |
| `List(ctx, *ListEmailsParams)` / `ListAll` | `GET /emails` |
| `GetEvents(ctx, id)` | `GET /emails/{id}/events` |
| `GetTags(ctx)` | `GET /emails/tags` |
| `Cancel(ctx, id)` | `POST /emails/{id}/cancel` |
| `Resend(ctx, id)` | `POST /emails/{id}/resend` |

### Domínios — `client.Domains`

| Método | Rota |
| --- | --- |
| `Create(ctx, name)` | `POST /domains` |
| `List(ctx)` | `GET /domains` |
| `Get(ctx, id)` | `GET /domains/{id}` |
| `Verify(ctx, id)` | `POST /domains/{id}/verify` |
| `Delete(ctx, id)` | `DELETE /domains/{id}` |
| `GetHealth(ctx, id)` | `POST /domains/{id}/health` |
| `GetWarmupStatus(ctx, id)` | `GET /domains/{id}/warmup` (nil sem aquecimento) |
| `RotateDKIM(ctx, id)` | `POST /domains/{id}/dkim/rotate` |
| `SetTrackingDomain(ctx, id, host)` | `PUT /domains/{id}/tracking-domain` |
| `RemoveTrackingDomain(ctx, id)` | `DELETE /domains/{id}/tracking-domain` (nil em caso de sucesso) |

`Get` traz também a entregabilidade do domínio: `DKIMRotation` (rotação de DKIM pendente, ou nil), `ReturnPath` e `TrackingDomain` (`*CustomHostStatus` com host, alvo do CNAME, verificação e registro DNS, ou nil) e `TrackingDomainsAvailable`. `GetWarmupStatus` informa a fase (`DomainWarmupWarming` ou `DomainWarmupCompleted`), `TotalDays`, `StartedAt` e `CompletedAt`.

`RotateDKIM` é idempotente: enquanto a rotação está pendente, devolve a mesma rotação; a chave em uso só muda quando o registro TXT do novo seletor aparece no DNS. Domínio não verificado responde `ErrConflict` com `APICode` `DOMAIN_NOT_VERIFIED`. Em `SetTrackingDomain`, host inválido responde `ErrValidation` (`INVALID_TRACKING_HOST`) e recurso desligado na plataforma, `ErrConflict` (`TRACKING_DOMAINS_DISABLED`).

### Templates — `client.Templates`

| Método | Rota |
| --- | --- |
| `Create(ctx, *CreateTemplateParams)` | `POST /templates` |
| `List(ctx)` | `GET /templates` |
| `Get(ctx, id)` | `GET /templates/{id}` |
| `Update(ctx, id, *UpdateTemplateParams)` | `PATCH /templates/{id}` |
| `Delete(ctx, id)` | `DELETE /templates/{id}` |
| `Preview(ctx, *PreviewTemplateParams)` | `POST /templates/preview` |
| `PreviewByID(ctx, id, variables)` | `GET /templates/{id}` + `POST /templates/preview` |
| `Format(ctx, *FormatTemplateParams)` | `POST /templates/format` |
| `TestRender(ctx, *TestRenderTemplateParams)` | `POST /templates/test-render` |
| `TestSend(ctx, id, *TestSendTemplateParams)` | `POST /templates/{id}/test-send` |

### Audiências e contatos — `client.Audiences`

| Método | Rota |
| --- | --- |
| `Create(ctx, *CreateAudienceParams)` | `POST /audiences` |
| `List(ctx, *OffsetParams)` | `GET /audiences` |
| `Get(ctx, id)` | `GET /audiences/{id}` |
| `Update(ctx, id, *UpdateAudienceParams)` | `PUT /audiences/{id}` |
| `Delete(ctx, id)` | `DELETE /audiences/{id}` |
| `Contacts.Create(ctx, audienceID, *CreateContactParams)` | `POST /audiences/{id}/contacts` |
| `Contacts.List(ctx, audienceID, *OffsetParams)` | `GET /audiences/{id}/contacts` |
| `Contacts.BulkAdd(ctx, audienceID, []CreateContactParams)` | `POST /audiences/{id}/contacts/bulk` (até 10 000) |
| `Contacts.Update(ctx, audienceID, contactID, *UpdateContactParams)` | `PUT /audiences/{id}/contacts/{contactId}` |
| `Contacts.Delete(ctx, audienceID, contactID)` | `DELETE /audiences/{id}/contacts/{contactId}` |

### Campanhas — `client.Broadcasts`

Toda operação exige chave `full_access`. Sem ela, o erro tem código `PERMISSION_DENIED`.

| Método | Rota |
| --- | --- |
| `Create(ctx, *CreateBroadcastParams)` | `POST /broadcasts` |
| `List(ctx, *ListBroadcastsParams)` / `ListAll` | `GET /broadcasts` |
| `Get(ctx, id)` | `GET /broadcasts/{id}` |
| `Send(ctx, id)` | `POST /broadcasts/{id}/send` |
| `TestSend(ctx, id, *TestSendBroadcastParams)` | `POST /broadcasts/{id}/test-send` |
| `Cancel(ctx, id)` | `POST /broadcasts/{id}/cancel` |

### Remetentes — `client.Senders`

| Método | Rota |
| --- | --- |
| `Create(ctx, *CreateSenderParams)` | `POST /senders` |
| `List(ctx)` | `GET /senders` |
| `Verify(ctx, token)` | `POST /senders/verify` |
| `Delete(ctx, id)` | `DELETE /senders/{id}` |

### Supressões — `client.Suppressions`

| Método | Rota |
| --- | --- |
| `Create(ctx, *CreateSuppressionParams)` | `POST /suppressions` |
| `List(ctx, *ListSuppressionsParams)` / `ListAll` | `GET /suppressions` |
| `Get(ctx, id)` | `GET /suppressions/{id}` |
| `Delete(ctx, id)` | `DELETE /suppressions/{id}` |
| `Reactivate(ctx, id)` | `POST /suppressions/{id}/reactivate` |
| `BulkCreate(ctx, []BulkSuppressionEntry)` | `POST /suppressions/bulk` (até 1000) |

### Webhooks — `client.Webhooks`

| Método | Rota |
| --- | --- |
| `Create(ctx, *CreateWebhookParams)` | `POST /webhooks` (o `Secret` vem só nesta resposta) |
| `List(ctx, *ListWebhooksParams)` | `GET /webhooks` |
| `Get(ctx, id)` | `GET /webhooks/{id}` |
| `Update(ctx, id, *UpdateWebhookParams)` | `PUT /webhooks/{id}` |
| `Toggle(ctx, id, status)` | `PATCH /webhooks/{id}` |
| `Delete(ctx, id)` | `DELETE /webhooks/{id}` |
| `RotateSecret(ctx, id)` | `POST /webhooks/{id}/rotate-secret` |
| `ListDeliveries(ctx, id, *ListWebhookDeliveriesParams)` | `GET /webhooks/{id}/deliveries` |
| `Test(ctx, id)` | `POST /webhooks/{id}/test` |

### Estatísticas — `client.Stats`

| Método | Rota |
| --- | --- |
| `Get(ctx, *GetStatsParams)` | `GET /stats` |

## 🔔 Webhooks

O subpacote `webhook` verifica a assinatura das entregas. Ele não depende do `Client` nem de chave de API:

```go
import "github.com/Coffee-Mail/coffee-mail-sdk-go/webhook"

http.HandleFunc("/webhooks/coffeemail", func(w http.ResponseWriter, r *http.Request) {
	event, err := webhook.ParseRequest(r, os.Getenv("COFFEEMAIL_WEBHOOK_SECRET"))
	if err != nil {
		http.Error(w, "assinatura inválida", http.StatusUnauthorized)
		return
	}
	log.Printf("%s: %v", event.EventType, event.Data)
	w.WriteHeader(http.StatusNoContent)
})
```

- A assinatura (`X-Webhook-Signature`) é `sha256=` + HMAC-SHA256 em hexadecimal de `timestamp + "." + idempotencyKey + "." + corpo bruto`, comparada em tempo constante.
- Entregas com `X-Webhook-Timestamp` fora da janela (padrão 5 minutos, `webhook.WithTolerance`) são rejeitadas para impedir replay.
- Já leu o corpo? Use `webhook.Parse(payload, r.Header, secret)` ou `webhook.Verify(webhook.DeliveryFromHeaders(payload, r.Header), secret)`. Nunca re-serialize o JSON antes de verificar.
- `webhook.Sign` calcula a assinatura esperada, útil para testar o seu handler.
- `X-Webhook-Idempotency-Key` muda a cada tentativa de entrega e não serve para deduplicar retentativas. O corpo (`event_type`, `webhook_id`, `data`) não traz o id do email.

## 🔖 Versionamento e releases

- [Versionamento Semântico](https://semver.org/lang/pt-BR/). A política da casa vale desde a `0.x`: a API pública só recebe **adições e depreciações**, nunca remoção ou renomeação. O CI compara a API com a última versão publicada (`gorelease`).
- A versão está em `coffeemail.Version` e vai no `User-Agent` (`coffeemail-go/<versão>`).
- Uma tag `vX.Y.Z` dispara o workflow de release: confere a tag com `Version` e o `CHANGELOG.md`, roda os gates, cria a GitHub Release e publica no proxy de módulos do Go.
- Mudanças ficam no [CHANGELOG.md](CHANGELOG.md).

## 🧪 Testes e qualidade

```bash
go run ./internal/tools/standards ./...   # regras da casa (sem else/switch/comentários fora de doc, profundidade ≤ 3)
go vet ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run ./...
go test -race -cover ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
```

- Os testes usam um `RoundTripper` falso injetado por `WithHTTPClient`; nenhum teste acessa a rede.
- `TestConformanceEveryRequestMatchesTheSpec` valida as requisições das 65 operações contra `spec/product.json` e falha se alguma operação do spec ficar sem método.
- `webhook` roda os vetores compartilhados de `testdata/webhook_vectors.json`, gerados pelo assinador da API.
- Cobertura mínima de 85% no CI.

Para atualizar o snapshot do spec:

```bash
go run ./internal/tools/syncspec                    # do spec público
COFFEEMAIL_OPENAPI_SPEC=caminho/product.json go run ./internal/tools/syncspec
go run ./internal/tools/syncspec -check             # só verifica divergência
```

## 🗂️ Estrutura do repositório

```
├── client.go, options.go, retry.go     Client, opções e política de retry
├── errors.go, error_mapping.go         Error, sentinelas e mapeamento status → código
├── requester.go, operation.go          transporte: headers, timeout, retry, decodificação
├── emails.go … stats.go                um arquivo por recurso
├── pagination.go                       iteradores ListAll
├── webhook/                            verificação de assinatura
├── internal/                           rotas, headers, i18n, backoff, idempotência
│   └── tools/                          standards (regras da casa) e syncspec
├── spec/product.json                   snapshot do OpenAPI de produto
├── testdata/                           vetores de webhook
└── examples/                           exemplos executáveis
```

## 🤝 Contribuição

Leia o [CONTRIBUTING.md](CONTRIBUTING.md). Vulnerabilidades: [SECURITY.md](SECURITY.md).

## 🔗 Projetos relacionados

| Repositório | Papel |
| --- | --- |
| [coffee-mail-sdk-node](https://github.com/Coffee-Mail/coffee-mail-sdk-node) | SDK oficial Node.js/TypeScript (`@coffeemail/node`) |
| [coffee-mail-sdk-python](https://github.com/Coffee-Mail/coffee-mail-sdk-python) | SDK oficial Python (`coffeemail`) |
| [coffee-mail-sdk-php](https://github.com/Coffee-Mail/coffee-mail-sdk-php) | SDK oficial PHP (`coffeemail/coffeemail-php`) |
| [coffee-mail-sdk-java](https://github.com/Coffee-Mail/coffee-mail-sdk-java) | SDK oficial Java (`br.com.coffeemail:coffeemail-java`) |
| **coffee-mail-sdk-go** (você está aqui) | SDK oficial Go (`github.com/Coffee-Mail/coffee-mail-sdk-go`) |
| [coffee-mail-sdk-ruby](https://github.com/Coffee-Mail/coffee-mail-sdk-ruby) | SDK oficial Ruby (`coffeemail`) |
| [coffee-mail-laravel](https://github.com/Coffee-Mail/coffee-mail-laravel) | Integração Laravel (`coffeemail/coffeemail-laravel`) |
| [coffee-mail-sdk-workbench](https://github.com/Coffee-Mail/coffee-mail-sdk-workbench) | Bancada de validação dos SDKs por spec |

## 📄 Licença

[MIT](LICENSE) © 2026 Dev-Etto.
