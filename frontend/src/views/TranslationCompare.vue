<template>
  <div class="translation-compare">
    <BreadcrumbNav :items="breadcrumbItems" />
    <el-header>
      <div class="header-content">
        <div class="left">
          <el-button @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <h2>{{ entityTitle }} - 翻译对比</h2>
        </div>
        <div class="right">
          <el-select v-model="sourceLang" class="lang-select">
            <el-option label="英文" value="en" />
            <el-option label="中文" value="zh" />
          </el-select>
          <span class="arrow">→</span>
          <el-select v-model="targetLang" class="lang-select">
            <el-option label="中文" value="zh" />
            <el-option label="英文" value="en" />
          </el-select>
          <el-button type="primary" @click="handleTranslate" :loading="translating">
            <el-icon><Translate /></el-icon>
            {{ targetLang === 'zh' ? '翻译为中文' : '翻译为英文' }}
          </el-button>
        </div>
      </div>
    </el-header>

    <el-main>
      <div class="compare-container">
        <div class="compare-panel source">
          <div class="panel-header">
            <span class="lang-label">{{ sourceLang === 'en' ? 'English' : '中文' }}</span>
            <el-button size="small" @click="copySource">
              <el-icon><CopyDocument /></el-icon>
              复制
            </el-button>
          </div>
          <div class="panel-content">
            <pre>{{ sourceText }}</pre>
          </div>
        </div>

        <div class="compare-panel target">
          <div class="panel-header">
            <span class="lang-label">{{ targetLang === 'zh' ? '中文' : 'English' }}</span>
            <div class="panel-actions">
              <el-button size="small" @click="copyTarget">
                <el-icon><CopyDocument /></el-icon>
                复制
              </el-button>
              <el-button
                v-if="targetText && !isSaved"
                size="small"
                type="primary"
                @click="handleApply"
              >
                <el-icon><Check /></el-icon>
                应用翻译
              </el-button>
            </div>
          </div>
          <div class="panel-content">
            <div v-if="loading" class="loading-state">
              <el-icon class="is-loading"><Loading /></el-icon>
              <span>翻译中...</span>
            </div>
            <pre v-else-if="targetText">{{ targetText }}</pre>
            <div v-else class="empty-state">
              <el-icon><Translate /></el-icon>
              <p>点击"翻译"按钮获取翻译结果</p>
            </div>
          </div>
        </div>
      </div>
    </el-main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import BreadcrumbNav from '../components/BreadcrumbNav.vue'

const breadcrumbItems = computed(() => {
  const type = route.params.type
  if (type === 'prompts') return [{ name: '提示词', path: '/prompts' }, { name: '翻译' }]
  if (type === 'skills') return [{ name: 'Skills', path: '/skills' }, { name: '翻译' }]
  if (type === 'agents') return [{ name: 'Agents', path: '/agents' }, { name: '翻译' }]
  return []
})

const router = useRouter()
const route = useRoute()

const entityType = ref('prompt') // prompt/skill/agent
const entityId = ref('')
const entityTitle = ref('')
const sourceLang = ref('en')
const targetLang = ref('zh')
const sourceText = ref('')
const targetText = ref('')
const translating = ref(false)
const loading = ref(false)
const isSaved = ref(false)

const fetchEntity = async () => {
  // 根据路由名称判断实体类型
  const routeName = route.name || ''
  if (routeName.includes('Skill')) {
    entityType.value = 'skill'
  } else if (routeName.includes('Agent')) {
    entityType.value = 'agent'
  } else {
    entityType.value = 'prompt'
  }
  entityId.value = route.params.id

  const endpoints = {
    prompt: `/api/prompts/${entityId.value}`,
    skill: `/api/skills/${entityId.value}`,
    agent: `/api/agents/${entityId.value}`
  }

  try {
    const res = await axios.get(endpoints[entityType.value])
    if (res.data.success) {
      const data = res.data.data
      entityTitle.value = data.name || data.title || data.role || ''

      if (sourceLang.value === 'en') {
        sourceText.value = data.content || ''
        targetText.value = data.content_cn || ''
      } else {
        sourceText.value = data.content_cn || data.content || ''
        targetText.value = data.content || ''
      }

      if (targetText.value) {
        isSaved.value = true
      }
    }
  } catch (err) {
    ElMessage.error('获取数据失败')
  }
}

const handleTranslate = async () => {
  if (!sourceText.value) {
    ElMessage.warning('源文本为空')
    return
  }

  translating.value = true
  try {
    const res = await axios.post(`/api/translate/${entityType.value}/${entityId.value}`, {
      source_lang: sourceLang.value,
      target_lang: targetLang.value
    })

    if (res.data.success) {
      targetText.value = res.data.data.target_text
      isSaved.value = false
      ElMessage.success('翻译成功')
    }
  } catch (err) {
    ElMessage.error('翻译失败')
  } finally {
    translating.value = false
  }
}

const handleApply = async () => {
  if (!targetText.value) return

  const endpoints = {
    prompt: `/api/prompts/${entityId.value}`,
    skill: `/api/skills/${entityId.value}`,
    agent: `/api/agents/${entityId.value}`
  }

  try {
    const updateData = targetLang.value === 'zh'
      ? { content_cn: targetText.value }
      : {}

    await axios.put(endpoints[entityType.value], updateData)
    isSaved.value = true
    ElMessage.success('翻译已保存')
  } catch (err) {
    ElMessage.error('保存失败')
  }
}

const copySource = () => {
  navigator.clipboard.writeText(sourceText.value)
  ElMessage.success('已复制')
}

const copyTarget = () => {
  navigator.clipboard.writeText(targetText.value)
  ElMessage.success('已复制')
}

const goBack = () => {
  const routes = {
    prompt: '/prompts',
    skill: '/skills',
    agent: '/agents'
  }
  router.push(`${routes[entityType.value]}/${entityId.value}`)
}

onMounted(fetchEntity)
</script>

<style scoped>
.translation-compare {
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

.right {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.lang-select {
  width: 100px;
}

.arrow {
  color: var(--color-text-muted);
  font-weight: var(--font-weight-semibold);
}

.el-main {
  padding: var(--spacing-5);
  height: calc(100vh - 64px);
}

.compare-container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-4);
  height: 100%;
}

.compare-panel {
  display: flex;
  flex-direction: column;
  background: var(--color-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
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

.lang-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-secondary);
}

.panel-actions {
  display: flex;
  gap: var(--spacing-2);
}

.panel-content {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-4);
}

.panel-content pre {
  margin: 0;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: var(--font-size-sm);
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
  color: var(--color-text-primary);
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: var(--color-text-muted);
  gap: var(--spacing-2);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: var(--color-text-muted);
  text-align: center;
}

.empty-state .el-icon {
  font-size: 48px;
  margin-bottom: var(--spacing-3);
}

@media (max-width: 768px) {
  .compare-container {
    grid-template-columns: 1fr;
  }
}
</style>
