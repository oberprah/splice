package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Checkpoint represents a point where a text snapshot should be taken
type Checkpoint struct {
	Name    string   // Optional name for the checkpoint
	Actions []string // Actions to perform before this checkpoint
}

type Config struct {
	Width       int    // Terminal width in columns
	Height      int    // Terminal height in rows
	Format      string // Output format: "text", "png", or "both"
	Checkpoints []Checkpoint
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	config, err := parseArgs()
	if err != nil {
		return err
	}

	// Build splice first
	if err := buildSplice(); err != nil {
		return fmt.Errorf("failed to build splice: %w", err)
	}

	// Create output directory
	outputDir, err := createOutputDir()
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Printf("Output directory: %s\n", outputDir)

	// Run test with tmux
	if err := runWithTmux(config, outputDir); err != nil {
		return fmt.Errorf("failed to run test: %w", err)
	}

	// Print appropriate message based on format
	switch config.Format {
	case "text":
		fmt.Printf("Text snapshots saved to: %s\n", outputDir)
	case "png":
		fmt.Printf("PNG images saved to: %s\n", outputDir)
	case "both":
		fmt.Printf("Text snapshots and PNG images saved to: %s\n", outputDir)
	}
	return nil
}

func parseArgs() (*Config, error) {
	config := &Config{
		Width:  120,    // Terminal columns
		Height: 40,     // Terminal rows
		Format: "both", // Default to both text and PNG
	}

	args := os.Args[1:]
	var currentActions []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "-c" || arg == "--checkpoint":
			// Create checkpoint with actions collected so far
			checkpoint := Checkpoint{
				Actions: currentActions,
			}

			// Check if next arg is a checkpoint name
			if i+1 < len(args) && isCheckpointName(args[i+1]) {
				checkpoint.Name = args[i+1]
				i++ // Skip the name in next iteration
			}

			config.Checkpoints = append(config.Checkpoints, checkpoint)
			currentActions = nil

		case arg == "--width":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--width requires a value")
			}
			i++
			_, err := fmt.Sscanf(args[i], "%d", &config.Width)
			if err != nil {
				return nil, fmt.Errorf("invalid width value: %s", args[i])
			}

		case arg == "--height":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--height requires a value")
			}
			i++
			_, err := fmt.Sscanf(args[i], "%d", &config.Height)
			if err != nil {
				return nil, fmt.Errorf("invalid height value: %s", args[i])
			}

		case arg == "--format":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--format requires a value (text, png, or both)")
			}
			i++
			format := args[i]
			if format != "text" && format != "png" && format != "both" {
				return nil, fmt.Errorf("invalid format value: %s (must be text, png, or both)", format)
			}
			config.Format = format

		case strings.HasPrefix(arg, "-"):
			return nil, fmt.Errorf("unknown flag: %s", arg)

		default:
			// Regular action (key sequence)
			currentActions = append(currentActions, arg)
		}
	}

	if len(config.Checkpoints) == 0 {
		return nil, fmt.Errorf("no checkpoints specified, use -c to add checkpoints")
	}

	return config, nil
}

// isCheckpointName returns true if the string looks like a checkpoint name
// rather than an action/key sequence
func isCheckpointName(s string) bool {
	// Starts with < means it's a special key like <enter>
	if strings.HasPrefix(s, "<") {
		return false
	}

	// Contains - or _ likely means it's a name like "after-nav" or "initial_state"
	if strings.Contains(s, "-") || strings.Contains(s, "_") {
		return true
	}

	// Single lowercase word without repetition is likely a name
	// e.g., "initial", "final", "loaded"
	if len(s) > 0 && len(s) < 20 {
		// Check if it's all same character repeated (like 'jjj') - that's an action
		allSame := true
		first := rune(s[0])
		for _, r := range s {
			if r != first {
				allSame = false
				break
			}
		}
		if allSame && len(s) > 1 {
			return false // It's an action like 'jjj'
		}
		// Otherwise it's likely a name
		return true
	}

	return false
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
	dir := filepath.Join("test-output", timestamp)
	return dir, os.MkdirAll(dir, 0755)
}

func runWithTmux(config *Config, outputDir string) error {
	sessionName := fmt.Sprintf("splice-test-%d", time.Now().Unix())

	// Check if freeze is installed for PNG output
	if config.Format == "png" || config.Format == "both" {
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

	// Process checkpoints
	checkpointNum := 1
	for _, checkpoint := range config.Checkpoints {
		// Perform actions first
		for _, action := range checkpoint.Actions {
			if err := sendAction(sessionName, action); err != nil {
				return err
			}
			time.Sleep(200 * time.Millisecond)
		}

		// Base filename
		baseFilename := fmt.Sprintf("%03d", checkpointNum)
		if checkpoint.Name != "" {
			baseFilename += "-" + checkpoint.Name
		}

		// Capture text if requested
		if config.Format == "text" || config.Format == "both" {
			textPath := filepath.Join(outputDir, baseFilename+".txt")
			cmd := exec.Command("tmux", "capture-pane", "-t", sessionName, "-p")
			output, err := cmd.Output()
			if err != nil {
				return fmt.Errorf("failed to capture text pane: %w", err)
			}
			if err := os.WriteFile(textPath, output, 0644); err != nil {
				return fmt.Errorf("failed to write text snapshot: %w", err)
			}
		}

		// Capture PNG if requested
		if config.Format == "png" || config.Format == "both" {
			pngPath := filepath.Join(outputDir, baseFilename+".png")
			if err := capturePNG(sessionName, pngPath); err != nil {
				return fmt.Errorf("failed to capture PNG: %w", err)
			}
		}

		checkpointNum++
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
