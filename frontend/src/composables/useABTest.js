import { ref } from 'vue'

// Mock data for A/B testing UI demonstration
export const mockABTests = ref([
  {
    id: 1,
    name: 'Code Review Prompt 优化测试',
    prompt_id: 1,
    prompt_title: '代码审查专家',
    created_at: '2026-03-27T14:30:00Z',
    status: 'completed',
    winner: 'variant_b',
    total_runs: 48,
    variants: [
      {
        id: 'a',
        name: 'Variant A (原版)',
        description: '原始 prompt 模板',
        content: '你是一位资深代码审查专家。请审查以下代码，关注：\n1. 代码质量与可读性\n2. 潜在的 bug 和安全漏洞\n3. 性能问题\n4. 最佳实践遵循',
        metrics: {
          avg_score: 3.2,
          avg_latency: 1200,
          token_usage: 850
        },
        runs: 24,
        test_records: [
          {
            id: 101,
            input: 'function add(a, b) { return a + b; }',
            response: '这个函数实现简洁有效。建议：\n1. 添加 TypeScript 类型注解\n2. 考虑参数验证\n3. 简洁的命名',
            score: 3,
            latency: 1150,
            created_at: '2026-03-27T10:00:00Z'
          },
          {
            id: 102,
            input: 'for (let i = 0; i < items.length; i++) { process(items[i]); }',
            response: '建议使用 forEach 或 map 替代传统的 for 循环，提高可读性。',
            score: 3,
            latency: 1280,
            created_at: '2026-03-27T11:30:00Z'
          },
          {
            id: 103,
            input: 'const x = data.filter(item => item.active).map(item => item.id);',
            response: '链式调用很优雅。性能方面，如果数据量大可考虑单次遍历。',
            score: 4,
            latency: 1100,
            created_at: '2026-03-27T12:45:00Z'
          }
        ]
      },
      {
        id: 'b',
        name: 'Variant B (结构化)',
        description: '增加输出格式约束',
        content: '你是一位资深代码审查专家。请按以下结构审查代码：\n\n## 优点\n- 列出代码的积极方面\n\n## 问题\n- 列出发现的问题（严重程度：高/中/低）\n\n## 建议\n- 提供具体的改进建议\n\n## 代码片段\n- 标注具体的行号和问题代码',
        metrics: {
          avg_score: 4.5,
          avg_latency: 1450,
          token_usage: 1200
        },
        runs: 24,
        test_records: [
          {
            id: 201,
            input: 'function add(a, b) { return a + b; }',
            response: '## 优点\n- 函数逻辑清晰\n\n## 问题\n- 中：缺少类型注解\n\n## 建议\n- 添加 TypeScript：`function add(a: number, b: number): number`',
            score: 5,
            latency: 1420,
            created_at: '2026-03-27T10:05:00Z'
          },
          {
            id: 202,
            input: 'for (let i = 0; i < items.length; i++) { process(items[i]); }',
            response: '## 优点\n- 无副作用\n\n## 问题\n- 高：数组长度在循环中重新计算可能影响性能\n\n## 建议\n```js\nfor (let i = 0, len = items.length; i < len; i++)\n```',
            score: 4,
            latency: 1500,
            created_at: '2026-03-27T11:35:00Z'
          },
          {
            id: 203,
            input: 'const x = data.filter(item => item.active).map(item => item.id);',
            response: '## 优点\n- 使用了函数式编程风格\n\n## 问题\n- 中：两次遍历可合并\n\n## 建议\n`const x = data.reduce(...)` 或直接使用 `for` 循环单次遍历',
            score: 4,
            latency: 1430,
            created_at: '2026-03-27T12:50:00Z'
          }
        ]
      }
    ]
  },
  {
    id: 2,
    name: 'SQL 生成 Prompt 对比',
    prompt_id: 2,
    prompt_title: 'SQL 生成助手',
    created_at: '2026-03-26T09:15:00Z',
    status: 'completed',
    winner: 'variant_a',
    total_runs: 32,
    variants: [
      {
        id: 'a',
        name: 'Variant A (简洁版)',
        description: '直接生成 SQL',
        content: '根据用户需求生成 SQL 查询。\n需求：{{requirement}}\n数据库类型：{{db_type}}',
        metrics: { avg_score: 4.1, avg_latency: 980, token_usage: 620 },
        runs: 16
      },
      {
        id: 'b',
        name: 'Variant B (带解释)',
        description: 'SQL + 解释',
        content: '根据需求生成 SQL 并解释。\n需求：{{requirement}}\n\n格式：\n```sql\n-- SQL 语句\n```\n\n说明：\n- 解释关键部分',
        metrics: { avg_score: 3.8, avg_latency: 1350, token_usage: 950 },
        runs: 16
      }
    ]
  },
  {
    id: 3,
    name: '翻译 Prompt 润色测试',
    prompt_id: 3,
    prompt_title: '翻译润色助手',
    created_at: '2026-03-25T16:00:00Z',
    status: 'running',
    winner: null,
    total_runs: 12,
    variants: [
      {
        id: 'a',
        name: 'Variant A',
        description: '直译风格',
        content: '翻译为{{target_lang}}：{{text}}',
        metrics: { avg_score: 3.5, avg_latency: 800, token_usage: 450 },
        runs: 6
      },
      {
        id: 'b',
        name: 'Variant B',
        description: '意译风格',
        content: '翻译为{{target_lang}}，保持原文风格：\n{{text}}',
        metrics: { avg_score: 3.7, avg_latency: 950, token_usage: 580 },
        runs: 6
      }
    ]
  }
])

export function useABTest() {
  return { mockABTests }
}
