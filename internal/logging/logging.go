// Package logging configura o logger estruturado (slog) da aplicação.
package logging

import (
	"io"
	"log/slog"
)

// New cria um logger estruturado em JSON, escrevendo em w no nível informado.
func New(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level}))
}
