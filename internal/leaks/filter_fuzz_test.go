package leaks

import (
	"strings"
	"testing"
)

// FuzzFilterByTerm exercita o parsing do termo dos comandos contra entradas
// arbitrárias. Invariantes: nunca entra em pânico; o resultado é subconjunto da
// entrada; e um item só permanece se todos os tokens do termo estão no título.
func FuzzFilterByTerm(f *testing.F) {
	seeds := []struct{ term, title string }{
		{"gta", "GTA 6 leak"},
		{"", "qualquer coisa"},
		{"   ", "espaços"},
		{"zelda totk", "Zelda TOTK news"},
		{"日本語", "rumor 日本語 vazado"},
	}
	for _, s := range seeds {
		f.Add(s.term, s.title)
	}

	f.Fuzz(func(t *testing.T, term, title string) {
		items := []Leak{{Title: title}}

		got := FilterByTerm(items, term)

		if len(got) > len(items) {
			t.Fatalf("resultado maior que a entrada: %d > %d", len(got), len(items))
		}

		tokens := strings.Fields(strings.ToLower(term))
		if len(got) == 1 {
			// Item mantido: todos os tokens devem aparecer no título.
			lower := strings.ToLower(title)
			for _, tok := range tokens {
				if !strings.Contains(lower, tok) {
					t.Fatalf("item mantido sem o token %q no título %q", tok, title)
				}
			}
		} else if len(tokens) == 0 {
			// Termo sem tokens deveria manter tudo.
			t.Fatalf("termo vazio %q deveria manter o item", term)
		}
	})
}
