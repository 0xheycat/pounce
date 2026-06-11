package store

import (
	"testing"

	"github.com/0xheycat/pounce/internal/model"
)

func TestSaveLoadDelete(t *testing.T) {
	dir := t.TempDir()
	s, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	d := &model.Download{
		ID:       "abc123",
		URL:      "http://example.com/file.bin",
		Filename: "file.bin",
		Status:   model.StatusQueued,
	}
	if err := s.Save(d); err != nil {
		t.Fatalf("Save: %v", err)
	}

	all, err := s.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(all) != 1 || all[0].ID != "abc123" {
		t.Fatalf("expected to load the saved download, got %+v", all)
	}

	if err := s.Delete("abc123"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	all, _ = s.LoadAll()
	if len(all) != 0 {
		t.Fatalf("expected no downloads after delete, got %d", len(all))
	}
}
