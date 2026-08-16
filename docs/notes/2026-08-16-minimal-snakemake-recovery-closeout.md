# Note 2026-08-16: Minimal Snakemake Recovery Closeout

## Scope

The seven-input production comparison uses Craftmake only. Snakemake remains an
explicit compatibility path, so this closeout is deliberately isolated from the
comparison cells.

## Planned real-workflow evidence

- Resolve a new, immutable `otter.run/v1` for the completed
  `mouse-rrbs-SRR10025242` project with `--executor snakemake`.
- Submit a full explicit Snakemake compatibility run through Otter on Slurm.
- Cancel the controller in a controlled manner while the real workflow is
  active, then run the same immutable snapshot with `--resume`.
- Require successful artifact publication and `otter artifact verify` after
  recovery. Existing Craftmake snapshots and published artifacts remain
  untouched.

## First execution result

- Controller `41460152` reached Otter's real Snakemake compatibility path but
  stopped before workflow submission with `snakemake not found in PATH`.
- No Snakemake child job, workflow output, snapshot mutation, or artifact
  publication was produced. The failure is therefore an environment-preflight
  blocker, not evidence of a workflow or recovery failure.
- The next attempt must provide a sealed, compute-visible Snakemake executable
  through the controller environment before the controlled interruption is
  attempted.

## Asset staging recovery

- The failed snapshot `run-20260816T022709Z-mzmkvr` is retained as the
  missing-asset diagnostic record and will not be reused.
- The isolated project now contains the repository's version-controlled
  `inst/snakefiles` (at project root) and `inst/rules` assets, staged read-only.
  These are the assets required by the explicit Snakemake compatibility engine.
- A new snapshot must be resolved after staging so the root `*.snakemake` files
  and `rules/` directory are covered by the immutable workflow-asset digest.

## Controlled interruption

- Asset-sealed snapshot `run-20260816T024341Z-cijlua` launched controller
  `41460256`; its real Snakemake step1 child `41460261` entered `RUNNING`.
- The controller was then deliberately cancelled. Slurm accounting records
  `41460256` as `CANCELLED`, while `41460261` remained running at the time of
  cancellation. This is the required real controller-interruption state.
- The next controller will run the same immutable snapshot with `--resume` to
  reconcile that retained worker state and complete publication.

## Resume lock repair

- The recovery attempt exposed a stale `.snakemake` lock after controller
  cancellation. Otter's explicit Snakemake resume path now calls
  `snakemake --unlock` from the project directory before it re-enters the
  workflow manager.
- The unlock is limited to actual resume runs; dry-runs and normal Snakemake
  executions retain their prior behavior.
- `go test ./cmd` passes for the repair before deploying the closeout binary.

## Reasoning

This supplies the required genuine Snakemake controller interruption/resume
and publication evidence without using Snakemake as a production comparison
executor or mutating any completed Craftmake run.
