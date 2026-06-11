package download

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildSegmentsKnownSize(t *testing.T) {
	segs := buildSegments(1000, 4)
	if len(segs) != 4 {
		t.Fatalf("expected 4 segments, got %d", len(segs))
	}
	if segs[0].Start != 0 {
		t.Fatalf("first segment must start at 0, got %d", segs[0].Start)
	}
	if last := segs[len(segs)-1]; last.End != 999 {
		t.Fatalf("last segment must end at 999, got %d", last.End)
	}
	// Segments must be contiguous and non-overlapping.
	for i := 1; i < len(segs); i++ {
		if segs[i].Start != segs[i-1].End+1 {
			t.Fatalf("segment %d not contiguous: start=%d prevEnd=%d", i, segs[i].Start, segs[i-1].End)
		}
	}
}

func TestBuildSegmentsUnknownSize(t *testing.T) {
	segs := buildSegments(0, 8)
	if len(segs) != 1 {
		t.Fatalf("unknown size must yield a single segment, got %d", len(segs))
	}
	if segs[0].End != -1 {
		t.Fatalf("unknown-size segment must have End -1 (open-ended), got %d", segs[0].End)
	}
}

func TestBuildSegmentsSingleConnection(t *testing.T) {
	segs := buildSegments(500, 1)
	if len(segs) != 1 || segs[0].Start != 0 || segs[0].End != 499 {
		t.Fatalf("single-connection download should be one full segment, got %+v", segs)
	}
}

func TestUniqueName(t *testing.T) {
	dir := t.TempDir()
	if name := uniqueName(dir, "file.bin"); name != "file.bin" {
		t.Fatalf("expected original name when no conflict, got %q", name)
	}
	if err := os.WriteFile(filepath.Join(dir, "file.bin"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	name := uniqueName(dir, "file.bin")
	if name == "" || name == "file.bin" {
		t.Fatalf("expected a new unique name when the file already exists, got %q", name)
	}
}

func TestNewIDUnique(t *testing.T) {
	a, b := newID(), newID()
	if a == "" || a == b {
		t.Fatalf("ids must be non-empty and unique, got %q and %q", a, b)
	}
}
