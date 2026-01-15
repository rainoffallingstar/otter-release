# xdxtools-go

A bioinformatics workflow management tool for RRBS, WGBS, RNA-seq, and PDX analysis, rewritten in Go from the original R package.

## Features

- **Multiple Workflow Modes**: RRBS, WGBS, RNA-seq, and PDX analysis
- **Per-Sample Adapter Generation**: Automatic adapter generation with barcode support and reverse complement
- **Intelligent Input Processing**: Automatic FASTQ pairing and pdata validation
- **Snakemake Integration**: Full compatibility with existing Snakemake workflows
- **SLURM Job Array Parallelization**: Efficient multi-sample parallel processing with automatic task distribution
- **Unified Parallelization Control**: Single `--parallel-jobs` parameter controls both local and SLURM execution
- **Local Parallel Execution**: Worker pool pattern for local multi-sample parallelization
- **Step-Based Execution Modes**: Steps 2&3 use single-sample mode, Step 1&checkers use all-samples mode
- **Resource Inheritance**: Checker steps automatically inherit resources from main steps
- **Standard Snakemake Parameters**: Uses `--config SIDs=[sample]` for single-sample override (no non-standard parameters)
- **Chinese Column Name Support**: Automatic mapping from Chinese to English column names
- **Group Levels Calculation**: Automatic calculation of unique groups from pdata
- **Multiple Execution Engines**: Slurm, SlurmArray (Job Array), and local execution (Docker support removed in architecture simplification)
- **Excel Support**: Direct .xlsx/.xls file support for pdata (no conversion needed)
- **TUI Interface**: Interactive terminal user interface for workflow management
- **Dynamic Reference Configuration**: Automatic generation of reference genome paths based on species and mode
- **Comprehensive Testing**: 175+ tests with 100% pass rate

## Git Submodules

This project uses git submodules to integrate two external tools for enhanced functionality:

### Submodules Overview

| Submodule | Purpose | URL |
|-----------|---------|-----|
| **enva** | Lightweight micromamba environment manager with auto-detection of conda/mamba/micromamba | [rainoffallingstar/enva](https://github.com/rainoffallingstar/enva) |
| **rv** | Fast, reproducible R package manager with conda environment support | [rainoffallingstar/rv](https://github.com/rainoffallingstar/rv) |

### Why Submodules?

- **enva**: Provides 2-5x faster environment activation compared to standard conda, with automatic detection of the fastest available package manager
- **rv**: Manages R dependencies reproducibly with support for conda environments, automatic dependency discovery from R scripts, and fast binary package installation

### Initializing Submodules

When cloning this repository, use one of these methods:

```bash
# Method 1: Clone with submodules (recommended)
git clone --recurse-submodules https://github.com/xdxtools/xdxtools-go.git
cd xdxtools-go

# Method 2: Clone and init separately
git clone https://github.com/xdxtools/xdxtools-go.git
cd xdxtools-go
git submodule init
git submodule update

# Method 3: If you already cloned without --recurse-submodules
git submodule update --init --recursive
```

### Updating Submodules

To update submodules to their latest versions:

```bash
# Update all submodules
git submodule update --remote --merge

# Update specific submodule
git submodule update --remote --merge enva
git submodule update --remote --merge rv

# After updating, rebuild the project
go build -o xdxtools
```

### Checking Submodule Status

```bash
# Check submodule status
git submodule status

# View submodule commits
git submodule summary
```

## Installation

### From Source

```bash
# Clone the repository with submodules
git clone --recurse-submodules https://github.com/xdxtools/xdxtools-go.git
cd xdxtools-go

# Alternatively, clone and init submodules separately
git clone https://github.com/xdxtools/xdxtools-go.git
cd xdxtools-go
git submodule init
git submodule update

# Build the binary
go build -o xdxtools

# Install to PATH (optional)
sudo mv xdxtools /usr/local/bin/
```

**Note**: This repository uses `enva` and `rv` as git submodules. If you cloned without `--recurse-submodules`, run:
```bash
git submodule update --init --recursive
```

See the [Git Submodules](#git-submodules) section above for more details.

### Requirements

- Go 1.21+
- Snakemake
- **enva** (recommended, auto-detects conda/mamba/micromamba)
- Conda/Miniconda (fallback if enva not available)
- R (optional, for R scripts)
- Python 3.8+ (optional, for Python scripts)

## enva Integration - Enhanced Package Management

### What is enva?

**enva** is a lightweight environment manager that automatically detects and uses the fastest available package manager (conda → mamba → micromamba) for optimal performance (2-5x faster than standard conda).

### Why enva?

| Feature | conda | mamba | micromamba | **enva (auto)** |
|---------|-------|-------|------------|-----------------|
| Startup Time | 2-3s | 0.8-1s | 0.5-0.7s | **0.5-3s*** |
| Environment Activation | 1-2s | 0.3-0.5s | 0.2-0.4s | **0.2-2s*** |
| Relative Performance | 1x (baseline) | 2-3x faster | 3-5x faster | **2-5x faster*** |

*enva performance depends on the fastest available PM detected

### Installation

```bash
# Download enva (Linux x86_64)
wget https://github.com/xdxtools/enva/releases/latest/download/enva-linux-x86_64
chmod +x enva-linux-x86_64
sudo mv enva-linux-x86_64 /usr/local/bin/enva

# Verify installation
enva --version
# enva v0.1.0
```

### Usage in Snakemake Rules

All xdxtools Snakemake rules now use `enva run` with `--` separator:

```python
# Standard format (recommended)
rule fastqc:
  shell:
    """
    enva run fastqc -- fastqc -o {params.dir} -t {threads} --extract {input.R1}
    """

# Multi-line commands
rule multiqc:
  shell:
    """
    enva run multiqc -- multiqc {params.readir} \
      -o {params.outdir} \
      -f
    """
```

**Key Features**:
- ✅ Automatic PM detection (conda → mamba → micromamba)
- ✅ Clean syntax: `enva run <env> -- <command>`
- ✅ Backslash newline support
- ✅ Backward compatible with `conda run -n`

### Environment Override

```bash
# Force specific package manager
ENVA_PACKAGE_MANAGER=mamba enva run fastqc -- fastqc --version
```

### Automatic Detection

When you run `xdxtools init`, enva availability is automatically checked:

```
Checking package manager support...
✓ enva detected - will use fastest available package manager
```

If enva is not found, you'll see installation guidance:
```
⚠️ enva not found in PATH
For best performance (2-5x faster), install enva:
  wget https://github.com/xdxtools/enva/releases/latest/download/enva-linux-x86_64
  ...
Falling back to conda run (slower)
```

### Performance Benefits

With enva + mamba/micromamba:
- **Workflow startup**: 2-5x faster
- **Environment activation**: 2-5x faster
- **Command execution**: Significantly reduced overhead

For a typical RRBS workflow with 10 samples:
- **Without enva**: ~8-12 minutes overhead
- **With enva + mamba**: ~2-3 minutes overhead
- **Time saved**: 5-10 minutes per run!

### Migration from conda run

Old format:
```python
shell:
    """
    conda run -n fastqc fastqc -o {params.dir} -t {threads}
    """
```

New format:
```python
shell:
    """
    enva run fastqc -- fastqc -o {params.dir} -t {threads}
    """
```

**All 67 .smk files have been updated automatically!**

## Snakemake Environment Automatic Fallback

### What is Automatic Fallback?

When your specified conda environment is invalid or doesn't exist, xdxtools automatically falls back to the `xdxtools-snakemake` environment to ensure your workflow can run.

### How It Works

1. **Validation**: Before running Snakemake, the tool validates the specified environment
2. **Fallback**: If validation fails, it automatically retries with `xdxtools-snakemake`
3. **Clear Logging**: All validation and fallback actions are logged

### Configuration

Add to your `config.yaml`:

```yaml
engine:
  type: auto
  conda_env: my-custom-env      # Primary environment
  fallback_env: xdxtools-snakemake  # Fallback (optional, default: xdxtools-snakemake)
  no_fallback: false            # Disable fallback (optional, default: false)
```

### Log Output Example

```
Validating conda environment: my-custom-env
Environment 'my-custom-env' validation failed: ...
Falling back to 'xdxtools-snakemake' environment...
Environment validation successful: xdxtools-snakemake
```

### Command Line Override

```bash
# Specify conda environment (will fallback if needed)
xdxtools run --config config.yaml --conda-env my-custom-env

# Disable fallback (fail immediately if environment invalid)
# Edit config.yaml: no_fallback: true
```

### Benefits

- ✅ **Resilient**: Automatically handles environment issues
- ✅ **Transparent**: Clear logging shows what's happening
- ✅ **Configurable**: Can be disabled if needed
- ✅ **Safe**: Only falls back if primary environment fails

## Quick Start - Three Command Workflow

xdxtools follows a simple three-command workflow:

### 1️⃣ `init` - Install Snakemake Workflow Files

Copy Snakemake workflow files, R scripts, and rules to your project directory.

```bash
# Initialize a new project directory
xdxtools init my_project

# Initialize with specific engine type (root/rootless)
xdxtools init my_project --engine-type rootless

# Initialize in current directory
xdxtools init .
```

This copies:
- 22 Snakemake workflow files
- 15+ R/Python scripts
- 27 Snakemake rules (active)
- 7 archived rules (in .depress/rules/)
- 3 gene annotation database files (.rda format for RNA-seq analysis)

### 2️⃣ `create` - Create Analysis Project

Scan FASTQ files, validate samples, create directory structure, and generate config.yaml.

```bash
# Basic usage - automatically generates jobid
xdxtools create --fastq /data/fastq --mode RRBS

# Complete example with all options
xdxtools create \
    --fastq /data/fastq \
    --pdata /data/pdata/samples.csv \
    --mode RRBS \
    --species1 human \
    --species2 mouse \
    --output userspace \
    --jobid my_project_2024
```

**What it does:**
1. Scans FASTQ directory for paired samples (R1/R2)
2. Validates sample pairing
3. Loads and validates pdata (if provided)
4. Generates unique job ID (or uses user-specified)
5. Creates complete directory structure in `userspace/{jobid}/`
6. Generates `config.yaml` with per-sample adapters
7. Calculates group levels from pdata

### 3️⃣ `run` - Execute Workflow

Execute the Snakemake workflow using the generated config.

```bash
# Run the workflow
xdxtools run --config userspace/my_project/config/config.yaml

# Run with specific engine
xdxtools run --config userspace/my_project/config/config.yaml --engine slurm

# Dry run (test configuration)
xdxtools run --config userspace/my_project/config/config.yaml --dry-run
```

## Command Reference

### Commands

#### init

Install Snakemake workflow files to a project directory.

```bash
xdxtools init <project-directory> [flags]
```

**Flags:**
- `--engine-type`: Engine type (root/rootless, default: rootless)

**Examples:**

```bash
# Initialize current directory
xdxtools init .

# Initialize new directory
xdxtools init my_project

# Initialize with rootless rules (non-container)
xdxtools init my_project --engine-type rootless
```

#### create

Create a new analysis project with validated samples and generated configuration.

```bash
xdxtools create --fastq <path> [flags]
```

**Required Flags:**
- `--fastq, -f`: FASTQ files directory

**Optional Flags:**
- `--pdata, -p`: Phenotype data file (Excel/CSV)
- `--mode, -m`: Workflow mode (RRBS/WGBS/RNASEQ, default: RRBS)
- `--species1`: Primary species (default: human)
- `--species2`: Secondary species (enables PDX mode when specified)
- `--output, -o`: Output directory for projects (default: userspace)
- `--jobid`: Custom job ID (default: auto-generated 40-char hex)
- `--suffix1`: R1 file suffix (default: _R1.fastq.gz)
- `--suffix2`: R2 file suffix (auto-derived if empty)
- `--genome1-fasta`: Primary species genome FASTA file
- `--genome1-index`: Primary species genome index directory
- `--genome2-fasta`: Secondary species genome FASTA file (PDX mode)
- `--genome2-index`: Secondary species genome index directory (PDX mode)
- `--gtf1`: Primary species GTF annotation file (for RNA-seq)
- `--gtf2`: Secondary species GTF annotation file (for PDX RNA-seq)
- `--star-index1`: Primary species STAR index directory (for RNA-seq)
- `--star-index2`: Secondary species STAR index directory (for PDX RNA-seq)

**Examples:**

```bash
# Basic RRBS project
xdxtools create --fastq /data/fastq --mode RRBS

# WGBS with pdata
xdxtools create --fastq /data/fastq --pdata /data/pdata.csv --mode WGBS

# PDX mode (automatic when species2 is specified)
xdxtools create \
    --fastq /data/fastq \
    --mode RRBS \
    --species1 human \
    --species2 mouse

# Custom output and jobid
xdxtools create \
    --fastq /data/fastq \
    --output /custom/path \
    --jobid experiment_001

# Custom reference genome files
xdxtools create \
    --fastq /data/fastq \
    --mode RRBS \
    --genome1-fasta inst/hg38/hg38.fasta \
    --genome1-index inst/hg38/

# RNA-seq with custom GTF and STAR index
xdxtools create \
    --fastq /data/fastq \
    --mode RNASEQ \
    --species1 human \
    --gtf1 inst/rnaseq/hg38/hg38.ensGene_sorted.gtf \
    --star-index1 inst/rnaseq/hg38/

# PDX mode with custom reference files for both species
xdxtools create \
    --fastq /data/fastq \
    --mode RRBS \
    --species1 human \
    --species2 mouse \
    --genome1-fasta inst/hg38/hg38.fasta \
    --genome1-index inst/hg38/ \
    --genome2-fasta inst/mm10/mm10.fasta \
    --genome2-index inst/mm10/
```

#### run

Execute a Snakemake workflow using a configuration file.

```bash
xdxtools run --config <config-file> [flags]
```

**Required Flags:**
- `--config, -c`: Configuration file path

**Optional Flags:**
- `--engine`: Execution engine (auto/slurm/local, default: auto)
- `--dry-run`: Perform a dry run without executing
- `--verbose, -v`: Verbose output
- `--resume, -r`: Resume from last completed step
- `--conda-env`: Conda environment for Snakemake
- `--slurm-partition`: Unified SLURM partition for all steps (overrides config and individual step partitions)
- `--slurm-unified-partition`: Legacy unified partition parameter (kept for backward compatibility)
- `--slurm-cores`: Default SLURM CPU cores for all steps
- `--slurm-memory`: Default SLURM memory for all steps (e.g., 16G)
- `--step1-cores`: Step 1 CPU cores
- `--step1-memory`: Step 1 memory (e.g., 8G)
- `--step1-partition`: Step 1 partition
- `--step2-cores`: Step 2 CPU cores
- `--step2-memory`: Step 2 memory (e.g., 32G)
- `--step2-partition`: Step 2 partition
- `--step3-cores`: Step 3 CPU cores
- `--step3-memory`: Step 3 memory (e.g., 16G)
- `--step3-partition`: Step 3 partition
- `--parallel-jobs`: Max parallel jobs for local/Snakemake execution (default: 2)

**Examples:**

```bash
# Run with auto-detected engine (auto-detects Slurm or Local)
xdxtools run --config config/config.yaml

# Run on Slurm cluster
xdxtools run --config config/config.yaml --engine slurm

# Run locally
xdxtools run --config config/config.yaml --engine local

# Test configuration
xdxtools run --config config/config.yaml --dry-run

# Run with unified partition for all steps
xdxtools run --config config/config.yaml --slurm-partition cpu --engine slurm

# Run with custom resources for specific steps
xdxtools run --config config/config.yaml \
  --step1-cores 20 --step1-memory 100G \
  --step2-cores 40 --step2-memory 200G \
  --step3-cores 10 --step3-memory 300G \
  --engine slurm
```

**Note**: Docker engine support has been removed in architecture simplification. Only Slurm and Local engines are now supported.

#### status

Display the status of a running or completed workflow.

```bash
xdxtools status [project-dir]
```

**Arguments:**
- `project-dir`: Project directory to check (default: current directory)

**Examples:**

```bash
# Check status of current directory
xdxtools status

# Check status of specific project
xdxtools status userspace/my_project
```

The status command displays:
- Job ID and workflow status
- Start time, last update, and duration
- Configuration summary (mode, species, samples, engine)
- Step-by-step progress with completion times
- SLURM job IDs for running/completed steps
- Overall progress statistics

**Resume Workflow:**

If a workflow is interrupted, you can resume from the last completed step:

```bash
xdxtools run --config config.yaml --resume
# or
xdxtools run --config config.yaml -r
```

#### config

Validate and manage configuration files.

```bash
xdxtools config [command]
```

**Commands:**
- `validate`: Validate a configuration file without running the workflow

**Examples:**

```bash
# Validate a configuration file
xdxtools config validate --config userspace/my_project/config/config.yaml
```

#### tui

Launch the interactive Terminal User Interface (TUI) for xdxtools.

```bash
xdxtools tui [flags]
```

**Optional Flags:**
- `--config, -c`: Optional configuration file to load on startup

**Examples:**

```bash
# Launch TUI with interactive menu
xdxtools tui

# Launch TUI with pre-loaded configuration
xdxtools tui --config userspace/my_project/config/config.yaml
```

The TUI provides:
- Interactive menu for creating new projects
- Workflow dashboard and status monitoring
- Configuration management interface
- Help and documentation

## Workflow Modes

### RRBS (Reduced Representation Bisulfite Sequencing)
```bash
xdxtools create --fastq /data/fastq --mode RRBS
```

### WGBS (Whole Genome Bisulfite Sequencing)
```bash
xdxtools create --fastq /data/fastq --mode WGBS
```

### RNA-seq
```bash
xdxtools create --fastq /data/fastq --mode RNASEQ
```

### PDX (Patient-Derived Xenograft)
```bash
xdxtools create \
    --fastq /data/fastq \
    --mode RRBS \
    --species1 human \
    --species2 mouse
```

PDX mode is automatically enabled when both `species1` and `species2` are specified. The workflow switches to "BeaverPDX" configuration and creates species-specific subdirectories.

## Default Resource Configuration

### Step Resources

The tool uses optimized default resource configurations for each workflow step:

#### RRBS / WGBS / RNASEQ Modes

| Step | CPU Cores | Memory | Partition | Threads | JobArray |
|------|-----------|--------|-----------|---------|----------|
| Step 1 | 20 | 100GB | cpu | 10 | No |
| Step 2 | 40 | 200GB | cpu | 20 | Yes |
| Step 3 | 10 | 300GB | cpu | 5 | Yes |
| Step 2 Checker | 40 | 200GB | cpu | 20 | - |
| Step 3 Checker | 10 | 300GB | cpu | 5 | - |

**Note**: Checker steps automatically inherit resources from their corresponding main steps.

#### PDX Mode

PDX mode uses the same resource configuration as non-PDX mode (no multiplier applied).

### Customizing Resources

Override default resources using command-line flags:

```bash
# Set unified partition for all steps
xdxtools run --config config.yaml \
  --slurm-unified-partition cpu \
  --engine slurm

# Override specific step resources
xdxtools run --config config.yaml \
  --step1-cores 20 --step1-memory 100G \
  --step2-cores 40 --step2-memory 200G \
  --step3-cores 10 --step3-memory 300G \
  --engine slurm

# Priority: specific step partition > unified partition > default partition
xdxtools run --config config.yaml \
  --slurm-unified-partition cpu \
  --step2-partition gpu \
  --engine slurm
# Result: Step1=cpu, Step2=gpu (override), Step3=cpu
```

## Configuration

### Generated Config.yaml

The `create` command generates a complete `config.yaml` file with the following key features:

#### Per-Sample Adapters
```yaml
# Automatically generated per-sample adapters
trimSeq1:
  - "NO_ADAPTER_CAL_USE_DEFAULT"  # For samples without barcode
  - "TGACGATAGATCGGAAGAGC"        # For samples with barcode (RRBS mode)
trimSeq2:
  - "NO_ADAPTER_CAL_USE_DEFAULT"
  - "ACGATAGATCGGAAGAGC"

# Sample identifiers
SIDs:
  - "sample1"
  - "sample2"
```

#### Group Levels
Automatically calculated from pdata:
```yaml
group_levels: 2  # Number of unique groups (control, treatment)
```

#### User and Job Information
```yaml
userid: "auto_generated_or_custom"
jobid: "a1b2c3d4e5f6..."
```

#### Dynamic Reference Configuration
Automatically generated based on species and mode:
```yaml
# Single species RRBS
reference:
  genome: human
  genome_fasta: [inst/pdx/homo_sapiens/human.fasta]
  genome_index: [inst/pdx/homo_sapiens/]

# PDX RRBS
reference:
  genome: human
  genome_fasta:
    - inst/pdx/homo_sapiens/human.fasta
    - inst/pdx/mouse/mouse.fasta
  genome_index:
    - inst/pdx/homo_sapiens/
    - inst/pdx/mouse/

# Single species RNA-seq
reference:
  genome: human
  genome_fasta: [inst/pdx/homo_sapiens/human.fasta]
  rnaseq_gtf: inst/rnaseq/homo_sapiens/human.ensGene_sorted.gtf
  rnaseq_ref: inst/rnaseq/homo_sapiens/

# PDX RNA-seq (arrays for multiple species)
reference:
  genome: human
  genome_fasta:
    - inst/pdx/homo_sapiens/human.fasta
    - inst/pdx/mouse/mouse.fasta
  rnaseq_gtf:
    - inst/rnaseq/homo_sapiens/human.ensGene_sorted.gtf
    - inst/rnaseq/mouse/mouse.ensGene_sorted.gtf
  rnaseq_ref:
    - inst/rnaseq/homo_sapiens/
    - inst/rnaseq/mouse/
```

### FASTQ File Naming

The tool supports various FASTQ file naming conventions:

- Standard: `Sample_R1.fastq.gz` + `Sample_R2.fastq.gz`
- Alternative: `Sample_1.fastq.gz` + `Sample_2.fastq.gz`
- Custom suffixes supported

**Example:**
```bash
# Files detected:
Sample1_R1.fastq.gz
Sample1_R2.fastq.gz
Sample2_R1.fastq.gz
Sample2_R2.fastq.gz

# Command
xdxtools create --fastq /data/fastq --mode RRBS

# Result: 2 paired samples detected
```

### PData Format

#### CSV Format (Recommended)
```bash
# Save as CSV (Excel files not yet supported)
sampleid,inline_barcode_sequence,condition
sample1,ATCG,control
sample2,,treatment
sample3,GCTA,treatment
```

#### Chinese Column Names
Automatically mapped to English:
- "样本编号" / "样本ID" → `sampleid`
- "条件" → `condition`
- "样本分组" / "分组" → `sample_group`

#### Barcode Support
- Barcode column: `inline_barcode_sequence` or `barcode`
- Empty barcode → `NO_ADAPTER_CAL_USE_DEFAULT` (auto-detect adapter)
- With barcode → Generates adapter with reverse complement

### Per-Sample Adapter Generation

The tool automatically generates adapters for each sample:

#### RRBS Mode Example
```bash
# Input pdata
sample1: barcode = ATCG
sample2: barcode = (empty)

# Generated adapters (RRBS mode)
trimSeq1:
  - "TGACGATAGATCGGAAGAGC"  # TGA + CGAT(reverse of ATCG) + AGATCGGAAGAGC
  - "NO_ADAPTER_CAL_USE_DEFAULT"

trimSeq2:
  - "ACGATAGATCGGAAGAGC"    # A + CGAT(reverse of ATCG) + AGATCGGAAGAGC
  - "NO_ADAPTER_CAL_USE_DEFAULT"
```

#### WGBS Mode Example
```bash
# Same input
# Generated adapters (WGBS mode - no TGA/A prefix)
trimSeq1:
  - "CGATAGATCGGAAGAGC"     # CGAT(reverse of ATCG) + AGATCGGAAGAGC
  - "NO_ADAPTER_CAL_USE_DEFAULT"
```

### Project Directory Structure

After running `xdxtools create`, the following structure is created:

```
userspace/{jobid}/
├── config/
│   └── config.yaml          # Generated configuration
├── data/                    # FASTQ files (soft links)
├── logs/                    # Log files
│   ├── xdxtools.log         # Main xdxtools log (all levels)
│   ├── slurm.out            # SLURM stdout
│   ├── slurm.err            # SLURM stderr
│   └── snakemake/           # Snakemake logs
├── analysis/                # Analysis results
│   ├── betaM
│   ├── clubcpg/
│   │   ├── coverage_before/
│   │   ├── coverage_impute/
│   │   └── model/
│   ├── DMR
│   ├── GCbias
│   ├── logsummary
│   ├── methrixh5
│   ├── qc_summary
│   ├── RData
│   └── uxm_summary
└── workflow/                # Workflow outputs
    ├── QC
    ├── fastqc_raw
    ├── fastqc_clean
    ├── trim
    ├── bsmap/
    │   ├── tmp/
    │   │   ├── {species1}/  # PDX mode only
    │   │   └── {species2}/  # PDX mode only
    │   ├── {species1}/      # PDX mode only
    │   └── {species2}/      # PDX mode only
    ├── mCall
    ├── mhap
    ├── qualimap
    └── umx
```

### Workflow Modes

- **RRBS**: Reduced Representation Bisulfite Sequencing
- **WGBS**: Whole Genome Bisulfite Sequencing
- **RNASEQ**: RNA Sequencing
- **PDX**: Patient-Derived Xenograft (automatic when species2 is specified)

### Execution Engines

#### Auto-Detection

The tool automatically detects the execution environment:

1. **Slurm**: Checks for `SLURM_JOB_ID` environment variable
2. **Local**: Falls back to local execution

#### Engine Types

- **slurm**: Execute on Slurm cluster
- **slurm_array**: Execute on Slurm cluster with Job Array for parallel processing
- **local**: Execute on local machine
- **local_parallel**: Execute locally with parallel worker pool

**Note**: Docker engine support has been removed in architecture simplification.

### SLURM Job Array Parallelization

The tool provides efficient multi-sample parallel processing using SLURM Job Arrays.

#### Key Features

- **Unified Control**: Single `--parallel-jobs` parameter controls both local and SLURM execution
- **Single Submission**: Submit one job that automatically distributes N tasks
- **Full Resource Allocation**: Each task gets complete resource allocation (e.g., 16 cores per task)
- **Progress Tracking**: Real-time monitoring of job array progress
- **Configurable Concurrency**: Control maximum concurrent tasks with `%max` syntax

#### Execution Strategy

The tool uses unified parallelization control via `--parallel-jobs` parameter:

- **parallel-jobs = 1**: Sequential execution (one sample at a time)
- **parallel-jobs > 1**: Parallel execution (local worker pool or SLURM Job Array with N concurrent jobs)

#### Step-Based Modes

Different workflow steps use different execution modes:

- **Steps 2 & 3**: Single-sample mode (each sample runs independently)
  - Enables efficient parallel processing for compute-intensive steps
  - Uses `--config "SIDs=[sample]"` to override config for single-sample execution

- **Step 1 & Checkers**: All-samples mode (all samples processed together)
  - Setup and validation steps typically run fast enough sequentially
  - Ensures proper coordination across all samples

#### Example: SLURM Job Array Execution

For 10 samples with Step 2 configuration (40 cores, 200G memory) and `--parallel-jobs=5`:

```bash
# Generated SLURM script
#!/bin/bash
#SBATCH --job-name=xdxtools_step2_array
#SBATCH --partition=cpu
#SBATCH --cpus-per-task=40
#SBATCH --mem=200G
#SBATCH --array=0-9%5

SAMPLES[0]="sample1"
SAMPLES[1]="sample2"
...
SAMPLES[9]="sample10"

SAMPLE_NAME=${SAMPLES[$SLURM_ARRAY_TASK_ID]}

conda run -n snakemake snakemake --cores all \
  --snakefile BeaverBS_step2.snakemake \
  --config "SIDs=[$SAMPLE_NAME]"
```

**Result**: Single job submission, 10 tasks with max 5 concurrent jobs controlled by SLURM scheduler

#### Performance Characteristics

| Samples | Sequential (parallel-jobs=1) | Parallel (parallel-jobs=2) | Parallel (parallel-jobs=4) |
|---------|-------------------------------|----------------------------|----------------------------|
| 5      | 5x time                       | ~2.5x time                 | ~1.25x time               |
| 10     | 10x time                      | ~5x time                   | ~2.5x time                |
| 20     | 20x time                      | ~10x time                  | ~5x time                  |

*Actual speedup depends on hardware resources and cluster load*

#### Resource Management

- **Per-Task Resources**: Each array task receives full resource allocation
- **No Resource Sharing**: Resources are not averaged across array size
- **Predictable Performance**: Consistent performance per sample

#### Local Parallel Execution

For local environments, the tool uses a worker pool pattern controlled by `--parallel-jobs`:

```bash
# For 10 samples with local parallel execution (parallel-jobs=2)
snakemake --cores all --snakefile BeaverBS_step2.snakemake --config "SIDs=[sample1]"
snakemake --cores all --snakefile BeaverBS_step2.snakemake --config "SIDs=[sample2]"
...
# Executed in parallel via worker pool (max 2 concurrent jobs by default)
```

**Features**:
- Worker pool pattern with `--parallel-jobs` control
- Unified parameter for both local and SLURM execution
- Automatic resource management
- 24-hour timeout per task
- Success/failure aggregation reporting

### FASTQ File Naming

The tool supports various FASTQ file naming conventions:

- Standard: `Sample_R1.fastq.gz` + `Sample_R2.fastq.gz`
- Simplified: `Sample_1.fastq.gz` + `Sample_2.fastq.gz`
- Custom: Configure with `suffix1` and `suffix2` parameters

**Example:**

```yaml
suffix1: "_R1.fastq.gz"
suffix2: "_R2.fastq.gz"  # Auto-derived if not specified
```

### PDX Mode

PDX mode is automatically enabled when both `species1` and `species2` are specified:

```yaml
species1: "human"    # graft species
species2: "mouse"    # host species
```

This automatically switches to the PDX workflow configuration.

### Environment Variables

Configuration can be overridden with environment variables:

```bash
export XDXTOOLS_MODE="WGBS"
export XDXTOOLS_SPECIES1="human"
export XDXTOOLS_SPECIES2="mouse"
export XDXTOOLS_ENGINE="slurm"
export XDXTOOLS_FASTQ_DIR="/data/fastq"
export XDXTOOLS_PARALLEL_WORKERS="8"
```

Priority: Command line > Environment variables > Configuration file > Defaults

## Complete Examples

### Example 1: Basic RRBS Analysis

```bash
# Step 1: Initialize project
xdxtools init rrbs_project

# Step 2: Create analysis project
xdxtools create \
    --fastq /data/fastq \
    --mode RRBS \
    --output userspace \
    --jobid rrbs_experiment_001

# Step 3: Run workflow
xdxtools run --config userspace/rrbs_experiment_001/config/config.yaml
```

### Example 2: WGBS with PData

```bash
# Prepare pdata.csv
cat > /data/pdata.csv << EOF
sampleid,inline_barcode_sequence,condition
sample1,ATCG,control
sample2,GCTA,treatment
sample3,TAGC,control
sample4,CGAT,treatment
EOF

# Create project
xdxtools create \
    --fastq /data/fastq \
    --pdata /data/pdata.csv \
    --mode WGBS \
    --output userspace \
    --jobid wgbs_analysis

# Run
xdxtools run --config userspace/wgbs_analysis/config/config.yaml
```

### Example 3: PDX Mode (Human + Mouse)

```bash
# Create PDX project
xdxtools create \
    --fastq /data/fastq \
    --mode RRBS \
    --species1 human \
    --species2 mouse \
    --output userspace \
    --jobid pdx_human_mouse

# Run on Slurm
xdxtools run \
    --config userspace/pdx_human_mouse/config/config.yaml \
    --engine slurm
```

### Example 4: RNA-seq Analysis

```bash
# Create RNA-seq project
xdxtools create \
    --fastq /data/fastq \
    --mode RNASEQ \
    --output userspace \
    --jobid rnaseq_analysis

# Run with dry-run first
xdxtools run \
    --config userspace/rnaseq_analysis/config/config.yaml \
    --dry-run
```

### Example 5: Parallel Execution Control

```bash
# Sequential execution (1 job)
xdxtools run \
    --config userspace/my_project/config/config.yaml \
    --parallel-jobs 1

# Parallel execution with 2 jobs (default)
xdxtools run \
    --config userspace/my_project/config/config.yaml

# Parallel execution with 5 jobs
xdxtools run \
    --config userspace/my_project/config/config.yaml \
    --parallel-jobs 5

# Works for both SLURM and local
xdxtools run \
    --config userspace/my_project/config/config.yaml \
    --engine slurm \
    --parallel-jobs 3
```

### Example 6: Custom Job ID and Output Path

```bash
# Use custom jobid and output directory
xdxtools create \
    --fastq /data/fastq \
    --pdata /data/pdata.csv \
    --mode RRBS \
    --output /custom/output/path \
    --jobid experiment_20240109_001

# Verify project structure
ls -la /custom/output/path/experiment_20240109_001/
```

### Example 6: Chinese PData

```bash
# Prepare Chinese pdata.csv
cat > /data/pdata_chinese.csv << EOF
样本编号,inline_barcode_sequence,条件
样本1,ATCG,对照组
样本2,GCTA,治疗组
样本3,TAGC,对照组
样本4,CGAT,治疗组
EOF

# Create project (automatically maps Chinese columns)
xdxtools create \
    --fastq /data/fastq \
    --pdata /data/pdata_chinese.csv \
    --mode RRBS \
    --output userspace \
    --jobid chinese_pdata_test

# Verify group levels in config
cat userspace/chinese_pdata_test/config/config.yaml | grep group_levels
# Output: group_levels: 2
```

### Example 7: Barcode Processing

```bash
# Prepare pdata with barcodes
cat > /data/pdata_barcodes.csv << EOF
sampleid,inline_barcode_sequence,condition
Tumor1,ATCG,tumor
Tumor2,GCTA,tumor
Normal1,,normal
Normal2,,normal
EOF

# Create project
xdxtools create \
    --fastq /data/fastq \
    --pdata /data/pdata_barcodes.csv \
    --mode RRBS \
    --output userspace \
    --jobid barcode_test

# Check generated adapters
cat userspace/barcode_test/config/config.yaml | grep -A 10 "trimSeq1"
```

Expected output shows:
- Tumor1 with ATCG barcode: adapter = "TGACGATAGATCGGAAGAGC"
- Normal1 without barcode: adapter = "NO_ADAPTER_CAL_USE_DEFAULT"

## Troubleshooting

### Common Issues

#### 1. No FASTQ files found
```bash
# Error: No FASTQ files found in directory

# Solution: Check file naming
ls /data/fastq/

# Should show files like:
# Sample1_R1.fastq.gz
# Sample1_R2.fastq.gz
# Sample2_R1.fastq.gz
# Sample2_R2.fastq.gz
```

#### 2. No valid paired samples
```bash
# Error: No valid paired samples found

# Solution: Check suffix settings
xdxtools create \
    --fastq /data/fastq \
    --mode RRBS \
    --suffix1 "_1.fastq.gz" \
    --suffix2 "_2.fastq.gz"
```

#### 3. Sample not found in pdata
```bash
# Error: sample 'Sample1' not found in pdata

# Solution: Ensure sample names match
# FASTQ: Sample1_R1.fastq.gz
# pdata.csv: Sample1,...
```

#### 4. Excel file encoding issues
```bash
# Error: Excel file encoding or format issues

# Solution: Excel files (.xlsx/.xls) are now supported natively
# Just specify your Excel file directly:
xdxtools create --fastq /data/fastq --pdata /data/pdata.xlsx --mode RRBS

# If you encounter encoding issues, convert to CSV
# From Excel: Save As → CSV (UTF-8)
# Or use command line:
iconv -f GB2312 -t UTF-8 pdata.xlsx > pdata.csv
```

### Debug Mode

Enable verbose logging:
```bash
xdxtools create \
    --fastq /data/fastq \
    --mode RRBS \
    --verbose

xdxtools run \
    --config config.yaml \
    --verbose
```

### Validation

Validate configuration without running:
```bash
xdxtools run --config config.yaml --dry-run
```

Check config.yaml syntax:
```bash
# Install yq if needed: pip install yq
yq eval config.yaml

# Or use Python
python -c "import yaml; yaml.safe_load(open('config.yaml'))"
```

## Advanced Usage

### Custom File Suffixes

```bash
# For non-standard naming
xdxtools create \
    --fastq /data/fastq \
    --mode RRBS \
    --suffix1 "_R1.fastq" \
    --suffix2 "_R2.fastq"
```

### Batch Processing

```bash
# Process multiple projects
for mode in RRBS WGBS RNASEQ; do
    xdxtools create \
        --fastq /data/fastq \
        --mode $mode \
        --output userspace \
        --jobid batch_${mode}
done
```

### Engine Selection

```bash
# Auto-detect (recommended - detects Slurm or Local)
xdxtools run --config config.yaml

# Force specific engine
xdxtools run --config config.yaml --engine local
xdxtools run --config config.yaml --engine slurm
```

**Note**: Docker engine support has been removed. Only Slurm and Local engines are available.

## FAQ

**Q: What FASTQ file formats are supported?**
A: .fastq, .fastq.gz, .fq, .fq.gz

**Q: Does the tool support Excel pdata files?**
A: Yes! Excel files (.xlsx/.xls) are fully supported. Just specify the Excel file directly:
   `xdxtools create --fastq /data/fastq --pdata /data/pdata.xlsx --mode RRBS`

**Q: Can I use custom adapters?**
A: Yes, edit the generated config.yaml and modify trimSeq1/trimSeq2 arrays.

**Q: How is group_levels calculated?**
A: Counts unique values in sample_group or condition column (priority: sample_group > condition).

**Q: What is the job ID format?**
A: 40-character hexadecimal string (e.g., a1b2c3d4e5f6...)

**Q: How do I enable PDX mode?**
A: Specify both --species1 and --species2 flags.

**Q: What is NO_ADAPTER_CAL_USE_DEFAULT?**
A: Special marker telling Snakemake to auto-detect adapters.

**Q: Does the tool support dynamic reference genome paths?**
A: Yes! The tool automatically generates reference genome paths based on the specified species and mode. Single species returns strings, while PDX mode returns arrays for compatibility with Snakemake rules.

**Q: What species are supported for reference genomes?**
A: Currently supports: human/homo_sapiens and mouse/mus_musculus. Additional species can be added by extending the species mapping in the code.

**Q: How many samples are supported?**
A: No hard limit, depends on system resources and Snakemake configuration.

**Q: Can I resume a failed workflow?**
A: Yes, Snakemake automatically resumes from the last successful step.

**Q: How do I change the number of parallel jobs?**
A: Use the `--parallel-jobs` parameter when running the workflow: `xdxtools run --config config.yaml --parallel-jobs 4`

**Q: Does --parallel-jobs work for both SLURM and local execution?**
A: Yes! The `--parallel-jobs` parameter provides unified control for both SLURM and local execution. For SLURM, it controls the max concurrent Job Array tasks. For local execution, it controls the worker pool size.

**Q: Where are the log files stored?**
A: All logs are stored in the `logs/` directory within your project:
   - `xdxtools.log` - Main xdxtools log with all log levels
   - `slurm.out` - SLURM stdout
   - `slurm.err` - SLURM stderr
   - `snakemake/` - Snakemake-specific logs

**Q: How do I check the status of a running workflow?**
A: Use the `xdxtools status` command to view workflow progress, step completion status, and SLURM job IDs.

**Q: What happens if my workflow is interrupted?**
A: xdxtools automatically saves state. Use `--resume` or `-r` flag to continue from the last completed step. The tool will also warn you if it detects an incomplete workflow without the resume flag.

## Development

### Building

```bash
# Build for current platform
go build -o xdxtools

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o xdxtools-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o xdxtools-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o xdxtools.exe
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -v ./internal/input -run TestAdapterGenerator
```

### Project Structure

```
xdxtools-go/
├── cmd/                          # CLI commands
│   ├── root.go                   # Root command
│   ├── init.go                   # Init command
│   ├── create.go                # Create command
│   └── run.go                    # Run command
│
├── internal/                     # Internal packages
│   ├── assets/                  # Asset management
│   │   └── assets.go           # Resource copier
│   ├── config/                   # Configuration management
│   │   ├── config.go            # Configuration types
│   │   ├── loader.go            # Configuration loader
│   │   ├── validator.go         # Configuration validator
│   │   ├── generator.go         # Snakemake YAML generator
│   │   └── defaults.go          # Default configuration
│   │
│   ├── engine/                   # Execution engines
│   │   ├── engine.go            # Engine interface
│   │   ├── slurm.go             # Slurm engine
│   │   ├── local.go             # Local engine
│   │   └── factory.go           # Engine factory
│   │
│   ├── workflow/                 # Workflow management
│   │   ├── types.go             # Workflow types
│   │   ├── snakemake.go         # Snakemake executor
│   │   └── manager.go            # Workflow manager
│   │
│   ├── input/                    # Input processing
│   │   ├── types.go              # Input types
│   │   ├── fastq.go             # FASTQ scanner
│   │   ├── pdata.go             # pdata parser
│   │   ├── adapter.go           # Adapter generator
│   │   ├── adapter_test.go      # Adapter tests
│   │   └── validator.go         # Input validator
│   │
│   ├── script/                   # Script execution
│   │   └── executor.go           # Script executor
│   │
│   └── logger/                   # Logging
│       └── logger.go            # Logger
│
├── pkg/                          # Public packages
│   └── utils/                    # Utilities
│
├── docs/                         # Documentation
│   └── active_context.md        # System context
│
├── embed.go                      # Embedded resources
├── go.mod
└── main.go                       # Entry point
```

## License

This project is licensed under the MIT License.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgments

- Original xdxtools R package
- Snakemake workflow engine
- The bioinformatics community

## Support

For issues, questions, or contributions, please visit our GitHub repository.
