# System Context (Updated: 2026-02-28)

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
- **Data Flow**: Shell Command → enva → conda/mamba/micromamba
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
| tui | cmd/tui.go | 交互式终端 UI |

## 4. Git Submodules (8 个)

| Submodule | Binary | Purpose |
|-----------|--------|---------|
| enva | enva | micromamba 环境管理器 |
| xenofilter-go | xenofilter | 污染过滤 |
| Paireads | paireads | 配对 reads 处理 |
| htseq2matrix-go | htseq2matrix | HTSeq 矩阵转换 |
| methrix-cli | methrix-cli | 甲基化分析 |
| qctb | qctb | 质量控制工具箱 |
| fastqc-rs | fqc | FastQC Rust 实现 |
| gomats | gomats | rMATS 可变剪接分析 |

## 5. 待解决的技术债

- [ ] 更新 Snakemake 规则文档 (中优先级)
- [ ] 添加更多集成测试用例 (中优先级)
- [ ] 优化大项目并行性能 (低优先级)
