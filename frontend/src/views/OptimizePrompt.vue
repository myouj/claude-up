<template>
  <div class="optimize-prompt">
    <BreadcrumbNav :items="[{ name: '提示词', path: '/prompts' }, { name: 'AI 优化' }]" />
    <el-header>
      <div class="header-content">
        <div class="left">
          <el-button class="back-btn" @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <h2>AI 优化</h2>
        </div>
        <div class="right">
          <el-select v-model="optimizeMode" class="mode-select">
            <el-option label="一键优化" value="improve" />
            <el-option label="结构优化" value="structure" />
            <el-option label="风格调整" value="style" />
            <el-option label="优化建议" value="suggest" />
          </el-select>
          <el-button type="primary" @click="handleOptimize" :loading="loading">
            <el-icon><MagicStick /></el-icon>
            {{ optimizeMode === 'suggest' ? '获取建议' : '优化' }}
          </el-button>
        </div>
      </div>
    </el-header>

    <el-main>
      <div class="optimize-container">
        <div class="optimize-panel original">
          <div class="panel-header">
            <div class="panel-title">
              <el-icon><Document /></el-icon>
              <h3>原始提示词</h3>
            </div>
            <el-tag type="info" effect="plain">v{{ currentVersion }}</el-tag>
          </div>
          <el-input
            v-model="originalContent"
            type="textarea"
            :rows="14"
            placeholder="原始提示词内容..."
            class="content-input"
          />
        </div>

        <div class="optimize-panel optimized">
          <div class="panel-header">
            <div class="panel-title">
              <el-icon><MagicStick /></el-icon>
              <h3>{{ optimizeMode === 'suggest' ? '优化建议' : '优化结果' }}</h3>
            </div>
            <div class="actions" v-if="optimizedContent">
              <el-button size="small" @click="copyOptimized">
                <el-icon><CopyDocument /></el-icon>
                复制
              </el-button>
              <el-button
                size="small"
                type="primary"
                @click="applyOptimized"
                :disabled="optimizeMode === 'suggest'"
              >
                <el-icon><Check /></el-icon>
                应用
              </el-button>
            </div>
          </div>
          <div v-if="optimizedContent" class="optimized-content">
            <pre v-if="optimizeMode === 'suggest'" class="suggestion-content">{{ optimizedContent }}</pre>
            <el-input
              v-else
              v-model="optimizedContent"
              type="textarea"
              :rows="14"
              class="content-input"
            />
          </div>
          <div v-else class="empty-state">
            <div class="empty-icon">
              <svg width="80" height="80" viewBox="0 0 80 80" fill="none">
                <path d="M40 10L50 30H70L54 44L60 66L40 52L20 66L26 44L10 30H30L40 10Z" stroke="var(--color-border)" stroke-width="2" stroke-linejoin="round"/>
              </svg>
            </div>
            <p>点击"优化"按钮获取 AI 优化结果</p>
            <span>AI 将帮助你改进提示词</span>
          </div>
        </div>
      </div>

      <div class="mode-descriptions">
        <el-card class="mode-card">
          <template #header>
            <div class="card-header">
              <el-icon><InfoFilled /></el-icon>
              <span>优化模式说明</span>
            </div>
          </template>
          <div class="mode-grid">
            <div class="mode-item">
              <div class="mode-icon improve">
                <el-icon><MagicStick /></el-icon>
              </div>
              <div class="mode-content">
                <h4>一键优化</h4>
                <p>对提示词进行全面优化，提升清晰度、有效性和具体性</p>
              </div>
            </div>
            <div class="mode-item">
              <div class="mode-icon structure">
                <el-icon><Tickets /></el-icon>
              </div>
              <div class="mode-content">
                <h4>结构优化</h4>
                <p>添加角色定义、上下文、任务描述、输出格式和约束条件</p>
              </div>
            </div>
            <div class="mode-item">
              <div class="mode-icon style">
                <el-icon><EditPen /></el-icon>
              </div>
              <div class="mode-content">
                <h4>风格调整</h4>
                <p>调整提示词的语气、长度和风格</p>
              </div>
            </div>
            <div class="mode-item">
              <div class="mode-icon suggest">
                <el-icon><ChatLineSquare /></el-icon>
              </div>
              <div class="mode-content">
                <h4>优化建议</h4>
                <p>列出 3-5 条具体可行的改进建议，而非直接修改</p>
              </div>
            </div>
          </div>
        </el-card>
      </div>
    </el-main>
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

const originalContent = ref('')
const optimizedContent = ref('')
const optimizeMode = ref('improve')
const loading = ref(false)
const currentVersion = ref(1)

const fetchPrompt = async () => {
  try {
    const res = await axios.get(`/api/prompts/${route.params.id}`)
    if (res.data.success) {
      originalContent.value = res.data.data.content
    }
  } catch (err) {
    ElMessage.error('获取提示词失败')
  }
}

const handleOptimize = async () => {
  if (!originalContent.value.trim()) {
    ElMessage.warning('请输入提示词内容')
    return
  }

  loading.value = true
  try {
    const res = await axios.post(`/api/prompts/${route.params.id}/optimize`, {
      content: originalContent.value,
      mode: optimizeMode.value
    })

    if (res.data.success) {
      optimizedContent.value = res.data.data.optimized
    }
  } catch (err) {
    ElMessage.error('优化请求失败')
  } finally {
    loading.value = false
  }
}

const copyOptimized = () => {
  navigator.clipboard.writeText(optimizedContent.value)
  ElMessage.success('已复制到剪贴板')
}

const applyOptimized = async () => {
  try {
    await axios.put(`/api/prompts/${route.params.id}`, {
      content: optimizedContent.value,
      comment: `AI ${optimizeMode.value} 模式优化`
    })
    ElMessage.success('已应用到提示词')
    originalContent.value = optimizedContent.value
  } catch (err) {
    ElMessage.error('应用失败')
  }
}

const goBack = () => router.back()

onMounted(fetchPrompt)
</script>

<style scoped>
.optimize-prompt {
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

.back-btn {
  padding: var(--spacing-2);
}

.right {
  display: flex;
  gap: var(--spacing-3);
}

.mode-select {
  width: 140px;
}

.el-main {
  padding: var(--spacing-5);
}

.optimize-container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-5);
  margin-bottom: var(--spacing-5);
}

.optimize-panel {
  background: var(--color-surface);
  border-radius: var(--radius-lg);
  padding: var(--spacing-5);
  border: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-4);
}

.panel-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.panel-title h3 {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0;
}

.actions {
  display: flex;
  gap: var(--spacing-2);
}

.content-input :deep(.el-textarea__inner) {
  flex: 1;
  font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
  font-size: var(--font-size-sm);
  line-height: 1.7;
}

.optimized-content {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.suggestion-content {
  flex: 1;
  background: var(--color-bg);
  padding: var(--spacing-4);
  border-radius: var(--radius-md);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  line-height: var(--line-height-relaxed);
  white-space: pre-wrap;
  word-break: break-word;
  margin: 0;
  overflow-y: auto;
}

.empty-state {
  flex: 1;
  min-height: 300px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  text-align: center;
}

.empty-icon {
  margin-bottom: var(--spacing-4);
  opacity: 0.5;
}

.empty-state p {
  font-size: var(--font-size-md);
  margin-bottom: var(--spacing-1);
}

.empty-state span {
  font-size: var(--font-size-xs);
}

.mode-descriptions {
  margin-top: var(--spacing-2);
}

.mode-card :deep(.el-card__header) {
  padding: var(--spacing-3) var(--spacing-4);
  border-bottom: 1px solid var(--color-border);
}

.card-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-secondary);
}

.mode-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-4);
}

@media (max-width: 1200px) {
  .mode-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.mode-item {
  display: flex;
  gap: var(--spacing-3);
  padding: var(--spacing-3);
  background: var(--color-bg);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.mode-item:hover {
  background: var(--color-primary-light);
}

.mode-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.mode-icon.improve {
  background: var(--color-primary-light);
  color: var(--color-primary);
}

.mode-icon.structure {
  background: var(--color-success-light);
  color: var(--color-success);
}

.mode-icon.style {
  background: var(--color-warning-light);
  color: var(--color-warning);
}

.mode-icon.suggest {
  background: var(--color-info-light);
  color: var(--color-info);
}

.mode-content h4 {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-1) 0;
}

.mode-content p {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  margin: 0;
  line-height: var(--line-height-relaxed);
}
</style>
