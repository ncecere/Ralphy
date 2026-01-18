package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ncecere/ralphy/internal/config"
	"github.com/spf13/cobra"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available models for an AI engine",
	Long: `List available models for the specified AI engine.

Use --set-default to interactively select a model and save it as the default.
Use --model with --set-default to set a model non-interactively.

Examples:
  ralphy models --opencode              # List OpenCode models
  ralphy models --claude                # List Claude models
  ralphy models --opencode --set-default        # Interactive selection
  ralphy models --opencode --set-default --model anthropic/claude-sonnet-4-5 --local`,
	RunE: runModels,
}

func init() {
	modelsCmd.Flags().Bool("claude", false, "List Claude Code models")
	modelsCmd.Flags().Bool("opencode", false, "List OpenCode models")
	modelsCmd.Flags().Bool("codex", false, "List Codex models")
	modelsCmd.Flags().Bool("cursor", false, "List Cursor models")
	modelsCmd.Flags().Bool("set-default", false, "Set a model as the default for this engine")
	modelsCmd.Flags().String("model", "", "Model to set (use with --set-default for non-interactive)")
	modelsCmd.Flags().Bool("local", false, "Save to local config (./ralphy.yaml) instead of global")
	rootCmd.AddCommand(modelsCmd)
}

func runModels(cmd *cobra.Command, args []string) error {
	// Determine which engine was selected
	engine := ""
	if v, _ := cmd.Flags().GetBool("claude"); v {
		engine = config.AIEngineClaude
	}
	if v, _ := cmd.Flags().GetBool("opencode"); v {
		if engine != "" {
			return fmt.Errorf("only one engine flag allowed")
		}
		engine = config.AIEngineOpenCode
	}
	if v, _ := cmd.Flags().GetBool("codex"); v {
		if engine != "" {
			return fmt.Errorf("only one engine flag allowed")
		}
		engine = config.AIEngineCodex
	}
	if v, _ := cmd.Flags().GetBool("cursor"); v {
		if engine != "" {
			return fmt.Errorf("only one engine flag allowed")
		}
		engine = config.AIEngineCursor
	}

	if engine == "" {
		return fmt.Errorf("specify an engine: --claude, --opencode, --codex, or --cursor")
	}

	setDefault, _ := cmd.Flags().GetBool("set-default")
	modelFlag, _ := cmd.Flags().GetString("model")
	local, _ := cmd.Flags().GetBool("local")

	// Get the model list
	models, err := getModelList(engine)
	if err != nil {
		fmt.Printf("Could not get model list: %s\n", err)
		return nil // Exit 0 as requested
	}

	// If no set-default, just print the list
	if !setDefault {
		for _, m := range models {
			fmt.Println(m)
		}
		return nil
	}

	// Determine which model to set
	var selectedModel string

	if modelFlag != "" {
		// Non-interactive: use the provided model
		selectedModel = modelFlag
	} else {
		// Interactive: show numbered list and prompt
		if len(models) == 0 {
			return fmt.Errorf("no models available to select")
		}

		fmt.Printf("Available models for %s:\n\n", engine)
		for i, m := range models {
			fmt.Printf("  %d) %s\n", i+1, m)
		}
		fmt.Println()

		// Prompt for selection
		fmt.Print("Select model number (or type model name): ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
		input = strings.TrimSpace(input)

		// Try to parse as number
		if num, err := strconv.Atoi(input); err == nil {
			if num < 1 || num > len(models) {
				return fmt.Errorf("invalid selection: %d", num)
			}
			selectedModel = models[num-1]
		} else {
			// Use as model name directly
			selectedModel = input
		}
	}

	// Determine config path
	var configPath string
	if local {
		configPath = config.LocalConfigPath()
	} else {
		configPath = config.GlobalConfigPath()
	}

	// Set the model
	if err := config.SetEngineModel(configPath, engine, selectedModel); err != nil {
		return fmt.Errorf("failed to set model: %w", err)
	}

	fmt.Printf("Set default model for %s: %s\n", engine, selectedModel)
	fmt.Printf("Saved to: %s\n", configPath)

	return nil
}

// getModelList returns the list of available models for an engine.
func getModelList(engine string) ([]string, error) {
	switch engine {
	case config.AIEngineClaude:
		// Claude doesn't have a CLI list command; return known aliases
		return []string{"opus", "sonnet", "haiku"}, nil

	case config.AIEngineOpenCode:
		return getOpenCodeModels()

	case config.AIEngineCodex:
		return nil, fmt.Errorf("Codex CLI does not expose model listing. Run `codex --help` for options")

	case config.AIEngineCursor:
		return nil, fmt.Errorf("Cursor does not expose a CLI model list. Check Cursor settings")

	default:
		return nil, fmt.Errorf("unknown engine: %s", engine)
	}
}

// getOpenCodeModels runs `opencode models` and parses the output.
func getOpenCodeModels() ([]string, error) {
	// Check if opencode is installed
	_, err := exec.LookPath("opencode")
	if err != nil {
		return nil, fmt.Errorf("opencode not found in PATH")
	}

	cmd := exec.Command("opencode", "models")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running opencode models: %w", err)
	}

	var models []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			models = append(models, line)
		}
	}

	return models, nil
}
