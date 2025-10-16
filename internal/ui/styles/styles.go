package styles

import "github.com/charmbracelet/lipgloss"

var (
	// HashStyle is used for commit hashes (amber/yellow)
	HashStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	// MessageStyle is used for commit messages (bright white/default)
	MessageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	// AuthorStyle is used for author names (cyan)
	AuthorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))

	// TimeStyle is used for relative timestamps (muted gray)
	TimeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	// Selected line styles (bold + brighter colors)
	SelectedHashStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true) // Brighter yellow
	SelectedMessageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Bold(true) // Bright white
	SelectedAuthorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)  // Bright cyan
	SelectedTimeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Bold(true) // Lighter gray
)
