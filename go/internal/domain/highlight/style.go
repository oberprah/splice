package highlight

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
)

var activeStyle *chroma.Style

func init() {
	// Select style based on terminal background
	if isDarkTerminal() {
		activeStyle = styles.Get("monokai")
	} else {
		activeStyle = styles.Get("github")
	}
}

// isDarkTerminal detects if the terminal has a dark background
func isDarkTerminal() bool {
	// Use Lip Gloss to detect terminal background
	// HasDarkBackground returns true if background is dark
	return lipgloss.HasDarkBackground()
}

// StyleForToken returns a Lip Gloss style for the given token type.
func StyleForToken(tokenType chroma.TokenType) lipgloss.Style {
	entry := activeStyle.Get(tokenType)
	style := lipgloss.NewStyle()

	if entry.Colour.IsSet() {
		style = style.Foreground(lipgloss.Color(entry.Colour.String()))
	}

	return style
}
