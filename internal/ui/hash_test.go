package ui

import "testing"

func TestToShortHash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"full hash", "a4c3a8a7b2c1d3e4f5g6h7i8j9k0l1m2n3o4p5q6", "a4c3a8a"},
		{"empty string", "", ""},
		{"short hash", "abc", "abc"},
		{"exactly 7 chars", "1234567", "1234567"},
		{"8 chars", "12345678", "1234567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToShortHash(tt.input)
			if result != tt.expected {
				t.Errorf("ToShortHash(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}
