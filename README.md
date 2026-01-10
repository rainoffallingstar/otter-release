# xdxtools-go

A bioinformatics workflow management tool for RRBS, WGBS, RNA-seq, and PDX analysis, rewritten in Go from the original R package.

## Features

- **Multiple Workflow Modes**: RRBS, WGBS, RNA-seq, and PDX analysis
- **Per-Sample Adapter Generation**: Automatic adapter generation with barcode support and reverse complement
- **Intelligent Input Processing**: Automatic FASTQ pairing and pdata validation
- **Snakemake Integration**: Full compatibility with existing Snakemake workflows
- **Chinese Column Name Support**: Automatic mapping from Chinese to English column names
- **Group Levels Calculation**: Automatic calculation of unique groups from pdata
- **Multiple Execution Engines**: Slurm and local execution (Docker support removed in architecture simplification)
- **Excel Support**: Direct .xlsx/.xls file support for pdata (no conversion needed)
- **TUI Interface**: Interactive terminal user interface for workflow management
- **Dynamic Reference Configuration**: Automatic generation of reference genome paths based on species and mode
- **Comprehensive Testing**: 175+ tests with 100% pass rate

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/xdxtools/xdxtools-go.git
cd xdxtools-go

# Build the binary
go build -o xdxtools

# Install to PATH (optional)
sudo mv xdxtools /usr/local/bin/
```

### Requirements

- Go 1.21+
- Snakemake
- Conda/Miniconda (for workflow execution)
- R (optional, for R scripts)
- Python 3.8+ (optional, for Python scripts)

## Quick Start - Three Command Workflow

xdxtools follows a simple three-command workflow:

### 1️⃣ `init` - Install Snakemake Workflow Files

Copy Snakemake workflow files, R scripts, and rules to your project directory.

```bash
# Initialize a new project directory
xdxtools init my_project

# Initialize with specific engine type (root/rootless)
xdxtools init my_project --engine-type rootless
```

This copies:
- 22 Snakemake workflow files
- 15+ R/Python scripts
- 34 Snakemake rules

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
```

**Note**: Docker engine support has been removed in architecture simplification. Only Slurm and Local engines are now supported.

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
├── log/                     # Log files
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
- **local**: Execute on local machine

**Note**: Docker engine support has been removed in architecture simplification.

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

### Example 5: Custom Job ID and Output Path

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

**Q: How do I change the number of workers?**
A: Edit the generated config.yaml: parallel.workers: 8

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
