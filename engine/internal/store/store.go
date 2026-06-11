// Package store persists downloads as JSON files so the engine can resume them
// after a restart. The interface is intentionally tiny; a SQLite-backed Store
// is a great first contribution (see docs/ROADMAP.md).
package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/0xheycat/pounce/internal/model"
)

// Store reads and writes download metadata to a directory of JSON files.
type Store struct {
	dir string
	mu  sync.Mutex
}

// New creates the data directory if needed and returns a Store.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

func (s *Store) path(id string) string {
	return filepath.Join(s.dir, id+".json")
}

// Save atomically writes a download's metadata.
func (s *Store) Save(d *model.Download) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path(d.ID) + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path(d.ID))
}

// Delete removes a download's metadata file.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := os.Remove(s.path(id))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// LoadAll returns every persisted download.
func (s *Store) LoadAll() ([]*model.Download, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var out []*model.Download
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			continue
		}
		var d model.Download
		if err := json.Unmarshal(b, &d); err != nil {
			continue
		}
		out = append(out, &d)
	}
	return out, nil
}
