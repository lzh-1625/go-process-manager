# GPM
### 🚀 Lightweight & High-Performance Process and Task Management Platform
An all-in-one platform integrating **process management, task orchestration, log collection, event auditing, status push, and access control**.

Supports Web UI / CLI / Webhook / cgroup resource limiting / log retrieval / terminal sharing.

<p>
  <img src="https://img.shields.io/badge/Go-1.20+-00ADD8?style=for-the-badge&logo=go" />
  <img src="https://img.shields.io/badge/Linux-Supported-success?style=for-the-badge&logo=linux" />
  <img src="https://img.shields.io/badge/WebUI-Built--in-blue?style=for-the-badge" />
  <img src="https://img.shields.io/badge/License-MIT-orange?style=for-the-badge" />
</p>

## [中文](./README_CN.md)
---

# ✨ Project Introduction
GPM (General Process Manager) is a modern process management platform designed for service, script and task workflow scenarios.

Compared with traditional tools such as `nohup`, `pm2` and `supervisor`, GPM delivers:
- More comprehensive task orchestration capabilities
- Powerful log collection
- Flexible status push mechanism
- Modern built-in Web UI
- Complete permission system
- Rich event auditing features

## Use Cases
- Service hosting
- Automated script execution
- AI inference services
- Game servers
- Crawler tasks
- DevOps automation
- CI/CD auxiliary tasks
- Remote terminal sharing

---

# 📦 Core Features
## 🖥️ Process Management
Start processes in GPM with custom launch commands, working directories, environment variables and more.

GPM automatically handles:
- Process lifecycle management
- Output log collection
- Status monitoring
- cgroup resource limiting
- performance metrics collection
- status event push

### Capabilities
- Web UI visual management
- CLI command-line interaction
- In-browser terminal access
- Temporary terminal share links
- Real-time process status overview
- Resource usage monitoring

---

## ⚙️ Task Management
GPM features a flexible automated task system.

### Trigger Modes
- Scheduled trigger
- API trigger
- Task linkage trigger
- Process status change trigger

### Capabilities
- Auto start/stop processes
- Conditional execution
- Task chain orchestration
- Workflow execution
- Subsequent task triggering

---

## 📜 Log Collection
Automatically collect and centrally store process logs.

### Supported Storage Backends
- SQLite (Default)
- Elasticsearch
- Bleve (Non-slim version)

### Capabilities
- Web-based log query
- Full-text log search
- Historical log archiving
- Unified log management for multiple processes
- Full-text indexing (Elasticsearch / Bleve)

---

## 📡 Status Push
Push process and task status via Webhook.

### Supported Methods
- GET Webhook
- POST Webhook

### Features
- Custom placeholder variables
- Custom push payload
- Real-time status change notifications

---

## 📋 Event Auditing
Record critical system activities.

### Monitored Events
- Process status changes
- Task status changes
- Terminal access records
- User operation logs

### Capabilities
- Event query & filtering
- Operation audit trail
- Automatic log cleanup

---

## 👥 User & Access Control
Built-in multi-role RBAC permission system.

| Role | Permissions |
|---|---|
| root | Full system privileges |
| admin | All privileges except process creation/deletion, role management and system configuration |
| user | Access only assigned processes and logs |

---

## ⚙️ System Configuration
Configure via:
- `config.json`
- Web UI panel

### Configurable Items
- Max restart attempts
- Log storage engine
- Timeout thresholds
- Log level
- Global system parameters

---

# 🌐 Online Demo
## Demo Address
http://106.54.154.228:8790/

## Test Account
| Username | Password |
|---|---|
| root | root |

---

# 🚀 Quick Start
## 1. Download Binary
```bash
wget <release-url>
```

## 2. Grant Execute Permission
```bash
chmod +x ./gpm
```

## 3. Launch Service
### Run Directly
```bash
./gpm run
```

### Install as System Service
```bash
./gpm service install
./gpm service start
```
After installation, use the `gpm` CLI to interact with the background service.

---

# 📂 Data Storage Paths
| Type | Path |
|---|---|
| Config File | `{User Home}/.gpm/config.json` |
| Database | `{User Home}/.gpm/data.db` |
| Log File | `{User Home}/.gpm/info.log` |

---

# 💻 CLI Usage
Ensure the GPM backend service is running:
```bash
./gpm -h
```
View full command documentation.

---

# 🏗️ Architecture Highlights
- Lightweight deployment
- Single binary distribution
- Built-in Web UI
- Multi-user RBAC system
- Highly extensible design
- Automatic log harvesting
- Native workflow support
- Multiple storage backends
- Webhook integration
- Complete event auditing

---

# 📌 Application Scenarios
## AI Service Hosting
- LLM inference services
- Stable Diffusion deployment
- Model scheduling & orchestration

## DevOps
- Automated task scripts
- CI/CD pipeline jobs
- Service health inspection

## Game Servers
- Minecraft
- Rust
- Palworld
- Steam game servers

## Operation & Maintenance
- Background service management
- Centralized log management
- Remote terminal collaboration

---

# 🔐 Security
- RBAC access control
- User resource isolation
- Full operation audit logs
- Terminal access recording
- Event traceability

---

# 🤝 Contribution
Welcome to submit:
- Issues
- Pull Requests
- Feature Requests

Let’s build GPM together.

---

# ⭐ Star History
If this project helps you, please give it a Star ⭐

---

# 📄 License
MIT License

---

<div align="center">

### GPM — The Modern All-in-One Process Management Platform 🚀

</div>