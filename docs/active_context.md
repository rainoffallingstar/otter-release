# System Context (Updated: 2026-07-22)

## 1. 已实现的核心模块 (Modules)

### config
- **Path**: `internal/config/`
- **Public Methods**:
  - `Load(path string): *XDXToolsConfig` - 加载 YAML 配置
  - `Validate(cfg *XDXToolsConfig) error` - 验证配置完整性
  - `Generate(opts Options) string` - 生成 YAML 配置
- **Data Flow**: YAML → Load → Validate → Engine Selection
- **Dependencies**: viper, yaml

### engine
- **Path**: `internal/engine/`
- **Public Methods**:
  - `CreateEngineFromConfig(cfg *EngineConfig) Engine` - 工厂方法
  - `Execute(cmd []string) error` - 执行命令
  - `GetStatus() *Status` - 获取状态
- **Data Flow**: Config → Engine Factory → Execute → Snakemake
- **Dependencies**: config

### input
- **Path**: `internal/input/`
- **Public Methods**:
  - `ScanFastq(dir string) ([]FastqFile, error)` - 扫描 FASTQ
  - `LoadPData(path string) (*PData, error)` - 加载 pdata
  - `GenerateAdapters(samples []Sample) error` - 生成接头序列
- **Data Flow**: FASTQ Directory → Scan → Pair → Validate → Config
- **Dependencies**: excelize

### workflow
- **Path**: `internal/workflow/`
- **Public Methods**:
  - `CreateDirectoryStructure(cfg *XDXToolsConfig) error` - 创建目录
  - `RunSnakemake(cfg *XDXToolsConfig, engine Engine) error` - 执行流程
- **Data Flow**: Config → Directory Setup → Snakemake Execution
- **Dependencies**: engine, config

### assets (Embedded Resources)
- **Path**: `internal/assets/`
- **Data**: Snakemake rules, Rscripts, conda envs
- **Dependencies**: go:embed

### enva
- **Path**: `internal/enva/`
- **Public Methods**:
  - `Run(env, binary string, args ...string) error` - 执行 conda 环境命令
- **Data Flow**: Shell Command → enva → rattler-native envs / compatibility package managers
- **Dependencies**: subprocess

### gomats (New)
- **Path**: `gomats/`
- **Public Methods**:
  - `gomats run --root --pdata --gtf --seqlengthQC` - 执行 rMATS 分析
- **Data Flow**: pdata + BAM → scan → compute N50 → pairwise combos → rmats.py
- **Dependencies**: excelize, cobra

## 2. 全局数据结构 (Global Types)

| Type Name | File Path | Key Fields | 使用场景 |
|-----------|-----------|------------|----------|
| XDXToolsConfig | internal/config/config.go | Workflow, Input, Output, Reference, Engine | 完整配置 |
| Sample | internal/input/fastq.go | Name, R1, R2, Metadata | FASTQ 样本 |
| PData | internal/input/pdata.go | Samples, Columns, Data | pdata 表格 |
| EngineConfig | internal/engine/types.go | Type, ParallelJobs, Partition | 引擎配置 |
| ContrastTask | gomats/internal/types/types.go | Species, Combination, B1Paths, B2Paths | rMATS 任务 |

## 3. CLI 端点注册表

| Command | File | Purpose |
|---------|------|---------|
| init | cmd/init.go | 安装 Snakemake 工作流文件 |
| create | cmd/create.go | 扫描 FASTQ，生成配置 |
| run | cmd/run.go | 执行 Snakemake 工作流 |
| config | cmd/config.go | 验证配置文件 |
| status | cmd/status.go | 显示工作流状态 |

## 4. Git Submodules (9 个)

| Submodule | Binary | Purpose | Review Status |
|-----------|--------|---------|---------------|
| enva | enva | rattler-first 环境管理器 | 待系统审查 |
| xenofilter-go | xenofilter | 污染过滤 | 整改提交 `6ae7f84` 已推送，主仓指针更新至该提交；0 Critical / 0 High / 0 Medium，真实 PDX 工作流集成待工具链 |
| Paireads | paireads | 配对 reads 处理 | 整改提交 `1ece0ba` 已推送，主仓指针更新至该提交；0 Critical / 0 High / 0 Medium，Bismark extractor 集成待工具链 |
| htseq2matrix-go | htseq2matrix | HTSeq 矩阵转换 | 待系统审查 |
| methrix-cli | methrix-cli | 甲基化分析 | 待系统审查 |
| qctb | qctb | 质量控制工具箱 | 待系统审查 |
| fastqc-rs | fqc | FastQC Rust 实现 | 待系统审查 |
| gomats | gomats | rMATS 可变剪接分析 | 待系统审查 |
| bamdriver-go | library | BAM/BGZF、BAI/FAI、排序与 NM 计算 | 已完成系统审查和整改 |

系统审查计划：`docs/review/submodule_review_plan_2026-07-21.md`。Wave 1 已完成 `xenofilter-go` 和 `Paireads` 的代码整改与本地动态审查，整改计划见 `docs/review/wave1_remediation_plan_2026-07-21.md`。两个子仓的已知 Critical/High/Medium 均已修复。`Paireads` 已补齐所有 BAM/BAI 目标路径的输入别名防护、备份阶段回滚错误传播、事务性多产物发布、加固 BAM 依赖、显式 64 MiB 最终排序预算、single group streaming、dual two-way streaming merge、显式 primary-only 语义和完整 CI 门禁；`go test`、`go vet`、`go test -race`、`govulncheck`、静态构建及 samtools 1.24 的 quickcheck/count/坐标排序/BAI 区域查询均通过。当前环境缺少 `bismark_methylation_extractor`、`snakemake`、`enva` 和 `conda`，因此主仓普通与 PDX Bismark extractor 真实集成仍待具备工具链的环境执行。

## 5. 待解决的技术债

- [ ] 在具备 Bismark/Snakemake 工具链的环境完成 Wave 1 普通与 PDX Bismark extractor 最小集成门禁
- [ ] 按子仓库系统审查计划完成其余 6 个子仓审查（高优先级）
- [ ] 更新 Snakemake 规则文档 (中优先级)
- [ ] 添加更多集成测试用例 (中优先级)
- [ ] 优化大项目并行性能 (低优先级)
