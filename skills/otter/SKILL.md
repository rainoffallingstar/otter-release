---
name: otter
description: "Use when working in the otter bioinformatics workflow repositories: the otter root CLI or its direct submodules enva, pairbam, bamdriver, fastqcx, matsrun, seq2mat, methx, qctb, and xenofilx. Helps Codex choose the right Go or Rust toolchain, respect submodule boundaries, run the right build and test commands, and understand how each repo fits into the workflow stack."
---

# otter

Use this skill when the current workspace is the `otter` root repository or one of its direct submodules listed in [references/repo-map.md](references/repo-map.md).

## First pass

- Identify the active repo from `pwd`, top-level files, and the nearest `README.md`.
- Treat each submodule as an independent module unless the task explicitly spans multiple repos.
- Keep stable CLI flags, workflow file names, and user-facing outputs unless the task explicitly changes them.
- When editing user-facing setup docs, update both `README.md` and `README_zh.md` in the root repo when both exist.

## Toolchain selection

- Use `conda activate go-env` for `otter`, `pairbam`, `bamdriver`, `matsrun`, `seq2mat`, and `xenofilx`.
- Use `conda activate rust_build` for `enva`, `fastqcx`, `methx`, and `qctb`.
- Load the matching commands from [references/build-and-test.md](references/build-and-test.md) before building or testing.

## Working rules

1. Detect the target repo, then read only the matching section from the references.
2. Prefer the local repo's `README.md`, entrypoint files, and existing tests over assumptions.
3. Validate with the smallest relevant command set first: formatter, build, unit tests, then heavier integration flows.
4. If a change touches the root repo plus a submodule, call out the coupling explicitly in the final response.

## Repo guardrails

- `otter` root owns workflow orchestration, embedded assets under `inst/`, and end-user CLI flows.
- `bamdriver` is shared library code for BAM and BGZF I/O used by `xenofilx` and `pairbam`; exported API changes can cascade.
- `enva` owns environment creation, adoption, and execution logic; do not duplicate that behavior in `otter` unless the task is integration glue.
- `matsrun`, `seq2mat`, `methx`, `qctb`, `xenofilx`, `pairbam`, and `fastqcx` are focused CLIs; keep interfaces explicit and narrow.
- `fastqcx` is a standalone Rust FASTQ QC tool; preserve its crate-style CLI and report outputs unless the task explicitly targets behavior changes there.

## References

- Repo map: [references/repo-map.md](references/repo-map.md)
- Build and test matrix: [references/build-and-test.md](references/build-and-test.md)
