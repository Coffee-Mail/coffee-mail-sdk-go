# Contribuindo com o CoffeeMail Go SDK

Obrigado pelo interesse! Este guia resume como propor mudanças.

## Antes de começar

- Bugs e sugestões: abra uma issue usando um dos modelos.
- Vulnerabilidades: **não** abra issue pública; siga o [SECURITY.md](SECURITY.md).
- Mudanças grandes ou que alterem a API pública: abra uma issue para discussão antes do PR.

## Ambiente

- Go 1.24 ou superior. O `go.mod` declara `go 1.24` e o CI falha se a diretiva mudar: não atualize dependências que exijam uma versão mais nova do Go.
- Antes de qualquer PR, rode:

```bash
go run ./internal/tools/standards ./...
go vet ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run ./...
go test -race -cover ./...
```

## Regras de código

O SDK segue as regras de código da CoffeeMail, verificadas automaticamente:

- Código em inglês; documentação, issues, commits e PRs em português do Brasil.
- Sem comentários, exceto a documentação (doc comment) de identificadores exportados e do pacote. Diretivas `//go:` são permitidas; `//nolint` não: exceções de lint ficam no `.golangci.yml`, com justificativa.
- Sem `else`, `else if`, `switch` e type switch: use retorno antecipado, tabelas (`map`) e `errors.As`.
- Profundidade máxima de 3 blocos aninhados por função.
- Sem `panic`, `print`/`println`, `fmt.Print*` e `log.*` em código de biblioteca.
- Sem números e textos mágicos: use constantes nomeadas (rotas em `internal/routes`, headers em `internal/headers`).
- Sem dependências de runtime: só a biblioteca padrão. Dependências de teste precisam de justificativa no PR.
- Todo método que faz I/O recebe `context.Context` como primeiro parâmetro e devolve `(T, error)` com `*Error`.
- Cada comportamento novo vem com testes: caminho feliz e ao menos dois caminhos de erro. Os testes usam o `fakeTransport` de `helpers_test.go`; nenhum teste acessa a rede.

## Compatibilidade

O SDK é usado em produção. Mudanças na API pública devem **adicionar e depreciar** (`// Deprecated:` no doc comment), nunca renomear ou remover. O job `compatibilidade da API pública` do CI compara a API com a última versão publicada.

O endereço da API é fixo: não adicione opção de URL base. Testes trocam o transporte, nunca a URL.

## Mudanças no contrato da API

Quando o OpenAPI da API mudar:

1. `go run ./internal/tools/syncspec` (ou com `COFFEEMAIL_OPENAPI_SPEC` apontando para um arquivo local);
2. ajuste tipos e serviços;
3. adicione o caso em `conformance_test.go` e mantenha `TestConformanceEveryRequestMatchesTheSpec` verde.

## Commits e PRs

- Mensagens em português, diretas, no formato `tipo(escopo): descrição` (ex.: `feat(emails): expõe metadata no envio`).
- Atualize o `CHANGELOG.md` na seção "Não lançado".
- Preencha o modelo de PR.
