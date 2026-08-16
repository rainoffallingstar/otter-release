# Note 2026-08-16: Gate 6 RNA Step2-Check Submission

## Purpose

After the three standalone RNA-seq projects completed step2, submit their
`step2-check` phase through Otter with Craftmake. This phase supplies the
remaining comparison evidence: expression-matrix construction and RNA splicing
analysis.

## Submission boundary

- Controllers will invoke `otter run --foreground --executor craftmake --phase step2-check`.
- Each controller uses its existing immutable `otter.run/v1` `run.yaml` and the
deployed workflow catalog.
- Controllers are Slurm wrappers only; no workflow task is submitted directly
with `sbatch`.
- Target projects: `human-rnaseq-SRR1039508`, `human-rnaseq-SRR018258`, and
  `mouse-rnaseq-SRR037954`.

## Expected evidence

Each project should complete four tasks: `construct_expression_matrix`,
`rnaseq_splicing`, and their declared comparisons/summary tasks as defined by
the deployed `BeaverRNA` `step2-check` workflow.
