# 第四章：分析模式详解

> **本章你将学到：**
> - RRBS、WGBS、RNA-seq、PDX 四种模式各自的应用场景
> - 每种模式的分析步骤和输出文件
> - 如何选择适合你的分析模式

---

## 如何选择分析模式？

```
你的实验是什么类型？
        │
        ├── 甲基化测序（亚硫酸氢盐处理的）
        │       │
        │       ├── 只测部分 CpG 位点（酶切富集）→ 🧬 RRBS
        │       └── 测全基因组所有 CpG 位点      → 🌐 WGBS
        │
        ├── 基因表达测序（mRNA）                 → 🧪 RNA-seq
        │
        └── 人源肿瘤种到小鼠体内的样本          → 🐭 PDX 模式
                （上面任一类型 + species2 参数）
```

---

## 🧬 RRBS — 限制性代表性重亚硫酸氢盐测序

### 白话解释

RRBS（Reduced Representation Bisulfite Sequencing）是一种**甲基化测序的经济型方案**。

通俗来说：DNA 甲基化就像给基因加了一把"锁"，影响基因是否表达。RRBS 用限制性内切酶（MspI）切割 DNA，只保留富含 CpG 的片段来测序，这样可以用更少的测序量覆盖更多有意义的 CpG 位点，大大降低成本。

**适用场景**：
- 临床样本（样本量大，成本敏感）
- 探索性研究（初步筛选差异甲基化区域）
- 需要高 CpG 覆盖度但预算有限

### 命令示例

```bash
xdxtools create \
    --fastq ./fastq \
    --mode RRBS \
    --pdata samples.xlsx \
    --species1 hg38
```

### 分析步骤（共 3 步）

| 步骤 | 名称 | 主要工作 | 主要输出 |
|------|------|---------|---------|
| Step 1 | 质控与比对 | Trim Galore 去接头，Bismark 比对到参考基因组 | BAM 文件、比对统计报告 |
| Step 2 | 甲基化提取 | Bismark methylation extractor 提取甲基化信息 | CpG 覆盖文件（`.cov.gz`） |
| Step 3 | 报告与 QC | methrix-cli 汇总，qctb 生成质控报告 | 甲基化矩阵、QC Excel 报告 |

### 主要输出目录

```text
my_project/userspace/<jobid>/
├── config/                 ← `config.yaml`
├── data/                   ← 预处理后的输入数据
├── workflow/
│   ├── trim/               ← Trim Galore 结果
│   ├── bsmap/              ← Bismark/BSMAP 比对与中间文件
│   ├── mCall/              ← 甲基化调用结果
│   └── QC/                 ← 流程级质控产物
├── analysis/               ← 汇总矩阵、DMR、QC 摘要等分析结果
└── log/                    ← 运行日志与 SLURM 输出
```

---

## 🌐 WGBS — 全基因组甲基化测序

### 白话解释

WGBS（Whole Genome Bisulfite Sequencing）测量全基因组所有 CpG 位点的甲基化状态，是甲基化分析的"金标准"。代价是测序成本更高，数据量更大。

### WGBS 与 RRBS 的区别

| 特性 | RRBS | WGBS |
|------|------|------|
| 覆盖范围 | ~3-5% 的 CpG 位点（富集区域） | >90% 的 CpG 位点（全基因组） |
| 测序深度要求 | 较低（~30M reads/样本） | 较高（~100M reads/样本） |
| 成本 | 较低 | 较高 |
| 适配器处理 | 无需 inline barcode 处理 | 无需 inline barcode 处理 |
| 分析步骤 | 3 步（与 WGBS 相同） | 3 步 |

### 命令示例

```bash
xdxtools create \
    --fastq ./fastq \
    --mode WGBS \
    --pdata samples.xlsx \
    --species1 hg38
```

> **注意**：WGBS 模式下，`inline_barcode_sequence` 列如果为空，不会添加任何前缀（与 RRBS 的处理略有不同）。

---

## 🧪 RNA-seq — 转录组测序

### 白话解释

RNA-seq 通过测量 mRNA 的丰度来了解基因的表达情况，常用于：
- 找出两组样本之间表达量有显著差异的基因（差异表达分析）
- 了解某种处理或疾病状态下哪些信号通路被激活

### 命令示例

```bash
xdxtools create \
    --fastq ./fastq \
    --mode RNASEQ \
    --pdata samples.xlsx \
    --species1 hg38
```

### 分析步骤（共 2 步）

| 步骤 | 名称 | 主要工作 | 主要输出 |
|------|------|---------|---------|
| Step 1 | 比对与计数 | STAR 比对，HTSeq-count 计数 | BAM 文件、每样本计数文件 |
| Step 2 | 矩阵与 QC | htseq2matrix 合并矩阵，qctb 生成报告 | 表达矩阵、QC Excel 报告 |

### 主要输出目录

```text
my_project/userspace/<jobid>/
├── config/
├── workflow/
│   ├── star/               ← STAR 比对结果
│   └── htseq/              ← HTSeq 计数中间文件
├── analysis/
│   ├── counts/             ← 计数矩阵及衍生文件
│   └── DEG/                ← 差异分析结果
└── log/
```

**矩阵格式说明**：

```
gene_id    sample1   sample2   sample3   sample4
ENSG000001    245       312       198       401
ENSG000002   1023      887       1156      954
...
```

---

## 🐭 PDX — 人源肿瘤异种移植

### 白话解释

PDX（Patient-Derived Xenograft）模型是把**人的肿瘤组织移植到免疫缺陷小鼠体内**生长，用来研究肿瘤生物学和测试药物疗效。

测序 PDX 样本时，序列读段（reads）来自**人和小鼠两个物种**混合在一起。分析的第一步必须把人源读段和鼠源读段分开，才能进行后续分析。这个步骤叫做**物种污染过滤**（xenofilter + paireads）。

### 启用 PDX 模式

只需在 `create` 命令中加上 `--species2` 参数，xdxtools 会**自动启用** PDX 模式：

```bash
# PDX + RRBS（人类肿瘤甲基化）
xdxtools create \
    --fastq ./fastq \
    --mode RRBS \
    --pdata samples.xlsx \
    --species1 hg38 \
    --species2 mm10

# PDX + WGBS
xdxtools create \
    --fastq ./fastq \
    --mode WGBS \
    --pdata samples.xlsx \
    --species1 hg38 \
    --species2 mm10

# PDX + RNA-seq（人类肿瘤转录组）
xdxtools create \
    --fastq ./fastq \
    --mode RNASEQ \
    --pdata samples.xlsx \
    --species1 hg38 \
    --species2 mm10
```

### 常见物种参数

| 物种 | 参数值 |
|------|--------|
| 人类 (hg38) | `hg38` |
| 人类 (hg19) | `hg19` |
| 小鼠 (mm10) | `mm10` |
| 小鼠 (mm39) | `mm39` |

### PDX 分析的额外步骤

PDX 模式在标准流程中**额外插入物种过滤步骤**：

```
正常 RRBS 流程：  Step1（比对）→ Step2（甲基化提取）→ Step3（报告）

PDX RRBS 流程：   Step1（双基因组比对）
                       ↓
                  xenofilter（过滤鼠源 reads）
                       ↓
                  paireads（恢复配对 reads）
                       ↓
                  Step2（甲基化提取，仅人源 reads）
                       ↓
                  Step3（报告）
```

**xenofilter** 的工作原理：将 reads 同时比对到人类和小鼠基因组，只保留比对到人类基因组更好的 reads。

**paireads** 的工作原理：xenofilter 过滤后，配对 reads 可能被打散（一条保留，一条丢失），paireads 负责恢复完整的 read pair。

### 主要输出目录（PDX 特有）

```text
my_project/userspace/<jobid>/
├── workflow/
│   ├── bsmap/<species1>/   ← 人源比对结果
│   ├── bsmap/<species2>/   ← 鼠源比对结果
│   ├── bsmap/Filtered_bams/← xenofilter / paireads 后的 BAM
│   └── ...                 ← 后续步骤与普通模式一致
├── analysis/
└── log/
```

**filter_stats.txt 示例**：

```
Sample      Total_Reads   Human_Reads   Mouse_Reads   Human_Ratio
sample1     10000000      8500000       1500000       85.0%
sample2     12000000      9800000       2200000       81.7%
```

---

**下一章：** [⚙️ 第五章：高级用法](05-advanced-usage.md)

**上一章：** [🚀 第三章：快速上手](03-quickstart.md)
