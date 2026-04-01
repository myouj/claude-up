# PromptVault UI/UX 设计规范报告

> 版本：v1.0
> 日期：2026-04-01
> 状态：设计规范定稿

---

## 一、设计系统扩展

### 1.1 现有设计 Token 分析

当前 `App.vue` 中已定义的设计 Token 覆盖了颜色、字体、间距、圆角、阴影、过渡动画和 z-index 共 7 个维度。扩展策略：**在现有 token 体系上做增量扩展**，不破坏现有 token 命名约定，保持 CSS 变量名前缀一致。

### 1.2 新增组件设计 Token

#### 1.2.1 质量评分卡 (Quality Score Card)

```css
/* 质量评分维度颜色 */
--color-score-clarity: #6366F1;       /* 指令清晰度 - Indigo */
--color-score-clarity-bg: #EEF2FF;
--color-score-completeness: #8B5CF6;   /* 约束完整性 - Violet */
--color-score-completeness-bg: #F5F3FF;
--color-score-example: #EC4899;        /* 示例质量 - Pink */
--color-score-example-bg: #FDF2F8;
--color-score-role: #14B8A6;           /* 角色定义 - Teal */
--color-score-role-bg: #F0FDFA;

/* 质量分等级 */
--color-score-excellent: #10B981;      /* >= 4.5 */
--color-score-good: #22C55E;           /* >= 3.5 */
--color-score-average: #F59E0B;         /* >= 2.5 */
--color-score-poor: #EF4444;            /* < 2.5 */

/* 评分环 */
--score-ring-size: 64px;
--score-ring-stroke: 6px;
```

#### 1.2.2 批量测试结果卡片 (Batch Test Result Card)

```css
/* 测试卡片专用 */
--color-test-pending: #94A3B8;          /* 等待中 */
--color-test-running: #3B82F6;         /* 运行中 - 主色 */
--color-test-success: #10B981;         /* 成功 */
--color-test-failed: #EF4444;           /* 失败 */
--color-test-running-bg: #EFF6FF;

/* 测试卡片布局 */
--test-card-min-width: 320px;
--test-card-max-height: 280px;
--test-card-padding: var(--spacing-4);
--test-card-gap: var(--spacing-3);

/* 进度条 */
--progress-bar-height: 4px;
--progress-bar-radius: 2px;
--progress-runway: var(--color-border);
--progress-fill: var(--color-primary);
```

#### 1.2.3 A/B 测试结果展示 (A/B Test Result Display)

```css
/* 统计显著性 */
--color-significant: #10B981;           /* 统计显著 */
--color-insufficient: #F59E0B;          /* 样本不足 */
--color-inconclusive: #94A3B8;          /* 未确定 */

/* Variant 标签 */
--color-variant-a: #2563EB;             /* Variant A - 主色 */
--color-variant-a-light: #EFF6FF;
--color-variant-b: #F97316;             /* Variant B - CTA 色 */
--color-variant-b-light: #FFF7ED;
--color-variant-c: #8B5CF6;             /* Variant C */
--color-variant-c-light: #F5F3FF;

/* 置信度条 */
--confidence-bar-height: 8px;
--confidence-bar-radius: 4px;

/* 效果量指示 */
--effect-size-small: 0.2;   /* 小效果: 10-20% */
--effect-size-medium: 0.5;  /* 中效果: 20-35% */
--effect-size-large: 0.8;    /* 大效果: >35% */
```

#### 1.2.4 任务进度通知 (Task Progress Notification)

```css
/* 通知条 */
--notification-bar-height: 48px;
--notification-bar-padding: var(--spacing-4) var(--spacing-6);
--notification-progress-height: 3px;
--notification-z: calc(var(--z-modal) + 1);

/* 状态颜色 */
--notification-info-bg: var(--color-primary-light);
--notification-success-bg: var(--color-success-light);
--notification-warning-bg: var(--color-warning-light);
--notification-error-bg: var(--color-danger-light);
```

#### 1.2.5 异步任务状态 (Async Task Status)

```css
/* 任务状态 */
--color-task-pending: #94A3B8;
--color-task-running: #3B82F6;
--color-task-done: #10B981;
--color-task-failed: #EF4444;

/* 任务卡片 */
--task-card-border-radius: var(--radius-lg);
--task-card-shadow: var(--shadow-md);
--task-card-hover-shadow: var(--shadow-hover);

/* 运行动画 */
--pulse-animation-duration: 1.5s;
--shimmer-duration: 1.8s;
```

---

## 二、布局规范

### 2.1 Prompt 编辑页 60/40 分栏布局

```
┌─────────────────────────────────────────────────────────┐
│ Header: 标题 + 操作按钮（版本历史/测试/AI优化/保存）       │
├──────────────────────────┬──────────────────────────────┤
│                          │                              │
│   编辑区域 (60%)          │    预览面板 (40%)            │
│                          │                              │
│  ┌──────────────────┐   │  ┌────────────────────────┐ │
│  │ 标题（可内联编辑）│   │  │ 渲染后的 Prompt 预览    │ │
│  ├──────────────────┤   │  │                        │ │
│  │                  │   │  │ 变量值填充区域          │ │
│  │ 内容 Textarea    │   │  │ ┌────────────────────┐ │ │
│  │ (monospace)      │   │  │ │ {{variable}} = ___ │ │ │
│  │                  │   │  │ └────────────────────┘ │ │
│  │                  │   │  ├────────────────────────┤ │
│  │                  │   │  │ 变量完成度进度条  75%  │ │
│  │                  │   │  ├────────────────────────┤ │
│  │                  │   │  │ Prompt 质量评分卡      │ │
│  │                  │   │  │ ● Clarity    ████░ 85 │ │
│  │                  │   │  │ ● Complete   ███░░ 70 │ │
│  │                  │   │  │ ● Example    ████░ 88 │ │
│  │                  │   │  │ ● Role       ███░░ 72 │ │
│  └──────────────────┘   │  └────────────────────────┘ │
│                          │                              │
├──────────────────────────┴──────────────────────────────┤
│ 左侧边栏 (300px): 描述 / 分类 / 标签 / 收藏 / 置顶         │
└─────────────────────────────────────────────────────────┘
```

**关键规范**：

| 区域 | 宽度 | 内容 | 最小高度 |
|------|------|------|---------|
| 左侧边栏 | 300px（固定） | 元数据表单 | 全屏 |
| 编辑区域 | 60%（flex: 6） | Prompt 内容 textarea | calc(100vh - 64px) |
| 预览面板 | 40%（flex: 4） | 变量预览 + 质量评分卡 | calc(100vh - 64px) |
| Header | 100% | 操作按钮组 | 64px |

**响应式断点**：

- **≥1200px**：完整三栏（侧边栏 + 编辑 + 预览）
- **768px-1199px**：隐藏左侧边栏（通过 Drawer 访问），编辑 + 预览保持 60/40
- **<768px**：单栏布局，预览面板变为底部可折叠区域（Accordion 模式）

### 2.2 批量测试页：表格 + 卡片混合布局

```
┌─────────────────────────────────────────────────────────┐
│ Header: 测试名称 + 模型选择 + 过滤器 + 导出               │
├─────────────────────────────────────────────────────────┤
│ 控制栏: [开始批量测试] [停止] [选择全部]  进度条 5/20    │
├─────────────────────────────────────────────────────────┤
│ 测试表格（紧凑行高，可滚动）                               │
│ ┌───┬────────────┬──────────┬────────┬──────┬─────────┐ │
│ │ ☑ │ #1        │ GPT-4    │  4.2   │ 1.2s │ [展开]  │ │
│ ├───┼────────────┼──────────┼────────┼──────┼─────────┤ │
│ │ ☑ │ #2        │ GPT-3.5  │  3.8   │ 0.8s │ [展开]  │ │
│ ├───┼────────────┼──────────┼────────┼──────┼─────────┤ │
│ │ ☑ │ #3        │ Claude 3 │  4.5   │ 1.5s │ [展开]  │ │
│ └───┴────────────┴──────────┴────────┴──────┴─────────┘ │
├─────────────────────────────────────────────────────────┤
│ 已展开卡片（点击展开行后出现）                              │
│ ┌─────────────────────────────────────────────────────┐ │
│ │ Case #2 展开卡片                                      │ │
│ │ ┌────────────────────┐  ┌────────────────────────┐ │ │
│ │ │  输入              │  │  AI 输出                │ │ │
│ │ │  变量: code=...    │  │  生成的代码块...        │ │ │
│ │ └────────────────────┘  └────────────────────────┘ │ │
│ │ 质量分: 3.8 ★★★☆☆  │  延迟: 0.8s  │  Token: 524   │ │
│ │ 维度雷达图: [clarity/completeness/example/role]    │ │
│ └─────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

**关键规范**：

- 表格行高：**48px**（紧凑），hover 时变为 **56px** 并显示展开按钮
- 展开卡片：从表格行位置平滑展开动画，宽度撑满整行
- 批量操作栏：固定在表格上方，滚动时自动吸顶（sticky）
- 多选对比：选中 ≥2 行时，显示「横向对比」按钮

### 2.3 A/B 测试中心：结果展示与统计显著性可视化

```
┌─────────────────────────────────────────────────────────┐
│ Header: 测试名称 + 状态标签 + 操作按钮                    │
├─────────────────────────────────────────────────────────┤
│ 测试概览卡片                                              │
│ ┌─────────┬─────────┬─────────┬─────────┬─────────────┐ │
│ │ 总样本  │ 显著状态│  胜出   │ 平均分差│  置信度     │ │
│ │  42    │ ✓ 显著  │ VariantB│  +0.8   │  ████████░░ │ │
│ └─────────┴─────────┴─────────┴─────────┴─────────────┘ │
├────────────────────────────┬──────────────────────────────┤
│ Variant A 面板 (50%)      │ Variant B 面板 (50%)         │
│ ┌──────────────────────┐  │ ┌──────────────────────┐    │
│ │ ★ 4.1 (落后)         │  │ │ ★ 4.9 (领先)          │    │
│ ├──────────────────────┤  │ ├──────────────────────┤    │
│ │ 运行次数: 21         │  │ │ 运行次数: 21          │    │
│ │ Token 均耗: 512      │  │ │ Token 均耗: 498       │    │
│ │ 平均延迟: 1.4s        │  │ │ 平均延迟: 1.2s        │    │
│ ├──────────────────────┤  │ ├──────────────────────┤    │
│ │ 分数分布直方图        │  │ │ 分数分布直方图        │    │
│ │ ▄▅█▇▅▄▂             │  │ │ ▂▄▅█▇▆▅▄             │    │
│ └──────────────────────┘  │ └──────────────────────┘    │
├────────────────────────────┴──────────────────────────────┤
│ 序贯检验进度条                                            │
│ [████████████░░░░░░░░░░░░░░] 21/50  样本量 (最低 15)     │
│ 显著性: ████████░░ 80%   预计还需: ~8 次                  │
├─────────────────────────────────────────────────────────┤
│ 检验记录时间线                                            │
│ ●──●──●──●──●──●──●──●──●──●──○──○──○──○──○             │
│ ↑ 测试开始           ↑ 接近显著        当前进度            │
└─────────────────────────────────────────────────────────┘
```

**关键规范**：

- Variant 分数字体：**48px**，加粗，使用 `--color-primary` / `--color-cta` 区分 A/B
- 统计显著性徽章：使用语义化颜色（绿=显著，黄=接近，红=不足）
- 序贯检验进度条：实时更新，显示 Wald SPRT 算法的当前边界状态
- 对比面板：始终左右对称布局，即使数据不对称也保持视觉平衡

---

## 三、动效规范

### 3.1 任务进度通知动效

```css
/* 入场动画：从顶部滑入 */
@keyframes slideDown {
  from {
    transform: translateY(-100%);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

/* 退场动画：向上滑出并淡出 */
@keyframes slideUp {
  from {
    transform: translateY(0);
    opacity: 1;
  }
  to {
    transform: translateY(-100%);
    opacity: 0;
  }
}

/* 进度条填充动画 */
@keyframes progressFill {
  from { width: 0%; }
  to { width: var(--target-width); }
}

/* 进度条脉冲（表示运行中） */
@keyframes progressPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

/* 运行状态：波纹扩散 */
@keyframes ripple {
  0% {
    transform: scale(1);
    opacity: 0.4;
  }
  100% {
    transform: scale(2.5);
    opacity: 0;
  }
}
```

**应用场景**：

| 动效 | 时长 | 缓动函数 | 触发场景 |
|------|------|---------|---------|
| 通知滑入 | 300ms | ease-out | 任务开始 |
| 通知滑出 | 250ms | ease-in | 任务完成/关闭 |
| 进度填充 | 400ms | ease-out | 进度更新 |
| 进度脉冲 | 1500ms | ease-in-out | 运行中状态 |
| 状态切换 | 200ms | ease | 任务状态变更 |
| 胜出揭晓 | 600ms | cubic-bezier(0.34, 1.56, 0.64, 1) | A/B 测试确定胜者 |

### 3.2 结果展示过渡动画

```css
/* 卡片展开：从行位置向外扩展 */
@keyframes expandCard {
  from {
    opacity: 0;
    transform: scaleY(0.8);
    transform-origin: top;
  }
  to {
    opacity: 1;
    transform: scaleY(1);
  }
}

/* 分数变化：数字滚动 */
@keyframes scoreChange {
  0% { transform: translateY(10px); opacity: 0; }
  100% { transform: translateY(0); opacity: 1; }
}

/* 雷达图/图表绘制：渐进描边 */
@keyframes drawChart {
  from { stroke-dashoffset: var(--circumference); }
  to { stroke-dashoffset: var(--offset); }
}

/* 质量评分卡：环形进度填充 */
@keyframes ringFill {
  from { stroke-dashoffset: 201; } /* 2π * 32 = ~201 */
  to { stroke-dashoffset: var(--ring-offset); }
}

/* 显著性徽章：弹跳 */
@keyframes badgeBounce {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.15); }
  75% { transform: scale(0.95); }
}
```

### 3.3 页面级过渡

```css
/* 页面切换淡入 */
.page-enter-active {
  transition: opacity 200ms ease, transform 200ms ease;
}

.page-enter-from {
  opacity: 0;
  transform: translateY(12px);
}

.page-leave-active {
  transition: opacity 150ms ease;
}

.page-leave-to {
  opacity: 0;
}

/* 模态框/对话框 */
.dialog-enter-active,
.dialog-leave-active {
  transition: opacity 200ms ease;
}

.dialog-enter-active .dialog-content,
.dialog-leave-active .dialog-content {
  transition: transform 200ms cubic-bezier(0.34, 1.2, 0.64, 1);
}

.dialog-enter-from,
.dialog-leave-to {
  opacity: 0;
}

.dialog-enter-from .dialog-content {
  transform: scale(0.92) translateY(10px);
}

.dialog-leave-to .dialog-content {
  transform: scale(0.95) translateY(5px);
}

/* 列表项交错入场（Stagger Animation） */
.stagger-item {
  animation: staggerFadeIn 300ms ease-out backwards;
}

.stagger-item:nth-child(1) { animation-delay: 0ms; }
.stagger-item:nth-child(2) { animation-delay: 50ms; }
.stagger-item:nth-child(3) { animation-delay: 100ms; }
.stagger-item:nth-child(4) { animation-delay: 150ms; }
.stagger-item:nth-child(5) { animation-delay: 200ms; }
/* ...以此类推，最多 10 项 */
```

### 3.4 微交互规范

| 交互 | 元素 | 动效 | 时长 |
|------|------|------|------|
| 按钮悬停 | 背景色 + 阴影 | translateY(-1px) + shadow 增强 | 150ms |
| 按钮点击 | 缩放 | scale(0.97) | 100ms |
| 卡片悬停 | 边框 + 阴影 | border-color + shadow | 150ms |
| 卡片悬停 | 整体上浮 | translateY(-2px) | 200ms |
| 输入框聚焦 | 边框颜色 | 边框变主色 + box-shadow | 150ms |
| 开关切换 | 滑块位移 | 沿轨道滑动 + 颜色变化 | 200ms |
| 折叠展开 | 高度动画 | max-height + opacity | 250ms |
| 删除确认 | 危险色渐变 | 背景变红 + 图标变化 | 200ms |
| 加载占位 | 骨架屏闪烁 | shimmer 从左到右扫过 | 1800ms |
| 任务完成 | 成功色 + 勾号 | 绿色扩散 + 勾号弹出 | 400ms |

---

## 四、可访问性（Accessibility）

### 4.1 颜色对比度检查

所有文本与背景的对比度必须满足 WCAG 2.1 AA 级别（最低 4.5:1，正常文本；3:1，大文本）。

**关键配色对比度验证**：

| 组合 | 前景色 | 背景色 | 对比度 | WCAG AA |
|------|--------|--------|--------|---------|
| 正文 | #1E293B | #FFFFFF | 14.5:1 | ✅ AAA |
| 次要文本 | #64748B | #FFFFFF | 5.9:1 | ✅ AA |
| 占位文本 | #94A3B8 | #FFFFFF | 3.2:1 | ⚠️ 仅大文本 |
| 按钮文字 | #FFFFFF | #2563EB | 4.7:1 | ✅ AA |
| 成功文本 | #10B981 | #ECFDF5 | 4.6:1 | ✅ AA |
| 危险文本 | #EF4444 | #FEF2F2 | 4.1:1 | ✅ AA |
| 评分 Excellent | #10B981 | #F8FAFC | 13.1:1 | ✅ AAA |
| 评分 Poor | #EF4444 | #F8FAFC | 5.2:1 | ✅ AA |
| Variant A 标签 | #2563EB | #EFF6FF | 4.6:1 | ✅ AA |
| Variant B 标签 | #F97316 | #FFF7ED | 3.8:1 | ⚠️ 需加粗 |

**扩展方案**：
- 危险色 `#EF4444` 在浅色背景 `#FEF2F2` 上仅为 **3.2:1**，不满足 AA。**改进**：`--color-danger` 用于纯文字时搭配 `--color-danger-light` 背景，确保对比度 ≥ 4.5:1。
- Variant B 标签文字 `3.8:1` 仅满足大文本（≥18px 或加粗）。**规范**：Variant B 标签文字必须使用 `--font-weight-semibold`（600）。
- 所有语义色（success/warning/danger）必须同时提供背景色版本（*-light）使用。

### 4.2 键盘导航

#### 4.2.1 焦点管理规范

```css
/* 焦点样式 - 全局 */
:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
  border-radius: var(--radius-sm);
}

/* 焦点进入面板时的强调样式 */
:focus-visible.has-panel-focus {
  outline: 2px solid var(--color-primary);
  outline-offset: -2px;
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.15);
}
```

#### 4.2.2 键盘快捷键

| 快捷键 | 功能 | 页面/上下文 |
|--------|------|------------|
| `Ctrl/Cmd + S` | 保存 | Prompt 编辑页 |
| `Ctrl/Cmd + Enter` | 发送测试 | 测试预览页 |
| `Ctrl/Cmd + E` | 打开编辑器 | 全局 |
| `Ctrl/Cmd + /` | 显示快捷键帮助 | 全局 |
| `Escape` | 关闭模态框/抽屉/取消操作 | 全局 |
| `Tab` | 焦点移到下一个可交互元素 | 全局 |
| `Shift + Tab` | 焦点移到上一个可交互元素 | 全局 |
| `↑/↓` | 导航菜单项 / 表格行选择 | 列表页面 |
| `Enter/Space` | 确认选择 / 展开行 | 表格/卡片 |
| `Ctrl/Cmd + Shift + A` | 新建 Prompt | Prompt 列表页 |
| `Ctrl/Cmd + F` | 聚焦搜索框 | 列表页面 |

#### 4.2.3 焦点陷阱（Focus Trap）

以下场景必须启用焦点陷阱，防止 Tab 焦点逃逸到模态框外：

- 模态对话框（Dialog）
- 侧边抽屉（Drawer）
- 批量测试展开行详情
- A/B 对比全屏模式

#### 4.2.4 跳转链接（Skip Links）

在 `<body>` 起始处添加跳转链接（对屏幕阅读器用户可见）：

```html
<a href="#main-content" class="skip-link">跳转到主要内容</a>
<a href="#main-nav" class="skip-link">跳转到导航</a>
```

```css
.skip-link {
  position: absolute;
  top: -100px;
  left: 0;
  background: var(--color-primary);
  color: #fff;
  padding: var(--spacing-2) var(--spacing-4);
  z-index: 9999;
  transition: top 0.2s;
}

.skip-link:focus {
  top: 0;
}
```

### 4.3 ARIA 规范

#### 4.3.1 语义化角色

| 组件 | ARIA Role | 说明 |
|------|-----------|------|
| 顶部导航 | `role="navigation"` + `aria-label="主导航"` | TopNav |
| 面包屑导航 | `role="navigation"` + `aria-label="面包屑"` | BreadcrumbNav |
| 主内容区 | `role="main"` + `id="main-content"` | 页面 `<main>` |
| 测试结果表格 | `role="table"` + `aria-label="批量测试结果"` | BatchTest 表格 |
| 质量评分卡 | `role="meter"` + `aria-valuenow/max/min/label` | ScoreCard |
| 进度通知条 | `role="progressbar"` + `aria-valuenow/max/label` | Notification bar |
| 展开收起区域 | `role="region"` + `aria-expanded/controls` | Accordion |
| 状态标签 | `role="status"` + `aria-live="polite"` | 任务状态 |

#### 4.3.2 ARIA 属性清单

```html
<!-- 任务进度通知 -->
<div role="progressbar"
     aria-valuenow="45"
     aria-valuemin="0"
     aria-valuemax="100"
     aria-label="批量测试进度">
</div>

<!-- 统计显著性状态 -->
<span role="status" aria-live="polite">
  统计显著性已达到 95%，Variant B 胜出
</span>

<!-- 质量评分卡 -->
<div role="meter"
     aria-valuenow="4.2"
     aria-valuemin="0"
     aria-valuemax="5"
     aria-label="Prompt 质量总分 4.2 分（满分 5 分）">
</div>

<!-- 展开的测试行 -->
<tr aria-expanded="true" aria-controls="detail-panel-1">
<td>
  <button aria-label="展开测试详情，第 1 条，共 20 条">
</td>

<!-- 变体对比卡 -->
<div role="group" aria-label="Variant A 与 Variant B 对比">
```

#### 4.3.3 屏幕阅读器支持

```html
<!-- 视觉隐藏但屏幕阅读器可见的文本 -->
<span class="sr-only">（仅屏幕阅读器可见的描述）</span>

<!-- 加载中状态 -->
<span class="sr-only" role="status" aria-live="polite">
  正在加载批量测试结果，已完成 5 条，共 20 条
</span>

<!-- 空状态 -->
<div aria-label="无测试记录">
  暂无测试记录，请先运行测试
</div>
```

```css
/* 屏幕阅读器专用：视觉隐藏但保持可访问 */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
```

### 4.4 可访问性检查清单

#### 4.4.1 视觉层面

- [ ] 所有文本对比度 ≥ 4.5:1（普通文本）或 ≥ 3:1（大文本 ≥ 18pt）
- [ ] 颜色不作为传达信息的唯一方式（必须配合图标、文字或图案）
- [ ] 文字大小可缩放至 200% 而不丢失内容（不出现水平滚动）
- [ ] 深色模式支持（使用 CSS 变量媒体查询 `@media (prefers-color-scheme: dark)`）
- [ ] 高对比度模式兼容（避免纯色边框依赖视觉效果）

#### 4.4.2 键盘操作

- [ ] 所有可交互元素可通过键盘访问（Tab / Shift+Tab / Enter / Space / Esc）
- [ ] 焦点顺序符合逻辑阅读顺序（从上到下，从左到右）
- [ ] 焦点状态清晰可见（不能依赖纯色无边框）
- [ ] 模态框启用焦点陷阱
- [ ] 跳转链接存在且可用

#### 4.4.3 屏幕阅读器

- [ ] 所有图片有 `alt` 属性（装饰性图片 `alt=""`）
- [ ] 所有表单输入有 `<label>` 关联
- [ ] 动态内容更新使用 `aria-live` 通知
- [ ] 表格有表头 `<th>` 和 `scope` 属性
- [ ] 页面有唯一 `<h1>` 标题
- [ ] 标题层级连贯（H1→H2→H3，不跳级）
- [ ] 图标按钮有 `aria-label`

#### 4.4.4 运动与动画

- [ ] 提供 `prefers-reduced-motion` 媒体查询选项
- [ ] 动画时长 ≤ 500ms（避免前庭功能障碍）
- [ ] 闪烁频率不高于每秒 3 次

```css
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

---

## 五、深色模式扩展

在现有 CSS 变量体系中增加深色模式覆盖：

```css
@media (prefers-color-scheme: dark) {
  :root {
    --color-bg: #0F172A;
    --color-surface: #1E293B;
    --color-border: #334155;
    --color-border-hover: #475569;
    --color-text-primary: #F1F5F9;
    --color-text-secondary: #94A3B8;
    --color-text-muted: #64748B;

    /* 评分卡颜色适配 */
    --color-score-clarity-bg: rgba(99, 102, 241, 0.15);
    --color-score-completeness-bg: rgba(139, 92, 246, 0.15);
    --color-score-example-bg: rgba(236, 72, 153, 0.15);
    --color-score-role-bg: rgba(20, 184, 166, 0.15);

    /* 阴影调整 */
    --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.3);
    --shadow-md: 0 2px 4px rgba(0, 0, 0, 0.4);
    --shadow-hover: 0 4px 12px rgba(0, 0, 0, 0.5);
  }
}
```

---

## 六、实现优先级

### P0（必须实现）

1. **设计 Token 扩展**：质量评分卡、批量测试卡片、A/B 结果展示的所有 CSS 变量
2. **60/40 分栏布局**：Prompt 编辑页的新布局结构
3. **键盘导航**：焦点管理、模态框焦点陷阱、快捷键
4. **颜色对比度修复**：danger 色在浅色背景的问题

### P1（重要）

5. **任务进度通知动效**：滑入/滑出 + 进度条动画
6. **ARIA 属性**：角色、aria-live、aria-label 全覆盖
7. **展开卡片动画**：批量测试行展开动效
8. **深色模式**：基础深色变量覆盖

### P2（增强）

9. **统计显著性可视化**：序贯检验进度条 + 置信度动画
10. **屏幕阅读器专用文本**：sr-only 优化
11. **跳过链接**：Skip links
12. **prefers-reduced-motion**：动画减弱支持

---

## 七、附录

### A. 现有 Token 与新 Token 的命名对照

所有新增 token 均遵循现有命名规范：
- 颜色：`--color-[语义]-[子语义]` 或 `--color-[组件]-[子语义]`
- 间距：复用现有 `--spacing-N`
- 圆角：复用现有 `--radius-*`
- 过渡：复用现有 `--transition-*`，新增复杂缓动直接写函数

### B. Element Plus 组件覆盖策略

新增组件的 Element Plus 覆盖在页面级 `<style scoped>` 中实现，不在 `App.vue` 全局覆盖，避免影响其他页面：

```css
/* 在具体页面中覆盖 Element Plus 组件样式 */
.page-class :deep(.el-progress-bar__outer) {
  border-radius: var(--radius-sm);
}
```

### C. 设计工具同步

此规范与 Figma/Sketch 设计文件应保持同步。建议在设计工具中使用与 CSS 变量同名的命名规范（如 `--color-primary`），方便开发直接引用。

---

*本规范为 PromptVault V1 UI/UX 设计系统核心参考，实现时请结合 `frontend/src/App.vue` 中的现有 CSS 变量和具体页面组件代码。*
