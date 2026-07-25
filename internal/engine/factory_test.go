package engine

import (
	"testing"

	"github.com/rainoffallingstar/otter/internal/config"
	"github.com/rainoffallingstar/otter/internal/logger"
)

func init() {
	logger.Init(false)
}

func TestNewEngine_Slurm(t *testing.T) {
	factory := &EngineFactory{}

	engineCfg := &config.EngineConfig{
		Type: "slurm",
		Slurm: config.SlurmConfig{
			Partition:  "compute",
			Cores:      16,
			Memory:     "32G",
			JobName:    "test_job",
			MaxRetries: 3,
		},
	}

	engine, err := factory.NewEngine(EngineSlurm, engineCfg)

	if err != nil {
		t.Fatalf("NewEngine() returned unexpected error: %v", err)
	}

	if engine == nil {
		t.Fatal("NewEngine() returned nil")
	}

	if engine.GetName() != EngineSlurm {
		t.Errorf("Expected engine name to be SLURM, got %s", engine.GetName())
	}
}

func TestNewEngine_Local(t *testing.T) {
	factory := &EngineFactory{}

	engineCfg := &config.EngineConfig{
		Type: "local",
		Local: config.LocalConfig{
			MaxCores:  8,
			MaxMemory: "16G",
		},
	}

	engine, err := factory.NewEngine(EngineLocal, engineCfg)

	if err != nil {
		t.Fatalf("NewEngine() returned unexpected error: %v", err)
	}

	if engine == nil {
		t.Fatal("NewEngine() returned nil")
	}

	if engine.GetName() != EngineLocal {
		t.Errorf("Expected engine name to be LOCAL, got %s", engine.GetName())
	}
}

func TestNewEngine_UnsupportedType(t *testing.T) {
	factory := &EngineFactory{}

	engineCfg := &config.EngineConfig{
		Type: "unsupported",
	}

	engine, err := factory.NewEngine(EngineType("unsupported"), engineCfg)

	if err == nil {
		t.Error("NewEngine() should have returned an error for unsupported engine type")
	}

	if engine != nil {
		t.Error("NewEngine() should have returned nil for unsupported engine type")
	}

	expectedErrMsg := "unsupported engine type: unsupported"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestDetectEngine(t *testing.T) {
	// This test checks the default behavior
	// In a real environment, it would detect SLURM or Local

	engineType := DetectEngine()

	// Just verify it returns one of the valid types
	validTypes := []EngineType{EngineSlurm, EngineLocal}

	valid := false
	for _, vt := range validTypes {
		if engineType == vt {
			valid = true
			break
		}
	}

	if !valid {
		t.Errorf("DetectEngine() returned invalid engine type: %s", engineType)
	}
}
