# Note 2026-08-15: Gate 6 Step2 Execution and SLURM Controller Incident

> Follow-up to `docs/notes/2026-08-15-gate6-step2-submission.md`.

## What happened

- The six step2 controllers (`41458048`–`41458053`) began Bismark (RRBS) and STAR (RNA-seq/RNA-PDX) alignment. BAM outputs are being produced: RRBS projects show multiple Bismark BAMs under `work/bsmap/<species>/`; RNA projects show the STAR BAM; `human-rnaseq-SRR1039508` also produced its `expression/` matrices.
- **Incident (documented)**: `human-rnaseq-SRR018258` step2 (controller `41458051`) failed with a classified exit code 5. Its STAR `map_and_sort` task succeeded (job `41458056`), but the dependent `qualimap` (job `41458069`) and `count_expression` (job `41458068`) submissions hit a transient SLURM controller error:
  `srun: error: Unable to confirm allocation for job ...: Unexpected message received`.
  This is the same transient controller fault class previously recorded for SRA decode. The task attempt dirs contain only `srun-launch.err` and no worker result, confirming the allocation-confirmation failure rather than a workflow defect.
- **Recovery**: resume controller `41458109` was submitted for the same immutable run (`run-20260815T134119Z-vdwxod`), phase `step2`, so the cached STAR BAM is reused and only `qualimap`/`count_expression` are re-attempted.
- **Second incident (same fault class)**: `rna-pdx-SRR30880970` step2 (controller `41458053`) also failed with exit code 5. Its `map_and_sort/species=hg38` and `qualimap/species=hg38` succeeded, but `map_and_sort/species=mm10` (job `41458071`) failed with the same transient `srun: Unable to confirm allocation ... Unexpected message received` error. Resume controller `41458152` was submitted for `run-20260815T132155Z-vhesje` phase `step2`, reusing the cached hg38 BAM.

## Running state

| Item | Status |
|---|---|
| step2 controllers (5 remaining) | running (Bismark/STAR alignment) |
| SRR018258 step2 resume | running |
| bs-pdx step1 (`41457463`) | running (~4.5 h), raw FastQC under node contention; autonomous watcher active |
