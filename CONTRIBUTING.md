# Guia de Contribuição

Obrigado pelo interesse em contribuir com o **leaks&promo**! Este documento
descreve o fluxo de trabalho e as convenções do projeto.

> Ao participar, você concorda em seguir o nosso
> [Código de Conduta](CODE_OF_CONDUCT.md).

## Pré-requisitos

- [Go](https://go.dev/dl/) (versão definida em [`go.mod`](go.mod)).
- [lefthook](https://github.com/evilmartians/lefthook) para os hooks de git locais *(config a ser adicionada)*.

## Fluxo de trabalho (GitHub Flow)

1. A branch `main` é protegida — nada de push direto.
2. Crie uma branch curta a partir de `main` (ex.: `feat/preco-handler`).
3. Faça commits pequenos e atômicos (ver convenções abaixo).
4. Abra um Pull Request; aguarde o CI ficar verde.
5. O merge é feito por *squash*, usando o título do PR (que também segue Conventional Commits).

## Build e testes

```sh
go build ./...
go test ./...
go test -race ./...
```

> Quando o `Makefile` estiver disponível, use `make help` para ver os atalhos.

## Convenções de commit

Usamos [Conventional Commits](https://www.conventionalcommits.org/) em **inglês**:

```
tipo(escopo opcional): descrição no imperativo, sem ponto final
```

Tipos permitidos: `feat`, `fix`, `docs`, `chore`, `ci`, `build`, `test`,
`refactor`, `perf`, `style`, `revert`.

As mensagens são validadas localmente por um *linter* próprio em Go
(`tools/commit-lint/`, plugado no hook `commit-msg` via lefthook) e, no CI, pelo
check do título do PR.

## Idiomas

- **Commits:** inglês.
- **Documentação, comentários de código e ADRs:** PT-BR.

## Estilo de código

- Go idiomático: erros encapsulados (`%w`, `errors.Is/As`), `context` propagado,
  injeção de dependência via construtores.
- Formatação com `gofmt`/`gofumpt` e lint com `golangci-lint`.
- Comentários só quando agregam (decisão, motivo, sutileza) — sem comentários
  óbvios.

## Reportando problemas

Use os modelos de *issue* do repositório *(a serem adicionados)*. Para questões
de **segurança**, siga a [política de segurança](SECURITY.md) — não abra issue
pública.
