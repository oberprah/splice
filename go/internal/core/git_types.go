package core

import "time"

// RefType represents the type of a git reference
type RefType int

const (
	RefTypeBranch       RefType = iota // Local branch (e.g., "main")
	RefTypeRemoteBranch                // Remote branch (e.g., "origin/main")
	RefTypeTag                         // Tag (e.g., "v1.0")
)

// RefInfo represents a git reference (branch, tag, or HEAD pointer)
type RefInfo struct {
	Name   string  // e.g., "main", "v1.0", "origin/main"
	Type   RefType // Branch, RemoteBranch, or Tag
	IsHead bool    // true if this is the current HEAD
}

// GitCommit represents a single git commit with all necessary display information
type GitCommit struct {
	Hash         string    // Full 40-char hash
	ParentHashes []string  // Parent commit hashes (empty for root commits, space-separated in git output)
	Refs         []RefInfo // Branch/tag decorations
	Message      string    // First line of commit message (subject)
	Body         string    // Commit message body (everything after subject line)
	Author       string    // Author name (not email)
	Date         time.Time // Commit timestamp
}

// FileChange represents a file that was changed in a commit
type FileChange struct {
	Path      string // File path relative to repository root
	Status    string // Git status: M (modified), A (added), D (deleted), R (renamed), etc.
	Additions int    // Number of lines added
	Deletions int    // Number of lines deleted
	IsBinary  bool   // True if the file is binary
}
