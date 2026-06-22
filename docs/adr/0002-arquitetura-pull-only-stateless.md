# Arquitetura pull-only e stateless na v1

- Status: aceito
- Data: 2026-06-22
- Decisores: João Victor Ventura Martins

## Contexto e problema

O bot oferece preços e leaks. É preciso decidir se ele coleta dados de forma
proativa (push, com coletor agendado e banco) ou apenas sob demanda (pull), e se
mantém estado.

## Opções consideradas

- **Pull-only e stateless** — tudo é "comando → busca → resposta", sem banco.
- **Push com coletor agendado + banco** — notificações proativas e dedup
  persistente desde o início.

## Decisão

Adotar **pull-only e stateless** na v1. Cada comando dispara uma busca nas fontes
e responde; um **cache em memória com TTL** reduz chamadas externas. Sem banco e
sem persistência de dados de usuário.

## Consequências

### Positivas

- Superfície de segurança mínima: sem dado de usuário a vazar (ver
  [PRIVACY.md](../../PRIVACY.md)).
- Operação e deploy simples; custo zero.

### Negativas

- Sem notificações proativas e sem histórico na v1 (adiados para a Fase 3 / v2.0,
  quando o SQLite reaparece).
- Cache volátil: reinícios esvaziam o cache.
