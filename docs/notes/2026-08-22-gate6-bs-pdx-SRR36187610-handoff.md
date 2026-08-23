# Handoff 2026-08-22: Gate 6 BS-PDX 数据集切换与 SRR36187610 流水线

## 1. 本次任务目标

用户要求：
1. 解决 BS-PDX `SRR23802966` 数据集体量异常庞大（3.88 亿配对读段、29.3 GiB hg38 BAM）导致的 Xenofilx 排序内存 OOM 问题；
2. 在 NCBI SRA 中寻找一个体量更小、可处理的公共 BS-PDX 数据集；
3. 用 `getdown` 工具完成新数据的本地下载 → 上传 → 解码 → 验证 → provenance 固化 → 分析全流程。

## 2. 决策与数据集切换

### 2.1 为什么切换

- `SRR23802966` 是 Whole-Genome Bisulfite 30X，`387,683,390` 配对读段（116.3 Gb），hg38 BAM 达 **29.3 GiB**。
- 即使将 Xenofilx `--sort-memory` 调至 12G（峰值 RSS ~35 GiB），现代 worker `41565810` 仍在 6.5 小时后被 Slurm OOM killer 杀死——根因是宿主节点 `e0706` 100% 共驻（`CPUAlloc=128/128`、`CPULoad=96.55`），主机级内存被共驻租户挤占，而非单进程堆超限。
- 用户决定**废弃 SRR23802966 主线，全面切换到新数据集**。

### 2.2 选定的新数据集

- **`SRR36187610`**（LTL331R PDX eRRBS，项目 PRJNA1368947 / GSE311321）
- 人前列腺肿瘤移植于小鼠宿主（genuine PDX），同时比对到 `hg38`（graft）+ `mm10`（host）
- 体量：**71,242,412 配对读段**（约为 SRR23802966 的 1/5.4），archive 仅 2.42 GB

## 3. 已完成的获取流程（getdown 工作流）

### 3.1 本地下载（getdown）

- 工具：`/home/fallingstar10/.cache/getdown-sra-20260813T041223Z-sdl-fallback`（原生 Go NCBI SDL 解析器）
- 命令：`getdown sra --accession SRR36187610 --kind sra --decode none --jobs 4`
- 产物：`SRR36187610.sra`，**2,424,481,661 字节**，MD5 `1f531242d95d33fd6547e274a53fd086`，SHA-256 `713d1eb153b3db279e8aff6df7139c48687458e79a79008f2ceed5248b2569bb`
- 本地 staging：`/home/fallingstar10/.cache/otter-gate6-sra-20260813T041223Z-local-getdown/bs-pdx-SRR36187610/`

### 3.2 上传与解码

- 远端 root：`/public3/home/scg9946/otter-gate6/acquisitions/bs-pdx-SRR36187610-20260821T091900Z/`
- archive 上传后远程 SHA-256 校验通过
- 解码（Slurm job `41572083`）：`fasterq-dump --split-files -e 8` + `pigz -n`，**12 分 30 秒**完成
- 解码产物：R1 `aa8a2225…0706`（2,235,653,837 B）、R2 `2d564420…c05`（2,291,050,570 B），各 **71,242,412** 配对记录

### 3.3 独立验证

- 计算节点验证 job `41572209`：`sha256sum -c checksums.sha256`、`gzip -t`、SHA-256 重算、配对记录数审计全部通过。

### 3.4 Provenance 固化

- `otter acquisition publish` 生成不可变 `otter.sra-acquisition/v1`：
  - 路径：`…/bs-pdx-SRR36187610-20260821T091900Z/provenance/otter-sra-acquisition.json`
  - ID：`sra-20260821T111642Z-srr36187610`（`immutable: true`，444 只读）
  - 绑定 hg38（graft）+ mm10（host）reference manifest digest

## 4. 已完成的流水线执行

### 4.1 新 canonical project

- 项目：`bs-pdx-SRR36187610`（`/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR36187610/`）
- 配置：project.yaml、samples.tsv、references.lock.yaml、project.lock.yaml、data/ 符号链接
- 不可变 snapshot：**`run-20260821T112353Z-tunvxb`**（resolve job `41572235`）

### 4.2 Step1（Trim Galore + FastQC）

- Controller `41572248`：**COMPLETED `0:0`，28 分 14 秒**（对比 SRR23802966 的数小时）
- 产出：trim/、fastqc_raw/、fastqc_clean/、QC/

### 4.3 Step2（BeaverPDX 双参考 Bismark map_and_sort）

- Controller `41572422`：**COMPLETED `0:0`**
- 产出：
  - `SRR36187610_hg38.bam` **3,125,450,579 B（~2.9 GB）** + BAI
  - `SRR36187610_mm10.bam` 91,850,996 B（~88 MB）+ BAI
  - qualimap + GC-bias（hg38/mm10 各一套）全部成功
- 6/6 任务 succeeded

### 4.4 Step2-check（Xenofilx 宿主过滤）

- Controller `41576257`：**COMPLETED `0:0`，1 小时 46 分**
- Xenofilx worker `41576288.0`：峰值 RSS ~31.9 GB，CPU 3h12m，读 46.8 GB / 写 30.2 GB
- 产出：
  - `SRR36187610_fixed_hg38_Filtered.bam` **3,388,275,652 B（~3.16 GB）** + BAI
  - `filtered-bam-validation.json`
- **关键成果**：Xenofilx 在 1h45m 内成功完成（对比 SRR23802966 的 6.5h OOM），彻底解决内存问题

## 5. 性能分析结论（用户关注点）

### 5.1 为什么比 RNA-seq 慢很多？

这是**算法本质差异**，不是性能退化：

| 对比项 | RNA-seq step2-check | BS-PDX step2-check (Xenofilx) |
|---|---|---|
| 任务内容 | 表达矩阵 + 剪接分析（轻量） | 双参考宿主过滤 + NM 重算 + 亚硫酸盐分类（计算密集） |
| 输入 | 表达矩阵文本 | 2.91 GB hg38 BAM + 88 MB mm10 BAM |
| 核心计算 | 计数聚合 | 对 7124 万配对读段做全基因组比对特征解析、C-to-T 校正、人/鼠裁决 |

### 5.2 已发现的优化点

- Xenofilx 分类阶段 `--threads 4` 但实际仅用 **~1.85 核**（CPU 185%），存在**串行瓶颈或 I/O 受限**（读取 6.7 GB queryname BAM 时等待磁盘）。
- 排序阶段占用了大量时间（graft queryname 排序 ~41 分钟）。
- 内存完全安全（峰值 31.9 GB，远低于 128 GB 配额）。

### 5.3 待办：下载 BAM 到本地 profiling

用户希望下载 hg38 BAM 到本地测试分析性能瓶颈。本地磁盘有 802G 可用。建议：
1. 下载 `SRR36187610_hg38.bam`（2.9 GB）到本地；
2. 用 `go tool pprof` 分析 Xenofilx 分类阶段的 CPU 热点；
3. 确认是串行瓶颈还是 I/O 受限，评估是否值得优化 `--threads` 或并行化策略。

## 6. 下一步计划

1. **（可选）下载 BAM 到本地 profiling**：分析 Xenofilx 分类阶段性能瓶颈。
2. **推进 step3（甲基化提取）**：在 `run-20260821T112353Z-tunvxb` 上运行 `otter run --phase step3`（BeaverPDX 甲基化提取，20 cores / 64 GiB）。
3. **推进 step3-check**：生成 `qc_summary.xlsx`。
4. **artifact verify / compare**：执行 `otter artifact verify` 及跨版本 `otter artifact compare`。
5. **（可选）创建 legacy-equivalent 项目**：如需 modern/legacy 对比，需为 SRR36187610 创建 `bs-pdx-SRR36187610-legacy-equivalent` 项目（toolchain: legacy-equivalent）。

## 7. 关键约束（必须遵守）

- 只走 **Otter → Craftmake → Slurm 主线**，不手工 sbatch workflow task。
- 修改代码/catalog/配置前先提交已有工作，并在 `docs/notes/2026-08-16-gate6-bs-pdx-craftmake-execution.md` 记录。
- 不触碰无关 dirty 子模块：`enva`、`methx`、`qctb`（及 `inst/rules/methrix_object.smk`）。
- 新数据集已废弃 SRR23802966 主线；SRR23802966 相关任务已 scancel。

## 8. 相关提交记录

| 仓库 | 提交 | 说明 |
|---|---|---|
| 根仓库 | `0aceacf` | docs: handoff Xenofilx sort-memory optimization（上一轮） |
| 根仓库 | `9e026fc` | feat: wire Xenofilx sort memory optimization |
| craftmake | `9667551` | catalog 增加 `--sort-memory` + 编译器测试 |

## 9. 重要参考路径

- 执行记录：`docs/notes/2026-08-16-gate6-bs-pdx-craftmake-execution.md`（已补充 SRR36187610 切换章节）
- 远端获取根：`/public3/home/scg9946/otter-gate6/acquisitions/bs-pdx-SRR36187610-20260821T091900Z/`
- 远端项目：`/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR36187610/`
- 本地 staging：`/home/fallingstar10/.cache/otter-gate6-sra-20260813T041223Z-local-getdown/bs-pdx-SRR36187610/`
- 本地脚本：`scripts/bs-pdx-SRR36187610-*.sh`、`scripts/step{1,2,2-check}-bs-pdx-SRR36187610-controller.sh`
