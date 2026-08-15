# Gate 6 Paracloud Operations Record

> Updated: 2026-08-15. This record distinguishes accepted production evidence from future workflow policy.

## Connection Boundary

Paracloud is accessed through the configured `agentsshcli` connection named `paracloud`:

```bash
agentsshcli exec --no-cache --connection "paracloud" --command "<read-only command>"
agentsshcli exec --no-cache --connection "paracloud" --command-file /absolute/path/to/script.sh
agentsshcli upload --no-cache --connection "paracloud" --local /absolute/local/path --remote /absolute/remote/path
agentsshcli download --no-cache --connection "paracloud" --remote /absolute/remote/path --local /absolute/local/path
```

Use `--command-file` when a command contains shell variables, nested quotes, loops, or heredocs; this prevents local-shell interpolation from changing remote paths. The remote Gate 6 base directory is:

```text
/public3/home/scg9946/otter-gate6
```

The login node is only a submission and controller boundary. Genome building, release verification, visibility tests, SRA conversion, and scientific workflows run in Slurm allocations on compute nodes.

## RNA Executor-Parity r30 Record (2026-08-08)

- Immutable release `gate6-20260808T220000Z-rna-normalized-schema-r30` was built by compute job `41299761` (`COMPLETED 0:0`), independently verified with `sha256sum -c`, binary help checks, source-overlay checksum `f2319c7f0f9fea822f81b0c167c3a3c04459fb674fc0e52714749b173894190d`, and sealed permissions. `binaries/current` now points to r30.
- r30 fixes the Snakemake RNA normalized-matrix declaration: `expression-normalized-matrix` now uses `otter.expression-normalized-matrix/v1` while retaining the shared `expression-count-matrix/v1` scientific comparator. The release property `release.snakemake_normalized_matrix_schema` records this contract.
- r29 remains immutable historical evidence only: its paired controllers `41296545`/`41296546` completed all RNA phases, and both manifests verified, but their comparison cannot be accepted because the old normalized-matrix schemas differed.
- Fresh r30 paired attempts are retained as invalid operational evidence, not parity evidence. `41301957`/`41301958` exposed permissions on copied mutable `.snakemake` state; `41303148`/`41303149` retained a Craftmake Slurm allocation-confirmation failure; `41304469`/`41304470` established that mutable `.snakemake` content is included in `workflow_assets` revalidation and consequently causes a post-step1 immutable-snapshot drift failure. No executor substitution, fallback, snapshot mutation, or r29 artifact reuse occurred.
- The next release must correct the shared resolver asset selection so mutable `.snakemake` runtime state is excluded from the immutable workflow-asset digest. Only then may a fresh parity-identical pair be resolved and run through all RNA phases, `artifact verify`, and `artifact compare`.

## RNA Executor-Parity r31 Record (2026-08-08)

- Immutable release `gate6-20260808T233000Z-rna-mutable-snakemake-state-r31` was built by Slurm job `41306089` (`COMPLETED 0:0`) from the pinned r18 source/vendor baseline with the r30 normalized-matrix schema overlay, r29 Craftmake projection overlay, and r31 shared resolver asset-filter overlay.
- Independent verification passed release checksums, `otter --help`, `craftmake --help`, source-overlay inspection, sealed permissions, and release properties. The release records `release.workflow_asset_filter=regular-files-only-root-snakemake-state-excluded`. Its `checksums.sha256` digest is `ebdb00ae07d053a63d73551874c41fb5a9e0a001dae6f740de344a59a7afbf3f`.
- `binaries/current` was promoted atomically to r31 only after independent verification. The fresh project is `gate6-rnaseq-executor-parity-r31`; its input/reference/workflow assets were copied from the immutable historical parity fixture, while only project identity, release lock, run IDs, and run-local state were newly materialized.
- Fresh paired snapshots were resolved with `SLURM_CLUSTER_NAME=zc-m6`: `run-20260808T233500Z-rabcde` for Craftmake and `run-20260808T233501Z-rsabcd` for explicit Snakemake. Strict normalized snapshot parity passed; the pair differed only in the permitted executor/run metadata and run-local paths. Both snapshots contain no `.snakemake` workflow asset, while all 27 regular workflow assets remain covered by the shared digest `sha256:4c5fdf8ef6fa8a2ae235dae8a912b02b0abf49d9a40ab79263c55a973be06a63`.
- Craftmake controller `41316038` and explicit Snakemake controller `41316039` completed `step1 → step2 → step2-check → publish` successfully. Snakemake child jobs `41316042`, `41316082`, and `41316091` completed successfully; no immutable snapshot drift was reported after `.snakemake` logs/locks were written.
- Both fresh `results/artifacts.json` manifests passed `otter artifact verify`. `otter artifact compare` returned `passed: true`, `comparable: true`, comparing `expression-count-matrix`, `expression-normalized-matrix`, `qc-summary`, and `splicing-outcome`. The RNA executor-parity gate is accepted for this canary. WGBS `SRR6373947` remains deferred.

## RNA r1-r31 Lessons and PDX Preflight (2026-08-09)

The RNA lineage is now treated as a failure-prevention checklist for PDX rather than as a single successful run. The durable lessons are:

| Lineage issue | Root cause | Required safeguard before PDX |
|---|---|---|
| Missing phase or broken verifier invocation | controller and verification harness defects | validate controller arguments and verifier scripts before Slurm submission |
| Missing `group_levels`, species, or immutable `pdata.xlsx` projection | compatibility config was incomplete or inferred from mutable state | derive every compatibility field from the sealed snapshot and test the projection directly |
| Craftmake task failures and scheduler allocation-confirmation failures | product task defects must be separated from Slurm/control-plane incidents | retain classified exit codes, reconcile with `squeue`/`sacct`, and reject operationally invalid pairs |
| Normalized-matrix schema mismatch | artifact payloads matched but declarations differed | compare artifact IDs, schemas, comparators, and payloads before accepting scientific parity |
| Non-writable `.snakemake` state and r30 workflow-asset drift | mutable runtime state was copied/sealed and then hashed as immutable workflow input | recreate runtime state writable and exclude only mutable `.snakemake` directories from shared asset digesting |
| Reuse of historical or invalid evidence | old runs were convenient but not fresh | every formal pair uses fresh immutable snapshots and fresh evidence; failed attempts remain diagnostic only |

Before the first PDX compute-node pair, the shared contracts were remediated as follows:

- BS-PDX and RNA-PDX filtered BAM/BAI publication IDs use the same executor-neutral `graft-alignment-bam-*` and `graft-alignment-bai-*` namespace. A regression test rejects the old RNA-PDX-specific IDs.
- Craftmake and Snakemake normal PDX filtering consume original step2 graft/host BAMs. Xenofilx is the sole producer of the fixed-name filtered BAM and matching BAI; both executors validate the BAM with `samtools quickcheck` and publish the Xenofilx-produced BAI without regenerating it with `samtools index`.
- Normal Xenofilx execution requires `--recalculate-nm`; BS-PDX additionally requires `--bisulfite`, while RNA-PDX must omit it. Picard patch rules are retained only for an explicit original-XenofilteR comparison and are excluded from normal PDX executor paths.
- Snakemake Methrix reference preparation now uses the same `extract-cp-gs` preference, `extract-cpgs` fallback, and empty-CpG contig retry as Craftmake, while consuming the resolved immutable GTF. Craftmake Methrix declares the complete producer output set required by Snakemake QC.
- Static Craftmake compiler and Otter publication tests cover these contracts, including RNA-PDX ID parity, Xenofilx-provided PDX BAI preservation, full Methrix output declarations, and idempotent publication retry. Real Snakemake execution, real Slurm recovery, and scientific semantic parity remain unaccepted until the fresh PDX pair completes.

The resulting order is: BS-PDX canary, RNA-PDX canary, genuine Snakemake interruption/retry, scientific semantic comparison, then the representative production matrix. WGBS remains explicitly deferred.

All releases below were built through Craftmake `ReferenceBuild`, published by staging plus atomic rename, verified with `sha256sum -c` on a compute node, sealed read-only, and read from a separate compute-node visibility job.

| Requested ID | Physical release root | Manifest SHA-256 |
|---|---|---|
| `hg19@GRCh37.p13-gencode-v19` | `references/genomes/hg19/GRCh37.p13-gencode-v19` | `33dfd7d4ec0a90c6e11fdc45d02b2d4e9b6d82e4a607148d1ab467a0e555accc` |
| `hg38@GRCh38-gencode-v44` | `references/genomes/hg38/GRCh38-gencode-v44` | `bda77ed9591ec4b9b6c4fadf032e9bfeafec0a6e03e7854e4d706e4f8909a132` |
| `mm10@GRCm38-gencode-M25` | `references/genomes/mm10/GRCm38-gencode-M25` | `777158ba49da3f43c76b449c635878ba65d26dfc2ebd44cd2af212a3aabf29e2` |
| `mm9@NCBIM37-gencode-M1` | `references/genomes/mm9/NCBIM37-gencode-M1` | `149d17fbcf5a37bb3740c8c9284fe3a58f25eea9a704762faf3137e4ad3be0c9` |

`mm38@GRCm38-gencode-M25` is an alias of the canonical `mm10` physical release; it does not own duplicate indexes. Compute-node alias acceptance job `40951842` resolved requested ID `mm38` to:

```text
/public3/home/scg9946/otter-gate6/references/genomes/mm10/GRCm38-gencode-M25
```

with the canonical `mm10` manifest digest.

## Release Evidence

- `mm9` complete integrity and sealing verification: job `40951326`.
- `mm10/mm38` complete integrity and sealing verification: job `40951561`.
- `hg19` complete integrity and sealing verification: job `40951815`.
- `hg38` complete integrity and sealing verification: job `40951816`.
- Four-release compute-node visibility verification: job `40951840`.

Each accepted release contains FASTA, FAI, GTF, Bismark CT/GA conversion indexes, Bowtie2 indexes, STAR indexes, `reference.yaml`, `manifest.json`, and `checksums.sha256`. Release directories are `0555`; regular release files are `0444`.

## ReferenceBuild Resource Policy

The completed production build used the earlier `8 CPU / 96 GiB` publish allocation with `index_build_threads: 8`. That historic immutable evidence must not be rewritten.

All **future** `ReferenceBuild` workflow revisions use the doubled resource policy below:

| Task | CPUs | Memory | Time | Index invocation |
|---|---:|---:|---|---|
| `acquire_sources` | 2 | 8 GiB | 3 days | checksum-bound resumable archive download |
| `prepare_assets` | 4 | 32 GiB | 4 hours | full or declared-contig FASTA/GTF derivation |
| `publish_release` | 16 | 192 GiB | 24 hours | `--index-build-threads 16` |

With `index_build_threads: 16`, the release builder invokes:

```text
bismark_genome_preparation --bowtie2 --parallel 8
bowtie2-build --threads 16
STAR --runMode genomeGenerate --runThreadN 16
```

Bismark runs CT and GA conversion indexers concurrently, so `--parallel 8` consumes the 16 allocated CPU threads during that stage. New immutable ReferenceBuild configurations must set `index_build_threads: 16` to match the `publish_release` allocation.

### Bismark Version and Rust Migration Compatibility

The immutable releases currently accepted on Paracloud were built with the legacy Perl implementation, not the Rust suite. Their `reference.yaml` metadata records `bismark_genome_preparation` with `--bowtie2 --parallel 4`, matching the historic `8 CPU / 96 GiB` allocation. At build time, its runtime executable was:

```text
/public3/home/scg9946/otter-gate6/environments/rattler-root/envs/otter-core-trimgalore0610-candidate/bin/bismark_genome_preparation
```

It was a Perl executable reporting `Bismark Genome Preparation Version: v0.25.1`; its files and Conda metadata have now been removed from the support prefix. Existing accepted reference releases remain immutable historical evidence and must not be rebuilt or altered in place. They remain usable by the Rust suite through the retained Bismark index layout; all new production reference builds use the Rust runtime.

Upstream Bismark Rust suite `bismark-rust-v3.1.0` retains a compatibility executable named `bismark_genome_preparation` and also supports the new `bismark prepare` spelling. Its documented and source-validated contract preserves the Otter invocation shape:

```text
bismark_genome_preparation --bowtie2 --parallel <N> <genome-folder>
```

Rust `--parallel` remains per conversion/indexing process, requires `N >= 2`, and runs CT and GA conversion indexes concurrently; `--parallel 8` therefore remains correct for a 16-CPU `publish_release` allocation. The Rust implementation also retains the required output tree `Bisulfite_Genome/CT_conversion` and `Bisulfite_Genome/GA_conversion`, which Otter validates before publication. The migration policy is Rust-only for active execution: no legacy-versus-Rust Bismark benchmark or output-comparison gate is required. Future reference construction uses the pinned Rust runtime and creates a new immutable release only when a new reference revision is otherwise needed; merely replacing an executable under an existing runtime or accepted release path is prohibited.

### Active Executor Binding and Conda Policy

All maintained Craftmake workflows and both current and retained legacy Snakemake rule assets bind Bismark commands to the versioned runtime name:

```text
otter-core-bismark-rust-3.1.0-r2
```

This revision is pinned to upstream Bismark Rust source tag `bismark-rust-v3.1.0` at commit `e552b8f307a7041bcebed8f8e5a764ebcf7b046c`; the locally acquired tag source archive has SHA-256 `638d7b068031b23fca9c3b3ca5a6b0db675a4da9d48975420fa642b584b5e411`. The locally built static Bismark package SHA-256 is `a6d27c20be57587a8f03c7e5d01619bf08aea715c56098068cd44594bbe11a1f`; its contained `bismark` executable SHA-256 is `c318a33479bb643141c5b66560498ebb63b96e04fee2131f0936d68c7b038ffe`. On Paracloud, `ldd` reports `statically linked`; `file` may still classify a static PIE as a shared object, which is ELF type metadata rather than a dynamic-library dependency. The upstream Linux x86_64 release asset SHA-256 is `8f7b355840523e98c87d7dc61e94227ece9a07a79b48ecce613451539cb3bdab`, but that vendor binary requires glibc 2.39 and cannot run on Paracloud glibc 2.17. It also pins upstream Bowtie2 `v2.5.4` Linux x86_64 release asset SHA-256 `32de7d9363124452296d45d7cb7da48d41c60e9ebaeebd2feafe1cc0418ae00c`. The runtime overlays Bismark and Bowtie2 over a read-only Conda base that provides supporting tools such as Samtools and STAR.

The base `otter-core` Conda specifications intentionally no longer declare either the Bioconda `bismark` package or Conda `bowtie2`: a package solver could drift either tool version. The versioned `otter-core-bismark-rust-3.1.0-r2` environment specifications explicitly document the pinned upstream assets rather than falsely representing them as Conda package constraints. Craftmake default workflows and explicit Snakemake compatibility rules use the Rust runtime so the two executors cannot silently select different Bismark or Bowtie2 implementations.

#### Global invocation policy

The active production commands are available directly through the user-global `.cargo/bin` PATH:

```bash
bismark [arguments]
bismark_genome_preparation [arguments]
bismark_methylation_extractor [arguments]
bowtie2 [arguments]
bowtie2-build [arguments]
```

Compute-node migration job `41045990` removed all legacy Perl Bismark executables and `conda-meta/bismark-*.json` records from the prior support prefix, then created global symbolic links to the immutable `otter-core-bismark-rust-3.1.0-r2` runtime. Independent compute-node job `41046006` confirmed the legacy commands and metadata are absent; global Bismark commands report Rust `3.1.0`; global Bowtie2 and Bowtie2-build report `2.5.4`; and every global link targets the r2 runtime. `enva run otter-core` remains valid for other support tools but does not package Bismark. Use bare global commands for Bismark and Bowtie2; `enva run otter-core-bismark-rust-3.1.0-r2` remains an equivalent explicit runtime boundary when desired.

The first full Conda-clone provisioning task `41045448` timed out before publication. The lightweight overlay task `41045633` then failed safely due to a malformed Bowtie2 checksum pin, and corrected task `41045639` verified both release assets but failed safely because the vendor Bismark executable requires glibc 2.39. Revision `r1` was intentionally not accepted: its static runtime ran successfully, but its manifest contained pre-rename staging paths and could not meet immutable evidence requirements. The accepted `r2` runtime was published by compute job `41045847`; its manifest SHA-256 is `f08d7a63a8b39b55b128111d22c5afa965a1b3b3d84359ea8ed02c2bca4c2f56`. Independent compute-node job `41045904` verified the read-only prefix, Enva name resolution, compatibility shims, Bismark Rust 3.1.0, Bowtie2 2.5.4, Samtools, STAR, executable checksums, and the static linkage assertion. The same-name symbolic link in the established Conda `envs/` search directory is used only for Enva name discovery; it does not copy or alter the legacy runtime.

## Local-first SRA Acquisition Plan

WGBS `SRR6373947` is intentionally deferred and must not be downloaded, uploaded, decoded, or included in the current canary matrix. The active runs are all paired-end and are acquired as **original SRA archives**, not as provider-derived FASTQ:

| Scenario | Accession | NCBI SDL SRA size | NCBI SDL MD5 |
|---|---|---:|---|
| RRBS | `SRR31480456` | 3.11 GB | `499ec5422c8bad0b6c6ae567e38690bf` |
| RNA-seq | `SRR1039508` | 1.67 GB | `b55775f72aa66e2632adf9d5a5bf0e84` |
| BS-PDX | `SRR23802966` | 42.67 GB | `485cb13d3b970e00395814a5d39cbca5` |
| RNA-PDX | `SRR30880970` | 2.44 GB | `467d3c4da50957a429f742fef4459804` |

The four archives total 49.82 GB (46.40 GiB). Local storage checked on 2026-07-30 has 868 GiB available; the download plus a conservative two-fold staging reservation fit with substantial headroom.

### Native archive retrieval and compute-node decode

The active retrieval tool is the locally built `rainoffallingstar/getdown` revision based on upstream `1f781f8898b348f6383672498b6469d1cccaeda0`. ENA exposes no `sra_ftp` links for these records, so `getdown sra --kind sra --decode none` now uses a native Go NCBI SDL resolver:

1. Query `https://locate.ncbi.nlm.nih.gov/sdl/2/retrieve?acc=<SRR>`.
2. Select the matching `type: "sra"` archive location.
3. Download its official archive URL through the existing Go HTTP Range-resume implementation.
4. Require the SDL-reported byte count and MD5 to match before treating the archive as acquired.
5. Generate a separate SHA-256 manifest for local-to-Paracloud transfer verification.

`prefetch` is deliberately not used. `--decode none` is mandatory on the local host: no local SRA parsing or FASTQ generation is allowed. Only after a checksum-matched upload to `/public3/home/scg9946/otter-gate6/cache/gate6-sra-20260730-native-sdl/archives/` may a Slurm compute-node job use SRA Toolkit `fasterq-dump 3.2.1` and gzip the resulting FASTQ files.

Required acquisition sequence:

1. Download each archive locally into the dedicated Gate 6 staging directory; preserve `.part` files for safe HTTP resume.
2. Retain `runinfo.tsv`, `links.tsv`, `metadata.json`, SDL MD5, archive byte count, and local SHA-256.
3. Upload only the SDL-MD5-verified archive and its evidence to the isolated Paracloud staging path using resumable `agentsshcli upload`.
4. Verify remote SHA-256 before submitting a compute-node-only `fasterq-dump` + gzip task.
5. Validate paired FASTQ integrity and then apply deterministic, seed-bound downsampling.

## Completed Active SRA Evidence

All four active archives completed the local native-Go SDL acquisition path, including SDL byte-size and MD5 verification. Each archive was uploaded to the isolated Paracloud staging path with a local SHA-256 transfer manifest, then decoded only in a Slurm compute-node allocation. WGBS `SRR6373947` remains deferred and was not acquired or processed.

| Scenario | Accession | Decode job | Decode result | FASTQ integrity job | Result |
|---|---|---:|---|---:|---|
| RRBS | `SRR31480456` | `40985517` | `COMPLETED (0:0)` | `40987819` | paired FASTQ published and verified |
| RNA-seq | `SRR1039508` | `40985515` | `COMPLETED (0:0)` | `40986554` | paired FASTQ published and verified |
| BS-PDX | `SRR23802966` | `41001757` | `COMPLETED (0:0)` | `41039599` | paired FASTQ published and verified |
| RNA-PDX | `SRR30880970` | `40985516` | `COMPLETED (0:0)` | `40987818` | paired FASTQ published and verified |

Every completed decode first matched the archive's remote SHA-256 against the uploaded transfer evidence, then invoked `fasterq-dump --split-files -e 8`, compressed the paired outputs with `gzip -n`, generated `checksums.sha256`, and atomically published the final FASTQ directory. The independent integrity jobs ran `sha256sum -c checksums.sha256` and `gzip -t` against each published pair.

BS-PDX requires an explicit note because its two gzip outputs differ in size. This does not indicate a multi-sample concatenation: its immutable `runinfo.tsv` identifies one `PAIRED` run, and decode job `41001757` reported `387,683,390` spots and `775,366,780` reads written. Dependent compute-node record-count audit `41040286` completed successfully and recorded exactly `387,683,390` FASTQ records in both `SRR23802966_1.fastq.gz` and `SRR23802966_2.fastq.gz`; its evidence is `SRR23802966.fastq-count-41040286.json`. Compression sizes are not used as a pairing or sample-identity criterion.

The native SDL resolver is published in `rainoffallingstar/getdown` commit `666cbf7`; automatic retry after an interrupted HTTP response body is published in commit `a632b2c`. The BS-PDX upload used the retry path after a transient connection failure, retaining and resuming the remote `.part` archive before atomically publishing the final archive. No SRA archive was decoded locally and `prefetch` was not used.

## Standalone Craftmake RRBS Decode Requalification (2026-08-13)

The new strict `craftmake.standalone/v1` route was exercised on an isolated, create-only RRBS root rather than any historical Gate 6 path. A locally acquired `SRR31480456.sra` archive was checked before transfer and again by the compute-node decode task: `3,109,362,710` bytes, NCBI SDL MD5 `499ec5422c8bad0b6c6ae567e38690bf`, and SHA-256 `d501c0360b93ca9f0b5ca02e69d46d71eb998a36d4fbe1b1af65ee619493ab9e`.

- `SRAArchiveDecode/decode` compiled and validated locally and on Paracloud using a static Craftmake binary; it ran as Slurm job `41423708` on `amd_512`, finishing `COMPLETED (0:0)` in `10m36s`.
- The workflow revalidated archive size/MD5/SHA-256/transfer checksum before `fasterq-dump --split-files -e 8`, used deterministic `pigz -n`, performed structural FASTQ and mate-count audit, and atomically published a read-only `decoded/` directory.
- Independent compute-node verifier `41423871` finished `COMPLETED (0:0)` in `5m20s`, rechecked `checksums.sha256`, `gzip -t`, non-symlink/read-only publication, final manifest paths, and `28,865,648` records in each mate. The final SHA-256 values are R1 `5050278313b8198185244b02df6cca7d2b4538e65f3da6f971951488690ad6a1` and R2 `2419d75dfcd179d0740b8340dce80bfa0620778adf2654372cba6d70c743dc44`.
- First-pass job `41423490` completed (`0:0`, `11m39s`) but exposed a staging-path reference in its decode manifest. It is retained only as diagnostic execution evidence. The fixed r2 workflow records the final `decoded/R1.fastq.gz` and `decoded/R2.fastq.gz` paths and is the accepted generic-decode evidence.

This record is not an `otter.sra-acquisition/v1` manifest and does not yet clear the RRBS production-input gate. The final provenance manifest must bind the frozen archive/FASTQ identity and the RRBS primary reference role before an immutable production `run.yaml` may use these inputs.

## Ready Decoded FASTQ Summary (2026-08-15)

The following seven archives are fully acquired and decoded on Paracloud through the isolated `craftmake.standalone/v1` create-only path. Every row passed a compute-node-only decode (`fasterq-dump --split-files -e 8`, deterministic `pigz -n`, structural FASTQ/mate-count audit, atomic read-only publication) and an independent compute-node verifier (`gzip -t`, `checksums.sha256`, final-path manifest, record count, SHA-256). All archives remain on Paracloud; local `.sra` copies were removed after processing. These records are `otter.sra-fastq-decode/v1` decode evidence only and still require final `otter.sra-acquisition/v1` provenance binding before feeding immutable production `run.yaml` inputs.

| Accession | Scenario | Decode root | R1 bytes | R2 bytes | Paired records | R1 SHA-256 | R2 SHA-256 |
|---|---|---|---|---|---|---|---|
| `SRR31480456` | RRBS (human) | `rrbs-SRR31480456-20260813T041223Z-r2` | 1,994,492,101 | 2,148,536,502 | 28,865,648 | `5050278313…ad6a1` | `2419d75dfc…43dc44` |
| `SRR1039508` | RNA-seq (human) | `rnaseq-SRR1039508-20260814T0411Z` | 1,371,724,331 | 1,357,718,211 | 22,935,521 | `1749a8a290…62869` | `02926c7f30…82aa7` |
| `SRR23802966` | BS-PDX (human graft / mouse host) | `bs-pdx-SRR23802966-20260814T0149Z` | 28,258,736,221 | 29,595,686,034 | 387,683,390 | `8b4a489ff4…4983f` | `329f4bf75a…d2e2cc` |
| `SRR30880970` | RNA-PDX (human graft / mouse host) | `rna-pdx-SRR30880970-20260814T0756Z` | 2,194,429,653 | 2,238,892,137 | 39,065,482 | `97dbba27dc…a2aa9` | `c2d1b30f38…d6be26` |
| `SRR018258` | Human RNA-seq (candidate) | `human-rnaseq-SRR018258-20260814T1450Z` | 1,042,632,531 | 1,043,429,319 | 15,314,364 | `753d3c7e3a…71fc` | `602f3883cb…bfda` |
| `SRR037954` | Mouse RNA-seq (candidate) | `mouse-rnaseq-SRR037954-20260814T1450Z` | 1,750,937,820 | 1,671,632,236 | 23,060,060 | `2c016df249…62a0` | `4c1c131d49…c91cd` |
| `SRR10025242` | Mouse RRBS (candidate) | `mouse-rrbs-SRR10025242-20260814T1450Z` | 504,131,054 | 596,921,250 | 11,935,038 | `b5a565e4b7…f5e5` | `795a0b49da…f683d` |

The first-pass RRBS root `rrbs-SRR31480456-20260813T041223Z` (job `41423490`) is retained only as diagnostic evidence because its decode manifest referenced staging paths; the `-r2` root is the accepted RRBS decode. WGBS `SRR6373947` remains deferred and is not present here.

## Otter-Craftmake Seven-Input Comparison Control Plane (2026-08-15)

The seven ready decoded pairs are the planned toolchain-comparison corpus, not yet production-qualified workflow inputs. Before any comparison task is submitted, Otter must create an immutable `otter.sra-acquisition/v1` manifest for the selected scenario input and then run `otter config validate` plus `otter config resolve`. The resulting read-only `otter.run/v1` snapshot is the only accepted task configuration.

Otter owns input/provenance/reference validation and immutable snapshot creation. Craftmake owns the complete workflow control plane through `otter run --executor craftmake`, including plan, Slurm submission, resume, status, logs, and report. Do not hand-author production `run.yaml` files and do not call `sbatch` directly for workflow tasks. Direct scheduler use remains limited to cluster administration and historical diagnostic evidence, never as a replacement for Craftmake submission.

Every selected paired FASTQ follows raw modern QC and legacy FastQC + SeqKit QC, a common pinned Trim Galore transformation, then trimmed modern and legacy QC. Craftmake must provide both QC implementations for this comparison; mixing Craftmake modern QC with a Snakemake legacy QC run is executor-mixed evidence and cannot be claimed as toolchain parity. `pairbam`/`bamdriver` are limited to BS (RRBS/WGBS) and BS-PDX paired-BAM stages. RNA-seq and RNA-PDX comparisons omit those operators. Snakemake remains only for the separately scoped interruption/retry, publication/recovery, and compatibility-smoke closeout.

## Immutable Acquisition Provenance and Otter Snapshots (2026-08-15)

All seven decoded inputs now have create-only `otter.sra-acquisition/v1` records published under each acquisition root at `provenance/otter-sra-acquisition.json` (mode `0444`). Each publication re-verified archive/FASTQ regular-file identity, byte size, SHA-256, paired record count, and the scenario reference binding. Reference bindings: human RRBS → `hg19@GRCh37.p13-gencode-v19`; human RNA-seq → `hg38@GRCh38-gencode-v44`; mouse RNA-seq and mouse RRBS → `mm10@GRCm38-gencode-M25`; BS-PDX and RNA-PDX → graft `hg38@GRCh38-gencode-v44` + host `mm10@GRCm38-gencode-M25`.

An isolated comparison runtime was deployed at `/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime` with a static Otter binary, a static current-session Craftmake binary, and the current workflow catalog (`catalog/BeaverBS|BeaverRNA|BeaverPDX|BeaverRNASEQPDX`). The catalog step1 YAML files carry the new legacy QC tasks (`legacy_fastqc_before`, `legacy_fastqc_after`, `legacy_seqkit_statistics`) that invoke `enva run otter-core -- fastqc|seqkit`; the `otter-core` environment provides FastQC 0.12.1 and SeqKit 2.13.0 on Paracloud.

Seven canonical `otter.run/v1` snapshots were resolved through `otter config resolve --site paracloud-gate6` with `craftmake` executor, `slurm` backend, and the `executor-phase-envelope/v1` phase resources. Each project exposes a read-only `data/<sample>_R1.fastq.gz`/`data/<sample>_R2.fastq.gz` symlink pair to the acquisition-declared `decoded/R1.fastq.gz`/`R2.fastq.gz` files so the workflow `<sample.id>_R1.fastq.gz` input convention resolves to the provenance-verified inputs. Frozen sample digests in each snapshot equal the acquisition manifest digests.

| Project | Scenario | Authoritative run ID |
|---|---|---|
| `human-rrbs-SRR31480456` | RRBS | `run-20260815T093050Z-ykunpt` |
| `human-rnaseq-SRR1039508` | RNA-seq | `run-20260815T093045Z-kwgebr` |
| `human-rnaseq-SRR018258` | RNA-seq | `run-20260815T093040Z-qkzjbu` |
| `mouse-rnaseq-SRR037954` | RNA-seq | `run-20260815T093057Z-ywlznw` |
| `mouse-rrbs-SRR10025242` | RRBS | `run-20260815T093102Z-liwxbq` |
| `bs-pdx-SRR23802966` | BS-PDX | `run-20260815T094416Z-rjwvkx` |
| `rna-pdx-SRR30880970` | RNA-PDX | `run-20260815T093104Z-patmyt` |

A previous resolve round used direct `decoded/R1.fastq.gz` sample paths and produced diagnostic snapshots only (e.g. `run-20260815T085710Z-htkaee` and `run-20260815T092853Z-hjhtuh`); they are retained as evidence that the workflow input convention requires per-sample `_R1/_R2.fastq.gz` names.

Step1 plans compile through Otter → Craftmake for all seven scenarios (BeaverBS/BeaverRNA/BeaverPDX/BeaverRNASEQPDX, 6 tasks per single-sample project: modern `fastqc_before`/`fastqc_after`, shared `trim_reads`, legacy `fastqc_before`/`fastqc_after`/`seqkit_statistics`). Slurm controller jobs `41457463`–`41457468` run `otter run --foreground --executor craftmake --phase step1` per project. Mouse RRBS controller `41457428` completed successfully (`status: succeeded`, controller exit 0), producing raw modern/legacy QC, Trim Galore paired outputs (`SRR10025242_val_1.fq.gz`, `SRR10025242_val_2.fq.gz`), trimmed modern/legacy QC, and `work/QC/SRR10025242_seqkit_stat.txt`. Initial background-attempt controller `41457420` was killed with its worker when its Slurm job ended; the authoritative mouse RRBS run uses the foreground controller path.

## r16b RRBS Executor-Parity Evidence

The accepted native release is `gate6-20260805T102500Z-aca58fb-8bb5d64-snakemake-projection-reuse-r16`. Its clean paired RRBS snapshots are `run-20260806T013408Z-craftm` and `run-20260806T013408Z-snakem`, each bound to the same deterministic RRBS canary, reference release, toolchain, and `executor-phase-envelope/v1` resources. Both completed all analysis phases on Paracloud. Their read-only artifact manifests passed `otter artifact verify`, and `otter artifact compare` passed for Methrix HDF5, Bismark summary HTML, and QC XLSX. Create-only phase accounting evidence is retained at:

```text
/public3/home/scg9946/otter-gate6/evidence/gate6-20260805T102500Z-aca58fb-8bb5d64-snakemake-projection-reuse-r16/rrbs-executor-parity-r16b
```

The initial Snakemake `step3-check` controller failed while child job `41185805` continued and eventually completed. The same immutable snapshot was published and verified through the recovery path. The controller now reconciles empty or transiently unavailable `squeue` results through `sacct` before assigning a terminal status; the r17b confirmation controller `41186233` and child `41186234` both completed with `0:0`, with read-only evidence `rrbs-reconciliation-smoke.json`. r18 then demonstrated real controller interruption/resume for both executors, create-only publication retry, and manual retry of a classified failed publish task. Immutable accounting, semantic, and recovery evidence is retained under `gate6-20260806T050000Z-aca58fb-8bb5d64-slurm-timeout-submit-retry-r18`. Controllers must also expose the release `bin/` directory in `PATH`; Snakemake controllers isolate inherited Java settings and use the Rust runtime library directory so Picard and Methx resolve compatible JVM/HDF5 dependencies.

## Next Gate 6 Execution Stage

The next stage is **paired production-grade Slurm canary validation for the remaining non-WGBS scenarios**, not another mock or local-only test. RRBS artifact parity, controller/child reconciliation, controller interruption/resume, create-only publication retry, failed-task retry, Methrix semantic parity, and normalized QC equality have passed. RNA-seq, BS-PDX, and RNA-PDX each still require their paired compute-node canaries. WGBS remains deferred.

For every active scenario, run both required executor paths independently on Paracloud compute nodes:

1. Craftmake as the default executor.
2. Explicit Snakemake compatibility executor; no automatic fallback is allowed.

Each canary must produce artifact manifests, collect Slurm/task metrics, pass structural artifact validation, and receive a scenario-specific semantic parity report before progressing to recovery experiments and the representative 20-cell matrix. The full 20-cell × 3-run representative benchmark is not yet authorized by this acquisition evidence alone; it follows only after the production-grade canaries and their parity/recovery gates pass.

## Production Canary Controls

The immutable input/executor/toolchain contract is [Gate 6 Canary Matrix](gate6-canary-matrix.md). It freezes RRBS, RNA-seq, BS-PDX, and RNA-PDX selection seeds, source checksums, reference releases, executor pairing, and promotion gates. WGBS remains deferred. Executor parity holds toolchain and resolved resources constant while comparing Craftmake-default with explicit Snakemake compatibility; toolchain parity holds executor constant. Rust Bismark `3.1.0` and Bowtie2 `2.5.4` remain pinned common dependencies and are not compared to the removed Perl runtime.

### Deterministic input publication

The checksum-pinned publication assets are stored under:

```text
/public3/home/scg9946/otter-gate6/evidence/gate6-canary-inputs-20260801/
```

They use `sha256-qname-modulus/v1` selection of paired FASTQ records, validate source checksums before streaming, write deterministic gzip output, publish by staging plus atomic rename, and reject pre-existing output paths. Each published output has `R1.fastq.gz`, `R2.fastq.gz`, `canary-inputs.json`, and `checksums.sha256`; acceptance additionally requires `gzip -t`, mate-pair record-count audit, and a compute-node visibility check. The input manifest is create-only; later immutable `run.yaml` snapshots reference its digest rather than modifying it.

Compute-node array job `41046679` completed on `amd_512` with 4 CPUs, 16 GiB, and a 12-hour limit. Its four create-only outputs were accepted only after independent compute-node job `41048591` rechecked `checksums.sha256`, `gzip -t`, the `otter.canary-inputs/v1` manifest identity, FASTQ record structure, mate cardinality, read-only permissions, and visibility. The verified paired-record counts are RRBS `57,418`, RNA-seq `45,729`, BS-PDX `193,159`, and RNA-PDX `77,883`; the four immutable input-manifest SHA-256 values are retained in `compute-node-input-manifests-41048591.sha256`. Verification attempt `41048586` is retained as non-acceptance diagnostic evidence: it failed after RRBS validation because a Bash loop redeclared a `readonly` variable, then job `41048591` reran the complete verifier successfully.

### Executor resource-equivalence prerequisite

No executor-parity workflow can be submitted yet. Craftmake resolves per-job `cores`/memory/partition from workflow YAML, while the explicit Snakemake compatibility path uses one per-phase allocation from `run.yaml`; the latter cannot currently express Craftmake batch-worker resources or per-rule memory. It also defaults partition differently and current rule-level threads diverge for selected checker/Picard/Xenofilx paths. The resource contract must become a single machine-readable source, with per-job Craftmake equality checks and Snakemake phase allocations derived as the maximum job requirement, before a Craftmake-vs-Snakemake result may be attributed to executor behavior. This is a gating implementation task, not a waiver candidate.

### Runtime incident and evidence collection

Every canary, failed canary, cancellation, and resume must retain a `craftmake.run-evidence/v1` bundle: run/task identity, controller JSONL path, task metrics, step timing, allocations, and any `otter.runtime-incident/v1` records. Incident categories are input/reference/digest, environment/tool, scheduler submission, queue timeout, resource exhaustion, filesystem/I/O, network acquisition, tool invocation, scientific/QC, artifact integrity, cancellation/recovery, and internal/unknown. Only bounded scheduler-submission or acquisition failures are safe for automatic retry. Unknown or blocker-class incidents require classification, root cause, and a new immutable run before promotion.

For every completed execution cell, collect `craftmake report --refresh-metrics`, controller JSONL, `sacct` job/step accounting, `otter artifact verify`, `otter artifact compare`, and the scenario semantic report. Never use login-node execution for the workflow, input generation, recovery injection, or accounting validation.