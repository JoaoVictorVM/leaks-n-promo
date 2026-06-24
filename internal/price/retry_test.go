package price

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeProvider falha as primeiras failTimes chamadas e depois retorna offers.
type fakeProvider struct {
	calls     int
	failTimes int
	err       error
	offers    []Offer
}

func (f *fakeProvider) Search(_ context.Context, _ string) ([]Offer, error) {
	f.calls++
	if f.calls <= f.failTimes {
		return nil, f.err
	}
	return f.offers, nil
}

// newTestRetrying cria o decorator com sleep instantâneo (sem esperar de fato).
func newTestRetrying(next PriceProvider, attempts int) *retrying {
	r := NewRetrying(next, RetryConfig{Attempts: attempts, BaseDelay: time.Millisecond, MaxDelay: time.Second}).(*retrying)
	r.sleep = func(context.Context, time.Duration) error { return nil }
	return r
}

func TestRetrySucceedsAfterFailures(t *testing.T) {
	want := []Offer{{Title: "Celeste", Store: "Steam"}}
	fake := &fakeProvider{failTimes: 2, err: errors.New("falha transitória"), offers: want}
	r := newTestRetrying(fake, 3)

	offers, err := r.Search(context.Background(), "celeste")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(offers) != 1 || offers[0].Title != "Celeste" {
		t.Fatalf("ofertas inesperadas: %+v", offers)
	}
	if fake.calls != 3 {
		t.Errorf("esperava 3 tentativas, obtive %d", fake.calls)
	}
}

func TestRetryExhaustsAttempts(t *testing.T) {
	sentinel := errors.New("indisponível")
	fake := &fakeProvider{failTimes: 99, err: sentinel}
	r := newTestRetrying(fake, 3)

	_, err := r.Search(context.Background(), "celeste")
	if err == nil {
		t.Fatal("esperava erro após esgotar as tentativas")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("erro final deveria envolver o original: %v", err)
	}
	if fake.calls != 3 {
		t.Errorf("esperava 3 tentativas, obtive %d", fake.calls)
	}
}

func TestRetryStopsOnContextCancel(t *testing.T) {
	fake := &fakeProvider{failTimes: 99, err: errors.New("falha")}
	r := NewRetrying(fake, RetryConfig{Attempts: 5, BaseDelay: time.Millisecond}).(*retrying)
	// sleep simula cancelamento do contexto durante o backoff.
	r.sleep = func(context.Context, time.Duration) error { return context.Canceled }

	_, err := r.Search(context.Background(), "celeste")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("esperava context.Canceled, obtive %v", err)
	}
	if fake.calls != 1 {
		t.Errorf("esperava parar após a 1ª tentativa, obtive %d", fake.calls)
	}
}

func TestRetryNoErrorNoRetry(t *testing.T) {
	fake := &fakeProvider{failTimes: 0, offers: nil}
	r := newTestRetrying(fake, 3)

	if _, err := r.Search(context.Background(), "celeste"); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if fake.calls != 1 {
		t.Errorf("sucesso não deveria gerar retry; tentativas = %d", fake.calls)
	}
}

func TestAttemptAppliesTimeout(t *testing.T) {
	var hadDeadline bool
	probe := providerFunc(func(ctx context.Context, _ string) ([]Offer, error) {
		_, hadDeadline = ctx.Deadline()
		return nil, nil
	})
	r := NewRetrying(probe, RetryConfig{Attempts: 1, Timeout: time.Second}).(*retrying)

	if _, err := r.Search(context.Background(), "celeste"); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !hadDeadline {
		t.Error("esperava que a tentativa recebesse um contexto com deadline")
	}
}

func TestNextDelayCapsAtMax(t *testing.T) {
	if got := nextDelay(800*time.Millisecond, time.Second); got != time.Second {
		t.Errorf("nextDelay = %v, esperava 1s (teto)", got)
	}
	if got := nextDelay(100*time.Millisecond, time.Second); got != 200*time.Millisecond {
		t.Errorf("nextDelay = %v, esperava 200ms", got)
	}
}

func TestContextSleepCompletes(t *testing.T) {
	if err := contextSleep(context.Background(), time.Millisecond); err != nil {
		t.Fatalf("esperava conclusão sem erro, obtive %v", err)
	}
	// d<=0 retorna imediatamente.
	if err := contextSleep(context.Background(), 0); err != nil {
		t.Fatalf("d<=0 deveria retornar nil, obtive %v", err)
	}
}

func TestContextSleepCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := contextSleep(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("esperava context.Canceled, obtive %v", err)
	}
}

type providerFunc func(context.Context, string) ([]Offer, error)

func (f providerFunc) Search(ctx context.Context, game string) ([]Offer, error) {
	return f(ctx, game)
}
