# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

xdxtools is a bioinformatics workflow management CLI for RRBS, WGBS, RNA-seq, and PDX analysis, rewritten in Go from an R package. It orchestrates Snakemake-based workflows with support for SLURM job arrays and local parallel execution.

## Build & Test Commands

```bash
# Build
go build -o xdxtools

# Static build (for distribution)
CGO_ENABLED=0 go build -ldflags="-s -w" -o target/release/xdxtools-linux-amd64

# Run all tests
go test ./...

# Run specific test
go test -v ./internal/input -run TestAdapterGenerator

# Run with coverage
go test -cover ./...
```

## Git Submodules

This project uses two git submodules:
- **enva**: Lightweight micromamba environment manager (2-5x faster than conda)
- **rv**: Fast, reproducible R package manager

```bash
# Clone with submodules (recommended)
git clone --recurse-submodules https://github.com/xdxtools/xdxtools-go.git

# Or initialize after cloning
git submodule update --init --recursive
```

## Three-Command Workflow

```bash
# 1. Initialize project (copies Snakemake files, R scripts, rules)
xdxtools init my_project

# 2. Create analysis (scans FASTQ, validates samples, generates config)
xdxtools create --fastq /data/fastq --mode RRBS --pdata samples.csv

# 3. Execute workflow
xdxtools run --config userspace/my_project/config/config.yaml --engine slurm
```

## CLI Commands

| Command | Purpose |
|---------|---------|
| `init` | Install Snakemake workflow files to project directory |
| `create` | Scan FASTQ, validate samples, generate config.yaml |
| `run` | Execute Snakemake workflow |
| `status` | Display workflow status and progress |
| `config` | Validate configuration files |
| `tui` | Interactive terminal UI |

### Key Run Flags

- `--engine`: Execution engine (auto/slurm/local)
- `--parallel-jobs`: Unified parallelization control (1 = sequential, >1 = parallel)
- `--dry-run`: Test configuration without executing
- `--resume/-r`: Resume from last completed step
- `--slurm-partition`: SLURM partition for all steps

## Architecture

### CLI Layer (cmd/)
Cobra-based commands: `init`, `create`, `run`, `config`, `tui`

### Core Modules (internal/)

| Module | Purpose | Key Files |
|--------|---------|-----------|
| **config** | YAML config loading, validation, generation | `config.go`, `generator.go`, `defaults.go` |
| **engine** | Execution backends (Slurm, SlurmArray, Local) | `slurm.go`, `slurm_array.go`, `local.go`, `factory.go` |
| **input** | FASTQ scanning, pdata parsing, adapter generation | `fastq.go`, `pdata.go`, `adapter.go` |
| **workflow** | Directory structure, Snakemake integration | `manager.go`, `snakemake.go` |
| **assets** | Embedded resources (Snakemake files, R scripts) | Uses Go `embed` package |

### Engine Types

Three execution engines are available (defined in `internal/engine/types.go`):
- `EngineSlurm` - Submit jobs to SLURM cluster
- `EngineSlurmArray` - Use SLURM Job Array for parallel sample processing
- `EngineLocal` - Run locally with worker pool

Use `engine.CreateEngineFromConfig(cfg)` to create the appropriate engine.

### Data Flow
```
CLI Command → Config Loading → Input Processing → Engine Selection → Snakemake Execution
```

### Engine Factory Pattern
```go
engine.CreateEngineFromConfig(cfg) // Returns SlurmEngine, SlurmArrayEngine, or LocalEngine
```

## Key Implementation Details

### Workflow Mode Mapping
- RRBS/WGBS/BSSEQ → "BeaverBS" (3 steps)
- RNASEQ → "BeaverRNA" (2 steps)
- PDX mode → "BeaverPDX" / "BeaverRNASEQPDX" (auto-enabled when species2 is set)

### Parallelization Strategy
- **< 5 samples**: Sequential execution
- **>= 5 samples**: SLURM Job Array or local worker pool
- Steps 2 & 3 use single-sample mode (`--config "SIDs=[sample]"`)
- Step 1 & checkers use all-samples mode

### Adapter Generation
- Reads `inline_barcode_sequence` from pdata
- Computes reverse complement
- RRBS: adds "TGA" (R1) / "A" (R2) prefix
- WGBS: no prefix added
- Empty barcode → "NO_ADAPTER_CAL_USE_DEFAULT"

### Group Levels
- Counts unique values in `sample_group` or `condition` column
- Priority: `sample_group` > `condition`

### Chinese Column Name Support
Automatically maps Chinese columns to English:
- "样本编号" / "样本ID" → `sampleid`
- "条件" → `condition`
- "样本分组" / "分组" → `sample_group`

## Embedded Resources

Located in `inst/`:
- `Rscripts/` - R/Python scripts for Snakemake rules
- `root_rules/` & `rootless_rules/` - Snakemake rule files
- `snakefiles/` - Main Snakemake workflow files
- `envs/` - Conda environment definitions

Copied to project during `xdxtools init`.

## Key Interfaces

### Engine Interface
```go
// internal/engine/engine.go
type Engine interface {
    Execute(cmd []string) error
    ExecuteWithOutput(cmd []string) (string, error)
    GetName() EngineType
    GetStatus() *Status
    Wait() error
    Kill() error
    SetLogDir(dir string) error
}
```

### Status States
```go
const (
    StatusPending   = "PENDING"
    StatusRunning   = "RUNNING"
    StatusCompleted = "COMPLETED"
    StatusFailed    = "FAILED"
    StatusKilled    = "KILLED"
)
```

### Key Configuration Types
```go
// internal/config/config.go
type XDXToolsConfig struct {
    Workflow      WorkflowConfig
    Input         InputConfig
    Output        OutputConfig
    Reference     ReferenceConfig
    Engine        EngineConfig
}

type WorkflowConfig struct {
    Mode      string   // RRBS, WGBS, RNASEQ
    Species   SpeciesConfig
    Adapters  AdapterConfig
    Samples   []SampleConfig
}
```

## Test Data

Centralized in `testdata/`:
- `fastq/` - Sample FASTQ files
- `pdata/` - Phenotype data (CSV/Excel)
- `configs/` - Example configurations
- `e2e/` - End-to-end test fixtures

## Conventions

- Go 1.21+
- YAML configuration with struct tags
- FASTQ naming: `*_R1.fastq.gz` / `*_R2.fastq.gz` (or `*_1.fastq.gz` / `*_2.fastq.gz`)
- Project output: `userspace/{jobid}/`
- Job ID: 40-character hexadecimal string (auto-generated)
- Excel (.xlsx/.xls) pdata files supported natively
