# AGENTS.md

Orientações para agentes de código neste repositório. As regras completas estão no [CONTRIBUTING.md](CONTRIBUTING.md); este arquivo resume o que mais importa.

## Comandos

```bash
go run ./internal/tools/standards ./...   # regras da casa
go vet ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run ./...
go test -race -cover ./...
go test -run TestConformance ./...        # contrato com spec/product.json
go run ./internal/tools/syncspec -check   # snapshot do spec atualizado?
```

## Regras inegociáveis

- Contrato idêntico aos SDKs Node.js, Python, PHP, Java e Ruby: mesmas operações, códigos de erro canônicos, headers, mensagens locais e esquema de webhook.
- API pública: só adicionar e depreciar. Nunca renomear, remover ou mudar assinatura.
- Sem opção de URL base. Testes injetam `*http.Client` com o `fakeTransport`.
- Zero dependências de runtime; `go.mod` fica em `go 1.24`.
- Sem comentários fora de doc de exportado; sem `else`/`switch`/type switch; profundidade ≤ 3; sem `panic` e saída em console; sem números e textos mágicos.
- Todo método de I/O: `ctx context.Context` primeiro, retorno `(T, error)` com `*Error`.
- Retry só em erro de rede, 5xx e 429 (exceto `PLAN_QUOTA_EXCEEDED`), e nunca em `POST` sem chave de idempotência.
- A chave de API e o segredo de webhook nunca aparecem em mensagens de erro ou representações textuais.
- Documentação, commits e PRs em português do Brasil; código em inglês.
- Este repositório é público: não cite nem linke repositórios privados.
