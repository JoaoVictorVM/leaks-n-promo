# Referência de Comandos

Comandos slash do **leaks&promo**. As respostas do bot são em PT-BR.

## `/preco`

Consulta os preços de um jogo de PC nas lojas digitais (via CheapShark).

### Parâmetros

| Parâmetro | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `jogo` | string | sim | Nome do jogo a consultar (ex.: `Celeste`). |

### Uso

```
/preco jogo: Celeste
```

### Resposta

Um embed com até 10 ofertas, cada uma mostrando:

- **Loja** e **preço atual** em **USD**.
- Preço cheio riscado quando há desconto (ex.: `$4.99 ~~$19.99~~`).
- Link **"ver oferta"** apontando para o redirect do CheapShark (exigência dos
  termos de uso da API).

O bot exibe "pensando…" enquanto consulta e então edita a mensagem com o
resultado.

### Casos de borda

| Situação | Resposta |
|---|---|
| Termo vazio | Mensagem efêmera pedindo o nome do jogo. |
| Nenhum resultado | "Nenhum resultado encontrado para **<jogo>**." |
| Fonte indisponível / rate limit (HTTP 429) | Mensagem de fallback amigável, sem detalhe interno. |

### Observações

- **Moeda:** preços em **USD** (o CheapShark opera em dólar). Preço em BRL/lojas
  brasileiras é um item de evolução futura (v2.0, via ITAD).
- **Escopo:** apenas **jogos de PC**.
- **Cache:** resultados recentes são memorizados em memória por um TTL curto
  (configurável em `CACHE_TTL`), reduzindo chamadas à API.
- **Resiliência:** cada consulta tem timeout por tentativa e retry com backoff.

## `/leaks`

> Em desenvolvimento (Fase 2). Será documentado quando disponível.
