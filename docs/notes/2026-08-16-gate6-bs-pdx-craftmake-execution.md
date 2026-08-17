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

## Long-running execution checkpoint

- Craftmake submitted the two real `BeaverPDX/step2/map_and_sort` tasks:
  Slurm `41462104` for graft `hg38` on `f0802`, and Slurm `41462105` for host
  `mm10` on `f0904`.
- Both workers remain `RUNNING` with `0:0` exit status after approximately two
  hours. Their active Bowtie2 alignment processes and increasing disk writes
  confirm that this is live dual-reference Bismark computation, not a stalled
  controller or a failed publication step.
- The combined BS-PDX mapping workspace is approximately `985 GiB`; peak RSS
  remains below the 320 GiB envelope per allocation and the shared filesystem
  has sufficient free capacity. No recovery action is justified while the
  workers continue to make progress.
- The accepted modern snapshot fixes `workflow.toolchain: modern`. A
  legacy-equivalent run must therefore be resolved later from a distinct
  canonical project configuration; the toolchain cannot be overridden at run
  invocation. It will use Craftmake with the same inputs, references, site,
  and phase resource contract before immutable artifact comparison.

## Classified hg38 mapping timeout

- Host `mm10` mapper `41462105` completed successfully with exit status `0:0`
  in `07:47:20`, publishing a `7,398,668,009`-byte Bismark paired-end BAM and
  its mapping report.
- Graft `hg38` mapper `41462104` remained active without a final BAM until its
  fixed `08:00:00` allocation expired. Slurm classified it as `TIMEOUT` after
  `08:00:22`, with exit status `0:0`; there was no OOM or application exit
  failure evidence.
- Because the dual-reference `map_and_sort` dependency was incomplete, the
  outer Otter/Craftmake controller `41462094` ended `FAILED 5:0` after
  `08:03:04`. This is a classified resource-time incident, not a result from
  the modern production pipeline.
- The immutable `run-20260816T064000Z-pdxmem` snapshot and its work products
  remain untouched. Recovery requires a new canonical configuration with an
  extended `step2` time envelope and a fresh immutable Otter snapshot.
- Craftmake cache identity includes the resolved configuration fingerprint and
  its SQLite state is scoped to the run snapshot. Therefore an extended-time
  snapshot cannot safely resume the failed snapshot's state or adopt the mm10
  output as a cross-snapshot cache hit. The recovery run will reuse only the
  accepted step1 links and will rerun both step2 species; the successful mm10
  BAM remains incident evidence, not an accepted recovery input.
- The approved recovery changes only the canonical `step2.time` envelope from
  `08:00:00` to `12:00:00`. It retains the validated 80-core, 320-GiB,
  `amd_512` allocation contract, references, inputs, and modern toolchain.
  Otter will resolve a new snapshot with the accepted step1 run as its lineage
  parent; the failed mapping snapshot is preserved solely as incident evidence.
- The remote canonical project was updated with a guarded, atomic YAML rewrite,
  then passed `otter config validate`. Its revised contract is exactly
  `80 cores / 320 GiB / 12:00:00 / amd_512`; the 1,075-byte file passed an
  ASCII-safe serialization check before a new snapshot was requested.
- Preflight controller `41466787` stopped before creating a snapshot because
  its resolve command incorrectly used a nonexistent comparison-local
  `references` directory. The accepted snapshot identifies the valid shared
  reference root as `/public3/home/scg9946/otter-gate6/references`; the retry
  will use that exact path. No workflow task or recovery snapshot was created
  by the failed preflight.
- Retried compute-node controller `41466815` completed successfully and
  published `run-20260817T001228Z-vclyug`. The snapshot records the accepted
  step1 parent and `80 cores / 320 GiB / 12:00:00 / amd_512` for `step2`.
  Its Craftmake dry-run passed, planning one 80-core allocation per species
  and two exclusive 40-core, 160-GiB workers within each allocation.
- Production recovery controller `41466855` started Otter-to-Craftmake step2
  on the new snapshot. Craftmake submitted mapper `41466864` for hg38 and
  `41466865` for mm10, each under the approved 80-core, 320-GiB, 12-hour
  allocation contract. Both began as explicit cache misses because the new
  snapshot has no accepted step2 attempt; this confirms the failed run's mm10
  BAM was not silently promoted across the immutable snapshot boundary.
- Initial scheduler sampling preceded worker startup and showed no Bowtie2
  children. A direct worker-node inspection then confirmed the hg38 Bismark
  parent process running with the resolved reference index, paired trimmed
  reads, and run-local temporary directory. The recovery mappers are therefore
  live; early empty output directories do not indicate a startup failure.
- After `06:44` of worker runtime, hg38 retained 16 active Bowtie2 children
  while mm10 declined to 12, and the combined `bsmap` workspace fell from
  roughly 1,014 GiB to 897 GiB. The coordinated sub-aligner reduction and
  intermediate-file reclamation indicate mm10 is progressing through Bismark
  direction completion rather than stalling. Both Slurm workers remain
  `RUNNING` with `0:0`; no final BAM has yet been published.
- Recovery mm10 mapper `41466865` completed successfully with `0:0` after
  `07:36:59`, publishing the run-local `7,398,668,004`-byte paired-end BAM
  and its Bismark report. This independently recomputed output belongs to the
  12-hour immutable recovery snapshot; it is not a promotion of the prior
  timed-out snapshot's mm10 artifact.
- At the same checkpoint, hg38 mapper `41466864` remained `RUNNING` with
  `0:0`, 16 active Bowtie2 children, and more than three hours of its approved
  allocation remaining. The combined mapping workspace had reduced to
  approximately 545 GiB as mm10 temporary files were reclaimed.
- Craftmake accepted the completed mm10 branch and submitted downstream work
  under its state-managed controller; the newly observed child submission
  `41494525` is running. The outer Otter controller `41466855` remains
  `RUNNING`; final step2 acceptance still depends on the hg38 mapper and all
  required downstream tasks.
- The recovery work tree now contains the standardized mm10 alignment pair
  `bsmap/SRR23802966_mm10.bam` (`5,571,590,889` bytes) and its
  `2,171,392`-byte BAI. Craftmake submitted `41494525` as the mm10 Qualimap
  task, consuming that run-local BAM under a `2 CPU / 40 GiB` contract. It
  remains `RUNNING` with `0:0`; hg38 mapper `41466864` remains the active
  mapping critical path with 16 Bowtie2 children.
- hg38 mapper `41466864` subsequently exited its Bowtie2 direction workers
  and reclaimed the mapping workspace from approximately 546 GiB to 63 GiB,
  then began materializing its Bismark BAM. The file grew from roughly 3.88 GiB
  to 5.57 GiB during direct sampling while the Slurm job remained
  `RUNNING 0:0`; it is therefore a live output write, not yet an accepted
  mapper completion or a valid downstream input.

## Legacy comparison preparation

- A read-only audit of the Gate 6 comparison root found no existing
  `legacy-equivalent` canonical project or immutable legacy snapshot for
  `SRR23802966`. The modern BS-PDX project fixes the required comparison
  contract: the same paired SRA inputs, `hg38@GRCh38-gencode-v44` graft and
  `mm10@GRCm38-gencode-M25` host references, Craftmake/Slurm executor and
  `paracloud-gate6` site, and the established phase resource envelopes.
- The eventual legacy run must therefore be a separate canonical project with
  `workflow.toolchain: legacy-equivalent`, preserving this input, reference,
  execution, site, and resource contract. It will receive a fresh immutable
  Otter snapshot and will not alter or reuse the in-flight modern snapshot's
  state.
- A separate `bs-pdx-SRR23802966-legacy-equivalent` canonical project was
  created with copied input and lock files plus symlinked shared workflow,
  rules, environment, schema, and data assets. Its only semantic project
  changes are the distinct project identity/description and
  `workflow.toolchain: legacy-equivalent`; the 1,122-byte YAML passed the
  ASCII serialization check.
- The first validation command used Otter's unsupported `--project` flag and
  stopped before validation or snapshot resolution. This is a CLI invocation
  incident only: no legacy snapshot, Craftmake plan, or Slurm production task
  was created. Validation will retry with the supported `--config` flag.
- Retrying with `otter config validate --config <legacy-project>/project.yaml`
  passed. Otter identifies the project as
  `bs-pdx-SRR23802966-legacy-equivalent (bs-pdx)` and confirms
  `workflow.toolchain: legacy-equivalent`; each shared input/workflow asset
  remains a project-local symlink to the matching modern comparison asset.
- Compute-visible resolve controller `41503216` failed with `1:0` before
  publishing a legacy snapshot. The resolver attempted to hash the
  project-local `environments` asset and rejected the directory symlink as an
  unsupported file read. This is an asset-layout failure, not an input,
  reference, resource, or toolchain-contract failure; no Craftmake plan or
  production task was submitted.
- Recovery will materialize independent regular asset directories in the
  legacy project from the same modern comparison inputs before retrying the
  fresh immutable resolve. Their contents and the legacy canonical contract
  will remain unchanged.
- The restricted shell operation intended to replace the five legacy asset
  symlinks was rejected before execution, leaving the project unchanged. A
  subsequent guarded operation asserted each link type, replaced it with the
  corresponding ordinary empty directory, and reran canonical validation
  successfully. The legacy project now matches the modern project's asset
  directory layout while preserving its distinct legacy-equivalent toolchain
  declaration and all comparison inputs.
- Retry resolve controller `41503802` then failed immediately with `1:0`
  before snapshot publication because the legacy project's now-regular empty
  `data/` directory did not contain the relative FASTQ paths declared by its
  copied sample manifest. This confirms the resolver correctly binds inputs to
  the legacy project root; no references, resources, Craftmake plan, or Slurm
  scientific task were reached.
- Recovery will read the accepted modern immutable snapshot's input identities
  and create matching legacy-project input links before a new resolve. It will
  preserve the exact paired files and provenance rather than copying or
  regenerating FASTQ data.
- Direct inspection of the modern snapshot's declared input paths confirmed
  that both source FASTQs remain regular files: R1 is `28,258,736,221` bytes
  and R2 is `29,595,686,034` bytes. The earlier root-wide search yielded no
  output because of its query/output behavior, not because the bound inputs
  were absent.
- The legacy `data/` directory now links exactly those two verified source
  FASTQs at the copied relative manifest paths. `otter config validate` again
  passed, preserving the intended same-input comparison contract without
  copying, modifying, or reacquiring read data.

## Completion criteria

- Complete BS-PDX `step2`, its applicable checker phase, and methylation
  `step3` through the same immutable Craftmake snapshot or explicit
  Craftmake resume if a classified Slurm incident occurs.
- Preserve all controller and worker evidence.
- Collect verified artifacts for a toolchain comparison only after an eligible
  modern/legacy Craftmake pair is available; no incompatible snapshots will be
  compared.
