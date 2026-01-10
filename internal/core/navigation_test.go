package core

import (
	"testing"

	"github.com/oberprah/splice/internal/domain/diff"
)

func TestPushFilesScreenMsg_WithCommitRangeDiffSource(t *testing.T) {
	start := GitCommit{Hash: "abc123"}
	end := GitCommit{Hash: "def456"}
	files := []FileChange{
		{Path: "file1.go", Status: "M"},
	}

	msg := PushFilesScreenMsg{
		Source: CommitRangeDiffSource{
			Start: start,
			End:   end,
			Count: 5,
		},
		Files:     files,
		ExitOnPop: false,
	}

	// Verify fields are set correctly
	if msg.Files[0].Path != "file1.go" {
		t.Errorf("Expected file path file1.go, got %s", msg.Files[0].Path)
	}
	if msg.ExitOnPop {
		t.Error("Expected ExitOnPop to be false")
	}

	// Verify DiffSource is the correct type
	commitRange, ok := msg.Source.(CommitRangeDiffSource)
	if !ok {
		t.Fatal("Expected Source to be CommitRangeDiffSource")
	}
	if commitRange.Start.Hash != "abc123" {
		t.Errorf("Expected start hash abc123, got %s", commitRange.Start.Hash)
	}
	if commitRange.End.Hash != "def456" {
		t.Errorf("Expected end hash def456, got %s", commitRange.End.Hash)
	}
	if commitRange.Count != 5 {
		t.Errorf("Expected count 5, got %d", commitRange.Count)
	}
}

func TestPushFilesScreenMsg_WithUncommittedChangesDiffSource(t *testing.T) {
	files := []FileChange{
		{Path: "file1.go", Status: "M"},
		{Path: "file2.go", Status: "A"},
	}

	msg := PushFilesScreenMsg{
		Source: UncommittedChangesDiffSource{
			Type: UncommittedTypeUnstaged,
		},
		Files:     files,
		ExitOnPop: true,
	}

	// Verify fields are set correctly
	if len(msg.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(msg.Files))
	}
	if !msg.ExitOnPop {
		t.Error("Expected ExitOnPop to be true")
	}

	// Verify DiffSource is the correct type
	uncommitted, ok := msg.Source.(UncommittedChangesDiffSource)
	if !ok {
		t.Fatal("Expected Source to be UncommittedChangesDiffSource")
	}
	if uncommitted.Type != UncommittedTypeUnstaged {
		t.Errorf("Expected type UncommittedTypeUnstaged, got %v", uncommitted.Type)
	}
}

func TestPushFilesScreenMsg_ExitOnPopVariants(t *testing.T) {
	source := CommitRangeDiffSource{
		Start: GitCommit{Hash: "abc"},
		End:   GitCommit{Hash: "def"},
		Count: 1,
	}

	tests := []struct {
		name      string
		exitOnPop bool
	}{
		{"ExitOnPop true", true},
		{"ExitOnPop false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := PushFilesScreenMsg{
				Source:    source,
				Files:     []FileChange{},
				ExitOnPop: tt.exitOnPop,
			}

			if msg.ExitOnPop != tt.exitOnPop {
				t.Errorf("Expected ExitOnPop %v, got %v", tt.exitOnPop, msg.ExitOnPop)
			}
		})
	}
}

func TestPushDiffScreenMsg_WithCommitRangeDiffSource(t *testing.T) {
	start := GitCommit{Hash: "abc123"}
	end := GitCommit{Hash: "def456"}
	file := FileChange{Path: "file1.go", Status: "M"}
	alignedDiff := &diff.AlignedFileDiff{}
	indices := []int{0, 5, 10}

	msg := PushDiffScreenMsg{
		Source: CommitRangeDiffSource{
			Start: start,
			End:   end,
			Count: 3,
		},
		File:          file,
		Diff:          alignedDiff,
		ChangeIndices: indices,
	}

	// Verify fields are set correctly
	if msg.File.Path != "file1.go" {
		t.Errorf("Expected file path file1.go, got %s", msg.File.Path)
	}
	if msg.Diff == nil {
		t.Error("Expected Diff to be non-nil")
	}
	if len(msg.ChangeIndices) != 3 {
		t.Errorf("Expected 3 change indices, got %d", len(msg.ChangeIndices))
	}

	// Verify DiffSource is the correct type
	commitRange, ok := msg.Source.(CommitRangeDiffSource)
	if !ok {
		t.Fatal("Expected Source to be CommitRangeDiffSource")
	}
	if commitRange.Start.Hash != "abc123" {
		t.Errorf("Expected start hash abc123, got %s", commitRange.Start.Hash)
	}
	if commitRange.End.Hash != "def456" {
		t.Errorf("Expected end hash def456, got %s", commitRange.End.Hash)
	}
}

func TestPushDiffScreenMsg_WithUncommittedChangesDiffSource(t *testing.T) {
	file := FileChange{Path: "main.go", Status: "A"}
	alignedDiff := &diff.AlignedFileDiff{}

	msg := PushDiffScreenMsg{
		Source: UncommittedChangesDiffSource{
			Type: UncommittedTypeStaged,
		},
		File:          file,
		Diff:          alignedDiff,
		ChangeIndices: []int{},
	}

	// Verify fields are set correctly
	if msg.File.Path != "main.go" {
		t.Errorf("Expected file path main.go, got %s", msg.File.Path)
	}

	// Verify DiffSource is the correct type
	uncommitted, ok := msg.Source.(UncommittedChangesDiffSource)
	if !ok {
		t.Fatal("Expected Source to be UncommittedChangesDiffSource")
	}
	if uncommitted.Type != UncommittedTypeStaged {
		t.Errorf("Expected type UncommittedTypeStaged, got %v", uncommitted.Type)
	}
}

func TestPushDiffScreenMsg_AllUncommittedTypes(t *testing.T) {
	file := FileChange{Path: "test.go", Status: "M"}
	alignedDiff := &diff.AlignedFileDiff{}

	types := []UncommittedType{
		UncommittedTypeUnstaged,
		UncommittedTypeStaged,
		UncommittedTypeAll,
	}

	for _, uncommittedType := range types {
		t.Run(uncommittedType.String(), func(t *testing.T) {
			msg := PushDiffScreenMsg{
				Source: UncommittedChangesDiffSource{
					Type: uncommittedType,
				},
				File:          file,
				Diff:          alignedDiff,
				ChangeIndices: []int{},
			}

			uncommitted, ok := msg.Source.(UncommittedChangesDiffSource)
			if !ok {
				t.Fatal("Expected Source to be UncommittedChangesDiffSource")
			}
			if uncommitted.Type != uncommittedType {
				t.Errorf("Expected type %v, got %v", uncommittedType, uncommitted.Type)
			}
		})
	}
}

func TestNavigationMessages_TypeSwitch(t *testing.T) {
	// Test that type switch works correctly with DiffSource in navigation messages
	commitRangeSource := CommitRangeDiffSource{
		Start: GitCommit{Hash: "abc"},
		End:   GitCommit{Hash: "def"},
		Count: 1,
	}
	uncommittedSource := UncommittedChangesDiffSource{
		Type: UncommittedTypeAll,
	}

	filesMsg1 := PushFilesScreenMsg{Source: commitRangeSource, Files: []FileChange{}}
	filesMsg2 := PushFilesScreenMsg{Source: uncommittedSource, Files: []FileChange{}}

	// Verify type switch works for PushFilesScreenMsg
	switch filesMsg1.Source.(type) {
	case CommitRangeDiffSource:
		// Expected
	case UncommittedChangesDiffSource:
		t.Error("filesMsg1 should be CommitRangeDiffSource")
	default:
		t.Error("Unexpected type for filesMsg1.Source")
	}

	switch filesMsg2.Source.(type) {
	case CommitRangeDiffSource:
		t.Error("filesMsg2 should be UncommittedChangesDiffSource")
	case UncommittedChangesDiffSource:
		// Expected
	default:
		t.Error("Unexpected type for filesMsg2.Source")
	}

	diffMsg1 := PushDiffScreenMsg{
		Source:        commitRangeSource,
		File:          FileChange{},
		Diff:          &diff.AlignedFileDiff{},
		ChangeIndices: []int{},
	}
	diffMsg2 := PushDiffScreenMsg{
		Source:        uncommittedSource,
		File:          FileChange{},
		Diff:          &diff.AlignedFileDiff{},
		ChangeIndices: []int{},
	}

	// Verify type switch works for PushDiffScreenMsg
	switch diffMsg1.Source.(type) {
	case CommitRangeDiffSource:
		// Expected
	case UncommittedChangesDiffSource:
		t.Error("diffMsg1 should be CommitRangeDiffSource")
	default:
		t.Error("Unexpected type for diffMsg1.Source")
	}

	switch diffMsg2.Source.(type) {
	case CommitRangeDiffSource:
		t.Error("diffMsg2 should be UncommittedChangesDiffSource")
	case UncommittedChangesDiffSource:
		// Expected
	default:
		t.Error("Unexpected type for diffMsg2.Source")
	}
}
