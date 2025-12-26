package styles

import "github.com/charmbracelet/lipgloss"

var (
	// HashStyle is used for commit hashes (amber/yellow)
	// Light: darker amber (#d78700), Dark: bright amber (#ffaf00)
	HashStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "172",
		Dark:  "214",
	})

	// MessageStyle is used for commit messages
	// Light: dark gray (#3a3a3a), Dark: bright white (#d0d0d0)
	MessageStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "237",
		Dark:  "252",
	})

	// AuthorStyle is used for author names (cyan)
	// Light: darker cyan (#00af87), Dark: bright cyan (#5fd7af)
	AuthorStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "36",
		Dark:  "86",
	})

	// TimeStyle is used for relative timestamps (gray)
	// Light: medium gray (#6c6c6c), Dark: muted gray (#808080)
	TimeStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "242",
		Dark:  "244",
	})

	// Selected line styles (bold + brighter colors)
	SelectedHashStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "208", // Bright orange (#ff8700)
		Dark:  "220", // Brighter yellow (#ffd700)
	}).Bold(true)

	SelectedMessageStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "16",  // Black (#000000)
		Dark:  "231", // Bright white (#ffffff)
	}).Bold(true)

	SelectedAuthorStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "30",  // Darker bright cyan (#008787)
		Dark:  "51",  // Bright cyan (#00ffff)
	}).Bold(true)

	SelectedTimeStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "240", // Dark gray (#585858)
		Dark:  "250", // Lighter gray (#bcbcbc)
	}).Bold(true)

	// File path styles
	FilePathStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "237",
		Dark:  "252",
	})

	SelectedFilePathStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "16",
		Dark:  "231",
	}).Bold(true)

	// Addition count style (green)
	AdditionsStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "28",  // Dark green
		Dark:  "46",  // Bright green
	})

	// Diff line addition style (green background)
	DiffAdditionsStyle = lipgloss.NewStyle().Background(lipgloss.AdaptiveColor{
		Light: "#e8f5e9", // Very subtle pale green
		Dark:  "#1e3a1e", // Very subtle dark green
	})

	SelectedAdditionsStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "22",  // Darker green
		Dark:  "82",  // Brighter green
	}).Bold(true)

	// Deletion count style (red)
	DeletionsStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "124", // Dark red
		Dark:  "196", // Bright red
	})

	// Diff line deletion style (red background)
	DiffDeletionsStyle = lipgloss.NewStyle().Background(lipgloss.AdaptiveColor{
		Light: "#ffebee", // Very subtle pale pink
		Dark:  "#3a1e1e", // Very subtle dark red
	})

	// Brighter deletion style for inline highlighting (1.3× brighter)
	DiffDeletionsBrightStyle = lipgloss.NewStyle().Background(lipgloss.AdaptiveColor{
		Light: "#ffcdd2", // Brighter pale pink (1.3× brighter)
		Dark:  "#4d2626", // Brighter dark red (1.3× brighter)
	})

	SelectedDeletionsStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "88",  // Darker red
		Dark:  "9",   // Brighter red
	}).Bold(true)

	// Diff line addition style (green background) - moved here for inline bright version
	// Brighter addition style for inline highlighting (1.3× brighter)
	DiffAdditionsBrightStyle = lipgloss.NewStyle().Background(lipgloss.AdaptiveColor{
		Light: "#c8e6c9", // Brighter pale green (1.3× brighter)
		Dark:  "#264d26", // Brighter dark green (1.3× brighter)
	})

	// Header styles for commit info
	HeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "243",
		Dark:  "248",
	})
)
