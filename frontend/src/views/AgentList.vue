<template>
  <div class="agent-list">
    <el-container>
      <el-header>
        <div class="header-content">
          <div class="left-group">
            <el-button class="mobile-menu-btn" @click="showSidebar = true">
              <el-icon><Menu /></el-icon>
            </el-button>
            <el-button class="back-btn" @click="goBack">
              <el-icon><ArrowLeft /></el-icon>
            </el-button>
            <h1>Agents</h1>
          </div>
          <div class="actions-group">
            <el-button type="primary" @click="showCreateDialog = true">
              <el-icon><Plus /></el-icon>
              <span class="btn-text">新建 Agent</span>
            </el-button>
            <el-button @click="handleExport">
              <el-icon><Download /></el-icon>
              <span class="btn-text">导出</span>
            </el-button>
            <el-button @click="showImportDialog = true">
              <el-icon><Upload /></el-icon>
              <span class="btn-text">导入</span>
            </el-button>
          </div>
        </div>
      </el-header>

      <el-container>
        <!-- Desktop sidebar -->
        <el-aside width="240px" class="sidebar">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索 Agents..."
            :prefix-icon="Search"
            clearable
            class="search-input"
          />

          <div class="nav-section">
            <el-menu :default-active="activeCategory" @select="handleCategorySelect" :ellipsis="false">
              <el-menu-item index="">
                <el-icon><Document /></el-icon>
                <span>全部 Agents</span>
                <span class="count">{{ agents.length }}</span>
              </el-menu-item>
            </el-menu>
          </div>

          <div class="category-section">
            <h3>分类</h3>
            <div class="category-tags">
              <el-tag
                v-for="cat in categories"
                :key="cat"
                :type="activeCategory === cat ? 'primary' : 'info'"
                class="category-tag"
                :effect="activeCategory === cat ? 'dark' : 'light'"
                @click="handleCategorySelect(cat)"
              >
                {{ cat }}
              </el-tag>
            </div>
          </div>
        </el-aside>

        <!-- Mobile sidebar drawer -->
        <el-drawer v-model="showSidebar" title="筛选" size="280px" direction="ltr" class="mobile-sidebar-drawer">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索 Agents..."
            :prefix-icon="Search"
            clearable
            class="search-input"
          />

          <div class="nav-section">
            <el-menu :default-active="activeCategory" @select="(key) => { handleCategorySelect(key); showSidebar = false }" :ellipsis="false">
              <el-menu-item index="">
                <el-icon><Document /></el-icon>
                <span>全部 Agents</span>
                <span class="count">{{ agents.length }}</span>
              </el-menu-item>
            </el-menu>
          </div>

          <div class="category-section">
            <h3>分类</h3>
            <div class="category-tags">
              <el-tag
                v-for="cat in categories"
                :key="cat"
                :type="activeCategory === cat ? 'primary' : 'info'"
                class="category-tag"
                :effect="activeCategory === cat ? 'dark' : 'light'"
                @click="() => { handleCategorySelect(cat); showSidebar = false }"
              >
                {{ cat }}
              </el-tag>
            </div>
          </div>
        </el-drawer>

        <el-main>
      <div v-if="agents.length > 0" class="agent-grid">
        <el-card
          v-for="agent in paginatedAgents"
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
              <div class="footer-left">
                <el-button
                  size="small"
                  text
                  @click.stop="handleClone(agent)"
                  class="clone-btn"
                  title="克隆"
                  :disabled="agent.source === 'builtin'"
                >
                  <el-icon><CopyDocument /></el-icon>
                </el-button>
                <el-button size="small" @click.stop="goToTranslate(agent.id)">
                  <el-icon><Translate /></el-icon>
                  翻译
                </el-button>
              </div>
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

      <div v-if="totalAgents > pageSize" class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="totalAgents"
          layout="prev, pager, next"
          background
        />
      </div>
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

    <el-dialog v-model="showImportDialog" title="导入 Agents" width="560px">
      <el-form :model="{ importType, importText }" label-position="top">
        <el-form-item label="导入格式">
          <el-radio-group v-model="importType">
            <el-radio label="json">JSON</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="导入内容">
          <el-input
            v-model="importText"
            type="textarea"
            :rows="8"
            placeholder="粘贴 JSON 数据..."
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showImportDialog = false">取消</el-button>
        <el-button type="primary" @click="handleImport">导入</el-button>
      </template>
    </el-dialog>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Menu, Search } from '@element-plus/icons-vue'

const router = useRouter()
const agents = ref([])
const searchKeyword = ref('')
const activeCategory = ref('')
const showCreateDialog = ref(false)
const showImportDialog = ref(false)
const currentPage = ref(1)
const pageSize = ref(12)
const totalAgents = ref(0)
const importType = ref('json')
const importText = ref('')
const showSidebar = ref(false)

const newAgent = ref({
  name: '',
  role: '',
  capabilities: '',
  content: '',
  category: ''
})

const fetchAgents = async () => {
  try {
    const res = await axios.get('/api/agents', {
      params: { page: currentPage.value, limit: pageSize.value }
    })
    if (res.data.success) {
      agents.value = res.data.data
      if (res.data.meta) {
        totalAgents.value = res.data.meta.total
      }
    }
  } catch (err) {
    console.error('Failed to fetch agents:', err)
  }
}

const paginatedAgents = computed(() => agents.value)

const categories = computed(() => {
  const cats = new Set(agents.value.map(a => a.category).filter(Boolean))
  return Array.from(cats).sort()
})

const handleCategorySelect = (key) => {
  activeCategory.value = key
  currentPage.value = 1
}

const handleClone = async (agent) => {
  try {
    const res = await axios.post(`/api/agents/${agent.id}/clone`)
    if (res.data.success) {
      ElMessage.success('克隆成功')
      fetchAgents()
    }
  } catch (err) {
    ElMessage.error('克隆失败')
  }
}

const handleExport = async () => {
  try {
    const res = await axios.get('/api/agents/export')
    if (res.data.success) {
      const content = JSON.stringify(res.data.data, null, 2)
      const blob = new Blob([content], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'agents.json'
      a.click()
      URL.revokeObjectURL(url)
      ElMessage.success('导出成功')
    }
  } catch (err) {
    ElMessage.error('导出失败')
  }
}

const handleImport = () => {
  if (!importText.value.trim()) {
    ElMessage.warning('请输入要导入的内容')
    return
  }
  try {
    const payload = JSON.parse(importText.value)
    axios.post('/api/agents/import', { agents: payload.agents || payload })
      .then(res => {
        if (res.data.success) {
          ElMessage.success(`成功导入 ${res.data.imported} 条 Agents`)
          showImportDialog.value = false
          importText.value = ''
          currentPage.value = 1
          fetchAgents()
        }
      })
      .catch(() => ElMessage.error('导入失败'))
  } catch (err) {
    ElMessage.error('JSON 解析失败，请检查格式')
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

watch(currentPage, () => fetchAgents())
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

.left-group {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.mobile-menu-btn {
  display: none;
}

.left-group h1 {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.actions-group {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
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

.agent-card:hover {
  transform: translateY(-2px);
  border-color: var(--color-border-hover);
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
  align-items: center;
}

.footer-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
}

.clone-btn {
  padding: 2px 4px;
  color: var(--color-text-muted);
}

.clone-btn:hover {
  color: var(--color-primary);
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: var(--spacing-6);
  padding-bottom: var(--spacing-4);
}

.sidebar {
  background: var(--color-surface);
  padding: var(--spacing-4);
  border-right: 1px solid var(--color-border);
}

.search-input {
  margin-bottom: var(--spacing-4);
}

.nav-section {
  margin-bottom: var(--spacing-6);
}

.nav-section :deep(.el-menu) {
  border: none;
}

.nav-section :deep(.el-menu-item) {
  height: 40px;
  line-height: 40px;
  display: flex;
  align-items: center;
}

.nav-section :deep(.el-menu-item span) {
  flex: 1;
}

.count {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  background: var(--color-bg);
  padding: 2px 8px;
  border-radius: var(--radius-full);
}

.category-section h3 {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: var(--spacing-3);
}

.category-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-2);
}

.category-tag {
  cursor: pointer;
  transition: all var(--transition-fast);
}

.category-tag:hover {
  transform: translateY(-1px);
}

/* Responsive - Tablet */
@media (max-width: 1024px) {
  .agent-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

/* Responsive - Mobile */
@media (max-width: 768px) {
  .mobile-menu-btn {
    display: flex;
  }

  .sidebar {
    display: none;
  }

  .header-content {
    gap: var(--spacing-2);
  }

  .actions-group {
    gap: var(--spacing-1);
  }

  .btn-text {
    display: none;
  }

  .agent-grid {
    grid-template-columns: 1fr;
  }

  .el-main {
    padding: var(--spacing-3);
  }
}
</style>
