package states

import (
	"flag"
	"testing"
	"time"

	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/graph"
)

var update = flag.Bool("update", false, "update golden files")

func assertLogViewGolden(t *testing.T, output *ViewBuilder, filename string) {
	t.Helper()
	assertGolden(t, output.String(), "log_view/"+filename, *update)
}

// createLogStateWithGraph creates a LogState with computed GraphLayout
func createLogStateWithGraph(commits []git.GitCommit) LogState {
	// Convert to graph.Commits
	graphCommits := make([]graph.Commit, len(commits))
	for i, commit := range commits {
		graphCommits[i] = graph.Commit{
			Hash:    commit.Hash,
			Parents: commit.ParentHashes,
		}
	}

	// Compute layout
	layout := graph.ComputeLayout(graphCommits)

	return LogState{
		Commits:     commits,
		GraphLayout: layout,
	}
}

func TestLogState_View_RendersCommits(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	commits := []git.GitCommit{
		{Hash: "abc123", ParentHashes: []string{"def456"}, Message: "First commit", Body: "", Author: "Alice", Date: fixedTime},
		{Hash: "def456", ParentHashes: []string{}, Message: "Second commit", Body: "", Author: "Bob", Date: fixedTime.Add(time.Hour)},
	}

	s := createLogStateWithGraph(commits)
	s.Cursor = 0
	s.ViewportStart = 0
	s.Preview = PreviewNone{}
	ctx := mockContext{width: 80, height: 24}

	output := s.View(ctx)

	assertLogViewGolden(t, output, "renders_commits.golden")
}

func TestLogState_View_SelectionIndicator(t *testing.T) {
	commits := createTestCommits(3)

	tests := []struct {
		name       string
		cursor     int
		goldenFile string
	}{
		{"first commit selected", 0, "selection_first.golden"},
		{"second commit selected", 1, "selection_second.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := createLogStateWithGraph(commits)
			s.Cursor = tt.cursor
			s.ViewportStart = 0
			s.Preview = PreviewNone{}
			ctx := mockContext{width: 80, height: 24}

			output := s.View(ctx)

			assertLogViewGolden(t, output, tt.goldenFile)
		})
	}
}

func TestLogState_View_ViewportLimits(t *testing.T) {
	commits := createTestCommits(20)

	s := createLogStateWithGraph(commits)
	s.Cursor = 10
	s.ViewportStart = 5
	s.Preview = PreviewNone{}
	ctx := mockContext{width: 80, height: 10}

	output := s.View(ctx)

	assertLogViewGolden(t, output, "viewport_limits.golden")
}

func TestLogState_View_EmptyViewport(t *testing.T) {
	commits := createTestCommits(5)

	s := createLogStateWithGraph(commits)
	s.Cursor = 0
	s.ViewportStart = 10 // Beyond end
	s.Preview = PreviewNone{}
	ctx := mockContext{width: 80, height: 10}

	output := s.View(ctx)

	assertLogViewGolden(t, output, "empty_viewport.golden")
}

func TestLogState_View_LineTruncation(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	commits := []git.GitCommit{
		{
			Hash:         "abc123def456",
			ParentHashes: []string{},
			Message:      "This is a very long commit message that should be truncated when the terminal is narrow",
			Body:         "",
			Author:       "VeryLongAuthorNameThatShouldAlsoGetTruncated",
			Date:         fixedTime,
		},
	}

	s := createLogStateWithGraph(commits)
	s.Cursor = 0
	s.ViewportStart = 0
	s.Preview = PreviewNone{}

	ctx := mockContext{width: 40, height: 24}

	output := s.View(ctx)

	assertLogViewGolden(t, output, "line_truncation.golden")
}

func TestLogState_View_SplitView_WideTerminal(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	commits := []git.GitCommit{
		{Hash: "abc123", ParentHashes: []string{"def456"}, Message: "First commit", Body: "This is the body", Author: "Alice", Date: fixedTime},
		{Hash: "def456", ParentHashes: []string{}, Message: "Second commit", Body: "", Author: "Bob", Date: fixedTime.Add(time.Hour)},
	}

	files := []git.FileChange{
		{Path: "src/main.go", Status: "M", Additions: 10, Deletions: 5},
		{Path: "README.md", Status: "A", Additions: 20, Deletions: 0},
	}

	s := createLogStateWithGraph(commits)
	s.Cursor = 0
	s.ViewportStart = 0
	s.Preview = PreviewLoaded{ForHash: "abc123", Files: files}

	ctx := mockContext{width: 160, height: 24}

	output := s.View(ctx)

	assertLogViewGolden(t, output, "split_view_wide.golden")
}

func TestLogState_View_SplitView_NarrowTerminal(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	commits := []git.GitCommit{
		{Hash: "abc123", ParentHashes: []string{}, Message: "First commit", Body: "This is the body", Author: "Alice", Date: fixedTime},
	}

	files := []git.FileChange{
		{Path: "src/main.go", Status: "M", Additions: 10, Deletions: 5},
	}

	s := createLogStateWithGraph(commits)
	s.Cursor = 0
	s.ViewportStart = 0
	s.Preview = PreviewLoaded{ForHash: "abc123", Files: files}

	ctx := mockContext{width: 100, height: 24}

	output := s.View(ctx)

	assertLogViewGolden(t, output, "split_view_narrow.golden")
}

func TestLogState_View_SplitView_PreviewLoading(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	commits := []git.GitCommit{
		{Hash: "abc123", ParentHashes: []string{}, Message: "First commit", Body: "", Author: "Alice", Date: fixedTime},
	}

	s := createLogStateWithGraph(commits)
	s.Cursor = 0
	s.ViewportStart = 0
	s.Preview = PreviewLoading{ForHash: "abc123"}

	ctx := mockContext{width: 160, height: 24}

	output := s.View(ctx)

	assertLogViewGolden(t, output, "split_view_loading.golden")
}

func TestLogState_View_SplitView_PreviewError(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	commits := []git.GitCommit{
		{Hash: "abc123", ParentHashes: []string{}, Message: "First commit", Body: "", Author: "Alice", Date: fixedTime},
	}

	s := createLogStateWithGraph(commits)
	s.Cursor = 0
	s.ViewportStart = 0
	s.Preview = PreviewError{ForHash: "abc123", Err: nil}

	ctx := mockContext{width: 160, height: 24}

	output := s.View(ctx)

	assertLogViewGolden(t, output, "split_view_error.golden")
}

func TestLogState_View_MergeBranchGraph(t *testing.T) {
	// Create a simple feature branch merge scenario:
	// E (merge commit) <- merges B and D
	// D (feature branch)
	// C (feature branch)
	// B (main branch)
	// A (initial commit)
	//
	// Expected graph:
	// ├─╮  E: Merge feature-x
	// │ ├  D: Add feature X part 2
	// │ ├  C: Add feature X part 1
	// ├ │  B: Fix bug on main
	// ├─╯  A: Initial commit

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	commits := []git.GitCommit{
		{
			Hash:         "eeeeeeee",
			ParentHashes: []string{"bbbbbbbb", "dddddddd"}, // Merge commit
			Message:      "Merge feature-x",
			Body:         "",
			Author:       "Alice",
			Date:         fixedTime,
		},
		{
			Hash:         "dddddddd",
			ParentHashes: []string{"cccccccc"},
			Message:      "Add feature X part 2",
			Body:         "",
			Author:       "Bob",
			Date:         fixedTime.Add(time.Hour),
		},
		{
			Hash:         "cccccccc",
			ParentHashes: []string{"aaaaaaaa"},
			Message:      "Add feature X part 1",
			Body:         "",
			Author:       "Bob",
			Date:         fixedTime.Add(2 * time.Hour),
		},
		{
			Hash:         "bbbbbbbb",
			ParentHashes: []string{"aaaaaaaa"},
			Message:      "Fix bug on main",
			Body:         "",
			Author:       "Alice",
			Date:         fixedTime.Add(3 * time.Hour),
		},
		{
			Hash:         "aaaaaaaa",
			ParentHashes: []string{},
			Message:      "Initial commit",
			Body:         "",
			Author:       "Alice",
			Date:         fixedTime.Add(4 * time.Hour),
		},
	}

	s := createLogStateWithGraph(commits)
	s.Cursor = 0
	s.ViewportStart = 0
	s.Preview = PreviewNone{}
	ctx := mockContext{width: 80, height: 24}

	output := s.View(ctx)

	assertLogViewGolden(t, output, "merge_branch_graph.golden")
}

// Helper function tests - testing internal formatting logic
func TestLogState_formatFileEntry(t *testing.T) {
	s := LogState{}

	tests := []struct {
		name     string
		file     git.FileChange
		width    int
		expected string
	}{
		{
			name:     "modified file",
			file:     git.FileChange{Path: "src/main.go", Status: "M", Additions: 10, Deletions: 5},
			width:    80,
			expected: "M + 10 - 5  src/main.go",
		},
		{
			name:     "added file",
			file:     git.FileChange{Path: "README.md", Status: "A", Additions: 20, Deletions: 0},
			width:    80,
			expected: "A + 20 - 0  README.md",
		},
		{
			name:     "deleted file",
			file:     git.FileChange{Path: "old.txt", Status: "D", Additions: 0, Deletions: 15},
			width:    80,
			expected: "D + 0 - 15  old.txt",
		},
		{
			name:     "binary file",
			file:     git.FileChange{Path: "image.png", Status: "M", IsBinary: true},
			width:    80,
			expected: "M (binary)  image.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.formatFileEntry(tt.file, tt.width)

			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

// Helper function tests - testing internal text wrapping logic
func TestLogState_wrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected []string
	}{
		{
			name:     "short text",
			text:     "Hello world",
			width:    20,
			expected: []string{"Hello world"},
		},
		{
			name:  "long text",
			text:  "This is a very long line that should be wrapped into multiple lines",
			width: 20,
			expected: []string{
				"This is a very long",
				"line that should be",
				"wrapped into",
				"multiple lines",
			},
		},
		{
			name:     "zero width",
			text:     "Hello",
			width:    0,
			expected: []string{"Hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapText(tt.text, tt.width)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d\nExpected: %v\nGot: %v",
					len(tt.expected), len(result), tt.expected, result)
				return
			}

			for i, line := range result {
				if line != tt.expected[i] {
					t.Errorf("Line %d mismatch\nExpected: %q\nGot: %q", i, tt.expected[i], line)
				}
			}
		})
	}
}

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
			expected: "> ├─╮ abc123d (main) This is a ...",
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
			expected: 2 + 10 + 7 + 1 + 7 + 13 + 3 + 5 + 1 + 10, // "> ├─╮ abc123d (main) Merge feature - Alice 2 days ago"
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
			expected: 2 + 4 + 7 + 1 + 26 + 11,
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
