package cmd

import (
	"bytes"
	"encoding/json"
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

func TestTopJsonOutput(t *testing.T) {
	profilePath := filepath.Join("..", "tests", "testdata", "fixture.json")
	var out bytes.Buffer
	var errBuf bytes.Buffer
	if err := Execute([]string{"top", "-p", profilePath, "--format", "json"}, &out, &errBuf); err != nil {
		t.Fatalf("top --format json failed: %v stderr=%s", err, errBuf.String())
	}
	var result []map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON array: %v\noutput: %s", err, out.String())
	}
	if len(result) == 0 {
		t.Fatalf("expected non-empty JSON array")
	}
	for _, item := range result {
		if item["name"] == nil || item["self_samples"] == nil {
			t.Fatalf("expected name and self_samples fields in JSON object, got %v", item)
		}
	}
}

func TestSummaryJsonOutput(t *testing.T) {
	profilePath := filepath.Join("..", "tests", "testdata", "fixture.json")
	var out bytes.Buffer
	var errBuf bytes.Buffer
	if err := Execute([]string{"summary", "-p", profilePath, "--format", "json"}, &out, &errBuf); err != nil {
		t.Fatalf("summary --format json failed: %v stderr=%s", err, errBuf.String())
	}
	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON object: %v\noutput: %s", err, out.String())
	}
	if result["profile_name"] == nil || result["total_samples"] == nil {
		t.Fatalf("expected profile_name and total_samples fields in JSON, got %v", result)
	}
}

func TestThreadsJsonOutput(t *testing.T) {
	profilePath := filepath.Join("..", "tests", "testdata", "fixture.json")
	var out bytes.Buffer
	var errBuf bytes.Buffer
	if err := Execute([]string{"threads", "-p", profilePath, "--format", "json"}, &out, &errBuf); err != nil {
		t.Fatalf("threads --format json failed: %v stderr=%s", err, errBuf.String())
	}
	var result []map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON array: %v\noutput: %s", err, out.String())
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 threads in JSON output, got %d", len(result))
	}
}
