// Package price define o contrato de busca de preços de jogos e os tipos de
// domínio correspondentes. A implementação concreta (CheapShark) fica em um
// subpacote, mantendo a interface no consumidor.
package price

import "context"

// Offer representa uma oferta de um jogo em uma loja específica. Preços em USD,
// conforme a fonte (CheapShark).
type Offer struct {
	Title  string  // nome do jogo
	Store  string  // nome da loja
	Price  float64 // preço atual em USD
	Retail float64 // preço cheio (sem desconto) em USD
	URL    string  // link do CheapShark para a oferta
}

// PriceProvider busca ofertas de preço para um termo de jogo. Um resultado vazio
// (sem erro) significa "nada encontrado"; erros são reservados a falhas reais
// (rede, indisponibilidade, rate limit).
type PriceProvider interface {
	Search(ctx context.Context, game string) ([]Offer, error)
}
