# Roadmap

Este roadmap resume as fases de desenvolvimento do **leaks&promo**. A
especificação completa está no `PRD.md` (seção 12); este arquivo é a visão
pública e resumida.

> Legenda: ⬜ pendente · 🚧 em andamento · ✅ concluído

## Fase 0 — Esqueleto andante 🚧

Repositório profissional, sem feature de produto ainda.

- 🚧 Módulo Go, layout base e arquivos de configuração de repositório.
- 🚧 Documentação base (README, licença, código de conduta, contribuição,
  segurança, privacidade, NOTICE, changelog, roadmap).
- ⬜ Tooling de qualidade: golangci-lint, Makefile, lefthook e linter próprio de
  Conventional Commits.
- ⬜ CI/CD: lint, testes, CodeQL, dependency review, OpenSSF Scorecard e release
  com GoReleaser/SBOM.
- ⬜ Containerização: Dockerfile multi-stage e docker-compose.
- ⬜ Documentação de arquitetura: ADRs, diagramas C4, sequência, threat model e
  runbook.
- ⬜ Fundações da aplicação: config, logging estruturado, graceful shutdown,
  conexão ao gateway do Discord e injeção de versão em build-time.

## Fase 1 — Comando `/preco` (MVP) ⬜

- ⬜ Interface `PriceProvider` e client do CheapShark (com testes via fixtures).
- ⬜ Cache em memória com TTL; timeout e backoff nas consultas.
- ⬜ Registro e handler do comando `/preco` (com testes e documentação).

## Fase 2 — Comando `/leaks` ⬜

- ⬜ Interface `LeakSource` e fonte RSS (backbone), com filtragem por termo.
- ⬜ Fonte Reddit (enhancement), habilitada apenas se o app for aprovado.
- ⬜ Merge e deduplicação dos resultados por URL.
- ⬜ Registro e handler do comando `/leaks` (com testes, fuzzing e documentação).

## Fase 3 — Polimento & push opcional ⬜

- ⬜ Persistência com SQLite para deduplicação.
- ⬜ Notificações push via GitHub Actions agendado (cron).
- ⬜ Observabilidade (métricas).

## v2.0 / Futuro ⬜

- ⬜ Migração/adição do ITAD (preços em BRL e lojas brasileiras).
- ⬜ Notificações push proativas por preço-alvo.
- ⬜ Observabilidade completa (Prometheus + Grafana) e assinatura de release
  (cosign).
