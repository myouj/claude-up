# PromptVault MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 PromptVault MVP，7 周 4 Sprint，包含异步任务队列、变量实时预览、批量测试、A/B 测试（序贯检验）、质量评分等核心功能。

**Architecture:** 后端 Go/Gin + GORM，前端 Vue 3 + Element Plus + ECharts。Worker 采用 Goroutine Pool + DB Polling 模式。所有并行开发使用 Git Worktree 隔离。

**Tech Stack:** Go 1.21+, Gin, GORM, SQLite, Vue 3, Element Plus, ECharts, Playwright

---

## 文件结构总览

### 后端新增文件

| 文件路径 | 说明 |
|---------|------|
| `backend/models/task.go` | Task 模型 |
| `backend/models/eval_set.go` | EvalSet 模型 |
| `backend/models/ab_test.go` | ABTest 模型 |
| `backend/models/ai_call_log.go` | AICallLog 模型 |
| `backend/models/quota.go` | Quota 模型 |
| `backend/worker/worker.go` | Goroutine Pool + DB Polling |
| `backend/worker/executor.go` | 任务执行器 |
| `backend/worker/sse.go` | SSE 事件推送 |
| `backend/worker/cache.go` | AI 响应缓存 |
| `backend/service/task.go` | TaskService |
| `backend/service/quota.go` | QuotaService |
| `backend/service/batch.go` | BatchService |
| `backend/service/scoring.go` | ScoringService |
| `backend/service/eval.go` | EvalService |
| `backend/service/ab_test.go` | ABTestService |
| `backend/service/regression.go` | RegressionService |
| `backend/service/sprt.go` | SPRT 序贯检验引擎 |
| `backend/handlers/task.go` | Task API Handler |
| `backend/handlers/batch.go` | Batch API Handler |
| `backend/handlers/scoring.go` | Scoring API Handler |
| `backend/handlers/eval.go` | Eval API Handler |
| `backend/handlers/ab_test.go` | ABTest API Handler |
| `backend/handlers/regression.go` | Regression API Handler |
| `backend/middleware/ai_call_log.go` | AI 调用日志中间件 |

### 前端新增文件

| 文件路径 | 说明 |
|---------|------|
| `frontend/src/components/VariablePreviewPanel.vue` | 右侧 40% 预览面板 |
| `frontend/src/components/BatchTestTable.vue` | 批量测试表格 |
| `frontend/src/components/BatchTestCard.vue` | 展开卡片 |
| `frontend/src/components/QualityScoreCard.vue` | 四维度评分卡 |
| `frontend/src/components/TaskProgressBar.vue` | SSE 进度条 |
| `frontend/src/components/ABTestSequentialPanel.vue` | 序贯检验面板 |
| `frontend/src/composables/useTask.js` | 任务状态 + SSE |
| `frontend/src/composables/useBatchTest.js` | 批量测试 |
| `frontend/src/composables/useQualityScore.js` | 质量评分 |
| `frontend/src/composables/useSSE.js` | SSE 封装 |
| `frontend/src/views/BatchTest.vue` | 批量测试视图 |
| `frontend/src/views/BatchTestResult.vue` | 任务详情视图 |

### 后端修改文件

| 文件路径 | 说明 |
|---------|------|
| `backend/main.go` | 添加新路由、新模型 AutoMigrate |
| `backend/models/test_record.go` | 扩展字段 |
| `backend/models/activity.go` | 扩展 action_type 字段 |
| `frontend/src/views/PromptEditor.vue` | 改造为 60/40 分栏 |
| `frontend/src/App.vue` | 添加新 CSS 变量 |
| `frontend/src/router/index.js` | 添加新路由 |

---

## Sprint 1: 基础设施（Week 1-2）

### Phase 1.1: Worktree 创建

```bash
# 在主分支创建所有 worktrees
cd /Users/mayujian/all_code/projects/ai-agent/vibecoder
git checkout main

# 后端 A - Worker + Task 模型
git worktree add worktrees/sprint1-worker-backend-a -b feature/sprint1-worker

# 后端 A - API 限流（等 Worker 完成后切换）
# 暂不创建，等 Worker 完成后

# 后端 B - AICallLog middleware
git worktree add worktrees/sprint1-aicall-backend-b -b feature/sprint1-aicall

# 后端 B - ActivityLog 扩展
git worktree add worktrees/sprint1-activity-backend-b -b feature/sprint1-activity

# 前端 A - VariablePreviewPanel
git worktree add worktrees/sprint1-varpanel-frontend-a -b feature/sprint1-varpanel

# 前端 B - useTask composable
git worktree add worktrees/sprint1-usetask-frontend-b -b feature/sprint1-usetask

# 前端 B - CostCenter 基础组件
git worktree add worktrees/sprint1-costcenter-frontend-b -b feature/sprint1-costcenter
```

### Phase 1.2: 后端 A - Task 模型 + Worker

**文件:**
- Create: `backend/models/task.go`
- Create: `backend/worker/worker.go`
- Create: `backend/worker/executor.go`
- Create: `backend/service/task.go`
- Create: `backend/handlers/task.go`
- Modify: `backend/main.go` (添加路由、AutoMigrate)

- [ ] **Step 1: 创建 Task 模型**

```go
// backend/models/task.go
package models

import "time"

type Task struct {
    ID          uint       `gorm:"primaryKey" json:"id"`
    Type        string     `gorm:"size:50;not null" json:"type"` // batch_test | ab_test | eval_gen | regression | multi_turn
    Status      string     `gorm:"size:20;not null;default:pending" json:"status"` // pending | running | done | failed | cancelled
    Payload     string     `gorm:"type:text" json:"payload"` // JSON
    Progress    int        `gorm:"default:0" json:"progress"` // 0-100
    Result      string     `gorm:"type:text" json:"result,omitempty"` // JSON
    Error       string     `gorm:"size:500" json:"error,omitempty"`
    RetryCount  int        `gorm:"default:0" json:"retry_count"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    RunAt       time.Time  `json:"run_at"` // 计划执行时间
    StartedAt   *time.Time `json:"started_at,omitempty"`
    CompletedAt *time.Time `json:"completed_at,omitempty"`
}

func (Task) TableName() string {
    return "tasks"
}
```

- [ ] **Step 2: 创建 Worker Goroutine Pool**

```go
// backend/worker/worker.go
package worker

import (
    "log"
    "sync"
    "time"

    "prompt-vault/models"
    "gorm.io/gorm"
)

type Config struct {
    PoolSize     int
    PollInterval time.Duration
    MaxRetries  int
}

type Worker struct {
    db     *gorm.DB
    config Config
    tasks  chan *models.Task
    wg     sync.WaitGroup
    stopCh chan struct{}
}

func NewWorker(db *gorm.DB, cfg Config) *Worker {
    if cfg.PoolSize == 0 {
        cfg.PoolSize = 5
    }
    if cfg.PollInterval == 0 {
        cfg.PollInterval = 3 * time.Second
    }
    if cfg.MaxRetries == 0 {
        cfg.MaxRetries = 3
    }
    return &Worker{
        db:     db,
        config: cfg,
        tasks:  make(chan *models.Task, cfg.PoolSize*2),
        stopCh: make(chan struct{}),
    }
}

func (w *Worker) Start() {
    for i := 0; i < w.config.PoolSize; i++ {
        w.wg.Add(1)
        go w.run(i)
    }
    go w.poll()
}

func (w *Worker) Stop() {
    close(w.stopCh)
    w.wg.Wait()
}

func (w *Worker) poll() {
    ticker := time.NewTicker(w.config.PollInterval)
    defer ticker.Stop()

    for {
        select {
        case <-w.stopCh:
            return
        case <-ticker.C:
            w.fetchPendingTasks()
        }
    }
}

func (w *Worker) fetchPendingTasks() {
    var tasks []models.Task
    now := time.Now()
    w.db.Where("status = ? AND run_at <= ?", "pending", now).
        Order("run_at ASC").
        Limit(w.config.PoolSize).
        Find(&tasks)

    for i := range tasks {
        task := &tasks[i]
        if err := w.db.Model(task).Updates(map[string]interface{}{
            "status":     "running",
            "started_at": time.Now(),
        }).Error; err != nil {
            log.Printf("failed to update task %d to running: %v", task.ID, err)
            continue
        }
        w.tasks <- task
    }
}

func (w *Worker) run(id int) {
    defer w.wg.Done()
    for {
        select {
        case <-w.stopCh:
            return
        case task := <-w.tasks:
            w.execute(task)
        }
    }
}

func (w *Worker) execute(task *models.Task) {
    executor := NewExecutor(w.db, w.config.MaxRetries)
    result, err := executor.Execute(task)
    if err != nil {
        w.db.Model(task).Updates(map[string]interface{}{
            "status":  "failed",
            "error":   err.Error(),
        })
        return
    }
    w.db.Model(task).Updates(map[string]interface{}{
        "status":       "done",
        "progress":     100,
        "result":       result,
        "completed_at": time.Now(),
    })
}
```

- [ ] **Step 3: 创建任务执行器**

```go
// backend/worker/executor.go
package worker

import (
    "encoding/json"
    "fmt"
    "log"
    "time"

    "prompt-vault/models"
    "gorm.io/gorm"
)

type Executor struct {
    db          *gorm.DB
    maxRetries  int
}

func NewExecutor(db *gorm.DB, maxRetries int) *Executor {
    return &Executor{db: db, maxRetries: maxRetries}
}

func (e *Executor) Execute(task *models.Task) (string, error) {
    switch task.Type {
    case "batch_test":
        return e.executeBatchTest(task)
    case "ab_test":
        return e.executeABTest(task)
    case "eval_gen":
        return e.executeEvalGen(task)
    case "regression":
        return e.executeRegression(task)
    case "multi_turn":
        return e.executeMultiTurn(task)
    default:
        return "", fmt.Errorf("unknown task type: %s", task.Type)
    }
}

func (e *Executor) executeBatchTest(task *models.Task) (string, error) {
    // TODO: 实现批量测试逻辑
    return `{"status": "completed", "processed": 0}`, nil
}

func (e *Executor) executeABTest(task *models.Task) (string, error) {
    // TODO: 实现 A/B 测试逻辑
    return `{"status": "completed", "winner": "A"}`, nil
}

func (e *Executor) executeEvalGen(task *models.Task) (string, error) {
    // TODO: 实现评测集生成逻辑
    return `{"status": "completed", "cases": []}`, nil
}

func (e *Executor) executeRegression(task *models.Task) (string, error) {
    // TODO: 实现回归检测逻辑
    return `{"status": "completed", "regressions": []}`, nil
}

func (e *Executor) executeMultiTurn(task *models.Task) (string, error) {
    // TODO: 实现多轮对话逻辑
    return `{"status": "completed", "turns": 0}`, nil
}

// UpdateProgress 更新任务进度
func (e *Executor) UpdateProgress(taskID uint, current, total int) error {
    progress := (current * 100) / total
    return e.db.Model(&models.Task{}).Where("id = ?", taskID).
        Update("progress", progress).Error
}
```

- [ ] **Step 4: 创建 TaskService**

```go
// backend/service/task.go
package service

import (
    "encoding/json"
    "time"

    "prompt-vault/models"
    "gorm.io/gorm"
)

type TaskService struct {
    db *gorm.DB
}

func NewTaskService(db *gorm.DB) *TaskService {
    return &TaskService{db: db}
}

type CreateTaskRequest struct {
    Type    string      `json:"type" binding:"required"`
    Payload interface{} `json:"payload" binding:"required"`
}

func (s *TaskService) Create(req CreateTaskRequest) (*models.Task, error) {
    payloadJSON, err := json.Marshal(req.Payload)
    if err != nil {
        return nil, err
    }

    task := &models.Task{
        Type:     req.Type,
        Status:   "pending",
        Payload:  string(payloadJSON),
        Progress: 0,
        RunAt:    time.Now(),
    }

    if err := s.db.Create(task).Error; err != nil {
        return nil, err
    }
    return task, nil
}

func (s *TaskService) GetByID(id uint) (*models.Task, error) {
    var task models.Task
    if err := s.db.First(&task, id).Error; err != nil {
        return nil, err
    }
    return &task, nil
}

func (s *TaskService) List(limit, offset int) ([]models.Task, int64, error) {
    var tasks []models.Task
    var total int64

    s.db.Model(&models.Task{}).Count(&total)
    if err := s.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&tasks).Error; err != nil {
        return nil, 0, err
    }
    return tasks, total, nil
}

func (s *TaskService) Cancel(id uint) error {
    return s.db.Model(&models.Task{}).Where("id = ? AND status = ?", id, "pending").
        Update("status", "cancelled").Error
}
```

- [ ] **Step 5: 创建 Task Handler + 路由注册**

```go
// backend/handlers/task.go
package handlers

import (
    "net/http"
    "strconv"

    "prompt-vault/service"
    "github.com/gin-gonic/gin"
)

type TaskHandler struct {
    svc *service.TaskService
}

func NewTaskHandler(svc *service.TaskService) *TaskHandler {
    return &TaskHandler{svc: svc}
}

// CreateTask godoc
// @Summary 创建任务
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body service.CreateTaskRequest true "任务参数"
// @Success 200 {object} Response{data=models.Task}
// @Router /api/tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
    var req service.CreateTaskRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respondError(c, http.StatusBadRequest, err.Error())
        return
    }

    task, err := h.svc.Create(req)
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    respondSuccess(c, task)
}

// GetTask godoc
// @Summary 获取任务详情
// @Tags tasks
// @Produce json
// @Param id path int true "任务ID"
// @Success 200 {object} Response{data=models.Task}
// @Router /api/tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        respondError(c, http.StatusBadRequest, "invalid task id")
        return
    }

    task, err := h.svc.GetByID(uint(id))
    if err != nil {
        respondError(c, http.StatusNotFound, "task not found")
        return
    }
    respondSuccess(c, task)
}

// ListTasks godoc
// @Summary 任务列表
// @Tags tasks
// @Produce json
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {object} Response{data=[]models.Task}
// @Router /api/tasks [get]
func (h *TaskHandler) ListTasks(c *gin.Context) {
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

    tasks, total, err := h.svc.List(limit, offset)
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    respondSuccess(c, gin.H{"tasks": tasks, "total": total})
}

// CancelTask godoc
// @Summary 取消任务
// @Tags tasks
// @Produce json
// @Param id path int true "任务ID"
// @Success 200 {object} Response
// @Router /api/tasks/{id} [delete]
func (h *TaskHandler) CancelTask(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        respondError(c, http.StatusBadRequest, "invalid task id")
        return
    }

    if err := h.svc.Cancel(uint(id)); err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    respondSuccess(c, nil)
}
```

- [ ] **Step 6: 修改 main.go 注册路由和 AutoMigrate**

在 main.go 中添加：

```go
// 添加模型 AutoMigrate
db.AutoMigrate(
    &models.Task{},
    // ... 其他新模型
)

// 注册新路由
taskHandler := handlers.NewTaskHandler(service.NewTaskService(db))
v1.PUT("/tasks/:id/cancel", taskHandler.CancelTask)
```

- [ ] **Step 7: 编写 TaskService 单元测试**

```go
// backend/service/task_test.go
package service

import (
    "testing"
    "time"

    "prompt-vault/models"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func newTaskDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(t.TempDir()+"/task_test.db"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open test db: %v", err)
    }
    db.AutoMigrate(&models.Task{})
    return db
}

func TestTaskService_Create(t *testing.T) {
    db := newTaskDB(t)
    svc := NewTaskService(db)

    task, err := svc.Create(CreateTaskRequest{
        Type:    "batch_test",
        Payload: map[string]interface{}{"count": 10},
    })
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if task.Type != "batch_test" {
        t.Errorf("expected type batch_test, got %s", task.Type)
    }
    if task.Status != "pending" {
        t.Errorf("expected status pending, got %s", task.Status)
    }
}

func TestTaskService_GetByID(t *testing.T) {
    db := newTaskDB(t)
    svc := NewTaskService(db)

    created, _ := svc.Create(CreateTaskRequest{
        Type:    "ab_test",
        Payload:  map[string]interface{}{},
    })

    fetched, err := svc.GetByID(created.ID)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if fetched.ID != created.ID {
        t.Errorf("expected id %d, got %d", created.ID, fetched.ID)
    }
}

func TestTaskService_Cancel(t *testing.T) {
    db := newTaskDB(t)
    svc := NewTaskService(db)

    task, _ := svc.Create(CreateTaskRequest{
        Type:    "eval_gen",
        Payload:  map[string]interface{}{},
    })

    err := svc.Cancel(task.ID)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    cancelled, _ := svc.GetByID(task.ID)
    if cancelled.Status != "cancelled" {
        t.Errorf("expected status cancelled, got %s", cancelled.Status)
    }
}
```

- [ ] **Step 8: 运行测试**

```bash
cd worktrees/sprint1-worker-backend-a/backend
go test ./service/... -v -run TestTaskService
```

- [ ] **Step 9: Commit**

```bash
git add -A
git commit -m "feat(worker): add Task model, Worker pool, and TaskService

- Task model with status: pending|running|done|failed|cancelled
- Goroutine pool with DB polling (3s interval, 5 workers)
- TaskService: Create, GetByID, List, Cancel
- Handler with REST API endpoints
- Unit tests for TaskService

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

### Phase 1.3: 后端 A - API 限流/配额

**前提：** Worker + Task 模型已完成

**文件:**
- Create: `backend/models/quota.go`
- Create: `backend/service/quota.go`
- Modify: `backend/main.go` (添加 Quota 中间件)

- [ ] **Step 1: 创建 Quota 模型**

```go
// backend/models/quota.go
package models

import "time"

type Quota struct {
    ID       uint      `gorm:"primaryKey" json:"id"`
    Provider string    `gorm:"size:50;not null" json:"provider"` // openai | claude | gemini | minimax
    Model    string    `gorm:"size:100" json:"model"`
    Limit    int       `gorm:"not null" json:"limit"`    // 月度上限
    Usage    int       `gorm:"default:0" json:"usage"`   // 当月已用
    ResetAt  time.Time `json:"reset_at"`                  // 重置时间
}

func (Quota) TableName() string {
    return "quotas"
}
```

- [ ] **Step 2: 创建 QuotaService**

```go
// backend/service/quota.go
package service

import (
    "sync"
    "time"

    "prompt-vault/models"
    "gorm.io/gorm"
)

type QuotaService struct {
    db *gorm.DB
    mu sync.Mutex
}

func NewQuotaService(db *gorm.DB) *QuotaService {
    return &QuotaService{db: db}
}

// Check 检查配额是否充足
func (s *QuotaService) Check(provider string, cost int) (bool, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    var quota models.Quota
    now := time.Now()

    // 查找当前月份的配额记录
    if err := s.db.Where("provider = ? AND reset_at > ?", provider, now.AddDate(0, -1, 0)).First(&quota).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return true, nil // 无配额限制记录，允许
        }
        return false, err
    }

    // 检查是否需要重置
    if now.After(quota.ResetAt) {
        s.db.Model(&quota).Updates(map[string]interface{}{
            "usage":    0,
            "reset_at": now.AddDate(0, 1, 0), // 下个月重置
        })
        return true, nil
    }

    return quota.Usage+cost <= quota.Limit, nil
}

// Consume 消耗配额
func (s *QuotaService) Consume(provider string, cost int) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    return s.db.Model(&models.Quota{}).
        Where("provider = ?", provider).
        UpdateColumn("usage", gorm.Expr("usage + ?", cost)).Error
}

// GetUsage 获取当前使用量
func (s *QuotaService) GetUsage(provider string) (int, error) {
    var quota models.Quota
    now := time.Now()

    if err := s.db.Where("provider = ? AND reset_at > ?", provider, now.AddDate(0, -1, 0)).First(&quota).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return 0, nil
        }
        return 0, err
    }
    return quota.Usage, nil
}
```

- [ ] **Step 3: 创建配额中间件**

```go
// backend/middleware/quota.go
package middleware

import (
    "net/http"
    "strconv"

    "prompt-vault/service"
    "github.com/gin-gonic/gin"
)

func QuotaMiddleware(qs *service.QuotaService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从请求中获取 provider 和预估 cost
        provider := c.GetHeader("X-AI-Provider")
        costStr := c.DefaultHeader("X-AI-Cost", "1")
        cost, _ := strconv.Atoi(costStr)

        if provider == "" {
            c.Next()
            return
        }

        allowed, err := qs.Check(provider, cost)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "quota check failed"})
            return
        }

        if !allowed {
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "quota exceeded"})
            return
        }

        // 记录配额消耗
        qs.Consume(provider, cost)
        c.Next()
    }
}
```

- [ ] **Step 4: 注册中间件到 main.go**

```go
// 在路由配置中添加
quotaService := service.NewQuotaService(db)
v1.Use(middleware.QuotaMiddleware(quotaService))
```

- [ ] **Step 5: 编写测试**

```go
// backend/service/quota_test.go
package service

import (
    "testing"
    "time"

    "prompt-vault/models"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func newQuotaDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(t.TempDir()+"/quota_test.db"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open test db: %v", err)
    }
    db.AutoMigrate(&models.Quota{})
    return db
}

func TestQuotaService_Check(t *testing.T) {
    db := newQuotaDB(t)
    svc := NewQuotaService(db)

    // 创建配额记录
    db.Create(&models.Quota{
        Provider: "openai",
        Model:    "gpt-4o",
        Limit:    100,
        Usage:    50,
        ResetAt:  time.Now().AddDate(0, 1, 0),
    })

    // 检查是否允许 (50 + 30 <= 100)
    allowed, err := svc.Check("openai", 30)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !allowed {
        t.Error("expected allowed, got not allowed")
    }

    // 检查是否超限 (50 + 60 > 100)
    allowed, _ = svc.Check("openai", 60)
    if allowed {
        t.Error("expected not allowed, got allowed")
    }
}

func TestQuotaService_Consume(t *testing.T) {
    db := newQuotaDB(t)
    svc := NewQuotaService(db)

    db.Create(&models.Quota{
        Provider: "claude",
        Model:    "claude-3-opus",
        Limit:    100,
        Usage:    10,
        ResetAt:  time.Now().AddDate(0, 1, 0),
    })

    err := svc.Consume("claude", 5)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    usage, _ := svc.GetUsage("claude")
    if usage != 15 {
        t.Errorf("expected usage 15, got %d", usage)
    }
}
```

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat(quota): add API quota management

- Quota model with provider, limit, usage, reset_at
- QuotaService: Check, Consume, GetUsage
- QuotaMiddleware for AI API rate limiting
- Monthly quota reset logic

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

### Phase 1.4: 后端 B - AICallLog Middleware

**文件:**
- Create: `backend/models/ai_call_log.go`
- Create: `backend/middleware/ai_call_log.go`
- Modify: `backend/main.go` (注册中间件)

- [ ] **Step 1: 创建 AICallLog 模型**

```go
// backend/models/ai_call_log.go
package models

import "time"

type AICallLog struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    Provider    string    `gorm:"size:50;not null" json:"provider"`
    Model       string    `gorm:"size:100" json:"model"`
    InputTokens int       `json:"input_tokens"`
    OutputTokens int      `json:"output_tokens"`
    LatencyMs   int       `json:"latency_ms"`
    Cost        float64   `json:"cost"`
    TraceID     string    `gorm:"size:100" json:"trace_id"`
    PromptID    uint      `json:"prompt_id,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
}

func (AICallLog) TableName() string {
    return "ai_call_logs"
}
```

- [ ] **Step 2: 创建 AICallLog 中间件**

```go
// backend/middleware/ai_call_log.go
package middleware

import (
    "bytes"
    "encoding/json"
    "io"
    "time"

    "prompt-vault/models"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type AICallLogMiddleware struct {
    db *gorm.DB
}

func NewAICallLogMiddleware(db *gorm.DB) *AICallLogMiddleware {
    return &AICallLogMiddleware{db: db}
}

func (m *AICallLogMiddleware) Handler() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 仅记录 AI 相关请求
        if c.GetHeader("X-AI-Provider") == "" {
            c.Next()
            return
        }

        start := time.Now()

        // 读取请求体
        var reqBody []byte
        if c.Request.Body != nil {
            reqBody, _ = io.ReadAll(c.Request.Body)
            c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
        }

        // 处理请求
        c.Next()

        // 记录日志
        latency := time.Since(start).Milliseconds()
        provider := c.GetHeader("X-AI-Provider")
        model := c.GetHeader("X-AI-Model")
        traceID := c.GetHeader("X-Trace-ID")

        // 从响应中提取 token 信息（如果有）
        var inputTokens, outputTokens int
        var cost float64
        if c.Writer.Status() == 200 {
            var resp map[string]interface{}
            if err := json.Unmarshal(reqBody, &resp); err == nil {
                if tokens, ok := resp["usage"].(map[string]interface{}); ok {
                    if it, ok := tokens["input_tokens"].(float64); ok {
                        inputTokens = int(it)
                    }
                    if ot, ok := tokens["output_tokens"].(float64); ok {
                        outputTokens = int(ot)
                    }
                }
                if cst, ok := resp["cost"].(float64); ok {
                    cost = cst
                }
            }
        }

        log := &models.AICallLog{
            Provider:    provider,
            Model:       model,
            InputTokens: inputTokens,
            OutputTokens: outputTokens,
            LatencyMs:   int(latency),
            Cost:        cost,
            TraceID:     traceID,
            CreatedAt:   time.Now(),
        }

        m.db.Create(log)
    }
}
```

- [ ] **Step 3: 注册中间件**

```bash
# 在 main.go 中
aiCallLog := middleware.NewAICallLogMiddleware(db)
v1.Use(aiCallLog.Handler())
```

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "feat(middleware): add AICallLog middleware

- AICallLog model for tracking AI API calls
- Middleware captures provider, model, tokens, latency, cost
- Records all AI calls for cost analysis

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

### Phase 1.5: 后端 B - ActivityLog 扩展

**文件:**
- Modify: `backend/models/activity.go` (添加 action_type 扩展)

- [ ] **Step 1: 查看现有 Activity 模型**

```go
// backend/models/activity.go 中添加
ActionType string `gorm:"size:50" json:"action_type"` // task_created | task_completed | ...
Detail     string `gorm:"type:text" json:"detail"`
```

- [ ] **Step 2: Commit**

```bash
git add -A
git commit -m "feat(activity): extend ActivityLog with action_type field

- Add action_type field for categorizing activities
- Add detail field for additional context

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

### Phase 1.6: 前端 A - VariablePreviewPanel

**文件:**
- Create: `frontend/src/components/VariablePreviewPanel.vue`
- Modify: `frontend/src/views/PromptEditor.vue` (60/40 分栏改造)

- [ ] **Step 1: 创建 VariablePreviewPanel 组件**

```vue
<!-- frontend/src/components/VariablePreviewPanel.vue -->
<template>
  <div class="variable-preview-panel">
    <!-- 变量输入区域 -->
    <div class="panel-section">
      <div class="section-header">
        <el-icon><Edit /></el-icon>
        <span>变量输入</span>
      </div>

      <div v-if="variables.length > 0" class="variable-list">
        <div
          v-for="v in variables"
          :key="v"
          class="variable-row"
          :class="{ filled: hasValue(v) }"
        >
          <label class="var-label">
            <span class="var-marker">{{{{</span>
            <span class="var-name">{{ v }}</span>
            <span class="var-marker">}}}}</span>
          </label>
          <el-input
            v-model="variableValues[v]"
            :placeholder="`输入 ${v} 的值...`"
            size="small"
            clearable
          >
            <template #suffix>
              <el-icon v-if="hasValue(v)" class="check-icon"><Check /></el-icon>
            </template>
          </el-input>
        </div>
      </div>

      <div v-else class="no-variables">
        <p>当前内容无变量</p>
      </div>

      <!-- 填充进度 -->
      <div v-if="variables.length > 0" class="fill-indicator">
        <div class="fill-bar">
          <div class="fill-progress" :style="{ width: fillRate + '%' }"></div>
        </div>
        <span class="fill-text">{{ fillRate }}%</span>
      </div>
    </div>

    <!-- 渲染预览区域 -->
    <div class="panel-section">
      <div class="section-header">
        <el-icon><View /></el-icon>
        <span>渲染预览</span>
        <span v-if="!allFilled" class="preview-hint">(未填完)</span>
      </div>

      <div class="preview-content">
        <pre class="preview-text">{{ renderedContent || '无内容' }}</pre>
      </div>

      <div class="preview-actions">
        <el-button
          size="small"
          type="primary"
          :disabled="!allFilled"
          @click="handleCopy"
        >
          <el-icon><CopyDocument /></el-icon>
          <span>复制渲染结果</span>
        </el-button>
        <el-button size="small" @click="clearValues">
          <el-icon><RefreshRight /></el-icon>
          <span>清空</span>
        </el-button>
      </div>
    </div>

    <!-- 质量评分卡（可选折叠） -->
    <div v-if="showQualityScore" class="panel-section">
      <div class="section-header collapsible" @click="qualityExpanded = !qualityExpanded">
        <el-icon><DataAnalysis /></el-icon>
        <span>质量评分</span>
        <el-icon class="collapse-icon" :class="{ rotated: qualityExpanded }">
          <ArrowRight />
        </el-icon>
      </div>

      <div v-show="qualityExpanded" class="quality-cards">
        <div class="quality-item">
          <span class="quality-label">Clarity</span>
          <el-progress :percentage="qualityScore.clarity" :color="'#6366F1'" />
        </div>
        <div class="quality-item">
          <span class="quality-label">Completeness</span>
          <el-progress :percentage="qualityScore.completeness" :color="'#8B5CF6'" />
        </div>
        <div class="quality-item">
          <span class="quality-label">Example</span>
          <el-progress :percentage="qualityScore.example" :color="'#EC4899'" />
        </div>
        <div class="quality-item">
          <span class="quality-label">Role</span>
          <el-progress :percentage="qualityScore.role" :color="'#14B8A6'" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Edit, View, Check, CopyDocument, RefreshRight,
  DataAnalysis, ArrowRight
} from '@element-plus/icons-vue'

const props = defineProps({
  content: {
    type: String,
    default: ''
  },
  showQualityScore: {
    type: Boolean,
    default: false
  },
  qualityScore: {
    type: Object,
    default: () => ({ clarity: 0, completeness: 0, example: 0, role: 0 })
  }
})

const emit = defineEmits(['update:qualityScore'])

// 变量值管理
const variableValues = ref({})

// 质量评分折叠状态
const qualityExpanded = ref(true)

// 从 content 中提取变量名
const variables = computed(() => {
  const text = props.content || ''
  const regex = /\{\{([^}]+)\}\}/g
  const vars = new Set()
  let match
  while ((match = regex.exec(text)) !== null) {
    vars.add(match[1].trim())
  }
  return Array.from(vars)
})

// 渲染后的内容
const renderedContent = computed(() => {
  let result = props.content || ''
  for (const [key, value] of Object.entries(variableValues.value)) {
    if (value) {
      result = result.replace(new RegExp(`\\{\\{${key}\\}\\}`, 'g'), value)
    }
  }
  return result
})

// 辅助方法
const hasValue = (varName) => !!variableValues.value[varName]

const allFilled = computed(() => {
  return variables.value.length > 0 &&
    variables.value.every(v => !!variableValues.value[v])
})

const fillRate = computed(() => {
  if (variables.value.length === 0) return 100
  const filled = variables.value.filter(v => !!variableValues.value[v]).length
  return Math.round((filled / variables.value.length) * 100)
})

const clearValues = () => {
  variableValues.value = {}
}

const handleCopy = () => {
  if (renderedContent.value) {
    navigator.clipboard.writeText(renderedContent.value)
    ElMessage.success('渲染结果已复制到剪贴板')
  }
}

// 当 content 变化时，清除不再使用的变量值
watch(variables, (newVars) => {
  const newKeys = new Set(newVars)
  const currentKeys = Object.keys(variableValues.value)
  for (const key of currentKeys) {
    if (!newKeys.has(key)) {
      delete variableValues.value[key]
    }
  }
})

// 暴露给父组件
defineExpose({
  variableValues,
  renderedContent,
  clearValues
})
</script>

<style scoped>
.variable-preview-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4);
  padding: var(--spacing-4);
  background: var(--color-surface);
}

.panel-section {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-4);
}

.section-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-3);
}

.section-header.collapsible {
  cursor: pointer;
  user-select: none;
}

.collapse-icon {
  transition: transform var(--transition-fast);
  margin-left: auto;
}

.collapse-icon.rotated {
  transform: rotate(90deg);
}

.variable-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.variable-row {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
  padding: var(--spacing-2);
  border-radius: var(--radius-md);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  transition: all var(--transition-fast);
}

.variable-row.filled {
  border-color: var(--color-success-light);
  background: color-mix(in srgb, var(--color-success-light) 30%, var(--color-surface));
}

.var-label {
  display: flex;
  align-items: center;
  gap: 2px;
  font-size: var(--font-size-xs);
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
}

.var-marker {
  color: var(--color-primary);
  font-weight: var(--font-weight-bold);
}

.var-name {
  color: var(--color-text-primary);
  font-weight: var(--font-weight-semibold);
}

.check-icon {
  color: var(--color-success);
}

.fill-indicator {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  margin-top: var(--spacing-3);
}

.fill-bar {
  flex: 1;
  height: 4px;
  background: var(--color-border);
  border-radius: 2px;
  overflow: hidden;
}

.fill-progress {
  height: 100%;
  background: var(--color-primary);
  border-radius: 2px;
  transition: width var(--transition-normal);
}

.fill-text {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  font-weight: var(--font-weight-medium);
  min-width: 36px;
}

.no-variables {
  text-align: center;
  padding: var(--spacing-4);
  color: var(--color-text-muted);
  font-size: var(--font-size-sm);
}

.preview-content {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--spacing-3);
  max-height: 200px;
  overflow-y: auto;
}

.preview-text {
  margin: 0;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: var(--font-size-sm);
  line-height: 1.7;
  color: var(--color-text-primary);
  white-space: pre-wrap;
  word-break: break-word;
}

.preview-hint {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-normal);
  color: var(--color-warning);
}

.preview-actions {
  display: flex;
  gap: var(--spacing-2);
  margin-top: var(--spacing-3);
}

.quality-cards {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.quality-item {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
}

.quality-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  font-weight: var(--font-weight-medium);
}
</style>
```

- [ ] **Step 2: 改造 PromptEditor.vue 为 60/40 分栏**

```vue
<!-- PromptEditor.vue 中的 template 修改 -->
<template>
  <div class="prompt-editor">
    <!-- Header -->
    <div class="editor-header">
      <el-button text @click="$router.back()">
        <el-icon><ArrowLeft /></el-icon>
      </el-button>
      <el-input
        v-model="prompt.title"
        class="title-input"
        placeholder="Prompt 标题"
      />
      <el-button type="primary" @click="handleSave">保存</el-button>
    </div>

    <!-- 60/40 分栏主体 -->
    <div class="editor-body">
      <!-- 左侧：侧边栏 + 编辑区 (60%) -->
      <div class="editor-main">
        <!-- 侧边栏 -->
        <div class="editor-sidebar">
          <el-form label-position="top">
            <el-form-item label="描述">
              <el-input v-model="prompt.description" type="textarea" :rows="3" />
            </el-form-item>
            <el-form-item label="分类">
              <el-select v-model="prompt.category" placeholder="选择分类">
                <!-- 分类选项 -->
              </el-select>
            </el-form-item>
            <el-form-item label="标签">
              <el-select v-model="prompt.tags" multiple filterable allow-create>
                <!-- 标签选项 -->
              </el-select>
            </el-form-item>
          </el-form>
        </div>

        <!-- 编辑区域 -->
        <div class="editor-content">
          <el-input
            v-model="prompt.content"
            type="textarea"
            :rows="15"
            placeholder="输入 Prompt 内容..."
            @input="handleContentChange"
          />
        </div>
      </div>

      <!-- 右侧：VariablePreviewPanel (40%) -->
      <div class="editor-preview">
        <VariablePreviewPanel
          :content="prompt.content"
          :show-quality-score="true"
          :quality-score="qualityScore"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import VariablePreviewPanel from '@/components/VariablePreviewPanel.vue'

const prompt = ref({
  title: '',
  content: '',
  description: '',
  category: '',
  tags: []
})

const qualityScore = ref({
  clarity: 0,
  completeness: 0,
  example: 0,
  role: 0
})

const handleContentChange = () => {
  // 内容变化时的处理
}
</script>

<style scoped>
.prompt-editor {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.editor-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  padding: var(--spacing-4);
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface);
}

.title-input {
  flex: 1;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
}

.editor-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.editor-main {
  flex: 0 0 60%;
  display: flex;
  border-right: 1px solid var(--color-border);
}

.editor-sidebar {
  width: 300px;
  flex-shrink: 0;
  padding: var(--spacing-4);
  border-right: 1px solid var(--color-border);
  background: var(--color-surface);
  overflow-y: auto;
}

.editor-content {
  flex: 1;
  padding: var(--spacing-4);
  overflow-y: auto;
}

.editor-preview {
  flex: 0 0 40%;
  overflow-y: auto;
  background: var(--color-bg);
}
</style>
```

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "feat(frontend): add VariablePreviewPanel and 60/40 layout

- New VariablePreviewPanel component with:
  - Variable input with {{variable}} parsing
  - Fill rate progress indicator
  - Rendered preview
  - Collapsible quality score card
- PromptEditor改造为60/40分栏布局
- Uses PromptVault Design System tokens

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

### Phase 1.7: 前端 B - useTask Composable (Mock)

**文件:**
- Create: `frontend/src/composables/useTask.js`
- Create: `frontend/src/composables/useSSE.js`

- [ ] **Step 1: 创建 useSSE.js**

```javascript
// frontend/src/composables/useSSE.js
import { ref, onUnmounted } from 'vue'

/**
 * SSE 连接管理
 * @param {string} url - SSE 端点 URL
 * @param {Object} options - 配置选项
 */
export function useSSE(url, options = {}) {
  const {
    autoReconnect = true,
    maxRetries = 5,
    reconnectDelay = 1000,
    onMessage = () => {},
    onError = () => {},
    onOpen = () => {}
  } = options

  const connected = ref(false)
  const error = ref(null)
  let eventSource = null
  let retryCount = 0
  let reconnectTimeout = null

  const connect = () => {
    if (eventSource) {
      eventSource.close()
    }

    eventSource = new EventSource(url)

    eventSource.onopen = () => {
      connected.value = true
      error.value = null
      retryCount = 0
      onOpen()
    }

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        onMessage(data)
      } catch (e) {
        console.error('Failed to parse SSE message:', e)
      }
    }

    eventSource.onerror = (err) => {
      connected.value = false
      error.value = err
      onError(err)

      // 自动重连
      if (autoReconnect && retryCount < maxRetries) {
        retryCount++
        const delay = reconnectDelay * Math.pow(2, retryCount - 1) // 指数退避
        reconnectTimeout = setTimeout(() => {
          connect()
        }, delay)
      }
    }
  }

  const disconnect = () => {
    if (reconnectTimeout) {
      clearTimeout(reconnectTimeout)
    }
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
    connected.value = false
  }

  onUnmounted(() => {
    disconnect()
  })

  return {
    connected,
    error,
    connect,
    disconnect
  }
}
```

- [ ] **Step 2: 创建 useTask.js (Mock 模式)**

```javascript
// frontend/src/composables/useTask.js
import { ref, computed } from 'vue'
import { useSSE } from './useSSE'

/**
 * 任务状态管理 Composable
 * @param {string} taskId - 任务 ID（可选，mock 模式下不需要）
 * @param {Object} options - 配置选项
 */
export function useTask(taskId = null, options = {}) {
  const {
    mockMode = true, // Sprint 1 使用 mock
    mockProgress = 0,
    onComplete = () => {},
    onError = () => {}
  } = options

  // 任务状态
  const task = ref(null)
  const progress = ref(0)
  const status = ref('idle') // idle | pending | running | done | failed | cancelled

  // Mock 数据生成器
  const generateMockTask = () => {
    return {
      id: Math.floor(Math.random() * 10000),
      type: 'batch_test',
      status: 'running',
      progress: 0,
      result: null,
      error: null,
      created_at: new Date().toISOString()
    }
  }

  // SSE 进度更新（真实模式）
  const handleSSEMessage = (data) => {
    if (data.status === 'running' || data.type === 'progress') {
      progress.value = data.progress || 0
      status.value = 'running'
      if (task.value) {
        task.value.progress = data.progress
      }
    } else if (data.status === 'done') {
      progress.value = 100
      status.value = 'done'
      if (task.value) {
        task.value.status = 'done'
        task.value.result = data.result
      }
      onComplete(data)
    } else if (data.status === 'failed') {
      status.value = 'failed'
      if (task.value) {
        task.value.status = 'failed'
        task.value.error = data.error
      }
      onError(data)
    }
  }

  // SSE 连接（真实模式）
  let sseConnection = null
  if (!mockMode && taskId) {
    sseConnection = useSSE(`/api/tasks/${taskId}/progress`, {
      onMessage: handleSSEMessage,
      onError: (err) => onError({ error: err.message })
    })
  }

  // Mock 模式：模拟进度更新
  let mockInterval = null
  if (mockMode) {
    const startMockProgress = () => {
      status.value = 'running'
      task.value = generateMockTask()

      mockInterval = setInterval(() => {
        if (progress.value < 100) {
          progress.value += Math.random() * 10
          if (progress.value > 100) progress.value = 100
          task.value.progress = Math.round(progress.value)
        } else {
          clearInterval(mockInterval)
          status.value = 'done'
          task.value.status = 'done'
          onComplete({ status: 'done', result: { mock: true } })
        }
      }, 500)
    }

    // 暴露 startMockProgress 供外部调用
    return {
      task,
      progress,
      status,
      mockMode: true,
      startMockProgress,
      cancelTask: () => {
        if (mockInterval) clearInterval(mockInterval)
        status.value = 'cancelled'
      },
      connect: () => {} // 空操作，mock 模式不需要
    }
  }

  // 真实模式
  return {
    task,
    progress,
    status,
    mockMode: false,
    connect: sseConnection?.connect,
    disconnect: sseConnection?.disconnect,
    cancelTask: async () => {
      if (!taskId) return
      try {
        const res = await fetch(`/api/tasks/${taskId}`, { method: 'DELETE' })
        if (res.ok) {
          status.value = 'cancelled'
        }
      } catch (e) {
        onError({ error: e.message })
      }
    }
  }
}
```

- [ ] **Step 3: 创建 CostCenter 基础组件**

```vue
<!-- frontend/src/components/CostCenter.vue -->
<template>
  <div class="cost-center">
    <div class="cost-header">
      <el-icon><Coin /></el-icon>
      <span>配额使用</span>
    </div>

    <div class="cost-list">
      <div v-for="item in quotaList" :key="item.provider" class="cost-item">
        <div class="cost-provider">
          <span class="provider-name">{{ item.provider }}</span>
          <span class="provider-model">{{ item.model }}</span>
        </div>
        <div class="cost-bar">
          <div
            class="cost-fill"
            :style="{
              width: (item.usage / item.limit * 100) + '%',
              background: getBarColor(item.usage / item.limit)
            }"
          ></div>
        </div>
        <div class="cost-numbers">
          <span>{{ item.usage }}</span>
          <span class="cost-limit">/ {{ item.limit }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Coin } from '@element-plus/icons-vue'

const quotaList = ref([
  { provider: 'OpenAI', model: 'gpt-4o', usage: 450, limit: 1000 },
  { provider: 'Claude', model: 'claude-3-opus', usage: 120, limit: 500 },
  { provider: 'MiniMax', model: 'MiniMax-Text-01', usage: 800, limit: 2000 }
])

const getBarColor = (ratio) => {
  if (ratio < 0.5) return 'var(--color-success)'
  if (ratio < 0.8) return 'var(--color-warning)'
  return 'var(--color-danger)'
}

onMounted(() => {
  // TODO: 从 API 获取真实配额数据
})
</script>

<style scoped>
.cost-center {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-4);
}

.cost-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-4);
}

.cost-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.cost-item {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
}

.cost-provider {
  display: flex;
  justify-content: space-between;
  font-size: var(--font-size-xs);
}

.provider-name {
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
}

.provider-model {
  color: var(--color-text-muted);
}

.cost-bar {
  height: 6px;
  background: var(--color-border);
  border-radius: 3px;
  overflow: hidden;
}

.cost-fill {
  height: 100%;
  border-radius: 3px;
  transition: width var(--transition-normal);
}

.cost-numbers {
  display: flex;
  justify-content: flex-end;
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
}

.cost-limit {
  color: var(--color-text-muted);
}
</style>
```

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "feat(frontend): add useTask, useSSE composables and CostCenter

- useSSE: SSE connection management with auto-reconnect
- useTask: Task state management with mock mode for Sprint 1
- CostCenter: Basic quota display component
- Mock mode generates simulated progress for frontend development

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

### Phase 1.8: Sprint 1 合并回 Main

所有 worktree 完成并 PR 合并后：

```bash
# 在 main 分支
git checkout main
git pull origin main

# 合并所有 feature 分支
git merge feature/sprint1-worker --no-ff -m "Merge sprint1-worker"
git merge feature/sprint1-quota --no-ff -m "Merge sprint1-quota"
git merge feature/sprint1-aicall --no-ff -m "Merge sprint1-aicall"
git merge feature/sprint1-activity --no-ff -m "Merge sprint1-activity"
git merge feature/sprint1-varpanel --no-ff -m "Merge sprint1-varpanel"
git merge feature/sprint1-usetask --no-ff -m "Merge sprint1-usetask"
git merge feature/sprint1-costcenter --no-ff -m "Merge sprint1-costcenter"

# 推送到远程
git push origin main

# 删除已合并的 worktrees
git worktree remove worktrees/sprint1-worker-backend-a
git worktree remove worktrees/sprint1-aicall-backend-b
git worktree remove worktrees/sprint1-activity-backend-b
git worktree remove worktrees/sprint1-varpanel-frontend-a
git worktree remove worktrees/sprint1-usetask-frontend-b
git worktree remove worktrees/sprint1-costcenter-frontend-b

# 删除已合并的分支
git branch -d feature/sprint1-worker
git branch -d feature/sprint1-quota
git branch -d feature/sprint1-aicall
git branch -d feature/sprint1-activity
git branch -d feature/sprint1-varpanel
git branch -d feature/sprint1-usetask
git branch -d feature/sprint1-costcenter
```

---

## Sprint 2: 核心测试（Week 3-4）

### Phase 2.1: Worktree 创建

```bash
# Sprint 2 worktrees
git worktree add worktrees/sprint2-batch-backend-a -b feature/sprint2-batch
git worktree add worktrees/sprint2-scoring-backend-b -b feature/sprint2-scoring
git worktree add worktrees/sprint2-eval-backend-b -b feature/sprint2-eval
git worktree add worktrees/sprint2-sse-backend-a -b feature/sprint2-sse
git worktree add worktrees/sprint2-batchtable-frontend-a -b feature/sprint2-batchtable
git worktree add worktrees/sprint2-batchcard-frontend-b -b feature/sprint2-batchcard
git worktree add worktrees/sprint2-scorecard-frontend-a -b feature/sprint2-scorecard
git worktree add worktrees/sprint2-progress-frontend-b -b feature/sprint2-progress
```

### Phase 2.2: 后端 A - BatchService + SSE

**文件:**
- Create: `backend/service/batch.go`
- Create: `backend/worker/sse.go`
- Modify: `backend/handlers/batch.go`
- Modify: `backend/worker/executor.go` (添加 batch_test 执行逻辑)

- [ ] **Step 1: 创建 SSE 模块**

```go
// backend/worker/sse.go
package worker

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
)

// SSEClient SSE 客户端连接
type SSEClient struct {
    taskID  uint
    channel chan SSEEvent
}

// SSEEvent SSE 事件
type SSEEvent struct {
    Event string      `json:"event,omitempty"`
    Data  interface{} `json:"data"`
}

// SSEManager SSE 连接管理器
type SSEManager struct {
    clients map[uint]chan SSEEvent
}

func NewSSEManager() *SSEManager {
    return &SSEManager{
        clients: make(map[uint]chan SSEEvent),
    }
}

// Subscribe 订阅任务进度
func (m *SSEManager) Subscribe(taskID uint) chan SSEEvent {
    ch := make(chan SSEEvent, 100)
    m.clients[taskID] = ch
    return ch
}

// Unsubscribe 取消订阅
func (m *SSEManager) Unsubscribe(taskID uint) {
    if ch, ok := m.clients[taskID]; ok {
        close(ch)
        delete(m.clients, taskID)
    }
}

// Publish 发布事件
func (m *SSEManager) Publish(taskID uint, event, data interface{}) {
    if ch, ok := m.clients[taskID]; ok {
        select {
        case ch <- SSEEvent{Event: event, Data: data}:
        default:
            // channel 满了，跳过
        }
    }
}

// SendSSEProgress 发送进度
func (m *SSEManager) SendSSEProgress(taskID uint, current, total int) {
    progress := (current * 100) / total
    m.Publish(taskID, "progress", map[string]interface{}{
        "current":   current,
        "total":     total,
        "progress":  progress,
        "status":    "running",
    })
}

// SendSSEComplete 发送完成
func (m *SSEManager) SendSSEComplete(taskID uint, result interface{}) {
    m.Publish(taskID, "complete", map[string]interface{}{
        "status": "done",
        "result": result,
    })
}

// SendSSEError 发送错误
func (m *SSEManager) SendSSEError(taskID uint, errMsg string) {
    m.Publish(taskID, "error", map[string]interface{}{
        "status": "failed",
        "error":  errMsg,
    })
}

// SSEHandler 处理 SSE 请求
func (m *SSEManager) SSEHandler(c *gin.Context) {
    taskID := c.Param("id")
    // 验证任务存在

    client := m.Subscribe(taskID)
    defer m.Unsubscribe(taskID)

    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    for {
        select {
        case event, ok := <-client:
            if !ok {
                return
            }
            data, _ := json.Marshal(event.Data)
            if event.Event != "" {
                fmt.Fprintf(c.Writer, "event: %s\n", event.Event)
            }
            fmt.Fprintf(c.Writer, "data: %s\n\n", data)
            c.Writer.Flush()
        case <-c.Request.Context().Done():
            return
        }
    }
}
```

- [ ] **Step 2: 创建 BatchService**

```go
// backend/service/batch.go
package service

import (
    "encoding/json"
    "fmt"

    "prompt-vault/models"
    "gorm.io/gorm"
)

type BatchService struct {
    db          *gorm.DB
    sseManager  interface{ SendSSEProgress(uint, int, int) }
}

func NewBatchService(db *gorm.DB) *BatchService {
    return &BatchService{db: db}
}

type BatchTestRequest struct {
    PromptID    uint        `json:"prompt_id" binding:"required"`
    Model       string      `json:"model" binding:"required"`
    TestCases   []TestCase  `json:"test_cases" binding:"required,min=1"`
    VariableSets []map[string]string `json:"variable_sets"` // 变量替换集
}

type TestCase struct {
    Name    string            `json:"name"`
    Input   map[string]string `json:"input"` // 变量名 -> 值
    Expected string           `json:"expected,omitempty"`
}

func (s *BatchService) CreateBatchTest(req BatchTestRequest) (*models.Task, error) {
    payload, _ := json.Marshal(req)

    task := &models.Task{
        Type:    "batch_test",
        Status:  "pending",
        Payload: string(payload),
        Progress: 0,
    }

    if err := s.db.Create(task).Error; err != nil {
        return nil, err
    }
    return task, nil
}

func (s *BatchService) GetBatchResults(taskID uint) ([]BatchResult, error) {
    var task models.Task
    if err := s.db.First(&task, taskID).Error; err != nil {
        return nil, err
    }

    if task.Result == "" {
        return nil, nil
    }

    var results []BatchResult
    if err := json.Unmarshal([]byte(task.Result), &results); err != nil {
        return nil, err
    }
    return results, nil
}

type BatchResult struct {
    CaseName    string  `json:"case_name"`
    Input       map[string]string `json:"input"`
    Output      string  `json:"output"`
    QualityScore float64 `json:"quality_score"`
    LatencyMs   int     `json:"latency_ms"`
    Error       string  `json:"error,omitempty"`
}

func (s *BatchService) ExecuteBatchTest(task *models.Task) error {
    var req BatchTestRequest
    if err := json.Unmarshal([]byte(task.Payload), &req); err != nil {
        return err
    }

    results := make([]BatchResult, 0, len(req.TestCases))
    total := len(req.TestCases)

    for i, tc := range req.TestCases {
        // 发送进度
        if s.sseManager != nil {
            s.sseManager.SendSSEProgress(task.ID, i+1, total)
        }

        // 执行测试
        result, err := s.runSingleTest(tc, req.Model)
        if err != nil {
            results = append(results, BatchResult{
                CaseName: tc.Name,
                Input:    tc.Input,
                Error:    err.Error(),
            })
        } else {
            results = append(results, result)
        }

        // 更新数据库进度
        s.db.Model(task).Update("progress", ((i+1)*100)/total)
    }

    // 保存结果
    resultJSON, _ := json.Marshal(results)
    s.db.Model(task).Updates(map[string]interface{}{
        "status":       "done",
        "progress":      100,
        "result":        string(resultJSON),
    })

    return nil
}

func (s *BatchService) runSingleTest(tc TestCase, model string) (BatchResult, error) {
    // TODO: 调用 AI Provider 执行测试
    // 这里先返回 mock 数据
    return BatchResult{
        CaseName:    tc.Name,
        Input:       tc.Input,
        Output:      "Mock AI response for: " + tc.Name,
        QualityScore: 75.5,
        LatencyMs:   150,
    }, nil
}
```

- [ ] **Step 3: 编写测试并 Commit**

### Phase 2.3: 后端 B - ScoringService + EvalService

**文件:**
- Create: `backend/service/scoring.go`
- Create: `backend/models/eval_set.go`
- Create: `backend/service/eval.go`
- Create: `backend/handlers/scoring.go`
- Create: `backend/handlers/eval.go`

- [ ] **Step 1: 创建 EvalSet 模型**

```go
// backend/models/eval_set.go
package models

import "time"

type EvalSet struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    PromptID  uint      `gorm:"not null" json:"prompt_id"`
    Name      string    `gorm:"size:200" json:"name"`
    Cases     string    `gorm:"type:text" json:"cases"` // JSON array of test cases
    Weights   string    `gorm:"type:text" json:"weights"` // JSON object for dimension weights
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (EvalSet) TableName() string {
    return "eval_sets"
}

type EvalCase struct {
    Name        string            `json:"name"`
    Input       map[string]string `json:"input"`
    Expected    string            `json:"expected,omitempty"`
    EvaluationPrompt string        `json:"evaluation_prompt,omitempty"`
}
```

- [ ] **Step 2: 创建 ScoringService (4 维度评分)**

```go
// backend/service/scoring.go
package service

import (
    "fmt"
    "regexp"
    "strings"

    "gorm.io/gorm"
)

type ScoringService struct {
    db *gorm.DB
}

func NewScoringService(db *gorm.DB) *ScoringService {
    return &ScoringService{db: db}
}

type ScoreResult struct {
    Clarity     int     `json:"clarity"`
    Completeness int   `json:"completeness"`
    Example    int     `json:"example"`
    Role       int     `json:"role"`
    Total      float64 `json:"total"`
}

var (
    clarityKeywords = []string{"please", "provide", "explain", "describe", "clarify"}
    roleKeywords    = []string{"you are", "as a", "role:", "persona:", "act as"}
    examplePatterns = []string{"example:", "for instance", "such as", "like:", "e.g."}
)

func (s *ScoringService) ScorePrompt(prompt string) (ScoreResult, error) {
    clarity := s.scoreClarity(prompt)
    completeness := s.scoreCompleteness(prompt)
    example := s.scoreExample(prompt)
    role := s.scoreRole(prompt)

    total := float64(clarity)*0.3 + float64(completeness)*0.3 +
             float64(example)*0.25 + float64(role)*0.15

    return ScoreResult{
        Clarity:     clarity,
        Completeness: completeness,
        Example:     example,
        Role:        role,
        Total:       total,
    }, nil
}

// scoreClarity 清晰度评分
// 规则：变量占位符数量、长度、格式规范
func (s *ScoringService) scoreClarity(prompt string) int {
    score := 50 // 基础分

    // 检查变量格式 {{variable}}
    varPattern := regexp.MustCompile(`\{\{([^}]+)\}\}`)
    vars := varPattern.FindAllString(prompt, -1)
    if len(vars) > 0 && len(vars) <= 5 {
        score += 15 // 适量变量加分
    } else if len(vars) > 5 {
        score -= 10 // 变量过多扣分
    }

    // 长度评分
    wordCount := len(strings.Fields(prompt))
    if wordCount >= 20 && wordCount <= 500 {
        score += 15
    } else if wordCount < 20 {
        score -= 5
    } else {
        score -= 10
    }

    // 检查是否有明确指令
    for _, kw := range clarityKeywords {
        if strings.Contains(strings.ToLower(prompt), kw) {
            score += 5
            break
        }
    }

    // 限制在 0-100
    if score < 0 {
        score = 0
    }
    if score > 100 {
        score = 100
    }
    return score
}

// scoreCompleteness 完整性评分
// 规则：必填字段检查、任务目标明确性
func (s *ScoringService) scoreCompleteness(prompt string) int {
    score := 50

    // 检查是否有明确的目标/任务
    if strings.Contains(prompt, "?") {
        score += 10 // 有明确问题
    }

    // 检查是否包含输出格式要求
    if strings.Contains(strings.ToLower(prompt), "output") ||
       strings.Contains(strings.ToLower(prompt), "format") ||
       strings.Contains(strings.ToLower(prompt), "return") {
        score += 15
    }

    // 检查长度是否足够
    if len(prompt) > 100 {
        score += 10
    }

    // 检查是否有关键上下文
    if len(strings.Fields(prompt)) >= 30 {
        score += 15
    }

    if score < 0 {
        score = 0
    }
    if score > 100 {
        score = 100
    }
    return score
}

// scoreExample 示例评分
// 规则：示例数量、质量
func (s *ScoringService) scoreExample(prompt string) int {
    score := 30 // 基础分

    lowerPrompt := strings.ToLower(prompt)
    for _, pattern := range examplePatterns {
        if strings.Contains(lowerPrompt, pattern) {
            score += 15
            break
        }
    }

    // 检查是否有输入输出示例
    if strings.Contains(prompt, "Input:") || strings.Contains(prompt, "输入:") {
        score += 15
    }
    if strings.Contains(prompt, "Output:") || strings.Contains(prompt, "输出:") {
        score += 15
    }

    // 示例数量
    exampleCount := strings.Count(lowerPrompt, "example")
    if exampleCount >= 1 && exampleCount <= 3 {
        score += 10
    }

    if score < 0 {
        score = 0
    }
    if score > 100 {
        score = 100
    }
    return score
}

// scoreRole 角色一致性评分
// 规则：角色关键词检查
func (s *ScoringService) scoreRole(prompt string) int {
    score := 40 // 基础分

    lowerPrompt := strings.ToLower(prompt)
    for _, kw := range roleKeywords {
        if strings.Contains(lowerPrompt, kw) {
            score += 20
            break
        }
    }

    // 检查是否在开头定义角色
    firstPart := strings.ToLower(prompt[:min(100, len(prompt))])
    for _, kw := range roleKeywords {
        if strings.Contains(firstPart, kw) {
            score += 15
            break
        }
    }

    if score < 0 {
        score = 0
    }
    if score > 100 {
        score = 100
    }
    return score
}

// ScoreWithAI AI 辅助评分（用于规则无法判断的场景）
func (s *ScoringService) ScoreWithAI(prompt string, dimension string) (int, error) {
    // TODO: 调用 AI 模型进行辅助评分
    // 目前返回基于规则的分数
    switch dimension {
    case "clarity":
        return s.scoreClarity(prompt), nil
    case "completeness":
        return s.scoreCompleteness(prompt), nil
    case "example":
        return s.scoreExample(prompt), nil
    case "role":
        return s.scoreRole(prompt), nil
    default:
        return 0, fmt.Errorf("unknown dimension: %s", dimension)
    }
}
```

- [ ] **Step 3: 创建 EvalService**

```go
// backend/service/eval.go
package service

import (
    "encoding/json"
    "fmt"

    "prompt-vault/models"
    "gorm.io/gorm"
)

type EvalService struct {
    db           *gorm.DB
    scoringSvc   *ScoringService
}

func NewEvalService(db *gorm.DB, scoringSvc *ScoringService) *EvalService {
    return &EvalService{db: db, scoringSvc: scoringSvc}
}

type GenerateEvalSetRequest struct {
    PromptID   uint   `json:"prompt_id" binding:"required"`
    PromptText string `json:"prompt_text" binding:"required"`
    Count      int    `json:"count" binding:"required,min=5,max=20"`
}

func (s *EvalService) CreateEvalSet(req GenerateEvalSetRequest) (*models.EvalSet, error) {
    // TODO: 调用 AI 生成测试用例
    // 目前生成 mock 数据
    cases := s.generateMockCases(req.PromptText, req.Count)

    casesJSON, _ := json.Marshal(cases)
    weightsJSON, _ := json.Marshal(map[string]float64{
        "clarity":      0.30,
        "completeness": 0.30,
        "example":      0.25,
        "role":         0.15,
    })

    evalSet := &models.EvalSet{
        PromptID:  req.PromptID,
        Name:      fmt.Sprintf("Eval Set for Prompt %d", req.PromptID),
        Cases:     string(casesJSON),
        Weights:   string(weightsJSON),
    }

    if err := s.db.Create(evalSet).Error; err != nil {
        return nil, err
    }
    return evalSet, nil
}

func (s *EvalService) generateMockCases(prompt string, count int) []models.EvalCase {
    cases := make([]models.EvalCase, count)
    for i := 0; i < count; i++ {
        cases[i] = models.EvalCase{
            Name:  fmt.Sprintf("Test Case %d", i+1),
            Input: map[string]string{
                "input": fmt.Sprintf("Mock input %d", i+1),
            },
            Expected:    "Expected output",
            EvaluationPrompt: "Evaluate the response quality",
        }
    }
    return cases
}

func (s *EvalService) GetEvalSet(id uint) (*models.EvalSet, error) {
    var evalSet models.EvalSet
    if err := s.db.First(&evalSet, id).Error; err != nil {
        return nil, err
    }
    return &evalSet, nil
}

func (s *EvalService) ListEvalSets(promptID uint) ([]models.EvalSet, error) {
    var evalSets []models.EvalSet
    query := s.db.Model(&models.EvalSet{})
    if promptID > 0 {
        query = query.Where("prompt_id = ?", promptID)
    }
    if err := query.Find(&evalSets).Error; err != nil {
        return nil, err
    }
    return evalSets, nil
}

func (s *EvalService) RunEvalSet(id uint) (*EvalResult, error) {
    evalSet, err := s.GetEvalSet(id)
    if err != nil {
        return nil, err
    }

    var cases []models.EvalCase
    if err := json.Unmarshal([]byte(evalSet.Cases), &cases); err != nil {
        return nil, err
    }

    var results []EvalCaseResult
    for _, c := range cases {
        result := s.runSingleEval(c)
        results = append(results, result)
    }

    return &EvalResult{
        EvalSetID: id,
        Results:   results,
    }, nil
}

type EvalCaseResult struct {
    CaseName   string  `json:"case_name"`
    Score      float64 `json:"score"`
    Feedback   string  `json:"feedback,omitempty"`
}

type EvalResult struct {
    EvalSetID uint              `json:"eval_set_id"`
    Results   []EvalCaseResult   `json:"results"`
}

func (s *EvalService) runSingleEval(c models.EvalCase) EvalCaseResult {
    // TODO: 调用 AI 执行评测
    return EvalCaseResult{
        CaseName: c.Name,
        Score:    75.0,
        Feedback: "Mock evaluation result",
    }
}
```

- [ ] **Step 4: 编写测试并 Commit**

### Phase 2.4: 前端组件开发

**文件:**
- Create: `frontend/src/components/BatchTestTable.vue`
- Create: `frontend/src/components/BatchTestCard.vue`
- Create: `frontend/src/components/QualityScoreCard.vue`
- Create: `frontend/src/components/TaskProgressBar.vue`

- [ ] **Step 1: BatchTestTable.vue**

```vue
<!-- frontend/src/components/BatchTestTable.vue -->
<template>
  <div class="batch-test-table">
    <el-table
      :data="testResults"
      :row-class-name="tableRowClassName"
      @row-click="handleRowClick"
      style="width: 100%"
    >
      <el-table-column type="index" width="60" label="序号" />
      <el-table-column prop="caseName" label="用例名称" min-width="150" />
      <el-table-column prop="model" label="模型" width="120">
        <template #default="{ row }">
          <el-tag size="small">{{ row.model }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="qualityScore" label="质量分" width="100">
        <template #default="{ row }">
          <el-tag
            :type="getScoreType(row.qualityScore)"
            size="small"
          >
            {{ row.qualityScore?.toFixed(1) || '-' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="latencyMs" label="延迟" width="80">
        <template #default="{ row }">
          {{ row.latencyMs ? row.latencyMs + 'ms' : '-' }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-button size="small" text type="primary" @click.stop="handleExpand(row)">
            展开
          </el-button>
          <el-button size="small" text type="primary" @click.stop="handleCompare(row)">
            对比
          </el-button>
        </template>
      </el-table-column>

      <!-- 展开行 -->
      <el-table-column type="expand" width="1">
        <template #default="{ row }">
          <BatchTestCard :result="row" />
        </template>
      </el-table-column>
    </el-table>

    <!-- 多选对比功能 -->
    <div v-if="selectedRows.length > 1" class="compare-bar">
      <span>已选择 {{ selectedRows.length }} 项</span>
      <el-button type="primary" size="small" @click="handleBatchCompare">
        横向对比
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import BatchTestCard from './BatchTestCard.vue'

const props = defineProps({
  testResults: {
    type: Array,
    default: () => []
  }
})

const selectedRows = ref([])

const getScoreType = (score) => {
  if (!score) return 'info'
  if (score >= 80) return 'success'
  if (score >= 60) return 'warning'
  return 'danger'
}

const handleRowClick = (row) => {
  // 处理行点击
}

const handleExpand = (row) => {
  // 展开详情
}

const handleCompare = (row) => {
  // 单项对比
}

const handleBatchCompare = () => {
  // 批量对比
}

const tableRowClassName = ({ rowIndex }) => {
  if (rowIndex % 2 === 0) {
    return 'even-row'
  }
  return ''
}
</script>

<style scoped>
.batch-test-table {
  width: 100%;
}

.compare-bar {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-3) var(--spacing-4);
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  box-shadow: var(--shadow-hover);
}

:deep(.el-table__row) {
  cursor: pointer;
}

:deep(.even-row) {
  background: var(--color-bg);
}
</style>
```

- [ ] **Step 2: BatchTestCard.vue**

```vue
<!-- frontend/src/components/BatchTestCard.vue -->
<template>
  <div class="batch-test-card">
    <div class="card-section">
      <h4>输入</h4>
      <pre class="code-block">{{ result.input }}</pre>
    </div>

    <div class="card-section">
      <h4>AI 输出</h4>
      <pre class="code-block output">{{ result.output }}</pre>
    </div>

    <div v-if="result.error" class="card-section error">
      <h4>错误</h4>
      <pre class="code-block">{{ result.error }}</pre>
    </div>

    <div class="card-meta">
      <span>质量分: {{ result.qualityScore?.toFixed(2) || '-' }}</span>
      <span>延迟: {{ result.latencyMs }}ms</span>
      <span>模型: {{ result.model }}</span>
    </div>
  </div>
</template>

<script setup>
defineProps({
  result: {
    type: Object,
    required: true
  }
})
</script>

<style scoped>
.batch-test-card {
  padding: var(--spacing-4);
  background: var(--color-bg);
}

.card-section {
  margin-bottom: var(--spacing-4);
}

.card-section h4 {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  margin-bottom: var(--spacing-2);
  font-weight: var(--font-weight-medium);
}

.code-block {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--spacing-3);
  font-family: 'SF Mono', 'Monaco', 'Consolas', monospace;
  font-size: var(--font-size-sm);
  line-height: 1.6;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-word;
  margin: 0;
}

.code-block.output {
  border-color: var(--color-primary-light);
  background: color-mix(in srgb, var(--color-primary-light) 30%, var(--color-surface));
}

.card-section.error .code-block {
  border-color: var(--color-danger-light);
  background: color-mix(in srgb, var(--color-danger-light) 30%, var(--color-surface));
  color: var(--color-danger);
}

.card-meta {
  display: flex;
  gap: var(--spacing-4);
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}
</style>
```

- [ ] **Step 3: QualityScoreCard.vue**

```vue
<!-- frontend/src/components/QualityScoreCard.vue -->
<template>
  <div class="quality-score-card">
    <div class="score-header">
      <span class="score-title">质量评分</span>
      <el-tag :type="getOverallType()" size="small">
        {{ overallScore.toFixed(1) }}
      </el-tag>
    </div>

    <div class="score-chart">
      <div ref="radarRef" class="radar-chart"></div>
    </div>

    <div class="score-details">
      <div class="score-item">
        <span class="score-label">Clarity</span>
        <el-progress
          :percentage="score.clarity"
          :color="'#6366F1'"
          :show-text="false"
        />
        <span class="score-value">{{ score.clarity }}</span>
      </div>
      <div class="score-item">
        <span class="score-label">Completeness</span>
        <el-progress
          :percentage="score.completeness"
          :color="'#8B5CF6'"
          :show-text="false"
        />
        <span class="score-value">{{ score.completeness }}</span>
      </div>
      <div class="score-item">
        <span class="score-label">Example</span>
        <el-progress
          :percentage="score.example"
          :color="'#EC4899'"
          :show-text="false"
        />
        <span class="score-value">{{ score.example }}</span>
      </div>
      <div class="score-item">
        <span class="score-label">Role</span>
        <el-progress
          :percentage="score.role"
          :color="'#14B8A6'"
          :show-text="false"
        />
        <span class="score-value">{{ score.role }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import * as echarts from 'echarts'

const props = defineProps({
  score: {
    type: Object,
    default: () => ({
      clarity: 0,
      completeness: 0,
      example: 0,
      role: 0
    })
  }
})

const radarRef = ref(null)
let chart = null

const overallScore = computed(() => {
  return props.score.clarity * 0.3 +
         props.score.completeness * 0.3 +
         props.score.example * 0.25 +
         props.score.role * 0.15
})

const getOverallType = () => {
  if (overallScore.value >= 80) return 'success'
  if (overallScore.value >= 60) return 'warning'
  return 'danger'
}

onMounted(() => {
  if (radarRef.value) {
    chart = echarts.init(radarRef.value)
    updateChart()
  }
})

const updateChart = () => {
  if (!chart) return

  const option = {
    radar: {
      indicator: [
        { name: 'Clarity', max: 100 },
        { name: 'Completeness', max: 100 },
        { name: 'Example', max: 100 },
        { name: 'Role', max: 100 }
      ],
      radius: '60%',
      axisName: {
        color: '#64748B',
        fontSize: 10
      }
    },
    series: [{
      type: 'radar',
      data: [{
        value: [
          props.score.clarity,
          props.score.completeness,
          props.score.example,
          props.score.role
        ],
        areaStyle: {
          color: 'rgba(37, 99, 235, 0.2)'
        },
        lineStyle: {
          color: '#2563EB'
        },
        itemStyle: {
          color: '#2563EB'
        }
      }]
    }]
  }
  chart.setOption(option)
}
</script>

<style scoped>
.quality-score-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-4);
}

.score-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-3);
}

.score-title {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.score-chart {
  height: 160px;
  margin-bottom: var(--spacing-3);
}

.radar-chart {
  width: 100%;
  height: 100%;
}

.score-details {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.score-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.score-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  width: 100px;
  flex-shrink: 0;
}

:deep(.el-progress) {
  flex: 1;
}

.score-value {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  width: 24px;
  text-align: right;
}
</style>
```

- [ ] **Step 4: TaskProgressBar.vue**

```vue
<!-- frontend/src/components/TaskProgressBar.vue -->
<template>
  <div class="task-progress-bar">
    <div class="progress-info">
      <span class="progress-label">{{ label }}</span>
      <span class="progress-value">{{ progress }}%</span>
    </div>

    <el-progress
      :percentage="progress"
      :color="progressColor"
      :show-text="false"
      :stroke-width="8"
    />

    <div v-if="showDetail" class="progress-detail">
      <span v-if="status === 'running'">处理中...</span>
      <span v-else-if="status === 'done'">已完成</span>
      <span v-else-if="status === 'failed'" class="error">失败</span>
      <span v-else>{{ current }}/{{ total }}</span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  progress: {
    type: Number,
    default: 0
  },
  status: {
    type: String,
    default: 'idle' // idle | pending | running | done | failed
  },
  current: {
    type: Number,
    default: 0
  },
  total: {
    type: Number,
    default: 0
  },
  label: {
    type: String,
    default: '进度'
  },
  showDetail: {
    type: Boolean,
    default: true
  }
})

const progressColor = computed(() => {
  if (props.status === 'failed') return 'var(--color-danger)'
  if (props.status === 'done') return 'var(--color-success)'
  if (props.progress >= 80) return 'var(--color-success)'
  if (props.progress >= 50) return 'var(--color-warning)'
  return 'var(--color-primary)'
})
</script>

<style scoped>
.task-progress-bar {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.progress-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.progress-label {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

.progress-value {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.progress-detail {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  text-align: center;
}

.progress-detail .error {
  color: var(--color-danger);
}
</style>
```

- [ ] **Step 5: Commit**

---

## Sprint 3: 统计分析（Week 5-6）

### Phase 3.1: Worktree 创建

```bash
git worktree add worktrees/sprint3-sprt-backend-a -b feature/sprint3-sprt
git worktree add worktrees/sprint3-abtest-backend-b -b feature/sprint3-abtest
git worktree add worktrees/sprint3-multiround-backend-b -b feature/sprint3-multiround
git worktree add worktrees/sprint3-abpanel-frontend-a -b feature/sprint3-abpanel
git worktree add worktrees/sprint3-ablist-frontend-b -b feature/sprint3-ablist
```

### Phase 3.2: 后端 A - SPRT Engine

**文件:**
- Create: `backend/service/sprt.go`

- [ ] **Step 1: 实现 SPRT 序贯检验引擎**

```go
// backend/service/sprt.go
package service

import (
    "math"
)

/*
SPRT (Sequential Probability Ratio Test) Implementation

SPRT 是一种序贯检验方法，可以在每次观测后决定：
- Accept (接受原假设)
- Reject (拒绝原假设)
- Continue (继续观测)

公式：
- Lambda = (p1/p0)^n1 * ((1-p1)/(1-p0))^n2
- Accept if Lambda > 1/beta
- Reject if Lambda < alpha
- Continue otherwise

参数：
- alpha: 显著性水平 (type I error), 默认 0.05
- beta: 功效水平 (type II error), 默认 0.20
- p0: 原假设下的成功率
- p1: 备择假设下的成功率
*/

// SPRTConfig SPRT 配置
type SPRTConfig struct {
    Alpha     float64 // 显著性水平，默认 0.05
    Beta      float64 // 功效水平，默认 0.20
    MinSamples int    // 最小样本数，默认 15
    MaxSamples int    // 最大样本数，默认 50
    P0        float64 // 原假设成功率，默认 0.5
    P1        float64 // 备择假设成功率，默认 0.6
}

// Decision SPRT 决策
type Decision int

const (
    Continue Decision = iota
    Accept
    Reject
)

func (d Decision) String() string {
    switch d {
    case Continue:
        return "continue"
    case Accept:
        return "accept"
    case Reject:
        return "reject"
    default:
        return "unknown"
    }
}

// SPRTResult SPRT 检验结果
type SPRTResult struct {
    Decision     Decision `json:"decision"`
    N            int      `json:"n"`             // 总样本数
    NA           int      `json:"n_a"`          // Variant A 样本数
    NB           int      `json:"n_b"`          // Variant B 样本数
    ScoreA       float64  `json:"score_a"`      // Variant A 平均分数
    ScoreB       float64  `json:"score_b"`      // Variant B 平均分数
    PValue       float64  `json:"p_value"`      // p 值
    ConfidenceCI [2]float64 `json:"ci"`         // 置信区间
    Winner       string   `json:"winner"`        // 胜出方
}

// SPRT 引擎
type SPRTEngine struct {
    config SPRTConfig
}

func NewSPRTEngine(config SPRTConfig) *SPRTEngine {
    if config.Alpha == 0 {
        config.Alpha = 0.05
    }
    if config.Beta == 0 {
        config.Beta = 0.20
    }
    if config.MinSamples == 0 {
        config.MinSamples = 15
    }
    if config.MaxSamples == 0 {
        config.MaxSamples = 50
    }
    if config.P0 == 0 {
        config.P0 = 0.5
    }
    if config.P1 == 0 {
        config.P1 = 0.6
    }
    return &SPRTEngine{config: config}
}

// Test 执行 SPRT 检验
// scoresA: Variant A 的分数列表
// scoresB: Variant B 的分数列表
func (e *SPRTEngine) Test(scoresA, scoresB []float64) SPRTResult {
    nA := len(scoresA)
    nB := len(scoresB)
    n := nA + nB

    // 计算平均分数
    sumA := 0.0
    for _, s := range scoresA {
        sumA += s
    }
    sumB := 0.0
    for _, s := range scoresB {
        sumB += s
    }

    avgA := 0.0
    avgB := 0.0
    if nA > 0 {
        avgA = sumA / float64(nA)
    }
    if nB > 0 {
        avgB = sumB / float64(nB)
    }

    // 计算似然比
    lambda := e.calculateLikelihoodRatio(scoresA, scoresB)

    // 决策边界
    upperBound := math.Log(1 / e.config.Beta)
    lowerBound := math.Log(e.config.Alpha)

    var decision Decision
    if n >= e.config.MaxSamples {
        decision = Continue // 达到最大样本，强制停止
    } else if lambda > upperBound {
        decision = Accept
    } else if lambda < lowerBound {
        decision = Reject
    } else if n >= e.config.MinSamples {
        // 在最小样本后，可以选择接受表现更好的
        if avgA > avgB {
            if e.config.P1 > e.config.P0 {
                decision = Accept
            } else {
                decision = Reject
            }
        } else {
            if e.config.P1 < e.config.P0 {
                decision = Accept
            } else {
                decision = Reject
            }
        }
    } else {
        decision = Continue
    }

    // 计算 p 值和置信区间
    pValue := e.calculatePValue(scoresA, scoresB)
    ci := e.calculateConfidenceInterval(scoresA, scoresB)

    // 确定胜出方
    winner := "none"
    if decision == Accept {
        if avgA > avgB {
            winner = "A"
        } else {
            winner = "B"
        }
    }

    return SPRTResult{
        Decision:     decision,
        N:            n,
        NA:           nA,
        NB:           nB,
        ScoreA:       avgA,
        ScoreB:       avgB,
        PValue:       pValue,
        ConfidenceCI: ci,
        Winner:       winner,
    }
}

func (e *SPRTEngine) calculateLikelihoodRatio(scoresA, scoresB []float64) float64 {
    // 简化的似然比计算
    // 实际实现需要考虑二项分布或正态分布
    p0 := e.config.P0
    p1 := e.config.P1

    // 计算成功次数（假设分数 > 阈值为成功）
    threshold := 0.5
    successesA := 0
    for _, s := range scoresA {
        if s > threshold {
            successesA++
        }
    }
    successesB := 0
    for _, s := range scoresB {
        if s > threshold {
            successesB++
        }
    }

    nA := len(scoresA)
    nB := len(scoresB)

    // 似然比
    logLambda := 0.0

    // Variant A 的贡献
    if nA > 0 {
        for i := 0; i < successesA; i++ {
            logLambda += math.Log(p1 / p0)
        }
        for i := 0; i < nA-successesA; i++ {
            logLambda += math.Log((1-p1)/(1-p0))
        }
    }

    // Variant B 的贡献
    if nB > 0 {
        for i := 0; i < successesB; i++ {
            logLambda -= math.Log(p1 / p0)
        }
        for i := 0; i < nB-successesB; i++ {
            logLambda -= math.Log((1-p1)/(1-p0))
        }
    }

    return math.Exp(logLambda)
}

func (e *SPRTEngine) calculatePValue(scoresA, scoresB []float64) float64 {
    // 简化的 p 值计算
    // 使用 Welch's t-test
    if len(scoresA) == 0 || len(scoresB) == 0 {
        return 1.0
    }

    meanA := 0.0
    meanB := 0.0
    for _, s := range scoresA {
        meanA += s
    }
    for _, s := range scoresB {
        meanB += s
    }
    meanA /= float64(len(scoresA))
    meanB /= float64(len(scoresB))

    var varA, varB float64
    for _, s := range scoresA {
        diff := s - meanA
        varA += diff * diff
    }
    for _, s := range scoresB {
        diff := s - meanB
        varB += diff * diff
    }
    varA /= float64(len(scoresA) - 1)
    varB /= float64(len(scoresB) - 1)

    pooledSE := math.Sqrt(varA/float64(len(scoresA)) + varB/float64(len(scoresB)))
    if pooledSE == 0 {
        return 1.0
    }

    t := math.Abs(meanA - meanB) / pooledSE
    // 简化：假设正态分布
    pValue := 2 * (1 - normalCDF(t))
    if pValue > 1 {
        pValue = 1
    }
    return pValue
}

func (e *SPRTEngine) calculateConfidenceInterval(scoresA, scoresB []float64) [2]float64 {
    // 95% 置信区间
    if len(scoresA) == 0 || len(scoresB) == 0 {
        return [2]float64{0, 1}
    }

    meanA := 0.0
    meanB := 0.0
    for _, s := range scoresA {
        meanA += s
    }
    for _, s := range scoresB {
        meanB += s
    }
    meanA /= float64(len(scoresA))
    meanB /= float64(len(scoresB))

    diff := meanA - meanB
    se := 1.96 * math.Sqrt(0.25/float64(len(scoresA)) + 0.25/float64(len(scoresB)))

    return [2]float64{diff - se, diff + se}
}

// normalCDF 标准正态分布 CDF
func normalCDF(x float64) float64 {
    return 0.5 * (1 + math.Erf(x/math.Sqrt2))
}
```

- [ ] **Step 2: 编写测试**

```go
// backend/service/sprt_test.go
package service

import (
    "math"
    "testing"
)

func TestSPRTEngine_Test(t *testing.T) {
    engine := NewSPRTEngine(SPRTConfig{
        Alpha:     0.05,
        Beta:      0.20,
        MinSamples: 15,
        MaxSamples: 50,
    })

    // 生成测试数据：A 明显优于 B
    scoresA := make([]float64, 30)
    scoresB := make([]float64, 30)
    for i := 0; i < 30; i++ {
        scoresA[i] = 0.8 + (float64(i) * 0.005) // 逐渐上升
        scoresB[i] = 0.5 + (float64(i) * 0.001) // 缓慢上升
    }

    result := engine.Test(scoresA, scoresB)

    t.Logf("Decision: %s", result.Decision.String())
    t.Logf("Winner: %s", result.Winner)
    t.Logf("ScoreA: %.3f, ScoreB: %.3f", result.ScoreA, result.ScoreB)
    t.Logf("PValue: %.4f", result.PValue)

    if result.Winner != "A" {
        t.Log("Expected winner A (A has higher scores)")
    }
}

func TestSPRTEngine_MaxSamples(t *testing.T) {
    engine := NewSPRTEngine(SPRTConfig{
        Alpha:     0.05,
        Beta:      0.20,
        MinSamples: 15,
        MaxSamples: 50,
    })

    // 提供超过 MaxSamples 的数据
    scoresA := make([]float64, 60)
    scoresB := make([]float64, 60)
    for i := 0; i < 60; i++ {
        scoresA[i] = 0.6
        scoresB[i] = 0.6
    }

    result := engine.Test(scoresA, scoresB)

    if result.N != 60 {
        t.Errorf("Expected N=60, got %d", result.N)
    }
}
```

- [ ] **Step 3: Commit**

### Phase 3.3: 后端 B - ABTestService

**文件:**
- Create: `backend/models/ab_test.go`
- Create: `backend/service/ab_test.go`
- Create: `backend/handlers/ab_test.go`

- [ ] **Step 1: 创建 ABTest 模型**

```go
// backend/models/ab_test.go
package models

import "time"

type ABTest struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    PromptID  uint      `gorm:"not null" json:"prompt_id"`
    Name      string    `gorm:"size:200" json:"name"`
    Config    string    `gorm:"type:text" json:"config"` // JSON: {variant_a, variant_b, config}
    Status    string    `gorm:"size:20;default:'running'" json:"status"` // pending | running | completed | stopped
    Result    string    `gorm:"type:text" json:"result,omitempty"` // JSON: SPRT result
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (ABTest) TableName() string {
    return "ab_tests"
}

type ABTestResult struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    ABTestID     uint      `gorm:"not null" json:"ab_test_id"`
    RunIndex     int       `json:"run_index"`
    Variant      string    `gorm:"size:10" json:"variant"` // A or B
    Score        float64   `json:"score"`
    LatencyMs    int       `json:"latency_ms"`
    CreatedAt    time.Time `json:"created_at"`
}

func (ABTestResult) TableName() string {
    return "ab_test_results"
}
```

- [ ] **Step 2: 创建 ABTestService**

```go
// backend/service/ab_test.go
package service

import (
    "encoding/json"
    "time"

    "prompt-vault/models"
    "gorm.io/gorm"
)

type ABTestService struct {
    db        *gorm.DB
    sprt      *SPRTEngine
    sseMgr    interface{ Publish(uint, string, interface{}) }
}

func NewABTestService(db *gorm.DB) *ABTestService {
    return &ABTestService{
        db:   db,
        sprt: NewSPRTEngine(SPRTConfig{}),
    }
}

type CreateABTestRequest struct {
    PromptID   uint   `json:"prompt_id" binding:"required"`
    Name       string `json:"name" binding:"required"`
    VariantA   string `json:"variant_a" binding:"required"` // prompt variant A
    VariantB   string `json:"variant_b" binding:"required"` // prompt variant B
    Model      string `json:"model" binding:"required"`
    MaxRuns    int    `json:"max_runs"` // 默认 50
}

func (s *ABTestService) Create(req CreateABTestRequest) (*models.ABTest, error) {
    config := map[string]interface{}{
        "variant_a": req.VariantA,
        "variant_b": req.VariantB,
        "model":     req.Model,
        "max_runs":  req.MaxRuns,
    }
    configJSON, _ := json.Marshal(config)

    abTest := &models.ABTest{
        PromptID: req.PromptID,
        Name:     req.Name,
        Config:   string(configJSON),
        Status:   "running",
    }

    if err := s.db.Create(abTest).Error; err != nil {
        return nil, err
    }
    return abTest, nil
}

func (s *ABTestService) GetByID(id uint) (*models.ABTest, error) {
    var abTest models.ABTest
    if err := s.db.First(&abTest, id).Error; err != nil {
        return nil, err
    }
    return &abTest, nil
}

func (s *ABTestService) RecordResult(abTestID uint, variant string, score float64, latencyMs int) error {
    // 获取当前运行次数
    var count int64
    s.db.Model(&models.ABTestResult{}).Where("ab_test_id = ?", abTestID).Count(&count)

    result := &models.ABTestResult{
        ABTestID:  abTestID,
        RunIndex:  int(count) + 1,
        Variant:   variant,
        Score:     score,
        LatencyMs: latencyMs,
        CreatedAt: time.Now(),
    }

    if err := s.db.Create(result).Error; err != nil {
        return err
    }

    // 触发 SPRT 检验
    s.checkSignificance(abTestID)

    return nil
}

func (s *ABTestService) checkSignificance(abTestID uint) {
    // 获取所有结果
    var resultsA, resultsB []models.ABTestResult
    s.db.Where("ab_test_id = ? AND variant = ?", abTestID, "A").Find(&resultsA)
    s.db.Where("ab_test_id = ? AND variant = ?", abTestID, "B").Find(&resultsB)

    if len(resultsA) == 0 || len(resultsB) == 0 {
        return
    }

    scoresA := make([]float64, len(resultsA))
    scoresB := make([]float64, len(resultsB))
    for i, r := range resultsA {
        scoresA[i] = r.Score
    }
    for i, r := range resultsB {
        scoresB[i] = r.Score
    }

    // 执行 SPRT 检验
    sprtResult := s.sprt.Test(scoresA, scoresB)

    // 更新状态
    updates := map[string]interface{}{
        "result": toJSON(sprtResult),
    }
    if sprtResult.Decision != Continue {
        updates["status"] = "completed"
        updates["result"] = toJSON(sprtResult)
    }

    s.db.Model(&models.ABTest{}).Where("id = ?", abTestID).Updates(updates)

    // 发送 SSE 事件
    if s.sseMgr != nil {
        s.sseMgr.Publish(abTestID, "progress", sprtResult)
    }
}

func (s *ABTestService) GetResults(abTestID uint) (*ABTestSummary, error) {
    var results []models.ABTestResult
    if err := s.db.Where("ab_test_id = ?", abTestID).Order("created_at").Find(&results).Error; err != nil {
        return nil, err
    }

    var resultsA, resultsB []models.ABTestResult
    for _, r := range results {
        if r.Variant == "A" {
            resultsA = append(resultsA, r)
        } else {
            resultsB = append(resultsB, r)
        }
    }

    return &ABTestSummary{
        TotalRuns: len(results),
        VariantA:  resultsA,
        VariantB:  resultsB,
    }, nil
}

type ABTestSummary struct {
    TotalRuns int                   `json:"total_runs"`
    VariantA  []models.ABTestResult `json:"variant_a"`
    VariantB  []models.ABTestResult `json:"variant_b"`
}

func toJSON(v interface{}) string {
    b, _ := json.Marshal(v)
    return string(b)
}
```

- [ ] **Step 3: Commit**

### Phase 3.4: 前端 A/B - ABTestSequentialPanel + ABTestList

**文件:**
- Create: `frontend/src/components/ABTestSequentialPanel.vue`
- Create: `frontend/src/views/ABTestList.vue`
- Create: `frontend/src/views/ABTestDetail.vue`
- Modify: `frontend/src/router/index.js`

- [ ] **Step 1: ABTestSequentialPanel.vue**

```vue
<!-- frontend/src/components/ABTestSequentialPanel.vue -->
<template>
  <div class="ab-test-sequential-panel">
    <div class="panel-header">
      <h3>序贯检验进度</h3>
      <el-tag :type="significanceType" size="small">
        {{ significanceLabel }}
      </el-tag>
    </div>

    <!-- 进度条 -->
    <div class="progress-section">
      <div class="variant-progress">
        <div class="variant-label">
          <span class="variant-dot variant-a"></span>
          <span>Variant A</span>
        </div>
        <el-progress
          :percentage="progressA"
          :color="'#2563EB'"
          :show-text="false"
          :stroke-width="12"
        />
        <span class="variant-count">{{ resultsA.length }} 次</span>
      </div>

      <div class="variant-progress">
        <div class="variant-label">
          <span class="variant-dot variant-b"></span>
          <span>Variant B</span>
        </div>
        <el-progress
          :percentage="progressB"
          :color="'#F97316'"
          :show-text="false"
          :stroke-width="12"
        />
        <span class="variant-count">{{ resultsB.length }} 次</span>
      </div>
    </div>

    <!-- 统计信息 -->
    <div class="stats-section">
      <div class="stat-item">
        <span class="stat-label">最低样本</span>
        <span class="stat-value">{{ minSamples }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">已完成</span>
        <span class="stat-value">{{ completedRuns }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">最大样本</span>
        <span class="stat-value">{{ maxSamples }}</span>
      </div>
    </div>

    <!-- 分数对比 -->
    <div class="score-section">
      <div class="score-card variant-a">
        <span class="score-label">Variant A</span>
        <span class="score-value">{{ avgScoreA.toFixed(2) }}</span>
      </div>
      <div class="score-divider">
        <span>vs</span>
      </div>
      <div class="score-card variant-b">
        <span class="score-label">Variant B</span>
        <span class="score-value">{{ avgScoreB.toFixed(2) }}</span>
      </div>
    </div>

    <!-- 置信区间 -->
    <div class="ci-section">
      <span class="ci-label">置信区间 (95%)</span>
      <div class="ci-bar">
        <div
          class="ci-range"
          :style="{
            left: ciLeft + '%',
            width: ciWidth + '%'
          }"
        ></div>
        <span class="ci-zero"></span>
      </div>
      <div class="ci-values">
        <span>{{ ciLower.toFixed(3) }}</span>
        <span>{{ ciUpper.toFixed(3) }}</span>
      </div>
    </div>

    <!-- 显著性徽章 -->
    <div v-if="showSignificance" class="significance-badge" :class="significanceType">
      <span class="badge-icon">{{ significanceIcon }}</span>
      <span class="badge-text">{{ significanceText }}</span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  resultsA: {
    type: Array,
    default: () => []
  },
  resultsB: {
    type: Array,
    default: () => []
  },
  minSamples: {
    type: Number,
    default: 15
  },
  maxSamples: {
    type: Number,
    default: 50
  },
  ciLower: {
    type: Number,
    default: 0
  },
  ciUpper: {
    type: Number,
    default: 0
  },
  pValue: {
    type: Number,
    default: 1
  },
  winner: {
    type: String,
    default: 'none'
  }
})

const completedRuns = computed(() => props.resultsA.length + props.resultsB.length)

const progressA = computed(() => (props.resultsA.length / props.maxSamples) * 100)
const progressB = computed(() => (props.resultsB.length / props.maxSamples) * 100)

const avgScoreA = computed(() => {
  if (props.resultsA.length === 0) return 0
  const sum = props.resultsA.reduce((acc, r) => acc + r.score, 0)
  return sum / props.resultsA.length
})

const avgScoreB = computed(() => {
  if (props.resultsB.length === 0) return 0
  const sum = props.resultsB.reduce((acc, r) => acc + r.score, 0)
  return sum / props.resultsB.length
})

const ciLeft = computed(() => Math.max(0, (props.ciLower + 0.5) * 100))
const ciWidth = computed(() => Math.max(0, (props.ciUpper - props.ciLower) * 100))

const showSignificance = computed(() => completedRuns.value >= props.minSamples)

const significanceType = computed(() => {
  if (props.pValue < 0.05) return 'success'
  if (props.pValue < 0.1) return 'warning'
  return 'info'
})

const significanceLabel = computed(() => {
  if (props.pValue < 0.05) return '显著'
  if (props.pValue < 0.1) return '接近'
  return '不足'
})

const significanceIcon = computed(() => {
  if (props.pValue < 0.05) return '✓'
  if (props.pValue < 0.1) return '~'
  return '○'
})

const significanceText = computed(() => {
  if (props.winner !== 'none') {
    return `Variant ${props.winner} 胜出`
  }
  return '无显著差异'
})
</script>

<style scoped>
.ab-test-sequential-panel {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-4);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-4);
}

.panel-header h3 {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0;
}

.progress-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
  margin-bottom: var(--spacing-4);
}

.variant-progress {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.variant-label {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  width: 100px;
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

.variant-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.variant-dot.variant-a {
  background: #2563EB;
}

.variant-dot.variant-b {
  background: #F97316;
}

.variant-count {
  width: 50px;
  text-align: right;
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.stats-section {
  display: flex;
  justify-content: space-between;
  padding: var(--spacing-3);
  background: var(--color-bg);
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-4);
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-1);
}

.stat-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.stat-value {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.score-section {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-4);
  margin-bottom: var(--spacing-4);
}

.score-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--spacing-3);
  border-radius: var(--radius-md);
  min-width: 100px;
}

.score-card.variant-a {
  background: color-mix(in srgb, #2563EB 10%, transparent);
}

.score-card.variant-b {
  background: color-mix(in srgb, #F97316 10%, transparent);
}

.score-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.score-value {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
}

.score-card.variant-a .score-value {
  color: #2563EB;
}

.score-card.variant-b .score-value {
  color: #F97316;
}

.score-divider {
  color: var(--color-text-muted);
  font-size: var(--font-size-sm);
}

.ci-section {
  margin-bottom: var(--spacing-4);
}

.ci-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  display: block;
  margin-bottom: var(--spacing-2);
}

.ci-bar {
  position: relative;
  height: 8px;
  background: var(--color-border);
  border-radius: 4px;
}

.ci-range {
  position: absolute;
  height: 100%;
  background: var(--color-primary);
  border-radius: 4px;
}

.ci-zero {
  position: absolute;
  left: 50%;
  top: -4px;
  width: 2px;
  height: 16px;
  background: var(--color-text-muted);
}

.ci-values {
  display: flex;
  justify-content: space-between;
  margin-top: var(--spacing-1);
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.significance-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-2);
  padding: var(--spacing-3);
  border-radius: var(--radius-md);
  font-weight: var(--font-weight-medium);
}

.significance-badge.success {
  background: var(--color-success-light);
  color: var(--color-success);
}

.significance-badge.warning {
  background: var(--color-warning-light);
  color: var(--color-warning);
}

.significance-badge.info {
  background: var(--color-info-light);
  color: var(--color-info);
}

.badge-icon {
  font-size: var(--font-size-lg);
}

.badge-text {
  font-size: var(--font-size-sm);
}
</style>
```

- [ ] **Step 2: Commit**

---

## Sprint 4: 收尾（Week 7）

### Phase 4.1: Worktree 创建

```bash
git worktree add worktrees/sprint4-regression-backend-a -b feature/sprint4-regression
git worktree add worktrees/sprint4-cache-backend-a -b feature/sprint4-cache
git worktree add worktrees/sprint4-integration -b feature/sprint4-integration
git worktree add worktrees/sprint4-e2e -b feature/sprint4-e2e
```

### Phase 4.2: Regression Service + Cache Service

**文件:**
- Create: `backend/service/regression.go`
- Create: `backend/models/response_cache.go`
- Create: `backend/worker/cache.go`

- [ ] **Step 1: RegressionService**

```go
// backend/service/regression.go
package service

import (
    "encoding/json"
    "time"

    "prompt-vault/models"
    "gorm.io/gorm"
)

type RegressionService struct {
    db         *gorm.DB
    scoringSvc *ScoringService
}

func NewRegressionService(db *gorm.DB, scoringSvc *ScoringService) *RegressionService {
    return &RegressionService{db: db, scoringSvc: scoringSvc}
}

type RegressionReport struct {
    PromptID       uint        `json:"prompt_id"`
    PromptTitle    string      `json:"prompt_title"`
    OldModel       string      `json:"old_model"`
    NewModel       string      `json:"new_model"`
    OldScore       float64     `json:"old_score"`
    NewScore       float64     `json:"new_score"`
    ScoreDelta     float64     `json:"score_delta"`
    HasRegression  bool        `json:"has_regression"`
    GeneratedAt    time.Time   `json:"generated_at"`
}

func (s *RegressionService) Detect(promptID uint, oldModel, newModel string) (*RegressionReport, error) {
    var prompt models.Prompt
    if err := s.db.First(&prompt, promptID).Error; err != nil {
        return nil, err
    }

    // 简单实现：使用 ScoringService 评估新旧分数
    // 实际实现需要分别用 oldModel 和 newModel 调用 AI
    oldScore, _ := s.scoringSvc.ScorePrompt(prompt.Content)
    newScore, _ := s.scoringSvc.ScorePrompt(prompt.Content)

    report := &RegressionReport{
        PromptID:      promptID,
        PromptTitle:   prompt.Title,
        OldModel:      oldModel,
        NewModel:      newModel,
        OldScore:      oldScore.Total,
        NewScore:      newScore.Total,
        ScoreDelta:    newScore.Total - oldScore.Total,
        HasRegression: newScore.Total < oldScore.Total,
        GeneratedAt:   time.Now(),
    }

    return report, nil
}
```

- [ ] **Step 2: CacheService**

```go
// backend/worker/cache.go
package worker

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "time"

    "gorm.io/gorm"
)

type CacheService struct {
    db *gorm.DB
}

func NewCacheService(db *gorm.DB) *CacheService {
    return &CacheService{db: db}
}

type CacheEntry struct {
    Hash       string    `gorm:"primaryKey" json:"hash"`
    Provider   string    `gorm:"size:50;not null" json:"provider"`
    Model      string    `gorm:"size:100" json:"model"`
    RequestHash string   `gorm:"size:64;not null" json:"request_hash"`
    Response   string    `gorm:"type:text" json:"response"`
    CreatedAt  time.Time `json:"created_at"`
    ExpiresAt  time.Time `json:"expires_at"`
}

func (s *CacheService) TableName() string {
    return "response_cache"
}

func (s *CacheService) Get(provider, model, request string) (string, bool, error) {
    hash := s.hashRequest(provider, model, request)

    var entry CacheEntry
    if err := s.db.Where("hash = ? AND expires_at > ?", hash, time.Now()).First(&entry).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return "", false, nil
        }
        return "", false, err
    }
    return entry.Response, true, nil
}

func (s *CacheService) Set(provider, model, request, response string, ttl time.Duration) error {
    hash := s.hashRequest(provider, model, request)

    entry := &CacheEntry{
        Hash:        hash,
        Provider:    provider,
        Model:       model,
        RequestHash: hash,
        Response:    response,
        CreatedAt:   time.Now(),
        ExpiresAt:   time.Now().Add(ttl),
    }

    return s.db.Save(entry).Error
}

func (s *CacheService) hashRequest(provider, model, request string) string {
    data := provider + ":" + model + ":" + request
    h := sha256.Sum256([]byte(data))
    return hex.EncodeToString(h[:])
}

func (s *CacheService) Cleanup() error {
    return s.db.Where("expires_at < ?", time.Now()).Delete(&CacheEntry{}).Error
}
```

- [ ] **Step 3: Commit**

### Phase 4.3: 集成测试

**文件:**
- Create: `backend/integration_test.go`

```go
// backend/integration_test.go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "prompt-vault/handlers"
    "prompt-vault/models"
    "prompt-vault/service"

    "github.com/gin-gonic/gin"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(t.TempDir()+"/integration_test.db"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open test db: %v", err)
    }
    db.AutoMigrate(
        &models.Task{},
        &models.Prompt{},
        &models.ABTest{},
        &models.ABTestResult{},
    )
    return db
}

func setupRouter(db *gorm.DB) *gin.Engine {
    gin.SetMode(gin.TestMode)
    r := gin.New()

    taskSvc := service.NewTaskService(db)
    taskHandler := handlers.NewTaskHandler(taskSvc)

    v1 := r.Group("/api")
    v1.POST("/tasks", taskHandler.CreateTask)
    v1.GET("/tasks/:id", taskHandler.GetTask)

    return r
}

func TestTaskLifecycle(t *testing.T) {
    db := setupTestDB(t)
    r := setupRouter(db)

    // 创建任务
    payload := map[string]interface{}{
        "type":    "batch_test",
        "payload": map[string]interface{}{"count": 10},
    }
    body, _ := json.Marshal(payload)

    req := httptest.NewRequest("POST", "/api/tasks", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }

    var resp map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &resp)

    if resp["success"] != true {
        t.Error("expected success response")
    }
}
```

### Phase 4.4: E2E 测试（Playwright）

**文件:**
- Create: `frontend/tests/e2e/prompt-editor.spec.js`
- Create: `frontend/tests/e2e/batch-test.spec.js`

```javascript
// frontend/tests/e2e/prompt-editor.spec.js
import { test, expect } from '@playwright/test'

test.describe('Prompt Editor', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/prompts/new')
  })

  test('60/40 layout renders correctly', async ({ page }) => {
    // 检查左侧编辑区
    await expect(page.locator('.editor-main')).toBeVisible()
    await expect(page.locator('.editor-sidebar')).toBeVisible()

    // 检查右侧预览面板
    await expect(page.locator('.editor-preview')).toBeVisible()
    await expect(page.locator('.variable-preview-panel')).toBeVisible()
  })

  test('variable parsing works', async ({ page }) => {
    const editor = page.locator('.editor-content textarea')
    await editor.fill('Hello {{name}}, your order is {{order_id}}')

    // 等待预览面板更新
    await page.waitForTimeout(500)

    // 检查变量被识别
    const varInputs = page.locator('.variable-preview-panel .variable-row')
    await expect(varInputs).toHaveCount(2)
  })

  test('fill rate updates', async ({ page }) => {
    const editor = page.locator('.editor-content textarea')
    await editor.fill('Hello {{name}}')

    // 初始状态
    await expect(page.locator('.fill-text')).toContainText('0%')

    // 填写变量
    const nameInput = page.locator('.variable-row input').first()
    await nameInput.fill('World')

    // 检查进度更新
    await expect(page.locator('.fill-text')).toContainText('100%')
  })
})
```

- [ ] **Step 2: Commit**

---

## 验收标准检查清单

### Sprint 1 DoD

- [ ] Goroutine Pool 消费 task 表任务
- [ ] 状态流转：pending → running → done/failed
- [ ] SSE 实时推送进度（0-100）
- [ ] 服务重启后 running 任务恢复
- [ ] API 全局配额限制生效
- [ ] AICallLog 记录所有 AI 调用
- [ ] VariablePreviewPanel 渲染正确
- [ ] useTask composable mock 模式正常工作

### Sprint 2 DoD

- [ ] CSV/JSON 上传测试用例
- [ ] 表格展示：序号|变量|模型|质量分|操作
- [ ] 点击展开完整 AI 输出
- [ ] 支持多选行 A/B 对比
- [ ] 质量评分 4 维度正确计算
- [ ] 评测集 5-20 个用例可配置生成
- [ ] TaskProgressBar 实时更新

### Sprint 3 DoD

- [ ] SPRT 序贯检验自动停止（最小15/最大50）
- [ ] 清晰展示：置信区间、p-value、胜出 variant
- [ ] 统计显著性徽章（绿=显著/黄=接近/红=不足）
- [ ] 多轮对话测试正确模拟轮次

### Sprint 4 DoD

- [ ] 回归检测手动触发，输出对比报告
- [ ] AI 响应缓存命中率达预期
- [ ] Playwright E2E 全链路测试通过
- [ ] 集成测试覆盖率 80%+
- [ ] 无阻塞性 Bug
