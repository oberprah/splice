# Comprehensive Tree Navigation Test
# This tape tests tree view with different commit structures

Width 120
Height 40

# Wait for initial load
Sleep 1s
Textshot 01-initial-log

# Test commit with simple file structure
Send <enter>
Sleep 500ms
Textshot 02-simple-tree

# Back to log
Send q
Sleep 300ms

# Navigate to different commit with more files
Send j
Sleep 200ms
Send j
Sleep 200ms
Send j
Sleep 200ms
Send j
Sleep 200ms
Textshot 03-log-at-fifth-commit

# Open files for commit with more complex structure
Send <enter>
Sleep 500ms
Textshot 04-complex-tree-expanded

# Test collapsing and expanding multiple folders
# Collapse first folder
Send <enter>
Sleep 300ms
Textshot 05-first-folder-collapsed

# Move to next folder
Send j
Sleep 200ms
Textshot 06-cursor-on-second-folder

# Collapse second folder
Send <enter>
Sleep 300ms
Textshot 07-second-folder-collapsed

# Expand first folder again
Send k
Sleep 200ms
Send <enter>
Sleep 300ms
Textshot 08-first-folder-expanded

# Navigate through items when some are collapsed
Send j
Sleep 200ms
Textshot 09-nav-mixed-state-1

Send j
Sleep 200ms
Textshot 10-nav-mixed-state-2

# Back to log to test another commit
Send q
Sleep 300ms

# Navigate to a commit with many files (e.g., the implementation commit)
Send j
Sleep 200ms
Send j
Sleep 200ms
Send j
Sleep 200ms
Send j
Sleep 200ms
Send j
Sleep 200ms
Textshot 11-log-many-files-commit

Send <enter>
Sleep 500ms
Textshot 12-many-files-tree

# Test viewport scrolling with many items
Send G
Sleep 300ms
Textshot 13-viewport-at-bottom

Send g
Sleep 300ms
Textshot 14-viewport-at-top

# Test rapid navigation
Send jjjjj
Sleep 300ms
Textshot 15-rapid-down

Send kkk
Sleep 300ms
Textshot 16-rapid-up

# Navigate to a file deep in tree and open it
Send jjjjjjjj
Sleep 300ms
Textshot 17-cursor-deep-item

Send <enter>
Sleep 500ms
Textshot 18-diff-view-deep-file

# Go back
Send q
Sleep 300ms
Textshot 19-back-from-diff

# Test collapsing parent folder when deep in tree
Send k
Sleep 200ms
Send k
Sleep 200ms
Send k
Sleep 200ms
Send k
Sleep 200ms
Send k
Sleep 200ms
Send k
Sleep 200ms
Send k
Sleep 200ms
Send k
Sleep 200ms
Textshot 20-navigate-to-parent

Send <enter>
Sleep 300ms
Textshot 21-parent-collapsed

# Expand again
Send <enter>
Sleep 300ms
Textshot 22-parent-expanded

# Test edge case: collapse all folders
Send g
Sleep 200ms
Send <enter>
Sleep 300ms
Textshot 23-collapse-all-1

Send j
Sleep 200ms
Send <enter>
Sleep 300ms
Textshot 24-collapse-all-2

Send j
Sleep 200ms
Send <enter>
Sleep 300ms
Textshot 25-collapse-all-3

# Now we should have a very compact view
Textshot 26-all-collapsed

# Expand all back
Send k
Sleep 200ms
Send k
Sleep 200ms
Send <enter>
Sleep 300ms
Textshot 27-expand-back-1

Send j
Sleep 200ms
Send <enter>
Sleep 300ms
Textshot 28-expand-back-2

Send j
Sleep 200ms
Send <enter>
Sleep 300ms
Textshot 29-expand-back-3

Textshot 30-all-expanded-again

# Test folder with collapsed path (single-child chain)
# Navigate back to log
Send q
Sleep 300ms

# Find a commit with a collapsed path
Send jjjjjjjj
Sleep 300ms
Textshot 31-find-collapsed-path-commit

Send <enter>
Sleep 500ms
Textshot 32-collapsed-path-tree

# Test toggling collapsed path
Send j
Sleep 200ms
Textshot 33-cursor-on-collapsed-path

Send <enter>
Sleep 300ms
Textshot 34-collapsed-path-toggled

Send <enter>
Sleep 300ms
Textshot 35-collapsed-path-back

# Test arrow keys on collapsed path
Send <left>
Sleep 300ms
Textshot 36-collapsed-path-left

Send <right>
Sleep 300ms
Textshot 37-collapsed-path-right

# Back to log
Send q
Sleep 300ms

# Test with E2E test commit (has test files)
Send j
Sleep 200ms
Send j
Sleep 200ms
Send j
Sleep 200ms
Textshot 38-test-files-commit

Send <enter>
Sleep 500ms
Textshot 39-test-files-tree

# Navigate through test structure
Send jjj
Sleep 300ms
Textshot 40-nav-test-files

# Collapse test folder
Send k
Sleep 200ms
Send k
Sleep 200ms
Send k
Sleep 200ms
Send <enter>
Sleep 300ms
Textshot 41-test-folder-collapsed

# Final cleanup
Send q
Sleep 300ms
Send q
Sleep 300ms
