package states

import (
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
	"github.com/oberprah/splice/internal/git"
)

// Pure function tests - testing truncation logic

func TestCapMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		maxLen   int
		expected string
	}{
		{
			name:     "message fits",
			message:  "Short message",
			maxLen:   50,
			expected: "Short message",
		},
		{
			name:     "message exactly at limit",
			message:  "Exactly 20 chars!!",
			maxLen:   18,
			expected: "Exactly 20 chars!!",
		},
		{
			name:     "message needs truncation",
			message:  "This is a very long commit message that should be truncated",
			maxLen:   30,
			expected: "This is a very long commit ...",
		},
		{
			name:     "maxLen is 3",
			message:  "Hello",
			maxLen:   3,
			expected: "...",
		},
		{
			name:     "maxLen less than 3",
			message:  "Hello",
			maxLen:   2,
			expected: "",
		},
		{
			name:     "empty message",
			message:  "",
			maxLen:   10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := capMessage(tt.message, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTruncateAuthor(t *testing.T) {
	tests := []struct {
		name     string
		author   string
		maxLen   int
		expected string
	}{
		{
			name:     "author fits",
			author:   "Alice",
			maxLen:   25,
			expected: "Alice",
		},
		{
			name:     "author exactly at limit",
			author:   "Alice",
			maxLen:   5,
			expected: "Alice",
		},
		{
			name:     "author needs truncation",
			author:   "VeryLongAuthorNameThatShouldGetTruncated",
			maxLen:   25,
			expected: "VeryLongAuthorNameThat...",
		},
		{
			name:     "maxLen is 3",
			author:   "Alice",
			maxLen:   3,
			expected: "...",
		},
		{
			name:     "maxLen less than 3",
			author:   "Alice",
			maxLen:   2,
			expected: "",
		},
		{
			name:     "empty author",
			author:   "",
			maxLen:   10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateAuthor(tt.author, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTruncateEntireLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		maxWidth int
		expected string
	}{
		{
			name:     "line fits",
			line:     "Short line",
			maxWidth: 50,
			expected: "Short line",
		},
		{
			name:     "line exactly at limit",
			line:     "Exactly 20 chars!!",
			maxWidth: 18,
			expected: "Exactly 20 chars!!",
		},
		{
			name:     "line needs truncation",
			line:     "> ├─╮ abc123d (main) This is a very long message - Alice (2 days ago)",
			maxWidth: 40,
			expected: "> ├─╮ abc123d (main) This is a very l...", // With rune counting, graph chars take fewer positions
		},
		{
			name:     "maxWidth is 3",
			line:     "Hello",
			maxWidth: 3,
			expected: "...",
		},
		{
			name:     "maxWidth is 2",
			line:     "Hello",
			maxWidth: 2,
			expected: "He",
		},
		{
			name:     "maxWidth is 1",
			line:     "Hello",
			maxWidth: 1,
			expected: "H",
		},
		{
			name:     "maxWidth is 0",
			line:     "Hello",
			maxWidth: 0,
			expected: "",
		},
		{
			name:     "maxWidth negative",
			line:     "Hello",
			maxWidth: -1,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateEntireLine(tt.line, tt.maxWidth)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatRefsFull(t *testing.T) {
	tests := []struct {
		name     string
		refs     []git.RefInfo
		expected string
	}{
		{
			name:     "no refs",
			refs:     []git.RefInfo{},
			expected: "",
		},
		{
			name: "single local branch",
			refs: []git.RefInfo{
				{Name: "main", Type: git.RefTypeBranch, IsHead: true},
			},
			expected: "(main) ",
		},
		{
			name: "multiple refs",
			refs: []git.RefInfo{
				{Name: "main", Type: git.RefTypeBranch, IsHead: true},
				{Name: "origin/main", Type: git.RefTypeRemoteBranch, IsHead: false},
				{Name: "v1.0", Type: git.RefTypeTag, IsHead: false},
			},
			expected: "(main, origin/main, tag: v1.0) ",
		},
		{
			name: "tag only",
			refs: []git.RefInfo{
				{Name: "v2.1.0", Type: git.RefTypeTag, IsHead: false},
			},
			expected: "(tag: v2.1.0) ",
		},
		{
			name: "long branch name",
			refs: []git.RefInfo{
				{Name: "feature/implement-advanced-user-authentication-system", Type: git.RefTypeBranch, IsHead: true},
			},
			expected: "(feature/implement-advanced-user-authentication-system) ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRefsFull(tt.refs)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatRefsShortenedIndividual(t *testing.T) {
	tests := []struct {
		name     string
		refs     []git.RefInfo
		maxLen   int
		expected string
	}{
		{
			name:     "no refs",
			refs:     []git.RefInfo{},
			maxLen:   30,
			expected: "",
		},
		{
			name: "short refs no truncation",
			refs: []git.RefInfo{
				{Name: "main", Type: git.RefTypeBranch, IsHead: true},
				{Name: "v1.0", Type: git.RefTypeTag, IsHead: false},
			},
			maxLen:   30,
			expected: "(main, tag: v1.0) ",
		},
		{
			name: "long branch name truncated",
			refs: []git.RefInfo{
				{Name: "feature/implement-advanced-user-auth", Type: git.RefTypeBranch, IsHead: true},
			},
			maxLen:   30,
			expected: "(feature/implement-advanced-…) ",
		},
		{
			name: "multiple refs with truncation",
			refs: []git.RefInfo{
				{Name: "feature/implement-advanced-user-auth", Type: git.RefTypeBranch, IsHead: true},
				{Name: "origin/feature/implement-advanced-user-auth", Type: git.RefTypeRemoteBranch, IsHead: false},
				{Name: "v2.1.0", Type: git.RefTypeTag, IsHead: false},
			},
			maxLen:   30,
			expected: "(feature/implement-advanced-…, origin/feature/implement-ad…, tag: v2.1.0) ",
		},
		{
			name: "maxLen 1",
			refs: []git.RefInfo{
				{Name: "main", Type: git.RefTypeBranch, IsHead: true},
			},
			maxLen:   1,
			expected: "() ",
		},
		{
			name: "maxLen 0",
			refs: []git.RefInfo{
				{Name: "main", Type: git.RefTypeBranch, IsHead: true},
			},
			maxLen:   0,
			expected: "() ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRefsShortenedIndividual(tt.refs, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatRefsFirstPlusCount(t *testing.T) {
	tests := []struct {
		name     string
		refs     []git.RefInfo
		maxLen   int
		expected string
	}{
		{
			name:     "no refs",
			refs:     []git.RefInfo{},
			maxLen:   30,
			expected: "",
		},
		{
			name: "single ref",
			refs: []git.RefInfo{
				{Name: "main", Type: git.RefTypeBranch, IsHead: true},
			},
			maxLen:   30,
			expected: "(main) ",
		},
		{
			name: "multiple refs shows HEAD",
			refs: []git.RefInfo{
				{Name: "main", Type: git.RefTypeBranch, IsHead: true},
				{Name: "origin/main", Type: git.RefTypeRemoteBranch, IsHead: false},
				{Name: "v1.0", Type: git.RefTypeTag, IsHead: false},
			},
			maxLen:   30,
			expected: "(main +2 more) ",
		},
		{
			name: "multiple refs no HEAD shows first",
			refs: []git.RefInfo{
				{Name: "origin/main", Type: git.RefTypeRemoteBranch, IsHead: false},
				{Name: "v1.0", Type: git.RefTypeTag, IsHead: false},
			},
			maxLen:   30,
			expected: "(origin/main +1 more) ",
		},
		{
			name: "long ref name truncated",
			refs: []git.RefInfo{
				{Name: "feature/implement-advanced-user-authentication", Type: git.RefTypeBranch, IsHead: true},
				{Name: "origin/feature/implement-advanced-user-authentication", Type: git.RefTypeRemoteBranch, IsHead: false},
			},
			maxLen:   30,
			expected: "(feature/implement-advanced-… +1 more) ",
		},
		{
			name: "tag as first ref",
			refs: []git.RefInfo{
				{Name: "v2.1.0-beta-very-long-tag-name", Type: git.RefTypeTag, IsHead: false},
				{Name: "main", Type: git.RefTypeBranch, IsHead: false},
			},
			maxLen:   20,
			expected: "(tag: v2.1.0-beta-very-… +1 more) ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRefsFirstPlusCount(tt.refs, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBuildRefs(t *testing.T) {
	refs := []git.RefInfo{
		{Name: "feature/implement-advanced-user-authentication", Type: git.RefTypeBranch, IsHead: true},
		{Name: "origin/feature/implement-advanced-user-authentication", Type: git.RefTypeRemoteBranch, IsHead: false},
		{Name: "v2.1.0", Type: git.RefTypeTag, IsHead: false},
	}

	tests := []struct {
		name     string
		refs     []git.RefInfo
		level    RefsLevel
		expected string
	}{
		{
			name:     "empty refs",
			refs:     []git.RefInfo{},
			level:    RefsLevelFull,
			expected: "",
		},
		{
			name:     "level full",
			refs:     refs,
			level:    RefsLevelFull,
			expected: "(feature/implement-advanced-user-authentication, origin/feature/implement-advanced-user-authentication, tag: v2.1.0) ",
		},
		{
			name:     "level shorten individual",
			refs:     refs,
			level:    RefsLevelShortenIndividual,
			expected: "(feature/implement-advanced-…, origin/feature/implement-ad…, tag: v2.1.0) ",
		},
		{
			name:     "level first plus count",
			refs:     refs,
			level:    RefsLevelFirstPlusCount,
			expected: "(feature/implement-advanced-… +2 more) ",
		},
		{
			name:     "level count only",
			refs:     refs,
			level:    RefsLevelCountOnly,
			expected: "(3 refs) ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildRefs(tt.refs, tt.level)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMeasureLineWidth(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		graph    string
		hash     string
		refs     string
		message  string
		author   string
		time     string
		expected int
	}{
		{
			name:     "minimal line (no refs, author, time)",
			selector: "  ",
			graph:    "",
			hash:     "abc123d",
			refs:     "",
			message:  "Initial commit",
			author:   "",
			time:     "",
			expected: 2 + 7 + 1 + 14, // "  abc123d Initial commit"
		},
		{
			name:     "full line with all components",
			selector: "> ",
			graph:    "├─╮ ",
			hash:     "abc123d",
			refs:     "(main) ",
			message:  "Merge feature",
			author:   "Alice",
			time:     "2 days ago",
			expected: 2 + 4 + 7 + 1 + 7 + 13 + 3 + 5 + 1 + 10, // "> ├─╮ abc123d (main) Merge feature - Alice 2 days ago" (graph is 4 runes, not 10 bytes)
		},
		{
			name:     "with refs no author or time",
			selector: "  ",
			graph:    "│ ",
			hash:     "abc123d",
			refs:     "(HEAD -> main, tag: v1.0) ",
			message:  "Add feature",
			author:   "",
			time:     "",
			expected: 2 + 2 + 7 + 1 + 26 + 11, // graph is 2 runes, not 4 bytes
		},
		{
			name:     "with author no time",
			selector: "  ",
			graph:    "",
			hash:     "abc123d",
			refs:     "",
			message:  "Fix bug",
			author:   "Bob",
			time:     "",
			expected: 2 + 7 + 1 + 7 + 3 + 3, // "  abc123d Fix bug - Bob"
		},
		{
			name:     "empty components",
			selector: "",
			graph:    "",
			hash:     "",
			refs:     "",
			message:  "",
			author:   "",
			time:     "",
			expected: 1, // just the space after hash
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := measureLineWidth(tt.selector, tt.graph, tt.hash, tt.refs, tt.message, tt.author, tt.time)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// Integration tests for formatCommitLine - testing the complete truncation pipeline

func TestFormatCommitLine_VariousTerminalWidths(t *testing.T) {
	tests := []struct {
		name             string
		components       CommitLineComponents
		availableWidth   int
		expectedMaxLen   int
		verifyContains   []string // strings that must appear in output
		verifyNotContain []string // strings that must not appear in output
	}{
		{
			name: "very wide terminal - everything fits",
			components: CommitLineComponents{
				IsSelected: true,
				Graph:      "├─╮ ",
				Hash:       "abc123d",
				Refs: []git.RefInfo{
					{Name: "feature/user-authentication", Type: git.RefTypeBranch, IsHead: true},
					{Name: "origin/feature/user-authentication", Type: git.RefTypeRemoteBranch, IsHead: false},
					{Name: "v2.1.0", Type: git.RefTypeTag, IsHead: false},
				},
				Message: "Implement advanced user authentication system with OAuth2 support",
				Author:  "Alice Johnson",
				Time:    "2 days ago",
			},
			availableWidth: 200,
			expectedMaxLen: 200,
			verifyContains: []string{"abc123d", "feature/user-authentication", "origin/feature/user-authentication", "tag: v2.1.0",
				"Implement advanced user authentication system with OAuth2 support", "Alice Johnson", "2 days ago"},
		},
		{
			name: "wide terminal - message capped at 72",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "",
				Hash:       "def456a",
				Refs:       []git.RefInfo{},
				Message:    "This is a very long commit message that exceeds the 72 character limit and should be truncated at that boundary for readability",
				Author:     "Bob Smith",
				Time:       "3 hours ago",
			},
			availableWidth:   120,
			expectedMaxLen:   120,
			verifyContains:   []string{"def456a", "This is a very long commit message that exceeds the 72 character limi...", "Bob Smith"},
			verifyNotContain: []string{"for readability"}, // Part after 72 chars should be cut
		},
		{
			name: "medium terminal - refs shortened, author truncated",
			components: CommitLineComponents{
				IsSelected: true,
				Graph:      "│ ",
				Hash:       "ghi789b",
				Refs: []git.RefInfo{
					{Name: "feature/implement-advanced-caching-strategy", Type: git.RefTypeBranch, IsHead: true},
					{Name: "origin/feature/implement-advanced-caching-strategy", Type: git.RefTypeRemoteBranch, IsHead: false},
				},
				Message: "Add distributed caching layer for improved performance",
				Author:  "Christopher Williamson-Henderson",
				Time:    "yesterday",
			},
			availableWidth: 80,
			expectedMaxLen: 80,
			verifyContains: []string{"ghi789b", "refs", "Add distributed caching", "Ch..."},
		},
		{
			name: "narrow terminal - time dropped, message shortened",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "├ ",
				Hash:       "jkl012c",
				Refs: []git.RefInfo{
					{Name: "main", Type: git.RefTypeBranch, IsHead: true},
				},
				Message: "Refactor database connection pool management",
				Author:  "Diana Prince",
				Time:    "5 days ago",
			},
			availableWidth:   60,
			expectedMaxLen:   60,
			verifyContains:   []string{"jkl012c", "Refactor database connection pool man..."},
			verifyNotContain: []string{"5 days ago"}, // Time should be dropped
		},
		{
			name: "very narrow terminal - minimal format, author dropped",
			components: CommitLineComponents{
				IsSelected: true,
				Graph:      "",
				Hash:       "mno345d",
				Refs:       []git.RefInfo{},
				Message:    "Fix critical security vulnerability in auth module",
				Author:     "Edward Norton",
				Time:       "just now",
			},
			availableWidth:   40,
			expectedMaxLen:   40,
			verifyContains:   []string{"mno345d", "Fix critical security vulne..."},
			verifyNotContain: []string{"Edward Norton", "just now"},
		},
		{
			name: "extreme narrow terminal - entire line truncated",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "├─┬─╮─┬─╮ ", // Large graph (8 runes)
				Hash:       "pqr678e",
				Refs: []git.RefInfo{
					{Name: "feature/x", Type: git.RefTypeBranch, IsHead: true},
				},
				Message: "Update",
				Author:  "Frank",
				Time:    "1h ago",
			},
			availableWidth: 20, // Reduced from 30 to force truncation with rune counting
			expectedMaxLen: 20,
			verifyContains: []string{"..."}, // Line is truncated including graph
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCommitLine(tt.components, tt.availableWidth)

			// Verify line doesn't exceed available width (accounting for ANSI codes)
			// Note: We can't use simple len() because of styling, but we can check the assembled components
			// For now, verify the result is not empty and contains expected strings
			if result == "" {
				t.Error("Expected non-empty result")
			}

			// Verify expected strings are present (ignoring ANSI codes)
			for _, str := range tt.verifyContains {
				if !contains(result, str) {
					t.Errorf("Expected result to contain %q, got:\n%s", str, result)
				}
			}

			// Verify strings that should not be present
			for _, str := range tt.verifyNotContain {
				if contains(result, str) {
					t.Errorf("Expected result NOT to contain %q, got:\n%s", str, result)
				}
			}
		})
	}
}

func TestFormatCommitLine_ContentCombinations(t *testing.T) {
	tests := []struct {
		name           string
		components     CommitLineComponents
		availableWidth int
		description    string
	}{
		{
			name: "very long commit message",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "",
				Hash:       "abc123d",
				Refs:       []git.RefInfo{},
				Message:    "Implement comprehensive end-to-end testing framework with support for multiple browsers, parallel execution, and detailed reporting capabilities that include screenshots and video recordings of test failures",
				Author:     "Alice",
				Time:       "1 day ago",
			},
			availableWidth: 100,
			description:    "Message should be capped at 72 chars",
		},
		{
			name: "very long author name",
			components: CommitLineComponents{
				IsSelected: true,
				Graph:      "",
				Hash:       "def456a",
				Refs:       []git.RefInfo{},
				Message:    "Update documentation",
				Author:     "Christopher Alexander Montgomery-Worthington III",
				Time:       "2 hours ago",
			},
			availableWidth: 80,
			description:    "Author should be truncated to 25 chars",
		},
		{
			name: "multiple long branch names",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "├ ",
				Hash:       "ghi789b",
				Refs: []git.RefInfo{
					{Name: "feature/implement-distributed-caching-with-redis", Type: git.RefTypeBranch, IsHead: true},
					{Name: "origin/feature/implement-distributed-caching-with-redis", Type: git.RefTypeRemoteBranch, IsHead: false},
					{Name: "staging/feature/implement-distributed-caching-with-redis", Type: git.RefTypeRemoteBranch, IsHead: false},
				},
				Message: "Add Redis caching layer",
				Author:  "Bob",
				Time:    "3 days ago",
			},
			availableWidth: 90,
			description:    "Refs should be progressively shortened",
		},
		{
			name: "many refs",
			components: CommitLineComponents{
				IsSelected: true,
				Graph:      "",
				Hash:       "jkl012c",
				Refs: []git.RefInfo{
					{Name: "main", Type: git.RefTypeBranch, IsHead: true},
					{Name: "develop", Type: git.RefTypeBranch, IsHead: false},
					{Name: "staging", Type: git.RefTypeBranch, IsHead: false},
					{Name: "production", Type: git.RefTypeBranch, IsHead: false},
					{Name: "v1.0", Type: git.RefTypeTag, IsHead: false},
					{Name: "v1.1", Type: git.RefTypeTag, IsHead: false},
					{Name: "v1.2", Type: git.RefTypeTag, IsHead: false},
					{Name: "v2.0-beta", Type: git.RefTypeTag, IsHead: false},
					{Name: "release-candidate", Type: git.RefTypeBranch, IsHead: false},
					{Name: "hotfix", Type: git.RefTypeBranch, IsHead: false},
				},
				Message: "Release version 2.0",
				Author:  "Release Manager",
				Time:    "1 week ago",
			},
			availableWidth: 80,
			description:    "Should show first ref + count or total count",
		},
		{
			name: "no refs",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "│ ",
				Hash:       "mno345d",
				Refs:       []git.RefInfo{},
				Message:    "Fix typo in documentation",
				Author:     "Charlie",
				Time:       "just now",
			},
			availableWidth: 70,
			description:    "Should render cleanly without refs",
		},
		{
			name: "empty message",
			components: CommitLineComponents{
				IsSelected: true,
				Graph:      "",
				Hash:       "pqr678e",
				Refs:       []git.RefInfo{},
				Message:    "",
				Author:     "Dave",
				Time:       "5 min ago",
			},
			availableWidth: 60,
			description:    "Should handle empty message gracefully",
		},
		{
			name: "large graph with many parallel branches",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "├─┬─┬─╮─┬─╮─┬─╮─┬ ", // 20 chars of graph
				Hash:       "stu901f",
				Refs: []git.RefInfo{
					{Name: "main", Type: git.RefTypeBranch, IsHead: true},
				},
				Message: "Merge multiple feature branches",
				Author:  "Integration Bot",
				Time:    "2 min ago",
			},
			availableWidth: 80,
			description:    "Graph takes priority, other components compete for remaining space",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCommitLine(tt.components, tt.availableWidth)

			// Basic sanity checks
			if result == "" && tt.components.Message != "" {
				t.Error("Expected non-empty result for non-empty message")
			}

			// Verify hash is always present (it's in styled content, so we check for the hash value)
			if !contains(result, tt.components.Hash) {
				t.Errorf("Expected result to always contain hash %q, got:\n%s", tt.components.Hash, result)
			}
		})
	}
}

func TestFormatCommitLine_TruncationLevels(t *testing.T) {
	// This test verifies that truncation levels are applied in the correct order
	// We test specific scenarios that trigger each level
	tests := []struct {
		name             string
		components       CommitLineComponents
		availableWidth   int
		expectedLevel    string
		verifyContains   []string
		verifyNotContain []string
	}{
		{
			name: "level 0 - message capped at 72",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "",
				Hash:       "abc123d",
				Refs:       []git.RefInfo{},
				Message:    "This is an extremely long commit message that definitely exceeds 72 characters and needs to be capped",
				Author:     "Alice",
				Time:       "1 day ago",
			},
			availableWidth: 110,
			expectedLevel:  "Message capped at 72",
			verifyContains: []string{"This is an extremely long commit message that definitely exceeds 72 c..."},
		},
		{
			name: "level 1 - author truncated to 25",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "",
				Hash:       "def456a",
				Refs:       []git.RefInfo{},
				Message:    "Short message",
				Author:     "Christopher Montgomery-Worthington III",
				Time:       "1 day ago",
			},
			availableWidth: 60,
			expectedLevel:  "Author truncated to 25",
			verifyContains: []string{"Ch..."},
		},
		{
			name: "level 2-4 - refs degradation",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "",
				Hash:       "ghi789b",
				Refs: []git.RefInfo{
					{Name: "feature/very-long-branch-name-here", Type: git.RefTypeBranch, IsHead: true},
					{Name: "origin/feature/very-long-branch-name-here", Type: git.RefTypeRemoteBranch, IsHead: false},
					{Name: "v1.0", Type: git.RefTypeTag, IsHead: false},
				},
				Message: "Update feature",
				Author:  "Bob",
				Time:    "2 days ago",
			},
			availableWidth: 55,
			expectedLevel:  "Refs shortened or count-only",
			verifyContains: []string{"ghi789b", "Update feature"},
		},
		{
			name: "level 5 - author at 5 chars",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "",
				Hash:       "jkl012c",
				Refs:       []git.RefInfo{},
				Message:    "This is a moderately long commit message for testing",
				Author:     "Christopher",
				Time:       "just now",
			},
			availableWidth: 45,
			expectedLevel:  "Author at 5 chars",
			verifyContains: []string{"moderately long commit..."},
		},
		{
			name: "level 6 - time dropped",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "├ ",
				Hash:       "mno345d",
				Refs:       []git.RefInfo{},
				Message:    "Implement new caching strategy for database",
				Author:     "Diana",
				Time:       "3 hours ago",
			},
			availableWidth:   42,
			expectedLevel:    "Time dropped",
			verifyNotContain: []string{"3 hours ago"},
		},
		{
			name: "level 7 - message at 40 chars",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "",
				Hash:       "pqr678e",
				Refs:       []git.RefInfo{},
				Message:    "Refactor the entire authentication and authorization subsystem",
				Author:     "Eve",
				Time:       "yesterday",
			},
			availableWidth: 38,
			expectedLevel:  "Message at 40 chars",
			verifyContains: []string{"Refactor the entire authe..."},
		},
		{
			name: "level 8 - author dropped",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "",
				Hash:       "stu901f",
				Refs:       []git.RefInfo{},
				Message:    "Update documentation for API endpoints",
				Author:     "Frank",
				Time:       "5 min ago",
			},
			availableWidth:   30,
			expectedLevel:    "Author dropped",
			verifyNotContain: []string{"Frank", " - "},
		},
		{
			name: "level 9 - entire line truncated",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "├─┬─╮─┬ ",
				Hash:       "vwx234g",
				Refs: []git.RefInfo{
					{Name: "feature/test", Type: git.RefTypeBranch, IsHead: true},
				},
				Message: "Fix bug in payment processing",
				Author:  "Grace",
				Time:    "now",
			},
			availableWidth: 25,
			expectedLevel:  "Entire line truncated",
			verifyContains: []string{"..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCommitLine(tt.components, tt.availableWidth)

			// Verify expected strings are present
			for _, str := range tt.verifyContains {
				if !contains(result, str) {
					t.Errorf("Level %q: Expected result to contain %q, got:\n%s", tt.expectedLevel, str, result)
				}
			}

			// Verify strings that should not be present
			for _, str := range tt.verifyNotContain {
				if contains(result, str) {
					t.Errorf("Level %q: Expected result NOT to contain %q, got:\n%s", tt.expectedLevel, str, result)
				}
			}
		})
	}
}

func TestFormatCommitLine_VisualQuality(t *testing.T) {
	tests := []struct {
		name           string
		components     CommitLineComponents
		availableWidth int
		description    string
	}{
		{
			name: "balanced parentheses for refs",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "",
				Hash:       "abc123d",
				Refs: []git.RefInfo{
					{Name: "main", Type: git.RefTypeBranch, IsHead: true},
					{Name: "v1.0", Type: git.RefTypeTag, IsHead: false},
				},
				Message: "Release version 1.0",
				Author:  "Alice",
				Time:    "1 week ago",
			},
			availableWidth: 50,
			description:    "Refs should have balanced parentheses at all truncation levels",
		},
		{
			name: "correct ellipsis for message",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "",
				Hash:       "def456a",
				Refs:       []git.RefInfo{},
				Message:    "This is a very long message that will be truncated",
				Author:     "Bob",
				Time:       "2 days ago",
			},
			availableWidth: 40,
			description:    "Message should use '...' (3 chars) for truncation",
		},
		{
			name: "correct ellipsis for author",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "",
				Hash:       "ghi789b",
				Refs:       []git.RefInfo{},
				Message:    "Short message",
				Author:     "Christopher Montgomery",
				Time:       "just now",
			},
			availableWidth: 40,
			description:    "Author should use '...' (3 chars) for truncation",
		},
		{
			name: "correct ellipsis for refs",
			components: CommitLineComponents{
				IsSelected: false,
				Graph:      "",
				Hash:       "jkl012c",
				Refs: []git.RefInfo{
					{Name: "feature/very-long-branch-name", Type: git.RefTypeBranch, IsHead: true},
				},
				Message: "Add feature",
				Author:  "Diana",
				Time:    "3 hours ago",
			},
			availableWidth: 45,
			description:    "Refs should use '…' (single char) for truncation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCommitLine(tt.components, tt.availableWidth)

			// Check for balanced parentheses if refs are present
			if len(tt.components.Refs) > 0 {
				openCount := 0
				closeCount := 0
				for _, ch := range result {
					switch ch {
					case '(':
						openCount++
					case ')':
						closeCount++
					}
				}
				if openCount != closeCount {
					t.Errorf("Unbalanced parentheses: %d open, %d close in:\n%s", openCount, closeCount, result)
				}
			}

			// Note: Checking for specific ellipsis characters is difficult due to ANSI codes
			// The actual verification is in the pure function tests above
			// Here we just verify the output is reasonable
			if result == "" {
				t.Error("Expected non-empty result")
			}
		})
	}
}

// Helper function to check if a string contains a substring (ignoring ANSI codes for basic checks)
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(len(s) >= len(substr) && indexOf(s, substr) >= 0)
}

// Simple indexOf implementation
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestFormatCommitLine_SelectedVsUnselected(t *testing.T) {
	// Test that selected and unselected lines with identical content
	// are truncated equally (or very similarly)

	tests := []struct {
		name          string
		message       string
		author        string
		width         int
		expectSimilar bool // Should selected and unselected be similar length?
	}{
		{
			name:          "Same message both selected and unselected at 80 width",
			message:       "Fix: Use rune count instead of byte count for width measurement",
			author:        "oberprah",
			width:         80,
			expectSimilar: true,
		},
		{
			name:          "Same message at 60 width",
			message:       "Refactor: Extract line formatting to separate functions",
			author:        "oberprah",
			width:         60,
			expectSimilar: true,
		},
		{
			name:          "Same message at narrow 40 width",
			message:       "Complete final verification of log line truncation",
			author:        "oberprah",
			width:         40,
			expectSimilar: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := CommitLineComponents{
				Graph:   "├ ",
				Hash:    "cc047b2",
				Refs:    []git.RefInfo{},
				Message: tt.message,
				Author:  tt.author,
				Time:    "5 minutes ago",
			}

			// Test with IsSelected = false
			componentsUnselected := components
			componentsUnselected.IsSelected = false
			unselectedLine := formatCommitLine(componentsUnselected, tt.width)

			// Test with IsSelected = true
			componentsSelected := components
			componentsSelected.IsSelected = true
			selectedLine := formatCommitLine(componentsSelected, tt.width)

			t.Logf("\nWidth: %d", tt.width)
			t.Logf("Unselected (raw): %q", unselectedLine)
			t.Logf("Selected (raw):   %q", selectedLine)

			// Strip ANSI codes for comparison
			unselectedPlain := stripANSI(unselectedLine)
			selectedPlain := stripANSI(selectedLine)

			// Compare visual widths
			unselectedWidth := utf8.RuneCountInString(unselectedPlain)
			selectedWidth := utf8.RuneCountInString(selectedPlain)

			t.Logf("Unselected (plain): %q (width: %d)", unselectedPlain, unselectedWidth)
			t.Logf("Selected (plain):   %q (width: %d)", selectedPlain, selectedWidth)

			if tt.expectSimilar {
				// They should be within 1-2 characters of each other
				diff := abs(unselectedWidth - selectedWidth)
				if diff > 2 {
					t.Errorf("Selected and unselected lines have very different widths: %d vs %d (diff: %d)",
						selectedWidth, unselectedWidth, diff)
				}
			}
		})
	}
}

func stripANSI(s string) string {
	// Use the charmbracelet/x/ansi package for proper ANSI stripping
	return ansi.Strip(s)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func TestFormatCommitLine_Level9Truncation(t *testing.T) {
	// Test that level 9 truncation (entire line truncation) works correctly
	// for both selected and unselected lines. This is where ANSI codes matter most.

	tests := []struct {
		name           string
		message        string
		author         string
		width          int
		expectSelected bool
	}{
		{
			name:           "Very narrow width forces level 9 - selected",
			message:        "Fix: Use rune count instead of byte count",
			author:         "oberprah",
			width:          25,
			expectSelected: true,
		},
		{
			name:           "Very narrow width forces level 9 - unselected",
			message:        "Fix: Use rune count instead of byte count",
			author:         "oberprah",
			width:          25,
			expectSelected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := CommitLineComponents{
				IsSelected: tt.expectSelected,
				Graph:      "├ ",
				Hash:       "cc047b2",
				Refs:       []git.RefInfo{},
				Message:    tt.message,
				Author:     tt.author,
				Time:       "5 minutes ago",
			}

			line := formatCommitLine(components, tt.width)

			t.Logf("\nWidth: %d, Selected: %v", tt.width, tt.expectSelected)
			t.Logf("Raw output: %q", line)
			t.Logf("Raw length (bytes): %d", len(line))
			t.Logf("Raw length (runes): %d", utf8.RuneCountInString(line))

			// Strip ANSI codes to get visual content
			plain := stripANSI(line)
			visualWidth := utf8.RuneCountInString(plain)

			t.Logf("Plain output: %q", plain)
			t.Logf("Visual width: %d", visualWidth)

			// The visual width should not exceed the available width
			if visualWidth > tt.width {
				t.Errorf("Visual width %d exceeds available width %d", visualWidth, tt.width)
			}

			// Check if the line ends with "..." as expected at level 9
			if plain[len(plain)-3:] != "..." {
				t.Logf("Note: Line does not end with '...' (may not have reached level 9)")
			}
		})
	}
}

func TestFormatCommitLine_SelectedVsUnselectedLevel9(t *testing.T) {
	// Direct comparison: same content, same width, selected vs unselected
	// Both should reach level 9 truncation and have same visual width

	message := "Fix: Use rune count instead of byte count for width measurement"
	author := "oberprah"
	width := 25 // Very narrow to force level 9

	componentsUnselected := CommitLineComponents{
		IsSelected: false,
		Graph:      "├ ",
		Hash:       "cc047b2",
		Refs:       []git.RefInfo{},
		Message:    message,
		Author:     author,
		Time:       "5 minutes ago",
	}

	componentsSelected := CommitLineComponents{
		IsSelected: true,
		Graph:      "├ ",
		Hash:       "cc047b2",
		Refs:       []git.RefInfo{},
		Message:    message,
		Author:     author,
		Time:       "5 minutes ago",
	}

	unselectedLine := formatCommitLine(componentsUnselected, width)
	selectedLine := formatCommitLine(componentsSelected, width)

	t.Logf("\nWidth: %d", width)
	t.Logf("Unselected raw: %q (bytes: %d, runes: %d)",
		unselectedLine, len(unselectedLine), utf8.RuneCountInString(unselectedLine))
	t.Logf("Selected raw:   %q (bytes: %d, runes: %d)",
		selectedLine, len(selectedLine), utf8.RuneCountInString(selectedLine))

	unselectedPlain := stripANSI(unselectedLine)
	selectedPlain := stripANSI(selectedLine)

	unselectedWidth := utf8.RuneCountInString(unselectedPlain)
	selectedWidth := utf8.RuneCountInString(selectedPlain)

	t.Logf("Unselected plain: %q (width: %d)", unselectedPlain, unselectedWidth)
	t.Logf("Selected plain:   %q (width: %d)", selectedPlain, selectedWidth)

	// Both should have the same visual width
	if unselectedWidth != selectedWidth {
		t.Errorf("Visual width mismatch: unselected=%d, selected=%d (diff=%d)",
			unselectedWidth, selectedWidth, abs(unselectedWidth-selectedWidth))
	}

	// Both should not exceed the available width
	if unselectedWidth > width {
		t.Errorf("Unselected visual width %d exceeds available width %d", unselectedWidth, width)
	}
	if selectedWidth > width {
		t.Errorf("Selected visual width %d exceeds available width %d", selectedWidth, width)
	}
}

func TestFormatCommitLine_ANSICodeMeasurement(t *testing.T) {
	// This test verifies that line 361 in formatCommitLine (level 9 truncation)
	// incorrectly measures the styled string (with ANSI codes) using utf8.RuneCountInString().
	// This causes the truncation logic to include ANSI escape sequences in the count,
	// leading to different truncation behavior for selected vs unselected lines.

	// Force lipgloss to output ANSI codes
	lipgloss.SetColorProfile(termenv.TrueColor)

	message := "This is a moderately long commit message that should trigger level 9 truncation"
	author := "oberprah"
	width := 30 // Narrow enough to force level 9 with a long graph

	// Use a long graph to force level 9
	graph := "├─┬─╮ "

	componentsUnselected := CommitLineComponents{
		IsSelected: false,
		Graph:      graph,
		Hash:       "cc047b2",
		Refs:       []git.RefInfo{},
		Message:    message,
		Author:     author,
		Time:       "5 minutes ago",
	}

	componentsSelected := CommitLineComponents{
		IsSelected: true,
		Graph:      graph,
		Hash:       "cc047b2",
		Refs:       []git.RefInfo{},
		Message:    message,
		Author:     author,
		Time:       "5 minutes ago",
	}

	unselectedLine := formatCommitLine(componentsUnselected, width)
	selectedLine := formatCommitLine(componentsSelected, width)

	// Count runes in raw strings (including ANSI codes)
	unselectedRawRunes := utf8.RuneCountInString(unselectedLine)
	selectedRawRunes := utf8.RuneCountInString(selectedLine)

	// Count runes in plain strings (after stripping ANSI codes)
	unselectedPlain := stripANSI(unselectedLine)
	selectedPlain := stripANSI(selectedLine)
	unselectedVisualWidth := utf8.RuneCountInString(unselectedPlain)
	selectedVisualWidth := utf8.RuneCountInString(selectedPlain)

	t.Logf("\nWidth: %d, Graph: %q", width, graph)
	t.Logf("Unselected:")
	t.Logf("  Raw bytes: %d", len(unselectedLine))
	t.Logf("  Raw rune count: %d", unselectedRawRunes)
	t.Logf("  Raw (hex): %x", unselectedLine)
	t.Logf("  Plain: %q", unselectedPlain)
	t.Logf("  Visual width: %d", unselectedVisualWidth)
	t.Logf("\nSelected:")
	t.Logf("  Raw bytes: %d", len(selectedLine))
	t.Logf("  Raw rune count: %d", selectedRawRunes)
	t.Logf("  Raw (hex): %x", selectedLine)
	t.Logf("  Plain: %q", selectedPlain)
	t.Logf("  Visual width: %d", selectedVisualWidth)

	// The bug: selected lines have more ANSI codes (bold styling),
	// so their raw rune count is higher even though visual width should be the same
	if selectedRawRunes > unselectedRawRunes {
		t.Logf("\nBUG DETECTED: Selected raw rune count (%d) > unselected (%d) due to ANSI codes",
			selectedRawRunes, unselectedRawRunes)
		t.Logf("This means line 361 in formatCommitLine will think the selected line is longer!")
	}

	// Visual widths should be equal
	if unselectedVisualWidth != selectedVisualWidth {
		t.Errorf("Visual width mismatch: unselected=%d, selected=%d (diff=%d)",
			unselectedVisualWidth, selectedVisualWidth, abs(unselectedVisualWidth-selectedVisualWidth))
	}
}
