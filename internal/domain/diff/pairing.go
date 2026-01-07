package diff

import (
	"regexp"
	"sort"
)

// similarityThreshold is the minimum Dice coefficient score required for two lines
// to be considered a pair. Lines with similarity below this threshold are left unpaired.
// Range: 0.0 (no overlap) to 1.0 (identical). Value of 0.5 is a reasonable default
// for code diffs where we want fairly strong similarity signals.
const similarityThreshold = 0.5

// tokenPattern matches non-alphanumeric characters for splitting text into tokens.
// This is used to tokenize lines for similarity comparison.
var tokenPattern = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// tokenize splits the input text into tokens by splitting on non-alphanumeric characters.
// Returns a slice of lowercase tokens, ignoring empty strings.
//
// Examples:
//   - "fmt.Println(name)" -> ["fmt", "println", "name"]
//   - "x == y" -> ["x", "y"]
//   - "" -> []
//   - "   " -> []
func tokenize(text string) []string {
	if text == "" {
		return []string{}
	}

	parts := tokenPattern.Split(text, -1)
	tokens := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			// Normalize to lowercase for case-insensitive comparison
			tokens = append(tokens, part)
		}
	}
	return tokens
}

// diceSimilarity computes the Dice coefficient similarity score between two lines.
// The score is based on token overlap and ranges from 0.0 (no overlap) to 1.0 (identical).
//
// Formula: Dice(A, B) = 2 × |tokens(A) ∩ tokens(B)| / (|tokens(A)| + |tokens(B)|)
//
// Special cases:
//   - If both lines are empty, returns 0.0 (empty lines should not match)
//   - If either line is empty (but not both), returns 0.0
//   - If lines have no common tokens, returns 0.0
func diceSimilarity(line1, line2 *AlignedLine) float64 {
	text1 := line1.Text()
	text2 := line2.Text()

	// Special case: empty lines should not match with each other
	// This prevents false positives from blank lines
	if text1 == "" && text2 == "" {
		return 0.0
	}

	tokens1 := tokenize(text1)
	tokens2 := tokenize(text2)

	// If either line has no tokens, they have no similarity
	if len(tokens1) == 0 || len(tokens2) == 0 {
		return 0.0
	}

	// Count token frequencies for both lines
	freq1 := make(map[string]int)
	freq2 := make(map[string]int)

	for _, token := range tokens1 {
		freq1[token]++
	}
	for _, token := range tokens2 {
		freq2[token]++
	}

	// Calculate intersection size (sum of min frequencies for common tokens)
	intersection := 0
	for token, count1 := range freq1 {
		if count2, exists := freq2[token]; exists {
			if count1 < count2 {
				intersection += count1
			} else {
				intersection += count2
			}
		}
	}

	// Dice coefficient: 2 × intersection / (size1 + size2)
	return 2.0 * float64(intersection) / float64(len(tokens1)+len(tokens2))
}

// linePair represents a potential pairing between a removed line and an added line
// along with their similarity score.
type linePair struct {
	removedIdx int     // Index in the removed lines slice
	addedIdx   int     // Index in the added lines slice
	score      float64 // Dice coefficient similarity score
}

// pairLines takes slices of removed and added lines and returns pairings between them
// based on similarity scores. Uses a greedy matching algorithm:
//  1. Compute similarity scores for all (removed, added) pairs
//  2. Filter pairs below the similarity threshold (0.5)
//  3. Sort remaining pairs by score in descending order
//  4. Greedily match: take highest-scoring pair, mark both lines as used, repeat
//
// Returns two values:
//   - pairs: slice of matched (removedIdx, addedIdx) pairs
//   - unpairedRemoved: indices of removed lines that couldn't be paired
//   - unpairedAdded: indices of added lines that couldn't be paired
//
// Note: indices in the returned pairs are relative to the input slices, not absolute file positions.
func pairLines(removed, added []AlignedLine) (pairs [][2]int, unpairedRemoved, unpairedAdded []int) {
	if len(removed) == 0 || len(added) == 0 {
		// No lines to pair, return all as unpaired
		for i := range removed {
			unpairedRemoved = append(unpairedRemoved, i)
		}
		for i := range added {
			unpairedAdded = append(unpairedAdded, i)
		}
		return nil, unpairedRemoved, unpairedAdded
	}

	// Step 1: Compute similarity matrix and collect candidate pairs
	candidates := make([]linePair, 0, len(removed)*len(added))
	for i := range removed {
		for j := range added {
			score := diceSimilarity(&removed[i], &added[j])
			if score >= similarityThreshold {
				candidates = append(candidates, linePair{
					removedIdx: i,
					addedIdx:   j,
					score:      score,
				})
			}
		}
	}

	// Step 2: Sort by score (descending)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	// Step 3: Greedy matching
	usedRemoved := make(map[int]bool)
	usedAdded := make(map[int]bool)

	for _, candidate := range candidates {
		// Skip if either line is already paired
		if usedRemoved[candidate.removedIdx] || usedAdded[candidate.addedIdx] {
			continue
		}

		// Create the pair
		pairs = append(pairs, [2]int{candidate.removedIdx, candidate.addedIdx})
		usedRemoved[candidate.removedIdx] = true
		usedAdded[candidate.addedIdx] = true
	}

	// Step 4: Collect unpaired lines
	for i := range removed {
		if !usedRemoved[i] {
			unpairedRemoved = append(unpairedRemoved, i)
		}
	}
	for i := range added {
		if !usedAdded[i] {
			unpairedAdded = append(unpairedAdded, i)
		}
	}

	return pairs, unpairedRemoved, unpairedAdded
}
