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

这条命令会在当前目录下创建 `my_project/` 项目目录，并复制分析所需的流程资产。

### 生成的目录结构

```text
my_project/
├── config/                     ← 项目级配置目录
├── data/                       ← 输入数据目录
├── envs/                       ← conda 环境定义文件
├── inst/                       ← 参考资源目录
├── R/                          ← R/Python 脚本
├── rules/                      ← Snakemake 规则文件
├── userspace/                  ← 每次 create 生成一个 job 子目录
└── *.snakemake                 ← 主工作流文件
```

### 常用参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `<项目名>` | 项目名称（必填） | `my_rrbs_project` |
| `--legacy` | 使用旧版 rules 目录布局 | `xdxtools init my_project --legacy` |

---

## 🔍 Step 2：创建分析配置

```bash
xdxtools create --fastq /path/to/fastq --mode RRBS --pdata samples.xlsx --output my_project/userspace --jobid demo_rrbs
```

这条命令会：
1. 扫描 `--fastq` 目录下的所有 FASTQ 文件
2. 读取 `--pdata` 样本信息表，匹配样本
3. 自动生成适配器序列
4. 使用 `--jobid` 指定 Job ID；未指定时会自动生成唯一 ID
5. 在 `my_project/userspace/<jobid>/config/config.yaml` 生成配置文件

### 三种分析模式的命令示例

**RRBS（限制性甲基化测序）**：

```bash
xdxtools create \
    --fastq ./fastq \
    --mode RRBS \
    --pdata samples.xlsx \
    --output rrbs_2024/userspace \
    --jobid demo_rrbs \
    --species1 hg38
```

**WGBS（全基因组甲基化测序）**：

```bash
xdxtools create \
    --fastq ./fastq \
    --mode WGBS \
    --pdata samples.xlsx \
    --output rrbs_2024/userspace \
    --jobid demo_rrbs \
    --species1 hg38
```

**RNA-seq（转录组测序）**：

```bash
xdxtools create \
    --fastq ./fastq \
    --mode RNASEQ \
    --pdata samples.xlsx \
    --output rrbs_2024/userspace \
    --jobid demo_rrbs \
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
📄 Config saved to: my_project/userspace/demo_rrbs/config/config.yaml
🆔 Job ID: demo_rrbs
```

- **Job ID**：默认会自动生成唯一 ID；如使用 `--jobid`，则采用你指定的值（例如 `demo_rrbs`）
- **config.yaml**：包含所有分析参数的配置文件，路径格式为 `my_project/userspace/<jobid>/config/config.yaml`

记下这个路径，下一步 `run` 命令需要用到。

---

## ▶️ Step 3：运行分析

```bash
xdxtools run --config my_project/userspace/demo_rrbs/config/config.yaml
```

> 如果你没有显式设置 `--jobid demo_rrbs`，请将这里替换为实际输出的 Job ID。

### 本地运行（适合少量样本或测试）

```bash
xdxtools run \
    --config my_project/userspace/<jobid>/config/config.yaml \
    --engine local \
    --parallel-jobs 4
```

### SLURM 集群运行（适合大批量样本，推荐）

```bash
xdxtools run \
    --config my_project/userspace/<jobid>/config/config.yaml \
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
    --config my_project/userspace/<jobid>/config/config.yaml \
    --dry-run
```

---

## 📡 Step 4：查看运行进度

```bash
xdxtools status my_project/userspace/<jobid>
```

输出示例：

```text
Workflow Status
================
Job ID:      demo_rrbs
Status:      Running

Configuration:
  Mode:      RRBS
  Species:   hg38
  Engine:    slurm

Steps:
  ✓ Step 1 (quality_control): completed
  → Step 2 (alignment): running
  ○ Step 3 (methylation_calling): pending

Progress: 1 completed, 1 running, 1 pending
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
    --output rrbs_2024/userspace \
    --jobid demo_rrbs \
    --species1 hg38

# 4. 确认配置文件已经生成
ls rrbs_2024/userspace/demo_rrbs/config/

# 5. 先做干跑测试（推荐）
xdxtools run \
    --config rrbs_2024/userspace/demo_rrbs/config/config.yaml \
    --dry-run

# 6. 正式运行（SLURM 集群）
xdxtools run \
    --config rrbs_2024/userspace/demo_rrbs/config/config.yaml \
    --engine slurm \
    --slurm-partition normal \
    --parallel-jobs 8

# 7. 查看进度
xdxtools status rrbs_2024/userspace/demo_rrbs
```

> **提示：** 如果你没有显式设置 `--jobid demo_rrbs`，请将示例中的 `demo_rrbs` 替换为实际输出的 Job ID；`normal` 需要替换为你集群的分区名。

---

**下一章：** [🔬 第四章：分析模式详解](04-analysis-modes.md)

**上一章：** [📁 第二章：数据准备](02-data-preparation.md)
