package highlight

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// Token represents a syntax-highlighted token
type Token struct {
	Type  chroma.TokenType // e.g., chroma.Keyword, chroma.String, chroma.Text
	Value string           // the actual text
}

// TokenizeFile tokenizes file content and returns tokens grouped by line.
// Always returns valid tokens - uses Text tokens for unsupported languages.
func TokenizeFile(content, filename string) [][]Token {
	// Try to find a lexer for this file type
	lexer := lexers.Match(filename)
	if lexer == nil {
		// No lexer found - return Text tokens, one per line
		return textTokensForContent(content)
	}

	// Tokenize the entire file
	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		// Tokenization failed - fall back to Text tokens
		return textTokensForContent(content)
	}

	// Split tokens by line
	return splitTokensByLine(iterator)
}

// textTokensForContent creates Text tokens for content (one per line)
func textTokensForContent(content string) [][]Token {
	lines := strings.Split(content, "\n")
	result := make([][]Token, len(lines))

	for i, line := range lines {
		if line != "" {
			result[i] = []Token{{Type: chroma.Text, Value: line}}
		} else {
			// Empty lines get an empty token slice
			result[i] = []Token{}
		}
	}

	return result
}

// splitTokensByLine takes a token iterator and groups tokens by line
func splitTokensByLine(iterator chroma.Iterator) [][]Token {
	var lines [][]Token
	var currentLine []Token

	for token := iterator(); token != chroma.EOF; token = iterator() {
		value := token.Value

		// Handle tokens that contain newlines
		if strings.Contains(value, "\n") {
			parts := strings.Split(value, "\n")

			for i, part := range parts {
				if i > 0 {
					// Start a new line
					lines = append(lines, currentLine)
					currentLine = []Token{}
				}

				if part != "" {
					currentLine = append(currentLine, Token{
						Type:  token.Type,
						Value: part,
					})
				}
			}
		} else {
			// Token doesn't contain newlines - add to current line
			currentLine = append(currentLine, Token{
				Type:  token.Type,
				Value: value,
			})
		}
	}

	// Add the last line
	lines = append(lines, currentLine)

	return lines
}
