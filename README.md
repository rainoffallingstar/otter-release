# otter

`otter` is a bioinformatics workflow CLI for RRBS, WGBS, RNA-seq, and PDX analysis. The canonical repository is [rainoffallingstar/otter](https://github.com/rainoffallingstar/otter).

## Product stack

```text
otter → craftmake → enva → operators → bamdriver
```

- `otter`: user-facing project, configuration, task, and workflow coordinator.
- `craftmake`: Go workflow execution layer intended to replace Snakemake.
- `enva`: environment lifecycle and command isolation for `otter-core`, `otter-snakemake`, and `otter-extra`.
- Operators: `fastqcx`, `xenofilx`, `pairbam`, `seq2mat`, `matsrun`, `qctb`, and `methx`.
- `bamdriver`: shared low-level BAM operation layer.

`craftmake` integration is still migrating. The current runtime is dual-track: existing workflows continue through Snakemake while `craftmake` compatibility and adoption are completed. Documentation must not be read as claiming that Snakemake has already been removed.

## Features

- RRBS, WGBS, RNA-seq, and PDX workflows
- FASTQ pairing, pdata validation, and per-sample adapter generation
- SLURM, SLURM Job Array, and local execution
- Background task status and log recovery
- Excel/CSV pdata with Chinese column aliases
- Rattler-first environment management through `enva`
- Specialized native operators with stable scientific formats where applicable

## Requirements

- Go 1.24+
- Snakemake for the current production path
- `craftmake` for migration and compatibility validation
- [enva](https://github.com/rainoffallingstar/enva)
- `conda`, `mamba`, or `micromamba` only for compatible/adopted environments

## Installation

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/rainoffallingstar/otter/main/scripts/install.sh)
```

Build from source:

```bash
git clone --recurse-submodules https://github.com/rainoffallingstar/otter.git
cd otter
conda activate go-env
go build -o otter .
```

The repository is in a naming migration. Some source symbols, state files, or older release assets may still use `xdxtools`; treat those as compatibility names, not as the current product name.

## Quick start

```bash
otter init my_project

otter create --fastq /data/fastq --mode RRBS --pdata samples.csv \
  --output my_project/userspace --jobid demo_rrbs

otter run --config my_project/userspace/demo_rrbs/config/config.yaml

otter task list
otter task status <task-id>
otter task logs <task-id> --follow
```

`otter run` currently dispatches the established Snakemake path. `craftmake` is the Go replacement execution layer under migration and is not yet the sole production backend.

## Commands

| Command | Purpose |
|---|---|
| `init` | Install workflow assets into a project directory |
| `create` | Scan inputs, validate samples, and generate configuration |
| `run` | Execute a workflow, using the current dual-track runtime policy |
| `task` | List, inspect, follow, or stop background tasks |
| `status` | Show project workflow progress |
| `config` | Validate configuration |

## Environment names

| Environment | Purpose |
|---|---|
| `otter-core` | Core bioinformatics runtime and operators |
| `otter-snakemake` | Current Snakemake compatibility runtime |
| `otter-extra` | Additional analysis and visualization tools |

## Operator names

| Current name | Scientific role | Historical repository name |
|---|---|---|
| `fastqcx` | FASTQ quality control; FastQC/MultiQC-compatible outputs | `fastqc-rs` |
| `xenofilx` | PDX xenograft read classification | `xenofilter-go` |
| `pairbam` | Paired-read BAM filtering | `Paireads` |
| `seq2mat` | HTSeq count-to-matrix conversion | `htseq2matrix-go` |
| `matsrun` | rMATS orchestration | `gomats` |
| `qctb` | QC aggregation | unchanged |
| `methx` | methylation/HDF5 processing; Methrix domain | `methrix-cli` |
| `bamdriver` | shared BAM operations | `bamdriver-go` |

Historical review and archive documents intentionally retain the names that were true when their evidence was recorded.

## Documentation

- [Architecture](docs/architecture.md)
- [Installation](docs/installation.md)
- [Build guide](docs/build.md)
- [User manual](docs/manual/README.md)
- [Submodule guide](docs/submodules-build-guide.md)
- [Current context](docs/active_context.md)

## License

MIT
