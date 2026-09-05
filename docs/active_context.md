# Current project context

This page is a short current-state index, not the primary evidence record. Detailed run IDs, checksums, scheduler output, benchmark measurements, and scientific comparison tables belong in the linked reports and evidence register.

## Current decision

Gate 6 is closed for its accepted bounded scope:

- real-SLURM Craftmake–Snakemake executor comparison;
- corrected Gate A/B preservation evidence;
- corrected Gate C NM/oracle and CT/GA controls;
- corrected Gate D disagreement accounting;
- Methx/Methrix production CpG matrix parity and HDF5 conversion guidance.

The project remains dual-track. Canonical v1 runs use Craftmake after immutable resolution. Established `otter.yaml` projects use explicit Snakemake compatibility. No automatic executor fallback is allowed.

## Deferred extensions

The following were not run and are not current release claims:

- fresh seven-input legacy-equivalent matrix;
- representative `20 samples × 3 repeats` matrix;
- production-scale throughput and scheduler-pressure testing;
- WGBS `SRR6373947` requalification;
- additional Snakemake interruption/recovery studies.

## Current product map

```text
otter → craftmake → enva → operators → bamdriver
```

The root CLI owns project/configuration and task control. Craftmake owns native workflow execution and state. Enva owns runtime environments. Operators own domain outputs. Bamdriver provides shared BAM/BGZF primitives.

## Current user paths

Legacy-compatible project:

```bash
otter init my_project
otter create --fastq ./fastq --mode RRBS --pdata ./samples.xlsx \
  --output my_project/userspace --jobid demo_rrbs
otter run \
  --config my_project/userspace/demo_rrbs/config/otter.yaml \
  --executor snakemake --engine local --foreground
```

Canonical v1 project:

```bash
otter config validate --config project.yaml --schema v1
otter config resolve --project project.yaml --backend local
otter run --config runs/<run-id>/run.yaml \
  --executor craftmake --phase step1 --backend local --foreground
```

## Evidence navigation

- [Release readiness](release-readiness.md) — formal release checklist and open decisions.
- [Gate 6 comparison report](gate6-toolchain-comparison-report.md) — detailed scientific and executor evidence.
- [Gate 6 evidence register](gate6-closeout-evidence-register.json) — machine-readable acceptance boundary.
- [Benchmark plan](benchmark-plan.md) — metrics, deferred matrices, and comparison tiers.
- [Paracloud operations](gate6-paracloud-operations.md) — runtime, reference, and scheduler evidence.
- [Xenofilx benchmark](../xenofilx/benchmark/) — PDX mixture and classification evidence.

## Maintenance rule

Update this page only when the current decision, supported entry points, or deferred-work boundary changes. Put experiment-specific detail in a dated note or the relevant evidence report, then link it here.

[Back to the documentation hub](README.md)
