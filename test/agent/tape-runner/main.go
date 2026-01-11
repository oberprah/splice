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
	Output string // Output directory
	Width  int    // Terminal width in columns
	Height int    // Terminal height in rows
}

// Command types
type OutputCmd struct{ path string }
type WidthCmd struct{ width int }
type HeightCmd struct{ height int }
type SendCmd struct{ keys string }
type SleepCmd struct{ duration time.Duration }
type TextshotCmd struct{ name string }
type AnishotCmd struct{ name string }
type SnapshotCmd struct{ name string }

func (c *OutputCmd) Execute(ctx *TapeContext) error {
	ctx.config.Output = c.path
	return nil
}

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
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`tape-runner - Test splice binary with tape files

USAGE:
    tape-runner <tape-file>
    tape-runner --help

DESCRIPTION:
    Runs splice in a tmux session and captures snapshots based on commands
    in a tape file. Tape files use a simple line-based format similar to VHS.

    Note: This tool always builds and tests splice from the current source code.

TAPE FILE FORMAT:
    # Comments start with #
    Output .test-output     # Output directory (required)
    Width 120               # Terminal columns (default: 120)
    Height 40               # Terminal rows (default: 40)

    Sleep 1s                # Wait before next command
    Send jjj                # Send keys to app
    Textshot initial        # Capture plain text (.txt)
    Ansishot initial        # Capture text with ANSI codes (.ansi)
    Snapshot initial        # Capture PNG image (.png)

COMMANDS:

  Configuration (applied at parse time):
    Output <path>           Output directory (required)
    Width <cols>            Terminal width (default: 120)
    Height <rows>           Terminal height (default: 40)

  Execution commands:
    Send <keys>             Send keys to the application
    Sleep <duration>        Wait (e.g., 200ms, 1s, 1.5s)
    Textshot [name]         Capture plain text without ANSI codes (.txt)
    Ansishot [name]         Capture text with ANSI color codes (.ansi)
    Snapshot [name]         Capture PNG image (.png)

SPECIAL KEYS (use with Send):
    <enter>    <esc>       <tab>       <space>     <backspace>
    <up>       <down>      <left>      <right>     <ctrl-c>

EXAMPLE TAPE FILE:
    # Test navigation
    Output .test-output
    Width 120
    Height 40

    Sleep 1s
    Textshot initial
    Snapshot initial

    Send jjj
    Sleep 200ms
    Textshot after-nav

    Send <enter>
    Sleep 200ms
    Textshot files-view

    Send q

OUTPUT:
    Creates numbered snapshots in Output/<timestamp>/:
    - 001-initial.txt       Plain text (~1-4KB)
    - 002-initial.png       PNG image (~500KB-1.7MB)
    - 003-after-nav.txt     Plain text
    - 004-files-view.txt    Plain text

DEPENDENCIES:
    - tmux (required)
    - freeze (required for Snapshot commands)
      Install: brew install charmbracelet/tap/freeze

EXAMPLES:
    # Run a test tape
    tape-runner test.tape

    # Create your own tape file
    cat > my-test.tape <<EOF
    Output .test-output
    Sleep 1s
    Textshot start
    Send j
    Textshot after-move
    EOF
    tape-runner my-test.tape
`)
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: %s <tape-file>\n\nFor help, run: %s --help", os.Args[0], os.Args[0])
	}

	// Handle --help flag
	if os.Args[1] == "--help" || os.Args[1] == "-h" {
		printHelp()
		return nil
	}

	tapeFile := os.Args[1]
	commands, config, err := parseTapeFile(tapeFile)
	if err != nil {
		return fmt.Errorf("failed to parse tape file: %w", err)
	}

	if config.Output == "" {
		return fmt.Errorf("tape file must specify Output directive")
	}

	// Build splice first
	if err := buildSplice(); err != nil {
		return fmt.Errorf("failed to build splice: %w", err)
	}

	// Create output directory with timestamp
	outputDir, err := createOutputDir(config.Output)
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

	config := &Config{
		Width:  120, // Default terminal columns
		Height: 40,  // Default terminal rows
	}

	var commands []TapeCommand
	scanner := bufio.NewScanner(file)
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
			return nil, nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		// Apply config commands immediately, queue execution commands
		switch c := cmd.(type) {
		case *OutputCmd, *WidthCmd, *HeightCmd:
			if err := c.Execute(&TapeContext{config: config}); err != nil {
				return nil, nil, fmt.Errorf("line %d: %w", lineNum, err)
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
	case "Output":
		if args == "" {
			return nil, fmt.Errorf("output requires a path")
		}
		return &OutputCmd{path: args}, nil

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

func createOutputDir(basePath string) (string, error) {
	timestamp := time.Now().Format("2006-01-02-150405")
	dir := filepath.Join(basePath, timestamp)
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

func checkFreezeInstalled() error {
	cmd := exec.Command("freeze", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("freeze not installed. Install with: brew install charmbracelet/tap/freeze")
	}
	return nil
}

func capturePNG(sessionName, outputPath string) error {
	// Capture pane with ANSI codes preserved
	captureCmd := exec.Command("tmux", "capture-pane", "-e", "-p", "-t", sessionName)

	// Pipe to freeze with terminal-optimized settings
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
