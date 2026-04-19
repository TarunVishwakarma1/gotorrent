package parser

import (
	"reflect"
	"testing"
)

// ---- Decode tests ----

func TestDecodeInteger(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"positive", "i42e", 42},
		{"zero", "i0e", 0},
		{"negative", "i-1e", -1},
		{"large", "i1000000e", 1000000},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Decode(tc.input)
			if got != tc.want {
				t.Errorf("Decode(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestDecodeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "4:spam", "spam"},
		{"empty", "0:", ""},
		{"single char", "1:x", "x"},
		{"with spaces", "5:hello", "hello"},
		{"longer", "11:hello world", "hello world"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Decode(tc.input)
			if got != tc.want {
				t.Errorf("Decode(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestDecodeList(t *testing.T) {
	t.Run("mixed list", func(t *testing.T) {
		got := Decode("l4:spami42ee")
		want := []any{"spam", 42}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Decode list = %v, want %v", got, want)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		got := Decode("le")
		want := []any(nil)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Decode empty list = %v, want %v", got, want)
		}
	})

	t.Run("list of strings", func(t *testing.T) {
		got := Decode("l3:foo3:bare")
		want := []any{"foo", "bar"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Decode list of strings = %v, want %v", got, want)
		}
	})

	t.Run("nested list", func(t *testing.T) {
		got := Decode("lli1eee")
		want := []any{[]any{1}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Decode nested list = %v, want %v", got, want)
		}
	})
}

func TestDecodeDict(t *testing.T) {
	t.Run("simple dict", func(t *testing.T) {
		got := Decode("d3:bar4:spam3:fooi42ee")
		want := map[string]any{"bar": "spam", "foo": 42}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Decode dict = %v, want %v", got, want)
		}
	})

	t.Run("empty dict", func(t *testing.T) {
		got := Decode("de")
		want := map[string]any{}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Decode empty dict = %v, want %v", got, want)
		}
	})

	t.Run("nested dict", func(t *testing.T) {
		got := Decode("d4:infod6:lengthi100eee")
		want := map[string]any{"info": map[string]any{"length": 100}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Decode nested dict = %v, want %v", got, want)
		}
	})
}

func TestDecodeEmpty(t *testing.T) {
	got := Decode("")
	if got != nil {
		t.Errorf("Decode(\"\") = %v, want nil", got)
	}
}

// ---- Encode tests ----

func TestEncodeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "spam", "4:spam"},
		{"empty string", "", "0:"},
		{"single char", "x", "1:x"},
		{"with spaces", "hello world", "11:hello world"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Encode(tc.input)
			if got != tc.want {
				t.Errorf("Encode(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestEncodeInteger(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  string
	}{
		{"positive", 42, "i42e"},
		{"zero", 0, "i0e"},
		{"negative", -1, "i-1e"},
		{"large", 1000000, "i1000000e"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Encode(tc.input)
			if got != tc.want {
				t.Errorf("Encode(%d) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestEncodeList(t *testing.T) {
	t.Run("mixed list", func(t *testing.T) {
		got := Encode([]any{"spam", 42})
		want := "l4:spami42ee"
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

	t.Run("nested list", func(t *testing.T) {
		got := Encode([]any{[]any{1}})
		want := "lli1eee"
		if got != want {
			t.Errorf("Encode(nested list) = %q, want %q", got, want)
		}
	})
}

func TestEncodeDict(t *testing.T) {
	t.Run("keys sorted lexicographically", func(t *testing.T) {
		got := Encode(map[string]any{"foo": 42, "bar": "spam"})
		// keys must be sorted: bar before foo
		want := "d3:bar4:spam3:fooi42ee"
		if got != want {
			t.Errorf("Encode(dict) = %q, want %q", got, want)
		}
	})

	t.Run("empty dict", func(t *testing.T) {
		got := Encode(map[string]any{})
		want := "de"
		if got != want {
			t.Errorf("Encode(empty dict) = %q, want %q", got, want)
		}
	})

	t.Run("nested dict", func(t *testing.T) {
		got := Encode(map[string]any{"info": map[string]any{"length": 100}})
		want := "d4:infod6:lengthi100eee"
		if got != want {
			t.Errorf("Encode(nested dict) = %q, want %q", got, want)
		}
	})
}

func TestEncodeUnknownType(t *testing.T) {
	// Unsupported type should produce empty string
	got := Encode(3.14)
	if got != "" {
		t.Errorf("Encode(float64) = %q, want %q", got, "")
	}
}

// ---- Roundtrip tests ----

func TestDecodeEncodeRoundtrip(t *testing.T) {
	tests := []string{
		"i42e",
		"i0e",
		"4:spam",
		"0:",
		"l4:spami42ee",
		"d3:bar4:spam3:fooi42ee",
		"d4:infod6:lengthi100eee",
		"le",
		"de",
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			decoded := Decode(input)
			encoded := Encode(decoded)
			if encoded != input {
				t.Errorf("roundtrip(%q): got %q", input, encoded)
			}
		})
	}
}

func TestEncodeDecodeRoundtrip(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{"int", 99},
		{"string", "hello"},
		{"list", []any{"a", 1}},
		{"dict", map[string]any{"key": "val", "num": 7}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encoded := Encode(tc.value)
			decoded := Decode(encoded)
			if !reflect.DeepEqual(decoded, tc.value) {
				t.Errorf("Encode then Decode(%v): got %v", tc.value, decoded)
			}
		})
	}
}