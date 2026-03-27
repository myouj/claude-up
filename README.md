# PromptVault

A prompt management system for organizing, versioning, testing, and optimizing AI prompts. Built with Go/Gin backend and Vue 3/Element Plus frontend.

## Features

- **Prompt Library** - Create, edit, and organize AI prompts with categories and tags
- **Version Control** - Track changes with automatic versioning, compare versions side-by-side
- **AI Testing** - Test prompts in a chat preview with mock or real AI responses
- **AI Optimization** - Get AI-powered suggestions to improve your prompts
- **Translation** - Translate prompts between English and Chinese
- **Skills Management** - Reusable slash commands for common tasks
- **Agent Personas** - Define and manage different AI agent roles

## Tech Stack

- **Backend**: Go, Gin, GORM, SQLite
- **Frontend**: Vue 3, Vite, Element Plus, CodeMirror

## Quick Start

### Backend

```bash
cd backend
go build -o server .
./server
```

Server runs on `http://localhost:8080`

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend runs on `http://localhost:3000` and proxies API requests to the backend.

## API

Base URL: `http://localhost:8080/api`

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /prompts | List all prompts |
| POST | /prompts | Create prompt |
| GET | /prompts/:id | Get prompt |
| PUT | /prompts/:id | Update prompt |
| DELETE | /prompts/:id | Delete prompt |
| POST | /prompts/:id/test | Test prompt |
| POST | /prompts/:id/optimize | Optimize prompt |
| GET | /skills | List skills |
| GET | /agents | List agents |
| GET | /stats | Get counts |

## Environment Variables

Create `backend/.env` to configure:

```
OPENAI_API_KEY=your-api-key  # Optional - enables real AI responses
```

Without `OPENAI_API_KEY`, the system uses mock responses for testing.

## Project Structure

```
vibecoder/
├── backend/
│   ├── main.go         # Entry point, router, DB setup
│   ├── handlers/       # HTTP request handlers
│   └── models/         # Data models
├── frontend/
│   └── src/
│       ├── views/      # Vue components
│       └── router/     # Vue Router config
└── CLAUDE.md           # Developer documentation
```
