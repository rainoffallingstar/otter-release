# Otter Benchmark Plan

> 状态：Gate 6 实施计划。Paracloud compute-node runtime 与 SRA metadata-only canary selection 已完成；本地 failure injection + metrics schema 已完成。真实集群 canary、representative 与 scale 尚未执行，先决条件是经校验且 compute-node 可见的 reference registry 与冻结的 canary FASTQ。

## 1. Benchmark Matrix

每场景 × toolchain 组合的测试矩阵：

| Scenario | Modern | Legacy-equivalent | Canary (1-sample) | Repr (20-cell × 3 runs) | Scale |
|---|---|---|---|---|---|
| RRBS | ✅ fastqcx + methx + qctb | FastQC + Bismark + Methrix | □ | □ | □ |
| WGBS | ✅ fastqcx + methx + qctb | FastQC + Bismark + Methrix | □ | □ | □ |
| RNA-seq | ✅ fastqcx + seq2mat + matsrun + qctb | FastQC + STAR + HTSeq + rMATS | □ | □ | □ |
| BS-PDX | ✅ fastqcx + xenofilx + methx + qctb | FastQC + XenofilteR + Bismark + Methrix | □ | □ | □ |
| RNA-PDX | ✅ fastqcx + xenofilx + seq2mat + matsrun + qctb | FastQC + XenofilteR + STAR + HTSeq + rMATS | □ | □ | □ |

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
| canary | 已接受的 Paracloud runtime + reference registry + 受控、已校验的测试 FASTQ (RRBS/WGBS/RNA/PDX 各 1) | 阻塞：reference registry 与 FASTQ acquisition/downsampling；真实 run 未开始 |
| representative | SLURM + 20-cell 数据 + 每 cell ≥ 3 次重复运行 | 阻塞：canary/parity 通过后需 60 次 SLURM 提交 |
| failure injection | 可在 local backend 验证 cancel/test-failure/digest-drift；controller-loss 需 sbatch | 本地已完成框架；真实 workflow recovery 待 canary |
| scale | 生产规模 + scheduler pressure 监控 | 阻塞：真实 canary/representative 与生产数据 |
| metrics | 报告格式和生成器已实现 | ✅ |

## 4. Gate 6 Reference-build Baseline

Gate 6 reference releases are built on Paracloud compute nodes, not login nodes. `mm10` and `mm38` are aliases for the same `GRCm38` assembly and therefore share one immutable release rather than creating divergent duplicate indexes.

| Logical ID | Assembly / annotation | Source release | Gate 6 aliases | Required assets |
|---|---|---|---|---|
| `hg19` | GRCh37.p13 / GENCODE v19 | GENCODE human release 19 | `human`, `grch37` | FASTA, FAI, GTF, Bismark/Bowtie2/STAR indexes |
| `hg38` | GRCh38 / GENCODE v44 | GENCODE human release 44 | `human`, `grch38` | FASTA, FAI, GTF, Bismark/Bowtie2/STAR indexes |
| `mm10` | GRCm38 / GENCODE M25 | GENCODE mouse release M25 | `mouse`, `mm10`, `mm38`, `grcm38` | FASTA, FAI, GTF, Bismark/Bowtie2/STAR indexes |

The candidate GENCODE source URLs and provider MD5 values were independently reachable from Paracloud on 2026-07-29. Before acceptance, the compute-node build must download to a controlled cache with HTTP Range resume, verify the provider MD5, decompress to the canonical build inputs, record SHA-256, construct indexes with the accepted `otter-core` runtime, atomically publish through `otter reference build`, and verify the final release plus compute-node visibility. Production acquisition allows three days because each compressed primary-assembly FASTA is approximately 0.8–0.9 GB at observed external network throughput. A build run must retain its Craftmake state, controller log, Slurm accounting, generated source-manifest, and final manifest digest as immutable Gate 6 evidence.

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
| `controller loss` | 进程崩溃后 resume | orphan task 检测 → 重新调度 → 复用 cache | ✅ 已实现（`resume_integration_test.go`） |
| `reference digest mismatch` | reference.yaml 声明 digest ≠ 实际文件 | `VerifyChecksums` 失败 → fail before sbatch | ✅ 已实现（`manifest.go`） |

## 7. Run Metrics Collection

每个 run 写入 `run.yaml` 后，Craftmake 执行过程中产生：

- `<run_root>/metrics/run.json` —  wall clock、CPU、memory、I/O
- `<run_root>/metrics/tasks.json` — per-task 指标
- state.sqlite 中记录的 controller events（log-structured）

benchmark 聚合工具读取 N 个 run 的 metrics 并生成 `benchmark.json`。

## 8. 实施检查清单

- [x] 6.1 — Benchmark plan + metrics schema
- [x] 6.2 — Failure injection tests (local) 
- [ ] 6.3 — Canary（需 SLURM + reference registry）
- [ ] 6.4 — Representative (20-cell × 3 runs)
- [ ] 6.5 — Scale (production throughput)
- [ ] 6.6 — Publish immutable benchmark report
- [ ] 6.7 — Parity reports per scenario × toolchain
