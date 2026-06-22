# syntax=docker/dockerfile:1

# Estágio de build: compila um binário estático (sem CGO).
FROM golang:1.26-alpine AS build
WORKDIR /src

# Baixa as dependências primeiro para aproveitar o cache de camadas. O padrão
# go.sum* casa tanto quando o arquivo ainda não existe (sem dependências) quanto
# depois que ele passa a existir.
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o /out/leaks-n-promo ./cmd/bot

# Imagem final mínima: distroless static, com CA certs e usuário non-root.
# Sem HEALTHCHECK por ora — o bot é um worker de gateway (sem porta HTTP), então
# não há endpoint para sondar; será adicionado junto do health endpoint (Fase 3).
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/leaks-n-promo /usr/local/bin/leaks-n-promo
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/leaks-n-promo"]
