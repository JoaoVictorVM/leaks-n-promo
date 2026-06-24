package cache

import (
	"sync"
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	c := New[string, int](time.Minute)
	c.Set("a", 42)

	v, ok := c.Get("a")
	if !ok || v != 42 {
		t.Fatalf("Get(a) = %v, %v; esperava 42, true", v, ok)
	}
}

func TestGetMissing(t *testing.T) {
	c := New[string, int](time.Minute)
	if v, ok := c.Get("x"); ok {
		t.Fatalf("Get(x) = %v, true; esperava ausente", v)
	}
}

func TestExpiration(t *testing.T) {
	c := New[string, string](time.Minute)
	current := time.Unix(0, 0)
	c.now = func() time.Time { return current }

	c.Set("k", "v")

	// Ainda dentro do TTL.
	current = current.Add(59 * time.Second)
	if _, ok := c.Get("k"); !ok {
		t.Fatal("esperava hit dentro do TTL")
	}

	// Após o TTL.
	current = current.Add(2 * time.Second)
	if _, ok := c.Get("k"); ok {
		t.Fatal("esperava miss após o TTL")
	}
	// Entrada expirada deve ter sido removida no acesso.
	if c.Len() != 0 {
		t.Fatalf("esperava 0 entradas após expiração, obtive %d", c.Len())
	}
}

func TestSetRenewsExpiry(t *testing.T) {
	c := New[string, int](time.Minute)
	current := time.Unix(0, 0)
	c.now = func() time.Time { return current }

	c.Set("k", 1)
	current = current.Add(50 * time.Second)
	c.Set("k", 2) // renova o prazo a partir de agora

	current = current.Add(50 * time.Second) // 100s desde o 1º Set, 50s desde o 2º
	v, ok := c.Get("k")
	if !ok || v != 2 {
		t.Fatalf("Get(k) = %v, %v; esperava 2, true (prazo renovado)", v, ok)
	}
}

func TestConcurrentAccess(t *testing.T) {
	c := New[int, int](time.Minute)
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Set(i, i*2)
			_, _ = c.Get(i)
			_, _ = c.Get(i - 1)
		}()
	}
	wg.Wait()

	// Sanidade: algumas chaves devem estar presentes.
	if _, ok := c.Get(10); !ok {
		t.Error("esperava chave 10 presente")
	}
}
