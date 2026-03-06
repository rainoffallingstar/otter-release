# 第三章：快速上手（三命令工作流）

> **本章你将学到：**
> - 用三条命令完成一次完整的生信分析
> - 了解每个命令做了什么
> - 查看分析进度

---

## 🗺️ 三命令工作流全貌

```
┌─────────────────────────────────────────────────────────┐
│                  xdxtools 三命令工作流                    │
└─────────────────────────────────────────────────────────┘

  xdxtools init      →    xdxtools create    →    xdxtools run
       │                        │                       │
       ▼                        ▼                       ▼
  建立项目骨架            扫描数据生成配置           提交分析任务
       │                        │                       │
  创建目录结构            自动识别样本               本地 or 集群
  复制流程文件            生成 config.yaml          运行 Snakemake
  复制 R/Python 脚本      分配唯一 Job ID           产出分析结果
```

---

## 🏗️ Step 1：初始化项目

```bash
xdxtools init my_project
```

这条命令会在当前目录下创建 `userspace/my_project/` 目录，并复制分析所需的所有流程文件。

### 生成的目录结构

```
userspace/my_project/
├── Snakefile                   ← 主工作流文件
├── config/                     ← 配置文件目录（第二步会生成 config.yaml）
├── rules/                      ← Snakemake 规则文件
├── Rscripts/                   ← R 分析脚本
├── envs/                       ← conda 环境定义文件
└── logs/                       ← 日志目录
```

### 常用参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `<项目名>` | 项目名称（必填） | `my_rrbs_project` |
| `--output <路径>` | 指定输出根目录（默认 `userspace/`） | `--output /data/projects` |

---

## 🔍 Step 2：创建分析配置

```bash
xdxtools create --fastq /path/to/fastq --mode RRBS --pdata samples.xlsx
```

这条命令会：
1. 扫描 `--fastq` 目录下的所有 FASTQ 文件
2. 读取 `--pdata` 样本信息表，匹配样本
3. 自动生成适配器序列
4. 生成唯一的 **Job ID**（一个 40 位十六进制字符串）
5. 在 `userspace/<jobid>/config/config.yaml` 生成配置文件

### 三种分析模式的命令示例

**RRBS（限制性甲基化测序）**：

```bash
xdxtools create \
    --fastq ./fastq \
    --mode RRBS \
    --pdata samples.xlsx \
    --species1 hg38
```

**WGBS（全基因组甲基化测序）**：

```bash
xdxtools create \
    --fastq ./fastq \
    --mode WGBS \
    --pdata samples.xlsx \
    --species1 hg38
```

**RNA-seq（转录组测序）**：

```bash
xdxtools create \
    --fastq ./fastq \
    --mode RNASEQ \
    --pdata samples.xlsx \
    --species1 hg38
```

### PDX 模式（双物种）额外说明

如果你的样本来自 PDX（人源肿瘤异种移植）模型，需要额外指定第二个物种：

```bash
xdxtools create \
    --fastq ./fastq \
    --mode RRBS \
    --pdata samples.xlsx \
    --species1 hg38 \
    --species2 mm10          # ← 添加第二物种（小鼠）
```

### 输出说明

命令运行后，终端会显示类似：

```
✅ Found 8 samples
✅ Adapters generated
📄 Config saved to: userspace/a3f8e1c2.../config/config.yaml
🆔 Job ID: a3f8e1c2d4b6f8a0e2c4d6f8a0b2c4d6e8f0a2b4
```

- **Job ID**：每次 `create` 都会生成一个唯一的 ID，用于追踪这次分析
- **config.yaml**：包含所有分析参数的配置文件，路径格式为 `userspace/<jobid>/config/config.yaml`

记下这个路径，下一步 `run` 命令需要用到。

---

## ▶️ Step 3：运行分析

```bash
xdxtools run --config userspace/a3f8e1c2.../config/config.yaml
```

> 将 `a3f8e1c2...` 替换为你实际的 Job ID。

### 本地运行（适合少量样本或测试）

```bash
xdxtools run \
    --config userspace/<jobid>/config/config.yaml \
    --engine local \
    --parallel-jobs 4
```

### SLURM 集群运行（适合大批量样本，推荐）

```bash
xdxtools run \
    --config userspace/<jobid>/config/config.yaml \
    --engine slurm \
    --slurm-partition your_partition \
    --parallel-jobs 10
```

> 将 `your_partition` 替换为你集群上的分区名称，可以用 `sinfo` 命令查看。

### 关键参数说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--config <路径>` | 配置文件路径（必填） | 无 |
| `--engine` | 执行引擎：`auto`/`slurm`/`local` | `auto` |
| `--parallel-jobs <N>` | 并发任务数（1=串行，>1=并行） | `2` |
| `--slurm-partition <名>` | SLURM 分区名称 | 无 |
| `--dry-run` | 只测试配置，不实际运行 | 否 |
| `--resume` / `-r` | 从上次中断处继续 | 否 |

### 先做干跑测试（强烈推荐新用户）

```bash
# 先用 --dry-run 验证配置，不会真正提交任务
xdxtools run \
    --config userspace/<jobid>/config/config.yaml \
    --dry-run
```

---

## 📡 Step 4：查看运行进度

```bash
xdxtools status
```

输出示例：

```
┌─────────────────────────────────────────────────┐
│ Job ID: a3f8e1c2...                             │
│ Mode: RRBS | Engine: SLURM                      │
│ Status: RUNNING                                 │
├──────┬──────────────────┬────────┬──────────────┤
│ Step │ Description      │ Status │ Progress     │
├──────┼──────────────────┼────────┼──────────────┤
│  1   │ Trimming & Align │ DONE   │ 8/8 samples  │
│  2   │ Methylation Call │ RUN    │ 3/8 samples  │
│  3   │ Report & QC      │ WAIT   │ -            │
└──────┴──────────────────┴────────┴──────────────┘
```

---

## 🎯 完整示例（从头到尾复制即可运行）

以下是一个完整的 RRBS 分析示例，你可以直接复制并修改路径后运行：

```bash
# ============================================
# xdxtools RRBS 完整分析示例
# ============================================

# 1. 进入你的工作目录
cd /data/my_analysis

# 2. 初始化项目（"rrbs_2024" 是你给这次分析起的名字）
xdxtools init rrbs_2024

# 3. 创建分析配置
#    --fastq: 你的 FASTQ 文件目录
#    --mode: 分析类型（RRBS/WGBS/RNASEQ）
#    --pdata: 样本信息表
#    --species1: 参考基因组（hg38/hg19/mm10）
xdxtools create \
    --fastq ./fastq \
    --mode RRBS \
    --pdata samples.xlsx \
    --species1 hg38

# 4. 查看生成的 Job ID 和配置文件路径
#    （记下终端输出中的 config.yaml 路径）
ls userspace/

# 5. 先做干跑测试（推荐）
xdxtools run \
    --config userspace/<your_jobid>/config/config.yaml \
    --dry-run

# 6. 正式运行（SLURM 集群）
xdxtools run \
    --config userspace/<your_jobid>/config/config.yaml \
    --engine slurm \
    --slurm-partition normal \
    --parallel-jobs 8

# 7. 查看进度
xdxtools status
```

> **提示：** 将 `<your_jobid>` 替换为第 3 步中输出的实际 Job ID，将 `normal` 替换为你集群的分区名。

---

**下一章：** [🔬 第四章：分析模式详解](04-analysis-modes.md)

**上一章：** [📁 第二章：数据准备](02-data-preparation.md)
