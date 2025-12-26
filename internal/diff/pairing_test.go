package diff

import (
	"testing"

	"github.com/alecthomas/chroma/v2"
	"github.com/oberprah/splice/internal/highlight"
)

// Helper function to create an AlignedLine from a plain text string
func makeAlignedLine(text string) AlignedLine {
	if text == "" {
		return AlignedLine{Tokens: []highlight.Token{}}
	}
	return AlignedLine{
		Tokens: []highlight.Token{
			{Value: text, Type: chroma.Text},
		},
	}
}

// ═══════════════════════════════════════════════════════════
// Tokenization Tests
// ═══════════════════════════════════════════════════════════

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple function call",
			input:    "fmt.Println(name)",
			expected: []string{"fmt", "Println", "name"},
		},
		{
			name:     "comparison operator",
			input:    "x == y",
			expected: []string{"x", "y"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: []string{},
		},
		{
			name:     "punctuation only",
			input:    "...()",
			expected: []string{},
		},
		{
			name:     "single word",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "mixed case",
			input:    "HelloWorld",
			expected: []string{"HelloWorld"},
		},
		{
			name:     "numbers",
			input:    "value123 == 456",
			expected: []string{"value123", "456"},
		},
		{
			name:     "underscores",
			input:    "some_variable_name",
			expected: []string{"some", "variable", "name"},
		},
		{
			name:     "complex expression",
			input:    "if (user.name != null) {",
			expected: []string{"if", "user", "name", "null"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tokenize(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("tokenize(%q) length = %d, want %d", tt.input, len(result), len(tt.expected))
				t.Errorf("  got: %v", result)
				t.Errorf("  want: %v", tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("tokenize(%q)[%d] = %q, want %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// ═══════════════════════════════════════════════════════════
// Dice Similarity Tests
// ═══════════════════════════════════════════════════════════

func TestDiceSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		line1    string
		line2    string
		expected float64
	}{
		{
			name:     "identical lines",
			line1:    "fmt.Println(name)",
			line2:    "fmt.Println(name)",
			expected: 1.0,
		},
		{
			name:     "no overlap",
			line1:    "hello world",
			line2:    "foo bar",
			expected: 0.0,
		},
		{
			name:     "partial overlap",
			line1:    "fmt.Println(name)",
			line2:    "fmt.Println(fullName)",
			expected: 0.6666666666666666, // tokens1: [fmt, Println, name], tokens2: [fmt, Println, fullName], intersection: 2, dice: 2*2/(3+3) = 4/6
		},
		{
			name:     "both empty",
			line1:    "",
			line2:    "",
			expected: 0.0, // Empty lines should not match
		},
		{
			name:     "one empty",
			line1:    "hello",
			line2:    "",
			expected: 0.0,
		},
		{
			name:     "whitespace only (both)",
			line1:    "   ",
			line2:    "  ",
			expected: 0.0,
		},
		{
			name:     "different cases, same word",
			line1:    "Hello",
			line2:    "Hello",
			expected: 1.0,
		},
		{
			name:     "word added",
			line1:    "Hello",
			line2:    "Hello World",
			expected: 0.6666666666666666, // 1 common token out of 3 total: 2*1/(1+2) = 2/3
		},
		{
			name:     "word removed",
			line1:    "Hello World",
			line2:    "Hello",
			expected: 0.6666666666666666, // 1 common token out of 3 total: 2*1/(2+1) = 2/3
		},
		{
			name:     "word changed",
			line1:    "Hello World",
			line2:    "Hello Universe",
			expected: 0.5, // 1 common token (Hello) out of 4 total: 2*1/(2+2) = 2/4 = 0.5
		},
		{
			name:     "punctuation differences",
			line1:    "x == y",
			line2:    "x === y",
			expected: 1.0, // Both tokenize to ["x", "y"]
		},
		{
			name:     "repeated tokens",
			line1:    "x + x + x",
			line2:    "x + x",
			expected: 0.8, // tokens1 = [x, x, x], tokens2 = [x, x], intersection = min(3,2) = 2, dice = 2*2/(3+2) = 4/5 = 0.8
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line1 := makeAlignedLine(tt.line1)
			line2 := makeAlignedLine(tt.line2)
			result := diceSimilarity(&line1, &line2)

			// Use a small epsilon for float comparison
			epsilon := 0.0001
			if result < tt.expected-epsilon || result > tt.expected+epsilon {
				t.Errorf("diceSimilarity(%q, %q) = %f, want %f", tt.line1, tt.line2, result, tt.expected)
			}
		})
	}
}

// ═══════════════════════════════════════════════════════════
// Pairing Algorithm Tests
// ═══════════════════════════════════════════════════════════

func TestPairLines_EmptyInputs(t *testing.T) {
	t.Run("both empty", func(t *testing.T) {
		pairs, unpaired1, unpaired2 := pairLines([]AlignedLine{}, []AlignedLine{})
		if len(pairs) != 0 {
			t.Errorf("expected no pairs, got %d", len(pairs))
		}
		if len(unpaired1) != 0 {
			t.Errorf("expected no unpaired removed lines, got %d", len(unpaired1))
		}
		if len(unpaired2) != 0 {
			t.Errorf("expected no unpaired added lines, got %d", len(unpaired2))
		}
	})

	t.Run("only removed", func(t *testing.T) {
		removed := []AlignedLine{
			makeAlignedLine("line 1"),
			makeAlignedLine("line 2"),
		}
		pairs, unpaired1, unpaired2 := pairLines(removed, []AlignedLine{})
		if len(pairs) != 0 {
			t.Errorf("expected no pairs, got %d", len(pairs))
		}
		if len(unpaired1) != 2 {
			t.Errorf("expected 2 unpaired removed lines, got %d", len(unpaired1))
		}
		if len(unpaired2) != 0 {
			t.Errorf("expected no unpaired added lines, got %d", len(unpaired2))
		}
	})

	t.Run("only added", func(t *testing.T) {
		added := []AlignedLine{
			makeAlignedLine("line 1"),
			makeAlignedLine("line 2"),
		}
		pairs, unpaired1, unpaired2 := pairLines([]AlignedLine{}, added)
		if len(pairs) != 0 {
			t.Errorf("expected no pairs, got %d", len(pairs))
		}
		if len(unpaired1) != 0 {
			t.Errorf("expected no unpaired removed lines, got %d", len(unpaired1))
		}
		if len(unpaired2) != 2 {
			t.Errorf("expected 2 unpaired added lines, got %d", len(unpaired2))
		}
	})
}

func TestPairLines_PerfectMatches(t *testing.T) {
	t.Run("identical lines", func(t *testing.T) {
		removed := []AlignedLine{
			makeAlignedLine("func hello() {"),
		}
		added := []AlignedLine{
			makeAlignedLine("func hello() {"),
		}

		pairs, unpaired1, unpaired2 := pairLines(removed, added)
		if len(pairs) != 1 {
			t.Errorf("expected 1 pair, got %d", len(pairs))
		}
		if len(unpaired1) != 0 || len(unpaired2) != 0 {
			t.Errorf("expected no unpaired lines, got %d removed and %d added", len(unpaired1), len(unpaired2))
		}
		if pairs[0][0] != 0 || pairs[0][1] != 0 {
			t.Errorf("expected pair [0, 0], got [%d, %d]", pairs[0][0], pairs[0][1])
		}
	})

	t.Run("multiple similar lines", func(t *testing.T) {
		removed := []AlignedLine{
			makeAlignedLine("fmt.Println(name)"),
			makeAlignedLine("fmt.Println(age)"),
		}
		added := []AlignedLine{
			makeAlignedLine("fmt.Println(fullName)"),
			makeAlignedLine("fmt.Println(userAge)"),
		}

		pairs, unpaired1, unpaired2 := pairLines(removed, added)
		if len(pairs) != 2 {
			t.Errorf("expected 2 pairs, got %d", len(pairs))
		}
		if len(unpaired1) != 0 || len(unpaired2) != 0 {
			t.Errorf("expected no unpaired lines, got %d removed and %d added", len(unpaired1), len(unpaired2))
		}
	})
}

func TestPairLines_NoGoodMatches(t *testing.T) {
	t.Run("completely different lines", func(t *testing.T) {
		removed := []AlignedLine{
			makeAlignedLine("hello world"),
			makeAlignedLine("foo bar"),
		}
		added := []AlignedLine{
			makeAlignedLine("xyz abc"),
			makeAlignedLine("qwe rty"),
		}

		pairs, unpaired1, unpaired2 := pairLines(removed, added)
		if len(pairs) != 0 {
			t.Errorf("expected no pairs (below threshold), got %d", len(pairs))
		}
		if len(unpaired1) != 2 {
			t.Errorf("expected 2 unpaired removed lines, got %d", len(unpaired1))
		}
		if len(unpaired2) != 2 {
			t.Errorf("expected 2 unpaired added lines, got %d", len(unpaired2))
		}
	})
}

func TestPairLines_UnbalancedCounts(t *testing.T) {
	t.Run("more removed than added", func(t *testing.T) {
		removed := []AlignedLine{
			makeAlignedLine("fmt.Println(x)"),
			makeAlignedLine("fmt.Println(y)"),
			makeAlignedLine("fmt.Println(z)"),
		}
		added := []AlignedLine{
			makeAlignedLine("fmt.Println(x)"),
		}

		pairs, unpaired1, unpaired2 := pairLines(removed, added)
		// Only one can be paired (the best match)
		if len(pairs) != 1 {
			t.Errorf("expected 1 pair, got %d", len(pairs))
		}
		if len(unpaired1) != 2 {
			t.Errorf("expected 2 unpaired removed lines, got %d", len(unpaired1))
		}
		if len(unpaired2) != 0 {
			t.Errorf("expected no unpaired added lines, got %d", len(unpaired2))
		}
	})

	t.Run("more added than removed", func(t *testing.T) {
		removed := []AlignedLine{
			makeAlignedLine("fmt.Println(x)"),
		}
		added := []AlignedLine{
			makeAlignedLine("fmt.Println(x)"),
			makeAlignedLine("fmt.Println(y)"),
			makeAlignedLine("fmt.Println(z)"),
		}

		pairs, unpaired1, unpaired2 := pairLines(removed, added)
		// Only one can be paired (the best match)
		if len(pairs) != 1 {
			t.Errorf("expected 1 pair, got %d", len(pairs))
		}
		if len(unpaired1) != 0 {
			t.Errorf("expected no unpaired removed lines, got %d", len(unpaired1))
		}
		if len(unpaired2) != 2 {
			t.Errorf("expected 2 unpaired added lines, got %d", len(unpaired2))
		}
	})
}

func TestPairLines_GreedyMatching(t *testing.T) {
	t.Run("prefers best matches", func(t *testing.T) {
		// Create a scenario where greedy algorithm should pick the best matches
		removed := []AlignedLine{
			makeAlignedLine("fmt.Println(hello)"),      // Best match: added[0]
			makeAlignedLine("something else entirely"), // No good match
		}
		added := []AlignedLine{
			makeAlignedLine("fmt.Println(hello world)"), // Best match: removed[0]
			makeAlignedLine("fmt.Println(goodbye)"),     // Could match removed[0] but worse score
		}

		pairs, _, _ := pairLines(removed, added)

		// Should pair removed[0] with added[0] (highest similarity)
		if len(pairs) != 1 {
			t.Errorf("expected 1 pair, got %d", len(pairs))
		}
		if len(pairs) > 0 && pairs[0][0] != 0 {
			t.Errorf("expected removed[0] to be paired, got removed[%d]", pairs[0][0])
		}
		if len(pairs) > 0 && pairs[0][1] != 0 {
			t.Errorf("expected added[0] to be paired, got added[%d]", pairs[0][1])
		}
	})
}

func TestPairLines_ThresholdFiltering(t *testing.T) {
	t.Run("filters pairs below threshold", func(t *testing.T) {
		// Create lines with similarity score just below 0.5 threshold
		removed := []AlignedLine{
			makeAlignedLine("Hello World"),
		}
		added := []AlignedLine{
			makeAlignedLine("Hello Universe"), // similarity = 0.5 (1 common token out of 4)
		}

		pairs, unpaired1, unpaired2 := pairLines(removed, added)

		// With threshold of 0.5, this should match (>= threshold)
		if len(pairs) != 1 {
			t.Errorf("expected 1 pair (at threshold), got %d", len(pairs))
		}
		if len(unpaired1) != 0 || len(unpaired2) != 0 {
			t.Errorf("expected no unpaired lines")
		}
	})

	t.Run("filters pairs just below threshold", func(t *testing.T) {
		// Create lines with similarity score just below 0.5 threshold
		removed := []AlignedLine{
			makeAlignedLine("Hello World Foo"),
		}
		added := []AlignedLine{
			makeAlignedLine("Hello Universe Bar"), // similarity = 0.333 (1 common token out of 6)
		}

		pairs, unpaired1, unpaired2 := pairLines(removed, added)

		// With threshold of 0.5, this should not match
		if len(pairs) != 0 {
			t.Errorf("expected no pairs (below threshold), got %d", len(pairs))
		}
		if len(unpaired1) != 1 || len(unpaired2) != 1 {
			t.Errorf("expected both lines unpaired")
		}
	})
}

func TestPairLines_MultipleTokensInLine(t *testing.T) {
	t.Run("lines with multiple syntax tokens", func(t *testing.T) {
		// Simulate lines with actual syntax highlighting tokens
		removed := []AlignedLine{
			{
				Tokens: []highlight.Token{
					{Value: "fmt", Type: chroma.Keyword},
					{Value: ".", Type: chroma.Text},
					{Value: "Println", Type: chroma.Name},
					{Value: "(", Type: chroma.Text},
					{Value: "name", Type: chroma.Name},
					{Value: ")", Type: chroma.Text},
				},
			},
		}
		added := []AlignedLine{
			{
				Tokens: []highlight.Token{
					{Value: "fmt", Type: chroma.Keyword},
					{Value: ".", Type: chroma.Text},
					{Value: "Println", Type: chroma.Name},
					{Value: "(", Type: chroma.Text},
					{Value: "fullName", Type: chroma.Name},
					{Value: ")", Type: chroma.Text},
				},
			},
		}

		pairs, unpaired1, unpaired2 := pairLines(removed, added)

		// Should successfully pair (high similarity: fmt.Println vs fmt.Println)
		if len(pairs) != 1 {
			t.Errorf("expected 1 pair, got %d", len(pairs))
		}
		if len(unpaired1) != 0 || len(unpaired2) != 0 {
			t.Errorf("expected no unpaired lines")
		}
	})
}
