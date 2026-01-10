package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xdxtools/xdxtools-go/internal/logger"
)

func init() {
	logger.Init(false)
}

func TestGenerateSnakemakeConfig_RRBS(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	config := &WorkflowConfig{
		Mode:     "RRBS",
		UserID:   "test_user",
		JobID:    "test_job",
		Species1: "human",
		Suffix1:  "_R1.fastq.gz",
		Suffix2:  "_R2.fastq.gz",
		Adapters: map[string]string{
			"adapter1": "AGATCGGAAGAGC",
			"adapter2": "AGATCGGAAGAGC",
		},
		Alignment: AlignmentConfig{
			ErrorRate: 0.2,
			C1:        "7",
			C2:        "9",
			T1:        0,
			T2:        0,
		},
		Output: OutputConfig{
			BaseDir:     "./results",
			WorkflowDir: "/workflow",
			AnalysisDir: "/analysis",
			QCDir:       "/QC",
			TrimDir:     "/trim",
			OutDirMCall: "/mCall",
			OutDirUmx:   "/uxm",
			LogDir:      "/log",
		},
		Reference: ReferenceConfig{
			GenomeIndex: []string{"inst/pdx/homo_sapiens/"},
			GenomeFasta: []string{"inst/pdx/homo_sapiens/hg19.fasta"},
		},
		UserEmail: "test@example.com",
	}

	samples := []string{"sample1", "sample2", "sample3"}

	outputPath := filepath.Join(tmpDir, "snakemake.yaml")

	err := GenerateSnakemakeConfig(config, samples, outputPath)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check if file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Snakemake config file was not created")
	}

	// Verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	contentStr := string(content)
	if contentStr == "" {
		t.Error("Config file is empty")
	}

	// Check for expected fields
	if !containsString(contentStr, "Mode: RRBS") {
		t.Error("Config file does not contain 'Mode: RRBS'")
	}

	if !containsString(contentStr, "userid: test_user") {
		t.Error("Config file does not contain 'userid: test_user'")
	}

	if !containsString(contentStr, "jobid: test_job") {
		t.Error("Config file does not contain 'jobid: test_job'")
	}

	if !containsString(contentStr, "SIDs:") {
		t.Error("Config file does not contain 'SIDs:'")
	}
}

func TestGenerateSnakemakeConfig_PDX(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	config := &WorkflowConfig{
		Mode:     "RRBS",
		UserID:   "test_user",
		JobID:    "test_job",
		Species1: "human",
		Species2: "mouse",
		Suffix1:  "_R1.fastq.gz",
		Suffix2:  "_R2.fastq.gz",
		Adapters: map[string]string{
			"adapter1": "AGATCGGAAGAGC",
			"adapter2": "AGATCGGAAGAGC",
		},
		Alignment: AlignmentConfig{
			ErrorRate: 0.2,
		},
		Output: OutputConfig{
			BaseDir:     "./results",
			WorkflowDir: "/workflow",
			AnalysisDir: "/analysis",
			QCDir:       "/QC",
			TrimDir:     "/trim",
			OutDirMCall: "/mCall",
			OutDirUmx:   "/uxm",
			LogDir:      "/log",
		},
		Reference: ReferenceConfig{
			GenomeIndex: []string{"inst/pdx/homo_sapiens/"},
		},
	}

	samples := []string{"sample1", "sample2"}

	outputPath := filepath.Join(tmpDir, "snakemake_pdx.yaml")

	err := GenerateSnakemakeConfig(config, samples, outputPath)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	contentStr := string(content)

	// Check for PDX-specific fields
	if !containsString(contentStr, "Mode: RRBS") {
		t.Error("PDX config should contain 'Mode: RRBS'")
	}

	if !containsString(contentStr, "graft:") {
		t.Error("PDX config should contain 'graft:'")
	}

	if !containsString(contentStr, "host:") {
		t.Error("PDX config should contain 'host:'")
	}
}

func TestInferGraftHost_SingleSpecies(t *testing.T) {
	config := &WorkflowConfig{
		Species1: "human",
		Species2: "",
	}

	graft, host := inferGraftHost(config)

	if graft != "human" {
		t.Errorf("Expected graft to be 'human', got '%s'", graft)
	}

	if host != "" {
		t.Errorf("Expected host to be empty, got '%s'", host)
	}
}

func TestInferGraftHost_PDX_HumanMouse(t *testing.T) {
	config := &WorkflowConfig{
		Species1: "human",
		Species2: "mouse",
	}

	graft, host := inferGraftHost(config)

	if graft != "human" {
		t.Errorf("Expected graft to be 'human', got '%s'", graft)
	}

	if host != "mouse" {
		t.Errorf("Expected host to be 'mouse', got '%s'", host)
	}
}

func TestInferGraftHost_PDX_MouseHuman(t *testing.T) {
	config := &WorkflowConfig{
		Species1: "mouse",
		Species2: "human",
	}

	graft, host := inferGraftHost(config)

	if graft != "mouse" {
		t.Errorf("Expected graft to be 'mouse', got '%s'", graft)
	}

	if host != "human" {
		t.Errorf("Expected host to be 'human', got '%s'", host)
	}
}

func TestInferGraftHost_PDX_CustomSpecies(t *testing.T) {
	config := &WorkflowConfig{
		Species1: "rat",
		Species2: "mouse",
	}

	graft, host := inferGraftHost(config)

	if graft != "rat" {
		t.Errorf("Expected graft to be 'rat', got '%s'", graft)
	}

	if host != "mouse" {
		t.Errorf("Expected host to be 'mouse', got '%s'", host)
	}
}

func TestInferGraftHost_ScientificNames(t *testing.T) {
	config := &WorkflowConfig{
		Species1: "homo_sapiens",
		Species2: "mus_musculus",
	}

	graft, host := inferGraftHost(config)

	if graft != "human" {
		t.Errorf("Expected graft to be 'human', got '%s'", graft)
	}

	if host != "mouse" {
		t.Errorf("Expected host to be 'mouse', got '%s'", host)
	}
}

func TestGetSpeciesSlice_SingleSpecies(t *testing.T) {
	config := &WorkflowConfig{
		Species1: "human",
		Species2: "",
	}

	species := getSpeciesSlice(config)

	if len(species) != 1 {
		t.Errorf("Expected 1 species, got %d", len(species))
	}

	if species[0] != "human" {
		t.Errorf("Expected species[0] to be 'human', got '%s'", species[0])
	}
}

func TestGetSpeciesSlice_TwoSpecies(t *testing.T) {
	config := &WorkflowConfig{
		Species1: "human",
		Species2: "mouse",
	}

	species := getSpeciesSlice(config)

	if len(species) != 2 {
		t.Errorf("Expected 2 species, got %d", len(species))
	}

	if species[0] != "human" {
		t.Errorf("Expected species[0] to be 'human', got '%s'", species[0])
	}

	if species[1] != "mouse" {
		t.Errorf("Expected species[1] to be 'mouse', got '%s'", species[1])
	}
}

func TestGenerateSnakemakeConfig_EmptySamples(t *testing.T) {
	tmpDir := t.TempDir()

	config := &WorkflowConfig{
		Mode:      "RRBS",
		UserID:    "test_user",
		JobID:     "test_job",
		Species1:  "human",
		Output:    OutputConfig{},
		Reference: ReferenceConfig{},
	}

	samples := []string{}

	outputPath := filepath.Join(tmpDir, "snakemake.yaml")

	err := GenerateSnakemakeConfig(config, samples, outputPath)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	contentStr := string(content)

	// Should still generate config even with empty samples
	if !containsString(contentStr, "Mode: RRBS") {
		t.Error("Config file should contain 'Mode: RRBS' even with empty samples")
	}
}

func TestGenerateSnakemakeConfig_DirectoryCreation(t *testing.T) {
	// Test that nested directories are created
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "deep", "config.yaml")

	config := &WorkflowConfig{
		Mode:      "RRBS",
		Output:    OutputConfig{},
		Reference: ReferenceConfig{},
	}

	err := GenerateSnakemakeConfig(config, []string{}, nestedPath)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check if file was created in nested directory
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("Config file was not created in nested directory")
	}
}

// Helper function to check if a string contains a substring
func containsString(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(str) > len(substr) &&
		(str[:len(substr)] == substr || str[len(str)-len(substr):] == substr ||
			indexOf(str, substr) >= 0))
}

// Simple indexOf implementation for testing
func indexOf(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
