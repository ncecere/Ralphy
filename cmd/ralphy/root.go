package main

import (
	"fmt"
	"os"

	"github.com/ncecere/ralphy/internal/app"
	"github.com/ncecere/ralphy/internal/config"
	"github.com/spf13/cobra"
)

var configFile string

var rootCmd = &cobra.Command{
	Use:           "ralphy",
	Short:         "Autonomous AI coding loop",
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       config.Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithFile(configFile)
		if err != nil {
			return err
		}

		if err := config.ApplyFlagOverrides(cmd, &cfg); err != nil {
			return err
		}

		return app.Run(cmd.Context(), cfg)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Path to config file (overrides default locations)")
	config.BindFlags(rootCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
