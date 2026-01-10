package engine

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/xdxtools/xdxtools-go/internal/logger"
)

// LocalEngine represents a local execution engine
type LocalEngine struct {
	maxCores  int
	maxMemory string
	cmd       *exec.Cmd
	ctx       context.Context
	cancel    context.CancelFunc
	status    *Status
}

// LocalConfig represents local execution configuration
type LocalConfig struct {
	MaxCores  int    `json:"max_cores"`
	MaxMemory string `json:"max_memory"`
}

// NewLocalEngine creates a new local engine
func NewLocalEngine(config *LocalConfig) *LocalEngine {
	return &LocalEngine{
		maxCores:  config.MaxCores,
		maxMemory: config.MaxMemory,
		status: &Status{
			State:     StatusPending,
			StartTime: time.Now(),
		},
	}
}

// Execute executes a command locally
func (e *LocalEngine) Execute(cmd []string) error {
	e.status.State = StatusRunning
	e.status.Message = "Executing locally"

	// Create context with timeout
	e.ctx, e.cancel = context.WithTimeout(context.Background(), 24*time.Hour)
	defer e.cancel()

	// Create command
	e.cmd = exec.CommandContext(e.ctx, cmd[0], cmd[1:]...)

	// Start execution
	logger.Debugf("Executing command: %s", strings.Join(cmd, " "))

	if err := e.cmd.Start(); err != nil {
		e.status.State = StatusFailed
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Wait for completion
	if err := e.cmd.Wait(); err != nil {
		e.status.State = StatusFailed
		return fmt.Errorf("command execution failed: %w", err)
	}

	e.status.State = StatusCompleted
	e.status.EndTime = time.Now()
	logger.Info("Local execution completed successfully")

	return nil
}

// ExecuteWithOutput executes a command and returns output
func (e *LocalEngine) ExecuteWithOutput(cmd []string) (string, error) {
	e.status.State = StatusRunning
	e.status.Message = "Executing locally"

	// Create command
	cmdExec := exec.Command(cmd[0], cmd[1:]...)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmdExec.Stdout = &stdout
	cmdExec.Stderr = &stderr

	logger.Debugf("Executing command: %s", strings.Join(cmd, " "))

	// Run command
	err := cmdExec.Run()

	e.status.State = StatusCompleted
	e.status.EndTime = time.Now()

	if err != nil {
		return "", fmt.Errorf("command execution failed: %w\nstderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// GetName returns the engine name
func (e *LocalEngine) GetName() EngineType {
	return EngineLocal
}

// GetStatus returns the current status
func (e *LocalEngine) GetStatus() *Status {
	if e.cmd != nil && e.cmd.ProcessState != nil {
		if e.cmd.ProcessState.Exited() {
			if e.cmd.ProcessState.ExitCode() == 0 {
				e.status.State = StatusCompleted
			} else {
				e.status.State = StatusFailed
			}
		} else {
			e.status.State = StatusRunning
		}
	}
	return e.status
}

// Wait waits for the command to complete
func (e *LocalEngine) Wait() error {
	if e.cmd == nil {
		return fmt.Errorf("no command running")
	}
	return e.cmd.Wait()
}

// Kill terminates the running command
func (e *LocalEngine) Kill() error {
	if e.cancel != nil {
		e.cancel()
	}
	if e.cmd != nil && e.cmd.Process != nil {
		if err := e.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}
	e.status.State = StatusKilled
	e.status.EndTime = time.Now()
	logger.Info("Local process terminated")
	return nil
}
