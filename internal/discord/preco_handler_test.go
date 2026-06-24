package discord

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/JoaoVictorVM/leaks-n-promo/internal/cache"
	"github.com/JoaoVictorVM/leaks-n-promo/internal/price"
)

type fakeProvider struct {
	calls  int
	offers []price.Offer
	err    error
}

func (f *fakeProvider) Search(_ context.Context, _ string) ([]price.Offer, error) {
	f.calls++
	return f.offers, f.err
}

func testHandler(p price.PriceProvider) (*PrecoHandler, *cache.Cache[string, []price.Offer]) {
	c := cache.New[string, []price.Offer](time.Minute)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewPrecoHandler(p, c, logger), c
}

func TestLookupCacheHit(t *testing.T) {
	fp := &fakeProvider{err: errors.New("não deveria ser chamado")}
	h, c := testHandler(fp)
	c.Set("celeste", []price.Offer{{Title: "Celeste"}})

	offers, err := h.lookup(context.Background(), "Celeste") // normaliza para "celeste"
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(offers) != 1 || offers[0].Title != "Celeste" {
		t.Fatalf("ofertas inesperadas: %+v", offers)
	}
	if fp.calls != 0 {
		t.Errorf("provider não deveria ser chamado em cache hit (calls=%d)", fp.calls)
	}
}

func TestLookupCacheMissThenCaches(t *testing.T) {
	fp := &fakeProvider{offers: []price.Offer{{Title: "Hades"}}}
	h, _ := testHandler(fp)

	if _, err := h.lookup(context.Background(), "Hades"); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	// Segunda chamada deve vir do cache (provider chamado só uma vez).
	if _, err := h.lookup(context.Background(), "Hades"); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if fp.calls != 1 {
		t.Errorf("esperava 1 chamada ao provider, obtive %d", fp.calls)
	}
}

func TestLookupErrorNotCached(t *testing.T) {
	fp := &fakeProvider{err: errors.New("indisponível")}
	h, c := testHandler(fp)

	if _, err := h.lookup(context.Background(), "Celeste"); err == nil {
		t.Fatal("esperava erro")
	}
	if _, ok := c.Get("celeste"); ok {
		t.Error("não deveria armazenar no cache em caso de erro")
	}
}

func TestOptionString(t *testing.T) {
	opts := []*discordgo.ApplicationCommandInteractionDataOption{
		{Name: "jogo", Type: discordgo.ApplicationCommandOptionString, Value: "Celeste"},
	}
	if got := optionString(opts, "jogo"); got != "Celeste" {
		t.Errorf("optionString(jogo) = %q, esperava Celeste", got)
	}
	if got := optionString(opts, "ausente"); got != "" {
		t.Errorf("optionString(ausente) = %q, esperava vazio", got)
	}
}

func TestBuildPrecoEmbed(t *testing.T) {
	offers := []price.Offer{
		{Title: "Celeste", Store: "Steam", Price: 4.99, Retail: 19.99, URL: "https://x/redirect?dealID=1"},
		{Title: "Celeste", Store: "GOG", Price: 5.49, Retail: 5.49, URL: ""},
	}
	embed := buildPrecoEmbed("celeste", offers)

	if embed.Title != "Celeste" {
		t.Errorf("título = %q, esperava Celeste", embed.Title)
	}
	desc := embed.Description
	if !strings.Contains(desc, "**Steam** — $4.99") {
		t.Errorf("faltou a linha da Steam: %q", desc)
	}
	if !strings.Contains(desc, "~~$19.99~~") {
		t.Errorf("esperava retail riscado quando há desconto: %q", desc)
	}
	if !strings.Contains(desc, "[ver oferta](https://x/redirect?dealID=1)") {
		t.Errorf("esperava link da oferta: %q", desc)
	}
	// GOG sem desconto (retail == price) não deve ter strikethrough na sua linha.
	if strings.Contains(desc, "**GOG** — $5.49 ~~") {
		t.Errorf("não deveria riscar retail sem desconto: %q", desc)
	}
}

func TestBuildPrecoEmbedTruncates(t *testing.T) {
	offers := make([]price.Offer, 0, maxEmbedOffers+5)
	for range maxEmbedOffers + 5 {
		offers = append(offers, price.Offer{Title: "Jogo", Store: "Steam", Price: 1})
	}
	embed := buildPrecoEmbed("jogo", offers)
	if lines := strings.Count(embed.Description, "\n"); lines != maxEmbedOffers {
		t.Errorf("esperava %d linhas, obtive %d", maxEmbedOffers, lines)
	}
}

func TestBuildPrecoEmbedFallsBackToGameTitle(t *testing.T) {
	embed := buildPrecoEmbed("Termo Buscado", []price.Offer{{Store: "Steam", Price: 1}})
	if embed.Title != "Termo Buscado" {
		t.Errorf("título = %q, esperava o termo buscado", embed.Title)
	}
}
