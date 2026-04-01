# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**PromptVault** - A prompt management system for organizing, versioning, testing, and optimizing AI prompts. Built with Go/Gin backend and Vue 3/Element Plus frontend. Also manages reusable Skills (slash commands) and Agent personas.

## Architecture

```
vibecoder/
├── backend/                    # Go API server
│   ├── main.go               # Entry point, router, middleware, DB migration
│   ├── .air.toml             # Hot reload config (air)
│   ├── handlers/             # HTTP handlers (input validation, response formatting)
│   ├── service/             # Business logic (transactions, cloning, validation)
│   ├── models/              # GORM models and DTOs
│   ├── middleware/          # Gin middleware (trace, request logging, CORS, rate limiting)
│   ├── utils/               # Utilities (structured JSON logger with levels and traceId)
│   ├── docs/                # API documentation
│   ├── docs/                # API documentation
│   └── */*_test.go         # Unit tests (~85% coverage)
├── frontend/                  # Vue 3 SPA
│   ├── src/
│   │   ├── views/           # Page components (20+ views)
│   │   ├── components/      # Reusable UI components
│   │   ├── composables/     # Vue composition functions (API calls, state)
│   │   └── router/          # Vue Router config
│   └── vite.config.js       # Dev server with /api proxy to :8080
└── CLAUDE.md
```

### Backend Layer Separation

- **handlers/** - HTTP request/response handling only; no business logic
- **service/** - Business logic, transactions, data operations; injected with DB
- **models/** - GORM models, DTOs, response structures

## Commands

### Backend (Go)
```bash
cd backend
go build -o server .    # Build
./server                 # Run server on :8080
go test ./...           # Run all tests
go test ./handlers/... -v -run TestName  # Run specific test
go test -cover         # Check coverage
go vet                 # Lint
```

#### Hot Reload (Recommended for Development)
```bash
cd backend
~/go/bin/air            # Start with hot reload (auto-restart on code changes)
# Or if air is in PATH:
air
```

**Note**: `air` is installed via `go install github.com/cosmtrek/air@v1.27.8`. Configuration is in `backend/.air.toml`. It watches all `.go` files and rebuilds automatically when changes are detected.

#### Structured Logging
All logs are JSON-formatted with `trace_id` for request correlation.

- **Log levels**: DEBUG / INFO / WARN / ERROR / FATAL (set via `LOG_LEVEL` env var, default: INFO)
- **Trace ID**: Each request gets a UUID (`X-Trace-ID` header). All logs within a request share this ID.
- **Request logging**: Every HTTP request logs start/completion with method, path, status, latency, client IP
- **Panic recovery**: Stack traces are captured and logged on recovery

Usage:
```bash
LOG_LEVEL=DEBUG ./server  # Enable debug logging
```

### Frontend (Vue)
```bash
cd frontend
npm install             # Install dependencies
npm run dev             # Dev server on :3000
npm run build           # Production build
npm run preview         # Preview production build
```

### Development Workflow
1. Start backend: `cd backend && ~/go/bin/air` (recommended, with hot reload) or `./server`
2. Start frontend: `cd frontend && npm run dev`
3. Frontend dev server proxies API requests to backend

## API Design

Base URL: `http://localhost:8080/api`

### Response Format
```json
{
  "success": true,
  "data": { ... }
}
```

### Core Endpoints

**Prompts** - CRUD, versioning, testing, optimization
**Skills** - CRUD, cloning, import/export
**Agents** - CRUD, cloning, import/export
**Translation** - Entity translation and free text translation
**Settings** - Encrypted key-value storage (AES-256-GCM)
**Activity Log** - Operation audit trail
**Stats** - Dashboard counts

### Environment Variables
- `OPENAI_API_KEY` - Optional. When set, uses real OpenAI API for test/optimize
- `TRANSLATE_PROVIDER` / `TRANSLATE_MODEL` - Translation provider config
- `ENCRYPTION_KEY` - 32-byte key for settings encryption
- `.env` file in backend/ for local env vars (gitignored)

## Frontend Routes

| Path | View | Description |
|------|------|-------------|
| / | Dashboard | Overview with stats |
| /prompts | PromptList | Prompt library |
| /prompts/:id | PromptEditor | Edit prompt |
| /prompts/:id/versions | VersionHistory | Version timeline |
| /prompts/:id/compare | VersionCompare | Side-by-side diff |
| /prompts/:id/test | PromptTester | Chat preview with AI |
| /prompts/:id/optimize | OptimizePrompt | AI optimization |
| /skills | SkillList | Skill library |
| /skills/:id | SkillEditor | Edit skill |
| /agents | AgentList | Agent personas |
| /agents/:id | AgentEditor | Edit agent |
| /templates | TemplateMarketplace | Template marketplace (v2) |
| /templates/:id | TemplateDetail | Template details (v2) |
| /teams | TeamList | Team list (v2) |
| /teams/:id/members | TeamMemberList | Member management (v2) |
| /teams/:id/settings | TeamSettings | Team settings (v2) |
| /ab-tests | ABTestList | A/B test history (v2) |
| /ab-tests/:id | ABTestDetail | A/B test results (v2) |
| /api-docs | ApiDocs | API documentation |
| /activity | ActivityLog | Operation logs |
| /settings | Settings | App settings |

## Design System

CSS variables in `App.vue`:
- `--color-primary: #2563EB` - Primary blue
- Semantic colors: success/warning/danger/info
- Typography: Plus Jakarta Sans
- Spacing: 4pt grid scale
- Border radius: sm/md/lg/xl
- Transitions: fast (150ms) / normal (200ms) / slow (300ms)

## Key Implementation Notes

- **Auto-versioning**: PUT to `/api/prompts/:id` with changed `content` auto-creates a version
- **Tags storage**: Tags as JSON string in SQLite, parsed on API response
- **Mock AI**: Without `OPENAI_API_KEY`, test/optimize return mock responses
- **AI Providers**: OpenAI, Claude, Gemini, MiniMax unified interface in `handlers/ai_provider.go`
- **Encryption**: Settings use AES-256-GCM encryption; secrets never exposed to frontend
- **Translation**: Entities have `content` and `content_cn` fields; translation via MiniMax API
- **Variable Preview**: Frontend parses `{{variable}}` syntax for real-time preview in VariablePreviewer component
