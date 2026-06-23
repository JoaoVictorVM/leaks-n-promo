// Command bot é o entrypoint do leaks&promo, um bot de Discord pull-only
// para consulta de preços de jogos de PC e agregação de vazamentos de games.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/JoaoVictorVM/leaks-n-promo/internal/config"
	"github.com/JoaoVictorVM/leaks-n-promo/internal/logging"
)

func main() {
	if err := run(); err != nil {
		// O logger pode não existir se a config falhar; reportamos em stderr.
		fmt.Fprintln(os.Stderr, "erro fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := logging.New(os.Stdout, cfg.LogLevel)

	// Cancela o contexto ao receber SIGINT/SIGTERM, disparando o encerramento
	// gracioso.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serve(ctx, logger, cfg)
	return nil
}

// serve mantém o bot rodando até o contexto ser cancelado (por sinal). A lógica
// fica separada de run para ser testável sem depender de sinais do SO.
func serve(ctx context.Context, logger *slog.Logger, cfg *config.Config) {
	logger.Info("bot iniciando",
		"guild_scoped", cfg.DiscordGuildID != "",
		"reddit_enabled", cfg.Reddit.Enabled(),
	)

	<-ctx.Done()

	logger.Info("sinal de encerramento recebido, finalizando")
}
