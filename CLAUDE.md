# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working in this repository.

## Project Overview

xdxtools-go is a bioinformatics workflow management tool for RRBS, WGBS, RNA-seq, and PDX analysis, rewritten in Go from the original R package. It provides a CLI and TUI interface for managing Snakemake-based bioinformatics workflows.

## Quick Start

```bash
# Build the binary
go build -o xdxtools

# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -v ./internal/input -run TestAdapterGenerator
```

## Build and Development Commands

```bash
# Build for current platform
go build -o xdxtools

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o xdxtools-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o xdxtools-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o xdxtools.exe

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests verbosely
go test -v ./...

# Run specific package tests
go test ./internal/input/...
go test ./internal/config/...
go test ./internal/engine/...
go test ./cmd/...

# Run integration tests only
go test ./cmd/... -v

# Test with timeout
go test -timeout 30s ./...
```

## Three-Command Workflow

xdxtools follows a simple three-command workflow:

### 1. `init` - Install Snakemake Workflow Files
```bash
xdxtools init my_project
xdxtools init my_project --engine-type rootless
```

### 2. `create` - Create Analysis Project
```bash
xdxtools create --fastq /data/fastq --mode RRBS
xdxtools create --fastq /data/fastq --pdata samples.csv --mode RRBS --species1 human --species2 mouse
```

### 3. `run` - Execute Workflow
```bash
xdxtools run --config config.yaml
xdxtools run --config config.yaml --engine slurm
```

## High-Level Architecture

### CLI Commands (cmd/)

The CLI is built using Cobra framework with 5 main commands:

1. **root** - Base command with global flags
2. **init** - Initialize project structure and copy embedded assets
3. **create** - Scan FASTQ, validate samples, create directory structure, generate config
4. **run** - Execute Snakemake workflows
5. **config** - Configuration management (validate only)
6. **tui** - Interactive Terminal User Interface

### Internal Modules (internal/)

#### 1. **assets** - Embedded Resource Management
- Uses Go 1.16+ `embed` package to embed Snakemake files, R scripts, rules
- Copies embedded assets to project directory during `init`
- Supports both root and rootless engine types

#### 2. **config** - Configuration Management
- Loads and validates YAML configuration files
- Generates Snakemake-compatible configs
- Supports workflow mode mapping (RRBS/WGBS/RNASEQ/PDX)
- Key types: `WorkflowConfig`, `ReferenceConfig`, `ParallelConfig`

#### 3. **engine** - Execution Engine Abstraction
- Three implementations: SlurmEngine, LocalEngine, DockerEngine
- Factory pattern: `CreateEngineFromConfig()`
- Auto-detection of execution environment
- Key interface: `Engine` with `Execute(cmd []string) error`

#### 4. **workflow** - Workflow Orchestration
- Manages workflow lifecycle: Initialize → ExecuteAll
- Creates directory structures (30+ subdirectories)
- Integrates with Snakemake execution
- PDX mode support with species-specific directories

#### 5. **input** - Input Processing
- **FASTQ Scanning**: Scans directories, pairs R1/R2 files
- **PData Parsing**: Excel (.xlsx/.xls) and CSV support
- **Adapter Generation**: Per-sample adapters with barcode support
- **Validation**: Input validation and error reporting
- Key components: Scanner, PDataParser, AdapterGenerator, Validator

#### 6. **script** - External Script Execution
- Executes R/Python scripts via Conda environments
- Integrates with execution engines (Slurm/Local/Docker)
- Encapsulates `conda run` invocations

#### 7. **tui** - Terminal User Interface
- Simple command-line based TUI (not Bubble Tea)
- Interactive menu system
- Workflow dashboard and status monitoring
- Configuration management interface

#### 8. **logger** - Logging
- Built on logrus
- Configurable verbosity
- Structured logging

### Data Flow

```
CLI Command (cmd/)
    ↓
Internal Module (internal/*/)
    ↓
Engine Interface (internal/engine/)
    ↓
Execution Backend (Slurm/Local/Docker)
    ↓
Snakemake Workflow
    ↓
Bioinformatics Tools (Bismark, FastQC, etc.)
```

## Key Implementation Details

### Workflow Mode Mapping
- RRBS/WGBS/BSSEQ → "BeaverBS" (3 steps)
- RNASEQ → "BeaverRNA" (2 steps)
- PDX + RRBS/WGBS → "BeaverPDX" (3 steps)
- PDX + RNASEQ → "BeaverRNASEQPDX" (3 steps)

### PDX Mode Detection
Automatically enabled when both `species1` and `species2` are specified:
- Creates species-specific subdirectories
- Switches workflow to PDX variants
- Supports graft (species1) and host (species2) species

### Per-Sample Adapter Generation
- Reads barcodes from `inline_barcode_sequence` column
- Computes reverse complement
- RRBS mode: adds "TGA" (R1) / "A" (R2) prefix
- WGBS mode: no prefix
- Generates "NO_ADAPTER_CAL_USE_DEFAULT" for empty barcodes

### Group Levels Calculation
- Counts unique values in `sample_group` or `condition` column
- Priority: `sample_group` > `condition`
- Used for downstream differential analysis

## Testing Strategy

- **175+ test cases** across all modules
- **Unit tests**: Each internal module has comprehensive tests
- **Integration tests**: cmd/cli_integration_test.go tests full CLI workflow
- **TUI tests**: internal/tui/tui_test.go
- **Coverage**: ~69.3% (approaching 80% target)

Test file organization:
```
*_test.go files in each package
cmd/cli_integration_test.go - Full CLI workflow tests
internal/*/*_test.go - Module-specific unit tests
```

## Test Data Organization

Test data is centralized in `testdata/`:
- `fastq/` - FASTQ test files (3 samples, 6 files)
- `pdata/` - Phenotype data files (CSV/Excel)
- `configs/` - Configuration examples
- `projects/` - Sample generated projects (13 different scenarios)
- `e2e/` - End-to-end test scenarios

## Project Structure

```
xdxtools-go/
├── cmd/                          # CLI commands
│   ├── root.go                   # Root command
│   ├── init.go                   # Init command
│   ├── create.go                # Create command
│   ├── run.go                    # Run command
│   ├── config.go                 # Config command
│   ├── tui.go                    # TUI command
│   └── cli_integration_test.go   # Integration tests
│
├── internal/                      # Core modules
│   ├── assets/                   # Embedded resources
│   ├── config/                   # Configuration
│   ├── engine/                   # Execution engines
│   ├── workflow/                 # Workflow management
│   ├── input/                    # Input processing
│   ├── script/                    # Script execution
│   ├── tui/                      # Terminal UI
│   └── logger/                   # Logging
│
├── pkg/                           # Public utilities
├── testdata/                     # Test data (centralized)
│   ├── fastq/
│   ├── pdata/
│   ├── configs/
│   ├── projects/
│   └── e2e/
│
├── docs/                          # Documentation
│   ├── active_context.md         # System state
│   ├── architecture.md           # Architecture
│   └── requirements.md            # Requirements
│
└── embed.go                       # Embedded resources
```

## Common Development Tasks

### Adding a New CLI Command
1. Create new file in `cmd/`
2. Register in `cmd/root.go`
3. Add tests in `cmd/` directory
4. Update this CLAUDE.md if adding significant functionality

### Adding a New Module
1. Create directory in `internal/`
2. Define types and interfaces
3. Write unit tests with `*_test.go`
4. Integrate with appropriate CLI command
5. Update `docs/active_context.md`

### Modifying Configuration
- Configuration types in `internal/config/config.go`
- Default values in `internal/config/defaults.go`
- Generation logic in `internal/config/generator.go`

### Testing Changes
```bash
# Run all tests
go test ./...

# Run tests for specific module
go test ./internal/input/... -v

# Run with coverage
go test -cover ./...

# Run integration tests
go test ./cmd/... -v

# Run specific test
go test -v ./internal/input -run TestAdapterGenerator
```

## Key Conventions

- **Go version**: Requires Go 1.21+
- **Error handling**: Use `error` returns, avoid `log.Fatal` in libraries
- **Logging**: Use `internal/logger` package
- **Configuration**: YAML format, struct tags for parsing
- **FASTQ naming**: `_R1.fastq.gz` and `_R2.fastq.gz` convention
- **PDX mode**: Automatically enabled when `species1` and `species2` both specified
- **Project output**: Stored in `userspace/{jobid}/` directory structure

## Embedded Resources

The project uses Go's `embed` package to bundle:
- 22 Snakemake workflow files
- 15+ R/Python scripts
- 34 Snakemake rules
- 5 workflow configuration templates

These are copied to project directory during `xdxtools init`.

## Dependencies

- **github.com/spf13/cobra** - CLI framework
- **github.com/sirupsen/logrus** - Logging
- **gopkg.in/yaml.v3** - YAML parsing
- **github.com/360EntSecGroup-Skylar/excelize** - Excel file support

See `go.mod` for complete dependency list.

## Environment Detection

The tool auto-detects execution environment:
1. Check `SLURM_JOB_ID` → Slurm
2. Check Docker availability → Docker
3. Default → Local

## Troubleshooting

### Tests Failing
```bash
# Clean test cache
go clean -testcache

# Run with verbose output
go test -v ./...

# Check for race conditions
go test -race ./...
```

### Build Issues
```bash
# Clean build cache
go clean -cache

# Rebuild dependencies
go mod tidy
go mod download
```

## Documentation

- **README.md** - User guide (English)
- **README_zh.md** - User guide (Chinese)
- **docs/active_context.md** - System state and implementation status
- **docs/architecture.md** - Detailed architecture
- **docs/requirements.md** - Requirements and specs
- **testdata/README.md** - Test data documentation

## Notes

- This is a production-ready codebase with 175+ tests
- Supports RRBS, WGBS, RNA-seq, and PDX workflows
- Integrates with Snakemake for workflow execution
- Provides both CLI and TUI interfaces
- Excel (.xlsx/.xls) support for pdata files
- Chinese column name mapping support
