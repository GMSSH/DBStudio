# 🗄️ DBStudio — Database Manager

> A modern, web-based MySQL management tool for [GMSSH](https://gmssh.com). Lightweight, fast, and built for server operations.

<p align="center">
  <img src="www/public/icon.png" width="96" alt="DBStudio Icon" />
</p>

<p align="center">
  <strong>DBStudio</strong> is a database management plugin for the GMSSH desktop system, providing a visual database experience similar to Navicat / DBeaver, purpose-built for server administration.
</p>

<p align="center">
  <a href="README_zh-CN.md">📖 中文文档</a>
</p>

---

## ✨ Features

### Connection Management
- 🔐 AES-256-GCM encrypted credential storage
- 🔌 Multiple concurrent connections with automatic pool recycling
- ✅ Connection testing & one-click save

### Data Operations
- 📊 Paginated data browsing (Grid View)
- 📝 Inline editing — modify data directly in the grid
- 📋 Form View — view and edit individual records vertically
- ➕ Add rows / 🗑️ Delete rows with batch commit
- 🔍 Safe CRUD for tables with and without primary keys

### SQL Query
- ⚡ SQL Editor powered by CodeMirror 6
- 📜 Query history with re-execution support
- 🎯 Multi-statement execution, `Ctrl+Enter` shortcut

### Table Structure Designer
- 🏗️ Visual structure editor (columns, indexes, foreign keys)
- 📝 DDL preview with real-time diff comparison
- 🔄 Auto-generated ALTER TABLE statements

### Import / Export
- 📤 Database export (mysqldump compatible + pure Go fallback)
- 📥 SQL file import with automatic database creation
- 📑 Table-level data export (CSV / JSON / SQL formats)
- 📋 Task management panel with real-time progress tracking

### Database Overview
- 📈 Table/view listing with row count, size, and engine info
- 🔎 Object search & filter
- 📄 Quick DDL preview & copy

### Internationalization
- 🌐 Chinese / English dual-language support
- 🔤 Auto-detects GMSSH system language

---

## 🏗️ Architecture

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

| Layer | Tech Stack |
|-------|-----------|
| **Frontend** | Vue 3 + Naive UI + CodeMirror 6 + Pinia |
| **Backend** | Go + simplejrpc-go (JSON-RPC over Unix Socket) |
| **Communication** | GMSSH proxy → Unix Socket → JSON-RPC |
| **Storage** | AES-256-GCM encrypted JSON (connection configs) |

---

## 📁 Project Structure

```
DBStudio/
├── backend/                 # Go backend
│   ├── handler/rpc/         # JSON-RPC route handlers
│   ├── service/             # Business logic layer
│   ├── pkg/
│   │   ├── config/          # Connection store (encrypted)
│   │   ├── db/              # Database abstraction layer
│   │   ├── files/           # File utilities
│   │   └── rpcutil/         # RPC utilities
│   ├── i18n/                # Backend i18n
│   ├── Makefile             # Build script
│   └── main.go              # Entry point
│
├── www/                     # Vue frontend
│   ├── src/
│   │   ├── components/      # Core components
│   │   ├── views/           # Page layouts
│   │   ├── stores/          # Pinia state management
│   │   ├── styles/          # CSS design system
│   │   ├── i18n/            # Frontend i18n
│   │   └── utils/           # API client & SDK wrappers
│   └── vite.config.js
│
└── docs/                    # Design docs
    └── designs/             # Design specs (tokens, components)
```

---

## 🚀 Getting Started

### Prerequisites

- **Go** ≥ 1.21
- **Node.js** ≥ 18 + **pnpm**
- **GMSSH** desktop system (runtime environment)

### Development

```bash
# 1. Frontend dev server
cd www
pnpm install
pnpm dev          # http://localhost:5173

# 2. Backend dev server
cd backend
go mod tidy
go run main.go
```

### Build & Package

```bash
cd backend

make help         # Show all build targets
make amd64        # Build linux/amd64
make arm64        # Build linux/arm64
make package      # Build frontend + dual-arch packages (tar.gz)
```

Build output structure:
```
build/linux-amd64/
├── app/
│   ├── bin/      # Backend binary + config + i18n
│   └── www/      # Frontend static files
├── data/         # Runtime data
├── logs/         # Logs
└── tmp/          # Temporary files
```

---

## 🔌 GMSSH Integration

DBStudio runs as a GMSSH external application and communicates with the host via:

```javascript
// Frontend communicates with GMSSH through window.$gm
window.$gm.request(url, options)   // API request proxy
window.$gm.execShell({ cmd })      // Remote command execution
window.$gm.chooseFile(callback)    // File picker
window.$gm.message.success(msg)    // UI feedback
window.$gm.userName                // Current logged-in user
```

---

## 📸 Screenshots

> *Coming soon*

---

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork this repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Commit your changes: `git commit -m 'feat: add some feature'`
4. Push to the branch: `git push origin feature/your-feature`
5. Open a Pull Request

### Development Guidelines

- **Backend**: Layered architecture `handler → service → pkg`, i18n error messages
- **Frontend**: Follow design specs under `docs/designs/` (tokens, typography, spacing)
- **Commits**: Use [Conventional Commits](https://www.conventionalcommits.org/)

---

## 📄 License

[MIT](LICENSE) © [GMSSH](https://gmssh.com)

---

<p align="center">
  <sub>Built with ❤️ for the GMSSH ecosystem</sub>
</p>
