package cli

import (
	"strings"
	"testing"
)

func TestEngramPathGuidanceFish(t *testing.T) {
	msg := engramPathGuidance("/usr/bin/fish")
	if want := "fish_user_paths"; !strings.Contains(msg, want) {
		t.Fatalf("engramPathGuidance(fish) missing %q: %s", want, msg)
	}
}

func TestEngramPathGuidanceZsh(t *testing.T) {
	msg := engramPathGuidance("/bin/zsh")
	if want := ".zshrc"; !strings.Contains(msg, want) {
		t.Fatalf("engramPathGuidance(zsh) missing %q: %s", want, msg)
	}
}

func TestEngramPathGuidanceDefault(t *testing.T) {
	msg := engramPathGuidance("")
	if want := "$HOME/go/bin"; !strings.Contains(msg, want) {
		t.Fatalf("engramPathGuidance(default) missing %q: %s", want, msg)
	}
}
