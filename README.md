<p align="right">
  <strong>English</strong> · <a href="./README_zh.md">简体中文</a>
</p>

# otter

**A reproducible bioinformatics workflow CLI for RRBS, WGBS, RNA-seq, and PDX analysis.**

Otter turns FASTQ inputs and sample metadata into validated workflow projects, then runs them locally or on SLURM with task tracking, resumability, and auditable outputs.

<p align="center">
  <img src="./docs/otter-workflow-stack.svg" width="100%" alt="Otter workflow stack from project control through Craftmake, Enva, domain operators, and Bamdriver">
</p>

## What you can do

- Start a workflow project with `init` and generate a validated analysis configuration with `create`.
- Run RRBS, WGBS, RNA-seq, BS-PDX, and RNA-PDX workflows.
- Use local execution or SLURM, including per-step CPU, memory, and partition controls on the current Snakemake path.
- Track background runs with `task list`, `task status`, `task logs`, `task stop`, and `task report`.
- Resolve canonical project files into immutable `otter.run/v1` snapshots with reference and resource identity.
- Validate published artifact manifests and checksums at the run boundary.

## The workflow stack

```text
otter → craftmake → enva → operators → bamdriver
  │         │         │         │          │
  │         │         │         │          └─ shared BAM/BGZF primitives
  │         │         │         └─ fastqcx · xenofilx · pairbam · seq2mat
  │         │         │            matsrun · qctb · methx
  │         │         └─ rattler-first runtime environments
  │         └─ native Local/SLURM executor and task state
  └─ project, configuration, workflow, and task control plane
```

The runtime is intentionally dual-track. Existing production workflows use the Snakemake compatibility path; `craftmake` is the native Go execution layer being integrated and validated. Otter does not claim that Snakemake has already been removed.

## Evidence and release boundary

The accepted Gate 6 scope covers bounded executor comparison, corrected read/BAM classification evidence, and Methx/Methrix scientific parity. It does **not** claim production-scale throughput, a fresh seven-input legacy-equivalent matrix, complete WGBS qualification, or universal Snakemake replacement. See the [workflow catalog](docs/workflow-catalog.md) and [Gate 6 evidence register](docs/gate6-closeout-evidence-register.json).

## Quick start

### Install a release

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/rainoffallingstar/otter/main/scripts/install.sh)
```

### Build from source

```bash
git clone --recurse-submodules https://github.com/rainoffallingstar/otter.git
cd otter
conda activate go-env
go build -o otter .
```

### Create and run a legacy-compatible project

```bash
otter init my_project

otter create \
  --fastq /data/fastq \
  --mode RRBS \
  --pdata /data/samples.xlsx \
  --output my_project/userspace \
  --jobid demo_rrbs

otter run \
  --config my_project/userspace/demo_rrbs/config/otter.yaml \
  --executor snakemake \
  --engine local \
  --foreground
```

For a cluster run, replace `--engine local` with `--engine slurm` and provide the required partition/resource settings for your site. The default background mode prints a task ID; inspect it with:

```bash
otter task list
otter task status <task-id>
otter task logs <task-id> --follow
```

### Use the canonical immutable run model

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

`craftmake` requires an immutable `otter.run/v1` snapshot. The executor, backend, phase, references, and resource envelope are resolved before execution; runtime flags cannot silently override that snapshot.

## Components

| Component | Role | Repository |
| --- | --- | --- |
| `otter` | Workflow project and task control plane | [rainoffallingstar/otter](https://github.com/rainoffallingstar/otter) |
| `craftmake` | Native workflow compiler and Local/SLURM executor | [rainoffallingstar/craftmake](https://github.com/rainoffallingstar/craftmake) |
| `enva` | Rattler-first environment lifecycle manager | [rainoffallingstar/enva](https://github.com/rainoffallingstar/enva) |
| `fastqcx` | FASTQ QC with FastQC-compatible summary output | [rainoffallingstar/fastqcx](https://github.com/rainoffallingstar/fastqcx) |
| `xenofilx` | Graft/host read classification for PDX data | [rainoffallingstar/xenofilx](https://github.com/rainoffallingstar/xenofilx) |
| `pairbam` | Paired-end BAM filtering and ordering | [rainoffallingstar/pairbam](https://github.com/rainoffallingstar/pairbam) |
| `seq2mat` | HTSeq count-to-expression-matrix conversion | [rainoffallingstar/seq2mat](https://github.com/rainoffallingstar/seq2mat) |
| `matsrun` | rMATS pairwise splicing orchestration | [rainoffallingstar/matsrun](https://github.com/rainoffallingstar/matsrun) |
| `qctb` | Versioned QC summary reporting | [rainoffallingstar/qctb](https://github.com/rainoffallingstar/qctb) |
| `methx` | Bismark coverage and methylation HDF5 processing | [rainoffallingstar/methx](https://github.com/rainoffallingstar/methx) |
| `bamdriver` | Shared pure-Go BAM/BGZF library | [rainoffallingstar/bamdriver](https://github.com/rainoffallingstar/bamdriver) |

> The links above follow the public repository naming contract. Historical source symbols and old asset names may still contain `xdxtools`, `fastqc-rs`, `xenofilter-go`, `Paireads`, `htseq2matrix-go`, `gomats`, `methrix-cli`, or `bamdriver-go`; those names are retained only where compatibility or historical evidence requires them.

## Documentation

- [Documentation hub](docs/README.md) — current contracts, tutorials, operations, and historical evidence map.
- [User manual](docs/manual/README.md) — installation, data preparation, quick start, modes, advanced usage, components, and FAQ.
- [Architecture](docs/architecture.md) — system boundaries and migration model.
- [Workflow catalog](docs/workflow-catalog.md) — scenarios, phases, artifacts, and comparison ownership.
- [Configuration and run snapshots](docs/configuration.md) — legacy and canonical configuration models.
- [Installation](docs/installation.md) — release, source, environment, and troubleshooting details.
- [Build and submodules](docs/build.md) · [submodule build guide](docs/submodules-build-guide.md).
- [Release readiness](docs/release-readiness.md) — current release checklist and deferred evidence boundary.
- [Methx → native Methrix HDF5 exporter](methx/scripts/export_methrix_hdf5.R) — R-side interoperability path.

## Development

```bash
conda activate go-env
go test -v ./...
go vet ./...
```

Rust submodules use the `rust_build` environment. Each submodule is an independent repository; see its own README for focused build and test commands.

## Naming note

The repository and product name is `otter`. Some source-level command and state symbols are still compatibility-era `xdxtools` names. Documentation distinguishes the current product name from those implementation aliases instead of pretending the rename is complete.

## License

MIT
