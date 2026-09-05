# Release readiness

This page is the release decision checklist for the current documentation and evidence boundary. It intentionally separates “the code builds,” “the bounded evidence is accepted,” and “the product is ready for a formal public release.”

## Current position

The project is in a **dual-track, conditional release-readiness** state:

- the root CLI, canonical v1 configuration path, workflow catalog, artifact contracts, and component documentation are available;
- bounded Gate 6 evidence is accepted for the documented Craftmake–Snakemake comparison, corrected classification controls, and Methx/Methrix parity;
- legacy projects still require explicit Snakemake compatibility routing;
- production-scale qualification, fresh representative matrices, complete WGBS qualification, and universal Snakemake retirement are not current claims.

## Readiness checklist

### Documentation and user entry points

- [x] Root README and Chinese README explain value, first use, architecture, and limitations.
- [x] `docs/README.md` indexes current manuals, contracts, build instructions, and evidence.
- [x] Each component README explains inputs, outputs, installation, an example, limitations, and tests.
- [x] Current documentation uses current product names; historical evidence keeps historical names.
- [ ] Rendered SVG preview has been checked on GitHub and a narrow viewport; local XML and safety validation is complete.

### Runtime and compatibility

- [x] Legacy `otter.yaml` examples explicitly select `--executor snakemake`.
- [x] Canonical v1 examples resolve an immutable `run.yaml` before Craftmake execution.
- [x] Craftmake does not silently fall back to Snakemake.
- [x] Site/backend selection and reference validation have fail-closed rules.
- [ ] A formal release build has been produced from the final versioned source and checked on the supported environments.

### Evidence and scientific claims

- [x] Accepted bounded executor-comparison evidence is indexed.
- [x] Corrected Gate A–D classification evidence is indexed.
- [x] Methx/Methrix parity and the R conversion path are documented.
- [x] Deferred work is visible and excluded from current claims.
- [ ] Any new release-specific acceptance requirement has an owner, evidence path, and decision record.

## Deferred, not silently missing

The following are deliberately deferred extensions, not undocumented completed work:

- fresh seven-input legacy-equivalent matrix;
- representative `20 samples × 3 repeats` matrix;
- production-scale throughput and scheduler pressure;
- WGBS `SRR6373947` requalification;
- additional Snakemake interruption/recovery studies.

A release may proceed only if the release owner explicitly accepts these limitations. Otherwise, each item needs its own plan and evidence entry; none should be inferred from local tests or README completeness.

## Recommended final release sequence

1. Correct any remaining cross-language command or link inconsistency.
2. Run root tests, component CI, and the final release build.
3. Verify README links, SVG metadata, and rendered images.
4. Freeze the evidence register and write a release decision record.
5. Publish with explicit `dual-track` wording and the deferred-work list.
6. Treat any future Snakemake retirement as a separate migration decision.

[Back to the documentation hub](README.md)
