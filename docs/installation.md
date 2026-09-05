# Otter installation

Otter is installed as a root CLI plus independent workflow components. Installation is currently dual-track:

- **Snakemake compatibility:** required for established production projects generated around `otter.yaml`.
- **Canonical v1/Craftmake:** required when resolving `project.yaml` into an immutable `run.yaml`.

Installing Craftmake does not remove or replace Snakemake automatically.

## Prerequisites

- Linux or macOS; Linux with SLURM is the primary production target.
- Git with recursive submodule support.
- Go 1.24 or newer for the root CLI and Go components.
- Rust toolchain for Rust components.
- HDF5 development/runtime libraries for `methx`.
- Access to the required package channels and external scientific tools.

## Release installation

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/rainoffallingstar/otter/main/scripts/install.sh)
otter --help
```

Verify the release and its bundled assets before using it in production. If private assets are supported by the selected installer, provide credentials through the environment; never put a token in a committed URL or configuration file.

## Source installation

```bash
git clone --recurse-submodules https://github.com/rainoffallingstar/otter.git
cd otter
conda activate go-env
go build -trimpath -o otter .
install -m 755 otter "$HOME/.cargo/bin/otter"
otter --help
```

For an existing checkout:

```bash
git submodule sync --recursive
git submodule update --init --recursive
```

The current submodule directories are:

```text
craftmake enva fastqcx xenofilx pairbam seq2mat matsrun qctb methx bamdriver
```

Use [build.md](build.md) and [submodules-build-guide.md](submodules-build-guide.md) for component builds.

## Runtime environments

Create the managed environments with Enva:

```bash
enva create --core
enva create --snakemake
enva create --extra
enva list --detailed
enva validate --all
```

| Environment | Role |
| --- | --- |
| `otter-core` | Core dependencies and modern operators |
| `otter-snakemake` | Current Snakemake compatibility path |
| `otter-extra` | Optional analysis and visualization dependencies |

Environment names are part of the runtime contract. If a checkout still emits historical `xdxtools-*` names, inspect its assets and lock files rather than manually renaming directories.

## Verify the installation

```bash
command -v otter craftmake enva
command -v fastqcx xenofilx pairbam seq2mat matsrun qctb methx
otter --help
craftmake --help
enva --version
```

`bamdriver` is primarily a shared Go package and may not expose a standalone executable in every checkout. Some operators expose `--help` but not `--version`.

## First compatibility dry-run

Use this path for an established project layout:

```bash
otter init my_project
otter create \
  --fastq /data/fastq \
  --mode RRBS \
  --pdata /data/samples.csv \
  --output my_project/userspace \
  --jobid demo_rrbs

otter run \
  --config my_project/userspace/demo_rrbs/config/otter.yaml \
  --executor snakemake \
  --engine local \
  --dry-run \
  --foreground
```

The compatibility path consumes the generated `otter.yaml` and uses Snakemake explicitly.

## First canonical validation

For a new reproducible project, validate and resolve before execution:

```bash
otter config validate --config project.yaml --schema v1
otter config resolve --project project.yaml --backend local
otter run \
  --config runs/<run-id>/run.yaml \
  --executor craftmake \
  --phase step1 \
  --backend local \
  --foreground
```

The snapshot fixes inputs, references, executor, backend, resources, workflow assets, and digests. Runtime flags cannot silently mutate it.

## Troubleshooting

- **Empty submodules:** run `git submodule update --init --recursive`.
- **HDF5 failure in `methx`:** activate the Rust/HDF5 environment and set `HDF5_DIR`, `PKG_CONFIG_PATH`, and the platform library path.
- **Snakemake unavailable:** validate `otter-snakemake`; Craftmake is not a blanket fallback for legacy workflows.
- **SLURM partially available:** use `otter site validate`; production resolution fails closed rather than silently selecting Local.
- **Historical product prefix:** inspect the environment assets and repository revision before deployment.
- **Bootstrap 404:** verify the canonical repository path and whether the selected release assets are private.

## Uninstallation

Remove only binaries and environments you own:

```bash
rm -f "$HOME/.cargo/bin/otter" "$HOME/.cargo/bin/craftmake" "$HOME/.cargo/bin/enva"
```

Use `enva remove` for managed `otter-*` environments. Do not recursively delete a project, reference registry, or shared environment root without confirming ownership.

[Back to the documentation hub](README.md)
