# xdxtools

用于 RRBS、WGBS、RNA-seq 和 PDX 分析的生物信息学工作流 CLI 工具。

## 功能特性

- 工作流模式：RRBS / WGBS / RNA-seq / PDX
- 自动 FASTQ 配对 + 每样本适配器生成（支持 barcode）
- SLURM Job Array + 本地工作池并行化
- Excel/CSV pdata，支持中文列名自动映射
- Snakemake 集成，内嵌工作流文件
- 通过 [enva](https://github.com/rainoffallingstar/enva) 进行 rattler 优先的环境管理，并兼容历史 conda 环境
- 主 CLI 为单二进制；工作流运行默认依赖 Snakemake 与 enva（历史 conda 兼容环境也可继续使用）

## 系统要求

- Go 1.24+
- Snakemake
- [enva](https://github.com/rainoffallingstar/enva)（推荐）
- `conda` / `mamba` / `micromamba` 仅用于历史兼容或被接管的环境

## 安装

### 快速安装（交互式）

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh)
```

脚本将从 GitHub Releases 下载预编译二进制文件，并以交互方式完成 conda 环境配置。现在默认会优先从公开镜像仓库 `rainoffallingstar/flightlight` 拉取二进制资产；如果目标版本尚未镜像，则自动回退到 `rainoffallingstar/xdxtools-go`；环境 YAML 也会从同一个已选中的 release 仓库下载。安装开始时会先让你选择中文或英文，也可以用 `--lang en` 或 `--lang zh` 强制指定界面语言。`scripts/setup.sh` 仅用于本地源码构建。当检测到已存在的二进制文件时，安装器只会统一询问一次是否覆盖，后续所有二进制都沿用这次选择。

如果仓库本身是私有的，匿名访问 `raw.githubusercontent.com` 会返回 `404`，而 `wget -qO-` 会把这个错误静默吞掉。此时应改用带认证头的启动命令。如果 release 仓库或 release 资产是私有的，请先导出 `GITHUB_TOKEN`（或 `GH_TOKEN` / `GITHUB_PAT`）；如果是私有或自定义仓库布局，还可以设置 `GITHUB_RELEASES_REPO=<owner>/<repo>` 与 `GITHUB_FALLBACK_RELEASES_REPO=<owner>/<repo>`。在交互模式下，如果 GitHub 访问失败且当前没有配置 token，安装器可以在终端里提示你做隐藏输入，并自动重试一次。

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

## Codex 技能

本仓库自带一个可安装的 Codex skill，目录为 `skills/xdxtools`。该 skill 面向根仓库 `xdxtools` 以及与之配套的子仓库：`enva`、`Paireads`、`bamdriver-go`、`fastqc-rs`、`gomats`、`htseq2matrix-go`、`methrix-cli`、`qctb` 和 `xenofilter-go`。

可在 Codex 环境中使用内置的 skill installer 从 GitHub 安装：

```bash
python ~/.codex/skills/.system/skill-installer/scripts/install-skill-from-github.py --repo rainoffallingstar/xdxtools-go --path skills/xdxtools
```

如果你使用的是 fork 或非默认分支，请替换 `--repo`，并按需附加 `--ref <branch-or-tag>`。

安装完成后，重启 Codex，让新 skill 生效。

随后可以在 Codex 提示词中显式调用，例如：

```text
Use $xdxtools to inspect the root workflow CLI and update the install docs.
Use $xdxtools to work on qctb without breaking xdxtools submodule boundaries.
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
