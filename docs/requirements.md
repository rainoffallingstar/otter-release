# otter 需求文档

## 产品目标

`otter` 将 RRBS、WGBS、RNA-seq 和 PDX 分析统一为可配置、可恢复、可在 local/SLURM 运行的 CLI 工作流。主仓是 `rainoffallingstar/otter`。

目标层级：

```text
otter → craftmake → enva → 算子 → bamdriver
```

## 功能需求

### 1. 协调层：otter

- 提供 `init`、`create`、`run`、`task`、`status`、`config` 命令。
- 扫描配对 FASTQ，解析 Excel/CSV pdata，生成并验证 `OtterConfig`。
- 支持 RRBS、WGBS、RNA-seq 和通过第二物种启用的 PDX。
- 支持后台任务、日志、停止、状态恢复和 dry-run。
- 源码、CLI、默认配置与状态路径直接使用 `otter`/`OtterConfig`，不提供旧产品名兼容别名。

### 2. 执行层：craftmake

- 作为 Snakemake 的 Go 替代执行层，提供 typed workflow spec、DAG、local 和 SLURM 后端。
- 在接入完成前与 Snakemake 双轨运行。
- 对现有 workflow 的任务依赖、资源、参数引用、失败传播和恢复语义提供等价验证。
- 未完成四种模式集成门禁前，Snakemake 仍是生产兼容路径。

### 3. 环境层：enva

- rattler-first 创建、发现、运行、安装、接管和删除环境。
- 预定义环境统一为 `otter-core`、`otter-snakemake`、`otter-extra`。
- conda/mamba/micromamba 只作为兼容发现或显式接管路径。
- 名称歧义、路径越界和不受支持操作必须 fail closed。

### 4. 算子层

| 算子 | 需求 |
|---|---|
| `fastqcx` | FASTQ 质控；保留 FastQC/MultiQC 可消费格式 |
| `xenofilx` | PDX graft/host 读段分类 |
| `pairbam` | 保留完整配对 BAM 读段 |
| `seq2mat` | 严格解析 HTSeq 计数并生成可追溯矩阵 |
| `matsrun` | 安全编排 rMATS 对比任务 |
| `qctb` | 为 BS-seq、RNA-seq、PDX 等模式生成版本化 QC 结果 |
| `methx` | 生成和验证原生甲基化/HDF5 产物，并准确声明 Methrix 兼容边界 |

### 5. BAM 基础层

- `bamdriver` 提供可复用的 BAM 读取、写入、过滤和相关低层操作。
- 上层算子不得各自复制不一致的 BAM 边界逻辑。

## 非功能需求

### 可重复性

- 配置、环境、算子版本和产物 schema 可追溯。
- 产物优先通过 staging、验证和原子发布生成。
- 外部标准名和科学语义不得因产品重命名而改变。

### 性能与执行

- 主 CLI 保持单二进制部署能力。
- 支持 local 与 SLURM 并发；资源覆盖必须显式且可验证。
- 大 FASTQ、BAM、count 和 HDF5 数据链需要受控内存与确定性输出。

### 兼容与迁移

- 迁移期保留 Snakemake 与 `otter-snakemake` 执行路径；产品和算子旧命令名不提供兼容别名。
- `craftmake` 替换完成必须由测试证据定义，而不是由文档声明。
- 历史归档和日期化审查报告不得批量重写。

## 技术约束

- 父仓 Go 1.24+。
- Rust 算子使用锁定依赖并通过 fmt/check/clippy/test。
- Go 算子通过 test/vet，适用时包含 race/static build。
- HDF5、Bismark、HTSeq、FastQC、MultiQC、Methrix、rMATS 等外部契约按实际集成保留。

## 验收标准

- [ ] `otter` 文档和发布入口统一指向 `rainoffallingstar/otter`。
- [ ] 三个环境均使用 `otter-*` 名称并通过干净环境验证。
- [ ] 新子模块名在当前文档、安装清单和构建清单一致。
- [ ] RRBS/WGBS/RNA-seq/PDX 在当前 Snakemake 路径均通过 smoke test。
- [ ] craftmake 双轨路径通过任务图、资源、恢复和关键产物等价测试。
- [ ] 满足上述门禁后，才可移除 Snakemake 兼容路径。
