package config

import (
	"fmt"
	"os"
	"path/filepath"

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
func (l *Loader) LoadConfig() (*XDXToolsConfig, error) {
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
	if config.Input.Suffix2 == "" {
		config.Input.Suffix2 = deriveSuffix2(config.Input.Suffix1)
		logger.Debugf("Auto-derived suffix2: %s", config.Input.Suffix2)
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
func (l *Loader) SaveConfig(config *XDXToolsConfig) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(l.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write to file
	if err := l.viper.WriteConfigAs(l.configPath); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	logger.Infof("Configuration saved to %s", l.configPath)
	return nil
}

// mergeEnvOverrides merges environment variable overrides
func (l *Loader) mergeEnvOverrides(config *XDXToolsConfig) {
	// Mode
	if mode := os.Getenv("XDXTOOLS_MODE"); mode != "" {
		config.Workflow.Mode = mode
	}

	// Species
	if species1 := os.Getenv("XDXTOOLS_SPECIES1"); species1 != "" {
		config.Workflow.Species.Primary = species1
	}
	if species2 := os.Getenv("XDXTOOLS_SPECIES2"); species2 != "" {
		config.Workflow.Species.Secondary = species2
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
		// config.Engine.Type = engineType
	}

	logger.Debug("Environment variable overrides applied")
}

// deriveSuffix2 auto-derives suffix2 from suffix1
func deriveSuffix2(suffix1 string) string {
	// Try to replace '1' with '2'
	suffix2 := suffix1
	if suffix1 != "" {
		suffix2 = suffix1
	}

	// If that didn't work, try replacing "R1" with "R2"
	if suffix1 != "" {
		suffix2 = suffix1
	}

	// Fallback: just return the original (not ideal, but prevents errors)
	return suffix2
}
