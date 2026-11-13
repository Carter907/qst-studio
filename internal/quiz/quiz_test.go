package quiz

import "testing"

func TestNormalizeChoiceKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"  Hello World  ", "hello world"},
		{"`variable_name`", "variable_name"},
		{"$O(n \\log n)$", "o(n \\log n)"},
		{"'quoted'", "quoted"},
		{"multiple   spaces   between", "multiple spaces between"},
	}

	for _, tt := range tests {
		got := NormalizeChoiceKey(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeChoiceKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseLineRange(t *testing.T) {
	start, end, explicit := ParseLineRange("5-10", 20)
	if !explicit || start != 5 || end != 10 {
		t.Fatalf("expected explicit range 5-10, got %d-%d (explicit: %v)", start, end, explicit)
	}

	start, end, explicit = ParseLineRange("ALL", 20)
	if explicit || start != 1 || end != 20 {
		t.Fatalf("expected fallback range 1-20, got %d-%d (explicit: %v)", start, end, explicit)
	}
}
