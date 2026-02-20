# xdxtools Build Guide

本指南记录 xdxtools 项目从源码构建的关键步骤和常见问题解决方案。

## 环境要求

| 组件 | 版本要求 | 说明 |
|------|----------|------|
| Go | 1.24+ | 主项目语言 |
| Rust | 1.85+ | enva 可用，rv 需要 1.92+ (Edition 2024) |
| Git | 2.0+ | 用于 Go 模块下载 |
| Conda/Miniconda/Micromamba | - | 运行时环境管理 |

## 构建步骤

### 1. 克隆仓库（含子模块）

```bash
git clone --recurse-submodules https://github.com/xdxtools/xdxtools-go.git
cd xdxtools-go

# 如果已克隆但没有子模块
git submodule update --init --recursive
```

### 2. 构建 xdxtools (Go)

**使用 conda 环境构建（推荐）：**

```bash
# 创建 Go 构建环境
conda create -n go-env go=1.24 git -c conda-forge -y
conda activate go-env

# 构建静态二进制
CGO_ENABLED=0 go build -ldflags="-s -w" -o xdxtools .
```

**Go 模块代理问题：**

如果遇到 `proxy.golang.org Service Unavailable` 错误，使用国内代理：

```bash
GOPROXY=https://goproxy.cn,direct go build -o xdxtools .
```

### 3. 构建 enva (Rust)

```bash
# 使用系统 Rust 或 conda 环境
conda activate rust_build  # 或其他 Rust 环境

cd enva
cargo build --release

# 安装
cp target/release/enva $HOME/.cargo/bin/
```

### 4. 构建 rv (Rust)

rv 使用 Rust Edition 2024，需要 Rust 1.85+：

```bash
# 检查 Rust 版本
rustc --version

# 如果版本 < 1.85，使用 conda 环境
conda create -n rust-env rust=1.92 -c conda-forge -y
conda activate rust-env

# 构建
cd rv
cargo build --release --features=cli

# 安装
cp target/release/rv $HOME/.cargo/bin/
```

### 5. 安装所有二进制

```bash
mkdir -p $HOME/.cargo/bin
cp xdxtools $HOME/.cargo/bin/
cp enva/target/release/enva $HOME/.cargo/bin/
cp rv/target/release/rv $HOME/.cargo/bin/

# 添加到 PATH
export PATH="$HOME/.cargo/bin:$PATH"
```

### 6. 验证安装

```bash
xdxtools --version
enva --version
rv --version
```

## 常见问题与解决方案

### 1. Go 模块下载失败

**问题：**
```
go: proxy.golang.org Service Unavailable
```

**解决方案：**
```bash
# 使用国内 Go 代理
export GOPROXY=https://goproxy.cn,direct
# 或
export GOPROXY=https://goproxy.cn
```

### 2. Git 版本过旧

**问题：**
```
git ls-remote: exit status 129
usage: git ls-remote ...
```

**解决方案：**
```bash
# 使用 conda 安装新版 Git
conda install -n go-env git -c conda-forge -y
```

### 3. Rust 版本不足

**问题：**
```
error: failed to parse: Unknown edition: 2024
```

**解决方案：**
```bash
# 创建包含 Rust 1.92+ 的 conda 环境
conda create -n rust-env rust=1.92 -c conda-forge -y
conda activate rust-env

# 重新构建
cd rv
cargo build --release --features=cli
```

### 4. Rust 编译时网络问题

**问题：** crates 下载失败

**解决方案：** 使用国内 crates 镜像

在 `~/.cargo/config.toml` 中添加：
```toml
[source.crates-io]
replace-with = "ustc"

[source.ustc]
registry = "sparse+https://mirrors.ustc.edu.cn/crates.io-index/"
```

### 5. conda 环境中的 Go 找不到

**问题：** `go: command not found`

**解决方案：**
```bash
# 激活 conda 环境后使用完整路径
source /path/to/conda/bin/activate go-env
which go  # 确认路径
```

## 使用 setup.sh 自动化构建

项目提供了自动化脚本：

```bash
# 完整构建（包含环境创建）
./scripts/setup.sh

# 跳过部分步骤
./scripts/setup.sh --skip-envs       # 跳过 conda 环境
./scripts/setup.sh --skip-r-packages # 跳过 R 包
./scripts/setup.sh --dry-run         # 模拟运行
```

## 构建产物

| 文件 | 大小 | 说明 |
|------|------|------|
| `xdxtools` | ~12MB | 主 CLI 工具（静态链接） |
| `enva` | ~5MB | 环境管理器 |
| `rv` | ~30MB | R 包管理器 |

## 开发相关命令

```bash
# 运行测试
go test ./...

# 带覆盖率的测试
go test -cover ./...

# 特定测试
go test -v ./internal/input -run TestAdapterGenerator

# 静态分析
go vet ./...

# 格式化
gofmt -w .
```

## 环境变量参考

| 变量 | 用途 | 示例 |
|------|------|------|
| `GOPROXY` | Go 模块代理 | `https://goproxy.cn,direct` |
| `CGO_ENABLED` | 是否启用 CGO | `0`（静态构建） |
| `CARGO_HOME` | Rust 缓存目录 | `~/.cargo` |
| `RUSTUP_TOOLCHAIN` | 指定 Rust 版本 | `stable`, `1.92` |

## 相关文件

- `scripts/setup.sh` - 自动化构建脚本
- `scripts/build.sh` - Go 构建脚本
- `go.mod` - Go 依赖定义
- `enva/Cargo.toml` - enva 依赖
- `rv/Cargo.toml` - rv 依赖
