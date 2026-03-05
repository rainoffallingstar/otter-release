# 第一章：安装指南

> **本章你将学到：**
> - 安装 xdxtools 及全套子工具
> - 配置 PATH 环境变量，让终端能找到工具
> - 验证安装是否成功

---

## 📋 安装前检查

在开始安装之前，请先确认以下条件：

### 系统要求

| 条件 | 要求 |
|------|------|
| 操作系统 | Linux（推荐 CentOS 7 / Ubuntu 18.04+） |
| 架构 | x86-64（64 位） |
| 网络 | 能访问 GitHub（下载工具）和 conda 源（创建环境） |

### 检查 conda/micromamba 是否存在

```bash
# 检查 conda
conda --version
# 期望输出类似：conda 23.x.x

# 或者检查 micromamba
micromamba --version
```

如果两者都没有，请先安装 [Miniconda](https://docs.conda.io/en/latest/miniconda.html)：

```bash
# 下载并安装 Miniconda（仅限首次）
wget https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-x86_64.sh
bash Miniconda3-latest-Linux-x86_64.sh
# 按提示操作，最后选择 yes 初始化
```

---

## 🚀 一键安装（推荐方式）

### 运行安装脚本

```bash
bash <(wget -qO- https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh)
```

安装脚本会逐步询问你的选项，下面是每个问题的说明：

```
❓ Select installation directory [default: ~/.cargo/bin]:
   → 直接回车使用默认路径（推荐），或输入自定义路径

❓ Install all subtools? [Y/n]:
   → 输入 Y 或直接回车，安装所有 8 个子工具（推荐）
   → 输入 n 跳过，后续可单独安装

❓ Create conda environments? [Y/n]:
   → 输入 Y 创建分析所需的 conda 环境（首次安装推荐）
   → 已有环境可输入 n 跳过

❓ xdxtools version [default: latest]:
   → 直接回车安装最新版本
```

安装过程大约需要 **15-30 分钟**（视网络速度），请耐心等待。

### 📍 安装日志示例

```
[INFO] Downloading xdxtools v1.x.x ...
[INFO] Installing to ~/.cargo/bin/xdxtools ...
[INFO] Installing enva ...
[INFO] Installing xenofilter ...
...
[SUCCESS] All tools installed successfully!
```

---

## ⚙️ 安装选项说明

如果你熟悉命令行，可以使用以下选项跳过交互式问答：

| 选项 | 说明 | 示例 |
|------|------|------|
| `--non-interactive` | 全自动安装，使用所有默认选项 | `bash install.sh --non-interactive` |
| `--skip-envs` | 跳过 conda 环境创建 | `bash install.sh --skip-envs` |
| `--version <v>` | 安装指定版本 | `bash install.sh --version v1.2.0` |
| `--prefix <dir>` | 指定安装目录 | `bash install.sh --prefix ~/mybin` |
| `--help` | 查看所有选项 | `bash install.sh --help` |

---

## ✅ 验证安装

安装完成后，运行以下命令验证：

```bash
# 验证主工具
xdxtools --version
# 期望输出：xdxtools version 1.x.x

# 验证子工具
enva --version
xenofilter --version
paireads --version
htseq2matrix --version
methrix-cli --version
qctb --version
fqc --version
gomats --version
```

如果某个命令报 `command not found`，请看下一节 PATH 配置。

---

## 🔧 PATH 配置说明

工具默认安装到 `~/.cargo/bin/`，如果终端找不到命令，需要将该目录加入 `PATH`。

### 查看当前 PATH

```bash
echo $PATH
# 检查是否包含 ~/.cargo/bin
```

### 添加到 PATH（永久生效）

```bash
# 编辑 ~/.bashrc（bash 用户）或 ~/.zshrc（zsh 用户）
echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> ~/.bashrc

# 使配置立即生效
source ~/.bashrc
```

### 验证 PATH 配置

```bash
which xdxtools
# 期望输出：/home/你的用户名/.cargo/bin/xdxtools
```

---

## 🐍 Conda 环境说明

xdxtools 的分析流程依赖以下三个 conda 环境：

| 环境名 | 用途 | 占用空间 |
|--------|------|---------|
| `go-build` | 编译 Go 子工具 | ~200 MB |
| `rust_build` | 编译 Rust 工具（fqc） | ~500 MB |
| `methrix-cli` 相关 | methrix-cli 运行时依赖（含 HDF5） | ~1 GB |

查看已创建的环境：

```bash
conda env list
# 或
enva list
```

---

## 🧪 HDF5 配置提示

如果你需要使用 **methrix-cli**（甲基化 HDF5 分析），需要额外配置 HDF5 库路径。

将以下内容添加到 `~/.bashrc`：

```bash
# HDF5 环境变量（methrix-cli 专用）
export HDF5_DIR="$HOME/miniconda3/envs/rust_build"
export HDF5_INCLUDE_DIR="$HDF5_DIR/include"
export HDF5_LIB_DIR="$HDF5_DIR/lib"
export LD_LIBRARY_PATH="$HDF5_DIR/lib:$LD_LIBRARY_PATH"
export PKG_CONFIG_PATH="$HDF5_DIR/lib/pkgconfig:$PKG_CONFIG_PATH"
```

> ⚠️ **注意**：将 `$HOME/miniconda3` 替换为你实际的 conda 安装路径。

运行后生效：

```bash
source ~/.bashrc
methrix-cli --version
```

---

## 🆘 安装遇到问题？

- 命令找不到 → [FAQ：`xdxtools: command not found`](07-faq.md#command-not-found)
- enva 安装失败 → [FAQ：enva 安装失败](07-faq.md#enva-install-fail)
- methrix 报 HDF5 错误 → [FAQ：HDF5 错误](07-faq.md#hdf5-error)

---

**下一章：** [📁 第二章：数据准备](02-data-preparation.md)
