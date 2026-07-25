# System Context (Updated: 2026-07-25)

## 1. 已实现的核心模块 (Modules)

### Main CLI
- **Path**: `cmd/`, `internal/config/`, `internal/input/`, `internal/engine/`, `internal/workflow/`
- **Public Flow**: `init` → `create` → `run` → `status`
- **Data Flow**: FASTQ/pdata → typed config → engine selection → Snakemake → submodule tools
- **Current Gate**: `go test -count=1 ./...` 与 `go vet ./...` 通过

### Workflow Assets
- **Path**: `inst/rules/`, `internal/assets/`
- **Current State**: gomats/qctb/fqc 命令参数使用 Snakemake `{value:q}`；config、pdata、GTF 已作为显式 input；静态契约测试已添加
- **Pending**: 真实 Snakemake、rMATS、Bismark 与四条关键数据链 smoke test

### Submodule Remediation
- **Plan**: `docs/review/wave2_wave3_remediation_plan_2026-07-22.md`
- **Scope Decision**:
  - `qctb` 采用版本化原生 Excel/TSV schema，不兼容历史 R QC，不生成 RDS
  - `methrix-cli` 采用 `methrix-cli.custom-hdf5/1.0.0`，不兼容 R loader
- **Delivery Rule**: 当前改动未提交、未推送；完成子仓门禁后再更新主仓指针

## 2. 全局数据结构 (Global Types)

| Type Name | File Path | Key Fields | 使用场景 |
|---|---|---|---|
| `XDXToolsConfig` | `internal/config/config.go` | Workflow, Input, Output, Reference, Engine | 主仓配置契约 |
| `Engine` | `internal/engine/engine.go` | Execute, Status, Wait, Kill | SLURM/local 执行边界 |
| `EnvironmentName` | `enva/src/backend/types.rs` | 单一正常路径组件 | 环境路径安全边界 |
| `EnvironmentResolution<T>` | `enva/src/backend/types.rs` | NotFound, Unique, Ambiguous | 名称解析的 typed fail-closed 结果；mutating/run 操作要求唯一，歧义时要求显式 prefix |
| `BackendCapabilities` | `enva/src/backend/types.rs` | Native, Delegated, Hybrid, Unsupported | rattler 与 CLI compatibility 后端的 typed capability matrix；命令入口在初始化具体操作前验证支持级别 |
| `CondaPrefixLayout` | `enva/src/backend/rattler.rs` | BaseRoot, ManagedEnvironment, ExternalEnvironment | 活跃 conda prefix 的 fail-closed root 推导 |
| `ContrastTask` | `gomats/internal/types/types.go` | Species, Combination, BAM paths | rMATS 任务契约 |
| `Count` | `htseq2matrix-go/internal/htseq/types.go` | 精确非负整数 | HTSeq count 解析 |
| Custom HDF5 Schema | `methrix-cli/src/hdf5/` | schema name/version, assays, row/col metadata | 原生甲基化产物契约 |

## 3. 子仓状态

| Submodule | 已完成 | 当前门禁 | 剩余发布阻塞 |
|---|---|---|---|
| `enva` | 路径/删除边界、argv 保留、symlink 拒绝、ownership fail-closed、force staging/rollback、版本化 journal、prefix/cache 跨进程锁、cache marker、fail-closed `CONDA_PREFIX` root 推导、canonical prefix identity 去重、同名环境 unique-or-ambiguous fail-closed 解析、typed capability matrix、external remove 显式 adoption、rattler 默认后端、兼容后端只发现已安装二进制、MatchSpec 边界、结构化退出码、rattler dry-run 真实 solve、native create/install sibling staging 原子发布、内部 hard-link 保留且不回链源 prefix、逃逸 symlink 拒绝、文本与 symlink relocation、二进制残留拒绝、无 reflink 大 prefix 门禁、conda/mamba/micromamba 真实 compatibility smoke matrix | fmt/locked all-target/all-feature check/严格 clippy/test 通过（101 deterministic library + 4 CLI）；diff/workflow 静态检查通过 | GitHub-hosted compatibility smoke matrix 待 CI 实跑；Windows 延期 |
| `gomats` | pdata/BAM 完整性、安全 contrast、零任务失败、staging、manifest、产物验证与原子发布 | test/vet/race/static build 通过 | 真实 rMATS、PDX species→GTF 契约与主仓集成 |
| `fastqc-rs` | malformed/truncated FASTQ fail-closed、IUPAC、碱基级质量统计、A-T/G-C、确定性输出、单一 parser、typed k-mer/read/position/length 上限、离线安全 HTML 生成、FastQC-compatible summary 与 SeqKit 2.13 语义 | locked fmt/check/严格 clippy/test/build/diff check 通过（23 unit + 4 CLI）；100,000×150 bp 预算 28.847 秒/14,208 KiB；FastQC 0.12.1、SeqKit 2.13.0、MultiQC 1.35 和 qctb 真实消费通过 | GitHub-hosted compatibility workflow 待远端实跑；commit/push 与主仓指针待授权 |
| `htseq2matrix-go` | 严格非负整数解析、one-to-many mapping 保留、显式 unmapped retain/one-to-many expand 契约、union gene universe、缺失值补零、duplicate symbol 逐列最大值、symbol ascending 稳定行序、mapping/matrix 版本化 provenance、count/norm/manifest 三产物同事务发布与故障回滚 | test/vet/race/build/diff check 通过 | production mouse mapping namespace 仍需独立验证或再生成 |
| `qctb` | 当前 config/Methrix 路径、严格 FQC、原子输出 | fmt/clippy/test 通过 | 原生 schema、四模式 golden、其余 parser、主仓集成 |
| `methrix-cli` | custom schema 元数据、多产物事务、下载边界 | fmt/clippy/test 通过 | 原生 validator、分块写、genome hash、主仓/qctb 集成 |
| `xenofilter-go` | Critical/High/Medium 已清零 | 已推送 | 真实 PDX 集成 |
| `Paireads` | Critical/High/Medium 已清零 | 已推送 | 真实 Bismark 集成 |
| `bamdriver-go` | 系统审查和整改完成 | 通过 | 无当前阻塞 |

## 4. 实施批次

1. **S0 稳定基线**: `enva` 测试稳定化、Git 凭据轮换、工作树清理
2. **S1 防止数据破坏/虚假成功**: gomats staging；enva transaction/lock/cache；兼容后端边界简化
3. **S2 冻结原生契约**: htseq matrix、qctb report、methrix custom HDF5
4. **S3 规模与异常输入**: methrix 流式写入/genome provenance；fastqc 资源与格式验证
5. **S4 跨仓集成**: RNA splicing、QC、甲基化、RNA count
6. **S5 发布交付**: 子仓提交推送、主仓指针、干净环境复验和报告收口

## 5. 待解决的技术债

- [x] `enva` 默认并行完整测试连续三次通过
- [ ] 轮换 Git remote 暴露凭据并清理明文 URL（主仓、gomats、htseq2matrix-go、methrix-cli；需要账户侧轮换）
- [x] `gomats` staging/原子发布和 rMATS 关键产物校验（fake rMATS 门禁通过；真实 rMATS 集成待工具链）
- [x] `enva` force transaction、prefix/cache 跨进程锁和 cache marker
- [x] `enva` 明确以 rattler 为默认后端；micromamba 自动下载已退出正式路径，兼容模式仅接受 `ENVA_MICROMAMBA_PATH` 或 `PATH` 中已有且可执行的二进制
- [x] `enva` 中断 journal/recovery：追加式版本化 journal、截断尾记录容忍、create/install/adopt/remove/run/list/find 入口恢复、强制进程退出和 backup cleanup/parent sync 故障注入
- [x] `enva` rattler dry-run 真实 solve、`--with` 求解、MatchSpec 逗号边界、退出码 141 传播和显式 adoption 边界
- [x] `enva` 真实 rattler create/install/remove、native install staging/原子发布、solver 失败与 Installer 后故障保留旧环境、Python 最终 prefix 运行校验
- [x] `enva` 真实 cache 并发集成：真实 conda-forge solve 持有 `CacheUse` 时，独立 cleanup 子进程保持阻塞；释放后 cleanup 删除 `pkgs`/`repodata`/`run-exports` 并保留 ownership marker
- [x] `enva` `CONDA_PREFIX` root 推导：仅 `<root>/envs/<name>` 或 `CONDA_DEFAULT_ENV=base` 可生成 root；外部、自定义和相对 prefix 不再被误当 root
- [x] `enva` canonical environment identity 与歧义名称收口：存在 prefix 以 canonical path 去重；run/activate/install/adopt/remove 名称解析要求唯一；不同物理 prefix 同名时 fail closed，install/remove 支持显式 `--prefix`，歧义失败不写 ownership 且不删除任何 prefix
- [x] `enva` typed capability matrix 与 external remove ownership 收口：rattler/CLI compatibility 操作声明为 Native/Delegated/Hybrid/Unsupported；命令入口 fail closed；rattler remove 不再隐式 adopt，external 环境必须先显式 `enva adopt` 或由原包管理器删除；拒绝路径不写 marker、不删除 prefix
- [x] `enva` native install 边缘：escaping symlink 拒绝、内部 hard-link 保留且不回链源 prefix、文本/绝对 symlink relocation、二进制 residual 拒绝、无 reflink 2,000 文件大 prefix 门禁；Windows 兼容延期
- [x] 冻结 `htseq2matrix-go` mapping/missing/unmapped 原生契约与 mapping/matrix provenance；count/norm/manifest 同事务发布
- [ ] 验证或重新生成 production mouse mapping namespace
- [ ] 建立 `qctb.report` 版本和 RRBS/WGBS/RNA/PDX golden fixtures
- [ ] 实现 methrix custom HDF5 原生 validator 与分块写入
- [x] 完成 FastQC/MultiQC/SeqKit 外部验证、qctb 真实消费和 100,000×150 bp 资源预算；GitHub-hosted compatibility job 待远端实跑
- [ ] 完成主仓四条关键数据链 smoke tests
- [ ] 更新六份审查报告、提交子仓并更新主仓子模块指针
