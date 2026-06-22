# Linguagem Go e modelo gateway (discordgo)

- Status: aceito
- Data: 2026-06-22
- Decisores: João Victor Ventura Martins

## Contexto e problema

O bot precisa de uma linguagem e de um modelo de conexão com o Discord. As duas
abordagens principais são um processo de longa duração conectado ao **gateway
(websocket)** ou um modelo **serverless via HTTP Interactions**.

## Opções consideradas

- **Go + gateway (`discordgo`)** — processo de longa duração.
- **Serverless / HTTP Interactions** — funções acionadas por webhook.
- Outra linguagem (Node/TS, Python) com qualquer um dos modelos.

## Decisão

Usar **Go** com o modelo **gateway** via `discordgo`. Go é idiomático para
serviços de longa duração e concorrência, e é a linguagem-alvo deste portfólio.
O modelo serverless empurraria a stack para WASM/JS e adicionaria complexidade
de hospedagem sem ganho para o caso de uso pull.

## Consequências

### Positivas

- Código idiomático, com boa história de concorrência, testes e tooling.
- Processo único e simples de conteinerizar e self-hostar.

### Negativas

- Exige um processo sempre ativo (não escala a zero como serverless).
- Acoplamento à biblioteca `discordgo` (mitigado por manter a lógica de domínio
  fora dos handlers).
