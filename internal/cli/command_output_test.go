package cli

import (
	"strings"
	"testing"
)

func TestExecuteCommandQuietModeIncludesCapturedOutputOnFailure(t *testing.T) {
	restore := SetCommandOutputStreaming(false)
	defer restore()

	err := executeCommand("bash", "-c", "echo boom && exit 1")
	if err == nil {
		t.Fatal("executeCommand() error = nil, want non-nil")
	}

	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("executeCommand() error = %q, want captured output", err.Error())
	}
}

func TestSetCommandOutputStreamingRestore(t *testing.T) {
	streamCommandOutput = true
	restore := SetCommandOutputStreaming(false)

	if streamCommandOutput {
		t.Fatal("streamCommandOutput should be false after SetCommandOutputStreaming(false)")
	}

	restore()
	if !streamCommandOutput {
		t.Fatal("restore should reset streamCommandOutput to previous value")
	}
}
