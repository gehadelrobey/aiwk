package cache

import (
	"path/filepath"
	"testing"
)

func TestPutGet(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "c.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	const (
		q  = "print second column"
		fs = ""
	)
	if _, hit, err := s.Get(q, fs, false, "openai"); err != nil {
		t.Fatal(err)
	} else if hit {
		t.Fatal("unexpected cache hit")
	}
	const prog = "{ print $2 }"
	if err := s.Put(q, fs, false, "openai", prog); err != nil {
		t.Fatal(err)
	}
	got, hit, err := s.Get(q, fs, false, "openai")
	if err != nil {
		t.Fatal(err)
	}
	if !hit || got != prog {
		t.Fatalf("Get = %q, hit=%v; want %q, true", got, hit, prog)
	}
}

func TestClear(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "x.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.Put("q", "", false, "p", "awk"); err != nil {
		t.Fatal(err)
	}
	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}
	_, hit, err := s.Get("q", "", false, "p")
	if err != nil {
		t.Fatal(err)
	}
	if hit {
		t.Fatal("expected cache empty after clear")
	}
}
