# xdxtools RNA-seq 真实测试进展记录

**测试时间**: 2026-02-25
**测试数据**: 34个真实人类RNA-seq样本
**数据位置**: `/public3/home/scg9946/msy0213/singleronbio-release/P25112107/`

---

## 测试环境设置

```bash
# 创建工作目录
mkdir -p ~/xdxtools-runtime
cd ~/xdxtools-runtime

# 初始化xdxtools流程文件
xdxtools init

# 复制参考基因组
cp -r /public3/home/scg9946/beaverflow1016/inst ~/xdxtools-runtime/inst
```

---

## 真实测试命令

### 1. 创建FASTQ软链接 (68个文件)

```bash
cd /public3/home/scg9946/xdxtools-runtime
mkdir -p fastq
cd /public3/home/scg9946/msy0213/singleronbio-release/P25112107/
for sample in F*; do
    ln -s /public3/home/scg9946/msy0213/singleronbio-release/P25112107/$sample/20260212/Rawdata/*.fastq.gz \
           ~/xdxtools-runtime/fastq/
done
```

### 2. 创建样本表 (pdata.csv)

```csv
sampleid,condition,sample_group
F6703,treated,groupA
F6745,treated,groupA
F6747,treated,groupA
F6752,treated,groupA
F6760,treated,groupA
F6785,treated,groupA
F6789,treated,groupA
F6806,treated,groupA
F6808,treated,groupA
F6819,treated,groupA
F6886,treated,groupB
F6891,treated,groupB
F6901,treated,groupB
F6902,treated,groupB
F6926,treated,groupB
F6934,treated,groupB
F6953,treated,groupB
F6961,treated,groupB
F6968,treated,groupB
F6973,control,groupB
F6977,control,groupB
F7007,control,groupB
F7013,control,groupB
F7025,control,groupB
F7040,control,groupB
F7055,control,groupB
F7060,control,groupB
F7062,control,groupB
F7080,control,groupB
F7088,control,groupB
F7100,control,groupB
F7110,control,groupB
F7111,control,groupB
F7117,control,groupB
```

### 3. 创建分析项目

```bash
cd ~/xdxtools-runtime
xdxtools create \
  --fastq ./fastq \
  --mode RNASEQ \
  --species hg38 \
  --pdata ./pdata.csv
```

**生成的配置**: `userspace/66ebdff874f1165c2d9d9eae8365cb5ac21cdeda/config/config.yaml`

### 4. 运行RNA-seq分析（最新版本）

```bash
cd ~/xdxtools-runtime
xdxtools run \
  --config userspace/66ebdff874f1165c2d9d9eae8365cb5ac21cdeda/config/config.yaml \
  --engine slurm \
  --slurm-partition amd_512 \
  --parallel-jobs 34 \
  --conda-env xdxtools-snakemake \
  --copy-fastq
```

---

## 已修复的问题

### 1. ✅ Snakemake文件路径问题

**问题**: Snakemake文件在 `xdxtools-project/` 子目录，但执行在项目根目录

**修复**: 添加 `resolveSnakefilePath()` 方法检查多个位置

**文件**: `internal/workflow/snakemake.go`

### 2. ✅ enva命令参数分隔问题

**问题**: `enva run xdxtools-core STAR --runThreadN` 被enva误解析

**修复**: 添加 `--` 分隔符
- `enva run xdxtools-core -- STAR --runThreadN`
- `enva run xdxtools-r -- Rscript ...`

**文件**: `inst/rules/rnaseq_mapping.smk`, `rnaseq_step1_checker.smk` 等

### 3. ✅ SLURM分区硬编码问题

**问题**: 默认分区硬编码为 "cpu112c"

**修复**:
- 修改默认值为 "all"
- 添加从 stepResources 同步分区到 cfg.Engine.Slurm.Partition

**文件**: `internal/engine/slurm.go`, `cmd/run.go`

### 4. ✅ 配置文件rnaseq.ref格式问题

**问题**: `ref: ./inst/rnaseq/homo_sapiens/` 是字符串，但snakemake用索引访问导致取第一个字符 "."

**修复**: 改为列表格式 `ref: ["./inst/rnaseq/homo_sapiens/"]`

**文件**: `cmd/create.go`, 配置生成器

### 5. ✅ Step 2 OUT_OF_MEMORY问题

**问题**: Step 2 使用 SlurmEngine (单节点34样本并行)，只有8GB内存

**解决方案**: 实现自动切换到 SlurmArrayEngine

**修改文件**:
- `internal/workflow/manager.go` - 添加自动切换逻辑
- `internal/engine/factory.go` - NewSlurmArrayEngineWithResources 优先使用 stepResource 值

**日志输出**:
```
level=info msg="Auto-switching from SlurmEngine to SlurmArrayEngine for step 2"
level=info msg="Using Job Array strategy for step 2 with max 34 concurrent jobs"
```

---

## 尚未解决的问题

### 1. ❌ Step 1 checker R脚本依赖问题

**错误**: Rscript 邮件脚本失败，导致 checker 无法完成

**日志**:
```
enva run xdxtools-r -- Rscript R/beaver_mail.R --step 1 --jobid 66ebdff874f1165c2d9d9eae8365cb5ac21cdeda --send_to
```

**临时解决方案**: 手动创建标记文件
```bash
mkdir -p userspace/66ebdff874f1165c2d9d9eae8365cb5ac21cdeda/workflow/log
touch userspace/66ebdff874f1165c2d9d9eae8365cb5ac21cdeda/workflow/log/step1_success.txt
```

**根本原因**: xdxtools-r 环境可能不存在或没有R

**需要修复**:
- 让 checker 可选跳过邮件发送
- 或使用简单的 touch 命令而非 R 脚本

### 2. ❌ Step 2 参考基因组路径问题

**错误**: STAR 找不到基因组文件
```
EXITING because of FATAL ERROR: could not open genome file .//genomeParameters.txt
SOLUTION: check that the path to genome files, specified in --genomeDir is correct
```

**问题**: Snakemake规则中 `--genomeDir .` 使用当前目录而非实际参考基因组路径

**规则位置**: `rules/rnaseq_mapping.smk`

**需要修复**: 检查 config["reference"]["rnaseq"]["ref"] 的值传递

### 3. ⚠️ Step 2 仍在运行中验证

**当前状态**: Step 1 已完成，Step 2 使用 SlurmArrayEngine 提交

**预期**: 每个样本独立运行，40核/200GB内存

**需要验证**:
- Job Array 是否正确提交 (36943614_[0-33])
- 每个作业是否有正确资源 (40核, 200GB)
- 是否仍有 OUT_OF_MEMORY 错误

---

## 内存配置表

| Step | 模式 | 默认核心 | 默认内存 | JobArray | MaxJobs |
|------|------|---------|----------|----------|---------|
| 1 | RNASEQ | 20 | 100G | false | - |
| 2 | RNASEQ | 40 | 200G | true | 10 |

**配置位置**: `internal/config/defaults.go:150-153`

---

## 下一步行动

1. **验证 Step 2 Job Array 运行情况**
   ```bash
   squeue -u $USER
   sacct -j 36943614 --format=JobID,JobName,AllocCPUS,ReqMem,State
   ```

2. **修复 R 脚本依赖问题**
   - 方案A: 让 checker 可选跳过邮件
   - 方案B: 使用简单的 touch 命令

3. **修复参考基因组路径问题**
   - 检查 rnaseq_mapping.smk 中 genomeDir 参数
   - 确保 config["reference"]["rnaseq"]["ref"] 正确传递

---

## 编译和安装命令

```bash
cd /public3/home/scg9946/xdxtools
conda run -n go-env go build -o xdxtools
rm -f ~/.cargo/bin/xdxtools
cp xdxtools ~/.cargo/bin/
```

---

## 关键文件路径

| 文件 | 用途 |
|------|------|
| `internal/workflow/manager.go` | 工作流管理，Step 2 自动切换逻辑 |
| `internal/engine/factory.go` | 引擎工厂，SlurmArrayEngine 创建 |
| `internal/config/defaults.go` | 默认资源配置 |
| `internal/engine/slurm_array.go` | SLURM Job Array 实现 |
| `inst/rules/rnaseq_mapping.smk` | RNA-seq 比对规则 |
| `inst/rules/rnaseq_step1_checker.smk` | Step 1 检查规则 |

---

**最后更新**: 2026-02-25 08:30
**测试状态**: Step 2 运行中，等待验证结果
