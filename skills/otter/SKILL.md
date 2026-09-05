---
name: otter
description: "Use when working in the otter bioinformatics workflow repositories: the root CLI or its independent submodules craftmake, enva, fastqcx, xenofilx, pairbam, seq2mat, matsrun, qctb, methx, and bamdriver. Covers repository boundaries, toolchain selection, documentation maintenance, CLI-contract verification, and evidence-aware release language."
---

# otter repository skill

Use this skill for code, documentation, release, and validation work in the `otter` root repository or one of its direct submodules.

## First pass

1. Identify the active repository from the working directory, top-level files, nearest README, and git remote.
2. Treat every submodule as an independent repository. A parent gitlink update is required when a submodule change is intended to affect the parent checkout.
3. Read the owning README, CLI entrypoint, package manifest, and relevant tests before changing a user-facing contract.
4. Keep current product names and external scientific standards distinct. See [repo-map.md](references/repo-map.md) for historical aliases.
5. For documentation work, read [documentation.md](references/documentation.md) before restructuring an entrypoint.

## Product model

```text
otter → craftmake → enva → operators → bamdriver
```

- `otter` owns project initialization, input/configuration, run orchestration, task control, embedded workflow assets, and user documentation.
- `craftmake` owns native workflow compilation, Local/SLURM execution, SQLite state, recovery, cancellation, and reports.
- `enva` owns rattler-first environment lifecycle and explicit compatibility with conda, mamba, and micromamba.
- `fastqcx`, `xenofilx`, `pairbam`, `seq2mat`, `matsrun`, `qctb`, and `methx` are focused operator CLIs.
- `bamdriver` owns shared BAM/BGZF primitives used by BAM-consuming operators.

The runtime is dual-track. Existing production workflows use Snakemake compatibility assets; Craftmake integration is still being validated. Do not write documentation that implies complete Snakemake replacement unless the source and accepted evidence explicitly support that claim.

## Toolchain selection

- Go work: `conda activate go-env` for the root, `craftmake`, `pairbam`, `bamdriver`, `matsrun`, `seq2mat`, and `xenofilx`.
- Rust work: `conda activate rust_build` for `enva`, `fastqcx`, `methx`, and `qctb`.
- Load focused commands from [build-and-test.md](references/build-and-test.md) before building or testing.

## Documentation mode

When asked to redesign or beautify a README, use README mode: improve the whole information architecture, not only the decoration. Apply this order unless the repository has a stronger need:

```text
Value → Proof → Mechanism → First use → Detail
```

The first screen should answer what the repository is, who benefits, and where to go next. Prefer real CLI examples, output contracts, tests, benchmarks, and repository-native diagrams over generic decoration. Do not create a hero image merely to fill space; if a visual asset is useful, keep commands and essential instructions in Markdown.

For the root repository:

- Update `README.md` and `README_zh.md` together when user-facing setup or product language changes.
- Link current docs through `docs/README.md` and the task-oriented manual through `docs/manual/README.md`.
- Keep `docs/archive/`, `docs/review/`, and dated `docs/notes/` as historical evidence unless a specific correction is requested.
- Preserve visible limitations around Gate 6, WGBS, production scale, deferred matrices, and the Snakemake/Craftmake boundary.

For a submodule:

- Make its README independently useful: one-sentence value, input/output contract, install, minimal example, limits, tests, and repository link.
- Do not import root-only assumptions into a standalone operator README.
- Update root integration docs only when the submodule interface or workflow contract changes.

## Validation

Use the smallest relevant validation first:

1. Markdown links, command names, and paths match source.
2. README image references and SVG metadata pass the README audit when the audit script is available.
3. Go: `gofmt`, focused tests, `go test ./...`, and `go vet ./...` as appropriate.
4. Rust: `cargo fmt --all -- --check`, `cargo clippy --all-targets --all-features --locked -- -D warnings`, and `cargo test --all-targets --all-features --locked` when the repository workflow uses them.
5. Run the linter on edited files after substantive changes.

Never claim a command was tested when only the documentation was edited. Never change a submodule and silently leave the parent gitlink stale.

## Safe operating rules

- Preserve unrelated work and start with read-only inspection.
- Do not rewrite historical evidence to apply current branding.
- Do not commit, push, tag, publish assets, or open a PR without explicit authorization.
- Do not add secrets, tokens, private paths, or undocumented local-machine assumptions to examples.
