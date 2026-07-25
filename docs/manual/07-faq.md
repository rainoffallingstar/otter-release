# 第七章：常见问题与排查

## `otter: command not found`

```bash
command -v otter
printf '%s\n' "$PATH"
```

确认安装目录已加入 `PATH`。若源码构建出的程序仍显示 `xdxtools`，这是父仓代码命名迁移尚未完成，不要仅靠重命名文件掩盖根命令和状态路径差异。

## 安装 URL 返回 404

确认使用：

```text
https://github.com/rainoffallingstar/otter
```

私有仓库或 release 资产需要配置 `GITHUB_TOKEN`、`GH_TOKEN` 或 `GITHUB_PAT`。不要把 token 写入仓库 URL。

## 找不到子模块

```bash
git submodule sync --recursive
git submodule update --init --recursive
git submodule status --recursive
```

当前目录应包含 `craftmake`、`enva`、`fastqcx`、`xenofilx`、`pairbam`、`seq2mat`、`matsrun`、`qctb`、`methx`、`bamdriver`。

## 环境名仍是旧前缀

当前环境统一为：

```text
otter-core
otter-snakemake
otter-extra
```

若 `enva list` 仍显示旧前缀，先核对环境 YAML、安装器和运行规则是否仍引用旧名。不要直接重命名物理目录造成前缀元数据不一致。

## Snakemake 不可用时能否直接改用 craftmake

不能一概而论。`craftmake` 正在接入，当前是双轨；只有已通过任务图、资源、恢复和科学产物等价测试的流程才能切换。生产兼容路径仍需要 `otter-snakemake`。

## `methx` 报 HDF5 错误

```bash
export HDF5_DIR="$CONDA_PREFIX"
export HDF5_INCLUDE_DIR="$HDF5_DIR/include"
export HDF5_LIB_DIR="$HDF5_DIR/lib"
export PKG_CONFIG_PATH="$HDF5_DIR/lib/pkgconfig:$PKG_CONFIG_PATH"
export LD_LIBRARY_PATH="$HDF5_DIR/lib:$LD_LIBRARY_PATH"
```

确认 HDF5 header、library 与运行时来自兼容版本。

## FASTQ 配对失败

- 检查 R1/R2 后缀是否与 `--suffix1`、`--suffix2` 一致。
- 检查 pdata 的 `sampleid` 与文件名前缀大小写是否一致。
- 确保每个样本同时存在 R1 和 R2。

## SLURM 提交失败

```bash
sinfo
squeue -u "$(whoami)"
otter run --config config.yaml --engine local --parallel-jobs 4
```

核对分区、QOS、配额、CPU 和内存。切换本地引擎前先确认输入规模。

## 如何查看日志和恢复

```bash
otter task list
otter task status <task-id>
otter task logs <task-id> --follow
otter run --config <config.yaml> --resume
```

当前源码可能仍使用 `.xdxtools_state.json` 或旧 XDG 状态目录；这是实现兼容事实。只有代码迁移完成后，文档才能把状态文件写成 `.otter_state.json`。

## 外部名称为何没有改

FastQC、MultiQC、FASTQ、BAM、Bismark、HTSeq、Methrix、HDF5 和 rMATS 是标准、格式、外部工具或科学领域名，保留是正确的，不属于旧产品名残留。

## 获取帮助

- `otter --help`
- <https://github.com/rainoffallingstar/otter/issues>
