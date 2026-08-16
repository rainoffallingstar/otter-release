# Note 2026-08-16: Gate 6 BS-PDX Craftmake Execution

## Scope

The production seven-input Gate 6 comparison is Craftmake-only. This record
covers the remaining BS-PDX project and does not use the completed Snakemake
compatibility path.

## Starting state

- Immutable snapshot `run-20260815T094416Z-rjwvkx` for
  `bs-pdx-SRR23802966` completed its Craftmake `step1` phase and retains the
  shared QC and Trim Galore outputs under its isolated run root.
- Craftmake `step2` dry-run completed successfully against that exact snapshot.
- The next action is Otter's foreground Craftmake controller for `step2`; it
  will be the only route used to submit BS-PDX downstream Slurm work.

## Completion criteria

- Complete BS-PDX `step2`, its applicable checker phase, and methylation
  `step3` through the same immutable Craftmake snapshot or explicit
  Craftmake resume if a classified Slurm incident occurs.
- Preserve all controller and worker evidence.
- Collect verified artifacts for a toolchain comparison only after an eligible
  modern/legacy Craftmake pair is available; no incompatible snapshots will be
  compared.
