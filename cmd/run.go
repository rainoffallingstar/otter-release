package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xdxtools/xdxtools-go/internal/config"
	"github.com/xdxtools/xdxtools-go/internal/engine"
	"github.com/xdxtools/xdxtools-go/internal/input"
	"github.com/xdxtools/xdxtools-go/internal/logger"
	"github.com/xdxtools/xdxtools-go/internal/workflow"
)

var (
	runConfigFile string
	runEngine     string
	dryRun        bool
	verboseRun    bool
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a workflow",
	Long: `Execute a bioinformatics workflow with the specified configuration.

Examples:
  xdxtools run --config config.yaml
  xdxtools run --config config.yaml --engine slurm
  xdxtools run --config config.yaml --dry-run`,
	RunE: runRun,
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&runConfigFile, "config", "c", "config.yaml", "Configuration file")
	runCmd.Flags().StringVarP(&runEngine, "engine", "e", "auto", "Execution engine (auto/slurm/local)")
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Perform a dry run without executing")
	runCmd.Flags().BoolVarP(&verboseRun, "verbose", "v", false, "Verbose output")
}

func runRun(cmd *cobra.Command, args []string) error {
	// Load configuration
	loader := config.NewLoader(runConfigFile)
	cfg, err := loader.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override engine if specified
	if runEngine != "auto" {
		cfg.Engine.Type = runEngine
	}

	// Scan FASTQ files
	scanner := input.NewScanner(&input.ScanOptions{
		FastqDir: cfg.Input.FastqDir,
		Suffix1:  cfg.Suffix1,
		Suffix2:  cfg.Suffix2,
	})

	samples, err := scanner.Scan()
	if err != nil {
		return fmt.Errorf("failed to scan FASTQ files: %w", err)
	}

	if len(samples) == 0 {
		return fmt.Errorf("no FASTQ files found")
	}

	// Pair samples
	pairedSamples, err := scanner.PairSamples(samples, nil)
	if err != nil {
		return fmt.Errorf("failed to pair samples: %w", err)
	}

	// Get sample names
	sampleNames := input.GetSampleNames(pairedSamples)

	// Create workflow
	w := workflow.NewWorkflow(cfg, sampleNames)

	// Create engine
	eng, err := engine.CreateEngineFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}

	w.SetEngine(eng)

	// Create workflow manager
	manager := workflow.NewManager(w)

	if dryRun {
		logger.Info("Dry run mode - no actual execution")
		logger.Infof("Configuration loaded: %s", cfg.Mode)
		logger.Infof("Samples found: %d", len(sampleNames))
		logger.Infof("Engine: %s", cfg.Engine.Type)
		return nil
	}

	// Initialize workflow
	if err := manager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize workflow: %w", err)
	}

	// Execute workflow
	if err := manager.ExecuteAll(); err != nil {
		return fmt.Errorf("workflow execution failed: %w", err)
	}

	logger.Info("Workflow completed successfully")

	return nil
}
