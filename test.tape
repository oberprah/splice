# Test tape for splice navigation
Output test-output
Width 120
Height 40

# Wait for app to load
Sleep 1s
Textshot initial
Ansishot initial-ansi
Snapshot initial-visual

# Navigate down a few lines
Send jjj
Sleep 200ms
Textshot after-nav
Snapshot after-nav

# Open files view
Send <enter>
Sleep 200ms
Textshot files-view

# Quit
Send q
