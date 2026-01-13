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

## Three-Command Workflow

```bash
# 1. Initialize project (copies Snakemake files, R scripts, rules)
xdxtools init my_project

# 2. Create analysis (scans FASTQ, validates samples, generates config)
xdxtools create --fastq /data/fastq --mode RRBS --pdata samples.csv

# 3. Execute workflow
xdxtools run --config userspace/my_project/config/config.yaml --engine slurm
```

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
- Empty barcode → "NO_ADAPTER_CAL_USE_DEFAULT"

### Group Levels
- Counts unique values in `sample_group` or `condition` column
- Priority: `sample_group` > `condition`

## Embedded Resources

Located in `inst/`:
- `Rscripts/` - R/Python scripts for Snakemake rules
- `root_rules/` & `rootless_rules/` - Snakemake rule files
- `snakefiles/` - Main Snakemake workflow files
- `envs/` - Conda environment definitions

Copied to project during `xdxtools init`.

## Test Data

Centralized in `testdata/`:
- `fastq/` - Sample FASTQ files
- `pdata/` - Phenotype data (CSV/Excel)
- `configs/` - Example configurations
- `e2e/` - End-to-end test fixtures

## Conventions

- Go 1.21+
- YAML configuration with struct tags
- FASTQ naming: `*_R1.fastq.gz` / `*_R2.fastq.gz`
- Project output: `userspace/{jobid}/`
- Chinese column names auto-mapped to English

## Key Configuration Types

```go
// internal/config/config.go
type WorkflowConfig struct {
    Mode      string   // RRBS, WGBS, RNASEQ
    Species1  string   // Primary species
    Species2  string   // Secondary species (enables PDX)
    SIDs      []string // Sample IDs
    TrimSeq1  []string // Per-sample R1 adapters
    TrimSeq2  []string // Per-sample R2 adapters
}
```

## Engine Interface

```go
// internal/engine/types.go
type Engine interface {
    Execute(cmd []string) error
    GetType() string
}
```
