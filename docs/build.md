# otter Build Guide

## Prerequisites

- Go 1.24+
- Rust toolchain compatible with locked submodule dependencies
- Git 2+
- HDF5 development libraries for `methx`
- Recommended build environments: `go-env` and `rust_build`

## Build the parent CLI

```bash
conda activate go-env
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o otter .
go test -count=1 ./...
go vet ./...
```

The source root command may still identify as `xdxtools` during the code migration. Producing an `otter` filename does not by itself complete that migration; validate `./otter --help` before packaging.

## Build the execution and environment layers

```bash
# craftmake: Go replacement execution layer, integration still migrating
(cd craftmake && CGO_ENABLED=0 go build -trimpath -o ../target/bin/craftmake .)

# enva
(cd enva && cargo build --locked --release)
```

`craftmake` and Snakemake are currently dual-track. Building `craftmake` does not mean the parent CLI dispatches all production workflows through it.

## Build operators

| Component | Language | Expected binary | Special dependency |
|---|---|---|---|
| `fastqcx` | Rust | `fastqcx` | FastQC/MultiQC compatibility fixtures |
| `xenofilx` | Go | `xenofilx` | BAM data |
| `pairbam` | Go | `pairbam` | BAM data |
| `seq2mat` | Go | `seq2mat` | HTSeq mapping data |
| `matsrun` | Go | `matsrun` | external rMATS runtime |
| `qctb` | Rust | `qctb` | mode-specific report fixtures |
| `methx` | Rust | `methx` | HDF5 |
| `bamdriver` | Go | `bamdriver` | BAM fixtures |

Prefer each submodule's own build instructions because package layouts and CLI entrypoints are independently versioned. The consolidated commands are documented in [submodules-build-guide.md](submodules-build-guide.md).

## Environment assets

Build/release verification must agree on:

```text
otter-core
otter-snakemake
otter-extra
```

Search generated packages and release manifests for old environment prefixes before publication. Do not edit scripts or CI as part of a documentation-only migration.

## Validation

Minimum parent checks:

```bash
go test -count=1 ./...
go vet ./...
./otter --help
```

Minimum ecosystem checks:

- each Go submodule: `go test ./...`, `go vet ./...`, and race/static checks where supported;
- each Rust submodule: `cargo fmt --check`, `cargo check --locked`, strict `cargo clippy`, `cargo test --locked`;
- Snakemake RRBS/WGBS/RNA-seq/PDX smoke tests in `otter-snakemake`;
- craftmake task-graph/resource/result equivalence checks before switching a workflow;
- external compatibility checks retaining FastQC, MultiQC, Methrix, Bismark, HTSeq and rMATS names where scientifically required.

## Release gate

Do not announce a complete Snakemake replacement until `otter → craftmake` integration, all four workflow modes, local/SLURM behavior, recovery, and key scientific outputs pass. Until then release notes must say “migration in progress” or “dual-track.”
