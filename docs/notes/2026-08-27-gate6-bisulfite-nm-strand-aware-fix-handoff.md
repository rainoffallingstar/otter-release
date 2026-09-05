# Gate 6 Bisulfite NM Strand-Aware 修复交班

> 更新时间：2026-08-27
>
> 本文是下一位 agent 的接续入口，专门记录本次 bisulfite NM 语义修复的根因、源码改动、验证状态、部署授权与未完成步骤。完整比较报告见 `docs/gate6-toolchain-comparison-report.md`，前置计划见 `docs/notes/2026-08-26-gate6-bam-nm-classification-parity-plan.md`。

## 1. 背景与根因

用户要求：探查 modern Xenofilx bisulfite NM 与 strand-aware CT/GA Picard 控制不一致的原因，并修改源码使二者对齐。

逐条 read 审计已经证明（历史证据，修复前基线）：

- Picard conventional NM == 独立 conventional oracle：RNA `200,000/200,000`、BS-PDX host/graft 均 `200,000/200,000`；
- Picard CT/GA converted-reference NM == 对应 arm 的 converted-reference conventional oracle：所有 arm 全部一致；
- Xenofilx 旧实现 == 旧契约 oracle：`200,000/200,000`（旧契约即无条件豁免 C->T 与 G->A）；
- 跨契约差异：host forward CT `72,547/100,038`、host reverse GA `72,880/99,962`、graft forward CT `81,816/100,032`、graft reverse GA `81,957/99,968` 条 read 与原参考 Xenofilx bisulfite 值不同。

根因（源码确认）：`bamdriver/pkg/bamnative/nmtag.go` 的 `CalculateNMCheckedWindow` 在 bisulfite 模式下对每条 read 无条件同时豁免 `reference C -> read T` 和 `reference G -> read A`，不检查 SAM `FLAG 0x10`；且 `CIGAR X` 分支直接累加全部错配，完全绕过转换判断。

## 2. 源码修改（工作树，未提交）

### `bamdriver/`（子仓库，独立提交）

`pkg/bamnative/nmtag.go`：

- `M` 分支逐碱基判断改为 `isBisulfiteConversion(record, referenceBase, readBase, isBisulfite)`；
- 新增 helper：正链（`FLAG 0x10` 未设置）仅豁免 `C->T`，反链（`FLAG 0x10` 已设置）仅豁免 `G->A`；
- `X` 分支：conventional 模式仍按 `operation.Len` 计入 NM（保持标准 SAM 语义）；bisulfite 模式先按长度计入，再仅对链匹配的转换位点逐一扣除；
- `Record.IsReverse()` 已存在（`bamnative.go`，`FlagReverse = 1 << 4`），无需新增 flag。

`cmd/nmoracle/main.go`（独立 oracle，独立 CIGAR/SEQ/reference 遍历，不调用 bamdriver 的 NM 计算）：

- `M`：正链豁免 `C->T`，反链豁免 `G->A`；
- `X`：conventional 按长度计 mismatch；bisulfite 先按长度计 mismatch，再统计链匹配转换位点进 `BisulfiteConversionsIgnored`；
- 汇总公式不变：`xenofilx_bisulfite_nm = conventional_nm - bisulfite_conversions_ignored`，`classification_score = xenofilx_bisulfite_nm + insertions + soft_clips`。

### 测试

`pkg/bamnative/faidx_nmtag_test.go`：

- `TestCalculateNMCheckedBisulfiteConversionsAreStrandAware`：正链 C->T 豁免、正链 G->A 计错配、反链 G->A 豁免、反链 C->T 计错配、`X` C->T/G->A 同规则；
- `TestCalculateNMCheckedMismatchOperationCountsInConventionalMode`：conventional `X` 即使碱基相同也计 1；
- 既有 `TestCalculateNMCheckedBisulfiteConversions` 期望值由 `2/0` 改为 `2/1`（record 无 `FLAG 0x10`，`G->A` 现在计错配）。

`cmd/nmoracle/main_test.go`（新增）：`TestScoreAlignmentAppliesStrandAwareBisulfiteConversions`。

### `xenofilx/`（子仓库，独立提交）

- `internal/classifier/scientific_correctness_test.go`：新增 `TestBisulfiteScoringUsesReadStrand`，通过 `NewEditDistanceCalculatorWithRef` + `CalculateWithRefDetails` 验证 wrapper 到共享 bamdriver 实现的链路；
- `internal/classifier/edit_distance.go`：本会话未改（此前的 `ReferenceScore` API 为更早会话改动，保留）；
- `go.mod`：本会话曾临时加 `replace github.com/rainoffallingstar/bamdriver => ../bamdriver` 用于联调，已移除。**Xenofilx 无 vendor 目录**，发布构建依赖 bamdriver 模块版本。

## 3. 验证状态

### 已通过（本次会话）

- `bamdriver`：`go test ./pkg/bamnative ./cmd/nmoracle -count=1`（含新增 strand-aware 与 `X` 测试）OK；
- `bamdriver`：`go test ./...` 在修复过程中通过过一轮；
- `xenofilx`：临时 replace 下 `go test -mod=mod ./internal/classifier -count=1` OK；
- `go mod verify`（xenofilx）OK；
- `git diff --check`（bamdriver、xenofilx）OK；
- 编辑文件 lint：无诊断。

### 未完成/受阻塞

- 最后一次调整 `X` 的 conventional 语义后，本地命令通道持续无响应：`go test`、`ps`、`agentsshcli --help` 均无法完成（卡在后台）。因此**最新一轮全量测试没有完成输出**，`go vet` 也未执行；仅定向测试通过不足以作为最终验收。
- 曾尝试用子 agent 完成构建/部署/提交，子 agent 启动时被中止。
- 未构建、未上传、未提交任何 Slurm job；远端 runtime binary 仍是旧契约版本。

## 4. 部署授权与目标（已获用户确认）

用户已明确授权覆盖（仅限以下两个 runtime binary，及配套审计脚本副本）：

- `/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-nmoracle`
- `/public3/home/scg9946/otter-gate6/toolchain-comparison-20260815T070000Z/runtime/gate6-xenofilx-scoreaudit`

约束：

- 覆盖前先备份旧 binary（参考 `docs/notes/2026-08-20-gate6-bs-pdx-xenofilx-sort-memory-handoff.md` 的 `xenofilx.r34-backup` 惯例）；
- 记录新旧 SHA-256；
- 证据目录必须全新（各 controller 均拒绝覆盖已有目录）；
- 不触碰其他 runtime 文件与历史证据。

## 5. 下一步精确顺序

1. **本地验证**（命令通道恢复后）：
   - `cd bamdriver && gofmt -l . && go test ./pkg/bamnative ./cmd/nmoracle -count=1 && go vet ./pkg/bamnative ./cmd/nmoracle`
   - `cd xenofilx && (临时加 replace 后) go test -mod=mod ./internal/classifier -count=1`，完成后移除 replace。
2. **构建（静态，参考 r35 惯例）**：
   - `CGO_ENABLED=0 go build -o gate6-nmoracle ./cmd/nmoracle`（bamdriver）
   - scoreaudit 需在 `xenofilx` 内以临时 `replace` 指向本地 `../bamdriver` 后 `CGO_ENABLED=0 go build -o gate6-xenofilx-scoreaudit ./cmd/scoreaudit`。
3. **部署**：备份旧 binary -> 上传新 binary -> `chmod 755` -> 记录 SHA-256（本地与远端）。
4. **提交 read-level jobs**（`scripts/gate6-nm-read-parity-controller.sh`，均为 200,000 前缀、新证据目录）：
   - RNA：`GATE6_NM_MODE=rna` + `SRR1039508` 输入（对照历史 job `41712621`/`41714006` 的参数）；
   - BS-PDX host：`GATE6_NM_MODE=bsseq` + mm10 BAM/参考；
   - BS-PDX graft：`GATE6_NM_MODE=bsseq` + hg38 BAM/参考。
   - 成功后提交 `scripts/gate6-nm-converted-reference-control-controller.sh`（host/graft 各一）。
5. **验收标准**：Xenofilx bisulfite NM == 独立 oracle `0` 差异；CT/GA arm 内 Picard/oracle/Xenofilx 一致；转换差异只作为分类语义保留。
6. **Gate D**：read-level 通过后，用新 xenofilx binary 重跑 `scripts/gate6-modern-xenofilx-controller.sh` + `scripts/gate6-fragment-membership-controller.sh`，重新计算 fragment disagreement。旧 `41,892,990` 仅为修复前基线。

## 6. 关键文件

| 路径 | 说明 |
|---|---|
| `bamdriver/pkg/bamnative/nmtag.go` | 共享 NM 实现（修复主体） |
| `bamdriver/pkg/bamnative/faidx_nmtag_test.go` | strand-aware 回归测试 |
| `bamdriver/cmd/nmoracle/main.go` | 独立 NM oracle |
| `bamdriver/cmd/nmoracle/main_test.go` | oracle 回归测试 |
| `xenofilx/internal/classifier/scientific_correctness_test.go` | wrapper 链路回归测试 |
| `scripts/gate6-nm-read-parity-controller.sh` | read-level 审计（已加 binary/脚本 SHA-256 记录） |
| `scripts/gate6-nm-converted-reference-control-controller.sh` | CT/GA cross-control |
| `scripts/gate6-modern-xenofilx-controller.sh` | modern 分类重跑 |
| `scripts/gate6-fragment-membership-controller.sh` | Gate D membership |
| `docs/gate6-toolchain-comparison-report.md` | 比较报告（已更新 §8 源码修正） |
| `docs/notes/2026-08-26-gate6-bam-nm-classification-parity-plan.md` | 计划（已更新 §5.2 bisulfite contract） |
| `docs/active_context.md` | 系统上下文（已更新 §0） |

## 7. 未关闭事项（沿用既有 Gate 6 列表）

- 七输入 fresh modern vs legacy-equivalent 全流程 parity；
- representative `20 samples x 3 repeats`；
- production-scale scheduler-pressure gate；
- WGBS `SRR6373947` reference requalification；
- BS-PDX publish / artifact manifest verification；
- 真实 Snakemake PDX interruption/retry 与 scientific comparison。

不要重跑已完成的 Methx benchmark，不要把实现性能写成整个工具链总 speedup，不要在没有明确 push 授权时 push。
