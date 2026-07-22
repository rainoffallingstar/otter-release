# methrix-cli 系统审查报告（2026-07-22）

## 1. 审查结论

- **仓库基线**：`c35185a58d0a58e5633d7c727dac2ba11065fec6`
- **分支**：`main`
- **审查状态**：只读审查完成，发现 2 Critical / 6 High / 6 Medium / 多 Low
- **发布结论**：**阻断发布**。Rust 静态门禁通过，但存在 Git 凭据暴露、核心 R 兼容契约不成立、发布门禁脱钩与真实规模内存/原子输出风险。

## 2. 门禁结果

| 命令 | 结果 |
|---|---|
| `cargo fmt --all -- --check` | 通过 |
| `cargo clippy --all-targets --all-features -- -D warnings` | 通过 |
| `cargo test --all-features` | 通过（27 passed） |
| R `rhdf5` 可用 | 可用 |
| R `HDF5Array` / `methrix` 可用 | 不可用 |

本地 R 环境缺少 `HDF5Array` 与 `methrix`，因此标准 `loadHDF5SummarizedExperiment()` 兼容性未动态验证。

## 3. 架构与坐标契约

处理链路：RON 参考或 FASTA 提取 CpG → 扫描排序 Bismark cov → 建立 `(contig, 0-based start) → CpG index` → 按样本并行读入全文件 → 对齐参考 → 全局 `Array2<f32>` beta 与 `Array2<u32>` coverage → 可选过滤未覆盖 → 输出。

坐标契约内部自洽：FASTA CpG 为 0-based end-exclusive；Bismark start 1-based 减一；HDF5/Excel 输出恢复 1-based closed。

Coverage 已从文档承诺的 `u16` 改为 `u32`，70,000 coverage 回归测试通过，不再截断。

## 4. Critical 发现

### METHRIX-C01：Git remote URL 包含明文 GitHub 凭据

- **范围**：Git 元数据，非受跟踪源码行。
- **影响**：仓库读写权限可能泄露，凭据可能进入日志与历史。
- **整改**：立即吊销/轮换凭据，改用无凭据 HTTPS 或 SSH，排查暴露范围。
- **跨仓观察**：此问题同样被 `htseq2matrix-go` 审查独立确认。

### METHRIX-C02：当前输出不满足标准 `loadHDF5SummarizedExperiment()` 契约

- **位置**：`src/hdf5/se_compat.rs:15-40`、`src/cli/process.rs:296-309`。
- **证据**：只写自定义 HDF5 根目录 dataset 与属性，不生成标准 `saveHDF5SummarizedExperiment()` 目录契约所需的 `se.rds` 等 R 端序列化对象；R 测试仅用 `rhdf5::h5read()` 手工读取后内存重建，未调用真实 loader；CI 只装 `r-bioc-rhdf5`。
- **影响**：只能证明“R 能用 rhdf5 读取自定义 schema”，不能证明标准 SummarizedExperiment/methrix loader 能直接加载，与项目承诺不符。
- **整改**：若承诺标准兼容，生成真实可加载目录与 `se.rds`；否则移除误导性兼容声明并提供正式 R loader。

## 5. High 发现

### METHRIX-H01：主仓 release 可绕过测试任务直接发布

- **位置**：`.github/workflows/release.yml:20-49`。
- **证据**：`build-and-release` 没有 `needs: test`；`test` job 不运行 cargo/R 门禁；主仓普通 CI 与 release test 均不测试 methrix。
- **整改**：给 `build-and-release` 加 `needs: test`，在当前子模块指针上运行 Rust 与真实 R 消费者门禁。

### METHRIX-H02：正式发布版 `download-genome` 默认不可用

- **位置**：`Cargo.toml:66-68`、`src/main.rs:174-189`、`.github/workflows/release.yml:186-191`、`scripts/build-all-submodules.sh:129-139`。
- **证据**：下载在可选 feature 中，默认 feature 为空；release 与构建脚本均不带 `--features download`。
- **整改**：正式构建启用 `download` 或从 CLI 隐藏。

### METHRIX-H03：核心产物非事务性发布

- **位置**：`src/hdf5/se_compat.rs:21`、`src/cli/process.rs:301-308`、`src/qc/report.rs:43`、`src/annotation/mod.rs:219`、`src/genome/cpg.rs:210-216`。
- **证据**：HDF5、Excel、RON 与 annotation 直接 `File::create` 最终路径；`assays.h5` 与兼容副本 `methrix_data.h5` 分两步；中途失败留截断或混合版本文件。
- **整改**：同目录临时文件写入，flush/sync 后原子 rename。

### METHRIX-H04：并行设计导致多份峰值内存

- **位置**：`src/cli/process.rs:59-60,62-78,110-115`、`src/processing/filter.rs:47-59`、`src/hdf5/se_compat.rs:68-82`。
- **证据**：主矩阵加 per-sample 结果加过滤矩阵加 HDF5 转置复制；2,800 万 CpG × 100 样本仅 beta+coverage 即约 44.8 GB，过滤/转置期接近 67 GB。
- **整改**：流式 Bismark 解析；分块 HDF5 写入；不要同时保留全部 `ProcessedSample`。

### METHRIX-H05：默认 annotation 报告超过 Excel 行上限

- **位置**：`src/annotation/mod.rs:169-217,225-251`、`src/main.rs:52-54`、`inst/rules/methrix_object.smk:76-84`。
- **证据**：annotation 默认开启，主仓调用不带 `--skip-annotation`；WGBS 超过 1,048,575 CpG 上限会在 HDF5/QC 已发布后失败。
- **整改**：明细改 TSV/Parquet/HDF5 或分 sheet，Excel 只保留汇总。

### METHRIX-H06：下载缺少完整性验证与原子安装

- **位置**：`src/genome/download.rs:39-74`。
- **证据**：直接 `reqwest::blocking::get`、直接写最终路径、无校验和、无 `.part`、无 `sync_all`、无大小/超时上限、无截断回归测试。
- **整改**：固定 URL、校验和、`.part` 原子安装、超时与大小上限。

## 6. Medium 发现

- **METHRIX-M01**：`--threads` 非端到端线程上限（自定义 pool 仅包围样本处理，后续过滤/stats 用全局 Rayon pool）。
- **METHRIX-M02**：coverage 从 `u16` 改 `u32` 但 schema 未版本化，消费者契约可能漂移。
- **METHRIX-M03**：Bismark `end` 只检查不早于 start，匹配时完全忽略，可能掩盖错误输入。
- **METHRIX-M04**：主仓规则找不到精确 GTF 时静默选目录中首个 GTF。
- **METHRIX-M05**：主仓未声明全部 side effect，`methrix_data.h5` 不受 Snakemake 生命周期管理。
- **METHRIX-M06**：兼容性与异常输入门禁不完整（缺真实 loader、多线程等价、截断 gzip、空文件、内存预算、R 差分；`tests/integration/test_full_pipeline.rs` 未被引入）。

## 7. Low 发现

`canonical_contig_name` 不处理 `Chr1`；标准染色体排除 `chrM/MT`；FASTA contig 长度用 `as u32` 可能截断；未固定 Rust toolchain；release `*-static` 未做链接验证；`find_bismark_files` 未先确认 regular file。

## 8. 主仓调用

- 已对齐：主仓调用 `methrix-cli`，本地 build 与安装器把 `methrix` 安装为 `methrix-cli`，名称一致；扫描有确定性排序；样本名冲突被拒绝。
- 缺失：主仓 CI/release 不测试 methrix；release 不依赖 test；子仓 CI 无 fmt/严格 clippy/完整 test；R CI 仅验证自定义 schema；release 默认未启用 download；无 Snakemake 集成测试；无原子输出门禁。

## 9. 完成门禁缺口

- [ ] Git remote 凭据已轮换。
- [ ] HDF5 产物满足真实 `loadHDF5SummarizedExperiment()` / `load_HDF5_methrix()` 契约。
- [ ] `build-and-release` 依赖 test 并运行真实 R 门禁。
- [ ] 所有产物原子发布。
- [ ] 流式/分块处理满足真实规模内存预算。
- [ ] annotation 不依赖 Excel 行上限。
- [ ] 下载固定版本、校验和、原子安装。
- [ ] 真实 R 消费者、多线程等价、异常输入门禁。
