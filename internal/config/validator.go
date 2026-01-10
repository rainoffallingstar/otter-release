package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xdxtools/xdxtools-go/internal/logger"
)

// ValidateConfig validates the workflow configuration
func ValidateConfig(config *WorkflowConfig) error {
	// Validate mode
	validModes := []string{"RRBS", "WGBS", "BSSEQ", "RNASEQ"}
	if !contains(validModes, strings.ToUpper(config.Mode)) {
		return fmt.Errorf("invalid mode: %s. Valid modes: %v", config.Mode, validModes)
	}

	// Detect PDX mode
	pdxMode := config.Species1 != "" && config.Species2 != ""
	if pdxMode {
		logger.Info("PDX mode detected (species1 and species2 specified)")
	}

	// Validate species
	if config.Species1 == "" {
		return fmt.Errorf("species1 must be specified")
	}

	// Validate FASTQ directories
	if config.Input.FastqDir != "" {
		if absPath, err := filepath.Abs(config.Input.FastqDir); err == nil {
			logger.Debugf("FASTQ directory: %s", absPath)
		}
	}

	// Validate suffix patterns
	if config.Suffix1 == "" {
		return fmt.Errorf("suffix1 must be specified")
	}
	if config.Suffix2 == "" {
		logger.Warn("suffix2 is empty, will be auto-derived from suffix1")
	}

	// Validate engine configuration
	if config.Engine.Type == "" {
		config.Engine.Type = "auto"
		logger.Info("Engine type not specified, using auto-detection")
	}

	validEngineTypes := []string{"auto", "slurm", "local"}
	if !contains(validEngineTypes, config.Engine.Type) {
		return fmt.Errorf("invalid engine type: %s. Valid types: %v", config.Engine.Type, validEngineTypes)
	}

	// Validate parallel settings
	if config.Parallel.Workers <= 0 {
		logger.Warn("Invalid workers setting, using default: 4")
		config.Parallel.Workers = 4
	}

	// Validate reference configuration
	if len(config.Reference.GenomeFasta) > 0 {
		for _, fasta := range config.Reference.GenomeFasta {
			if !fileExists(fasta) {
				logger.Warnf("Reference FASTA not found: %s", fasta)
			}
		}
	}

	// Validate alignment parameters
	if config.Alignment.ErrorRate < 0 || config.Alignment.ErrorRate > 1 {
		return fmt.Errorf("error_rate must be between 0 and 1, got: %f", config.Alignment.ErrorRate)
	}

	logger.Debug("Configuration validation passed")
	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

// DetectPDXMode detects if PDX mode should be enabled
func DetectPDXMode(config *WorkflowConfig) bool {
	return config.Species1 != "" && config.Species2 != ""
}

// GetWorkflowName returns the workflow name based on mode and PDX flag
func GetWorkflowName(mode string, pdx bool) string {
	switch {
	case pdx && mode == "RNASEQ":
		return "BeaverRNASEQPDX"
	case pdx:
		return "BeaverPDX"
	case mode == "RNASEQ":
		return "BeaverRNA"
	default: // RRBS/WGBS/BSSEQ
		return "BeaverBS"
	}
}

// GetStepCount returns the number of steps for the workflow
func GetStepCount(mode string, pdx bool) int {
	if mode == "RNASEQ" && !pdx {
		return 2 // RNASEQ non-PDX only has 2 steps
	}
	return 3 // All other modes have 3 steps
}
