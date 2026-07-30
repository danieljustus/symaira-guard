package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewReporter_Default(t *testing.T) {
	r := NewReporter("")
	if r == nil {
		t.Fatal("NewReporter(\"\") returned nil")
	}
	// Should default to table
	var buf bytes.Buffer
	err := r.Print(&buf, "hello")
	if err != nil {
		t.Fatalf("Print() error = %v", err)
	}
	if !strings.Contains(buf.String(), "hello") {
		t.Errorf("expected 'hello', got %q", buf.String())
	}
}

func TestNewReporter_Table(t *testing.T) {
	r := NewReporter("table")
	if r == nil {
		t.Fatal("NewReporter(\"table\") returned nil")
	}
}

func TestNewReporter_JSON(t *testing.T) {
	r := NewReporter("json")
	if r == nil {
		t.Fatal("NewReporter(\"json\") returned nil")
	}
}

func TestNewReporter_Unknown(t *testing.T) {
	r := NewReporter("unknown")
	if r == nil {
		t.Fatal("NewReporter(\"unknown\") returned nil")
	}
	// Should fall back to table
}

func TestJSONReporter_Print(t *testing.T) {
	r := NewReporter("json")
	result := map[string]any{
		"name":  "test",
		"count": 42,
	}
	var buf bytes.Buffer
	err := r.Print(&buf, result)
	if err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}
	if decoded["name"] != "test" {
		t.Errorf("name = %v, want test", decoded["name"])
	}
	if decoded["count"] != float64(42) {
		t.Errorf("count = %v, want 42", decoded["count"])
	}
}

func TestTableReporter_PrintString(t *testing.T) {
	r := NewReporter("table")
	var buf bytes.Buffer
	err := r.Print(&buf, "scan complete")
	if err != nil {
		t.Fatalf("Print() error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "scan complete") {
		t.Errorf("expected 'scan complete', got %q", out)
	}
}

func TestTableReporter_PrintStruct(t *testing.T) {
	r := NewReporter("table")
	result := struct {
		Name string `json:"name"`
	}{Name: "test"}
	var buf bytes.Buffer
	err := r.Print(&buf, result)
	if err != nil {
		t.Fatalf("Print() error = %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output for struct")
	}
}

func TestFormatNames(t *testing.T) {
	names := FormatNames()
	if len(names) < 2 {
		t.Errorf("expected at least 2 formats, got %d", len(names))
	}
	hasTable := false
	hasJSON := false
	for _, n := range names {
		if n == "table" {
			hasTable = true
		}
		if n == "json" {
			hasJSON = true
		}
	}
	if !hasTable || !hasJSON {
		t.Error("FormatNames() missing table or json")
	}
}
