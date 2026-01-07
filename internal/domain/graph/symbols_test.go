package graph

import "testing"

func TestGraphSymbol_String(t *testing.T) {
	tests := []struct {
		name     string
		symbol   GraphSymbol
		expected string
	}{
		{"empty", SymbolEmpty, "  "},
		{"branch pass", SymbolBranchPass, "│ "},
		{"branch cross", SymbolBranchCross, "│─"},
		{"commit", SymbolCommit, "├ "},
		{"merge commit", SymbolMergeCommit, "├─"},
		{"branch top", SymbolBranchTop, "╮ "},
		{"branch bottom", SymbolBranchBottom, "╯ "},
		{"merge join", SymbolMergeJoin, "┤ "},
		{"octopus", SymbolOctopus, "┬─"},
		{"diverge", SymbolDiverge, "┴─"},
		{"merge cross", SymbolMergeCross, "┼─"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.symbol.String()
			if got != tt.expected {
				t.Errorf("GraphSymbol.String() = %q, want %q", got, tt.expected)
			}
			// Verify all symbols are exactly 2 characters wide
			// Note: Unicode box-drawing characters are 3 bytes but 1 rune
			runeCount := 0
			for range got {
				runeCount++
			}
			if runeCount != 2 {
				t.Errorf("GraphSymbol.String() = %q has %d runes, want 2", got, runeCount)
			}
		})
	}
}

func TestGraphSymbol_String_Unknown(t *testing.T) {
	// Unknown symbol values should default to empty
	unknown := GraphSymbol(999)
	got := unknown.String()
	if got != "  " {
		t.Errorf("Unknown GraphSymbol.String() = %q, want %q", got, "  ")
	}
}

func TestRenderRow(t *testing.T) {
	tests := []struct {
		name     string
		row      Row
		expected string
	}{
		{
			name:     "empty row",
			row:      Row{Symbols: []GraphSymbol{}},
			expected: "",
		},
		{
			name:     "single commit",
			row:      Row{Symbols: []GraphSymbol{SymbolCommit}},
			expected: "├ ",
		},
		{
			name:     "linear continuation",
			row:      Row{Symbols: []GraphSymbol{SymbolCommit, SymbolEmpty}},
			expected: "├   ",
		},
		{
			name:     "merge commit row",
			row:      Row{Symbols: []GraphSymbol{SymbolMergeCommit, SymbolBranchTop}},
			expected: "├─╮ ",
		},
		{
			name:     "branch passing",
			row:      Row{Symbols: []GraphSymbol{SymbolBranchPass, SymbolCommit}},
			expected: "│ ├ ",
		},
		{
			name:     "complex row with crossing",
			row:      Row{Symbols: []GraphSymbol{SymbolMergeCommit, SymbolBranchCross, SymbolBranchTop}},
			expected: "├─│─╮ ",
		},
		{
			name:     "convergence",
			row:      Row{Symbols: []GraphSymbol{SymbolMergeCommit, SymbolBranchBottom}},
			expected: "├─╯ ",
		},
		{
			name:     "octopus merge",
			row:      Row{Symbols: []GraphSymbol{SymbolMergeCommit, SymbolOctopus, SymbolBranchTop}},
			expected: "├─┬─╮ ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderRow(tt.row)
			if got != tt.expected {
				t.Errorf("RenderRow() = %q, want %q", got, tt.expected)
			}
		})
	}
}
