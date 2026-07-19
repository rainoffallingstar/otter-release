# xdxtools 用户手册

> 写给生信新手和湿实验人员的完整操作指南

---

## 🧬 一句话介绍

**xdxtools** 是一套面向生物信息分析的一站式流程工具，帮助你用 **三条命令** 完成 RRBS（甲基化）、WGBS（全基因组甲基化）、RNA-seq（转录组）和 PDX（肿瘤移植模型）等多种测序数据的全流程分析——即使你从来没有写过代码，也可以上手。

---

## 📚 章节导航

| # | 章节 | 你想做什么 |
|---|------|-----------|
| 📦 | [第一章：安装指南](01-installation.md) | 第一次安装 xdxtools |
| 📁 | [第二章：数据准备](02-data-preparation.md) | 整理 FASTQ 文件和样本信息表 |
| 🚀 | [第三章：快速上手](03-quickstart.md) | 用三条命令跑起来一个分析 |
| 🔬 | [第四章：分析模式详解](04-analysis-modes.md) | 了解 RRBS/WGBS/RNA-seq/PDX 的区别 |
| ⚙️ | [第五章：高级用法](05-advanced-usage.md) | 集群投递、断点续传、参数调优 |
| 🛠️ | [第六章：子工具参考手册](06-subtools.md) | 单独使用 enva/xenofilter/qctb 等工具 |
| ❓ | [第七章：常见问题与排查](07-faq.md) | 遇到报错，查这里 |

---

## 🗺️ 我该从哪里开始？

```
你是第一次使用 xdxtools 吗？
        │
        ▼
    ┌── 是 ──► 先看 📦 第一章：安装指南
    │
    └── 否
        │
        ▼
  xdxtools 已经装好了？
        │
   ┌─── 是 ──► 直接看 🚀 第三章：快速上手
   │
   └─── 否 ──► 先看 📦 第一章：安装指南
                         │
                         ▼
              安装完成后，看 📁 第二章：数据准备
                         │
                         ▼
              数据整理好后，看 🚀 第三章：快速上手
```

**遇到报错？** 直接跳到 [❓ 第七章：常见问题与排查](07-faq.md)

---

## 🧰 工具清单总览

### 主工具

| 工具 | 用途 |
|------|------|
| `xdxtools` | 主命令，管理完整的分析工作流（init / create / run / status） |

### 8 个子工具

| 子工具 | 用途 | 文档位置 |
|--------|------|---------|
| 🌿 `enva` | rattler 优先的环境管理器，可统一管理和接管历史 conda 环境 | [第六章 §1](06-subtools.md#enva) |
| 🧹 `xenofilter` | PDX 样本物种污染过滤（去掉鼠源读段） | [第六章 §2](06-subtools.md#xenofilter) |
| 🔗 `paireads` | 配对 reads 过滤，用于 PDX 双 BAM 处理 | [第六章 §3](06-subtools.md#paireads) |
| 📊 `htseq2matrix` | 将 HTSeq 计数文件合并为表达矩阵 | [第六章 §4](06-subtools.md#htseq2matrix) |
| 💎 `methrix-cli` | 甲基化数据的 HDF5 提取、处理与 QC 报告 | [第六章 §5](06-subtools.md#methrix-cli) |
| 📈 `qctb` | 质控报告生成器（BS-seq 和 RNA-seq 两种模式） | [第六章 §6](06-subtools.md#qctb) |
| ⚡ `fqc` | 快速 FASTQ 质控（FastQC 的 Rust 替代品） | [第六章 §7](06-subtools.md#fqc) |
| 🔀 `gomats` | RNA 可变剪接分析（rMATS 流程编排器） | [第六章 §8](06-subtools.md#gomats) |

---

## 💡 快速参考

| 场景 | 命令 |
|------|------|
| 查看版本 | `xdxtools --version` |
| 新建项目 | `xdxtools init my_project` |
| 扫描 FASTQ 并生成配置 | `xdxtools create --fastq /data/fastq --mode RRBS --pdata samples.xlsx --output my_project/userspace --jobid demo_rrbs` |
| 在 SLURM 上后台运行 | `xdxtools run --config my_project/userspace/demo_rrbs/config/config.yaml --engine slurm` |
| 查看后台任务 | `xdxtools task list` |
| 查看任务状态 | `xdxtools task status <task-id>` |
| 跟随任务日志 | `xdxtools task logs <task-id> --follow` |
| 停止后台任务 | `xdxtools task stop <task-id>` |
| 前台运行 | `xdxtools run --config my_project/userspace/demo_rrbs/config/config.yaml --foreground` |
| 查看项目工作流进度 | `xdxtools status my_project/userspace/demo_rrbs` |
| 验证配置文件 | `xdxtools config validate --config my_project/userspace/demo_rrbs/config/config.yaml` |

---

## 📞 获取帮助

- 查看命令帮助：`xdxtools --help` 或 `xdxtools <命令> --help`
- 提交问题：[GitHub Issues](https://github.com/rainoffallingstar/xdxtools-go/issues)
- 常见报错排查：[第七章 FAQ](07-faq.md)
