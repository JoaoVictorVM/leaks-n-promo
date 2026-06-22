# Diagramas de Sequência — leaks&promo

Fluxo de um comando no modelo pull-only: **Discord → handler → cache → fonte →
resposta**, com cache (TTL) e mensagem de fallback em caso de falha.

## `/preco <jogo>`

```mermaid
sequenceDiagram
    autonumber
    actor U as Usuário
    participant D as Discord Gateway
    participant H as Handler /preco
    participant C as Cache (TTL)
    participant P as PriceProvider (CheapShark)
    participant API as CheapShark API

    U->>D: /preco <jogo>
    D->>H: Interação
    H->>H: Valida input

    H->>C: get(jogo)
    alt cache hit
        C-->>H: ofertas (em cache)
    else cache miss
        C-->>H: vazio
        H->>P: BuscarPrecos(ctx, jogo)
        Note over P,API: context com timeout + retry/backoff
        P->>API: GET deals (User-Agent)
        alt sucesso
            API-->>P: ofertas
            P-->>H: ofertas
            H->>C: set(jogo, ofertas, TTL)
        else 429 / indisponível
            API-->>P: erro/429
            P-->>H: erro
            H-->>D: embed de fallback (sem detalhe interno)
            D-->>U: "Não consegui consultar agora, tente mais tarde"
        end
    end

    H-->>D: embed (loja, preço, link do CheapShark)
    D-->>U: resposta
```

## `/leaks [termo]`

```mermaid
sequenceDiagram
    autonumber
    actor U as Usuário
    participant D as Discord Gateway
    participant H as Handler /leaks
    participant C as Cache (TTL)
    participant R as RSS (backbone)
    participant RD as Reddit (opcional)

    U->>D: /leaks [termo]
    D->>H: Interação
    H->>C: get(termo)
    alt cache hit
        C-->>H: itens (em cache)
    else cache miss
        C-->>H: vazio
        par Consulta em paralelo
            H->>R: Buscar(ctx, termo)
            R-->>H: itens RSS
        and
            H->>RD: Buscar(ctx, termo)
            RD-->>H: itens Reddit (ou nada, se indisponível)
        end
        H->>H: merge + dedupe por URL + filtro por termo
        H->>C: set(termo, itens, TTL)
    end

    alt há resultados
        H-->>D: lista (título, fonte, link, data)
    else nenhuma fonte respondeu / sem resultado
        H-->>D: mensagem de fallback informativa
    end
    D-->>U: resposta
```

> O RSS é o backbone: se o Reddit estiver indisponível ou não aprovado, o
> `/leaks` continua respondendo apenas com os itens do RSS.
