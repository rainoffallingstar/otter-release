# fastqc-rs 系统审查报告（2026-07-22）

## 1. 审查结论

- **仓库基线**：`14ccb7b`（分支 `library-latest`，工作树干净）
- **审查状态**：只读审查完成，发现 1 Critical / 4 High / 8 Medium / 7 Low
- **发布结论**：**阻断发布**。非法碱基 panic、Q20/Q30/AvgQual 公式按 read 而非碱基级（直接污染 qctb 上报）、clippy 失败、doctest 失败致 `cargo test` exit 101。

## 2. 门禁结果

| 命令 | 结果 |
|---|---|
| `cargo fmt --all -- --check` | 通过 |
| `cargo clippy --all-targets --all-features -- -D warnings` | **失败**（lib 18 个 error，lib test 17 个 error） |
| `cargo test --all-features` | **失败**（exit 101）：单测/集成通过，但 `src/lib.rs:8` doctest 用不存在的 `sample.fastq` 触发 panic |
| 系统 `fastqc` / `multiqc` / `seqkit` | 均未安装，无法做外部兼容验证 |

工具链：rustc/cargo 1.97.1，edition 2018，包 `fastqc-rs 0.3.4`，二进制 `fqc`，needletail 0.5.1。

## 3. 架构与数据流

CLI `fqc`（clap v4）：`-q/--fastq`（必填）、`-k/--kmer`（默认 5）、`-s/--summary`（输出 `fastqc_data.txt` 目录）、`--no-html`（跳过 HTML 到 stdout，主仓实际使用模式）。

`analyze(filename, k) -> Result<FastqAnalysisResult>`（库入口，只统计）；`process(...)`（CLI 入口，统计+渲染；两处解析循环几乎逐行重复）。

数据流：`parse_fastx_file`（needletail）逐 record → read_count/read_lengths/gc_content/mean_read_qualities/base_quality_count/base_count/kmers → 派生 warning 等级 → 嵌入 Vega-Lite spec + Tera 渲染 HTML（或 `fastqc_data.txt`）。

## 4. Critical 发现

### FASTQC-C01：非法碱基触发 panic，解析直接崩溃

- **位置**：`src/process.rs:140-149`（analyze）、`:343-352`（process，同逻辑）。
- **触发**：FASTQ sequence 含 `A/C/G/T/N`（含大小写）以外字节（IUPAC `R/Y/S/W/K/M/B/D/H/V`、`U`、`*`、`-`、空白等）。
- **证据**：`_ => panic!("Invalid base")`。
- **影响**：含 IUPAC 模糊碱基的 FASTQ（参考基因组导出、contig、UMI、变体调用输出）让 `fqc` 整个进程 panic，无优雅退出；库 API `analyze` 同样 panic 而非 `Result`。
- **整改**：返回 `io::Error`/自定义 error，或将非标准碱基并入 N 桶，至少 `panic!` 改 `return Err`。

## 5. High 发现

### FASTQC-H01：Q20/Q30/AvgQual 公式语义错误（按 read 平均而非按碱基）

- **位置**：`src/process.rs:619-642`（summary 计算）。
- **证据**：`mean_read_qualities` 是 `HashMap<mean_quality_of_read, count_of_reads>`；`q20_count` 统计 **mean quality ≥ 20 的 read 数**除以 read 总数得 "Q20%"；`avg_qual = Σ(q*c)/Σc` 是 reads 的平均 mean quality。与行业惯例（FastQC/seqkit 碱基级）、MultiQC 预期不符。
- **级联影响**：qctb `parsers/fqc.rs` 把 `fqc` 的 `Q20(%)/Q30(%)` 当碱基级使用，对长 read（150bp）严重高估；**直接污染 qctb 上报的 Q20/Q30**。
- **整改**：遍历 `base_quality_count[pos][score]`（已按碱基的 94-bin 直方图）聚合，按碱基计 Q20/Q30/AvgQual。

### FASTQC-H02：Per base sequence content 警告比较对象错误（T vs G 而非 T vs A）

- **位置**：`src/process.rs:399-404`。
- **证据**：`tg_diff = |bases[T] - bases[G]| / sum`；判定 `gc_diff(G-C)` 和 `tg_diff(T-G)` 任一 ≥0.20 fail / ≥0.10 warn。FastQC 原始模块比较 **A-T 和 G-C**（Chargaff 平衡），这里写成 **T-G** 与 **G-C**。
- **影响**：会误报/漏报 sequence content 警告，与 Babraham FastQC 不一致。
- **整改**：改 `at_diff = |A-T|/sum` 与 `gc_diff = |G-C|/sum`。

### FASTQC-H03：HashMap 输出顺序不确定，fastqc_data.txt 非确定性

- **位置**：`src/process.rs:370-405`、`fastqc_summary.txt.tera`（`base_count`/`read_lengths`/`gc_per_base` 用 `{% for pos, base_map in ... %}`）。
- **证据**：`rustc-hash::FxHashMap` 故意不保序；Per base sequence content / Sequence Length Distribution / Per base GC content 行顺序在不同运行间不可复现。
- **影响**：人工比对/回归测试困难；MultiQC 对部分模块期望按位置升序。
- **整改**：导出前对 key `usize` 排序后输出（`BTreeMap` 或 `collect`+`sort`）。

### FASTQC-H04：embed_source 同步阻塞拉取 6 个外部 CDN，无超时/重试/失败处理

- **位置**：`src/process.rs:717-724`（`embed_source` filter）+ `report.html.tera` head（bootstrap CSS、jquery、popper、bootstrap JS、vega、vega-lite、vega-embed 共 6 个 URL）。
- **证据**：`reqwest::blocking::get(&url).unwrap().text().unwrap()` 两次 unwrap，无 timeout/重试/缓存。
- **影响**：离线/受限网络/CDN 故障时 panic；reqwest 0.11 默认无连接超时可能挂起数十秒。主仓 snakemake 规则全用 `--no-html` 规避，但交互式使用仍危险。
- **整改**：设 `Client::builder().timeout(Duration::from_secs(10))`，失败回退外链 `<script src=url>` 或内嵌资源到二进制。

### FASTQC-H05：计数与内存无上限，k-mer map 与 per-position 直方图可能 OOM

- **位置**：`src/process.rs`（analyze/process）。
- **触发**：超大 FASTQ 或可变长 read（Nanopore 长度可达 10^5+）；`base_count`/`base_quality_count` key 为位置随最大 read 长度膨胀；`kmers` 在 k 大时（`u8` 允许到 255）指数爆炸。
- **影响**：无流式/采样策略（Babraham FastQC 前 200000 read 采样），大文件 RSS 可能数 GB。
- **整改**：k-mer 限长（≤7）、read 长度截断/采样上限、或文档化内存预算。

## 6. Medium 发现

- **FASTQC-M01**：`quartiles` 函数 `assert!(sum != 0)`，空向量 panic 而非错误（`process.rs:188`）。
- **FASTQC-M02**：截断 gzip/malformed FASTQ 仅置 `broken_read=true` 不返回错误也不停止；`fastqc_data.txt` 无字段反映；截断 gzip 可能产生部分 record 后才报错（`process.rs:155-157,333-335`）。
- **FASTQC-M03**：`gc_percentage` 整数除法丢精度（`gc_bases * 100 / called_bases`，`process.rs:104`）。
- **FASTQC-M04**：`phred33_score` 仅支持 Phred+33 范围 0..=93，对 +64/Solexa 静默拒绝，无自动检测（`process.rs:111-123`）。
- **FASTQC-M05**：CRLF/空 read 处理：0 长度 read 不报错但静默计入 `read_lengths[0]`，拉低 `avg_read_length`。
- **FASTQC-M06**：超长/可变长 read 无 capped binning（Vega-Lite 图表 X 轴爆炸）。
- **FASTQC-M07**：doctest 失败（`src/lib.rs:8-11` 用不存在 `sample.fastq`，`cargo test --all-features` exit 101）。
- **FASTQC-M08**：`fastqc_data.txt` 缺标准 FastQC 模块/字段，MultiQC 兼容性有限；自定义 `>>Seqkit Statistics` MultiQC 不识别；Per base sequence quality 列名与数据错位（`lower/upper` 实为 IQR fence 非 10th/90th 百分位）。

## 7. Low 发现

- **FASTQC-L01**：模板 GitHub 链接拼写错误 `hhttps://`（`report.html.tera:48`）。
- **FASTQC-L02**：内联外部 CSS/JS 无 SRI、无离线回退。
- **FASTQC-L03**：`FastqAnalysisResult::new` 12 参数 + `analyze`/`process` 逻辑几乎逐行重复（clippy `too_many_arguments`）。
- **FASTQC-L04**：`quartiles` 注释自认 lower/upper 实为 1.5×IQR fence 而非 min/max，UI 却当 min/max，与 FastQC 10th/90th 百分位语义不同（`process.rs:204-206` TODO）。
- **FASTQC-L05**：lib 构建下 `process`/`quartiles` dead_code 误报（`pub(crate)`）。
- **FASTQC-L06**：README/CLAUDE.md 未提 `--no-html`/`analyze` 库 API/`>>Seqkit Statistics`；作者邮箱 Cargo.toml 与 main.rs 不一致。
- **FASTQC-L07**：`mean_read_quality` 用 `u64` 截断除法丢精度（`process.rs:166`）。

## 8. 主仓实际调用契约

主仓 snakemake 规则（`inst/rules/01fqcAtfirst.smk`、`03-0-fqcAtclean.smk`）**只用 `--no-html` + `-s` 模式**：`fqc -q {input.R1} -s {params.R1_dir} --no-html`。输出契约：`<dir>/fastqc_data.txt`（`fqc` 内部 `create_dir_all`）。

qctb 解析契约（`qctb/src/qc_summary/parsers/fqc.rs`）：定位 `>>Seqkit Statistics` 模块，表头 19 字段，数据行末尾 `>>END_MODULE` 紧贴（`trim_end_matches`），消费 num_seqs/sum_len/Q20(%)/Q30(%)/min_len/avg_len/max_len 七字段。

**关键级联**：FASTQC-H01 的 Q20/Q30 公式错误**直接污染 qctb 上报的 Q20/Q30**；M2 malformed read 不落字段 qctb 无感知；`>>END_MODULE` 紧贴数据行是硬契约，模板空行/换行改动会破坏 qctb 解析。

构建契约：`scripts/build-all-submodules.sh:138` `build_rust "fastqc-rs" "fastqc-rs" "fqc"`；`scripts/install.sh:65` `"fqc:fqc:static"`。

## 9. 外部兼容验证缺口

系统无 `fastqc`/`multiqc`/`seqkit`，无法：
- 对比 fastqc-rs 与 Babraham FastQC 对同一 FASTQ 的 `fastqc_data.txt` 数值差异。
- 用真实 MultiQC 解析 `fqc -s` 产出验证模块名/字段名。
- 用 seqkit 验证 `>>Seqkit Statistics` 的 Q20/Q30/AvgQual（H1 已证明公式不同源）。

## 10. 完成门禁缺口

- [ ] Critical/High 风险为零。
- [ ] 非法碱基不再 panic。
- [ ] Q20/Q30/AvgQual 改碱基级。
- [ ] Per base sequence content 警告改 A-T。
- [ ] summary 输出排序后确定性。
- [ ] `cargo clippy --all-targets --all-features -- -D warnings` 通过。
- [ ] doctest 修复（`cargo test --all-features` exit 0）。
- [ ] embed_source 超时/回退。
- [ ] 内存/k-mer 上限或文档化预算。
- [ ] 真实 fastqc/multiqc/seqkit 差分门禁。
