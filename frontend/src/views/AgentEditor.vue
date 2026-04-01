<template>
  <div class="agent-editor">
    <BreadcrumbNav :items="[{ name: 'Agents', path: '/agents' }, { name: '编辑' }]" />
    <el-header>
      <div class="header-content">
        <div class="left">
          <el-button @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <el-avatar :size="32" class="agent-avatar">
            {{ agent.name?.charAt(0).toUpperCase() || 'A' }}
          </el-avatar>
          <h2>{{ agent.role || agent.name || 'Agent 详情' }}</h2>
          <el-tag v-if="agent.source === 'builtin'" type="success" size="small">内置</el-tag>
        </div>
        <div class="right">
          <el-button @click="goToTranslate">
            <el-icon><Translate /></el-icon>
            翻译对比
          </el-button>
          <el-button type="primary" @click="handleSave" :disabled="agent.source === 'builtin'">
            <el-icon><Check /></el-icon>
            保存
          </el-button>
        </div>
      </div>
    </el-header>

    <el-container>
      <el-aside width="300px" class="sidebar">
        <el-form :model="agent" label-position="top">
          <el-form-item label="名称">
            <el-input v-model="agent.name" :disabled="agent.source === 'builtin'" />
          </el-form-item>
          <el-form-item label="角色">
            <el-input v-model="agent.role" :disabled="agent.source === 'builtin'" />
          </el-form-item>
          <el-form-item label="能力描述">
            <el-input v-model="agent.capabilities" type="textarea" :rows="3" />
          </el-form-item>
          <el-form-item label="分类">
            <el-select v-model="agent.category" :disabled="agent.source === 'builtin'" allow-create filterable clearable>
              <el-option label="development" value="development" />
              <el-option label="security" value="security" />
              <el-option label="docs" value="docs" />
              <el-option label="devops" value="devops" />
            </el-select>
          </el-form-item>
        </el-form>
      </el-aside>

      <el-main>
        <div class="editor-container">
          <div class="editor-header">
            <span>系统提示词</span>
          </div>
          <el-input
            v-model="agent.content"
            type="textarea"
            class="content-editor"
            :disabled="agent.source === 'builtin'"
            placeholder="Agent 系统提示词..."
          />
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import BreadcrumbNav from '../components/BreadcrumbNav.vue'

const router = useRouter()
const route = useRoute()

const agent = ref({
  id: null,
  name: '',
  role: '',
  content: '',
  content_cn: '',
  capabilities: '',
  category: '',
  source: 'custom'
})

const fetchAgent = async () => {
  try {
    const res = await axios.get(`/api/agents/${route.params.id}`)
    if (res.data.success) {
      agent.value = res.data.data
    }
  } catch (err) {
    ElMessage.error('获取 Agent 失败')
  }
}

const handleSave = async () => {
  try {
    await axios.put(`/api/agents/${agent.value.id}`, {
      name: agent.value.name,
      role: agent.value.role,
      content: agent.value.content,
      capabilities: agent.value.capabilities,
      category: agent.value.category
    })
    ElMessage.success('保存成功')
  } catch (err) {
    ElMessage.error('保存失败')
  }
}

const goBack = () => router.push('/agents')
const goToTranslate = () => router.push(`/agents/${route.params.id}/translate`)

onMounted(fetchAgent)
</script>

<style scoped>
.agent-editor {
  height: 100vh;
  background: var(--color-bg);
}

.el-header {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  padding: 0 var(--spacing-6);
  height: 64px;
}

.header-content {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.left {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.left h2 {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.agent-avatar {
  background: var(--color-warning);
  color: white;
  font-weight: var(--font-weight-semibold);
}

.right {
  display: flex;
  gap: var(--spacing-2);
  flex-shrink: 0;
}

.sidebar {
  background: var(--color-surface);
  padding: var(--spacing-5);
  border-right: 1px solid var(--color-border);
}

.el-main {
  padding: var(--spacing-5);
  background: var(--color-bg);
}

.editor-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--color-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
  overflow: hidden;
}

.editor-header {
  padding: var(--spacing-3) var(--spacing-4);
  background: var(--color-bg);
  border-bottom: 1px solid var(--color-border);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-secondary);
}

.content-editor {
  flex: 1;
}

.content-editor :deep(.el-textarea__inner) {
  height: 100%;
  padding: var(--spacing-4);
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: var(--font-size-sm);
  line-height: 1.7;
  border: none;
  border-radius: 0;
  resize: none;
}
</style>
