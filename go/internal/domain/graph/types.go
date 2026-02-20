package graph

// Commit is a minimal commit representation for graph layout computation.
// It contains only the data needed for graph computation, not full git metadata.
type Commit struct {
	Hash    string   // Commit hash (can be abbreviated)
	Parents []string // Parent commit hashes (first parent is primary branch)
}

// Layout holds the computed graph layout for a sequence of commits.
// Each row corresponds to a commit in the same order as the input.
type Layout struct {
	Rows []Row
}

// Row represents the graph symbols for a single commit line.
type Row struct {
	Symbols []GraphSymbol
}

// GraphSymbol represents a single cell in the graph.
// Each symbol renders as exactly 2 characters.
type GraphSymbol int

const (
	// SymbolEmpty represents an empty cell: "  "
	SymbolEmpty GraphSymbol = iota
	// SymbolBranchPass represents a branch line passing through: "│ "
	SymbolBranchPass
	// SymbolBranchCross represents a branch line crossed by merge line: "│─"
	SymbolBranchCross
	// SymbolCommit represents a commit node: "├ "
	SymbolCommit
	// SymbolMergeCommit represents a commit starting a merge line: "├─"
	SymbolMergeCommit
	// SymbolBranchTop represents top of a feature branch (merge point): "╮ "
	SymbolBranchTop
	// SymbolBranchBottom represents bottom of a feature branch (common ancestor): "╯ "
	SymbolBranchBottom
	// SymbolMergeJoin represents merge line joining into existing branch: "┤ "
	SymbolMergeJoin
	// SymbolOctopus represents octopus merge - new branch goes down: "┬─"
	SymbolOctopus
	// SymbolDiverge represents branches diverging - branch goes up: "┴─"
	SymbolDiverge
	// SymbolMergeCross represents merge line crossing a branch: "┼─"
	SymbolMergeCross
)
