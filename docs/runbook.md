# Runbook — leaks&promo

Guia operacional para rodar, atualizar e diagnosticar o bot em um host
self-hosted (máquina própria ou Raspberry Pi). Régua de portfólio:
**containerizado + deploy documentado** (não uptime 24/7).

## Pré-requisitos

- Docker instalado no host.
- Um **bot do Discord** criado em https://discord.com/developers/applications,
  com o token em mãos e os intents mínimos habilitados.

## Configuração

Toda a config é via variáveis de ambiente (12-factor). Veja
[`.env.example`](../.env.example) para a lista completa.

```sh
cp .env.example .env
# edite .env e preencha ao menos DISCORD_TOKEN e DISCORD_APP_ID
```

> O `.env` é **gitignored** e nunca deve ser commitado.

## Executando

### Opção A — docker compose (dev local)

```sh
docker compose up --build -d   # sobe em background
docker compose logs -f         # acompanha os logs
docker compose down            # encerra
```

### Opção B — imagem publicada no GHCR

```sh
docker run --rm --name leaks-n-promo \
  --env-file ./.env \
  ghcr.io/joaovictorvm/leaks-n-promo:latest
```

## Execução como serviço (systemd)

Para manter o bot rodando e reiniciando automaticamente, use uma unit que
gerencia o container Docker.

1. Coloque o `.env` em `/etc/leaks-n-promo/.env` (permissão `600`).
2. Crie `/etc/systemd/system/leaks-n-promo.service`:

```ini
[Unit]
Description=leaks&promo Discord bot
After=docker.service network-online.target
Requires=docker.service
Wants=network-online.target

[Service]
Restart=always
RestartSec=10
# Garante estado limpo e imagem atualizada a cada (re)start.
ExecStartPre=-/usr/bin/docker rm -f leaks-n-promo
ExecStartPre=/usr/bin/docker pull ghcr.io/joaovictorvm/leaks-n-promo:latest
ExecStart=/usr/bin/docker run --name leaks-n-promo \
  --env-file /etc/leaks-n-promo/.env \
  ghcr.io/joaovictorvm/leaks-n-promo:latest
ExecStop=/usr/bin/docker stop leaks-n-promo

[Install]
WantedBy=multi-user.target
```

3. Habilite e inicie:

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now leaks-n-promo
```

## Operações comuns

| Tarefa | Comando |
|---|---|
| Ver status | `sudo systemctl status leaks-n-promo` |
| Ver logs | `journalctl -u leaks-n-promo -f` |
| Reiniciar | `sudo systemctl restart leaks-n-promo` |
| Parar | `sudo systemctl stop leaks-n-promo` |
| Atualizar | `sudo systemctl restart leaks-n-promo` (faz `docker pull` no `ExecStartPre`) |

### Atualizar para uma versão fixa (em vez de `latest`)

Troque a tag da imagem na unit (ex.: `:1.2.0`) e rode
`sudo systemctl daemon-reload && sudo systemctl restart leaks-n-promo`.

### Rollback

Aponte a tag para a versão anterior conhecida e reinicie. As imagens ficam
disponíveis em **Packages** do repositório no GHCR.

## Diagnóstico

| Sintoma | Possível causa | Ação |
|---|---|---|
| Container reinicia em loop | `DISCORD_TOKEN` ausente/ inválido | Confira o `.env`; veja os logs. |
| Bot conecta mas comandos não aparecem | Comandos não registrados / propagação global lenta | Use `DISCORD_GUILD_ID` em dev (propaga na hora). |
| `/preco` retornando fallback | CheapShark indisponível ou 429 | Transitório; aguarde. O cache e o backoff reduzem o impacto. |
| `/leaks` sem itens do Reddit | App do Reddit não aprovado/indisponível | Esperado — o RSS é o backbone e segue respondendo. |

> Não há `HEALTHCHECK` na imagem: o bot é um worker de gateway (sem porta HTTP).
> A saúde é observada pelos logs e pelo estado do serviço no systemd.
