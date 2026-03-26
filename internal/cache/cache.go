package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"database/sql"
	_ "modernc.org/sqlite"
)

// Store persists generated awk keyed by natural language + options.
type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if path != "" && path != ":memory:" {
		if dir := filepath.Dir(path); dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, err
			}
		}
	}
	// modernc sqlite driver name is "sqlite"
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS cache (
		key TEXT PRIMARY KEY,
		awk TEXT NOT NULL,
		created_at INTEGER NOT NULL
	)`); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// DefaultPath returns ~/.cache/aiwk/cache.db (or XDG_CACHE_HOME).
func DefaultPath() (string, error) {
	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".cache")
	}
	return filepath.Join(base, "aiwk", "cache.db"), nil
}

func cacheKey(query, fieldSep string, explain bool, provider string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s\n%s\n%v\n%s", strings.TrimSpace(query), fieldSep, explain, strings.ToLower(strings.TrimSpace(provider)))
	return hex.EncodeToString(h.Sum(nil))
}

// Get returns cached awk if present.
func (s *Store) Get(query, fieldSep string, explain bool, provider string) (string, bool, error) {
	key := cacheKey(query, fieldSep, explain, provider)
	var awk string
	err := s.db.QueryRowContext(context.Background(), `SELECT awk FROM cache WHERE key = ?`, key).Scan(&awk)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return awk, true, nil
}

// Put stores awk for the given key inputs.
func (s *Store) Put(query, fieldSep string, explain bool, provider, awk string) error {
	key := cacheKey(query, fieldSep, explain, provider)
	_, err := s.db.ExecContext(context.Background(),
		`INSERT INTO cache(key, awk, created_at) VALUES(?,?,?)
		 ON CONFLICT(key) DO UPDATE SET awk=excluded.awk, created_at=excluded.created_at`,
		key, awk, time.Now().Unix())
	return err
}

// Clear removes all entries.
func (s *Store) Clear() error {
	_, err := s.db.ExecContext(context.Background(), `DELETE FROM cache`)
	return err
}

func (s *Store) Close() error {
	return s.db.Close()
}
