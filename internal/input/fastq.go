package input

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xdxtools/xdxtools-go/internal/logger"
)

// Scanner scans directories for FASTQ files
type Scanner struct {
	options *ScanOptions
}

// NewScanner creates a new FASTQ scanner
func NewScanner(options *ScanOptions) *Scanner {
	// Set default extensions if not provided
	if options.Extensions == nil || len(options.Extensions) == 0 {
		options.Extensions = []string{".fastq.gz", ".fastq", ".fq.gz", ".fq"}
	}

	// Set default suffix if not provided
	if options.Suffix1 == "" {
		options.Suffix1 = "_R1.fastq.gz"
	}
	if options.Suffix2 == "" {
		options.Suffix2 = "" // Will be auto-derived
	}

	return &Scanner{
		options: options,
	}
}

// Scan scans the directory for FASTQ files
func (s *Scanner) Scan() ([]Sample, error) {
	logger.Infof("Scanning FASTQ directory: %s", s.options.FastqDir)

	// Check if directory exists
	if _, err := os.Stat(s.options.FastqDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("FASTQ directory not found: %s", s.options.FastqDir)
	}

	// Find all FASTQ files
	files, err := filepath.Glob(filepath.Join(s.options.FastqDir, "*"))
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// Filter FASTQ files
	fastqFiles := []string{}
	for _, file := range files {
		// Skip directories
		if info, err := os.Stat(file); err != nil || info.IsDir() {
			continue
		}

		// Check if it's a FASTQ file
		if s.isFastqFile(file) {
			fastqFiles = append(fastqFiles, file)
		}
	}

	logger.Debugf("Found %d FASTQ files", len(fastqFiles))

	// Group files by sample name
	samples, err := s.groupFiles(fastqFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to group files: %w", err)
	}

	logger.Infof("Identified %d samples", len(samples))

	return samples, nil
}

// isFastqFile checks if a file is a FASTQ file based on extension
func (s *Scanner) isFastqFile(filePath string) bool {
	fileName := filepath.Base(filePath)
	lowerName := strings.ToLower(fileName)
	for _, validExt := range s.options.Extensions {
		if strings.HasSuffix(lowerName, validExt) {
			return true
		}
	}
	return false
}

// groupFiles groups FASTQ files by sample name
func (s *Scanner) groupFiles(files []string) ([]Sample, error) {
	samples := make(map[string][]string)

	// Auto-derive suffix2 if not set
	suffix2 := s.options.Suffix2
	if suffix2 == "" {
		suffix2 = deriveSuffix2(s.options.Suffix1)
	}

	logger.Debugf("Using suffix1=%s, suffix2=%s", s.options.Suffix1, suffix2)

	for _, file := range files {
		fileName := filepath.Base(file)

		// Remove suffix1 or suffix2 to get sample name
		sampleName := fileName

		// Try removing suffix1
		if strings.HasSuffix(fileName, s.options.Suffix1) {
			sampleName = strings.TrimSuffix(fileName, s.options.Suffix1)
		} else if strings.HasSuffix(fileName, suffix2) {
			// Try removing suffix2
			sampleName = strings.TrimSuffix(fileName, suffix2)
		}

		samples[sampleName] = append(samples[sampleName], file)
	}

	// Convert to Sample structs
	var result []Sample
	for sampleName, fileList := range samples {
		sample := Sample{
			Name:      sampleName,
			Metadata:  make(map[string]string),
			CreatedAt: time.Now(),
		}

		// Determine R1 and R2 files
		for _, file := range fileList {
			fileName := filepath.Base(file)
			// Check if this is an R1 file
			if strings.Contains(fileName, "_R1") || strings.Contains(fileName, "_1.") || strings.HasSuffix(fileName, s.options.Suffix1) {
				sample.FastqR1 = file
			} else if strings.Contains(fileName, "_R2") || strings.Contains(fileName, "_2.") || strings.HasSuffix(fileName, suffix2) {
				// Check if this is an R2 file
				sample.FastqR2 = file
			} else {
				// If no clear R1/R2 indicator, assign to R1
				if sample.FastqR1 == "" {
					sample.FastqR1 = file
				}
			}
		}

		result = append(result, sample)
	}

	return result, nil
}

// PairSamples pairs R1 and R2 files for each sample
func (s *Scanner) PairSamples(samples []Sample, options *PairOptions) ([]PairedSample, error) {
	if options == nil {
		options = &PairOptions{
			StrictMatching: true,
			AllowSingleEnd: false,
			MinFiles:       1,
		}
	}

	var pairedSamples []PairedSample

	for _, sample := range samples {
		paired := PairedSample{
			Name:   sample.Name,
			R1Path: sample.FastqR1,
			R2Path: sample.FastqR2,
			Valid:  false,
		}

		// Check if sample is valid
		if paired.R1Path != "" && paired.R2Path != "" {
			// Both R1 and R2 exist
			paired.Valid = true
		} else if paired.R1Path != "" && options.AllowSingleEnd {
			// Only R1 exists and single-end is allowed
			paired.Valid = true
		} else if options.StrictMatching {
			// Strict matching requires both R1 and R2
			logger.Warnf("Sample %s is missing R2 file", sample.Name)
			continue
		}

		pairedSamples = append(pairedSamples, paired)
	}

	logger.Infof("Successfully paired %d/%d samples", len(pairedSamples), len(samples))

	return pairedSamples, nil
}

// deriveSuffix2 auto-derives suffix2 from suffix1
func deriveSuffix2(suffix1 string) string {
	// Try replacing '1' with '2'
	if strings.Contains(suffix1, "1") {
		return strings.Replace(suffix1, "1", "2", 1)
	}

	// Try replacing "R1" with "R2"
	if strings.Contains(suffix1, "R1") {
		return strings.Replace(suffix1, "R1", "R2", 1)
	}

	// Fallback: return original
	return suffix1
}
