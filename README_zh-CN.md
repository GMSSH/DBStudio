# 🗄️ GM Database — 数据库管理器

> A modern, web-based MySQL management tool for [GMSSH](https://gmssh.com). Lightweight, fast, and built for server operations.

<p align="center">
  <img src="www/public/icon.png" width="96" alt="GM Database Icon" />
</p>

<p align="center">
  <strong>GM Database</strong> 是 GMSSH 桌面系统的数据库管理插件，提供类似 Navicat / DBeaver 的可视化数据库操作体验，专为服务器运维场景设计。
</p>

<p align="center">
  <a href="README.md">📖 English</a>
</p>

---

## ✨ 功能特性

### 连接管理
- 🔐 AES-256-GCM 加密存储连接凭证
- 🔌 多连接并行管理，连接池自动回收
- ✅ 连接测试，一键保存常用连接

### 数据操作
- 📊 分页浏览表数据（Grid View）
- 📝 行内编辑 — 直接在表格中修改数据
- 📋 表单模式（Form View）— 单条记录纵向查看与编辑
- ➕ 新增行 / 🗑️ 删除行，批量提交变更
- 🔍 支持主键与非主键表的安全 CRUD

### SQL 查询
- ⚡ SQL 编辑器（基于 CodeMirror 6）
- 📜 查询历史记录，支持回溯执行
- 🎯 支持多语句执行、`Ctrl+Enter` 快速运行

### 表结构设计
- 🏗️ 可视化表结构编辑器（列、索引、外键）
- 📝 DDL 预览与实时 diff 对比
- 🔄 ALTER TABLE 自动生成

### 导入 / 导出
- 📤 数据库导出（mysqldump 兼容 + 纯 Go 兼容模式）
- 📥 SQL 文件导入，支持自动建库
- 📑 表级数据导出（CSV / JSON / SQL 格式）
- 📋 任务管理面板，实时进度跟踪

### 数据库概览
- 📈 表/视图列表、行数、大小、引擎信息
- 🔎 对象搜索过滤
- 📄 DDL 快速预览与复制

### 国际化
- 🌐 中文 / English 双语支持
- 🔤 自动跟随 GMSSH 系统语言

---

## 🏗️ 技术架构

```
┌─────────────────────────────────────────┐
│           GMSSH Desktop (Host)          │
│  ┌───────────────────────────────────┐  │
│  │         Iframe (Plugin)           │  │
│  │  ┌─────────┐    ┌─────────────┐  │  │
│  │  │   Vue 3  │◄──►│  Go Backend │  │  │
│  │  │ Naive UI │    │  JSON-RPC   │  │  │
│  │  └─────────┘    └──────┬──────┘  │  │
│  └─────────────────────────┼─────────┘  │
│                            │            │
│                    Unix Socket          │
│                    (app.sock)           │
└─────────────────────────────────────────┘
                     │
              ┌──────▼──────┐
              │    MySQL     │
              │   Server     │
              └─────────────┘
```

| 层 | 技术栈 |
|---|--------|
| **前端** | Vue 3 + Naive UI + CodeMirror 6 + Pinia |
| **后端** | Go + simplejrpc-go (JSON-RPC over Unix Socket) |
| **通信** | GMSSH 代理 → Unix Socket → JSON-RPC |
| **存储** | AES-256-GCM 加密 JSON（连接配置） |

---

## 📁 项目结构

```
dbmanager/
├── backend/                 # Go 后端
│   ├── handler/rpc/         # JSON-RPC 路由层
│   ├── service/             # 业务逻辑层
│   ├── pkg/
│   │   ├── config/          # 连接存储（加密）
│   │   ├── db/              # 数据库抽象层
│   │   ├── files/           # 文件工具
│   │   └── rpcutil/         # RPC 工具
│   ├── i18n/                # 后端国际化
│   ├── Makefile             # 构建脚本
│   └── main.go              # 入口
│
├── www/                     # Vue 前端
│   ├── src/
│   │   ├── components/      # 核心组件
│   │   ├── views/           # 页面布局
│   │   ├── stores/          # Pinia 状态
│   │   ├── styles/          # CSS 设计系统
│   │   ├── i18n/            # 前端国际化
│   │   └── utils/           # API 客户端 & SDK 封装
│   └── vite.config.js
│
└── docs/                    # 设计文档
    └── designs/             # 设计规范 (tokens, components)
```

---

## 🚀 快速开始

### 环境要求

- **Go** ≥ 1.21
- **Node.js** ≥ 18 + **pnpm**
- **GMSSH** 桌面系统（运行环境）

### 开发模式

```bash
# 1. 前端开发
cd www
pnpm install
pnpm dev          # http://localhost:5173

# 2. 后端开发
cd backend
go mod tidy
go run main.go
```

### 构建发布

```bash
cd backend

make help         # 查看所有构建目标
make amd64        # 构建 linux/amd64
make arm64        # 构建 linux/arm64
make package      # 构建前端 + 双架构打包 (tar.gz)
```

构建产物结构：
```
build/linux-amd64/
├── app/
│   ├── bin/      # 后端二进制 + config + i18n
│   └── www/      # 前端静态文件
├── data/         # 运行时数据
├── logs/         # 日志
└── tmp/          # 临时文件
```

---

## 🔌 GMSSH 集成

GM Database 作为 GMSSH 外置应用运行，通过以下接口与宿主通信：

```javascript
// 前端通过 window.$gm 与 GMSSH 交互
window.$gm.request(url, options)   // API 请求代理
window.$gm.execShell({ cmd })      // 远程命令执行
window.$gm.chooseFile(callback)    // 文件选择器
window.$gm.message.success(msg)    // UI 反馈
window.$gm.userName                // 当前登录用户
```

---

## 📸 Screenshots

> *Coming soon*

---

## 🤝 Contributing

欢迎贡献代码！请遵循以下规范：

1. Fork 本仓库
2. 创建特性分支：`git checkout -b feature/your-feature`
3. 提交变更：`git commit -m 'feat: add some feature'`
4. 推送分支：`git push origin feature/your-feature`
5. 创建 Pull Request

### 开发规范

- **后端**：分层架构 `handler → service → pkg`，错误信息国际化
- **前端**：遵循 `docs/designs/` 下的设计规范（tokens, 字体, 间距）
- **提交**：使用 [Conventional Commits](https://www.conventionalcommits.org/)

---

## 📄 License

[MIT](LICENSE) © [GMSSH](https://gmssh.com)

---

<p align="center">
  <sub>Built with ❤️ for the GMSSH ecosystem</sub>
</p>
