package parser

import (
	"testing"
)

// ---- Decode tests ----

func TestDecodeInteger(t *testing.T) {
	result := Decode("i42e")
	if result != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestDecodeNegativeInteger(t *testing.T) {
	result := Decode("i-7e")
	if result != -7 {
		t.Errorf("expected -7, got %v", result)
	}
}

func TestDecodeZeroInteger(t *testing.T) {
	result := Decode("i0e")
	if result != 0 {
		t.Errorf("expected 0, got %v", result)
	}
}

func TestDecodeString(t *testing.T) {
	result := Decode("4:spam")
	if result != "spam" {
		t.Errorf("expected \"spam\", got %v", result)
	}
}

func TestDecodeEmptyString(t *testing.T) {
	result := Decode("0:")
	if result != "" {
		t.Errorf("expected empty string, got %v", result)
	}
}

func TestDecodeStringWithColon(t *testing.T) {
	result := Decode("5:a:b:c")
	// length 5 → reads "a:b:c"
	if result != "a:b:c" {
		t.Errorf("expected \"a:b:c\", got %v", result)
	}
}

func TestDecodeList(t *testing.T) {
	result := Decode("li1ei2ei3ee")
	list, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(list))
	}
	expected := []int{1, 2, 3}
	for idx, exp := range expected {
		if list[idx] != exp {
			t.Errorf("list[%d]: expected %d, got %v", idx, exp, list[idx])
		}
	}
}

func TestDecodeEmptyList(t *testing.T) {
	result := Decode("le")
	list, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got length %d", len(list))
	}
}

func TestDecodeDict(t *testing.T) {
	result := Decode("d3:fooi99ee")
	d, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result)
	}
	if d["foo"] != 99 {
		t.Errorf("expected d[\"foo\"]=99, got %v", d["foo"])
	}
}

func TestDecodeEmptyDict(t *testing.T) {
	result := Decode("de")
	d, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result)
	}
	if len(d) != 0 {
		t.Errorf("expected empty dict, got %d keys", len(d))
	}
}

func TestDecodeNestedDict(t *testing.T) {
	// {"outer": {"inner": 5}}
	encoded := "d5:outerd5:inneri5eee"
	result := Decode(encoded)
	outer, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result)
	}
	inner, ok := outer["outer"].(map[string]any)
	if !ok {
		t.Fatalf("expected inner map, got %T", outer["outer"])
	}
	if inner["inner"] != 5 {
		t.Errorf("expected inner[\"inner\"]=5, got %v", inner["inner"])
	}
}

func TestDecodeListOfStrings(t *testing.T) {
	result := Decode("l3:foo3:bare")
	list, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}
	if len(list) != 2 || list[0] != "foo" || list[1] != "bar" {
		t.Errorf("expected [\"foo\", \"bar\"], got %v", list)
	}
}

func TestDecodeEmptyInput(t *testing.T) {
	result := Decode("")
	if result != nil {
		t.Errorf("expected nil for empty input, got %v", result)
	}
}

func TestDecodeTorrentLike(t *testing.T) {
	// Simulate a minimal torrent-like structure
	encoded := "d8:announce35:http://tracker.example.com/announce4:infod6:lengthi1000e4:name8:test.iso12:piece lengthi256e6:pieces20:AAAAAAAAAAAAAAAAAAAAee"
	result := Decode(encoded)
	d, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if d["announce"] != "http://tracker.example.com/announce" {
		t.Errorf("unexpected announce: %v", d["announce"])
	}
	info, ok := d["info"].(map[string]any)
	if !ok {
		t.Fatalf("expected info map, got %T", d["info"])
	}
	if info["length"] != 1000 {
		t.Errorf("expected length=1000, got %v", info["length"])
	}
	if info["name"] != "test.iso" {
		t.Errorf("expected name=test.iso, got %v", info["name"])
	}
}

// ---- Encode tests ----

func TestEncodeString(t *testing.T) {
	result := Encode("spam")
	expected := "4:spam"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEncodeEmptyString(t *testing.T) {
	result := Encode("")
	expected := "0:"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEncodeInteger(t *testing.T) {
	result := Encode(42)
	expected := "i42e"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEncodeNegativeInteger(t *testing.T) {
	result := Encode(-7)
	expected := "i-7e"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEncodeZero(t *testing.T) {
	result := Encode(0)
	expected := "i0e"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEncodeList(t *testing.T) {
	result := Encode([]any{1, 2, 3})
	expected := "li1ei2ei3ee"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEncodeEmptyList(t *testing.T) {
	result := Encode([]any{})
	expected := "le"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEncodeDict(t *testing.T) {
	result := Encode(map[string]any{"foo": 99})
	expected := "d3:fooi99ee"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEncodeDictKeySorting(t *testing.T) {
	// Keys must be sorted lexicographically per bencode spec
	result := Encode(map[string]any{
		"zebra": "last",
		"apple": "first",
		"mango": "middle",
	})
	expected := "d5:apple5:first5:mango6:middle5:zebra4:laste"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEncodeEmptyDict(t *testing.T) {
	result := Encode(map[string]any{})
	expected := "de"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEncodeNestedDict(t *testing.T) {
	result := Encode(map[string]any{
		"outer": map[string]any{
			"inner": 5,
		},
	})
	expected := "d5:outerd5:inneri5eee"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEncodeUnknownTypeReturnsEmpty(t *testing.T) {
	result := Encode(3.14)
	if result != "" {
		t.Errorf("expected empty string for unknown type, got %q", result)
	}
}

// ---- Roundtrip tests ----

func TestRoundtripInteger(t *testing.T) {
	original := "i12345e"
	encoded := Encode(Decode(original))
	if encoded != original {
		t.Errorf("roundtrip failed: expected %q, got %q", original, encoded)
	}
}

func TestRoundtripString(t *testing.T) {
	original := "11:hello world"
	encoded := Encode(Decode(original))
	if encoded != original {
		t.Errorf("roundtrip failed: expected %q, got %q", original, encoded)
	}
}

func TestRoundtripList(t *testing.T) {
	original := "li1e4:teste"
	encoded := Encode(Decode(original))
	if encoded != original {
		t.Errorf("roundtrip failed: expected %q, got %q", original, encoded)
	}
}

func TestRoundtripDict(t *testing.T) {
	// bencode dicts are ordered by key, so use already-sorted keys
	original := "d3:bari42e3:foo4:teste"
	encoded := Encode(Decode(original))
	if encoded != original {
		t.Errorf("roundtrip failed: expected %q, got %q", original, encoded)
	}
}

func TestRoundtripNestedStructure(t *testing.T) {
	original := "d4:infod6:lengthi1000e4:name8:test.isoeee"
	// After decode/encode keys are sorted - "info" is the only key, inner keys: "length" < "name"
	reencoded := Encode(Decode(original))
	// Re-decode both and compare semantically
	orig := Decode(original).(map[string]any)
	reenc := Decode(reencoded).(map[string]any)
	origInfo := orig["info"].(map[string]any)
	reencInfo := reenc["info"].(map[string]any)
	if origInfo["length"] != reencInfo["length"] {
		t.Errorf("roundtrip length mismatch: %v vs %v", origInfo["length"], reencInfo["length"])
	}
	if origInfo["name"] != reencInfo["name"] {
		t.Errorf("roundtrip name mismatch: %v vs %v", origInfo["name"], reencInfo["name"])
	}
}