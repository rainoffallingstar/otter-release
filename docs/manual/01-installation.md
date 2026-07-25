# 第一章：安装指南

## 支持状态

主仓统一为 `rainoffallingstar/otter`。当前运行时仍需 Snakemake；`craftmake` 是正在接入的 Go 替代执行层，暂时与 Snakemake 双轨。

## 安装前要求

- Linux 或 macOS；SLURM 集群以 Linux 为主
- Git、Go 1.24+
- 构建 Rust 组件时需要 Rust 工具链
- `methx` 构建/运行需要 HDF5
- 能访问 GitHub 和配置的软件包频道

## Release 安装

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/rainoffallingstar/otter/main/scripts/install.sh)
```

安装脚本和 release 资产也处于命名迁移范围。如果当前 release 仍包含旧二进制名或旧环境前缀，应作为兼容资产记录，不要把它误写成已完成迁移。

## 源码安装

```bash
git clone --recurse-submodules https://github.com/rainoffallingstar/otter.git
cd otter
conda activate go-env
go build -o otter .
install -m 755 otter "$HOME/.cargo/bin/otter"
```

当前源码根命令可能仍显示 `xdxtools`。输出文件名改为 `otter` 不等于代码迁移完成，打包前必须运行 `./otter --help` 核对。

## 子模块

```text
craftmake enva fastqcx xenofilx pairbam seq2mat matsrun qctb methx bamdriver
```

若目录为空：

```bash
git submodule update --init --recursive
```

## 环境

```bash
enva create --core
enva create --snakemake
enva create --extra
enva list --detailed
enva validate --all
```

期望环境名：

| 环境 | 用途 |
|---|---|
| `otter-core` | 核心生信工具与算子 |
| `otter-snakemake` | 当前 Snakemake 兼容路径 |
| `otter-extra` | 附加分析和可视化工具 |

## 验证

```bash
command -v otter craftmake enva
command -v fastqcx xenofilx pairbam seq2mat matsrun qctb methx bamdriver

otter --help
craftmake --help
enva --version
```

`craftmake --help` 成功不代表 `otter run` 已切换到 craftmake。

## HDF5

```bash
export HDF5_DIR="$CONDA_PREFIX"
export HDF5_INCLUDE_DIR="$HDF5_DIR/include"
export HDF5_LIB_DIR="$HDF5_DIR/lib"
export PKG_CONFIG_PATH="$HDF5_DIR/lib/pkgconfig:$PKG_CONFIG_PATH"
export LD_LIBRARY_PATH="$HDF5_DIR/lib:$LD_LIBRARY_PATH"
```

下一章：[数据准备](02-data-preparation.md)
