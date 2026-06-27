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

Lista vazamentos e rumores recentes de games, agregados de fontes RSS e
(opcionalmente) do Reddit.

### Parâmetros

| Parâmetro | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `termo` | string | não | Filtra por um termo (ex.: nome de um jogo). Sem ele, retorna os mais recentes. |

### Uso

```
/leaks
/leaks termo: gta
```

### Resposta

Um embed com até 10 itens, cada um mostrando:

- **Título** com link para a publicação original.
- **Fonte** (nome do feed ou "Reddit").
- **Data** de publicação (quando disponível).

Os resultados são **deduplicados por URL** e ordenados do mais recente para o
mais antigo. O bot exibe "pensando…" enquanto busca.

### Fontes (modelo em camadas)

- **Backbone: RSS** — feeds de sites de notícias/leaks de games. Funciona sem
  autenticação e independe do Reddit. O filtro por termo é aplicado localmente.
- **Enhancement: Reddit** — busca no subreddit r/GamingLeaksAndRumours. Só é
  habilitado quando há credenciais (`REDDIT_CLIENT_ID`/`REDDIT_CLIENT_SECRET`) e
  o app está aprovado. Se indisponível, o `/leaks` continua respondendo só com o
  RSS.

### Casos de borda

| Situação | Resposta |
|---|---|
| Nenhum resultado (com termo) | "Nenhum vazamento encontrado para **<termo>**." |
| Nenhum resultado (sem termo) | "Nenhum vazamento recente encontrado." |
| Todas as fontes indisponíveis | Mensagem de fallback amigável, sem detalhe interno. |

### Observações

- **Resiliência:** as fontes são consultadas em paralelo; falha de uma fonte não
  derruba as demais. Só há erro quando **todas** falham.
- **Cache:** resultados recentes são memorizados em memória por um TTL curto
  (`CACHE_TTL`), inclusive a listagem sem termo.
- O bot apenas **agrega e direciona** à fonte original; não verifica a
  veracidade dos rumores.
