# Otter Workflow Catalog

> 状态：目标 catalog，双轨实现已完成（Craftmake + Snakemake）。每个场景已拥有完整的 workflow YAML 和 snakefile。

## 实施状态 (2026-07)

| Scenario | Craftmake workflow YAML | Snakemake snakefile | Phase chain | 验证状态 |
|---|---|---|---|---|
| RRBS | `craftmake/workflows/BeaverBS/step1–3.yaml` (+ checks) | `inst/snakefiles/BeaverBS*.snakemake` | 7 phases | ✅ |
| WGBS | 同上（共享 BeaverBS） | 同上 | 7 phases | ✅ |
| RNA-seq | `craftmake/workflows/BeaverRNA/step1–2.yaml` (+ check) | `inst/snakefiles/BeaverRNA*.snakemake` | 7 phases (no step3) | ✅ |
| BS-PDX | `craftmake/workflows/BeaverPDX/step1–3.yaml` (+ checks) | `inst/snakefiles/BeaverPDX*.snakemake` | 8 phases (+ separate) | ✅ |
| RNA-PDX | `craftmake/workflows/BeaverRNASEQPDX/step1–3.yaml` (+ checks) | `inst/snakefiles/BeaverRNASEQPDX*.snakemake` | 8 phases (+ separate) | ✅ |

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

使用 RNA-specific species separation 与 STAR/GTF quantification。Snakemake 兼容资产中历史 `XenofilteR_RNA.smk` 引用在进入生产 matrix 前必须修正并锁定 digest。

关键 artifact：classification counts、graft RNA FASTQ/BAM、STAR BAM、gene counts/matrix、splicing tables、QC summary。

## Toolchain 对应关系

| 领域 | Modern | Legacy-equivalent | 主 parity |
|---|---|---|---|
| FASTQ QC | `fastqcx` | FastQC/MultiQC compatible path | 是 |
| PDX separation | `xenofilx` | XenofilteR-compatible path | 是 |
| paired BAM | `pairbam`/`bamdriver` | Paireads-compatible path | 是 |
| count matrix | `seq2mat` | legacy HTSeq matrix path | 是 |
| splicing | `matsrun` | legacy rMATS orchestrator | 是 |
| methylation | `methx` | legacy Methrix/R path | 是 |
| QC aggregate | `qctb` | legacy report path | 是 |
| CluBCpG/mHap/CCGG/insert length | 无完整一一替代 | legacy extension | 否，单列 |

“等价”要求输入、reference、工具参数和 artifact schema 可比，不要求内部实现或字节序列完全相同。

## Artifact contract

每个发布 artifact 至少记录：

```yaml
id: methylation-matrix
path: results/methylation/matrix.h5
media_type: application/x-hdf5
schema: methrix-se/v1
checksum: sha256:...
comparison:
  tier: scientific
  comparator: methrix-semantics/v1
```

比较 tier：

- `exact`：规范化文本、manifest、ID 集合等应完全一致。
- `structural`：文件格式、字段、维度、样本顺序、reference identity 一致。
- `scientific`：采用领域容差和 invariant。
- `informational`：日志、图像渲染、时间戳等只保存，不阻断。

## 完成定义

一个 scenario 只有满足以下条件才视为 Craftmake 完整：

- Craftmake 和 Snakemake 均能从同一 `run.yaml` 适配执行；
- modern 与 legacy-equivalent 映射已冻结；
- artifact schema/comparator 已实现；
- cancel、失败注入、resume 在 sbatch 验证；
- 生产数据重复运行通过科学 parity；
- benchmark evidence 可追溯到 config、reference、workflow 和工具 digest。
