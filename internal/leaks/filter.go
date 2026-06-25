package leaks

import "strings"

// FilterByTerm mantém os itens cujo título contém todos os tokens do termo
// (busca AND, case-insensitive). Um termo vazio mantém todos os itens. Centraliza
// o filtro para que todas as fontes e o agregador apliquem a mesma regra.
func FilterByTerm(items []Leak, term string) []Leak {
	tokens := strings.Fields(strings.ToLower(term))
	if len(tokens) == 0 {
		return items
	}

	out := make([]Leak, 0, len(items))
	for _, it := range items {
		if matchesAll(strings.ToLower(it.Title), tokens) {
			out = append(out, it)
		}
	}
	return out
}

func matchesAll(haystack string, tokens []string) bool {
	for _, tok := range tokens {
		if !strings.Contains(haystack, tok) {
			return false
		}
	}
	return true
}
