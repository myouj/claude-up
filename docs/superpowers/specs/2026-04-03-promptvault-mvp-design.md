# PromptVault MVP Implementation Design

> 日期：2026-04-03
> 参与人：技术经理、产品经理、前端开发 x2、后端开发 x2、UI 设计师
> 参考文档：`docs/落地实施方案.md`

---

## 一、团队与流程

### 1.1 团队构成

| 角色 | 人数 | 职责 |
|------|------|------|
| 后端开发 A | 1 | Worker、Task 模型、API 限流 |
| 后端开发 B | 1 | SPRT 引擎、评分服务、评测集服务 |
| 前端开发 A | 1 | VariablePreviewPanel、Prompt 编辑页改造 |
| 前端开发 B | 1 | useTask composable、BatchTest、CostCenter |
| 技术负责人 | - | Code Review、PR 审批、Sprint 协调 |

### 1.2 任务管理

- **工具**：GitHub Issues + Projects（Table 视图）
- **Sprint 管理**：7 周 4 个 Sprint，每 Sprint 末生成 release notes

### 1.3 完成标准

- [ ] **Code Review**：PR 必须经过技术负责人 review 通过
- [ ] **自动化测试**：单元测试（service 层）+ 集成测试（Gin Mock）+ E2E（Playwright）
- [ ] **测试覆盖率**：后端 80%+，前端组件测试覆盖新增组件

### 1.4 开发隔离

**并行开发必须使用 Git Worktree**：
- 每个模块在独立 worktree 中开发
- 完成后 PR 合并回 main
- Worktree 路径命名：`worktrees/{feature-name}-{developer}/`
- 示例：`worktrees/sprint1-worker-backend-a/`

---

## 二、Sprint 规划

### Sprint 1：基础设施（Week 1-2）

| 模块 | 负责人 | Worktree 路径 | 依赖 |
|------|--------|---------------|------|
| Task 模型 + Worker | 后端 A | `worktrees/sprint1-worker-backend-a/` | 无 |
| API 限流/配额 | 后端 A | `worktrees/sprint1-quota-backend-a/` | Task 模型 |
| AICallLog middleware | 后端 B | `worktrees/sprint1-aicall-backend-b/` | 无 |
| ActivityLog 扩展 | 后端 B | `worktrees/sprint1-activity-backend-b/` | 无 |
| VariablePreviewPanel | 前端 A | `worktrees/sprint1-varpanel-frontend-a/` | 无（前端独立） |
| useTask composable (mock) | 前端 B | `worktrees/sprint1-usetask-frontend-b/` | 无（前端独立） |
| CostCenter 基础组件 | 前端 B | `worktrees/sprint1-costcenter-frontend-b/` | 无（前端独立） |

**并行策略**：
- 有依赖的模块按顺序：Worker → Quota
- 无依赖的模块并行：AICallLog + ActivityLog + 前端组件

**接口约定**：
- Sprint 1 开始前约定 Task API Schema（见第三节）
- 后端完成 API 后用 Swaggo 生成 Swagger
- 前端对接时根据 Swagger 调整

### Sprint 2：核心测试（Week 3-4）

| 模块 | 负责人 | Worktree 路径 | 依赖 |
|------|--------|---------------|------|
| BatchService | 后端 A | `worktrees/sprint2-batch-backend-a/` | Worker |
| ScoringService | 后端 B | `worktrees/sprint2-scoring-backend-b/` | AI Provider |
| EvalService | 后端 B | `worktrees/sprint2-eval-backend-b/` | AI Provider |
| SSE 进度推送 | 后端 A | `worktrees/sprint2-sse-backend-a/` | Worker |
| BatchTestTable | 前端 A | `worktrees/sprint2-batchtable-frontend-a/` | useTask |
| BatchTestCard | 前端 B | `worktrees/sprint2-batchcard-frontend-b/` | useTask |
| QualityScoreCard | 前端 A | `worktrees/sprint2-scorecard-frontend-a/` | 无（纯前端） |
| TaskProgressBar | 前端 B | `worktrees/sprint2-progress-frontend-b/` | useTask |

### Sprint 3：统计分析（Week 5-6）

| 模块 | 负责人 | Worktree 路径 | 依赖 |
|------|--------|---------------|------|
| SPRT 序贯检验 Engine | 后端 A | `worktrees/sprint3-sprt-backend-a/` | BatchService |
| ABTestService | 后端 B | `worktrees/sprint3-abtest-backend-b/` | SPRT Engine |
| MultiRound Service | 后端 B | `worktrees/sprint3-multiround-backend-b/` | BatchService |
| ABTestSequentialPanel | 前端 A | `worktrees/sprint3-abpanel-frontend-a/` | useTask + SSE |
| ABTestList + ABTestDetail | 前端 B | `worktrees/sprint3-ablist-frontend-b/` | 后端 API |

### Sprint 4：收尾（Week 7）

| 模块 | 负责人 | Worktree 路径 | 依赖 |
|------|--------|---------------|------|
| Regression Service | 后端 A | `worktrees/sprint4-regression-backend-a/` | EvalService |
| Cache Service | 后端 A | `worktrees/sprint4-cache-backend-a/` | 无 |
| 集成测试 | 后端 A+B | `worktrees/sprint4-integration/` | 所有后端模块 |
| E2E 测试 | 前端 A+B | `worktrees/sprint4-e2e/` | 所有前端模块 |
| Bug 修复 + 发布 | 所有人 | main 分支 | - |

---

## 三、后端架构

### 3.1 目录结构

```
backend/
├── main.go                      # 入口、路由、中间件
├── handlers/                     # HTTP 层
│   ├── task.go                  # 任务管理 API [NEW]
│   ├── batch.go                 # 批量测试 API [NEW]
│   ├── ab_test.go               # A/B 测试 API [NEW]
│   ├── scoring.go               # 质量评分 API [NEW]
│   ├── eval.go                  # 评测集 API [NEW]
│   ├── regression.go             # 回归检测 API [NEW]
│   └── ...
├── service/                     # 业务逻辑层
│   ├── task.go                  # TaskService [NEW]
│   ├── batch.go                 # BatchService [NEW]
│   ├── ab_test.go               # ABTestService [NEW]
│   ├── scoring.go                # ScoringService [NEW]
│   ├── eval.go                  # EvalService [NEW]
│   ├── regression.go             # RegressionService [NEW]
│   └── ...
├── worker/                      # Worker 层 [NEW]
│   ├── worker.go                # Goroutine Pool + DB Polling
│   ├── executor.go              # 任务执行器
│   ├── sse.go                  # SSE 事件推送
│   └── cache.go                # AI 响应缓存
├── models/                      # 数据模型
│   ├── task.go                  # Task 模型 [NEW]
│   ├── eval_set.go              # EvalSet 模型 [NEW]
│   ├── ab_test.go               # ABTest 模型 [NEW]
│   ├── ai_call_log.go           # AICallLog 模型 [NEW]
│   └── ...
├── middleware/
│   └── ai_call_log.go          # AI 调用日志拦截 [NEW]
└── docs/                        # Swagger 文档
```

### 3.2 Task 模型

```go
type Task struct {
    ID          uint      `gorm:"primaryKey"`
    Type        string    // batch_test | ab_test | eval_gen | regression | multi_turn
    Status      string    // pending | running | done | failed | cancelled
    Payload     string    // JSON，任务参数
    Progress    int       // 0-100
    Result      string    // JSON，执行结果
    Error       string    // 错误信息
    RetryCount  int       // 重试次数
    CreatedAt   time.Time
    UpdatedAt   time.Time
    RunAt       time.Time // 计划执行时间
    StartedAt   *time.Time
    CompletedAt *time.Time
}
```

### 3.3 Worker 配置

| 参数 | 默认值 | 说明 |
|------|--------|------|
| PoolSize | 5 | Goroutine 池大小 |
| PollInterval | 3s | SQLite 轮询间隔 |
| MaxRetries | 3 | 失败重试次数 |

### 3.4 Task API（Swaggo Schema）

```go
// POST /api/tasks - 创建任务
type CreateTaskRequest struct {
    Type    string `json:"type" binding:"required"`
    Payload string `json:"payload" binding:"required"`
}

// GET /api/tasks/:id - 任务详情
type TaskResponse struct {
    ID         uint      `json:"id"`
    Type       string    `json:"type"`
    Status     string    `json:"status"`
    Progress   int       `json:"progress"`
    Result     string    `json:"result,omitempty"`
    Error      string    `json:"error,omitempty"`
    CreatedAt  time.Time `json:"created_at"`
}

// GET /api/tasks/:id/progress - SSE 进度流
// event: progress
// data: {"current": 5, "total": 20, "progress": 25}
```

### 3.5 SSE 事件格式

```go
// 进度事件
event: progress
data: {"current": 5, "total": 20, "progress": 25, "status": "running"}

// 完成事件
event: complete
data: {"status": "done", "result": {...}}

// 错误事件
event: error
data: {"status": "failed", "error": "..."}
```

### 3.6 SPRT 序贯检验实现

SPRT（Sequential Probability Ratio Test）直接实现，约 150 行 Go 代码：

```go
// 核心逻辑：计算似然比，判断是否接受/拒绝/继续
func SPRT(nA, nB int, scoresA, scoresB []float64, alpha, beta float64) Decision
// Decision: accept | reject | continue
```

**参数配置**：
- 最小样本：15 次
- 最大样本：50 次
- 显著性水平 α：0.05
- 功效水平 β：0.2

### 3.7 质量评分服务

4 维度评分（规则 + AI 结合）：

| 维度 | 权重 | 评分方式 |
|------|------|----------|
| Clarity | 30% | 规则：变量占位符数量、长度、格式规范 |
| Completeness | 30% | 规则：必填字段检查 + AI 评估 |
| Example | 25% | 规则：示例数量、质量 + AI 评估 |
| Role | 15% | 规则：角色关键词 + AI 评估 |

**AI 评分 Prompt**：由后端构造，调用统一 AI Provider。

### 3.8 API 限流

- **维度**：全局总配额（不区分用户）
- **存储**：SQLite `quotas` 表
- **升级路径**：后续扩展 Redis

```go
type Quota struct {
    ID         uint   `gorm:"primaryKey"`
    Provider   string // openai | claude | gemini | minimax
    Model      string
    Limit      int    // 月度上限
    Usage      int    // 当月已用
    ResetAt    time.Time
}
```

### 3.9 数据库新增表

| 表名 | 说明 |
|------|------|
| `tasks` | 异步任务队列 |
| `eval_sets` | 评测集 |
| `ab_tests` | A/B 测试配置 |
| `ab_test_results` | A/B 测试结果 |
| `ai_call_logs` | AI 调用日志 |
| `response_cache` | AI 响应缓存 |
| `quotas` | 配额管理 |

---

## 四、前端架构

### 4.1 目录结构

```
frontend/src/
├── views/
│   ├── PromptEditor.vue       [改造] 60/40 分栏
│   ├── BatchTest.vue          [NEW] 批量测试视图
│   ├── BatchTestResult.vue    [NEW] 任务详情视图
│   └── ...
├── components/
│   ├── VariablePreviewPanel.vue [NEW] 右侧 40% 预览面板
│   ├── BatchTestTable.vue      [NEW] 批量测试表格
│   ├── BatchTestCard.vue       [NEW] 展开卡片
│   ├── QualityScoreCard.vue    [NEW] 四维度评分卡
│   ├── TaskProgressBar.vue     [NEW] SSE 进度条
│   └── ABTestSequentialPanel.vue [NEW] 序贯检验面板
└── composables/
    ├── useTask.js              [NEW] 任务状态 + SSE
    ├── useBatchTest.js         [NEW] 批量测试
    ├── useQualityScore.js      [NEW] 质量评分
    └── useSSE.js               [NEW] SSE 封装
```

### 4.2 组件设计

#### VariablePreviewPanel

**位置**：Prompt 编辑页右侧 40% 区域

**功能**：
- 变量输入（自动解析 `{{variable}}`）
- 填充进度指示器
- 渲染预览（实时替换变量值）
- 质量评分卡（可折叠）

**布局**：
```
┌─────────────────────────────────────────────────────────┐
│ 变量预览面板                                    40%    │
├─────────────────────────────────────────────────────────┤
│ 📝 变量输入                                          │
│ ┌─────────────────────────────────────────────────┐   │
│ │ {{variable}}  [已填写]                        │   │
│ │ {{name}}      [输入 name 的值...]             │   │
│ │ ▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░ 50%                │   │
│ └─────────────────────────────────────────────────┘   │
│                                                         │
│ 👁 渲染预览                                           │
│ ┌─────────────────────────────────────────────────┐   │
│ │ 这里是渲染后的内容                               │   │
│ └─────────────────────────────────────────────────┘   │
│                                                         │
│ 📊 质量评分                                            │
│ ┌─────────────────────────────────────────────────┐   │
│ │ Clarity 85  Complete 72  Example 60  Role 90  │   │
│ └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

**色彩**：统一使用 PromptVault Design System

| Token | 值 | 用途 |
|-------|-----|------|
| `--color-primary` | #2563EB | 主色调、按钮、链接 |
| `--color-bg` | #F8FAFC | 背景色 |
| `--color-surface` | #FFFFFF | 卡片背景 |
| `--color-border` | #E2E8F0 | 边框 |
| `--color-text-primary` | #1E293B | 主文本 |
| `--color-text-secondary` | #64748B | 次要文本 |

### 4.3 useTask Composable

```javascript
// 接口约定
const { task, progress, status, startTask, cancelTask } = useTask(taskId)

// SSE 事件
// - progress: { current, total, progress }
// - complete: { status, result }
// - error: { status, error }

// Mock 模式（Sprint 1）
// - 前端独立开发时，返回模拟数据
// - 后端 API 就绪后，对接真实 SSE
```

### 4.4 批量测试表格

**布局**：
- el-table 紧凑模式（48px 行高）
- 固定列：序号、变量、模型、质量分、操作
- 点击展开行显示完整 AI 输出（卡片形式）
- 多选行支持横向对比

### 4.5 A/B 测试面板

**序贯检验进度条**：
```
Variant A ████████████░░░░░░░░░░░░░░░░░░ 35次
Variant B ██████████████░░░░░░░░░░░░░░░░░ 35次
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
最低样本: 15 | 已完成: 35 | 最大: 50
[████████░░░░] 统计显著性: 显著 (p < 0.05)
```

---

## 五、验收标准（DoD）

### Sprint 1：基础设施

- [ ] Goroutine Pool 消费 task 表任务
- [ ] 状态流转：pending → running → done/failed
- [ ] SSE 实时推送进度（0-100）
- [ ] 服务重启后 running 任务恢复
- [ ] API 全局配额限制生效
- [ ] AICallLog 记录所有 AI 调用
- [ ] VariablePreviewPanel 渲染正确
- [ ] useTask composable mock 模式正常工作

### Sprint 2：核心测试

- [ ] CSV/JSON 上传测试用例
- [ ] 表格展示：序号|变量|模型|质量分|操作
- [ ] 点击展开完整 AI 输出
- [ ] 支持多选行 A/B 对比
- [ ] 质量评分 4 维度正确计算
- [ ] 评测集 5-20 个用例可配置生成
- [ ] TaskProgressBar 实时更新

### Sprint 3：统计分析

- [ ] SPRT 序贯检验自动停止（最小15/最大50）
- [ ] 清晰展示：置信区间、p-value、胜出 variant
- [ ] 统计显著性徽章（绿=显著/黄=接近/红=不足）
- [ ] 多轮对话测试正确模拟轮次

### Sprint 4：收尾

- [ ] 回归检测手动触发，输出对比报告
- [ ] AI 响应缓存命中率达预期
- [ ] Playwright E2E 全链路测试通过
- [ ] 集成测试覆盖率 80%+
- [ ] 无阻塞性 Bug

---

## 六、风险与缓解

| 风险 | 级别 | 缓解方案 |
|------|------|----------|
| SPRT 序贯检验复杂度 | 高 | Phase 1 固定30次，Phase 2 再引入完整 SPRT |
| SQLite 并发瓶颈 | 中 | Worker 池控制并发；预留 PostgreSQL 升级路径 |
| AI API 成本超支 | 中 | Sprint 1 实现配额管理 + AI Call Log |
| 批量测试前端性能 | 低 | el-table 虚拟滚动（规模大时） |
| SSE 断线 | 低 | 前端自动重连（指数退避） |
| Worktree 合并冲突 | 中 | 每日同步 main 分支；冲突及时沟通 |

---

## 七、文件变更汇总

| 类别 | 数量 | 说明 |
|------|------|------|
| 后端新建 | 12+ | worker/, models/, handlers/, service/ |
| 后端修改 | 3 | main.go、现有 handler 扩展 |
| 前端新建 | 11 | 6 组件 + 2 视图 + 3 composable |
| 前端修改 | 8 | 现有视图/组件改造 |
| 数据库迁移 | 1 | GORM AutoMigrate |

**总文件变更**：约 34 个
