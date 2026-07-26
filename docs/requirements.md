# Otter 需求文档

> 更新日期：2026-07-26。未标记已实现的条目均属于目标需求。

## 产品目标

Otter 统一 RRBS、WGBS、RNA-seq、BS-PDX 和 RNA-PDX 的项目、配置和任务入口。目标层级为：

```text
otter → craftmake → enva → 算子 → bamdriver
```

当前生产兼容路径仍使用 Snakemake；目标默认 executor 是 Craftmake。Snakemake 保留为显式兼容与 benchmark 路径，直到独立退场门禁通过。

## 功能需求

### 项目与配置

- `otter init` 创建 versioned `project.yaml`、`samples.tsv`、reference/workflow locks 和 pinned workflow assets。
- canonical project config 使用 `otter.project/v1` JSON Schema；历史字段只由 Otter legacy adapter 读取并保留两个 minor release warning。
- 禁止 `interface{}`、隐式数组下标 reference pairing 和静默字段冲突进入新模型。
- Otter 生成唯一、不可变 `runs/<run_id>/run.yaml`；Craftmake 只解析该文件。
- Run ID 格式为 `run-YYYYMMDDTHHMMSSZ-abcdef`：UTC、秒级、6 位安全随机小写英文、原子创建。
- resume 复用 snapshot/run ID；改变 samples、reference、executor、toolchain 或 DAG 的请求必须创建新 run。

### Executor 与 Backend

- `executor` 和 `backend` 是正交维度。
- executor 默认 `craftmake`；仅 config 或 `--executor snakemake` 可选择 Snakemake。
- Craftmake 失败后不得自动回退 Snakemake。
- backend 默认 `auto`，支持 `local | slurm`。
- SLURM 工具链和资源完整时选择 slurm；完全无 SLURM 时选择 Local；部分 SLURM、资源或共享路径不完整时 fail closed。
- Otter 与 Craftmake 使用版本化 CLI JSON/JSONL 协议，并关联 task/run/submission/SLURM IDs。
- cancel、signal、retry、failure propagation 和 resume 必须有可审计状态。

### Workflow 与工具链

- 提供五个独立场景：`rrbs`、`wgbs`、`rnaseq`、`bs-pdx`、`rna-pdx`。
- 每场景都有 Craftmake workflow 和 Snakemake compatibility adapter，发布相同 artifact contract。
- toolchain 支持 `modern | legacy-equivalent`。
- modern 子仓工具与其 legacy-equivalent 路径需有显式映射；CluBCpG、mHap、CCGG、insert-length 等 legacy-only extensions 单列。
- executor 比较固定 toolchain；工具比较固定 executor。
- 产物比较分为 exact、structural、scientific 和 informational tiers。

### Reference Registry

- 大型 genome 由共享 `$OTTER_REFERENCE_ROOT` registry 管理，不复制进项目。
- reference entry 必须记录 ID、release、assembly、FASTA、annotation、index、checksums、构建 provenance 和兼容场景。
- project lock 固定 reference ID/release/manifest digest，不固定 site mount root。
- sbatch 前验证文件、checksum、FASTA/index 对应关系、scenario compatibility 和 compute-node visibility。
- 单次 run 可覆盖项目默认 genome，但不修改项目 lock。
- `otter reference promote <run_id>` 必须经过验证、diff preview 和显式确认。

### 运行目录与 provenance

- 每个 run 拥有独立 `input/work/results/logs/state/metrics`。
- 双轨和 benchmark cell 不共享可变 work/state/cache。
- 产物通过 staging、验证和原子发布；`artifacts.json` 记录 schema、checksum 和 comparator。
- config、samples、workflow、environment、tool、reference 和 source commit digest 可追溯。

### 环境与算子

- Enva 提供 `otter-core`、`otter-snakemake`、`otter-extra` 的创建、发现、运行、安装、接管和删除。
- `fastqcx`、`xenofilx`、`pairbam`、`seq2mat`、`matsrun`、`qctb`、`methx` 保持稳定 CLI/文件契约。
- BAM 操作优先复用 `bamdriver`，避免上层算子复制不同语义。
- FastQC/MultiQC、Bismark、HTSeq、Methrix/HDF5、rMATS 等外部标准名与格式保持不变。

## 验证需求

### Local contract-only

Local 允许编译、lint、unit tests、schema/examples、legacy migration、DAG、CLI protocol、fake backend、run ID、auto detection 和 cancel/resume 状态机测试。Local 结果不得作为真实科学流程或性能结论。

### sbatch production

所有真实生信流程、科学 parity、SLURM cancel/resume/retry 和性能 benchmark 在 sbatch 集群执行。基础矩阵为：

```text
5 scenarios × 2 executors × 2 toolchains = 20 cells
```

每 cell 至少 3 次成功重复，cold/warm cache 分开，保留失败证据，并收集 queue、wall time、CPU、RSS、I/O、retry、artifact 指标。

## 兼容和迁移

- 当前 Snakemake 是生产兼容路径；Craftmake 的“默认目标”不能被描述为当前已完成状态。
- 默认切换和 Snakemake 删除是两个独立 gate。
- 历史归档、日期化审查报告和 remediation 证据不批量改写。
- 完整 Gate 0–7 见 [迁移路线](migration/craftmake-adoption.md)。

## 验收标准

- [x] 产品文档与发布入口使用 Otter 当前名称。
- [x] Craftmake、Enva 和领域算子存在独立模块边界。
- [x] Gate 0 的项目/config/reference/execution/workflow/benchmark 目标契约已写入 docs。
- [ ] Canonical typed resolver 生成 schema-valid、不可变 `run.yaml`。
- [ ] Craftmake 成为默认 executor，且只解析 `run.yaml`。
- [ ] backend/site auto detection 和 partial-SLURM fail-closed 通过测试。
- [ ] Reference registry、run override 和 explicit promote 完成。
- [ ] 五场景 Craftmake/Snakemake 与 modern/legacy-equivalent workflow 完整。
- [ ] 20-cell sbatch representative/scale matrix、故障注入和科学 parity 通过。
- [ ] 观察窗口、rollback 和迁移说明完成后，另行决定 Snakemake 退场。
