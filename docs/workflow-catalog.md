# Otter Workflow Catalog

> 状态：四个 Craftmake workflow families（覆盖五个 scenario）和 Snakemake compatibility completion path 都可生成同一 immutable artifact manifest contract；但尚未满足完整 workflow orchestration、真实 Snakemake/SLURM execution、科学 parity 或 sbatch 生产验证的完成条件。

## 实施状态 (2026-07)

| Scenario | Craftmake workflow YAML | Snakemake snakefile | 当前 publish 状态 | 验证状态 |
|---|---|---|---|---|
| RRBS | `craftmake/workflows/BeaverBS/step1–3.yaml` (+ checks, `publish.yaml`) | `inst/snakefiles/BeaverBS*.snakemake` | Craftmake and post-success Snakemake compatibility stage Methrix HDF5、Bismark HTML 与 QC workbook through the shared immutable manifest API | contract-only |
| WGBS | 同上（共享 BeaverBS） | 同上 | 同 RRBS；artifact declaration 未按 WGBS 单独分化 | contract-only |
| RNA-seq | `craftmake/workflows/BeaverRNA/step1–2.yaml` (+ check, `publish.yaml`) | `inst/snakefiles/BeaverRNA*.snakemake` | Both paths publish count/normalized matrices、QC workbook 与 typed splicing outcome/实际 splicing files | contract-only |
| BS-PDX | `craftmake/workflows/BeaverPDX/step1–3.yaml` (+ checks, `publish.yaml`) | `inst/snakefiles/BeaverPDX*.snakemake` | Both paths stage graft BAM/BAI、samtools mapped-read classification、Methrix HDF5、Bismark HTML 与 QC workbook | contract-only |
| RNA-PDX | `craftmake/workflows/BeaverRNASEQPDX/step1–3.yaml` (+ checks, `publish.yaml`) | `inst/snakefiles/BeaverRNASEQPDX*.snakemake` | Both paths publish graft BAM/BAI、classification、matrices、QC 与 typed splicing outcome/实际 splicing files | contract-only |

## 正交维度

```text
scenario × executor × toolchain × backend
```

- `scenario`：`rrbs | wgbs | rnaseq | bs-pdx | rna-pdx`
- `executor`：`craftmake | snakemake`
- `toolchain`：`modern | legacy-equivalent`
- `backend`：开发 contract 使用 `local`；真实流程与 benchmark 使用 `slurm`

比较 executor 时固定 toolchain；比较工具时固定 executor。禁止用 `craftmake+modern` 直接对比 `snakemake+legacy-equivalent` 并归因执行层。

## 统一 phase contract

| Phase | 责任 | 典型产物 |
|---|---|---|
| `ingest` | 输入、sample/reference preflight | input manifest、preflight report |
| `qc_raw` | 原始 FASTQ QC | per-sample QC、summary |
| `prepare` | trim、adapter 和必要预处理 | cleaned FASTQ、metrics |
| `separate` | PDX graft/host 分离 | classified FASTQ/BAM、counts |
| `align` | reference alignment | sorted/indexed BAM、alignment metrics |
| `quantify` | methylation/expression/splicing 量化 | cov/count/matrix/event tables |
| `qc_final` | 聚合科学 QC | QC workbook/HTML/JSON summary |
| `publish` | artifact validation 与发布 | `artifacts.json`、checksums |

不适用的 phase 可以省略，但 phase 名称和输出类别不得由 executor 改写。

## RRBS

```text
ingest → qc_raw → prepare → align → quantify → qc_final → publish
```

- 比对：Bismark/Bowtie2 等价配置。
- 量化：CpG coverage、methylation beta/coverage、Methrix-compatible HDF5。
- RRBS 特有：barcode/adapter、MspI/CCGG 语义作为 typed parameters。
- 现代算子：`fastqcx`、`methx`、`qctb`。

关键 artifact：trim metrics、BAM/BAI、Bismark coverage、CpG table、HDF5、QC summary。

## WGBS

```text
ingest → qc_raw → prepare → align → quantify → qc_final → publish
```

与 RRBS 共用 BS phase interface，但 adapter 默认、restriction-site 逻辑和覆盖阈值不同。RRBS/WGBS 必须是两个 scenario，不能仅靠运行时字符串分支隐藏差异。

关键 artifact：clean FASTQ、BAM/BAI、coverage、methylation matrix/HDF5、QC summary。

## RNA-seq

```text
ingest → qc_raw → prepare → align → quantify → qc_final → publish
```

- 比对：STAR 与锁定 GTF/reference。
- 量化：gene counts、规范化输入矩阵、可选 rMATS splicing。
- 现代算子：`fastqcx`、`seq2mat`、`matsrun`、`qctb`。

关键 artifact：BAM/BAI、STAR metrics、per-sample counts、merged expression matrix、splicing tables、QC summary。

## BS-PDX

```text
ingest → qc_raw → prepare → separate → align → quantify → qc_final → publish
```

- references：必须同时有 `graft` 和 `host` role。
- species separation：现代 `xenofilx/pairbam/bamdriver` 与锁定 legacy-equivalent 实现。
- 分离后 graft reads 才进入 BS alignment/quantification；host/ambiguous/unmapped 统计必须保留。

关键 artifact：classification counts、graft/host FASTQ 或 BAM、graft BAM/BAI、coverage/HDF5、QC summary。

## RNA-PDX

```text
ingest → qc_raw → prepare → separate → align → quantify → qc_final → publish
```

使用 RNA-specific species separation 与 STAR/GTF quantification。Snakemake compatibility assets use the immutable run's digest-bound Picard fixed-BAM and common Xenofilx filtered BAM/BAI contract; actual Snakemake execution remains required before it enters a production matrix.

关键 artifact：classification counts、graft RNA FASTQ/BAM、STAR BAM、gene counts/matrix、splicing tables、QC summary。

## Toolchain 对应关系

| 领域 | Modern | Legacy-equivalent | 主 parity |
|---|---|---|---|
| FASTQ QC | `fastqcx` | FastQC-compatible outputs; MultiQC remains an external-format compatibility consumer | 是 |
| PDX separation | `xenofilx` | XenofilteR-compatible path | 是 |
| paired BAM | `pairbam`/`bamdriver` | Paireads-compatible path | 是 |
| count matrix | `seq2mat` | legacy HTSeq matrix path | 是 |
| splicing | `matsrun` | legacy rMATS orchestrator | 是 |
| methylation | `methx` | legacy Methrix/R path | 是 |
| QC aggregate | `qctb` | legacy report path | 是 |
| CluBCpG/mHap/CCGG/insert length | 无完整一一替代 | legacy extension | 否，单列 |

`pairbam`/`bamdriver` 的适用范围固定为 BS（RRBS/WGBS）和 BS-PDX 的 paired-BAM 阶段；RNA-seq 与 RNA-PDX 不调度这些工具。七个 ready FASTQ 的比较配置必须由 Otter `config validate`/`config resolve` 生成 immutable `otter.run/v1` snapshot，再通过 `otter run --executor craftmake` 进入 Craftmake 的 plan/run/resume/status/log/report 控制面。共同 Trim Galore 是 modern/legacy 两条链路之间的固定预处理，不是本轮 trimming 比较轴；Snakemake 只保留单独的 compatibility/recovery 收尾。

“等价”要求输入、reference、工具参数和 artifact schema 可比，不要求内部实现或字节序列完全相同。

## Artifact contract

发布器必须在 publish 边界写入 `<run>/results/artifacts.json`，其 JSON Schema 为 `docs/schema/otter-artifacts-v1.schema.json`。manifest 绑定 immutable `run_id`、`run.yaml` digest、scenario/toolchain/executor/backend；所有 artifact path 相对 `results/`，并固定 checksum 与比较器。

`otter artifact publish <run>/run.yaml <declarations.json>` 只接受 `docs/schema/otter-artifact-declarations-v1.schema.json` 定义的 versioned declaration document；该 document 声明 artifact metadata 与 comparator，但 checksum 只能由 publisher 在读取稳定 regular file 后计算。

Phase-local sample validation uses `docs/schema/otter-sample-artifacts-validation-v1.schema.json` where migrated. `otter.sample-artifacts-validation/v1` records workflow/phase, validated sample dimensions, and every regular non-empty input file's path, media type, byte size, and SHA-256; it is not an artifact-publication manifest. PDX aggregate filtered BAM validation uses `docs/schema/otter-filtered-bam-validation-v1.schema.json`; `otter.filtered-bam-validation/v1` binds every graft sample to its regular filtered BAM/BAI paths, sizes, SHA-256 checksums, and mapped-read count.

Craftmake `BeaverBS/publish.yaml`、`BeaverRNA/publish.yaml`、`BeaverPDX/publish.yaml`、`BeaverRNASEQPDX/publish.yaml` 与 Snakemake compatibility completion path 都使用相同的 versioned declaration / immutable manifest APIs。BeaverBS stage Methrix HDF5、Bismark summary 与 QC workbook；BeaverRNA stage count/normalized expression matrices、QC workbook、`otter.rna-splicing-outcome/v1` outcome 与 status 为 `produced` 时的实际 splicing files；两条 PDX producer 都验证每个 graft BAM/BAI pair 和 mapped-read count，并分别 stage 自己的科学产物。RNA-PDX 使用相同 typed outcome contract，不再将 success marker 声明为 scientific artifact。`inst/rules/` 与 `inst/rules_legacy/` 的 Snakemake splicing rule 也生成相同的 typed outcome，且当 splicing 工具成功但没有产生文件时 fail closed。兼容发布仅在 `ExecuteAll()` 成功后运行、重新验证 immutable snapshot、在 results-local staging 构建 producer-owned directories、以 create-only manifest publication 完成并立即 checksum verify；当前只有 fixture/local contract evidence，不能视作真实 Snakemake/SLURM publish 证据。

```json
{
  "id": "methylation-matrix",
  "path": "methylation/matrix.h5",
  "media_type": "application/x-hdf5",
  "schema": "methrix-se/v1",
  "checksum": "sha256:...",
  "comparison": {
    "tier": "scientific",
    "comparator": "methrix-semantics/v1"
  }
}
```

`otter artifact verify <run>/run.yaml` 会重验 immutable snapshot、manifest identity 和所有已声明 artifact checksum。workflow 在生成真实 publish producer 前不得用 success marker 冒充 manifest。

比较 tier：

- `exact`：规范化文本、manifest、ID 集合等应完全一致。
- `structural`：文件格式、字段、维度、样本顺序、reference identity 一致。
- `scientific`：采用领域容差和 invariant。
- `informational`：日志、图像渲染、时间戳等只保存，不阻断。

## Gate 6 Comparison Ownership

Executor parity fixes `workflow.toolchain` and compares only Craftmake-default with explicit Snakemake compatibility using the same immutable canary input, reference digest, Bismark Rust/Bowtie2 runtime, parameters, and resolved resources. Toolchain parity fixes `execution.executor` and compares modern with legacy-equivalent implementations. Perl Bismark is excluded from both comparisons.

| Scenario | Blocking semantic/QC indicators | Structural blockers | Informational only |
|---|---|---|---|
| RRBS | CpG beta and coverage distributions, conversion metrics, trimmed-read retention, mapping rate | paired BAM/BAI, coverage/HDF5 schema, sample/reference identity | Bismark HTML and rendered QC assets |
| RNA-seq | feature/count concordance, assignment rate, splicing outcome and PSI where emitted, mapping rate | matrix headers/order, BAM/BAI, typed splicing outcome | rendered QC assets |
| BS-PDX | graft/host/ambiguous proportions, graft mapped reads, CpG beta/coverage, conversion metrics | graft BAM/BAI, classification rows bound to manifest, reference roles | Bismark/Qualimap HTML/PDF assets |
| RNA-PDX | graft/host/ambiguous proportions, graft count concordance, assignment rate, splicing outcome/PSI | graft BAM/BAI, classification rows, matrix and typed outcome identity | rendered QC assets |

A completed manifest is necessary but never enough for scientific parity. The current comparator registry must fail closed when an artifact declares an unavailable comparator. Byte or signature checks for HDF5, BAM/BAI, XLSX, HTML, or splicing metadata must be reported as their actual tier and cannot be represented as a tolerance-based scientific result until the relevant semantic comparator is implemented.

## 完成定义

一个 scenario 只有满足以下条件才视为 Craftmake 完整：

- Craftmake 和 Snakemake 均能从同一 `run.yaml` 适配执行；
- modern 与 legacy-equivalent 映射已冻结；
- artifact schema/comparator 已实现；
- cancel、失败注入、resume 在 sbatch 验证；
- 生产数据重复运行通过科学 parity；
- benchmark evidence 可追溯到 config、reference、workflow 和工具 digest。
