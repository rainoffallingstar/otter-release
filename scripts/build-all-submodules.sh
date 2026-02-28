#!/bin/bash
# Build all submodules and install to $HOME/.cargo/bin

set -e

CARGO_BIN="$HOME/.cargo/bin"
mkdir -p "$CARGO_BIN"

# Check for Go
GO_CMD=""
if command -v go &> /dev/null; then
    GO_CMD="go"
elif conda env list | grep -q "^go-build "; then
    GO_CMD="conda run -n go-build go"
else
    echo "Error: Go not found. Please install Go first."
    exit 1
fi

# Check for Rust
RUST_CMD=""
if command -v cargo &> /dev/null; then
    RUST_CMD=""
elif conda env list | grep -q "^rust-build "; then
    RUST_CMD="conda run -n rust-build"
else
    echo "Warning: Rust not found in PATH, will try to use conda environment..."
fi

echo "======================================"
echo "Building all xdxtools submodules"
echo "======================================"
echo "Go: $GO_CMD"
echo "Rust: $RUST_CMD"
echo ""

# Function to build Go projects
build_go() {
    local name=$1
    local dir=$2
    local binary=$3
    local main_pkg=$4

    echo "Building $name..."
    cd "$dir"

    if [ -n "$GO_CMD" ]; then
        echo "  Building $binary..."
        $GO_CMD build -o "$CARGO_BIN/$binary" "$main_pkg"
        echo "  ✓ Installed to $CARGO_BIN/$binary"
    else
        echo "  ✗ Go not found, skipping $name"
    fi

    cd ..
    echo ""
}

# Function to build Rust projects
build_rust() {
    local name=$1
    local dir=$2
    local binary=$3

    echo "Building $name..."
    cd "$dir"

    if [ -n "$RUST_CMD" ]; then
        echo "  Building $binary..."
        $RUST_CMD cargo build --release
        cp "target/release/$binary" "$CARGO_BIN/$binary"
    elif command -v cargo &> /dev/null; then
        echo "  Building $binary..."
        cargo build --release
        cp "target/release/$binary" "$CARGO_BIN/$binary"
    else
        echo "  ✗ Rust/Cargo not found, skipping $name"
        cd ..
        echo ""
        return
    fi

    echo "  ✓ Installed to $CARGO_BIN/$binary"
    cd ..
    echo ""
}

# ============================================
# Go Projects
# ============================================

build_go "xenofilter-go" "xenofilter-go" "xenofilter" "./cmd/xenofilter"
build_go "Paireads" "Paireads" "paireads" "./cmd/paireads"
build_go "gomats" "gomats" "gomats" "./cmd/gomats"

# ============================================
# Rust Projects
# ============================================

build_rust "enva" "enva" "enva"
build_rust "methrix-cli" "methrix-cli-local" "methrix-cli"
build_rust "qctb" "qctb" "qctb"

# ============================================
# Verify installations
# ============================================

echo "======================================"
echo "Verifying installations"
echo "======================================"

for binary in enva xenofilter paireads methrix-cli qctb gomats; do
    if [ -x "$CARGO_BIN/$binary" ]; then
        VERSION=$("$CARGO_BIN/$binary" --version 2>&1 | head -1 || echo "OK")
        echo "✓ $binary - $VERSION"
    else
        echo "✗ $binary - NOT FOUND or NOT EXECUTABLE"
    fi
done

echo ""
echo "Build complete!"
echo "Binaries installed to: $CARGO_BIN"
