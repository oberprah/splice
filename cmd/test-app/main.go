package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Checkpoint represents a point where a screenshot should be taken
type Checkpoint struct {
	Name    string   // Optional name for the checkpoint
	Actions []string // Actions to perform before this checkpoint
}

type Config struct {
	Width       int
	Height      int
	FontSize    int
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

	// Generate tape file
	tapeFile := filepath.Join(outputDir, "test.tape")
	if err := generateTapeFile(config, outputDir, tapeFile); err != nil {
		return fmt.Errorf("failed to generate tape file: %w", err)
	}

	// Run VHS
	if err := runVHS(tapeFile); err != nil {
		return fmt.Errorf("failed to run VHS: %w", err)
	}

	fmt.Printf("Screenshots saved to: %s\n", outputDir)
	return nil
}

func parseArgs() (*Config, error) {
	config := &Config{
		Width:    1200,
		Height:   800,
		FontSize: 12,
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

		case arg == "--font-size":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--font-size requires a value")
			}
			i++
			_, err := fmt.Sscanf(args[i], "%d", &config.FontSize)
			if err != nil {
				return nil, fmt.Errorf("invalid font-size value: %s", args[i])
			}

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

func generateTapeFile(config *Config, outputDir, tapeFile string) error {
	var sb strings.Builder

	// VHS settings - use paths relative to project root (no leading /)
	sb.WriteString(fmt.Sprintf("Output %s\n", filepath.Join(outputDir, "recording.gif")))
	sb.WriteString(fmt.Sprintf("Set Width %d\n", config.Width))
	sb.WriteString(fmt.Sprintf("Set Height %d\n", config.Height))
	sb.WriteString(fmt.Sprintf("Set FontSize %d\n", config.FontSize))
	sb.WriteString("Set Shell bash\n")
	sb.WriteString("\n")

	// Start the application
	sb.WriteString("Type \"./splice\"\n")
	sb.WriteString("Enter\n")
	sb.WriteString("Sleep 1s\n")
	sb.WriteString("\n")

	// Initial screenshot if first checkpoint has no actions
	checkpointNum := 1
	if len(config.Checkpoints) > 0 && len(config.Checkpoints[0].Actions) == 0 {
		filename := fmt.Sprintf("%03d", checkpointNum)
		if config.Checkpoints[0].Name != "" {
			filename += "-" + config.Checkpoints[0].Name
		}
		filename += ".png"
		screenshotPath := filepath.Join(outputDir, filename)
		sb.WriteString(fmt.Sprintf("Screenshot %s\n", screenshotPath))
		sb.WriteString("\n")
		checkpointNum++
	}

	// Process each checkpoint
	for i, checkpoint := range config.Checkpoints {
		// Skip first checkpoint if it had no actions (already handled above)
		if i == 0 && len(checkpoint.Actions) == 0 {
			continue
		}

		// Perform actions first
		for _, action := range checkpoint.Actions {
			if err := writeAction(&sb, action); err != nil {
				return err
			}
			sb.WriteString("Sleep 200ms\n")
		}

		// Then take screenshot
		filename := fmt.Sprintf("%03d", checkpointNum)
		if checkpoint.Name != "" {
			filename += "-" + checkpoint.Name
		}
		filename += ".png"

		screenshotPath := filepath.Join(outputDir, filename)
		sb.WriteString(fmt.Sprintf("Screenshot %s\n", screenshotPath))
		sb.WriteString("\n")
		checkpointNum++
	}

	return os.WriteFile(tapeFile, []byte(sb.String()), 0644)
}

func writeAction(sb *strings.Builder, action string) error {
	// Parse action for special keys like <enter>, <esc>, <ctrl-c>
	if strings.HasPrefix(action, "<") && strings.HasSuffix(action, ">") {
		// Special key
		key := strings.ToLower(action[1 : len(action)-1])
		switch key {
		case "enter":
			sb.WriteString("Enter\n")
		case "esc", "escape":
			sb.WriteString("Escape\n")
		case "tab":
			sb.WriteString("Tab\n")
		case "space":
			sb.WriteString("Space\n")
		case "backspace":
			sb.WriteString("Backspace\n")
		case "up":
			sb.WriteString("Up\n")
		case "down":
			sb.WriteString("Down\n")
		case "left":
			sb.WriteString("Left\n")
		case "right":
			sb.WriteString("Right\n")
		case "ctrl-c":
			sb.WriteString("Ctrl+C\n")
		default:
			return fmt.Errorf("unknown special key: %s", action)
		}
	} else {
		// Regular key sequence
		fmt.Fprintf(sb, "Type \"%s\"\n", escapeQuotes(action))
	}
	return nil
}

func escapeQuotes(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}

func runVHS(tapeFile string) error {
	fmt.Println("Running VHS...")
	cmd := exec.Command("vhs", tapeFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
