package workflow

import (
	"fmt"
	"strings"

	"github.com/xdxtools/xdxtools-go/internal/enva"
	"github.com/xdxtools/xdxtools-go/internal/logger"
)

// SnakemakeExecutor executes Snakemake workflows
type SnakemakeExecutor struct {
	WorkflowIdx string
	Step        int
	ConfigFile  string
	Options     *WorkflowOptions
	CondaEnv    string // Conda environment for Snakemake
}

// NewSnakemakeExecutor creates a new Snakemake executor
func NewSnakemakeExecutor(workflowIdx string, step int, configFile string, options *WorkflowOptions, condaEnv string) *SnakemakeExecutor {
	return &SnakemakeExecutor{
		WorkflowIdx: workflowIdx,
		Step:        step,
		ConfigFile:  configFile,
		Options:     options,
		CondaEnv:    condaEnv,
	}
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
