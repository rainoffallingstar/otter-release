# Otter 配置契约

> 状态：canonical v1 的 typed model、严格解析、samples manifest、resolver、immutable snapshot 与 legacy adapter 已实现。生产 `run` 路由和 backend auto detection 分别属于 Gate 2、Gate 3。

## 配置分层

```text
project.yaml + samples.tsv + references.lock.yaml
                  + site/profile/auto detection
                  + CLI overrides
                              ↓
                         Otter resolver
                              ↓
                 runs/<run_id>/run.yaml
                              ↓
                 Craftmake 或 Snakemake adapter
```

- `project.yaml`：用户可编辑的项目意图。
- `samples.tsv`：独立、版本化样本清单。
- `references.lock.yaml`：项目默认 reference 的 release 与 manifest digest。
- `run.yaml`：完全展开、不可变、唯一可执行配置。
- Craftmake 只解析 `run.yaml`；不读取项目文件或 legacy config。
- QCTB 直接解析同一不可变 `run.yaml`：`qctb --config <run.yaml>`，或 `qctb --config-dir <run-root>`（等价于 `<run-root>/run.yaml`）。QCTB 不读取 Snakemake compatibility YAML。

## Canonical project v1

```yaml
schema_version: otter.project/v1
project:
  id: cohort-a
workflow:
  scenario: rrbs
  toolchain: modern
execution:
  executor: craftmake
  backend: auto
  site: auto
samples:
  manifest: samples.tsv
references:
  primary: hg38@GRCh38.p14
resources:
  defaults:
    cores: 8
    memory: 32GiB
observability:
  metrics: true
  retain_logs: true
```

### 默认路由

- `execution.executor` 默认 `craftmake`。
- `snakemake` 只能由 config 或 `--executor snakemake` 显式选择，不是失败 fallback。
- `execution.backend` 默认 `auto`。
- 完整 SLURM 能力验证通过时使用 `slurm`；完全不存在 SLURM 时使用 `local`；部分可用或资源/共享路径不完整时 fail closed。
- 真实生信流程只允许在 sbatch 生产门禁运行；Local 仅承担 contract tests。

### 覆盖优先级

```text
CLI > site profile/auto detection > project.yaml > workflow defaults
```

每个最终值必须在 `run.yaml` 中记录 `value` 和 `source`，禁止运行时再次隐式读取环境变量改变结果。

## 样本清单

`sample_id`、`r1`、`r2` 为必填列；其余列按流程验证。

```tsv
sample_id	r1	r2	group	batch	adapter_r1	adapter_r2
S01	data/S01_R1.fastq.gz	data/S01_R2.fastq.gz	case	B1	AUTO	AUTO
```

约束：

- `sample_id` 在项目内唯一，且不得包含路径分隔符。
- FASTQ 路径相对 `project.yaml` 解析；`run.yaml` 固化规范化绝对路径与输入 digest/metadata。
- PDX 的物种角色属于 reference/species，不编码进 sample ID。
- 修改样本清单后必须创建新 run；resume 使用原快照。

## Workflow 场景

使用五个独立 scenario：

- `rrbs`
- `wgbs`
- `rnaseq`
- `bs-pdx`
- `rna-pdx`

PDX 由两个显式 species role 表达，不使用隐式数组下标：

```yaml
references:
  species:
    - role: graft
      selection: hg38@GRCh38.p14
    - role: host
      selection: mm39@GRCm39
```

## Toolchain

- `modern`：使用 `fastqcx/xenofilx/pairbam/seq2mat/matsrun/methx/qctb` 等当前算子。
- `legacy-equivalent`：只包含有明确现代替代关系、可进行语义比较的旧实现。
- `legacy_extensions`：CluBCpG、mHap、CCGG、insert-length 等额外科学步骤；不进入主 parity 门禁。

## Run snapshot

`run.yaml` 至少固化：

- run ID、UTC 创建时间、project/config/sample digests；
- scenario、executor、backend、toolchain、site 的最终值和来源；
- 完整 samples/species；
- reference ID、release、manifest、绝对路径和 checksum；
- workflow/rules/environment/tool digest；
- 每个 phase/job 的资源；
- run root 下全部路径；
- observability、benchmark 和 parity policy。

生成后不得原地修改。任何会改变 DAG、输入、reference、执行器或科学结果的 override 都必须创建新 run。

## Run 级 reference override

```bash
otter run --reference-primary hg19@GRCh37.p13
```

该 override 只进入新 run：

```yaml
references:
  project_reference: hg38@GRCh38.p14
  effective_reference: hg19@GRCh37.p13
  override: true
  override_source: cli
```

Gate 1 已在 `otter config resolve --reference-primary ...` 中实现 snapshot 级 override；Gate 2 接入生产 `otter run` 后沿用同一 resolver。成功 run 不自动修改项目默认。显式提升使用：

```bash
otter reference promote <run_id>
```

promote 先验证资产与兼容性，再生成新的 project/reference lock 变更预览；用户确认后才发布。

## Legacy migration

历史字段只在 Otter legacy adapter 中读取，保留两个 minor release 的 warning 期。迁移表至少覆盖：

| Legacy | Canonical v1 |
|---|---|
| 顶层 `SIDs` / `metadata.SIDs` | `samples.tsv` |
| 顶层 `mode` / `workflow.mode` | `workflow.scenario` |
| `species1/species2`、`graft/host/name` | `references.species[].role/selection` |
| `C1/C2/T1/T2` 或小写变体 | typed workflow parameters |
| 多种 FASTA/GTF/index 数组 | reference registry selection |
| `engine.type` | `execution.backend` |
| Snakemake conda/fallback 字段 | explicit Snakemake adapter config |
| 旧 directory keys | run-root derived paths |

已实现命令：

```bash
otter config validate --schema auto --config project.yaml
otter config migrate --from legacy --to v1 --input old.yaml --output project.yaml \
  --reference-primary hg38@GRCh38.p14
otter config resolve --project project.yaml --reference-root /shared/otter/references \
  --backend slurm
```

Gate 3 完成前，`config resolve` 遇到 `backend=auto` 会 fail closed，必须显式传入 `--backend local|slurm`；这不会启动科学流程。

迁移必须输出采用字段、弃用字段、冲突和无法推断项；冲突不得静默选择。
