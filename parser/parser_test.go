package parser

import (
	"testing"
)

// --- Decode tests ---

func TestDecode_Integer(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"positive integer", "i42e", 42},
		{"zero", "i0e", 0},
		{"negative integer", "i-7e", -7},
		{"large integer", "i1000000e", 1000000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Decode(tt.input)
			v, ok := got.(int)
			if !ok {
				t.Fatalf("expected int, got %T", got)
			}
			if v != tt.want {
				t.Errorf("Decode(%q) = %d, want %d", tt.input, v, tt.want)
			}
		})
	}
}

func TestDecode_String(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"basic string", "4:spam", "spam"},
		{"single char", "1:a", "a"},
		{"string with spaces", "5:hello", "hello"},
		{"longer string", "11:hello world", "hello world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Decode(tt.input)
			v, ok := got.(string)
			if !ok {
				t.Fatalf("expected string, got %T", got)
			}
			if v != tt.want {
				t.Errorf("Decode(%q) = %q, want %q", tt.input, v, tt.want)
			}
		})
	}
}

func TestDecode_List(t *testing.T) {
	t.Run("list of integers", func(t *testing.T) {
		got := Decode("li1ei2ei3ee")
		list, ok := got.([]any)
		if !ok {
			t.Fatalf("expected []any, got %T", got)
		}
		if len(list) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(list))
		}
		expected := []int{1, 2, 3}
		for i, want := range expected {
			if list[i].(int) != want {
				t.Errorf("list[%d] = %v, want %d", i, list[i], want)
			}
		}
	})

	t.Run("list of strings", func(t *testing.T) {
		got := Decode("l4:spam4:eggse")
		list, ok := got.([]any)
		if !ok {
			t.Fatalf("expected []any, got %T", got)
		}
		if len(list) != 2 {
			t.Fatalf("expected 2 elements, got %d", len(list))
		}
		if list[0].(string) != "spam" {
			t.Errorf("list[0] = %q, want %q", list[0], "spam")
		}
		if list[1].(string) != "eggs" {
			t.Errorf("list[1] = %q, want %q", list[1], "eggs")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		got := Decode("le")
		list, ok := got.([]any)
		if !ok {
			t.Fatalf("expected []any, got %T", got)
		}
		if len(list) != 0 {
			t.Errorf("expected empty list, got %v", list)
		}
	})

	t.Run("nested list", func(t *testing.T) {
		got := Decode("lli1ei2eel3:fooe")
		outer, ok := got.([]any)
		if !ok {
			t.Fatalf("expected []any, got %T", got)
		}
		if len(outer) != 2 {
			t.Fatalf("expected 2 elements, got %d", len(outer))
		}
		inner, ok := outer[0].([]any)
		if !ok {
			t.Fatalf("expected inner list, got %T", outer[0])
		}
		if len(inner) != 2 {
			t.Errorf("expected 2 inner elements, got %d", len(inner))
		}
	})
}

func TestDecode_Dict(t *testing.T) {
	t.Run("basic dict", func(t *testing.T) {
		got := Decode("d3:bar4:spam3:fooi42ee")
		dict, ok := got.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", got)
		}
		if dict["bar"].(string) != "spam" {
			t.Errorf("dict[bar] = %v, want %q", dict["bar"], "spam")
		}
		if dict["foo"].(int) != 42 {
			t.Errorf("dict[foo] = %v, want 42", dict["foo"])
		}
	})

	t.Run("nested dict", func(t *testing.T) {
		got := Decode("d5:outerdi1ei2eee")
		dict, ok := got.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", got)
		}
		inner, ok := dict["outer"].(map[string]any)
		if !ok {
			t.Fatalf("expected nested dict, got %T", dict["outer"])
		}
		if inner["1"].(int) != 2 {
			t.Errorf("inner[1] = %v, want 2", inner["1"])
		}
	})

	t.Run("torrent-like dict", func(t *testing.T) {
		input := "d8:announce35:http://tracker.example.com/announce4:infod6:lengthi1000e4:name8:test.iso12:piece lengthi256e6:pieces20:AAAAAAAAAAAAAAAAAAAAee"
		got := Decode(input)
		dict, ok := got.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", got)
		}
		if dict["announce"].(string) != "http://tracker.example.com/announce" {
			t.Errorf("unexpected announce: %v", dict["announce"])
		}
		info, ok := dict["info"].(map[string]any)
		if !ok {
			t.Fatalf("expected info dict")
		}
		if info["length"].(int) != 1000 {
			t.Errorf("info[length] = %v, want 1000", info["length"])
		}
		if info["name"].(string) != "test.iso" {
			t.Errorf("info[name] = %v, want test.iso", info["name"])
		}
	})
}

func TestDecode_EmptyInput(t *testing.T) {
	got := Decode("")
	if got != nil {
		t.Errorf("Decode(\"\") = %v, want nil", got)
	}
}

// --- Encode tests ---

func TestEncode_String(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"basic string", "spam", "4:spam"},
		{"single char", "a", "1:a"},
		{"hello", "hello", "5:hello"},
		{"string with colon", "a:b", "3:a:b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Encode(tt.input)
			if got != tt.want {
				t.Errorf("Encode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEncode_EmptyString(t *testing.T) {
	got := Encode("")
	if got != "0:" {
		t.Errorf("Encode(\"\") = %q, want %q", got, "0:")
	}
}

func TestEncode_Integer(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  string
	}{
		{"positive", 42, "i42e"},
		{"zero", 0, "i0e"},
		{"negative", -7, "i-7e"},
		{"large", 1000000, "i1000000e"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Encode(tt.input)
			if got != tt.want {
				t.Errorf("Encode(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEncode_List(t *testing.T) {
	t.Run("list of strings", func(t *testing.T) {
		got := Encode([]any{"spam", "eggs"})
		want := "l4:spam4:eggse"
		if got != want {
			t.Errorf("Encode(list) = %q, want %q", got, want)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		got := Encode([]any{})
		want := "le"
		if got != want {
			t.Errorf("Encode(empty list) = %q, want %q", got, want)
		}
	})

	t.Run("list of integers", func(t *testing.T) {
		got := Encode([]any{1, 2, 3})
		want := "li1ei2ei3ee"
		if got != want {
			t.Errorf("Encode(int list) = %q, want %q", got, want)
		}
	})

	t.Run("mixed list", func(t *testing.T) {
		got := Encode([]any{"foo", 42})
		want := "l3:fooi42ee"
		if got != want {
			t.Errorf("Encode(mixed list) = %q, want %q", got, want)
		}
	})
}

func TestEncode_Dict(t *testing.T) {
	t.Run("basic dict", func(t *testing.T) {
		// Keys must be sorted: bar < foo
		got := Encode(map[string]any{"foo": "bar", "baz": 42})
		want := "d3:bazi42e3:foo3:bare"
		if got != want {
			t.Errorf("Encode(dict) = %q, want %q", got, want)
		}
	})

	t.Run("keys are sorted lexicographically", func(t *testing.T) {
		got := Encode(map[string]any{"z": 1, "a": 2, "m": 3})
		want := "d1:ai2e1:mi3e1:zi1ee"
		if got != want {
			t.Errorf("Encode(dict with many keys) = %q, want %q", got, want)
		}
	})

	t.Run("empty dict", func(t *testing.T) {
		got := Encode(map[string]any{})
		want := "de"
		if got != want {
			t.Errorf("Encode(empty dict) = %q, want %q", got, want)
		}
	})
}

func TestEncode_UnknownType(t *testing.T) {
	// Unsupported type should return empty string
	got := Encode(3.14)
	if got != "" {
		t.Errorf("Encode(float) = %q, want empty string", got)
	}
}

// --- Round-trip tests: Decode then Encode ---

func TestRoundTrip_Integer(t *testing.T) {
	original := "i12345e"
	decoded := Decode(original)
	reencoded := Encode(decoded)
	if reencoded != original {
		t.Errorf("round-trip integer: got %q, want %q", reencoded, original)
	}
}

func TestRoundTrip_String(t *testing.T) {
	original := "5:hello"
	decoded := Decode(original)
	reencoded := Encode(decoded)
	if reencoded != original {
		t.Errorf("round-trip string: got %q, want %q", reencoded, original)
	}
}

func TestRoundTrip_List(t *testing.T) {
	original := "l4:spam4:eggse"
	decoded := Decode(original)
	reencoded := Encode(decoded)
	if reencoded != original {
		t.Errorf("round-trip list: got %q, want %q", reencoded, original)
	}
}

func TestRoundTrip_Dict(t *testing.T) {
	// Keys already in sorted order: bar < foo
	original := "d3:bar4:spam3:fooi42ee"
	decoded := Decode(original)
	reencoded := Encode(decoded)
	if reencoded != original {
		t.Errorf("round-trip dict: got %q, want %q", reencoded, original)
	}
}

func TestRoundTrip_NestedDict(t *testing.T) {
	// Simulates a torrent info dict structure with keys already sorted
	original := "d6:lengthi1000e4:name8:test.iso12:piece lengthi256e6:pieces20:AAAAAAAAAAAAAAAAAAAAe"
	decoded := Decode(original)
	reencoded := Encode(decoded)
	if reencoded != original {
		t.Errorf("round-trip nested dict: got %q, want %q", reencoded, original)
	}
}