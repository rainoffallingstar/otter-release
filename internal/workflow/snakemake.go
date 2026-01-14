package workflow

import (
	"fmt"
	"strings"

	"github.com/xdxtools/xdxtools-go/internal/enva"
	"github.com/xdxtools/xdxtools-go/internal/logger"
)

const DefaultFallbackEnv = "xdxtools-snakemake"

// SnakemakeExecutor executes Snakemake workflows
type SnakemakeExecutor struct {
	WorkflowIdx  string
	Step         int
	ConfigFile   string
	Options      *WorkflowOptions
	CondaEnv     string // Conda environment for Snakemake
	FallbackEnv  string // Fallback environment (default: xdxtools-snakemake)
	NoFallback   bool   // Disable automatic fallback
}

// NewSnakemakeExecutor creates a new Snakemake executor
func NewSnakemakeExecutor(workflowIdx string, step int, configFile string, options *WorkflowOptions, condaEnv string) *SnakemakeExecutor {
	return &SnakemakeExecutor{
		WorkflowIdx:  workflowIdx,
		Step:         step,
		ConfigFile:   configFile,
		Options:      options,
		CondaEnv:     condaEnv,
		FallbackEnv:  DefaultFallbackEnv,
		NoFallback:   false,
	}
}

// SetFallbackConfig sets the fallback configuration
func (e *SnakemakeExecutor) SetFallbackConfig(fallbackEnv string, noFallback bool) {
	e.FallbackEnv = fallbackEnv
	e.NoFallback = noFallback
}

// ValidateAndFallback validates the conda environment and falls back if needed
func (e *SnakemakeExecutor) ValidateAndFallback() error {
	if e.NoFallback {
		logger.Debug("Fallback disabled, skipping validation")
		return nil
	}

	// Get the environment to validate
	envToValidate := e.CondaEnv
	logger.Infof("Validating conda environment: %s", getEnvDisplayName(envToValidate))

	// Validate the environment
	if err := enva.ValidateEnvironment(envToValidate); err != nil {
		// Validation failed
		if envToValidate == "" || envToValidate == e.FallbackEnv {
			// Already using fallback or system snakemake, no more options
			return fmt.Errorf("snakemake environment validation failed: %w", err)
		}

		// Fall back to default environment
		logger.Warnf("Environment '%s' validation failed: %v", getEnvDisplayName(envToValidate), err)
		logger.Infof("Falling back to '%s' environment...", e.FallbackEnv)

		e.CondaEnv = e.FallbackEnv

		// Validate the fallback environment
		if err := enva.ValidateEnvironment(e.CondaEnv); err != nil {
			return fmt.Errorf("fallback environment '%s' also failed: %w", e.FallbackEnv, err)
		}

		logger.Infof("Successfully using fallback environment: %s", e.FallbackEnv)
	} else {
		logger.Infof("Environment validation successful: %s", getEnvDisplayName(envToValidate))
	}

	return nil
}

// getEnvDisplayName returns a display name for the environment
func getEnvDisplayName(env string) string {
	if env == "" {
		return "system"
	}
	return env
}

// BuildCommand builds the Snakemake command
func (e *SnakemakeExecutor) BuildCommand() []string {
	cmd := []string{}

	// If conda environment is specified, use enva (if available) or conda run
	if e.CondaEnv != "" {
		if enva.IsAvailable() {
			// 使用 enva: enva run <env> -- snakemake <args...>
			cmd = append(cmd, "enva", "run", e.CondaEnv, "--")
			logger.Debugf("Using enva for optimal performance")
		} else {
			// 回退到 conda: conda run -n <env> --no-capture-output snakemake <args...>
			cmd = append(cmd, "conda", "run", "-n", e.CondaEnv, "--no-capture-output")
			logger.Debugf("enva not found, using conda run")
		}
	}

	// Add Snakemake command
	cmd = append(cmd, "snakemake")

	// Dry run
	if e.Options.DryRun {
		cmd = append(cmd, "-n")
	}

	// Cores
	cmd = append(cmd, "--cores", "all")

	// Snakefile
	snakefile := fmt.Sprintf("%s_step%d.snakemake", e.WorkflowIdx, e.Step)
	if e.Options.Snakefile != "" {
		snakefile = e.Options.Snakefile
	}
	cmd = append(cmd, "--snakefile", snakefile)

	// Config file
	logger.Debugf("ConfigFile path: %s", e.ConfigFile)
	cmd = append(cmd, "--configfile", e.ConfigFile)

	// Rerun incomplete
	cmd = append(cmd, "--rerun-incomplete")

	// Jobs
	if e.Options.Jobs > 0 {
		cmd = append(cmd, "--jobs", fmt.Sprintf("%d", e.Options.Jobs))
	}

	// Force all
	if e.Options.ForceAll {
		cmd = append(cmd, "--forceall")
	}

	// Extra arguments
	if len(e.Options.ExtraArgs) > 0 {
		cmd = append(cmd, e.Options.ExtraArgs...)
	}

	logger.Debugf("Snakemake command: %s", strings.Join(cmd, " "))

	return cmd
}

// GetSnakefilePath returns the path to the snakefile
func (e *SnakemakeExecutor) GetSnakefilePath() string {
	return fmt.Sprintf("%s_step%d.snakemake", e.WorkflowIdx, e.Step)
}
