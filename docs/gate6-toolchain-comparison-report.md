# Gate 6 新旧工具链比较报告

> 更新时间：2026-09-05
>
> **当前状态：Gate 6 comparison scope complete.** 已接受的 Craftmake–Snakemake real-Slurm executor comparison、Gate A–D corrected evidence，以及 Methx/Methrix CpG parity 已完成。正式证据登记见 [`gate6-closeout-evidence-register.json`](gate6-closeout-evidence-register.json)。
>
> 本报告汇总 Gate 6 已获得的真实 Paracloud/Slurm 证据，并明确区分“现代工具链功能验收”“执行器 parity”“工具链 parity”“生产规模放行”和“明确延期工作”。

## 1. 结论摘要

Gate 6 的现代 Craftmake 主线已经在多个真实输入上完成了从 Otter immutable snapshot、Craftmake 计划与 Slurm 执行，到产物校验的闭环。已有历史 RRBS 与 RNA-seq executor-parity 证据通过；BS-PDX 与 RNA-PDX 已完成真实 Slurm 的 PDX executor-parity 有界证据。BS-PDX 还完成了新的 `SRR36187610` 数据集切换、Xenofilx 内存问题规避，以及 Methx 全量 CpG 注释优化和正式 `step3-check` 验收。

当前最重要的性能结论是：Methx 的 CpG 注释瓶颈不是“缺少 GTF”，而是旧实现对每个 CpG 在线性扫描 dense interval bucket。已验证的运行时 binary 使用 binary annotation index；当前本地源码 diff 可审计的核心改动是 sorted interval、prefix-maximum interval query 和 Rayon 并行处理。BS-PDX 7,290,833 个 CpG 的完整运行约 3 分 40 秒；正式 Craftmake `step3-check` 使用同一 v2 index 后约 1 分 55 秒完成。

Corrected Gate D fragment membership comparison has now completed as job `41967198`. Using the same frozen input fragment universe, corrected modern BAM and corrected strand-aware legacy BAM, it reports `44,005,684` graft/graft agreements, `39,106` discarded/discarded agreements, and `1,115,240` disagreements: `1,115,239` modern-only graft and `1` legacy-only graft. This exactly matches the provenance-corrected baseline, so the Bismark NM correction did not change the aggregate Gate D membership result under the current comparison setup.

The replacement stratification job `41972977` completed successfully in `00:06:14` against a deterministic sample of `10,000` modern-only disagreements. All sampled fragments had a complete graft primary pair and no legacy filtered primary records; `9,192` (`91.92%`) had no host primary alignment, while `808` (`8.08%`) had a complete host primary pair. The latter subset requires score-level comparison rather than mapping-state explanation. The first stratification attempt (`41967199`) failed before reading data because of an obsolete evidence-directory prefix; it produced no scientific result. The controller was corrected, and score-decision audit `41973407` completed with no missing score records. Its decision pairs were: `9,190` modern host-absent graft versus legacy host-absent discarded; `808` modern graft-better-total versus legacy threshold-discarded; and `2` discarded under both paths with different reasons. The full mapping stratification `41973923` has now completed successfully over all `1,115,239` modern-only disagreements. It found `1,026,316` host-absent fragments and `88,923` fragments with complete graft and host primary pairs; every fragment was a complete modern graft pair and absent from the legacy filtered output. The dependent full score-decision audit `41973925` also completed successfully with no missing score records: `1,025,732` host-absent modern graft versus legacy discarded, `88,923` modern graft-better-total versus legacy threshold-discarded, and `584` discarded under both paths with different reasons. These categories sum exactly to all `1,115,239` modern-only disagreements, so the Gate D disagreement accounting is closed. The bounded legacy/modern semantic difference is recorded as an accepted scientific limitation in the closeout evidence register; it is not a corrected-source implementation failure.

### 1.1 Corrected-source read-level audit update (2026-09-04)

The corrected-source read-level audit has completed successfully for all three submitted cells:

- RNA human `SRR1039508`: `41965556`, conventional mode, `200,000/200,000` Xenofilx-vs-independent-oracle score equality; Picard conventional NM also matched the conventional oracle for all `200,000` records.
- BS-PDX host mm10: `41965558`, bisulfite mode, `200,000/200,000` Xenofilx-vs-current-contract-oracle equality; explicit `XG=CT` and `XG=GA` control arms each matched the converted-reference oracle for `200,000/200,000` records.
- BS-PDX graft hg38: `41965559`, bisulfite mode, `200,000/200,000` Xenofilx-vs-current-contract-oracle equality; explicit `XG=CT` and `XG=GA` control arms each matched their converted-reference oracle for `200,000/200,000` records.

The BS-PDX host and graft comparisons record the expected distinction between conventional and strand-aware bisulfite NM: conversion-compatible bases create large conventional-versus-bisulfite differences, while Xenofilx matches the corrected bisulfite oracle exactly. This is expected behavior, not an audit failure. Gate C read-level acceptance is complete for the audited RNA and BS-PDX cells.

A dedicated RNA-seq mixture pilot has now been completed using human `SRR1039508` and mouse `SRR037954`, with STAR alignment to ordinary hg38/mm10 references and conventional reference-based NM. The 50:50 pilot used two independent 1,000,000-fragment replicates and evaluated both graft directions. Corrected jobs `41929279`–`41929282` all completed with exit `0`.

Xenofilx and XenofilteR produced identical fragment selections in every pilot cell. Across replicates, human-graft recall was `96.7456–96.7656%`, while mouse-graft recall was `75.5026–75.5986%`; host specificity was `99.2142–99.2214%` and `99.9854–99.9878%`, respectively. This establishes an RNA direction asymmetry but no modern-versus-legacy disagreement in the 50:50 pilot. The results are archived under `/public3/home/scg9946/otter-gate6/evidence/gate6-human-mouse-rna-mixtures-20260903/` and remain algorithmic benchmark evidence, not full seven-input toolchain acceptance.

The composition convergence extension is now complete for human fractions `0.60`, `0.70`, `0.80`, `0.90`, `0.95`, and `0.99`, with two replicates and both graft directions at every fraction. All 24 analysis jobs completed with exit `0`. Mean performance by fraction and direction was:

| Human fraction | Direction | Graft recall | Host specificity | Graft precision | F1 |
|---:|---|---:|---:|---:|---:|
| 0.60 | human graft | 96.7558% | 99.2111% | 99.4594% | 98.0890% |
| 0.60 | mouse graft | 75.5462% | 99.9876% | 99.9754% | 86.0608% |
| 0.70 | human graft | 96.7620% | 99.2172% | 99.6545% | 98.1869% |
| 0.70 | mouse graft | 75.5323% | 99.9879% | 99.9625% | 86.0470% |
| 0.80 | human graft | 96.7690% | 99.1972% | 99.7930% | 98.2578% |
| 0.80 | mouse graft | 75.5005% | 99.9881% | 99.9371% | 86.0169% |
| 0.90 | human graft | 96.7606% | 99.2100% | 99.9094% | 98.3098% |
| 0.90 | mouse graft | 75.5985% | 99.9879% | 99.8560% | 86.0504% |
| 0.95 | human graft | 96.7585% | 99.2040% | 99.9567% | 98.3316% |
| 0.95 | mouse graft | 75.6840% | 99.9877% | 99.6931% | 86.0450% |
| 0.99 | human graft | 96.7628% | 99.2250% | 99.9919% | 98.3508% |
| 0.99 | mouse graft | 75.4300% | 99.9878% | 98.4212% | 85.4053% |

Fragment counts and confusion matrices matched between Xenofilx and XenofilteR in all 24 gradient reports. The two tools' output BAM record counts can differ because of mate-record representation, but their evaluated fragment membership is identical. Human-graft recall remained essentially flat around `96.76%`; mouse-graft recall remained around `75.5%–75.7%` through `0.95` and was `75.43%` at `0.99`. The data therefore do not show a strong composition-driven recall collapse. The principal finding remains a stable direction-specific asymmetry, with high host specificity in both directions.

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
| Executor parity | 输入、参考、参数、资源、toolchain、run identity contract | Craftmake vs explicit Snakemake compatibility | RRBS、RNA-seq 和 PDX 的 accepted real-Slurm paired evidence 已完成；该比较范围已接受 |
| Toolchain parity | executor、输入、参考、参数和资源 | modern vs legacy-equivalent implementation | 当前接受范围不要求重新执行七输入 fresh legacy-equivalent matrix；该矩阵未运行并登记为 deferred limitation |

## 3. 真实运行证据

| 场景/输入 | 现代路径当前状态 | 旧路径/比较状态 | 结论 |
|---|---|---|---|
| Human RRBS `SRR31480456` | step3 与 step3-check 完成，Methx HDF5/annotation/QC 产物生成，publish/verify 完成 | 历史 RRBS clean executor-parity 与 semantic evidence 已通过 | RRBS executor parity 已有 accepted evidence；当前输入现代路径完成 |
| Mouse RRBS `SRR10025242` | 已纳入 immutable corpus，step1 decode/verifier evidence 完成 | fresh seven-input legacy-equivalent matrix 未执行，已登记为 deferred limitation | 不影响当前 Gate 6 scope closeout |
| Human RNA-seq `SRR1039508` | step1、step2、step2-check、publish 完成，artifact verify 通过 | RNA r31 clean Craftmake/Snakemake pair 已通过 compare | RNA executor parity accepted |
| Human RNA-seq `SRR018258` | 完成并通过 publish/verify；曾有 FastQC walltime recovery incident | RNA r31 recovery/compatibility evidence 已保留 | recovery incident 已分类，不是科学失败 |
| Mouse RNA-seq `SRR037954` | 完成并通过 publish/verify | 与 RNA executor-parity 证据相容 | 当前现代路径完成 |
| RNA-PDX `SRR30880970` | step2-check、step3、step3-check、publish 完成，artifact verify 通过 | 真实 Slurm PDX executor-parity 有界 evidence，step2-check 三次 paired scheduler repeats | 有界 parity 通过；性能结论仅为描述性 |
| BS-PDX `SRR36187610` | step1/step2/step2-check/step3 完成；正式 step3-check `controller 41687475` exit 0 | 真实 Slurm PDX paired evidence；Gate A–D corrected evidence 与 Methx/Methrix parity 已完成 | 当前接受范围通过；仅保留 closeout documentation |
| WGBS `SRR6373947` | deferred | 未提交 | 明确延期，不阻塞当前 Gate 6 closeout |

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

## 7. BAM, NM and PDX classification parity prerequisite

The corrected Gate A–D prerequisite work is complete for the accepted Gate 6 scope. The historical prerequisite plan remains linked for provenance, but it is no longer a blocker for the current closeout. Gate A and Gate B establish real-BAM and pair preservation; Gate C establishes corrected NM/oracle agreement on the audited RNA and BS-PDX cells; Gate D completes fragment-level disagreement accounting with zero unexplained disagreement after stratification.

The prerequisite gates are:

| Gate | Subject | Accepted evidence |
|---|---|---|
| A | bamdriver decode/encode | canonical record, header, auxiliary-field, and structural preservation on frozen real BAMs |
| B | pairbam pair handling | retained-record preservation plus complete fragment/mate accounting on the frozen host BAM |
| C | Xenofilx NM | independent CIGAR/SEQ/reference oracle, conventional RNA audit, bisulfite audit, and CT/GA controls |
| D | PDX classification | corrected modern-versus-legacy fragment comparison with complete mapping and score-decision reason partition |

Picard remains a legacy implementation baseline, not NM ground truth. The accepted evidence binds the independent oracle, corrected source contract, input/reference identity, and Slurm jobs rather than relying on output-tag comparison alone.

### 7.1 Per-read NM audit evidence, 2026-08-27

The per-read audit uses a 200,000 raw-record prefix for each frozen BAM, preserves stable alignment identity (`QNAME`, `FLAG`, `RNAME`, zero-based `POS`, `CIGAR`, duplicate occurrence), and records every input/reference checksum in the source manifest. The standalone oracle is the NM authority for these checks; Picard is an implementation baseline.

| Dataset / evidence | Conventional NM result | Xenofilx result | Picard bisulfite result |
|---|---|---|---|
| Human RNA `SRR1039508`, source job `41712621`, postprocess `41714006` | Picard conventional equals the independent conventional oracle for `200,000 / 200,000` mapped records | Xenofilx non-bisulfite recalculation and score equal the conventional oracle for `200,000 / 200,000`; zero missing identities | Not applicable |
| BS-PDX host mm10, source job `41713979`, strand postprocess `41725232` | Picard conventional equals the independent conventional oracle for `200,000 / 200,000` records | Xenofilx `--bisulfite` NM and score equal the current-contract oracle for `200,000 / 200,000`; conventional NM total `2,643,344`, Xenofilx bisulfite NM total `50,287`, conversion-compatible positions `2,593,057` | Picard `IS_BISULFITE_SEQUENCE=true` equals the current Xenofilx/oracle bisulfite NM for `97,571 / 200,000` records |
| BS-PDX graft hg38, source job `41715084`, strand postprocess `41725231` | Picard conventional equals the independent conventional oracle for `200,000 / 200,000` records | Xenofilx `--bisulfite` NM and score equal the current-contract oracle for `200,000 / 200,000`; conventional NM total `3,028,837`, Xenofilx bisulfite NM total `28,951`, conversion-compatible positions `2,999,886` | Picard `IS_BISULFITE_SEQUENCE=true` equals the current Xenofilx/oracle bisulfite NM for `99,219 / 200,000` records |

The RNA result closes the conventional NM recalculation question for the audited prefix: Picard conventional, the independent oracle, and Xenofilx conventional all agree exactly.

For both BS-PDX references, the large conventional-to-bisulfite reduction is a measured conversion-semantic difference, not an Oracle/Xenofilx implementation mismatch. Picard conventional remains exactly equal to conventional oracle NM. Picard's built-in bisulfite option matches the Xenofilx current contract for approximately half of the reads but not all reads; the discrepancy must therefore remain an explicit semantic difference rather than a generic Picard error claim.

The control also ran Picard after splitting records by `FLAG 0x10`, using C-to-T-converted reference for forward records and G-to-A-converted reference for reverse records. Together, the two arms cover the full `200,000` record audit prefix for each reference; each arm has its own strand-specific identity universe. Picard NM equaled the converted-reference conventional oracle for every record in each arm on both host and graft.

The identity-matched CT/GA cross-control completed as host job `41728385` and graft job `41728384`. It compares each strand-specific converted-reference Picard/Oracle/Xenofilx value to the original-reference Xenofilx current-contract bisulfite NM for the same record. The results establish that the two semantics are substantially different:

- host forward CT: `72,547 / 100,038` differ; host reverse GA: `72,880 / 99,962` differ;
- graft forward CT: `81,816 / 100,032` differ; graft reverse GA: `81,957 / 99,968` differ.

Within every control arm, Picard, the converted-reference oracle and Xenofilx evaluated on that converted reference agree with each other. The observed cross-control differences therefore come from the **reference and conversion semantics**, not from Picard's NM serialization or Xenofilx's reference traversal. Xenofilx's current original-reference contract ignores both reference-C/read-T and reference-G/read-A without conditioning on `FLAG 0x10`; the strand-specific CT/GA control represents a different contract. Both are now measured explicitly and must not be treated as interchangeable.

Accordingly, Gate C passes for the **current Xenofilx implementation contract** on all audited RNA, host and graft samples. Gate D is also closed for the accepted scope: its corrected comparison has a complete fragment universe, zero missing score records, and zero unexplained disagreement after the full mapping and score-decision partition. The remaining difference is a documented legacy/modern semantic limitation, not an unresolved implementation failure.

Gate A now also passes on the full frozen real BAMs. Slurm job `41695184` completed with exit `0` in 39m30s; decoded headers and ordered canonical record streams matched for all 90,320,060 graft and 1,885,462 host records, and both round-trip outputs passed `samtools quickcheck`. Gate B passes on the frozen mm10 host BAM: Slurm job `41696363` completed with exit `0` in 4m43s; pairbam retained all 942,731 complete primary fragments / 1,885,462 records, and the native-name-sort-aware comparator found equal retained canonical records, mate pointers/TLEN, reference declarations, and filtered-name manifest. The input had no incomplete fragments, so a mixed synthetic selection fixture remains a coverage enhancement rather than a real-data failure. Gate D fixture preflight passed as Slurm job `41696471`, and the full direct Picard + XenofilteR legacy run completed successfully as job `41696495` in 27m24s: Picard preserved all graft/host input record counts and XenofilteR emitted 6,434,546 filtered graft records. The first dependent membership comparison `41696588` is invalid for acceptance because its modern BAM resolved to prior run `tunvxb` rather than frozen `ndcwfa`; it reported 41,892,990 differences and is retained as a provenance incident. Corrected direct modern Xenofilx job `41696880` and dependent comparison `41696881` used the frozen `ndcwfa` original BAMs and confirmed the same physical-input `41,892,990` fragment-membership disagreements. They establish that Gate D is a real classification-semantics gap, not a provenance mismatch; the per-fragment reason partition remains required before closure.

## 8. Strand-aware NM source correction (2026-08-27)

源码审计确认此前 bisulfite NM 实现对每条 read 无条件同时豁免 `reference C -> read T` 和 `reference G -> read A`，没有检查 SAM `FLAG 0x10`。这使原始参考上的 Xenofilx bisulfite 语义不同于 forward/CT 与 reverse/GA 转换参考控制。

已修改 `bamdriver/pkg/bamnative/nmtag.go`：

- forward read（`FLAG 0x10` 未设置）仅豁免 `C -> T`；
- reverse read（`FLAG 0x10` 已设置）仅豁免 `G -> A`；
- `M` 与 bisulfite 模式的 `X` 都执行相同的链特异转换判断；
- conventional 模式的 `X` 保持按 CIGAR 长度计入 NM，不受转换规则影响。

独立 `cmd/nmoracle` 已同步为同一科学契约，但仍保留独立的 CIGAR/SEQ/reference 遍历实现。新增回归测试覆盖正反链、`M`/`X`、转换与非转换错配，以及 conventional `X` 的标准计数；Xenofilx classifier 还增加了通过其 reference-aware wrapper 验证共享 bamdriver 实现的测试。

本地已通过 `bamdriver` 的 `pkg/bamnative` 与 `cmd/nmoracle` 定向测试、`bamdriver` 全量测试（此前修复完成阶段），以及 Xenofilx `internal/classifier` 定向测试。后续必须使用包含本修复的 bamdriver/Xenofilx binary 重跑 CT/GA strand-specific audit、转换参考 cross-control、RNA/BS read audit，并在此之后重新评估 Gate D 的 fragment disagreement。此前记录的 `41,892,990` disagreement 不能沿用为修复后的结果。

最新会话状态：用户已授权覆盖远端 runtime 的 `gate6-nmoracle` 与 `gate6-xenofilx-scoreaudit`（先备份、记录 SHA-256、全新证据目录），但本地命令通道持续无响应，构建/上传/Slurm 提交均未发生。详细接续步骤见 `docs/notes/2026-08-27-gate6-bisulfite-nm-strand-aware-fix-handoff.md`。

## 9. Code and documentation scope

本次提交包含：

- 根仓库：Gate6 controller、Methrix workflow rule、publish/check controller 和比较/交班文档；
- `methx`：内存 annotation interval index、sorted interval/prefix-max query、并行 CpG 注释和 FASTA/CpG extraction robustness 与测试；
- `enva`：Rust Bismark 3.1.0 环境声明、Bowtie2 外部 pinned asset contract、Conda run separator 修正；
- `qctb`：immutable `otter.run/v1` / `--config-dir` 读取、canonical paths 派生和严格校验。

BS-PDX benchmark 所使用的 persistent binary-index serialization/CLI wiring 来自远端 staged runtime；它已作为运行时证据记录，但不应在本次 git commit 中声称为当前 `methx` 子仓库源码功能，除非后续把对应 source revision 同步进来。远端运行时的 BS-PDX 显式 index wrapper 是一次性 compatibility injection；它不改变 catalog digest，也不代表把 run-local wrapper 作为科学产物发布。

## 10. 当前放行状态与 closeout boundary

### 已通过

- 现代 Craftmake 路径在真实 Slurm 上的多场景 canary；
- 已接受的 RRBS、RNA-seq 和 PDX Craftmake–Snakemake executor-parity evidence；
- Gate A BAM round-trip preservation；
- Gate B pairbam fragment and mate integrity；
- Gate C corrected read-level NM/oracle and CT/GA control audits；
- Gate D corrected fragment comparison and complete disagreement accounting；
- BS-PDX `SRR36187610` Methx 全量 annotation benchmark 与正式 `step3-check`；
- Methx 与 Methrix production `stranded=TRUE, collapse_strands=TRUE` CpG matrix parity；
- Methx custom HDF5 到 native Methrix HDF5 的 R-only export and reload validation。

### 当前仅剩 closeout 项

- `gate6-closeout-evidence-register.json` 的正式归档与链接核验；
- Gate 6 decision log / final scope record；
- BS-PDX publish 与完整 artifact manifest verification，仅在本阶段要求正式发布时执行。

### 明确延期、非阻塞的扩展

以下项目保留为未来授权的 extension register，不属于当前 Gate 6 acceptance blockers：

- 七输入 fresh legacy-equivalent scientific matrix（未执行，绝不表示已完成）；
- representative `20 samples × 3 repeats`；
- production-scale scheduler-pressure qualification；
- WGBS `SRR6373947` reference/acquisition requalification；
- 额外 Snakemake PDX interruption/retry/recovery scientific comparison。

本报告保留历史 evidence 的 incident、baseline 和 limitation 描述；延期不等于失败，也不改变已接受的 Craftmake–Snakemake comparison scope。
