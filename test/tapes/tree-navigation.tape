# Tree Navigation Test
# This tape file tests the tree file view feature
# Testing FR1-FR7 from requirements

Width 120
Height 40

# Wait for initial load
Sleep 1s
Textshot 01-initial-log-view

# Navigate to files view (press Enter on first commit)
Send <enter>
Sleep 500ms
Textshot 02-files-tree-expanded

# Test FR1 & FR2: Tree structure with folders first, alphabetically sorted
# Verify box-drawing characters are correct

# Test FR6: Navigate down through tree items
Send j
Sleep 200ms
Textshot 03-cursor-second-item

Send j
Sleep 200ms
Textshot 04-cursor-third-item

Send j
Sleep 200ms
Textshot 05-cursor-fourth-item

# Test FR6: Navigate back up
Send k
Sleep 200ms
Textshot 06-cursor-back-up

# Test FR6: Jump to top with 'g'
Send g
Sleep 200ms
Textshot 07-jump-to-top

# Test FR6: Jump to bottom with 'G'
Send G
Sleep 200ms
Textshot 08-jump-to-bottom

# Go back to top to test collapse/expand
Send g
Sleep 200ms
Textshot 09-back-to-top

# Test FR5 & FR6: Collapse folder with Enter
Send <enter>
Sleep 300ms
Textshot 10-folder-collapsed-enter

# Verify collapsed state shows: foldername/ +N -N (X files)

# Test FR6: Expand folder with Enter again
Send <enter>
Sleep 300ms
Textshot 11-folder-expanded-enter

# Test FR6: Collapse with Space key
Send <space>
Sleep 300ms
Textshot 12-folder-collapsed-space

# Test FR6: Expand with Space key
Send <space>
Sleep 300ms
Textshot 13-folder-expanded-space

# Test FR6: Collapse with left arrow
Send <left>
Sleep 300ms
Textshot 14-folder-collapsed-left

# Test FR6: Try left arrow again (should be no-op)
Send <left>
Sleep 300ms
Textshot 15-folder-collapsed-left-noop

# Test FR6: Expand with right arrow
Send <right>
Sleep 300ms
Textshot 16-folder-expanded-right

# Test FR6: Try right arrow again (should be no-op)
Send <right>
Sleep 300ms
Textshot 17-folder-expanded-right-noop

# Test FR6 & FR7: Navigate to a file and enter to view diff
# Navigate down to find a file (not a folder)
Send j
Sleep 200ms
Textshot 18-cursor-on-file

Send <enter>
Sleep 500ms
Textshot 19-diff-view

# Test back navigation
Send q
Sleep 300ms
Textshot 20-back-to-files

# Test FR3: Check collapsed paths (deep nested single-child folders)
# Navigate to find a collapsed path if it exists
Send j
Sleep 200ms
Textshot 21-navigate-to-collapsed-path

# Test collapsing and expanding collapsed paths
Send <enter>
Sleep 300ms
Textshot 22-toggle-collapsed-path

Send <enter>
Sleep 300ms
Textshot 23-toggle-back-collapsed-path

# Test edge case: Navigate through all items rapidly
Send g
Sleep 200ms
Send jjjjjjjj
Sleep 300ms
Textshot 24-rapid-navigation

# Test edge case: At bottom, try to go down (should stay at bottom)
Send G
Sleep 200ms
Textshot 25-at-bottom

Send j
Sleep 200ms
Textshot 26-try-down-at-bottom

# Test edge case: At top, try to go up (should stay at top)
Send g
Sleep 200ms
Textshot 27-at-top

Send k
Sleep 200ms
Textshot 28-try-up-at-top

# Test FR7: Visual selection indicator (→) is visible
Send j
Sleep 200ms
Textshot 29-cursor-indicator-visible

# Test mixed navigation and collapse states
Send <enter>
Sleep 200ms
Textshot 30-collapse-first

Send j
Sleep 200ms
Textshot 31-move-to-next-after-collapse

Send <enter>
Sleep 200ms
Textshot 32-collapse-second

# Expand all back
Send k
Sleep 200ms
Send <enter>
Sleep 200ms
Textshot 33-expand-first-back

Send j
Sleep 200ms
Send <enter>
Sleep 200ms
Textshot 34-expand-second-back

# Test viewport scrolling with many items
# (If there are enough items to require scrolling)
Send g
Sleep 200ms
Textshot 35-top-viewport

Send G
Sleep 200ms
Textshot 36-bottom-viewport

# Final state check
Send g
Sleep 200ms
Textshot 37-final-state

# Exit
Send q
Sleep 200ms
Send q
Sleep 200ms
Textshot 38-back-to-log

Send q
Sleep 200ms
