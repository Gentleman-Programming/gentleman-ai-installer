package filemerge

import (
	"encoding/json"
	"testing"
)

func TestMergeJSONObjectsRecursively(t *testing.T) {
	base := []byte(`{"plugins":["a"],"settings":{"theme":"default","flags":{"x":true}}}`)
	overlay := []byte(`{"settings":{"theme":"gentleman","flags":{"y":true}},"extra":1}`)

	merged, err := MergeJSONObjects(base, overlay)
	if err != nil {
		t.Fatalf("MergeJSONObjects() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(merged, &got); err != nil {
		t.Fatalf("Unmarshal merged json error = %v", err)
	}

	settings := got["settings"].(map[string]any)
	flags := settings["flags"].(map[string]any)

	if settings["theme"] != "gentleman" {
		t.Fatalf("theme = %v", settings["theme"])
	}

	if flags["x"] != true || flags["y"] != true {
		t.Fatalf("flags = %#v", flags)
	}

	plugins := got["plugins"].([]any)
	if len(plugins) != 1 || plugins[0] != "a" {
		t.Fatalf("plugins = %#v", plugins)
	}
}
