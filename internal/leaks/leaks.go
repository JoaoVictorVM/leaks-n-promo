// Package leaks define o contrato de busca de vazamentos/rumores de games e os
// tipos de domínio. As implementações (RSS, Reddit) ficam em subpacotes,
// mantendo a interface no consumidor.
package leaks

import (
	"context"
	"time"
)

// Leak representa um item de vazamento/rumor agregado de uma fonte.
type Leak struct {
	Title     string    // título do item
	Source    string    // nome da fonte (ex.: nome do site/feed, "Reddit")
	URL       string    // link para o item original (usado também na deduplicação)
	Published time.Time // data de publicação
}

// LeakSource busca vazamentos. Um termo vazio significa "os mais recentes"; com
// termo, filtra por relevância ao termo. Resultado vazio sem erro significa
// "nada encontrado"; erros são reservados a falhas reais.
type LeakSource interface {
	Fetch(ctx context.Context, term string) ([]Leak, error)
}
