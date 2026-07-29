# System Context (Updated: 2026-07-29)

## 1. 已实现的核心模块 (Modules)

### Reference release builder
- **Path**: `cmd/reference.go`, `internal/reference/build.go`, `internal/reference/build_types.go`
- **Public Methods**:
  - `otter reference build`: stages and atomically publishes a reference release.
  - `reference.BuildRelease(BuildRequest): BuildResult`: copies FASTA/GTF, runs `samtools faidx`, builds selected real Bismark/Bowtie2/STAR indexes, writes typed metadata and verifies publication.
- **Data Flow**: source FASTA + GTF → sibling staging release → FAI + selected indexes → `reference.yaml` + `manifest.json` + `checksums.sha256` → atomic directory rename.
- **Dependencies**: `samtools`, selected index-build executables, shared registry filesystem.
- **Status**: default registry root is `$OTTER_REFERENCE_ROOT`, else `~/.otter/references`; complete releases require at least one real index and are sealed read-only. Gate 6 still requires trusted source provenance and a Paracloud compute-node visibility preflight before any built release can be used. The first bounded mouse build is `mm10-canary@GRCm38-gencode-M25-chr19-MT`, derived from GENCODE contigs `19,MT` only; it is technical-canary-only and cannot stand in for a full `mm10`/`mm38` reference.

### Typed run resolver
- **Path**: `internal/config/v1/`, `internal/config/resolver/`, `internal/run/`
- **Public Methods**:
  - `Resolver.Resolve(ProjectConfig): RunSnapshot` - 将 project/sample/reference/site 固化为 immutable run。
  - `RevalidateSnapshot(RunSnapshot): error` - 在 run/resume 前验证输入、reference 与 workflow assets drift。
- **Data Flow**: `project.yaml + samples.tsv + locks + site` → `runs/<id>/run.yaml`。
- **Dependencies**: reference registry、site profile、workflow asset digest。
- **Status**: FASTQ metadata、resolved resources、required workflow assets 均 fail closed；project/run examples and invalid mutations now execute through both draft 2020-12 JSON Schema and strict Go loaders. Immutable run validation now rejects malformed resolved reference selections and requires provenance for every resolved SLURM field; schema likewise rejects unknown resource/evidence fields and missing SLURM resources. Coverage must still expand to references lock/reference definition schemas and future contract fields.

### Executor boundary
- **Path**: `cmd/run_craftmake.go`, `cmd/run_snakemake_snapshot.go`, `internal/execution/`, `internal/craftmake/`
- **Public Methods**:
  - `LoadRunInvocation(path, Executor): RunInvocation` - 两个 executor 的统一 immutable snapshot boundary。
  - `craftmake.Execute(Request): Result` - protocol envelope 与 classified exit code。
  - `cmd.ExitCode(error): int` - 向 CLI 与后台 task 保留 Craftmake 分类退出码。
- **Data Flow**: immutable `run.yaml` → explicit executor → local/SLURM backend → task state。
- **Dependencies**: `internal/task/`, Craftmake JSON protocol, Snakemake compatibility config。
- **Status**: Craftmake 默认、Snakemake explicit compatibility；没有 automatic fallback。

### Artifact publication and comparison
- **Path**: `internal/artifact/`, `cmd/artifact.go`
- **Public Methods**:
  - `artifact publish|verify|compare` - immutable manifest 的构建、校验与跨 run 比较。
  - `ComparatorRegistry.Compare(...)` - comparator fail-closed dispatch。
- **Data Flow**: declaration document → stable regular files → checksummed `results/artifacts.json` → verify/compare。
- **Dependencies**: immutable run identity、workflow producer。
- **Status**: expression、PDX count、BAM pair、XLSX workbook structure、HTML、HDF5 signature and typed RNA splicing outcome comparators 已存在。RNA outcome comparison validates the v1 outcome contract and binds declared relative paths to manifest `splicing/files/*` entries while deliberately ignoring run-local absolute `source_root`; individual event files remain exact artifacts. PDX classification comparison now binds each sorted row's ordinal/BAM/BAI names to manifest-declared graft pairs. BAM/BAI comparison rejects symlinks; validates BGZF/BAM magic and reference declarations; and performs bounded streaming validation of each BAI reference's bin, chunk, and linear-index record layout before checking BAM/BAI reference counts. It does not map BAI virtual offsets to BAM records or compare alignment records/bin semantics. XLSX comparison validates required ZIP members and ordered workbook sheet titles, but not cells, formatting, or QC semantics; Methrix remains signature and byte-identity based.

### Workflow producers
- **Path**: `craftmake/workflows/`, `inst/rules/`, `inst/rules_legacy/`
- **Public Methods**: `BeaverBS`, `BeaverRNA`, `BeaverPDX`, `BeaverRNASEQPDX` publish phases。
- **Data Flow**: tool outputs → validated staged files → declaration → `otter artifact publish|verify`。
- **Dependencies**: `otter` CLI、samtools、methx/matsrun/qctb。
- **Status**: All four Craftmake producers and the explicit Snakemake compatibility completion path use results-local temporary staging for producer-owned directories, create-only directory publication, and retry payload tree verification. Snakemake derives declarations from immutable `run.yaml` only after `ExecuteAll()` succeeds, calls the same `artifact.Publish` API, and verifies the resulting manifest. All four `step1` phases, BeaverBS `step2-check` and `step3-check`, BeaverRNA `step2-check`, BeaverPDX `step2-check` and `step3-check`, and BeaverRNASEQPDX `step2-check` and `step3-check` terminate on their real analysis outputs rather than phase-success markers in their Craftmake and workflow-specific Snakemake entries. BeaverBS `step2-check`/`step3-check`, BeaverRNA `step2-check`, BeaverPDX `step2-check`/`step3-check`, and BeaverRNASEQPDX `step2-check`/`step3-check` sample validation now write dimension-aware `otter.sample-artifacts-validation/v1` JSON manifests with regular-file, size, and SHA-256 evidence rather than empty `.ready` files. BeaverPDX and BeaverRNASEQPDX `step2-check` now terminate in `otter.filtered-bam-validation/v1`, which records validated graft filtered BAM/BAI pairs and mapped-read counts; their Craftmake Picard, Xenofilx, filtered-artifact, and phase-success markers were removed. The matching Snakemake PDX/RNA-PDX checker entrypoints now include Picard as a real fixed-BAM producer and Xenofilx as a declared fixed-graft filtered BAM/BAI producer; their `rule all` targets and e2e fixtures no longer depend on patch, filter, or phase-success marker paths. This is static contract coverage only: genuine Snakemake execution, publication, and recovery evidence remain required. BeaverPDX retains its bisulfite-specific Picard/Xenofilx settings, GC-bias validation inputs, and MultiQC output. The old step1 checker rules are no longer referenced by those entrypoints; shared checker rules with remaining consumers are retained. RNA paths use `otter.rna-splicing-outcome/v1` (`produced` or `not_applicable`) and stage actual splicing files only for `produced`. Local publish-recovery evidence now covers RRBS, WGBS, RNA-seq, BS-PDX, and RNA-PDX. It includes matching/mismatched partial retries for WGBS and RNA-seq, interruption during BS-PDX staging, interruption after atomic BS-PDX `pdx/graft` publication, interruption after RNA-seq and RNA-PDX `splicing` publication, and interruption after immutable manifest creation; each covered retry verifies either safe completion or existing-manifest idempotency. These are direct Go publisher tests with fake tools, not genuine Snakemake or SLURM execution. All four Craftmake publishers now execute their compiled Bash publish commands through controlled local process interruption: BeaverBS is terminated during staging; BeaverRNA after `splicing` directory publication; BeaverPDX after `methylation` publication; and BeaverRNASEQPDX after `splicing` publication. Each regression verifies the expected partial tree has no manifest, then retries through existing-tree digest comparison, immutable-manifest permissions, and the verify-only branch using fake `otter`/tool commands. This is local compiler-and-shell recovery evidence only; a single all-artifact transaction and real Snakemake/SLURM evidence remain pending.

## 2. Global Types

| Type Name | File Path | Key Fields | 使用场景 |
|---|---|---|---|
| `ProjectConfig` / `RunSnapshot` | `internal/config/v1/types.go` | workflow, execution, paths, references | canonical project/run contract |
| `BuildRequest` / `BuildResult` | `internal/reference/build_types.go` | source FASTA/GTF, registry identity, selected indexes, published paths/digest | immutable reference-release build |
| `ReferenceBuild` configuration | `craftmake/workflows/ReferenceBuild/build.yaml` | source URLs/MD5, allowed contigs, release/tool paths, evidence paths | Gate 6 Slurm reference-build DAG |
| `RunInvocation` | `internal/execution/invocation.go` | snapshot path, project/run/state/results paths | executor-neutral run boundary |
| `SnakemakeArtifactPublicationRequest` | `internal/workflow/snakemake_artifacts.go` | immutable snapshot, snapshot path, samtools/declaration paths | post-success compatibility publication |
| `ExitCodeError` | `internal/craftmake/client.go` | command, classified code, cause | retain Craftmake exit classification |
| `Manifest` / `Declaration` | `internal/artifact/manifest.go` | identity, artifacts, comparison | artifact publication contract |
| `ComparisonReport` | `internal/artifact/comparator.go` | passed, comparable, issues | cross-run artifact comparison |

## 3. API / CLI 注册表

| Command | Input Type | Response Type |
|---|---|---|
| `otter config resolve` | project v1 inputs | immutable `run.yaml` |
| `otter reference build` | FASTA + GTF + build/index options | immutable reference release and digest |
| `otter run` | immutable `run.yaml` | Craftmake or explicit Snakemake task |
| `otter artifact publish` | `run.yaml` + declarations | immutable manifest |
| `otter artifact verify` | `run.yaml` | verification report |
| `otter artifact compare` | left/right `run.yaml` | comparison report |

## 4. 待解决的技术债

- [ ] Gate 0/1: Expand JSON Schema/Go shared fixtures as project and run contracts gain fields.
- [ ] Gate 5: run genuine Snakemake compatibility publish and interruption/retry evidence on real workflows.
- [ ] Gate 5: retain the content-bearing Craftmake validation-manifest coverage, replace any newly identified marker-only inter-phase edges, and obtain genuine runtime evidence for the declared PDX/RNA-PDX Snakemake fixed/filtered BAM contracts. The all-in-one RNA-PDX entrypoint now uses the digest-bound Picard fixed-BAM and common Xenofilx filtered BAM/BAI assets, but still requires a real Snakemake execution.
- [ ] Gate 5: complete scientific-semantic comparison for Methrix, BAM/BAI, XLSX QC, and splicing outputs before scientific parity claims; current structural checks do not substitute for it.
- [ ] Gate 6: Paracloud compute-node runtime provisioning is accepted for `otter-core`, `otter-snakemake`, and `otter-extra` (`gate6-runtime-acceptance.json`). Canary metadata-accessibility selection is immutable evidence at `gate6-canary-accessions-resolved-v1.json` (SHA-256 `d02f4ceeca9c8cb6f5a2bd25326db3467f345adcdae0d392df1047083b0fd43e`): RRBS `SRR31480456`, WGBS `SRR6373947`, RNA-seq `SRR1039508`, BS-PDX `SRR23802966`, and RNA-PDX `SRR30880970`. It does not establish data acquisition, downsampling, reference compatibility, workflow execution, recovery, or scientific parity. Acquire/checksum/downsample sources; publish and validate compute-visible reference registry assets; then run real dual-executor canaries, recovery evidence, the 20-cell sbatch parity matrix, and benchmark. Local tests are not substitutes.
- [ ] Gate 7: independently evaluate Snakemake retirement after stable Gate 6 evidence.
