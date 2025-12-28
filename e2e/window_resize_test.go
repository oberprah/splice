package main

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui"
)

// createLongCommits creates test commits with long messages and authors to demonstrate truncation
func createLongCommits(count int) []git.GitCommit {
	commits := make([]git.GitCommit, count)
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	messages := []string{
		"Refactor authentication module to support OAuth2.0 and JWT tokens with refresh capability",
		"Add comprehensive error handling for network failures in API client with retry logic",
		"Update database schema migration to include new user preferences and notification settings",
		"Implement dark mode toggle with system preference detection and custom theme support",
		"Fix critical security vulnerability in session management discovered during audit",
		"Optimize rendering performance for large commit lists using virtualization techniques",
		"Add internationalization support for multiple languages including right-to-left layouts",
		"Migrate legacy codebase from deprecated framework to modern architecture patterns",
		"Implement automated backup system with incremental snapshots and cloud storage integration",
		"Add real-time collaboration features with WebSocket support and conflict resolution",
	}

	authors := []string{
		"Alexandra Chen-Rodriguez",
		"Benjamin O'Sullivan-Smith",
		"Christina Van Der Berg",
	}

	for i := 0; i < count; i++ {
		commits[i] = git.GitCommit{
			Hash:    fmt.Sprintf("%040d", i),
			Message: messages[i%len(messages)],
			Body:    "Additional context about this change that provides more detailed information.",
			Author:  authors[i%len(authors)],
			Date:    baseTime.Add(time.Duration(-i) * time.Hour),
		}
	}

	return commits
}

func TestWindowResize(t *testing.T) {
	commits := createLongCommits(15)

	// Create long file paths to demonstrate truncation
	fileChanges := []git.FileChange{
		{Path: "internal/authentication/oauth2/providers/github/client.go", Additions: 245, Deletions: 123},
		{Path: "internal/authentication/oauth2/providers/google/client.go", Additions: 198, Deletions: 87},
		{Path: "pkg/database/migrations/v2/user_preferences_schema.sql", Additions: 67, Deletions: 12},
		{Path: "cmd/server/handlers/api/v1/users/profile_handler.go", Additions: 142, Deletions: 56},
		{Path: "web/components/settings/themes/dark_mode_toggle.tsx", Additions: 89, Deletions: 34},
	}

	// Create mock diff data with long lines to demonstrate truncation
	oldContent := `func authenticateUser(username, password string) (*User, error) {
	if username == "" || password == "" {
		return nil, errors.New("username and password required")
	}
	return nil, errors.New("not implemented")
}
`
	newContent := `func authenticateUser(username, password string, options *AuthenticationOptions) (*User, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("authentication failed: username and password are required fields")
	}

	// Validate password complexity requirements before attempting authentication
	if len(password) < options.MinPasswordLength {
		return nil, fmt.Errorf("password does not meet minimum length requirement of %d characters", options.MinPasswordLength)
	}

	return authenticateWithProvider(username, password, options.Provider)
}
`
	diffOutput := `@@ -1,6 +1,12 @@
-func authenticateUser(username, password string) (*User, error) {
+func authenticateUser(username, password string, options *AuthenticationOptions) (*User, error) {
 	if username == "" || password == "" {
-		return nil, errors.New("username and password required")
+		return nil, fmt.Errorf("authentication failed: username and password are required fields")
 	}
-	return nil, errors.New("not implemented")
+
+	// Validate password complexity requirements before attempting authentication
+	if len(password) < options.MinPasswordLength {
+		return nil, fmt.Errorf("password does not meet minimum length requirement of %d characters", options.MinPasswordLength)
+	}
+
+	return authenticateWithProvider(username, password, options.Provider)
 }
`

	m := ui.NewModel(
		ui.WithFetchCommits(func(limit int) ([]git.GitCommit, error) {
			if limit < len(commits) {
				return commits[:limit], nil
			}
			return commits, nil
		}),
		ui.WithFetchFileChanges(func(commitHash string) ([]git.FileChange, error) {
			return fileChanges, nil
		}),
		ui.WithFetchFullFileDiff(func(commitHash string, change git.FileChange) (*git.FullFileDiffResult, error) {
			return &git.FullFileDiffResult{
				OldContent: oldContent,
				NewContent: newContent,
				DiffOutput: diffOutput,
				OldPath:    change.Path,
				NewPath:    change.Path,
			}, nil
		}),
	)

	runner := NewE2ETestRunner(t, m)

	// 1. Start small (detail panel NOT shown, width < 160)
	runner.Send(tea.WindowSizeMsg{Width: 140, Height: 24})
	runner.AssertGolden("window_resize/1_log_simple.golden")

	// 2. Make it larger (detail panel shown, width >= 160)
	runner.Send(tea.WindowSizeMsg{Width: 180, Height: 40})
	runner.AssertGolden("window_resize/2_log_split.golden")

	// 3. Move down (detail panel updated)
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	runner.AssertGolden("window_resize/3_log_detail_updated.golden")

	// 4. Select commit (Enter to go to files view)
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	runner.AssertGolden("window_resize/4_files_view.golden")

	// 5. Make window smaller (file paths truncated)
	runner.Send(tea.WindowSizeMsg{Width: 50, Height: 20})
	runner.AssertGolden("window_resize/5_files_truncated.golden")

	// 6. Make window larger again (full file paths shown)
	runner.Send(tea.WindowSizeMsg{Width: 180, Height: 40})
	runner.AssertGolden("window_resize/6_files_full.golden")

	// 7. Select first file (Enter to go to diff view)
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	runner.AssertGolden("window_resize/7_diff_view.golden")

	// 8. Make window smaller (diff lines truncated)
	runner.Send(tea.WindowSizeMsg{Width: 60, Height: 20})
	runner.AssertGolden("window_resize/8_diff_truncated.golden")

	// 9. Make window larger again (full diff shown)
	runner.Send(tea.WindowSizeMsg{Width: 180, Height: 40})
	runner.AssertGolden("window_resize/9_diff_full.golden")

	// 10. Go back to files (q from diff view)
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	runner.AssertGolden("window_resize/10_back_to_files.golden")

	// 11. Go back to log (q key from files view)
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	runner.AssertGolden("window_resize/11_back_to_log.golden")

	// Quit (q from log view quits the app)
	runner.Quit()
}
