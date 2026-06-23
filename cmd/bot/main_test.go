package main

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/JoaoVictorVM/leaks-n-promo/internal/logging"
)

func TestServeShutsDownOnContextCancel(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.New(&buf, slog.LevelInfo)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() { serve(ctx, logger); close(done) }()

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("serve não encerrou após o cancelamento do contexto")
	}

	// Leitura segura: só ocorre após receber de done (happens-before).
	if out := buf.String(); !strings.Contains(out, "encerramento recebido") {
		t.Errorf("esperava log de encerramento; saída: %q", out)
	}
}
