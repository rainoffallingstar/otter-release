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

func TestNewLoader(t *testing.T) {
	loader := NewLoader("/path/to/config.yaml")

	if loader == nil {
		t.Fatal("NewLoader returned nil")
	}

	if loader.configPath != "/path/to/config.yaml" {
		t.Errorf("Expected configPath to be '/path/to/config.yaml', got '%s'", loader.configPath)
	}

	if loader.viper == nil {
		t.Error("Expected viper to be non-nil")
	}
}

func TestLoadConfig_Default(t *testing.T) {
	// Test loading with no config file (uses defaults)
	loader := NewLoader("/nonexistent/config.yaml")

	config, err := loader.LoadConfig()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	// Check default values
	if config.Mode != "RRBS" {
		t.Errorf("Expected Mode to be 'RRBS', got '%s'", config.Mode)
	}

	if config.Species1 != "human" {
		t.Errorf("Expected Species1 to be 'human', got '%s'", config.Species1)
	}

	if config.Suffix1 != "_R1.fastq.gz" {
		t.Errorf("Expected Suffix1 to be '_R1.fastq.gz', got '%s'", config.Suffix1)
	}
}

func TestLoadConfig_AutoDeriveSuffix2(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create a config file with suffix1 but no suffix2
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `mode: "RRBS"
species1: "human"
suffix1: "_R1.fastq.gz"
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	loader := NewLoader(configFile)

	config, err := loader.LoadConfig()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if config.Suffix2 != "_R2.fastq.gz" {
		t.Errorf("Expected Suffix2 to be auto-derived to '_R2.fastq.gz', got '%s'", config.Suffix2)
	}
}

func TestLoadConfig_LoadFromFile(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create a config file
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `mode: "WGBS"
species1: "mouse"
suffix1: "_R1.fq.gz"
suffix2: "_R2.fq.gz"
engine:
  type: "slurm"
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	loader := NewLoader(configFile)

	config, err := loader.LoadConfig()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if config.Mode != "WGBS" {
		t.Errorf("Expected Mode to be 'WGBS', got '%s'", config.Mode)
	}

	if config.Species1 != "mouse" {
		t.Errorf("Expected Species1 to be 'mouse', got '%s'", config.Species1)
	}

	if config.Suffix1 != "_R1.fq.gz" {
		t.Errorf("Expected Suffix1 to be '_R1.fq.gz', got '%s'", config.Suffix1)
	}

	if config.Suffix2 != "_R2.fq.gz" {
		t.Errorf("Expected Suffix2 to be '_R2.fq.gz', got '%s'", config.Suffix2)
	}
}

func TestDeriveSuffix2(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"_R1.fastq.gz", "_R2.fastq.gz"},
		{"_1.fastq.gz", "_2.fastq.gz"},
		{"_R1.fq", "_R2.fq"},
		{"_R1", "_R2"},
		{"_1", "_2"},
		{"sample_R1.fastq.gz", "sample_R2.fastq.gz"},
		{"sample_1.fastq.gz", "sample_2.fastq.gz"},
		{"no_number.fastq.gz", "no_number.fastq.gz"}, // No change
	}

	for _, test := range tests {
		result := deriveSuffix2(test.input)
		if result != test.expected {
			t.Errorf("deriveSuffix2(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestMergeEnvOverrides_Mode(t *testing.T) {
	// Set environment variable
	original := os.Getenv("XDXTOOLS_MODE")
	defer os.Setenv("XDXTOOLS_MODE", original)

	os.Setenv("XDXTOOLS_MODE", "WGBS")

	loader := NewLoader("/nonexistent/config.yaml")

	// Access the private method through loader
	// We can't directly test private methods, so we'll test through LoadConfig
	config2, err := loader.LoadConfig()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Environment variable should override default
	if config2.Mode != "WGBS" {
		t.Errorf("Expected Mode to be 'WGBS' from env var, got '%s'", config2.Mode)
	}
}

func TestMergeEnvOverrides_Species(t *testing.T) {
	// Save original values
	originalSpecies1 := os.Getenv("XDXTOOLS_SPECIES1")
	originalSpecies2 := os.Getenv("XDXTOOLS_SPECIES2")
	defer os.Setenv("XDXTOOLS_SPECIES1", originalSpecies1)
	defer os.Setenv("XDXTOOLS_SPECIES2", originalSpecies2)

	// Set environment variables
	os.Setenv("XDXTOOLS_SPECIES1", "mouse")
	os.Setenv("XDXTOOLS_SPECIES2", "human")

	loader := NewLoader("/nonexistent/config.yaml")

	config, err := loader.LoadConfig()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if config.Species1 != "mouse" {
		t.Errorf("Expected Species1 to be 'mouse' from env var, got '%s'", config.Species1)
	}

	if config.Species2 != "human" {
		t.Errorf("Expected Species2 to be 'human' from env var, got '%s'", config.Species2)
	}
}

func TestMergeEnvOverrides_Engine(t *testing.T) {
	original := os.Getenv("XDXTOOLS_ENGINE")
	defer os.Setenv("XDXTOOLS_ENGINE", original)

	os.Setenv("XDXTOOLS_ENGINE", "slurm")

	loader := NewLoader("/nonexistent/config.yaml")

	config, err := loader.LoadConfig()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if config.Engine.Type != "slurm" {
		t.Errorf("Expected Engine.Type to be 'slurm' from env var, got '%s'", config.Engine.Type)
	}
}

func TestSaveConfig(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	configFile := filepath.Join(tmpDir, "config.yaml")

	loader := NewLoader(configFile)

	config := LoadDefaults()
	config.Mode = "RNASEQ"
	config.Species1 = "mouse"

	err := loader.SaveConfig(config)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check if file was created
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify the content was written
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	contentStr := string(content)
	if contentStr == "" {
		t.Error("Config file is empty")
	}

	// Load it back and verify
	loader2 := NewLoader(configFile)
	config2, err := loader2.LoadConfig()

	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if config2.Mode != "RNASEQ" {
		t.Errorf("Expected Mode to be 'RNASEQ' after save/load, got '%s'", config2.Mode)
	}

	if config2.Species1 != "mouse" {
		t.Errorf("Expected Species1 to be 'mouse' after save/load, got '%s'", config2.Species1)
	}
}

func TestLoadConfig_InvalidConfig(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create an invalid config file (malformed YAML)
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `mode: "RRBS"
species1:
  invalid: yaml: syntax
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	loader := NewLoader(configFile)

	_, err := loader.LoadConfig()

	if err == nil {
		t.Error("Expected error for invalid config file, got nil")
	}
}

func TestLoadConfig_MergeFileAndDefaults(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create a config file with only some fields
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `mode: "RNASEQ"
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	loader := NewLoader(configFile)

	config, err := loader.LoadConfig()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Mode should be from file
	if config.Mode != "RNASEQ" {
		t.Errorf("Expected Mode to be 'RNASEQ' from file, got '%s'", config.Mode)
	}

	// Species1 should be from defaults
	if config.Species1 != "human" {
		t.Errorf("Expected Species1 to be 'human' from defaults, got '%s'", config.Species1)
	}

	// Other defaults should be present
	if config.Suffix1 != "_R1.fastq.gz" {
		t.Errorf("Expected Suffix1 to be '_R1.fastq.gz' from defaults, got '%s'", config.Suffix1)
	}
}
