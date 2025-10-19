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
)
