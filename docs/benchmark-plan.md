# Otter Benchmark Plan

> 状态（2026-08-13）：Paracloud runtime、immutable reference registry 与四个 active non-WGBS 的 deterministic canary inputs 已接受。RRBS、RNA-seq、BS-PDX 与 RNA-PDX 均已完成 Craftmake 对 explicit Snakemake compatibility executor 的有界 parity：输入/reference/toolchain/resource contract 固定，适用的 artifact verify/compare、BAM/BAI checksum 与 mapped-read checks 均通过。BS-PDX 与 RNA-PDX 的 `step2-check` 各已完成 3 个 balanced paired scheduler repeats（12 cells total）；结果仅支持有界的 controller-reconciliation 描述性比较。representative 20-cell matrix、scale、production input reacquisition 与 clean-host release gate 仍未完成；WGBS 继续暂缓。

> 生产数据状态：五个登记 accession 均为 `missing_reacquire_for_production`。当前 accepted parity 不能被重述为 production-data、representative 或 throughput acceptance。

## 0. Execution Plan and Delivery Gates

本计划将“已有对比数据”“可用于技术判断的对比数据”和“可用于生产放行的对比数据”严格区分。所有时间均按实际 Paracloud Slurm 队列可用性估算，不把排队时间伪装成执行时间。

| 阶段 | 工作内容 | 交付物 | 预计耗时 | 放行条件 |
|---|---|---|---:|---|
| P0–P1 | 运行时/参考前置条件与 RRBS recovery closeout | Completed | 已有 immutable runtime/reference 与 RRBS recovery/semantic evidence；历史任务描述保留作审计记录 |
| P2 | RNA-seq、BS-PDX、RNA-PDX paired canary | Completed, bounded | 三场景均已完成 fresh paired executor evidence；不等同于生产数据或代表性 acceptance |
| P3 | 至少三次 paired repeat 的小样本 benchmark | Partially completed | BS-PDX 与 RNA-PDX `step2-check` 已各完成三次；其余场景、其他 phase 与 topology-identical microbenchmark 仍开放 |
| P4 | 20-cell × 3 representative matrix | Not started | 等待 production input reacquisition、approved canary/recovery review 和 P3 扩展 |
| P5 | Production throughput 与 scheduler-pressure | Not started | 仅在 P4 通过后执行 |

**预计获得数据的时间点：**

- **第一批可用于技术判断的对比数据：** P1 完成后，约 1–3 个工作日；范围是 RRBS，包含现有 r16b artifact parity、r17b reconciliation 和新增 recovery/semantic 结果。
- **四个 active non-WGBS 场景的完整 canary 对比数据：** P2 完成后，约 4–8 个工作日；包括 RRBS、RNA-seq、BS-PDX、RNA-PDX 的 Craftmake vs Snakemake paired reports。
- **可用于代表性生产决策的统计数据：** P4 完成后，约 7–15 个工作日；这是 20-cell × 3-repeat 矩阵之后的中位数、范围和 incident 汇总，不是单次 canary 的性能结论。

若出现 queue timeout、resource exhaustion、tool invocation 或 artifact integrity incident，估算时间顺延；不得通过跳过 recovery 或 semantic review 来提前宣称通过。WGBS 在本计划中仍 deferred。

## 1. Benchmark Matrix

每场景 × toolchain 组合的测试矩阵：

| Scenario | Modern | Legacy-equivalent | Canary (1-sample) | Repr (20-cell × 3 runs) | Scale |
|---|---|---|---|---|---|
| RRBS | ✅ fastqcx + methx + qctb | FastQC + Bismark + Methrix | ✅ bounded executor parity, recovery suite, semantic review | □ | □ |
| WGBS | ✅ fastqcx + methx + qctb | FastQC + Bismark + Methrix | □ deferred | □ | □ |
| RNA-seq | ✅ fastqcx + seq2mat + matsrun + qctb | FastQC + STAR + HTSeq + rMATS | ✅ bounded executor parity through publication | □ | □ |
| BS-PDX | ✅ fastqcx + xenofilx + methx + qctb | FastQC + XenofilteR + Bismark + Methrix | ✅ bounded executor parity; 3-repeat `step2-check` scheduler evidence | □ | □ |
| RNA-PDX | ✅ fastqcx + xenofilx + seq2mat + matsrun + qctb | FastQC + XenofilteR + STAR + HTSeq + rMATS | ✅ bounded executor parity; 3-repeat `step2-check` scheduler evidence | □ | □ |

The `✅` marks in the Modern and Legacy-equivalent columns describe available toolchain mappings only; they do not mean that the corresponding scenario has passed a production-grade canary or parity gate. `△` records partial evidence that is intentionally insufficient for promotion.

## 1.1 Seven-Decoded-FASTQ Toolchain Comparison (2026-08-15)

The next comparison corpus consists of the seven Paracloud-ready paired FASTQ datasets recorded in `docs/gate6-paracloud-operations.md`: human RRBS `SRR31480456`, mouse RRBS `SRR10025242`, human RNA-seq `SRR1039508` and `SRR018258`, mouse RNA-seq `SRR037954`, BS-PDX `SRR23802966`, and RNA-PDX `SRR30880970`. WGBS `SRR6373947` remains deferred.

Every comparison run must be created by `otter config validate` and `otter config resolve`, which materialize a fresh immutable `otter.run/v1` snapshot. Otter owns project, sample, reference, acquisition-provenance, and input-identity validation. Craftmake is the only workflow control plane: `otter run --executor craftmake` delegates planning, submission, resume, status, logs, and reporting to Craftmake. Production workflow tasks must not be submitted through handwritten `run.yaml` files or direct `sbatch` invocations.

All seven inputs follow the same fixed pre-processing boundary: raw paired FASTQ receives modern `fastqcx` and legacy FastQC + SeqKit QC, then the same pinned Trim Galore invocation, then modern and legacy QC of the trimmed pair. Trim Galore is a shared fixed transformation in this comparison; it is not a new-versus-legacy trimming axis. The downstream cells are RRBS (`fastqcx + methx + qctb` versus `FastQC + Bismark + Methrix`), RNA-seq (`fastqcx + seq2mat + matsrun + qctb` versus `FastQC + STAR + HTSeq + rMATS`), BS-PDX (`fastqcx + xenofilx + methx + qctb` versus `FastQC + XenofilteR + Bismark + Methrix`), and RNA-PDX (`fastqcx + xenofilx + seq2mat + matsrun + qctb` versus `FastQC + XenofilteR + STAR + HTSeq + rMATS`). `pairbam` and `bamdriver` are compared only in BS (RRBS/WGBS) and BS-PDX paired-BAM phases; they are not RNA-seq or RNA-PDX operators.

Snakemake remains an explicit compatibility/rollback executor only. The planned seven-sample toolchain matrix is Craftmake-driven; only the separately documented real-workflow interruption/retry, publication/recovery, and compatibility-smoke debt may use Snakemake.

## 2. Parity Criteria

| Tier | 要求 | 通过条件 | 不通过处理 |
|---|---|---|---|
| **Exact** | 规范化输出完全一致 | byte-for-byte or sorted-text identical | blocker |
| **Structural** | 文件格式、字段、维度、样本顺序、reference identity 一致 | 字段级 diff = 0 | blocker，允许格式转换豁免 |
| **Scientific** | 领域容差内结果等价 | 预定义 tolerance（甲基化 ±0.01、counts ±5%、splicing ±1% PSI） | 需 root-cause 分析 + waiver |
| **Informational** | logs/images/timestamps 合理 | 无硬性要求 | 记录偏差，不阻断 |

## 3. Cluster Dependencies

| Gate 6 子项 | 集群需求 | 状态 |
|---|---|---|
| canary | Accepted Paracloud runtime/reference plus checksum-verified canary fixtures; production candidates still require reacquisition | RRBS, RNA-seq, BS-PDX, and RNA-PDX bounded executor parity accepted; production-data requalification remains open |
| representative | Slurm + approved 20-cell inputs + ≥3 runs per cell | Blocked on production input reacquisition and P3 expansion; no 60-run matrix submitted |
| failure injection | Local backend plus real-Slurm recovery evidence | Local contract plus RRBS real-Slurm interruption/resume, publication retry, and failed-task retry accepted; additional scenario-specific exercises are part of the canary expansion |
| scale | Production-scale data + scheduler-pressure monitoring | Blocked on representative acceptance and approved production inputs |
| metrics | Versioned metrics/evidence schema and report generation | ✅; PDX repeated scheduler source/summary/chart are retained under `craftmake/doc/benchmarks/` |

## 4. Gate 6 Reference-build Baseline

Gate 6 reference releases are built on Paracloud compute nodes, not login nodes. `mm10` and `mm38` are aliases for the same `GRCm38` assembly and therefore share one immutable release rather than creating divergent duplicate indexes.

| Logical ID | Assembly / annotation | Source release | Gate 6 aliases | Required assets |
|---|---|---|---|---|
| `hg19` | GRCh37.p13 / GENCODE v19 | GENCODE human release 19 | `human`, `grch37` | FASTA, FAI, GTF, Bismark/Bowtie2/STAR indexes |
| `hg38` | GRCh38 / GENCODE v44 | GENCODE human release 44 | `human`, `grch38` | FASTA, FAI, GTF, Bismark/Bowtie2/STAR indexes |
| `mm10` / `mm38` | GRCm38 / GENCODE M25 | GENCODE mouse release M25 | `mouse`, `mm10`, `mm38`, `grcm38` | FASTA, FAI, GTF, Bismark/Bowtie2/STAR indexes |
| `mm9` | NCBIM37 / GENCODE M1 | GENCODE mouse release M1 | `mouse`, `mm9`, `ncbi37` | FASTA, FAI, GTF, Bismark/Bowtie2/STAR indexes |

The four production releases are accepted: provider archives were downloaded to controlled caches with HTTP Range resume, provider checksums were verified, canonical FASTA/GTF assets and SHA-256 evidence were generated, and indexes were built with the accepted `otter-core` runtime before atomic publication. Each release retained Craftmake state, controller logs, Slurm accounting, source manifests, and final manifest digest. Their compute-node integrity, readonly sealing, visibility, and `mm38` alias evidence are recorded in `docs/gate6-paracloud-operations.md`.

The first bounded build is `mm10-canary@GRCm38-ensembl-100-chr19`: Ensembl release 100 GRCm38 chromosome `19` FASTA with the matching Ensembl release 100 GTF filtered to contig `19`, plus Bismark/Bowtie2/STAR indexes. Both source archives publish BSD `sum` provider checksums and use the same unprefixed contig namespace. This is deliberately a technical canary for Craftmake/Slurm/index/publication/recovery validation; it must never be reported as a full `mm10`/`mm38` result or used for scientific parity. A successful canary de-risks the larger `mm10`/`mm38` release but does not satisfy its acceptance criteria.

## 5. Benchmark Metrics Schema

```json
{
  "benchmark_id": "bs-repr-20260726T000000Z",
  "run_ids": ["run-20260725T120000Z-abc123", "run-20260725T140000Z-def456"],
  "scenario": "rrbs",
  "toolchain": "modern",
  "backend": "slurm",
  "metrics": {
    "wall_clock": {"min_ms": 0, "max_ms": 0, "avg_ms": 0},
    "cpu_seconds": {"min": 0, "max": 0, "avg": 0},
    "peak_memory_mb": {"min": 0, "max": 0, "avg": 0},
    "io_read_mb": {"min": 0, "max": 0, "avg": 0},
    "io_write_mb": {"min": 0, "max": 0, "avg": 0}
  },
  "parity": {
    "tier": "scientific",
    "status": "pending",
    "comparator": "methrix-semantics/v1",
    "modern_artifacts": ["results/methylation/matrix.h5"],
    "legacy_artifacts": ["results/methylation/legacy_matrix.h5"],
    "tolerance": {"methylation_beta": 0.01, "coverage": 0.05}
  },
  "evidence": {
    "config_digest": "sha256:...",
    "reference_digest": "sha256:...",
    "workflow_digest": "sha256:...",
    "tool_digests": {"fastqcx": "sha256:...", "methx": "sha256:..."}
  }
}
```

## 6. Failure Injection Tests (local backend)

| 测试 | 场景 | 预期行为 | 状态 |
|---|---|---|---|
| `cancel` | 运行中发送 SIGTERM → exit code 8 | craftmake 收到 cancel → cancelled state → state.sqlite 标记 cancelled | ✅ 已实现（`exit_codes_integration_test.go`） |
| `task failure` | 步骤故意失败（exit ≠ 0） | task 标记 failed → controller 记录 attempt.failed → exit code 5 | ✅ 已实现 |
| `digest drift` | snapshot digest 与 runtime 不一致 | 检测到 drift → fail before sbatch | ✅ 已实现（`run_test.go`） |
| `controller loss` | 进程崩溃后 resume | orphan task 检测 → 重新调度 → 复用 cache | ✅ local contract；RRBS r18 real-Slurm controller interruption/resume、publication retry、failed-task retry passed |
| `reference digest mismatch` | reference.yaml 声明 digest ≠ 实际文件 | `VerifyChecksums` 失败 → fail before sbatch | ✅ 已实现（`manifest.go`） |

## 7. Run Metrics Collection

每个 run 写入 `run.yaml` 后，Craftmake 执行过程中产生：

- `<run_root>/metrics/run.json` —  wall clock、CPU、memory、I/O
- `<run_root>/metrics/tasks.json` — per-task 指标
- state.sqlite 中记录的 controller events（log-structured）

benchmark 聚合工具读取 N 个 run 的 metrics 并生成 `benchmark.json`。

## 8. Production Validation Contract

The authoritative four-scenario input/executor/toolchain matrix is [Gate 6 Canary Matrix](gate6-canary-matrix.md). Its dimensions are intentionally independent: executor parity fixes toolchain and compares Craftmake-default to explicit Snakemake compatibility; toolchain parity fixes executor and compares modern to legacy-equivalent implementations. Rust Bismark 3.1.0 and Bowtie2 2.5.4 are pinned shared dependencies in both paths. Perl Bismark is not an input, oracle, performance baseline, or result-comparison axis.

### Required Metrics and Interpretation

| Group | Required fields | Gate interpretation |
|---|---|---|
| reproducibility | input/reference/workflow/tool/run digest, executor/backend/site, selection seed/rule, resource request/source, task/attempt/job IDs, cache decision, artifact manifest digest | missing or mismatched identity is a blocker |
| correctness | artifact verify result, exact/structural report, methylation beta/coverage, RNA feature/count agreement, PDX classifications/mapped reads, splicing outcome/PSI | exact and structural failures block; scientific differences require root cause and approved expiry-bound waiver |
| reliability | completion rate, typed incident category, retry count, queue delay, scheduler state/exit/signal, cancellation latency, recovery result, publication idempotency | any unclassified or unresolved blocker incident blocks representative work |
| resource efficiency | wall time, CPU seconds, allocated/requested CPU, peak RSS/requested memory, read/write I/O, artifact bytes, task critical path, queue delay, metric quality | report repeated-run median and range; do not promote based on a single run |
| scientific QC | raw/trimmed reads and retention, alignment/mapping rate, duplicates when emitted, bisulfite conversion/coverage, RNA assignment, PDX category proportions | deviations require predeclared scenario baseline or an incident/waiver |

Craftmake `report` now writes `task_metrics.csv`, `step_timings.csv`, `allocations.csv`, and `run-evidence.json`. The evidence bundle binds run identity and task statuses to retained controller logs, accounting exports, and the versioned runtime incident records.

### Runtime Incident Policy

Failed or cancelled attempts produce `otter.runtime-incident/v1` records. Categories are `input_reference_digest`, `environment_tool`, `scheduler_submission`, `queue_timeout`, `resource_exhaustion`, `filesystem_io`, `network_acquisition`, `tool_invocation`, `scientific_qc`, `artifact_integrity`, `cancellation_recovery`, and `internal_unknown`. Each record includes retry safety/policy, owner, escalation rule, remediation state, backend job identity, exit/signal, and bounded diagnostic/evidence paths. Free-form stderr remains human evidence rather than a stable classifier.

Automatic retry is permitted only for bounded scheduler-submission or network-acquisition incidents. Cancellation/recovery and tool invocation require reviewed manual retry. Resource exhaustion requires a new immutable run with reviewed resources. Input/reference, environment, filesystem, scientific-QC, and artifact-integrity incidents prohibit retry until their contract failure is corrected. `internal_unknown` prohibits promotion and must be classified before another run is accepted.

### Gate Sequence

1. Publish and verify seed-bound immutable inputs.
2. Run Craftmake and explicit Snakemake modern canaries on compute nodes.
3. Verify artifacts and perform executor parity plus scenario semantic review.
4. Run real Slurm cancel, task-failure, controller-loss/resume, and publication-retry evidence.
5. Only then run the 20-sample, three-repeat representative matrix, reporting median/range and incident review.
6. Run production-scale throughput/scheduler-pressure jobs only after representative acceptance.

## 9. 实施检查清单

- [x] 6.1 — Benchmark plan + metrics schema
- [x] 6.2 — Failure injection tests (local)
- [ ] 6.3 — Production-grade canary：为 RRBS、RNA-seq、BS-PDX、RNA-PDX 生成 seed-bound immutable downsampled inputs，并各自运行 Craftmake 与 explicit Snakemake 的 Slurm canary
- [ ] 6.4 — Canary semantic parity 与真实 Slurm recovery：artifact manifest/structural validation、scenario-specific comparator、cancel/retry/controller-loss evidence
- [ ] 6.5 — Representative (20-cell × 3 runs)：仅在 canary/parity/recovery 全部通过后启动
- [ ] 6.6 — Scale (production throughput)
- [ ] 6.7 — Publish immutable benchmark and parity reports per scenario × toolchain
