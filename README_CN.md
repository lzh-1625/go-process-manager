<div align="center">

# GPM

### 🚀 轻量级、高性能的进程与任务管理平台

一个集 **进程管理、任务编排、日志收集、事件审计、状态推送、权限控制** 于一体的通用服务管理平台。

支持 Web UI / CLI / Webhook / cgroup 限制 / 日志检索 / 终端共享。

<p>
  <img src="https://img.shields.io/badge/Go-1.20+-00ADD8?style=for-the-badge&logo=go" />
  <img src="https://img.shields.io/badge/Linux-Supported-success?style=for-the-badge&logo=linux" />
  <img src="https://img.shields.io/badge/WebUI-Built--in-blue?style=for-the-badge" />
  <img src="https://img.shields.io/badge/License-MIT-orange?style=for-the-badge" />
</p>

</div>

---

# ✨ 项目介绍

GPM（General Process Manager）是一个面向服务、脚本、任务流场景设计的现代化进程管理平台。

相比传统的 `nohup`、`pm2`、`supervisor` 等工具，GPM 提供：

- 更完整的任务编排能力
- 更强的日志收集能力
- 更灵活的状态推送机制
- 更现代化的 Web UI
- 更完善的权限体系
- 更丰富的事件审计能力

适用于：

- 服务托管
- 自动化脚本运行
- AI 推理服务
- 游戏服务器
- 爬虫任务
- DevOps 自动化
- CI/CD 辅助任务
- 远程终端共享

---

# 📦 功能特性

## 🖥️ 进程管理

通过启动命令、工作目录、环境变量等信息在 GPM 中启动进程。

GPM 将自动完成：

- 进程生命周期管理
- 输出日志收集
- 状态监听
- cgroup 资源限制
- 性能信息采集
- 状态推送

### 支持能力

- Web UI 在线管理
- CLI 命令行交互
- 在线终端访问
- 临时终端分享链接
- 进程实时状态查看
- 资源占用监控

---

## ⚙️ 任务管理

GPM 提供灵活的自动化任务系统。

### 支持触发方式

- 定时触发
- API 触发
- 任务联动触发
- 进程状态变更触发

### 支持能力

- 自动启动/停止进程
- 条件判断
- 任务链编排
- 工作流执行
- 后续任务触发

---

## 📜 日志收集

自动收集进程日志并统一存储。

### 支持存储后端

- SQLite（默认）
- Elasticsearch
- Bleve（非 slim 版本）

### 提供能力

- Web 日志查询
- 日志检索
- 历史日志存储
- 多进程日志统一管理
- 全文索引(es、bleve)

---

## 📡 状态推送

支持通过 Webhook 推送进程或任务状态。

### 支持方式

- GET Webhook
- POST Webhook

### 特性

- 自定义占位符
- 自定义推送内容
- 状态变更通知

---

## 📋 事件管理

记录系统中的关键行为。

### 支持记录

- 进程状态变更
- 任务状态变更
- 终端访问事件
- 用户操作行为

### 提供能力

- 事件查询
- 审计记录
- 定时清理

---

## 👥 用户与权限管理

提供多角色权限体系。

| 角色 | 权限 |
|---|---|
| root | 全部权限 |
| admin | 除创建/删除进程、角色管理、系统设置外的全部权限 |
| user | 可配置允许访问的进程与日志 |

---

## ⚙️ 系统设置

支持通过：

- `config.json`
- Web UI

进行系统配置。

### 可配置项

- 重启次数
- 存储引擎
- 超时时长
- 日志等级
- 系统行为参数

---

# 🌐 在线 Demo

## Demo 地址

http://106.54.154.228:8790/

## 测试账号

| 用户名 | 密码 |
|---|---|
| root | root |

---

# 🚀 快速开始

# 1️⃣ 下载二进制文件

```bash
wget <release-url>
````

---

# 2️⃣ 添加执行权限

```bash
chmod +x ./gpm
```

---

# 3️⃣ 启动服务

## 直接启动

```bash
./gpm run
```

## 注册为系统服务

```bash
./gpm service install
./gpm service start
```

注册为服务后，可通过 `gpm` 命令访问后台服务信息。

---

# 📂 数据存储位置

| 类型   | 路径                        |
| ---- | ------------------------- |
| 配置文件 | `{用户目录}/.gpm/config.json` |
| 数据库  | `{用户目录}/.gpm/data.db`     |
| 日志文件 | `{用户目录}/.gpm/info.log`    |

---

# 💻 CLI 使用

确保 GPM 已在后台运行：

```bash
./gpm -h
```

查看帮助命令。

---

# 🏗️ 架构特点

* 轻量级部署
* 单二进制运行
* Web UI 内置
* 多用户权限体系
* 高扩展性
* 自动日志采集
* 任务流支持
* 多存储后端
* Webhook 集成
* 事件审计能力

---

# 📌 使用场景

## AI 服务托管

* LLM 推理服务
* Stable Diffusion
* 模型调度

## DevOps

* 自动化脚本
* CI/CD 任务
* 服务巡检

## 游戏服务

* Minecraft
* Rust
* Palworld
* Steam 游戏服务器

## 运维场景

* 后台服务管理
* 日志统一管理
* 远程终端共享

---

# 🔐 安全性

* RBAC 权限控制
* 用户隔离
* 操作审计
* 终端访问记录
* 事件日志追踪

---

# 🤝 贡献

欢迎提交：

* Issue
* Pull Request
* Feature Request

一起让 GPM 变得更强大。

---

# ⭐ Star History

如果这个项目对你有帮助，欢迎点一个 Star ⭐

---

# 📄 License

MIT License

---

<div align="center">

### GPM — 一个真正现代化的进程管理平台 🚀

</div>