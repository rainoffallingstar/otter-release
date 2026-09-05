# 3. Quick start

This path creates a legacy-compatible project and runs it locally. It is the shortest route to verify installation and input contracts.

## 1. Initialize the project

```bash
otter init my_project
```

The command copies the current workflow assets, environments, scripts, rules, and an assets manifest into `my_project/`.

## 2. Create a configuration

```bash
otter create \
  --fastq /data/fastq \
  --mode RRBS \
  --pdata /data/samples.xlsx \
  --output my_project/userspace \
  --jobid demo_rrbs
```

The generated configuration is:

```text
my_project/userspace/demo_rrbs/config/otter.yaml
```

Use `otter create --help` for custom FASTQ suffixes and reference overrides.

## 3. Validate and dry-run

```bash
otter config validate \
  --config my_project/userspace/demo_rrbs/config/otter.yaml

otter run \
  --config my_project/userspace/demo_rrbs/config/otter.yaml \
  --executor snakemake \
  --engine local \
  --dry-run \
  --foreground
```

The current legacy configuration path dispatches through the Snakemake compatibility runtime. A dry-run checks the configuration and workflow planning; it does not establish scientific parity.

## 4. Run locally

```bash
otter run \
  --config my_project/userspace/demo_rrbs/config/otter.yaml \
  --executor snakemake \
  --engine local \
  --parallel-jobs 4 \
  --foreground
```

Without `--foreground`, Otter creates a background task and prints its task ID.

## 5. Run on SLURM

```bash
otter run \
  --config my_project/userspace/demo_rrbs/config/otter.yaml \
  --executor snakemake \
  --engine slurm \
  --slurm-partition normal \
  --parallel-jobs 8
```

Use the partition, CPU, memory, and checker settings defined for your site. See [advanced usage](05-advanced-usage.md).

## 6. Inspect progress

For a background task:

```bash
otter task list
otter task status <task-id>
otter task logs <task-id> --follow
```

For a project state directory:

```bash
otter status my_project/userspace/demo_rrbs
```

Stop a task safely:

```bash
otter task stop <task-id>
```

## Canonical run path

For new reproducible projects, use the v1 configuration and immutable run snapshot model:

```bash
otter config validate --config project.yaml --schema v1
otter config resolve --project project.yaml --backend local
otter run --config runs/<run-id>/run.yaml \
  --executor craftmake --phase step1 --backend local --foreground
```

The canonical path makes executor, backend, reference, and resource selection explicit. See [configuration](../configuration.md) and [execution contract](../execution-contract.md).

[Back to the manual](README.md) · [Next: analysis modes](04-analysis-modes.md)
