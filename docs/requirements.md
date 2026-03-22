# 需求文档 (Requirements)

## 项目概述

将现有 xdxtools R 包重构为 Go 技术栈的命令行工具，专注于生物信息学工作流管理，支持 RRBS、WGBS、RNA-seq 和 PDX 流程，目标是简化部署和运维。

## 功能需求

### 1. 工作流管理
- **RRBS** (Reduced Representation Bisulfite Sequencing) - 3步流程
- **WGBS** (Whole Genome Bisulfite Sequencing) - 3步流程
- **RNA-seq** (RNA Sequencing) - 2步流程
- **PDX** (Patient-Derived Xenograft) - 自动识别双物种，3步流程

### 2. 配置管理
- 支持 YAML/TOML/JSON 配置文件
- 环境变量覆盖配置
- 50+ 参数集中管理
- 配置文件验证和智能提示

### 3. 输入处理
- **FASTQ 文件管理**
  - 自动识别配对样本（R1/R2）
  - 可配置后缀（`_R1.fastq.gz` → `_R2.fastq.gz`）
  - 支持压缩/非压缩文件
  - 灵活命名约定（`_1/_2`, `_fwd/_rev`）
- **pdata 表格解析**
  - 支持 Excel (.xlsx) 和 CSV 格式
  - 验证样本 ID 匹配
  - 自动生成表型数据

### 4. 执行引擎
- **Slurm 集群** - Job Array 执行
- **SLURM Job Array** - 并行样本处理
- **本地执行** - 直接运行 Snakemake
- **统一接口** - 引擎切换无缝

### 5. 基因组构建
- **Bismark 索引** - 甲基化分析
- **STAR 索引** - RNA-seq 比对
- **引擎调用** - 使用统一执行引擎
- **多物种支持** - PDX 模式自动识别

### 6. 外部脚本调用
- **R 脚本执行** - 通过 Rscript 调用
- **Python 脚本执行** - 直接运行
- **Conda 环境隔离** - 每个工具独立环境
- **参数传递** - 灵活的参数映射

### 7. 用户界面
- **CLI 命令行工具**
  - `init` - 项目初始化
  - `run` - 执行工作流
  - `config` - 配置管理
- **终端输出** - 结构化日志与状态输出
  - 实时进度与错误信息
  - 适合 CLI/批处理环境

### 8. 结果处理
- **结果聚合** - 自动打包结果
- **日志管理** - 结构化日志输出
- **状态追踪** - 工作流状态记录

## 非功能需求

### 性能
- 启动时间 < 5秒
- 内存使用 < 100MB
- 并行处理支持

### 可用性
- 跨平台支持 (Linux/Windows/macOS)
- 主 CLI 可单二进制部署
- 运行仍依赖 Snakemake 与 Conda/Enva 环境

### 可维护性
- 模块化设计
- 统一引擎接口
- 配置集中管理

### 兼容性
- 兼容现有 Snakemake 文件
- 保持参数名称一致性
- 平滑迁移路径

## 约束与假设

### 技术约束
- 基于 Go 1.24+
- 使用 Snakemake 工作流引擎
- 依赖 Conda 环境管理
- 支持 Python 3.8+ 和 R 4.0+

### 输入约束
- FASTQ 文件需遵循命名规范
- pdata 文件需包含必要列
- 参考基因组需预构建或自动下载

### 运行环境
- 需要安装 Snakemake
- Slurm 模式需 Slurm 集群
- 本地模式需 worker pool 并行

## 关键用例

### 用例 1: 项目初始化
```
用户: xdxtools init my_project
系统: 创建项目目录 → 提取 Snakemake 文件 → 生成资产清单
```

### 用例 2: 配置生成
```
用户: xdxtools create --fastq ./fastq --pdata samples.csv --mode WGBS --output my_project/userspace
系统: 扫描 FASTQ/pdata → 生成默认配置 → 保存配置文件
```

### 用例 3: 工作流执行
```
用户: xdxtools run --config config.yaml --engine slurm
系统: 验证配置 → 执行 Snakemake → 输出日志与状态
```

## 验收标准

- [ ] 支持所有 4 种工作流模式
- [ ] 配置文件支持 3 种格式
- [ ] 执行引擎支持 2 种环境（Slurm/Local）
- [ ] 单二进制 < 20MB
- [ ] 文档完整且示例丰富
