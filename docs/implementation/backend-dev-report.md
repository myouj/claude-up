# PromptVault 后端实现方案

> **任务**: Task #4 - 后端实现方案：数据库设计、API 设计与 Worker 实现
> **负责人**: backend-dev
> **日期**: 2026-04-01
> **基于**: `docs/功能规划_v2.md` + 现有代码分析

---

## 一、现有代码分析

### 1.1 现有架构

- **框架**: Go + Gin + GORM + SQLite
- **现有模型**: `Prompt`, `PromptVersion`, `TestRecord`, `Skill`, `Agent`, `Translation`, `ActivityLog`, `Setting`
- **AI Provider**: OpenAI / Claude / Gemini / MiniMax，实现统一 `AIProvider` 接口（`Name()`, `Call()`）
- **中间件**: trace_id、请求日志、限流（IP 级）、CORS、panic recovery
- **无 Worker 系统**: 当前所有 AI 调用均为同步阻塞

### 1.2 需要新增的模块

| 模块 | 说明 |
|------|------|
| `Task` | 异步任务队列，SQLite 持久化 |
| `AICallLog` | AI 调用链路记录 |
| 扩展 `ActivityLog` | 增加 action_type、detail、trace_id |
| `Worker` | 内嵌 goroutine 池 + SQLite 轮询 |
| `BatchService` | 批量测试业务逻辑 |
| `ABTestService` | A/B 测试 + SPRT 统计引擎 |
| `EvalSet` | 评测集管理 |
| `CacheLayer` | AI 响应缓存（内存 + SQLite） |
| `TaskHandler` | 任务相关 API Handler |
| `EvalSetHandler` | 评测集 CRUD API Handler |

---

## 二、数据库设计

### 2.1 Task 表（异步任务队列）

```sql
CREATE TABLE tasks (
    id           TEXT PRIMARY KEY,                    -- UUID v4
    type         TEXT NOT NULL,                       -- batch_test | ab_test | eval_gen | optimize | regression | multi_turn
    entity_type  TEXT NOT NULL,                       -- prompt | skill | agent
    entity_id    INTEGER NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending',     -- pending | running | done | failed | cancelled
    priority     INTEGER NOT NULL DEFAULT 0,          -- 数值越大优先级越高
    payload      TEXT NOT NULL,                      -- JSON: {cases?: [], config?: {}, ...}
    result       TEXT,                                -- JSON: 完成后写入
    error        TEXT,                                -- 失败原因
    progress     INTEGER NOT NULL DEFAULT 0,          -- 0-100
    total        INTEGER NOT NULL DEFAULT 0,          -- 总任务数（如：测试用例总数）
    current      INTEGER NOT NULL DEFAULT 0,         -- 当前完成数
    worker_id    TEXT,                                -- 处理此任务的 Worker ID
    created_at   DATETIME NOT NULL,
    updated_at   DATETIME NOT NULL,
    started_at   DATETIME,                            -- 开始执行时间
    completed_at DATETIME                             -- 完成时间
);

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_entity ON tasks(entity_type, entity_id);
CREATE INDEX idx_tasks_priority ON tasks(priority DESC, created_at ASC);
```

### 2.2 AICallLog 表（AI 调用链路记录）

```sql
CREATE TABLE ai_call_logs (
    id             TEXT PRIMARY KEY,                  -- UUID v4
    task_id        TEXT,                              -- 关联任务（可为 NULL）
    trace_id       TEXT NOT NULL,                     -- 复用现有 trace_id 机制
    provider       TEXT NOT NULL,                     -- openai | claude | gemini | minimax
    model          TEXT NOT NULL,
    prompt_tokens  INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens   INTEGER NOT NULL DEFAULT 0,
    latency_ms     INTEGER NOT NULL DEFAULT 0,         -- 端到端延迟（毫秒）
    status_code    INTEGER NOT NULL DEFAULT 200,       -- HTTP 状态码
    error          TEXT,                               -- 错误信息
    cost_usd       REAL NOT NULL DEFAULT 0,            -- 美元成本（按 availableModels 中的单价计算）
    cached         BOOLEAN NOT NULL DEFAULT 0,         -- 是否命中缓存
    created_at     DATETIME NOT NULL
);

CREATE INDEX idx_ai_call_task ON ai_call_logs(task_id);
CREATE INDEX idx_ai_call_trace ON ai_call_logs(trace_id);
CREATE INDEX idx_ai_call_created ON ai_call_logs(created_at);
CREATE INDEX idx_ai_call_provider ON ai_call_logs(provider, created_at);
```

### 2.3 EvalSet 表（评测集）

```sql
CREATE TABLE eval_sets (
    id          TEXT PRIMARY KEY,                      -- UUID v4
    name        TEXT NOT NULL,
    description TEXT,
    entity_type TEXT NOT NULL,                         -- prompt | skill | agent
    entity_id   INTEGER NOT NULL,
    cases       TEXT NOT NULL,                         -- JSON: [{input: {}, expected?: "", tags?: []}, ...]
    weights     TEXT,                                   -- JSON: {clarity: 0.3, completeness: 0.3, example: 0.25, role: 0.15}
    version_id  INTEGER,                                -- 关联版本
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL
);

CREATE INDEX idx_eval_set_entity ON eval_sets(entity_type, entity_id);
```

### 2.4 ABTest 表（A/B 测试）

```sql
CREATE TABLE ab_tests (
    id                  TEXT PRIMARY KEY,              -- UUID v4
    name                TEXT NOT NULL,
    entity_type         TEXT NOT NULL,
    entity_id           INTEGER NOT NULL,
    version_a_id        INTEGER,                        -- 版本 A（通常为最新）
    version_b_id        INTEGER,                        -- 版本 B（历史版本或对照版本）
    model               TEXT NOT NULL,
    provider            TEXT NOT NULL,
    config              TEXT NOT NULL,                   -- JSON: {alpha, beta, mde, max_samples, min_samples}
    status              TEXT NOT NULL DEFAULT 'running', -- pending | running | done | stopped
    significance_result TEXT,                           -- done 时写入: {winner, confidence, p_value, sample_size_a, sample_size_b}
    created_at          DATETIME NOT NULL,
    updated_at          DATETIME NOT NULL,
    completed_at        DATETIME
);

CREATE INDEX idx_ab_test_entity ON ab_tests(entity_type, entity_id);

CREATE TABLE ab_test_results (
    id              TEXT PRIMARY KEY,
    ab_test_id      TEXT NOT NULL,
    run_index       INTEGER NOT NULL,                   -- 第几次运行（1-based）
    version         TEXT NOT NULL,                     -- 'a' | 'b'
    score           REAL NOT NULL,                     -- 归一化得分 0-1
    raw_response    TEXT,                               -- 原始 AI 输出
    latency_ms      INTEGER,
    tokens_used     INTEGER,
    created_at      DATETIME NOT NULL
);

CREATE INDEX idx_ab_results_test ON ab_test_results(ab_test_id);
```

### 2.5 扩展 ActivityLog 表

```sql
-- 新增字段（向后兼容，原有字段保留）
ALTER TABLE activity_logs ADD COLUMN action_type TEXT;       -- ai_call | batch_test | ab_test | eval_gen | ...
ALTER TABLE activity_logs ADD COLUMN detail TEXT;            -- JSON: {before: {}, after: {}, extra: {}}
-- 注: existing `details` column 保留复用，仅改 JSON 结构
-- trace_id 复用中间件的 trace ID，无需新列
```

### 2.6 ResponseCache 表（AI 响应缓存）

```sql
CREATE TABLE response_cache (
    hash        TEXT PRIMARY KEY,                  -- SHA256(request_payload)
    provider    TEXT NOT NULL,
    model       TEXT NOT NULL,
    request     TEXT NOT NULL,                    -- JSON 规范化后的请求体
    response    TEXT NOT NULL,
    tokens_used INTEGER NOT NULL,
    ttl_seconds INTEGER NOT NULL DEFAULT 3600,     -- 缓存 TTL，默认 1 小时
    created_at  DATETIME NOT NULL,
    accessed_at DATETIME NOT NULL
);

CREATE INDEX idx_cache_expire ON response_cache(ttl_seconds, created_at);
```

### 2.7 GORM AutoMigrate 更新

```go
// main.go 中需要添加
db.AutoMigrate(
    &models.Task{},
    &models.AICallLog{},
    &models.EvalSet{},
    &models.ABTest{},
    &models.ABTestResult{},
    &models.ResponseCache{},
)

// ActivityLog 扩展字段通过 ALTER TABLE 或手动 migration 添加
```

---

## 三、API 设计

### 3.1 任务管理 API

| 方法 | 路径 | 说明 | 请求体 | 响应 |
|------|------|------|--------|------|
| `POST` | `/api/tasks` | 创建任务（批量测试、A/B 测试、评测集生成等） | `TaskCreateRequest` | `Task` |
| `GET` | `/api/tasks` | 列出任务（支持分页、状态筛选、实体筛选） | - | `Task[]` |
| `GET` | `/api/tasks/:id` | 获取任务详情 | - | `Task` |
| `DELETE` | `/api/tasks/:id` | 取消任务 | - | `{success: true}` |
| `GET` | `/api/tasks/:id/progress` | 获取任务进度（SSE 流） | - | `SSE stream` |

**TaskCreateRequest**:
```json
{
  "type": "batch_test | ab_test | eval_gen | regression | multi_turn",
  "entity_type": "prompt | skill | agent",
  "entity_id": 123,
  "priority": 0,
  "payload": {
    "cases": [{ "variables": {"name": "Alice"}, "messages": [] }],
    "model": "gpt-4o",
    "provider": "openai",
    "eval_set_id": "uuid"       // 可选，用于评分
  }
}
```

**Task 响应结构**:
```json
{
  "id": "uuid",
  "type": "batch_test",
  "entity_type": "prompt",
  "entity_id": 123,
  "status": "running",
  "progress": 45,
  "total": 20,
  "current": 9,
  "result": null,
  "error": null,
  "created_at": "2026-04-01T10:00:00Z",
  "updated_at": "2026-04-01T10:05:00Z"
}
```

### 3.2 批量测试 API（快捷入口）

| 方法 | 路径 | 说明 | 请求体 | 响应 |
|------|------|------|--------|------|
| `POST` | `/api/prompts/:id/batch-test` | 创建批量测试任务 | `BatchTestRequest` | `Task` |
| `GET` | `/api/prompts/:id/batch-test` | 获取批量测试结果列表 | - | `BatchTestResult[]` |

**BatchTestRequest**:
```json
{
  "cases": [
    { "name": "Case 1", "variables": {"topic": "Python"}, "messages": [] },
    { "name": "Case 2", "variables": {"topic": "Go"}, "messages": [] }
  ],
  "model": "gpt-4o",
  "provider": "openai",
  "eval_set_id": "uuid"
}
```

**BatchTestResult** (任务完成后的 result 字段):
```json
{
  "task_id": "uuid",
  "summary": {
    "total": 20,
    "success": 18,
    "failed": 2,
    "avg_score": 0.85,
    "avg_latency_ms": 1250,
    "total_tokens": 45000,
    "total_cost_usd": 0.23
  },
  "cases": [
    {
      "index": 0,
      "name": "Case 1",
      "status": "success",
      "response": "...",
      "score": 0.92,
      "latency_ms": 1100,
      "tokens_used": 2100,
      "cost_usd": 0.011,
      "error": null
    }
  ]
}
```

### 3.3 A/B 测试 API

| 方法 | 路径 | 说明 | 请求体 | 响应 |
|------|------|------|--------|------|
| `POST` | `/api/prompts/:id/ab-tests` | 创建 A/B 测试 | `ABTestCreateRequest` | `ABTest` |
| `GET` | `/api/prompts/:id/ab-tests` | 获取 A/B 测试列表 | - | `ABTest[]` |
| `GET` | `/api/ab-tests/:id` | 获取 A/B 测试详情（含中间结果） | - | `ABTest` |
| `POST` | `/api/ab-tests/:id/stop` | 手动停止 A/B 测试 | - | `ABTest` |
| `GET` | `/api/ab-tests/:id/sse` | SSE 流推送实时结果 | - | `SSE stream` |

**ABTestCreateRequest**:
```json
{
  "name": "Prompt v3 vs v4 效果对比",
  "version_a_id": 10,
  "version_b_id": 9,
  "model": "gpt-4o",
  "provider": "openai",
  "cases": [{ "variables": {"topic": "Python"} }],
  "config": {
    "alpha": 0.05,
    "beta": 0.20,
    "mde": 0.10,
    "min_samples": 15,
    "max_samples": 50
  }
}
```

### 3.4 评测集 API

| 方法 | 路径 | 说明 | 请求体 | 响应 |
|------|------|------|--------|------|
| `POST` | `/api/eval-sets` | 创建评测集 | `EvalSetCreateRequest` | `EvalSet` |
| `GET` | `/api/eval-sets` | 列出评测集 | - | `EvalSet[]` |
| `GET` | `/api/eval-sets/:id` | 获取评测集详情 | - | `EvalSet` |
| `PUT` | `/api/eval-sets/:id` | 更新评测集 | `EvalSetCreateRequest` | `EvalSet` |
| `DELETE` | `/api/eval-sets/:id` | 删除评测集 | - | `{success: true}` |
| `POST` | `/api/eval-sets/:id/generate` | AI 生成评测用例 | `EvalGenerateRequest` | `Task` |

**EvalSetCreateRequest**:
```json
{
  "name": "通用助手评测集",
  "description": "用于测试通用问答类 prompt",
  "entity_type": "prompt",
  "entity_id": 123,
  "cases": [
    { "input": {"variables": {"query": "How do I write a Go goroutine?"}, "messages": []}, "expected": "..." }
  ],
  "weights": {
    "clarity": 0.30,
    "completeness": 0.30,
    "example": 0.25,
    "role": 0.15
  }
}
```

### 3.5 回归检测 API

| 方法 | 路径 | 说明 | 请求体 | 响应 |
|------|------|------|--------|------|
| `POST` | `/api/prompts/:id/regression` | 触发回归检测 | `RegressionRequest` | `Task` |
| `GET` | `/api/prompts/:id/regression/:taskId` | 获取回归检测结果 | - | `RegressionResult` |

**RegressionRequest**:
```json
{
  "type": "light | full | model_upgrade",
  "eval_set_id": "uuid",
  "baseline_version_id": 9
}
```

### 3.6 AI 调用日志 API

| 方法 | 路径 | 说明 | 查询参数 | 响应 |
|------|------|------|---------|------|
| `GET` | `/api/ai-call-logs` | 查询 AI 调用记录 | `task_id`, `provider`, `from`, `to`, `page`, `limit` | `AICallLog[]` |
| `GET` | `/api/ai-call-logs/stats` | 调用统计 | `from`, `to`, `group_by` | `AICallStats` |

### 3.7 成本分析 API

| 方法 | 路径 | 说明 | 响应 |
|------|------|------|------|
| `GET` | `/api/cost-stats` | 成本汇总统计 | `{total_cost, by_provider, by_model, daily_breakdown}` |

### 3.8 变量预览 API

| 方法 | 路径 | 说明 | 请求体 | 响应 |
|------|------|------|--------|------|
| `POST` | `/api/prompts/:id/variables/preview` | 变量预览（实时替换） | `VariablesPreviewRequest` | `VariablesPreviewResponse` |

---

## 四、Worker 实现

### 4.1 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                         main.go                              │
│  ┌──────────┐   ┌──────────────────┐   ┌──────────────────┐ │
│  │  Router  │   │  TaskScheduler   │   │  WorkerPool     │ │
│  │  (Gin)   │──▶│  (create tasks)  │──▶│  (execute tasks) │ │
│  └──────────┘   └──────────────────┘   └────────┬─────────┘ │
│       │                                          │           │
│  ┌────▼────┐                              ┌─────▼─────┐   │
│  │ SSE Hub  │◀─────────────────────────────│ SSE Chan  │   │
│  │(notify   │                              │(per task) │   │
│  │ clients) │                              └───────────┘   │
│  └──────────┘                                                │
└─────────────────────────────────────────────────────────────┘
                              │
                    ┌─────────▼─────────┐
                    │   SQLite DB        │
                    │  tasks table       │
                    │  ai_call_logs table│
                    └───────────────────┘
```

### 4.2 Worker 池配置

```go
// backend/worker/worker.go

type WorkerConfig struct {
    PoolSize      int           // goroutine 池大小，默认 5
    PollInterval  time.Duration // 轮询间隔，默认 3 秒
    MaxRetries    int           // 失败重试次数，默认 3
    RetryBackoff  time.Duration // 重试退避时间，默认 10 秒
}

// 全局配置（可通过环境变量覆盖）
const (
    DefaultPoolSize     = 5
    DefaultPollInterval = 3 * time.Second
    DefaultMaxRetries   = 3
)
```

**Goroutine 池大小选择依据**:
- MVP 阶段并发量可控（5 个 Worker 可同时处理 5 个任务）
- 每个 Worker 内部对 AI Provider 的调用仍有并发（batch_test 场景）
- 可通过 `WORKER_POOL_SIZE` 环境变量调整

### 4.3 Worker 伪代码

```go
package worker

import (
    "context"
    "database/sql"
    "encoding/json"
    "sync"
    "time"

    "prompt-vault/handlers"
    "prompt-vault/models"
)

type WorkerPool struct {
    db       *sql.DB
    config   WorkerConfig
    taskChan chan TaskJob        // 新任务通知 channel
    wg       sync.WaitGroup
    cancel   context.CancelFunc
    sseHub   *SSEHub             // SSE 通知中心
}

type TaskJob struct {
    TaskID  string
    Type    string
    Payload json.RawMessage
}

func NewWorkerPool(db *sql.DB, config WorkerConfig, sseHub *SSEHub) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    wp := &WorkerPool{
        db:       db,
        config:   config,
        taskChan: make(chan TaskJob, config.PoolSize*2),
        cancel:   cancel,
        sseHub:   sseHub,
    }
    return wp
}

// Start 启动 Worker 池和轮询器
func (wp *WorkerPool) Start() {
    // 启动 N 个 Worker goroutine
    for i := 0; i < wp.config.PoolSize; i++ {
        wp.wg.Add(1)
        go wp.worker(i)
    }

    // 启动轮询器：每 PollInterval 查询一次 pending 任务
    wp.wg.Add(1)
    go wp.pollLoop()
}

// pollLoop 轮询 pending 任务
func (wp *WorkerPool) pollLoop() {
    defer wp.wg.Done()
    ticker := time.NewTicker(wp.config.PollInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            tasks := wp.fetchPendingTasks()
            for _, task := range tasks {
                select {
                case wp.taskChan <- TaskJob{Type: task.Type, Payload: task.Payload, TaskID: task.ID}:
                    wp.updateTaskStatus(task.ID, "running")
                default:
                    // channel 满，等待下一轮
                }
            }
        }
    }
}

// fetchPendingTasks 查询待处理任务（按 priority 降序、created_at 升序）
func (wp *WorkerPool) fetchPendingTasks() []PendingTask {
    // SQL: SELECT * FROM tasks
    //       WHERE status = 'pending'
    //       ORDER BY priority DESC, created_at ASC
    //       LIMIT PoolSize
    // 返回包含 ID, Type, Payload 的结构体
}

// worker 处理单个任务
func (wp *WorkerPool) worker(id int) {
    defer wp.wg.Done()
    for job := range wp.taskChan {
        wp.processTask(job)
    }
}

func (wp *WorkerPool) processTask(job TaskJob) {
    // 更新状态为 running
    wp.updateTaskStatus(job.TaskID, "running")

    var err error
    var result interface{}

    switch job.Type {
    case "batch_test":
        result, err = wp.runBatchTest(job)
    case "ab_test":
        result, err = wp.runABTest(job)
    case "eval_gen":
        result, err = wp.runEvalGen(job)
    case "regression":
        result, err = wp.runRegression(job)
    case "multi_turn":
        result, err = wp.runMultiTurn(job)
    default:
        err = fmt.Errorf("unknown task type: %s", job.Type)
    }

    // 写入结果或错误
    if err != nil {
        wp.updateTaskError(job.TaskID, err.Error())
    } else {
        wp.updateTaskDone(job.TaskID, result)
    }

    // SSE 通知
    wp.sseHub.Broadcast(job.TaskID, map[string]interface{}{
        "type": "task_complete",
        "task_id": job.TaskID,
        "status": map[bool]string{true: "done", false: "failed"}[err == nil],
    })
}

// runBatchTest 执行批量测试
func (wp *WorkerPool) runBatchTest(job TaskJob) (*BatchTestResult, error) {
    var req BatchTestRequest
    json.Unmarshal(job.Payload, &req)

    total := len(req.Cases)
    result := &BatchTestResult{Total: total, Cases: make([]CaseResult, total)}

    for i, c := range req.Cases {
        // SSE 进度通知
        wp.sseHub.Broadcast(job.TaskID, map[string]interface{}{
            "type": "progress",
            "current": i + 1,
            "total":   total,
            "progress": (i + 1) * 100 / total,
        })
        wp.updateTaskProgress(job.TaskID, i+1, total)

        // 调用 AI Provider（带缓存）
        resp, tokens, latency, err := wp.callAIWithCache(req.Provider, req.Model, c)

        // 记录 AI 调用
        wp.logAICall(job.TaskID, req.Provider, req.Model, tokens, latency, err)

        if err != nil {
            result.Cases[i] = CaseResult{Index: i, Name: c.Name, Status: "failed", Error: err.Error()}
            result.Failed++
        } else {
            score := wp.scoreResponse(resp, req.EvalSetID)
            result.Cases[i] = CaseResult{
                Index: i, Name: c.Name, Status: "success",
                Response: resp, Score: score,
                LatencyMs: latency, TokensUsed: tokens,
                CostUSD: calcCost(req.Provider, req.Model, tokens),
            }
            result.Success++
            result.TotalTokens += tokens
        }
    }

    result.TotalCostUSD = calcTotalCost(result.Cases)
    return result, nil
}

// runABTest 执行 A/B 测试（SPRT 序贯检验）
func (wp *WorkerPool) runABTest(job TaskJob) (*ABTestResult, error) {
    // 实现 Wald's SPRT
    // 1. 初始化：每版本至少 10 次
    // 2. 循环：每轮运行后检查 SPRT 是否达到显著性
    // 3. 停止条件：达到显著性 或 达到 max_samples
}

// callAIWithCache 带缓存的 AI 调用
func (wp *WorkerPool) callAIWithCache(provider, model string, caseInput CaseInput) (string, int, int64, error) {
    // 1. 计算请求 hash
    // 2. 查询 response_cache 表
    // 3. 命中则返回（更新 accessed_at）
    // 4. 未命中则调用 AI Provider，记录 ai_call_log
    // 5. 写入缓存
}

// SSE Hub
type SSEHub struct {
    mu      sync.RWMutex
    clients map[string]chan SSEEvent  // task_id -> client channel
}

func (h *SSEHub) Broadcast(taskID string, event SSEEvent) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    if ch, ok := h.clients[taskID]; ok {
        select {
        case ch <- event:
        default:
        }
    }
}
```

### 4.4 SSE 通知端点

```go
// GET /api/tasks/:id/sse
func (h *TaskHandler) TaskSSE(c *gin.Context) {
    taskID := c.Param("id")

    // 设置 SSE headers
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    // 订阅任务事件
    ch := h.sseHub.Subscribe(taskID)
    defer h.sseHub.Unsubscribe(taskID, ch)

    // 先发送当前状态
    task := h.getTask(taskID)
    c.SSEvent("status", task)

    // 持续推送直到任务完成或客户端断开
    for event := range ch {
        c.SSEvent(event.Type, event.Data)
        c.Writer.Flush()
        if event.Type == "task_complete" {
            break
        }
    }
}
```

### 4.5 轮询策略

| 场景 | 轮询间隔 | Worker 数 | 说明 |
|------|---------|-----------|------|
| 默认 | 3s | 5 | 平衡响应速度与 DB 压力 |
| 高负载 | 5s | 10 | 通过 WORKER_POOL_SIZE=10 |
| 开发调试 | 10s | 1 | WORKER_POOL_SIZE=1, POLL_INTERVAL=10s |
| 批量测试 | 按需 | 5 | 通过 priority 区分紧急任务 |

---

## 五、AI Provider 扩展

### 5.1 扩展 AIProvider 接口

```go
// 现有接口（保持向后兼容）
type AIProvider interface {
    Name() string
    Call(messages []map[string]string, model string) (response string, tokens int, err error)
}

// 新增批量调用接口（用于 batch_test 和 ab_test）
type BatchAIProvider interface {
    AIProvider
    // BatchCall 并发执行多个请求，返回结果数组
    BatchCall(requests []BatchRequest, model string) ([]BatchResponse, error)
    // GetCost 计算单次调用的成本（美元）
    GetCost(model string, tokens int) float64
}

type BatchRequest struct {
    ID       string
    Messages []map[string]string
}

type BatchResponse struct {
    ID              string
    Response        string
    Tokens          int
    LatencyMs       int64
    Err             error
}
```

### 5.2 批量调用实现（OpenAI 示例）

```go
// OpenAI Batch API（OpenAI 官方 Batch Endpoint）
func (p *OpenAIProvider) BatchCall(requests []BatchRequest, model string) ([]BatchResponse, error) {
    apiKey := os.Getenv("OPENAI_API_KEY")
    baseURL := p.baseURL
    if baseURL == "" {
        baseURL = "https://api.openai.com/v1/chat/completions"
    }

    // 构建 OpenAI Batch 请求体
    batchReqs := make([]map[string]interface{}, len(requests))
    for i, r := range requests {
        batchReqs[i] = map[string]interface{}{
            "custom_id": r.ID,
            "method":    "POST",
            "url":       "/v1/chat/completions",
            "body": map[string]interface{}{
                "model":    model,
                "messages": r.Messages,
            },
        }
    }

    // 提交 Batch Job
    submitBody := map[string]interface{}{"input_file_content": batchReqs}
    // ... 提交到 OpenAI Batch API (POST /v1/batches)
    // 轮询 batch 状态
    // 返回 BatchResponse 数组
}
```

### 5.3 A/B 测试扩展

```go
// ABTestRunner A/B 测试运行器
type ABTestRunner struct {
    provider AIProvider
    scorer   PromptScorer  // 4 维度评分器
}

func (r *ABTestRunner) RunSPRT(ctx context.Context, cfg SPRTConfig, cases []CaseInput) (*SPRTResult, error) {
    // SPRT (Wald's Sequential Probability Ratio Test)
    // theta0: 零假设（两版本等价）
    // theta1: 备择假设（效果差为 MDE）
    //
    // 算法：
    // 1. 初始运行每个版本 10 次
    // 2. 每轮后计算似然比
    //    S = sum(log(Lambda_i))  // 累积 log 似然比
    //    if S > log(A): reject H0 (version_a wins)
    //    if S < log(B): reject H0 (version_b wins)
    //    else: continue sampling
    // 3. 若 sample_size >= max_samples，强制停止
    // 4. 若任一版本 < min_samples，继续采样
    //
    // OpenAI Batch API 使用：
    // - 每次提交 A 侧 10 个 + B 侧 10 个请求
    // - 等待 batch 完成，汇总得分，计算 SPRT
    // - SSE 推送每轮结果
}

// SPRT 配置
type SPRTConfig struct {
    Alpha      float64 // 显著性水平，默认 0.05
    Beta       float64 // 统计功效，默认 0.80
    MDE        float64 // 最小可检测效应，默认 0.10
    MinSamples int     // 最小样本量，默认 15
    MaxSamples int     // 最大样本量，默认 50
}

// Wald's SPRT 边界
// A = (1 - beta) / alpha
// B = beta / (1 - alpha)
```

### 5.4 Prompt 评分卡（4 维度）

```go
// PromptScorer 4 维度评分器
type PromptScorer struct {
    provider AIProvider
    Weights  ScoreWeights
}

type ScoreWeights struct {
    Clarity      float64 // 30%
    Completeness float64 // 30%
    Example      float64 // 25%
    Role         float64 // 15%
}

func (s *PromptScorer) Score(prompt string, response string, evalCase CaseInput) (float64, ScoreBreakdown, error) {
    // 构造评分 Prompt（few-shot）
    scoringPrompt := s.buildScoringPrompt(prompt, response, evalCase)
    messages := []map[string]string{{"role": "user", "content": scoringPrompt}}

    raw, _, err := s.provider.Call(messages, "")
    if err != nil {
        return 0, ScoreBreakdown{}, err
    }

    // 解析 JSON 响应：{clarity: 0.9, completeness: 0.8, example: 0.7, role: 0.95}
    var breakdown ScoreBreakdown
    json.Unmarshal([]byte(raw), &breakdown)

    // 加权求和
    total := breakdown.Clarity*s.Weights.Clarity +
        breakdown.Completeness*s.Weights.Completeness +
        breakdown.Example*s.Weights.Example +
        breakdown.Role*s.Weights.Role

    return total, breakdown, nil
}

type ScoreBreakdown struct {
    Clarity      float64 `json:"clarity"`
    Completeness float64 `json:"completeness"`
    Example      float64 `json:"example"`
    Role         float64 `json:"role"`
}
```

---

## 六、文件结构规划

```
backend/
├── main.go                          # 新增: Worker 初始化、Task/AICallLog AutoMigrate
├── worker/
│   ├── worker.go                    # WorkerPool + pollLoop
│   ├── executor.go                  # 任务执行器（batch_test, ab_test, eval_gen...）
│   ├── sse.go                       # SSE Hub 实现
│   ├── cache.go                     # AI 响应缓存层
│   └── sprt.go                      # SPRT 统计显著性算法
├── models/
│   ├── task.go                      # Task 模型
│   ├── ai_call_log.go               # AICallLog 模型
│   ├── eval_set.go                  # EvalSet 模型
│   ├── ab_test.go                   # ABTest + ABTestResult 模型
│   ├── response_cache.go            # ResponseCache 模型
│   └── common.go                    # (已有) PaginatedResponse, ImportResult
├── handlers/
│   ├── task.go                      # TaskHandler: CRUD + SSE
│   ├── batch_test.go                # BatchTestHandler
│   ├── ab_test.go                   # ABTestHandler
│   ├── eval_set.go                  # EvalSetHandler
│   ├── ai_call_log.go               # AICallLogHandler
│   └── regression.go                # RegressionHandler
├── service/
│   ├── task_service.go              # 任务创建、查询、取消
│   ├── batch_service.go             # 批量测试业务逻辑
│   ├── ab_test_service.go          # A/B 测试业务逻辑
│   ├── eval_service.go             # 评测集管理 + AI 生成
│   ├── score_service.go            # 评分卡计算
│   ├── regression_service.go        # 回归检测逻辑
│   └── cache_service.go             # 缓存读写
└── (existing files unchanged)
```

---

## 七、实现优先级

### Phase 1: 基础设施（P0）

1. **Task 模型 + Worker 框架** — 所有其他功能的基石
2. **AICallLog 模型** — AI 调用可观测性
3. **ResponseCache 缓存层** — 降低 token 消耗和延迟
4. **TaskHandler API** — 任务 CRUD + SSE 通知

### Phase 2: 核心功能（P0）

5. **BatchTestHandler + BatchService** — 批量测试
6. **PromptScorer + ScoreService** — 4 维度评分
7. **ABTestHandler + ABTestService** — A/B 测试 + SPRT

### Phase 3: 扩展功能（P1）

8. **EvalSetHandler + EvalService** — 评测集 CRUD + AI 生成
9. **RegressionHandler + RegressionService** — 回归检测
10. **AICallLogHandler** — 调用日志查询 + 成本统计

---

## 八、关键技术决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| 任务 ID 生成 | UUID v4 | 无中心依赖，合并分片数据时不冲突 |
| 轮询策略 | 固定间隔（3s） | SQLite 无 LISTEN/NOTIFY，固定间隔简单可靠 |
| SSE 实现 | Gin 内置 SSE | Gin 原生支持，无需额外库 |
| 缓存 Key | SHA256(request) | 确定性哈希，支持语义相同但文本不同的请求去重 |
| SPRT 算法 | Wald's SPRT | 论文验证，序贯检验，平均样本量减少 30-50% |
| 批量 AI 调用 | OpenAI Batch API / 并发 Call | OpenAI Batch API 有 24h SLA；需要快速返回时用并发 Call |
| 扩展接口 | 新增接口而非修改现有 | AIProvider 保持向后兼容，不破坏现有 test.go 的同步测试 |

---

## 九、向后兼容性

- **AIProvider 接口不变**: 现有 `test.go` 中的同步 `Test()` / `Optimize()` 方法继续工作
- **ActivityLog 扩展**: 仅新增字段，原有字段和 API 完全兼容
- **路由新增**: 所有新 API 以 `/api/tasks`, `/api/eval-sets`, `/api/ab-tests` 等新前缀添加，不影响现有路由
- **数据库迁移**: 使用 GORM AutoMigrate 新增表，对已有表仅做字段扩展（向后兼容）
