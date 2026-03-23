# xdxtools

用于 RRBS、WGBS、RNA-seq 和 PDX 分析的生物信息学工作流 CLI 工具。

## 功能特性

- 工作流模式：RRBS / WGBS / RNA-seq / PDX
- 自动 FASTQ 配对 + 每样本适配器生成（支持 barcode）
- SLURM Job Array + 本地工作池并行化
- Excel/CSV pdata，支持中文列名自动映射
- Snakemake 集成，内嵌工作流文件
- Conda 环境自动回退（支持 [enva](https://github.com/rainoffallingstar/enva)）
- 主 CLI 为单二进制；工作流运行仍依赖 Snakemake 与 conda/enva

## 系统要求

- Go 1.24+
- Snakemake
- conda / mamba / micromamba（或 [enva](https://github.com/rainoffallingstar/enva)）

## 安装

### 快速安装（交互式）

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh)
```

脚本将从 GitHub Releases 下载预编译二进制文件，并以交互方式完成 conda 环境配置。安装开始时会先让你选择中文或英文，也可以用 `--lang en` 或 `--lang zh` 强制指定界面语言。`scripts/setup.sh` 仅用于本地源码构建。

如果仓库本身是私有的，匿名访问 `raw.githubusercontent.com` 会返回 `404`，而 `wget -qO-` 会把这个错误静默吞掉。此时应改用带认证头的启动命令。如果 release 仓库或 release 资产是私有的，请先导出 `GITHUB_TOKEN`（或 `GH_TOKEN` / `GITHUB_PAT`）；如果是私有 fork，还需要设置 `GITHUB_RELEASES_REPO=<owner>/<repo>`。在交互模式下，如果 GitHub 访问失败且当前没有配置 token，安装器可以在终端里提示你做隐藏输入，并自动重试一次。

常用选项：

```bash
# 非交互模式，全部使用默认值
bash <(curl -fsSL https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh) --non-interactive

# 指定发布版本
bash <(curl -fsSL https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh) --version v0.3.0

# 私有仓库启动
export GITHUB_PAT=<your_pat>
bash <(curl -fsSL -H "Authorization: Bearer ${GITHUB_PAT}" \
  https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh)

# 私有 release fork
export GITHUB_PAT=<your_pat>
export GITHUB_RELEASES_REPO=<owner>/<repo>
bash <(curl -fsSL -H "Authorization: Bearer ${GITHUB_PAT}" \
  https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh)

# 跳过 conda 环境创建
bash <(curl -fsSL https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh) --skip-envs
```

### 从源码构建

```bash
git clone --recurse-submodules https://github.com/rainoffallingstar/xdxtools-go.git
cd xdxtools-go
go build -o xdxtools
```

## 快速上手

```bash
# 1. 初始化项目（复制 Snakemake 工作流文件）
xdxtools init my_project

# 2. 扫描 FASTQ，验证样本，生成配置文件
xdxtools create --fastq /data/fastq --mode RRBS --pdata samples.csv --output my_project/userspace --jobid demo_rrbs

# 3. 执行工作流
xdxtools run --config my_project/userspace/demo_rrbs/config/config.yaml
```

## 命令速查

| 命令 | 说明 |
|------|------|
| `init`   | 将 Snakemake 工作流文件复制到项目目录 |
| `create` | 扫描 FASTQ，验证样本，生成 config.yaml |
| `run`    | 执行 Snakemake 工作流 |
| `status` | 显示工作流进度 |
| `config` | 验证配置文件 |

### `run` 常用参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--engine` | `auto` | 执行引擎：`slurm` / `local` / `auto` |
| `--slurm-partition` | empty | 可选的 SLURM 分区覆盖 |
| `--parallel-jobs` | `2` | 最大并发作业数 |
| `--dry-run` | `false` | 试运行（不实际执行） |
| `--resume` / `-r` | `false` | 从上次完成的步骤恢复 |

使用 `--verbose` 查看详细输出，或使用 `--dry-run` 在不执行的情况下排查问题。

## 工作流模式

| 模式 | 参数 |
|------|------|
| RRBS（限制性酶切甲基化测序） | `--mode RRBS` |
| WGBS（全基因组甲基化测序） | `--mode WGBS` |
| RNA-seq（转录组测序） | `--mode RNASEQ` |
| PDX（人源肿瘤异种移植） | `--mode RRBS --species1 human --species2 mouse` |

同时指定 `--species1` 和 `--species2` 时，PDX 模式自动启用。

## 许可证

MIT
