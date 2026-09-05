# 1. Installation

This chapter covers release installation, source builds, runtime environments, and a first verification.

## Requirements

- Linux or macOS; Linux with SLURM is the primary production target.
- Git with recursive submodule support.
- Go 1.24+ for the root and Go submodules.
- Rust/Cargo for `enva`, `fastqcx`, `methx`, and `qctb`.
- HDF5 build/runtime support for `methx`.
- Access to GitHub and the configured package channels.

## Release installation

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/rainoffallingstar/otter/main/scripts/install.sh)
```

After installation, verify the selected binary and asset set rather than assuming the latest pre-release is production-qualified:

```bash
otter --help
otter --version
command -v otter
```

## Source installation

```bash
git clone --recurse-submodules https://github.com/rainoffallingstar/otter.git
cd otter
conda activate go-env
go build -o otter .
./otter --help
```

If the checkout was cloned without submodules:

```bash
git submodule sync --recursive
git submodule update --init --recursive
git submodule status --recursive
```

The independent component directories are `craftmake`, `enva`, `fastqcx`, `xenofilx`, `pairbam`, `seq2mat`, `matsrun`, `qctb`, `methx`, and `bamdriver`.

## Runtime environments

Use `enva` to create the standard environments:

```bash
enva create --all
enva list --detailed
enva validate --all
```

| Environment | Purpose |
| --- | --- |
| `otter-core` | Core workflow dependencies and operators. |
| `otter-snakemake` | Current Snakemake compatibility runtime. |
| `otter-extra` | Additional analysis and visualization tools. |

The production runtime remains dual-track. Install `otter-snakemake` for existing workflows and install/build `craftmake` for the canonical migration path.

## Verify component binaries

```bash
command -v otter craftmake enva
command -v fastqcx xenofilx pairbam seq2mat matsrun qctb methx
craftmake --help
enva --version
```

Not every operator exposes `--version`; use `--help` for those tools.

## HDF5 build environment

When building `methx` locally, configure HDF5 from the active environment or the platform installation:

```bash
export HDF5_DIR="$CONDA_PREFIX"
export HDF5_INCLUDE_DIR="$HDF5_DIR/include"
export HDF5_LIB_DIR="$HDF5_DIR/lib"
export PKG_CONFIG_PATH="$HDF5_DIR/lib/pkgconfig:${PKG_CONFIG_PATH:-}"
export LD_LIBRARY_PATH="$HDF5_DIR/lib:${LD_LIBRARY_PATH:-}"
```

See the [`methx` HDF5 guide](../../methx/docs/HDF5_DEPENDENCY.md) for component-specific details.

## Compatibility note

Some source-level symbols, generated state paths, or older release assets may still use `xdxtools`. That is a migration-era compatibility name; do not infer from a renamed output file that every code and runtime path has been renamed.

[Back to the manual](README.md) · [Documentation hub](../README.md)
