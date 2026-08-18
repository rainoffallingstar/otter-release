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
- Third compute-visible resolve controller `41504431` completed with `0:0` and
  published the legacy immutable snapshot
  `run-20260817T102532Z-emvbei`. It is an independent run with no modern
  parent state; the next step is to inspect its resolved comparison contract
  and run a Craftmake `step1` dry-run before any legacy scientific task is
  submitted.
- Legacy step1 Craftmake dry-run controller `41504704` completed with `0:0`.
  It produced the full planned Trim Galore task graph, including the paired
  verified inputs and the resolved 6-core step1 contract; its `53,851`-byte
  stdout contains the plan and its stderr is empty. No scientific task was
  submitted by this dry-run.
- Legacy production step1 controller `41504897` then started through
  Otter-to-Craftmake. Its first two Craftmake child submissions, `41504960`
  and `41504961`, are `RUNNING` and were explicitly evaluated as cache misses
  in the independent legacy state. In parallel, modern mm10 Qualimap
  `41494525` completed with `0:0`, while modern hg38 mapper `41466864` began
  merging its Bismark BAM into the standardized alignment through run-local
  sorted BAM chunks.
- Modern hg38 mapper `41466864` completed with `0:0` after `10:23:15`,
  publishing the standardized `29,303,457,511`-byte BAM and its
  `8,514,968`-byte BAI. Craftmake accepted that mapper completion and submitted
  independent hg38 `collect_gc_bias` and `qualimap` tasks (`41505526` and
  `41505527`), each `RUNNING` under their catalog-declared resource contracts.
  The outer modern step2 controller remains active until these final required
  QC dependencies finish.
- Modern hg38 GC-bias task `41505526` completed successfully with `0:0`.
  Craftmake recorded its Picard `CollectGcBiasMetrics` worker as `succeeded`
  with exit code `0` after approximately `24:36`; only the parallel hg38
  Qualimap task remains before outer modern step2 acceptance.
- Modern hg38 Qualimap `41505527` and outer step2 controller `41466855` then
  both completed with `0:0`. The completed recovery step2 now provides
  non-empty standardized BAM/BAI pairs, Qualimap reports, and GC-bias artifacts
  for both hg38 and mm10 under the modern immutable snapshot. It is eligible
  for the Craftmake `step2-check` artifact-validation phase.
- Modern `step2-check` controller `41507248` failed with `5:0` after
  `02:02:20`. Its two Craftmake `sample_artifacts` validators, Slurm jobs
  `41507292` and `41507294`, both reached `TIMEOUT` at `02:00:04` with
  `0:0`; Craftmake classified the resulting `context canceled` events as
  retry-safe `cancellation_recovery`. Their workers had no stderr output and
  had read approximately `22.5 GB` each, consistent with SHA-256 validation
  of the large BAM inputs rather than an artifact-content rejection. The
  checker needs a fresh immutable snapshot or accepted phase retry with a
  longer `step2-check` task time contract before modern downstream work can
  continue.
- Legacy-equivalent `step1` completed successfully through the independent
  Craftmake snapshot. Outer controller `41504897`, Trim Galore `41504960`, and
  pre-trim FastQC `41504961` all completed with `0:0`; the full phase duration
  was `08:52:40`. The controller evidence records successful Trim Galore and
  FastQC task attempts, and the run-local R1/R2 trimmed FASTQs plus validation
  reads are present. Slurm reason codes observed during execution were
  transient backend state notifications only and did not change the successful
  terminal status.
- The modern canonical project was atomically updated only at
  `resources.phases.step2-check.time`, from `02:00:00` to `08:00:00`, to cover
  the two observed large-BAM SHA-256 validators. Cores (`4`), memory (`16GiB`),
  partition (`amd_512`), toolchain, inputs, references, and all other phase
  contracts remain unchanged. The first combined update-and-validate shell
  invocation did not run validation because its `$root` variable was unset;
  the atomic update had already completed and was immediately validated with
  the explicit Otter runtime path. The 1,075-byte canonical YAML passed Otter
  validation and an ASCII-byte check. A new immutable snapshot will carry this
  contract and bind only accepted modern output directories from the completed
  parent run.
- Resolve controller `41533149` completed with `0:0` and published modern
  immutable snapshot `run-20260818T020535Z-xpwyeu`, whose lineage parent is
  `run-20260817T001228Z-vclyug`. It contains the validated eight-hour
  `step2-check` envelope. No workflow task was submitted by this resolve;
  the next action is to link only accepted parent `work/` outputs into the
  new snapshot, without copying its Craftmake state or cache.
- The initial protected recovery-output binding command failed before execution
  with a Python `SyntaxError` caused by shell-to-heredoc escaping. A direct
  inspection confirmed that the new snapshot's `work/` remained empty and that
  no `state` link exists; no output, cache, or immutable snapshot content was
  altered. The binding will be retried with a simpler guarded command.
- Legacy-equivalent `step2` Craftmake dry-run completed with exit code `0`
  against `run-20260817T102532Z-emvbei`. This validates its planned dual
  reference mapping and downstream QC graph without submitting a scientific
  task; production step2 can proceed independently of modern checker recovery.
- The recovery snapshot now binds only five accepted parent work directories:
  `trim`, `QC`, `fastqc_raw`, `fastqc_clean`, and `bsmap`. Each is a direct
  symlink to the completed `run-20260817T001228Z-vclyug` run's corresponding
  work directory. The recovery snapshot has no `state/runs` link, so Craftmake
  planning, attempts, cache evaluation, and validation state remain wholly
  isolated while the accepted scientific outputs are available to the checker.
- Modern recovery `step2-check` dry-run completed with exit code `0`, confirming
  that the eight-hour validator contract and accepted artifact bindings produce
  a valid Craftmake plan. The production checker has not yet submitted a task
  in this checkpoint.
- Legacy-equivalent production `step2` controller `41533289` is running through
  Otter foreground to Craftmake. Craftmake planned six tasks, and all six
  initially evaluated as explicit cache misses in the independent legacy
  snapshot, including both species `map_and_sort` tasks. No modern outputs or
  state are used by this legacy execution.
- Modern recovery `step2-check` production controller `41533552` started
  through Otter foreground to Craftmake. Craftmake created the hg38 and mm10
  `sample_artifacts` submissions as independent cache misses and started both
  one-core validators under the new eight-hour phase time envelope. Xenofilx
  remains dependency-blocked until the two validator manifests are accepted.
  The concurrent legacy and modern controllers retain separate snapshot roots,
  Craftmake state, and cache boundaries.
- The modern hg38 and mm10 `sample_artifacts` validators both succeeded with
  exit code `0` under the eight-hour recovery envelope. Craftmake then accepted
  both manifests and started its dependency-gated Xenofilx task as Slurm
  `41533581` under the resolved `4 cores / 16 GiB / 08:00:00` contract. This
  confirms the prior two-hour failure was an insufficient validation-time
  envelope rather than invalid mapping or QC artifacts.
- Modern Xenofilx `41533581` remains `RUNNING/0:0` under the new recovery
  snapshot; no filtered BAM or validation manifest is published yet. Its
  output is the critical path for accepting modern `step2-check` and starting
  modern `step3`.
- Legacy-equivalent mapper allocations remain active under controller
  `41533289`: hg38 and mm10 each have a Craftmake-managed 80-core allocation
  with a 40-core/160-GiB mapper worker. Both began as independent cache misses
  and remain `RUNNING/0:0`; no legacy BAM has been accepted yet.
- Modern Xenofilx `41533581` ended `OUT_OF_MEMORY` after `00:41:22`, and outer
  checker controller `41533552` then ended `FAILED 5:0`. The retained worker
  diagnostic shows Xenofilx entered bisulfite-mode processing and was killed
  with exit `137` while name-sorting the `29,303,457,511`-byte hg38 BAM;
  validation then correctly rejected the absent temporary filtered BAM/BAI.
  No filtered output was published. The Craftmake result classifies the task as
  a retry-safe tool-invocation incident, while Slurm supplies the root resource
  classification. Increasing only the phase envelope would be insufficient:
  the BeaverPDX Xenofilx catalog task itself declares `4 cores / 16 GiB`.
  Its memory contract must be increased, followed by a new catalog-bound,
  immutable recovery snapshot before retrying the checker.
- The BeaverPDX `step2-check` catalog contract was increased from `16G` to
  `64G` for Xenofilx while retaining four cores and all filtering semantics.
  The increase is based on the observed OOM during name-sort of the 29.3-GB
  graft BAM and leaves room for Xenofilx's paired-BAM filtering and output
  validation. The targeted Craftmake compiler suite
  (`go -C craftmake test ./internal/compiler`) passed after this change.
- The sealed remote BeaverPDX catalog was atomically updated to the committed
  `64G` Xenofilx contract, and modern canonical
  `resources.phases.step2-check.memory` was atomically raised from `16GiB` to
  `64GiB` while retaining four cores, the eight-hour time budget, and the
  `amd_512` partition. Both files passed ASCII-safe serialization checks within
  their guarded update, and Otter accepted the revised canonical project. A new
  immutable snapshot is required because the phase resource contract and
  catalog task definition have changed.
- Resolve controller `41534436` completed with `0:0` and published immutable
  modern snapshot `run-20260818T042539Z-kncxnc`, directly lineaged to the
  accepted modern step2 run `run-20260817T001228Z-vclyug`. The snapshot records
  `step2-check` as `4 cores / 64GiB / 08:00:00 / amd_512`. No workflow task
  was submitted by resolution; only accepted parent work directories may now be
  bound, while its Craftmake state remains new and isolated.
- `run-20260818T042539Z-kncxnc` now binds the same five accepted directories
  from the completed modern step2 run: `trim`, `QC`, `fastqc_raw`,
  `fastqc_clean`, and `bsmap`. Each binding is a direct symlink to the accepted
  parent `work/` directory; the new snapshot has no `state/runs` link, leaving
  its Craftmake state, cache decisions, and attempts independent for the
  64-GiB Xenofilx retry.
- The new snapshot's Otter-to-Craftmake `step2-check` dry-run completed with
  exit code `0` against the sealed updated catalog. The 64-GiB Xenofilx task
  request is therefore accepted by the `4 cores / 64GiB / 08:00:00` immutable
  phase envelope before production submission.
- Modern 64-GiB checker production controller `41534635` was submitted through
  Otter foreground to Craftmake against `run-20260818T042539Z-kncxnc`. Its
  first scheduler observation is `RUNNING/0:0`; controller planning is pending
  in this checkpoint. Legacy-equivalent step2 controller `41533289` remains
  independently `RUNNING/0:0` with its dual-reference mapper branches.
- Modern 64-GiB recovery controller `41534635` ended `FAILED 5:0` before
  Xenofilx was eligible. Its mm10 validator succeeded, but the hg38 validator
  allocation `41534638` completed without starting the worker: retained
  `srun-launch.err` reports `Unexpected message received` and `Expired or
  invalid job 41534638`. The expected hg38 `result.json` was consequently
  absent. This is a retry-safe Slurm allocation/startup incident, not a
  validator content failure, catalog error, or another Xenofilx OOM. The
  immutable 64-GiB snapshot and its accepted output bindings remain valid;
  recovery will use Otter-to-Craftmake `--resume` on that same snapshot so
  Craftmake owns the retry without reusing failed state as cache.
- Otter-to-Craftmake resume controller `41534649` was submitted against the
  unchanged `run-20260818T042539Z-kncxnc` snapshot and started as
  `RUNNING/0:0`. It is the only recovery action for the retry-safe hg38
  validator startup incident: no resource, catalog, input, output binding, or
  immutable snapshot change accompanies this resume.
- After 59 minutes of controller runtime, `41534649` remained `RUNNING/0:0`
  in Slurm accounting but had not appended a new Craftmake controller event or
  created a retry submission beyond the failed first round. The latest durable
  controller state is therefore a control-plane stall observation, not a
  completed retry or a validator result. No workflow child job was manually
  submitted, and no configuration, catalog, snapshot, output binding, or cache
  boundary was changed while this controller remained active.
- At the same checkpoint, independent legacy-equivalent `step2` controller
  `41533289` remained `RUNNING/0:0` after more than three hours. Its terminal
  mapper and downstream QC outcomes are not yet available, so it has no
  comparison result to record.
- A subsequent state-scope inspection corrected the prior controller-log
  interpretation: Craftmake's `resume` command creates a new run lineage, so
  its events are not appended to the source
  `run-20260818T042539Z-kncxnc--step2-check/controller.jsonl`. The Otter
  resume controller `41534649` owns resumed Craftmake run
  `f7b53160-ec64-4f4d-96c3-7d526c7134dd`; it is active rather than stalled.
- In that resumed run, the missing hg38 `sample_artifacts` validator was
  retried as Slurm `41534650` and succeeded with exit `0:0`. The already
  successful mm10 validation was reused from the source run, while the
  dependency-gated Xenofilx task remained a fresh resumed-run cache miss.
  Craftmake then submitted Xenofilx `41534653`, which remains `RUNNING/0:0`
  under the corrected `4 CPU / 64 GiB` contract. Its Craftmake attempt records
  exactly four cores and `68,719,476,736` bytes, and Slurm reports `ReqMem=64G`.
  This is an Otter-to-Craftmake-owned retry; no workflow child task was
  manually submitted and no immutable contract was changed.
- Xenofilx `41534653` subsequently ended `OUT_OF_MEMORY` after `02:26:11`
  with Slurm exit `0:125`, and its resumed Craftmake run
  `f7b53160-ec64-4f4d-96c3-7d526c7134dd` then ended failed. The retained
  `slurm-step-41534653.err` records one OOM-kill event and `srun-launch.err`
  reports `task 0: Out Of Memory`; its Craftmake result is a retry-safe
  `tool_invocation` failure. No filtered BAM, BAI, or filtered validation
  manifest was published. This establishes that 64 GiB remains insufficient
  for this Xenofilx workload; a further resource-contract recovery must be
  recorded, published, and resolved as a new immutable snapshot rather than
  mutating or reusing the failed one.
- At this same checkpoint, legacy-equivalent `step2` controller `41533289`
  and its two Craftmake mapper allocations `41533294` and `41533295` continue
  `RUNNING/0:0` under their 320-GiB allocations after more than five hours.
- The approved next recovery increases only the Xenofilx task and modern
  `step2-check` phase memory contract from 64 GiB to 128 GiB. It retains four
  CPUs, the eight-hour limit, `amd_512`, all validated input bindings, and the
  existing bisulfite filtering semantics. Because both prior 16-GiB and 64-GiB
  allocations were Slurm OOM-killed, no lower value has supporting evidence.
  The catalog update, canonical configuration update, and a new
  Otter-resolved immutable snapshot will be completed before the next
  Craftmake-owned production attempt.
- Craftmake catalog commit `1b44dfc` raises the BeaverPDX Xenofilx task to
  `4 CPU / 128 GiB`; its targeted compiler suite passed with
  `go -C craftmake test ./internal/compiler`. Root commit `cd29458` advances
  the pinned catalog pointer. The sealed remote catalog and modern canonical
  `step2-check` envelope were atomically updated to the same 128-GiB contract,
  with ASCII-safe serialization checks, and `otter config validate` passed.
- Resolve controller `41537335` completed `0:0`, publishing immutable modern
  snapshot `run-20260818T101652Z-jpsrgh` with accepted modern step2 run
  `run-20260817T001228Z-vclyug` as its direct parent. Its snapshot fixes
  Craftmake/Slurm execution and `4 CPU / 128GiB / 08:00:00 / amd_512` for
  `step2-check`. The new run binds only the accepted parent `trim`, `QC`,
  `fastqc_raw`, `fastqc_clean`, and `bsmap` work directories. It has no
  `state/runs` entry or state link, so its Craftmake state and cache remain
  independent from every failed checker attempt.
- The snapshot's Otter-to-Craftmake `step2-check` dry-run completed with exit
  `0` against the sealed 128-GiB catalog. The new independent checker state is
  therefore contract-valid before any production workflow child task is
  submitted.
- Production controller `41537404` then started Otter foreground to Craftmake
  against `run-20260818T101652Z-jpsrgh`. Craftmake created new-snapshot cache
  misses for both artifact validators and submitted them as Slurm `41537414`
  (hg38) and `41537415` (mm10); the dependency-gated 128-GiB Xenofilx task has
  not yet been submitted. These are new independent state attempts, not cache
  reuse from failed checker snapshots.
- Legacy-equivalent mm10 mapper allocation `41533295` completed `0:0` after
  `07:45:55`. Craftmake accepted that mapping branch and began its required
  mm10 GC-bias and Qualimap downstream work. The legacy hg38 mapper allocation
  `41533294` remains the active `step2` critical path.
- Both modern artifact validators then succeeded (`41537414` hg38 and
  `41537415` mm10). Craftmake accepted their manifests and dependency-submitted
  Xenofilx `41537429`; its attempt requests four cores and
  `137,438,953,472` bytes, while Slurm reports the active allocation at 128G.
  This is the first 128-GiB production filter attempt, wholly planned and
  submitted through the Otter-to-Craftmake execution path.
- At the next scheduler checkpoint, Xenofilx `41537429` remained
  `RUNNING/0:0` at 128G. In the independent legacy run, mm10 Qualimap
  `41537371` completed successfully with exit `0:0` after `04:38`; legacy
  hg38 mapper `41533294` remains the only active `step2` critical path.

## Completion criteria

- Complete BS-PDX `step2`, its applicable checker phase, and methylation
  `step3` through the same immutable Craftmake snapshot or explicit
  Craftmake resume if a classified Slurm incident occurs.
- Preserve all controller and worker evidence.
- Collect verified artifacts for a toolchain comparison only after an eligible
  modern/legacy Craftmake pair is available; no incompatible snapshots will be
  compared.
