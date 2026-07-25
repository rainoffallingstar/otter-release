# System Context (Updated: 2026-07-25)

## 1. 已实现的核心模块 (Modules)

### Otter 主协调层
- **Path**: `cmd/`, `internal/config/`, `internal/input/`, `internal/engine/`, `internal/workflow/`
- **Public Methods**: `init/create/run/task/status/config` - 项目、配置、执行与状态入口
- **Data Flow**: FASTQ/pdata → `OtterConfig` → 工作流执行层 → `enva` → 算子 → `bamdriver`
- **Dependencies**: Go、Cobra、嵌入工作流资产；根命令、状态目录与源码类型已硬切为 `otter`

### Fastqcx 质控工作流契约
- **Path**: `inst/rules/`, `inst/snakefiles/`, `internal/assets/`
- **Public Rules**: `fastqcxAtfirst`, `fastqcxAtclean`
- **Data Flow**: 原始/清洗 FASTQ → `fastqcx` → `*_fastqcx/fastqc_data.txt` → checker 与 qctb QC summary
- **Dependencies**: 当前规则文件为 `01fastqcxAtfirst.smk`、`03-0-fastqcxAtclean.smk`；外部 FastQC/MultiQC 协议名保持不变

### Craftmake 执行层
- **Path**: `craftmake/`
- **Public Methods**: `craftmake` CLI/protocol - Go 工作流编译与执行
- **Data Flow**: otter 工作流描述 → DAG/后端 → local/SLURM
- **Dependencies**: otter 适配器、执行后端
- **Integration Status**: 迁移中/双轨；生产路径仍使用 Snakemake，尚未完成单轨替换

### Enva 环境层
- **Path**: `enva/`
- **Public Methods**: `create/list/run/install/adopt/remove` - 环境生命周期
- **Data Flow**: 执行请求 → `otter-core`/`otter-snakemake`/`otter-extra` → 隔离命令
- **Dependencies**: rattler；conda/mamba/micromamba 仅作兼容发现或接管

### 生信算子与 BAM 基础层
- **Path**: `fastqcx/`, `xenofilx/`, `pairbam/`, `seq2mat/`, `matsrun/`, `qctb/`, `methx/`, `bamdriver/`
- **Public Methods**: 各 CLI 的稳定参数与文件契约
- **Data Flow**: FASTQ/BAM/count/coverage → 专用算子 → 科学结果与 QC；BAM 操作下沉到 `bamdriver`
- **Dependencies**: FastQC/MultiQC、Bismark、HTSeq、Methrix、rMATS 等外部标准或领域契约

## 2. 全局数据结构 (Global Types)

| Type Name | File Path | Key Fields | 使用场景 |
|---|---|---|---|
| `OtterConfig` | `internal/config/config.go` | Workflow, Input, Output, Reference, Engine | 主配置契约；源码直接实现且不提供旧类型别名 |
| `Engine` | `internal/engine/engine.go` | Execute, Status, Wait, Kill | local/SLURM 执行边界 |
| `WorkflowSpec` | `craftmake/internal/spec/` | DAG, tasks, resources | craftmake 工作流编译契约 |

## 3. API 端点注册表

| Method | Endpoint | Request Type | Response Type |
|---|---|---|---|
| CLI | `otter init/create/run/task/status/config` | typed CLI args / `OtterConfig` | files, task state, exit status |
| CLI | `craftmake`（迁移中） | workflow spec | DAG execution result |
| CLI | `enva` | environment command types | environment result |

## 4. 当前命名迁移

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

## 5. 待解决的技术债

- [x] 10 个子仓完成新身份提交与推送；父仓 gitlink 指向对应新提交
- [x] `bamdriver` 新 module path 已发布，`xenofilx` 与 `pairbam` 已固定可解析 pseudo-version
- [ ] 观察改名后各仓首次 CI 与 release workflow，处理平台特有失败
- [ ] 完成 `craftmake` 与 otter 的集成测试并确定 Snakemake 退场门禁
- [ ] 完成 RRBS/WGBS/RNA-seq/PDX 双轨 smoke tests
- [x] GitHub PAT 已轮换；本地 remote 已清理为无凭据 HTTPS
