# xdxtools Build Guide

## Prerequisites

- Go 1.24+
- Rust 1.85+
- Git 2.0+
- Conda/Mamba (recommended)

Recommended conda environments:

- Go: `go-env` (compatible fallback: `go-build`)
- Rust: `rust_build` (compatible fallback: `rust-build`)

## Build Main CLI

```bash
conda activate go-env
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o xdxtools .
```

If Go module download is unstable:

```bash
GOPROXY=https://goproxy.cn,direct go build -o xdxtools .
```

## Build All Submodule Tools

```bash
bash scripts/build-all-submodules.sh
```

The script builds and installs these binaries to `$HOME/.cargo/bin`:

- `enva`
- `fqc`
- `xenofilter`
- `paireads`
- `htseq2matrix`
- `methrix-cli`
- `qctb`
- `gomats`

By default the script runs in strict mode (`STRICT_MODE=1`) and exits non-zero when any required binary is missing.  
To downgrade to warnings only:

```bash
STRICT_MODE=0 bash scripts/build-all-submodules.sh
```

## Verify Toolchain Consistency

Before PR/release:

```bash
bash scripts/verify_toolchain_consistency.sh
```

This checks that commands used in `inst/rules/*.smk` are covered by:

- `scripts/install.sh` tool list
- `scripts/build-all-submodules.sh` required binary list

## Test and Validation

```bash
conda run -n go-env go test -v ./...
conda run -n go-env go vet ./...
```

Minimum runtime smoke check:

```bash
xdxtools run --dry-run --config <path-to-config.yaml>
```
