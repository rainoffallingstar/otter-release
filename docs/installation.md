# xdxtools Installation Guide

This guide provides step-by-step instructions for building and installing xdxtools, including its dependencies: enva (environment manager) and rv (R package manager).

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Detailed Installation](#detailed-installation)
  - [Step 1: Build and Install Binaries](#step-1-build-and-install-binaries)
  - [Step 2: Initialize Runtime Directory](#step-2-initialize-runtime-directory)
  - [Step 3: Create Conda Environments](#step-3-create-conda-environments)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)
- [Platform-Specific Notes](#platform-specific-notes)

## Overview

xdxtools consists of two main components:

| Component | Language | Purpose |
|-----------|----------|---------|
| **xdxtools** | Go 1.21+ | Main CLI for bioinformatics workflows |
| **enva** | Rust 1.92+ | Lightweight micromamba environment manager |

### Git Submodules

- **enva**: Environment manager for creating and managing conda environments
- **rv**: R package manager for reproducible package installation

## Prerequisites

### Build Dependencies

1. **Go 1.21 or later**
   ```bash
   go version  # Should be go1.21+
   ```

2. **Rust 1.92 or later** (for enva and rv)
   ```bash
   rustc --version  # Should be 1.92+
   ```

   **Note**: Rust 1.92+ is required for rv which uses Rust edition 2024.

3. **Git** (for submodule management)
   ```bash
   git --version
   ```

### Runtime Dependencies

These will be installed automatically by enva:

- **Conda/Miniconda** or **Micromamba**
- **Python 3.8+**
- **R 4.4+**

## Quick Start

### Automated Installation

Use the provided setup script for automated installation:

```bash
# Clone with submodules
git clone --recurse-submodules https://github.com/yourusername/xdxtools.git
cd xdxtools

# Run setup script
bash scripts/setup.sh
```

The setup script will:
1. Build xdxtools, enva, and rv
2. Install binaries to `~/.cargo/bin`
3. Initialize runtime directory at `~/xdxtools-runtime`
4. Create all conda environments
5. Install R packages

### Options

```bash
# Skip specific steps
bash scripts/setup.sh --skip-build          # Skip building binaries
bash scripts/setup.sh --skip-init           # Skip runtime directory init
bash scripts/setup.sh --skip-envs           # Skip conda environment creation
bash scripts/setup.sh --skip-r-packages     # Skip R package installation

# Dry run (show what would be done)
bash scripts/setup.sh --dry-run

# Show help
bash scripts/setup.sh --help
```

## Detailed Installation

### Step 1: Build and Install Binaries

#### 1.1 Build xdxtools

```bash
cd /path/to/xdxtools

# Option 1: Simple build
go build -o xdxtools .

# Option 2: Static build (recommended for portability)
CGO_ENABLED=0 go build -ldflags="-s -w" -o xdxtools .

# Install to PATH
mkdir -p ~/.cargo/bin
cp xdxtools ~/.cargo/bin/
export PATH="$PATH:$HOME/.cargo/bin"
```

#### 1.2 Initialize and Build rv Submodule

```bash
# Initialize submodule
git submodule update --init --recursive rv

# Build rv (requires Rust 1.92+)
cd rv
cargo build --release --features=cli

# Verify build
./target/release/rv --version

# Install to PATH
cp target/release/rv ~/.cargo/bin/
```

**Note**: If your system Rust version is < 1.92, use a conda environment with Rust 1.92+:

```bash
# Create conda environment with Rust 1.92
conda create -n rustenv rust=1.92 -y
conda activate rustenv

# Build rv
cargo build --release --features=cli
```

#### 1.3 Build enva

```bash
cd /path/to/xdxtools/enva

# Build
cargo build --release

# Verify build
./target/release/enva --version

# Install to PATH
cp target/release/enva ~/.cargo/bin/
```

#### 1.4 Verify Installation

```bash
# Check binaries are in PATH
which xdxtools enva rv

# Check versions
xdxtools --version
enva --version
rv --version
```

### Step 2: Initialize Runtime Directory

```bash
# Create runtime directory
mkdir -p ~/xdxtools-runtime
cd ~/xdxtools-runtime

# Initialize with xdxtools
xdxtools init .

# Verify structure
ls -la
# Expected output:
# config/  data/  envs/  R/  rules/  snakefiles/  workflows/
```

**What gets created:**

- `config/` - Configuration directory
- `data/` - Gene database files (human/mouse genomes)
- `envs/` - Conda environment YAML files
- `R/` - R scripts directory
- `rules/` - Snakemake rules
- `snakefiles/` - Main workflow files
- `workflows/` - Workflow configurations

### Step 3: Create Conda Environments

```bash
cd ~/xdxtools-runtime

# Create all 3 environments
enva create --all

# Expected output:
# ✓ Creating xdxtools-core environment...
# ✓ Creating xdxtools-snakemake environment...
# ✓ Creating xdxtools-extra environment...
```

**Or create individually:**

```bash
enva create --core          # Core bioinformatics tools (includes qualimap)
enva create --snakemake     # Snakemake workflow engine
enva create --extra         # Additional visualization tools
```

**Note**: The R environment (xdxtools-r) is no longer required - gomats replaces RNA_Splicing.R and htseq2matrix-go replaces htseq2matrix.R.

**Verify environments:**

```bash
# List all environments
enva list

# Validate environments
enva validate --all

# Detailed information
enva list --detailed | grep xdxtools
```
```

**Note**: Step 4 (Install R Packages) is no longer required. The following tools have been replaced:
- `gomats` replaces RNA_Splicing.R
- `htseq2matrix-go` replaces htseq2matrix.R

## Verification

### Test Binary Installation

```bash
# Test xdxtools
xdxtools --version
xdxtools init --help

# Test enva
enva --version
enva list

# Test rv
rv --version
rv --help
```

### Test Runtime Directory

```bash
cd ~/xdxtools-runtime

# Check structure
ls -la config/ data/ envs/ R/ rules/

# Test environment
enva list --detailed

# Test gomats
gomats --help
```

### Test Workflow

```bash
cd ~/xdxtools-runtime

# Create a test project (requires FASTQ files)
xdxtools create --fastq /path/to/fastq --pdata samples.csv

# Verify config
cat config/config.yaml
```

## Troubleshooting

### enva Not Found During Setup

**Problem**: `enva: command not found`

**Solution**: Add `~/.cargo/bin` to PATH:

```bash
export PATH="$PATH:$HOME/.cargo/bin"

# Add to ~/.bashrc or ~/.zshrc for persistence
echo 'export PATH="$PATH:$HOME/.cargo/bin"' >> ~/.bashrc
source ~/.bashrc
```

### Conda Environment Already Exists

**Problem**: `enva create` fails because environment already exists.

**Solution**: This is expected if you've already created the environment. enva will skip existing environments:

```bash
# Check what environments exist
enva list

# Create only missing environments
enva create --core    # Will skip if exists
```

### Go Build Fails

**Problem**: `go: module ...: no matching versions`

**Solution**: Update Go modules:

```bash
cd /path/to/xdxtools
go mod tidy
go mod download
go build -o xdxtools .
```

### Git Submodule Not Initialized

**Problem**: rv directory is empty

**Solution**: Initialize submodule:

```bash
git submodule update --init --recursive rv
```

## Platform-Specific Notes

### Linux

All components are fully supported on Linux. No additional steps required.

### macOS

All components are supported on macOS (Intel and Apple Silicon).

**Build for Apple Silicon:**

```bash
# Go build
CGO_ENABLED=0 go build -ldflags="-s -w" -o xdxtools .

# Rust build (cargo will detect architecture automatically)
cargo build --release
```

### Windows

xdxtools, enva, and rv support Windows via WSL (Windows Subsystem for Linux).

**Recommended: Use WSL2**

1. Install WSL2
2. Install Ubuntu or other Linux distribution
3. Follow Linux installation instructions

**Native Windows (not recommended):**

Requires:
- Git Bash or similar Unix-like environment
- Miniconda or Anaconda
- Go for Windows
- Rust for Windows (MSVC toolchain)

## Alternative Installation Methods

### Using Pre-built Binaries

If you don't want to build from source:

1. **xdxtools**: Download from [GitHub Releases](https://github.com/yourusername/xdxtools/releases)
2. **enva**: Download from [enva releases](https://github.com/rainoffallingstar/enva/releases/latest)
3. **rv**: Use the install script:
   ```bash
   curl -sSL https://raw.githubusercontent.com/A2-ai/rv/main/scripts/install.sh | bash
   ```

### Using Standard Conda (without enva)

If you prefer standard conda over enva:

```bash
# Create environments manually (3 environments, no R environment needed)
conda env create -f ~/xdxtools-runtime/envs/xdxtools-core.yaml
conda env create -f ~/xdxtools-runtime/envs/xdxtools-snakemake.yaml
conda env create -f ~/xdxtools-runtime/envs/xdxtools-extra.yaml
```

## Uninstallation

To remove xdxtools completely:

```bash
# Remove binaries
rm ~/.cargo/bin/xdxtools
rm ~/.cargo/bin/enva

# Remove runtime directory
rm -rf ~/xdxtools-runtime

# Remove conda environments (optional)
enva remove --all
# Or: conda env remove -n xdxtools-core
#      conda env remove -n xdxtools-snakemake
#      conda env remove -n xdxtools-extra
```

## Next Steps

After installation:

1. **Read the User Guide**: See README.md for usage examples
2. **Create Your First Analysis**:
   ```bash
   cd ~/xdxtools-runtime
   xdxtools create --fastq /path/to/fastq --mode RRBS --pdata samples.csv
   ```
3. **Run a Workflow**:
   ```bash
   xdxtools run --config config/config.yaml --engine slurm
   ```

## Getting Help

- **Documentation**: See README.md and docs/ directory
- **Issues**: Report bugs on [GitHub Issues](https://github.com/yourusername/xdxtools/issues)
- **Discussions**: Ask questions on [GitHub Discussions](https://github.com/yourusername/xdxtools/discussions)
