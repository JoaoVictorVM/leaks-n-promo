package cheapshark

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

// newTestServer monta um servidor que serve as fixtures e conta os hits por
// rota, validando o cabeçalho User-Agent.
func newTestServer(t *testing.T, hits map[string]int) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/stores", func(w http.ResponseWriter, r *http.Request) {
		hits["stores"]++
		if r.Header.Get("User-Agent") != testUserAgent {
			http.Error(w, "user-agent ausente", http.StatusBadRequest)
			return
		}
		_, _ = w.Write(fixture(t, "stores.json"))
	})
	mux.HandleFunc("/deals", func(w http.ResponseWriter, r *http.Request) {
		hits["deals"]++
		if r.Header.Get("User-Agent") != testUserAgent {
			http.Error(w, "user-agent ausente", http.StatusBadRequest)
			return
		}
		_, _ = w.Write(fixture(t, "deals.json"))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestSearchSuccess(t *testing.T) {
	hits := map[string]int{}
	srv := newTestServer(t, hits)
	c := New(srv.Client(), testUserAgent, WithBaseURL(srv.URL))

	offers, err := c.Search(context.Background(), "celeste")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(offers) != 3 {
		t.Fatalf("esperava 3 ofertas, obtive %d", len(offers))
	}

	first := offers[0]
	if first.Store != "Steam" {
		t.Errorf("loja[0] = %q, esperava Steam", first.Store)
	}
	if first.Price != 4.99 || first.Retail != 19.99 {
		t.Errorf("preços[0] = %v/%v, esperava 4.99/19.99", first.Price, first.Retail)
	}
	if !strings.HasPrefix(first.URL, "https://www.cheapshark.com/redirect") || !strings.Contains(first.URL, "dealID=abc123") {
		t.Errorf("URL[0] inesperada: %q", first.URL)
	}

	// storeID desconhecido cai no fallback.
	if offers[2].Store != "Loja 99" {
		t.Errorf("loja[2] = %q, esperava \"Loja 99\"", offers[2].Store)
	}
}

func TestStoresCached(t *testing.T) {
	hits := map[string]int{}
	srv := newTestServer(t, hits)
	c := New(srv.Client(), testUserAgent, WithBaseURL(srv.URL))

	for range 2 {
		if _, err := c.Search(context.Background(), "celeste"); err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
	}

	if hits["stores"] != 1 {
		t.Errorf("/stores chamado %d vezes, esperava 1 (cache)", hits["stores"])
	}
	if hits["deals"] != 2 {
		t.Errorf("/deals chamado %d vezes, esperava 2", hits["deals"])
	}
}

func TestSearchEmptyGameSkipsNetwork(t *testing.T) {
	hits := map[string]int{}
	srv := newTestServer(t, hits)
	c := New(srv.Client(), testUserAgent, WithBaseURL(srv.URL))

	offers, err := c.Search(context.Background(), "   ")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(offers) != 0 {
		t.Errorf("esperava nenhuma oferta, obtive %d", len(offers))
	}
	if hits["stores"] != 0 || hits["deals"] != 0 {
		t.Errorf("não deveria chamar a rede para termo vazio: %v", hits)
	}
}

func TestSearchRateLimited(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/stores", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(fixture(t, "stores.json"))
	})
	mux.HandleFunc("/deals", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := New(srv.Client(), testUserAgent, WithBaseURL(srv.URL))

	_, err := c.Search(context.Background(), "celeste")
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("esperava ErrRateLimited, obtive %v", err)
	}
}

func TestRedirectURLPreservesEncodedDealID(t *testing.T) {
	// O dealID já vem percent-encoded; o link não pode re-encodá-lo.
	dealID := "dPa6mWDMGSef4%2FAdNuQBsxOHDUbNZ3ttlethrFug6DQ%3D"
	got := redirectURL(dealID)

	want := "https://www.cheapshark.com/redirect?dealID=" + dealID
	if got != want {
		t.Errorf("redirectURL = %q, esperava %q", got, want)
	}
	if strings.Contains(got, "%25") {
		t.Errorf("dealID foi re-encodado (contém %%25): %q", got)
	}
}

func TestSearchUnexpectedStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/stores", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := New(srv.Client(), testUserAgent, WithBaseURL(srv.URL))

	if _, err := c.Search(context.Background(), "celeste"); err == nil {
		t.Fatal("esperava erro para status 500")
	}
}
