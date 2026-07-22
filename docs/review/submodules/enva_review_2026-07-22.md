# enva 系统审查报告（2026-07-22）

## 1. 审查结论

- **仓库基线**：本地 HEAD `b7f45e6336dfa33e9fbeaab89f2710a3a437524c`（`master`，相对 `origin/master` ahead 5）
- **工作树**：5 个源文件未暂存改动（`src/backend/mod.rs`、`src/backend/rattler.rs`、`src/lib.rs`、`src/main.rs`、`src/micromamba.rs`），属用户原有改动，审查保留未动
- **审查状态**：只读审查完成，发现 1 Critical / 7 High / 5 Medium / 2 Low
- **发布结论**：**阻断发布**。环境名可逃逸允许根目录并被 `create --force` 递归删除；`run -- <argv>` 丢失参数边界进入 shell，破坏主仓契约并允许注入。

本地未推送的 5 个提交为 shell hook、交互式多环境删除、toolchain 配置等；脏树改动（`bash -lc`→`bash -c`、补 `--`、禁 banner、banner 移 stderr）改善了主仓机器调用的输出污染和部分参数分隔，但**未修复 argv 边界丢失、shell 注入或 Windows 不可用**。

## 2. 门禁结果

| 命令 | 结果 |
|---|---|
| `cargo fmt --check` | 通过 |
| `cargo clippy --all-targets --all-features -- -D warnings` | **失败** |
| `cargo test --all-features` | 通过（74 passed） |

环境中无 `conda`/`mamba`/`micromamba`，使用系统 Rust 1.97.1，未执行任何真实环境操作。

Clippy 失败包括：`src/backend/mod.rs:26` test module 后仍有公开 item（来自脏树）；`rattler.rs` unit struct `default()`、`types.rs` 可派生 `Default`、测试中 `MutexGuard` 跨 `.await`、`print_literal`、`needless_borrows_for_generic_args` 等。

审查过程中 cargo 门禁生成的临时构建目录（`target-native-tests-20260722`、`target-paracloud-musl-20260722`）已清理，用户 5 个源文件改动保持原样。

## 3. Critical 发现

### ENVA-C01：环境名可逃逸允许根目录，并被 `create --force` 递归删除

- **位置**：`src/backend/rattler.rs:62-110,128-136,1410-1500`。
- **触发**：`enva create --yaml env.yaml --name /absolute/path/to/victim --force` 或 `--name ../../victim --force`。
- **证据**：`target_prefix_for_env_name()` 只拒绝 `"base"`；直接 `preferred_root_prefix().join("envs").join(env_name)`；Rust 路径连接遇绝对路径丢弃前根，`..` 不规范化/拒绝；目标含 `conda-meta/` 时 `--force` 执行 `remove_dir_all()`；无 canonicalize/symlink 策略/必须位于允许根目录内的检查。
- **影响**：可在 rattler root 外创建环境；对任意 conda-style 目录递归删除；root 为符号链接时也可能越界。
- **整改**：强类型 `EnvironmentName` 只允许一个正常路径组件；拒绝绝对/`.`/`..`/分隔符/Windows drive/NUL/空；canonicalize 父目录后验证位于 allowed root `envs/` 下；删除前 `symlink_metadata` 拒绝符号链接；`remove_dir_all` 前验证所有权 marker + 允许根目录 + 非 base + identity。

## 4. High 发现

### ENVA-H01：`run -- <argv>` 丢失参数边界进入 shell，破坏主仓契约且允许注入

- **位置**：`src/env_run.rs:71-93,420-438,461-474`、`src/backend/mod.rs:13-22`、`src/backend/rattler.rs:1235-1270`、`src/micromamba.rs:1300-1375`、主仓 `internal/enva/enva.go:16-38`、`internal/engine/local.go:286-301`。
- **证据**：主仓用独立 argv 调 `enva run env -- tool --output "a b" "$(payload)"`；enva 将 argv 重拼成空格分隔字符串交 `bash -c`；`--script` 也 `format!("Rscript {}", script.display())` 未引用。
- **影响**：`"a b"` 变两参数；`*`/`$()`/反引号/`;`/重定向/管道被 shell 解释；主仓 Go 侧 `exec.Command` 的安全 argv 保护在 enva 内被破坏。
- **整改**：建模 `Argv(Vec<OsString>)`（默认不经 shell）与 `ShellCommand(String)`（仅 `--command`）；`--script` 用独立 argv；主仓契约测试覆盖空格/引号/`$()`/通配/换行/前导 `-`。

### ENVA-H02：`create --force` 先删旧环境再求解安装，无事务/回滚

- **位置**：`src/backend/rattler.rs:1380-1535`、`src/micromamba.rs:1154-1288`。
- **影响**：网络/solver/磁盘错误或中断后原环境永久丢失；无恢复点。
- **整改**：临时 prefix 完成 solve/install/验证/marker，成功后原子 rename（旧→备份，新→正式），最终删除备份；中断 journal/recovery。

### ENVA-H03：锁只在单 backend 实例内生效，无法防跨进程/跨操作竞争

- **位置**：`src/backend/rattler.rs:35-59,1360-1370`、`src/backend/factory.rs:6-17`、`src/micromamba.rs:170-191,1154-1165`。
- **证据**：`creation_lock` 为实例内 `Arc<Mutex<()>>`，每次 `build_backend()` 新建；不跨进程；install/remove/cache cleanup 不共享 prefix 锁。
- **整改**：按 canonical prefix 跨进程 lock file；cache root 共享/排他锁；锁含 PID/时间/操作/transaction ID，处理 stale。

### ENVA-H04：普通 `run`/`install`/`remove` 隐式 adopt 外部环境，所有权边界失效

- **位置**：`src/backend/rattler.rs:817-853,937-980,1580-1695`、`src/ownership.rs:45-82`。
- **证据**：`run()`/`install_packages()` 调 `ensure_adopted_environment()` 直接写 rattler marker；remove 先 adopt；marker 写权限拒绝时 `b7f45e6` 明确"proceeding with direct removal"。
- **整改**：`run` 只读禁隐式 adopt；install 到外部要求 `--adopt`/`--external`；remove 外部需二次确认；marker 不可写时 fail closed。

### ENVA-H05：cache cleanup 对任意环境变量目录递归删除，后端语义不一致

- **位置**：`src/backend/rattler.rs:983-1037,1295-1345`、`src/env.rs:454-456`、`src/micromamba.rs:1071-1148`。
- **证据**：`RATTLER_CACHE_DIR` 无条件接受为 cache root，直接递归删 `pkgs/`/`repodata/`/`run-exports/`；无 allowlist/marker/symlink 验证；CLI 后端执行 `conda|mamba|micromamba clean --all` 影响全局。
- **整改**：只清理由 enva 创建且有 marker 的 cache；拒绝 `/`/home/工作区/当前目录及祖先；输出 canonical 列表要求确认；拆分 rattler cache cleanup 与全局 clean。

### ENVA-H06：自动下载 micromamba 缺版本固定/校验和/原子写入/安全解包

- **位置**：`src/micromamba.rs:648-665,686-870,879-948`。
- **证据**：可变 `latest` URL；只依赖 HTTPS 无 SHA-256/签名；直接写最终路径；外部 `tar -xjf` 解压无 entry 路径校验；存在即跳过。正面：HTTP 客户端用 rustls TLS，未发现禁用证书验证。
- **整改**：固定版本 + 签名清单；`.part` + `fsync` + 原子 rename；Rust archive 库逐项拒绝绝对/`..`/设备/越界 symlink；已存在也验证 hash。

### ENVA-H07：同名环境选择规则不一致，prefix 去重未 canonicalize

- **位置**：`src/backend/rattler.rs:690-720,866-912`、`src/prefix_registry.rs:125-160`、`src/env_run.rs:240-290`、`src/env.rs:700-900`。
- **证据**：去重 key 为原始 `PathBuf` 不 canonicalize；run 按优先级选首个仅 warning；install/adopt 要求唯一；remove 终端交互、非终端直接失败；create 另一套冲突规则。
- **整改**：统一 `EnvironmentIdentity { canonical_prefix, device/inode, uuid }`；多匹配默认失败除非 `--prefix`；主仓支持 canonical prefix。

## 5. Medium 发现

- **ENVA-M01**：ownership marker 非原子写入，解析失败被静默降级为 external（`.ok().flatten()` 丢弃错误）。
- **ENVA-M02**：YAML validation 报"可解析"但未真实求解，后端支持范围不同（`dependencies_resolvable` 为 placeholder）。
- **ENVA-M03**：逗号拆包破坏合法 MatchSpec，兼容安装覆盖原 channels（`numpy>=1.24,<2` 被拆两 token；强制 `-c conda-forge -c bioconda`）。
- **ENVA-M04**：运行路径硬编码 `bash`，与声明 PowerShell/Windows 支持矛盾（脏树集中到 `backend/mod.rs`）。
- **ENVA-M05**：非标准 `CONDA_PREFIX` 被当 rattler root 本身（`rattler.rs:75-91`）。

## 6. Low 发现

- **ENVA-L01**：所有退出码 141 被当成功（`env_run.rs:441-454`，按错误字符串匹配）。
- **ENVA-L02**：全局 `--quiet` 只控制 banner 不控制命令输出（`main.rs:15-20,38-57`）。

## 7. 主仓调用契约检查

主仓多处按标准 argv 调 `enva run <env> -- <command> <arg1> ...`（`internal/enva/enva.go:16-38,55-67`、`internal/engine/local.go:286-301`、`cmd/run.go:619-652`），预期 argv 保持、无 shell 二次解释、stdout 不被 banner 污染、退出状态准确。

当前：argv 保持失败；shell 二次解释存在；141 被吞；banner 脏树已改善/HEAD 仍 stdout 打印；同名环境主仓只传 name 可能选错 prefix。SLURM `fmt.Sprintf("enva run %s --", condaEnv)`（`slurm_array.go:95-101,438-444`）无 shell 引用。

## 8. 完成门禁缺口

- [ ] 环境名单一路径组件强类型验证。
- [ ] destructive prefix canonicalize 并验证位于允许根目录内。
- [ ] symlink 策略明确并测试。
- [ ] base/root/home/workspace/当前目录永不可递归删除。
- [ ] `run -- argv` 保留真实 argv 不经 shell。
- [ ] `--command` 与 argv 模式类型/CLI 分离。
- [ ] 主仓契约测试覆盖特殊字符/空格。
- [ ] 外部环境不被 run/install/remove 隐式 adopt。
- [ ] ownership marker 原子写入、损坏 fail closed。
- [ ] prefix 和 cache 跨进程锁。
- [ ] force replacement 用 staging/原子切换/rollback。
- [ ] 中断 journal/recovery。
- [ ] cache cleanup 有 marker/allowlist/预览。
- [ ] 自动下载固定版本 + 签名/SHA-256。
- [ ] rattler 与兼容后端 capability/semantic matrix。
- [ ] YAML dry-run 真实 solve。
- [ ] MatchSpec 不按裸逗号拆分。
- [ ] CI 覆盖 Linux/macOS/Windows。
- [ ] `cargo clippy --all-targets --all-features -- -D warnings` 通过。
- [ ] 安全负向测试在隔离临时目录通过。
