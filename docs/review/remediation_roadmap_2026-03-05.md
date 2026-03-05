# 整改路线图（更新于 2026-03-05）

## 目标
将整改从“实现补缺”切换到“发布门禁和回归证据沉淀”，确保每次发布都能证明主仓与子仓协作闭环仍然成立。

## 阶段 A（已完成，2026-03-05）

1. 安装命名统一（P0）
- `scripts/install.sh` 以 `methrix-cli` 为安装名。
- 保留 `methrix -> methrix-cli` 兼容链接。

2. 子仓构建覆盖（P1）
- `scripts/build-all-submodules.sh` 覆盖 `fqc`、`htseq2matrix`、`methrix-cli` 等必需命令。
- 默认严格模式：缺必需工具即失败退出。

3. 一致性门禁（P1）
- 已有 `scripts/verify_toolchain_consistency.sh`。
- 已接入 CI：
  - PR/Push：`.github/workflows/go.yml`
  - Release：`.github/workflows/release.yml`

## 阶段 B（下一迭代）

1. 最小 dry-run 回归固化（P1）
- RRBS：`xdxtools run --dry-run --engine local ...`
- RNASEQ：`xdxtools run --dry-run --engine slurm ...`
- 产物要求：保留命令、时间戳、关键输出到 `docs/review/`。

2. 发布前门禁清单化（P2）
- 使用 `docs/review/release_gates_2026-03-05.md` 作为发布勾稽清单。
- 每次发布前更新“通过/失败/备注”。

## DoD

1. rules/install/build/CI 一致性检查通过。
2. 8 个子工具构建产物可执行，版本/帮助输出可解析。
3. RRBS + RNASEQ 最小 dry-run 回归记录齐全。
4. 审查报告、证据索引、整改路线图三者一致。
