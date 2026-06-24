package price

import (
	"context"
	"fmt"
	"time"
)

// RetryConfig parametriza o decorator resiliente.
type RetryConfig struct {
	Attempts  int           // número total de tentativas (mínimo 1)
	BaseDelay time.Duration // atraso inicial entre tentativas
	MaxDelay  time.Duration // teto do atraso
	Timeout   time.Duration // timeout por tentativa (<=0 desativa)
}

type retrying struct {
	next  PriceProvider
	cfg   RetryConfig
	sleep func(context.Context, time.Duration) error // injetável para testes
}

// NewRetrying decora um PriceProvider com timeout por tentativa e retry com
// backoff exponencial. Erros disparam nova tentativa; um resultado vazio sem
// erro ("nada encontrado") não é considerado falha.
func NewRetrying(next PriceProvider, cfg RetryConfig) PriceProvider {
	if cfg.Attempts < 1 {
		cfg.Attempts = 1
	}
	return &retrying{next: next, cfg: cfg, sleep: contextSleep}
}

func (r *retrying) Search(ctx context.Context, game string) ([]Offer, error) {
	var lastErr error
	delay := r.cfg.BaseDelay

	for attempt := 1; attempt <= r.cfg.Attempts; attempt++ {
		offers, err := r.attempt(ctx, game)
		if err == nil {
			return offers, nil
		}
		lastErr = err

		if attempt == r.cfg.Attempts {
			break
		}
		if werr := r.sleep(ctx, delay); werr != nil {
			return nil, werr
		}
		delay = nextDelay(delay, r.cfg.MaxDelay)
	}

	return nil, fmt.Errorf("busca de preço falhou após %d tentativa(s): %w", r.cfg.Attempts, lastErr)
}

func (r *retrying) attempt(ctx context.Context, game string) ([]Offer, error) {
	if r.cfg.Timeout <= 0 {
		return r.next.Search(ctx, game)
	}
	attemptCtx, cancel := context.WithTimeout(ctx, r.cfg.Timeout)
	defer cancel()
	return r.next.Search(attemptCtx, game)
}

func nextDelay(current, max time.Duration) time.Duration {
	next := current * 2
	if max > 0 && next > max {
		return max
	}
	return next
}

func contextSleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return fmt.Errorf("aguardando backoff: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}
