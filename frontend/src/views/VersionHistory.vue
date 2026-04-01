<template>
  <div class="version-history">
    <BreadcrumbNav :items="[{ name: '提示词', path: '/prompts' }, { name: '版本历史' }]" />
    <el-header>
      <div class="header-content">
        <div class="left">
          <el-button class="back-btn" @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <div class="title-area">
            <h2>版本历史</h2>
            <span class="prompt-title">{{ promptTitle }}</span>
          </div>
        </div>
        <div class="right">
          <el-button
            @click="goToCompare"
            :disabled="selectedVersions.length !== 2"
            type="primary"
          >
            <el-icon><Connection /></el-icon>
            对比选中版本
          </el-button>
        </div>
      </div>
    </el-header>

    <el-main>
      <div class="timeline-container">
        <el-timeline v-if="versions.length > 0">
          <el-timeline-item
            v-for="version in versions"
            :key="version.id"
            :timestamp="version.created_at"
            placement="top"
            :hollow="!isSelected(version)"
          >
            <el-card
              class="version-card"
              :class="{ selected: isSelected(version), current: version.version === currentVersion }"
              @click="toggleSelect(version)"
            >
              <template #header>
                <div class="card-header">
                  <el-checkbox
                    :model-value="isSelected(version)"
                    @click.stop
                    @change="toggleSelect(version)"
                  >
                    <span class="version-number">v{{ version.version }}</span>
                    <el-tag v-if="version.version === currentVersion" size="small" type="success">当前</el-tag>
                  </el-checkbox>
                  <div class="actions" @click.stop>
                    <el-button size="small" @click="viewContent(version)">
                      查看内容
                    </el-button>
                    <el-button
                      v-if="version.version !== currentVersion"
                      type="primary"
                      size="small"
                      @click="rollbackTo(version)"
                    >
                      恢复此版本
                    </el-button>
                  </div>
                </div>
              </template>
              <div class="version-content">
                <p v-if="version.comment" class="comment">
                  <el-icon><ChatLineSquare /></el-icon>
                  {{ version.comment }}
                </p>
                <p v-else class="no-comment">无备注</p>
              </div>
            </el-card>
          </el-timeline-item>
        </el-timeline>

        <el-empty v-else description="暂无版本记录" class="empty-state">
          <template #image>
            <svg width="100" height="100" viewBox="0 0 100 100" fill="none">
              <circle cx="50" cy="50" r="40" stroke="var(--color-border)" stroke-width="2" stroke-dasharray="8 4"/>
              <path d="M50 30v25l15 10" stroke="var(--color-border)" stroke-width="2" stroke-linecap="round"/>
            </svg>
          </template>
        </el-empty>
      </div>
    </el-main>

    <!-- 内容查看对话框 -->
    <el-dialog
      v-model="showContentDialog"
      title="版本内容"
      width="640px"
      :close-on-click-modal="false"
    >
      <div class="content-viewer">
        <pre>{{ selectedContent }}</pre>
      </div>
      <template #footer>
        <el-button @click="copyContent">复制内容</el-button>
        <el-button type="primary" @click="showContentDialog = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import BreadcrumbNav from '../components/BreadcrumbNav.vue'

const router = useRouter()
const route = useRoute()

const versions = ref([])
const promptTitle = ref('')
const selectedVersions = ref([])
const showContentDialog = ref(false)
const selectedContent = ref('')
const currentVersion = ref(0)

const fetchVersions = async () => {
  try {
    const res = await axios.get(`/api/prompts/${route.params.id}/versions`)
    if (res.data.success) {
      versions.value = res.data.data
      if (versions.value.length > 0) {
        currentVersion.value = versions.value[0].version
      }
    }
  } catch (err) {
    ElMessage.error('获取版本历史失败')
  }
}

const fetchPromptInfo = async () => {
  try {
    const res = await axios.get(`/api/prompts/${route.params.id}`)
    if (res.data.success) {
      promptTitle.value = res.data.data.title
    }
  } catch (err) {
    console.error('Failed to fetch prompt info:', err)
  }
}

const viewContent = (version) => {
  selectedContent.value = version.content
  showContentDialog.value = true
}

const copyContent = () => {
  navigator.clipboard.writeText(selectedContent.value)
  ElMessage.success('已复制到剪贴板')
}

const toggleSelect = (version) => {
  const idx = selectedVersions.value.findIndex(v => v.id === version.id)
  if (idx >= 0) {
    selectedVersions.value.splice(idx, 1)
  } else {
    if (selectedVersions.value.length >= 2) {
      ElMessage.warning('最多选择两个版本进行对比')
      return
    }
    selectedVersions.value.push(version)
  }
}

const isSelected = (version) => {
  return selectedVersions.value.some(v => v.id === version.id)
}

const rollbackTo = async (version) => {
  try {
    await ElMessageBox.confirm(
      `确定恢复到 v${version.version} 吗？这将创建一个新版本。`,
      '恢复确认',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await axios.post(`/api/prompts/${route.params.id}/versions`, {
      content: version.content,
      comment: `回滚到 v${version.version}`
    })
    ElMessage.success('恢复成功')
    fetchVersions()
  } catch (err) {
    if (err !== 'cancel') ElMessage.error('恢复失败')
  }
}

const goBack = () => router.back()
const goToCompare = () => {
  const [v1, v2] = selectedVersions.value
  router.push({
    path: `/prompts/${route.params.id}/compare`,
    query: { v1: v1.id, v2: v2.id }
  })
}

onMounted(() => {
  fetchVersions()
  fetchPromptInfo()
})
</script>

<style scoped>
.version-history {
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

.title-area {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.title-area h2 {
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
  gap: var(--spacing-2);
}

.el-main {
  padding: var(--spacing-6);
}

.timeline-container {
  max-width: 800px;
  margin: 0 auto;
}

.version-card {
  cursor: pointer;
  transition: all var(--transition-normal);
  border: 1px solid var(--color-border);
}

.version-card:hover {
  transform: translateY(-2px);
  border-color: var(--color-border-hover);
}

.version-card.selected {
  border-color: var(--color-primary);
  background: var(--color-primary-light);
}

.version-card.current {
  border-left: 3px solid var(--color-success);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-3);
}

.card-header :deep(.el-checkbox__label) {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.version-number {
  font-weight: var(--font-weight-semibold);
  color: var(--color-primary);
}

.actions {
  display: flex;
  gap: var(--spacing-2);
}

.version-content {
  padding-top: var(--spacing-2);
}

.comment {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-2);
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  line-height: var(--line-height-relaxed);
}

.comment .el-icon {
  color: var(--color-text-muted);
  flex-shrink: 0;
  margin-top: 2px;
}

.no-comment {
  color: var(--color-text-muted);
  font-size: var(--font-size-sm);
  font-style: italic;
}

.content-viewer {
  background: var(--color-bg);
  border-radius: var(--radius-lg);
  padding: var(--spacing-4);
}

.content-viewer pre {
  margin: 0;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: var(--font-size-sm);
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 400px;
  overflow-y: auto;
}

.empty-state {
  padding: var(--spacing-12) 0;
}

/* Timeline customization */
:deep(.el-timeline-item__wrapper) {
  padding-left: 28px;
}

:deep(.el-timeline-item__node) {
  background: var(--color-primary);
}

:deep(.el-timeline-item__timestamp) {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}
</style>
