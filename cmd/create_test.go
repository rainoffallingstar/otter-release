package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xdxtools/xdxtools-go/internal/input"
	"github.com/xdxtools/xdxtools-go/internal/logger"
)

func init() {
	logger.Init(false)
}

func TestGenerateJobID(t *testing.T) {
	jobID := generateJobID()

	// Job ID should not be empty
	if jobID == "" {
		t.Error("generateJobID() returned empty string")
	}

	// Job ID should be a hex string (40 characters for 20 bytes)
	if len(jobID) != 40 {
		t.Errorf("Expected job ID length to be 40, got %d", len(jobID))
	}

	// Job ID should only contain hex characters
	for _, c := range jobID {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("Job ID contains non-hex character: %c", c)
			break
		}
	}

	// Multiple calls should generate different IDs
	jobID2 := generateJobID()
	if jobID == jobID2 {
		t.Error("Multiple calls to generateJobID() returned the same ID")
	}
}

func TestCreateProjectStructure_RRBS(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test_project")

	err := createProjectStructure(projectDir, "RRBS", "human", "")

	if err != nil {
		t.Fatalf("createProjectStructure() returned unexpected error: %v", err)
	}

	// Check if base directories were created
	expectedDirs := []string{
		"data",
		"workflow",
		"analysis",
		"config",
		"log",
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(projectDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Directory was not created: %s", fullPath)
		}
	}

	// Check workflow subdirectories
	workflowSubdirs := []string{
		"workflow/QC",
		"workflow/fastqc_raw",
		"workflow/fastqc_clean",
		"workflow/trim",
		"workflow/bsmap",
		"workflow/mCall",
	}

	for _, dir := range workflowSubdirs {
		fullPath := filepath.Join(projectDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Workflow subdirectory was not created: %s", fullPath)
		}
	}
}

func TestCreateProjectStructure_PDX(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test_project_pdx")

	err := createProjectStructure(projectDir, "RRBS", "human", "mouse")

	if err != nil {
		t.Fatalf("createProjectStructure() returned unexpected error: %v", err)
	}

	// Check if PDX-specific directories were created
	pdxDirs := []string{
		"workflow/bsmap/tmp/human",
		"workflow/bsmap/tmp/mouse",
		"workflow/bsmap/human",
		"workflow/bsmap/mouse",
		"workflow/bsmap/Filtered_bams",
	}

	for _, dir := range pdxDirs {
		fullPath := filepath.Join(projectDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("PDX directory was not created: %s", fullPath)
		}
	}
}

func TestCreateProjectStructure_RNAseq(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test_project_rnaseq")

	err := createProjectStructure(projectDir, "RNASEQ", "human", "")

	if err != nil {
		t.Fatalf("createProjectStructure() returned unexpected error: %v", err)
	}

	// Check if RNA-seq specific directories were created
	rnaseqDirs := []string{
		"workflow/star",
		"workflow/htseq",
		"workflow/splicing",
		"analysis/counts",
		"analysis/DEG",
	}

	for _, dir := range rnaseqDirs {
		fullPath := filepath.Join(projectDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("RNA-seq directory was not created: %s", fullPath)
		}
	}
}

func TestCreateProjectStructure_CustomSpecies(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test_project_custom")

	err := createProjectStructure(projectDir, "WGBS", "rat", "mouse")

	if err != nil {
		t.Fatalf("createProjectStructure() returned unexpected error: %v", err)
	}

	// Check if custom species directories were created
	customDirs := []string{
		"workflow/bsmap/tmp/rat",
		"workflow/bsmap/tmp/mouse",
		"workflow/bsmap/rat",
		"workflow/bsmap/mouse",
	}

	for _, dir := range customDirs {
		fullPath := filepath.Join(projectDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Custom species directory was not created: %s", fullPath)
		}
	}
}

func TestCreateProjectStructure_DirectoryCreationError(t *testing.T) {
	// Try to create in an invalid location
	// On most systems, MkdirAll("") might succeed or fail depending on the implementation
	// We'll just verify it doesn't panic
	err := createProjectStructure("", "RRBS", "human", "")

	// The function might handle empty paths gracefully or return an error
	// We just verify it doesn't panic
	if err == nil {
		t.Log("Note: Empty path was handled gracefully without error")
	}
}

func TestCalculateGroupLevels_WithPData(t *testing.T) {
	// Create mock pdata
	pdata := &input.PData{
		Data: map[string]map[string]string{
			"sample1": {"sample_group": "control"},
			"sample2": {"sample_group": "control"},
			"sample3": {"sample_group": "treatment"},
			"sample4": {"sample_group": "treatment"},
		},
	}

	samples := []string{"sample1", "sample2", "sample3", "sample4"}

	groupLevels := calculateGroupLevels(pdata, samples)

	if groupLevels != 2 {
		t.Errorf("Expected 2 group levels, got %d", groupLevels)
	}
}

func TestCalculateGroupLevels_WithoutPData(t *testing.T) {
	groupLevels := calculateGroupLevels(nil, []string{"sample1", "sample2"})

	if groupLevels != 0 {
		t.Errorf("Expected 0 group levels when pdata is nil, got %d", groupLevels)
	}
}

func TestCalculateGroupLevels_EmptyPData(t *testing.T) {
	// Create empty pdata
	pdata := &input.PData{
		Data: map[string]map[string]string{},
	}

	samples := []string{"sample1", "sample2"}

	groupLevels := calculateGroupLevels(pdata, samples)

	if groupLevels != 0 {
		t.Errorf("Expected 0 group levels for empty pdata, got %d", groupLevels)
	}
}

func TestCalculateGroupLevels_WithCondition(t *testing.T) {
	// Create pdata with 'condition' field instead of 'sample_group'
	pdata := &input.PData{
		Data: map[string]map[string]string{
			"sample1": {"condition": "ctrl"},
			"sample2": {"condition": "ctrl"},
			"sample3": {"condition": "treat"},
			"sample4": {"condition": "treat"},
		},
	}

	samples := []string{"sample1", "sample2", "sample3", "sample4"}

	groupLevels := calculateGroupLevels(pdata, samples)

	if groupLevels != 2 {
		t.Errorf("Expected 2 group levels from condition field, got %d", groupLevels)
	}
}

func TestCalculateGroupLevels_MixedFields(t *testing.T) {
	// Create pdata with both sample_group and condition
	pdata := &input.PData{
		Data: map[string]map[string]string{
			"sample1": {"sample_group": "group1", "condition": "ctrl"},
			"sample2": {"sample_group": "group2", "condition": "treat"},
			"sample3": {"sample_group": "group3", "condition": "ctrl"},
			"sample4": {"sample_group": "group4", "condition": "treat"},
		},
	}

	samples := []string{"sample1", "sample2", "sample3", "sample4"}

	groupLevels := calculateGroupLevels(pdata, samples)

	// Should prioritize sample_group over condition
	if groupLevels != 4 {
		t.Errorf("Expected 4 group levels from sample_group field, got %d", groupLevels)
	}
}

func TestCalculateGroupLevels_EmptyGroup(t *testing.T) {
	// Create pdata with empty sample_group
	pdata := &input.PData{
		Data: map[string]map[string]string{
			"sample1": {"sample_group": ""},
			"sample2": {"sample_group": "control"},
		},
	}

	samples := []string{"sample1", "sample2"}

	groupLevels := calculateGroupLevels(pdata, samples)

	// Empty group should be ignored
	if groupLevels != 1 {
		t.Errorf("Expected 1 group level (ignoring empty), got %d", groupLevels)
	}
}

func TestGenerateProjectConfig_RRBS(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := generateProjectConfig(
		configPath,
		"RRBS",
		"human",
		"",
		"/data/fastq",
		"",
		[]string{"sample1", "sample2"},
		tmpDir,
		[]string{"adapter1_seq1", "adapter1_seq2"},
		[]string{"adapter2_seq1", "adapter2_seq2"},
		nil,
		"test_job_123",
	)

	if err != nil {
		t.Fatalf("generateProjectConfig() returned unexpected error: %v", err)
	}

	// Check if config file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	contentStr := string(content)

	// Check for expected fields
	if !containsString(contentStr, "mode: RRBS") {
		t.Error("Config does not contain 'mode: RRBS'")
	}

	if !containsString(contentStr, "species1: human") {
		t.Error("Config does not contain 'species1: human'")
	}

	if !containsString(contentStr, "samples:") {
		t.Error("Config does not contain 'samples:'")
	}

	if !containsString(contentStr, "jobid: test_job_123") {
		t.Error("Config does not contain 'jobid: test_job_123'")
	}
}

func TestGenerateProjectConfig_PDX(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config_pdx.yaml")

	err := generateProjectConfig(
		configPath,
		"RRBS",
		"human",
		"mouse",
		"/data/fastq",
		"",
		[]string{"sample1", "sample2"},
		tmpDir,
		[]string{"adapter1_seq1", "adapter1_seq2"},
		[]string{"adapter2_seq1", "adapter2_seq2"},
		nil,
		"test_job_456",
	)

	if err != nil {
		t.Fatalf("generateProjectConfig() returned unexpected error: %v", err)
	}

	// Read and verify content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	contentStr := string(content)

	// Check for PDX-specific fields
	if !containsString(contentStr, "pdx_mode: true") {
		t.Error("PDX config does not contain 'pdx_mode: true'")
	}

	if !containsString(contentStr, "species2: mouse") {
		t.Error("PDX config does not contain 'species2: mouse'")
	}
}

func TestGenerateProjectConfig_EmptySamples(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config_empty.yaml")

	err := generateProjectConfig(
		configPath,
		"RRBS",
		"human",
		"",
		"/data/fastq",
		"",
		[]string{},
		tmpDir,
		[]string{},
		[]string{},
		nil,
		"test_job_empty",
	)

	if err != nil {
		t.Fatalf("generateProjectConfig() returned unexpected error: %v", err)
	}

	// Check if config file was created even with empty samples
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created for empty samples")
	}
}

func TestGenerateProjectConfig_NestedDirectory(t *testing.T) {
	// Test that nested directories are created
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "deep", "config.yaml")

	err := generateProjectConfig(
		nestedPath,
		"WGBS",
		"human",
		"",
		"/data/fastq",
		"",
		[]string{"sample1"},
		tmpDir,
		[]string{"adapter1"},
		[]string{"adapter2"},
		nil,
		"test_job_nested",
	)

	if err != nil {
		t.Fatalf("generateProjectConfig() returned unexpected error: %v", err)
	}

	// Check if nested config file was created
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("Nested config file was not created")
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
