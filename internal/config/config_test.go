package config

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

func baseEnv() map[string]string {
	return map[string]string{
		"DISCORD_TOKEN":         "tok",
		"DISCORD_APP_ID":        "app",
		"DISCORD_GUILD_ID":      "",
		"LOG_LEVEL":             "info",
		"CACHE_TTL":             "5m",
		"HTTP_TIMEOUT":          "10s",
		"CHEAPSHARK_USER_AGENT": "ua/test",
		"REDDIT_CLIENT_ID":      "",
		"REDDIT_CLIENT_SECRET":  "",
	}
}

func applyEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for k, v := range kv {
		t.Setenv(k, v)
	}
}

func TestLoadValid(t *testing.T) {
	env := baseEnv()
	env["LOG_LEVEL"] = "debug"
	env["CACHE_TTL"] = "1m"
	env["HTTP_TIMEOUT"] = "3s"
	env["DISCORD_GUILD_ID"] = "guild123"
	env["REDDIT_CLIENT_ID"] = "id"
	env["REDDIT_CLIENT_SECRET"] = "secret"
	applyEnv(t, env)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("esperava sucesso, obtive erro: %v", err)
	}

	if cfg.DiscordToken != "tok" || cfg.DiscordAppID != "app" {
		t.Errorf("credenciais do Discord incorretas: %+v", cfg)
	}
	if cfg.DiscordGuildID != "guild123" {
		t.Errorf("guild id = %q, esperava %q", cfg.DiscordGuildID, "guild123")
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("log level = %v, esperava debug", cfg.LogLevel)
	}
	if cfg.CacheTTL != time.Minute {
		t.Errorf("cache ttl = %v, esperava 1m", cfg.CacheTTL)
	}
	if cfg.HTTPTimeout != 3*time.Second {
		t.Errorf("http timeout = %v, esperava 3s", cfg.HTTPTimeout)
	}
	if !cfg.Reddit.Enabled() {
		t.Errorf("Reddit deveria estar habilitado")
	}
}

func TestLoadDefaults(t *testing.T) {
	applyEnv(t, baseEnv())
	// Zera os opcionais para garantir que os defaults entram em ação.
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("CACHE_TTL", "")
	t.Setenv("HTTP_TIMEOUT", "")
	t.Setenv("CHEAPSHARK_USER_AGENT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("esperava sucesso, obtive erro: %v", err)
	}

	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("log level default = %v, esperava info", cfg.LogLevel)
	}
	if cfg.CacheTTL != 5*time.Minute {
		t.Errorf("cache ttl default = %v, esperava 5m", cfg.CacheTTL)
	}
	if cfg.HTTPTimeout != 10*time.Second {
		t.Errorf("http timeout default = %v, esperava 10s", cfg.HTTPTimeout)
	}
	if cfg.CheapSharkUserAgent != defaultUserAgent {
		t.Errorf("user-agent default = %q", cfg.CheapSharkUserAgent)
	}
	if cfg.Reddit.Enabled() {
		t.Errorf("Reddit não deveria estar habilitado sem credenciais")
	}
}

func TestLoadErrors(t *testing.T) {
	tests := []struct {
		name         string
		overrides    map[string]string
		wantContains string
	}{
		{
			name:         "token ausente",
			overrides:    map[string]string{"DISCORD_TOKEN": ""},
			wantContains: "DISCORD_TOKEN",
		},
		{
			name:         "app id ausente",
			overrides:    map[string]string{"DISCORD_APP_ID": ""},
			wantContains: "DISCORD_APP_ID",
		},
		{
			name:         "log level invalido",
			overrides:    map[string]string{"LOG_LEVEL": "loud"},
			wantContains: "LOG_LEVEL",
		},
		{
			name:         "duracao invalida",
			overrides:    map[string]string{"CACHE_TTL": "abc"},
			wantContains: "CACHE_TTL",
		},
		{
			name:         "duracao nao positiva",
			overrides:    map[string]string{"HTTP_TIMEOUT": "0s"},
			wantContains: "HTTP_TIMEOUT",
		},
		{
			name:         "reddit parcial",
			overrides:    map[string]string{"REDDIT_CLIENT_ID": "só-id"},
			wantContains: "REDDIT_CLIENT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := baseEnv()
			for k, v := range tt.overrides {
				env[k] = v
			}
			applyEnv(t, env)

			_, err := Load()
			if err == nil {
				t.Fatalf("esperava erro, mas Load teve sucesso")
			}
			if !strings.Contains(err.Error(), tt.wantContains) {
				t.Fatalf("erro %q não contém %q", err.Error(), tt.wantContains)
			}
		})
	}
}

func TestLoadAggregatesErrors(t *testing.T) {
	env := baseEnv()
	env["DISCORD_TOKEN"] = ""
	env["DISCORD_APP_ID"] = ""
	applyEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("esperava erro agregado")
	}
	msg := err.Error()
	if !strings.Contains(msg, "DISCORD_TOKEN") || !strings.Contains(msg, "DISCORD_APP_ID") {
		t.Fatalf("erro agregado deveria citar ambas as variáveis: %q", msg)
	}
}
