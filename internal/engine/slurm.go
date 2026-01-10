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

	"github.com/xdxtools/xdxtools-go/internal/logger"
)

// SlurmEngine represents a Slurm cluster execution engine
type SlurmEngine struct {
	partition  string
	cores      int
	memory     string
	jobName    string
	maxRetries int
	scriptPath string
	status     *Status
}

// SlurmConfig represents Slurm-specific configuration
type SlurmConfig struct {
	Partition  string `json:"partition"`
	Cores      int    `json:"cores"`
	Memory     string `json:"memory"`
	JobName    string `json:"job_name"`
	MaxRetries int    `json:"max_retries"`
}

// NewSlurmEngine creates a new Slurm engine
func NewSlurmEngine(config *SlurmConfig) *SlurmEngine {
	return &SlurmEngine{
		partition:  config.Partition,
		cores:      config.Cores,
		memory:     config.Memory,
		jobName:    config.JobName,
		maxRetries: config.MaxRetries,
		status: &Status{
			State:     StatusPending,
			StartTime: time.Now(),
		},
	}
}

// Execute executes a command via Slurm
func (e *SlurmEngine) Execute(cmd []string) error {
	e.status.State = StatusRunning
	e.status.Message = "Executing via Slurm"

	// Generate Slurm script
	scriptPath, err := e.generateSlurmScript(cmd)
	if err != nil {
		return fmt.Errorf("failed to generate Slurm script: %w", err)
	}
	e.scriptPath = scriptPath

	// Submit job
	jobID, err := e.submitJob(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to submit Slurm job: %w", err)
	}

	e.status.JobID = jobID
	logger.Infof("Slurm job submitted: %s", jobID)

	// Wait for completion
	if err := e.waitForCompletion(); err != nil {
		return fmt.Errorf("job execution failed: %w", err)
	}

	return nil
}

// ExecuteWithOutput executes a command and returns output
func (e *SlurmEngine) ExecuteWithOutput(cmd []string) (string, error) {
	// For Slurm engine, ExecuteWithOutput first submits the job and then
	// waits for completion while collecting output

	// Generate and submit the job
	scriptPath, err := e.generateSlurmScript(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to generate Slurm script: %w", err)
	}
	e.scriptPath = scriptPath

	// Submit job
	jobID, err := e.submitJob(scriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to submit Slurm job: %w", err)
	}

	e.status.JobID = jobID
	logger.Infof("Slurm job submitted: %s", jobID)

	// Wait for completion and collect output
	if err := e.waitForCompletion(); err != nil {
		return "", fmt.Errorf("job execution failed: %w", err)
	}

	// Collect job output using sacct
	return e.collectJobOutput(jobID)
}

// GetName returns the engine name
func (e *SlurmEngine) GetName() EngineType {
	return EngineSlurm
}

// GetStatus returns the current status
func (e *SlurmEngine) GetStatus() *Status {
	return e.status
}

// Wait waits for the job to complete
func (e *SlurmEngine) Wait() error {
	return e.waitForCompletion()
}

// Kill terminates the Slurm job
func (e *SlurmEngine) Kill() error {
	if e.status.JobID == "" {
		return fmt.Errorf("no job ID available")
	}

	cmd := exec.Command("scancel", e.status.JobID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to cancel job: %w", err)
	}

	e.status.State = StatusKilled
	e.status.EndTime = time.Now()
	logger.Infof("Slurm job cancelled: %s", e.status.JobID)

	return nil
}

// generateSlurmScript generates a Slurm batch script
func (e *SlurmEngine) generateSlurmScript(cmd []string) (string, error) {
	// Create temporary directory
	tmpDir := os.TempDir()
	scriptPath := filepath.Join(tmpDir, fmt.Sprintf("xdxtools_%s.sh", e.jobName))

	// Slurm script template
	const scriptTemplate = `#!/bin/bash
#SBATCH --job-name={{.JobName}}
#SBATCH --partition={{.Partition}}
#SBATCH --cpus-per-task={{.Cores}}
#SBATCH --mem={{.Memory}}
#SBATCH --output={{.OutputFile}}
#SBATCH --error={{.ErrorFile}}

set -e

# Execute command
{{range .Commands}}{{.}}
{{end}}

# Touch success file
touch {{.SuccessFile}}
`

	data := struct {
		JobName     string
		Partition   string
		Cores       int
		Memory      string
		OutputFile  string
		ErrorFile   string
		SuccessFile string
		Commands    []string
	}{
		JobName:     e.jobName,
		Partition:   e.partition,
		Cores:       e.cores,
		Memory:      e.memory,
		OutputFile:  filepath.Join(tmpDir, fmt.Sprintf("%s.out", e.jobName)),
		ErrorFile:   filepath.Join(tmpDir, fmt.Sprintf("%s.err", e.jobName)),
		SuccessFile: filepath.Join(tmpDir, fmt.Sprintf("%s.success", e.jobName)),
		Commands:    cmd,
	}

	tmpl, err := template.New("slurm").Parse(scriptTemplate)
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

// submitJob submits a Slurm batch job
func (e *SlurmEngine) submitJob(scriptPath string) (string, error) {
	cmd := exec.Command("sbatch", scriptPath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("sbatch command failed: %w", err)
	}

	// Parse job ID from output
	outputStr := strings.TrimSpace(string(output))
	// Expected format: "Submitted batch job 12345"
	parts := strings.Fields(outputStr)
	if len(parts) >= 4 {
		return parts[3], nil
	}

	return "", fmt.Errorf("failed to parse job ID from output: %s", outputStr)
}

// waitForCompletion waits for the Slurm job to complete
func (e *SlurmEngine) waitForCompletion() error {
	if e.status.JobID == "" {
		return fmt.Errorf("no job ID available")
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status, err := e.checkJobStatus(e.status.JobID)
			if err != nil {
				logger.Warnf("Failed to check job status: %v", err)
				continue
			}

			e.status.Progress = status.Progress
			e.status.Message = status.Message

			switch status.State {
			case StatusCompleted:
				e.status.State = StatusCompleted
				e.status.EndTime = time.Now()
				logger.Info("Slurm job completed successfully")
				return nil
			case StatusFailed:
				e.status.State = StatusFailed
				e.status.EndTime = time.Now()
				return fmt.Errorf("Slurm job failed")
			case StatusRunning:
				logger.Debugf("Slurm job running: %s", status.Message)
			}
		}
	}
}

// checkJobStatus checks the status of a Slurm job
func (e *SlurmEngine) checkJobStatus(jobID string) (*Status, error) {
	cmd := exec.Command("squeue", "-j", jobID, "-o", "%T,%L")
	output, err := cmd.Output()
	if err != nil {
		// Job might have completed
		return &Status{
			State:   StatusCompleted,
			Message: "Job completed",
		}, nil
	}

	outputStr := strings.TrimSpace(string(output))
	parts := strings.SplitN(outputStr, ",", 2)

	if len(parts) >= 2 {
		state := parts[0]
		message := parts[1]

		switch state {
		case "RUNNING":
			return &Status{
				State:   StatusRunning,
				Message: message,
			}, nil
		case "FAILED":
			return &Status{
				State:   StatusFailed,
				Message: message,
			}, nil
		}
	}

	return &Status{
		State:   StatusRunning,
		Message: "Checking...",
	}, nil
}

// collectJobOutput collects the output of a completed Slurm job
func (e *SlurmEngine) collectJobOutput(jobID string) (string, error) {
	// Use sacct to get job output
	cmd := exec.Command("sacct", "-j", jobID, "-X", "-o", "JobID,JobName,State,ExitCode", "--parsable2")
	output, err := cmd.Output()
	if err != nil {
		logger.Warnf("Failed to collect job output: %v", err)
		return "", fmt.Errorf("failed to collect job output: %w", err)
	}

	outputStr := strings.TrimSpace(string(output))
	logger.Infof("Job output collected: %s", outputStr)

	return outputStr, nil
}
