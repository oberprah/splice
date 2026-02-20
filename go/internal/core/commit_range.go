package core

// CommitRange represents either a single commit or a range of commits.
// Start is always the older commit, End is the newer commit.
// For single commits, Start and End are the same.
type CommitRange struct {
	Start GitCommit
	End   GitCommit
	Count int // Number of commits in the range (1 for single commit)
}

// IsSingleCommit returns true if this range represents a single commit.
func (r CommitRange) IsSingleCommit() bool {
	return r.Count == 1
}

// NewSingleCommitRange creates a CommitRange for a single commit.
func NewSingleCommitRange(commit GitCommit) CommitRange {
	return CommitRange{Start: commit, End: commit, Count: 1}
}

// NewCommitRange creates a CommitRange for multiple commits.
// start should be the older commit, end should be the newer commit.
func NewCommitRange(start, end GitCommit, count int) CommitRange {
	return CommitRange{Start: start, End: end, Count: count}
}

// ToDiffSource converts a CommitRange to a CommitRangeDiffSource.
func (r CommitRange) ToDiffSource() CommitRangeDiffSource {
	return CommitRangeDiffSource(r)
}
