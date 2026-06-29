# Persistência opcional (SQLite) e push de leaks

- Status: aceito
- Data: 2026-06-29
- Decisores: João Victor Ventura Martins

## Contexto e problema

A Fase 3 prevê **notificações push** de novos vazamentos. Para não notificar o
mesmo item duas vezes, é preciso **lembrar o que já foi enviado** entre
reinícios — ou seja, persistência. Isso conflita com a decisão de ser
**stateless na v1** ([ADR-0002](0002-arquitetura-pull-only-stateless.md)).

## Opções consideradas

- **Manter stateless** e abrir mão do push.
- **Persistência opcional (opt-in)** com SQLite local, habilitada só quando
  configurada.
- **Serviço externo** de estado (Redis/DB gerenciado) — fere o custo zero.

## Decisão

Introduzir **persistência opcional** via **SQLite**, usada apenas para
deduplicação do push. Pontos da decisão:

- **Driver `modernc.org/sqlite` (SQLite puro em Go):** evita CGO, preservando o
  build estático (`CGO_ENABLED=0`) e a imagem final distroless
  ([ADR-0004](0004-deploy-gateway-self-host.md)).
- **Push como scheduler interno** ao processo do bot (ticker + `context`), com o
  arquivo SQLite em um **volume Docker**. Coerente com o deploy self-host; o
  estado persiste naturalmente, sem depender de runners efêmeros.
- **Opt-in:** sem configuração de persistência/push, o bot **continua
  stateless** — o comportamento padrão da v1 é preservado.

Esta decisão **amenda** a ADR-0002: o stateless deixa de ser absoluto e passa a
ser o padrão quando o push não está habilitado.

## Consequências

### Positivas

- Notificações proativas de leaks sem custo adicional.
- Dedup persistente sobrevive a reinícios.
- Mantém a toolchain sem CGO e a imagem distroless.
- O dado persistido são **URLs já notificadas** — não há PII de usuário, então a
  postura de privacidade ([PRIVACY.md](../../PRIVACY.md)) permanece válida.

### Negativas

- Quando habilitado, há estado a operar: volume, espaço em disco e eventual
  limpeza/retention da tabela de URLs.
- Amplia a superfície (arquivo em disco, novo dependente: o driver SQLite).
- Um scheduler interno adiciona um caminho concorrente ao gateway pull-only.
