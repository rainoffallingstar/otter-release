# Otter user manual

This manual is organized around the first successful run rather than the internal package layout.

## Choose a path

| Goal | Chapter |
| --- | --- |
| Install a release or build from source | [1. Installation](01-installation.md) |
| Prepare FASTQ files and sample metadata | [2. Data preparation](02-data-preparation.md) |
| Create and run your first project | [3. Quick start](03-quickstart.md) |
| Select RRBS, WGBS, RNA-seq, or PDX behavior | [4. Analysis modes](04-analysis-modes.md) |
| Use custom references, SLURM resources, recovery, and canonical snapshots | [5. Advanced usage](05-advanced-usage.md) |
| Understand the companion tools | [6. Component reference](06-subtools.md) |
| Diagnose common failures | [7. FAQ](07-faq.md) |

## The current execution model

```text
otter → craftmake → enva → operators → bamdriver
```

The production compatibility path remains Snakemake. Craftmake is the native Go executor under integration and validation. Use the executor explicitly when working with canonical immutable `otter.run/v1` snapshots; do not treat it as an undocumented fallback for every legacy project.

## Core command flow

```bash
otter init my_project
otter create --fastq ./fastq --mode RRBS --pdata ./samples.xlsx \
  --output my_project/userspace --jobid demo_rrbs
otter run --config my_project/userspace/demo_rrbs/config/otter.yaml \
  --executor snakemake \
  --engine local --foreground
```

For canonical projects:

```bash
otter config validate --config project.yaml --schema v1
otter config resolve --project project.yaml --backend local
otter run --config runs/<run-id>/run.yaml \
  --executor craftmake --phase step1 --backend local --foreground
```

## Before you begin

- Linux or macOS for local development; Linux with SLURM is the primary production target.
- Paired FASTQ files and a matching pdata file for `create`.
- Reference FASTA/index assets appropriate for the selected scenario.
- `otter-snakemake` for the current Snakemake compatibility path.
- `enva` and the managed runtime environments when using the release workflow setup.

[Back to documentation hub](../README.md)
