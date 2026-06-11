package settings

import (
	"path/filepath"
	"testing"
)

func TestDefaultsAndPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	s, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if s.Get().Theme != "midnight" {
		t.Fatalf("expected default theme midnight, got %q", s.Get().Theme)
	}

	next := s.Get()
	next.Theme = "aurora"
	next.DefaultConnections = 16
	if _, err := s.Save(next); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// A fresh Store must load the persisted values from disk.
	s2, err := New(path)
	if err != nil {
		t.Fatalf("New(reload): %v", err)
	}
	if s2.Get().Theme != "aurora" || s2.Get().DefaultConnections != 16 {
		t.Fatalf("settings did not persist: %+v", s2.Get())
	}
}

func TestSaveKeepsConnectionsWhenZero(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	s, _ := New(path)
	next := s.Get()
	next.DefaultConnections = 0 // invalid; should fall back to the previous value
	saved, err := s.Save(next)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if saved.DefaultConnections <= 0 {
		t.Fatalf("expected connections to fall back to a positive default, got %d", saved.DefaultConnections)
	}
}
