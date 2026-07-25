# 第五章：高级用法

> **本章你将学到：**
> - `create` / `run` 高级参数的场景化用法
> - 自定义参考文件与复杂 FASTQ 命名处理
> - SLURM 全局与分步骤资源覆盖策略
> - 运行前数据处理、断点续传与调试

---

## ⚠️ 参数口径说明

从当前版本开始，`create` 命令使用 `--species1` 指定主物种，`--species2` 指定第二物种（PDX）。

```bash
# 单物种
otter create --fastq ./fastq --mode RRBS --pdata samples.xlsx --species1 hg38

# 双物种（PDX）
otter create --fastq ./fastq --mode RRBS --pdata samples.xlsx --species1 hg38 --species2 mm10
```

---

## 场景 1：自定义参考文件（create，高级）

适用场景：
- 你使用的参考基因组不在 `inst/` 默认路径下
- RNA-seq 需要替换 GTF/STAR 索引
- PDX 需要同时指定两套参考资源

### RRBS/WGBS：覆盖默认 FASTA 与索引

```bash
otter create \
  --fastq ./fastq \
  --pdata samples.xlsx \
  --mode RRBS \
  --species1 hg38 \
  --genome1-fasta /ref/hg38/hg38.fa \
  --genome1-index /ref/hg38/bismark_index \
  --output userspace \
  --jobid rrbs_hg38_custom
```

### RNA-seq：覆盖 GTF 与 STAR 索引（含 PDX）

```bash
otter create \
  --fastq ./fastq \
  --pdata samples.xlsx \
  --mode RNASEQ \
  --species1 hg38 \
  --species2 mm10 \
  --gtf1 /ref/hg38/hg38.gtf \
  --star-index1 /ref/hg38/star \
  --gtf2 /ref/mm10/mm10.gtf \
  --star-index2 /ref/mm10/star \
  --genome1-fasta /ref/hg38/hg38.fa \
  --genome1-index /ref/hg38/bismark_index \
  --genome2-fasta /ref/mm10/mm10.fa \
  --genome2-index /ref/mm10/bismark_index
```

### create 高级参数速查

| 参数 | 作用 | 默认值 |
|------|------|--------|
| `--jobid` | 指定自定义任务 ID | 自动生成 |
| `--output` | 指定项目输出根目录 | `userspace` |
| `--suffix1` | 指定 R1 FASTQ 后缀 | `_R1.fastq.gz` |
| `--suffix2` | 指定 R2 FASTQ 后缀 | 自动推断 |
| `--conda-env` | 指定 Snakemake conda 环境名 | 空 |
| `--genome1-fasta` / `--genome1-index` | 主物种参考 FASTA / 索引目录 | 使用内置默认路径 |
| `--genome2-fasta` / `--genome2-index` | 第二物种 FASTA / 索引目录（PDX） | 使用内置默认路径 |
| `--gtf1` / `--star-index1` | 主物种 RNA-seq 注释与 STAR 索引 | 使用内置默认路径 |
| `--gtf2` / `--star-index2` | 第二物种 RNA-seq 注释与 STAR 索引（PDX） | 使用内置默认路径 |

---

## 场景 2：复杂 FASTQ 命名与目录扫描

适用场景：
- 文件名不是 `_R1.fastq.gz` / `_R2.fastq.gz`
- 需要稳定重跑（固定 jobid）

```bash
otter create \
  --fastq ./raw_fastq \
  --mode WGBS \
  --pdata samples.xlsx \
  --species1 hg38 \
  --suffix1 _1.fq.gz \
  --suffix2 _2.fq.gz \
  --jobid wgbs_batch_20260306
```

常见误用：
- `--suffix1` 与真实文件后缀不匹配，会导致扫描不到样本
- 只改 `--suffix1` 不改 `--suffix2`，可能出现配对失败

---

## 场景 3：SLURM 资源精细覆盖（run，高级）

otter 支持三层资源覆盖：
1. 全局：`--slurm-partition` / `--slurm-cores` / `--slurm-memory`
2. 分步骤：`--step1-*`、`--step2-*`、`--step3-*`
3. checker 步骤：`--step2-checker-*`、`--step3-checker-*`

### 全局资源 + 动态负载控制

```bash
otter run \
  --config my_project/userspace/<jobid>/config/config.yaml \
  --engine slurm \
  --slurm-partition normal \
  --slurm-cores 16 \
  --slurm-memory 64G \
  --parallel-jobs 12 \
  --load-ratio 0.8
```

### 分步骤覆盖（step2 大内存）

```bash
otter run \
  --config my_project/userspace/<jobid>/config/config.yaml \
  --engine slurm \
  --step1-cores 8 --step1-memory 32G --step1-partition normal \
  --step2-cores 24 --step2-memory 128G --step2-partition fat \
  --step3-cores 8 --step3-memory 16G --step3-partition normal \
  --step2-checker-cores 4 --step2-checker-memory 8G \
  --step3-checker-cores 2 --step3-checker-memory 4G
```

### run 高级参数速查

| 参数 | 作用 | 默认值 |
|------|------|--------|
| `--parallel-jobs` | 最大并行任务数（local/Snakemake） | `2` |
| `--load-ratio` | 动态负载比例（0.1-1.0，`0` 表示关闭） | `1.0` |
| `--slurm-unified-partition` | 统一覆盖所有步骤分区（兼容参数） | 空 |
| `--slurm-cores` / `--slurm-memory` | 覆盖全局 CPU/内存 | `0` / 空 |
| `--step1/2/3-cores` | 覆盖各步骤 CPU | `0` |
| `--step1/2/3-memory` | 覆盖各步骤内存 | 空 |
| `--step1/2/3-partition` | 覆盖各步骤分区 | 空 |
| `--step2/3-checker-cores` | 覆盖 checker CPU | `0` |
| `--step2/3-checker-memory` | 覆盖 checker 内存 | 空 |
| `--conda-env` | 指定 Snakemake conda 环境名 | 空 |
| `--verbose` / `-v` | 输出详细日志 | 关闭 |

---

## 场景 4：运行前处理、恢复与调试

### 运行前 FASTQ 预处理

```bash
# 仅压缩未压缩 FASTQ
otter run --config my_project/userspace/<jobid>/config/config.yaml --compress-fastq

# 将 FASTQ 复制到项目目录（保留原始数据）
otter run --config my_project/userspace/<jobid>/config/config.yaml --copy-fastq

# 将 FASTQ 移动到项目目录（原目录会变更）
otter run --config my_project/userspace/<jobid>/config/config.yaml --move-fastq
```

说明：
- `--copy-fastq` 和 `--move-fastq` 都会改动数据布局，建议先在测试项目验证。
- 若你只需要标准流程，不要同时启用这类参数。

### Dry-run 与断点续传

```bash
# 提交前检查（不执行）
otter run --config my_project/userspace/<jobid>/config/config.yaml --dry-run

# 从中断点恢复
otter run --config my_project/userspace/<jobid>/config/config.yaml --resume
```

状态文件位置在当前源码中仍可能是兼容名：

```
my_project/userspace/<jobid>/.xdxtools_state.json
```

代码命名迁移完成后才会统一为 `.otter_state.json`；不要在现有项目中手工改名。

---

## 🛠️ 配置文件手动编辑

有时候需要手动调整 `config.yaml`，建议先备份再修改。

### 常见手动修改场景

**场景 1：修改 Trim 参数**

```yaml
workflow:
  trim:
    quality: 20
    min_length: 20
    clip_r1: 0
    clip_r2: 0
```

**场景 2：切换执行引擎**

```yaml
engine:
  type: local
  parallel_jobs: 4
```

修改后请优先执行 `--dry-run` 或 `config validate` 检查配置一致性。

---

## ✅ 配置验证

在运行之前，可以先验证配置文件是否正确：

```bash
otter config validate --config my_project/userspace/<jobid>/config/config.yaml
```

---

**下一章：** [🛠️ 第六章：子工具参考手册](06-subtools.md)

**上一章：** [🔬 第四章：分析模式详解](04-analysis-modes.md)
