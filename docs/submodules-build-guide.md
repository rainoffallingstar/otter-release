# otter 子模块构建指南

## 边界

`otter` 父仓通过 Git submodule 组合 10 个独立仓库。子模块有各自的历史、测试和发布节奏；父仓文档修改不应进入子模块目录。

```text
otter → craftmake → enva → 算子 → bamdriver
```

`craftmake` 是 Snakemake 的 Go 替代执行层，但当前仍在接入，必须按双轨状态构建和验证。

## 当前子模块

| 层级 | 子模块目录 | 语言 | 目标二进制 | 角色 |
|---|---|---|---|---|
| 执行 | `craftmake/` | Go | `craftmake` | workflow spec/DAG/local/SLURM |
| 环境 | `enva/` | Rust | `enva` | rattler-first 环境管理 |
| 算子 | `fastqcx/` | Rust | `fastqcx` | FASTQ 质控，FastQC/MultiQC 兼容 |
| 算子 | `xenofilx/` | Go | `xenofilx` | PDX 物种过滤 |
| 算子 | `pairbam/` | Go | `pairbam` | 配对 BAM 过滤 |
| 算子 | `seq2mat/` | Go | `seq2mat` | HTSeq count-to-matrix |
| 算子 | `matsrun/` | Go | `matsrun` | rMATS 编排 |
| 算子 | `qctb/` | Rust | `qctb` | QC 汇总 |
| 算子 | `methx/` | Rust | `methx` | 甲基化/HDF5 处理 |
| 基础 | `bamdriver/` | Go | `bamdriver` | BAM 共享底层能力 |

## 初始化

```bash
git clone --recurse-submodules https://github.com/rainoffallingstar/otter.git
cd otter
```

已有 checkout：

```bash
git submodule sync --recursive
git submodule update --init --recursive
git submodule status --recursive
```

不要在没有审查的情况下运行 `git submodule update --remote --merge`，因为它会改变父仓记录的子模块指针。

## Go 子模块

推荐环境：

```bash
conda activate go-env
export CGO_ENABLED=0
```

对每个 Go 子模块，在其目录内按模块实际入口构建：

```bash
go test ./...
go vet ./...
go build ./...
```

适用目录：`craftmake/`、`xenofilx/`、`pairbam/`、`seq2mat/`、`matsrun/`、`bamdriver/`。

发布构建应使用该仓库声明的 `cmd/<name>` 或根包入口，不要根据历史仓库名猜测入口。

## Rust 子模块

推荐环境：

```bash
conda activate rust_build
```

在 `enva/`、`fastqcx/`、`qctb/`、`methx/` 中执行：

```bash
cargo fmt --check
cargo check --locked --all-targets
cargo clippy --locked --all-targets -- -D warnings
cargo test --locked
cargo build --locked --release
```

`methx` 需要 HDF5：

```bash
export HDF5_DIR="$CONDA_PREFIX"
export HDF5_INCLUDE_DIR="$HDF5_DIR/include"
export HDF5_LIB_DIR="$HDF5_DIR/lib"
export PKG_CONFIG_PATH="$HDF5_DIR/lib/pkgconfig:$PKG_CONFIG_PATH"
export LD_LIBRARY_PATH="$HDF5_DIR/lib:$LD_LIBRARY_PATH"
```

## 安装结果

建议把二进制安装到用户可写目录：

```bash
install -d "$HOME/.cargo/bin"
install -m 755 <built-binary> "$HOME/.cargo/bin/<current-name>"
```

验证：

```bash
command -v craftmake enva fastqcx xenofilx pairbam seq2mat matsrun qctb methx bamdriver
```

## 双轨验证

- Snakemake 继续在 `otter-snakemake` 环境中承担当前生产工作流。
- `craftmake` 构建成功仅证明执行层可编译，不证明已接入 `otter`。
- 切换前需对相同 fixture 比较任务图、资源参数、失败传播、恢复状态和关键科学产物。
- `otter-core` 与 `otter-extra` 中的算子解析必须使用当前名称。

## 历史映射

| 当前目录 | 历史目录/仓库名 |
|---|---|
| `fastqcx/` | `fastqc-rs/` |
| `xenofilx/` | `xenofilter-go/` |
| `pairbam/` | `Paireads/` |
| `seq2mat/` | `htseq2matrix-go/` |
| `matsrun/` | `gomats/` |
| `methx/` | `methrix-cli/` |
| `bamdriver/` | `bamdriver-go/` |

历史审查报告中的旧路径用于定位当时证据，不应被视为当前 checkout 指令。

## 父仓指针交付

只有在子模块自身门禁通过、变更已由有权限者提交/推送后，父仓才可更新 gitlink。本文档迁移没有修改任何子模块、gitlink、脚本或 CI，也没有提交或推送。
