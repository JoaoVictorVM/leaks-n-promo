package sqlite

import (
	"context"
	"path/filepath"
	"testing"
)

func openTemp(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "seen.db")
	s, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("abrindo store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestUnseenEmptyStore(t *testing.T) {
	s := openTemp(t)
	in := []string{"a", "b", "c"}

	got, err := s.Unseen(context.Background(), in)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("esperava todas as URLs como não-vistas, obtive %v", got)
	}
}

func TestMarkAndUnseenPreservesOrder(t *testing.T) {
	s := openTemp(t)
	ctx := context.Background()

	if err := s.MarkSeen(ctx, []string{"b"}); err != nil {
		t.Fatalf("MarkSeen: %v", err)
	}

	got, err := s.Unseen(ctx, []string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("Unseen: %v", err)
	}
	want := []string{"a", "c"}
	if len(got) != len(want) || got[0] != "a" || got[1] != "c" {
		t.Fatalf("Unseen = %v, esperava %v", got, want)
	}
}

func TestMarkSeenIdempotent(t *testing.T) {
	s := openTemp(t)
	ctx := context.Background()

	if err := s.MarkSeen(ctx, []string{"a"}); err != nil {
		t.Fatalf("1ª MarkSeen: %v", err)
	}
	if err := s.MarkSeen(ctx, []string{"a"}); err != nil {
		t.Fatalf("2ª MarkSeen deveria ser idempotente: %v", err)
	}

	got, err := s.Unseen(ctx, []string{"a"})
	if err != nil {
		t.Fatalf("Unseen: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("esperava nenhuma não-vista, obtive %v", got)
	}
}

func TestPersistsAcrossReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "seen.db")

	s1, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("1ª abertura: %v", err)
	}
	if err := s1.MarkSeen(ctx, []string{"a"}); err != nil {
		t.Fatalf("MarkSeen: %v", err)
	}
	if err := s1.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	s2, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("reabertura: %v", err)
	}
	defer func() { _ = s2.Close() }()

	got, err := s2.Unseen(ctx, []string{"a", "b"})
	if err != nil {
		t.Fatalf("Unseen: %v", err)
	}
	if len(got) != 1 || got[0] != "b" {
		t.Fatalf("estado não persistiu: Unseen = %v, esperava [b]", got)
	}
}

func TestEmptyInputs(t *testing.T) {
	s := openTemp(t)
	ctx := context.Background()

	if got, err := s.Unseen(ctx, nil); err != nil || got != nil {
		t.Fatalf("Unseen(nil) = %v, %v; esperava nil, nil", got, err)
	}
	if err := s.MarkSeen(ctx, nil); err != nil {
		t.Fatalf("MarkSeen(nil) deveria ser no-op, obtive %v", err)
	}
}
