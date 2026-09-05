# Documentation maintenance guide

## Information architecture

Use the sequence below for project homepages and focused tool READMEs:

```text
Value → Proof → Mechanism → First use → Detail
```

- **Value**: describe the user outcome in plain language.
- **Proof**: show a real output, tested capability, accepted evidence, or meaningful limitation.
- **Mechanism**: explain where the tool fits and what it consumes/produces.
- **First use**: provide one copyable path that can succeed end to end.
- **Detail**: link to advanced options, schemas, benchmarks, and development instructions.

Do not open with a long architecture explanation, a contributor guide, or a table of every flag. Keep limitations close to the claim they qualify.

## Root documentation layers

- Root README: product value, stack, release boundary, minimal install/run path, and links.
- `docs/README.md`: current documentation map and maintenance policy.
- `docs/manual/`: task-oriented tutorial for users.
- `docs/*.md`: current contracts, architecture, operations, release readiness, and evidence indexes.
- `docs/active_context.md`: concise current-state index; detailed experiment records belong in linked reports.
- `docs/review/`, `docs/archive/`, `docs/notes/`: historical evidence, retained without branding rewrites.
- `skills/otter/`: agent operating rules and validation matrix.

## Submodule README minimum

Every independent component README should answer:

1. What does this tool do?
2. What input does it accept?
3. What output does it write?
4. What is the shortest install path?
5. What is one real command example?
6. What is explicitly unsupported or compatibility-only?
7. How is it tested and where is the repository?

Keep package-specific details in the submodule. Explain root integration only in a short “Where it fits” section.

## Terminology

Use current names in new documentation:

- `otter` instead of `xdxtools`;
- `fastqcx` instead of `fastqc-rs`;
- `xenofilx` instead of `xenofilter-go`;
- `pairbam` instead of `Paireads`;
- `seq2mat` instead of `htseq2matrix-go`;
- `matsrun` instead of `gomats`;
- `methx` instead of `methrix-cli`;
- `bamdriver` instead of `bamdriver-go`.

Retain FastQC, MultiQC, FASTQ, Bismark, Methrix, HTSeq, rMATS, BAM, BGZF, HDF5, and Snakemake because they are external standards, tools, formats, or scientific concepts.

## Claim discipline

- Distinguish implemented, tested, accepted, bounded, deferred, and unsupported.
- Do not turn a benchmark into a general performance claim.
- Do not call a custom Methx HDF5 file a native Methrix HDF5 object; link to the explicit R exporter when native loading is needed.
- Do not describe Craftmake as the only production executor while Snakemake compatibility remains active.
- Do not use future or historical dates as evidence of current support without linking the source record.
