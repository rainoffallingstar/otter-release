package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xdxtools/xdxtools-go/internal/config"
	"github.com/xdxtools/xdxtools-go/internal/input"
	"github.com/xdxtools/xdxtools-go/internal/logger"
	"gopkg.in/yaml.v3"
)

// generateReferencePaths 根据模式和物种生成参考文件路径
func generateReferencePaths(mode, species1, species2 string) (fasta, index, gtf, ref []string, err error) {
	mode = strings.ToUpper(mode)

	// 物种名称映射
	speciesMap := map[string]string{
		"human":        "homo_sapiens",
		"mouse":        "mouse",
		"homo_sapiens": "homo_sapiens",
		"mus_musculus": "mouse",
	}

	// 添加 species1（graft）
	if dir, ok := speciesMap[species1]; ok {
		fasta = append(fasta, fmt.Sprintf("inst/pdx/%s/%s.fasta", dir, species1))
		index = append(index, fmt.Sprintf("inst/pdx/%s/", dir))
	} else {
		return nil, nil, nil, nil, fmt.Errorf("unsupported species1: %s", species1)
	}

	// 如果是PDX模式，添加 species2（host）
	if species2 != "" {
		if dir, ok := speciesMap[species2]; ok {
			fasta = append(fasta, fmt.Sprintf("inst/pdx/%s/%s.fasta", dir, species2))
			index = append(index, fmt.Sprintf("inst/pdx/%s/", dir))
		} else {
			return nil, nil, nil, nil, fmt.Errorf("unsupported species2: %s", species2)
		}
	}

	// RNAseq需要GTF和STAR索引
	if mode == "RNASEQ" {
		// graft (species1) 的GTF和索引
		if dir, ok := speciesMap[species1]; ok {
			gtf = append(gtf, fmt.Sprintf("inst/rnaseq/%s/%s.ensGene_sorted.gtf", dir, species1))
			ref = append(ref, fmt.Sprintf("inst/rnaseq/%s/", dir))
		}

		// 如果是PDX，添加 host (species2) 的GTF和索引
		if species2 != "" {
			if dir, ok := speciesMap[species2]; ok {
				gtf = append(gtf, fmt.Sprintf("inst/rnaseq/%s/%s.ensGene_sorted.gtf", dir, species2))
				ref = append(ref, fmt.Sprintf("inst/rnaseq/%s/", dir))
			}
		}
	}

	return
}

var (
	createFastqDir  string
	createPdataFile string
	createMode      string
	createSpecies1  string
	createSpecies2  string
	createOutputDir string
	createJobID     string
	createSuffix1   string
	createSuffix2   string
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new analysis project",
	Long: `Create a new analysis project after validating paired samples.

This command performs the following steps:
1. Scans the FASTQ directory for paired samples (R1/R2)
2. Validates sample pairing
3. Loads and validates pdata if provided
4. Generates a unique job ID (or uses user-specified one)
5. Creates project directory structure in userspace/{jobid}/
6. Generates config.yaml for the workflow

Examples:
  xdxtools create --fastq /data/fastq --mode RRBS
  xdxtools create --fastq /data/fastq --pdata samples.xlsx --mode WGBS
  xdxtools create --fastq /data/fastq --mode RNASEQ --species1 human --species2 mouse`,
	RunE: runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&createFastqDir, "fastq", "f", "", "FASTQ files directory (required)")
	createCmd.Flags().StringVarP(&createPdataFile, "pdata", "p", "", "Phenotype data file (Excel/CSV)")
	createCmd.Flags().StringVarP(&createMode, "mode", "m", "RRBS", "Workflow mode (RRBS/WGBS/RNASEQ)")
	createCmd.Flags().StringVar(&createSpecies1, "species1", "human", "Primary species")
	createCmd.Flags().StringVar(&createSpecies2, "species2", "", "Secondary species (enables PDX mode)")
	createCmd.Flags().StringVarP(&createOutputDir, "output", "o", "userspace", "Output directory for projects")
	createCmd.Flags().StringVar(&createJobID, "jobid", "", "Custom job ID (default: auto-generated)")
	createCmd.Flags().StringVar(&createSuffix1, "suffix1", "_R1.fastq.gz", "R1 file suffix")
	createCmd.Flags().StringVar(&createSuffix2, "suffix2", "", "R2 file suffix (auto-derived if empty)")

	createCmd.MarkFlagRequired("fastq")
}

func runCreate(cmd *cobra.Command, args []string) error {
	logger.Info("Creating analysis project...")

	// 1. Validate FASTQ directory exists
	if _, err := os.Stat(createFastqDir); os.IsNotExist(err) {
		return fmt.Errorf("FASTq directory does not exist: %s", createFastqDir)
	}

	// 2. Scan FASTQ files
	logger.Infof("Scanning FASTQ directory: %s", createFastqDir)
	scanner := input.NewScanner(&input.ScanOptions{
		FastqDir: createFastqDir,
		Suffix1:  createSuffix1,
		Suffix2:  createSuffix2,
	})

	samples, err := scanner.Scan()
	if err != nil {
		return fmt.Errorf("failed to scan FASTQ files: %w", err)
	}

	if len(samples) == 0 {
		return fmt.Errorf("no FASTQ files found in directory")
	}

	logger.Infof("Found %d FASTQ files", len(samples))

	// 3. Pair samples
	pairedSamples, err := scanner.PairSamples(samples, nil)
	if err != nil {
		return fmt.Errorf("failed to pair samples: %w", err)
	}

	// Filter valid pairs
	var validPairs []input.PairedSample
	for _, ps := range pairedSamples {
		if ps.Valid {
			validPairs = append(validPairs, ps)
		}
	}

	if len(validPairs) == 0 {
		return fmt.Errorf("no valid paired samples found. Check file naming conventions")
	}

	logger.Infof("Found %d valid paired samples", len(validPairs))

	// 4. Load pdata if provided
	var pdata *input.PData
	if createPdataFile != "" {
		logger.Infof("Loading pdata file: %s", createPdataFile)
		parser := input.NewPDataParser()
		pdata, err = parser.Load(createPdataFile)
		if err != nil {
			return fmt.Errorf("failed to load pdata: %w", err)
		}
		logger.Infof("Loaded pdata with %d samples", len(pdata.Samples))
	}

	// 5. Validate input
	validator := input.NewValidator()
	validationResult := validator.ValidateInput(createFastqDir, createPdataFile, validPairs, pdata)

	if len(validationResult.Warnings) > 0 {
		for _, w := range validationResult.Warnings {
			logger.Warn(w)
		}
	}

	if !validationResult.Valid {
		for _, e := range validationResult.Errors {
			logger.Error(e)
		}
		return fmt.Errorf("validation failed. Please fix the errors above")
	}

	// 6. Generate job ID
	jobID := createJobID
	if jobID == "" {
		jobID = generateJobID()
	}

	// Ensure jobID doesn't already exist
	projectDir := filepath.Join(createOutputDir, jobID)
	if _, err := os.Stat(projectDir); err == nil {
		if createJobID != "" {
			return fmt.Errorf("project with jobid '%s' already exists", jobID)
		}
		// Regenerate if auto-generated
		jobID = generateJobID()
		projectDir = filepath.Join(createOutputDir, jobID)
	}

	// 7. Detect PDX mode
	pdxMode := createSpecies2 != ""
	modeStr := strings.ToUpper(createMode)
	if pdxMode {
		logger.Infof("PDX mode enabled: %s + %s", createSpecies1, createSpecies2)
	}

	// 8. Create project directory structure
	logger.Infof("Creating project directory: %s", projectDir)
	if err := createProjectStructure(projectDir, modeStr, createSpecies1, createSpecies2); err != nil {
		return fmt.Errorf("failed to create project structure: %w", err)
	}

	// 9. Get sample names
	sampleNames := make([]string, len(validPairs))
	for i, ps := range validPairs {
		sampleNames[i] = ps.Name
	}

	// 10. Generate adapters for each sample
	adapterGen := input.NewAdapterGenerator(input.AdapterGeneratorOptions{
		BaseAdapter1: "AGATCGGAAGAGC",
		BaseAdapter2: "AGATCGGAAGAGC",
		Mode:         modeStr,
	})
	adapter1, adapter2, err := adapterGen.GenerateAdapters(sampleNames, pdata)
	if err != nil {
		return fmt.Errorf("failed to generate adapters: %w", err)
	}
	logger.Infof("Generated adapters for %d samples", len(sampleNames))

	// 11. Generate config.yaml
	configPath := filepath.Join(projectDir, "config", "config.yaml")
	if err := generateProjectConfig(configPath, modeStr, createSpecies1, createSpecies2,
		createFastqDir, createPdataFile, sampleNames, projectDir, adapter1, adapter2, pdata, jobID); err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// 12. Print summary
	logger.Info("===========================================")
	logger.Infof("Analysis project created successfully!")
	logger.Info("===========================================")
	logger.Infof("  Job ID:    %s", jobID)
	logger.Infof("  Mode:      %s%s", modeStr, func() string {
		if pdxMode {
			return " (PDX)"
		}
		return ""
	}())
	logger.Infof("  Samples:   %d paired samples", len(validPairs))
	logger.Infof("  Project:   %s", projectDir)
	logger.Infof("  Config:    %s", configPath)
	logger.Info("")
	logger.Info("Sample list:")
	for i, name := range sampleNames {
		if i < 5 {
			logger.Infof("  - %s", name)
		} else if i == 5 {
			logger.Infof("  ... and %d more", len(sampleNames)-5)
			break
		}
	}
	logger.Info("")
	logger.Info("Next step:")
	logger.Infof("  xdxtools run --config %s", configPath)

	return nil
}

// generateJobID generates a random job ID (similar to R's openssl::rand_bytes)
func generateJobID() string {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("job_%d", os.Getpid())
	}
	return hex.EncodeToString(b)
}

// createProjectStructure creates the complete project directory structure
func createProjectStructure(projectDir, mode, species1, species2 string) error {
	// Base directories
	baseDirs := []string{
		"data",
		"workflow",
		"analysis",
		"config",
		"log",
	}

	// Workflow subdirectories
	workflowDirs := []string{
		"workflow/QC",
		"workflow/fastqc_raw",
		"workflow/fastqc_clean",
		"workflow/trim",
		"workflow/bsmap",
		"workflow/bsmap/tmp",
		"workflow/mCall",
		"workflow/umx",
		"workflow/qualimap",
		"workflow/mhap",
	}

	// Analysis subdirectories
	analysisDirs := []string{
		"analysis/methrixh5",
		"analysis/GCbias",
		"analysis/clubcpg",
		"analysis/clubcpg/coverage_before",
		"analysis/clubcpg/model",
		"analysis/clubcpg/coverage_impute",
		"analysis/RData",
		"analysis/qc_summary",
		"analysis/betaM",
		"analysis/DMR",
		"analysis/uxm_summary",
		"analysis/logsummary",
	}

	// PDX-specific directories
	pdxMode := species2 != ""
	if pdxMode {
		workflowDirs = append(workflowDirs,
			fmt.Sprintf("workflow/bsmap/tmp/%s", species1),
			fmt.Sprintf("workflow/bsmap/tmp/%s", species2),
			fmt.Sprintf("workflow/bsmap/%s", species1),
			fmt.Sprintf("workflow/bsmap/%s", species2),
			"workflow/bsmap/Filtered_bams",
		)
	}

	// RNA-seq specific directories
	if strings.ToUpper(mode) == "RNASEQ" {
		workflowDirs = append(workflowDirs,
			"workflow/star",
			"workflow/htseq",
			"workflow/splicing",
		)
		analysisDirs = append(analysisDirs,
			"analysis/counts",
			"analysis/DEG",
		)
	}

	// Combine all directories
	allDirs := append(baseDirs, workflowDirs...)
	allDirs = append(allDirs, analysisDirs...)

	// Create all directories
	for _, dir := range allDirs {
		fullPath := filepath.Join(projectDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
		}
	}

	logger.Debugf("Created %d directories", len(allDirs))
	return nil
}

// calculateGroupLevels calculates the number of unique groups in pdata
// Returns 0 if no pdata or no group information available
func calculateGroupLevels(pdata *input.PData, samples []string) int {
	if pdata == nil || pdata.Data == nil {
		logger.Debug("No pdata provided, group_levels = 0")
		return 0
	}

	// Use a map to track unique groups
	groups := make(map[string]bool)

	for _, sample := range samples {
		if sampleData, ok := pdata.Data[sample]; ok {
			// Priority: sample_group > condition
			var group string
			var exists bool

			if group, exists = sampleData["sample_group"]; exists && group != "" {
				groups[group] = true
			} else if group, exists = sampleData["condition"]; exists && group != "" {
				groups[group] = true
			}
		}
	}

	groupCount := len(groups)
	logger.Debugf("Found %d unique groups: %v", groupCount, groups)

	return groupCount
}

// generateProjectConfig generates the config.yaml file
func generateProjectConfig(configPath, mode, species1, species2,
	fastqDir, pdataFile string, samples []string, projectDir string,
	adapter1, adapter2 []string, pdata *input.PData, jobID string) error {

	// Build configuration
	pdxMode := species2 != ""
	workflowName := config.GetWorkflowName(mode, pdxMode)
	stepCount := config.GetStepCount(mode, pdxMode)

	// Calculate group levels
	groupLevels := calculateGroupLevels(pdata, samples)

	// Generate reference paths dynamically
	fasta, index, gtf, ref, err := generateReferencePaths(mode, species1, species2)
	if err != nil {
		return fmt.Errorf("failed to generate reference paths: %w", err)
	}

	// Prepare genome_anno array
	genomeAnno := []string{species1}
	if species2 != "" {
		genomeAnno = append(genomeAnno, species2)
	}

	// Prepare rnaseq fields based on mode
	var rnaseqGTF interface{}
	var rnaseqRef interface{}
	if len(gtf) == 0 {
		rnaseqGTF = ""
		rnaseqRef = ""
	} else if len(gtf) == 1 {
		rnaseqGTF = gtf[0]
		rnaseqRef = ref[0]
	} else {
		rnaseqGTF = gtf
		rnaseqRef = ref
	}

	cfg := map[string]interface{}{
		// Basic settings
		"mode":     mode,
		"species1": species1,
		"species2": species2,
		"pdx_mode": pdxMode,
		"workflow": workflowName,
		"steps":    stepCount,

		// User and job info
		"userid": jobID,
		"jobid":  jobID,

		// File suffixes
		"suffix":  createSuffix1,
		"suffix2": createSuffix2,

		// Input
		"input": map[string]interface{}{
			"fastq_dir":  fastqDir,
			"pdata_file": pdataFile,
		},

		// Samples and per-sample adapters
		"samples":      samples,
		"SIDs":         samples, // Alias for samples (R compatibility)
		"trimSeq1":     adapter1,
		"trimSeq2":     adapter2,
		"group_levels": groupLevels,

		// Output directories (relative to project)
		"output": map[string]interface{}{
			"project_dir":  projectDir,
			"workflow_dir": filepath.Join(projectDir, "workflow"),
			"analysis_dir": filepath.Join(projectDir, "analysis"),
			"config_dir":   filepath.Join(projectDir, "config"),
			"log_dir":      filepath.Join(projectDir, "log"),
			"data_dir":     filepath.Join(projectDir, "data"),
		},

		// Processing parameters
		"error_rate": 0.2,
		"trim": map[string]int{
			"read1_5":  0,
			"read1_3":  0,
			"read2_5":  0,
			"read2_3":  0,
			"seq_deth": 10,
		},
		"alignment": map[string]interface{}{
			"C1": "7",
			"C2": "9",
			"T1": 0,
			"T2": 0,
		},

		// Reference (dynamically generated based on mode and species)
		"reference": map[string]interface{}{
			"genome":       strings.ToLower(species1),
			"genome_fasta": fasta,
			"genome_index": index,
			"genome_anno":  genomeAnno,
			"rnaseq_gtf":   rnaseqGTF,
			"rnaseq_ref":   rnaseqRef,
		},

		// Engine settings
		"engine": map[string]interface{}{
			"type": "auto",
		},

		// Parallel processing
		"parallel": map[string]int{
			"workers": 4,
		},
	}

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	// Write YAML file
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add header comment
	header := `# xdxtools Analysis Project Configuration
# Generated automatically by: xdxtools create
#
# Edit this file to customize your analysis parameters.
# Then run: xdxtools run --config <this-file>
#

`
	content := header + string(data)

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
