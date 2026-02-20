package format

// ToShortHash returns the short version of a commit hash (7 chars)
func ToShortHash(hash string) string {
	if len(hash) >= 7 {
		return hash[:7]
	}
	return hash
}
