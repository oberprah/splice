package graph

// findInLanes searches for a hash in the active lanes.
// Returns the index where the hash was found, or -1 if not found.
func findInLanes(hash string, lanes []string) int {
	for i, h := range lanes {
		if h == hash {
			return i
		}
	}
	return -1
}

// findEmptyLane finds the first empty (empty string) slot in lanes.
// Returns the index of the empty slot, or -1 if no empty slots exist.
func findEmptyLane(lanes []string) int {
	for i, h := range lanes {
		if h == "" {
			return i
		}
	}
	return -1
}

// assignColumn determines which column a commit should occupy.
// It first checks if the commit's hash already exists in a lane (branch continuation),
// then looks for an empty slot, and finally appends a new column if needed.
// Returns the column index and the potentially expanded lanes slice.
func assignColumn(hash string, lanes []string) (int, []string) {
	// First, check if this hash is already in a lane (we're continuing a branch)
	if idx := findInLanes(hash, lanes); idx >= 0 {
		return idx, lanes
	}

	// Look for an empty slot to reuse
	if idx := findEmptyLane(lanes); idx >= 0 {
		lanes[idx] = hash
		return idx, lanes
	}

	// No existing slot, append a new column
	lanes = append(lanes, hash)
	return len(lanes) - 1, lanes
}

// collapseTrailingEmpty removes trailing empty strings from lanes.
// This prevents unbounded width growth when branches complete.
func collapseTrailingEmpty(lanes []string) []string {
	// Find the last non-empty lane
	lastNonEmpty := -1
	for i := len(lanes) - 1; i >= 0; i-- {
		if lanes[i] != "" {
			lastNonEmpty = i
			break
		}
	}

	// Return slice up to and including last non-empty, or empty slice
	if lastNonEmpty < 0 {
		return []string{}
	}
	return lanes[:lastNonEmpty+1]
}
