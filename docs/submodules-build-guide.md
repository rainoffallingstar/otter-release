# xdxtools 子仓库编译指南

本文档记录了 xdxtools 项目所有子仓库的编译过程、依赖要求和安装方法。

## 目录

- [概述](#概述)
- [编译环境准备](#编译环境准备)
- [子仓库列表](#子仓库列表)
- [编译过程](#编译过程)
- [环境变量配置](#环境变量配置)
- [验证安装](#验证安装)
- [故障排除](#故障排除)

---

## 概述

xdxtools 项目包含 7 个子仓库（git submodules），用于不同的生物信息学分析任务：

| 子仓库 | 语言 | 功能 | 二进制文件 |
|--------|------|------|-----------|
| enva | Rust | micromamba 环境管理器 | enva |
| rv | Rust | R 包管理器 | rv |
| htseq2matrix-go | Go | HTSeq 表达矩阵转换 | htseq2matrix |
| methrix-cli | Rust | 甲基化分析 CLI | methrix-cli |
| xenofilter-go | Go | xenofilter 污染过滤 | xenofilter |
| Paireads | Go | 配对 reads 处理 | paireads |
| qctb | Rust | 质量控制工具箱 | qctb |

---

## 编译环境准备

### 1. Conda 环境创建

项目使用两个 conda 环境分别编译 Rust 和 Go 项目：

#### Rust 编译环境

```bash
# 创建 rust_build 环境（包含 Rust 工具链和 HDF5 库）
conda create -y -n rust_build rust

# 激活环境
conda activate rust_build
```

**注意**: `rust_build` 环境已包含 HDF5 库，用于编译 methrix-cli。

#### Go 编译环境

```bash
# 创建 go-build 环境
conda create -y -n go-build go

# 激活环境
conda activate go-build
```

### 2. 环境变量配置

将 HDF5 库路径添加到 `~/.bashrc`，使 methrix-cli 可以找到 HDF5 库：

```bash
# 添加到 ~/.bashrc
cat >> ~/.bashrc << 'EOF'

# HDF5 library paths (from rust_build conda environment)
export HDF5_DIR="/public3/home/scg9946/TTest/soft/MyMiniconda/envs/rust_build"
export HDF5_INCLUDE_DIR="$HDF5_DIR/include"
export HDF5_LIB_DIR="$HDF5_DIR/lib"
export LD_LIBRARY_PATH="$HDF5_DIR/lib:$LD_LIBRARY_PATH"
export PKG_CONFIG_PATH="$HDF5_DIR/lib/pkgconfig:$PKG_CONFIG_PATH"
EOF

# 重新加载配置
source ~/.bashrc
```

---

## 子仓库列表

### Rust 项目（5 个）

| 项目 | 版本 | 编译时间 | 特殊依赖 |
|------|------|----------|----------|
| enva | 0.1.0 | ~2 分钟 | 无 |
| rv | 0.17.1 | ~2 分钟 | 需要 `--features cli` |
| methrix-cli | 0.1.0 | ~2.5 分钟 | HDF5 库 |
| qctb | 0.1.0 | ~1 分钟 | 无 |

### Go 项目（2 个）

| 项目 | 编译时间 | 特殊依赖 | 编译选项 |
|------|----------|----------|----------|
| xenofilter-go | ~30 秒 | 无 | `CGO_ENABLED=0` |
| Paireads | ~30 秒 | 无 | `CGO_ENABLED=0` |
| htseq2matrix-go | ~30 秒 | 完整源码 | `CGO_ENABLED=0` |

---

## 编译过程

### 方法 1: 使用统一编译脚本

项目提供了 `scripts/build-all-submodules.sh` 脚本，可以一键编译所有子模块：

```bash
cd /public3/home/scg9946/xdxtools
bash scripts/build-all-submodules.sh
```

**脚本功能**:
- 自动检测 Go 和 Rust 编译环境
- 使用 conda 环境编译各个项目
- 将二进制文件安装到 `$HOME/.cargo/bin`

### 方法 2: 手动编译各个子模块

#### Rust 项目编译

```bash
# 设置环境变量
source ~/.bashrc

# enva
cd enva
conda run -n rust_build cargo build --release
cp target/release/enva $HOME/.cargo/bin/

# rv (需要 cli feature)
cd ../rv
conda run -n rust_build cargo build --release --features cli
cp target/release/rv $HOME/.cargo/bin/

# methrix-cli (需要 HDF5 环境变量)
cd ../methrix-cli-local
conda run -n rust_build cargo build --release
cp target/release/methrix $HOME/.cargo/bin/methrix-cli

# qctb
cd ../qctb
conda run -n rust_build cargo build --release
cp target/release/qctb $HOME/.cargo/bin/
```

#### Go 项目编译

```bash
# 设置 Go 代理（可选，加速依赖下载）
export GOPROXY=https://goproxy.cn,direct
export CGO_ENABLED=0

# xenofilter-go
cd xenofilter-go
conda run -n go-build go build -o $HOME/.cargo/bin/xenofilter ./cmd/xenofilter

# Paireads
cd ../Paireads
conda run -n go-build go build -o $HOME/.cargo/bin/paireads ./cmd/paireads

# htseq2matrix-go
cd ../htseq2matrix-go
conda run -n go-build go build -o $HOME/.cargo/bin/htseq2matrix ./cmd/htseq2matrix
```

---

## 环境变量配置

### HDF5 库配置

methrix-cli 需要 HDF5 库支持。以下环境变量已添加到 `~/.bashrc`：

```bash
# HDF5 library paths (from rust_build conda environment)
export HDF5_DIR="/public3/home/scg9946/TTest/soft/MyMiniconda/envs/rust_build"
export HDF5_INCLUDE_DIR="$HDF5_DIR/include"
export HDF5_LIB_DIR="$HDF5_DIR/lib"
export LD_LIBRARY_PATH="$HDF5_DIR/lib:$LD_LIBRARY_PATH"
export PKG_CONFIG_PATH="$HDF5_DIR/lib/pkgconfig:$PKG_CONFIG_PATH"
```

**验证 HDF5 配置**:

```bash
# 检查环境变量
echo $HDF5_DIR
echo $HDF5_INCLUDE_DIR
echo $HDF5_LIB_DIR

# 检查库文件
ls -la $HDF5_DIR/lib/libhdf5.so*
ls -la $HDF5_DIR/include/hdf5.h
```

### Go 编译选项

Go 项目编译时需要禁用 CGO：

```bash
export CGO_ENABLED=0
```

**原因**: conda go-build 环境缺少 C 编译器，但这些 Go 项目都是纯 Go 实现，不需要 CGO。

---

## 验证安装

### 检查所有二进制文件

```bash
# 检查文件是否存在
ls -la $HOME/.cargo/bin/ | grep -E "enva|rv|htseq2matrix|xenofilter|paireads|qctb|methrix"

# 测试命令行调用
which enva rv htseq2matrix xenofilter paireads qctb methrix-cli
```

### 版本信息验证

```bash
# enva
enva --version
# 输出: enva 0.1.0

# rv
rv --version
# 输出: rv 0.17.1

# qctb
qctb --version
# 输出: qctb 0.1.0

# htseq2matrix
htseq2matrix --help
# 输出: Usage of htseq2matrix...

# xenofilter
xenofilter --help
# 输出: Usage 信息

# paireads
paireads
# 输出: Usage: paireads <R1.bam> <R2.bam> <output_prefix>

# methrix-cli
methrix-cli --help
# 输出: Usage 信息
```

---

## 故障排除

### 1. Go 编译错误: CGO 相关

**问题**:
```
cgo: C compiler "x86_64-conda-linux-gnu-cc" not found
```

**解决方案**:
```bash
export CGO_ENABLED=0
```

### 2. Go 依赖下载超时

**问题**:
```
Get "https://proxy.golang.org/...": context deadline exceeded
```

**解决方案**:
```bash
export GOPROXY=https://goproxy.cn,direct
```

### 3. methrix-cli 编译失败: HDF5 未找到

**问题**:
```
Unable to locate HDF5 root directory and/or headers
```

**解决方案**:
```bash
# 确保 HDF5 环境变量已设置
source ~/.bashrc

# 验证 HDF5 路径
ls -la $HDF5_DIR/lib/libhdf5.so*
ls -la $HDF5_DIR/include/hdf5.h
```

### 4. rv 编译后找不到二进制文件

**问题**: rv 编译成功但 `target/release/rv` 不存在

**原因**: rv 需要使用 `--features cli` 编译

**解决方案**:
```bash
conda run -n rust_build cargo build --release --features cli
```

### 5. htseq2matrix-go 缺少 main.go

**问题**: cmd/htseq2matrix/main.go 不存在

**解决方案**: 使用完整版源码
```bash
# 从完整源码复制
cp -r $HOME/htseq2matrix-go/* /path/to/xdxtools/htseq2matrix-go/
```

---

## 性能参考

各项目在标准服务器上的编译时间（仅供参考）：

| 项目 | 编译时间 | 输出大小 |
|------|----------|----------|
| enva | ~2 分钟 | ~5.4 MB |
| rv | ~2 分钟 | ~8 MB |
| methrix-cli | ~2.5 分钟 | ~3.7 MB |
| qctb | ~1 分钟 | ~4 MB |
| xenofilter-go | ~30 秒 | ~2 MB |
| Paireads | ~30 秒 | ~2 MB |
| htseq2matrix-go | ~30 秒 | ~3 MB |

---

## 维护建议

### 定期更新子模块

```bash
# 更新所有子模块到最新版本
git submodule update --remote --merge

# 或更新特定子模块
cd <submodule>
git pull origin main
```

### 重新编译

```bash
# 清理旧的编译产物
cargo clean  # Rust 项目
go clean    # Go 项目

# 重新编译
cargo build --release
go build -o $HOME/.cargo/bin/<binary>
```

---

## 附录

### A. 完整的一键编译脚本

参见 `scripts/build-all-submodules.sh`

### B. 系统信息

- **操作系统**: Linux 3.10.0
- **Conda 版本**: MyMiniconda
- **Rust 版本**: 1.83.0 (通过 conda)
- **Go 版本**: 1.21+ (通过 conda)
- **HDF5 版本**: 1.10.x (在 rust_build 环境中)

### C. 相关文档

- [CLAUDE.md](../CLAUDE.md) - 项目开发指南
- [README.md](../README.md) - 项目概述
- [architecture.md](architecture.md) - 架构设计文档

---

**文档版本**: v1.0
**最后更新**: 2026-02-23
**维护者**: xdxtools 开发团队
