# Threat Model (STRIDE leve) — leaks&promo

Modelo de ameaças resumido, no estilo **STRIDE**, adequado ao escopo da v1:
bot **pull-only** e **stateless**. O princípio central é reduzir a superfície de
ataque por design — sem persistência, não há dado de usuário a vazar.

## Escopo e ativos

- **Token do Discord** (segredo de maior valor).
- **Credenciais do Reddit** (OAuth), quando habilitado.
- **Integridade do processo** e da imagem do container.
- **Entrada do usuário** nos comandos (`/preco`, `/leaks`).
- **Chamadas a APIs externas** (CheapShark, RSS, Reddit).

### Fora de escopo

- Dados pessoais de usuários — **não são coletados nem armazenados** (ver
  [PRIVACY.md](../PRIVACY.md)).
- Disponibilidade 24/7 / SLA — é projeto de portfólio self-hosted.

## Fronteiras de confiança

1. Discord ↔ bot (gateway autenticado por token).
2. Bot ↔ APIs externas (HTTPS de saída).
3. Host ↔ container (isolamento Docker, non-root).

## Análise STRIDE

| Categoria | Ameaça | Mitigação |
|---|---|---|
| **S**poofing | Uso indevido do token do Discord para se passar pelo bot | Token só em `.env` (gitignored) / GitHub Secrets; nunca commitado; `gitleaks` no CI; intents mínimos. |
| **T**ampering | Adulteração de dependências ou da imagem (supply chain) | Actions pinadas por SHA; `dependabot`; SBOM no release; imagem distroless; `Dependency Review` e `Trivy`. |
| **R**epudiation | Falta de rastro de ações | Logging estruturado (`slog`) **sem PII/segredo**; foco em diagnóstico, não em auditoria de usuário (stateless). |
| **I**nformation Disclosure | Vazamento de segredo ou de detalhe interno em erro/log | Erros contidos (mensagem de fallback, sem stack); logs sem segredo; sem persistência de dados de usuário. |
| **D**enial of Service | Abuso de comandos / esgotar rate limit das fontes | Cache em memória com TTL; `context` com timeout; retry com backoff; tratamento de HTTP 429 com fallback. |
| **E**levation of Privilege | Comprometer o host a partir do container | Container **non-root**, imagem mínima (sem shell), `CGO_ENABLED=0`; superfície reduzida. |

## Riscos residuais aceitos

- **Disponibilidade**: depende do host caseiro; sem redundância (decisão de
  portfólio — ver [ADR-0004](adr/0004-deploy-gateway-self-host.md)).
- **Confiança nas fontes externas**: o conteúdo de RSS/Reddit não é verificado
  quanto à veracidade; o bot apenas agrega e direciona à fonte original.

## Revisão

Este documento deve ser revisitado quando: (a) a v1 deixar de ser stateless
(Fase 3, SQLite/push), ou (b) novas fontes/integrações forem adicionadas.
