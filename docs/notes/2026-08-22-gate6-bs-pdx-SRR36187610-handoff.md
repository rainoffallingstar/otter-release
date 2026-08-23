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

### 4.5 Step3（BeaverPDX 甲基化提取）

- Controller `41576841`：**COMPLETED `0:0`，2 小时 19 分**
- 产出（`work/mCall/` 与 `work/bsmap/`）：
  - `SRR36187610_nsort.bismark.cov.gz` (98 MB) 全基因组 CpG 覆盖矩阵
  - `SRR36187610_nsort.bedGraph.gz` (96 MB)
  - `CpG_context_SRR36187610_nsort.txt.gz` (1.7 GB)
  - `Non_CpG_context_SRR36187610_nsort.txt.gz` (4.0 GB)
  - `SRR36187610_nsort.M-bias.txt`、`SRR36187610_nsort_splitting_report.txt`
  - `SRR36187610_nsort.bam` (6.7 GB)

### 4.6 Step3-check（QC 与 Methrix 汇总）

- Controller `41581091`：运行中（节点 `e0301`）。
- `prepare_methrix_reference`、`sample_artifacts`、`species_qc_artifacts` (hg38/mm10)、`bismark_report`、`bismark_summary` 全部已完成；`create_methrix_object`（任务 `41581092`）正在稳健运行中。

### 4.7 其它待跑项目推进状态（最新全部完成 ✅）

1. **3 个 RNA-seq 项目（全部完成并发布校验通过 ✅）**：
   - `human-rnaseq-SRR1039508`：`phase: publish` 完成，`otter artifact verify` 输出 `{"passed": true}`。
   - `human-rnaseq-SRR018258`：`phase: publish` 完成，`otter artifact verify` 输出 `{"passed": true}`。
   - `mouse-rnaseq-SRR037954`：`phase: publish` 完成，`otter artifact verify` 输出 `{"passed": true}`。
2. **人类 RRBS (`human-rrbs-SRR31480456`)（全部完成并发布校验通过 ✅）**：
   - Step 3（甲基化提取）COMPLETED `0:0`；
   - Step 3-Check（Methrix HDF5）COMPLETED `0:0`（产出 `methrix_data.h5`、`assays.h5`、`CpG_annotation_details.tsv.gz`）；
   - `phase: publish` COMPLETED `0:0`，`otter artifact verify` 输出 `{"passed": true}`。
3. **RNA-PDX (`rna-pdx-SRR30880970`)（全部完成并发布校验通过 ✅）**：
   - Step 2-Check（Xenofilx 1.77G 过滤BAM）COMPLETED `0:0`；
   - Step 3（HTSeq 表达定量 + 剪接分类）COMPLETED `0:0`（产出 `SRR30880970_hg38.txt`、`splicing-outcome.json`）；
   - Step 3-Check（表达矩阵 + 质控汇总）COMPLETED `0:0`（产出 `matrix_count.txt`、`matrix_norm.txt`、`qc_summary.xlsx`）；
   - `phase: publish` COMPLETED `0:0`，`otter artifact verify` 输出 `{"passed": true}`。

## 5. 性能分析结论与本地代码优化

### 5.1 为什么比 RNA-seq 慢很多？

这是**算法本质差异**，不是性能退化：

| 对比项 | RNA-seq step2-check | BS-PDX step2-check (Xenofilx) |
|---|---|---|
| 任务内容 | 表达矩阵 + 剪接分析（轻量） | 双参考宿主过滤 + NM 重算 + 亚硫酸盐分类（计算密集） |
| 输入 | 表达矩阵文本 | 2.91 GB hg38 BAM + 88 MB mm10 BAM |
| 核心计算 | 计数聚合 | 对 7124 万配对读段做全基因组比对特征解析、C-to-T 校正、人/鼠裁决 |

### 5.2 本地 Profiling 实验与瓶颈发现

基于本地下载的真实 `SRR36187610_mm10.bam`（88 MB / 188.5 万记录）进行了基准测试与 Go `pprof` CPU 剖析：
- **GC 占用 ~40% CPU**：单片段分类时频繁构建 `map[string]*ReadPair` 和 `map[string]Classification`，产生数十亿次短生命周期堆对象分配（每片段 10 次 malloc，928 B/op）。
- **单线程阶段多**：Graft 与 Host 的输入 BAM 排序原本为串行执行。

### 5.3 已完成的代码级性能优化（已提交）

在 `xenofilx` 子模块（commit `e010d68`）完成了以下深度优化：
1. **零分配单片段分类（Zero-Alloc Group Classifier）**：
   - 实现 `BuildSinglePair` / `BuildSingleRecord`，单片段分类直接在定长结构体中比较 Forward/Reverse 读段得分，**完全消除 map 堆分配**；
   - 基准测试：单次分类从 **1592 ns / 10 allocs / 928 B** 降至 **299.5 ns / 2 allocs / 64 B**（**提速 5.3 倍，内存分配减少 93%**）。
2. **双物种 BAM 并行排序**：
   - 将 Graft 与 Host 输入 BAM 的 Queryname 排序改为并发 goroutine 执行，充分利用多核资源。

## 6. 下一步计划

1. **监控 step3-check 完成**：等待生成 `qc_summary.xlsx` 及 Methrix 产物。
2. **artifact verify / compare**：执行 `otter artifact verify` 及对比。
3. **（可选）创建 legacy-equivalent 项目**：为 SRR36187610 创建 `bs-pdx-SRR36187610-legacy-equivalent` 项目并完成工具链对比。

## 7. 关键约束（必须遵守）

- 只走 **Otter → Craftmake → Slurm 主线**，不手工 sbatch workflow task。
- 修改代码/catalog/配置前先提交已有工作，并在 `docs/notes/2026-08-16-gate6-bs-pdx-craftmake-execution.md` 记录。
- 不触碰无关 dirty 子模块：`enva`、`methx`、`qctb`（及 `inst/rules/methrix_object.smk`）。
- 新数据集已废弃 SRR23802966 主线；SRR23802966 相关任务已 scancel。

## 8. 相关提交记录

| 仓库 | 提交 | 说明 |
|---|---|---|
| xenofilx | `e010d68` | perf: eliminate map allocations in per-fragment classification and sort species in parallel |
| craftmake | `700c592` | fix: tune BS-PDX Xenofilx sort memory contract to 12G |
| 根仓库 | `9b1dd62` | feat: optimize Xenofilx classification and switch BS-PDX to SRR36187610 |

## 9. 重要参考路径

- 执行记录：`docs/notes/2026-08-16-gate6-bs-pdx-craftmake-execution.md`（已补充 SRR36187610 切换章节）
- 远端获取根：`/public3/home/scg9946/otter-gate6/acquisitions/bs-pdx-SRR36187610-20260821T091900Z/`
- 远端项目：`/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/projects/bs-pdx-SRR36187610/`
- 本地数据：`/home/fallingstar10/shire/xdxtools/testdata/pdx-srr36187610/`
- 本地脚本：`scripts/bs-pdx-SRR36187610-*.sh`、`scripts/step{1,2,2-check,3,3-check}-bs-pdx-SRR36187610-controller.sh`
