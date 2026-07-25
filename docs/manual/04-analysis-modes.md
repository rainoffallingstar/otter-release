# 第四章：分析模式详解

## 共同运行模型

四种模式都由 `otter` 管理配置与任务。当前生产执行路径使用 Snakemake；`craftmake` 是迁移中的 Go 替代执行层，尚未完成四种模式的单轨接入。

## RRBS

RRBS（Reduced Representation Bisulfite Sequencing）用于富集 CpG 区域的甲基化测序。

```bash
otter create --fastq ./fastq --mode RRBS --pdata samples.xlsx --species1 hg38
```

主要数据链：

```text
FASTQ → fastqcx/Trim Galore → Bismark → methx → qctb
```

FastQC 兼容结果、Bismark coverage、Methrix 领域语义和 HDF5 名称保持不变。

## WGBS

WGBS（Whole Genome Bisulfite Sequencing）覆盖全基因组 CpG 位点，流程与 RRBS 类似但数据量和资源需求更高。

```bash
otter create --fastq ./fastq --mode WGBS --pdata samples.xlsx --species1 hg38
```

主要输出包括 BAM、Bismark coverage、甲基化矩阵/HDF5 和 QC 报告。

## RNA-seq

RNA-seq 当前使用 STAR 比对和 HTSeq 计数，由 `seq2mat` 合并矩阵，并可由 `matsrun` 编排 rMATS 可变剪接分析。

```bash
otter create --fastq ./fastq --mode RNASEQ --pdata samples.xlsx --species1 hg38
```

主要数据链：

```text
FASTQ → fastqcx → STAR → HTSeq → seq2mat → matsrun/rMATS → qctb
```

## PDX

指定第二物种后自动进入 PDX 路径：

```bash
otter create \
  --fastq ./fastq \
  --mode RRBS \
  --pdata samples.xlsx \
  --species1 hg38 \
  --species2 mm10
```

PDX 额外数据链：

```text
双物种比对 → xenofilx → pairbam → bamdriver → 下游 RRBS/WGBS/RNA-seq
```

- `xenofilx` 比较 graft/host 证据并过滤物种污染。
- `pairbam` 保留完整 paired-end 读段。
- `bamdriver` 提供共享 BAM 低层操作。

## 输出目录

```text
my_project/userspace/<jobid>/
├── config/       # OtterConfig YAML；源码旧类型名仍可能存在
├── workflow/     # 中间产物
├── analysis/     # 矩阵、报告和分析结果
└── log/          # workflow/SLURM 日志
```

具体文件名受当前 Snakemake 规则和算子版本约束。迁移到 `craftmake` 时必须验证关键产物等价，而不是仅验证命令成功。

下一章：[高级用法](05-advanced-usage.md)
