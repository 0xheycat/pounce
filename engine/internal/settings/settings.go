// Package settings persists user-configurable preferences (theme, default
// download options) to a small JSON file. It is concurrency-safe and writes
// atomically so a crash mid-save never corrupts the file.
package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Settings holds preferences shared between the engine and the dashboard.
type Settings struct {
	Theme              string `json:"theme"`
	DefaultDir         string `json:"defaultDir"`
	DefaultConnections int    `json:"defaultConnections"`
	DefaultSpeedLimit  int64  `json:"defaultSpeedLimit"`
	MaxConcurrent      int    `json:"maxConcurrent"`
	NotifyOnComplete   bool   `json:"notifyOnComplete"`
}

// Defaults returns the baseline settings used on first run.
func Defaults() Settings {
	return Settings{
		Theme:              "midnight",
		DefaultDir:         "",
		DefaultConnections: 8,
		DefaultSpeedLimit:  0,
		MaxConcurrent:      4,
		NotifyOnComplete:   true,
	}
}

// Store is a file-backed, concurrency-safe settings store.
type Store struct {
	path string
	mu   sync.RWMutex
	cur  Settings
}

// New loads settings from path, falling back to defaults when the file is
// missing or unreadable so the engine always starts.
func New(path string) (*Store, error) {
	s := &Store{path: path, cur: Defaults()}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}
	var loaded Settings
	if err := json.Unmarshal(b, &loaded); err != nil {
		return s, nil // corrupt file: keep defaults
	}
	s.cur = merge(Defaults(), loaded)
	return s, nil
}

// Get returns a copy of the current settings.
func (s *Store) Get() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cur
}

// Save merges, persists and returns the updated settings. Invalid/zero numeric
// fields fall back to the previous value so a partial update never wipes them.
func (s *Store) Save(next Settings) (Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	merged := merge(s.cur, next)
	if err := s.persist(merged); err != nil {
		return s.cur, err
	}
	s.cur = merged
	return s.cur, nil
}

func (s *Store) persist(v Settings) error {
	if dir := filepath.Dir(s.path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func merge(prev, next Settings) Settings {
	out := prev
	if next.Theme != "" {
		out.Theme = next.Theme
	}
	out.DefaultDir = next.DefaultDir
	if next.DefaultConnections > 0 {
		out.DefaultConnections = next.DefaultConnections
	}
	if next.DefaultSpeedLimit >= 0 {
		out.DefaultSpeedLimit = next.DefaultSpeedLimit
	}
	if next.MaxConcurrent > 0 {
		out.MaxConcurrent = next.MaxConcurrent
	}
	out.NotifyOnComplete = next.NotifyOnComplete
	return out
}
