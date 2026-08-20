# Handoff 2026-08-20: Gate 6 BS-PDX Xenofilx Sort-Memory Optimization

## 1. 本次任务目标

用户要求：
1. 对 Xenofilx 实施性能优化（将排序内存从硬编码 256 MiB 改为可配置）；
2. 用优化后的 Xenofilx 重新处理当前 Gate 6 BS-PDX `step2-check` 任务。

## 2. 已完成的实施

### 2.1 Xenofilx 代码优化（已提交）

- 仓库：`xenofilx/`（子模块，Go 实现）
- 提交：`1184b72 feat: configurable sort memory for large BAM inputs`
- 变更文件：
  - `internal/config/config.go`：新增 `SortMemoryBytes int64` 字段（0 = 历史默认 256 MiB）
  - `internal/filter/filter.go`：`filterTotalSortMemoryBudgetBytes(configuration)`、`effectiveFilterWorkerCount(requested, total)`、`sortMemoryLimitForWorkers(worker, total)` 改为接受可配置预算；`Filter`/`FilterParallel` 从配置读取
  - `internal/filter/filter_test.go`：适配新签名，新增自定义预算测试
  - `pkg/cli/run.go`：新增 `--sort-memory` 标志（支持 `K/M/G/T` 后缀，默认 256M），`parseMemorySize` 解析函数
- 验证：`go build ./...`、`go vet ./...`、`go test -race ./...` 全部通过

### 2.2 构建部署（已上传 paracloud）

- 本地构建：
  - 普通构建二进制因 glibc 过新（需要 GLIBC_2.34）无法在 paracloud 运行
  - 改用 `CGO_ENABLED=0 go build` 静态构建成功
  - 版本号：`0.1.0-direct-bamdriver-region-r35`
  - SHA-256：`caed65ea46c7f692a6f9b0b3226f79d0c9a0d67826c1e8633f3018046a8c34cc`
- 远端部署：
  - 上传至 `/public3/home/scg9946/.cargo/bin/xenofilx`（替换 r34）
  - 旧版备份为 `/public3/home/scg9946/.cargo/bin/xenofilx.r34-backup`
  - 已 `chmod 755` 并验证版本输出为 `xenofilx 0.1.0-direct-bamdriver-region-r35`

### 2.3 Sealed catalog 更新（已提交 + 远端已更新）

- 仓库：`craftmake/`（子模块）
- 提交：`9667551 feat: raise Xenofilx sort memory budget`
- 变更文件：
  - `workflows/BeaverPDX/step2-check.yaml`：Xenofilx 命令增加 `--sort-memory 96G`
  - `internal/compiler/beaverpdx_step2_check_test.go`：断言 `--sort-memory 96G`
- 远端 catalog `/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/catalog/BeaverPDX/step2-check.yaml` 已用 Python 原子更新（第 140 行加入 `--sort-memory 96G`）

### 2.4 根仓库指针与文档（已提交）

- 提交：`9e026fc feat: wire Xenofilx sort memory optimization`
- 变更：`craftmake`、`xenofilx` 子模块指针 + `docs/notes/2026-08-16-gate6-bs-pdx-craftmake-execution.md` 记录

## 3. 当前 Gate 6 BS-PDX 执行状态（截至交班）

### 3.1 关键问题：r34 在 12 小时窗口内 TIMEOUT

- **modern**：Xenofilx `41545883` 在 128 GiB / 12 小时窗口下 `TIMEOUT`（12:00:15），controller `41545831` FAILED 5:0
- **legacy**：Xenofilx 同样超时，validators 成功但 controller `41546541` FAILED 5:0；legacy 使用的仍是 r34
- 根因：r34 的 `bamnative.Sort` 在 256 MiB 预算下对 29.3 GiB hg38 BAM 做外部排序，产生 50+ GiB 写放大，导致排序阶段耗时超过整个 12 小时窗口

### 3.2 已部署 r35 的预期效果

- `--sort-memory 96G` 使排序预算提升 384 倍，外部排序应能在内存中完成
- 分类/过滤算法未改动，输出语义与 r34 一致
- **注意**：catalog 的 `--sort-memory` 只有新提交的 run 才会生效；当前运行的 r34 任务已失败，需要新 snapshot 重跑

### 3.3 待办

1. **为 modern 与 legacy 分别 resolve 新的 immutable snapshot**（因为 catalog 变更 = 科学设置变更，需新 snapshot）：
   - modern：以 `run-20260817T001228Z-vclyug` 为 parent（已接受的 step2）
   - legacy：以 `run-20260817T102532Z-emvbei` 为 parent（已接受的 step2）
   - 绑定 accepted parent `work/` 输出（trim、QC、fastqc_raw、fastqc_clean、bsmap）
   - 新 snapshot 的 `step2-check` 资源建议保持 `4 CPU / 128 GiB / 12h / amd_512`
2. **重跑 modern/legacy `step2-check`**（Otter → Craftmake，确保子模块 pointer 已更新到含 `--sort-memory` 的 catalog）
3. 完成 `step3` / `step3-check`、artifact verify、artifact compare

## 4. 关键约束（必须遵守）

- **只走 Otter → Craftmake → Slurm 主线**，不手工 `sbatch` workflow task
- modern 与 legacy 使用**独立 immutable snapshots**，不得跨 snapshot 复用 Craftmake state/cache
- 只允许把已接受的 parent output 目录作为 symlink 绑定到新 snapshot 的 `work/`
- 修改代码/catalog/配置前先提交已有工作，并在 `docs/notes/2026-08-16-gate6-bs-pdx-craftmake-execution.md` 记录
- 不触碰无关 dirty 子模块：`enva`、`methx`、`qctb`（以及工作区 `inst/rules/methrix_object.smk`）

## 5. 相关提交记录

| 仓库 | 提交 | 说明 |
|---|---|---|
| xenofilx | `1184b72` | `--sort-memory` 可配置排序内存 |
| craftmake | `9667551` | catalog 增加 `--sort-memory 96G` + 编译器测试 |
| 根仓库 | `9e026fc` | 子模块指针 + docs |

## 6. 重要参考路径

- 执行记录：`docs/notes/2026-08-16-gate6-bs-pdx-craftmake-execution.md`
- 远端 runtime：`/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/`
- modern project：`projects/bs-pdx-SRR23802966/`
- legacy project：`projects/bs-pdx-SRR23802966-legacy-equivalent/`
