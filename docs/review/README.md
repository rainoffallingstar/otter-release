# Release Review Workflow

## Purpose

`docs/review/` stores the human-readable release gate checklist and the generated evidence that proves the minimal release regressions still pass.

## Submodule review program

The repository contains 9 Git submodules. `bamdriver-go` has completed a systematic review and remediation; the remaining 8 submodules are scheduled for staged review.

- Plan: [`submodule_review_plan_2026-07-21.md`](submodule_review_plan_2026-07-21.md)
- Wave 1 remediation: [`wave1_remediation_plan_2026-07-21.md`](wave1_remediation_plan_2026-07-21.md)
- Wave 1 reports:
  - [`xenofilter-go`](submodules/xenofilter-go_review_2026-07-21.md)
  - [`Paireads`](submodules/Paireads_review_2026-07-21.md)
- Order: BAM data chain → scientific outputs → runtime/orchestration → cross-repository integration
- Completion requires code review, regression tests, external compatibility evidence, security/reliability checks, and main-repository integration evidence.

## Minimal release evidence

Public releases now require two committed dry-run records:

1. RRBS local dry-run
2. RNASEQ slurm dry-run

Generate them with:

```bash
bash scripts/capture_release_evidence.sh --slurm-partition <partition>
```

Artifacts written by the script:

- `docs/review/release_evidence_<YYYY-MM-DD>.md`
- `docs/review/release_evidence_<YYYY-MM-DD>/rrbs_local_dry_run.log`
- `docs/review/release_evidence_<YYYY-MM-DD>/rnaseq_slurm_dry_run.log`
- `docs/review/release_evidence_latest.md`
- `docs/review/release_evidence_latest.env`

## Verification

Before tagging a release, verify the committed evidence bundle:

```bash
bash scripts/verify_release_evidence.sh
```

The manual release script always runs this verification and will fail if the latest evidence manifest is missing or either dry-run is not `PASS`.

The GitHub release workflow verifies the committed bundle only when `docs/review/release_evidence_latest.env` is present in the repository. If no bundle has been committed yet, CI logs a warning and skips this gate instead of trying to regenerate SLURM-based evidence on `ubuntu-latest`.
