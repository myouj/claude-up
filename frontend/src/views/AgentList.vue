<template>
  <div class="agent-list">
    <el-header>
      <div class="header-content">
        <div class="left">
          <el-button class="back-btn" @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <h1>Agents</h1>
        </div>
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          新建 Agent
        </el-button>
      </div>
    </el-header>

    <el-main>
      <div v-if="agents.length > 0" class="agent-grid">
        <el-card
          v-for="agent in agents"
          :key="agent.id"
          class="agent-card"
          :class="{ builtin: agent.source === 'builtin' }"
          @click="goToEditor(agent.id)"
        >
          <template #header>
            <div class="card-header">
              <div class="title-row">
                <el-avatar :size="32" class="agent-avatar">
                  {{ agent.name.charAt(0).toUpperCase() }}
                </el-avatar>
                <div class="title-info">
                  <span class="name">{{ agent.role || agent.name }}</span>
                  <el-tag v-if="agent.source === 'builtin'" type="success" size="small">内置</el-tag>
                  <el-tag v-else type="info" size="small">自定义</el-tag>
                </div>
              </div>
            </div>
          </template>

          <div class="card-body">
            <p class="description">{{ agent.capabilities || '暂无能力描述' }}</p>
            <div class="meta">
              <el-tag v-if="agent.category" size="small" type="info">{{ agent.category }}</el-tag>
              <span v-if="agent.content_cn" class="translated-badge">
                <el-icon><Check /></el-icon>
                已翻译
              </span>
            </div>
          </div>

          <template #footer>
            <div class="card-footer">
              <el-button size="small" @click.stop="goToTranslate(agent.id)">
                <el-icon><Translate /></el-icon>
                翻译
              </el-button>
              <el-button
                v-if="agent.source !== 'builtin'"
                size="small"
                type="danger"
                @click.stop="handleDelete(agent)"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </template>
        </el-card>
      </div>

      <el-empty v-else description="暂无 Agents">
        <el-button type="primary" @click="showCreateDialog = true">
          创建第一个 Agent
        </el-button>
      </el-empty>
    </el-main>

    <el-dialog v-model="showCreateDialog" title="新建 Agent" width="560px">
      <el-form :model="newAgent" label-position="top">
        <el-form-item label="名称" required>
          <el-input v-model="newAgent.name" placeholder="如: code-reviewer" />
        </el-form-item>
        <el-form-item label="角色">
          <el-input v-model="newAgent.role" placeholder="如: Code Reviewer" />
        </el-form-item>
        <el-form-item label="能力描述">
          <el-input v-model="newAgent.capabilities" type="textarea" :rows="2" placeholder="如: 代码审查、安全检测" />
        </el-form-item>
        <el-form-item label="系统提示词" required>
          <el-input v-model="newAgent.content" type="textarea" :rows="6" placeholder="Agent 系统提示词..." />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="newAgent.category" placeholder="选择分类" allow-create filterable clearable>
            <el-option label="development" value="development" />
            <el-option label="security" value="security" />
            <el-option label="docs" value="docs" />
            <el-option label="devops" value="devops" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'

const router = useRouter()
const agents = ref([])
const showCreateDialog = ref(false)

const newAgent = ref({
  name: '',
  role: '',
  capabilities: '',
  content: '',
  category: ''
})

const fetchAgents = async () => {
  try {
    const res = await axios.get('/api/agents')
    if (res.data.success) {
      agents.value = res.data.data
    }
  } catch (err) {
    console.error('Failed to fetch agents:', err)
  }
}

const handleCreate = async () => {
  if (!newAgent.value.name || !newAgent.value.content) {
    ElMessage.warning('请填写名称和内容')
    return
  }
  try {
    await axios.post('/api/agents', newAgent.value)
    ElMessage.success('创建成功')
    showCreateDialog.value = false
    newAgent.value = { name: '', role: '', capabilities: '', content: '', category: '' }
    fetchAgents()
  } catch (err) {
    ElMessage.error('创建失败')
  }
}

const handleDelete = async (agent) => {
  try {
    await ElMessageBox.confirm(`确定删除 "${agent.name}" 吗？`, '删除确认', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await axios.delete(`/api/agents/${agent.id}`)
    ElMessage.success('删除成功')
    fetchAgents()
  } catch (err) {
    if (err !== 'cancel') ElMessage.error('删除失败')
  }
}

const goBack = () => router.push('/')
const goToEditor = (id) => router.push(`/agents/${id}`)
const goToTranslate = (id) => router.push(`/agents/${id}/translate`)

onMounted(fetchAgents)
</script>

<style scoped>
.agent-list {
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

.left h1 {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.el-main {
  padding: var(--spacing-6);
}

.agent-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: var(--spacing-5);
}

.agent-card {
  cursor: pointer;
  transition: all var(--transition-normal);
}

.agent-card.builtin {
  border-left: 3px solid var(--color-warning);
}

.card-header {
  display: flex;
  justify-content: space-between;
}

.title-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.agent-avatar {
  background: var(--color-warning);
  color: white;
  font-weight: var(--font-weight-semibold);
}

.title-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.name {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.card-body .description {
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  margin-bottom: var(--spacing-3);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.meta {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.translated-badge {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  font-size: var(--font-size-xs);
  color: var(--color-success);
}

.card-footer {
  display: flex;
  justify-content: space-between;
}
</style>
