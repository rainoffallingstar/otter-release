# xdxtools 项目全景总览

> 更新日期：2026-02-27

xdxtools 是一个面向生信分析的工作流管理 CLI，围绕 Snakemake 编排多个专用二进制工具，支持 RRBS、WGBS、RNA-seq 和 PDX 样本的全流程分析。项目由 **1 个主仓库 + 8 个 git 子仓库**组成。

---

## 子仓库汇总表

| 子仓库 | 路径 | 语言 | 版本 | 二进制 | 状态 | 流程角色 |
|--------|------|------|------|--------|------|----------|
| enva | `enva/` | Rust | 0.1.0 | `enva` | ✅ 已编译 | conda 环境驱动层 |
| fastqc-rs | `fastqc-rs/` | Rust | 0.3.4 | `fqc` | ✅ 已编译 | FASTQ 质控（替代 FastQC）|
| xenofilter-go | `xenofilter-go/` | Go | devel | `xenofilter` | ✅ 已编译 | PDX 宿主读段过滤 |
| Paireads | `Paireads/` | Go | devel | `paireads` | ✅ 已编译 | PDX 配对读段验证 |
| htseq2matrix-go | `htseq2matrix-go/` | Go | devel | `htseq2matrix` | ✅ 已编译 | RNA-seq 表达矩阵生成 |
| methrix-cli | `methrix-cli/` | Rust | 0.1.0 | `methrix-cli` | ✅ 已编译 | 甲基化 HDF5 对象生成 |
| qctb | `qctb/` | Rust | 0.1.0 | `qctb` | ✅ 已编译 | QC 汇总报告（Excel）|
| gomats | `gomats/` | Go | devel | `gomats` | ✅ 已编译 | rMATS 可变剪接分析 |

---

## 1. 主项目 xdxtools

| 字段 | 值 |
|------|----|
| 语言 | Go 1.24.0 |
| 模块名 | `github.com/xdxtools/xdxtools-go` |
| 二进制 | `xdxtools`（16.5 MB，已编译）|
| 框架 | Cobra CLI + Snakemake 编排 |

### 1.1 CLI 命令

| 命令 | 说明 |
|------|------|
| `init` | 初始化项目根目录，复制 Snakemake 工作流和脚本资产，并创建 `userspace/` 工作区 |
| `create` | 扫描 FASTQ、解析 pdata（Excel/CSV）、生成 `config.yaml` |
| `run` | 执行 Snakemake 工作流，支持三种执行引擎 |
| `status` | 显示工作流运行状态与进度 |
| `config validate` | 验证配置文件合法性 |

### 1.2 流程模式

| 模式 | Snakefile | 步骤数 | 说明 |
|------|-----------|--------|------|
| RRBS / WGBS | `BeaverBS.snakemake` | 3 步 | 亚硫酸盐测序 → 甲基化矩阵 |
| RNASEQ | `BeaverRNA.snakemake` | 2 步 | RNA 定量 → 表达矩阵 |
| PDX（BS） | `BeaverPDX.snakemake` | 3 步 | BS-seq + Xenograft 过滤 |
| PDX（RNA） | `BeaverRNASEQPDX.snakemake` | 3 步 | RNA-seq + Xenograft 过滤 |

> PDX 模式在 `species2` 字段有值时自动启用。

### 1.3 执行引擎（`internal/engine/`）

| 引擎 | 类型常量 | 触发条件 | 说明 |
|------|----------|----------|------|
| SLURM | `EngineSlurm` | 手动指定或默认集群模式 | 单任务 SLURM 提交 |
| SLURM Array | `EngineSlurmArray` | ≥5 样本时自动启用 | Job Array 并行，每样本独立作业 |
| Local | `EngineLocal` | `--engine local` | 本地 worker pool |

### 1.4 核心内部模块

| 模块 | 路径 | 职责 |
|------|------|------|
| config | `internal/config/` | YAML 配置加载、验证、生成 |
| engine | `internal/engine/` | 执行后端（Slurm / SlurmArray / Local）|
| input | `internal/input/` | FASTQ 扫描、pdata 解析、Adapter 计算 |
| workflow | `internal/workflow/` | 目录结构管理、Snakemake 集成、enva 检测 |
| assets | `internal/assets/` | Go `embed` 嵌入资源（Snakemake 文件、R 脚本）|

---

## 2. 子仓库详细说明

### 2.1 enva — rattler 优先的环境管理器

- **语言 / 版本**: Rust 0.1.0
- **二进制**: `enva`
- **功能**: 以 rattler 为一等公民的环境管理器，负责创建和运行 4 个预设环境（core / r / snakemake / extra），并兼容发现 / 接管历史 conda 环境
- **在 xdxtools 中的作用**: `internal/workflow/snakemake.go` 启动前优先检测 `enva`；若可用则以 `enva run <env> --` 调用 Snakemake，并统一兼容历史 conda 环境
- **调用位置**: 主程序调用层注入，不在 Snakemake rules 的 `shell` 块中直接出现

---

### 2.2 fastqc-rs（`fqc`）— FASTQ 质控

- **语言 / 版本**: Rust 0.3.4
- **二进制**: `fqc`
- **功能**: Rust 版 FastQC 替代品，输出 HTML 报告 + MultiQC 兼容的 `fastqc_data.txt`，内置 Seqkit 统计；替代原 FastQC + Seqkit 组合
- **调用规则**:
  - `inst/rules/01fqcAtfirst.smk` — 原始 FASTQ 质控
  - `inst/rules/03-0-fqcAtclean.smk` — 剪接后 FASTQ 质控
- **调用示例**:
  ```bash
  fqc -q {input.R1} -s {params.R1_dir} --no-html
  fqc -q {input.R2} -s {params.R2_dir} --no-html
  ```

---

### 2.4 xenofilter-go — Xenograft 读段过滤

- **语言 / 版本**: Go（devel）
- **二进制**: `xenofilter`
- **功能**: 纯 Go 版 XenofilteR，对比 graft（人）/ host（鼠）比对 BAM 的编辑距离（NM tag + soft-clip），保留来自 graft 的读段；支持 Bisulfite 模式
- **调用规则**: `inst/rules/XenofilteR.smk`
- **调用示例**:
  ```bash
  xenofilter run \
    --graft <graft_bam1> [<graft_bam2>...] \
    --host  <host_bam1>  [<host_bam2>...] \
    --output <output_dir> \
    --mm-threshold 4 \
    --threads <N> \
    [--bisulfite]
  ```
- **特点**: 纯 Go 无 C 依赖，评分 = NM_tag + clips

---

### 2.5 Paireads — 配对读段过滤

- **语言 / 版本**: Go（devel）
- **二进制**: `paireads`
- **功能**: 纯 Go 工具，从 BAM 中过滤出正确配对的双端读段，剔除孤立读段，常用于 XenofilteR 输出的后处理
- **调用位置**: PDX 流程中 `XenofilteR.smk` 之后的 BAM 后处理步骤
- **调用示例**:
  ```bash
  paireads <R1.bam> <R2.bam> <output_prefix>
  ```
- **输出**: `filtered_R1.bam`、`filtered_R2.bam`、未配对读名列表

---

### 2.6 htseq2matrix-go — 表达矩阵生成

- **语言 / 版本**: Go（devel）
- **二进制**: `htseq2matrix`
- **功能**: 将多样本 HTSeq 计数文件合并为基因表达矩阵，内嵌 ENSEMBL → Gene Symbol 映射数据库，支持 log2 归一化
- **调用规则**: `inst/rules/rnaseq_matrix.smk`
- **调用示例**:
  ```bash
  htseq2matrix \
    --htseq_dir  <htseq_output_dir> \
    --output_dir <matrix_output_dir> \
    --postfix    _human.txt
  ```
- **输出**:
  - `matrix_count.txt` — 原始计数矩阵
  - `matrix_norm.txt` — log2 归一化矩阵

---

### 2.7 methrix-cli — 甲基化矩阵处理

- **语言 / 版本**: Rust 0.1.0
- **二进制**: `methrix-cli`
- **功能**: 将 Bismark coverage 文件批量转换为 HDF5 格式，与 R `methrix` 包 100% 兼容，支持多线程
- **调用规则**: `inst/rules/methrix_object.smk`
- **调用示例**:
  ```bash
  methrix-cli process \
    --input   <mcall_dir> \
    --output  <methrix_dir> \
    --genome  <fasta_or_ron> \
    --threads <N>
  ```
- **特殊依赖**: 编译时需要 HDF5 库（`$HDF5_DIR` 环境变量）
- **输出**: HDF5 文件，含 beta 矩阵（f32）+ 覆盖度矩阵（u16）

---

### 2.8 qctb — QC 汇总报告

- **语言 / 版本**: Rust 0.1.0
- **二进制**: `qctb`
- **功能**: 汇总多样本 QC 统计信息（比对率、测序深度、覆盖度等），生成 Excel 格式综合报告
- **调用规则**:
  - `inst/rules/bs_qc_summary.smk` — BS-seq 模式 QC 汇总
  - `inst/rules/rna_qc_summary.smk` — RNA-seq 模式 QC 汇总
- **调用示例**:
  ```bash
  # BS-seq 模式
  qctb --config config.yaml --output qc_summary.xlsx
  # RNA-seq 模式
  qctb --config config.yaml --output qc_summary.xlsx --rnaseq
  ```

### 2.9 gomats — rMATS 可变剪接分析

- **语言 / 版本**: Go
- **二进制**: `gomats`
- **功能**: Go 实现的 rMATS 可变剪接分析编排器，替代 R 脚本 RNA_Splicing.R，消除嵌套 enva 调用
- **调用规则**:
  - `inst/rules/rnaseq_splicing.smk` — RNA-seq 模式可变剪接分析
- **调用示例**:
  ```bash
  gomats run --root /data/bam --pdata samples.xlsx --seqlengthQC /data/qc --gtf hg38.gtf
  ```

---

## 3. Snakemake 流程与工具对应关系

### 3.1 BS-seq 流程（RRBS / WGBS）

```
Step 1（质控与剪接）
  01fqcAtfirst.smk       ←  fqc（原始 FASTQ 质控）
  02fastq2trim.smk       ←  Trim Galore
  03-0-fqcAtclean.smk    ←  fqc（剪接后质控）

Step 2（比对与矩阵）
  04bsmap2sort_bismark.smk          ←  Bismark
  build_methy_matrix_bismark.smk    ←  Bismark coverage 汇总

Step 3（甲基化对象与 QC 报告）
  methrix_object.smk     ←  methrix-cli（生成 HDF5）
  bs_qc_summary.smk      ←  qctb（生成 Excel QC 报告）
```

### 3.2 RNA-seq 流程

```
Step 1（质控与剪接）
  01fqcAtfirst.smk         ←  fqc
  02fastq2trim.smk         ←  Trim Galore
  rnaseq_step1_checker.smk ←  校验器

Step 2（比对、计数与报告）
  rnaseq_mapping.smk       ←  STAR
  rnaseq_htseq.smk         ←  HTSeq-count
  rnaseq_matrix.smk        ←  htseq2matrix（生成表达矩阵）
  rna_qc_summary.smk       ←  qctb（生成 Excel QC 报告）
```

### 3.3 PDX 附加步骤（在 Step 2 比对后插入）

```
XenofilteR.smk             ←  xenofilter run（保留 graft 读段）
Paireads 后处理             ←  paireads（剔除孤立读段）
```

> PDX BS-seq 使用 `BeaverPDX.snakemake`，PDX RNA-seq 使用 `BeaverRNASEQPDX.snakemake`。

---

## 4. 相关文档

| 文档 | 路径 | 说明 |
|------|------|------|
| 子仓库编译指南 | `docs/submodules-build-guide.md` | 各子仓库详细编译步骤与环境配置 |
| 架构设计 | `docs/architecture.md` | 数据流与模块交互设计（含 Mermaid 图）|
| 需求文档 | `docs/requirements.md` | 功能需求与设计决策 |
| 安装说明 | `docs/installation.md` | 安装与环境配置说明 |
