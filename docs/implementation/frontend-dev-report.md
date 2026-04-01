# 前端实现方案报告

> 作者：前端开发
> 日期：2026-04-01
> 依据：功能规划_v2.md

---

## 一、现状分析

### 1.1 现有技术栈

| 类别 | 技术 | 说明 |
|------|------|------|
| 框架 | Vue 3 + Composition API | `<script setup>` 语法 |
| UI 库 | Element Plus | el-button, el-table, el-dialog 等 |
| 图表 | vue-chartjs + Chart.js | TestAnalytics 已使用 |
| 路由 | Vue Router | 路由配置在 `router/index.js` |
| HTTP | Axios | 封装在 `composables/useApi.js` |
| 状态 | Composables 模式 | 各业务模块独立 composable |
| 样式 | CSS Variables | 定义在 `App.vue` |

### 1.2 现有组件盘点

| 组件/视图 | 状态 | 需改造 |
|-----------|------|--------|
| `VariablePreviewer.vue` | 存在，独立面板 | 需支持 60/40 分栏布局 |
| `ABTestCompare.vue` | 存在，mock 数据 | 需对接真实 API |
| `ABTestList.vue` | 存在，mock 数据 | 需对接真实 API |
| `ABTestDetail.vue` | 存在，mock 数据 | 需支持序贯检验 UI |
| `PromptTester.vue` | 存在 | 需支持批量测试入口 |
| `TestAnalytics.vue` | 存在，图表完整 | 可复用图表组件 |
| `TestCompare.vue` | 存在 | 需支持批量测试结果 |
| `OptimizePrompt.vue` | 存在 | 无需改造 |

### 1.3 现有 composables

| 文件 | 职责 |
|------|------|
| `useApi.js` | Axios 封装，响应拦截 |
| `usePrompts.js` | Prompt CRUD API 封装 |
| `useVariablePreview.js` | 变量提取与渲染逻辑 |
| `useABTest.js` | Mock 数据，需替换为真实 API |
| `usePagination.js` | 分页逻辑 |

---

## 二、组件设计

### 2.1 组件清单

#### 新增组件（6 个）

| 组件 | 文件位置 | 描述 |
|------|----------|------|
| `VariablePreviewPanel` | `components/VariablePreviewPanel.vue` | 60/40 分栏的右侧预览面板 |
| `BatchTestTable` | `components/BatchTestTable.vue` | 批量测试结果表格 + 行展开卡片 |
| `BatchTestCard` | `components/BatchTestCard.vue` | 展开的 AI 回复卡片 |
| `QualityScoreCard` | `components/QualityScoreCard.vue` | 四维度质量评分卡 + 雷达图 |
| `TaskProgressBar` | `components/TaskProgressBar.vue` | 异步任务进度条（SSE 实时更新） |
| `ABTestSequentialPanel` | `components/ABTestSequentialPanel.vue` | 序贯检验进度面板 |

#### 修改组件（5 个）

| 组件 | 改动 |
|------|------|
| `VariablePreviewer.vue` | 支持 `side-by-side` 模式，提供左右分栏能力 |
| `ABTestList.vue` | 对接 `GET /api/ab-tests` 真实 API，支持状态筛选 |
| `ABTestDetail.vue` | 支持 SSE 实时进度，对接 `GET /api/ab-tests/:id` |
| `ABTestCompare.vue` | 集成 `QualityScoreCard`，支持统计显著性展示 |
| `PromptTester.vue` | 新增"批量测试"入口按钮，跳转到批量测试视图 |

#### 新增视图（2 个）

| 视图 | 路由 | 描述 |
|------|------|------|
| `BatchTest.vue` | `/prompts/:id/batch-test` | 批量测试创建与结果展示 |
| `BatchTestResult.vue` | `/prompts/:id/batch-test/:taskId` | 批量测试任务详情 |

### 2.2 组件布局示意

#### 2.2.1 变量实时预览面板（60/40 分栏）

```
┌─────────────────────────────────────────────────────────────────┐
│  PromptEditor.vue                                                │
│  ┌─────────────────────────────┬───────────────────────────────┤
│  │                             │                               │
│  │  PromptEditor 左侧 (60%)    │  VariablePreviewPanel (40%)   │
│  │  ┌───────────────────────┐  │  ┌─────────────────────────┐  │
│  │  │ 标题输入框            │  │  │ 变量预览面板             │  │
│  │  ├───────────────────────┤  │  │ ├─ 填充进度条 (x/5)      │  │
│  │  │                       │  │  │ ├─ 变量输入表单          │  │
│  │  │  Prompt 内容编辑      │  │  │ │  {{name}} __________   │  │
│  │  │  (textarea)           │  │  │ │  {{query}} __________  │  │
│  │  │                       │  │  │ │  {{lang}}  __________  │  │
│  │  │                       │  │  │ ├─ 渲染预览区            │  │
│  │  │                       │  │  │ │  实时渲染后的 prompt    │  │
│  │  └───────────────────────┘  │  │ ├─ [复制] [一键替换]    │  │
│  │                              │  └─────────────────────────┘  │
│  └─────────────────────────────┴───────────────────────────────┤
└─────────────────────────────────────────────────────────────────┘
```

**实现要点**：
- `VariablePreviewPanel` 接收 `content` prop，内部调用 `useVariablePreview` composable
- 当所有变量填充后，"一键替换"按钮将渲染结果写回父组件
- 右侧面板支持折叠（collapsed 状态保留填充值）
- 填充进度条实时更新，支持 `{{var|default}}` 默认值语法高亮

#### 2.2.2 批量测试结果表格 + 卡片展开

```
┌──────────────────────────────────────────────────────────────────┐
│  BatchTestResult.vue                                              │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │ 筛选: [模型 ▼] [时间范围] [状态]      导出 | 多选对比        │ │
│  ├──────────────────────────────────────────────────────────────┤ │
│  │ # │ 变量摘要      │ 模型    │ 质量分 │ 延迟  │ Token │ 操作  │ │
│  ├───┼───────────────┼─────────┼────────┼───────┼───────┼───────┤ │
│  │ 1 │ name: 小明   │ GPT-4   │ ★★★★☆ │ 1.2s  │ 850   │ ▶展开 │ │
│  │ 2 │ name: 小红   │ GPT-4   │ ★★★★★ │ 1.1s  │ 820   │ ▶展开 │ │
│  │ 3 │ name: 小刚   │ Claude3 │ ★★★☆☆ │ 2.3s  │ 1200  │ ▶展开 │ │
│  ├──────────────────────────────────────────────────────────────┤ │
│  │ ▶ 展开详情 (BatchTestCard)                                   │ │
│  │ ┌────────────────────────────────────────────────────────┐   │ │
│  │ │ 输入 (变量):                                            │   │ │
│  │ │ name = "小明" | query = "北京的天气" | lang = "中文"   │   │ │
│  │ │                                                        │   │ │
│  │ │ AI 回复:                                                │   │ │
│  │ │ 今天北京天气晴，气温15-22°C...                          │   │ │
│  │ │                                                        │   │ │
│  │ │ 质量评分: Clarity ★★★★☆ | Completeness ★★★☆☆          │   │ │
│  │ │         Example ★★★★☆ | Role ★★★★★                    │   │ │
│  │ └────────────────────────────────────────────────────────┘   │ │
│  └──────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────┘
```

**实现要点**：
- `BatchTestTable` 使用 `el-table` 的 `expand` 插槽展开行
- 展开后渲染 `BatchTestCard` 组件
- 支持多选行（checkbox）横向对比
- 表格列支持排序（质量分、延迟、Token）

#### 2.2.3 A/B 测试中心 UI

```
┌─────────────────────────────────────────────────────────────────┐
│  ABTestDetail.vue                                                │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │ [返回]  Code Review Prompt 优化测试        [重新测试] [导出] │ │
│  ├─────────────────────────────────────────────────────────────┤ │
│  │ 运行中 ● │ Prompt: 代码审查专家 │ 总运行: 12/50 │ 胜出: —   │ │
│  ├─────────────────────────────────────────────────────────────┤ │
│  │ 序贯检验进度:                                             │ │
│  │ Variant A: 6/15 ─────●─────○─────○─── 目标15  Variant B: 6 │ │
│  │ 当前胜率: A=48% B=52%  │  显著性: 未达显著 (需再跑3组)      │ │
│  ├────────────────────────┬────────────────────────────────────┤ │
│  │ Variant A (原版)       │ Variant B (结构化)                 │ │
│  │ ┌──────────────────┐  │ ┌──────────────────────────────┐   │ │
│  │ │ 平均分: 3.2  ★★★ │  │ │ 平均分: 4.5  ★★★★☆          │   │ │
│  │ │ 运行次数: 6      │  │ │ 运行次数: 6                  │   │ │
│  │ │ 平均延迟: 1200ms │  │ │ 平均延迟: 1450ms              │   │ │
│  │ │ Token: 850       │  │ │ Token: 1200                  │   │ │
│  │ ├──────────────────┤  │ ├──────────────────────────────┤   │ │
│  │ │ [质量评分卡 ▼]   │  │ │ [质量评分卡 ▼]               │   │ │
│  │ │ Clarity    3.0  │  │ │ Clarity    4.5               │   │ │
│  │ │ Complete   3.2  │  │ │ Complete   4.8               │   │ │
│  │ │ Example     3.0  │  │ │ Example     4.2               │   │ │
│  │ │ Role        3.5  │  │ │ Role        4.5               │   │ │
│  │ └──────────────────┘  │ └──────────────────────────────┘   │ │
│  └────────────────────────┴────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │ 统计汇总: 总运行 12 │ 平均分差 +1.3 │ 胜出: 待定 │ 置信度 62%│ │
│  └─────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

**实现要点**：
- 序贯检验进度条：动态显示 `已完成/最低要求/最大上限`
- Variant 对比卡片支持折叠展开质量评分详情
- SSE 连接实时接收任务进度更新
- 统计显著性结果后自动高亮胜出 Variant

#### 2.2.4 质量评分卡组件

```
┌─────────────────────────────────────┐
│  QualityScoreCard.vue               │
│  ┌───────────┬───────────────────┐  │
│  │ 雷达图    │ 综合分: 85/100     │  │
│  │           │ ████████████░░   │  │
│  │  Clarity  │ 权重: 30%  得分: 4.0 │  │
│  │    ▲      │ Completeness     │  │
│  │  Role / Example               │  │
│  └───────────┴───────────────────┘  │
│  [可配置权重] [重新评分]             │
└─────────────────────────────────────┘
```

**实现要点**：
- 使用 `vue-chartjs` 的 `Radar` 图表
- 四维度独立评分滑块（1-5 分），权重可配置
- 综合分 = Σ(维度分 × 权重)
- 支持自定义权重模板（代码类、翻译类、创意类等）

---

## 三、状态管理

### 3.1 新增 Composable

#### `useTask.js` — 异步任务状态管理

```javascript
// 职责：管理批量测试、A/B 测试等异步任务的生命周期
// 状态：
//   - taskList: 任务列表（id, type, status, progress, created_at）
//   - currentTask: 当前查看的任务
//   - sseConnection: SSE 连接引用
//
// 方法：
//   - fetchTasks() — 获取任务列表
//   - fetchTask(taskId) — 获取单个任务详情
//   - startSSE(taskId) — 建立 SSE 连接，实时更新 progress
//   - stopSSE() — 断开 SSE 连接
//   - createBatchTest(promptId, cases, model) — 创建批量测试任务
//   - createABTest(promptId, variants, runs) — 创建 A/B 测试任务
```

#### `useQualityScore.js` — 质量评分管理

```javascript
// 职责：管理 Prompt 质量评分的四维度计算
// 状态：
//   - scores: { clarity, completeness, example, role }
//   - weights: { clarity: 0.3, completeness: 0.3, example: 0.25, role: 0.15 }
//   - totalScore: 计算属性，综合分
//
// 方法：
//   - fetchScore(promptId) — 获取已有评分
//   - submitScore(promptId, scores) — 提交评分
//   - setWeights(template) — 应用权重模板
//   - resetWeights() — 恢复默认权重
```

#### `useBatchTest.js` — 批量测试状态管理

```javascript
// 职责：管理批量测试用例和结果
// 状态：
//   - testCases: 测试用例列表 [{ variables: {...}, expected: "" }]
//   - results: 测试结果列表
//   - running: 是否正在运行
//   - currentIndex: 当前运行到第几个
//
// 方法：
//   - addCase(case) — 添加测试用例
//   - removeCase(index) — 删除测试用例
//   - runBatch(promptId, model) — 运行批量测试
//   - importFromFile(file) — 从 CSV/JSON 导入用例
```

### 3.2 现有 Composable 改造

| 文件 | 改动 |
|------|------|
| `useABTest.js` | 移除 mock 数据，对接真实 API（见 API 对接章节） |
| `useVariablePreview.js` | 新增 `highlightedContent` 的高亮 HTML 输出，供 VariablePreviewPanel 使用 |

---

## 四、API 对接方案

### 4.1 API 清单

#### 4.1.1 任务队列 API

| 方法 | 路径 | 描述 | 前端调用 |
|------|------|------|----------|
| POST | `/api/tasks` | 创建任务（批量测试/A/B测试/评测集生成） | `useTask.createTask()` |
| GET | `/api/tasks` | 获取任务列表 | `useTask.fetchTasks()` |
| GET | `/api/tasks/:id` | 获取任务详情（含结果） | `useTask.fetchTask()` |
| GET | `/api/tasks/:id/progress` | SSE 流式进度推送 | `useTask.startSSE()` |
| DELETE | `/api/tasks/:id` | 取消任务 | `useTask.cancelTask()` |

**POST /api/tasks 请求体**：
```json
{
  "type": "batch_test",
  "payload": {
    "prompt_id": 1,
    "cases": [
      { "name": "测试1", "variables": { "var1": "值1" } },
      { "name": "测试2", "variables": { "var1": "值2" } }
    ],
    "model": "gpt-4",
    "eval_criteria": ["clarity", "completeness"]
  }
}
```

```json
{
  "type": "ab_test",
  "payload": {
    "prompt_id": 1,
    "variants": [
      { "id": "a", "name": "原版", "content": "..." },
      { "id": "b", "name": "优化版", "content": "..." }
    ],
    "target_runs": 20,
    "min_runs": 15,
    "max_runs": 50
  }
}
```

**GET /api/tasks/:id/progress SSE 事件格式**：
```
event: progress
data: {"progress": 45, "status": "running", "runs_a": 6, "runs_b": 5}

event: complete
data: {"status": "done", "winner": "b", "significance": true, "confidence": 0.94}
```

#### 4.1.2 批量测试 API

| 方法 | 路径 | 描述 | 前端调用 |
|------|------|------|----------|
| POST | `/api/prompts/:id/batch-test` | 创建批量测试任务 | `useBatchTest.runBatch()` |
| GET | `/api/prompts/:id/batch-test` | 获取批量测试历史 | `useBatchTest.fetchHistory()` |
| GET | `/api/prompts/:id/batch-test/:taskId` | 获取任务结果 | `useBatchTest.fetchResult()` |

#### 4.1.3 A/B 测试 API

| 方法 | 路径 | 描述 | 前端调用 |
|------|------|------|----------|
| GET | `/api/ab-tests` | 获取 A/B 测试列表 | `useABTest.fetchAll()` |
| GET | `/api/ab-tests/:id` | 获取 A/B 测试详情 | `useABTest.fetchOne()` |
| POST | `/api/ab-tests` | 创建 A/B 测试 | `useABTest.create()` |
| DELETE | `/api/ab-tests/:id` | 删除 A/B 测试 | `useABTest.delete()` |
| GET | `/api/ab-tests/:id/progress` | SSE 实时进度 | `useABTest.startSSE()` |

#### 4.1.4 质量评分 API

| 方法 | 路径 | 描述 | 前端调用 |
|------|------|------|----------|
| GET | `/api/prompts/:id/score` | 获取质量评分 | `useQualityScore.fetchScore()` |
| POST | `/api/prompts/:id/score` | 提交/更新评分 | `useQualityScore.submitScore()` |
| POST | `/api/prompts/:id/score/auto` | AI 自动评分 | `useQualityScore.autoScore()` |

**评分请求体**：
```json
{
  "scores": {
    "clarity": 4.0,
    "completeness": 3.5,
    "example": 4.0,
    "role": 3.0
  },
  "weights": {
    "clarity": 0.3,
    "completeness": 0.3,
    "example": 0.25,
    "role": 0.15
  }
}
```

#### 4.1.5 评测集 API

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/prompts/:id/eval-set` | 创建评测集（AI 辅助 + 人工） |
| GET | `/api/prompts/:id/eval-set` | 获取评测集 |
| PUT | `/api/prompts/:id/eval-set` | 更新评测集 |
| DELETE | `/api/prompts/:id/eval-set` | 删除评测集 |

### 4.2 API 封装

在 `composables/useApi.js` 中新增：

```javascript
// 新增 API 封装
export const tasksApi = {
  create: (data) => api.post('/tasks', data),
  list: (params) => api.get('/tasks', { params }),
  get: (id) => api.get(`/tasks/${id}`),
  delete: (id) => api.delete(`/tasks/${id}`)
}

export const abTestsApi = {
  list: (params) => api.get('/ab-tests', { params }),
  get: (id) => api.get(`/ab-tests/${id}`),
  create: (data) => api.post('/ab-tests', data),
  delete: (id) => api.delete(`/ab-tests/${id}`)
}

export const scoresApi = {
  get: (promptId) => api.get(`/prompts/${promptId}/score`),
  submit: (promptId, data) => api.post(`/prompts/${promptId}/score`, data),
  auto: (promptId) => api.post(`/prompts/${promptId}/score/auto`)
}
```

### 4.3 SSE 封装

```javascript
// composables/useSSE.js
export function useSSE(url) {
  const data = ref(null)
  const error = ref(null)
  const status = ref('disconnected') // disconnected | connecting | connected | error

  let eventSource = null

  const connect = () => {
    status.value = 'connecting'
    eventSource = new EventSource(url)

    eventSource.onopen = () => { status.value = 'connected' }

    eventSource.addEventListener('progress', (e) => {
      data.value = JSON.parse(e.data)
    })

    eventSource.addEventListener('complete', (e) => {
      data.value = JSON.parse(e.data)
      status.value = 'disconnected'
    })

    eventSource.onerror = (e) => {
      status.value = 'error'
      error.value = e
    }
  }

  const disconnect = () => {
    eventSource?.close()
    status.value = 'disconnected'
  }

  onUnmounted(disconnect)

  return { data, error, status, connect, disconnect }
}
```

---

## 五、技术选型

### 5.1 无需引入新库

| 原因 | 说明 |
|------|------|
| **图表已有 vue-chartjs** | TestAnalytics 已引入，复用 Radar、Bar、Line 图表 |
| **表格已有 Element Plus** | el-table 支持 expand 行展开，足够批量测试需求 |
| **无 Pinia 需求** | Composable 模式已满足跨组件状态共享，无需引入 Pinia |
| **无 WebSocket 库需求** | 浏览器原生 EventSource API 即可支持 SSE |

### 5.2 技术决策

| 决策项 | 结论 | 理由 |
|--------|------|------|
| 状态管理 | 继续用 Composable | 小型团队，简单直观，无需 Pinia 学习成本 |
| 图表库 | vue-chartjs（已引入） | 雷达图 (Radar)、柱状图 (Bar)、折线图 (Line) 均支持 |
| 实时推送 | 原生 SSE | 轻量、单向推送足够，无需 WebSocket |
| 批量测试 UI | el-table expand | Element Plus 原生支持，样式统一 |
| 变量预览 | 新建 VariablePreviewPanel | 独立面板组件，60/40 分栏布局 |

### 5.3 依赖变更

无需安装任何新 npm 包。现有依赖已覆盖全部需求。

---

## 六、实施计划

### Phase 1: 变量实时预览（优先级 P0）

1. 改造 `VariablePreviewer.vue` → 支持 side-by-side 模式
2. 新建 `VariablePreviewPanel.vue`（60/40 分栏包装）
3. 改造 `PromptEditor.vue` 布局，集成 VariablePreviewPanel
4. 更新 `useVariablePreview.js`，增加高亮 HTML 输出
5. 单元测试

### Phase 2: 批量测试 UI（优先级 P0）

1. 新建 `BatchTest.vue` 视图（含测试用例编辑表单）
2. 新建 `BatchTestTable.vue` + `BatchTestCard.vue` 组件
3. 新建 `useBatchTest.js` composable
4. 新建 `useTask.js` composable + SSE 集成
5. 对接 `POST /api/prompts/:id/batch-test` + `GET /api/tasks/:id/progress`
6. 更新 `PromptTester.vue`，新增批量测试入口按钮
7. 单元测试

### Phase 3: A/B 测试中心（优先级 P0）

1. 更新 `useABTest.js`，移除 mock，对接真实 API
2. 新建 `ABTestSequentialPanel.vue` 组件
3. 改造 `ABTestList.vue` + `ABTestDetail.vue` 对接真实 API
4. 集成 `useTask.js` 的 SSE 实现
5. 更新 `ABTestCompare.vue` 集成 `QualityScoreCard`
6. 单元测试

### Phase 4: 质量评分卡（优先级 P1）

1. 新建 `QualityScoreCard.vue` 组件（雷达图 + 权重配置）
2. 新建 `useQualityScore.js` composable
3. 集成到 `ABTestCompare.vue` 的 Variant 卡片详情中
4. 集成到 `PromptEditor.vue` 侧边栏
5. 单元测试

### Phase 5: 路由与导航整合（优先级 P1）

1. 更新 `router/index.js`，添加 `/prompts/:id/batch-test` 路由
2. 更新 `TopNav.vue`，在 Prompt 详情页添加入口按钮
3. 更新面包屑导航组件 `BreadcrumbNav.vue`
4. 端到端测试

---

## 七、关键风险与缓解

| 风险 | 级别 | 缓解方案 |
|------|------|----------|
| SSE 断线重连 | 中 | `useSSE` 实现自动重连逻辑（3 次重试，指数退避） |
| 批量测试 Token 消耗 | 高 | 前端显示预估消耗，确认后再执行；分批执行避免单次超时 |
| 序贯检验进度显示 | 中 | SSE 推送完整状态，前端只做渲染，不做计算 |
| Element Plus el-table 展开行性能 | 低 | 虚拟滚动（el-table-v2）可选优化，当前规模不需要 |
| 质量评分主观性 | 低 | 提供 AI 自动评分 + 人工评分双模式，用户可选择 |

---

## 八、文件变更汇总

| 操作 | 文件路径 | 数量 |
|------|----------|------|
| 新建 | `frontend/src/components/VariablePreviewPanel.vue` | 1 |
| 新建 | `frontend/src/components/BatchTestTable.vue` | 1 |
| 新建 | `frontend/src/components/BatchTestCard.vue` | 1 |
| 新建 | `frontend/src/components/QualityScoreCard.vue` | 1 |
| 新建 | `frontend/src/components/TaskProgressBar.vue` | 1 |
| 新建 | `frontend/src/components/ABTestSequentialPanel.vue` | 1 |
| 新建 | `frontend/src/views/BatchTest.vue` | 1 |
| 新建 | `frontend/src/composables/useTask.js` | 1 |
| 新建 | `frontend/src/composables/useQualityScore.js` | 1 |
| 新建 | `frontend/src/composables/useBatchTest.js` | 1 |
| 新建 | `frontend/src/composables/useSSE.js` | 1 |
| 修改 | `frontend/src/components/VariablePreviewer.vue` | 1 |
| 修改 | `frontend/src/views/PromptEditor.vue` | 1 |
| 修改 | `frontend/src/views/PromptTester.vue` | 1 |
| 修改 | `frontend/src/views/ABTestList.vue` | 1 |
| 修改 | `frontend/src/views/ABTestDetail.vue` | 1 |
| 修改 | `frontend/src/components/ABTestCompare.vue` | 1 |
| 修改 | `frontend/src/composables/useABTest.js` | 1 |
| 修改 | `frontend/src/composables/useVariablePreview.js` | 1 |
| 修改 | `frontend/src/composables/useApi.js` | 1 |
| 修改 | `frontend/src/router/index.js` | 1 |
| **合计** | | **22 个文件** |
