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
else
RM_BIN = rm -rf "$(BIN_DIR)"
RM_COVER = rm -f coverage.out
endif

.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo Targets disponiveis:
	@echo   help        mostra esta ajuda
	@echo   build       compila o binario em bin/
	@echo   run         executa o bot localmente
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
