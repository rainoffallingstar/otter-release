package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xdxtools/xdxtools-go/internal/assets"
	"github.com/xdxtools/xdxtools-go/internal/logger"
)

var (
	initMode    string
	projectPath string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new xdxtools project",
	Long: `Initialize a new xdxtools project with the beaverflow directory structure.

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
  xdxtools init my_project
  xdxtools init my_project --mode RRBS`,
	Args: cobra.MinimumNArgs(0),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&initMode, "mode", "m", "RRBS", "Workflow mode (RRBS/WGBS/RNASEQ)")
	initCmd.Flags().StringVarP(&projectPath, "path", "p", ".", "Parent directory for the project")
}

func runInit(cmd *cobra.Command, args []string) error {
	var projectName string
	if len(args) > 0 {
		projectName = args[0]
	} else {
		projectName = "xdxtools-project"
	}

	projectDir := filepath.Join(projectPath, projectName)

	// Check if directory already exists
	if _, err := os.Stat(projectDir); err == nil {
		logger.Warnf("Directory %s already exists", projectDir)
	}

	// Create asset copier
	copier := assets.NewAssetCopier(projectDir, "")

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

	// Step 4: Create README.md
	readmeContent := fmt.Sprintf(`# %s

This is an xdxtools project initialized with beaverflow structure.

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
4. Create config file: xdxtools create --fastq data --mode %s --output config/config.yaml
5. Run workflow: xdxtools run --config config/config.yaml

## Snakemake Workflows

Available workflows:
- BeaverBS_step1/2/3.snakemake   - RRBS/WGBS bisulfite sequencing
- BeaverPDX_step1/2/3.snakemake  - PDX bisulfite sequencing
- BeaverRNA_step1/2.snakemake    - RNA-seq analysis
- BeaverRNASEQPDX_step1/2/3.snakemake - PDX RNA-seq analysis

`, projectName, initMode, initMode)

	readmePath := filepath.Join(projectDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		logger.Warnf("Failed to create README: %v", err)
	}

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
	logger.Infof("4. Create config: xdxtools create --fastq data --mode %s --output %s/config/config.yaml", initMode, projectName)
	logger.Infof("5. Run workflow: xdxtools run --config %s/config/config.yaml", projectName)

	return nil
}
