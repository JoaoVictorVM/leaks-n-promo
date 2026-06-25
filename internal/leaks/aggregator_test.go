package leaks

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeSource struct {
	items []Leak
	err   error
}

func (f fakeSource) Fetch(_ context.Context, _ string) ([]Leak, error) {
	return f.items, f.err
}

func at(hoursFromBase int) time.Time {
	base := time.Unix(1_000_000, 0).UTC()
	return base.Add(time.Duration(hoursFromBase) * time.Hour)
}

func TestAggregatorMergesDedupesAndSorts(t *testing.T) {
	s1 := fakeSource{items: []Leak{
		{Title: "A", URL: "u1", Published: at(3)},
		{Title: "dup-vence", URL: "u2", Published: at(1)},
	}}
	s2 := fakeSource{items: []Leak{
		{Title: "dup-perde", URL: "u2", Published: at(2)},
		{Title: "C", URL: "u3", Published: at(5)},
	}}

	agg := NewAggregator(s1, s2)
	items, err := agg.Fetch(context.Background(), "")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("esperava 3 itens únicos, obtive %d", len(items))
	}
	// Ordenado por recência: u3(5) > u1(3) > u2(1).
	wantURLs := []string{"u3", "u1", "u2"}
	for i, want := range wantURLs {
		if items[i].URL != want {
			t.Errorf("ordem[%d] = %q, esperava %q", i, items[i].URL, want)
		}
	}
	// A primeira fonte tem prioridade na dedup de u2.
	if items[2].Title != "dup-vence" {
		t.Errorf("dedup deveria manter a 1ª fonte, obtive %q", items[2].Title)
	}
}

func TestAggregatorToleratesPartialFailure(t *testing.T) {
	ok := fakeSource{items: []Leak{{Title: "ok", URL: "u1", Published: at(1)}}}
	bad := fakeSource{err: errors.New("fonte fora do ar")}

	agg := NewAggregator(ok, bad)
	items, err := agg.Fetch(context.Background(), "")
	if err != nil {
		t.Fatalf("falha parcial não deveria retornar erro: %v", err)
	}
	if len(items) != 1 || items[0].URL != "u1" {
		t.Fatalf("esperava 1 item da fonte boa, obtive %+v", items)
	}
}

func TestAggregatorAllFail(t *testing.T) {
	a := fakeSource{err: errors.New("falha a")}
	b := fakeSource{err: errors.New("falha b")}

	if _, err := NewAggregator(a, b).Fetch(context.Background(), ""); err == nil {
		t.Fatal("esperava erro quando todas as fontes falham")
	}
}

func TestAggregatorEmptyNoError(t *testing.T) {
	agg := NewAggregator(fakeSource{}, fakeSource{})
	items, err := agg.Fetch(context.Background(), "")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("esperava nenhum item, obtive %d", len(items))
	}
}
