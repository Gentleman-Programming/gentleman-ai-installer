package gga

import "testing"

func TestWriteDefaultConfigIsIdempotent(t *testing.T) {
	home := t.TempDir()

	first, err := WriteDefaultConfig(home)
	if err != nil {
		t.Fatalf("WriteDefaultConfig() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("WriteDefaultConfig() first changed = false")
	}

	second, err := WriteDefaultConfig(home)
	if err != nil {
		t.Fatalf("WriteDefaultConfig() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("WriteDefaultConfig() second changed = true")
	}
}
