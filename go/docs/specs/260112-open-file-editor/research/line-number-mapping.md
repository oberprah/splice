# Research: Line Number Mapping in Aligned Diff

## Alignment Structure

**Location:** `/home/user/splice/internal/domain/diff/alignment.go:47-107`

The `AlignedFileDiff` contains:
- `Left` and `Right`: FileContent objects with lines (0-indexed)
- `Alignments`: Array of Alignment implementations, one per viewport row

Alignment is a sum type with four implementations:

```go
UnchangedAlignment {LeftIdx, RightIdx}    // Same line in both versions
ModifiedAlignment  {LeftIdx, RightIdx}    // Changed line (paired)
RemovedAlignment   {LeftIdx}               // Only in old version (no RightIdx)
AddedAlignment     {RightIdx}              // Only in new version (no LeftIdx)
```

**Key insight:** Each alignment occupies one display row. The Alignments array index directly corresponds to the viewport line position.

## DiffState Viewport Data

**Location:** `/home/user/splice/internal/ui/states/diff/state.go:9-35`

```go
type State struct {
    ViewportStart int                      // Current scroll position (index into Alignments)
    Diff *diff.AlignedFileDiff            // Contains Alignments and Right FileContent
}
```

## Mapping Algorithm

To map from viewport position to file line number:

1. **Get the alignment at viewport position:**
   ```go
   alignment := s.Diff.Alignments[s.ViewportStart]
   ```

2. **Extract the RightIdx based on alignment type:**
   ```go
   switch a := alignment.(type) {
   case UnchangedAlignment:
       rightIdx := a.RightIdx
   case ModifiedAlignment:
       rightIdx := a.RightIdx
   case AddedAlignment:
       rightIdx := a.RightIdx
   case RemovedAlignment:
       // No right line - this is a deleted line
       // Need to find adjacent alignment with RightIdx
   }
   ```

3. **Convert 0-indexed RightIdx to 1-indexed line number:**
   ```go
   lineNo := rightIdx + 1
   ```

**Code Reference:** LineNo conversion at `/home/user/splice/internal/domain/diff/alignment.go:37-41`

## Handling RemovedAlignment (Deleted Lines)

When `ViewportStart` points to a `RemovedAlignment`:
- The alignment has no RightIdx (no corresponding line in new file)
- The left column shows removed content, right column is blank

**Solution:**
- Skip forward to find the next alignment with a RightIdx
- If not found, skip backward
- Use that line number for the editor

## Edge Cases

1. **RemovedAlignment at viewport:** Find adjacent alignment with RightIdx
2. **AddedAlignment:** Has RightIdx, use directly
3. **Empty diffs:** Handle gracefully with error message
4. **Header lines:** 2 header lines exist but don't affect ViewportStart (already accounts for this)

**Code References:**
- View rendering: `/home/user/splice/internal/ui/states/diff/view.go:52-70`
- Jump navigation: `/home/user/splice/internal/ui/states/diff/update.go:86-128`
- RemovedAlignment rendering: `/home/user/splice/internal/ui/states/diff/view.go:179-184`

## Practical Implementation

For the editor feature:

```go
func (s *State) getCurrentFileLineNumber() (int, error) {
    if s.ViewportStart >= len(s.Diff.Alignments) {
        return 0, fmt.Errorf("viewport out of range")
    }

    alignment := s.Diff.Alignments[s.ViewportStart]

    switch a := alignment.(type) {
    case UnchangedAlignment:
        return a.RightIdx + 1, nil
    case ModifiedAlignment:
        return a.RightIdx + 1, nil
    case AddedAlignment:
        return a.RightIdx + 1, nil
    case RemovedAlignment:
        // Find next alignment with RightIdx
        for i := s.ViewportStart + 1; i < len(s.Diff.Alignments); i++ {
            if rightIdx := getRightIdx(s.Diff.Alignments[i]); rightIdx >= 0 {
                return rightIdx + 1, nil
            }
        }
        return 1, nil // Default to line 1 if no RightIdx found
    }
}
```
