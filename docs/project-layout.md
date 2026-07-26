# Otter 项目目录契约

> 状态：目标契约，尚未由当前 `otter init/run` 完整实现。

## 设计原则

项目保存可审查、可版本化的用户意图和轻量流程资产；每次运行使用独立且不可变的运行快照。大型参考基因组由共享 registry 管理，不复制到项目。

```text
project/
├── project.yaml
├── samples.tsv
├── references.lock.yaml
├── project.lock.yaml
├── workflows/
│   ├── snakemake/
│   └── craftmake/
├── rules/
├── environments/
├── schemas/
└── runs/
    └── <run_id>/
        ├── run.yaml
        ├── manifest.json
        ├── input/
        ├── work/
        ├── results/
        ├── logs/
        ├── state/
        ├── metrics/
        └── artifacts.json
```

## 项目级文件

| 路径 | 责任 | 可否由 run 修改 |
|---|---|---|
| `project.yaml` | 用户意图、默认流程、默认 executor/backend/toolchain | 否 |
| `samples.tsv` | 版本化样本、FASTQ、分组、批次和 adapter | 否 |
| `references.lock.yaml` | 项目默认 reference release 与 manifest digest | 否 |
| `project.lock.yaml` | 工作流、rules、schema、环境声明的 digest | 否 |
| `workflows/` | 项目锁定的 Snakefile 与 Craftmake workflow spec | 否 |
| `rules/` | Snakemake 兼容规则 | 否 |
| `environments/` | Enva 环境声明或 lock | 否 |
| `schemas/` | 本项目采用的配置与协议 schema | 否 |

`otter init` 应从安装资产生成上述轻量文件，并在 `project.lock.yaml` 中记录来源版本与 SHA-256。运行不得回退到未记录 digest 的安装目录资产。

## Run ID

默认 run ID 格式：

```text
run-YYYYMMDDTHHMMSSZ-abcdef
```

例如：

```text
run-20260726T013245Z-kxqjrm
```

约束：

- 时间戳使用 UTC，精度为秒。
- 后缀为安全随机源生成的 6 位小写英文字母。
- 使用原子 create-new 语义创建 `runs/<run_id>/`；冲突时只重试随机后缀。
- `--run-id` 自定义值必须符合 `^run-[0-9]{8}T[0-9]{6}Z-[a-z]{6}$`。
- `resume` 复用原 run ID，不生成新 ID；改变 reference/config/executor 时必须创建新 run。

## 运行目录

| 路径 | 内容 |
|---|---|
| `run.yaml` | Otter 解析后的唯一、不可变执行配置 |
| `manifest.json` | run、project、资产、工具、环境和 provenance 摘要 |
| `input/` | 输入清单与快照元数据；默认不复制 FASTQ |
| `work/` | executor 临时文件；不同 run 不共享 |
| `results/` | 科学产物；相对布局不得依赖 executor/backend/toolchain |
| `logs/` | Otter、executor、submission 和 task 日志 |
| `state/` | Otter task、Craftmake SQLite 或 Snakemake 兼容状态 |
| `metrics/` | sbatch accounting、任务 timing 和 benchmark 数据 |
| `artifacts.json` | 产物 schema、路径、checksum、比较策略和发布状态 |

## 隔离与发布

- 双轨或 benchmark 的每个 matrix cell 必须使用不同 run ID。
- 不允许 Snakemake 与 Craftmake 共享 `work/`、状态文件或会影响性能的缓存。
- 科学结果先写 `work/`，验证后发布到 `results/`；发布失败不得覆盖已有结果。
- 项目根只维护 run 索引，不使用“latest”可变目录作为权威来源；可提供只读 latest 指针。
- 删除 run 需要显式命令和 provenance 记录，不能由新 run 自动清理旧证据。
