package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xdxtools/xdxtools-go/internal/config"
	"github.com/xdxtools/xdxtools-go/internal/engine"
	"github.com/xdxtools/xdxtools-go/internal/logger"
)

// Manager manages workflow execution
type Manager struct {
	workflow        *Workflow
	state           *State
	dryRun          bool
	condaEnv        string
	fallbackEnv     string // Fallback environment (default: xdxtools-snakemake)
	noFallback      bool   // Disable automatic fallback
	samples         []string
	stepResources   map[int]*config.StepResource
	parallelJobs    int
}

// NewManager creates a new workflow manager
func NewManager(w *Workflow) *Manager {
	return &Manager{
		workflow:    w,
		dryRun:      false,
		condaEnv:    "",
		fallbackEnv: DefaultFallbackEnv,
		noFallback:  false,
	}
}

// SetDryRun sets the dry-run mode for the workflow
func (m *Manager) SetDryRun(dryRun bool) {
	m.dryRun = dryRun
}

// SetCondaEnv sets the conda environment for Snakemake execution
func (m *Manager) SetCondaEnv(condaEnv string) {
	m.condaEnv = condaEnv
}

// SetFallbackConfig sets the fallback environment configuration
func (m *Manager) SetFallbackConfig(fallbackEnv string, noFallback bool) {
	m.fallbackEnv = fallbackEnv
	m.noFallback = noFallback
}

// SetSamples sets the sample names for the workflow
func (m *Manager) SetSamples(samples []string) {
	m.samples = samples
}

// SetStepResources sets the step resources for the workflow
func (m *Manager) SetStepResources(resources map[int]*config.StepResource) {
	m.stepResources = resources
}

// SetParallelJobs sets the number of parallel jobs for both local and SLURM execution
func (m *Manager) SetParallelJobs(jobs int) {
	m.parallelJobs = jobs
}

// SetJobID updates the job ID for a specific step in the state
func (m *Manager) SetJobID(step int, jobID string) error {
	if m.state == nil {
		return nil // State tracking not enabled
	}
	if err := m.state.UpdateStepStatus(step, "running", time.Now(), jobID); err != nil {
		return fmt.Errorf("failed to update job ID: %w", err)
	}
	return m.state.Save()
}

// getStepResource returns the resource configuration for a specific step
func (m *Manager) getStepResource(step int) *config.StepResource {
	if resource, exists := m.stepResources[step]; exists {
		return resource
	}

	// Checker steps inherit from main steps if not explicitly configured
	if step == 102 { // Step 2 Checker
		if resource, exists := m.stepResources[2]; exists {
			// Create a copy to avoid modifying the original
			inherited := *resource
			return &inherited
		}
	} else if step == 103 { // Step 3 Checker
		if resource, exists := m.stepResources[3]; exists {
			// Create a copy to avoid modifying the original
			inherited := *resource
			return &inherited
		}
	}

	// Return default resource
	return config.GetDefaultStepResource(step, m.workflow.Config.Workflow.Mode, config.DetectPDXMode(m.workflow.Config))
}

// shouldUseSingleSampleMode returns true if the step should use single-sample mode
// Step 2 and 3 use single-sample mode, Step 1 and checkers use all-samples mode
func (m *Manager) shouldUseSingleSampleMode(step int) bool {
	// Step 2 and 3 use single-sample mode
	return step == 2 || step == 3
}

// Initialize initializes the workflow
func (m *Manager) Initialize() error {
	logger.Info("Initializing workflow...")

	// Create output directory
	if err := os.MkdirAll(m.workflow.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create logs directory
	logsDir := filepath.Join(m.workflow.OutputDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Initialize file logging
	logFilePath := filepath.Join(logsDir, "xdxtools.log")
	logger.InitWithFile(false, logFilePath)
	logger.Infof("Logging to file: %s", logFilePath)

	// Set log directory on engine if it supports it
	if m.workflow.Engine != nil {
		if err := m.workflow.Engine.SetLogDir(logsDir); err != nil {
			logger.Warnf("Failed to set engine log directory: %v", err)
		}
	}

	// Detect workflow mode and steps
	m.workflow.Status.State = engine.StatusRunning
	m.workflow.Status.Message = "Initializing workflow"

	// Determine steps based on mode
	pdxMode := config.DetectPDXMode(m.workflow.Config)
	m.workflow.Steps = config.GetStepCount(m.workflow.Config.Workflow.Mode, pdxMode)

	logger.Infof("Workflow initialized with %d steps", m.workflow.Steps)

	return nil
}

// ExecuteStep executes a specific step with intelligent parallelization strategy
func (m *Manager) ExecuteStep(step int) error {
	if step < 1 || step > m.workflow.Steps {
		return fmt.Errorf("invalid step: %d (valid range: 1-%d)", step, m.workflow.Steps)
	}

	logger.Infof("Executing step %d/%d", step, m.workflow.Steps)

	// Update state: step started
	if m.state != nil {
		if err := m.state.UpdateStepStatus(step, "running", time.Now(), ""); err != nil {
			logger.Warnf("Failed to update step status: %v", err)
		}
		_ = m.state.Save()
	}

	// Get step resource configuration
	stepResource := m.getStepResource(step)

	// Unified parallelization control via --parallel-jobs parameter
	// For step2 and step3: ensure JobArray is enabled when parallel-jobs > 1
	if m.shouldUseSingleSampleMode(step) && m.parallelJobs > 1 {
		// Force JobArray mode for single-sample steps when using parallel execution
		if !stepResource.JobArray {
			stepResource.JobArray = true
		}
	}

	// Set MaxJobs from parallelJobs for unified control
	// This ensures job-array and local parallel execution use the same value
	if m.parallelJobs > 0 {
		stepResource.MaxJobs = m.parallelJobs
	}

	// Determine workflow index
	workflowIdx := config.GetWorkflowName(m.workflow.Config.Workflow.Mode, config.DetectPDXMode(m.workflow.Config))

	// Create executor
	executor := NewSnakemakeExecutor(workflowIdx, step, m.getConfigFile(), &WorkflowOptions{
		DryRun: m.dryRun,
	}, m.condaEnv)

	// Set fallback configuration from manager or config
	if m.fallbackEnv != "" {
		executor.SetFallbackConfig(m.fallbackEnv, m.noFallback)
	} else if m.workflow.Config.Engine.FallbackEnv != "" {
		executor.SetFallbackConfig(m.workflow.Config.Engine.FallbackEnv, m.workflow.Config.Engine.NoFallback)
	}

	// Validate environment and apply fallback if needed
	if err := executor.ValidateAndFallback(); err != nil {
		return fmt.Errorf("environment validation failed: %w", err)
	}

	// Build command (may use fallback environment)
	cmd := executor.BuildCommand()

	// Unified parallelization strategy based on --parallel-jobs parameter
	// Step 2 and 3 use single-sample mode, Step 1 and checkers use all-samples mode
	if m.shouldUseSingleSampleMode(step) {
		// Step 2 and 3: Always use single-sample mode
		// Parallelization is controlled by --parallel-jobs parameter
		if m.parallelJobs > 1 {
			// Use parallel execution
			logger.Infof("Using parallel single-sample strategy for step %d with %d parallel jobs", step, m.parallelJobs)

			switch m.workflow.Engine.(type) {
			case *engine.SlurmArrayEngine:
				// Use Job Array for SLURM environments
				logger.Infof("Using Job Array strategy for step %d with max %d concurrent jobs", step, m.parallelJobs)
				if err := m.executeStepWithJobArray(step, stepResource, cmd); err != nil {
					return fmt.Errorf("step %d Job Array execution failed: %w", step, err)
				}

			case *engine.LocalEngine:
				// Use local parallel execution for local environments
				logger.Infof("Using local parallel strategy for step %d with %d workers", step, m.parallelJobs)
				if err := m.executeStepWithLocalParallel(step, stepResource, cmd); err != nil {
					return fmt.Errorf("step %d parallel execution failed: %w", step, err)
				}

			default:
				// Other engines: use standard execution
				logger.Infof("Using standard execution for step %d", step)
				if err := m.workflow.Engine.Execute(cmd); err != nil {
					return fmt.Errorf("step %d execution failed: %w", step, err)
				}
			}
		} else {
			// Sequential execution (parallel-jobs = 1)
			logger.Infof("Using single-sample sequential execution for step %d", step)
			if err := m.workflow.Engine.Execute(cmd); err != nil {
				return fmt.Errorf("step %d execution failed: %w", step, err)
			}
		}
	} else {
		// Step 1 and checkers: use all-samples execution (standard mode)
		logger.Infof("Using all-samples standard execution for step %d", step)
		if err := m.workflow.Engine.Execute(cmd); err != nil {
			return fmt.Errorf("step %d execution failed: %w", step, err)
		}
	}

	logger.Infof("Step %d completed successfully", step)

	// Get Job ID from engine (for SLURM jobs)
	jobID := ""
	if engineStatus := m.workflow.Engine.GetStatus(); engineStatus != nil && engineStatus.JobID != "" {
		jobID = engineStatus.JobID
	}

	// Update state: step completed
	if m.state != nil {
		if err := m.state.UpdateStepStatus(step, "completed", time.Now(), jobID); err != nil {
			logger.Warnf("Failed to update step status: %v", err)
		}
		_ = m.state.Save()
	}

	return nil
}

// executeStepWithJobArray executes a step using SLURM Job Array
func (m *Manager) executeStepWithJobArray(step int, resource *config.StepResource, cmd []string) error {
	// Check if engine is SlurmArrayEngine
	if arrayEngine, ok := m.workflow.Engine.(*engine.SlurmArrayEngine); ok {
		// Determine workflow file
		workflowIdx := config.GetWorkflowName(m.workflow.Config.Workflow.Mode, config.DetectPDXMode(m.workflow.Config))
		workflowFile := fmt.Sprintf("Beaver%s_step%d.snakemake", workflowIdx, step)

		if err := arrayEngine.ExecuteStepWithArray(step, m.condaEnv, workflowFile, resource); err != nil {
			return err
		}

		// Note: Job ID will be captured in ExecuteStep after this returns
		// The engine's GetStatus() will contain the submitted Job ID
		return nil
	}

	// Fallback to regular engine if SlurmArrayEngine not available
	logger.Warn("SlurmArrayEngine not available, falling back to regular execution")
	return m.workflow.Engine.Execute(cmd)
}

// executeStepWithLocalParallel executes a step using local parallel execution
func (m *Manager) executeStepWithLocalParallel(step int, resource *config.StepResource, cmd []string) error {
	// Check if engine supports parallel execution
	if localEngine, ok := m.workflow.Engine.(*engine.LocalEngine); ok {
		// Determine workflow file
		workflowIdx := config.GetWorkflowName(m.workflow.Config.Workflow.Mode, config.DetectPDXMode(m.workflow.Config))
		workflowFile := fmt.Sprintf("Beaver%s_step%d.snakemake", workflowIdx, step)

		// Pass parallelJobs for unified control
		return localEngine.ExecuteSamples(step, m.samples, m.condaEnv, workflowFile, m.parallelJobs)
	}

	// Fallback to regular engine
	logger.Warn("LocalEngine parallel execution not available, falling back to regular execution")
	return m.workflow.Engine.Execute(cmd)
}

// ExecuteAll executes all workflow steps
func (m *Manager) ExecuteAll() error {
	if err := m.Initialize(); err != nil {
		return err
	}

	// Initialize state
	m.state = NewState(m.workflow.Config.Output.BaseDir, m.workflow.Config.Workflow.JobID)

	// Check for resume mode
	if m.workflow.Options != nil && m.workflow.Options.Resume && m.state.Exists() {
		if err := m.state.Load(); err != nil {
			return fmt.Errorf("failed to load state: %w", err)
		}

		// Validate config consistency
		if err := m.validateConfigForResume(); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}

		// Find last completed step
		lastCompleted := m.state.GetLastCompletedStep()
		logger.Infof("Resuming from step %d (last completed: %d)",
			lastCompleted+1, lastCompleted)

		// Start from next step
		for step := lastCompleted + 1; step <= m.workflow.Steps; step++ {
			if err := m.ExecuteStep(step); err != nil {
				// Mark workflow as failed
				_ = m.state.MarkFailed()
				return fmt.Errorf("failed to execute step %d: %w", step, err)
			}
		}
	} else {
		// Initialize new state
		if err := m.initializeState(); err != nil {
			logger.Warnf("Failed to initialize state: %v (continuing without state tracking)", err)
		}

		// Normal execution: run all steps
		for step := 1; step <= m.workflow.Steps; step++ {
			if err := m.ExecuteStep(step); err != nil {
				// Mark workflow as failed
				if m.state != nil {
					_ = m.state.MarkFailed()
				}
				return fmt.Errorf("failed to execute step %d: %w", step, err)
			}
		}
	}

	m.workflow.Status.State = engine.StatusCompleted
	m.workflow.Status.Message = "Workflow completed"
	m.workflow.Status.EndTime = time.Now()

	// Mark state as completed
	if m.state != nil {
		_ = m.state.MarkCompleted()
	}

	logger.Info("All workflow steps completed successfully")

	return nil
}

// initializeState initializes the state with current workflow configuration
func (m *Manager) initializeState() error {
	if m.state == nil {
		return nil
	}

	// Get partition for state tracking
	partition := ""
	if m.workflow.Config.Engine.Slurm.Partition != "" {
		partition = m.workflow.Config.Engine.Slurm.Partition
	}

	return m.state.Initialize(
		m.workflow.Config.Workflow.JobID,
		m.workflow.Config.Workflow.Mode,
		m.workflow.Config.Workflow.Species.Primary,
		m.workflow.Config.Workflow.Species.Secondary,
		m.samples,
		m.workflow.Config.Engine.Type,
		partition,
	)
}

// validateConfigForResume checks if current config matches saved state
func (m *Manager) validateConfigForResume() error {
	stateData := m.state.GetData()

	// Check critical config parameters
	if stateData.Config.WorkflowMode != m.workflow.Config.Workflow.Mode {
		return fmt.Errorf("workflow mode changed: was %s, now %s",
			stateData.Config.WorkflowMode, m.workflow.Config.Workflow.Mode)
	}

	if stateData.Config.SampleCount != len(m.samples) {
		return fmt.Errorf("sample count changed: was %d, now %d",
			stateData.Config.SampleCount, len(m.samples))
	}

	// Check species consistency
	if stateData.Config.Species1 != m.workflow.Config.Workflow.Species.Primary {
		return fmt.Errorf("species changed: was %s, now %s",
			stateData.Config.Species1, m.workflow.Config.Workflow.Species.Primary)
	}

	// Warn on non-critical changes
	if stateData.Config.Partition != "" && m.workflow.Config.Engine.Slurm.Partition != "" {
		if stateData.Config.Partition != m.workflow.Config.Engine.Slurm.Partition {
			logger.Warnf("Partition changed: %s -> %s",
				stateData.Config.Partition, m.workflow.Config.Engine.Slurm.Partition)
		}
	}

	return nil
}

// CreateFilework creates the directory structure for the workflow
func (m *Manager) CreateFilework() error {
	logger.Info("Creating workflow directory structure...")

	dirs := []string{
		m.workflow.Config.Output.WorkflowDir,
		m.workflow.Config.Output.AnalysisDir,
		m.workflow.Config.Directories.QC.Main,
		m.workflow.Config.Output.TrimDir,
		m.workflow.Config.Directories.BSMAP.Main,
		m.workflow.Config.Directories.MethylationCall,
		m.workflow.Config.Directories.UMX,
		m.workflow.Config.Directories.SIDLog,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	logger.Info("Directory structure created successfully")

	return nil
}

// MoveFastq moves FASTQ files to workflow directory
func (m *Manager) MoveFastq(samples []string, method string) error {
	logger.Infof("Moving FASTQ files using method: %s", method)

	// This would copy/move FASTQ files from input directory to workflow directory
	// Implementation depends on specific requirements

	logger.Info("FASTQ files moved successfully")

	return nil
}

// AggregateResults aggregates workflow results
func (m *Manager) AggregateResults() error {
	logger.Info("Aggregating workflow results...")

	// This would aggregate results from different steps
	// Implementation depends on specific requirements

	logger.Info("Results aggregated successfully")

	return nil
}

// GetStatus returns the current workflow status
func (m *Manager) GetStatus() *engine.Status {
	return m.workflow.Status
}

// getConfigFile returns the path to the Snakemake config file
func (m *Manager) getConfigFile() string {
	configPath := filepath.Join(m.workflow.Config.Directories.Config, "config.yaml")

	// Convert to absolute path to ensure Snakemake can find it
	if !filepath.IsAbs(configPath) {
		// Get current working directory
		if wd, err := os.Getwd(); err == nil {
			configPath = filepath.Join(wd, configPath)
		}
	}

	return configPath
}
