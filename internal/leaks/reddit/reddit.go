// Package reddit implementa leaks.LeakSource consultando o subreddit
// r/GamingLeaksAndRumours via API OAuth do Reddit. É uma fonte de enhancement:
// só é habilitada quando há credenciais; o RSS continua sendo o backbone.
package reddit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/JoaoVictorVM/leaks-n-promo/internal/leaks"
)

const (
	defaultTokenURL   = "https://www.reddit.com/api/v1/access_token" //nolint:gosec // URL pública do endpoint de token, não é credencial
	defaultAPIBaseURL = "https://oauth.reddit.com"
	defaultSubreddit  = "GamingLeaksAndRumours"
	listingLimit      = "25"
	// tokenExpiryBuffer renova o token um pouco antes de expirar, evitando usar
	// um token no limite da validade.
	tokenExpiryBuffer = time.Minute
)

// ErrRateLimited indica que o Reddit respondeu HTTP 429 (rate limit).
var ErrRateLimited = errors.New("reddit: limite de requisições atingido (429)")

// Source busca leaks no Reddit e implementa leaks.LeakSource.
type Source struct {
	httpClient   *http.Client
	userAgent    string
	clientID     string
	clientSecret string
	subreddit    string
	tokenURL     string
	apiBaseURL   string

	mu       sync.Mutex
	token    string
	tokenExp time.Time
	now      func() time.Time // injetável para testes
}

// Option configura o Source na construção.
type Option func(*Source)

// WithTokenURL sobrescreve a URL do endpoint de token (útil em testes).
func WithTokenURL(u string) Option { return func(s *Source) { s.tokenURL = u } }

// WithAPIBaseURL sobrescreve a URL base da API (útil em testes).
func WithAPIBaseURL(u string) Option { return func(s *Source) { s.apiBaseURL = u } }

// WithSubreddit sobrescreve o subreddit consultado.
func WithSubreddit(name string) Option { return func(s *Source) { s.subreddit = name } }

// New cria a fonte Reddit. Se httpClient for nil, usa http.DefaultClient.
func New(httpClient *http.Client, userAgent, clientID, clientSecret string, opts ...Option) *Source {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	s := &Source{
		httpClient:   httpClient,
		userAgent:    userAgent,
		clientID:     clientID,
		clientSecret: clientSecret,
		subreddit:    defaultSubreddit,
		tokenURL:     defaultTokenURL,
		apiBaseURL:   defaultAPIBaseURL,
		now:          time.Now,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Garante em tempo de compilação que Source satisfaz a interface.
var _ leaks.LeakSource = (*Source)(nil)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type listingResponse struct {
	Data struct {
		Children []struct {
			Data struct {
				Title      string  `json:"title"`
				Permalink  string  `json:"permalink"`
				CreatedUTC float64 `json:"created_utc"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

// Fetch busca posts recentes (termo vazio) ou resultados da busca no subreddit.
func (s *Source) Fetch(ctx context.Context, term string) ([]leaks.Leak, error) {
	token, err := s.accessToken(ctx)
	if err != nil {
		return nil, err
	}

	var listing listingResponse
	if err := s.getJSON(ctx, s.listingURL(term), token, &listing); err != nil {
		return nil, err
	}

	out := make([]leaks.Leak, 0, len(listing.Data.Children))
	for _, child := range listing.Data.Children {
		d := child.Data
		out = append(out, leaks.Leak{
			Title:     strings.TrimSpace(d.Title),
			Source:    "Reddit",
			URL:       "https://www.reddit.com" + d.Permalink,
			Published: time.Unix(int64(d.CreatedUTC), 0).UTC(),
		})
	}
	return out, nil
}

func (s *Source) listingURL(term string) string {
	if strings.TrimSpace(term) == "" {
		return fmt.Sprintf("%s/r/%s/new?limit=%s", s.apiBaseURL, s.subreddit, listingLimit)
	}
	q := url.Values{
		"q":           {term},
		"restrict_sr": {"true"},
		"sort":        {"new"},
		"limit":       {listingLimit},
	}
	return fmt.Sprintf("%s/r/%s/search?%s", s.apiBaseURL, s.subreddit, q.Encode())
}

// accessToken retorna um token válido, renovando-o via client_credentials quando
// expirado. O cache é protegido por mutex.
func (s *Source) accessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token != "" && s.now().Before(s.tokenExp) {
		return s.token, nil
	}

	form := url.Values{"grant_type": {"client_credentials"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("criando requisição de token: %w", err)
	}
	req.SetBasicAuth(s.clientID, s.clientSecret)
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("obtendo token do reddit: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := statusError(resp.StatusCode, "token"); err != nil {
		return "", err
	}

	var tok tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return "", fmt.Errorf("decodificando token do reddit: %w", err)
	}

	s.token = tok.AccessToken
	s.tokenExp = s.now().Add(time.Duration(tok.ExpiresIn)*time.Second - tokenExpiryBuffer)
	return s.token, nil
}

func (s *Source) getJSON(ctx context.Context, endpoint, token string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("criando requisição: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", s.userAgent)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("consultando reddit: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := statusError(resp.StatusCode, "listagem"); err != nil {
		return err
	}

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decodificando resposta do reddit: %w", err)
	}
	return nil
}

func statusError(code int, what string) error {
	switch code {
	case http.StatusOK:
		return nil
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		return fmt.Errorf("reddit respondeu status %d em %s", code, what)
	}
}
