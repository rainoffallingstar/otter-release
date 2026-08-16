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

## Reasoning

This supplies the required genuine Snakemake controller interruption/resume
and publication evidence without using Snakemake as a production comparison
executor or mutating any completed Craftmake run.
