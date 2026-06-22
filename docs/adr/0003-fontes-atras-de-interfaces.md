# Fontes de dados atrás de interfaces

- Status: aceito
- Data: 2026-06-22
- Decisores: João Victor Ventura Martins

## Contexto e problema

As fontes de dados (preço e leaks) podem mudar ao longo do tempo (ex.: trocar
CheapShark por ITAD; depender ou não da aprovação do Reddit). O bot não deve ser
reescrito a cada troca de fonte.

## Opções consideradas

- **Interfaces no consumidor** (inversão de dependência) com implementações
  plugáveis.
- Acoplar os handlers diretamente a cada client de API.

## Decisão

Definir interfaces no consumidor e injetar implementações via construtores:

- `PriceProvider` → implementação atual **CheapShark** (USD, PC); futura: ITAD
  (v2.0).
- `LeakSource` → **RSS como backbone** (sem auth, sem aprovação) e **Reddit como
  enhancement**, habilitado apenas se/quando o app for aprovado.

O RSS funciona de forma independente; o bot **nunca** pode depender só do Reddit.

## Consequências

### Positivas

- Troca de fonte sem reescrever os handlers; testes com fixtures via `httptest`.
- Degradação graciosa: se o Reddit não estiver disponível/aprovado, o RSS atende.

### Negativas

- Camada de abstração extra a manter.
- Necessidade de normalizar formatos heterogêneos (RSS vs Reddit) num modelo
  comum.
