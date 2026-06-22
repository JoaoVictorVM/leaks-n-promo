# Deploy gateway + self-host containerizado

- Status: aceito
- Data: 2026-06-22
- Decisores: João Victor Ventura Martins

## Contexto e problema

Por ser um processo de gateway de longa duração (ver
[ADR-0001](0001-linguagem-go-e-modelo-gateway.md)) e um projeto de portfólio com
custo zero, é preciso definir como o bot é empacotado e executado.

## Opções consideradas

- **Self-host containerizado** (máquina própria ou Raspberry Pi).
- PaaS/serviço gerenciado de containers.
- Função serverless (descartada no ADR-0001).

## Decisão

Empacotar o bot em **imagem Docker multi-stage** (final `distroless`, non-root) e
fazer **self-host containerizado**, documentado no runbook (incluindo unit do
systemd de exemplo). A régua de portfólio é "containerizado + deploy
documentado", não uptime 24/7.

## Consequências

### Positivas

- Custo zero e portabilidade (roda em qualquer host com Docker).
- Imagem pequena e segura (sem shell, non-root).

### Negativas

- Disponibilidade depende do host caseiro; sem SLA.
- Operação manual (start/stop/atualização) a cargo do mantenedor.
