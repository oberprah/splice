package graph

// symbolStrings maps GraphSymbol values to their 2-character string representation.
var symbolStrings = map[GraphSymbol]string{
	SymbolEmpty:        "  ",
	SymbolBranchPass:   "│ ",
	SymbolBranchCross:  "│─",
	SymbolCommit:       "├ ",
	SymbolMergeCommit:  "├─",
	SymbolBranchTop:    "╮ ",
	SymbolBranchBottom: "╯ ",
	SymbolMergeJoin:    "┤ ",
	SymbolOctopus:      "┬─",
	SymbolDiverge:      "┴─",
	SymbolMergeCross:   "┼─",
}

// String returns the 2-character string representation of the symbol.
func (s GraphSymbol) String() string {
	if str, ok := symbolStrings[s]; ok {
		return str
	}
	return "  " // Default to empty for unknown symbols
}

// RenderRow renders a Row to its string representation.
// Each symbol in the row is exactly 2 characters wide.
func RenderRow(row Row) string {
	result := ""
	for _, sym := range row.Symbols {
		result += sym.String()
	}
	return result
}
