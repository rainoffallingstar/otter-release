#!/bin/bash
set -e

INSTALL_DIR="$HOME/.cargo/bin"
RUNTIME_DIR="$HOME/xdxtools-runtime"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "======================================"
echo "xdxtools Build and Installation Script"
echo "======================================"
echo ""

# Parse command line arguments
SKIP_BUILD=false
SKIP_INIT=false
SKIP_ENVS=false
SKIP_R_PACKAGES=false
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --skip-init)
            SKIP_INIT=true
            shift
            ;;
        --skip-envs)
            SKIP_ENVS=true
            shift
            ;;
        --skip-r-packages)
            SKIP_R_PACKAGES=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --skip-build        Skip building binaries"
            echo "  --skip-init         Skip initializing runtime directory"
            echo "  --skip-envs         Skip creating conda environments"
            echo "  --skip-r-packages   Deprecated no-op (R packages are no longer installed)"
            echo "  --dry-run           Show what would be done without executing"
            echo "  --help              Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

if [ "$DRY_RUN" = true ]; then
    echo "DRY RUN MODE - No actual changes will be made"
    echo ""
fi

# ============================================================================
# Step 1: Build and Install Binaries
# ============================================================================
if [ "$SKIP_BUILD" = false ]; then
    echo "=== Step 1: Build and Install Binaries ==="

    if [ "$DRY_RUN" = false ]; then
        # Build xdxtools
        echo "Building xdxtools..."
        cd "$PROJECT_ROOT"
        CGO_ENABLED=0 go build -ldflags="-s -w" -o xdxtools .
        mkdir -p "$INSTALL_DIR"
        cp xdxtools "$INSTALL_DIR/"
        echo "✓ xdxtools installed to $INSTALL_DIR/"

        # Build enva
        if [ -f "$PROJECT_ROOT/enva/Cargo.toml" ]; then
            echo "Building enva..."
            cd "$PROJECT_ROOT/enva"
            cargo build --release
            cp target/release/enva "$INSTALL_DIR/"
            echo "✓ enva installed to $INSTALL_DIR/"
        fi
    else
        echo "[DRY RUN] Would build and install:"
        echo "  - xdxtools → $INSTALL_DIR/xdxtools"
        echo "  - enva → $INSTALL_DIR/enva"
    fi

    echo ""
else
    echo "=== Step 1: Skipped (--skip-build) ==="
    echo ""
fi

# ============================================================================
# Step 2: Initialize Runtime Directory
# ============================================================================
if [ "$SKIP_INIT" = false ]; then
    echo "=== Step 2: Initialize Runtime Directory ==="

    if [ "$DRY_RUN" = false ]; then
        mkdir -p "$RUNTIME_DIR"
        cd "$RUNTIME_DIR"

        # Check if already initialized
        if [ -f "config/config.yaml" ] || [ -d "rules" ]; then
            echo "⚠ Runtime directory already initialized"
            echo "  To re-initialize, remove $RUNTIME_DIR and run again"
        else
            "$PROJECT_ROOT/xdxtools" init .
            echo "✓ Runtime directory initialized at $RUNTIME_DIR/"
        fi
    else
        echo "[DRY RUN] Would initialize runtime directory at $RUNTIME_DIR/"
    fi

    echo ""
else
    echo "=== Step 2: Skipped (--skip-init) ==="
    echo ""
fi

# ============================================================================
# Step 3: Create Conda Environments
# ============================================================================
if [ "$SKIP_ENVS" = false ]; then
    echo "=== Step 3: Create Conda Environments ==="

    # Check if enva is available
    ENVA_PATH="$INSTALL_DIR/enva"
    if [ ! -f "$ENVA_PATH" ]; then
        ENVA_PATH="$PROJECT_ROOT/enva/target/release/enva"
    fi

    if [ ! -f "$ENVA_PATH" ]; then
        echo "⚠ enva not found, skipping conda environment creation"
        echo "  Build enva first or use --skip-envs"
    else
        if [ "$DRY_RUN" = false ]; then
            cd "$RUNTIME_DIR"

            # Create environments using the built enva binary
            "$ENVA_PATH" create --all
            echo "✓ Conda environments created"
        else
            echo "[DRY RUN] Would create conda environments using enva"
        fi
    fi

    echo ""
else
    echo "=== Step 3: Skipped (--skip-envs) ==="
    echo ""
fi

# ============================================================================
# Step 4: Install R Packages (已废弃 - 使用 Go 替代方案)
# ============================================================================
if [ "$SKIP_R_PACKAGES" = false ]; then
    echo "=== Step 4: Skipped (R packages no longer required) ==="
    echo "  - gomats 替代 RNA_Splicing.R"
    echo "  - htseq2matrix-go 替代 htseq2matrix.R"
    echo ""
else
    echo "=== Step 4: Skipped (--skip-r-packages) ==="
    echo ""
fi

# ============================================================================
# Summary
# ============================================================================
echo "======================================"
echo "Installation Summary"
echo "======================================"
echo ""
echo "Installation directory: $INSTALL_DIR"
echo "Runtime directory: $RUNTIME_DIR"
echo ""

# Check if binaries are in PATH
if [ -f "$INSTALL_DIR/xdxtools" ]; then
    if command -v xdxtools &> /dev/null; then
        echo "✓ xdxtools is in PATH"
    else
        echo "⚠ Add to PATH: export PATH=\"\$PATH:$INSTALL_DIR\""
    fi
fi

if [ -f "$INSTALL_DIR/enva" ]; then
    if command -v enva &> /dev/null; then
        echo "✓ enva is in PATH"
    else
        echo "⚠ Add to PATH: export PATH=\"\$PATH:$INSTALL_DIR\""
    fi
fi


echo ""
echo "Next steps:"
echo "  1. Add binaries to PATH (if not already):"
echo "     export PATH=\"\$PATH:$INSTALL_DIR\""
echo ""
echo "  2. Verify installation:"
echo "     xdxtools --version"
echo "     enva --version"
echo ""
echo "  3. Create a new analysis:"
echo "     cd $RUNTIME_DIR"
echo "     xdxtools create --fastq /path/to/fastq --pdata samples.csv --output ./userspace --jobid demo_run"
echo "     xdxtools run --config ./userspace/demo_run/config/config.yaml --dry-run"
echo ""
