# 第六章：子工具参考手册

> **本章你将学到：**
> - 8 个子工具各自的功能和使用方法
> - 每个工具的典型使用场景和命令示例
> - 理解输出文件的格式和含义

---

## 工具清单

| # | 工具 | 功能简介 |
|---|------|---------|
| 1 | [🌿 enva](#enva) | 轻量级 conda 环境管理器 |
| 2 | [🧹 xenofilter](#xenofilter) | PDX 物种污染过滤 |
| 3 | [🔗 paireads](#paireads) | 配对 reads 过滤与恢复 |
| 4 | [📊 htseq2matrix](#htseq2matrix) | HTSeq 计数矩阵生成 |
| 5 | [💎 methrix-cli](#methrix-cli) | 甲基化 HDF5 处理 |
| 6 | [📈 qctb](#qctb) | 质控报告生成 |
| 7 | [⚡ fqc](#fqc) | FASTQ 质控（FastQC 替代品） |
| 8 | [🔀 gomats](#gomats) | RNA 可变剪接分析 |

---

## 🌿 enva {#enva}

**enva** 是一个以 **rattler 为主后端** 的 conda 环境管理器。xdxtools 默认通过 enva 管理依赖环境；如有历史 `conda` / `mamba` / `micromamba` 环境，也可以继续被发现、列举和接管。

### 验证安装

```bash
enva --version
enva list --detailed
```

### 主要命令

| 命令 | 说明 |
|------|------|
| `enva list` | 列出所有可访问环境，并合并同名环境 |
| `enva create --all` | 创建所有预定义环境 |
| `enva create --core` | 创建 `xdxtools-core` |
| `enva run <env> -- <cmd>` | 在指定环境中运行命令 |
| `enva install --name <env> <pkg...>` | 向环境安装一个或多个包 |
| `enva adopt --name <env>` | 将现有环境纳入 rattler ownership |
| `enva remove <env>` | 删除指定环境 |

### 使用示例

```bash
# 查看已有环境及其 owner/source
enva list --detailed

# 在 xdxtools-core 中运行 bismark
enva run xdxtools-core -- bismark --genome /ref/hg38 sample.fastq

# 批量创建 xdxtools 所需环境
enva create --all

# 向 xdxtools-core 安装额外工具
enva install --name xdxtools-core fastqc multiqc
```

---

## 🧹 xenofilter {#xenofilter}

**xenofilter** 用于 PDX（人源肿瘤异种移植）样本的**物种污染过滤**，将 reads 按物种来源分离。

### 验证安装

```bash
xenofilter --version
xenofilter --help
```

### 基本用法

```bash
# 基本模式：输入人类和小鼠的 BAM 文件，输出过滤后的人类 BAM
xenofilter \
    --human human_aligned.bam \
    --mouse mouse_aligned.bam \
    --output filtered_human.bam
```

### 含参考统计的高精度模式

```bash
xenofilter \
    --human human_aligned.bam \
    --mouse mouse_aligned.bam \
    --output filtered_human.bam \
    --min-score 3 \              # 最小比对分数差（人-鼠），默认 1
    --output-stats stats.txt     # 输出过滤统计报告
```

### 输出统计表解读

`stats.txt` 示例内容：

```
# xenofilter 过滤统计报告
Total reads:        10,000,000
Human-specific:      8,500,000  (85.0%)
Mouse-specific:      1,200,000  (12.0%)
Ambiguous:             300,000   (3.0%)
Retained (human):    8,500,000  (85.0%)
```

| 列名 | 含义 |
|------|------|
| Human-specific | 明确来自人类基因组的 reads |
| Mouse-specific | 明确来自小鼠基因组的 reads（被过滤） |
| Ambiguous | 无法区分物种来源的 reads（默认也被过滤） |
| Retained | 最终保留的 reads 数量 |

---

## 🔗 paireads {#paireads}

**paireads** 用于恢复经过过滤后打散的 paired-end reads 配对关系，是 PDX 流程中 xenofilter 之后的必要步骤。

> **背景**：xenofilter 过滤后，有时 R1 保留了但 R2 被过滤（或反之），导致配对不完整。paireads 负责重新筛选出完整的 read pairs。

### 验证安装

```bash
paireads --version
paireads --help
```

### 单 BAM 模式

```bash
# 从单个 BAM 文件中恢复完整配对
paireads \
    --input filtered.bam \
    --output paired_filtered.bam
```

### 双 BAM 模式（xenofilter 后使用）

```bash
# 同时处理 R1 和 R2 的 BAM，保留双方都保留的 reads
paireads \
    --r1 filtered_r1.bam \
    --r2 filtered_r2.bam \
    --output-r1 paired_r1.bam \
    --output-r2 paired_r2.bam
```

### `--coord-sort` 参数说明

`--coord-sort` 表示输入 BAM 文件按坐标排序（而非 read name 排序）：

```bash
# 如果 BAM 是坐标排序的，加此参数（会先自动按 read name 重新排序）
paireads --input filtered.bam --output paired.bam --coord-sort
```

---

## 📊 htseq2matrix {#htseq2matrix}

**htseq2matrix** 将多个 HTSeq-count 的输出文件合并为一个基因表达矩阵。

### 验证安装

```bash
htseq2matrix --version
htseq2matrix --help
```

### 基本用法

```bash
# 合并同一目录下所有 .count 文件
htseq2matrix \
    --input-dir ./counts \
    --output matrix_count.txt
```

### 人类 vs 小鼠模式

```bash
# 人类样本（过滤 ERCC 和非编码基因）
htseq2matrix \
    --input-dir ./counts \
    --species human \
    --output matrix_count.txt

# 小鼠样本
htseq2matrix \
    --input-dir ./counts \
    --species mouse \
    --output matrix_count.txt
```

### 输出文件说明

**matrix_count.txt**（原始计数矩阵）：

```
gene_id         sample1   sample2   sample3   sample4
ENSG00000001     1024       892       1156       784
ENSG00000002      345       412        298       501
```

**matrix_norm.txt**（归一化矩阵，TPM 或 RPKM）：

```
gene_id         sample1   sample2   sample3   sample4
ENSG00000001     12.34     10.78     14.02      9.56
ENSG00000002      4.21      5.03      3.64      6.12
```

---

## 💎 methrix-cli {#methrix-cli}

**methrix-cli** 是甲基化数据的处理工具，将 Bismark 输出的 CpG 覆盖文件转换为高效的 HDF5 格式（`.h5`），并提供 QC 报告生成功能。

> **什么是 HDF5？** HDF5 是一种高效的科学数据存储格式，可以存储海量甲基化位点数据，并支持快速随机访问。可以把它理解为"超大号的 Excel 文件"。

### 验证安装

```bash
methrix-cli --version
methrix-cli --help
```

### 完整处理流程

```bash
# Step 1：提取 CpG 位点，生成 HDF5 文件
methrix-cli extract-cpgs \
    --input ./bismark_coverage/*.cov.gz \
    --genome hg38 \
    --output methylation.h5

# Step 2：处理甲基化矩阵（过滤低覆盖度位点）
methrix-cli process \
    --input methylation.h5 \
    --min-coverage 10 \          # 至少 10x 覆盖深度的位点才保留
    --output methylation_filtered.h5

# Step 3：生成 QC 报告
methrix-cli qc-report \
    --input methylation_filtered.h5 \
    --output qc_report/
```

### HDF5 文件格式说明

HDF5 文件内部结构：

```
methylation.h5
├── /metadata            ← 样本信息和参数
├── /cpg_sites           ← 所有 CpG 位点的坐标
└── /samples/
    ├── /sample1/
    │   ├── /methylation    ← 甲基化比例（0-1）
    │   └── /coverage       ← 测序深度
    ├── /sample2/
    ...
```

### 在 R 中读取 HDF5 文件

```r
library(methrix)

# 读取 HDF5 文件
m <- load_HDF5_methrix("methylation_filtered.h5")

# 查看基本信息
print(m)

# 提取甲基化矩阵
meth_matrix <- get_matrix(m, type = "M")
coverage_matrix <- get_matrix(m, type = "C")
```

---

## 📈 qctb {#qctb}

**qctb**（Quality Control ToolBox）是质控报告生成器，整合多个工具的输出，生成一份 Excel 汇总报告。

### 验证安装

```bash
qctb --version
qctb --help
```

### BS-seq 模式（RRBS/WGBS）

```bash
qctb bsseq \
    --input-dir ./results \
    --samples sample1,sample2,sample3 \
    --output qc_report.xlsx
```

### RNA-seq 模式

```bash
qctb rnaseq \
    --input-dir ./results \
    --samples sample1,sample2,sample3 \
    --output qc_report.xlsx
```

### Excel 输出列说明

**BS-seq QC 报告列**：

| 列名 | 说明 |
|------|------|
| Sample | 样本名称 |
| Total_Reads | 原始 reads 总数 |
| Trimmed_Reads | 过滤后 reads 数 |
| Aligned_Reads | 成功比对的 reads 数 |
| Mapping_Rate | 比对率（%） |
| CpG_Coverage | CpG 位点覆盖数 |
| Median_Coverage | 中位覆盖深度 |
| Bisulfite_Conversion | 亚硫酸氢盐转化率（%） |
| Global_Methylation | 全局甲基化水平（%） |

**RNA-seq QC 报告列**：

| 列名 | 说明 |
|------|------|
| Sample | 样本名称 |
| Total_Reads | 原始 reads 总数 |
| Aligned_Reads | 成功比对 reads 数 |
| Mapping_Rate | 比对率（%） |
| Assigned_Reads | HTSeq 成功计数的 reads |
| Assignment_Rate | 计数成功率（%） |

---

## ⚡ fqc {#fqc}

**fqc** 是用 Rust 编写的 FASTQ 质控工具，是 FastQC 的替代品，速度更快，输出 `fastqc_data.txt` 格式（与 FastQC 兼容）并附带 Seqkit 风格的统计摘要。

### 验证安装

```bash
fqc --version
fqc --help
```

### 基本用法

```bash
# 对单个 FASTQ 文件进行质控
fqc --input sample1_R1.fastq.gz --output sample1_R1_qc/

# 同时处理 R1 和 R2
fqc --input sample1_R1.fastq.gz sample1_R2.fastq.gz --output sample1_qc/
```

### 批量处理（使用 GNU parallel）

```bash
# 并行处理所有样本（4 个并发）
ls fastq/*_R1.fastq.gz | \
    parallel -j 4 'fqc --input {} --output qc/$(basename {_R1.fastq.gz}_qc)'

# 或使用 while 循环
for f in fastq/*_R1.fastq.gz; do
    sample=$(basename ${f/_R1.fastq.gz/})
    fqc --input $f fastq/${sample}_R2.fastq.gz --output qc/${sample}/
done
```

### 与 MultiQC 整合

fqc 输出的 `fastqc_data.txt` 与 FastQC 格式兼容，可以直接用 MultiQC 汇总：

```bash
# 运行 fqc
for f in fastq/*_R1.fastq.gz; do
    sample=$(basename ${f/_R1.fastq.gz/})
    fqc --input $f --output qc/${sample}/
done

# 用 MultiQC 汇总所有 fqc 报告
multiqc qc/ --outdir multiqc_report/
```

---

## 🔀 gomats {#gomats}

**gomats** 是 rMATS（RNA 可变剪接分析工具）的 Go 语言编排器，简化了 rMATS 的参数配置和批量运行。

### 验证安装

```bash
gomats --version
gomats --help
```

### 基本用法

```bash
# 分析两组样本的可变剪接差异
gomats \
    --bam-group1 ctrl1.bam,ctrl2.bam \
    --bam-group2 treat1.bam,treat2.bam \
    --gtf /ref/hg38/annotation.gtf \
    --output rmats_output/
```

### PDX 模式

```bash
# PDX 样本（已经 xenofilter 过滤后的人源 BAM）
gomats \
    --bam-group1 human_ctrl1.bam,human_ctrl2.bam \
    --bam-group2 human_treat1.bam,human_treat2.bam \
    --gtf /ref/hg38/annotation.gtf \
    --output rmats_output/ \
    --pdx                    # 启用 PDX 模式（调整内部参数）
```

### 输出目录结构

```
rmats_output/
├── SE.MATS.JC.txt          ← 外显子跳过事件（Skipped Exon）
├── A5SS.MATS.JC.txt        ← 可变 5' 剪接位点
├── A3SS.MATS.JC.txt        ← 可变 3' 剪接位点
├── MXE.MATS.JC.txt         ← 互斥外显子
├── RI.MATS.JC.txt          ← 内含子保留
├── summary.txt             ← 各类型事件统计摘要
└── logs/                   ← 运行日志
```

**输出文件列说明（以 SE.MATS.JC.txt 为例）**：

| 列名 | 含义 |
|------|------|
| GeneID | 基因 Ensembl ID |
| geneSymbol | 基因名 |
| chr / strand | 染色体和链方向 |
| exonStart/End | 跳过外显子的坐标 |
| IncLevel1/2 | 两组的外显子包含率 |
| IncLevelDifference | 两组包含率的差值 |
| PValue | 统计显著性 p 值 |
| FDR | 校正后 p 值 |

---

**下一章：** [❓ 第七章：常见问题与排查](07-faq.md)

**上一章：** [⚙️ 第五章：高级用法](05-advanced-usage.md)
