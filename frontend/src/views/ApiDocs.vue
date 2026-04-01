<template>
  <div class="api-docs">
    <el-header>
      <div class="header-content">
        <div class="left">
          <div class="brand">
            <el-button class="back-btn" @click="goBack">
              <el-icon><ArrowLeft /></el-icon>
            </el-button>
            <h1>API 文档</h1>
          </div>
        </div>
        <div class="right">
          <el-input
            v-model="searchQuery"
            placeholder="搜索 API 端点..."
            :prefix-icon="Search"
            clearable
            class="search-input"
          />
        </div>
      </div>
    </el-header>

    <el-main>
      <div class="docs-container">
        <!-- Module Tabs -->
        <div class="module-tabs">
          <el-radio-group v-model="activeModule" class="module-radio">
            <el-radio-button label="all">全部</el-radio-button>
            <el-radio-button label="prompt">提示词</el-radio-button>
            <el-radio-button label="skill">Skill</el-radio-button>
            <el-radio-button label="agent">Agent</el-radio-button>
            <el-radio-button label="other">其他</el-radio-button>
          </el-radio-group>
        </div>

        <!-- API List -->
        <div v-if="filteredEndpoints.length > 0" class="api-list">
          <el-collapse v-model="expandedItems" accordion>
            <el-collapse-item
              v-for="endpoint in filteredEndpoints"
              :key="endpoint.method + endpoint.path"
              :name="endpoint.method + endpoint.path"
            >
              <template #title>
                <div class="endpoint-header">
                  <el-tag
                    :type="methodType(endpoint.method)"
                    effect="dark"
                    size="small"
                    class="method-tag"
                  >
                    {{ endpoint.method }}
                  </el-tag>
                  <code class="endpoint-path">{{ endpoint.path }}</code>
                  <span class="endpoint-name">{{ endpoint.name }}</span>
                  <el-tag v-if="endpoint.auth" size="small" type="warning" class="auth-tag">需认证</el-tag>
                </div>
              </template>

              <div class="endpoint-body">
                <!-- Description -->
                <div v-if="endpoint.description" class="endpoint-desc">
                  <p>{{ endpoint.description }}</p>
                </div>

                <!-- Parameters -->
                <div v-if="endpoint.params && endpoint.params.length > 0" class="params-section">
                  <h4>请求参数</h4>
                  <el-table :data="endpoint.params" stripe size="small" class="params-table">
                    <el-table-column prop="name" label="参数" width="160">
                      <template #default="{ row }">
                        <code class="param-name">{{ row.name }}</code>
                        <el-tag v-if="row.required" size="small" type="danger" class="required-tag">必填</el-tag>
                      </template>
                    </el-table-column>
                    <el-table-column prop="type" label="类型" width="100">
                      <template #default="{ row }">
                        <el-tag size="small" type="info">{{ row.type }}</el-tag>
                      </template>
                    </el-table-column>
                    <el-table-column prop="description" label="说明" />
                  </el-table>
                </div>

                <!-- Request Body -->
                <div v-if="endpoint.body" class="body-section">
                  <h4>请求体</h4>
                  <pre class="code-block"><code>{{ endpoint.body }}</code></pre>
                </div>

                <!-- Response -->
                <div v-if="endpoint.response" class="response-section">
                  <h4>响应示例</h4>
                  <pre class="code-block response"><code>{{ endpoint.response }}</code></pre>
                </div>

                <!-- Code Examples -->
                <div class="examples-section">
                  <h4>请求示例</h4>
                  <div class="example-tabs">
                    <el-radio-group v-model="endpoint.activeExample" size="small">
                      <el-radio-button label="curl">cURL</el-radio-button>
                      <el-radio-button label="js">JavaScript</el-radio-button>
                      <el-radio-button label="python">Python</el-radio-button>
                    </el-radio-group>
                  </div>
                  <div class="example-code">
                    <pre class="code-block"><code>{{ getExample(endpoint, 'curl') }}</code></pre>
                  </div>
                </div>
              </div>
            </el-collapse-item>
          </el-collapse>
        </div>

        <!-- Empty State -->
        <div v-else class="empty-state">
          <div class="empty-icon">
            <svg width="80" height="80" viewBox="0 0 80 80" fill="none">
              <rect x="15" y="15" width="50" height="50" rx="8" stroke="var(--color-border)" stroke-width="2"/>
              <path d="M25 30h30M25 40h20M25 50h25" stroke="var(--color-border)" stroke-width="2" stroke-linecap="round"/>
            </svg>
          </div>
          <p>未找到匹配的 API 端点</p>
          <span>尝试调整搜索关键词</span>
        </div>
      </div>
    </el-main>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Search } from '@element-plus/icons-vue'

const router = useRouter()
const searchQuery = ref('')
const activeModule = ref('all')
const expandedItems = ref('')

// API endpoints data
const endpoints = ref([
  // Prompt APIs
  {
    module: 'prompt',
    method: 'GET',
    path: '/api/prompts',
    name: '获取提示词列表',
    description: '获取所有提示词，支持分页、搜索和分类筛选',
    auth: true,
    params: [
      { name: 'page', type: 'int', description: '页码，默认 1', required: false },
      { name: 'limit', type: 'int', description: '每页数量，默认 20', required: false },
      { name: 'search', type: 'string', description: '搜索关键词', required: false },
      { name: 'category', type: 'string', description: '按分类筛选', required: false },
      { name: 'favorite', type: 'bool', description: '仅返回收藏的提示词', required: false }
    ],
    response: `{
  "success": true,
  "data": [
    {
      "id": 1,
      "title": "代码审查助手",
      "description": "专业的代码审查提示词",
      "category": "development",
      "tags": ["code", "review"],
      "version_count": 5,
      "is_favorite": true,
      "is_pinned": false,
      "updated_at": "2026-03-28T10:00:00Z"
    }
  ],
  "meta": { "total": 42, "page": 1, "limit": 20 }
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/prompts?page=1&limit=20" \\
  -H "Content-Type: application/json"`,
      js: `const res = await fetch('/api/prompts?page=1&limit=20');
const data = await res.json();
console.log(data.data);`,
      python: `import requests

response = requests.get('http://localhost:8080/api/prompts', params={
    'page': 1,
    'limit': 20
})
data = response.json()
print(data['data'])`
    }
  },
  {
    module: 'prompt',
    method: 'POST',
    path: '/api/prompts',
    name: '创建提示词',
    description: '创建新的提示词，自动生成第一个版本',
    auth: true,
    params: [],
    body: `{
  "title": "提示词标题",
  "content": "提示词内容...",
  "description": "简短描述（可选）",
  "category": "development",
  "tags": ["tag1", "tag2"]
}`,
    response: `{
  "success": true,
  "data": {
    "id": 1,
    "title": "提示词标题",
    "version": 1,
    "created_at": "2026-03-28T10:00:00Z"
  }
}`,
    examples: {
      curl: `curl -X POST "http://localhost:8080/api/prompts" \\
  -H "Content-Type: application/json" \\
  -d '{
    "title": "新提示词",
    "content": "提示词内容...",
    "category": "development"
  }'`,
      js: `const res = await fetch('/api/prompts', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    title: '新提示词',
    content: '提示词内容...',
    category: 'development'
  })
});
const data = await res.json();`,
      python: `import requests

response = requests.post('http://localhost:8080/api/prompts', json={
    'title': '新提示词',
    'content': '提示词内容...',
    'category': 'development'
})
data = response.json()`
    }
  },
  {
    module: 'prompt',
    method: 'GET',
    path: '/api/prompts/:id',
    name: '获取提示词详情',
    description: '获取单个提示词的详细信息',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: '提示词 ID', required: true }
    ],
    response: `{
  "success": true,
  "data": {
    "id": 1,
    "title": "代码审查助手",
    "content": "你是一个专业的代码审查员...",
    "content_cn": "你是一个专业的代码审查员...",
    "description": "专业的代码审查提示词",
    "category": "development",
    "tags": ["code", "review"],
    "version_count": 5,
    "is_favorite": true,
    "is_pinned": false,
    "created_at": "2026-03-01T10:00:00Z",
    "updated_at": "2026-03-28T10:00:00Z"
  }
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/prompts/1"`,
      js: `const res = await fetch('/api/prompts/1');
const data = await res.json();
console.log(data.data);`,
      python: `response = requests.get('http://localhost:8080/api/prompts/1')
data = response.json()`
    }
  },
  {
    module: 'prompt',
    method: 'PUT',
    path: '/api/prompts/:id',
    name: '更新提示词',
    description: '更新提示词内容，自动创建新版本（如果内容变化）',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: '提示词 ID', required: true }
    ],
    body: `{
  "title": "更新后的标题",
  "content": "更新后的内容...",
  "description": "更新后的描述",
  "category": "development",
  "tags": ["tag1", "tag2"],
  "is_favorite": true,
  "is_pinned": false,
  "comment": "版本备注（可选）"
}`,
    response: `{
  "success": true,
  "data": {
    "id": 1,
    "version": 6,
    "message": "更新成功"
  }
}`,
    examples: {
      curl: `curl -X PUT "http://localhost:8080/api/prompts/1" \\
  -H "Content-Type: application/json" \\
  -d '{
    "content": "更新后的内容...",
    "comment": "优化了描述"
  }'`,
      js: `const res = await fetch('/api/prompts/1', {
  method: 'PUT',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    content: '更新后的内容...',
    comment: '优化了描述'
  })
});`,
      python: `response = requests.put('http://localhost:8080/api/prompts/1', json={
    'content': '更新后的内容...',
    'comment': '优化了描述'
})`
    }
  },
  {
    module: 'prompt',
    method: 'DELETE',
    path: '/api/prompts/:id',
    name: '删除提示词',
    description: '删除提示词及其所有版本',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: '提示词 ID', required: true }
    ],
    response: `{
  "success": true,
  "message": "删除成功"
}`,
    examples: {
      curl: `curl -X DELETE "http://localhost:8080/api/prompts/1"`,
      js: `await fetch('/api/prompts/1', { method: 'DELETE' });`,
      python: `requests.delete('http://localhost:8080/api/prompts/1')`
    }
  },
  {
    module: 'prompt',
    method: 'GET',
    path: '/api/prompts/:id/versions',
    name: '获取版本历史',
    description: '获取提示词的所有历史版本',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: '提示词 ID', required: true }
    ],
    response: `{
  "success": true,
  "data": [
    {
      "id": 10,
      "version": 5,
      "content": "上一版本内容...",
      "comment": "优化了输出格式",
      "created_at": "2026-03-27T10:00:00Z"
    }
  ]
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/prompts/1/versions"`,
      js: `const res = await fetch('/api/prompts/1/versions');
const data = await res.json();
console.log(data.data);`,
      python: `response = requests.get('http://localhost:8080/api/prompts/1/versions')
data = response.json()`
    }
  },
  {
    module: 'prompt',
    method: 'POST',
    path: '/api/prompts/:id/test',
    name: '测试提示词',
    description: '使用 AI 模型测试提示词，返回模型响应',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: '提示词 ID', required: true }
    ],
    body: `{
  "content": "替换变量后的完整提示词",
  "model": "gpt-4",
  "messages": []
}`,
    response: `{
  "success": true,
  "data": {
    "response": "AI 模型的回复内容...",
    "tokens_used": 1234,
    "latency": 2.5,
    "model": "gpt-4",
    "test_record_id": 1
  }
}`,
    examples: {
      curl: `curl -X POST "http://localhost:8080/api/prompts/1/test" \\
  -H "Content-Type: application/json" \\
  -d '{
    "content": "完整提示词内容...",
    "model": "gpt-4"
  }'`,
      js: `const res = await fetch('/api/prompts/1/test', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    content: '完整提示词内容...',
    model: 'gpt-4'
  })
});
const data = await res.json();`,
      python: `response = requests.post('http://localhost:8080/api/prompts/1/test', json={
    'content': '完整提示词内容...',
    'model': 'gpt-4'
})
data = response.json()`
    }
  },
  {
    module: 'prompt',
    method: 'POST',
    path: '/api/prompts/:id/optimize',
    name: 'AI 优化提示词',
    description: '使用 AI 优化提示词内容，支持多种优化模式',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: '提示词 ID', required: true }
    ],
    body: `{
  "content": "原始提示词内容",
  "mode": "improve"
}`,
    response: `{
  "success": true,
  "data": {
    "optimized": "AI 优化后的提示词内容..."
  }
}`,
    examples: {
      curl: `curl -X POST "http://localhost:8080/api/prompts/1/optimize" \\
  -H "Content-Type: application/json" \\
  -d '{
    "content": "原始提示词内容",
    "mode": "improve"
  }'`,
      js: `const res = await fetch('/api/prompts/1/optimize', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    content: '原始提示词内容',
    mode: 'improve'
  })
});`,
      python: `response = requests.post('http://localhost:8080/api/prompts/1/optimize', json={
    'content': '原始提示词内容',
    'mode': 'improve'
})`
    }
  },

  // Skill APIs
  {
    module: 'skill',
    method: 'GET',
    path: '/api/skills',
    name: '获取 Skills 列表',
    description: '获取所有 Skills，支持分页和分类筛选',
    auth: true,
    params: [
      { name: 'page', type: 'int', description: '页码，默认 1', required: false },
      { name: 'limit', type: 'int', description: '每页数量，默认 20', required: false }
    ],
    response: `{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "/commit",
      "description": "智能生成 git commit message",
      "category": "git",
      "source": "builtin",
      "content_cn": "已翻译内容"
    }
  ],
  "meta": { "total": 5, "page": 1, "limit": 20 }
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/skills?page=1&limit=20"`,
      js: `const res = await fetch('/api/skills?page=1&limit=20');
const data = await res.json();`,
      python: `response = requests.get('http://localhost:8080/api/skills', params={
    'page': 1, 'limit': 20
})`
    }
  },
  {
    module: 'skill',
    method: 'POST',
    path: '/api/skills',
    name: '创建 Skill',
    description: '创建新的 Skill',
    auth: true,
    params: [],
    body: `{
  "name": "/commit",
  "description": "智能生成 git commit message",
  "content": "When given git diff...",
  "category": "git"
}`,
    response: `{
  "success": true,
  "data": {
    "id": 6,
    "name": "/commit"
  }
}`,
    examples: {
      curl: `curl -X POST "http://localhost:8080/api/skills" \\
  -H "Content-Type: application/json" \\
  -d '{"name": "/commit", "content": "When given git diff..."}'`,
      js: `await fetch('/api/skills', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ name: '/commit', content: 'When given...' })
});`,
      python: `requests.post('http://localhost:8080/api/skills', json={
    'name': '/commit', 'content': 'When given...'
})`
    }
  },
  {
    module: 'skill',
    method: 'PUT',
    path: '/api/skills/:id',
    name: '更新 Skill',
    description: '更新 Skill 内容和元数据',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: 'Skill ID', required: true }
    ],
    body: `{
  "name": "/commit",
  "description": "更新后的描述",
  "content": "更新后的内容...",
  "category": "git"
}`,
    response: `{
  "success": true,
  "message": "更新成功"
}`,
    examples: {
      curl: `curl -X PUT "http://localhost:8080/api/skills/6" \\
  -H "Content-Type: application/json" \\
  -d '{"content": "更新后的内容..."}'`,
      js: `await fetch('/api/skills/6', {
  method: 'PUT',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ content: '更新后的内容...' })
});`,
      python: `requests.put('http://localhost:8080/api/skills/6', json={
    'content': '更新后的内容...'
})`
    }
  },
  {
    module: 'skill',
    method: 'DELETE',
    path: '/api/skills/:id',
    name: '删除 Skill',
    description: '删除自定义 Skill（内置 Skill 不可删除）',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: 'Skill ID', required: true }
    ],
    response: `{
  "success": true,
  "message": "删除成功"
}`,
    examples: {
      curl: `curl -X DELETE "http://localhost:8080/api/skills/6"`,
      js: `await fetch('/api/skills/6', { method: 'DELETE' });`,
      python: `requests.delete('http://localhost:8080/api/skills/6')`
    }
  },

  // Agent APIs
  {
    module: 'agent',
    method: 'GET',
    path: '/api/agents',
    name: '获取 Agents 列表',
    description: '获取所有 Agent personas，支持分页',
    auth: true,
    params: [
      { name: 'page', type: 'int', description: '页码，默认 1', required: false },
      { name: 'limit', type: 'int', description: '每页数量，默认 20', required: false }
    ],
    response: `{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "code-reviewer",
      "role": "Code Reviewer",
      "capabilities": "代码审查、安全检测",
      "category": "development",
      "source": "builtin"
    }
  ],
  "meta": { "total": 3, "page": 1, "limit": 20 }
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/agents?page=1&limit=20"`,
      js: `const res = await fetch('/api/agents?page=1&limit=20');
const data = await res.json();`,
      python: `response = requests.get('http://localhost:8080/api/agents', params={
    'page': 1, 'limit': 20
})`
    }
  },
  {
    module: 'agent',
    method: 'POST',
    path: '/api/agents',
    name: '创建 Agent',
    description: '创建新的 Agent persona',
    auth: true,
    params: [],
    body: `{
  "name": "my-agent",
  "role": "Custom Role",
  "content": "You are a helpful assistant...",
  "capabilities": "Agent 的能力描述",
  "category": "development"
}`,
    response: `{
  "success": true,
  "data": {
    "id": 4,
    "name": "my-agent"
  }
}`,
    examples: {
      curl: `curl -X POST "http://localhost:8080/api/agents" \\
  -H "Content-Type: application/json" \\
  -d '{"name": "my-agent", "content": "You are a helpful assistant..."}'`,
      js: `await fetch('/api/agents', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ name: 'my-agent', content: 'You are...' })
});`,
      python: `requests.post('http://localhost:8080/api/agents', json={
    'name': 'my-agent', 'content': 'You are...'
})`
    }
  },
  {
    module: 'agent',
    method: 'PUT',
    path: '/api/agents/:id',
    name: '更新 Agent',
    description: '更新 Agent 内容和元数据',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: 'Agent ID', required: true }
    ],
    body: `{
  "name": "updated-agent",
  "role": "Updated Role",
  "content": "You are an updated assistant...",
  "capabilities": "更新后的能力描述",
  "category": "development"
}`,
    response: `{
  "success": true,
  "message": "更新成功"
}`,
    examples: {
      curl: `curl -X PUT "http://localhost:8080/api/agents/4" \\
  -H "Content-Type: application/json" \\
  -d '{"content": "更新后的内容..."}'`,
      js: `await fetch('/api/agents/4', {
  method: 'PUT',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ content: '更新后的内容...' })
});`,
      python: `requests.put('http://localhost:8080/api/agents/4', json={
    'content': '更新后的内容...'
})`
    }
  },
  {
    module: 'agent',
    method: 'DELETE',
    path: '/api/agents/:id',
    name: '删除 Agent',
    description: '删除自定义 Agent（内置 Agent 不可删除）',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: 'Agent ID', required: true }
    ],
    response: `{
  "success": true,
  "message": "删除成功"
}`,
    examples: {
      curl: `curl -X DELETE "http://localhost:8080/api/agents/4"`,
      js: `await fetch('/api/agents/4', { method: 'DELETE' });`,
      python: `requests.delete('http://localhost:8080/api/agents/4')`
    }
  },

  {
    module: 'prompt',
    method: 'POST',
    path: '/api/prompts/:id/favorite',
    name: '切换收藏状态',
    description: '切换提示词的收藏状态',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: '提示词 ID', required: true }
    ],
    response: `{
  "success": true,
  "data": {
    "is_favorite": true
  }
}`,
    examples: {
      curl: `curl -X POST "http://localhost:8080/api/prompts/1/favorite"`,
      js: `await fetch('/api/prompts/1/favorite', { method: 'POST' });`,
      python: `requests.post('http://localhost:8080/api/prompts/1/favorite')`
    }
  },
  {
    module: 'prompt',
    method: 'GET',
    path: '/api/prompts/categories',
    name: '获取分类列表',
    description: '获取所有提示词分类',
    auth: true,
    params: [],
    response: `{
  "success": true,
  "categories": ["development", "docs", "code"]
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/prompts/categories"`,
      js: `const res = await fetch('/api/prompts/categories');`,
      python: `response = requests.get('http://localhost:8080/api/prompts/categories')`
    }
  },
  {
    module: 'prompt',
    method: 'POST',
    path: '/api/prompts/:id/clone',
    name: '克隆提示词',
    description: '克隆提示词及其所有版本，自动追加"(Copy)"到标题',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: '源提示词 ID', required: true }
    ],
    response: `{
  "success": true,
  "data": {
    "id": 2,
    "title": "代码审查助手 (Copy)"
  }
}`,
    examples: {
      curl: `curl -X POST "http://localhost:8080/api/prompts/1/clone"`,
      js: `const res = await fetch('/api/prompts/1/clone', { method: 'POST' });`,
      python: `requests.post('http://localhost:8080/api/prompts/1/clone')`
    }
  },
  {
    module: 'prompt',
    method: 'GET',
    path: '/api/prompts/export',
    name: '导出提示词',
    description: '导出所有提示词为 JSON 格式（含版本历史）',
    auth: true,
    params: [],
    response: `{
  "success": true,
  "data": {
    "version": "1.0",
    "exported_at": "2026-03-28 10:00:00",
    "prompts": [...]
  }
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/prompts/export"`,
      js: `const res = await fetch('/api/prompts/export');`,
      python: `response = requests.get('http://localhost:8080/api/prompts/export')`
    }
  },
  {
    module: 'prompt',
    method: 'POST',
    path: '/api/prompts/import',
    name: '导入提示词',
    description: '批量导入提示词（支持选择性跳过重复项）',
    auth: true,
    params: [],
    body: `{
  "prompts": [
    {
      "title": "提示词标题",
      "content": "提示词内容...",
      "category": "development",
      "tags": ["tag1"]
    }
  ]
}`,
    response: `{
  "success": true,
  "imported": 5,
  "failed": 0,
  "total_count": 5
}`,
    examples: {
      curl: `curl -X POST "http://localhost:8080/api/prompts/import" \\
  -H "Content-Type: application/json" \\
  -d '{"prompts": [{"title": "New", "content": "..."}]}'`,
      js: `await fetch('/api/prompts/import', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ prompts: [...] })
});`,
      python: `requests.post('http://localhost:8080/api/prompts/import', json={
    'prompts': [...]
})`
    }
  },
  {
    module: 'prompt',
    method: 'POST',
    path: '/api/prompts/:id/versions',
    name: '创建版本',
    description: '手动创建提示词新版本（自动分配版本号）',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: '提示词 ID', required: true }
    ],
    body: `{
  "content": "新版本内容...",
  "comment": "版本备注"
}`,
    response: `{
  "success": true,
  "data": {
    "id": 10,
    "version": 6,
    "created_at": "2026-03-28T10:00:00Z"
  }
}`,
    examples: {
      curl: `curl -X POST "http://localhost:8080/api/prompts/1/versions" \\
  -H "Content-Type: application/json" \\
  -d '{"content": "新内容...", "comment": "优化了描述"}'`,
      js: `await fetch('/api/prompts/1/versions', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ content: '...', comment: '...' })
});`,
      python: `requests.post('http://localhost:8080/api/prompts/1/versions', json={
    'content': '...', 'comment': '...'
})`
    }
  },
  {
    module: 'prompt',
    method: 'GET',
    path: '/api/versions/:id',
    name: '获取指定版本',
    description: '通过版本 ID 获取单个版本详情',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: '版本 ID', required: true }
    ],
    response: `{
  "success": true,
  "data": {
    "id": 10,
    "prompt_id": 1,
    "version": 5,
    "content": "版本内容...",
    "comment": "优化了输出格式",
    "created_at": "2026-03-27T10:00:00Z"
  }
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/versions/10"`,
      js: `const res = await fetch('/api/versions/10');`,
      python: `response = requests.get('http://localhost:8080/api/versions/10')`
    }
  },
  {
    module: 'prompt',
    method: 'GET',
    path: '/api/prompts/:id/tests',
    name: '获取测试记录',
    description: '获取提示词的所有测试历史记录',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: '提示词 ID', required: true }
    ],
    response: `{
  "success": true,
  "data": [
    {
      "id": 1,
      "model": "gpt-4o",
      "tokens_used": 1234,
      "latency_ms": 2500,
      "created_at": "2026-03-28T10:00:00Z"
    }
  ],
  "meta": { "total": 20, "page": 1, "limit": 20 }
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/prompts/1/tests"`,
      js: `const res = await fetch('/api/prompts/1/tests');`,
      python: `response = requests.get('http://localhost:8080/api/prompts/1/tests')`
    }
  },
  {
    module: 'prompt',
    method: 'GET',
    path: '/api/prompts/:id/tests/compare',
    name: '对比测试结果',
    description: '对比提示词各版本的测试效果',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: '提示词 ID', required: true }
    ],
    response: `{
  "success": true,
  "data": [
    {
      "version_id": 1,
      "version": 1,
      "avg_tokens": 500,
      "test_count": 3
    },
    {
      "version_id": 2,
      "version": 2,
      "avg_tokens": 420,
      "test_count": 2
    }
  ]
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/prompts/1/tests/compare"`,
      js: `const res = await fetch('/api/prompts/1/tests/compare');`,
      python: `response = requests.get('http://localhost:8080/api/prompts/1/tests/compare')`
    }
  },
  {
    module: 'prompt',
    method: 'GET',
    path: '/api/prompts/:id/analytics',
    name: '测试分析数据',
    description: '获取提示词的测试分析统计（默认 30 天内）',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: '提示词 ID', required: true },
      { name: 'days', type: 'int', description: '分析天数（默认 30）', required: false }
    ],
    response: `{
  "success": true,
  "data": {
    "total_tests": 25,
    "avg_tokens": 456,
    "avg_latency_ms": 2300,
    "model_distribution": {
      "gpt-4o": 15,
      "claude-3-5-sonnet": 10
    }
  }
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/prompts/1/analytics?days=30"`,
      js: `const res = await fetch('/api/prompts/1/analytics?days=30');`,
      python: `response = requests.get('http://localhost:8080/api/prompts/1/analytics', params={'days': 30})`
    }
  },

  // Skill APIs - additional
  {
    module: 'skill',
    method: 'GET',
    path: '/api/skills/:id',
    name: '获取 Skill 详情',
    description: '获取单个 Skill 的详细信息',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: 'Skill ID', required: true }
    ],
    response: `{
  "success": true,
  "data": {
    "id": 1,
    "name": "/commit",
    "description": "智能生成 git commit message",
    "content": "When given git diff...",
    "content_cn": "已翻译内容",
    "category": "git",
    "source": "builtin"
  }
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/skills/1"`,
      js: `const res = await fetch('/api/skills/1');`,
      python: `response = requests.get('http://localhost:8080/api/skills/1')`
    }
  },
  {
    module: 'skill',
    method: 'GET',
    path: '/api/skills/categories',
    name: '获取分类列表',
    description: '获取所有 Skill 分类',
    auth: true,
    params: [],
    response: `{
  "success": true,
  "categories": ["git", "code", "docs"]
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/skills/categories"`,
      js: `const res = await fetch('/api/skills/categories');`,
      python: `response = requests.get('http://localhost:8080/api/skills/categories')`
    }
  },
  {
    module: 'skill',
    method: 'POST',
    path: '/api/skills/:id/clone',
    name: '克隆 Skill',
    description: '克隆 Skill，自动追加"(Copy)"到名称，source 设为 custom',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: '源 Skill ID', required: true }
    ],
    response: `{
  "success": true,
  "data": {
    "id": 6,
    "name": "/commit (Copy)"
  }
}`,
    examples: {
      curl: `curl -X POST "http://localhost:8080/api/skills/1/clone"`,
      js: `await fetch('/api/skills/1/clone', { method: 'POST' });`,
      python: `requests.post('http://localhost:8080/api/skills/1/clone')`
    }
  },
  {
    module: 'skill',
    method: 'GET',
    path: '/api/skills/export',
    name: '导出 Skills',
    description: '导出所有 Skills 为 JSON 格式',
    auth: true,
    params: [],
    response: `{
  "success": true,
  "data": {
    "version": "1.0",
    "exported_at": "2026-03-28 10:00:00",
    "skills": [...]
  }
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/skills/export"`,
      js: `const res = await fetch('/api/skills/export');`,
      python: `response = requests.get('http://localhost:8080/api/skills/export')`
    }
  },
  {
    module: 'skill',
    method: 'POST',
    path: '/api/skills/import',
    name: '导入 Skills',
    description: '批量导入 Skills',
    auth: true,
    params: [],
    body: `{
  "skills": [
    {
      "name": "/my-skill",
      "description": "My custom skill",
      "content": "Skill content...",
      "category": "custom"
    }
  ]
}`,
    response: `{
  "success": true,
  "imported": 3,
  "failed": 0,
  "total_count": 3
}`,
    examples: {
      curl: `curl -X POST "http://localhost:8080/api/skills/import" \\
  -H "Content-Type: application/json" \\
  -d '{"skills": [{"name": "/my-skill", "content": "..."}]}'`,
      js: `await fetch('/api/skills/import', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ skills: [...] })
});`,
      python: `requests.post('http://localhost:8080/api/skills/import', json={
    'skills': [...]
})`
    }
  },

  // Agent APIs - additional
  {
    module: 'agent',
    method: 'GET',
    path: '/api/agents/:id',
    name: '获取 Agent 详情',
    description: '获取单个 Agent persona 的详细信息',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: 'Agent ID', required: true }
    ],
    response: `{
  "success": true,
  "data": {
    "id": 1,
    "name": "code-reviewer",
    "role": "Code Reviewer",
    "content": "You are an expert code reviewer...",
    "capabilities": "Static analysis, Security review",
    "category": "development",
    "source": "builtin"
  }
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/agents/1"`,
      js: `const res = await fetch('/api/agents/1');`,
      python: `response = requests.get('http://localhost:8080/api/agents/1')`
    }
  },
  {
    module: 'agent',
    method: 'GET',
    path: '/api/agents/categories',
    name: '获取分类列表',
    description: '获取所有 Agent 分类',
    auth: true,
    params: [],
    response: `{
  "success": true,
  "categories": ["development", "security", "docs"]
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/agents/categories"`,
      js: `const res = await fetch('/api/agents/categories');`,
      python: `response = requests.get('http://localhost:8080/api/agents/categories')`
    }
  },
  {
    module: 'agent',
    method: 'POST',
    path: '/api/agents/:id/clone',
    name: '克隆 Agent',
    description: '克隆 Agent，自动追加"(Copy)"到名称，source 设为 custom',
    auth: true,
    params: [
      { name: 'id', type: 'int', description: '源 Agent ID', required: true }
    ],
    response: `{
  "success": true,
  "data": {
    "id": 4,
    "name": "code-reviewer (Copy)"
  }
}`,
    examples: {
      curl: `curl -X POST "http://localhost:8080/api/agents/1/clone"`,
      js: `await fetch('/api/agents/1/clone', { method: 'POST' });`,
      python: `requests.post('http://localhost:8080/api/agents/1/clone')`
    }
  },
  {
    module: 'agent',
    method: 'GET',
    path: '/api/agents/export',
    name: '导出 Agents',
    description: '导出所有 Agents 为 JSON 格式',
    auth: true,
    params: [],
    response: `{
  "success": true,
  "data": {
    "version": "1.0",
    "exported_at": "2026-03-28 10:00:00",
    "agents": [...]
  }
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/agents/export"`,
      js: `const res = await fetch('/api/agents/export');`,
      python: `response = requests.get('http://localhost:8080/api/agents/export')`
    }
  },
  {
    module: 'agent',
    method: 'POST',
    path: '/api/agents/import',
    name: '导入 Agents',
    description: '批量导入 Agents',
    auth: true,
    params: [],
    body: `{
  "agents": [
    {
      "name": "my-agent",
      "role": "Custom Role",
      "content": "You are a helpful assistant...",
      "capabilities": "Custom capability",
      "category": "custom"
    }
  ]
}`,
    response: `{
  "success": true,
  "imported": 2,
  "failed": 0,
  "total_count": 2
}`,
    examples: {
      curl: `curl -X POST "http://localhost:8080/api/agents/import" \\
  -H "Content-Type: application/json" \\
  -d '{"agents": [{"name": "my-agent", "content": "..."}]}'`,
      js: `await fetch('/api/agents/import', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ agents: [...] })
});`,
      python: `requests.post('http://localhost:8080/api/agents/import', json={
    'agents': [...]
})`
    }
  },

  // Other APIs - additional
  {
    module: 'other',
    method: 'GET',
    path: '/api/models',
    name: '获取模型列表',
    description: '获取所有支持的 AI 模型，按 provider 分组',
    auth: false,
    params: [
      { name: 'provider', type: 'string', description: '按 provider 筛选 (openai/claude/gemini/minimax)', required: false }
    ],
    response: `{
  "success": true,
  "data": [
    { "provider": "openai", "model": "gpt-4o", "input_cost_per_1m": 2.5, "output_cost_per_1m": 10.0 },
    { "provider": "claude", "model": "claude-3-5-sonnet-20241022", "input_cost_per_1m": 3.0, "output_cost_per_1m": 15.0 }
  ]
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/models"`,
      js: `const res = await fetch('/api/models');`,
      python: `response = requests.get('http://localhost:8080/api/models')`
    }
  },
  {
    module: 'other',
    method: 'GET',
    path: '/api/export',
    name: '全量导出',
    description: '导出所有数据（Prompts + Skills + Agents）为单个 JSON 文件',
    auth: true,
    params: [],
    response: `{
  "success": true,
  "data": {
    "version": "1.0",
    "exported_at": "2026-03-28 10:00:00",
    "prompts": [...],
    "skills": [...],
    "agents": [...]
  }
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/export"`,
      js: `const res = await fetch('/api/export');`,
      python: `response = requests.get('http://localhost:8080/api/export')`
    }
  },
  {
    module: 'other',
    method: 'GET',
    path: '/api/settings/:key',
    name: '获取单个设置',
    description: '获取指定 key 的设置值（secret 值自动解密返回）',
    auth: true,
    params: [
      { name: 'key', type: 'string', description: '设置键名', required: true }
    ],
    response: `{
  "success": true,
  "data": {
    "key": "openai_api_key",
    "value": "sk-...",
    "is_secret": true
  }
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/settings/openai_api_key"`,
      js: `const res = await fetch('/api/settings/openai_api_key');`,
      python: `response = requests.get('http://localhost:8080/api/settings/openai_api_key')`
    }
  },
  {
    module: 'other',
    method: 'DELETE',
    path: '/api/settings/:key',
    name: '删除设置',
    description: '删除指定 key 的设置',
    auth: true,
    params: [
      { name: 'key', type: 'string', description: '设置键名', required: true }
    ],
    response: `{
  "success": true,
  "message": "设置已删除"
}`,
    examples: {
      curl: `curl -X DELETE "http://localhost:8080/api/settings/openai_api_key"`,
      js: `await fetch('/api/settings/openai_api_key', { method: 'DELETE' });`,
      python: `requests.delete('http://localhost:8080/api/settings/openai_api_key')`
    }
  },

  // Other APIs
  {
    module: 'other',
    method: 'GET',
    path: '/api/stats',
    name: '获取统计数据',
    description: '获取仪表盘统计数据（提示词、Skills、Agents 数量）',
    auth: false,
    params: [],
    response: `{
  "success": true,
  "data": {
    "prompts": 42,
    "skills": 5,
    "agents": 3
  }
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/stats"`,
      js: `const res = await fetch('/api/stats');
const data = await res.json();
console.log(data.data);`,
      python: `response = requests.get('http://localhost:8080/api/stats')
data = response.json()`
    }
  },
  {
    module: 'other',
    method: 'POST',
    path: '/api/translate',
    name: '翻译文本',
    description: '翻译指定文本（不保存到实体）',
    auth: true,
    params: [],
    body: `{
  "text": "Text to translate",
  "source_lang": "en",
  "target_lang": "zh"
}`,
    response: `{
  "success": true,
  "data": {
    "translated_text": "要翻译的文本"
  }
}`,
    examples: {
      curl: `curl -X POST "http://localhost:8080/api/translate" \\
  -H "Content-Type: application/json" \\
  -d '{"text": "Hello", "source_lang": "en", "target_lang": "zh"}'`,
      js: `await fetch('/api/translate', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    text: 'Hello',
    source_lang: 'en',
    target_lang: 'zh'
  })
});`,
      python: `requests.post('http://localhost:8080/api/translate', json={
    'text': 'Hello',
    'source_lang': 'en',
    'target_lang': 'zh'
})`
    }
  },
  {
    module: 'other',
    method: 'POST',
    path: '/api/translate/:type/:id',
    name: '翻译并保存',
    description: '翻译实体内容并保存到数据库',
    auth: true,
    params: [
      { name: 'type', type: 'string', description: '实体类型: prompt / skill / agent', required: true },
      { name: 'id', type: 'int', description: '实体 ID', required: true }
    ],
    body: `{
  "source_lang": "en",
  "target_lang": "zh"
}`,
    response: `{
  "success": true,
  "data": {
    "target_text": "翻译后的内容"
  }
}`,
    examples: {
      curl: `curl -X POST "http://localhost:8080/api/translate/prompt/1" \\
  -H "Content-Type: application/json" \\
  -d '{"source_lang": "en", "target_lang": "zh"}'`,
      js: `await fetch('/api/translate/prompt/1', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    source_lang: 'en',
    target_lang: 'zh'
  })
});`,
      python: `requests.post('http://localhost:8080/api/translate/prompt/1', json={
    'source_lang': 'en',
    'target_lang': 'zh'
})`
    }
  },
  {
    module: 'other',
    method: 'GET',
    path: '/api/activity-logs',
    name: '获取活动日志',
    description: '获取系统操作活动日志',
    auth: true,
    params: [
      { name: 'page', type: 'int', description: '页码', required: false },
      { name: 'limit', type: 'int', description: '每页数量', required: false },
      { name: 'entity_type', type: 'string', description: '按实体类型筛选', required: false },
      { name: 'action', type: 'string', description: '按操作类型筛选', required: false }
    ],
    response: `{
  "success": true,
  "data": [
    {
      "id": 1,
      "entity_type": "prompt",
      "entity_id": 1,
      "action": "created",
      "details": "创建了提示词",
      "created_at": "2026-03-28T10:00:00Z"
    }
  ],
  "meta": { "total": 100, "page": 1, "limit": 20 }
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/activity-logs?page=1&limit=20"`,
      js: `const res = await fetch('/api/activity-logs?page=1&limit=20');
const data = await res.json();`,
      python: `response = requests.get('http://localhost:8080/api/activity-logs', params={
    'page': 1, 'limit': 20
})`
    }
  },
  {
    module: 'other',
    method: 'GET',
    path: '/api/settings',
    name: '获取设置列表',
    description: '获取所有系统设置',
    auth: true,
    params: [],
    response: `{
  "success": true,
  "data": [
    { "key": "openai_api_key", "value": "********", "is_secret": true }
  ]
}`,
    examples: {
      curl: `curl -X GET "http://localhost:8080/api/settings"`,
      js: `const res = await fetch('/api/settings');
const data = await res.json();`,
      python: `response = requests.get('http://localhost:8080/api/settings')
data = response.json()`
    }
  },
  {
    module: 'other',
    method: 'PUT',
    path: '/api/settings/:key',
    name: '更新设置',
    description: '更新系统设置（用于保存 API Key 等敏感信息，使用 AES-256-GCM 加密存储）',
    auth: true,
    params: [
      { name: 'key', type: 'string', description: '设置键名', required: true }
    ],
    body: `{
  "value": "your-api-key",
  "is_secret": true
}`,
    response: `{
  "success": true,
  "message": "设置已保存"
}`,
    examples: {
      curl: `curl -X PUT "http://localhost:8080/api/settings/openai_api_key" \\
  -H "Content-Type: application/json" \\
  -d '{"value": "sk-...", "is_secret": true}'`,
      js: `await fetch('/api/settings/openai_api_key', {
  method: 'PUT',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ value: 'sk-...', is_secret: true })
});`,
      python: `requests.put('http://localhost:8080/api/settings/openai_api_key', json={
    'value': 'sk-...', 'is_secret': True
})`
    }
  }
])

// Add activeExample to each endpoint
endpoints.value.forEach(ep => {
  ep.activeExample = 'curl'
})

const filteredEndpoints = computed(() => {
  let result = endpoints.value

  // Filter by module
  if (activeModule.value !== 'all') {
    result = result.filter(ep => ep.module === activeModule.value)
  }

  // Filter by search query
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(ep =>
      ep.path.toLowerCase().includes(query) ||
      ep.name.toLowerCase().includes(query) ||
      ep.description.toLowerCase().includes(query) ||
      ep.method.toLowerCase().includes(query)
    )
  }

  return result
})

const methodType = (method) => {
  const types = {
    GET: 'success',
    POST: 'primary',
    PUT: 'warning',
    DELETE: 'danger',
    PATCH: 'info'
  }
  return types[method] || 'info'
}

const getExample = (endpoint, type) => {
  return endpoint.examples[type] || endpoint.examples.curl
}

const goBack = () => router.push('/')
</script>

<style scoped>
.api-docs {
  height: 100vh;
  background: var(--color-bg);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.el-header {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  padding: 0 var(--spacing-6);
  height: 64px;
  flex-shrink: 0;
}

.header-content {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-4);
}

.left {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.brand {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.brand h1 {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0;
}

.back-btn {
  padding: var(--spacing-2);
}

.right {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.search-input {
  width: 320px;
}

.el-main {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-5);
}

.docs-container {
  max-width: 900px;
  margin: 0 auto;
}

.module-tabs {
  margin-bottom: var(--spacing-5);
}

.module-radio {
  display: flex;
  gap: var(--spacing-2);
}

.module-radio :deep(.el-radio-button__inner) {
  border-radius: var(--radius-md);
}

.api-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.api-list :deep(.el-collapse) {
  border: none;
}

.api-list :deep(.el-collapse-item__header) {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-3) var(--spacing-4);
  height: auto;
  line-height: normal;
  margin-bottom: 0;
}

.api-list :deep(.el-collapse-item__wrap) {
  border: 1px solid var(--color-border);
  border-top: none;
  border-radius: 0 0 var(--radius-lg) var(--radius-lg);
  margin-bottom: var(--spacing-3);
}

.api-list :deep(.el-collapse-item__content) {
  padding: 0;
}

.api-list :deep(.el-collapse-item.is-active .el-collapse-item__header) {
  border-radius: var(--radius-lg) var(--radius-lg) 0 0;
}

.endpoint-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  flex-wrap: wrap;
  width: 100%;
}

.method-tag {
  flex-shrink: 0;
  min-width: 60px;
  text-align: center;
}

.endpoint-path {
  font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
  font-size: var(--font-size-sm);
  color: var(--color-text-primary);
  flex-shrink: 0;
}

.endpoint-name {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  flex: 1;
  min-width: 100px;
}

.auth-tag {
  flex-shrink: 0;
}

.endpoint-body {
  padding: var(--spacing-4);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4);
}

.endpoint-desc p {
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  margin: 0;
  line-height: var(--line-height-relaxed);
}

.params-section h4,
.body-section h4,
.response-section h4,
.examples-section h4 {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-3) 0;
}

.params-table {
  font-size: var(--font-size-sm);
}

.param-name {
  font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
  color: var(--color-primary);
}

.required-tag {
  margin-left: var(--spacing-1);
}

.code-block {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--spacing-4);
  font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
  font-size: var(--font-size-xs);
  line-height: 1.7;
  overflow-x: auto;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

.code-block code {
  color: var(--color-text-secondary);
}

.code-block.response code {
  color: var(--color-text-primary);
}

.example-tabs {
  margin-bottom: var(--spacing-3);
}

.example-code {
  position: relative;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-12) 0;
  color: var(--color-text-muted);
  text-align: center;
}

.empty-icon {
  opacity: 0.5;
  margin-bottom: var(--spacing-4);
}

.empty-state p {
  font-size: var(--font-size-md);
  color: var(--color-text-secondary);
  margin: 0 0 var(--spacing-1) 0;
}

.empty-state span {
  font-size: var(--font-size-sm);
}

/* Responsive - Mobile */
@media (max-width: 768px) {
  .header-content {
    flex-direction: column;
    align-items: stretch;
    gap: var(--spacing-3);
  }

  .left {
    width: 100%;
  }

  .right {
    width: 100%;
  }

  .search-input {
    width: 100%;
  }

  .module-radio {
    flex-wrap: wrap;
  }

  .endpoint-header {
    gap: var(--spacing-2);
  }

  .endpoint-path {
    font-size: var(--font-size-xs);
  }

  .endpoint-name {
    width: 100%;
    order: 3;
    min-width: 0;
  }

  .el-main {
    padding: var(--spacing-3);
  }

  .docs-container {
    padding: 0;
  }
}
</style>
