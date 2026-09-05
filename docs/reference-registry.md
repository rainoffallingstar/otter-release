# Reference registry contract

The reference registry stores immutable genome releases outside project directories. Projects lock logical IDs and release digests; site resolution supplies the actual mount path when creating a run snapshot.

## Directory layout

```text
$OTTER_REFERENCE_ROOT/
└── genomes/<reference-id>/<release>/
    ├── reference.yaml
    ├── manifest.json
    ├── checksums.sha256
    ├── fasta/
    ├── annotations/
    └── indexes/
        ├── bismark/
        ├── bowtie2/
        └── star/
```

A release directory is write-once. Corrections create a new release or manifest identity; an existing release is never overwritten in place.

## Build a release

The root CLI currently exposes `reference build` and publishes through a staging directory and atomic rename:

```bash
export OTTER_REFERENCE_ROOT=/shared/otter/references

otter reference build \
  --id hg38 \
  --release GRCh38.p14 \
  --organism 'Homo sapiens' \
  --assembly GRCh38 \
  --fasta /staging/genome.fa.gz \
  --gtf /staging/genes.gtf.gz \
  --indexes bismark,bowtie2,star
```

The command uses `samtools faidx` for the FASTA index and real index builders for selected assets. Placeholder index files are not valid registry releases. The default registry root is `$OTTER_REFERENCE_ROOT`, then `~/.otter/references` when the variable is unset.

The dedicated Gate 6 reference-build workflow is a separate operational path. Its source archives, provider checksums, compute-node build logs, resource policy, and accepted releases are documented in [Paracloud operations](gate6-paracloud-operations.md). Those historical releases are evidence; they are not implied by running the minimal command above.

## Reference metadata

Each release contains identity, assets, checksums, tools, and compatibility information:

```yaml
schema_version: otter.reference/v1
reference:
  id: hg38
  release: GRCh38.p14
  organism: Homo sapiens
  assembly: GRCh38
assets:
  fasta:
    path: fasta/genome.fa.gz
    sha256: sha256:...
  annotations:
    - id: gencode-v44
      type: gtf
      path: annotations/genes.gtf.gz
      sha256: sha256:...
  indexes:
    - type: bismark
      path: indexes/bismark
      reference_fasta_sha256: sha256:...
      tool: bismark
      tool_version: 0.24.2
compatibility:
  scenarios: [rrbs, wgbs, rnaseq, bs-pdx, rna-pdx]
```

An index declaration must identify the FASTA digest, builder/version, relevant parameters, output files, and supported scenarios. Directory indexes are expanded to file-level manifest entries.

## Project lock and resolution

`references.lock.yaml` records logical identity and manifest digest, not a machine-specific mount path:

```yaml
schema_version: otter.references.lock/v1
references:
  primary:
    id: hg38
    release: GRCh38.p14
    manifest_digest: sha256:...
```

Resolution verifies the lock, release metadata, manifest, checksums, FASTA/FAI relationship, index/reference relationship, scenario compatibility, and compute-node visibility before a production submission. The resolved absolute paths and digests are copied into `run.yaml`.

```text
selection
  → registry root from site
  → reference.yaml
  → manifest/checksum verification
  → scenario asset selection
  → absolute paths in run.yaml
  → scheduler preflight
```

Workflows must consume typed resolved assets from `run.yaml`; they must not reconstruct FASTA, GTF, or index paths from strings.

## Run override and promotion

A reference override belongs to one new run and does not change the project lock. Promotion is an explicit, audited action:

```bash
otter reference promote runs/<run-id>/run.yaml
otter reference promote runs/<run-id>/run.yaml --confirm
```

The first command previews the lock diff. The second verifies the resolved release, atomically updates `references.lock.yaml`, and appends the old/new lock and source run ID to `.otter/reference-promotions.jsonl`.

Resume cannot switch references. Re-resolve and create a new run when the effective reference changes.

## Gate 6 boundary

The accepted Gate 6 reference evidence covers controlled acquisition, provider checksum verification, immutable publication, and compute-node visibility for the documented release set. The `mm10-canary` is a small technical reference for scheduler/index/publication checks; it is not a full-genome `mm10` or `mm38` production reference.

[Back to the documentation hub](README.md)
