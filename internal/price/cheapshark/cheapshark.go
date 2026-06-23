// Package cheapshark implementa price.PriceProvider consultando a API pública do
// CheapShark (preços de jogos de PC, em USD).
package cheapshark

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/JoaoVictorVM/leaks-n-promo/internal/price"
)

const (
	defaultBaseURL = "https://www.cheapshark.com/api/1.0"
	// redirectBaseURL gera o link da oferta. Os termos do CheapShark exigem
	// direcionar o usuário pelos links deles.
	redirectBaseURL = "https://www.cheapshark.com/redirect"
	// dealsLimit limita o número de ofertas retornadas por busca.
	dealsLimit = "12"
)

// ErrRateLimited indica que o CheapShark respondeu HTTP 429 (rate limit).
var ErrRateLimited = errors.New("cheapshark: limite de requisições atingido (429)")

// Garante em tempo de compilação que o Client satisfaz a interface.
var _ price.PriceProvider = (*Client)(nil)

// Client consulta o CheapShark e implementa price.PriceProvider.
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string

	storesMu sync.Mutex
	stores   map[string]string // storeID -> nome da loja (cache)
}

// Option configura o Client na construção.
type Option func(*Client)

// WithBaseURL sobrescreve a URL base da API (útil em testes).
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
}

// New cria um Client. Se httpClient for nil, usa http.DefaultClient.
func New(httpClient *http.Client, userAgent string, opts ...Option) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	c := &Client{
		httpClient: httpClient,
		baseURL:    defaultBaseURL,
		userAgent:  userAgent,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type dealDTO struct {
	Title       string `json:"title"`
	DealID      string `json:"dealID"`
	StoreID     string `json:"storeID"`
	SalePrice   string `json:"salePrice"`
	NormalPrice string `json:"normalPrice"`
}

type storeDTO struct {
	StoreID   string `json:"storeID"`
	StoreName string `json:"storeName"`
}

// Search busca ofertas para um termo de jogo. Resultado vazio (sem erro)
// significa "nada encontrado".
func (c *Client) Search(ctx context.Context, game string) ([]price.Offer, error) {
	game = strings.TrimSpace(game)
	if game == "" {
		return nil, nil
	}

	stores, err := c.loadStores(ctx)
	if err != nil {
		return nil, err
	}

	deals, err := c.fetchDeals(ctx, game)
	if err != nil {
		return nil, err
	}

	offers := make([]price.Offer, 0, len(deals))
	for _, d := range deals {
		offers = append(offers, price.Offer{
			Title:  d.Title,
			Store:  storeName(stores, d.StoreID),
			Price:  parsePrice(d.SalePrice),
			Retail: parsePrice(d.NormalPrice),
			URL:    redirectURL(d.DealID),
		})
	}
	return offers, nil
}

func (c *Client) fetchDeals(ctx context.Context, game string) ([]dealDTO, error) {
	query := url.Values{
		"title":    {game},
		"limit":    {dealsLimit},
		"pageSize": {dealsLimit},
	}
	var deals []dealDTO
	if err := c.getJSON(ctx, c.baseURL+"/deals?"+query.Encode(), &deals); err != nil {
		return nil, err
	}
	return deals, nil
}

// loadStores busca e memoriza o mapa storeID -> nome. O cache só é preenchido em
// caso de sucesso, permitindo nova tentativa após falha transitória.
func (c *Client) loadStores(ctx context.Context) (map[string]string, error) {
	c.storesMu.Lock()
	defer c.storesMu.Unlock()

	if c.stores != nil {
		return c.stores, nil
	}

	var raw []storeDTO
	if err := c.getJSON(ctx, c.baseURL+"/stores", &raw); err != nil {
		return nil, err
	}

	m := make(map[string]string, len(raw))
	for _, s := range raw {
		m[s.StoreID] = s.StoreName
	}
	c.stores = m
	return m, nil
}

func (c *Client) getJSON(ctx context.Context, endpoint string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("criando requisição: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executando requisição: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		return fmt.Errorf("cheapshark respondeu status inesperado: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decodificando resposta do cheapshark: %w", err)
	}
	return nil
}

func storeName(stores map[string]string, id string) string {
	if name, ok := stores[id]; ok && name != "" {
		return name
	}
	return "Loja " + id
}

func parsePrice(raw string) float64 {
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0
	}
	return v
}

func redirectURL(dealID string) string {
	return redirectBaseURL + "?" + url.Values{"dealID": {dealID}}.Encode()
}
