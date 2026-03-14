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

func TestMergeJSONObjectsSupportsJSONCBase(t *testing.T) {
	base := []byte(`{
	  // VS Code-style comments and trailing commas
	  "editor.fontSize": 14,
	  "files.exclude": {
	    "**/.git": true,
	  },
	}`)
	overlay := []byte(`{"chat.tools.autoApprove": true}`)

	merged, err := MergeJSONObjects(base, overlay)
	if err != nil {
		t.Fatalf("MergeJSONObjects() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(merged, &got); err != nil {
		t.Fatalf("Unmarshal merged json error = %v", err)
	}

	autoApprove, ok := got["chat.tools.autoApprove"].(bool)
	if !ok || !autoApprove {
		t.Fatalf("chat.tools.autoApprove = %#v", got["chat.tools.autoApprove"])
	}

	if got["editor.fontSize"] != float64(14) {
		t.Fatalf("editor.fontSize = %v", got["editor.fontSize"])
	}
}

func TestMergeJSONObjectsReturnsErrorForInvalidJSON(t *testing.T) {
	base := []byte(`{"ok": true`)
	overlay := []byte(`{"chat.tools.autoApprove": true}`)

	if _, err := MergeJSONObjects(base, overlay); err == nil {
		t.Fatal("expected error for invalid JSON input")
	}
}
