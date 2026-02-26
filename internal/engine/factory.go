package engine

import (
	"fmt"
	"os"

	"github.com/xdxtools/xdxtools-go/internal/config"
	"github.com/xdxtools/xdxtools-go/internal/logger"
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
	maxBatchSize int,
	loadRatio float64,
) (*SlurmArrayEngine, error) {
	// Use stepResource values if available, otherwise fallback to cfg.Slurm values
	cores := cfg.Slurm.Cores
	memory := cfg.Slurm.Memory

	logger.Debugf("NewSlurmArrayEngineWithResources: cfg.Slurm.Cores=%d, cfg.Slurm.Memory=%s", cores, memory)

	if stepResource != nil {
		logger.Debugf("stepResource.Cores=%d, stepResource.Memory=%s", stepResource.Cores, stepResource.Memory)
		if stepResource.Cores > 0 {
			cores = stepResource.Cores
		}
		if stepResource.Memory != "" {
			memory = stepResource.Memory
		}
	}

	logger.Debugf("Final cores=%d, memory=%s", cores, memory)

	slurmConfig := &SlurmConfig{
		Partition:  cfg.Slurm.Partition,
		Cores:      cores,
		Memory:     memory,
		JobName:    cfg.Slurm.JobName,
		MaxRetries: cfg.Slurm.MaxRetries,
	}

	eng := NewSlurmArrayEngine(slurmConfig, samples, stepResource)
	eng.maxBatchSize = maxBatchSize
	eng.loadRatio = loadRatio
	return eng, nil
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
