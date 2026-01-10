package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"github.com/xdxtools/xdxtools-go/internal/logger"
)

// SnakemakeConfig represents the configuration for Snakemake
type SnakemakeConfig struct {
	Mode      string            `yaml:"Mode"`
	UserID    string            `yaml:"userid"`
	JobID     string            `yaml:"jobid"`
	Species   []string         `yaml:"species"`
	Graft     string            `yaml:"graft"`
	Host      string            `yaml:"host"`
	Suffix    string            `yaml:"suffix"`
	Suffix2   string            `yaml:"suffix2"`
	Error     float64           `yaml:"error"`
	TrimSeq1  string            `yaml:"trimSeq1"`
	TrimSeq2  string            `yaml:"trimSeq2"`
	C1        string            `yaml:"C1"`
	C2        string            `yaml:"C2"`
	T1        int               `yaml:"T1"`
	T2        int               `yaml:"T2"`

	// Directory paths
	WorkDir      string `yaml:"workDir"`
	WorkflowDir  string `yaml:"workflowDir"`
	AnalysisDir  string `yaml:"analysisDir"`
	SelfConfig   string `yaml:"selfconfig"`
	QCDir       string `yaml:"qcDir"`
	TrimDir     string `yaml:"trimDir"`
	BsmapDir    string `yaml:"bsmapDir"`
	OutDirMCall string `yaml:"outDir_mCall"`
	OutDirUmx   string `yaml:"outDir_umx"`
	LogDir      string `yaml:"SID_log"`

	// Reference files
	GenomeFile   []string `yaml:"genomeFile"`
	GenomeFasta  []string `yaml:"gnome_fasta"`
	GenomeAnno   []string `yaml:"genomeAnno"`
	RNASEQGTF   string   `yaml:"rnaseq_gtf"`
	RNASEQRef   string   `yaml:"rnaseq_ref"`

	// Sample information
	SIDs []string `yaml:"SIDs"`

	// User info
	UserEmail string `yaml:"user_email"`
}

// GenerateSnakemakeConfig generates YAML configuration for Snakemake
func GenerateSnakemakeConfig(config *WorkflowConfig, samples []string, outputPath string) error {
	// Infer graft and host
	graft, host := inferGraftHost(config)

	snakeConfig := SnakemakeConfig{
		Mode:        strings.ToUpper(config.Mode),
		UserID:      config.UserID,
		JobID:       config.JobID,
		Species:     getSpeciesSlice(config),
		Graft:       graft,
		Host:        host,
		Suffix:      config.Suffix1,
		Suffix2:     config.Suffix2,
		Error:       config.Alignment.ErrorRate,
		TrimSeq1:    config.Adapters["adapter1"],
		TrimSeq2:    config.Adapters["adapter2"],
		C1:          config.Alignment.C1,
		C2:          config.Alignment.C2,
		T1:          config.Alignment.T1,
		T2:          config.Alignment.T2,

		WorkDir:      config.Output.BaseDir,
		WorkflowDir:  config.Output.WorkflowDir,
		AnalysisDir:  config.Output.AnalysisDir,
		SelfConfig:   config.SelfConfig,
		QCDir:       config.Output.QCDir,
		TrimDir:     config.Output.TrimDir,
		BsmapDir:    filepath.Join(config.Output.WorkflowDir, "bsmap"),
		OutDirMCall: config.Output.OutDirMCall,
		OutDirUmx:   config.Output.OutDirUmx,
		LogDir:      config.Output.LogDir,

		GenomeFile:   config.Reference.GenomeIndex,
		GenomeFasta:  config.Reference.GenomeFasta,
		GenomeAnno:   config.Reference.GenomeAnno,
		RNASEQGTF:    config.Reference.RNASEQGTF,
		RNASEQRef:    config.Reference.RNASEQRef,

		SIDs:      samples,
		UserEmail: config.UserEmail,
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(snakeConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	logger.Infof("Snakemake configuration generated: %s", outputPath)
	return nil
}

// inferGraftHost infers graft and host from species configuration
func inferGraftHost(config *WorkflowConfig) (graft, host string) {
	if config.Species2 == "" {
		// Single species mode
		return config.Species1, ""
	}

	// PDX mode: determine graft and host
	if config.Species1 == "human" || config.Species1 == "homo_sapiens" {
		graft = "human"
		host = "mouse"
	} else if config.Species1 == "mouse" || config.Species1 == "mus_musculus" {
		graft = "mouse"
		host = "human"
	} else {
		graft = config.Species1
		host = config.Species2
	}

	return
}

// getSpeciesSlice returns species as a slice
func getSpeciesSlice(config *WorkflowConfig) []string {
	if config.Species2 == "" {
		return []string{config.Species1}
	}
	return []string{config.Species1, config.Species2}
}
