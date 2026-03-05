# 第七章：常见问题与排查

> **本章你将学到：**
> - 快速定位和解决常见报错
> - 查找和阅读日志文件
> - 处理特殊情况和边界场景

---

## 快速定位问题

```
遇到问题了？
      │
      ├── 工具找不到（command not found）    → 🔧 安装问题 §1
      ├── 安装过程报错                       → 🔧 安装问题 §2-3
      │
      ├── FASTQ 文件识别问题                 → 📁 数据准备问题 §1
      ├── 样本信息表读取失败                 → 📁 数据准备问题 §2
      │
      ├── SLURM 作业提交失败                 → ▶️ 运行问题 §1
      ├── 如何查看日志                       → ▶️ 运行问题 §2
      ├── 某步骤失败，需要重跑吗             → ▶️ 运行问题 §3
      │
      └── 找不到输出文件 / 理解结果          → 📊 结果问题 §1-2
```

---

## 🔧 安装问题

### Q1：`xdxtools: command not found` {#command-not-found}

**症状**：运行 `xdxtools` 后出现 `command not found` 或 `bash: xdxtools: 未找到命令`。

**原因**：工具没有被安装到当前 `PATH` 包含的目录中。

**解决步骤**：

```bash
# 第一步：找到 xdxtools 被安装在哪里
find $HOME -name "xdxtools" -type f 2>/dev/null
# 常见安装位置：~/.cargo/bin/xdxtools

# 第二步：将安装目录加入 PATH
echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# 第三步：验证
which xdxtools
xdxtools --version
```

如果 `find` 命令找不到 xdxtools，说明工具未安装成功，请重新运行安装脚本：

```bash
bash <(wget -qO- https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh)
```

---

### Q2：enva 安装失败 {#enva-install-fail}

**症状**：安装脚本报错 `failed to install enva` 或 `enva: command not found`。

**解决方法 1**：手动编译安装

```bash
git clone https://github.com/rainoffallingstar/enva.git
cd enva
conda run -n go-build go build -o ~/.cargo/bin/enva
```

**解决方法 2**：检查 Go 编译环境

```bash
# 检查 go-build conda 环境是否存在
conda env list | grep go-build

# 如果不存在，先创建
conda create -n go-build golang -y
```

**解决方法 3**：直接使用 conda 代替 enva

enva 是 conda 的替代品，如果 enva 无法安装，可以直接用 `conda run` 替代 `enva run`：

```bash
# enva run bismark bismark --help
conda run -n bismark bismark --help
```

---

### Q3：methrix 运行报 HDF5 错误 {#hdf5-error}

**症状**：运行 `methrix-cli` 时报错类似：

```
error while loading shared libraries: libhdf5.so.xxx: cannot open shared object file
```

**原因**：系统找不到 HDF5 动态库文件。

**解决步骤**：

```bash
# 第一步：找到 HDF5 库的位置
find $HOME -name "libhdf5*.so*" 2>/dev/null | head -5
# 常见位置：~/miniconda3/envs/rust_build/lib/

# 第二步：将库路径加入 LD_LIBRARY_PATH
# （将下面的路径替换为你实际找到的路径）
export HDF5_DIR="$HOME/miniconda3/envs/rust_build"
export LD_LIBRARY_PATH="$HDF5_DIR/lib:$LD_LIBRARY_PATH"

# 第三步：将配置写入 ~/.bashrc 使其永久生效
echo 'export HDF5_DIR="$HOME/miniconda3/envs/rust_build"' >> ~/.bashrc
echo 'export LD_LIBRARY_PATH="$HDF5_DIR/lib:$LD_LIBRARY_PATH"' >> ~/.bashrc
source ~/.bashrc

# 第四步：验证
methrix-cli --version
```

如果 HDF5 环境不存在，先安装：

```bash
conda create -n rust_build -c conda-forge hdf5 -y
```

---

## 📁 数据准备问题

### Q4：FASTQ 配对失败（找不到 R2）

**症状**：`xdxtools create` 报错 `Could not find R2 for sample: sample1` 或类似错误。

**诊断**：

```bash
# 查看 FASTQ 目录
ls ./fastq/

# 检查文件命名（看 R1 和 R2 的后缀）
ls ./fastq/ | grep -E "_R[12]|_[12]\."
```

**常见原因和修复**：

**原因 1**：R2 文件后缀格式不对

```bash
# 错误：sample1_R1.fastq.gz 配对的 R2 是 sample1_2.fastq.gz（格式不统一）
# 修复：统一为 _R2 格式
mv sample1_2.fastq.gz sample1_R2.fastq.gz
```

**原因 2**：样本名大小写不一致

```bash
# 错误：pdata 中是 Sample1，但文件名是 sample1
# 修复：保持一致（推荐小写）
mv Sample1_R1.fastq.gz sample1_R1.fastq.gz
mv Sample1_R2.fastq.gz sample1_R2.fastq.gz
# 同时修改 pdata 中的 sampleid 为 sample1
```

**原因 3**：文件不完整（测序公司只交付了 R1）

```bash
# 检查 R1 和 R2 数量是否一致
ls ./fastq/*_R1.fastq.gz | wc -l
ls ./fastq/*_R2.fastq.gz | wc -l
# 两个数字应该相同，如果不同，联系测序公司补发数据
```

---

### Q5：pdata 中文列名不识别

**症状**：`xdxtools create` 报错 `column 'sampleid' not found` 或类似错误，但你的 pdata 里确实有"样本编号"列。

**解决步骤**：

```bash
# 查看 pdata 的实际列名（检查是否有隐藏字符或编码问题）
python3 -c "
import pandas as pd
df = pd.read_excel('samples.xlsx')
print('列名列表:', df.columns.tolist())
print('前3行:', df.head(3))
"
```

支持的中文列名对照：

| 你的列名 | xdxtools 识别为 |
|---------|----------------|
| 样本编号 | sampleid |
| 样本ID | sampleid |
| 条件 | condition |
| 样本分组 | sample_group |
| 分组 | sample_group |

如果你的列名不在上述列表中（如"编号"、"ID"等），有两种解决方法：

**方法 1**：直接改成英文列名（推荐）

在 Excel 中将"编号"改为 `sampleid`，保存后重新运行。

**方法 2**：提 Issue 请求支持新的中文列名

访问 [GitHub Issues](https://github.com/rainoffallingstar/xdxtools-go/issues) 提交你的列名，我们会在下一版本添加支持。

---

## ▶️ 运行问题

### Q6：SLURM 作业提交失败

**症状**：`xdxtools run` 报错 `sbatch: error: ...` 或任务提交后立即消失。

**诊断步骤**：

```bash
# 查看 SLURM 分区信息
sinfo

# 测试手动提交一个简单任务
echo '#!/bin/bash
echo hello' | sbatch --partition your_partition

# 查看作业队列
squeue -u $(whoami)
```

**常见原因**：

| 错误信息 | 原因 | 解决 |
|---------|------|------|
| `Invalid partition name` | 分区名称不对 | 用 `sinfo` 查看正确分区名 |
| `Unable to allocate resources` | 资源不足 | 等待资源释放或减少 `--parallel-jobs` |
| `Job violates accounting/QOS policy` | 超出用户配额 | 联系集群管理员 |
| `sbatch: command not found` | SLURM 未安装 | 改用 `--engine local` |

```bash
# 如果没有 SLURM，改用本地运行
xdxtools run \
    --config config.yaml \
    --engine local \
    --parallel-jobs 4
```

---

### Q7：如何查看详细日志

日志文件保存在以下位置：

```
userspace/<jobid>/
└── logs/
    ├── step1/
    │   ├── sample1.log     ← 各样本的运行日志
    │   ├── sample2.log
    │   └── ...
    ├── step2/
    └── step3/
```

**查看最新日志**：

```bash
# 查看某个样本的日志
cat userspace/<jobid>/logs/step1/sample1.log

# 实时监控日志（类似 tail -f）
tail -f userspace/<jobid>/logs/step1/sample1.log

# 查找错误信息
grep -i "error\|failed\|exception" userspace/<jobid>/logs/step1/*.log
```

**SLURM 作业日志**：

```bash
# 查看 SLURM 作业输出
ls userspace/<jobid>/logs/slurm/

# 查看指定作业的日志
cat userspace/<jobid>/logs/slurm/slurm_<jobid>.out
cat userspace/<jobid>/logs/slurm/slurm_<jobid>.err
```

---

### Q8：某个 Step 失败了，需要从头重跑吗？

**不需要！** 使用 `--resume` 参数从失败处继续：

```bash
xdxtools run \
    --config userspace/<jobid>/config/config.yaml \
    --engine slurm \
    --resume
```

xdxtools 会自动：
1. 读取 `.xdxtools_state.json` 状态文件
2. 跳过已成功完成的步骤
3. 从失败或未完成的步骤继续

**如果单个样本失败，其他样本是否受影响？**

在并行模式下，**不影响**。其他样本会继续运行，只有失败的样本会被标记为需要重跑。

**如果需要强制重跑某一步**：

```bash
# 手动编辑状态文件，将对应步骤的状态改为 PENDING
# （注意：.xdxtools_state.json 在 userspace/<jobid>/ 目录下）
vi userspace/<jobid>/.xdxtools_state.json
```

---

## 📊 结果问题

### Q9：输出文件在哪里？

所有输出文件都在 `userspace/<jobid>/` 目录下：

```bash
# 查看你的 jobid
ls userspace/

# 查看该 job 的所有输出
ls userspace/<jobid>/results/
```

**各分析模式的关键输出**：

| 模式 | 关键输出 | 位置 |
|------|---------|------|
| RRBS/WGBS | 甲基化矩阵（HDF5） | `results/methrix/*.h5` |
| RRBS/WGBS | QC 报告（Excel） | `results/qc_report.xlsx` |
| RNA-seq | 基因表达矩阵 | `results/matrix_count.txt` |
| RNA-seq | 归一化矩阵 | `results/matrix_norm.txt` |
| RNA-seq | QC 报告 | `results/qc_report.xlsx` |
| PDX | 物种过滤统计 | `results/xenofilter/stats.txt` |

---

### Q10：如何理解 QC 报告？

QC 报告（`qc_report.xlsx`）包含多个工作表（Sheet），各有侧重：

**Sheet 1：Summary（总览）**

- 绿色：指标在正常范围内
- 黄色：指标偏低，建议关注
- 红色：指标异常，建议排查

**关键指标参考值**：

| 指标 | 正常范围 | 低于此值需注意 |
|------|---------|--------------|
| 比对率（Mapping Rate） | >80% | <60% 说明样本质量差或参考基因组选择有误 |
| 亚硫酸氢盐转化率（BS Conversion Rate） | >99% | <98% 说明转化不完全，可能影响甲基化结果 |
| 全局甲基化水平（Global Methylation） | 因组织而异（5-80%） | 接近 0% 或 100% 可能有问题 |
| CpG 覆盖数 | >1M（RRBS） | 过少说明数据量不够 |
| HTSeq 计数率（Assignment Rate） | >50% | <30% 说明 GTF 注释与 BAM 不匹配 |

**Sheet 2：Per-Sample（各样本详情）**

每行一个样本，方便比较各样本的质量差异，找出离群值。

---

## 💡 其他常见问题

### Q11：如何更新 xdxtools 到最新版本？

```bash
# 重新运行安装脚本（会自动检测最新版本）
bash <(wget -qO- https://raw.githubusercontent.com/rainoffallingstar/xdxtools-go/main/scripts/install.sh)

# 或指定版本
bash install.sh --version v1.x.x
```

### Q12：如何在不同服务器上共享分析结果？

```bash
# 打包整个 job 目录
tar -czf my_analysis.tar.gz userspace/<jobid>/

# 传输到另一台服务器
scp my_analysis.tar.gz user@server:/data/
```

### Q13：分析结果的磁盘占用太大，哪些文件可以删除？

| 文件类型 | 可以删除？ | 说明 |
|---------|----------|------|
| 中间 BAM 文件 | 是（如确认结果正确后） | 通常最大，可节省几十 GB |
| 原始计数文件（`.count`） | 是（已合并成矩阵后） | |
| CpG 覆盖文件（`.cov.gz`） | 是（已生成 HDF5 后） | |
| HDF5 文件（`.h5`） | 否 | 最终结果，请妥善保存 |
| 矩阵文件（`.txt`） | 否 | 最终结果 |
| QC 报告（`.xlsx`） | 否 | 需要存档 |

```bash
# 清理中间 BAM 文件（确认分析完成后执行）
rm -f userspace/<jobid>/results/alignment/*.bam
rm -f userspace/<jobid>/results/alignment/*.bai
```

---

## 📞 获取更多帮助

如果本章没有解决你的问题：

1. **查看工具帮助**：`xdxtools --help` 或 `xdxtools <命令> --help`
2. **查看详细日志**：参见 [Q7：如何查看详细日志](#q7如何查看详细日志)
3. **提交 Issue**：[GitHub Issues](https://github.com/rainoffallingstar/xdxtools-go/issues)
   - 请附上：报错信息、操作系统版本、xdxtools 版本、相关日志片段

---

**上一章：** [🛠️ 第六章：子工具参考手册](06-subtools.md)

**返回首页：** [📚 手册首页](README.md)
