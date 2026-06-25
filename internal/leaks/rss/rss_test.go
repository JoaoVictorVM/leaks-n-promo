package rss

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const testUserAgent = "leaks-n-promo-test/1.0"

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("lendo fixture %s: %v", name, err)
	}
	return data
}

func feedServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != testUserAgent {
			http.Error(w, "user-agent ausente", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write(fixture(t, "feed.xml"))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func errorServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestFetchParsesAndCombines(t *testing.T) {
	a, b := feedServer(t), feedServer(t)
	src := New(http.DefaultClient, testUserAgent, []string{a.URL, b.URL})

	items, err := src.Fetch(context.Background(), "")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	// 2 itens válidos por feed (o sem link é descartado) × 2 feeds.
	if len(items) != 4 {
		t.Fatalf("esperava 4 itens, obtive %d", len(items))
	}
	for _, it := range items {
		if it.Source != "Game Leaks Daily" {
			t.Errorf("source = %q, esperava o título do canal", it.Source)
		}
		if it.URL == "" {
			t.Error("item sem URL não deveria ter passado")
		}
	}
}

func TestFetchFilterByTerm(t *testing.T) {
	a := feedServer(t)
	src := New(http.DefaultClient, testUserAgent, []string{a.URL})

	items, err := src.Fetch(context.Background(), "gta")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("esperava 1 item para \"gta\", obtive %d", len(items))
	}
	if items[0].URL != "https://example.com/gta6-map" {
		t.Errorf("URL inesperada: %q", items[0].URL)
	}
}

func TestFetchTolaratesPartialFailure(t *testing.T) {
	good, bad := feedServer(t), errorServer(t)
	src := New(http.DefaultClient, testUserAgent, []string{good.URL, bad.URL})

	items, err := src.Fetch(context.Background(), "")
	if err != nil {
		t.Fatalf("falha parcial não deveria retornar erro: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("esperava 2 itens do feed bom, obtive %d", len(items))
	}
}

func TestFetchAllFail(t *testing.T) {
	a, b := errorServer(t), errorServer(t)
	src := New(http.DefaultClient, testUserAgent, []string{a.URL, b.URL})

	if _, err := src.Fetch(context.Background(), ""); err == nil {
		t.Fatal("esperava erro quando todos os feeds falham")
	}
}

func TestParsePubDate(t *testing.T) {
	cases := map[string]bool{
		"Mon, 02 Jan 2006 15:04:05 -0700": true,
		"2006-01-02T15:04:05Z":            true,
		"data inválida":                   false,
	}
	for raw, wantParsed := range cases {
		got := parsePubDate(raw)
		if wantParsed && got.IsZero() {
			t.Errorf("parsePubDate(%q) deveria ter parseado", raw)
		}
		if !wantParsed && !got.IsZero() {
			t.Errorf("parsePubDate(%q) deveria ser zero, obtive %v", raw, got)
		}
	}
}
