# xdxtools Installation Guide

This guide provides step-by-step instructions for building and installing xdxtools and its runtime helper `enva`.

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
| **xdxtools** | Go 1.24+ | Main CLI for bioinformatics workflows |
| **enva** | Rust 1.92+ | Rattler-first environment manager |

### Git Submodules

- **enva**: Environment manager for creating and managing conda environments
- Other analysis helpers live as independent submodules and can be built separately with `./scripts/build-all-submodules.sh` when needed

## Prerequisites

### Build Dependencies

1. **Go 1.24 or later**
   ```bash
   go version  # Should be go1.24+
   ```

2. **Rust 1.92 or later** (for `enva` and Rust-based submodules)
   ```bash
   rustc --version  # Should be 1.92+
   ```

3. **Git** (for submodule management)
   ```bash
   git --version
   ```

### Runtime Dependencies

These will be installed automatically by enva:

- **Optional compatibility package manager**: `conda`, `mamba`, or `micromamba` for adoption / fallback scenarios
- **Python 3.8+**
- **R 4.4+**

## Quick Start

### Automated Installation

Use the release installer when you want to download pre-built binaries from GitHub Releases. Use `scripts/setup.sh` only when you want to build from source locally:

```bash
# Run directly from GitHub
bash <(curl -fsSL https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh)

# Or clone first and run locally
git clone --recurse-submodules https://github.com/rainoffallingstar/xdxtools-go.git
cd xdxtools-go
bash scripts/install.sh
```

If the repository itself is private, anonymous `raw.githubusercontent.com` downloads return `404`, and `wget -qO-` hides that failure. Use an authenticated bootstrap command instead. If the GitHub Releases repository or assets are private, export `GITHUB_TOKEN` (or `GH_TOKEN` / `GITHUB_PAT`) before running the installer. For a private fork, also set `GITHUB_RELEASES_REPO=<owner>/<repo>` or pass `--releases-repo <owner>/<repo>`. The first interactive prompt lets you choose English or Chinese; you can also force the interface language with `--lang en` or `--lang zh`. If GitHub access fails and no token is configured, the installer can prompt for a hidden token input and retry once for the current session.

The installer will:
1. Download release binaries into your install directory
2. Add the install directory to `PATH`
3. Optionally create the required conda environments
4. Print the next `xdxtools init/create/run` steps

### Options

```bash
# Skip conda environment creation
bash scripts/install.sh --skip-envs

# Skip HDF5 setup for methrix-cli
bash scripts/install.sh --skip-hdf5

# Pin a specific release tag
bash scripts/install.sh --version v0.3.0

# Force Chinese installer output
bash scripts/install.sh --lang zh

# Private repository bootstrap
export GITHUB_PAT=<your_pat>
bash <(curl -fsSL -H "Authorization: Bearer ${GITHUB_PAT}" \
  https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh)

# Private GitHub Releases
export GITHUB_PAT=<your_pat>
export GITHUB_RELEASES_REPO=<owner>/<repo>

# Or override per invocation
bash scripts/install.sh --releases-repo <owner>/<repo> --dry-run

# Dry run (show what would be done)
bash scripts/install.sh --dry-run

# Show help
bash scripts/install.sh --help
```

If you want to build from source instead of downloading release binaries, use `scripts/setup.sh` or follow the manual build steps below.

## Release Tag Model

`xdxtools` supports two release-tag patterns:

- Immutable release tags: `vX.Y.Z` or `vYYYY.MM.DD.N`
- Movable daily aliases: `daily-YYYYMMDD`

Recommended usage:

- Use `vYYYY.MM.DD.N` for each archived build you need to reproduce later.
- Use `daily-YYYYMMDD` as the same-day shared tag that can be updated to the latest build for that date.

Examples:

```bash
# Immutable date-based release
bash scripts/release.sh 2026.03.23.1

# Later the same day: new immutable build + update daily alias
bash scripts/release.sh 2026.03.23.2

# Immutable date-based release without moving the daily alias
bash scripts/release.sh 2026.03.23.3 --no-daily-alias
```

## Detailed Installation

### Step 1: Build and Install Binaries

#### 1.1 Build xdxtools

```bash
cd /path/to/xdxtools-go

# Option 1: Simple build
go build -o xdxtools .

# Option 2: Static build (recommended for portability)
CGO_ENABLED=0 go build -ldflags="-s -w" -o xdxtools .

# Install to PATH
mkdir -p ~/.cargo/bin
cp xdxtools ~/.cargo/bin/
export PATH="$PATH:$HOME/.cargo/bin"
```

#### 1.2 Build enva

```bash
cd /path/to/xdxtools-go/enva

# Build
cargo build --release

# Verify build
./target/release/enva --version

# Install to PATH
cp target/release/enva ~/.cargo/bin/
```

#### 1.3 Verify Installation

```bash
# Check binaries are in PATH
which xdxtools enva

# Check versions
xdxtools --version
enva --version
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
# Expected output includes:
# config/  data/  envs/  inst/  R/  rules/  userspace/
```

**What gets created:**

- `config/` - Project-level configuration directory
- `data/` - Input/reference data directory
- `envs/` - Conda environment YAML files
- `inst/` - Embedded workflow/runtime resources
- `R/` - R/Python helper scripts
- `rules/` - Snakemake rules
- `userspace/` - Per-run working directories created by `xdxtools create`

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

**Note**: Step 4 (Install R Packages) is no longer required. The following tools have been replaced:
- `gomats` replaces RNA_Splicing.R
- `htseq2matrix-go` replaces htseq2matrix.R

## Verification

### Test Binary Installation

```bash
# Test xdxtools
xdxtools --version
xdxtools init --help
xdxtools create --help

# Test enva
enva --version
enva list
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
xdxtools init my_project
xdxtools create --fastq /path/to/fastq --pdata samples.csv --output my_project/userspace --jobid demo_run

# Verify config
cat my_project/userspace/demo_run/config/config.yaml
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
cd /path/to/xdxtools-go
go mod tidy
go mod download
go build -o xdxtools .
```

### Git Submodules Not Initialized

**Problem**: one or more submodule directories are empty

**Solution**: initialize all required submodules:

```bash
git submodule update --init --recursive
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

xdxtools and enva support Windows via WSL (Windows Subsystem for Linux).

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

1. **xdxtools**: Download from [GitHub Releases](https://github.com/rainoffallingstar/xdxtools-go/releases)
2. **enva**: Download from [enva releases](https://github.com/rainoffallingstar/enva/releases/latest)

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
   xdxtools init my_project
   xdxtools create --fastq /path/to/fastq --mode RRBS --pdata samples.csv --output my_project/userspace --jobid demo_run
   ```
3. **Run a Workflow**:
   ```bash
   xdxtools run --config my_project/userspace/demo_run/config/config.yaml --engine slurm
   ```

## Getting Help

- **Documentation**: See README.md and docs/ directory
- **Issues**: Report bugs on [GitHub Issues](https://github.com/rainoffallingstar/xdxtools-go/issues)
