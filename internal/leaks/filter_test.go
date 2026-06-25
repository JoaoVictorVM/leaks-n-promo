package leaks

import "testing"

func TestFilterByTerm(t *testing.T) {
	items := []Leak{
		{Title: "GTA 6 map leaked online"},
		{Title: "New Zelda rumor surfaces"},
		{Title: "GTA Online weekly update"},
	}

	tests := []struct {
		name      string
		term      string
		wantCount int
	}{
		{name: "termo vazio mantém tudo", term: "", wantCount: 3},
		{name: "apenas espaços mantém tudo", term: "   ", wantCount: 3},
		{name: "token único", term: "gta", wantCount: 2},
		{name: "case-insensitive", term: "ZELDA", wantCount: 1},
		{name: "multi-token AND presente", term: "gta map", wantCount: 1},
		{name: "multi-token AND ausente", term: "gta zelda", wantCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterByTerm(items, tt.term)
			if len(got) != tt.wantCount {
				t.Fatalf("FilterByTerm(%q) = %d itens, esperava %d", tt.term, len(got), tt.wantCount)
			}
		})
	}
}
