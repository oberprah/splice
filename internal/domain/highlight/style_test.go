package highlight

import (
	"testing"

	"github.com/alecthomas/chroma/v2"
)

func TestStyleForToken_ReturnsStyle(t *testing.T) {
	// Test that StyleForToken returns a valid Lip Gloss style
	style := StyleForToken(chroma.Keyword)

	// Just verify we can call the function without panicking
	// We can't easily test the actual color without mocking terminal state
	_ = style
}

func TestStyleForToken_HandlesTextToken(t *testing.T) {
	// Text tokens might not have a specific color in the style
	style := StyleForToken(chroma.Text)

	// Should return a valid style (even if no color is set)
	_ = style
}

func TestStyleForToken_HandlesVariousTokenTypes(t *testing.T) {
	tokenTypes := []chroma.TokenType{
		chroma.Keyword,
		chroma.String,
		chroma.Comment,
		chroma.Name,
		chroma.Number,
		chroma.Operator,
		chroma.Punctuation,
		chroma.Text,
	}

	for _, tokenType := range tokenTypes {
		style := StyleForToken(tokenType)
		// Just verify we can get a style for each type
		_ = style
	}
}

func TestActiveStyleInitialized(t *testing.T) {
	// Verify that activeStyle was initialized in init()
	if activeStyle == nil {
		t.Error("activeStyle should be initialized")
	}
}

func TestIsDarkTerminal(t *testing.T) {
	// Just verify the function can be called
	// We can't test the actual return value as it depends on terminal state
	result := isDarkTerminal()
	_ = result
}
