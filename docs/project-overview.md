# otter 项目全景总览

> 更新日期：2026-07-25

`otter` 是面向 RRBS、WGBS、RNA-seq 和 PDX 的 Go 工作流 CLI。主仓为 `rainoffallingstar/otter`，产品层级统一为：

```text
otter → craftmake → enva → 算子 → bamdriver
```

## 当前组件

| 层级 | 组件 | 路径 | 角色 | 当前状态 |
|---|---|---|---|---|
| 协调 | `otter` | 父仓根目录 | 项目、配置、任务与流程入口 | 产品命名已统一；源码兼容名待迁移 |
| 执行 | `craftmake` | `craftmake/` | Snakemake 的 Go 替代执行层 | 迁移中，与 Snakemake 双轨 |
| 环境 | `enva` | `enva/` | rattler-first 环境生命周期与命令隔离 | 当前环境入口 |
| 算子 | `fastqcx` | `fastqcx/` | FASTQ 质控，保留 FastQC/MultiQC 契约 | 子模块 |
| 算子 | `xenofilx` | `xenofilx/` | PDX 物种过滤 | 子模块 |
| 算子 | `pairbam` | `pairbam/` | 配对 BAM 过滤与恢复 | 子模块 |
| 算子 | `seq2mat` | `seq2mat/` | HTSeq count-to-matrix | 子模块 |
| 算子 | `matsrun` | `matsrun/` | rMATS 流程编排 | 子模块 |
| 算子 | `qctb` | `qctb/` | QC 聚合报告 | 子模块 |
| 算子 | `methx` | `methx/` | 甲基化/HDF5 处理，保留 Methrix 领域语义 | 子模块 |
| 基础 | `bamdriver` | `bamdriver/` | 共享 BAM 操作层 | 子模块 |

## 运行路径

当前不是纯 `craftmake` 单轨：

1. `otter init/create` 负责工作区与 `OtterConfig`。
2. `otter run` 的既有生产路径仍调用 Snakemake。
3. `craftmake` 正在实现/验证 Go 原生 spec、DAG、local 与 SLURM 后端。
4. 两条路径都通过 `enva` 解析 `otter-core`、`otter-snakemake`、`otter-extra`。
5. `enva` 调用专用算子；需要 BAM 低层操作时由 `bamdriver` 提供能力。

完成四种模式的等价性、恢复与科学产物门禁前，不得声称 Snakemake 已被替换。

## 工作流模式

| 模式 | 当前兼容流程 | 主要算子 |
|---|---|---|
| RRBS/WGBS | Snakemake 3 步路径 | `fastqcx`, `methx`, `qctb` |
| RNA-seq | Snakemake 2 步路径 | `fastqcx`, `seq2mat`, `matsrun`, `qctb` |
| PDX | 双物种 Snakemake 路径 | `xenofilx`, `pairbam`, `bamdriver` |

Bismark、STAR、HTSeq、FastQC/MultiQC、Methrix、rMATS 等是外部工具、标准或科学领域名称，不属于产品重命名范围。

## 父仓主要目录

| 路径 | 职责 |
|---|---|
| `cmd/` | CLI 命令层 |
| `internal/config/` | `OtterConfig` 的目标契约；源码旧类型仍待迁移 |
| `internal/input/` | FASTQ/pdata 解析与验证 |
| `internal/engine/` | local/SLURM 执行边界 |
| `internal/workflow/` | 当前 Snakemake 集成与工作流状态 |
| `inst/` | 嵌入 workflow、rules、env 和辅助资产 |
| `docs/` | 当前文档、归档、审查证据 |

## 历史名称映射

| 当前名 | 历史名 |
|---|---|
| `otter` | `xdxtools`, `xdxtools-go` |
| `fastqcx` | `fastqc-rs` |
| `xenofilx` | `xenofilter-go` |
| `pairbam` | `Paireads` |
| `seq2mat` | `htseq2matrix-go` |
| `matsrun` | `gomats` |
| `methx` | `methrix-cli` |
| `bamdriver` | `bamdriver-go` |

`docs/archive/**`、`docs/review/submodules/**` 和已有 remediation 报告保留历史名称，以维持当时证据的可追溯性。

## 文档入口

- [架构设计](architecture.md)
- [安装指南](installation.md)
- [构建指南](build.md)
- [需求文档](requirements.md)
- [子模块指南](submodules-build-guide.md)
- [用户手册](manual/README.md)
