# Reference Registry 契约

> 状态：目标契约。大型 reference 资产不复制进项目，由共享 registry 统一管理。

## 标准目录

```text
$OTTER_REFERENCE_ROOT/
└── genomes/
    └── <genome_id>/
        └── <release>/
            ├── reference.yaml
            ├── manifest.json
            ├── checksums.sha256
            ├── fasta/
            ├── annotations/
            └── indexes/
                ├── bismark/
                ├── bowtie2/
                └── star/
```

- `genome_id` 是稳定逻辑名，例如 `hg38`、`mm39`。
- `release` 是精确 assembly/release，例如 `GRCh38.p14`。
- release 目录不可变；任何资产修订必须生成新 release 或新 manifest digest。
- 项目 lock 不硬编码集群 mount path；site resolver 在生成 `run.yaml` 时解析实际路径。

## Build a release

`otter reference build` 将用户提供的 FASTA 与 GTF 复制到 staging release，在该 staging 中生成 FAI 和真实 index，校验后将整个 release 目录原子 rename 到 registry。它生成规范的 `reference.yaml`、`manifest.json` 与 `checksums.sha256`；已有 release 不会被覆盖。

默认 registry root 为 `$OTTER_REFERENCE_ROOT`。该环境变量未设置时，默认值是 `~/.otter/references`。对 Gate 6 shared registry 必须显式设置环境变量或传入 `--registry-root`，并在发布后另行执行 compute-node visibility preflight。

```bash
export OTTER_REFERENCE_ROOT=/shared/otter/references

otter reference build \
  --id hg38 \
  --release GRCh38.p14-gencode-v44 \
  --organism 'Homo sapiens' \
  --assembly GRCh38 \
  --alias human \
  --fasta /staging/GRCh38.primary_assembly.genome.fa.gz \
  --gtf /staging/gencode.v44.primary_assembly.annotation.gtf.gz
```

默认会调用 `samtools faidx`，并构建 `bismark`、`bowtie2` 与 `star` indexes。通过 `--indexes bismark,bowtie2,star` 可显式选择；按选择推导兼容场景。`--indexes star` 仅生成 RNA-seq/RNA-PDX 兼容 release，`--indexes bismark` 仅生成 RRBS/WGBS/BS-PDX 兼容 release。完整 release 不能跳过真实 index 构建：schema 要求至少一个 index，命令不会以 placeholder 文件替代工具输出。

All future production `ReferenceBuild` runs use the doubled resource policy: `acquire_sources` requests 2 CPUs/8 GiB, `prepare_assets` requests 4 CPUs/32 GiB, and `publish_release` requests 16 CPUs/192 GiB. Its immutable configuration must set `index_build_threads: 16`, causing Bismark to invoke `--parallel 8` for each concurrent CT/GA conversion indexer while Bowtie2 and STAR use 16 threads. The completed Gate 6 releases remain immutable historical evidence built under their original 8 CPU/96 GiB policy. See [Gate 6 Paracloud Operations](gate6-paracloud-operations.md) for access commands, accepted release evidence, and local-first SRA acquisition policy.

### Gate 6 Craftmake reference builds

Gate 6 source acquisition、provider checksum validation、FASTA/GTF derivation and immutable release publication are compiled by the dedicated Craftmake `ReferenceBuild/build` workflow. It uses the explicit `--reference-build-config` contract rather than the historical `--legacy-config` adapter, and its immutable configuration fixes the release identity, URLs, provider checksum algorithm/value, accepted tool paths, registry root, evidence directory and declared contigs. The allowlisted provider checksum algorithms are `md5` and BSD `sum`; the latter supports Ensembl releases that publish `CHECKSUMS` rather than MD5 files. Provider archives download to `.part` paths with HTTP Range resume and are renamed only after a successful transfer and checksum verification. Source acquisition allows three days for complete production assets. `contigs: "*"` retains every source contig, while a comma-separated list produces an explicitly bounded technical reference. The workflow records Craftmake state/controller logs, Slurm accounting, compressed-source SHA-256 and derived-asset SHA-256 before `otter reference build` atomically publishes the release.

`mm10-canary` is the intentionally small Gate 6 technical reference. It uses Ensembl release 100 GRCm38 chromosome `19` FASTA and the matching Ensembl release 100 GTF filtered to `19`; both source assets use the same unprefixed contig namespace and provider-published BSD `sum` checksums. It is suitable only for scheduler, index, publication, interruption/recovery and minimal pipeline-canary checks. It is **not** an `mm10`/`mm38` substitute for production analysis, full-genome scientific parity, or a project default lock. Production builds use complete GENCODE primary-assembly sources: `hg19@GRCh37.p13-gencode-v19`, `hg38@GRCh38-gencode-v44`, and `mm10@GRCm38-gencode-M25` with aliases `mm10` and `mm38`.

## Reference metadata

`reference.yaml` 描述身份、资产和兼容性：

```yaml
schema_version: otter.reference/v1
reference:
  id: hg38
  release: GRCh38.p14
  organism: Homo sapiens
  assembly: GRCh38
  aliases: [human, hg38]
assets:
  fasta:
    path: fasta/genome.fa.gz
    sha256: sha256:...
    size_bytes: 0
    fai: fasta/genome.fa.gz.fai
  annotations:
    - id: gencode-v44
      type: gtf
      path: annotations/gencode.v44.gtf.gz
      sha256: sha256:...
  indexes:
    - type: bismark
      path: indexes/bismark
      reference_fasta_sha256: sha256:...
      tool: bismark
      tool_version: 0.24.2
compatibility:
  scenarios: [rrbs, wgbs, rnaseq, bs-pdx, rna-pdx]
  workflows: [BeaverBS, BeaverRNA, BeaverPDX, BeaverRNASEQPDX]
```

每个 index 必须声明：

- 对应 FASTA digest；
- 构建工具和版本；
- 影响结果的构建参数；
- 输出文件集合与 checksum；
- 支持的场景和架构。

## Manifest 与 checksum

- `manifest.json` 是 release 的规范化机器清单，按相对路径排序；它必须包含 `reference.yaml`，并覆盖 `fasta/`、`annotations/` 与 `indexes/` 下的每个普通文件。
- 运行/恢复和 promote 会将 manifest 的已声明 entry 与重新构建的目录快照逐项比较；缺失、未跟踪或 digest 变化均 fail closed。
- `checksums.sha256` 覆盖 metadata、FASTA、annotation 和 index 文件。
- directory index 必须展开到文件级 manifest，不能只 hash 目录名。
- registry 发布采用 staging、校验和原子 rename；已发布 release 不允许原地写入。

## 项目锁定

`references.lock.yaml` 锁定项目默认选择：

```yaml
schema_version: otter.references.lock/v1
references:
  primary:
    id: hg38
    release: GRCh38.p14
    manifest_digest: sha256:...
```

PDX 使用 `graft` 和 `host` 两个明确 role。lock 固定逻辑 ID、release 和 manifest digest，不固定 mount root。

## 解析流程

```text
selection
  -> site reference root
  -> reference.yaml schema
  -> manifest digest
  -> checksum verification
  -> scenario/toolchain asset selection
  -> absolute path expansion
  -> run.yaml snapshot
  -> sbatch preflight
```

Workflow 不得自行拼接 FASTA/GTF/index 路径，只能消费 `run.yaml` 中 typed resolved assets。

## Run override 与 promote

单次 run 可指定其他 release，但不改变项目默认。`run.yaml` 同时记录 project/effective selection、override source 和 resolved assets。resume 不允许切换 reference。

`otter reference promote <run_id>` 用于显式提升：

1. 读取 run 的 effective selection。
2. 重新验证 registry 和兼容性。
3. 生成 project/reference lock diff。
4. 用户确认后原子更新 lock。
5. 原子更新后将 run ID、旧/新 lock 和更新时间追加到项目的 `.otter/reference-promotions.jsonl`。

## Fail-closed 验证

提交 sbatch 前必须验证：

- reference/release/manifest 存在且 digest 一致；
- FASTA、FAI、annotation 和 index 文件可读；
- FAI 与 FASTA 一致；
- index 的 FASTA digest 与实际 FASTA 一致；
- RNA-seq 有匹配 GTF 和 STAR index；
- BS-seq 有匹配 Bismark/Bowtie2 index；
- PDX graft/host 资产分别完整；
- scenario、workflow 和 toolchain 在 compatibility 中允许；
- compute node 能访问实际解析出的 registry root；`config resolve --backend slurm` 会在写入 `run.yaml` 前执行 login/compute-node preflight。

失败时不得调用 sbatch；诊断必须给出资产角色、期望值、实际路径和 digest 差异。
