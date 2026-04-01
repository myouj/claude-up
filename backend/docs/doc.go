// Package docs contains PromptVault architecture specifications.
//
// # Layered Architecture (Handler / Service / Repository)
//
//   - handlers/    HTTP layer: request parsing, validation, response formatting
//   - service/     Business logic layer: data operations, validation (future)
//   - models/      Data models: GORM models, request/response DTOs
//   - middleware/  Gin middleware: pagination, CORS, rate limiting
//
// # API Response Format
//
// All endpoints use the following JSON format:
//
//	Success:  {"success": true, "data": {...}}
//	Paginated: {"success": true, "data": [...], "meta": {"total": N, "page": 1, "limit": 20, "total_pages": 5}}
//	Error:     {"success": false, "error": "description"}
//
// # Error Handling
//
//   - 4xx errors: return specific message to help client fix request
//   - 5xx errors: log details internally, return "Internal server error" externally
//   - Never expose internal error details (DB errors, stack traces) to clients
//   - All goroutines spawned via go func() must have defer recover()
//
// # AI Provider Architecture
//
// Located in handlers/ai_provider.go. Unified AIProvider interface supporting
// OpenAI, Claude (Anthropic), Gemini (Google), and MiniMax. When no API key is
// configured, falls back to mock responses for development.
//
// # API Endpoints
//
// Base URL: http://localhost:8080/api
//
// ## Prompts
//
//   - GET    /prompts              List prompts (params: page, limit, search, category, tag, favorite)
//   - POST   /prompts              Create prompt
//   - GET    /prompts/:id          Get single prompt
//   - PUT    /prompts/:id          Update prompt (auto-creates version on content change)
//   - DELETE /prompts/:id          Delete prompt + versions + tests
//   - POST   /prompts/:id/favorite Toggle favorite
//   - GET    /prompts/categories   List unique categories
//   - POST   /prompts/:id/clone   Clone prompt
//   - GET    /prompts/export      Export all prompts
//   - POST   /prompts/import       Import prompts (returns imported/failed counts)
//
// ## Versions
//
//   - GET    /prompts/:id/versions List versions for prompt
//   - POST   /prompts/:id/versions Create version manually
//   - GET    /versions/:id         Get version by ID
//
// ## Skills
//
//   - GET    /skills               List skills (params: page, limit, category, source)
//   - POST   /skills               Create skill
//   - GET    /skills/:id           Get skill
//   - PUT    /skills/:id           Update skill
//   - DELETE /skills/:id            Delete skill
//   - GET    /skills/categories    List categories
//   - POST   /skills/:id/clone     Clone skill (source → "custom")
//   - GET    /skills/export        Export all skills
//   - POST   /skills/import       Import skills
//
// ## Agents
//
//   - GET    /agents               List agents (params: page, limit, category, source)
//   - POST   /agents               Create agent
//   - GET    /agents/:id           Get agent
//   - PUT    /agents/:id           Update agent
//   - DELETE /agents/:id           Delete agent
//   - GET    /agents/categories    List categories
//   - POST   /agents/:id/clone    Clone agent (source → "custom")
//   - GET    /agents/export        Export all agents
//   - POST   /agents/import       Import agents
//
// ## Test & Optimize
//
//   - POST   /prompts/:id/test       Test prompt with AI (params: content, model, provider, messages)
//   - POST   /prompts/:id/optimize   Optimize prompt (params: mode: improve|structure|suggest)
//   - GET    /prompts/:id/tests       List test history (params: page, limit)
//   - GET    /prompts/:id/test-compare Compare two tests (params: v1, v2)
//   - GET    /prompts/:id/analytics   Test analytics
//   - GET    /models                 List available models with pricing
//
// ## Translation
//
//   - POST   /translate              Translate free text (params: text, source_lang, target_lang)
//   - POST   /translate/:type/:id     Translate entity :type (prompt|skill|agent), updates content_cn
//
// ## Activity Log
//
//   - GET    /activity-logs          List activity logs (params: page, limit, entity_type, entity_id, action)
//
// ## Settings
//
//   - GET    /settings               List all settings (secrets masked)
//   - GET    /settings/:key           Get setting value (secrets decrypted)
//   - PUT    /settings/:key          Set setting (secrets encrypted with AES-256-GCM)
//   - DELETE /settings/:key          Delete setting
//
// ## Stats
//
//   - GET    /stats                  Get counts: prompts, skills, agents
//   - GET    /export                 Full export: all prompts, skills, agents
//
// # Data Models
//
// ## Prompt
//
//	{id, title, content, content_cn, description, category, tags (JSON),
//	 variables (JSON), is_favorite, is_pinned, created_at, updated_at}
//
// ## PromptVersion
//
//	{id, prompt_id, version (int), content, comment, created_at}
//
// ## Skill
//
//	{id, name, description, content, content_cn, category, source (builtin|custom)}
//
// ## Agent
//
//	{id, name, role, content, content_cn, capabilities, category, source (builtin|custom)}
//
// ## TestRecord
//
//	{id, prompt_id, version_id, model, provider, prompt_text, response,
//	 tokens_used, latency_ms, created_at}
//
// ## ActivityLog
//
//	{id, entity_type, entity_id, action, user_id, details (JSON), created_at}
//
// ## Setting
//
//	{id, key (unique), value, is_secret, updated_at}
package docs
