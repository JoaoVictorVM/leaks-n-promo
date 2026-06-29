// Package store define a persistência de deduplicação usada pelo push de leaks:
// quais URLs já foram notificadas. As implementações devem ser seguras para uso
// concorrente.
package store

import "context"

// SeenStore registra quais itens (identificados por URL) já foram notificados.
//
// O consumo recomendado é: consultar Unseen, notificar e só então MarkSeen — de
// modo que uma falha no envio não marque o item como visto (at-least-once).
type SeenStore interface {
	// Unseen retorna, dentre as URLs dadas, as que ainda não foram registradas,
	// preservando a ordem de entrada.
	Unseen(ctx context.Context, urls []string) ([]string, error)
	// MarkSeen registra as URLs como já vistas (idempotente).
	MarkSeen(ctx context.Context, urls []string) error
}
