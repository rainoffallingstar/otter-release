# xdxtools

A bioinformatics workflow CLI for RRBS, WGBS, RNA-seq, and PDX analysis.

## Features

- Workflow modes: RRBS / WGBS / RNA-seq / PDX
- Auto FASTQ pairing + per-sample adapter generation (barcode support)
- SLURM Job Array + local worker pool parallelization
- Excel/CSV pdata with Chinese column name auto-mapping
- Snakemake integration with embedded workflow files
- Rattler-first environment management via `enva`, with conda-compatible environment fallback
- Single primary CLI binary; workflow runtime depends on Snakemake + `enva` (legacy conda-compatible envs also work)

## Requirements

- Go 1.24+
- Snakemake
- [enva](https://github.com/rainoffallingstar/enva) (preferred)
- `conda` / `mamba` / `micromamba` only for legacy compatibility or adopted environments

## Installation

### Quick install (interactive)

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh)
```

The script downloads pre-built binaries from GitHub Releases and sets up the required conda environments interactively. By default it now pulls binary assets from the public mirror repo `rainoffallingstar/flightlight`, falls back to `rainoffallingstar/xdxtools-go` if a requested release is not mirrored yet, and downloads env YAML files from the same selected release repo. The first interactive prompt lets you choose English or Chinese. Use `--lang en` or `--lang zh` to force the interface language, or use `scripts/setup.sh` only for source builds. When existing binaries are detected, the installer asks once whether to overwrite them, then applies that decision to all remaining binaries in the same run.

If the repository itself is private, anonymous `raw.githubusercontent.com` downloads return `404`, and `wget -qO-` hides that failure. Use an authenticated bootstrap command instead. If the release repo or assets are private, export `GITHUB_TOKEN` (or `GH_TOKEN` / `GITHUB_PAT`) first. For private or custom layouts, you can also set `GITHUB_RELEASES_REPO=<owner>/<repo>` and `GITHUB_FALLBACK_RELEASES_REPO=<owner>/<repo>`. In interactive mode, if GitHub access fails and no token is configured, the installer can prompt for a hidden token and retry once.

Common options:

```bash
# Non-interactive, use all defaults
bash <(curl -fsSL https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh) --non-interactive

# Specify a release version
bash <(curl -fsSL https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh) --version v0.3.0

# Private repository bootstrap
export GITHUB_PAT=<your_pat>
bash <(curl -fsSL -H "Authorization: Bearer ${GITHUB_PAT}" \
  https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh)

# Private release fork
export GITHUB_PAT=<your_pat>
export GITHUB_RELEASES_REPO=<owner>/<repo>
bash <(curl -fsSL -H "Authorization: Bearer ${GITHUB_PAT}" \
  https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh)

# Skip conda environment creation
bash <(curl -fsSL https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh) --skip-envs
```

### Build from source

```bash
git clone --recurse-submodules https://github.com/rainoffallingstar/xdxtools-go.git
cd xdxtools-go
go build -o xdxtools
```

## Codex Skill

This repository ships an installable Codex skill at `skills/xdxtools`. The skill is meant for the root `xdxtools` repo and its matched submodules: `enva`, `Paireads`, `bamdriver-go`, `fastqc-rs`, `gomats`, `htseq2matrix-go`, `methrix-cli`, `qctb`, and `xenofilter-go`.

Install it from GitHub with Codex's bundled skill installer:

```bash
python ~/.codex/skills/.system/skill-installer/scripts/install-skill-from-github.py --repo rainoffallingstar/xdxtools-go --path skills/xdxtools
```

If you are installing from a fork or a non-default branch, replace `--repo` and optionally add `--ref <branch-or-tag>`.

After installing, restart Codex to pick up the new skill.

You can then invoke it explicitly in Codex prompts, for example:

```text
Use $xdxtools to inspect the root workflow CLI and update the install docs.
Use $xdxtools to work on qctb without breaking xdxtools submodule boundaries.
```

## Quick Start

```bash
# 1. Initialize project (copies Snakemake workflow files)
xdxtools init my_project

# 2. Scan FASTQ, validate samples, generate config
xdxtools create --fastq /data/fastq --mode RRBS --pdata samples.csv --output my_project/userspace --jobid demo_rrbs

# 3. Execute workflow (submitted in the background by default)
xdxtools run --config my_project/userspace/demo_rrbs/config/config.yaml

# 4. Inspect the background task and follow its log
xdxtools task list
xdxtools task status <task-id>
xdxtools task logs <task-id> --follow
```

## Commands

| Command | Description |
|---------|-------------|
| `init`   | Copy Snakemake workflow files to project directory |
| `create` | Scan FASTQ, validate samples, generate config.yaml |
| `run`    | Execute Snakemake workflow as a background task by default |
| `task`   | List, inspect, follow logs, or stop background tasks |
| `status` | Show workflow progress for a project directory |
| `config` | Validate configuration file |

### Key `run` flags

| Flag | Default | Description |
|------|---------|-------------|
| `--engine` | `auto` | Execution engine: `slurm` / `local` / `auto` |
| `--slurm-partition` | empty | Optional SLURM partition override |
| `--parallel-jobs` | `2` | Max concurrent jobs |
| `--dry-run` | `false` | Validate without executing; always runs in the foreground |
| `--foreground` / `-F` | `false` | Keep the workflow attached to the current terminal |
| `--resume` / `-r` | `false` | Resume from last completed step |

`run` creates an independent background task by default, so disconnecting SSH does not stop the xdxtools coordinator. Task metadata and the complete startup log are stored under `$XDG_STATE_HOME/xdxtools/tasks/`, or `~/.local/state/xdxtools/tasks/` when `XDG_STATE_HOME` is unset. This restores status and log viewing, not an interactive tmux terminal session.

```bash
xdxtools task list
xdxtools task list --all
xdxtools task status <task-id>
xdxtools task logs <task-id> --follow
xdxtools task stop <task-id>
```

Use `--verbose` for detailed output or `--dry-run` to troubleshoot without running.

## Workflow Modes

| Mode | Flag |
|------|------|
| RRBS (Reduced Representation BS-seq) | `--mode RRBS` |
| WGBS (Whole Genome BS-seq) | `--mode WGBS` |
| RNA-seq | `--mode RNASEQ` |
| PDX (Patient-Derived Xenograft) | `--mode RRBS --species1 human --species2 mouse` |

PDX mode is enabled automatically when both `--species1` and `--species2` are specified.

## License

MIT
