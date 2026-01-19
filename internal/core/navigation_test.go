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
		Files: files,
	}

	// Verify fields are set correctly
	if msg.Files[0].Path != "file1.go" {
		t.Errorf("Expected file path file1.go, got %s", msg.Files[0].Path)
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
		Files: files,
	}

	// Verify fields are set correctly
	if len(msg.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(msg.Files))
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

func TestPushDiffScreenMsg_WithCommitRangeDiffSource(t *testing.T) {
	start := GitCommit{Hash: "abc123"}
	end := GitCommit{Hash: "def456"}
	file := FileChange{Path: "file1.go", Status: "M"}
	files := []FileChange{file}
	fileDiff := &diff.FileDiff{Path: "file1.go", Blocks: []diff.Block{}}

	msg := PushDiffScreenMsg{
		Source: CommitRangeDiffSource{
			Start: start,
			End:   end,
			Count: 3,
		},
		Files:     files,
		FileIndex: 0,
		File:      file,
		Diff:      fileDiff,
	}

	// Verify fields are set correctly
	if msg.File.Path != "file1.go" {
		t.Errorf("Expected file path file1.go, got %s", msg.File.Path)
	}
	if msg.Diff == nil {
		t.Error("Expected Diff to be non-nil")
	}
	if len(msg.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(msg.Files))
	}
	if msg.FileIndex != 0 {
		t.Errorf("Expected file index 0, got %d", msg.FileIndex)
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
	files := []FileChange{file}
	fileDiff := &diff.FileDiff{Path: "main.go", Blocks: []diff.Block{}}

	msg := PushDiffScreenMsg{
		Source: UncommittedChangesDiffSource{
			Type: UncommittedTypeStaged,
		},
		Files:     files,
		FileIndex: 0,
		File:      file,
		Diff:      fileDiff,
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
	files := []FileChange{file}
	fileDiff := &diff.FileDiff{Path: "test.go", Blocks: []diff.Block{}}

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
				Files:     files,
				FileIndex: 0,
				File:      file,
				Diff:      fileDiff,
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
		Source:    commitRangeSource,
		Files:     []FileChange{{}},
		FileIndex: 0,
		File:      FileChange{},
		Diff:      &diff.FileDiff{Blocks: []diff.Block{}},
	}
	diffMsg2 := PushDiffScreenMsg{
		Source:    uncommittedSource,
		Files:     []FileChange{{}},
		FileIndex: 0,
		File:      FileChange{},
		Diff:      &diff.FileDiff{Blocks: []diff.Block{}},
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
