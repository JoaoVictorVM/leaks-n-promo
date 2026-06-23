package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestNewRespectsLevel(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, slog.LevelInfo)

	log.Debug("abaixo do nível")
	log.Info("oi", "chave", "valor")

	out := buf.String()
	if strings.Contains(out, "abaixo do nível") {
		t.Fatalf("mensagem de debug não deveria aparecer no nível info: %q", out)
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &entry); err != nil {
		t.Fatalf("saída não é JSON válido: %v (%q)", err, out)
	}
	if entry["msg"] != "oi" {
		t.Errorf("msg = %v, esperava \"oi\"", entry["msg"])
	}
	if entry["level"] != "INFO" {
		t.Errorf("level = %v, esperava INFO", entry["level"])
	}
	if entry["chave"] != "valor" {
		t.Errorf("chave = %v, esperava \"valor\"", entry["chave"])
	}
}
