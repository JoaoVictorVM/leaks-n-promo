package reddit

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

const (
	testUserAgent = "leaks-n-promo-test/1.0"
	testClientID  = "id"
	testSecret    = "secret"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("lendo fixture %s: %v", name, err)
	}
	return data
}

type recorder struct {
	tokenHits  int
	newHits    int
	searchHits int
	lastQuery  string
}

func newServer(t *testing.T, rec *recorder) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/access_token", func(w http.ResponseWriter, r *http.Request) {
		rec.tokenHits++
		if u, p, ok := r.BasicAuth(); !ok || u != testClientID || p != testSecret {
			http.Error(w, "basic auth inválido", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"tok123","expires_in":3600}`))
	})

	mux.HandleFunc("/r/GamingLeaksAndRumours/new", func(w http.ResponseWriter, r *http.Request) {
		rec.newHits++
		if r.Header.Get("Authorization") != "Bearer tok123" {
			http.Error(w, "bearer ausente", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write(fixture(t, "listing.json"))
	})

	mux.HandleFunc("/r/GamingLeaksAndRumours/search", func(w http.ResponseWriter, r *http.Request) {
		rec.searchHits++
		rec.lastQuery = r.URL.Query().Get("q")
		_, _ = w.Write(fixture(t, "listing.json"))
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newSource(srv *httptest.Server) *Source {
	return New(srv.Client(), testUserAgent, testClientID, testSecret,
		WithTokenURL(srv.URL+"/api/v1/access_token"),
		WithAPIBaseURL(srv.URL),
	)
}

func TestFetchNewListing(t *testing.T) {
	rec := &recorder{}
	src := newSource(newServer(t, rec))

	items, err := src.Fetch(context.Background(), "")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("esperava 2 itens, obtive %d", len(items))
	}
	if rec.newHits != 1 || rec.searchHits != 0 {
		t.Errorf("termo vazio deveria usar /new (new=%d search=%d)", rec.newHits, rec.searchHits)
	}

	first := items[0]
	if first.Source != "Reddit" {
		t.Errorf("source = %q, esperava Reddit", first.Source)
	}
	if first.URL != "https://www.reddit.com/r/GamingLeaksAndRumours/comments/abc/gta6/" {
		t.Errorf("URL inesperada: %q", first.URL)
	}
	if first.Published.IsZero() {
		t.Error("esperava Published preenchido a partir de created_utc")
	}
}

func TestFetchSearch(t *testing.T) {
	rec := &recorder{}
	src := newSource(newServer(t, rec))

	if _, err := src.Fetch(context.Background(), "gta"); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if rec.searchHits != 1 || rec.newHits != 0 {
		t.Errorf("termo deveria usar /search (new=%d search=%d)", rec.newHits, rec.searchHits)
	}
	if rec.lastQuery != "gta" {
		t.Errorf("q = %q, esperava gta", rec.lastQuery)
	}
}

func TestTokenCached(t *testing.T) {
	rec := &recorder{}
	src := newSource(newServer(t, rec))

	for range 2 {
		if _, err := src.Fetch(context.Background(), ""); err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
	}
	if rec.tokenHits != 1 {
		t.Errorf("token deveria ser buscado 1 vez (cache), obtive %d", rec.tokenHits)
	}
}

func TestTokenRateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)
	src := New(srv.Client(), testUserAgent, testClientID, testSecret,
		WithTokenURL(srv.URL), WithAPIBaseURL(srv.URL))

	if _, err := src.Fetch(context.Background(), ""); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("esperava ErrRateLimited, obtive %v", err)
	}
}

func TestListingRateLimited(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/access_token", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"tok123","expires_in":3600}`))
	})
	mux.HandleFunc("/r/GamingLeaksAndRumours/new", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	src := New(srv.Client(), testUserAgent, testClientID, testSecret,
		WithTokenURL(srv.URL+"/api/v1/access_token"), WithAPIBaseURL(srv.URL))

	if _, err := src.Fetch(context.Background(), ""); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("esperava ErrRateLimited, obtive %v", err)
	}
}

func TestListingURL(t *testing.T) {
	s := New(nil, testUserAgent, testClientID, testSecret, WithAPIBaseURL("https://api"))

	if got := s.listingURL(""); !strings.Contains(got, "/r/GamingLeaksAndRumours/new") {
		t.Errorf("termo vazio = %q, esperava /new", got)
	}
	got := s.listingURL("gta 6")
	if !strings.Contains(got, "/search?") || !strings.Contains(got, "q=gta+6") {
		t.Errorf("com termo = %q, esperava /search com q codificado", got)
	}
}
