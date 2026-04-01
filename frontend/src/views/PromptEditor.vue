<template>
  <div class="prompt-editor">
    <BreadcrumbNav :items="[{ name: '提示词', path: '/prompts' }, { name: '编辑' }]" />
    <el-container>
      <el-header>
        <div class="header-content">
          <div class="left">
            <el-button class="mobile-menu-btn" @click="showSidebar = true">
              <el-icon><Menu /></el-icon>
            </el-button>
            <el-button class="back-btn" @click="goBack">
              <el-icon><ArrowLeft /></el-icon>
            </el-button>
            <div class="title-area">
              <el-input
                v-if="isEditingTitle"
                v-model="prompt.title"
                class="title-input"
                @blur="saveTitle"
                @keyup.enter="saveTitle"
                autofocus
              />
              <h2 v-else class="title" @click="isEditingTitle = true">
                {{ prompt.title || '未命名提示词' }}
                <el-icon class="edit-icon"><Edit /></el-icon>
              </h2>
            </div>
          </div>
          <div class="right">
            <el-button class="tool-btn" @click="goToVersions">
              <el-icon><Clock /></el-icon>
              <span class="btn-text">版本历史</span>
            </el-button>
            <el-button class="tool-btn" @click="goToCompare">
              <el-icon><Connection /></el-icon>
              <span class="btn-text">版本对比</span>
            </el-button>
            <el-button class="tool-btn" @click="goToTranslate">
              <el-icon><Translate /></el-icon>
              <span class="btn-text">翻译</span>
            </el-button>
            <el-button class="tool-btn" @click="goToTest">
              <el-icon><ChatDotRound /></el-icon>
              <span class="btn-text">测试</span>
            </el-button>
            <el-button class="tool-btn" @click="goToTestCompare">
              <el-icon><Connection /></el-icon>
              <span class="btn-text">对比测试</span>
            </el-button>
            <el-button class="tool-btn" @click="goToOptimize">
              <el-icon><MagicStick /></el-icon>
              <span class="btn-text">AI 优化</span>
            </el-button>
            <el-button class="tool-btn" @click="goToAnalytics">
              <el-icon><DataAnalysis /></el-icon>
              <span class="btn-text">分析</span>
            </el-button>
            <el-button type="primary" @click="handleSave">
              <el-icon><Check /></el-icon>
              <span class="btn-text">保存</span>
            </el-button>
          </div>
        </div>
      </el-header>

      <el-container>
        <el-aside width="300px" class="sidebar">
          <el-form :model="prompt" label-position="top" class="sidebar-form">
            <el-form-item label="描述">
              <el-input
                v-model="prompt.description"
                type="textarea"
                :rows="3"
                placeholder="简短描述这个提示词的用途"
                maxlength="500"
                show-word-limit
              />
            </el-form-item>

            <el-form-item label="分类">
              <el-select
                v-model="prompt.category"
                placeholder="选择分类"
                allow-create
                filterable
                clearable
                class="full-width"
              >
                <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
              </el-select>
            </el-form-item>

            <el-form-item label="标签">
              <el-select
                v-model="prompt.tags"
                multiple
                placeholder="添加标签"
                allow-create
                filterable
                clearable
                class="full-width"
              />
            </el-form-item>

            <el-form-item label="操作">
              <div class="action-buttons">
                <el-button
                  :type="prompt.is_favorite ? 'warning' : 'default'"
                  :icon="prompt.is_favorite ? 'Star' : 'StarFilled'"
                  @click="toggleFavorite"
                  class="action-btn"
                >
                  {{ prompt.is_favorite ? '已收藏' : '收藏' }}
                </el-button>
                <el-button
                  :type="prompt.is_pinned ? 'primary' : 'default'"
                  :icon="prompt.is_pinned ? 'Pin' : 'Pushpin'"
                  @click="togglePinned"
                  class="action-btn"
                >
                  {{ prompt.is_pinned ? '已置顶' : '置顶' }}
                </el-button>
              </div>
            </el-form-item>

            <VariablePreviewer
              ref="variablePreviewerRef"
              :content="prompt.content"
              @insert="insertVariables"
            />
          </el-form>

          <el-divider />

          <div class="info-section">
            <h4>提示词信息</h4>
            <div class="info-item">
              <span class="info-label">版本</span>
              <span class="info-value">v{{ versionCount }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">创建</span>
              <span class="info-value">{{ formatDate(prompt.created_at) }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">更新</span>
              <span class="info-value">{{ formatDate(prompt.updated_at) }}</span>
            </div>
          </div>
        </el-aside>

        <!-- Mobile sidebar drawer -->
        <el-drawer v-model="showSidebar" title="属性" size="300px" direction="ltr">
          <el-form :model="prompt" label-position="top" class="sidebar-form">
            <el-form-item label="描述">
              <el-input
                v-model="prompt.description"
                type="textarea"
                :rows="3"
                placeholder="简短描述这个提示词的用途"
                maxlength="500"
                show-word-limit
              />
            </el-form-item>
            <el-form-item label="分类">
              <el-select
                v-model="prompt.category"
                placeholder="选择分类"
                allow-create
                filterable
                clearable
                class="full-width"
              >
                <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
              </el-select>
            </el-form-item>
            <el-form-item label="标签">
              <el-select
                v-model="prompt.tags"
                multiple
                placeholder="添加标签"
                allow-create
                filterable
                clearable
                class="full-width"
              />
            </el-form-item>
            <el-form-item label="操作">
              <div class="action-buttons">
                <el-button
                  :type="prompt.is_favorite ? 'warning' : 'default'"
                  :icon="prompt.is_favorite ? 'Star' : 'StarFilled'"
                  @click="toggleFavorite"
                  class="action-btn"
                >
                  {{ prompt.is_favorite ? '已收藏' : '收藏' }}
                </el-button>
                <el-button
                  :type="prompt.is_pinned ? 'primary' : 'default'"
                  :icon="prompt.is_pinned ? 'Pin' : 'Pushpin'"
                  @click="togglePinned"
                  class="action-btn"
                >
                  {{ prompt.is_pinned ? '已置顶' : '置顶' }}
                </el-button>
              </div>
            </el-form-item>
            <VariablePreviewer
              ref="variablePreviewerRef"
              :content="prompt.content"
              @insert="insertVariables"
            />
            <el-divider />
            <div class="info-section">
              <h4>提示词信息</h4>
              <div class="info-item">
                <span class="info-label">版本</span>
                <span class="info-value">v{{ versionCount }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">创建</span>
                <span class="info-value">{{ formatDate(prompt.created_at) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">更新</span>
                <span class="info-value">{{ formatDate(prompt.updated_at) }}</span>
              </div>
            </div>
          </el-form>
        </el-drawer>

        <el-main>
          <div class="editor-container">
            <div class="editor-header">
              <div class="editor-title">
                <el-icon><Document /></el-icon>
                <span>提示词内容</span>
              </div>
              <div class="editor-actions">
                <el-input
                  v-model="saveComment"
                  placeholder="版本备注（可选）"
                  class="comment-input"
                />
                <el-button size="small" @click="insertTemplate">
                  <el-icon><Document /></el-icon>
                  插入模板
                </el-button>
              </div>
            </div>
            <el-input
              v-model="prompt.content"
              type="textarea"
              class="content-editor"
              placeholder="输入提示词内容...

提示词应该清晰地描述：
• 角色：你希望 AI 扮演什么角色
• 任务：需要 AI 完成什么
• 上下文：相关的背景信息
• 输出格式：期望的回复格式"
            />
          </div>
        </el-main>
      </el-container>
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
import VariablePreviewer from '../components/VariablePreviewer.vue'

const router = useRouter()
const route = useRoute()

const prompt = ref({
  id: null,
  title: '',
  content: '',
  description: '',
  category: '',
  tags: [],
  is_favorite: false,
  is_pinned: false,
  created_at: '',
  updated_at: ''
})
const versionCount = ref(0)
const isEditingTitle = ref(false)
const showSidebar = ref(false)
const saveComment = ref('')
const categories = ref([])
const variablePreviewerRef = ref(null)

const fetchPrompt = async () => {
  try {
    const res = await axios.get(`/api/prompts/${route.params.id}`)
    if (res.data.success) {
      prompt.value = res.data.data
    }
  } catch (err) {
    ElMessage.error('获取提示词失败')
  }
}

const fetchVersionCount = async () => {
  try {
    const res = await axios.get(`/api/prompts/${route.params.id}/versions`)
    if (res.data.success) {
      versionCount.value = res.data.data.length
    }
  } catch (err) {
    console.error('Failed to fetch versions:', err)
  }
}

const saveTitle = async () => {
  isEditingTitle.value = false
  if (!prompt.value.title.trim()) {
    ElMessage.warning('标题不能为空')
    return
  }
  await handleSave(false)
}

const handleSave = async (showMsg = true) => {
  if (!prompt.value.title.trim()) {
    ElMessage.warning('请填写标题')
    return
  }
  try {
    const res = await axios.put(`/api/prompts/${prompt.value.id}`, {
      title: prompt.value.title,
      content: prompt.value.content,
      description: prompt.value.description,
      category: prompt.value.category,
      tags: prompt.value.tags,
      comment: saveComment.value
    })
    if (res.data.success) {
      if (showMsg) ElMessage.success('保存成功')
      saveComment.value = ''
      fetchVersionCount()
    }
  } catch (err) {
    ElMessage.error('保存失败')
  }
}

const toggleFavorite = async () => {
  try {
    await axios.put(`/api/prompts/${prompt.value.id}`, { is_favorite: !prompt.value.is_favorite })
    prompt.value.is_favorite = !prompt.value.is_favorite
  } catch (err) {
    ElMessage.error('操作失败')
  }
}

const togglePinned = async () => {
  try {
    await axios.put(`/api/prompts/${prompt.value.id}`, { is_pinned: !prompt.value.is_pinned })
    prompt.value.is_pinned = !prompt.value.is_pinned
  } catch (err) {
    ElMessage.error('操作失败')
  }
}

const insertTemplate = () => {
  const template = `## Role
You are a [role/expertise]

## Task
[What you need the AI to do]

## Context
[Background information]

## Output Format
[Expected response structure]

## Constraints
- [Requirements or limitations]`
  prompt.value.content = prompt.value.content
    ? prompt.value.content + '\n\n' + template
    : template
}

const insertVariables = () => {
  if (variablePreviewerRef.value) {
    prompt.value.content = variablePreviewerRef.value.renderedContent.value
    variablePreviewerRef.value.clearValues()
    ElMessage.success('变量已替换')
  }
}

const goBack = () => router.back()
const goToVersions = () => router.push(`/prompts/${route.params.id}/versions`)
const goToCompare = () => router.push(`/prompts/${route.params.id}/compare`)
const goToTest = () => router.push(`/prompts/${route.params.id}/test`)
const goToTestCompare = () => router.push(`/prompts/${route.params.id}/test-compare`)
const goToOptimize = () => router.push(`/prompts/${route.params.id}/optimize`)
const goToTranslate = () => router.push(`/prompts/${route.params.id}/translate`)
const goToAnalytics = () => router.push(`/prompts/${route.params.id}/analytics`)

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return `${date.getFullYear()}/${date.getMonth() + 1}/${date.getDate()} ${date.getHours()}:${String(date.getMinutes()).padStart(2, '0')}`
}

onMounted(() => {
  fetchPrompt()
  fetchVersionCount()
})
</script>

<style scoped>
.prompt-editor {
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
  gap: var(--spacing-4);
}

.left {
  display: flex;
  align-items: center;
  gap: var(--spacing-4);
  min-width: 0;
}

.back-btn {
  flex-shrink: 0;
}

.title-area {
  min-width: 0;
}

.title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.edit-icon {
  color: var(--color-text-muted);
  opacity: 0;
  transition: opacity var(--transition-fast);
  flex-shrink: 0;
}

.title:hover .edit-icon {
  opacity: 1;
}

.title-input {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  width: 320px;
}

.right {
  display: flex;
  gap: var(--spacing-2);
  flex-shrink: 0;
}

.btn-text {
  margin-left: var(--spacing-1);
}

@media (max-width: 1024px) {
  .btn-text {
    display: none;
  }
}

.sidebar {
  background: var(--color-surface);
  padding: var(--spacing-5);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4);
}

.sidebar-form :deep(.el-form-item__label) {
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
  padding-bottom: var(--spacing-2);
}

.sidebar-form :deep(.el-form-item) {
  margin-bottom: var(--spacing-4);
}

.full-width {
  width: 100%;
}

.action-buttons {
  display: flex;
  gap: var(--spacing-2);
}

.action-btn {
  flex: 1;
}

.el-divider {
  margin: var(--spacing-2) 0;
}

.info-section {
  margin-top: auto;
}

.info-section h4 {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: var(--spacing-3);
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-2) 0;
  border-bottom: 1px solid var(--color-border);
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

.info-value {
  font-size: var(--font-size-sm);
  color: var(--color-text-primary);
  font-weight: var(--font-weight-medium);
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
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-3) var(--spacing-4);
  background: var(--color-bg);
  border-bottom: 1px solid var(--color-border);
}

.editor-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-secondary);
}

.editor-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.comment-input {
  width: 180px;
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
  background: var(--color-surface);
}

.content-editor :deep(.el-textarea__inner:focus) {
  box-shadow: none;
}

/* Responsive - Tablet */
@media (max-width: 1024px) {
  .right {
    gap: var(--spacing-1);
    overflow-x: auto;
    flex-shrink: 0;
  }
  .tool-btn {
    flex-shrink: 0;
  }
}

/* Responsive - Mobile */
@media (max-width: 768px) {
  .mobile-menu-btn {
    display: flex !important;
  }

  .sidebar {
    display: none;
  }

  .el-main {
    padding: var(--spacing-3);
  }

  .header-content {
    gap: var(--spacing-2);
  }

  .left {
    min-width: 0;
    overflow: hidden;
  }

  .back-btn :deep(.el-icon) {
    margin: 0;
  }

  .right {
    gap: var(--spacing-1);
    overflow-x: auto;
    flex-shrink: 0;
    padding-right: var(--spacing-2);
  }

  .right::-webkit-scrollbar {
    display: none;
  }

  .btn-text {
    display: none;
  }

  .title-area {
    min-width: 0;
  }

  .title {
    font-size: var(--font-size-md);
  }

  .title-input {
    width: 160px;
  }
}
</style>
