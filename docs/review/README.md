# Release Review Workflow

## Purpose

`docs/review/` stores the human-readable release gate checklist and the generated evidence that proves the minimal release regressions still pass.

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

The manual release script and `.github/workflows/release.yml` both run this verification and will fail if the latest evidence manifest is missing or either dry-run is not `PASS`.
