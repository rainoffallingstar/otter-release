# xdxtools GitHub Actions 二进制编译实施计划

## 项目概述

- **模块名**: github.com/xdxtools/xdxtools-go
- **Go 版本**: 1.24.0
- **当前版本**: 0.1.0 (硬编码在 cmd/root.go)
- **二进制名**: xdxtools
- **现有测试**: 14 个测试文件覆盖核心功能

---

## Phase 1: 版本注入系统改造

### 1.1 修改 cmd/root.go

**目标**: 从硬编码版本改为动态版本注入

**当前状态**:
```go
var rootCmd = &cobra.Command{
    ...
    Version: "0.1.0",
}
```

**修改后**:
```go
var (
    buildVersion   = "0.1.0"
    buildCommit    = ""
    buildDate      = ""
)

var rootCmd = &cobra.Command{
    ...
    Version: fmt.Sprintf("%s (commit: %s, date: %s)", buildVersion, buildCommit, buildDate),
}
```

**新增命令**:
- `xdxtools version` - 显示详细版本信息
- `xdxtools version --short` - 仅显示版本号

**涉及文件**:
- `cmd/root.go` - 修改版本变量定义
- `cmd/version.go` - 新增版本命令 (可选)

### 1.2 创建构建脚本 (scripts/build.sh)

**目标**: 提供本地构建能力和 CI 一致性

**功能**:
- 编译标志注入 (ldflags)
- 多平台支持 (GOOS/GOARCH)
- 优化标志 (-s -w, -trimpath)
- 静态链接

**脚本内容**:
```bash
#!/bin/bash
set -e

VERSION=${1:-"0.1.0"}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-X 'main.buildVersion=${VERSION}' -X 'main.buildCommit=${COMMIT}' -X 'main.buildDate=${DATE}' -s -w"

# 构建函数
build() {
    local os=$1
    local arch=$2
    local ext=$3

    echo "Building for ${os}/${arch}..."
    GOOS=${os} GOARCH=${arch} CGO_ENABLED=0 go build \
        -ldflags "${LDFLAGS}" \
        -trimpath \
        -o "dist/xdxtools-${VERSION}-${os}-${arch}${ext}" \
        .

    # 生成校验和
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "dist/xdxtools-${VERSION}-${os}-${arch}${ext}" > "dist/xdxtools-${VERSION}-${os}-${arch}${ext}.sha256"
    fi
}

# 创建输出目录
mkdir -p dist

# 构建多平台
build linux amd64 ""
build linux arm64 ""
build darwin amd64 ""
build darwin arm64 ""
build windows amd64 ".exe"

echo "Build complete!"
```

### 1.3 创建 Makefile (可选)

**目标**: 提供标准化的构建命令

**内容**:
```makefile
.PHONY: build test clean

VERSION ?= 0.1.0
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X 'main.buildVersion=${VERSION}' -X 'main.buildCommit=${COMMIT}' -X 'main.buildDate=${DATE}' -s -w

build:
	@./scripts/build.sh $(VERSION)

test:
	go test -v ./...

clean:
	rm -rf dist/

install:
	go install -ldflags "$(LDFLAGS)" .
```

---

## Phase 2: CI 工作流创建

### 2.1 创建 `.github/workflows/go.yml`

**触发条件**:
- Push 到 master/main 分支
- Pull Request 到 master/main 分支

**任务流程**:
1. 检出代码 (actions/checkout@v4)
2. 设置 Go (actions/setup-go@v5)
3. 缓存依赖 (actions/cache@v4)
4. 运行测试 (`go test -v ./...`)
5. 运行基准测试 (可选)
6. 构建多平台二进制
7. 上传构建产物 (actions/upload-artifact@v4)

**完整配置**:
```yaml
name: CI Build

on:
  push:
    branches: [ master, main ]
  pull_request:
    branches: [ master, main ]

env:
  GO_VERSION: '1.24.0'

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - name: Check out code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Cache Go modules
        uses: actions/cache@v4
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Run tests
        run: |
          go test -v ./...
          go test -race ./...  # 竞态检测

      - name: Generate coverage report
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out
          flags: unittests
          name: codecov-umbrella

  build:
    name: Build
    needs: test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, windows, darwin]
        goarch: [amd64, arm64]
        exclude:
          # Windows 不支持 ARM64 (如需支持需特殊配置)
          - goos: windows
            goarch: arm64
    steps:
      - name: Check out code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Cache Go modules
        uses: actions/cache@v4
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Get version
        id: version
        run: |
          if [[ $GITHUB_REF == refs/tags/v* ]]; then
            echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT
          else
            echo "VERSION=0.0.0-dev+${GITHUB_SHA::8}" >> $GITHUB_OUTPUT
          fi

      - name: Get commit hash
        id: commit
        run: echo "COMMIT=${GITHUB_SHA::8}" >> $GITHUB_OUTPUT

      - name: Build binary
        run: |
          VERSION=${{ steps.version.outputs.VERSION }}
          COMMIT=${{ steps.commit.outputs.COMMIT }}
          DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
          LDFLAGS="-X 'main.buildVersion=${VERSION}' -X 'main.buildCommit=${COMMIT}' -X 'main.buildDate=${DATE}' -s -w"

          EXT=""
          if [[ "${{ matrix.goos }}" == "windows" ]]; then
            EXT=".exe"
          fi

          mkdir -p dist
          GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} CGO_ENABLED=0 \
            go build -ldflags "${LDFLAGS}" -trimpath \
            -o "dist/xdxtools-${VERSION}-${{ matrix.goos }}-${{ matrix.goarch }}${EXT}" .

          # 生成校验和
          if [[ "${{ matrix.goos }}" == "linux" ]] && command -v sha256sum >/dev/null 2>&1; then
            sha256sum "dist/xdxtools-${VERSION}-${{ matrix.goos }}-${{ matrix.goarch }}${EXT}" \
              > "dist/xdxtools-${VERSION}-${{ matrix.goos }}-${{ matrix.goarch }}${EXT}.sha256"
          fi

      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: xdxtools-${{ matrix.goos }}-${{ matrix.goarch }}
          path: dist/*
          retention-days: 30

  # 构建后合并所有平台产物
  package:
    name: Package
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/master' || github.ref == 'refs/heads/main'
    steps:
      - name: Download all artifacts
        uses: actions/download-artifact@v4

      - name: Setup SHA256 tool
        run: |
          # 安装 shasum (macOS) 或 sha256sum (Linux)
          if [[ "${{ runner.os }}" == "macOS" ]]; then
            brew install --cask basictex || true
          fi

      - name: Generate checksums
        run: |
          mkdir -p dist
          find . -name "*.sha256" -exec cat {} \; > dist/checksums.txt

      - name: Upload to artifacts
        uses: actions/upload-artifact@v4
        with:
          name: xdxtools-snapshot
          path: dist/*
          retention-days: 7
```

### 2.2 创建 `.github/workflows/release.yml`

**触发条件**:
- 创建版本标签 (v*)

**功能**:
1. 构建多平台二进制
2. 生成 checksums
3. 创建 GitHub Release
4. 上传二进制到 Release

**完整配置**:
```yaml
name: Release

on:
  push:
    tags:
      - 'v*'  # 触发所有 v 开头的标签

permissions:
  contents: write  # 需要创建 releases

env:
  GO_VERSION: '1.24.0'

jobs:
  build:
    name: Build Release
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, windows, darwin]
        goarch: [amd64, arm64]
        exclude:
          - goos: windows
            goarch: arm64
    steps:
      - name: Check out code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Build binary
        env:
          VERSION: ${{ github.ref_name }}
        run: |
          COMMIT=$(git rev-parse --short HEAD)
          DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
          LDFLAGS="-X 'main.buildVersion=${VERSION}' -X 'main.buildCommit=${COMMIT}' -X 'main.buildDate=${DATE}' -s -w"

          EXT=""
          if [[ "${{ matrix.goos }}" == "windows" ]]; then
            EXT=".exe"
          fi

          mkdir -p release
          GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} CGO_ENABLED=0 \
            go build -ldflags "${LDFLAGS}" -trimpath \
            -o "release/xdxtools-${VERSION}-${{ matrix.goos }}-${{ matrix.goarch }}${EXT}" .

          # 生成校验和
          if command -v sha256sum >/dev/null 2>&1; then
            sha256sum "release/xdxtools-${VERSION}-${{ matrix.goos }}-${{ matrix.goarch }}${EXT}" \
              > "release/xdxtools-${VERSION}-${{ matrix.goos }}-${{ matrix.goarch }}${EXT}.sha256"
          fi

      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: release-${{ matrix.goos }}-${{ matrix.goarch }}
          path: release/*

  create-release:
    name: Create Release
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Download all artifacts
        uses: actions/download-artifact@v4

      - name: Prepare release
        run: |
          mkdir -p release
          find . -name "*.sha256" -exec cat {} \; > release/checksums.txt
          find . -type f -name "xdxtools-*" ! -name "*.sha256" -exec cp {} release/ \;

      - name: Generate changelog
        id: changelog
        run: |
          # 生成从上次标签到现在的变更日志
          if [[ $(git tag --list | wc -l) -gt 0 ]]; then
            LAST_TAG=$(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || git tag | head -1)
            echo "CHANGELOG<<EOF" >> $GITHUB_OUTPUT
            echo "## Changes since $LAST_TAG" >> $GITHUB_OUTPUT
            echo "" >> $GITHUB_OUTPUT
            git log --pretty=format:"- %s (%h)" $LAST_TAG..HEAD >> $GITHUB_OUTPUT
            echo "EOF" >> $GITHUB_OUTPUT
          else
            echo "CHANGELOG<<EOF" >> $GITHUB_OUTPUT
            echo "## Initial release" >> $GITHUB_OUTPUT
            echo "EOF" >> $GITHUB_OUTPUT
          fi

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v1
        with:
          name: Release ${{ github.ref_name }}
          body: |
            ${{ steps.changelog.outputs.CHANGELOG }}

            ## Downloads

            ### Checksums
            ```
            $(cat release/checksums.txt)
            ```

            ### Binaries
            - Linux (AMD64): `xdxtools-${{ github.ref_name }}-linux-amd64`
            - Linux (ARM64): `xdxtools-${{ github.ref_name }}-linux-arm64`
            - macOS (AMD64): `xdxtools-${{ github.ref_name }}-darwin-amd64`
            - macOS (ARM64): `xdxtools-${{ github.ref_name }}-darwin-arm64`
            - Windows (AMD64): `xdxtools-${{ github.ref_name }}-windows-amd64.exe`

            ## Installation

            ### Linux/macOS
            ```bash
            # 下载二进制
            curl -L https://github.com/xdxtools/xdxtools-go/releases/download/${{ github.ref_name }}/xdxtools-${{ github.ref_name }}-linux-amd64 -o xdxtools
            chmod +x xdxtools

            # 移动到 PATH
            sudo mv xdxtools /usr/local/bin/
            ```

            ### Windows (PowerShell)
            ```powershell
            # 使用 winget
            winget install xdxtools.xdxtools

            # 或手动下载
            Invoke-WebRequest -Uri "https://github.com/xdxtools/xdxtools-go/releases/download/${{ github.ref_name }}/xdxtools-${{ github.ref_name }}-windows-amd64.exe" -OutFile "xdxtools.exe"
            ```
          draft: false
          prerelease: false
          files: |
            release/*
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  # 验证构建的二进制
  verify:
    name: Verify Release Build
    needs: create-release
    runs-on: ubuntu-latest
    steps:
      - name: Download release artifacts
        uses: actions/download-artifact@v4
        with:
          name: release-linux-amd64
          path: verify/

      - name: Test binary
        run: |
          chmod +x verify/xdxtools-*
          ./verify/xdxtools-* version
```

---

## Phase 3: 文档和辅助配置

### 3.1 更新 `.gitignore`

**新增条目**:
```
# Build artifacts
dist/
release/
release-*/
```

### 3.2 创建安装文档 (INSTALL.md)

**内容包含**:
- 二进制安装方法
- 包管理器安装 (Homebrew, Scoop, winget)
- Docker 安装
- 从源码编译

**示例**:
```markdown
# 安装 xdxtools

## 方法 1: 直接下载二进制

从 [Releases](https://github.com/xdxtools/xdxtools-go/releases) 页面下载对应平台的二进制文件。

## 方法 2: 包管理器

### macOS (Homebrew)
```bash
brew tap xdxtools/tap
brew install xdxtools
```

### Windows (Scoop)
```bash
scoop install xdxtools
```

### Windows (winget)
```bash
winget install xdxtools.xdxtools
```

## 方法 3: Docker
```bash
docker run --rm -it xdxtools/xdxtools:latest
```

## 方法 4: 从源码编译
```bash
go install github.com/xdxtools/xdxtools-go@latest
```
```

### 3.3 创建发布脚本 (scripts/release.sh)

**功能**:
- 自动打标签
- 推送标签
- 触发 GitHub Actions

**内容**:
```bash
#!/bin/bash
set -e

VERSION=$1
if [[ -z "$VERSION" ]]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 1.0.0"
  exit 1
fi

# 验证版本格式
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: Version must match vX.Y.Z format"
  exit 1
fi

# 更新版本号 (如果需要)
read -p "Update version in cmd/root.go to $VERSION? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  sed -i.bak "s/Version: \".*\"/Version: \"$VERSION\"/" cmd/root.go
  echo "Version updated in cmd/root.go"
fi

# 创建标签
git tag -a "$VERSION" -m "Release $VERSION"

# 推送标签
read -p "Push tags to GitHub? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  git push origin "$VERSION"
  echo "Tag pushed. GitHub Actions will build and create release."
fi

echo "Done!"
```

---

## 实施顺序和依赖关系

### 顺序 1: 基础改造
1. ✅ 修改 `cmd/root.go` 添加版本变量
2. ✅ 创建 `scripts/build.sh` 构建脚本
3. ✅ 本地测试构建

**验证步骤**:
```bash
# 本地构建测试
./scripts/build.sh 0.1.0
ls dist/
# 验证版本信息
./dist/xdxtools-0.1.0-linux-amd64 version
```

### 顺序 2: CI 工作流
1. ✅ 创建 `.github/workflows/go.yml`
2. ✅ 提交并测试 CI 流程
3. ✅ 验证所有平台构建成功

**验证步骤**:
```bash
# 提交代码触发 CI
git add .
git commit -m "feat: Add multi-platform build workflow"
git push origin master
# 检查 GitHub Actions 页面
```

### 顺序 3: 发布工作流
1. ✅ 创建 `.github/workflows/release.yml`
2. ✅ 创建 `scripts/release.sh`
3. ✅ 创建 `INSTALL.md`
4. ✅ 首次测试发布

**验证步骤**:
```bash
# 首次测试发布
./scripts/release.sh v0.1.0
# 检查 GitHub Releases 页面
```

### 顺序 4: 优化和完善
1. ✅ 添加代码覆盖率报告
2. ✅ 添加竞态检测
3. ✅ 优化构建缓存
4. ✅ 添加安装文档

---

## 关键文件列表

### 需要修改的文件

1. **cmd/root.go**
   - 添加版本变量 (`buildVersion`, `buildCommit`, `buildDate`)
   - 修改 cobra 版本显示逻辑
   - 添加详细版本信息格式化

2. **.gitignore**
   - 添加构建产物目录 (`dist/`, `release/`)

### 需要创建的文件

3. **scripts/build.sh** (新增)
   - 多平台构建脚本
   - ldflags 版本注入
   - 校验和生成

4. **scripts/release.sh** (新增)
   - 自动化发布脚本
   - 标签管理

5. **.github/workflows/go.yml** (新增)
   - CI 工作流
   - 多平台矩阵构建
   - 测试和构建

6. **.github/workflows/release.yml** (新增)
   - 发布工作流
   - 自动创建 GitHub Release

7. **Makefile** (可选)
   - 标准化构建命令
   - 本地开发便利

8. **INSTALL.md** (可选)
   - 安装文档
   - 多平台安装说明

---

## 构建优化说明

### 1. 二进制体积优化
- `-s -w`: 去除符号表和调试信息，减少 20-30% 体积
- `-trimpath`: 去除文件系统路径，安全性提升
- `CGO_ENABLED=0`: 纯 Go 构建，无 C 依赖

### 2. 静态链接
- 默认 Go 编译是静态链接
- 无需额外配置即可获得独立二进制

### 3. 版本注入
使用 ldflags 注入构建时信息:
```bash
LDFLAGS="-X 'main.buildVersion=${VERSION}' -X 'main.buildCommit=${COMMIT}' -X 'main.buildDate=${DATE}'"
```

### 4. 多平台支持
通过 `GOOS` 和 `GOARCH` 环境变量实现:
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ...
```

---

## 测试策略

### 单元测试
- 现有 14 个测试文件自动运行
- 覆盖率报告上传到 Codecov

### 集成测试
- CI 流程中验证二进制可执行
- 发布流程中验证版本信息

### 跨平台测试
- 在 Linux (ubuntu-latest) 构建所有平台
- 实际部署时在各平台验证

---

## 发布流程

### 快照版本 (Snapshot)
- **触发**: Push 到 master/main 分支
- **行为**: 构建但不发布
- **用途**: 内部测试、CI 验证

### 正式版本 (Release)
- **触发**: 创建版本标签 (v*)
- **行为**: 构建 + 创建 GitHub Release
- **用途**: 正式发布

### 版本命名规范
- **开发版本**: `0.0.0-dev+<commit-hash>`
- **正式版本**: `vX.Y.Z` (遵循语义化版本)

---

## 预期成果

### 构建时间
- 单平台: ~2-3 分钟
- 全平台矩阵: ~5-8 分钟

### 二进制体积 (估算)
- Linux AMD64: ~15-20 MB
- macOS AMD64: ~15-20 MB
- Windows EXE: ~18-22 MB
- ARM64 平台: 类似 AMD64

### 支持的平台矩阵
| OS | Arch | 状态 | 备注 |
|----|------|------|------|
| Linux | amd64 | ✅ | 主要平台 |
| Linux | arm64 | ✅ | 服务器/ARM 设备 |
| macOS | amd64 | ✅ | Intel Mac |
| macOS | arm64 | ✅ | Apple Silicon |
| Windows | amd64 | ✅ | Windows 10/11 |

---

## 后续扩展建议

1. **包管理器集成**
   - 提交到 Homebrew
   - 提交到 Scoop (Windows)
   - 提交到 winget (Windows)

2. **容器优化**
   - 多阶段 Docker 构建
   - 最小化基础镜像

3. **签名和验证**
   - 代码签名 (Windows)
   - 签名验证 (macOS)

4. **自动更新**
   - 实现 self-update 命令
   - 更新检查机制

5. **CI 增强**
   - 并行构建优化
   - 构建缓存策略

---

## 成本评估

### GitHub Actions 费用
- **公共仓库**: 免费 (2000 分钟/月)
- **私有仓库**: $0.008/分钟
- **本项目**: 预计 50-100 分钟/月，免费

### 存储费用
- 构建产物: ~100 MB × 30 天 = 3 GB 免费额度内

---

## 总结

本方案提供了完整的 CI/CD 流程，支持:
- ✅ 多平台二进制构建
- ✅ 动态版本注入
- ✅ 测试优先策略
- ✅ 自动发布流程
- ✅ 构建优化

通过三个阶段的实施，可以建立起高效、可靠的发布流程，支持 xdxtools 项目的持续集成和自动发布需求。

---

**文档版本**: v1.0
**创建日期**: 2026-01-10
**维护者**: xdxtools 团队
