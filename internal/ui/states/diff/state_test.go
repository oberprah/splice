package diff

import (
	"testing"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/diff"
)

func TestFindFirstHunkSegmentIndex(t *testing.T) {
	tests := []struct {
		name     string
		segments []diff.Segment
		want     int
	}{
		{
			name:     "empty segments",
			segments: []diff.Segment{},
			want:     0,
		},
		{
			name:     "nil segments",
			segments: nil,
			want:     0,
		},
		{
			name: "only unchanged segments",
			segments: []diff.Segment{
				diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 5},
				diff.UnchangedSegment{LeftStart: 5, RightStart: 5, Count: 10},
			},
			want: 0,
		},
		{
			name: "hunk at beginning",
			segments: []diff.Segment{
				diff.HunkSegment{
					LeftLines:  []diff.HunkLine{{SourceIdx: 0, Type: diff.HunkLineRemoved}},
					RightLines: []diff.HunkLine{{SourceIdx: 0, Type: diff.HunkLineAdded}},
				},
				diff.UnchangedSegment{LeftStart: 1, RightStart: 1, Count: 10},
			},
			want: 0,
		},
		{
			name: "hunk after unchanged",
			segments: []diff.Segment{
				diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 5},
				diff.HunkSegment{
					LeftLines:  []diff.HunkLine{{SourceIdx: 5, Type: diff.HunkLineRemoved}},
					RightLines: []diff.HunkLine{{SourceIdx: 5, Type: diff.HunkLineAdded}},
				},
				diff.UnchangedSegment{LeftStart: 6, RightStart: 6, Count: 10},
			},
			want: 1,
		},
		{
			name: "multiple hunks returns first",
			segments: []diff.Segment{
				diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 5},
				diff.HunkSegment{
					LeftLines:  []diff.HunkLine{{SourceIdx: 5, Type: diff.HunkLineRemoved}},
					RightLines: []diff.HunkLine{},
				},
				diff.UnchangedSegment{LeftStart: 6, RightStart: 5, Count: 10},
				diff.HunkSegment{
					LeftLines:  []diff.HunkLine{},
					RightLines: []diff.HunkLine{{SourceIdx: 15, Type: diff.HunkLineAdded}},
				},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findFirstHunkSegmentIndex(tt.segments)
			if got != tt.want {
				t.Errorf("findFirstHunkSegmentIndex() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNew_SegmentBasedPosition(t *testing.T) {
	tests := []struct {
		name           string
		diff           *diff.AlignedFileDiff
		wantSegmentIdx int
	}{
		{
			name:           "nil diff",
			diff:           nil,
			wantSegmentIdx: 0,
		},
		{
			name: "diff with no segments",
			diff: &diff.AlignedFileDiff{
				Segments: []diff.Segment{},
			},
			wantSegmentIdx: 0,
		},
		{
			name: "diff with only unchanged segments",
			diff: &diff.AlignedFileDiff{
				Segments: []diff.Segment{
					diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 10},
				},
			},
			wantSegmentIdx: 0,
		},
		{
			name: "diff with hunk at start",
			diff: &diff.AlignedFileDiff{
				Segments: []diff.Segment{
					diff.HunkSegment{
						LeftLines:  []diff.HunkLine{{SourceIdx: 0, Type: diff.HunkLineRemoved}},
						RightLines: []diff.HunkLine{},
					},
					diff.UnchangedSegment{LeftStart: 1, RightStart: 0, Count: 10},
				},
			},
			wantSegmentIdx: 0,
		},
		{
			name: "diff with hunk after unchanged",
			diff: &diff.AlignedFileDiff{
				Segments: []diff.Segment{
					diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 5},
					diff.HunkSegment{
						LeftLines:  []diff.HunkLine{{SourceIdx: 5, Type: diff.HunkLineRemoved}},
						RightLines: []diff.HunkLine{{SourceIdx: 5, Type: diff.HunkLineAdded}},
					},
				},
			},
			wantSegmentIdx: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := New(
				core.NewSingleCommitRange(core.GitCommit{Hash: "abc123"}),
				core.FileChange{Path: "test.go"},
				tt.diff,
				nil,
			)

			if state.SegmentIndex != tt.wantSegmentIdx {
				t.Errorf("New().SegmentIndex = %d, want %d", state.SegmentIndex, tt.wantSegmentIdx)
			}
			if state.LeftOffset != 0 {
				t.Errorf("New().LeftOffset = %d, want 0", state.LeftOffset)
			}
			if state.RightOffset != 0 {
				t.Errorf("New().RightOffset = %d, want 0", state.RightOffset)
			}
		})
	}
}

func TestSegmentLeftLineCount(t *testing.T) {
	state := &State{
		Diff: &diff.AlignedFileDiff{
			Segments: []diff.Segment{
				diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 5},
				diff.HunkSegment{
					LeftLines: []diff.HunkLine{
						{SourceIdx: 5, Type: diff.HunkLineRemoved},
						{SourceIdx: 6, Type: diff.HunkLineRemoved},
						{SourceIdx: 7, Type: diff.HunkLineModified},
					},
					RightLines: []diff.HunkLine{
						{SourceIdx: 5, Type: diff.HunkLineAdded},
					},
				},
				diff.UnchangedSegment{LeftStart: 8, RightStart: 6, Count: 10},
			},
		},
	}

	tests := []struct {
		name   string
		segIdx int
		want   int
	}{
		{"unchanged segment", 0, 5},
		{"hunk segment", 1, 3},
		{"another unchanged segment", 2, 10},
		{"negative index", -1, 0},
		{"out of bounds index", 99, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := state.segmentLeftLineCount(tt.segIdx)
			if got != tt.want {
				t.Errorf("segmentLeftLineCount(%d) = %d, want %d", tt.segIdx, got, tt.want)
			}
		})
	}

	// Test with nil diff
	nilState := &State{Diff: nil}
	if got := nilState.segmentLeftLineCount(0); got != 0 {
		t.Errorf("segmentLeftLineCount with nil Diff = %d, want 0", got)
	}
}

func TestSegmentRightLineCount(t *testing.T) {
	state := &State{
		Diff: &diff.AlignedFileDiff{
			Segments: []diff.Segment{
				diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 5},
				diff.HunkSegment{
					LeftLines: []diff.HunkLine{
						{SourceIdx: 5, Type: diff.HunkLineRemoved},
					},
					RightLines: []diff.HunkLine{
						{SourceIdx: 5, Type: diff.HunkLineAdded},
						{SourceIdx: 6, Type: diff.HunkLineAdded},
						{SourceIdx: 7, Type: diff.HunkLineAdded},
						{SourceIdx: 8, Type: diff.HunkLineModified},
					},
				},
				diff.UnchangedSegment{LeftStart: 6, RightStart: 9, Count: 10},
			},
		},
	}

	tests := []struct {
		name   string
		segIdx int
		want   int
	}{
		{"unchanged segment", 0, 5},
		{"hunk segment", 1, 4},
		{"another unchanged segment", 2, 10},
		{"negative index", -1, 0},
		{"out of bounds index", 99, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := state.segmentRightLineCount(tt.segIdx)
			if got != tt.want {
				t.Errorf("segmentRightLineCount(%d) = %d, want %d", tt.segIdx, got, tt.want)
			}
		})
	}

	// Test with nil diff
	nilState := &State{Diff: nil}
	if got := nilState.segmentRightLineCount(0); got != 0 {
		t.Errorf("segmentRightLineCount with nil Diff = %d, want 0", got)
	}
}

func TestTotalLeftLines(t *testing.T) {
	tests := []struct {
		name  string
		state *State
		want  int
	}{
		{
			name:  "nil diff",
			state: &State{Diff: nil},
			want:  0,
		},
		{
			name: "empty segments",
			state: &State{
				Diff: &diff.AlignedFileDiff{
					Segments: []diff.Segment{},
				},
			},
			want: 0,
		},
		{
			name: "single unchanged segment",
			state: &State{
				Diff: &diff.AlignedFileDiff{
					Segments: []diff.Segment{
						diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 10},
					},
				},
			},
			want: 10,
		},
		{
			name: "mixed segments",
			state: &State{
				Diff: &diff.AlignedFileDiff{
					Segments: []diff.Segment{
						diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 5},
						diff.HunkSegment{
							LeftLines: []diff.HunkLine{
								{SourceIdx: 5, Type: diff.HunkLineRemoved},
								{SourceIdx: 6, Type: diff.HunkLineRemoved},
							},
							RightLines: []diff.HunkLine{
								{SourceIdx: 5, Type: diff.HunkLineAdded},
							},
						},
						diff.UnchangedSegment{LeftStart: 7, RightStart: 6, Count: 3},
					},
				},
			},
			want: 10, // 5 + 2 + 3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.state.totalLeftLines()
			if got != tt.want {
				t.Errorf("totalLeftLines() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTotalRightLines(t *testing.T) {
	tests := []struct {
		name  string
		state *State
		want  int
	}{
		{
			name:  "nil diff",
			state: &State{Diff: nil},
			want:  0,
		},
		{
			name: "empty segments",
			state: &State{
				Diff: &diff.AlignedFileDiff{
					Segments: []diff.Segment{},
				},
			},
			want: 0,
		},
		{
			name: "single unchanged segment",
			state: &State{
				Diff: &diff.AlignedFileDiff{
					Segments: []diff.Segment{
						diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 10},
					},
				},
			},
			want: 10,
		},
		{
			name: "mixed segments",
			state: &State{
				Diff: &diff.AlignedFileDiff{
					Segments: []diff.Segment{
						diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 5},
						diff.HunkSegment{
							LeftLines: []diff.HunkLine{
								{SourceIdx: 5, Type: diff.HunkLineRemoved},
							},
							RightLines: []diff.HunkLine{
								{SourceIdx: 5, Type: diff.HunkLineAdded},
								{SourceIdx: 6, Type: diff.HunkLineAdded},
								{SourceIdx: 7, Type: diff.HunkLineAdded},
							},
						},
						diff.UnchangedSegment{LeftStart: 6, RightStart: 8, Count: 3},
					},
				},
			},
			want: 11, // 5 + 3 + 3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.state.totalRightLines()
			if got != tt.want {
				t.Errorf("totalRightLines() = %d, want %d", got, tt.want)
			}
		})
	}
}
