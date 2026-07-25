# otter Installation Guide

## Current support model

The canonical repository is `rainoffallingstar/otter`. The product hierarchy is `otter → craftmake → enva → operators → bamdriver`.

`craftmake` is being integrated as the Go replacement for Snakemake. Installation remains dual-track: install Snakemake and the `otter-snakemake` environment for current production workflows, and install/build `craftmake` for migration validation. Do not remove Snakemake yet.

## Prerequisites

- Linux or macOS; Linux/SLURM is the primary production target
- Git with submodule support
- Go 1.24+
- Rust toolchain for Rust submodules
- HDF5 development/runtime libraries for `methx`
- Network access to GitHub and configured package channels

## Release installation

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/rainoffallingstar/otter/main/scripts/install.sh)
```

The installer in a checkout may still carry compatibility-era asset names until scripts and release artifacts are migrated. Verify the selected release before using it in production.

For private assets, export `GITHUB_TOKEN`, `GH_TOKEN`, or `GITHUB_PAT` as supported by the installer. Never place a token in a committed URL.

## Source installation

```bash
git clone --recurse-submodules https://github.com/rainoffallingstar/otter.git
cd otter
conda activate go-env
go build -o otter .
install -m 755 otter "$HOME/.cargo/bin/otter"
```

If an existing checkout was cloned without submodules:

```bash
git submodule update --init --recursive
```

Current submodule directories are:

```text
craftmake enva fastqcx xenofilx pairbam seq2mat matsrun qctb methx bamdriver
```

Build details are in [build.md](build.md) and [submodules-build-guide.md](submodules-build-guide.md).

## Runtime environments

Create the three environments with `enva`:

```bash
enva create --core
enva create --snakemake
enva create --extra
```

Expected names:

| Name | Purpose |
|---|---|
| `otter-core` | core bioinformatics dependencies and operators |
| `otter-snakemake` | current Snakemake compatibility runtime |
| `otter-extra` | optional analysis and visualization dependencies |

Verify:

```bash
enva list --detailed
enva validate --all
```

If the installed `enva` still emits `xdxtools-*`, that is an incomplete runtime-asset migration; do not relabel the environment manually without checking the corresponding YAML and workflow references.

## Verify binaries

```bash
command -v otter craftmake enva
command -v fastqcx xenofilx pairbam seq2mat matsrun qctb methx bamdriver

otter --version
craftmake --help
enva --version
```

Not every operator necessarily exposes `--version`; use `--help` when appropriate.

## Initialize and dry-run

```bash
otter init my_project

otter create --fastq /data/fastq --mode RRBS --pdata samples.csv \
  --output my_project/userspace --jobid demo_rrbs

otter run --config my_project/userspace/demo_rrbs/config/config.yaml --dry-run
```

The current production dry-run follows Snakemake. Run any `craftmake` compatibility check separately until `otter` integration explicitly exposes it.

## Compatibility note

The parent source snapshot may still compile a binary whose root command is `xdxtools`, and may still store state under old compatibility paths. This documentation defines the target/current product name but does not claim that the code rename has already occurred. If `otter --version` fails after a source build, inspect the produced root command before deploying; do not create undocumented symlink conventions in shared installations.

## Troubleshooting

- Empty submodule directory: run `git submodule update --init --recursive`.
- HDF5 link/load failure for `methx`: activate the Rust/HDF5 build environment and set `HDF5_DIR`, `PKG_CONFIG_PATH`, and the platform library path.
- Snakemake unavailable: validate `otter-snakemake`; `craftmake` is not yet a blanket runtime fallback.
- Environment still uses an old product prefix: treat it as a migration defect or explicit legacy environment, not as a new canonical name.
- GitHub bootstrap returns 404: authenticate if the repository/assets are private and verify the canonical `rainoffallingstar/otter` path.

## Uninstallation

Remove only paths you own and have verified:

```bash
rm -f "$HOME/.cargo/bin/otter"
rm -f "$HOME/.cargo/bin/craftmake"
rm -f "$HOME/.cargo/bin/enva"
```

Use `enva remove` for the three `otter-*` environments. Do not recursively delete a project or environment root without confirming its ownership and contents.
