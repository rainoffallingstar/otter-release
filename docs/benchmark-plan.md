# Craftmake 迁移验证与 Benchmark 计划

> 状态：目标验证计划。Local 只做协议/代码验证；所有真实生信流程、恢复、科学等价和性能比较统一在 sbatch 集群执行。

## 两层门禁

### Local contract-only

允许：

- Go/Rust 编译、fmt、vet、clippy、unit tests；
- JSON Schema、YAML 示例、legacy migration；
- DAG 编译、cycle/missing dependency/resource validation；
- CLI JSON/JSONL 和固定退出码；
- fake tools/fake SLURM、submission argv、cancel/resume 状态机；
- run ID、不可变 snapshot、reference override/promote preview；
- auto detection 的 no-SLURM、complete-SLURM、partial-SLURM fail-closed cases。

禁止将 Local 结果解释为真实流程、科学正确性或性能证据。

### sbatch production validation

必须覆盖：

- 真实工具、reference、FASTQ 和共享存储；
- job array、dependency、retry、cancel、resume；
- queue wait、scheduler overhead、I/O、CPU、RSS 和 wall time；
- 五个科学场景的 artifact contract 与语义等价；
- cold/warm cache 分离；
- executor 与 toolchain 两种正交比较。

## Matrix

每个 scenario 的核心矩阵：

| 比较目标 | 固定项 | 变化项 |
|---|---|---|
| Executor parity/performance | scenario、toolchain、reference、inputs、resources | Craftmake vs Snakemake |
| Tool parity/performance | scenario、executor、reference、inputs、resources | modern vs legacy-equivalent |

五场景共 20 个基础 cell：

```text
5 scenarios × 2 executors × 2 toolchains
```

每个 cell 至少 3 次成功重复；首次 cold-cache 和后续 warm-cache 分开报告。Legacy extensions 不进入这 20 个主 cell。

## 数据集层级

- `canary`：每场景最小真实样本，验证提交和 artifact contract。
- `representative`：代表性深度、样本数和 paired-end 特征，作为 parity 主门禁。
- `scale`：生产规模，评估吞吐、队列压力、job array 和恢复成本。

数据集、样本子集、reference release 和输入 checksum 在 benchmark manifest 中冻结。

## 性能指标

至少收集：

- end-to-end wall time；
- submit-to-start queue time；
- executor orchestration overhead；
- phase/task wall time；
- allocated/used CPU time 和 CPU efficiency；
- peak RSS、requested/used memory；
- read/write bytes 或站点可用 I/O 指标；
- SLURM job、array task、retry 和 failure counts；
- success/cancel/resume latency；
- artifact count/size/checksum。

`metrics/` 保存原始 `sacct` 快照、executor metrics 和规范化 `benchmark.json`。报告必须展示绝对值和相对变化，不只展示百分比。

## 科学 parity

### Exact

适合 sample IDs、reference identity、feature IDs、规范化 TSV/JSON、row/column ordering。忽略声明的 timestamp、absolute run root、executor metadata。

### Structural

验证：

- artifact 集合与 schema；
- matrix shape、dtype、row/column identity；
- BAM header/reference dictionary；
- HDF5 group/dataset/attributes；
- report 必需 section。

### Scientific

比较器由 artifact catalog 指定：

- alignment/count：mapped/unique/multi/ambiguous counts 和容差；
- methylation：locus identity、coverage、beta value、missingness；
- expression/splicing：feature universe、counts、event IDs 和数值容差；
- PDX：graft/host/ambiguous/unmapped 分类分布；
- QC：关键指标和 pass/warn/fail 语义。

阈值必须在运行前版本化，不得看到结果后修改当前 gate；需要调整时新建 parity policy version 并重跑。

## 公平性控制

- 相同 `run.yaml` 意图、samples/reference/workflow/tool digest；
- 相同 partition/QOS、CPU、memory、time、parallel limits；
- 每个 cell 独立 run root、work/state/log；
- 随机或轮换运行顺序，避免固定时段偏差；
- 不共享会改变性能的中间缓存；若使用共享 tool/reference cache必须记录 cache state；
- executor 比较固定 toolchain，工具比较固定 executor；
- 失败 run 保留，不从分母中静默删除。

## 故障与恢复测试

每个 scenario 至少注入：

- task exit failure；
- SLURM cancel；
- controller SIGTERM；
- transient submit/accounting error；
- output 已存在但 checksum 不一致；
- immutable config/reference digest 被修改。

验收：依赖阻断正确、无孤儿作业、状态可追溯、resume 只重跑必要任务、发布不覆盖有效产物。

## 退场门禁

Snakemake compatibility 只有同时满足以下条件才能考虑移除：

- 五场景 representative 与 scale matrix 全部通过；
- modern 和 legacy-equivalent 的科学 parity 都有证据；
- Craftmake SLURM cancel/resume/retry 通过故障注入；
- 性能无未解释的阻断级回退；
- 至少一个约定观察窗口内没有需要 Snakemake fallback 的生产事件；
- rollback、release note 和用户迁移说明已发布。

Craftmake 成为默认 executor 不等于 Snakemake 已可删除。

## 报告与判定

每批次发布一个不可变 benchmark manifest 和 summary：

- commit/submodule/workflow/environment/reference digests；
- matrix cell/run IDs/SLURM IDs；
- parity policy 与比较器版本；
- 原始和规范化 metrics；
- pass/fail/waiver；
- 未解释差异和 owner。

Waiver 必须有范围、理由、责任人和到期条件；不得把失败 cell 标记为成功。
