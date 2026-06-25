package leaks

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
)

// Aggregator combina várias LeakSource em uma só, deduplicando por URL e
// ordenando por recência. É resiliente: tolera falhas parciais e só retorna erro
// se todas as fontes falharem. Implementa a própria interface LeakSource.
type Aggregator struct {
	sources []LeakSource
}

// NewAggregator cria um agregador. A ordem das fontes define a prioridade na
// deduplicação (a primeira ocorrência de uma URL prevalece).
func NewAggregator(sources ...LeakSource) *Aggregator {
	return &Aggregator{sources: sources}
}

var _ LeakSource = (*Aggregator)(nil)

// Fetch consulta todas as fontes em paralelo e devolve o resultado combinado,
// deduplicado e ordenado do mais recente para o mais antigo.
func (a *Aggregator) Fetch(ctx context.Context, term string) ([]Leak, error) {
	results := make([][]Leak, len(a.sources))
	fetchErrs := make([]error, len(a.sources))

	var wg sync.WaitGroup
	for i, src := range a.sources {
		wg.Add(1)
		go func(i int, s LeakSource) {
			defer wg.Done()
			results[i], fetchErrs[i] = s.Fetch(ctx, term)
		}(i, src)
	}
	wg.Wait()

	var merged []Leak
	var errs []error
	for i := range a.sources {
		if fetchErrs[i] != nil {
			errs = append(errs, fetchErrs[i])
			continue
		}
		merged = append(merged, results[i]...)
	}

	if len(merged) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("todas as fontes de leaks falharam: %w", errors.Join(errs...))
	}

	return sortByRecency(dedupeByURL(merged)), nil
}

// dedupeByURL remove itens com URL repetida, mantendo a primeira ocorrência.
func dedupeByURL(items []Leak) []Leak {
	seen := make(map[string]struct{}, len(items))
	out := make([]Leak, 0, len(items))
	for _, it := range items {
		if it.URL != "" {
			if _, dup := seen[it.URL]; dup {
				continue
			}
			seen[it.URL] = struct{}{}
		}
		out = append(out, it)
	}
	return out
}

func sortByRecency(items []Leak) []Leak {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Published.After(items[j].Published)
	})
	return items
}
