package engine

import (
	"testing"

	"github.com/xdxtools/xdxtools-go/internal/config"
)

func TestSlurmArrayEngine(t *testing.T) {
	// Create a SLURM config
	slurmConfig := &SlurmConfig{
		Partition:  "cpu112c",
		Cores:      16,
		Memory:     "32G",
		JobName:    "test_job",
		MaxRetries: 3,
	}

	// Create sample list
	samples := []string{"sample1", "sample2", "sample3", "sample4", "sample5"}

	// Create step resource
	stepResource := &config.StepResource{
		Cores:      16,
		Memory:     "32G",
		Partition:  "cpu112c",
		Threads:    8,
		JobArray:   true,
		MaxJobs:    5,
	}

	// Create SlurmArrayEngine
	arrayEngine := NewSlurmArrayEngine(slurmConfig, samples, stepResource)

	// Test array size
	if arrayEngine.GetArraySize() != 5 {
		t.Errorf("Expected array size 5, got %d", arrayEngine.GetArraySize())
	}

	// Test samples
	if len(arrayEngine.GetSamples()) != len(samples) {
		t.Errorf("Expected %d samples, got %d", len(samples), len(arrayEngine.GetSamples()))
	}

	// Test step resource
	if arrayEngine.stepResource.Cores != 16 {
		t.Errorf("Expected cores 16, got %d", arrayEngine.stepResource.Cores)
	}
	if arrayEngine.stepResource.Memory != "32G" {
		t.Errorf("Expected memory 32G, got %s", arrayEngine.stepResource.Memory)
	}
	if arrayEngine.stepResource.Partition != "cpu112c" {
		t.Errorf("Expected partition cpu112c, got %s", arrayEngine.stepResource.Partition)
	}
	if arrayEngine.maxArrayJobs != 5 {
		t.Errorf("Expected max jobs 5, got %d", arrayEngine.maxArrayJobs)
	}
}

func TestSlurmArrayEngineEmptySamples(t *testing.T) {
	slurmConfig := &SlurmConfig{
		Partition: "cpu112c",
		Cores:     4,
		Memory:    "8G",
	}

	// Create with empty samples
	arrayEngine := NewSlurmArrayEngine(slurmConfig, []string{}, &config.StepResource{})

	if arrayEngine.GetArraySize() != 0 {
		t.Errorf("Expected array size 0, got %d", arrayEngine.GetArraySize())
	}
	if len(arrayEngine.GetSamples()) != 0 {
		t.Errorf("Expected empty samples, got %d samples", len(arrayEngine.GetSamples()))
	}
}

func TestEngineFactorySlurmArray(t *testing.T) {
	factory := &EngineFactory{}

	// Test with resources
	samples := []string{"sample1", "sample2"}
	stepResource := &config.StepResource{
		Cores:     8,
		Memory:    "16G",
		Partition: "cpu",
	}

	slurmCfg := &SlurmConfig{
		Partition: "cpu",
		Cores:     8,
		Memory:    "16G",
		JobName:  "test",
	}

	arrayEngine, err := factory.NewSlurmArrayEngineWithResources(
		&config.EngineConfig{}, // Minimal config, not used in this test
		samples,
		stepResource,
		0,   // maxBatchSize=0: no batching limit
		0.0, // loadRatio=0: disable dynamic pool for test
	)

	// Override with our specific config for testing
	arrayEngine.partition = slurmCfg.Partition
	arrayEngine.cores = slurmCfg.Cores
	arrayEngine.memory = slurmCfg.Memory

	if err != nil {
		t.Errorf("Should not error creating array engine, got: %v", err)
	}
	if arrayEngine == nil {
		t.Error("Array engine should not be nil")
	}
	if arrayEngine.GetArraySize() != 2 {
		t.Errorf("Expected array size 2, got %d", arrayEngine.GetArraySize())
	}
	if len(arrayEngine.GetSamples()) != 2 {
		t.Errorf("Expected 2 samples, got %d", len(arrayEngine.GetSamples()))
	}
}

func TestLocalEngineSetMaxParallel(t *testing.T) {
	config := &LocalConfig{
		MaxCores:  8,
		MaxMemory: "16G",
	}

	engine := NewLocalEngine(config)
	if engine.maxParallel != 8 {
		t.Errorf("Expected max parallel 8, got %d", engine.maxParallel)
	}

	engine.SetMaxParallelJobs(16)
	if engine.maxParallel != 16 {
		t.Errorf("Expected max parallel 16, got %d", engine.maxParallel)
	}

	engine.SetMaxParallelJobs(0)
	if engine.maxParallel != 16 {
		t.Errorf("Expected max parallel to remain 16, got %d", engine.maxParallel)
	}
}

func TestLocalEngineDefaultMaxParallel(t *testing.T) {
	config := &LocalConfig{
		MaxCores:  0,
		MaxMemory: "8G",
	}

	engine := NewLocalEngine(config)
	if engine.maxParallel != 4 {
		t.Errorf("Expected max parallel to default to 4, got %d", engine.maxParallel)
	}
}

