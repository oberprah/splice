package core

import (
	"testing"
	"time"

	"github.com/oberprah/splice/internal/git"
)

func makeTestCommit(hash string) git.GitCommit {
	return git.GitCommit{
		Hash:    hash,
		Message: "Test commit",
		Author:  "Test Author",
		Date:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

func TestIsSingleCommit_SingleCommit(t *testing.T) {
	commit := makeTestCommit("abc123")
	r := NewSingleCommitRange(commit)

	if !r.IsSingleCommit() {
		t.Errorf("IsSingleCommit() = false, want true for single commit")
	}
}

func TestIsSingleCommit_CommitRange(t *testing.T) {
	start := makeTestCommit("abc123")
	end := makeTestCommit("def456")
	r := NewCommitRange(start, end, 5)

	if r.IsSingleCommit() {
		t.Errorf("IsSingleCommit() = true, want false for multi-commit range")
	}
}

func TestNewSingleCommitRange(t *testing.T) {
	commit := makeTestCommit("abc123")
	r := NewSingleCommitRange(commit)

	if r.Start.Hash != "abc123" {
		t.Errorf("Start.Hash = %s, want abc123", r.Start.Hash)
	}
	if r.End.Hash != "abc123" {
		t.Errorf("End.Hash = %s, want abc123", r.End.Hash)
	}
	if r.Count != 1 {
		t.Errorf("Count = %d, want 1", r.Count)
	}
}

func TestNewCommitRange(t *testing.T) {
	start := makeTestCommit("abc123")
	end := makeTestCommit("def456")
	r := NewCommitRange(start, end, 5)

	if r.Start.Hash != "abc123" {
		t.Errorf("Start.Hash = %s, want abc123", r.Start.Hash)
	}
	if r.End.Hash != "def456" {
		t.Errorf("End.Hash = %s, want def456", r.End.Hash)
	}
	if r.Count != 5 {
		t.Errorf("Count = %d, want 5", r.Count)
	}
}

func TestNewCommitRange_OrderPreserved(t *testing.T) {
	// Test that the constructor preserves the order of commits
	older := makeTestCommit("older123")
	newer := makeTestCommit("newer456")
	r := NewCommitRange(older, newer, 3)

	if r.Start.Hash != "older123" {
		t.Errorf("Start.Hash = %s, want older123", r.Start.Hash)
	}
	if r.End.Hash != "newer456" {
		t.Errorf("End.Hash = %s, want newer456", r.End.Hash)
	}
}
