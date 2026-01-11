package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TapeCommand represents a command in the tape file
type TapeCommand interface {
	Execute(ctx *TapeContext) error
}

// TapeContext holds the execution state
type TapeContext struct {
	sessionName   string
	outputDir     string
	config        *Config
	checkpointNum int
}

// Config holds the tape configuration
type Config struct {
	Width  int // Terminal width in columns
	Height int // Terminal height in rows
}

// Command types
type WidthCmd struct{ width int }
type HeightCmd struct{ height int }
type SendCmd struct{ keys string }
type SleepCmd struct{ duration time.Duration }
type TextshotCmd struct{ name string }
type AnishotCmd struct{ name string }
type SnapshotCmd struct{ name string }

func (c *WidthCmd) Execute(ctx *TapeContext) error {
	ctx.config.Width = c.width
	return nil
}

func (c *HeightCmd) Execute(ctx *TapeContext) error {
	ctx.config.Height = c.height
	return nil
}

func (c *SendCmd) Execute(ctx *TapeContext) error {
	return sendAction(ctx.sessionName, c.keys)
}

func (c *SleepCmd) Execute(ctx *TapeContext) error {
	time.Sleep(c.duration)
	return nil
}

func (c *TextshotCmd) Execute(ctx *TapeContext) error {
	ctx.checkpointNum++
	baseFilename := fmt.Sprintf("%03d", ctx.checkpointNum)
	if c.name != "" {
		baseFilename += "-" + c.name
	}

	textPath := filepath.Join(ctx.outputDir, baseFilename+".txt")
	cmd := exec.Command("tmux", "capture-pane", "-t", ctx.sessionName, "-p")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to capture text pane: %w", err)
	}
	if err := os.WriteFile(textPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write text snapshot: %w", err)
	}

	return nil
}

func (c *AnishotCmd) Execute(ctx *TapeContext) error {
	ctx.checkpointNum++
	baseFilename := fmt.Sprintf("%03d", ctx.checkpointNum)
	if c.name != "" {
		baseFilename += "-" + c.name
	}

	ansiPath := filepath.Join(ctx.outputDir, baseFilename+".ansi")
	cmd := exec.Command("tmux", "capture-pane", "-e", "-p", "-t", ctx.sessionName)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to capture ansi pane: %w", err)
	}
	if err := os.WriteFile(ansiPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write ansi snapshot: %w", err)
	}

	return nil
}

func (c *SnapshotCmd) Execute(ctx *TapeContext) error {
	ctx.checkpointNum++
	baseFilename := fmt.Sprintf("%03d", ctx.checkpointNum)
	if c.name != "" {
		baseFilename += "-" + c.name
	}

	pngPath := filepath.Join(ctx.outputDir, baseFilename+".png")
	if err := capturePNG(ctx.sessionName, pngPath); err != nil {
		return fmt.Errorf("failed to capture PNG: %w", err)
	}

	return nil
}

func main() {
	// Ensure GOBIN is in PATH for freeze and other go-installed tools
	ensureGoBinInPath()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// ensureGoBinInPath adds $HOME/go/bin to PATH if not already present
func ensureGoBinInPath() {
	home := os.Getenv("HOME")
	if home == "" {
		return
	}

	gobin := filepath.Join(home, "go", "bin")
	currentPath := os.Getenv("PATH")

	// Check if already in PATH
	for _, dir := range strings.Split(currentPath, ":") {
		if dir == gobin {
			return
		}
	}

	// Add GOBIN to PATH for this process and all child processes
	if err := os.Setenv("PATH", gobin+":"+currentPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to add GOBIN to PATH: %v\n", err)
	}
}

func printHelp() {
	fmt.Print(`tape-runner - Test splice binary with tape commands

USAGE:
    ./run-tape - <<'EOF'
    ... tape commands ...
    EOF

    ./run-tape <tape-file>    # Alternative: read from file
    ./run-tape --help

DESCRIPTION:
    Runs splice in a tmux session and captures snapshots. Commands use a
    simple line-based format. Output saved to .test-output/<timestamp>/

COMMANDS:
    Width <cols>            Terminal width (default: 120)
    Height <rows>           Terminal height (default: 40)
    Sleep <duration>        Wait (e.g., 200ms, 1s, 1.5s)
    Send <keys>             Send keys to the application
    Textshot [name]         Capture plain text (.txt)
    Ansishot [name]         Capture with ANSI codes (.ansi)
    Snapshot [name]         Capture PNG image (.png)

SPECIAL KEYS:
    <enter> <esc> <tab> <space> <backspace>
    <up> <down> <left> <right> <ctrl-c>

EXAMPLE:
    ./run-tape - <<'EOF'
    # Test navigation
    Width 120
    Height 40

    Sleep 1s
    Textshot initial

    Send jjj
    Sleep 200ms
    Textshot after-nav

    Send <enter>
    Sleep 500ms
    Textshot files-view
    EOF
`)
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: %s - <<'EOF' ... EOF  OR  %s <tape-file>\n\nFor help, run: %s --help", os.Args[0], os.Args[0], os.Args[0])
	}

	// Handle --help flag
	if os.Args[1] == "--help" || os.Args[1] == "-h" {
		printHelp()
		return nil
	}

	tapeInput := os.Args[1]

	var commands []TapeCommand
	var config *Config
	var err error

	if tapeInput == "-" {
		// Read from stdin
		commands, config, err = parseTapeReader(os.Stdin, "<stdin>")
	} else {
		// Read from file
		commands, config, err = parseTapeFile(tapeInput)
	}

	if err != nil {
		return fmt.Errorf("failed to parse tape: %w", err)
	}

	// Check dependencies upfront (before building)
	if err := checkTmuxInstalled(); err != nil {
		return err
	}

	// Build splice first
	if err := buildSplice(); err != nil {
		return fmt.Errorf("failed to build splice: %w", err)
	}

	// Create output directory with timestamp
	outputDir, err := createOutputDir()
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Printf("Output directory: %s\n", outputDir)

	// Run test with tmux
	if err := runTape(commands, config, outputDir); err != nil {
		return fmt.Errorf("failed to run test: %w", err)
	}

	fmt.Printf("Snapshots saved to: %s\n", outputDir)
	return nil
}

// parseTapeFile parses a tape file and returns the commands and config
func parseTapeFile(path string) ([]TapeCommand, *Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	return parseTapeReader(file, path)
}

// parseTapeReader parses tape commands from an io.Reader
func parseTapeReader(reader *os.File, sourceName string) ([]TapeCommand, *Config, error) {
	config := &Config{
		Width:  120, // Default terminal columns
		Height: 40,  // Default terminal rows
	}

	var commands []TapeCommand
	scanner := bufio.NewScanner(reader)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		cmd, err := parseLine(line)
		if err != nil {
			return nil, nil, fmt.Errorf("%s:%d: %w", sourceName, lineNum, err)
		}

		// Apply config commands immediately, queue execution commands
		switch c := cmd.(type) {
		case *WidthCmd, *HeightCmd:
			if err := c.Execute(&TapeContext{config: config}); err != nil {
				return nil, nil, fmt.Errorf("%s:%d: %w", sourceName, lineNum, err)
			}
		default:
			commands = append(commands, cmd)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return commands, config, nil
}

// parseLine parses a single line into a command
func parseLine(line string) (TapeCommand, error) {
	// Match: Command <args>
	parts := strings.SplitN(line, " ", 2)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	command := parts[0]
	args := ""
	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}

	switch command {
	case "Width":
		if args == "" {
			return nil, fmt.Errorf("width requires a value")
		}
		width, err := strconv.Atoi(args)
		if err != nil {
			return nil, fmt.Errorf("invalid width value: %s", args)
		}
		return &WidthCmd{width: width}, nil

	case "Height":
		if args == "" {
			return nil, fmt.Errorf("height requires a value")
		}
		height, err := strconv.Atoi(args)
		if err != nil {
			return nil, fmt.Errorf("invalid height value: %s", args)
		}
		return &HeightCmd{height: height}, nil

	case "Send":
		if args == "" {
			return nil, fmt.Errorf("send requires keys")
		}
		return &SendCmd{keys: args}, nil

	case "Sleep":
		if args == "" {
			return nil, fmt.Errorf("sleep requires duration")
		}
		duration, err := parseDuration(args)
		if err != nil {
			return nil, fmt.Errorf("invalid sleep duration: %w", err)
		}
		return &SleepCmd{duration: duration}, nil

	case "Textshot":
		// Args are optional (name)
		return &TextshotCmd{name: args}, nil

	case "Ansishot":
		// Args are optional (name)
		return &AnishotCmd{name: args}, nil

	case "Snapshot":
		// Args are optional (name)
		return &SnapshotCmd{name: args}, nil

	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

// parseDuration parses durations like "500ms", "1s", "1.5s"
func parseDuration(s string) (time.Duration, error) {
	// Simple regex for common patterns
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)(ms|s)$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("invalid duration format (use 500ms or 1s)")
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, err
	}

	unit := matches[2]
	switch unit {
	case "ms":
		return time.Duration(value * float64(time.Millisecond)), nil
	case "s":
		return time.Duration(value * float64(time.Second)), nil
	default:
		return 0, fmt.Errorf("unsupported duration unit: %s", unit)
	}
}

func buildSplice() error {
	fmt.Println("Building splice...")
	cmd := exec.Command("go", "build", "-o", "splice", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func createOutputDir() (string, error) {
	timestamp := time.Now().Format("2006-01-02-150405")
	dir := filepath.Join(".test-output", timestamp)
	return dir, os.MkdirAll(dir, 0755)
}

func runTape(commands []TapeCommand, config *Config, outputDir string) error {
	sessionName := fmt.Sprintf("splice-test-%d", time.Now().Unix())

	// Check if freeze is installed if any Snapshot commands exist
	hasSnapshot := false
	for _, cmd := range commands {
		if _, ok := cmd.(*SnapshotCmd); ok {
			hasSnapshot = true
			break
		}
	}
	if hasSnapshot {
		if err := checkFreezeInstalled(); err != nil {
			return err
		}
	}

	// Start detached tmux session with splice
	fmt.Println("Starting tmux session...")
	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName,
		"-x", fmt.Sprintf("%d", config.Width),
		"-y", fmt.Sprintf("%d", config.Height),
		"./splice")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start tmux session: %w", err)
	}

	// Ensure cleanup on exit
	defer func() {
		_ = exec.Command("tmux", "kill-session", "-t", sessionName).Run()
	}()

	// Wait for app to initialize
	time.Sleep(1 * time.Second)

	// Execute commands
	ctx := &TapeContext{
		sessionName:   sessionName,
		outputDir:     outputDir,
		config:        config,
		checkpointNum: 0,
	}

	for _, cmd := range commands {
		if err := cmd.Execute(ctx); err != nil {
			return err
		}
	}

	return nil
}

func checkTmuxInstalled() error {
	cmd := exec.Command("tmux", "-V")
	if err := cmd.Run(); err != nil {
		// Try to install using unified installer
		installCmd := exec.Command("bash", "scripts/env-setup/install-tool.sh", "tmux")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("tmux installation failed: %w", err)
		}
		fmt.Println()
	}
	return nil
}

func checkFreezeInstalled() error {
	cmd := exec.Command("freeze", "--version")
	if err := cmd.Run(); err != nil {
		// Try to install using unified installer
		fmt.Println("🔧 Installing freeze for image snapshots (first time only)...")
		installCmd := exec.Command("bash", "scripts/env-setup/install-tool.sh", "freeze")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("freeze installation failed: %w", err)
		}
		fmt.Println()
	}
	return nil
}

func capturePNG(sessionName, outputPath string) error {
	// Capture pane with ANSI codes preserved
	captureCmd := exec.Command("tmux", "capture-pane", "-e", "-p", "-t", sessionName)

	// Pipe to freeze with terminal-optimized settings
	// (freeze is found via PATH, which includes GOBIN from ensureGoBinInPath)
	freezeCmd := exec.Command("freeze", "-o", outputPath,
		"--language", "txt", // Required for freeze to process terminal output
		"--font.family", "JetBrainsMono Nerd Font Mono", // Excellent monospace with box-drawing support
		"--font.size", "14",
		"--font.ligatures=false", // Terminals don't use ligatures
		"--window=false",         // Remove window chrome
		"--margin", "0",          // Remove extra margins
		"--padding", "10", // Minimal padding
	)

	// Connect pipes
	pipe, err := captureCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %w", err)
	}
	freezeCmd.Stdin = pipe

	// Start both commands
	if err := freezeCmd.Start(); err != nil {
		return fmt.Errorf("failed to start freeze: %w", err)
	}
	if err := captureCmd.Run(); err != nil {
		return fmt.Errorf("failed to capture pane: %w", err)
	}
	if err := freezeCmd.Wait(); err != nil {
		return fmt.Errorf("freeze failed: %w", err)
	}

	return nil
}

func sendAction(sessionName, action string) error {
	// Parse action for special keys like <enter>, <esc>, <ctrl-c>
	if strings.HasPrefix(action, "<") && strings.HasSuffix(action, ">") {
		// Special key
		key := strings.ToLower(action[1 : len(action)-1])
		var tmuxKey string
		switch key {
		case "enter":
			tmuxKey = "Enter"
		case "esc", "escape":
			tmuxKey = "Escape"
		case "tab":
			tmuxKey = "Tab"
		case "space":
			tmuxKey = "Space"
		case "backspace":
			tmuxKey = "BSpace"
		case "up":
			tmuxKey = "Up"
		case "down":
			tmuxKey = "Down"
		case "left":
			tmuxKey = "Left"
		case "right":
			tmuxKey = "Right"
		case "ctrl-c":
			tmuxKey = "C-c"
		default:
			return fmt.Errorf("unknown special key: %s", action)
		}
		cmd := exec.Command("tmux", "send-keys", "-t", sessionName, tmuxKey)
		return cmd.Run()
	}

	// Regular key sequence - send character by character to avoid issues
	for _, char := range action {
		cmd := exec.Command("tmux", "send-keys", "-t", sessionName, "-l", string(char))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to send key '%c': %w", char, err)
		}
	}
	return nil
}
