package diff

// ═══════════════════════════════════════════════════════════
// SEGMENT: Sum type representing contiguous regions of a diff
// ═══════════════════════════════════════════════════════════

// Segment represents a contiguous region of the diff.
// Each segment is either an unchanged region (content identical on both sides)
// or a hunk (region with changes).
//
// This is a sealed interface - only UnchangedSegment and HunkSegment implement it.
// Use a type switch to handle all cases in scrolling and rendering logic.
type Segment interface {
	segment() // unexported marker method prevents external implementation
}

// UnchangedSegment represents a region where the content is identical on both sides.
// During scrolling, both panels scroll together through this region.
type UnchangedSegment struct {
	LeftStart  int // Start index into Left.Lines
	RightStart int // Start index into Right.Lines
	Count      int // Number of lines (same for both sides)
}

// HunkSegment represents a region with changed content.
// Each side may have a different number of lines.
// During scrolling, panels use differential scrolling when this segment is centered.
type HunkSegment struct {
	LeftLines  []HunkLine // Lines on left (removals + modified-old)
	RightLines []HunkLine // Lines on right (additions + modified-new)
}

// Marker method implementations - seal the Segment interface
func (UnchangedSegment) segment() {}
func (HunkSegment) segment()      {}

// ═══════════════════════════════════════════════════════════
// HUNK LINE: Individual line within a hunk
// ═══════════════════════════════════════════════════════════

// HunkLine represents a single line within a hunk segment.
// It references the actual line content via an index into the FileContent.Lines slice.
type HunkLine struct {
	SourceIdx int          // Index into Left.Lines or Right.Lines
	Type      HunkLineType // Added, Removed, or Modified
}

// HunkLineType indicates the change type for a line within a hunk.
// This is distinct from LineType in parse.go which represents raw diff line types.
type HunkLineType int

const (
	HunkLineAdded    HunkLineType = iota // Line exists only in the new file
	HunkLineRemoved                      // Line exists only in the old file
	HunkLineModified                     // Line was modified (paired with a line on the other side)
)
