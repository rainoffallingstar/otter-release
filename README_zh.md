# xdxtools-go

xdxtools 的 Go 语言重写版本，用于 RRBS、WGBS、RNA-seq 和 PDX 分析的生物信息学工作流管理工具。

## 功能特性

- **多种工作流模式**：支持 RRBS、WGBS、RNA-seq 和 PDX 分析
- **每样本适配器生成**：支持 barcode 的自动适配器生成和反向互补计算
- **智能输入处理**：自动 FASTQ 文件配对和 pdata 验证
- **Snakemake 集成**：与现有 Snakemake 工作流完全兼容
- **中文列名支持**：中文列名自动映射为英文列名
- **分组级别计算**：从 pdata 自动计算唯一分组数量
- **多种执行引擎**：支持 Slurm 和本地执行（架构简化中移除了 Docker 支持）
- **Excel 文件支持**：原生支持 .xlsx/.xls 格式的 pdata 文件（无需转换）
- **TUI 界面**：交互式终端用户界面，用于工作流管理
- **动态参考配置**：根据物种和模式自动生成参考基因组路径
- **全面测试**：175+ 个测试用例，100% 通过率

## 安装

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/xdxtools/xdxtools-go.git
cd xdxtools-go

# 构建二进制文件
go build -o xdxtools

# 安装到 PATH（可选）
sudo mv xdxtools /usr/local/bin/
```

### 系统要求

- Go 1.21+
- Snakemake
- Conda/Miniconda（用于工作流执行）
- R（可选，用于 R 脚本）
- Python 3.8+（可选，用于 Python 脚本）

## 快速上手 - 三命令工作流

xdxtools 采用简单的三命令工作流：

### 1️⃣ `init` - 安装 Snakemake 工作流文件

将 Snakemake 工作流文件、R 脚本和规则复制到项目目录。

```bash
# 初始化新项目目录
xdxtools init my_project

# 指定引擎类型初始化（root/rootless）
xdxtools init my_project --engine-type rootless
```

这会复制：
- 22 个 Snakemake 工作流文件
- 15+ 个 R/Python 脚本
- 34 个 Snakemake 规则
- 5 个工作流配置模板

### 2️⃣ `create` - 创建分析项目

扫描 FASTQ 文件，验证样本，创建目录结构，生成 config.yaml。

```bash
# 基本用法 - 自动生成 jobid
xdxtools create --fastq /data/fastq --mode RRBS

# 完整示例
xdxtools create \
    --fastq /data/fastq \
    --pdata /data/pdata/samples.csv \
    --mode RRBS \
    --species1 human \
    --species2 mouse \
    --output userspace \
    --jobid my_project_2024
```

**功能说明：**
1. 扫描 FASTQ 目录查找配对样本（R1/R2）
2. 验证样本配对
3. 加载并验证 pdata（如果提供）
4. 生成唯一 job ID（或使用用户指定的）
5. 在 `userspace/{jobid}/` 下创建完整目录结构
6. 生成包含每样本适配器的 `config.yaml`
7. 从 pdata 计算分组级别

### 3️⃣ `run` - 执行工作流

使用生成的配置执行 Snakemake 工作流。

```bash
# 执行工作流
xdxtools run --config userspace/my_project/config/config.yaml

# 使用特定引擎执行
xdxtools run --config userspace/my_project/config/config.yaml --engine slurm

# 试运行（测试配置）
xdxtools run --config userspace/my_project/config/config.yaml --dry-run
```

## 命令参考

### 命令

#### init

将 Snakemake 工作流文件安装到项目目录。

```bash
xdxtools init <project-directory> [flags]
```

**参数：**
- `--engine-type`：引擎类型（root/rootless，默认：rootless）

**示例：**

```bash
# 初始化当前目录
xdxtools init .

# 初始化新目录
xdxtools init my_project

# 使用 rootless 规则初始化（非容器）
xdxtools init my_project --engine-type rootless
```

#### create

创建具有验证样本和生成配置的新的分析项目。

```bash
xdxtools create --fastq <path> [flags]
```

**必需参数：**
- `--fastq, -f`：FASTQ 文件目录

**可选参数：**
- `--pdata, -p`：表型数据文件（Excel/CSV）
- `--mode, -m`：工作流模式（RRBS/WGBS/RNASEQ，默认：RRBS）
- `--species1`：主要物种（默认：human）
- `--species2`：次要物种（指定时启用 PDX 模式）
- `--output, -o`：项目输出目录（默认：userspace）
- `--jobid`：自定义 job ID（默认：自动生成的 40 字符十六进制）
- `--suffix1`：R1 文件后缀（默认：_R1.fastq.gz）
- `--suffix2`：R2 文件后缀（为空时自动推导）

**示例：**

```bash
# 基本 RRBS 项目
xdxtools create --fastq /data/fastq --mode RRBS

# 带 pdata 的 WGBS
xdxtools create --fastq /data/fastq --pdata /data/pdata.csv --mode WGBS

# PDX 模式（指定 species2 时自动启用）
xdxtools create \
    --fastq /data/fastq \
    --mode RRBS \
    --species1 human \
    --species2 mouse

# 自定义输出和 jobid
xdxtools create \
    --fastq /data/fastq \
    --output /custom/path \
    --jobid experiment_001
```

#### run

使用配置文件执行 Snakemake 工作流。

```bash
xdxtools run --config <config-file> [flags]
```

**必需参数：**
- `--config, -c`：配置文件路径

**可选参数：**
- `--engine`：执行引擎（auto/slurm/local，默认：auto）
- `--dry-run`：试运行，不执行
- `--verbose, -v`：详细输出

**示例：**

```bash
# 使用自动检测的引擎运行（自动检测 Slurm 或 Local）
xdxtools run --config config/config.yaml

# 在 Slurm 集群上运行
xdxtools run --config config/config.yaml --engine slurm

# 本地运行
xdxtools run --config config/config.yaml --engine local

# 测试配置
xdxtools run --config config/config.yaml --dry-run
```

**注意**：架构简化中移除了 Docker 引擎支持，现在仅支持 Slurm 和 Local 引擎。

#### config

验证和管理配置文件。

```bash
xdxtools config [command]
```

**命令：**
- `validate`：验证配置文件而不运行工作流

**示例：**

```bash
# 验证配置文件
xdxtools config validate --config userspace/my_project/config/config.yaml
```

#### tui

启动 xdxtools 的交互式终端用户界面（TUI）。

```bash
xdxtools tui [flags]
```

**可选参数：**
- `--config, -c`：启动时可选加载的配置文件

**示例：**

```bash
# 启动带交互菜单的 TUI
xdxtools tui

# 启动带预加载配置的 TUI
xdxtools tui --config userspace/my_project/config/config.yaml
```

TUI 提供：
- 创建新项目的交互菜单
- 工作流仪表板和状态监控
- 配置管理界面
- 帮助和文档

## 工作流模式

### RRBS（限制性酶切甲基化测序）
```bash
xdxtools create --fastq /data/fastq --mode RRBS
```

### WGBS（全基因组甲基化测序）
```bash
xdxtools create --fastq /data/fastq --mode WGBS
```

### RNA-seq（转录组测序）
```bash
xdxtools create --fastq /data/fastq --mode RNASEQ
```

### PDX（人源肿瘤异种移植）
```bash
xdxtools create \
    --fastq /data/fastq \
    --mode RRBS \
    --species1 human \
    --species2 mouse
```

当同时指定 `species1` 和 `species2` 时自动启用 PDX 模式。工作流切换到 "BeaverPDX" 配置并创建物种特异性子目录。

## 配置说明

### 生成的 Config.yaml

`create` 命令生成包含以下关键特性的完整 `config.yaml` 文件：

#### 每样本适配器
```yaml
# 自动生成的每样本适配器
trimSeq1:
  - "NO_ADAPTER_CAL_USE_DEFAULT"  # 无 barcode 的样本
  - "TGACGATAGATCGGAAGAGC"        # 有 barcode 的样本（RRBS 模式）
trimSeq2:
  - "NO_ADAPTER_CAL_USE_DEFAULT"
  - "ACGATAGATCGGAAGAGC"

# 样本标识符
SIDs:
  - "sample1"
  - "sample2"
```

#### 分组级别
从 pdata 自动计算：
```yaml
group_levels: 2  # 唯一分组数量（control, treatment）
```

#### 用户和作业信息
```yaml
userid: "auto_generated_or_custom"
jobid: "a1b2c3d4e5f6..."
```

#### 动态参考配置
根据物种和模式自动生成：
```yaml
# 单物种 RRBS
reference:
  genome: human
  genome_fasta: [inst/pdx/homo_sapiens/human.fasta]
  genome_index: [inst/pdx/homo_sapiens/]

# PDX RRBS
reference:
  genome: human
  genome_fasta:
    - inst/pdx/homo_sapiens/human.fasta
    - inst/pdx/mouse/mouse.fasta
  genome_index:
    - inst/pdx/homo_sapiens/
    - inst/pdx/mouse/

# 单物种 RNA-seq
reference:
  genome: human
  genome_fasta: [inst/pdx/homo_sapiens/human.fasta]
  rnaseq_gtf: inst/rnaseq/homo_sapiens/human.ensGene_sorted.gtf
  rnaseq_ref: inst/rnaseq/homo_sapiens/

# PDX RNA-seq（多物种数组格式）
reference:
  genome: human
  genome_fasta:
    - inst/pdx/homo_sapiens/human.fasta
    - inst/pdx/mouse/mouse.fasta
  rnaseq_gtf:
    - inst/rnaseq/homo_sapiens/human.ensGene_sorted.gtf
    - inst/rnaseq/mouse/mouse.ensGene_sorted.gtf
  rnaseq_ref:
    - inst/rnaseq/homo_sapiens/
    - inst/rnaseq/mouse/
```

### FASTQ 文件命名

工具支持多种 FASTQ 文件命名约定：

- 标准：`Sample_R1.fastq.gz` + `Sample_R2.fastq.gz`
- 替代：`Sample_1.fastq.gz` + `Sample_2.fastq.gz`
- 支持自定义后缀

**示例：**
```bash
# 检测到的文件：
Sample1_R1.fastq.gz
Sample1_R2.fastq.gz
Sample2_R1.fastq.gz
Sample2_R2.fastq.gz

# 命令
xdxtools create --fastq /data/fastq --mode RRBS

# 结果：检测到 2 个配对样本
```

### PData 格式

#### CSV 格式（推荐）
```bash
# 保存为 CSV（Excel 文件暂不支持）
sampleid,inline_barcode_sequence,condition
sample1,ATCG,control
sample2,,treatment
sample3,GCTA,treatment
```

#### 中文列名
自动映射为英文：
- "样本编号" / "样本ID" → `sampleid`
- "条件" → `condition`
- "样本分组" / "分组" → `sample_group`

#### Barcode 支持
- Barcode 列：`inline_barcode_sequence` 或 `barcode`
- 空 barcode → `NO_ADAPTER_CAL_USE_DEFAULT`（自动检测适配器）
- 有 barcode → 生成带反向互补的适配器

### 每样本适配器生成

工具自动为每个样本生成适配器：

#### RRBS 模式示例
```bash
# 输入 pdata
sample1: barcode = ATCG
sample2: barcode = (空)

# 生成的适配器（RRBS 模式）
trimSeq1:
  - "TGACGATAGATCGGAAGAGC"  # TGA + CGAT(ATCG的反向互补) + AGATCGGAAGAGC
  - "NO_ADAPTER_CAL_USE_DEFAULT"

trimSeq2:
  - "ACGATAGATCGGAAGAGC"    # A + CGAT(ATCG的反向互补) + AGATCGGAAGAGC
  - "NO_ADAPTER_CAL_USE_DEFAULT"
```

#### WGBS 模式示例
```bash
# 相同输入
# 生成的适配器（WGBS 模式 - 无 TGA/A 前缀）
trimSeq1:
  - "CGATAGATCGGAAGAGC"     # CGAT(ATCG的反向互补) + AGATCGGAAGAGC
  - "NO_ADAPTER_CAL_USE_DEFAULT"
```

### 项目目录结构

运行 `xdxtools create` 后创建以下结构：

```
userspace/{jobid}/
├── config/
│   └── config.yaml          # 生成的配置
├── data/                    # FASTQ 文件（软链接）
├── log/                     # 日志文件
├── analysis/                # 分析结果
│   ├── betaM
│   ├── clubcpg/
│   │   ├── coverage_before/
│   │   ├── coverage_impute/
│   │   └── model/
│   ├── DMR
│   ├── GCbias
│   ├── logsummary
│   ├── methrixh5
│   ├── qc_summary
│   ├── RData
│   └── uxm_summary
└── workflow/                # 工作流输出
    ├── QC
    ├── fastqc_raw
    ├── fastqc_clean
    ├── trim
    ├── bsmap/
    │   ├── tmp/
    │   │   ├── {species1}/  # 仅 PDX 模式
    │   │   └── {species2}/  # 仅 PDX 模式
    │   ├── {species1}/      # 仅 PDX 模式
    │   └── {species2}/      # 仅 PDX 模式
    ├── mCall
    ├── mhap
    ├── qualimap
    └── umx
```

## 完整示例

### 示例 1：基本 RRBS 分析

```bash
# 步骤 1：初始化项目
xdxtools init rrbs_project

# 步骤 2：创建分析项目
xdxtools create \
    --fastq /data/fastq \
    --mode RRBS \
    --output userspace \
    --jobid rrbs_experiment_001

# 步骤 3：运行工作流
xdxtools run --config userspace/rrbs_experiment_001/config/config.yaml
```

### 示例 2：带 PData 的 WGBS

```bash
# 准备 pdata.csv
cat > /data/pdata.csv << EOF
sampleid,inline_barcode_sequence,condition
sample1,ATCG,control
sample2,GCTA,treatment
sample3,TAGC,control
sample4,CGAT,treatment
EOF

# 创建项目
xdxtools create \
    --fastq /data/fastq \
    --pdata /data/pdata.csv \
    --mode WGBS \
    --output userspace \
    --jobid wgbs_analysis

# 运行
xdxtools run --config userspace/wgbs_analysis/config/config.yaml
```

### 示例 3：PDX 模式（人 + 鼠）

```bash
# 创建 PDX 项目
xdxtools create \
    --fastq /data/fastq \
    --mode RRBS \
    --species1 human \
    --species2 mouse \
    --output userspace \
    --jobid pdx_human_mouse

# 在 Slurm 上运行
xdxtools run \
    --config userspace/pdx_human_mouse/config/config.yaml \
    --engine slurm
```

### 示例 4：RNA-seq 分析

```bash
# 创建 RNA-seq 项目
xdxtools create \
    --fastq /data/fastq \
    --mode RNASEQ \
    --output userspace \
    --jobid rnaseq_analysis

# 先试运行
xdxtools run \
    --config userspace/rnaseq_analysis/config/config.yaml \
    --dry-run
```

### 示例 5：自定义 Job ID 和输出路径

```bash
# 使用自定义 jobid 和输出目录
xdxtools create \
    --fastq /data/fastq \
    --pdata /data/pdata.csv \
    --mode RRBS \
    --output /custom/output/path \
    --jobid experiment_20240109_001

# 验证项目结构
ls -la /custom/output/path/experiment_20240109_001/
```

### 示例 6：中文 PData

```bash
# 准备中文 pdata.csv
cat > /data/pdata_chinese.csv << EOF
样本编号,inline_barcode_sequence,条件
样本1,ATCG,对照组
样本2,GCTA,治疗组
样本3,TAGC,对照组
样本4,CGAT,治疗组
EOF

# 创建项目（自动映射中文列）
xdxtools create \
    --fastq /data/fastq \
    --pdata /data/pdata_chinese.csv \
    --mode RRBS \
    --output userspace \
    --jobid chinese_pdata_test

# 验证配置中的分组级别
cat userspace/chinese_pdata_test/config/config.yaml | grep group_levels
# 输出：group_levels: 2
```

### 示例 7：Barcode 处理

```bash
# 准备带 barcodes 的 pdata
cat > /data/pdata_barcodes.csv << EOF
sampleid,inline_barcode_sequence,condition
Tumor1,ATCG,tumor
Tumor2,GCTA,tumor
Normal1,,normal
Normal2,,normal
EOF

# 创建项目
xdxtools create \
    --fastq /data/fastq \
    --pdata /data/pdata_barcodes.csv \
    --mode RRBS \
    --output userspace \
    --jobid barcode_test

# 检查生成的适配器
cat userspace/barcode_test/config/config.yaml | grep -A 10 "trimSeq1"
```

预期输出显示：
- Tumor1 有 ATCG barcode：适配器 = "TGACGATAGATCGGAAGAGC"
- Normal1 无 barcode：适配器 = "NO_ADAPTER_CAL_USE_DEFAULT"

## 故障排除

### 常见问题

#### 1. 未找到 FASTQ 文件
```bash
# 错误：未找到 FASTQ 文件

# 解决方案：检查文件命名
ls /data/fastq/

# 应显示类似文件：
# Sample1_R1.fastq.gz
# Sample1_R2.fastq.gz
# Sample2_R1.fastq.gz
# Sample2_R2.fastq.gz
```

#### 2. 无有效配对样本
```bash
# 错误：未找到有效配对样本

# 解决方案：检查后缀设置
xdxtools create \
    --fastq /data/fastq \
    --mode RRBS \
    --suffix1 "_1.fastq.gz" \
    --suffix2 "_2.fastq.gz"
```

#### 3. pdata 中未找到样本
```bash
# 错误：pdata 中未找到样本 'Sample1'

# 解决方案：确保样本名称匹配
# FASTQ: Sample1_R1.fastq.gz
# pdata.csv: Sample1,...
```

#### 4. Excel 文件编码问题
```bash
# 错误：Excel 文件编码或格式问题

# 解决方案：Excel 文件（.xlsx/.xls）现在已原生支持
# 直接指定您的 Excel 文件：
xdxtools create --fastq /data/fastq --pdata /data/pdata.xlsx --mode RRBS

# 如果遇到编码问题，请转换为 CSV
# 从 Excel：另存为 → CSV (UTF-8)
# 或使用命令行：
iconv -f GB2312 -t UTF-8 pdata.xlsx > pdata.csv
```

### 调试模式

启用详细日志：
```bash
xdxtools create \
    --fastq /data/fastq \
    --mode RRBS \
    --verbose

xdxtools run \
    --config config.yaml \
    --verbose
```

### 验证

无需运行即可验证配置：
```bash
xdxtools run --config config.yaml --dry-run
```

检查 config.yaml 语法：
```bash
# 安装 yq（如果需要）：pip install yq
yq eval config.yaml

# 或使用 Python
python -c "import yaml; yaml.safe_load(open('config.yaml'))"
```

## 高级用法

### 自定义文件后缀

```bash
# 用于非标准命名
xdxtools create \
    --fastq /data/fastq \
    --mode RRBS \
    --suffix1 "_R1.fastq" \
    --suffix2 "_R2.fastq"
```

### 批量处理

```bash
# 处理多个项目
for mode in RRBS WGBS RNASEQ; do
    xdxtools create \
        --fastq /data/fastq \
        --mode $mode \
        --output userspace \
        --jobid batch_${mode}
done
```

### 引擎选择

```bash
# 自动检测（推荐 - 检测 Slurm 或 Local）
xdxtools run --config config.yaml

# 强制使用特定引擎
xdxtools run --config config.yaml --engine local
xdxtools run --config config.yaml --engine slurm
```

**注意**：已移除 Docker 引擎支持，仅支持 Slurm 和 Local 引擎。

## 常见问题

**Q: 支持哪些 FASTQ 文件格式？**
A: .fastq, .fastq.gz, .fq, .fq.gz

**Q: 是否支持 Excel pdata 文件？**
A: 是的！Excel 文件（.xlsx/.xls）已完全支持。只需直接指定 Excel 文件：
   `xdxtools create --fastq /data/fastq --pdata /data/pdata.xlsx --mode RRBS`

**Q: 可以使用自定义适配器吗？**
A: 可以。编辑生成的 config.yaml 并修改 trimSeq1/trimSeq2 数组。

**Q: group_levels 如何计算？**
A: 计算 sample_group 或 condition 列中的唯一值（优先级：sample_group > condition）。

**Q: job ID 格式是什么？**
A: 40 字符十六进制字符串（例如：a1b2c3d4e5f6...）

**Q: 如何启用 PDX 模式？**
A: 同时指定 --species1 和 --species2 参数。

**Q: 什么是 NO_ADAPTER_CAL_USE_DEFAULT？**
A: 告诉 Snakemake 自动检测适配器的特殊标记。

**Q: 工具是否支持动态参考基因组路径？**
A: 是的！工具会根据指定的物种和模式自动生成参考基因组路径。单物种返回字符串，PDX 模式返回数组以兼容 Snakemake 规则。

**Q: 支持哪些物种的参考基因组？**
A: 目前支持：human/homo_sapiens 和 mouse/mus_musculus。可以通过扩展代码中的物种映射来添加更多物种。

**Q: 支持多少样本？**
A: 无硬性限制，取决于系统资源和 Snakemake 配置。

**Q: 可以恢复失败的工作流吗？**
A: 可以，Snakemake 会自动从最后一个成功步骤恢复。

**Q: 如何更改工作线程数？**
A: 编辑生成的 config.yaml：parallel.workers: 8

## 开发

### 构建

```bash
# 为当前平台构建
go build -o xdxtools

# 为多平台构建
GOOS=linux GOARCH=amd64 go build -o xdxtools-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o xdxtools-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o xdxtools.exe
```

### 测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 运行特定测试
go test -v ./internal/input -run TestAdapterGenerator
```

### 项目结构

```
xdxtools-go/
├── cmd/                          # CLI 命令
│   ├── root.go                   # 根命令
│   ├── init.go                   # init 命令
│   ├── create.go                # create 命令
│   └── run.go                    # run 命令
│
├── internal/                     # 内部包
│   ├── assets/                  # 资源管理
│   │   └── assets.go           # 资源复制器
│   ├── config/                   # 配置管理
│   │   ├── config.go            # 配置类型
│   │   ├── loader.go            # 配置加载器
│   │   ├── validator.go         # 配置验证器
│   │   ├── generator.go         # Snakemake YAML 生成器
│   │   └── defaults.go          # 默认配置
│   │
│   ├── engine/                   # 执行引擎
│   │   ├── engine.go            # 引擎接口
│   │   ├── slurm.go             # Slurm 引擎
│   │   ├── local.go             # 本地引擎
│   │   └── factory.go           # 引擎工厂
│   │
│   ├── workflow/                 # 工作流管理
│   │   ├── types.go             # 工作流类型
│   │   ├── snakemake.go         # Snakemake 执行器
│   │   └── manager.go            # 工作流管理器
│   │
│   ├── input/                    # 输入处理
│   │   ├── types.go              # 输入类型
│   │   ├── fastq.go             # FASTQ 扫描器
│   │   ├── pdata.go             # pdata 解析器
│   │   ├── adapter.go           # 适配器生成器
│   │   ├── adapter_test.go      # 适配器测试
│   │   └── validator.go         # 输入验证器
│   │
│   ├── script/                   # 脚本执行
│   │   └── executor.go           # 脚本执行器
│   │
│   └── logger/                   # 日志
│       └── logger.go            # 日志记录器
│
├── pkg/                          # 公共包
│   └── utils/                    # 工具函数
│
├── docs/                         # 文档
│   └── active_context.md        # 系统上下文
│
├── embed.go                      # 嵌入资源
├── go.mod
└── main.go                       # 入口点
```

## 许可证

本项目采用 MIT 许可证。

## 贡献

欢迎贡献！请随时提交 Pull Request。

## 致谢

- 原始 xdxtools R 包
- Snakemake 工作流引擎
- 生物信息学社区

## 支持

如有问题、疑问或贡献，请访问我们的 GitHub 仓库。
