package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rainoffallingstar/otter/internal/assets"
	"github.com/rainoffallingstar/otter/internal/logger"
	"github.com/spf13/cobra"
)

var (
	initMode   = "RRBS" // Default mode, can be changed by editing config later
	legacyFlag bool     // Use legacy rules instead of new rules
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new otter project",
	Long: `Initialize a new otter project with the beaverflow directory structure.

This command creates the complete project structure and copies all required
Snakemake workflow files, R/Python scripts, rules, and conda environments
from the embedded assets.

Directory structure created:
  config/          # Configuration files
  data/            # Input data (FASTQ, pdata)
  envs/            # Conda environments (copied from package)
  inst/            # Genome reference files
  R/               # R/Python scripts (copied from package)
  rules/           # Snakemake rules (copied from package)
  saveRDS/         # R intermediate results
  temp/            # Temporary files
  userspace/       # User workspace
  www/             # Web resources
  *.snakemake      # Snakemake workflow files (copied from package)

Examples:
  otter init my_project
  otter init my_project --mode RRBS`,
	Args: cobra.MinimumNArgs(0),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&legacyFlag, "legacy", false, "Use legacy rules instead of new rules")
}

func runInit(cmd *cobra.Command, args []string) error {
	var projectName string
	var projectDir string

	if len(args) > 0 {
		projectName = args[0]
		if filepath.IsAbs(projectName) {
			// If absolute path, use directly
			projectDir = projectName
		} else {
			// If relative path, use from current directory
			projectDir = projectName
		}
	} else {
		// Default project name if not provided
		projectName = "otter-project"
		projectDir = projectName
	}

	// Check if directory already exists
	if _, err := os.Stat(projectDir); err == nil {
		logger.Warnf("Directory %s already exists", projectDir)
	}

	// Determine rules type
	rulesType := "rootless"
	if legacyFlag {
		rulesType = "legacy"
	}

	// Create asset copier
	copier := assets.NewAssetCopier(projectDir, rulesType)

	// Step 1: Create directory structure
	logger.Info("Creating project directory structure...")
	if err := copier.CreateDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Step 2: Copy all workflow assets
	logger.Info("Copying workflow assets...")
	if err := copier.CopyAll(); err != nil {
		return fmt.Errorf("failed to copy assets: %w", err)
	}

	// Step 3: Create snakefile entry point
	snakefilePath := filepath.Join(projectDir, "snakefile")
	if err := os.WriteFile(snakefilePath, []byte("# Main snakefile entry point\n"), 0644); err != nil {
		logger.Warnf("Failed to create snakefile: %v", err)
	}

	// Step 3.5: Stamp workflow assets manifest (includes deprecated R/ assets).
	logger.Info("Stamping workflow assets manifest...")
	manifest, err := assets.StampWorkflowAssets(projectDir, buildVersion, buildCommit, buildDate)
	if err != nil {
		return fmt.Errorf("failed to stamp assets manifest: %w", err)
	}
	deprecatedCount := 0
	for _, e := range manifest.Entries {
		if e.Deprecated {
			deprecatedCount++
		}
	}
	if deprecatedCount > 0 {
		logger.Warnf("Deprecated assets detected: %d files under R/ (still verified for integrity)", deprecatedCount)
	}
	logger.Infof("Assets manifest written: %s (%d files)", assets.ManifestPath(projectDir), len(manifest.Entries))

	// Step 4: Create README.md
	readmeContent := fmt.Sprintf(`# %s

This is an otter project initialized with beaverflow structure.

## Workflow Mode: %s

## Directory Structure

- config/          - Configuration files
- data/            - Input data (FASTQ files, pdata)
- envs/            - Conda environments (copied from package)
- inst/            - Genome reference files
- R/               - R/Python scripts
- rules/           - Snakemake rules
- saveRDS/         - R intermediate results
- temp/            - Temporary files
- userspace/       - User workspace
- www/             - Web resources
- *.snakemake      - Snakemake workflow files

## Usage

1. Copy your FASTQ files to data/
2. Place your pdata file (Excel/CSV) in data/
3. Download reference genomes:
   - Visit: https://huggingface.co/datasets/Genomiclab/xdxtools-genomes/tree/main
   - Download and extract genomes to inst/ directory
4. Create config file: otter create --fastq data --mode %s --output %s/userspace --jobid demo_run
5. Run workflow: otter run --config %s/userspace/demo_run/config/otter.yaml

## Snakemake Workflows

Available workflows:
- BeaverBS_step1/2/3.snakemake   - RRBS/WGBS bisulfite sequencing
- BeaverPDX_step1/2/3.snakemake  - PDX bisulfite sequencing
- BeaverRNA_step1/2.snakemake    - RNA-seq analysis
- BeaverRNASEQPDX_step1/2/3.snakemake - PDX RNA-seq analysis

`, projectName, initMode, initMode, projectName, projectName)

	readmePath := filepath.Join(projectDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		logger.Warnf("Failed to create README: %v", err)
	}

	// Check enva support
	checkEnvSupport()

	// Summary
	logger.Info("===========================================")
	logger.Infof("Project '%s' initialized successfully!", projectName)
	logger.Infof("Location: %s", projectDir)
	logger.Info("===========================================")
	logger.Info("Next steps:")
	logger.Info("1. Copy FASTQ files to data/")
	logger.Info("2. Place pdata file (Excel/CSV) in data/")
	logger.Info("3. Download reference genomes:")
	logger.Info("   Visit: https://huggingface.co/datasets/Genomiclab/xdxtools-genomes/tree/main")
	logger.Info("   Download and extract genomes to inst/ directory")
	logger.Infof("4. Create config: otter create --fastq data --output %s/userspace --jobid demo_run", projectName)
	logger.Infof("5. Run workflow: otter run --config %s/userspace/demo_run/config/otter.yaml", projectName)

	return nil
}

// checkEnvSupport checks if enva is available and provides installation guidance
func checkEnvSupport() {
	logger.Info("")
	logger.Info("Checking package manager support...")

	if _, err := exec.LookPath("enva"); err != nil {
		// enva not found
		logger.Warn("────────────────────────────────────────────────────────")
		logger.Warn("enva not found in PATH")
		logger.Warn("")
		logger.Warn("For the best environment workflow, install enva:")
		logger.Warn("  enva is rattler-first and can interoperate with existing conda/mamba/micromamba environments")
		logger.Warn("")
		logger.Warn("Installation:")
		logger.Warn("  wget https://github.com/rainoffallingstar/enva/releases/latest/download/enva-linux-x86_64")
		logger.Warn("  chmod +x enva-linux-x86_64")
		logger.Warn("  sudo mv enva-linux-x86_64 /usr/local/bin/enva")
		logger.Warn("")
		logger.Warn("Or build from source:")
		logger.Warn("  git clone https://github.com/rainoffallingstar/enva")
		logger.Warn("  cd enva && cargo build --release")
		logger.Warn("  cp target/release/enva /usr/local/bin/enva")
		logger.Warn("")
		logger.Warn("Falling back to conda run (slower)")
		logger.Warn("────────────────────────────────────────────────────────")
	} else {
		// enva found
		logger.Info("✓ enva detected - rattler-first environment management is available")
		logger.Info("")
	}
}
