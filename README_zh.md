# otter

`otter` 是面向 RRBS、WGBS、RNA-seq 和 PDX 分析的生物信息学工作流 CLI。GitHub 主仓统一为 [rainoffallingstar/otter](https://github.com/rainoffallingstar/otter)。

## 产品体系

```text
otter → craftmake → enva → 算子 → bamdriver
```

- `otter`：面向用户的项目、配置、任务和工作流协调入口。
- `craftmake`：用于替代 Snakemake 的 Go 工作流执行层。
- `enva`：管理 `otter-core`、`otter-snakemake`、`otter-extra` 环境并隔离命令执行。
- 算子：`fastqcx`、`xenofilx`、`pairbam`、`seq2mat`、`matsrun`、`qctb`、`methx`。
- `bamdriver`：共享的底层 BAM 操作层。

`craftmake` 仍在接入迁移中。当前运行时是双轨状态：既有生产流程继续使用 Snakemake，同时推进 `craftmake` 的兼容验证和接入。本文档不声称 Snakemake 已经被完全替换。

## 主要能力

- RRBS、WGBS、RNA-seq、PDX 工作流
- FASTQ 自动配对、pdata 校验和逐样本接头生成
- SLURM、SLURM Job Array 与本地执行
- 后台任务状态和日志恢复
- Excel/CSV pdata 与中文列名映射
- 通过 `enva` 进行 rattler 优先的环境管理
- 专用原生算子，并保留适用的科学格式兼容性

## 系统要求

- Go 1.24+
- 当前生产路径需要 Snakemake
- 迁移和兼容验证使用 `craftmake`
- [enva](https://github.com/rainoffallingstar/enva)
- `conda`、`mamba`、`micromamba` 仅用于兼容或接管已有环境

## 安装

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/rainoffallingstar/otter/main/scripts/install.sh)
```

从源码构建：

```bash
git clone --recurse-submodules https://github.com/rainoffallingstar/otter.git
cd otter
conda activate go-env
go build -o otter .
```

仓库仍处于命名迁移期，部分源码符号、状态文件或旧 release 资产可能保留 `xdxtools`；这些是兼容名，不是当前产品名。

## 快速上手

```bash
otter init my_project

otter create --fastq /data/fastq --mode RRBS --pdata samples.csv \
  --output my_project/userspace --jobid demo_rrbs

otter run --config my_project/userspace/demo_rrbs/config/config.yaml

otter task list
otter task status <task-id>
otter task logs <task-id> --follow
```

`otter run` 当前仍可调度既有 Snakemake 路径；`craftmake` 是正在迁移接入的 Go 替代执行层，尚未成为唯一生产后端。

## 环境名称

| 环境 | 用途 |
|---|---|
| `otter-core` | 核心生信运行时与算子 |
| `otter-snakemake` | 当前 Snakemake 兼容运行时 |
| `otter-extra` | 额外分析和可视化工具 |

## 算子名称映射

| 当前名称 | 科学角色 | 历史仓库名 |
|---|---|---|
| `fastqcx` | FASTQ 质控；保留 FastQC/MultiQC 兼容输出 | `fastqc-rs` |
| `xenofilx` | PDX 异种移植读段分类 | `xenofilter-go` |
| `pairbam` | 配对读段 BAM 过滤 | `Paireads` |
| `seq2mat` | HTSeq 计数转表达矩阵 | `htseq2matrix-go` |
| `matsrun` | rMATS 编排 | `gomats` |
| `qctb` | QC 汇总 | 名称不变 |
| `methx` | 甲基化/HDF5 处理；保留 Methrix 领域术语 | `methrix-cli` |
| `bamdriver` | 共享 BAM 操作 | `bamdriver-go` |

历史审查和归档文档保留证据产生时的旧名称，不做批量改写。

## 文档入口

- [架构](docs/architecture.md)
- [安装](docs/installation.md)
- [构建](docs/build.md)
- [用户手册](docs/manual/README.md)
- [子模块指南](docs/submodules-build-guide.md)
- [当前上下文](docs/active_context.md)

## 许可证

MIT
