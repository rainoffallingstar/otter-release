# Otter 项目全景总览

> 更新日期：2026-07-26

Otter 是面向 RRBS、WGBS、RNA-seq、BS-PDX 和 RNA-PDX 的 Go 工作流 CLI。产品层级统一为：

```text
otter → craftmake → enva → 算子 → bamdriver
```

## 当前组件

| 层级 | 组件 | 路径 | 角色 | 当前状态 |
|---|---|---|---|---|
| 协调 | `otter` | 父仓根目录 | 项目、配置、任务与流程入口 | 当前生产入口 |
| 执行 | `craftmake` | `craftmake/` | DAG、Local/SLURM 和状态 | 迁移中，尚未完整接入 |
| 兼容 | Snakemake | `inst/` | 当前生产兼容 workflow | 显式兼容目标，退场需门禁 |
| 环境 | `enva` | `enva/` | Rattler-first 环境生命周期 | 当前环境入口 |
| 算子 | `fastqcx`、`xenofilx`、`pairbam`、`seq2mat`、`matsrun`、`qctb`、`methx` | 对应子模块 | QC、PDX、BAM、matrix、splicing、methylation | 独立演进 |
| 基础 | `bamdriver` | `bamdriver/` | 共享 BAM 操作层 | 子模块 |

## 当前路径和目标路径

当前：

1. `otter init/create` 生成已有 `OtterConfig` 与工作区。
2. `otter run` 生产路径调用 Snakemake。
3. Craftmake 已有 spec、DAG、Local/SLURM、状态和 CLI 能力，但 Otter 接入尚未完成。

目标：

1. `project.yaml`、`samples.tsv`、reference/workflow locks 表达项目意图。
2. Otter 解析 site、reference 和 CLI override，生成不可变 `runs/<run_id>/run.yaml`。
3. 默认 Craftmake 只解析 `run.yaml`；Snakemake 仅显式兼容，且不作为失败 fallback。
4. Local 只做 contract tests；真实流程、parity、恢复和 benchmark 全部在 sbatch 集群验证。

Run ID 为 `run-YYYYMMDDTHHMMSSZ-abcdef`，使用 UTC 时间戳和 6 位安全随机小写英文后缀。

## 五场景与比较维度

| 场景 | 主要领域 |
|---|---|
| RRBS | restriction-aware BS alignment、CpG/methylation |
| WGBS | whole-genome BS alignment、CpG/methylation |
| RNA-seq | STAR、counts/matrix、splicing |
| BS-PDX | graft/host separation + BS downstream |
| RNA-PDX | graft/host separation + RNA downstream |

比较维度彼此正交：`executor=craftmake|snakemake`、`backend=local|slurm|auto`、`toolchain=modern|legacy-equivalent`。主 sbatch benchmark 是 5 × 2 × 2 的 20-cell 矩阵。

## 新项目模型

```text
project/
├── project.yaml
├── samples.tsv
├── references.lock.yaml
├── project.lock.yaml
├── workflows/ rules/ environments/ schemas/
└── runs/<run_id>/
    ├── run.yaml
    └── input/ work/ results/ logs/ state/ metrics/
```

轻量 workflow assets 复制并锁定在项目；大型 genome 保存在共享 reference registry。项目可锁定默认 genome，单次 run 可覆盖；只有显式 `reference promote` 才能修改默认。

## 父仓主要目录

| 路径 | 职责 |
|---|---|
| `cmd/` | Otter CLI |
| `internal/config/` | 当前配置；后续 canonical typed resolver |
| `internal/input/` | FASTQ/pdata；后续 samples manifest |
| `internal/engine/` | 当前 local/SLURM 边界；后续收敛到 executor adapter |
| `internal/workflow/` | 当前 Snakemake 集成 |
| `inst/` | 当前嵌入 workflow、rules、env 和辅助资产 |
| `docs/` | 当前契约、迁移计划、归档与审查证据 |

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

历史归档与日期化审查证据保留当时名称。

## 文档入口

- [架构设计](architecture.md)
- [配置契约](configuration.md)
- [项目目录](project-layout.md)
- [执行协议](execution-contract.md)
- [Site 与 Backend](site-profiles.md)
- [Reference Registry](reference-registry.md)
- [Workflow Catalog](workflow-catalog.md)
- [Benchmark 计划](benchmark-plan.md)
- [Craftmake 迁移路线](migration/craftmake-adoption.md)
- [需求文档](requirements.md)
- [安装指南](installation.md)
- [构建指南](build.md)
- [用户手册](manual/README.md)
