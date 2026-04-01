# PromptVault V1 技术方案细化报告

> 技术经理出品 | 日期：2026-04-01
> 基于 `docs/功能规划_v2.md` 及现有代码库分析

---

## 一、技术架构图

### 1.1 整体架构（单体 + 内嵌 Worker）

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend (Vue 3)                       │
│   PromptEditor │ PromptTester │ ABTest │ BatchTest │ ...     │
│                     ↕ SSE / HTTP                             │
└────────────────────────────┬────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────┐
│                     Backend (Go/Gin)                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │ Handlers │  │ Service  │  │  Worker  │  │ Middleware│  │
│  │          │  │          │  │  (Task)  │  │          │  │
│  │ prompt   │  │ prompt   │  │  Pool    │  │ CORS     │  │
│  │ test     │  │ test     │  │  SSE     │  │ RateLim  │  │
│  │ ab_test  │  │ ab_test  │  │  Notify  │  │ TraceID  │  │
│  │ batch    │  │ scoring  │  │          │  │ Recovery │  │
│  │ scoring  │  │ eval_gen │  │          │  │          │  │
│  └────┬─────┘  └────┬─────┘  └───┬──────┘  └──────────┘  │
│       │             │            │                           │
│       └─────────────┼────────────┘                           │
│                     ▼                                         │
│  ┌──────────────────────────────────────────────────────┐    │
│  │              Data Layer (GORM + SQLite)               │    │
│  │  Prompt │ Version │ TestRecord │ Activity │ Setting │    │
│  │  Skill  │ Agent   │ Translation│   NEW:   │  NEW:   │    │
│  │         │         │            │  Task   │EvalSet  │    │
│  │         │         │            │AICallLog│ Quota   │    │
│  └──────────────────────────────────────────────────────┘    │
│                     │                                         │
│                     ▼                                         │
│  ┌──────────────────────────────────────────────────────┐    │
│  │           AI Provider (OpenAI/Claude/Gemini/MiniMax)│    │
│  └──────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 后端模块依赖关系

```
main.go (入口)
    ├── handlers/        (HTTP 层 — 请求/响应)
    │   ├── prompt.go    ← service.PromptService
    │   ├── test.go     ← service.TestService, AI Provider
    │   ├── ab_test.go  ← service.ABTestService, Worker
    │   ├── batch.go    ← service.BatchService, Worker
    │   ├── scoring.go  ← service.ScoringService, AI Provider
    │   ├── eval.go     ← service.EvalGenService, AI Provider
    │   ├── regression.go← service.RegressionService, Worker
    │   └── activity.go
    │
    ├── service/         (业务逻辑层 — 无 HTTP 依赖)
    │   ├── prompt.go
    │   ├── test.go     (测试执行、结果评分)
    │   ├── ab_test.go  (序贯检验、SPRT 统计引擎)
    │   ├── batch.go    (批量测试分发)
    │   ├── scoring.go  (质量评分卡：4 维度)
    │   ├── eval.go     (评测集生成)
    │   ├── regression.go(回归检测：三触发机制)
    │   ├── task.go     (Worker 任务调度)
    │   └── worker.go   (内嵌 Worker 池)
    │
    ├── models/          (数据模型)
    │   ├── task.go     [NEW]  异步任务队列
    │   ├── eval_set.go [NEW]  评测集
    │   ├── ai_call_log.go[NEW] AI 调用日志
    │   ├── quota.go    [NEW]  配额管理
    │   └── (现有: prompt, test_record, activity...)
    │
    ├── middleware/      (Gin 中间件 — 复用现有)
    │   └── trace.go, request_logger.go, pagination.go
    │
    └── utils/logger.go (复用现有 trace_id 机制)
```

### 1.3 前端模块结构

```
frontend/src/
├── views/
│   ├── PromptEditor.vue     [增强] 右侧 40% 实时预览面板
│   ├── PromptTester.vue     [增强] 多轮对话、批量测试入口
│   ├── BatchTest.vue        [NEW] 表格+卡片批量测试
│   ├── ABTestList.vue       [已有骨架] 完善后端联调
│   ├── ABTestDetail.vue     [NEW] 序贯检验可视化
│   ├── ScoreCard.vue        [NEW] 质量评分卡
│   ├── RegressionAlert.vue  [NEW] 回归告警面板
│   └── CostCenter.vue       [NEW] 成本分析看板
├── components/
│   ├── VariablePreviewer.vue [已有] 增强：支持默认值语法
│   ├── TestResultCard.vue    [NEW] 测试结果卡片
│   ├── TestResultTable.vue   [NEW] 测试结果表格
│   ├── ABTestChart.vue       [NEW] A/B 测试序贯图
│   └── ScoreRadar.vue        [NEW] 评分雷达图
└── composables/
    ├── useABTest.js          [NEW] A/B 测试状态管理
    ├── useTask.js            [NEW] SSE 任务进度
    └── useCost.js            [NEW] 成本统计
```

---

## 二、技术债务评估

### 2.1 与现有代码的冲突点

| # | 冲突点 | 当前状态 | 影响分析 | 兼容性建议 |
|---|--------|----------|----------|------------|
| 1 | **TestRecord 模型** | 现有 `TestRecord` 只存储单次测试结果，无评测集、多轮对话关联 | P1 功能（批量测试、质量评分）需要扩展字段 | 向前兼容：新增字段不修改现有字段，新增 `test_type` 字段区分单测/批量/多轮/A/B |
| 2 | **test.go Handler** | 当前 `Test()` 方法同步调用 AI，阻塞 HTTP | P0 批量测试需要异步化 | 重构：将 AI 调用移入 Worker，`Test()` 改为提交任务并返回 task_id |
| 3 | **ActivityLog 表** | 现有 `ActivityLog` 只记录 CRUD 操作 | 新功能需要记录 AI 调用、批量测试、A/B 测试等活动 | 扩展：新增 `action_type` 枚举值，新增 `trace_id` 关联 AI 调用链 |
| 4 | **VariablePreviewer** | 当前使用 `{{var}}` 简单正则 | P0 变量实时预览需要支持 `{{var\|default}}` 默认值语法 | 向前兼容：正则改为 `/\{\{([^}\|]+)(?:\|([^}]*))?\}\}/g`，现有语法不变 |
| 5 | **AI Provider** | 现有 `AIProvider` 接口只有 `Call()` 方法 | 评分、评测集生成等新场景需要不同调用模式 | 扩展接口：新增 `CallWithScore()`、`GenerateEvalSet()` 等专用方法 |
| 6 | **Rate Limiter** | 现有限流器基于 IP，无用户/配额维度 | P1 配额管理需要用户维度的 token 配额 | 扩展 `rateLimiter`：增加 `userID` key，支持 per-user 配额 |
| 7 | **Prompt Versioning** | 现有版本只记录 content diff | 回归检测需要关联评测集评分历史 | 新增 `PromptVersion.score_history` JSON 字段 |
| 8 | **Frontend Router** | 现有路由固定，新视图需要注册 | 新增批量测试、A/B 详情等路由 | 直接添加，无需冲突 |

### 2.2 数据库迁移计划

```sql
-- Sprint 1: 新增 Task 表（异步任务队列）
CREATE TABLE tasks (
  id TEXT PRIMARY KEY,
  type TEXT NOT NULL,           -- "batch_test" | "ab_test" | "eval_gen" | "optimize" | "regression"
  prompt_id INTEGER,
  payload TEXT,                 -- JSON
  status TEXT NOT NULL,         -- "pending" | "running" | "done" | "failed"
  progress INTEGER DEFAULT 0,    -- 0-100
  result TEXT,                  -- JSON nullable
  error TEXT,
  created_at DATETIME,
  started_at DATETIME,
  completed_at DATETIME
);

-- Sprint 1: 扩展 TestRecord 表
ALTER TABLE test_records ADD COLUMN test_type TEXT DEFAULT 'single';
ALTER TABLE test_records ADD COLUMN eval_set_id INTEGER;
ALTER TABLE test_records ADD COLUMN batch_id TEXT;
ALTER TABLE test_records ADD COLUMN quality_score REAL;
ALTER TABLE test_records ADD COLUMN score_clarity REAL;
ALTER TABLE test_records ADD COLUMN score_completeness REAL;
ALTER TABLE test_records ADD COLUMN score_example REAL;
ALTER TABLE test_records ADD COLUMN score_role REAL;
ALTER TABLE test_records ADD COLUMN round_index INTEGER DEFAULT 0;
ALTER TABLE test_records ADD COLUMN trace_id TEXT;

-- Sprint 2: 新增 EvalSet 表
CREATE TABLE eval_sets (
  id TEXT PRIMARY KEY,
  prompt_id INTEGER NOT NULL,
  name TEXT,
  cases TEXT NOT NULL,           -- JSON array of test cases
  auto_generated BOOLEAN DEFAULT FALSE,
  created_at DATETIME
);

-- Sprint 2: 新增 AICallLog 表
CREATE TABLE ai_call_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  trace_id TEXT,
  provider TEXT,
  model TEXT,
  prompt_tokens INTEGER,
  completion_tokens INTEGER,
  total_tokens INTEGER,
  latency_ms INTEGER,
  cost REAL,
  model_version TEXT,
  created_at DATETIME
);

-- Sprint 3: 新增 ABTest 表
CREATE TABLE ab_tests (
  id TEXT PRIMARY KEY,
  prompt_id INTEGER NOT NULL,
  name TEXT,
  variant_a_id TEXT,
  variant_b_id TEXT,
  status TEXT DEFAULT 'running',
  winner TEXT,
  total_runs INTEGER DEFAULT 0,
  current_significance REAL,
  target_significance REAL DEFAULT 0.05,
  min_samples INTEGER DEFAULT 15,
  max_samples INTEGER DEFAULT 50,
  created_at DATETIME,
  completed_at DATETIME
);

-- Sprint 3: 新增 Quota 表
CREATE TABLE quotas (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  provider TEXT NOT NULL,
  daily_limit INTEGER,
  monthly_limit INTEGER,
  daily_used INTEGER DEFAULT 0,
  monthly_used INTEGER DEFAULT 0,
  reset_at DATETIME
);
```

### 2.3 API 兼容性策略

| 变更类型 | 策略 | 示例 |
|----------|------|------|
| 新增 API | 直接添加 | `POST /api/prompts/:id/batch-test` |
| 扩展现有 API | 添加可选字段 | `POST /api/prompts/:id/test` 新增 `eval_set_id`、`batch_id` 参数 |
| 破坏性变更 | 新增端点 | 不修改现有 `GET /prompts/:id/tests`，新增 `GET /prompts/:id/tests-v2` |
| 删除旧 API | v2 稳定后废弃 | 保留旧端点 1 个 Sprint，过期删除 |

---

## 三、Sprint 里程碑规划（7 周）

```
Week 1-2 (Sprint 1): 基础设施 — 异步任务队列 + 变量预览
Week 3-4 (Sprint 2): 核心测试 — 批量测试 + 质量评分卡 + 评测集
Week 5-6 (Sprint 3): 统计分析 — A/B 测试 + 序贯检验 + 多轮对话
Week 7   (Sprint 4): 收尾 + 回归检测 + 配额管理
```

### Sprint 1: 基础设施（Week 1-2）

| 任务 | 后端 | 前端 | 依赖 | 交付物 |
|------|------|------|------|--------|
| 异步任务队列 (Task 模型 + Worker 池) | Task handler, Worker, SSE | useTask composable | 无 | `POST /api/tasks`, `GET /api/tasks/:id`, SSE 进度推送 |
| test.go 重构为异步提交 | Test handler 改为提交任务 | 任务提交 + 轮询 UI | Task 队列 | `POST /api/prompts/:id/test` 返回 task_id |
| VariablePreviewer 支持默认值语法 | 无 | 增强正则，支持 `{{var\|default}}` | 无 | 现有功能向后兼容 |
| 右侧实时预览面板（60/40 布局） | 无 | PromptEditor 重构 | VariablePreviewer | 新布局通过评审 |
| ActivityLog 扩展字段 | activity handler 支持新 action_type | 无 | 无 | `test_started`, `batch_started`, `ab_test_started` |
| AI Call Log 表 + 记录埋点 | middleware 层拦截 AI 调用 | 无 | 无 | `ai_call_logs` 表数据正确 |
| 单元测试补充 | Task handler、Worker、变量解析测试 | VariablePreviewer 组件测试 | 无 | 覆盖率 >85% |

### Sprint 2: 核心测试功能（Week 3-4）

| 任务 | 后端 | 前端 | 依赖 | 交付物 |
|------|------|------|------|--------|
| 批量测试 Handler | Batch handler, Batch service | 无 | Task 队列, AI Call Log | `POST /api/prompts/:id/batch-test` |
| 批量测试结果展示 | 无 | BatchTest.vue（表格+卡片） | 后端 API | 表格：序号\|变量摘要\|模型\|质量分\|操作；卡片：展开完整输出 |
| 质量评分卡（4 维度） | Scoring service：Clarity 30% + Completeness 30% + Example 25% + Role 15% | ScoreCard.vue, ScoreRadar.vue | AI Provider | 可配置权重，雷达图展示 |
| 自动评测集生成 | EvalGen service，AI 辅助 + 人工补充 | EvalSetEditor.vue | AI Provider | `POST /api/eval-sets`, `POST /api/eval-sets/:id/cases` |
| TestRecord 扩展字段 | 迁移脚本，新增 test_type、quality_score 等 | 无 | 数据库迁移 | 向后兼容 |
| SSE 任务进度推送 | Worker 每完成一条测试推送进度 | useTask.js SSE 监听 | Task 队列 | 前端实时进度条 |
| 单元测试补充 | Scoring, EvalGen, Batch service 测试 | BatchTest 组件测试 | 无 | 覆盖率 >85% |

### Sprint 3: 统计分析（Week 5-6）

| 任务 | 后端 | 前端 | 依赖 | 交付物 |
|------|------|------|------|--------|
| A/B 测试 Engine | ABTest service, SPRT 序贯检验 | 无 | Task 队列, TestRecord | 15-50 次动态停止，α=0.05，β=0.80 |
| A/B 测试 Handler | `POST /api/prompts/:id/ab-test`, `GET /api/ab-tests/:id` | 无 | ABTest service | RESTful API |
| A/B 测试前端完善 | 无 | ABTestList.vue 联调, ABTestDetail.vue | 后端 API | 序贯检验可视化（折线图）、胜出标记 |
| 多轮对话流程测试 | MultiRound service，轮次模拟 | MultiRoundTester.vue | Batch test | 对话轮次可配置，暴露上下文累积问题 |
| A/B 测试结果卡片/表格 | 无 | ABTestCompare.vue | ABTest API | 横向对比 A/B 变体输出 |
| 质量评分卡 UI 集成 | 无 | PromptEditor 集成 ScoreCard | Scoring service | 编辑页显示评分 |
| 单元测试补充 | SPRT, MultiRound service 测试 | A/B 组件测试 | 无 | 覆盖率 >85% |

### Sprint 4: 收尾（Week 7）

| 任务 | 后端 | 前端 | 依赖 | 交付物 |
|------|------|------|------|--------|
| Prompt 回归检测 | Regression service（保存轻量 + 定时全量 + 模型升级专项） | RegressionAlert.vue | EvalSet, AI Call Log | 低于基线 15% 告警 |
| API 限流与配额管理 | Quota service，per-provider/token 配额 | CostCenter.vue | 无 | `GET /api/quotas`, `GET /api/cost-stats` |
| AI 响应缓存层 | Cache service（内存 + 可选 Redis） | 无 | 无 | 相同 prompt+model 命中缓存 |
| 成本分析中心 | 无 | CostCenter.vue | AICallLog, Quota | Token 预算 + 月度看板 |
| 集成测试 | 全链路 API 测试 | E2E 测试（Playwright） | 所有 Sprint 产出 | 80%+ 覆盖率 |
| 文档更新 | API 文档补全 | 无 | 无 | docs/API.md 更新 |
| Bug 修复 + 优化 | 性能优化、边界情况处理 | UI 打磨 | 集成测试发现 | Sprint 4 结束发布 |

---

## 四、风险评估

| 风险项 | 概率 | 影响 | 风险等级 | 缓解方案 |
|--------|------|------|----------|----------|
| **SPRT 序贯检验实现复杂度高** | 中 | 高 | **高** | 1. 优先实现固定样本量版本（简单）<br>2. Phase 2 再引入 SPRT<br>3. 参考开源实现（如 scipy.stats） |
| **SQLite 并发写入瓶颈** | 中 | 中 | **中** | 1. Worker 池控制并发度（默认 5 个 worker）<br>2. 批量 insert 替代逐条 insert<br>3. Phase 2 可考虑升级到 PostgreSQL |
| **AI API 成本超预算** | 高 | 中 | **中** | 1. Sprint 1 就实现 AI Call Log + 配额管理<br>2. 批量测试限制最大次数<br>3. 缓存层减少重复调用 |
| **批量测试前端性能** | 低 | 中 | **低** | 1. 表格虚拟滚动（el-table-v2）<br>2. 分页加载测试历史<br>3. 卡片懒加载 |
| **回归检测误报** | 中 | 中 | **中** | 1. 保存轻量版阈值保守（15%）<br>2. 告警可配置<br>3. 人工确认后再计入 |
| **VariablePreviewer 正则兼容性** | 低 | 高 | **低** | 1. 新正则完全向后兼容 `{{var}}`<br>2. 充分单元测试覆盖 |
| **多轮对话上下文累积问题** | 中 | 中 | **中** | 1. 多轮测试独立存储 `round_index`<br>2. 提供"重置上下文"按钮<br>3. 可配置最大轮次 |
| **Task Worker 崩溃恢复** | 低 | 高 | **低** | 1. Task 表 status 持久化<br>2. 服务重启后 Worker 自动恢复<br>3. 超时任务自动标记 failed |
| **前端 SSE 连接断线** | 中 | 低 | **低** | 1. 前端自动重连（ exponential backoff）<br>2. 轮询降级方案 |
| **模型升级触发回归检测风暴** | 低 | 高 | **低** | 1. 错峰执行，控制并发<br>2. 优先级队列<br>3. 告警抑制（相同问题不重复告警） |

---

## 五、技术决策总结

| 议题 | 决策 | 理由 |
|------|------|------|
| 异步任务队列 | 内嵌 Worker + SQLite Task 表 + SSE | 保持单体架构，运维简单，扩展只需换实现 |
| 变量校验 | 正则先行（Phase 1），AST 处理（Phase 2） | Phase 1 需求简单，`{{var}}` 和 `{{var\|default}}` 正则足够 |
| A/B 显著性检验 | Phase 1 固定 30 次，Phase 2 再引入 SPRT | SPRT 实现复杂，先保证功能可用再优化效率 |
| 批量测试展示 | 表格为主 + 卡片展开 | 测试 case >5 个时表格效率更高，与规划一致 |
| 日志/可观测性 | 扩展 activity 表 + 新增 ai_call_log 表 | 与规划一致，复用现有 trace_id 机制 |
| AI 缓存 | Phase 1 内存缓存，Phase 2 可选 Redis | Phase 1 并发量小，内存缓存足够 |
| 数据库 | 保持 SQLite | 单体架构，SQLite 够用；Phase 2 按需升级 |

---

## 六、实施优先级矩阵

```
         难度
           │
  高       │  [AI 缓存层]     [SPRT 序贯检验]
           │                  [多轮对话]
           │                  [成本分析中心]
           │
  中       │  [批量测试后端]  [A/B 测试 Engine]
           │  [质量评分卡]    [回归检测]
           │  [评测集生成]
           │
  低       │  [Task 队列]    [VariablePreviewer]
           │  [SSE 推送]     [ABTest 前端完善]
           │  [配额管理]
           │
           └─────────────────────────────────
                    低           高
                       业务价值
```

---

*报告版本：1.0 | 撰写人：技术经理 | 审核：待定*
