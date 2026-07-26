# System Context (Updated: 2026-07-26)

## 1. 已实现的核心模块 (Modules)

### Otter 主协调层
- **Path**: `cmd/`, `internal/config/`, `internal/input/`, `internal/engine/`, `internal/workflow/`
- **Public Methods**: `init/create/run/task/status/config`
- **Current Data Flow**: FASTQ/pdata → `OtterConfig` → Snakemake → `enva` → 算子
- **Target Data Flow**: project/samples/reference/site → immutable `run.yaml` → Craftmake(default) 或 Snakemake(explicit compatibility)
- **Status**: 当前生产路径仍是 Snakemake；Gate 1 canonical resolver 已实现，Craftmake 默认生产路由尚未接入

### Craftmake 执行层
- **Path**: `craftmake/`
- **Public Methods**: `validate/plan/run/resume/status/logs/report/cancel/doctor`
- **Data Flow**: workflow/config → DAG → Local/SLURM → SQLite/metrics
- **Status**: 核心执行能力存在；目标改为只解析 resolved `run.yaml`，尚未完整接入 Otter

### Enva 环境层
- **Path**: `enva/`
- **Public Methods**: `create/list/run/install/adopt/remove`
- **Data Flow**: 环境请求 → sibling staging → residual 验证/原子发布 → 隔离命令
- **Status**: Rattler 增量安装、active root 和 Mamba 2 argv 修复已发布；GitHub E2E 已通过验收

### 生信算子与 BAM 基础层
- **Path**: `fastqcx/`, `xenofilx/`, `pairbam/`, `seq2mat/`, `matsrun/`, `qctb/`, `methx/`, `bamdriver/`
- **Data Flow**: FASTQ/BAM/count/coverage → 专用算子 → 科学结果/QC
- **Dependencies**: FastQC/MultiQC、Bismark、HTSeq、Methrix/HDF5、rMATS 等外部契约

## 2. 已冻结的目标契约

| Contract | Path | Key Decision |
|---|---|---|
| Project layout | `docs/project-layout.md` | 项目/run 分离；pinned workflow assets |
| Configuration | `docs/configuration.md` | canonical project v1 → immutable run v1 |
| Execution | `docs/execution-contract.md` | Craftmake 默认；Snakemake 显式；单 `run.yaml` |
| Site/backend | `docs/site-profiles.md` | auto 完整 SLURM→slurm、无 SLURM→Local、partial→fail closed |
| References | `docs/reference-registry.md` | shared registry、project lock、run override、explicit promote |
| Workflows | `docs/workflow-catalog.md` | RRBS/WGBS/RNA-seq/BS-PDX/RNA-PDX；modern/legacy-equivalent |
| Validation | `docs/benchmark-plan.md` | Local contract-only；真实流程与 benchmark 统一 sbatch |
| Migration | `docs/migration/craftmake-adoption.md` | Gate 0–7；默认切换与 Snakemake 退场分离 |

## 3. 全局数据结构 (Global Types)

| Type Name | File Path | Status |
|---|---|---|
| `OtterConfig` | `internal/config/config.go` | legacy compatibility；迁移读取不应用环境变量 override |
| `ProjectConfig` / `RunSnapshot` | `internal/config/v1/types.go` | canonical v1 typed model 已实现 |
| `SampleRecord` | `internal/config/v1/types.go` | `samples.tsv` typed row；严格列与唯一 ID |
| `ReferencesLock` / `ReferenceDefinition` | `internal/config/v1/types.go` | lock/registry typed model 与 digest 校验 |
| `Resolver` | `internal/config/resolver/resolver.go` | project/samples/lock/registry → immutable snapshot |
| `Engine` | `internal/engine/engine.go` | 当前 local/SLURM 边界；Gate 2 后收敛为 adapter |
| `WorkflowSpec` | `craftmake/internal/spec/` | 当前 Craftmake DAG 契约 |

## 4. Run 与验证决策

- Run ID：`run-YYYYMMDDTHHMMSSZ-abcdef`；UTC 秒级时间戳 + 6 位安全随机小写英文
- Executor：默认 `craftmake`；`snakemake` 只能显式选择且不作为 fallback
- Backend：默认 `auto`；Local 只做代码/协议验证
- Production：所有真实科学流程、cancel/resume、parity、benchmark 使用 sbatch
- Matrix：5 scenarios × 2 executors × 2 toolchains = 20 cells；每 cell 至少 3 次成功重复
- Reference：项目默认可由单次 run 覆盖；resume 不可切换；只有 explicit promote 修改默认

## 5. 待解决的技术债

- [x] Gate 0 文档、JSON Schema、YAML 示例、五场景 catalog、benchmark 与 Gate 0–7 路线已建立
- [x] 验证 schema/examples 和文档链接后冻结 Gate 0
- [x] Gate 1：实现 typed v1、严格 samples/reference 解析、resolver、run ID、immutable snapshot、digest drift 与 legacy adapter
- [ ] Gate 2：将 production `otter run` 接入 Otter–Craftmake JSON 协议和默认 executor 路由
- [ ] Gate 3：实现 site/backend auto detection；当前 `config resolve` 对 auto fail closed
- [ ] Gate 4：实现 registry publication、site mount、reference promote 与 sbatch preflight
- [ ] Gate 5：实现五场景双 executor、双 toolchain workflow/artifact contract
- [ ] Gate 6：执行 sbatch 20-cell parity/benchmark 和故障注入
- [ ] Gate 7：观察默认稳定性后独立评估 Snakemake 退场
- [x] Enva GitHub E2E 已通过三个环境与 compatibility manager 验收
