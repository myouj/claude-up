<template>
  <div class="test-compare">
    <BreadcrumbNav :items="[{ name: '提示词', path: '/prompts' }, { name: '对比测试' }]" />
    <el-header>
      <div class="header-content">
        <div class="left">
          <el-button class="mobile-menu-btn" @click="showSidebar = true">
            <el-icon><Menu /></el-icon>
          </el-button>
          <el-button class="back-btn" @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <h2 class="page-title">测试对比</h2>
          <span class="prompt-title">{{ promptTitle }}</span>
        </div>
        <div class="right">
          <el-select v-model="selectedVersionId" placeholder="选择版本" class="version-select" clearable>
            <el-option label="全部版本" :value="null" />
            <el-option v-for="v in versions" :key="v.id" :label="`v${v.version}`" :value="v.id" />
          </el-select>
          <el-select v-model="selectedModel" class="model-select">
            <el-option label="MiniMax" value="MiniMax-Text-01" />
            <el-option label="阿里百炼 (Qwen)" value="qwen-turbo" />
          </el-select>
          <el-button type="primary" @click="runNewTest" :loading="running">
            <el-icon><Promotion /></el-icon>
            <span class="btn-text">运行新测试</span>
          </el-button>
        </div>
      </div>
    </el-header>

    <el-container>
      <el-aside width="360px" class="sidebar">
        <div class="sidebar-header">
          <span class="section-title">测试记录</span>
          <el-tag size="small" type="info">{{ filteredRecords.length }} 条</el-tag>
        </div>

        <div v-if="loadingRecords" class="loading-state">
          <el-icon class="is-loading"><Loading /></el-icon>
        </div>

        <div v-else-if="filteredRecords.length === 0" class="empty-state">
          <p>暂无测试记录</p>
          <span>运行测试后在此查看对比</span>
        </div>

        <div v-else class="test-list">
          <div
            v-for="record in filteredRecords"
            :key="record.id"
            class="test-item"
            :class="{ selected: selectedRecordIds.includes(record.id) }"
            @click="toggleSelect(record)"
          >
            <div class="test-item-header">
              <el-checkbox
                :model-value="selectedRecordIds.includes(record.id)"
                @click.stop
                @change="toggleSelect(record)"
              />
              <el-tag size="small" type="info">{{ record.model }}</el-tag>
              <span class="test-time">{{ formatTime(record.created_at) }}</span>
            </div>
            <p class="test-preview">{{ record.response?.substring(0, 60) }}...</p>
            <div class="test-meta">
              <span v-if="record.tokens_used > 0">{{ record.tokens_used }} tokens</span>
            </div>
          </div>
        </div>
      </el-aside>

      <!-- Mobile sidebar drawer -->
      <el-drawer v-model="showSidebar" title="测试记录" size="300px" direction="ltr">
        <div class="sidebar-header">
          <span class="section-title">测试记录</span>
          <el-tag size="small" type="info">{{ filteredRecords.length }} 条</el-tag>
        </div>
        <div v-if="loadingRecords" class="loading-state">
          <el-icon class="is-loading"><Loading /></el-icon>
        </div>
        <div v-else-if="filteredRecords.length === 0" class="empty-state">
          <p>暂无测试记录</p>
        </div>
        <div v-else class="test-list">
          <div
            v-for="record in filteredRecords"
            :key="record.id"
            class="test-item"
            :class="{ selected: selectedRecordIds.includes(record.id) }"
            @click="() => { toggleSelect(record); showSidebar = false }"
          >
            <div class="test-item-header">
              <el-checkbox
                :model-value="selectedRecordIds.includes(record.id)"
                @click.stop
                @change="() => { toggleSelect(record); showSidebar = false }"
              />
              <el-tag size="small" type="info">{{ record.model }}</el-tag>
              <span class="test-time">{{ formatTime(record.created_at) }}</span>
            </div>
            <p class="test-preview">{{ record.response?.substring(0, 60) }}...</p>
          </div>
        </div>
      </el-drawer>

      <el-main>
        <div v-if="compareRecords.length === 0" class="compare-placeholder">
          <div class="placeholder-content">
            <svg width="80" height="80" viewBox="0 0 80 80" fill="none">
              <rect x="10" y="15" width="26" height="50" rx="4" stroke="var(--color-border)" stroke-width="2"/>
              <rect x="44" y="15" width="26" height="50" rx="4" stroke="var(--color-border)" stroke-width="2"/>
              <path d="M23 30h8M23 38h5M23 46h6" stroke="var(--color-border)" stroke-width="1.5" stroke-linecap="round"/>
              <path d="M57 30h8M57 38h5M57 46h6" stroke="var(--color-border)" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
            <p>选择测试记录进行对比</p>
            <span>勾选左侧列表中的记录（至少2条）即可查看对比</span>
          </div>
        </div>

        <div v-else class="compare-container">
          <div class="compare-header">
            <span class="compare-count">对比 {{ compareRecords.length }} 条测试结果</span>
            <el-button size="small" @click="clearSelection">清除选择</el-button>
          </div>

          <div class="compare-panels">
            <div
              v-for="record in compareRecords"
              :key="record.id"
              class="compare-panel"
            >
              <div class="panel-header">
                <div class="panel-info">
                  <el-tag size="small" type="primary">{{ record.model }}</el-tag>
                  <span class="panel-time">{{ formatDate(record.created_at) }}</span>
                </div>
                <div class="panel-actions">
                  <el-button size="small" text @click="copyResponse(record)">
                    <el-icon><CopyDocument /></el-icon>
                  </el-button>
                  <el-button size="small" text type="danger" @click="removeFromCompare(record)">
                    <el-icon><Close /></el-icon>
                  </el-button>
                </div>
              </div>
              <div class="panel-body">
                <div class="prompt-section">
                  <div class="section-label">输入</div>
                  <pre class="prompt-text">{{ record.prompt_text }}</pre>
                </div>
                <div class="response-section">
                  <div class="section-label">
                    输出
                    <span v-if="record.tokens_used > 0" class="tokens-badge">{{ record.tokens_used }} tokens</span>
                  </div>
                  <pre class="response-text">{{ record.response }}</pre>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { Menu } from '@element-plus/icons-vue'
import BreadcrumbNav from '../components/BreadcrumbNav.vue'

const router = useRouter()
const route = useRoute()

const promptTitle = ref('')
const versions = ref([])
const selectedVersionId = ref(null)
const selectedModel = ref('gpt-4')
const testRecords = ref([])
const selectedRecordIds = ref([])
const loadingRecords = ref(false)
const running = ref(false)
const showSidebar = ref(false)

const filteredRecords = computed(() => {
  if (selectedVersionId.value === null) return testRecords.value
  return testRecords.value.filter(r => r.version_id === selectedVersionId.value)
})

const compareRecords = computed(() => {
  return testRecords.value.filter(r => selectedRecordIds.value.includes(r.id))
})

const fetchVersions = async () => {
  try {
    const res = await axios.get(`/api/prompts/${route.params.id}/versions`)
    if (res.data.success) {
      versions.value = res.data.data
    }
  } catch (err) {
    console.error('Failed to fetch versions:', err)
  }
}

const fetchPromptInfo = async () => {
  try {
    const res = await axios.get(`/api/prompts/${route.params.id}`)
    if (res.data.success) {
      promptTitle.value = res.data.data.title
    }
  } catch (err) {
    console.error('Failed to fetch prompt:', err)
  }
}

const fetchTestRecords = async () => {
  loadingRecords.value = true
  try {
    const res = await axios.get(`/api/prompts/${route.params.id}/tests`)
    if (res.data.success) {
      testRecords.value = res.data.data || []
    }
  } catch (err) {
    console.error('Failed to fetch test records:', err)
  } finally {
    loadingRecords.value = false
  }
}

const toggleSelect = (record) => {
  const idx = selectedRecordIds.value.indexOf(record.id)
  if (idx >= 0) {
    selectedRecordIds.value.splice(idx, 1)
  } else {
    selectedRecordIds.value.push(record.id)
  }
}

const removeFromCompare = (record) => {
  const idx = selectedRecordIds.value.indexOf(record.id)
  if (idx >= 0) selectedRecordIds.value.splice(idx, 1)
}

const clearSelection = () => {
  selectedRecordIds.value = []
}

const copyResponse = (record) => {
  navigator.clipboard.writeText(record.response)
  ElMessage.success('已复制')
}

const runNewTest = async () => {
  running.value = true
  try {
    const res = await axios.get(`/api/prompts/${route.params.id}`)
    if (res.data.success) {
      const content = res.data.data.content
      const testRes = await axios.post(`/api/prompts/${route.params.id}/test`, {
        content: content,
        model: selectedModel.value
      })
      if (testRes.data.success) {
        ElMessage.success('测试完成')
        fetchTestRecords()
        // Auto-select the new test
        if (testRes.data.data.test_record_id) {
          selectedRecordIds.value = [testRes.data.data.test_record_id]
        }
      }
    }
  } catch (err) {
    ElMessage.error('测试失败')
  } finally {
    running.value = false
  }
}

const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  return `${date.getMonth() + 1}/${date.getDate()} ${date.getHours()}:${String(date.getMinutes()).padStart(2, '0')}`
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

const goBack = () => router.back()

onMounted(() => {
  fetchVersions()
  fetchPromptInfo()
  fetchTestRecords()
})
</script>

<style scoped>
.test-compare {
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

.back-btn {
  padding: var(--spacing-2);
}

.left h2 {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.prompt-title {
  color: var(--color-text-muted);
  font-size: var(--font-size-sm);
  padding-left: var(--spacing-3);
  border-left: 1px solid var(--color-border);
}

.right {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.version-select {
  width: 130px;
}

.model-select {
  width: 140px;
}

.sidebar {
  background: var(--color-surface);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-4);
  border-bottom: 1px solid var(--color-border);
}

.section-title {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-secondary);
}

.loading-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
}

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  text-align: center;
  padding: var(--spacing-6);
}

.empty-state p {
  font-size: var(--font-size-sm);
  margin-bottom: var(--spacing-1);
}

.empty-state span {
  font-size: var(--font-size-xs);
}

.test-list {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-3);
}

.test-item {
  padding: var(--spacing-3);
  background: var(--color-bg);
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-2);
  cursor: pointer;
  transition: all var(--transition-fast);
  border: 1px solid transparent;
}

.test-item:hover {
  border-color: var(--color-border-hover);
}

.test-item.selected {
  border-color: var(--color-primary);
  background: var(--color-primary-light);
}

.test-item-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-2);
}

.test-time {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  margin-left: auto;
}

.test-preview {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  line-height: var(--line-height-relaxed);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin: 0 0 var(--spacing-1) 0;
}

.test-meta {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.el-main {
  padding: var(--spacing-5);
  background: var(--color-bg);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.compare-placeholder {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.placeholder-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  color: var(--color-text-muted);
}

.placeholder-content p {
  font-size: var(--font-size-md);
  margin: var(--spacing-4) 0 var(--spacing-1);
}

.placeholder-content span {
  font-size: var(--font-size-xs);
}

.compare-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.compare-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-4);
}

.compare-count {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-secondary);
}

.compare-panels {
  flex: 1;
  display: flex;
  gap: var(--spacing-4);
  overflow-x: auto;
  overflow-y: hidden;
  padding-bottom: var(--spacing-3);
}

.compare-panel {
  flex: 1;
  min-width: 300px;
  max-width: 500px;
  background: var(--color-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-3) var(--spacing-4);
  background: var(--color-bg);
  border-bottom: 1px solid var(--color-border);
}

.panel-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.panel-time {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.panel-actions {
  display: flex;
  gap: var(--spacing-1);
}

.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-4);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4);
}

.section-label {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: var(--spacing-2);
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.tokens-badge {
  font-weight: var(--font-weight-regular);
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  background: var(--color-bg);
  padding: 1px 6px;
  border-radius: var(--radius-sm);
}

.prompt-section,
.response-section {
  display: flex;
  flex-direction: column;
}

.prompt-text {
  font-size: var(--font-size-sm);
  line-height: 1.6;
  color: var(--color-text-secondary);
  background: var(--color-bg);
  padding: var(--spacing-3);
  border-radius: var(--radius-md);
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 120px;
  overflow-y: auto;
}

.response-text {
  font-size: var(--font-size-sm);
  line-height: 1.7;
  color: var(--color-text-primary);
  background: var(--color-surface);
  padding: var(--spacing-3);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  flex: 1;
  overflow-y: auto;
}

/* Responsive - Mobile */
@media (max-width: 768px) {
  .mobile-menu-btn {
    display: flex !important;
  }

  .sidebar {
    display: none;
  }

  .left {
    gap: var(--spacing-2);
    min-width: 0;
  }

  .page-title {
    font-size: var(--font-size-md);
  }

  .prompt-title {
    display: none;
  }

  .right {
    gap: var(--spacing-1);
    overflow-x: auto;
    flex-shrink: 0;
  }

  .right::-webkit-scrollbar {
    display: none;
  }

  .btn-text {
    display: none;
  }

  .el-main {
    padding: var(--spacing-3);
  }

  .compare-panels {
    flex-direction: column;
    overflow-x: hidden;
  }

  .compare-panel {
    max-width: 100%;
    min-width: 0;
  }
}
</style>
