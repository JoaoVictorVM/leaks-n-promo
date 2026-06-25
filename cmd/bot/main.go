// Command bot é o entrypoint do leaks&promo, um bot de Discord pull-only
// para consulta de preços de jogos de PC e agregação de vazamentos de games.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JoaoVictorVM/leaks-n-promo/internal/cache"
	"github.com/JoaoVictorVM/leaks-n-promo/internal/config"
	"github.com/JoaoVictorVM/leaks-n-promo/internal/discord"
	"github.com/JoaoVictorVM/leaks-n-promo/internal/leaks"
	"github.com/JoaoVictorVM/leaks-n-promo/internal/leaks/reddit"
	"github.com/JoaoVictorVM/leaks-n-promo/internal/leaks/rss"
	"github.com/JoaoVictorVM/leaks-n-promo/internal/logging"
	"github.com/JoaoVictorVM/leaks-n-promo/internal/price"
	"github.com/JoaoVictorVM/leaks-n-promo/internal/price/cheapshark"
)

const (
	priceAttempts  = 3
	priceBaseDelay = 200 * time.Millisecond
	priceMaxDelay  = 2 * time.Second
)

// Informações de build injetadas via ldflags (-X) pelo GoReleaser/Dockerfile.
// Os defaults valem para builds de desenvolvimento (go run/go build sem flags).
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
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

	bot, err := discord.New(cfg.DiscordToken, logger)
	if err != nil {
		return err
	}

	// Cadeia de preço: cliente CheapShark → resiliência (timeout/backoff). O
	// cache fica no handler. O timeout por tentativa vem da config.
	priceProvider := price.NewRetrying(
		cheapshark.New(&http.Client{}, cfg.CheapSharkUserAgent),
		price.RetryConfig{
			Attempts:  priceAttempts,
			BaseDelay: priceBaseDelay,
			MaxDelay:  priceMaxDelay,
			Timeout:   cfg.HTTPTimeout,
		},
	)
	priceCache := cache.New[string, []price.Offer](cfg.CacheTTL)
	bot.AddInteractionHandler(discord.NewPrecoHandler(priceProvider, priceCache, logger))

	// Fonte de leaks: RSS é o backbone; o Reddit só entra se houver credenciais.
	httpClient := &http.Client{}
	leakSources := []leaks.LeakSource{rss.New(httpClient, cfg.CheapSharkUserAgent, rss.DefaultFeeds)}
	if cfg.Reddit.Enabled() {
		leakSources = append(leakSources, reddit.New(httpClient, cfg.CheapSharkUserAgent, cfg.Reddit.ClientID, cfg.Reddit.ClientSecret))
	}
	leakCache := cache.New[string, []leaks.Leak](cfg.CacheTTL)
	bot.AddInteractionHandler(discord.NewLeaksHandler(leaks.NewAggregator(leakSources...), leakCache, logger))

	// Cancela o contexto ao receber SIGINT/SIGTERM, disparando o encerramento
	// gracioso.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("bot iniciando",
		"version", version,
		"commit", commit,
		"date", date,
		"guild_scoped", cfg.DiscordGuildID != "",
		"reddit_enabled", cfg.Reddit.Enabled(),
	)

	if err := bot.Open(); err != nil {
		return err
	}
	defer func() {
		if cerr := bot.Close(); cerr != nil {
			logger.Error("falha ao fechar a sessão do discord", "erro", cerr)
		}
	}()

	if err := bot.RegisterCommands(cfg.DiscordAppID, cfg.DiscordGuildID); err != nil {
		return err
	}

	serve(ctx, logger)
	return nil
}

// serve bloqueia até o contexto ser cancelado (por sinal). Fica separada de run
// para ser testável sem depender de sinais do SO nem de conexão real.
func serve(ctx context.Context, logger *slog.Logger) {
	<-ctx.Done()
	logger.Info("sinal de encerramento recebido, finalizando")
}
