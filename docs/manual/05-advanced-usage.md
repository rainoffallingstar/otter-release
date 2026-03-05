# 第五章：高级用法

> **本章你将学到：**
> - 精细控制 SLURM 集群资源
> - 并行化策略与自动化规则
> - 断点续传（任务中断后如何恢复）
> - TUI 交互界面的使用
> - 手动编辑配置文件
> - 配置文件验证

---

## 🏃 SLURM 集群资源精细调控

### 全局参数 vs 分步骤参数

xdxtools 支持两个层次的资源配置：

**全局参数**（应用到所有步骤）：

```bash
xdxtools run \
    --config config.yaml \
    --engine slurm \
    --slurm-partition normal \           # SLURM 分区
    --parallel-jobs 10 \                 # 最大并发任务数
    --load-ratio 0.8                     # 负载比例（0.0-1.0）
```

**分步骤参数**（在 config.yaml 中设置，覆盖全局）：

```yaml
# config.yaml 中的引擎配置部分
engine:
  type: slurm
  partition: normal
  parallel_jobs: 10
  steps:
    step1:
      partition: fat          # 大内存分区（比对步骤消耗内存多）
      cpus: 16
      memory: "64G"
      time: "24:00:00"
    step2:
      partition: normal
      cpus: 8
      memory: "32G"
      time: "12:00:00"
    step3:
      partition: normal
      cpus: 4
      memory: "16G"
      time: "4:00:00"
```

### 各分析模式推荐资源配置

**RRBS / WGBS**：

| 步骤 | 描述 | 推荐 CPU | 推荐内存 | 推荐时间 |
|------|------|---------|---------|---------|
| Step 1 | 质控 + 比对 | 16 | 64G | 24h |
| Step 2 | 甲基化提取 | 8 | 32G | 12h |
| Step 3 | 报告生成 | 4 | 16G | 4h |

**RNA-seq**：

| 步骤 | 描述 | 推荐 CPU | 推荐内存 | 推荐时间 |
|------|------|---------|---------|---------|
| Step 1 | STAR 比对 + HTSeq 计数 | 16 | 48G | 12h |
| Step 2 | 矩阵合并 + QC | 4 | 8G | 2h |

### `--load-ratio` 的意义

`--load-ratio` 控制系统负载的阈值（取值 0.0 到 1.0）：

- `0.8`（默认）：当系统负载达到 80% 时，暂停提交新任务
- `1.0`：尽可能填满集群资源
- `0.5`：保守模式，只使用一半资源

```bash
# 保守模式，不要占用太多集群资源
xdxtools run --config config.yaml --engine slurm --load-ratio 0.5
```

---

## ⚡ 并行化策略

xdxtools 会根据样本数量自动选择并行策略：

| 样本数 | 默认策略 | 说明 |
|--------|---------|------|
| < 5 | 串行（Sequential） | 按顺序逐个处理，日志清晰易排查 |
| ≥ 5 | SLURM Job Array 或本地 worker pool | 并行处理，大幅缩短总时间 |

### SLURM Job Array 模式

当样本数 ≥ 5 时，Step 2 和 Step 3 会以 **SLURM Job Array** 的方式提交，每个样本作为一个独立的 array task：

```bash
# xdxtools 自动生成类似这样的 SLURM 命令：
sbatch --array=1-8 submit_step2.sh
```

**优点**：
- 所有样本同时处理，大幅缩短总时间
- SLURM 自动管理资源分配
- 单个样本失败不影响其他样本

### 本地 Worker Pool 模式

使用 `--engine local` 时，通过 worker pool 并行处理：

```bash
xdxtools run \
    --config config.yaml \
    --engine local \
    --parallel-jobs 4    # 同时处理 4 个样本
```

---

## 🔄 断点续传

### 何时使用 `--resume`

以下情况可以使用断点续传，从上次中断处继续，**避免重复已完成的步骤**：

- 集群节点故障导致任务中断
- 网络超时导致连接断开
- 手动取消后（`Ctrl+C`）需要继续
- SLURM 队列等待超时

### 使用方法

```bash
# 在原来的命令后加 --resume 即可
xdxtools run \
    --config userspace/<jobid>/config/config.yaml \
    --engine slurm \
    --slurm-partition normal \
    --resume              # ← 添加这个参数
```

或者简写：

```bash
xdxtools run --config config.yaml -r
```

### 状态文件说明

xdxtools 在每次运行时会维护一个状态文件，记录哪些步骤已完成：

```
userspace/<jobid>/
└── .xdxtools_state.json    ← 状态文件（自动维护，不要手动修改）
```

状态文件内容示例：

```json
{
  "job_id": "a3f8e1c2...",
  "steps": {
    "step1": {"status": "COMPLETED", "finished_at": "2024-03-01T10:30:00Z"},
    "step2": {"status": "FAILED", "failed_samples": ["sample3"]},
    "step3": {"status": "PENDING"}
  }
}
```

恢复运行时，xdxtools 会跳过 `COMPLETED` 的步骤，从 `FAILED` 或 `PENDING` 的步骤继续。

---

## 🔧 TUI 交互界面

`xdxtools tui` 提供了一个基于终端的交互式图形界面，适合不熟悉命令行参数的用户。

### 启动 TUI

```bash
xdxtools tui
```

### 适用场景

- 不记得具体参数时，通过菜单逐步填写
- 需要实时监控多个任务的状态
- 快速查看日志和报错信息

### TUI 功能

| 功能 | 说明 |
|------|------|
| 项目管理 | 创建/查看/切换项目 |
| 任务提交 | 图形化填写参数并提交任务 |
| 状态监控 | 实时查看各步骤进度 |
| 日志查看 | 直接在 TUI 中查看运行日志 |

> **操作提示**：使用方向键导航，回车键确认，`q` 或 `Ctrl+C` 退出。

---

## 🛠️ 配置文件手动编辑

有时候需要手动修改 `config.yaml` 来调整分析参数。

### config.yaml 全字段说明

```yaml
# ========================================
# xdxtools 配置文件
# ========================================

workflow:
  mode: RRBS                    # 分析模式：RRBS / WGBS / RNASEQ
  species:
    primary: hg38               # 主要物种参考基因组
    secondary: ""               # 第二物种（PDX 模式填 mm10）

input:
  fastq_dir: ./fastq            # FASTQ 文件目录
  pdata: samples.xlsx           # 样本信息表路径

output:
  dir: userspace/               # 输出根目录
  job_id: a3f8e1c2...           # Job ID（自动生成，不要手动修改）

reference:
  genome: /path/to/hg38/genome  # 参考基因组目录

engine:
  type: slurm                   # 执行引擎：slurm / local / auto
  partition: normal             # SLURM 分区（仅 slurm 模式）
  parallel_jobs: 10             # 并发任务数
  load_ratio: 0.8               # 负载比例

# 每个样本的配置（由 create 命令自动生成）
samples:
  - id: sample1
    r1: ./fastq/sample1_R1.fastq.gz
    r2: ./fastq/sample1_R2.fastq.gz
    adapter_r1: TGAATCACG       # 自动计算的适配器序列
    adapter_r2: ACGATGAT        # 自动计算的适配器序列
    condition: treated
    sample_group: Group_A
```

### 常见手动修改场景

**场景 1：修改适配器序列**

如果 `create` 生成的适配器不对，手动修改：

```yaml
samples:
  - id: sample1
    adapter_r1: "TGAATCACG"    # 修改为正确的序列
    adapter_r2: "ACGATGAT"
```

**场景 2：调整 Trim 参数**

```yaml
workflow:
  trim:
    quality: 20                 # 质量分数阈值（默认 20）
    min_length: 20              # 最小 read 长度（默认 20）
    clip_r1: 0                  # 从 R1 5' 端额外裁剪的碱基数
    clip_r2: 0                  # 从 R2 5' 端额外裁剪的碱基数
```

**场景 3：切换执行引擎**

```yaml
engine:
  type: local                   # 从 slurm 改为 local（本地运行）
  parallel_jobs: 4
```

---

## ✅ 配置验证

在运行之前，可以先验证配置文件是否正确：

```bash
xdxtools config validate --config userspace/<jobid>/config/config.yaml
```

**输出示例（验证通过）**：

```
✅ Config file syntax: OK
✅ FASTQ files found: 16 files (8 samples)
✅ Pdata samples matched: 8/8
✅ Reference genome accessible: /path/to/hg38
✅ Engine config: SLURM (partition=normal)
✅ All checks passed. Ready to run.
```

**输出示例（有问题）**：

```
✅ Config file syntax: OK
❌ FASTQ files: sample3_R2.fastq.gz not found
⚠️  Reference genome: /path/to/hg38 not accessible (check permissions)
```

---

**下一章：** [🛠️ 第六章：子工具参考手册](06-subtools.md)

**上一章：** [🔬 第四章：分析模式详解](04-analysis-modes.md)
