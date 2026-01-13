package engine

import (
	"fmt"
	"os"

	"github.com/xdxtools/xdxtools-go/internal/config"
)

// EngineFactory creates engines based on type and configuration
type EngineFactory struct{}

// NewEngine creates an engine based on the specified type and configuration
func (f *EngineFactory) NewEngine(engineType EngineType, cfg *config.EngineConfig) (Engine, error) {
	switch engineType {
	case EngineSlurm:
		slurmConfig := &SlurmConfig{
			Partition:  cfg.Slurm.Partition,
			Cores:      cfg.Slurm.Cores,
			Memory:     cfg.Slurm.Memory,
			JobName:    cfg.Slurm.JobName,
			MaxRetries: cfg.Slurm.MaxRetries,
		}
		return NewSlurmEngine(slurmConfig), nil

	case EngineLocal:
		localConfig := &LocalConfig{
			MaxCores:  cfg.Local.MaxCores,
			MaxMemory: cfg.Local.MaxMemory,
		}
		return NewLocalEngine(localConfig), nil

	case EngineSlurmArray:
		slurmConfig := &SlurmConfig{
			Partition:  cfg.Slurm.Partition,
			Cores:      cfg.Slurm.Cores,
			Memory:     cfg.Slurm.Memory,
			JobName:    cfg.Slurm.JobName,
			MaxRetries: cfg.Slurm.MaxRetries,
		}
		// For array engine, samples and stepResource need to be set later
		return NewSlurmArrayEngine(slurmConfig, nil, nil), nil

	default:
		return nil, fmt.Errorf("unsupported engine type: %s", engineType)
	}
}

// NewSlurmArrayEngineWithResources creates a SlurmArrayEngine with samples and step resources
func (f *EngineFactory) NewSlurmArrayEngineWithResources(
	cfg *config.EngineConfig,
	samples []string,
	stepResource *config.StepResource,
) (*SlurmArrayEngine, error) {
	slurmConfig := &SlurmConfig{
		Partition:  cfg.Slurm.Partition,
		Cores:      cfg.Slurm.Cores,
		Memory:     cfg.Slurm.Memory,
		JobName:    cfg.Slurm.JobName,
		MaxRetries: cfg.Slurm.MaxRetries,
	}

	return NewSlurmArrayEngine(slurmConfig, samples, stepResource), nil
}

// DetectEngine automatically detects the execution environment
func DetectEngine() EngineType {
	// Check for Slurm environment variables
	if os.Getenv("SLURM_JOB_ID") != "" || os.Getenv("SLURM_CLUSTER_NAME") != "" {
		return EngineSlurm
	}

	// Default to local
	return EngineLocal
}

// CreateEngineFromConfig creates an engine from configuration
// It handles auto-detection if engine type is "auto"
func CreateEngineFromConfig(cfg *config.XDXToolsConfig) (Engine, error) {
	factory := &EngineFactory{}

	engineType := EngineType(cfg.Engine.Type)

	// Auto-detect if specified
	if engineType == "auto" {
		engineType = DetectEngine()
	}

	return factory.NewEngine(engineType, &cfg.Engine)
}
