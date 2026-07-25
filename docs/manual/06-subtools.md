# 第六章：组件参考手册

## 层级

```text
otter → craftmake → enva → 算子 → bamdriver
```

## craftmake

`craftmake` 是 Snakemake 的 Go 替代执行层，包含 workflow spec、DAG、local/SLURM 后端与 otter 适配边界。它仍在接入迁移中，当前不能视为所有 `otter run` 的唯一后端。

```bash
craftmake --help
```

在切换任一流程前，应与 Snakemake 对同一 fixture 比较任务图、资源、失败传播、恢复和关键产物。

## enva

`enva` 是 rattler-first 环境层，管理：

```text
otter-core
otter-snakemake
otter-extra
```

```bash
enva list --detailed
enva create --all
enva run otter-core -- <command>
enva validate --all
```

## 算子

### fastqcx

FASTQ 质控算子，保留 FastQC/MultiQC 可消费格式。FastQC 是外部标准名，不应改写。

```bash
fastqcx --help
```

历史仓库名：`fastqc-rs`；历史二进制可能为 `fqc`。

### xenofilx

PDX graft/host 读段分类算子。

```bash
xenofilx --help
```

历史仓库名：`xenofilter-go`，历史二进制名：`xenofilter`。

### pairbam

过滤或恢复完整 paired-end BAM 读段，通常消费 `xenofilx` 结果。

```bash
pairbam --help
```

历史仓库名：`Paireads`，历史二进制名：`paireads`。

### seq2mat

将多样本 HTSeq count 合并为可追溯表达矩阵。HTSeq 是外部工具/格式名。

```bash
seq2mat --help
```

历史仓库名：`htseq2matrix-go`，历史二进制名：`htseq2matrix`。

### matsrun

编排 rMATS 可变剪接任务；rMATS 是外部科学工具名。

```bash
matsrun --help
```

历史仓库名/二进制名：`gomats`。

### qctb

聚合 BS-seq、RNA-seq、PDX 等模式的 QC 输入并生成版本化结果。

```bash
qctb --help
```

### methx

处理 Bismark coverage 与甲基化/HDF5 产物。Methrix 是科学领域/兼容边界名称，应准确声明原生 schema 与兼容程度。

```bash
methx --help
```

历史仓库名/二进制名：`methrix-cli`。

## bamdriver

`bamdriver` 是共享 BAM 低层操作层，供 `xenofilx`、`pairbam`、`matsrun` 等组件复用。除非其 CLI 明确稳定，普通用户应优先通过上层算子调用。

```bash
bamdriver --help
```

历史仓库名：`bamdriver-go`。

## 命名兼容规则

旧名只用于定位旧 release、历史审查和兼容二进制。新安装、构建清单和当前目录应使用本章当前名。

下一章：[常见问题与排查](07-faq.md)
