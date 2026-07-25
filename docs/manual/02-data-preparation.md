# 第二章：数据准备

> **本章你将学到：**
> - FASTQ 文件的正确命名规范
> - 如何制作样本信息表（pdata）
> - 推荐的项目目录结构

---

## 📁 FASTQ 文件命名规范

FASTQ 文件（测序原始数据）必须按照特定规则命名，otter 才能正确识别成对的读段（R1 和 R2）。

### ✅ 正确命名示例

| 格式 | R1 文件 | R2 文件 |
|------|---------|---------|
| 下划线格式（推荐） | `sample1_R1.fastq.gz` | `sample1_R2.fastq.gz` |
| 数字格式（也支持） | `sample1_1.fastq.gz` | `sample1_2.fastq.gz` |

> **提示：** `fastq.gz` 表示经过 gzip 压缩的 FASTQ 文件，这是测序公司最常提供的格式。

### ❌ 常见错误命名

| 错误命名 | 问题所在 | 修正方法 |
|---------|---------|---------|
| `sample1.R1.fastq.gz` | 用了点号分隔 | 改为下划线 `_R1` |
| `sample1_r1.fastq.gz` | 小写 r1 | 改为大写 `_R1` |
| `sample1_read1.fastq.gz` | 不认识 read1 | 改为 `_R1` |
| `Sample1_R1.fastq.gz` | 大写 S（与 pdata 大小写不一致）| 保持与 pdata 中 sampleid 大小写一致 |
| `sample1.fastq.gz` | 缺少 R1/R2 后缀 | 补充 `_R1` 或 `_R2` |

### 🔄 批量重命名命令

如果你的文件命名不规范，可以用以下命令批量修改：

```bash
# 将 .R1. 改为 _R1.（点号改下划线）
for f in *.R1.fastq.gz; do
    mv "$f" "${f/.R1./_R1.}"
done

# 将小写 r1 改为大写 R1
for f in *_r1.fastq.gz; do
    mv "$f" "${f/_r1.fastq.gz/_R1.fastq.gz}"
done

# 查看重命名后的结果
ls -la *.fastq.gz
```

### 📌 命名规则总结

```
样本ID_R1.fastq.gz   ← 正向读段
样本ID_R2.fastq.gz   ← 反向读段
   │
   └── 这里的"样本ID"要与样本信息表中的 sampleid 完全一致！
```

---

## 📊 样本信息表（pdata）制作指南

样本信息表（pdata）是一个 Excel 或 CSV 文件，记录每个样本的基本信息。otter 会读取这个表来匹配 FASTQ 文件并生成分析配置。

### 必填列

| 列名 | 说明 | 示例值 |
|------|------|--------|
| `sampleid` | 样本编号，必须与 FASTQ 文件名前缀完全一致 | `sample1`, `ctrl_01` |

### 选填列

| 列名 | 说明 | 示例值 | 什么情况下需要 |
|------|------|--------|--------------|
| `inline_barcode_sequence` | 内联条形码序列 | `ATCACG` | 使用了 inline barcode 的 RRBS/WGBS 实验 |
| `condition` | 实验条件/处理组 | `treated`, `control` | RNA-seq 差异分析 |
| `sample_group` | 样本分组（优先级高于 condition） | `Group_A`, `Group_B` | 有多分组设计时 |

### 支持中文列名

otter 支持中文列名，会自动识别并转换：

| 中文列名 | 等价英文列名 |
|---------|------------|
| 样本编号 | `sampleid` |
| 样本ID | `sampleid` |
| 条件 | `condition` |
| 样本分组 | `sample_group` |
| 分组 | `sample_group` |

### 📝 样本信息表示例

**Excel 格式（推荐）**：

| sampleid | inline_barcode_sequence | condition | sample_group |
|----------|------------------------|-----------|-------------|
| sample1 | ATCACG | treated | Group_A |
| sample2 | CGATGT | treated | Group_A |
| sample3 | TTAGGC | control | Group_B |
| sample4 | TGACCA | control | Group_B |

**CSV 格式**：

```csv
sampleid,inline_barcode_sequence,condition,sample_group
sample1,ATCACG,treated,Group_A
sample2,CGATGT,treated,Group_A
sample3,TTAGGC,control,Group_B
sample4,TGACCA,control,Group_B
```

> **提示：** 测序公司通常会提供一个样本信息表，你只需要参照上面的格式整理即可。

### 关于 inline_barcode_sequence（适配器）的说明

**什么是 inline barcode？**

在某些甲基化测序（RRBS/WGBS）实验中，会在引物上加入一小段独特的 DNA 序列（barcode），用来区分不同样本。这段序列就叫做 inline barcode。

**如何填写：**

- 如果你的实验**使用了** inline barcode：填入实际的 barcode 碱基序列（如 `ATCACG`）
- 如果**没有**使用 inline barcode：留空或不填这一列

**留空时的行为：**

如果该列为空，otter 会生成配置 `NO_ADAPTER_CAL_USE_DEFAULT`，表示使用 Trim Galore 的默认适配器检测模式——这对大多数没有 inline barcode 的样本来说完全够用。

---

## 📂 推荐的项目目录结构

在开始分析之前，建议按照以下结构组织你的文件：

```
my_project/               ← 你的工作目录
├── fastq/                ← 把所有 FASTQ 文件放这里
│   ├── sample1_R1.fastq.gz
│   ├── sample1_R2.fastq.gz
│   ├── sample2_R1.fastq.gz
│   ├── sample2_R2.fastq.gz
│   └── ...
└── samples.xlsx          ← 样本信息表放这里
```

> **提示：** 你不需要手动创建 `userspace/` 等目录，`otter init` 命令会自动为你创建分析所需的所有目录。

### 检查数据准备是否完成

在进入下一步之前，请核查：

```bash
# 进入项目目录
cd my_project

# 查看 FASTQ 文件数量（应为样本数的 2 倍）
ls fastq/*.fastq.gz | wc -l

# 查看 FASTQ 文件列表
ls fastq/

# 确认每个样本都有 R1 和 R2
for sample in $(ls fastq/*_R1.fastq.gz | sed 's/_R1.fastq.gz//'); do
    basename_s=$(basename $sample)
    if [ ! -f "fastq/${basename_s}_R2.fastq.gz" ]; then
        echo "⚠️  缺少 R2: ${basename_s}"
    fi
done
echo "✅ 配对检查完成"
```

---

**下一章：** [🚀 第三章：快速上手](03-quickstart.md)

**上一章：** [📦 第一章：安装指南](01-installation.md)
