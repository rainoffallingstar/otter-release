# Gate 6 新旧工具链比较报告

> 更新时间：2026-08-26
>
> 本报告汇总 Gate 6 当前已经获得的真实 Paracloud/Slurm 证据，并明确区分“现代工具链功能验收”“执行器 parity”“工具链 parity”“生产规模放行”。单次 canary 或单个数据集的成功，不自动等同于代表性生产 acceptance。

## 1. 结论摘要

Gate 6 的现代 Craftmake 主线已经在多个真实输入上完成了从 Otter immutable snapshot、Craftmake 计划与 Slurm 执行，到产物校验的闭环。已有历史 RRBS 与 RNA-seq executor-parity 证据通过；BS-PDX 与 RNA-PDX 已完成真实 Slurm 的 PDX executor-parity 有界证据。BS-PDX 还完成了新的 `SRR36187610` 数据集切换、Xenofilx 内存问题规避，以及 Methx 全量 CpG 注释优化和正式 `step3-check` 验收。

当前最重要的性能结论是：Methx 的 CpG 注释瓶颈不是“缺少 GTF”，而是旧实现对每个 CpG 在线性扫描 dense interval bucket。已验证的运行时 binary 使用 binary annotation index；当前本地源码 diff 可审计的核心改动是 sorted interval、prefix-maximum interval query 和 Rayon 并行处理。BS-PDX 7,290,833 个 CpG 的完整运行约 3 分 40 秒；正式 Craftmake `step3-check` 使用同一 v2 index 后约 1 分 55 秒完成。

这份证据支持将现代 Methx 注释实现作为 Gate 6 的候选生产实现继续推进，但尚不足以关闭 representative 20-cell × 3-repeat、全场景新旧工具链 scientific parity、WGBS 和 scale gates。

## 2. 比较边界

### 2.1 现代工具链

```text
Otter immutable run.yaml
  -> Craftmake planner/controller
  -> Slurm
  -> fastqcx / Rust Bismark 3.1.0 / xenofilx / methx / seq2mat / matsrun / qctb
```

现代路径的关键契约包括：

- 输入、参考、工具、资源和 workflow assets 由 immutable `otter.run/v1` 固化；
- Craftmake 是生产 workflow 的计划、提交、resume、状态和报告控制面；
- BS-PDX/RNA-PDX 使用 Xenofilx 生成固定名称的 filtered BAM/BAI；
- Bismark Rust 3.1.0 和 Bowtie2 2.5.4 是共享的 checksum-pinned runtime 依赖；
- 不把 Perl-versus-Rust Bismark 当作本报告的比较轴；
- Methx 输出使用版本化 `methx.custom-hdf5` schema，并同时产生 HDF5、QC 和 annotation 产物。

### 2.2 旧工具链 / legacy-equivalent

旧路径表示既有 workflow 语义和兼容任务映射，例如 legacy FastQC/SeqKit QC、旧式目录/配置投影、Methrix 兼容输出和 XenofilteR 兼容工作流路径。它用于验证 workflow 语义和执行器兼容性，不代表可以把旧 Perl Bismark 与新 Rust Bismark 的差异归因于 Craftmake 或 Otter。

### 2.3 两个独立比较轴

| 轴 | 固定内容 | 改变内容 | 当前证据 |
|---|---|---|---|
| Executor parity | 输入、参考、参数、资源、toolchain、run identity contract | Craftmake vs explicit Snakemake compatibility | RRBS 和 RNA-seq 已有 accepted clean pair；PDX 有界真实 Slurm paired evidence |
| Toolchain parity | executor、输入、参考、参数和资源 | modern vs legacy-equivalent implementation | 仍需扩大到当前七输入 corpus 的完整 fresh paired scientific comparison |

## 3. 真实运行证据

| 场景/输入 | 现代路径当前状态 | 旧路径/比较状态 | 结论 |
|---|---|---|---|
| Human RRBS `SRR31480456` | step3 与 step3-check 完成，Methx HDF5/annotation/QC 产物生成，publish/verify 完成 | 历史 RRBS clean executor-parity 与 semantic evidence 已通过 | RRBS executor parity 已有 accepted evidence；当前输入现代路径完成 |
| Mouse RRBS `SRR10025242` | 已纳入七输入 immutable corpus，step1 decode/verifier evidence 完成 | 完整 fresh modern/legacy matrix 仍待整理 | 可继续作为 RRBS 扩展样本，不能单独关闭 representative gate |
| Human RNA-seq `SRR1039508` | step1、step2、step2-check、publish 完成，artifact verify 通过 | RNA r31 clean Craftmake/Snakemake pair 已通过 compare | RNA executor parity accepted；需避免把历史 pair 当作当前所有输入的 toolchain parity |
| Human RNA-seq `SRR018258` | 完成并通过 publish/verify；曾有 FastQC walltime recovery incident | RNA r31 recovery/compatibility evidence 已保留 | recovery incident 已分类，不是科学失败 |
| Mouse RNA-seq `SRR037954` | 完成并通过 publish/verify | 与 RNA executor-parity 证据相容 | 当前现代路径完成 |
| RNA-PDX `SRR30880970` | step2-check、step3、step3-check、publish 完成，artifact verify 通过 | 真实 Slurm PDX executor-parity 有界 evidence，step2-check 三次 paired scheduler repeats | 有界 parity 通过；性能结论仅为描述性 |
| BS-PDX `SRR36187610` | step1/step2/step2-check/step3 完成；正式 step3-check `controller 41687475` exit 0 | 真实 Slurm PDX paired evidence；完整 legacy-equivalent scientific pair 尚未完成 | 当前 Methx 与现代 PDX 路径验收通过；publish/完整 toolchain parity 仍开放 |
| WGBS `SRR6373947` | deferred | 未提交 | 需要重新确认 primary reference 与 acquisition provenance |

## 4. BS-PDX 数据集切换与 Xenofilx 结果

原始 `SRR23802966` 是约 387,683,390 对 reads 的大型 WGBS 输入，hg38 BAM 约 29.3 GiB。Xenofilx 在共驻节点上经历约 6.5 小时后被 OOM killer 终止；该事件不能简单归因于单个进程的 sort-memory 参数。

新主线使用 genuine PDX eRRBS 数据集 `SRR36187610`：

- 配对 reads：71,242,412；
- 参考：graft `hg38@GRCh38-gencode-v44`，host `mm10@GRCm38-gencode-M25`；
- archive：2,424,481,661 bytes；
- Step2 双参考 mapping、Step2-check Xenofilx 和 Step3 methylation 均成功；
- Step2-check Xenofilx 约 1 小时 46 分完成，峰值 RSS 约 31.9 GiB；
- 原始大型输入保留为已分类 incident evidence，不再作为 BS-PDX 主验收输入。

此前完成的 Xenofilx zero-allocation classifier 和双物种并行排序优化属于独立子仓库提交，不应与 Methx CpG annotation benchmark 混为同一性能数字。

## 5. Methx CpG 注释优化结果

### 5.1 旧瓶颈

旧注释实现虽然按 chromosome bucket 组织 GTF 区间，但每个 CpG 查询仍会线性扫描对应 bucket 的全部候选区间。对于 hg38 dense buckets 和数百万 CpG，这使注释阶段成为主要耗时，并导致旧 BS-PDX `step3-check` 在 8 小时内没有生成最终产物。

### 5.2 当前实现与部署边界

当前源码中的 annotation 查询实现包括：

- 按 chromosome 构建内存索引；
- bucket 内按区间排序；
- 维护 prefix maximum end；
- 查询时先按 start 二分，再跳过不可能重叠的前缀；
- 保持半开区间 `[start, end)`、deterministic transcript tie-breaking 和原 priority/distance 语义；
- 使用 Rayon 并行处理 CpG 查询，`--threads` 控制处理线程。

BS-PDX benchmark 使用的 staged/production binary 还加载了版本化 v2 `.annotation-index.bin`，并通过 `--annotation-index` 或目录自动发现路径完成正式运行。需要注意：当前本地 `methx` 子仓库工作树的待提交 diff 主要覆盖内存 interval-query/FASTA 提取实现；如果 persistent binary-index serialization/CLI wiring 来自独立远端 staged source revision，则必须在后续 release 中确保该 source revision 与此 git commit 对齐，不能仅凭运行时 binary 证据推断本地源码已经包含全部 index 功能。

因此，本报告把 binary index benchmark 作为已部署运行时证据记录，同时把本地源码可审计的 interval-query 改动与持久化 index 功能分开标注。

### 5.3 全量 BS-PDX benchmark

| 指标 | 结果 |
|---|---:|
| CpG 输入 | 7,290,833 |
| annotation details 行数 | 7,290,834（含表头） |
| binary index 构建时间 | 38.15 秒 |
| binary index 大小 | 432,934,757 bytes |
| 完整 process wall time | 约 3 分 40 秒 |
| process exit | 0 |
| 完整 process 峰值 RSS | 约 9.51 GiB |
| HDF5 shape | `(1, 7290833)` |
| 产物 | 5/5 全部生成 |

五类产物为：

```text
assays.h5
methrix_data.h5
CpG_annotation_details.tsv.gz
CpG_annotation_report.xlsx
CpG_coverage.xlsx
```

正式 Craftmake checker 通过 controller `41687475` 和 child `41687489` 验收：

- controller exit `0`；
- controller elapsed 2 分 21 秒；
- Methx child elapsed 1 分 55 秒；
- annotation details 行数 `7,290,834`；
- `beta`/`cov` shape 均为 `(1, 7290833)`；
- `assays.h5` 与 `methrix_data.h5` SHA-256 相同：
  `eef88a4de5329de2772cf8f2dacc3fb4b3b64f0c6cb52372cf9d3e930e6d0444`。

正式 checker 的第一次异常运行停留在 `Preparing CpG annotation report`，随后通过 controller 注入显式 v2 index 路径而成功。该兼容注入不修改 immutable catalog 或 run snapshot，避免了 workflow digest drift。

## 6. 旧实现与新实现的性能解释

| 项目 | 旧实现/旧事件 | 新实现/当前证据 |
|---|---|---|
| GTF 使用方式 | 每个 CpG 在线性扫描 dense bucket | 预处理 binary index + sorted interval + prefix max end |
| 索引载入 | 大型 RON 反序列化，约 3.29 GB | binary index，约 432.9 MB |
| BS-PDX Methx checker | 8 小时内无最终产物 | 正式 child 约 1 分 55 秒完成 |
| 7.29M CpG 隔离 benchmark | 不可接受/未完成 | 约 3 分 40 秒，exit 0 |
| 结论 | 主要是区间查询算法和索引载入成本 | 已达到当前 canary 的合理运行时间 |

这些数值不能直接表述为新旧生物学工具链的总体 speedup：输入、索引格式和查询算法同时发生变化，且正式 checker 与隔离 benchmark 的调度环境不同。它们只能支持 Methx 注释实现的 bounded performance acceptance。

## 7. 代码与文档变更范围

本次提交包含：

- 根仓库：Gate6 controller、Methrix workflow rule、publish/check controller 和比较/交班文档；
- `methx`：内存 annotation interval index、sorted interval/prefix-max query、并行 CpG 注释和 FASTA/CpG extraction robustness 与测试；
- `enva`：Rust Bismark 3.1.0 环境声明、Bowtie2 外部 pinned asset contract、Conda run separator 修正；
- `qctb`：immutable `otter.run/v1` / `--config-dir` 读取、canonical paths 派生和严格校验。

BS-PDX benchmark 所使用的 persistent binary-index serialization/CLI wiring 来自远端 staged runtime；它已作为运行时证据记录，但不应在本次 git commit 中声称为当前 `methx` 子仓库源码功能，除非后续把对应 source revision 同步进来。远端运行时的 BS-PDX 显式 index wrapper 是一次性 compatibility injection；它不改变 catalog digest，也不代表把 run-local wrapper 作为科学产物发布。

## 8. 当前放行状态与下一步

### 已通过

- 现代 Craftmake 路径在真实 Slurm 上的多场景 canary；
- 历史 RRBS clean executor parity 与 semantic comparison；
- RNA-seq r31 clean executor parity、artifact verify/compare；
- PDX 三次 paired `step2-check` scheduler evidence（描述性）；
- BS-PDX `SRR36187610` Methx 全量 annotation benchmark；
- BS-PDX 正式 `step3-check` 五类产物、行数和 HDF5 shape 验收。

### 仍未关闭

- 七输入完整 modern vs legacy-equivalent fresh toolchain parity；
- representative `20 samples × 3 repeats`；
- production-scale scheduler-pressure gate；
- WGBS `SRR6373947` reference requalification；
- BS-PDX publish 与完整 artifact manifest verification（若本阶段需要发布）；
- 真正的全流程 Snakemake PDX interruption/retry 和 scientific comparison。

下一位 agent 应先读取本报告、`docs/notes/2026-08-26-gate6-final-handoff.md` 和各仓库最近 commit，再继续执行未关闭 gate，不要重新提交已完成的 BS-PDX Methx benchmark 或重复创建新的 binary index。
