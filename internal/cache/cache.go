// Package cache fornece um cache em memória com TTL, seguro para uso
// concorrente. A expiração é preguiçosa (verificada no acesso).
package cache

import (
	"sync"
	"time"
)

type entry[V any] struct {
	value   V
	expires time.Time
}

// Cache é um mapa com expiração por TTL. K deve ser comparável; V é arbitrário.
type Cache[K comparable, V any] struct {
	mu    sync.Mutex
	ttl   time.Duration
	items map[K]entry[V]
	now   func() time.Time // injetável para testes
}

// New cria um cache cujas entradas expiram após ttl.
func New[K comparable, V any](ttl time.Duration) *Cache[K, V] {
	return &Cache[K, V]{
		ttl:   ttl,
		items: make(map[K]entry[V]),
		now:   time.Now,
	}
}

// Get retorna o valor associado à chave e true se presente e não expirado.
// Entradas expiradas são removidas no acesso.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	if !c.now().Before(e.expires) {
		delete(c.items, key)
		var zero V
		return zero, false
	}
	return e.value, true
}

// Set armazena o valor sob a chave, renovando o prazo de expiração.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = entry[V]{value: value, expires: c.now().Add(c.ttl)}
}

// Len retorna a quantidade de entradas armazenadas (incluindo expiradas ainda
// não removidas). Útil principalmente para testes.
func (c *Cache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.items)
}
