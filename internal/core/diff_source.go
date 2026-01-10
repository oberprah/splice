package core

// DiffSource is a sum type representing any source of a git diff.
// It can be either a range of commits or uncommitted changes.
//
// This is a sealed interface - only the concrete types below can implement it.
// Use a type switch to handle all cases when working with DiffSource values.
type DiffSource interface {
	diffSource() // unexported marker method prevents external implementation
}

// CommitRangeDiffSource represents a diff between two commits.
// Start is the older commit, End is the newer commit.
// For single commits, Start and End are the same with Count = 1.
type CommitRangeDiffSource struct {
	Start GitCommit
	End   GitCommit
	Count int // Number of commits in range (1 for single commit)
}

// UncommittedChangesDiffSource represents a diff of uncommitted changes
// in the working tree and/or staging area.
type UncommittedChangesDiffSource struct {
	Type UncommittedType
}

// UncommittedType specifies which uncommitted changes to include in the diff.
type UncommittedType int

const (
	UncommittedTypeUnstaged UncommittedType = iota // Working tree vs index (git diff)
	UncommittedTypeStaged                          // Index vs HEAD (git diff --staged)
	UncommittedTypeAll                             // Working tree vs HEAD (git diff HEAD)
)

// Marker method implementations - seal the DiffSource interface
func (CommitRangeDiffSource) diffSource()        {}
func (UncommittedChangesDiffSource) diffSource() {}
