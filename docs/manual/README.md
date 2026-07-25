# otter 用户手册

`otter` 面向 RRBS、WGBS、RNA-seq 和 PDX 分析，提供项目初始化、输入验证、配置生成、工作流执行和任务状态管理。

## 当前运行模型

```text
otter → craftmake → enva → 算子 → bamdriver
```

`craftmake` 是 Snakemake 的 Go 替代执行层，但仍在接入。当前为双轨：生产流程继续使用 Snakemake，`craftmake` 用于迁移和等价验证。手册中的 `otter` 是当前产品名；若现有源码构建仍显示 `xdxtools`，这是待完成的代码兼容名迁移。

## 章节

| 章节 | 内容 |
|---|---|
| [安装指南](01-installation.md) | 主仓、环境和组件安装 |
| [数据准备](02-data-preparation.md) | FASTQ 与 pdata |
| [快速上手](03-quickstart.md) | init/create/run/status |
| [分析模式](04-analysis-modes.md) | RRBS/WGBS/RNA-seq/PDX |
| [高级用法](05-advanced-usage.md) | SLURM、恢复和参数覆盖 |
| [组件参考](06-subtools.md) | craftmake、enva、算子、bamdriver |
| [FAQ](07-faq.md) | 安装和运行排查 |

## 组件

| 层级 | 组件 |
|---|---|
| 协调 | `otter` |
| 执行 | `craftmake`（迁移中）/ Snakemake（当前兼容路径） |
| 环境 | `enva`；`otter-core`、`otter-snakemake`、`otter-extra` |
| 算子 | `fastqcx`, `xenofilx`, `pairbam`, `seq2mat`, `matsrun`, `qctb`, `methx` |
| BAM 基础 | `bamdriver` |

FastQC、MultiQC、Methrix、Bismark、HTSeq 和 rMATS 是外部标准、工具或科学领域名称，继续按实际契约使用。

## 快速参考

```bash
otter init my_project

otter create --fastq /data/fastq --mode RRBS --pdata samples.xlsx \
  --output my_project/userspace --jobid demo_rrbs

otter run --config my_project/userspace/demo_rrbs/config/config.yaml --dry-run
otter task list
otter status my_project/userspace/demo_rrbs
```

- 命令帮助：`otter --help`
- GitHub Issues: <https://github.com/rainoffallingstar/otter/issues>
