# leaks&promo

> Bot de Discord, open-source, escrito em Go, para o mundo dos games: consulta de **preços de jogos de PC** e agregação de **vazamentos (leaks) de games**, sob demanda.

<!-- Badges (a adicionar conforme os workflows de CI/CD entrarem no ar):
     build · cobertura · go report card · openssf scorecard · license · release -->

## Sobre

**leaks&promo** é um bot de Discord com arquitetura **pull-only** (comando → busca → resposta) e dois comandos:

- `/preco <jogo>` — preços do jogo nas lojas de PC (via CheapShark, em USD).
- `/leaks [termo]` — vazamentos e rumores recentes de games (via RSS e, opcionalmente, Reddit).

É, antes de tudo, um **projeto de portfólio**: o foco é demonstrar excelência de engenharia (arquitetura, testes, CI/CD, segurança e documentação), com **custo zero** de operação.

## Comandos

| Comando | Descrição |
|---|---|
| `/preco <jogo>` | Lista as lojas de PC e os preços de um jogo. |
| `/leaks [termo]` | Lista vazamentos/rumores recentes, com filtro opcional por termo. |

> Referência detalhada: `docs/commands.md` *(a criar)*.

## Stack

- **Go** (gateway via `discordgo`), logging com `log/slog`, config 12-factor.
- Cache em memória com TTL; resiliência com `context`, timeout e retry com backoff.
- Docker multi-stage (imagem final non-root); GitHub Actions; GoReleaser → GHCR.

## Desenvolvimento

> Em construção. Instruções de build, testes e execução local serão adicionadas conforme as fases do roadmap avançam.

```sh
go build ./...
go test ./...
```

## Arquitetura

Pull-only e stateless na v1. As fontes de dados ficam atrás de interfaces (`PriceProvider`, `LeakSource`) para permitir troca sem reescrita.

> Diagramas C4 e de sequência: `docs/architecture/` *(a criar)*.

## Contribuindo

Veja `CONTRIBUTING.md` *(a criar)* e o `CODE_OF_CONDUCT.md` *(a criar)*.

## Segurança e privacidade

- Política de segurança: `SECURITY.md` *(a criar)*.
- Privacidade: o bot é **stateless** — não persiste ID de usuário nem consultas. Detalhes em `PRIVACY.md` *(a criar)*.

## Licença

Distribuído sob a licença [MIT](LICENSE).

Atribuições das fontes de dados em `NOTICE` *(a criar)*.
