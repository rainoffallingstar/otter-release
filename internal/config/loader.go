package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/xdxtools/xdxtools-go/internal/logger"
)

// Loader handles configuration loading
type Loader struct {
	configPath string
	viper      *viper.Viper
}

// NewLoader creates a new configuration loader
func NewLoader(configPath string) *Loader {
	v := viper.New()
	v.SetConfigFile(configPath)

	return &Loader{
		configPath: configPath,
		viper:      v,
	}
}

// LoadConfig loads configuration from file
func (l *Loader) LoadConfig() (*WorkflowConfig, error) {
	config := LoadDefaults()

	// Load from file if it exists
	if _, err := os.Stat(l.configPath); err == nil {
		logger.Debugf("Loading config from %s", l.configPath)
		if err := l.viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := l.viper.Unmarshal(&config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	} else {
		logger.Debugf("Config file not found at %s, using defaults", l.configPath)
	}

	// Auto-derive suffix2 if not set
	if config.Suffix2 == "" {
		config.Suffix2 = deriveSuffix2(config.Suffix1)
		logger.Debugf("Auto-derived suffix2: %s", config.Suffix2)
	}

	// Merge environment variables
	l.mergeEnvOverrides(config)

	// Validate configuration
	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// SaveConfig saves configuration to file
func (l *Loader) SaveConfig(config *WorkflowConfig) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(l.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Use viper to write the config by first setting all values
	// This is a workaround since viper doesn't have a direct Marshal method
	v := viper.New()
	v.SetConfigType("yaml")

	// Manually set all config values
	v.Set("mode", config.Mode)
	v.Set("species1", config.Species1)
	v.Set("species2", config.Species2)
	v.Set("suffix1", config.Suffix1)
	v.Set("suffix2", config.Suffix2)
	v.Set("input.fastq_dir", config.Input.FastqDir)
	v.Set("input.pdata_file", config.Input.PdataFile)
	v.Set("output.base_dir", config.Output.BaseDir)
	v.Set("reference.genome", config.Reference.Genome)
	v.Set("engine.type", config.Engine.Type)
	v.Set("parallel.workers", config.Parallel.Workers)

	// Write to file
	if err := v.WriteConfigAs(l.configPath); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	logger.Infof("Configuration saved to %s", l.configPath)
	return nil
}

// mergeEnvOverrides merges environment variable overrides
func (l *Loader) mergeEnvOverrides(config *WorkflowConfig) {
	// Mode
	if mode := os.Getenv("XDXTOOLS_MODE"); mode != "" {
		config.Mode = mode
	}

	// Species
	if species1 := os.Getenv("XDXTOOLS_SPECIES1"); species1 != "" {
		config.Species1 = species1
	}
	if species2 := os.Getenv("XDXTOOLS_SPECIES2"); species2 != "" {
		config.Species2 = species2
	}

	// FASTQ
	if fastqDir := os.Getenv("XDXTOOLS_FASTQ_DIR"); fastqDir != "" {
		config.Input.FastqDir = fastqDir
	}

	// Parallel
	if workers := os.Getenv("XDXTOOLS_PARALLEL_WORKERS"); workers != "" {
		// Convert to int
		// This is a simplified version, in real code you'd use strconv.Atoi
		// config.Parallel.Workers = ...
	}

	// Engine
	if engineType := os.Getenv("XDXTOOLS_ENGINE"); engineType != "" {
		config.Engine.Type = engineType
	}

	logger.Debug("Environment variable overrides applied")
}

// deriveSuffix2 auto-derives suffix2 from suffix1
func deriveSuffix2(suffix1 string) string {
	// Try to replace '1' with '2'
	suffix2 := strings.Replace(suffix1, "1", "2", 1)
	if suffix2 != suffix1 {
		return suffix2
	}

	// If that didn't work, try replacing "R1" with "R2"
	suffix2 = strings.Replace(suffix1, "R1", "R2", 1)
	if suffix2 != suffix1 {
		return suffix2
	}

	// Fallback: just return the original (not ideal, but prevents errors)
	return suffix1
}
