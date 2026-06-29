// Package config carrega e valida a configuração do bot a partir do ambiente
// (12-factor), falhando no boot quando algo obrigatório está ausente ou inválido.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

// defaultUserAgent identifica a aplicação nas chamadas HTTP. O CheapShark exige
// um User-Agent; as demais fontes também o enviam.
const defaultUserAgent = "leaks-n-promo/0.1 (+https://github.com/JoaoVictorVM/leaks-n-promo)"

// Config reúne toda a configuração resolvida do bot.
type Config struct {
	DiscordToken   string
	DiscordAppID   string
	DiscordGuildID string // opcional: vazio = registro global de comandos
	LogLevel       slog.Level
	CacheTTL       time.Duration
	HTTPTimeout    time.Duration
	UserAgent      string
	RSSFeeds       []string // vazio = usar os feeds padrão da fonte RSS
	Reddit         RedditConfig
}

// RedditConfig guarda as credenciais opcionais do Reddit (fonte enhancement).
type RedditConfig struct {
	ClientID     string
	ClientSecret string
}

// Enabled indica se a fonte Reddit deve ser habilitada (ambas credenciais presentes).
func (r RedditConfig) Enabled() bool {
	return r.ClientID != "" && r.ClientSecret != ""
}

// Load lê a configuração do ambiente e a valida. Os problemas encontrados são
// reportados de uma vez (via errors.Join) para facilitar a correção.
func Load() (*Config, error) {
	var errs []error

	token, err := requireEnv("DISCORD_TOKEN")
	errs = appendErr(errs, err)

	appID, err := requireEnv("DISCORD_APP_ID")
	errs = appendErr(errs, err)

	level, err := parseLevel(lookupEnv("LOG_LEVEL", "info"))
	errs = appendErr(errs, err)

	cacheTTL, err := parseDuration("CACHE_TTL", "5m")
	errs = appendErr(errs, err)

	httpTimeout, err := parseDuration("HTTP_TIMEOUT", "10s")
	errs = appendErr(errs, err)

	reddit := RedditConfig{
		ClientID:     strings.TrimSpace(os.Getenv("REDDIT_CLIENT_ID")),
		ClientSecret: strings.TrimSpace(os.Getenv("REDDIT_CLIENT_SECRET")),
	}
	errs = appendErr(errs, validateReddit(reddit))

	if len(errs) > 0 {
		return nil, fmt.Errorf("configuração inválida: %w", errors.Join(errs...))
	}

	return &Config{
		DiscordToken:   token,
		DiscordAppID:   appID,
		DiscordGuildID: strings.TrimSpace(os.Getenv("DISCORD_GUILD_ID")),
		LogLevel:       level,
		CacheTTL:       cacheTTL,
		HTTPTimeout:    httpTimeout,
		UserAgent:      lookupEnv("HTTP_USER_AGENT", defaultUserAgent),
		RSSFeeds:       parseList(os.Getenv("LEAK_RSS_FEEDS")),
		Reddit:         reddit,
	}, nil
}

func appendErr(errs []error, err error) []error {
	if err != nil {
		return append(errs, err)
	}
	return errs
}

func requireEnv(key string) (string, error) {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("variável %s é obrigatória", key)
}

func lookupEnv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func parseDuration(key, def string) (time.Duration, error) {
	raw := lookupEnv(key, def)
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("variável %s inválida (%q): %w", key, raw, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("variável %s deve ser positiva (%q)", key, raw)
	}
	return d, nil
}

func parseLevel(raw string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("variável LOG_LEVEL inválida (%q): use debug, info, warn ou error", raw)
	}
}

// parseList separa uma lista CSV, descartando espaços e itens vazios. Retorna
// nil quando não há itens (o consumidor aplica o padrão).
func parseList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func validateReddit(r RedditConfig) error {
	if (r.ClientID == "") != (r.ClientSecret == "") {
		return errors.New("as variáveis REDDIT_CLIENT_ID e REDDIT_CLIENT_SECRET devem ser definidas juntas")
	}
	return nil
}
