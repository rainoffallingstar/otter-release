# xdxtools

A bioinformatics workflow CLI for RRBS, WGBS, RNA-seq, and PDX analysis.

## Features

- Workflow modes: RRBS / WGBS / RNA-seq / PDX
- Auto FASTQ pairing + per-sample adapter generation (barcode support)
- SLURM Job Array + local worker pool parallelization
- Excel/CSV pdata with Chinese column name auto-mapping
- Snakemake integration with embedded workflow files
- TUI interface for interactive management
- Conda environment auto-fallback (enva supported)
- Single binary, no runtime dependencies

## Requirements

- Go 1.21+
- Snakemake
- conda / mamba / micromamba (or [enva](https://github.com/rainoffallingstar/enva))

## Installation

### Quick install (interactive)

```bash
bash <(wget -qO- https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh)
```

The script downloads pre-built binaries from GitHub Releases and sets up the required conda environments interactively.

Common options:

```bash
# Non-interactive, use all defaults
bash <(wget -qO- https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh) --non-interactive

# Specify a release version
bash <(wget -qO- https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh) --version v0.3.0

# Skip conda environment creation
bash <(wget -qO- https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh) --skip-envs
```

### Build from source

```bash
git clone --recurse-submodules https://github.com/rainoffallingstar/xdxtools-go.git
cd xdxtools-go
go build -o xdxtools
```

## Quick Start

```bash
# 1. Initialize project (copies Snakemake workflow files)
xdxtools init my_project

# 2. Scan FASTQ, validate samples, generate config
xdxtools create --fastq /data/fastq --mode RRBS --pdata samples.csv

# 3. Execute workflow
xdxtools run --config userspace/my_project/config/config.yaml
```

## Commands

| Command | Description |
|---------|-------------|
| `init`   | Copy Snakemake workflow files to project directory |
| `create` | Scan FASTQ, validate samples, generate config.yaml |
| `run`    | Execute Snakemake workflow |
| `status` | Show workflow progress |
| `config` | Validate configuration file |

### Key `run` flags

| Flag | Default | Description |
|------|---------|-------------|
| `--engine` | `auto` | Execution engine: `slurm` / `local` / `auto` |
| `--slurm-partition` | `cpu` | SLURM partition for all steps |
| `--parallel-jobs` | `2` | Max concurrent jobs |
| `--dry-run` | `false` | Test configuration without executing |
| `--resume` / `-r` | `false` | Resume from last completed step |

Use `--verbose` for detailed output or `--dry-run` to troubleshoot without running.

## Workflow Modes

| Mode | Flag |
|------|------|
| RRBS (Reduced Representation BS-seq) | `--mode RRBS` |
| WGBS (Whole Genome BS-seq) | `--mode WGBS` |
| RNA-seq | `--mode RNASEQ` |
| PDX (Patient-Derived Xenograft) | `--mode RRBS --species1 human --species2 mouse` |

PDX mode is enabled automatically when both `--species1` and `--species2` are specified.

## License

MIT
