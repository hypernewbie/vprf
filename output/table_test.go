package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestWriteTable(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"name", "value", "count"}
	rows := [][]string{
		{"foo", "10", "100"},
		{"bar", "2", "20"},
		{"baz", "200", "3"},
	}
	WriteTable(&buf, headers, rows)
	out := buf.String()
	lines := splitLines(out)
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(lines))
	}
	if !contains(out, "name") || !contains(out, "foo") {
		t.Fatalf("expected headers and data in output, got %q", out)
	}
}

func TestWriteTableEmpty(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"col1", "col2"}
	WriteTable(&buf, headers, nil)
	out := buf.String()
	lines := splitLines(out)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line (header only), got %d", len(lines))
	}
	if !contains(out, "col1") {
		t.Fatalf("expected header in output, got %q", out)
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	data := struct {
		Name string `json:"name"`
		Val  int    `json:"value"`
	}{Name: "test", Val: 42}
	err := WriteJSON(&buf, data)
	if err != nil {
		t.Fatalf("WriteJSON error: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["name"] != "test" {
		t.Fatalf("expected name=test, got %v", parsed["name"])
	}
}

func TestWriteJSONSlice(t *testing.T) {
	var buf bytes.Buffer
	data := []string{"a", "b", "c"}
	err := WriteJSON(&buf, data)
	if err != nil {
		t.Fatalf("WriteJSON error: %v", err)
	}
	var parsed []string
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(parsed) != 3 || parsed[0] != "a" {
		t.Fatalf("expected [a,b,c], got %v", parsed)
	}
}

func splitLines(s string) []string {
	var lines []string
	for _, line := range bytes.Split([]byte(s), []byte{'\n'}) {
		if len(line) > 0 {
			lines = append(lines, string(line))
		}
	}
	return lines
}

func contains(s, sub string) bool {
	return bytes.Contains([]byte(s), []byte(sub))
}
