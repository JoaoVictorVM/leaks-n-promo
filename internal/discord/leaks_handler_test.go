package discord

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/JoaoVictorVM/leaks-n-promo/internal/cache"
	"github.com/JoaoVictorVM/leaks-n-promo/internal/leaks"
)

type fakeLeakSource struct {
	calls int
	items []leaks.Leak
	err   error
}

func (f *fakeLeakSource) Fetch(_ context.Context, _ string) ([]leaks.Leak, error) {
	f.calls++
	return f.items, f.err
}

func testLeaksHandler(src leaks.LeakSource) (*LeaksHandler, *cache.Cache[string, []leaks.Leak]) {
	c := cache.New[string, []leaks.Leak](time.Minute)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewLeaksHandler(src, c, logger), c
}

func TestLeaksLookupCacheHit(t *testing.T) {
	fs := &fakeLeakSource{err: errors.New("não deveria ser chamado")}
	h, c := testLeaksHandler(fs)
	c.Set("gta", []leaks.Leak{{Title: "GTA leak"}})

	items, err := h.lookup(context.Background(), "GTA")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(items) != 1 || fs.calls != 0 {
		t.Fatalf("esperava cache hit sem chamar a fonte: items=%d calls=%d", len(items), fs.calls)
	}
}

func TestLeaksLookupCacheMissThenCaches(t *testing.T) {
	fs := &fakeLeakSource{items: []leaks.Leak{{Title: "Recentes"}}}
	h, _ := testLeaksHandler(fs)

	// Termo vazio (mais recentes) também deve ser memorizado.
	for range 2 {
		if _, err := h.lookup(context.Background(), ""); err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
	}
	if fs.calls != 1 {
		t.Errorf("esperava 1 chamada à fonte (cache), obtive %d", fs.calls)
	}
}

func TestLeaksLookupErrorNotCached(t *testing.T) {
	fs := &fakeLeakSource{err: errors.New("indisponível")}
	h, c := testLeaksHandler(fs)

	if _, err := h.lookup(context.Background(), "gta"); err == nil {
		t.Fatal("esperava erro")
	}
	if _, ok := c.Get("gta"); ok {
		t.Error("não deveria cachear em caso de erro")
	}
}

func TestNoLeaksMessage(t *testing.T) {
	if msg := noLeaksMessage(""); !strings.Contains(msg, "recente") {
		t.Errorf("mensagem sem termo inesperada: %q", msg)
	}
	if msg := noLeaksMessage("gta"); !strings.Contains(msg, "gta") {
		t.Errorf("mensagem com termo deveria citar o termo: %q", msg)
	}
}

func TestBuildLeaksEmbed(t *testing.T) {
	published := time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)
	items := []leaks.Leak{
		{Title: "GTA 6 leak", Source: "VGC", URL: "https://x/gta", Published: published},
		{Title: "Sem data", Source: "Reddit", URL: "https://x/y"},
	}

	embed := buildLeaksEmbed("gta", items)
	if embed.Title != "Vazamentos: gta" {
		t.Errorf("título = %q, esperava \"Vazamentos: gta\"", embed.Title)
	}
	desc := embed.Description
	if !strings.Contains(desc, "[GTA 6 leak](https://x/gta)") {
		t.Errorf("faltou o link mascarado: %q", desc)
	}
	if !strings.Contains(desc, "*VGC*") {
		t.Errorf("faltou a fonte em itálico: %q", desc)
	}
	if !strings.Contains(desc, "02/01/2026") {
		t.Errorf("faltou a data formatada: %q", desc)
	}
	// Item sem data não deve produzir " · " na sua linha.
	if strings.Contains(desc, "*Reddit* · ") {
		t.Errorf("item sem data não deveria mostrar data: %q", desc)
	}
}

func TestBuildLeaksEmbedDefaultTitleAndTruncation(t *testing.T) {
	if embed := buildLeaksEmbed("", []leaks.Leak{{Title: "x", URL: "u"}}); embed.Title != "Vazamentos recentes" {
		t.Errorf("título sem termo = %q, esperava \"Vazamentos recentes\"", embed.Title)
	}

	items := make([]leaks.Leak, 0, maxEmbedLeaks+5)
	for range maxEmbedLeaks + 5 {
		items = append(items, leaks.Leak{Title: "x", URL: "u"})
	}
	embed := buildLeaksEmbed("", items)
	if lines := strings.Count(embed.Description, "\n"); lines != maxEmbedLeaks {
		t.Errorf("esperava %d linhas, obtive %d", maxEmbedLeaks, lines)
	}
}
