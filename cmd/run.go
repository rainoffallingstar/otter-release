package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xdxtools/xdxtools-go/internal/assets"
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
	runCondaEnv   string
	copyFastq     bool
	moveFastq     bool

	// SLURM global settings (acts as unified partition)
	slurmPartition string
	slurmCores     int
	slurmMemory    string

	// Per-step resources
	step1Cores     int
	step1Memory    string
	step1Partition string
	step2Cores     int
	step2Memory    string
	step2Partition string
	step3Cores     int
	step3Memory    string
	step3Partition string

	// Checker resources
	step2CheckerCores  int
	step2CheckerMemory string
	step3CheckerCores  int
	step3CheckerMemory string

	// Parallel control
	parallelJobs int

	// Dynamic pool load ratio (replaces fixed batching for SLURM)
	loadRatio float64

	// Legacy unified partition parameter (kept for backward compatibility)
	slurmUnifiedPartition string

	// FASTQ compression option
	compressFastq bool

	// Resume option
	resumeFlag bool

	// Asset integrity options
	runProjectDir string
	verifyAssets  bool
	strictAssets  bool
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
	runCmd.Flags().StringVar(&runCondaEnv, "conda-env", "", "Conda environment for Snakemake")
	runCmd.Flags().BoolVar(&copyFastq, "copy-fastq", false, "Copy FASTQ files to project data directory (default for testing)")
	runCmd.Flags().BoolVar(&moveFastq, "move-fastq", false, "Move FASTQ files to project data directory")
	runCmd.Flags().StringVar(&runProjectDir, "project-dir", "", "Project directory (defaults to directory containing --config)")
	runCmd.Flags().BoolVar(&verifyAssets, "verify-assets", true, "Verify workflow assets manifest before running (if present)")
	runCmd.Flags().BoolVar(&strictAssets, "strict-assets", false, "Fail fast if workflow assets differ from the manifest")

	// SLURM global settings
	runCmd.Flags().StringVar(&slurmPartition, "slurm-partition", "", "Default SLURM partition (overrides config)")
	runCmd.Flags().IntVar(&slurmCores, "slurm-cores", 0, "Default SLURM CPU cores")
	runCmd.Flags().StringVar(&slurmMemory, "slurm-memory", "", "Default SLURM memory (e.g., 16G)")

	// Per-step resources
	runCmd.Flags().IntVar(&step1Cores, "step1-cores", 0, "Step 1 CPU cores")
	runCmd.Flags().StringVar(&step1Memory, "step1-memory", "", "Step 1 memory (e.g., 8G)")
	runCmd.Flags().StringVar(&step1Partition, "step1-partition", "", "Step 1 partition")

	runCmd.Flags().IntVar(&step2Cores, "step2-cores", 0, "Step 2 CPU cores")
	runCmd.Flags().StringVar(&step2Memory, "step2-memory", "", "Step 2 memory (e.g., 32G)")
	runCmd.Flags().StringVar(&step2Partition, "step2-partition", "", "Step 2 partition")

	runCmd.Flags().IntVar(&step3Cores, "step3-cores", 0, "Step 3 CPU cores")
	runCmd.Flags().StringVar(&step3Memory, "step3-memory", "", "Step 3 memory (e.g., 16G)")
	runCmd.Flags().StringVar(&step3Partition, "step3-partition", "", "Step 3 partition")

	// Checker resources
	runCmd.Flags().IntVar(&step2CheckerCores, "step2-checker-cores", 0, "Step 2 checker CPU cores")
	runCmd.Flags().StringVar(&step2CheckerMemory, "step2-checker-memory", "", "Step 2 checker memory")
	runCmd.Flags().IntVar(&step3CheckerCores, "step3-checker-cores", 0, "Step 3 checker CPU cores")
	runCmd.Flags().StringVar(&step3CheckerMemory, "step3-checker-memory", "", "Step 3 checker memory")

	// Parallel control
	runCmd.Flags().IntVar(&parallelJobs, "parallel-jobs", 2, "Max parallel jobs (local/Snakemake)")

	// Dynamic pool load ratio for SLURM
	runCmd.Flags().Float64Var(&loadRatio, "load-ratio", 1.0,
		"Job pool load ratio (0.1-1.0). SLURM: slot_limit=floor(min(parallel-jobs, MaxSubmitJobs-1)×ratio). "+
			"Local: slot_limit=floor(parallel-jobs×ratio). Set to 0 to disable dynamic pool.")

	// Legacy unified partition parameter (kept for backward compatibility)
	runCmd.Flags().StringVar(&slurmUnifiedPartition, "slurm-unified-partition", "", "Unified SLURM partition for all steps (overrides individual step partitions)")

	// FASTQ compression option
	runCmd.Flags().BoolVar(&compressFastq, "compress-fastq", false, "Compress uncompressed FASTQ files to .gz format")

	// Resume option
	runCmd.Flags().BoolVarP(&resumeFlag, "resume", "r", false, "Resume from last completed step")
}

func runRun(cmd *cobra.Command, args []string) error {
	projectDir := runProjectDir
	if projectDir == "" {
		projectDir = filepath.Dir(runConfigFile)
		if projectDir == "" {
			projectDir = "."
		}
	}
	projectDir = filepath.Clean(projectDir)

	// Load configuration
	loader := config.NewLoader(runConfigFile)
	cfg, err := loader.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if verifyAssets {
		m, err := assets.LoadManifest(projectDir)
		if err != nil {
			if os.IsNotExist(err) {
				logger.Warnf("Assets manifest not found: %s (run `xdxtools assets stamp --project %s` to generate)", assets.ManifestPath(projectDir), projectDir)
			} else {
				logger.Warnf("Failed to load assets manifest: %v", err)
			}
		} else {
			deprecated := 0
			for _, e := range m.Entries {
				if e.Deprecated {
					deprecated++
				}
			}
			if deprecated > 0 {
				logger.Warnf("Deprecated assets present: %d files under R/ (still verified for integrity)", deprecated)
			}

			diff, err := assets.VerifyWorkflowAssets(projectDir, m)
			if err != nil {
				logger.Warnf("Assets manifest verification failed: %v", err)
			} else if !diff.IsClean() {
				msg := fmt.Sprintf("Assets differ from manifest (missing=%d extra=%d modified=%d)", len(diff.Missing), len(diff.Extra), len(diff.Modified))
				if strictAssets {
					return fmt.Errorf("%s", msg)
				}
				logger.Warn(msg)
			} else {
				logger.Infof("Assets verified against manifest: %s", assets.ManifestPath(projectDir))
			}
		}
	}

	// Override engine if specified
	if runEngine != "auto" {
		cfg.Engine.Type = runEngine
	}

	// Validate SLURM partition if specified
	if slurmPartition != "" {
		if err := engine.ValidateSlurmPartition(slurmPartition); err != nil {
			return fmt.Errorf("SLURM partition validation failed: %w", err)
		}
		cfg.Engine.Slurm.Partition = slurmPartition
	}

	// Override SLURM resources
	if slurmCores > 0 {
		cfg.Engine.Slurm.Cores = slurmCores
	}
	if slurmMemory != "" {
		cfg.Engine.Slurm.Memory = slurmMemory
	}

	// Compress FASTQ files if requested
	if compressFastq {
		logger.Info("Compressing uncompressed FASTQ files...")
		if err := compressFastqFiles(cfg.Input.FastqDir); err != nil {
			return fmt.Errorf("failed to compress FASTQ files: %w", err)
		}
	}

	// Scan FASTQ files
	scanner := input.NewScanner(&input.ScanOptions{
		FastqDir: cfg.Input.FastqDir,
		Suffix1:  cfg.Input.Suffix1,
		Suffix2:  cfg.Input.Suffix2,
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

	// Check for existing state and warn user
	state := workflow.NewState(cfg.Output.BaseDir, cfg.Workflow.JobID)
	if state.Exists() && !resumeFlag {
		logger.Warn("==============================================")
		logger.Warn("Found incomplete workflow state!")
		logger.Warn("To resume from last completed step, use --resume or -r flag")
		logger.Warn("To start from beginning, remove the state file:")
		logger.Warnf("  rm %s", state.GetFilePath())
		logger.Warn("==============================================")
	}

	// Set workflow options
	w.Options = &workflow.WorkflowOptions{
		DryRun: dryRun,
		Resume: resumeFlag,
	}

	// Build step resources from command line arguments
	stepResources := buildStepResources()

	// Sync partition from stepResources to cfg.Engine.Slurm.Partition
	// This ensures --slurm-unified-partition is passed to the engine
	if stepResources != nil && cfg.Engine.Slurm.Partition == "" {
		// Find the first non-empty partition from stepResources
		for _, res := range stepResources {
			if res != nil && res.Partition != "" {
				cfg.Engine.Slurm.Partition = res.Partition
				break
			}
		}
	}

	// Set step resources in config
	cfg.StepResources = stepResources

	// Create workflow manager
	manager := workflow.NewManager(w)

	// Set samples, resources, and parallel jobs for the manager
	manager.SetSamples(sampleNames)
	manager.SetStepResources(stepResources)
	manager.SetParallelJobs(parallelJobs)
	if loadRatio > 0 {
		manager.SetLoadRatio(loadRatio)
	}

	// Create engine (after setting resources so manager can make intelligent decisions)
	eng, err := engine.CreateEngineFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}

	w.SetEngine(eng)

	// Warn if --parallel-jobs explicitly set alongside --load-ratio for SLURM
	if loadRatio > 0 && cmd.Flags().Changed("parallel-jobs") {
		if eng.GetName().String() == "slurm" || eng.GetName().String() == "slurm_array" {
			logger.Warn("--parallel-jobs is deprecated for SLURM step2/3 when --load-ratio is set; load-ratio takes priority")
		}
	}

	// Validate resources before execution
	if err := validateResources(eng.GetName().String(), cfg, stepResources, parallelJobs); err != nil {
		return fmt.Errorf("resource validation failed: %w", err)
	}

	// Auto-detect and adjust parallel jobs based on SLURM limits
	if eng.GetName().String() == "slurm" || eng.GetName().String() == "slurm_array" {
		adjustedJobs := adjustParallelJobsForSlurmLimits(parallelJobs)
		if adjustedJobs != parallelJobs {
			logger.Infof("Auto-adjusted parallel jobs from %d to %d based on SLURM limits",
				parallelJobs, adjustedJobs)
			parallelJobs = adjustedJobs
		}
	}

	// Set dry-run mode if requested
	if dryRun {
		manager.SetDryRun(true)
		logger.Info("Dry run mode - validating workflow with Snakemake")
		logger.Infof("Configuration loaded: %s", cfg.Workflow.Mode)
		logger.Infof("Samples found: %d", len(sampleNames))
		logger.Infof("Engine: %s", cfg.Engine.Type)
	}

	// Set conda environment (from flag or config)
	condaEnv := runCondaEnv
	if condaEnv == "" {
		condaEnv = cfg.Engine.CondaEnv
	}
	if condaEnv != "" {
		manager.SetCondaEnv(condaEnv)
		logger.Infof("Using conda environment: %s", condaEnv)
	}

	// Set fallback configuration from config
	fallbackEnv := cfg.Engine.FallbackEnv
	if fallbackEnv == "" {
		fallbackEnv = "xdxtools-snakemake" // Default fallback
	}
	manager.SetFallbackConfig(fallbackEnv, cfg.Engine.NoFallback)

	// Preflight check for rnaseq_splicing dependency chain:
	// gomats -> rmats.py in xdxtools-core environment via enva.
	if err := validateRNAsplicingDependencies(cfg); err != nil {
		return err
	}

	// Prepare FASTQ files if requested
	if copyFastq || moveFastq {
		if err := prepareFastqFiles(cfg, moveFastq); err != nil {
			return fmt.Errorf("failed to prepare FASTQ files: %w", err)
		}
	}

	// Execute workflow
	if err := manager.ExecuteAll(); err != nil {
		logger.Close() // Close log file before returning error
		return fmt.Errorf("workflow execution failed: %w", err)
	}

	if dryRun {
		logger.Info("Dry run completed - workflow validation successful")
	} else {
		logger.Info("Workflow completed successfully")
	}

	// Close log file
	logger.Close()

	return nil
}

func shouldPreflightRNAsplicing(cfg *config.XDXToolsConfig) bool {
	if strings.ToUpper(strings.TrimSpace(cfg.Workflow.Mode)) != "RNASEQ" {
		return false
	}
	return cfg.Metadata.GroupLevels >= 2
}

func validateRNAsplicingDependencies(cfg *config.XDXToolsConfig) error {
	if !shouldPreflightRNAsplicing(cfg) {
		return nil
	}

	if _, err := exec.LookPath("enva"); err != nil {
		return fmt.Errorf(
			"RNA splicing preflight failed: enva not found in PATH: %w\n"+
				"Required for rnaseq_splicing: enva + xdxtools-core environment with gomats and rmats.py.\n"+
				"Try: enva run xdxtools-core -- gomats --help\n"+
				"Try: enva run xdxtools-core -- rmats.py --help", err)
	}

	checks := [][]string{
		{"enva", "run", "xdxtools-core", "--", "gomats", "--help"},
		{"enva", "run", "xdxtools-core", "--", "rmats.py", "--help"},
	}

	for _, check := range checks {
		cmd := exec.Command(check[0], check[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf(
				"RNA splicing preflight failed while running `%s`: %w\noutput:\n%s\n"+
					"Required for rnaseq_splicing: enva + xdxtools-core with gomats and rmats.py.\n"+
					"Try: enva run xdxtools-core -- gomats --help\n"+
					"Try: enva run xdxtools-core -- rmats.py --help",
				strings.Join(check, " "), err, string(output))
		}
	}

	return nil
}

// prepareFastqFiles prepares FASTQ files by copying or moving them to the project data directory
func prepareFastqFiles(cfg *config.XDXToolsConfig, move bool) error {
	fastqDir := cfg.Input.FastqDir
	dataDir := cfg.Output.RawDir

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Find all FASTQ files (support various extensions)
	patterns := []string{
		filepath.Join(fastqDir, "*.fastq.gz"),
		filepath.Join(fastqDir, "*.fq.gz"),
		filepath.Join(fastqDir, "*.fastq"),
		filepath.Join(fastqDir, "*.fq"),
	}

	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			logger.Debugf("Error globbing %s: %v", pattern, err)
			continue
		}
		files = append(files, matches...)
	}

	if len(files) == 0 {
		return fmt.Errorf("no FASTQ files found in %s", fastqDir)
	}

	operation := "copying"
	if move {
		operation = "moving"
	}
	logger.Infof("%s %d FASTQ files to %s", operation, len(files), dataDir)

	// Process each file
	for _, src := range files {
		dst := filepath.Join(dataDir, filepath.Base(src))

		// Check if destination already exists
		if _, err := os.Stat(dst); err == nil {
			logger.Debugf("File already exists, skipping: %s", filepath.Base(dst))
			continue
		}

		if move {
			// Move file
			if err := os.Rename(src, dst); err != nil {
				logger.Warnf("Failed to move %s: %v", filepath.Base(src), err)
				// Try copy as fallback
				if err := copyFile(src, dst); err != nil {
					return fmt.Errorf("failed to copy %s: %w", filepath.Base(src), err)
				}
			}
		} else {
			// Copy file
			if err := copyFile(src, dst); err != nil {
				return fmt.Errorf("failed to copy %s: %w", filepath.Base(src), err)
			}
		}
		logger.Debugf("Processed: %s", filepath.Base(src))
	}

	logger.Info("FASTQ file preparation completed")
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy contents
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Copy file mode
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, info.Mode())
}

// compressFastqFiles compresses uncompressed FASTQ files to .gz format
func compressFastqFiles(fastqDir string) error {
	patterns := []string{
		filepath.Join(fastqDir, "*.fastq"),
		filepath.Join(fastqDir, "*.fq"),
	}

	var uncompressedFiles []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			logger.Debugf("Error globbing %s: %v", pattern, err)
			continue
		}
		uncompressedFiles = append(uncompressedFiles, matches...)
	}

	if len(uncompressedFiles) == 0 {
		logger.Info("No uncompressed FASTQ files found")
		return nil
	}

	logger.Infof("Found %d uncompressed FASTQ files to compress", len(uncompressedFiles))

	// Check if gzip is available
	if _, err := exec.LookPath("gzip"); err != nil {
		return fmt.Errorf("gzip not found in PATH, cannot compress files: %w", err)
	}

	for _, file := range uncompressedFiles {
		compressedPath := file + ".gz"

		// Skip if already compressed
		if _, err := os.Stat(compressedPath); err == nil {
			logger.Debugf("Compressed file already exists, skipping: %s", filepath.Base(compressedPath))
			continue
		}

		logger.Infof("Compressing: %s", filepath.Base(file))

		// Run gzip command
		cmd := exec.Command("gzip", "-f", file)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to compress %s: %w\noutput: %s", filepath.Base(file), err, string(output))
		}

		logger.Debugf("Compressed: %s", filepath.Base(file))
	}

	logger.Info("FASTQ compression completed")
	return nil
}

// validateResources 验证本地和SLURM资源
func validateResources(engineType string, cfg *config.XDXToolsConfig, stepResources map[int]*config.StepResource, parallelJobs int) error {
	logger.Info("Validating resources...")

	// Get workflow mode and PDX status
	mode := cfg.Workflow.Mode
	pdxMode := config.DetectPDXMode(cfg)

	// Validate each step
	for step := 1; step <= 3; step++ {
		// Get default step resource
		defaultRes := config.GetDefaultStepResource(step, mode, pdxMode)

		// Get final step resource (command line overrides defaults)
		stepRes := &config.StepResource{
			Cores:     defaultRes.Cores,
			Memory:    defaultRes.Memory,
			Partition: defaultRes.Partition,
		}

		if customRes, exists := stepResources[step]; exists {
			if customRes.Cores > 0 {
				stepRes.Cores = customRes.Cores
			}
			if customRes.Memory != "" {
				stepRes.Memory = customRes.Memory
			}
			if customRes.Partition != "" {
				stepRes.Partition = customRes.Partition
			}
		}

		// Validate based on engine type
		switch engineType {
		case "local":
			// Validate local resources
			if err := engine.ValidateLocalResources(stepRes.Cores, stepRes.Memory); err != nil {
				return fmt.Errorf("step %d local resource validation failed: %w", step, err)
			}

		case "slurm", "slurm_array":
			// Validate SLURM resources
			partition := stepRes.Partition
			if partition == "" {
				partition = "cpu" // Default partition
			}

			// Validate partition exists
			if err := engine.ValidateSlurmPartition(partition); err != nil {
				return fmt.Errorf("step %d SLURM partition validation failed: %w", step, err)
			}

			// Validate node resources (检查是否有节点满足需求)
			if err := engine.ValidateSlurmNodeResources(partition, stepRes.Cores, stepRes.Memory); err != nil {
				return fmt.Errorf("step %d SLURM node resource validation failed: %w", step, err)
			}

		default:
			logger.Debugf("Engine type %s, skipping resource validation", engineType)
		}
	}

	// Validate parallel jobs
	if parallelJobs > 0 {
		if engineType == "local" {
			if err := engine.ValidateParallelJobs(parallelJobs); err != nil {
				return fmt.Errorf("parallel jobs validation failed: %w", err)
			}
		} else if parallelJobs > 20 {
			logger.Warnf("High parallel jobs (%d) may cause SLURM queue congestion", parallelJobs)
		}
	}

	logger.Info("Resource validation completed successfully")
	return nil
}

// buildStepResources builds the step resources from command line arguments
func buildStepResources() map[int]*config.StepResource {
	resources := make(map[int]*config.StepResource)

	// Determine partition for each step
	// Priority: specific partition > unified partition (slurmUnifiedPartition or slurmPartition) > default
	step1Part := step1Partition
	step2Part := step2Partition
	step3Part := step3Partition

	// Determine unified partition from either parameter
	// slurmPartition acts as unified partition if slurmUnifiedPartition is not set
	unifiedPartition := slurmUnifiedPartition
	if unifiedPartition == "" && slurmPartition != "" {
		unifiedPartition = slurmPartition
	}

	// Apply unified partition if no specific partition is set
	if unifiedPartition != "" {
		if step1Part == "" {
			step1Part = unifiedPartition
		}
		if step2Part == "" {
			step2Part = unifiedPartition
		}
		if step3Part == "" {
			step3Part = unifiedPartition
		}
	}

	// Step 1
	if step1Cores > 0 || step1Memory != "" || step1Part != "" {
		resources[1] = &config.StepResource{
			Cores:     step1Cores,
			Memory:    step1Memory,
			Partition: step1Part,
		}
	}

	// Step 2
	if step2Cores > 0 || step2Memory != "" || step2Part != "" {
		resources[2] = &config.StepResource{
			Cores:     step2Cores,
			Memory:    step2Memory,
			Partition: step2Part,
		}
	}

	// Step 3
	if step3Cores > 0 || step3Memory != "" || step3Part != "" {
		resources[3] = &config.StepResource{
			Cores:     step3Cores,
			Memory:    step3Memory,
			Partition: step3Part,
		}
	}

	// Step 2 Checker
	if step2CheckerCores > 0 || step2CheckerMemory != "" {
		resources[102] = &config.StepResource{
			Cores:  step2CheckerCores,
			Memory: step2CheckerMemory,
		}
	}

	// Step 3 Checker
	if step3CheckerCores > 0 || step3CheckerMemory != "" {
		resources[103] = &config.StepResource{
			Cores:  step3CheckerCores,
			Memory: step3CheckerMemory,
		}
	}

	return resources
}

// adjustParallelJobsForSlurmLimits 输出 SLURM 限制信息，实际分批逻辑由 engine 层处理
// 返回原始 requestedJobs，不做调整
func adjustParallelJobsForSlurmLimits(requestedJobs int) int {
	limits := engine.GetSlurmUserLimits()
	if limits.MaxSubmitJobs > 0 {
		logger.Infof("SLURM limits: MaxSubmit=%d, Current=%d, Available=%d",
			limits.MaxSubmitJobs, limits.CurrentJobCount, limits.AvailableJobs)
	}
	// 分批逻辑由 engine 层（SlurmArrayEngine.executeStepWithBatches）处理
	return requestedJobs
}
