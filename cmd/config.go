package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xdxtools/xdxtools-go/internal/config"
	"github.com/xdxtools/xdxtools-go/internal/logger"
)

var (
	validateConfigFile string
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long: `Configuration management commands for xdxtools.

Note: Use 'xdxtools create' to generate complete project configurations
with sample information, adapters, and directory structure.

Examples:
  xdxtools config validate --config my_config.yaml
  xdxtools create --fastq /data --mode RRBS  # Generates full config`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a configuration file",
	Long: `Validate a configuration file for correctness.

This command checks if a configuration file is properly formatted
and contains all required fields for workflow execution.

Examples:
  xdxtools config validate --config my_config.yaml`,
	RunE: runConfigValidate,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(validateCmd)

	// Flags for validate command
	validateCmd.Flags().StringVarP(&validateConfigFile, "config", "c", "config.yaml", "Configuration file to validate")
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
	// Load configuration
	loader := config.NewLoader(validateConfigFile)
	cfg, err := loader.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	logger.Info("Configuration validation passed")
	logger.Infof("Mode: %s", cfg.Mode)
	logger.Infof("Engine: %s", cfg.Engine.Type)
	logger.Infof("FASTQ Directory: %s", cfg.Input.FastqDir)
	logger.Infof("Samples: %d", cfg.Parallel.Workers)

	return nil
}
