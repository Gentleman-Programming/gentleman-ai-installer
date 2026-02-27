package filemerge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomicCreatesAndIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "config.json")
	content := []byte("{\"ok\":true}\n")

	first, err := WriteFileAtomic(path, content, 0o644)
	if err != nil {
		t.Fatalf("WriteFileAtomic() first write error = %v", err)
	}

	if !first.Changed || !first.Created {
		t.Fatalf("WriteFileAtomic() first write result = %+v", first)
	}

	second, err := WriteFileAtomic(path, content, 0o644)
	if err != nil {
		t.Fatalf("WriteFileAtomic() second write error = %v", err)
	}

	if second.Changed || second.Created {
		t.Fatalf("WriteFileAtomic() second write result = %+v", second)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if string(got) != string(content) {
		t.Fatalf("file content = %q", string(got))
	}
}
