BINARY := leaks-n-promo
CMD := ./cmd/bot
BIN_DIR := bin

GO ?= go

# Shell e comandos de remoção diferem entre Windows (cmd.exe) e Unix (sh).
# Fixar o shell por SO torna os recipes determinísticos, não importa de onde o
# make seja chamado (PowerShell, cmd ou git bash). Por isso o help usa só echo,
# sem depender de grep/sed/awk, que não existem no cmd.
ifeq ($(OS),Windows_NT)
SHELL := cmd.exe
.SHELLFLAGS := /c
RM_BIN = if exist "$(BIN_DIR)" rmdir /s /q "$(BIN_DIR)"
RM_COVER = if exist coverage.out del /q coverage.out
# Carrega o .env (KEY=VALUE, ignorando comentários) e roda o bot. No Windows via
# PowerShell; no Unix lendo o arquivo no shell atual (preserva valores com espaços).
DEV = powershell -NoProfile -ExecutionPolicy Bypass -Command "Get-Content .env | Where-Object { $$_ -match '^[^\#].*=' } | ForEach-Object { $$k,$$v = $$_ -split '=',2; [Environment]::SetEnvironmentVariable($$k.Trim(), $$v.Trim()) }; go run ./cmd/bot"
else
RM_BIN = rm -rf "$(BIN_DIR)"
RM_COVER = rm -f coverage.out
DEV = while IFS='=' read -r k v; do case "$$k" in ''|\#*) :;; *) export "$$k=$$v";; esac; done < .env; $(GO) run $(CMD)
endif

.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo Targets disponiveis:
	@echo   help        mostra esta ajuda
	@echo   build       compila o binario em bin/
	@echo   run         executa o bot localmente
	@echo   dev         carrega o .env e roda o bot
	@echo   test        roda os testes
	@echo   test-race   roda os testes com o detector de data race
	@echo   cover       gera e abre o relatorio de cobertura
	@echo   lint        roda o golangci-lint
	@echo   fmt         formata o codigo
	@echo   vet         roda o go vet
	@echo   tidy        organiza as dependencias do go.mod
	@echo   clean       remove artefatos de build e cobertura

.PHONY: build
build:
	$(GO) build -o $(BIN_DIR)/$(BINARY) $(CMD)

.PHONY: run
run:
	$(GO) run $(CMD)

.PHONY: dev
dev:
	@$(DEV)

.PHONY: test
test:
	$(GO) test ./...

.PHONY: test-race
test-race:
	$(GO) test -race ./...

.PHONY: cover
cover:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	golangci-lint fmt ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: clean
clean:
	$(RM_BIN)
	$(RM_COVER)
