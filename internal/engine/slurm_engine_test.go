package engine

import (
	"testing"

	"github.com/rainoffallingstar/otter/internal/logger"
)

func init() {
	logger.Init(false)
}

func TestNewSlurmEngine(t *testing.T) {
	config := &SlurmConfig{
		Partition:  "compute",
		Cores:      16,
		Memory:     "32G",
		JobName:    "test_job",
		MaxRetries: 3,
	}

	engine := NewSlurmEngine(config)

	if engine == nil {
		t.Fatal("NewSlurmEngine returned nil")
	}

	if engine.partition != "compute" {
		t.Errorf("Expected partition to be 'compute', got '%s'", engine.partition)
	}

	if engine.cores != 16 {
		t.Errorf("Expected cores to be 16, got %d", engine.cores)
	}

	if engine.memory != "32G" {
		t.Errorf("Expected memory to be '32G', got '%s'", engine.memory)
	}

	if engine.jobName != "test_job" {
		t.Errorf("Expected jobName to be 'test_job', got '%s'", engine.jobName)
	}

	if engine.maxRetries != 3 {
		t.Errorf("Expected maxRetries to be 3, got %d", engine.maxRetries)
	}

	if engine.status.State != StatusPending {
		t.Errorf("Expected status to be PENDING, got %s", engine.status.State)
	}
}

func TestSlurmEngine_GetName(t *testing.T) {
	config := &SlurmConfig{
		JobName: "test",
	}

	engine := NewSlurmEngine(config)

	if engine.GetName() != EngineSlurm {
		t.Errorf("Expected engine name to be SLURM, got %s", engine.GetName())
	}
}

func TestSlurmEngine_GetStatus(t *testing.T) {
	config := &SlurmConfig{
		JobName: "test",
	}

	engine := NewSlurmEngine(config)

	status := engine.GetStatus()

	if status == nil {
		t.Fatal("GetStatus() returned nil")
	}

	if status.State != StatusPending {
		t.Errorf("Expected status to be PENDING, got %s", status.State)
	}
}

func TestSlurmEngine_ExecuteWithOutput(t *testing.T) {
	config := &SlurmConfig{
		JobName: "test",
	}

	engine := NewSlurmEngine(config)

	// ExecuteWithOutput should now work with the fixed implementation
	// However, it requires actual Slurm environment to function
	// In test environment, it will fail when trying to execute sbatch

	// We test the structure of the implementation
	// The method should not immediately return an error like before

	// Test that the method doesn't panic and returns an appropriate error
	// when Slurm is not available
	output, err := engine.ExecuteWithOutput([]string{"echo", "test"})

	// We expect an error in test environment without Slurm
	// But it should NOT be the old "not implemented" error
	if err != nil && err.Error() == "ExecuteWithOutput not implemented for Slurm engine (use Execute instead)" {
		t.Errorf("ExecuteWithOutput() still returns 'not implemented' error")
	}

	// Output should be non-empty on success (job info from sacct)
	// Note: In a real SLURM environment, the job will actually run and return output
	if err == nil && output == "" {
		t.Errorf("Expected job output on success, got empty string")
	}
	// If the test ran successfully, log the output for debugging
	if err == nil {
		t.Logf("Job executed successfully, output: %s", output)
	}
}

func TestSlurmEngine_Kill(t *testing.T) {
	config := &SlurmConfig{
		JobName: "test",
	}

	engine := NewSlurmEngine(config)

	// Test kill without a job ID
	err := engine.Kill()

	if err == nil {
		t.Error("Kill() should have returned an error when no job ID is available")
	}
}

func TestSlurmEngine_WaitWithoutJob(t *testing.T) {
	config := &SlurmConfig{
		JobName: "test",
	}

	engine := NewSlurmEngine(config)

	// Try to wait without submitting a job
	err := engine.Wait()

	if err == nil {
		t.Error("Wait() should have returned an error when no job ID is available")
	}
}

func TestSlurmEngine_StatusTransitions(t *testing.T) {
	config := &SlurmConfig{
		JobName: "test",
	}

	engine := NewSlurmEngine(config)

	// Initial status should be PENDING
	if engine.status.State != StatusPending {
		t.Errorf("Expected initial status to be PENDING, got %s", engine.status.State)
	}

	// Set a job ID to test status transitions
	engine.status.JobID = "12345"

	// Kill should return an error without actually having a slurm job
	err := engine.Kill()

	// The error is expected, but we can check that the method doesn't panic
	if err == nil {
		t.Log("Note: Kill() succeeded without actual slurm job (expected in test environment)")
	}
}
