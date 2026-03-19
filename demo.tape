# Splice demo recording
Output assets/demo.gif

Set Width 1400
Set Height 800
Set FontSize 20
Set Theme "Catppuccin Mocha"
Set Framerate 30

Require splice
Require git

# Launch splice (hidden so we don't show typing)
Hide
Type "splice"
Enter
Sleep 3s
Show

Sleep 1.5s

# --- Log view: scroll down to "Smooth scrolling" commit ---
Type "j"
Sleep 0.3s
Type "j"
Sleep 0.3s
Type "j"
Sleep 0.3s
Type "j"
Sleep 0.3s
Type "j"
Sleep 0.3s
Type "j"
Sleep 0.3s
Type "j"
Sleep 0.8s

# --- Enter files view for that commit ---
Enter
Sleep 2s

# --- Show commit message expansion ---
Type "m"
Sleep 2s

# Collapse it back
Type "m"
Sleep 1.5s

# --- Go back to log ---
Type "q"
Sleep 1.5s

# --- Multi-select: enter visual mode and select a range ---
Type "v"
Sleep 0.5s
Type "k"
Sleep 0.3s
Type "k"
Sleep 0.3s
Type "k"
Sleep 0.8s

# Open the multi-commit files view
Enter
Sleep 2s

# --- Collapse folders ---
# Cursor starts on first file (child of docs/spec/)
# Move up to docs/spec/ folder
Type "k"
Sleep 0.5s

# Collapse docs/spec/
Left
Sleep 1s

# Move down to go/ folder
Type "j"
Sleep 0.5s

# Collapse go/ — dramatic collapse of 405 files
Left
Sleep 1.5s

# Now the tree is compact: docs/spec (collapsed), go/ (collapsed), src/, tests/
# Navigate down to src/ area — should be just 2 down from go/
Type "j"
Sleep 0.3s

# We should now be on src/ — expand into it
Type "j"
Sleep 0.3s

# Navigate into src/app/ folder
Type "j"
Sleep 0.3s

# Open diff on first src file (should be a .rs file with real code changes)
Enter
Sleep 2s

# --- Diff view: scroll through code ---
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 0.2s
Type "j"
Sleep 1s

# Jump to next diff hunk
Type "n"
Sleep 1.5s

# Jump again
Type "n"
Sleep 1.5s

# Jump again
Type "n"
Sleep 1.5s

# Jump to previous
Type "p"
Sleep 1.5s

# Exit
Type "q"
Sleep 1s
Type "q"
Sleep 1s
Type "q"
Sleep 0.5s
