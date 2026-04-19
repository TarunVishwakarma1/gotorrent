package parser

import (
	"testing"
)

// ---- Decode tests ----

func TestDecode_Integer(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"i42e", 42},
		{"i0e", 0},
		{"i-1e", -1},
		{"i1000e", 1000},
	}
	for _, tt := range tests {
		got := Decode(tt.input)
		v, ok := got.(int)
		if !ok {
			t.Errorf("Decode(%q): expected int, got %T", tt.input, got)
			continue
		}
		if v != tt.expected {
			t.Errorf("Decode(%q): expected %d, got %d", tt.input, tt.expected, v)
		}
	}
}

func TestDecode_String(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"4:spam", "spam"},
		{"3:foo", "foo"},
		{"0:", ""},
		{"11:hello world", "hello world"},
	}
	for _, tt := range tests {
		got := Decode(tt.input)
		v, ok := got.(string)
		if !ok {
			t.Errorf("Decode(%q): expected string, got %T", tt.input, got)
			continue
		}
		if v != tt.expected {
			t.Errorf("Decode(%q): expected %q, got %q", tt.input, tt.expected, v)
		}
	}
}

func TestDecode_List(t *testing.T) {
	got := Decode("l4:spami42ee")
	list, ok := got.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", got)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(list))
	}
	if list[0].(string) != "spam" {
		t.Errorf("expected list[0]=%q, got %q", "spam", list[0])
	}
	if list[1].(int) != 42 {
		t.Errorf("expected list[1]=%d, got %v", 42, list[1])
	}
}

func TestDecode_EmptyList(t *testing.T) {
	got := Decode("le")
	list, ok := got.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", got)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %v", list)
	}
}

func TestDecode_Dict(t *testing.T) {
	got := Decode("d3:foo3:bare")
	d, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", got)
	}
	if d["foo"].(string) != "bar" {
		t.Errorf("expected d[foo]=%q, got %v", "bar", d["foo"])
	}
}

func TestDecode_NestedDict(t *testing.T) {
	// d8:announce35:http://tracker.example.com/announce4:infod6:lengthi1000eee
	input := "d8:announce35:http://tracker.example.com/announce4:infod6:lengthi1000eee"
	got := Decode(input)
	d, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", got)
	}
	if d["announce"].(string) != "http://tracker.example.com/announce" {
		t.Errorf("unexpected announce: %v", d["announce"])
	}
	info, ok := d["info"].(map[string]any)
	if !ok {
		t.Fatalf("expected info to be map, got %T", d["info"])
	}
	if info["length"].(int) != 1000 {
		t.Errorf("expected info.length=1000, got %v", info["length"])
	}
}

func TestDecode_EmptyInput(t *testing.T) {
	got := Decode("")
	if got != nil {
		t.Errorf("expected nil for empty input, got %v", got)
	}
}

func TestDecode_ListWithMultipleTypes(t *testing.T) {
	// list containing string and nested list
	got := Decode("l3:fool3:baree")
	list, ok := got.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", got)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(list))
	}
	if list[0].(string) != "foo" {
		t.Errorf("expected list[0]=%q, got %v", "foo", list[0])
	}
	inner, ok := list[1].([]any)
	if !ok {
		t.Fatalf("expected inner []any, got %T", list[1])
	}
	if inner[0].(string) != "bar" {
		t.Errorf("expected inner[0]=%q, got %v", "bar", inner[0])
	}
}

// ---- Encode tests ----

func TestEncode_String(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"spam", "4:spam"},
		{"foo", "3:foo"},
		{"", "0:"},
		{"hello world", "11:hello world"},
	}
	for _, tt := range tests {
		got := Encode(tt.input)
		if got != tt.expected {
			t.Errorf("Encode(%q): expected %q, got %q", tt.input, tt.expected, got)
		}
	}
}

func TestEncode_Integer(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{42, "i42e"},
		{0, "i0e"},
		{-1, "i-1e"},
		{1000, "i1000e"},
	}
	for _, tt := range tests {
		got := Encode(tt.input)
		if got != tt.expected {
			t.Errorf("Encode(%d): expected %q, got %q", tt.input, tt.expected, got)
		}
	}
}

func TestEncode_List(t *testing.T) {
	input := []any{"spam", 42}
	got := Encode(input)
	expected := "l4:spami42ee"
	if got != expected {
		t.Errorf("Encode(list): expected %q, got %q", expected, got)
	}
}

func TestEncode_EmptyList(t *testing.T) {
	got := Encode([]any{})
	expected := "le"
	if got != expected {
		t.Errorf("Encode(emptylist): expected %q, got %q", expected, got)
	}
}

func TestEncode_Dict(t *testing.T) {
	input := map[string]any{"foo": "bar"}
	got := Encode(input)
	expected := "d3:foo3:bare"
	if got != expected {
		t.Errorf("Encode(dict): expected %q, got %q", expected, got)
	}
}

func TestEncode_DictKeysSorted(t *testing.T) {
	// Keys must be sorted lexicographically per bencode spec
	input := map[string]any{
		"zebra": "last",
		"apple": "first",
		"mango": "middle",
	}
	got := Encode(input)
	expected := "d5:apple5:first5:mango6:middle5:zebra4:laste"
	if got != expected {
		t.Errorf("Encode(dict sorted keys): expected %q, got %q", expected, got)
	}
}

func TestEncode_NestedDict(t *testing.T) {
	input := map[string]any{
		"info": map[string]any{
			"length": 1000,
		},
	}
	got := Encode(input)
	expected := "d4:infod6:lengthi1000eee"
	if got != expected {
		t.Errorf("Encode(nested dict): expected %q, got %q", expected, got)
	}
}

func TestEncode_UnsupportedType(t *testing.T) {
	// Unsupported types should produce empty string (no panic)
	got := Encode(3.14)
	if got != "" {
		t.Errorf("Encode(float64): expected empty string, got %q", got)
	}
}

// ---- Round-trip tests ----

func TestDecodeEncode_RoundTrip_Integer(t *testing.T) {
	original := "i42e"
	decoded := Decode(original)
	encoded := Encode(decoded)
	if encoded != original {
		t.Errorf("round-trip integer: expected %q, got %q", original, encoded)
	}
}

func TestDecodeEncode_RoundTrip_String(t *testing.T) {
	original := "4:spam"
	decoded := Decode(original)
	encoded := Encode(decoded)
	if encoded != original {
		t.Errorf("round-trip string: expected %q, got %q", original, encoded)
	}
}

func TestDecodeEncode_RoundTrip_List(t *testing.T) {
	original := "l4:spami42ee"
	decoded := Decode(original)
	encoded := Encode(decoded)
	if encoded != original {
		t.Errorf("round-trip list: expected %q, got %q", original, encoded)
	}
}

func TestDecodeEncode_RoundTrip_Dict(t *testing.T) {
	// Single-key dict to avoid key ordering ambiguity
	original := "d3:foo3:bare"
	decoded := Decode(original)
	encoded := Encode(decoded)
	if encoded != original {
		t.Errorf("round-trip dict: expected %q, got %q", original, encoded)
	}
}

func TestDecodeEncode_RoundTrip_ComplexTorrentDict(t *testing.T) {
	// Simulated minimal torrent-like bencode
	pieces := "AAAAAAAAAAAAAAAAAAAA" // 20 bytes
	original := "d8:announce35:http://tracker.example.com/announce4:infod6:lengthi1000e4:name8:test.iso12:piece lengthi262144e6:pieces20:" + pieces + "ee"
	decoded := Decode(original)
	encoded := Encode(decoded)
	if encoded != original {
		t.Errorf("round-trip complex dict: expected %q, got %q", original, encoded)
	}
}