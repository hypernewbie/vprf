package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSummaryAndTopOnRealProfile(t *testing.T) {
	profilePath := filepath.Join("..", "real-profile-debug.json.gz")
	if _, err := os.Stat(profilePath); err != nil {
		t.Skip("real samply profile not present")
	}

	var out bytes.Buffer
	var errBuf bytes.Buffer
	if err := Execute([]string{"summary", "-p", profilePath}, &out, &errBuf); err != nil {
		t.Fatalf("summary failed: %v stderr=%s", err, errBuf.String())
	}
	if !strings.Contains(out.String(), "Top function:") {
		t.Fatalf("expected summary output to contain top function line, got %q", out.String())
	}

	out.Reset()
	errBuf.Reset()
	if err := Execute([]string{"top", "-p", profilePath, "--limit", "10"}, &out, &errBuf); err != nil {
		t.Fatalf("top failed: %v stderr=%s", err, errBuf.String())
	}
	if !strings.Contains(out.String(), "main.innerLoop") || !strings.Contains(out.String(), "main.outer") {
		t.Fatalf("expected top output to contain Go symbols, got %q", out.String())
	}
}
