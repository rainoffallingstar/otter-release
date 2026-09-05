# Gate 6 BAM 保真、NM 与 PDX 分类 Parity 计划

> 状态：执行中
>
> 创建日期：2026-08-26
>
> 主验证输入：BS-PDX `SRR36187610`。本计划是七输入 modern vs legacy-equivalent 全流程科学比较的前置门。

## 1. 目的与边界

在比较现代和 legacy-equivalent 工具链之前，必须先把四个问题分开证明：

1. `bamdriver` 的 decode/encode 不改变 BAM header、record 或 auxiliary fields；
2. `pairbam` 的读写、过滤和配对处理不破坏 record 内容或 mate relationship；
3. Xenofilx `--recalculate-nm --bisulfite` uses reference-aware NM recomputation to derive the **classification score**; it does not currently rewrite the selected BAM record's `NM` auxiliary field;
4. Xenofilx's reference-aware score and fragment classification are compared with an independent oracle and legacy `Picard NM patch + XenofilteR` results.

原始 BAM 的文件 SHA-256 不用作 round-trip 唯一验收条件，因为 BGZF block 布局、压缩字节、排序、`@PG` 和 BAI 可以合法不同。验收对象是 canonical record 内容、pair relationship、允许变更白名单，以及 fragment-level classification。

## 2. 冻结输入与版本

| 项目 | 值 |
|---|---|
| 项目 | `bs-pdx-SRR36187610` |
| immutable run | `run-20260823T034624Z-ndcwfa` |
| graft input | `work/bsmap/SRR36187610_hg38.bam` + BAI |
| host input | `work/bsmap/SRR36187610_mm10.bam` + BAI |
| modern filtered output | `work/bsmap/Filtered_bams/SRR36187610_fixed_hg38_Filtered.bam` + BAI |
| graft reference | `hg38@GRCh38-gencode-v44` |
| host reference | `mm10@GRCm38-gencode-M25` |
| Xenofilx production binary | `0.1.0-direct-bamdriver-region-r34` at initial collection; source gitlink is separately recorded |
| Pairbam production binary | `0.1.0-direct-bamdriver-region-r34` at initial collection |
| NM configuration | `--recalculate-nm --bisulfite --mm-threshold 6 --unmapped-penalty 8` |

Each comparison manifest must additionally record tool paths, `--version` output, command line, input/reference SHA-256, header text, Slurm job and step identifiers, and output checksums.

## 3. Gate A: BAM decode/encode preservation

### A1. Synthetic fixture coverage

The test fixture set must include: primary paired alignments, orphaned pairs, unmapped mate, secondary/supplementary records, duplicate flags, soft and hard clipping, insertion/deletion, `=`/`X` CIGAR operations, reverse reads, repeated QNAMEs, `SA`/`MC`/`MD`/`NM` tags, missing NM, incorrect NM, nil qualities, and non-scalar auxiliary arrays.

`bamdriver` already has unit coverage for headers, IUPAC bases, qualities and array auxiliary fields. This gate adds a full canonical-record comparator rather than assuming serializer tests prove real BAM preservation.

### A2. Real BAM round-trip

For each original hg38 and mm10 BAM:

```text
input BAM -> bamdriver reader/writer no-op -> roundtrip BAM
```

The comparator must match all records by a stable alignment identity, not QNAME alone. It must include QNAME, read1/read2/secondary/supplementary bits, reference, position, CIGAR and an occurrence ordinal for duplicate alignment identities.

The required equality fields are:

```text
QNAME FLAG RNAME POS MAPQ CIGAR RNEXT PNEXT TLEN SEQ QUAL
all auxiliary fields and types
```

Allowed changes must be explicitly declared before execution. The default allowlist is only BGZF layout, BAI bytes, a tool-added `@PG`, and intentionally recalculated tags in a designated NM test. Any other changed field is a failure.

### A3. Structural checks

For original and generated BAMs record:

```text
samtools quickcheck -v
samtools view -H
samtools idxstats
samtools flagstat
samtools stats
samtools view -c
```

Pass condition:

```text
unexpected record loss = 0
unexpected record duplication = 0
unexpected record field mutation = 0
BAM/BAI structural failure = 0
```

## 4. Gate B: Pairbam preservation and pair integrity

Pairbam is assessed separately from bamdriver because it operates on paired records and may reorder or select records.

For every retained output record, compare the same canonical alignment fields from Gate A. At fragment level, construct a canonical object from QNAME plus occurrence ordinal and both mate records. Verify:

- each retained fragment resolves to exactly the expected mate records;
- read1/read2 identities are unchanged;
- mate reference, mate position and template length are unchanged unless a declared pair repair is under test;
- secondary/supplementary membership is unchanged;
- pairbam selection is represented as an explicit subset, never a silent mutation.

Pass condition:

```text
unexpected fragment loss/split/duplication = 0
unexpected mate-pointer or TLEN mutation = 0
unexpected retained-record mutation = 0
```

## 5. Gate C: Xenofilx reference-aware score correctness

Picard is a legacy baseline, not NM ground truth. The source of truth is an independently implemented oracle that does not import or call Xenofilx, bamdriver or pairbam NM calculation code. It may use bamdriver's BAM/FASTA reader types only for file decoding and reference retrieval; the CIGAR traversal and scoring logic are implemented independently.

`--recalculate-nm` currently recalculates an in-memory reference-aware value for classification. The selected graft record is then written without changing its `NM` auxiliary tag. Consequently, the real-BAM validation cannot infer Xenofilx's recalculated value by reading the output BAM `NM:i`; the execution harness must emit a per-alignment score audit before records are written.

### C1. Compared values

For every mapped alignment retain:

```text
NM_original                 # original BAM auxiliary tag
NM_picard                   # legacy Picard-patched auxiliary tag
NM_oracle_conventional      # mismatch + inserted bases + deleted bases
NM_oracle_bisulfite         # conversion-aware variant of conventional NM
Xenofilx_recalculated_nm    # instrumented in-memory value
Xenofilx_classification_score
Oracle_xenofilx_score
```

The independent oracle consumes reference FASTA, POS, CIGAR, SEQ and FLAG. It reports mismatch, insertion, deletion and soft-clip contributions separately. The oracle must reproduce the currently implemented score contract explicitly:

```text
Xenofilx_recalculated_nm = mismatch + insertion + deletion
Xenofilx_classification_score = Xenofilx_recalculated_nm + insertion + soft_clip
```

This is deliberately separate from conventional NM because the current classifier adds insertion length after calling the recalculator. The audit must expose that component rather than hide it behind a single `NM` column. Soft/hard clips, reverse orientation, `M`/`=`/`X`, skipped regions and supplementary records are handled explicitly.

### C2. Bisulfite contract

The source implementation now uses the accepted strand-aware contract when `--bisulfite` is enabled:

- forward reads (`FLAG 0x10` absent) ignore only reference `C` to read `T`;
- reverse reads (`FLAG 0x10` present) ignore only reference `G` to read `A`;
- `M` and bisulfite-mode `X` inspect reference/read bases for that conversion rule;
- conventional-mode `X` remains an unconditional mismatch operation counted by length.

Before the source correction, the implementation excluded both reference `C` to read `T` and reference `G` to read `A` for every read and did not condition the rule on strand. The earlier audit therefore measured a historical/current-contract baseline, not the corrected scientific contract. Those reports remain useful as diagnosis evidence but cannot serve as post-fix acceptance.

The independent oracle consumes reference FASTA, POS, CIGAR, SEQ and FLAG and now implements the corrected contract independently. Picard remains an implementation baseline rather than an absolute oracle and must continue to be measured in both configurations:

```text
Picard SetNmMdAndUqTags                         # conventional reference-aware NM
Picard SetNmMdAndUqTags IS_BISULFITE_SEQUENCE=true  # Picard bisulfite-aware NM
```

The per-read audit records original BAM `NM:i`, both Picard-patched `NM:i` values, independent conventional and conversion-aware NM, Xenofilx's actual recalculated NM, and the number of conversion-compatible bases. RNA-seq uses the conventional Picard route and Xenofilx without `--bisulfite`. BS-seq uses both Picard routes and Xenofilx with `--bisulfite`; the BS-seq control splits records by `FLAG 0x10`, evaluates forward records on a C->T-converted reference and reverse records on a G->A-converted reference, then compares each arm within its own identity universe.

### C3. Acceptance

```text
Xenofilx recalculated NM != independent current-contract oracle: 0 records
Xenofilx classification score != independent current-contract oracle: 0 records
all conventional-vs-bisulfite differences are retained as explicitly classified conversion semantics, not treated as equality failures
```

A small hand-auditable fixture is required before the full BAM scan. It must cover exact match, mismatch, insertion, deletion, soft clipping, reverse read, C-to-T, G-to-A, `=`/`X`, missing NM, incorrect NM and paired records.

## 6. Gate D: Xenofilx vs Picard + XenofilteR classification

Both paths begin from identical original hg38/mm10 BAMs, references and classification thresholds.

```text
legacy: original BAMs -> Picard NM patch -> XenofilteR -> legacy classification/BAM
modern: original BAMs -> Xenofilx --recalculate-nm --bisulfite -> modern classification/BAM
```

Raw reports are normalized to a shared fragment-level schema:

```text
fragment_id
mate_occurrence
modern_classification
legacy_classification
graft_nm
host_nm
graft_mapq
host_mapq
graft_cigar
host_cigar
selected_species
classification_reason
threshold
```

Standard categories are `graft`, `host`, `ambiguous`, `unmapped`, and `discarded`.

Required outputs:

- confusion matrix;
- complete fragment accounting for each input fragment;
- gzip-compressed disagreement table;
- reason-stratified disagreement summary for ties, threshold boundaries, missing mates, conversion-aware NM, secondary/supplementary alignments and unmapped reads;
- filtered-BAM subset/preservation report.

Pass condition:

```text
all input fragments accounted for
all retained modern BAM records trace to original graft BAM records
all record changes are in the declared allowlist
0 unexplained modern/legacy classification disagreement
```

No percentage-agreement threshold is accepted before every disagreement category has been explained. A waiver requires a documented scientific cause and bounded condition.

## 7. Execution sequence

1. Collect readonly baseline metadata and structural statistics from real `SRR36187610` BAMs.
2. Run synthetic bamdriver/pairbam/NM oracle fixtures in CI and on a Slurm compute node.
3. Run real-BAM no-op round-trip and pair integrity comparators.
4. Run Picard NM patch and XenofilteR only after their exact executable/package versions pass preflight.
5. Run Xenofilx NM/classification capture with the frozen reference and parameters.
6. Normalize results, resolve every NM/classification disagreement, and publish the Gate A-D report.
7. Only after Gates A-D pass, resolve and execute the fourteen fresh immutable runs for seven-input modern vs legacy-equivalent full-pipeline comparison.

## 8. Initial environment and baseline findings

At the first 2026-08-26 preflight, `samtools`, `xenofilx`, `pairbam`, `Rscript`, and `XenofilteR 1.6` were found. A standalone `bamdriver` CLI was not found, which is expected because bamdriver is a shared library rather than a production CLI.

The initial readonly real-BAM baseline collection completed as Slurm job `41694759` with exit `0` in 52m43s. Evidence is stored under `/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx/baseline-41694759`. It recorded 90,320,060 mapped primary graft records, 1,885,462 mapped primary host records, and 90,220,046 mapped primary records in the existing modern filtered graft BAM. All three have complete `NM:i` coverage and passed `samtools quickcheck`. This is an input inventory only; it does not yet prove round-trip or classification parity.

The inherited `picard-3.4.0-0` wrapper remains unusable because it selects a broken Conda Java binary (`JLI_StringDup` dynamic-link failure); system Java 8 cannot run Picard's Java 17 class files. The dedicated Enva environment `gate6-picard-java17`, declared in `enva/src/configs/gate6-picard-java17.yaml`, was created at `/public3/home/scg9946/.local/share/mamba/envs/gate6-picard-java17`. Verification job `41695130` passed with Java `17.0.18-internal`, Picard `3.4.0`, and an available `SetNmMdAndUqTags` command. Gate D must call the Enva prefix Java and Picard JAR directly; it must not call the inherited wrapper.

Gate A real-data round-trip completed as Slurm job `41695184` with exit `0` in 39m30s. Evidence is stored under `/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx/bamdriver-roundtrip-41695184`. The round-trip BAMs passed `samtools quickcheck`; decoded headers and ordered canonical record streams matched exactly for all `90,320,060` graft records and `1,885,462` host records. The graft digest was `e679a5901946d6d51416fafc7e3155f019031e13d03ed9fca187b0013bee554d` and the host digest was `1f7894ba0f3fb18672f2cc9d6826600da29809289d69e7a4e8135195fa1c2992` on both sides of each comparison. Gate A passes the current no-op reader/writer preservation contract.

Gate B real-data pair integrity completed in validation report from Slurm job `41696363`. Evidence is stored under `/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx/pairbam-integrity-41696363`. On the frozen mm10 host BAM, pairbam retained all `942,731` fragment names and `1,885,462` records as complete primary pairs; the comparator found equal headers/reference declarations, retained canonical records, RNEXT/PNEXT/TLEN relationships, and filtered-name manifest, with identical retained-stream digest `016e4d1e4d472aec1120d550c1a4e685956b24e3011bf21abb772492aad95645`. Gate B passes this complete-primary-mates contract; this test had zero filtered fragments because the selected input was already complete.

Gate C real-data sample audit completed as Slurm job `41695328` with exit `0` in 53 seconds. Evidence is stored under `/public3/home/scg9946/otter-gate6/evidence/gate6-bam-nm-parity-20260826-bs-pdx/nm-score-audit-41695328`. On the first 200,000 mapped records from each original BAM, the independent oracle and Xenofilx's actual reference-aware calculator agreed on identity, recalculated bisulfite NM, insertion bases, soft-clip bases, and final classification score for all 400,000 records: zero differences, zero malformed alignments, and zero reference lookup failures. The observed aggregate conventional/bisulfite NM totals were `3,028,837/28,951` for graft and `2,643,344/50,287` for host, confirming that conventional NM is not an acceptable bisulfite truth baseline. This validates the current implementation contract only; strand-aware scientific equivalence remains a separate Gate D interpretation question.

Gate D legacy execution completed as Slurm job `41696495` with exit `0` in 27m24s. Direct Java 17/Picard 3.4.0 retained all 90,320,060 graft and 1,885,462 host records while reference-specifically patching NM/MD/UQ; XenofilteR 1.6 with `MM_threshold=6`, `Unmapped_penalty=8`, and `NM_id=NM` emitted 6,434,546 filtered graft records (3,217,273 fragment names). The first membership comparison job `41696588` must not be used for scientific acceptance: its modern BAM path resolved to the prior `run-20260821T112353Z-tunvxb` output, while legacy used frozen `run-20260823T034624Z-ndcwfa` inputs. Its 41,892,990 selection disagreements are therefore a documented provenance incident, not parity evidence. Corrected direct Xenofilx job `41696880` and dependent same-frozen-input membership comparison job `41696881` were submitted; Gate D remains open pending their evidence.
