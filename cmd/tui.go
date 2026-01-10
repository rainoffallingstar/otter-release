package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xdxtools/xdxtools-go/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive TUI interface",
	Long: `Launch the interactive Terminal User Interface (TUI) for xdxtools.

The TUI provides an interactive menu-driven interface for managing
bioinformatics workflows. You can:

- Create new projects from FASTQ files
- Run workflows with real-time monitoring
- View workflow status and progress
- Manage configuration files

Examples:
  xdxtools tui
  xdxtools tui --config /path/to/config.yaml`,
	RunE: runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
	tuiCmd.Flags().String("config", "", "Optional configuration file to load on startup")
}

func runTUI(cmd *cobra.Command, args []string) error {
	// Load config if provided
	configPath, _ := cmd.Flags().GetString("config")
	if configPath != "" {
		fmt.Printf("Loading configuration from: %s\n", configPath)
		// TODO: Load configuration
	}

	// Start the TUI
	if err := tui.StartTUI(); err != nil {
		return fmt.Errorf("TUI execution failed: %w", err)
	}

	return nil
}
