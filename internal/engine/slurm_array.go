package engine

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/xdxtools/xdxtools-go/internal/config"
	"github.com/xdxtools/xdxtools-go/internal/logger"
)

// SlurmArrayEngine represents a SLURM Job Array engine for multi-sample parallelization
type SlurmArrayEngine struct {
	*SlurmEngine
	samples       []string
	stepResource  *config.StepResource
	arraySize    int
	maxArrayJobs int
}

// NewSlurmArrayEngine creates a new SlurmArrayEngine
func NewSlurmArrayEngine(config *SlurmConfig, samples []string, stepResource *config.StepResource) *SlurmArrayEngine {
	slurmEngine := NewSlurmEngine(config)

	return &SlurmArrayEngine{
		SlurmEngine:  slurmEngine,
		samples:       samples,
		stepResource:  stepResource,
		arraySize:    len(samples),
		maxArrayJobs: stepResource.MaxJobs,
	}
}

// ExecuteStepWithArray executes a workflow step using SLURM Job Array for multi-sample parallelization
func (e *SlurmArrayEngine) ExecuteStepWithArray(step int, condaEnv string, workflowFile string, stepResource *config.StepResource) error {
	e.status.State = StatusRunning
	e.status.Message = fmt.Sprintf("Executing Step %d with Job Array (%d samples)", step, e.arraySize)

	// Use updated MaxJobs from stepResource for unified parallelization control
	maxJobs := stepResource.MaxJobs
	if maxJobs > 0 {
		e.maxArrayJobs = maxJobs
	}

	// Generate Job Array script
	scriptPath, err := e.generateArrayScript(step, condaEnv, workflowFile)
	if err != nil {
		return fmt.Errorf("failed to generate Job Array script: %w", err)
	}

	e.scriptPath = scriptPath

	// Submit Job Array
	jobID, err := e.submitArrayJob(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to submit Job Array: %w", err)
	}

	e.status.JobID = jobID
	logger.Infof("SLURM Job Array submitted: %s (Array size: %d)", jobID, e.arraySize)

	// Wait for completion
	if err := e.waitForArrayJob(jobID); err != nil {
		return fmt.Errorf("Job Array execution failed: %w", err)
	}

	return nil
}

// generateArrayScript generates a SLURM Job Array batch script
func (e *SlurmArrayEngine) generateArrayScript(step int, condaEnv string, workflowFile string) (string, error) {
	// Use log directory if set, otherwise fall back to temp directory
	tmpDir := e.logDir
	if tmpDir == "" {
		tmpDir = os.TempDir()
	}
	scriptPath := filepath.Join(tmpDir, fmt.Sprintf("xdxtools_array_step%d_%s.sh", step, e.jobName))

	// Prepare sample array for script
	sampleLines := make([]string, len(e.samples))
	for i, sample := range e.samples {
		sampleLines[i] = fmt.Sprintf(`SAMPLES[%d]=%q`, i, sample)
	}

	// Job Array script template
	const scriptTemplate = `#!/bin/bash
#SBATCH --job-name={{.JobName}}
#SBATCH --partition={{.Partition}}
#SBATCH --cpus-per-task={{.Cores}}
#SBATCH --mem={{.Memory}}
#SBATCH --output={{.OutputFile}}
#SBATCH --error={{.ErrorFile}}
#SBATCH --array=0-{{.ArraySize}}{{.MaxArrayJobs}}
#SBATCH --time=24:00:00

set -e

# Sample array
{{.SampleLines}}

# Change to working directory
cd {{.WorkDir}}

# Get sample name for this array task
SAMPLE_NAME=${SAMPLES[$SLURM_ARRAY_TASK_ID]}

# Execute step using conda environment
{{.CondaCommand}} snakemake --cores all --snakefile {{.WorkflowFile}} --config "SIDs=[$SAMPLE_NAME]"

# Touch success file for this task
touch {{.SuccessFile}}.$SLURM_ARRAY_TASK_ID
`

	data := struct {
		JobName       string
		Partition     string
		Cores         int
		Memory        string
		WorkDir       string
		OutputFile    string
		ErrorFile     string
		SuccessFile   string
		ArraySize     int
		MaxArrayJobs  string
		SampleLines   string
		CondaCommand  string
		Step          int
		WorkflowFile  string
	}{
		JobName:     fmt.Sprintf("%s_step%d_array", e.jobName, step),
		Partition:   e.partition,
		Cores:       e.stepResource.Cores,
		Memory:      e.stepResource.Memory,
		WorkDir:     e.getWorkDir(),
		OutputFile:  filepath.Join(tmpDir, fmt.Sprintf("%s_step%d_array_%%A_%%a.out", e.jobName, step)),
		ErrorFile:   filepath.Join(tmpDir, fmt.Sprintf("%s_step%d_array_%%A_%%a.err", e.jobName, step)),
		SuccessFile: filepath.Join(tmpDir, fmt.Sprintf("%s_step%d_array_success", e.jobName, step)),
		ArraySize:   e.arraySize,
		MaxArrayJobs: func() string {
			if e.maxArrayJobs > 0 {
				return fmt.Sprintf("%%%d", e.maxArrayJobs)
			}
			return ""
		}(),
		SampleLines: strings.Join(sampleLines, "\n"),
		CondaCommand: func() string {
			if condaEnv != "" {
				return fmt.Sprintf("conda run -n %s", condaEnv)
			}
			return ""
		}(),
		Step:         step,
		WorkflowFile: workflowFile,
	}

	tmpl, err := template.New("slurm_array").Parse(scriptTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	// Write script to file
	if err := os.WriteFile(scriptPath, buf.Bytes(), 0755); err != nil {
		return "", err
	}

	return scriptPath, nil
}

// submitArrayJob submits a SLURM Job Array batch job
func (e *SlurmArrayEngine) submitArrayJob(scriptPath string) (string, error) {
	cmd := exec.Command("sbatch", scriptPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		stdoutStr := stdout.String()
		if stderrStr != "" {
			return "", fmt.Errorf("sbatch command failed: %w\nstderr: %s", err, stderrStr)
		}
		return "", fmt.Errorf("sbatch command failed: %w\nstdout: %s", err, stdoutStr)
	}

	// Parse job ID from output
	outputStr := strings.TrimSpace(stdout.String())
	// Expected format: "Submitted batch job 12345[0-9]"
	parts := strings.Fields(outputStr)
	if len(parts) >= 4 {
		return parts[3], nil
	}

	return "", fmt.Errorf("failed to parse Job Array ID from output: %s", outputStr)
}

// waitForArrayJob waits for the SLURM Job Array to complete
func (e *SlurmArrayEngine) waitForArrayJob(jobID string) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	completedTasks := 0
	totalTasks := e.arraySize

	for {
		select {
		case <-ticker.C:
			status, err := e.checkArrayJobStatus(jobID)
			if err != nil {
				logger.Warnf("Failed to check Job Array status: %v", err)
				continue
			}

			completedTasks = status.Completed
			e.status.Progress = (completedTasks * 100) / totalTasks
			e.status.Message = fmt.Sprintf("Job Array running: %d/%d tasks completed", completedTasks, totalTasks)

			logger.Infof("Job Array progress: %d/%d tasks completed", completedTasks, totalTasks)

			if completedTasks >= totalTasks {
				e.status.State = StatusCompleted
				e.status.EndTime = time.Now()
				e.status.Message = "Job Array completed successfully"
				logger.Info("Job Array completed successfully")
				return nil
			}

			// Check if any tasks failed
			if status.Failed > 0 {
				e.status.State = StatusFailed
				e.status.EndTime = time.Now()
				e.status.Message = fmt.Sprintf("Job Array failed: %d tasks failed", status.Failed)
				return fmt.Errorf("Job Array execution failed: %d tasks failed", status.Failed)
			}
		}
	}
}

// ArrayJobStatus represents the status of a Job Array
type ArrayJobStatus struct {
	Completed int
	Running   int
	Pending   int
	Failed    int
}

// checkArrayJobStatus checks the status of a SLURM Job Array
func (e *SlurmArrayEngine) checkArrayJobStatus(jobID string) (*ArrayJobStatus, error) {
	// Extract base job ID (without array index)
	baseJobID := strings.Split(jobID, "[")[0]

	// Use squeue to check job array status
	cmd := exec.Command("squeue", "-j", baseJobID, "-o", "%T", "--noheader")
	output, err := cmd.Output()
	if err != nil {
		// Job might have completed, use sacct to get final status
		return e.getArrayJobFinalStatus(baseJobID)
	}

	outputStr := strings.TrimSpace(string(output))
	lines := strings.Split(outputStr, "\n")

	status := &ArrayJobStatus{
		Completed: 0,
		Running:   0,
		Pending:   0,
		Failed:    0,
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		switch line {
		case "RUNNING":
			status.Running++
		case "PENDING":
			status.Pending++
		case "COMPLETED":
			status.Completed++
		case "FAILED", "CANCELLED", "TIMEOUT":
			status.Failed++
		}
	}

	return status, nil
}

// getArrayJobFinalStatus gets the final status using sacct
func (e *SlurmArrayEngine) getArrayJobFinalStatus(jobID string) (*ArrayJobStatus, error) {
	cmd := exec.Command("sacct", "-j", jobID, "-X", "-o", "State", "--noheader", "--parsable2")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get job status: %w", err)
	}

	outputStr := strings.TrimSpace(string(output))
	lines := strings.Split(outputStr, "\n")

	status := &ArrayJobStatus{
		Completed: 0,
		Running:   0,
		Pending:   0,
		Failed:    0,
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "JobID" {
			continue
		}

		switch line {
		case "COMPLETED":
			status.Completed++
		case "FAILED", "CANCELLED", "TIMEOUT":
			status.Failed++
		}
	}

	return status, nil
}

// GetArraySize returns the size of the job array
func (e *SlurmArrayEngine) GetArraySize() int {
	return e.arraySize
}

// GetSamples returns the samples assigned to this array job
func (e *SlurmArrayEngine) GetSamples() []string {
	return e.samples
}
