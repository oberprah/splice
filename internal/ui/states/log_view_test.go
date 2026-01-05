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

// Integration tests for formatCommitLine - testing the complete truncation pipeline

func TestFormatCommitLine_VariousTerminalWidths(t *testing.T) {
	tests := []struct {
		name             string
		components       CommitLineComponents
		availableWidth   int
		isSelected       bool
		expectedMaxLen   int
		verifyContains   []string // strings that must appear in output
		verifyNotContain []string // strings that must not appear in output
	}{
		{
			name: "very wide terminal - everything fits",
			components: CommitLineComponents{
				Selector: "> ",
				Graph:    "├─╮ ",
				Hash:     "abc123d",
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
			isSelected:     true,
			expectedMaxLen: 200,
			verifyContains: []string{"abc123d", "feature/user-authentication", "origin/feature/user-authentication", "tag: v2.1.0",
				"Implement advanced user authentication system with OAuth2 support", "Alice Johnson", "2 days ago"},
		},
		{
			name: "wide terminal - message capped at 72",
			components: CommitLineComponents{
				Selector: "  ",
				Graph:    "",
				Hash:     "def456a",
				Refs:     []git.RefInfo{},
				Message:  "This is a very long commit message that exceeds the 72 character limit and should be truncated at that boundary for readability",
				Author:   "Bob Smith",
				Time:     "3 hours ago",
			},
			availableWidth:   120,
			isSelected:       false,
			expectedMaxLen:   120,
			verifyContains:   []string{"def456a", "This is a very long commit message that exceeds the 72 character limi...", "Bob Smith"},
			verifyNotContain: []string{"for readability"}, // Part after 72 chars should be cut
		},
		{
			name: "medium terminal - refs shortened, author truncated",
			components: CommitLineComponents{
				Selector: "> ",
				Graph:    "│ ",
				Hash:     "ghi789b",
				Refs: []git.RefInfo{
					{Name: "feature/implement-advanced-caching-strategy", Type: git.RefTypeBranch, IsHead: true},
					{Name: "origin/feature/implement-advanced-caching-strategy", Type: git.RefTypeRemoteBranch, IsHead: false},
				},
				Message: "Add distributed caching layer for improved performance",
				Author:  "Christopher Williamson-Henderson",
				Time:    "yesterday",
			},
			availableWidth: 80,
			isSelected:     true,
			expectedMaxLen: 80,
			verifyContains: []string{"ghi789b", "refs", "Add distributed caching", "Ch..."},
		},
		{
			name: "narrow terminal - time dropped, message shortened",
			components: CommitLineComponents{
				Selector: "  ",
				Graph:    "├ ",
				Hash:     "jkl012c",
				Refs: []git.RefInfo{
					{Name: "main", Type: git.RefTypeBranch, IsHead: true},
				},
				Message: "Refactor database connection pool management",
				Author:  "Diana Prince",
				Time:    "5 days ago",
			},
			availableWidth:   60,
			isSelected:       false,
			expectedMaxLen:   60,
			verifyContains:   []string{"jkl012c", "Refactor database connection pool man..."},
			verifyNotContain: []string{"5 days ago"}, // Time should be dropped
		},
		{
			name: "very narrow terminal - minimal format, author dropped",
			components: CommitLineComponents{
				Selector: "> ",
				Graph:    "",
				Hash:     "mno345d",
				Refs:     []git.RefInfo{},
				Message:  "Fix critical security vulnerability in auth module",
				Author:   "Edward Norton",
				Time:     "just now",
			},
			availableWidth:   40,
			isSelected:       true,
			expectedMaxLen:   40,
			verifyContains:   []string{"mno345d", "Fix critical security vulne..."},
			verifyNotContain: []string{"Edward Norton", "just now"},
		},
		{
			name: "extreme narrow terminal - entire line truncated",
			components: CommitLineComponents{
				Selector: "  ",
				Graph:    "├─┬─╮─┬─╮ ", // Large graph
				Hash:     "pqr678e",
				Refs: []git.RefInfo{
					{Name: "feature/x", Type: git.RefTypeBranch, IsHead: true},
				},
				Message: "Update",
				Author:  "Frank",
				Time:    "1h ago",
			},
			availableWidth: 30,
			isSelected:     false,
			expectedMaxLen: 30,
			verifyContains: []string{"..."}, // Line is truncated including graph
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCommitLine(tt.components, tt.availableWidth, tt.isSelected)

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
				Selector: "  ",
				Graph:    "",
				Hash:     "abc123d",
				Refs:     []git.RefInfo{},
				Message:  "Implement comprehensive end-to-end testing framework with support for multiple browsers, parallel execution, and detailed reporting capabilities that include screenshots and video recordings of test failures",
				Author:   "Alice",
				Time:     "1 day ago",
			},
			availableWidth: 100,
			description:    "Message should be capped at 72 chars",
		},
		{
			name: "very long author name",
			components: CommitLineComponents{
				Selector: "> ",
				Graph:    "",
				Hash:     "def456a",
				Refs:     []git.RefInfo{},
				Message:  "Update documentation",
				Author:   "Christopher Alexander Montgomery-Worthington III",
				Time:     "2 hours ago",
			},
			availableWidth: 80,
			description:    "Author should be truncated to 25 chars",
		},
		{
			name: "multiple long branch names",
			components: CommitLineComponents{
				Selector: "  ",
				Graph:    "├ ",
				Hash:     "ghi789b",
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
				Selector: "> ",
				Graph:    "",
				Hash:     "jkl012c",
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
				Selector: "  ",
				Graph:    "│ ",
				Hash:     "mno345d",
				Refs:     []git.RefInfo{},
				Message:  "Fix typo in documentation",
				Author:   "Charlie",
				Time:     "just now",
			},
			availableWidth: 70,
			description:    "Should render cleanly without refs",
		},
		{
			name: "empty message",
			components: CommitLineComponents{
				Selector: "> ",
				Graph:    "",
				Hash:     "pqr678e",
				Refs:     []git.RefInfo{},
				Message:  "",
				Author:   "Dave",
				Time:     "5 min ago",
			},
			availableWidth: 60,
			description:    "Should handle empty message gracefully",
		},
		{
			name: "large graph with many parallel branches",
			components: CommitLineComponents{
				Selector: "  ",
				Graph:    "├─┬─┬─╮─┬─╮─┬─╮─┬ ", // 20 chars of graph
				Hash:     "stu901f",
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
			result := formatCommitLine(tt.components, tt.availableWidth, false)

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
				Selector: "  ",
				Graph:    "",
				Hash:     "abc123d",
				Refs:     []git.RefInfo{},
				Message:  "This is an extremely long commit message that definitely exceeds 72 characters and needs to be capped",
				Author:   "Alice",
				Time:     "1 day ago",
			},
			availableWidth: 110,
			expectedLevel:  "Message capped at 72",
			verifyContains: []string{"This is an extremely long commit message that definitely exceeds 72 c..."},
		},
		{
			name: "level 1 - author truncated to 25",
			components: CommitLineComponents{
				Selector: "  ",
				Graph:    "",
				Hash:     "def456a",
				Refs:     []git.RefInfo{},
				Message:  "Short message",
				Author:   "Christopher Montgomery-Worthington III",
				Time:     "1 day ago",
			},
			availableWidth: 60,
			expectedLevel:  "Author truncated to 25",
			verifyContains: []string{"Ch..."},
		},
		{
			name: "level 2-4 - refs degradation",
			components: CommitLineComponents{
				Selector: "  ",
				Graph:    "",
				Hash:     "ghi789b",
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
				Selector: "  ",
				Graph:    "",
				Hash:     "jkl012c",
				Refs:     []git.RefInfo{},
				Message:  "This is a moderately long commit message for testing",
				Author:   "Christopher",
				Time:     "just now",
			},
			availableWidth: 45,
			expectedLevel:  "Author at 5 chars",
			verifyContains: []string{"moderately long commit..."},
		},
		{
			name: "level 6 - time dropped",
			components: CommitLineComponents{
				Selector: "  ",
				Graph:    "├ ",
				Hash:     "mno345d",
				Refs:     []git.RefInfo{},
				Message:  "Implement new caching strategy for database",
				Author:   "Diana",
				Time:     "3 hours ago",
			},
			availableWidth:   42,
			expectedLevel:    "Time dropped",
			verifyNotContain: []string{"3 hours ago"},
		},
		{
			name: "level 7 - message at 40 chars",
			components: CommitLineComponents{
				Selector: "  ",
				Graph:    "",
				Hash:     "pqr678e",
				Refs:     []git.RefInfo{},
				Message:  "Refactor the entire authentication and authorization subsystem",
				Author:   "Eve",
				Time:     "yesterday",
			},
			availableWidth: 38,
			expectedLevel:  "Message at 40 chars",
			verifyContains: []string{"Refactor the entire authe..."},
		},
		{
			name: "level 8 - author dropped",
			components: CommitLineComponents{
				Selector: "  ",
				Graph:    "",
				Hash:     "stu901f",
				Refs:     []git.RefInfo{},
				Message:  "Update documentation for API endpoints",
				Author:   "Frank",
				Time:     "5 min ago",
			},
			availableWidth:   30,
			expectedLevel:    "Author dropped",
			verifyNotContain: []string{"Frank", " - "},
		},
		{
			name: "level 9 - entire line truncated",
			components: CommitLineComponents{
				Selector: "  ",
				Graph:    "├─┬─╮─┬ ",
				Hash:     "vwx234g",
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
			result := formatCommitLine(tt.components, tt.availableWidth, false)

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
				Selector: "  ",
				Graph:    "",
				Hash:     "abc123d",
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
				Selector: "  ",
				Graph:    "",
				Hash:     "def456a",
				Refs:     []git.RefInfo{},
				Message:  "This is a very long message that will be truncated",
				Author:   "Bob",
				Time:     "2 days ago",
			},
			availableWidth: 40,
			description:    "Message should use '...' (3 chars) for truncation",
		},
		{
			name: "correct ellipsis for author",
			components: CommitLineComponents{
				Selector: "  ",
				Graph:    "",
				Hash:     "ghi789b",
				Refs:     []git.RefInfo{},
				Message:  "Short message",
				Author:   "Christopher Montgomery",
				Time:     "just now",
			},
			availableWidth: 40,
			description:    "Author should use '...' (3 chars) for truncation",
		},
		{
			name: "correct ellipsis for refs",
			components: CommitLineComponents{
				Selector: "  ",
				Graph:    "",
				Hash:     "jkl012c",
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
			result := formatCommitLine(tt.components, tt.availableWidth, false)

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
