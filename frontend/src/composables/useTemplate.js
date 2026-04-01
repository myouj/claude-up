import { ref } from 'vue'

export const mockTemplates = ref([
  {
    id: 1,
    name: '代码审查专家',
    category: 'development',
    description: '专业的代码审查助手，自动检测潜在问题、安全漏洞和性能优化点。',
    content: '你是一位资深代码审查专家。请按以下结构审查代码：\n\n## 优点\n- 列出代码的积极方面\n\n## 问题\n- 列出发现的问题（严重程度：高/中/低）\n\n## 建议\n- 提供具体的改进建议',
    author: { name: '张明', avatar: '张' },
    score: 4.8,
    installs: 1243,
    tags: ['code-review', 'security', 'quality'],
    preview_image: 'code-review',
    created_at: '2026-01-10T10:00:00Z',
    comments: [
      {
        id: 1,
        author: { name: '李华', avatar: '李' },
        content: '非常好用的模板！帮我发现了几个隐藏的安全问题。',
        score: 5,
        created_at: '2026-02-15T14:30:00Z'
      },
      {
        id: 2,
        author: { name: '王芳', avatar: '王' },
        content: '结构清晰，但希望增加更多 OWASP 相关的检查项。',
        score: 4,
        created_at: '2026-02-20T09:15:00Z'
      }
    ]
  },
  {
    id: 2,
    name: 'SQL 生成助手',
    category: 'data',
    description: '根据自然语言描述自动生成高效的 SQL 查询语句。',
    content: '根据用户需求生成 SQL 查询。\n需求：{{requirement}}\n数据库类型：{{db_type}}',
    author: { name: '李华', avatar: '李' },
    score: 4.5,
    installs: 892,
    tags: ['sql', 'database', 'generator'],
    preview_image: 'sql-gen',
    created_at: '2026-01-20T10:00:00Z',
    comments: []
  },
  {
    id: 3,
    name: '技术文档撰写',
    category: 'docs',
    description: '自动生成结构清晰、技术准确的 API 文档和 README。',
    content: '你是一位技术文档专家。请为以下代码生成文档：\n\n## 功能概述\n描述主要功能\n\n## 参数说明\n列出所有参数\n\n## 示例\n提供使用示例',
    author: { name: '王芳', avatar: '王' },
    score: 4.3,
    installs: 567,
    tags: ['documentation', 'api', 'readme'],
    preview_image: 'docs',
    created_at: '2026-02-01T10:00:00Z',
    comments: [
      {
        id: 3,
        author: { name: '赵伟', avatar: '赵' },
        content: '生成的文档格式很规范，省了我很多时间！',
        score: 5,
        created_at: '2026-02-25T16:45:00Z'
      }
    ]
  },
  {
    id: 4,
    name: 'Bug 修复助手',
    category: 'debug',
    description: '分析错误信息，提供根因分析和修复建议。',
    content: '分析以下错误信息，提供根因分析和修复建议：\n\n错误信息：\n{{error}}\n\n代码上下文：\n{{code}}',
    author: { name: '赵伟', avatar: '赵' },
    score: 4.6,
    installs: 721,
    tags: ['debug', 'bug', 'fix'],
    preview_image: 'debug',
    created_at: '2026-02-05T10:00:00Z',
    comments: []
  },
  {
    id: 5,
    name: '产品需求分析',
    category: 'product',
    description: '将模糊的需求描述转化为结构化的产品需求文档。',
    content: '分析以下需求，生成结构化的 PRD：\n\n需求：{{requirement}}\n\n请提供：\n1. 问题陈述\n2. 目标用户\n3. 功能需求\n4. 非功能需求\n5. 验收标准',
    author: { name: '陈静', avatar: '陈' },
    score: 4.2,
    installs: 345,
    tags: ['product', 'prd', 'analysis'],
    preview_image: 'product',
    created_at: '2026-02-10T10:00:00Z',
    comments: []
  },
  {
    id: 6,
    name: 'Git Commit 生成器',
    category: 'git',
    description: '根据代码变更自动生成符合规范的 Git Commit 消息。',
    content: '根据以下代码变更生成 Git Commit 消息：\n\n变更：\n{{diff}}\n\n要求：\n- 使用 Conventional Commits 格式\n- 不超过 72 字符\n- 清晰描述变更内容',
    author: { name: '周杰', avatar: '周' },
    score: 4.9,
    installs: 2108,
    tags: ['git', 'commit', 'automation'],
    preview_image: 'git',
    created_at: '2026-01-25T10:00:00Z',
    comments: [
      {
        id: 4,
        author: { name: '吴婷', avatar: '吴' },
        content: '完美！生成的 commit 信息非常规范，团队都爱用。',
        score: 5,
        created_at: '2026-03-01T11:20:00Z'
      }
    ]
  },
  {
    id: 7,
    name: '测试用例生成',
    category: 'testing',
    description: '根据函数签名和文档自动生成单元测试用例。',
    content: '为以下函数生成单元测试：\n\n函数：\n{{function}}\n\n要求：\n- 使用 Jest 语法\n- 覆盖主要路径和边界情况\n- 包含 mock 示例',
    author: { name: '吴婷', avatar: '吴' },
    score: 4.4,
    installs: 489,
    tags: ['testing', 'unit-test', 'jest'],
    preview_image: 'testing',
    created_at: '2026-02-15T10:00:00Z',
    comments: []
  },
  {
    id: 8,
    name: '代码翻译',
    category: 'translation',
    description: '在多种编程语言之间转换代码，同时保持逻辑一致。',
    content: '将以下 {{source_lang}} 代码翻译为 {{target_lang}}：\n\n源代码：\n{{code}}\n\n注意保持原始逻辑和注释。',
    author: { name: '张明', avatar: '张' },
    score: 3.9,
    installs: 267,
    tags: ['translation', 'conversion', 'multilang'],
    preview_image: 'translation',
    created_at: '2026-02-20T10:00:00Z',
    comments: []
  }
])

export const templateCategories = ref([
  { label: '全部', value: 'all', icon: 'Grid' },
  { label: '开发', value: 'development', icon: 'Code' },
  { label: '数据', value: 'data', icon: 'DataAnalysis' },
  { label: '文档', value: 'docs', icon: 'Document' },
  { label: '调试', value: 'debug', icon: 'Bug' },
  { label: '产品', value: 'product', icon: 'Goods' },
  { label: 'Git', value: 'git', icon: 'Branch' },
  { label: '测试', value: 'testing', icon: 'CircleCheck' },
  { label: '翻译', value: 'translation', icon: 'Translate' }
])

export const installedTemplates = ref(new Set([1, 6]))

export function useTemplate() {
  return { mockTemplates, templateCategories, installedTemplates }
}
