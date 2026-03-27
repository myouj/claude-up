# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**PromptVault** - A prompt management system for organizing, versioning, testing, and optimizing AI prompts. Built with Go/Gin backend and Vue 3/Element Plus frontend.

## Architecture

```
prompt-vault/
├── backend/              # Go API server (Gin + GORM + SQLite)
│   ├── main.go          # Entry point, router setup
│   ├── handlers/        # HTTP request handlers (prompt, version, test)
│   ├── models/          # Data models (Prompt, PromptVersion, TestRecord)
│   └── prompt-vault.db  # SQLite database
├── frontend/             # Vue 3 SPA (Vite + Element Plus)
│   ├── src/
│   │   ├── views/       # Page components (6 views)
│   │   └── router/      # Vue Router config
│   └── vite.config.js   # Dev server with API proxy
└── CLAUDE.md
```

## Commands

### Backend (Go)
```bash
cd backend
go build -o server .    # Build
./server                # Run server on :8080
```

### Frontend (Vue)
```bash
cd frontend
npm install             # Install dependencies
npm run dev             # Dev server on :3000 (proxies /api to :8080)
npm run build           # Production build to dist/
npm run preview         # Preview production build
```

### Development Workflow
1. Start backend: `cd backend && ./server`
2. Start frontend: `cd frontend && npm run dev`
3. Frontend dev server proxies API requests to backend

## API Design

Base URL: `http://localhost:8080/api`

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /prompts | List all prompts |
| POST | /prompts | Create prompt |
| GET | /prompts/:id | Get prompt |
| PUT | /prompts/:id | Update prompt (auto-creates version if content changes) |
| DELETE | /prompts/:id | Delete prompt and versions |
| GET | /prompts/:id/versions | List versions |
| POST | /prompts/:id/versions | Create version manually |
| GET | /versions/:id | Get version |
| POST | /prompts/:id/test | Test prompt with AI |
| POST | /prompts/:id/optimize | AI optimize prompt |
| GET | /prompts/:id/tests | List test history |

### Response Format
```json
{
  "success": true,
  "data": { ... }
}
```

### Environment Variables
- `OPENAI_API_KEY` - Optional. When set, uses real OpenAI API for test/optimize. Otherwise uses mock responses.

## Frontend Routes

| Path | View | Description |
|------|------|-------------|
| / | → /prompts | Redirect |
| /prompts | PromptList | Prompt library with grid view |
| /prompts/:id | PromptEditor | Edit prompt content/metadata |
| /prompts/:id/versions | VersionHistory | Version timeline |
| /prompts/:id/compare | VersionCompare | Side-by-side diff |
| /prompts/:id/test | PromptTester | Chat preview |
| /prompts/:id/optimize | OptimizePrompt | AI optimization |

## Design System

CSS variables in `App.vue`:
- Colors: `--color-primary: #2563EB`, semantic colors for success/warning/danger
- Typography: Plus Jakarta Sans font
- Spacing: 4/8pt scale (`--spacing-1` through `--spacing-12`)
- Border radius: `--radius-sm/md/lg/xl`
- Transitions: `--transition-fast/normal/slow` (150/200/300ms)

## Key Implementation Notes

- **Auto-versioning**: PUT to `/api/prompts/:id` with changed `content` automatically creates a new version
- **Tags storage**: Tags stored as JSON string in SQLite, parsed on API response
- **Mock AI**: When `OPENAI_API_KEY` not set, test/optimize endpoints return predefined mock responses
- **Diff library**: Frontend uses `diff` npm package for version comparison
