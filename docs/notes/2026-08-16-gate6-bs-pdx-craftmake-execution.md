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

## Controller preflight incident

- The first `step2` controller, `41461969`, exited `1:0` before Craftmake
  submitted any workflow task because its Slurm batch environment did not
  expose `craftmake` on `PATH`.
- The failure is confined to Otter's Craftmake-binary preflight. It neither
  mutated the immutable snapshot nor created downstream task outputs.
- The recovery controller will explicitly pass the sealed
  `runtime/craftmake` binary and `runtime/catalog` root to Otter. This retains
  the same snapshot and delegates all task submission to Craftmake.

## Step2 resource-envelope incident

- Recovery controller `41461982` correctly loaded the sealed Craftmake runtime,
  then stopped before a workflow task was submitted because the immutable
  snapshot's `step2` envelope declares 40 cores while BeaverPDX requires an
  80-core allocation.
- The exact preflight failure was `phase resource envelope "step2" has 40
  cores but allocation requests 80`. No task output was created and the
  original snapshot remains unchanged.
- BS-PDX must now receive a fresh Otter-resolved immutable snapshot with a
  compliant 80-core `step2` resource envelope. The accepted `step1` work will
  be reused only through the established preserved-output mechanism.

## Canonical configuration serialization incident

- The remote `yq` implementation used for the 80-core update wrote control
  characters into `project.yaml`; Otter rejected the canonical configuration
  during validation before it could create a new snapshot.
- The remediation will reconstruct the canonical configuration from the
  previously accepted immutable snapshot and apply only the verified
  `step2.cores: 80` change with a YAML-safe writer before resolving a new run.

## Compliant immutable snapshot

- The canonical configuration was rewritten without control characters and
  passed `otter config validate` with `step2.cores: 80`.
- The initial direct resolve attempt created only the incomplete diagnostic
  directory described below. It did not publish an immutable snapshot.
- Compute-node validation controller `41462041` then passed against the
  canonical configuration. Durable resolve controller `41462044` published
  `run-20260816T063000Z-pdxrsl` with
  `run-20260815T094416Z-rjwvkx` as its immutable lineage parent.
- Only accepted `step1` directories are linked into the new run. The sealed
  Craftmake `step2` dry-run passed against that exact snapshot.

## Incomplete resolve diagnostic

- The first 80-core resolve attempt left directory
  `run-20260816T061209Z-sktqup` without an immutable `run.yaml`. Its
  subsequent Craftmake plan was therefore rejected before planning or task
  submission.
- The accepted step1 links placed in that incomplete directory are diagnostic
  only. It will not be used or repaired; a new Otter resolve will receive a
  different immutable run ID.

## Compute-node configuration visibility incident

- Resolve controller `41462033` confirmed that residual control characters
  remained visible on the shared filesystem even after the login-node
  validation appeared to pass. It exited before creating a snapshot.
- The configuration will be sanitized to its ASCII YAML character set, then
  validated from an allocated compute-node context before any further resolve
  submission.

## Step2 memory-envelope incident

- The 80-core controller `41462059` reached Craftmake's resource-contract
  validation but stopped before task submission: BeaverPDX requests
  343,597,383,680 bytes (320 GiB), while the new snapshot declares 160 GiB.
- BS-PDX `step2` requires a combined `80 cores / 320 GiB` envelope. A new
  immutable snapshot will be resolved with that complete allocation contract;
  all prior snapshots remain diagnostic evidence only.

## Complete resource-contract snapshot

- Compute-node validation controller `41462075` accepted the canonical
  `step2` envelope of 80 cores and 320 GiB.
- Resolve controller `41462078` published immutable snapshot
  `run-20260816T064000Z-pdxmem`, again with the completed step1 run as its
  lineage parent.
- The new run links only accepted step1 work directories, and its explicit
  Craftmake step2 dry-run completed successfully with the full resource
  contract.

- Full Craftmake controller `41462094` was submitted against this snapshot
  with the sealed Craftmake binary and catalog. Otter passed preflight and
  invoked `craftmake run`; child-task submission is pending its Slurm plan.

## Completion criteria

- Complete BS-PDX `step2`, its applicable checker phase, and methylation
  `step3` through the same immutable Craftmake snapshot or explicit
  Craftmake resume if a classified Slurm incident occurs.
- Preserve all controller and worker evidence.
- Collect verified artifacts for a toolchain comparison only after an eligible
  modern/legacy Craftmake pair is available; no incompatible snapshots will be
  compared.
