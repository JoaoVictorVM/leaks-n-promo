package main

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/JoaoVictorVM/leaks-n-promo/internal/config"
	"github.com/JoaoVictorVM/leaks-n-promo/internal/logging"
)

func TestServeShutsDownOnContextCancel(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.New(&buf, slog.LevelInfo)
	cfg := &config.Config{DiscordToken: "tok", DiscordAppID: "app"}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() { serve(ctx, logger, cfg); close(done) }()

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("serve não encerrou após o cancelamento do contexto")
	}

	// Leitura segura: só ocorre após receber de done (happens-before).
	out := buf.String()
	if !strings.Contains(out, "bot iniciando") {
		t.Errorf("esperava log de início; saída: %q", out)
	}
	if !strings.Contains(out, "encerramento recebido") {
		t.Errorf("esperava log de encerramento; saída: %q", out)
	}
}
