# Site and backend profiles

Site resolution turns portable project intent into an execution-ready run snapshot. The selected backend, site, resource envelope, and evidence supporting that choice must be recorded before execution.

## Default selection

```yaml
execution:
  executor: craftmake
  backend: auto
  site: auto
```

Use `--backend local` for development or contract tests. Use `--backend slurm` for real workflow execution when the site is configured and visible from compute nodes.

## Backend detection

```text
backend=auto
     ↓
check sbatch/squeue/sacct/scancel/srun
     ↓
none available ─────────→ local
all available ──────────→ validate cluster and paths
partial/inconsistent ───→ fail closed
```

A complete SLURM toolchain is not enough. Detection also validates the cluster, selected partition/account/QOS, resource limits, shared project/reference paths, scratch policy, and compute-node visibility.

Otter must not silently choose Local when SLURM is partially available or when a production reference path cannot be proven accessible.

## Inspect the current environment

```bash
otter site list
otter site validate
otter site validate production-cluster
```

The validation result reports backend, site ID, source, reason, cluster, and detected commands.

## Site profile

A profile can provide stable site defaults without embedding credentials:

```yaml
schema_version: otter.site/v1
site:
  id: production-cluster
  backend: slurm
slurm:
  partition: cpu
  account: genomics
  qos: normal
  max_jobs: 100
  default_time: 24:00:00
paths:
  reference_root: /shared/otter/references
  scratch_root: /scratch/otter
```

Profiles may define defaults and constraints. Authentication, tokens, and private credentials belong to the site's standard environment, not to committed YAML.

## Resource precedence

```text
explicit CLI value > site profile/detection > project value > workflow default
```

For canonical immutable runs, the effective values are written to `run.yaml`. Craftmake then rejects runtime overrides that would change the snapshot. For legacy Snakemake runs, the existing `otter run` resource flags remain available:

```bash
otter run \
  --config my_project/userspace/<jobid>/config/otter.yaml \
  --executor snakemake \
  --engine slurm \
  --slurm-partition cpu \
  --slurm-cores 16 \
  --slurm-memory 64G
```

## Fail-closed rules

Resolution must stop before `sbatch` when:

- required SLURM commands are missing or inconsistent;
- the requested partition/account/QOS is unavailable;
- CPU, memory, time, or submission limits cannot satisfy the run;
- shared project, reference, or scratch paths are not visible where required;
- a reference manifest or workflow asset digest differs;
- an automatic choice cannot be explained deterministically.

The diagnostic should identify the candidate, rejected constraint, expected value, actual value, and affected path or resource.

## Snapshot stability

Site state can change after a snapshot is written. That must not mutate the existing run. Re-resolve to create a new run when the selected site, backend, partition, resource envelope, or reference visibility changes.

[Back to the documentation hub](README.md)
