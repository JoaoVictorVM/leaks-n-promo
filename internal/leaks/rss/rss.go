// Package rss implementa leaks.LeakSource consumindo feeds RSS 2.0 de sites de
// notícias/leaks de games. É o backbone da fonte de leaks: funciona sem
// autenticação e independe do Reddit.
package rss

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/JoaoVictorVM/leaks-n-promo/internal/leaks"
)

// DefaultFeeds são feeds de exemplo; a lista é configurável via construtor.
var DefaultFeeds = []string{
	"https://www.videogameschronicle.com/feed/",
	"https://insider-gaming.com/feed/",
}

// maxFeedBytes limita o tamanho lido de cada feed, evitando respostas abusivas.
const maxFeedBytes = 5 << 20 // 5 MiB

// Source busca leaks a partir de uma lista de feeds RSS.
type Source struct {
	httpClient *http.Client
	userAgent  string
	feeds      []string
}

// New cria a fonte RSS. Se httpClient for nil, usa http.DefaultClient.
func New(httpClient *http.Client, userAgent string, feeds []string) *Source {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Source{httpClient: httpClient, userAgent: userAgent, feeds: feeds}
}

// Garante em tempo de compilação que Source satisfaz a interface.
var _ leaks.LeakSource = (*Source)(nil)

type rssDocument struct {
	Channel struct {
		Title string    `xml:"title"`
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
}

// Fetch busca em todos os feeds em paralelo e combina os resultados, filtrando
// por termo localmente. Falhas parciais são toleradas; só retorna erro se todos
// os feeds falharem.
func (s *Source) Fetch(ctx context.Context, term string) ([]leaks.Leak, error) {
	type result struct {
		items []leaks.Leak
		err   error
	}

	results := make(chan result, len(s.feeds))
	var wg sync.WaitGroup
	for _, feed := range s.feeds {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			items, err := s.fetchFeed(ctx, url)
			results <- result{items: items, err: err}
		}(feed)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	var all []leaks.Leak
	var errs []error
	for r := range results {
		if r.err != nil {
			errs = append(errs, r.err)
			continue
		}
		all = append(all, r.items...)
	}

	if len(all) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("todas as fontes RSS falharam: %w", errors.Join(errs...))
	}

	return filterByTerm(all, term), nil
}

func (s *Source) fetchFeed(ctx context.Context, url string) ([]leaks.Leak, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("criando requisição para %s: %w", url, err)
	}
	req.Header.Set("User-Agent", s.userAgent)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("buscando feed %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed %s respondeu status %d", url, resp.StatusCode)
	}

	var doc rssDocument
	if err := xml.NewDecoder(io.LimitReader(resp.Body, maxFeedBytes)).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decodificando feed %s: %w", url, err)
	}

	source := strings.TrimSpace(doc.Channel.Title)
	if source == "" {
		source = "RSS"
	}

	items := make([]leaks.Leak, 0, len(doc.Channel.Items))
	for _, it := range doc.Channel.Items {
		link := strings.TrimSpace(it.Link)
		if link == "" {
			continue // sem URL não há como deduplicar nem direcionar o usuário
		}
		items = append(items, leaks.Leak{
			Title:     strings.TrimSpace(it.Title),
			Source:    source,
			URL:       link,
			Published: parsePubDate(it.PubDate),
		})
	}
	return items, nil
}

var pubDateLayouts = []string{
	time.RFC1123Z,
	time.RFC1123,
	time.RFC822Z,
	time.RFC822,
	"Mon, 2 Jan 2006 15:04:05 -0700",
	time.RFC3339,
}

func parsePubDate(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	for _, layout := range pubDateLayouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t
		}
	}
	return time.Time{}
}

func filterByTerm(items []leaks.Leak, term string) []leaks.Leak {
	term = strings.TrimSpace(strings.ToLower(term))
	if term == "" {
		return items
	}
	out := make([]leaks.Leak, 0, len(items))
	for _, it := range items {
		if strings.Contains(strings.ToLower(it.Title), term) {
			out = append(out, it)
		}
	}
	return out
}
