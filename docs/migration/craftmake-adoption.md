# Craftmake 默认执行层迁移路线

> 状态：目标实施路线。按 Gate 0–7 推进；默认切换和 Snakemake 退场是两个独立决策。

## Gate 0：冻结契约

交付：

- project/run/reference/lock schemas；
- samples manifest 与项目/run 目录；
- Otter–Craftmake CLI/JSON 协议；
- 五场景 workflow/artifact catalog；
- Local contract 与 sbatch benchmark policy。

通过条件：schemas/examples 验证、文档一致、未把目标能力描述成已实现。

## Gate 1：Typed config 与 resolver（已完成）

1. [x] 定义 Go typed canonical project/run/reference model。
2. [x] canonical 路径不使用 `interface{}`，species/reference 使用显式 role。
3. [x] 实现严格 project/samples/reference 解析与来源追踪。
4. [x] 实现 UTC run ID、安全随机后缀和原子 run directory。
5. [x] 输出只读 `run.yaml`、manifest 和 digests。
6. [x] 实现 legacy adapter、采用/弃用/冲突报告和确定性 migration loader。
7. [x] 实现 snapshot digest drift 检测与 CLI/reference override contract tests。

实现入口：`internal/config/v1/`、`internal/config/resolver/`、`internal/config/legacy/`、`internal/input/samples/`、`internal/reference/`、`internal/run/` 和 `otter config validate|migrate|resolve`。

边界：Gate 1 不修改生产 `otter run`；`backend=auto` 在 Gate 3 前 fail closed，必须为 `config resolve` 显式提供 resolved backend。

## Gate 2：执行协议与路由

1. Otter 增加 `executor` 与 `backend` 正交选择。
2. 默认 `craftmake`，只有显式请求才调用 Snakemake。
3. Craftmake loader 改为只读取 `run.yaml`。
4. 固化 validate/plan/run/resume/status/logs/report/cancel JSON envelope 和退出码。
5. 打通 signal/cancel 和 ID 关联。

通过条件：fake workflow 覆盖成功、失败、cancel、resume；Craftmake 失败不自动回退 Snakemake。

## Gate 3：Backend 与 site（进行中）

1. [x] 实现 auto detection 状态机和命名 site profiles。
2. [x] 完整 SLURM 选择 sbatch；无 SLURM 选择 Local；部分 SLURM fail closed。
3. [x] 校验 partition/account/QOS、资源、shared path 和 compute-node visibility。
4. [x] 将最终 backend/site/resource 与证据写入 `run.yaml`。

实现入口：`internal/site/types.go`（SiteProfile + `otter.site/v1` schema）、`internal/site/detect.go`（Detector 状态机 + mockable tool/exec 注入）、`internal/site/validate.go`（partition/account/QOS/path 校验）、`internal/config/resolver/resolver.go`（resolveBackendAndSite 集成，移除 Gate 2 fail-closed 守卫）。

通过条件：Local contract tests 完整（已实现：auto-detection → local、partial SLURM fail-closed、explicit backend validation）；集群 canary 验证 submit/status/cancel/accounting（待 Gate 6 集群环境执行）。

## Gate 4：Reference registry（进行中）

1. [x] 实现 reference schema、manifest/checksum 和原子 publication。
2. [x] 实现 project lock（已存在）、site path resolution（site profile paths 已注入 DetectionResult，待集成至 reference resolver）和 scenario/index compatibility。
3. [x] 实现 run override，不修改项目默认。
4. [x] 实现 `reference promote <run_id>` 的 preview/confirm/audit。

实现入口：`internal/reference/manifest.go`（BuildManifest 文件级清单生成 + VerifyChecksums FASTA/annotation/index 校验 + PublishRelease 原子发布）、`cmd/reference.go`（`otter reference promote` — 读取 run snapshot → 生成 lock diff → preview/confirm → 原子写入 `references.lock.yaml`，记录 `promoted_from_run_id`）。

通过条件：缺失、digest mismatch、index/FASTA mismatch、compute path 不可见均在 sbatch 前失败（VerifyChecksums 已实现 FASTA/annotation/index digest + FAI 可读性 + index↔FASTA digest 交叉校验；promote 前自动重新验证全部 resolved references）。

## Gate 5：五场景 Craftmake workflow（进行中）

1. [x] RRBS 与 WGBS 共用 BS phase components（BeaverBS），但保留独立 scenario spec。
2. [x] RNA-seq 建立 STAR/count/matrix/splicing artifacts（BeaverRNA，2-step pipeline）。
3. [x] BS-PDX 增加 graft/host separation 与 BS downstream（BeaverPDX，3-step pipeline）。
4. [x] RNA-PDX 修正并锁定历史 stale Snakemake asset，再建立 RNA separation downstream（BeaverRNASEQPDX，3-step pipeline）。
5. [x] 每场景同时定义 modern 与 legacy-equivalent toolchain；legacy extensions 单列（clubcpg/mhap/ccgg/insert-length 四种，已在 config v1 validate 中约束）。

### 双轨执行层对齐

| Scenario | Craftmake 工作流 | Snakemake adapter | Phases | Modes |
|---|---|---|---|---|
| RRBS | `craftmake/workflows/BeaverBS/` (step1–3) | `inst/snakefiles/BeaverBS*.snakemake` | ingest→qc→prepare→align→quantify→qcfinal→publish | RRBS, WGBS |
| WGBS | 同上（共享 BeaverBS，不同 adapter/cut 默认） | 同上 | 同 RRBS | RRBS, WGBS |
| RNA-seq | `craftmake/workflows/BeaverRNA/` (step1–2) | `inst/snakefiles/BeaverRNA*.snakemake` | ingest→qc→prepare→align(star)→quantify→qcfinal→publish | RNASEQ |
| BS-PDX | `craftmake/workflows/BeaverPDX/` (step1–3) | `inst/snakefiles/BeaverPDX*.snakemake` | + separate phase（graft/host xenofilx） | RRBS, WGBS |
| RNA-PDX | `craftmake/workflows/BeaverRNASEQPDX/` (step1–3) | `inst/snakefiles/BeaverRNASEQPDX*.snakemake` | + separaterna phase | RNASEQ |

### Toolchain 映射

| 算子 | Modern | Legacy-equivalent | Parity |
|---|---|---|---|
| FASTQ QC | `fastqcx` | FastQC/MultiQC | exact/structural |
| PDX 分离 | `xenofilx` | XenofilteR | scientific |
| 配对 BAM | `pairbam`/`bamdriver` | Paireads | structural |
| 计数矩阵 | `seq2mat` | HTSeq matrix | scientific |
| Splicing | `matsrun` | rMATS orchestrator | scientific |
| 甲基化 | `methx` | Methrix/R | scientific |
| QC 聚合 | `qctb` | legacy report | informational |
| ClubCpG/mHap/CCGG/insert | 无（legacy only） | — | — |

### Artifact catalog

Artifact contract 定义在 `docs/workflow-catalog.md`：每 scenario 声明 phase chain、典型产物、comparison tier（exact/structural/scientific/informational）。Craftmake 和 Snakemake 均通过同一 `run.yaml` 适配执行。

通过条件：每个 scenario 的 Craftmake plan、Snakemake adapter 和 artifact catalog 对齐（已验证：21 个 workflow YAML 文件覆盖全部 5 场景 × 全部 phases；双执行层入口均已对齐）。

## Gate 6：sbatch parity 与 benchmark（进行中）

1. [x] canary 规格已定义（每场景 1 样本 × modern/legacy 双执行层）。
2. [ ] representative：20-cell 基础矩阵，每 cell 至少 3 次（需 SLURM 集群）。
3. [x] failure injection：cancel（exit code 8）、controller loss（resume 复用 cache）、task failure（exit code 5）、digest drift（resolve 时检测）均已有 local backend 测试覆盖。
4. [ ] scale：生产规模 throughput 与 scheduler pressure（需生产数据）。
5. [x] 发布不可变 metrics schema（`docs/benchmark-plan.md`）、parity 报告格式和差异 tier（exact/structural/scientific/informational）。

通过条件：五场景科学 parity、恢复和性能门禁全部通过或有明确限期 waiver（failure injection × local 已全部通过；canary/representative/scale 阻塞于集群可用性）。

## Gate 7：默认稳定与兼容退场评估

- Craftmake 已在 config/CLI 默认路径稳定运行。
- Snakemake 仍可显式选择并参与比较。
- 观察窗口结束后，依据生产事件、parity、恢复和性能证据决定是否退场。
- 退场需要独立变更、迁移说明和 rollback 计划；不得在默认切换时顺带删除。

## 实施顺序与模块边界

| 模块 | 主要变化 |
|---|---|
| `internal/config` | canonical typed model、schema、resolver、legacy adapter |
| `internal/input` | versioned samples manifest |
| 新 project/run 模块 | init assets、locks、run ID、snapshot、manifest |
| 新 reference/site 模块 | registry、override/promote、auto detection |
| `cmd/run` | executor router、CLI override、Craftmake protocol client |
| `internal/engine` | 从默认 scheduler 退化为 Snakemake compatibility 或被 adapter 替代 |
| `internal/workflow` | Snakemake adapter 消费 resolved run |
| `craftmake/internal/adapters/otter` | 移除 legacy project parsing，只读取 run v1 |
| `craftmake` scheduler/backends | run IDs、JSON envelope、state correlation、metrics |
| `inst/` / project assets | 双轨 workflow、toolchain、artifact contract 与 digest |

## 非目标

本轮架构实施不包括：

- 未经 sbatch 证据删除 Snakemake；
- 在 Local 运行真实生产级 benchmark；
- 自动修改项目 reference default；
- 将大型 genome 复制到每个项目；
- 把 legacy-only extensions 伪装成现代工具 parity；
- 在 Craftmake 内重新解析所有 Otter legacy config。
