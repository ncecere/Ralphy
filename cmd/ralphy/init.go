package main

import (
	"fmt"

	"github.com/ncecere/ralphy/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a ralphy configuration file",
	Long: `Initialize a ralphy configuration file.

By default, creates a global config at ~/.config/ralphy/ralphy.yaml.
Use --local to create a project-local config at ./ralphy.yaml.

Config precedence (highest to lowest):
  1. Command-line flags
  2. Environment variables
  3. Local config (./ralphy.yaml)
  4. Global config (~/.config/ralphy/ralphy.yaml)`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().Bool("local", false, "Create config in current directory (./ralphy.yaml)")
	initCmd.Flags().Bool("force", false, "Overwrite existing config file")
	initCmd.Flags().String("engine", "", "Set default AI engine (claude, opencode, codex, cursor)")
	initCmd.Flags().String("model", "", "Set default model for the engine")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	local, _ := cmd.Flags().GetBool("local")
	force, _ := cmd.Flags().GetBool("force")
	engine, _ := cmd.Flags().GetString("engine")
	model, _ := cmd.Flags().GetString("model")

	var configPath string
	if local {
		configPath = config.LocalConfigPath()
	} else {
		configPath = config.GlobalConfigPath()
	}

	// Check if config already exists
	if config.ConfigExists(configPath) && !force {
		return fmt.Errorf("config file already exists at %s (use --force to overwrite)", configPath)
	}

	// Write default config
	if err := config.WriteDefaultConfig(configPath); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("Created config file: %s\n", configPath)

	// Set engine if provided
	if engine != "" {
		if err := config.SetDefaultEngine(configPath, engine); err != nil {
			return fmt.Errorf("failed to set engine: %w", err)
		}
		fmt.Printf("Set default engine: %s\n", engine)
	}

	// Set model if provided (requires engine)
	if model != "" {
		if engine == "" {
			return fmt.Errorf("--model requires --engine to be specified")
		}
		if err := config.SetEngineModel(configPath, engine, model); err != nil {
			return fmt.Errorf("failed to set model: %w", err)
		}
		fmt.Printf("Set default model for %s: %s\n", engine, model)
	}

	return nil
}
