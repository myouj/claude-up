<template>
  <div class="variable-preview-panel">
    <!-- Panel Header -->
    <div class="panel-header">
      <div class="header-title">
        <el-icon><View /></el-icon>
        <span>变量预览面板</span>
      </div>
    </div>

    <!-- Panel Body -->
    <div class="panel-body">
      <!-- Variable Input Section -->
      <div class="section">
        <div class="section-header">
          <div class="section-title">
            <el-icon><Edit /></el-icon>
            <span>变量输入</span>
          </div>
          <el-tag v-if="variables.length > 0" size="small" type="info">
            {{ filledCount }}/{{ variables.length }}
          </el-tag>
        </div>

        <!-- Fill Progress -->
        <div v-if="variables.length > 0" class="fill-progress">
          <div class="progress-bar">
            <div
              class="progress-fill"
              :style="{ width: fillRate + '%' }"
              :class="{ complete: fillRate === 100 }"
            ></div>
          </div>
          <span class="progress-text">{{ fillRate }}%</span>
        </div>

        <!-- Variable List -->
        <div v-if="variables.length > 0" class="variable-list">
          <div
            v-for="v in variables"
            :key="v"
            class="variable-item"
            :class="{ filled: hasValue(v) }"
          >
            <label class="var-label">
              <span class="var-marker">{{</span>
              <span class="var-name">{{ v }}</span>
              <span class="var-marker">}}</span>
              <el-icon v-if="hasValue(v)" class="check-icon"><Check /></el-icon>
            </label>
            <el-input
              v-model="variableValues[v]"
              :placeholder="`输入 ${v} 的值...`"
              size="small"
              clearable
            />
          </div>
        </div>

        <!-- No Variables State -->
        <div v-else class="no-variables">
          <el-icon class="no-var-icon"><CircleCheck /></el-icon>
          <p>当前内容无变量</p>
          <span>使用 &#123;&#123;变量名&#125;&#125; 语法定义变量</span>
        </div>
      </div>

      <!-- Rendered Preview Section -->
      <div class="section">
        <div class="section-header">
          <div class="section-title">
            <el-icon><View /></el-icon>
            <span>渲染预览</span>
          </div>
          <el-button
            v-if="variables.length > 0"
            size="small"
            :type="allFilled ? 'success' : 'default'"
            :disabled="!allFilled"
            @click="handleCopy"
          >
            <el-icon><CopyDocument /></el-icon>
            <span>复制</span>
          </el-button>
        </div>
        <div class="preview-box">
          <pre class="preview-text">{{ renderedContent || '无内容' }}</pre>
        </div>
      </div>

      <!-- Quality Score Cards (Collapsible) -->
      <div class="section quality-section" :class="{ collapsed: !qualityExpanded }">
        <div class="section-header clickable" @click="qualityExpanded = !qualityExpanded">
          <div class="section-title">
            <el-icon><DataAnalysis /></el-icon>
            <span>质量评分</span>
          </div>
          <el-icon class="collapse-icon" :class="{ rotated: qualityExpanded }">
            <ArrowRight />
          </el-icon>
        </div>
        <div v-show="qualityExpanded" class="quality-cards">
          <div class="quality-card">
            <span class="quality-label">Clarity</span>
            <div class="quality-bar">
              <div class="quality-fill" :style="{ width: '85%' }"></div>
            </div>
            <span class="quality-value">85</span>
          </div>
          <div class="quality-card">
            <span class="quality-label">Complete</span>
            <div class="quality-bar">
              <div class="quality-fill" :style="{ width: '72%' }"></div>
            </div>
            <span class="quality-value">72</span>
          </div>
          <div class="quality-card">
            <span class="quality-label">Example</span>
            <div class="quality-bar">
              <div class="quality-fill" :style="{ width: '60%' }"></div>
            </div>
            <span class="quality-value">60</span>
          </div>
          <div class="quality-card">
            <span class="quality-label">Role</span>
            <div class="quality-bar">
              <div class="quality-fill" :style="{ width: '90%' }"></div>
            </div>
            <span class="quality-value">90</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'

const props = defineProps({
  content: {
    type: String,
    default: ''
  }
})

const variableValues = ref({})
const qualityExpanded = ref(true)

// Extract unique variable names from content
const variables = computed(() => {
  const text = props.content || ''
  const regex = /\{\{([^}]+)\}\}/g
  const vars = new Set()
  let match
  while ((match = regex.exec(text)) !== null) {
    vars.add(match[1].trim())
  }
  return Array.from(vars)
})

const filledCount = computed(() => {
  return variables.value.filter(v => !!variableValues.value[v]).length
})

const hasValue = (varName) => !!variableValues.value[varName]

const allFilled = computed(() => {
  return variables.value.length > 0 &&
    variables.value.every(v => !!variableValues.value[v])
})

const fillRate = computed(() => {
  if (variables.value.length === 0) return 100
  return Math.round((filledCount.value / variables.value.length) * 100)
})

// Replace all {{var}} with their values
const renderedContent = computed(() => {
  let result = props.content || ''
  const varMap = Object.entries(variableValues.value).filter(([, v]) => v)
  if (varMap.length === 0) return result
  // Build a single regex to replace all variables at once
  const pattern = varMap.map(([key]) => `\\{\\{${key}\\}\\}`).join('|')
  const regex = new RegExp(pattern, 'g')
  return result.replace(regex, (match) => {
    const entry = varMap.find(([key]) => match === `{{${key}}}`)
    return entry ? entry[1] : match
  })
})

// Sync variable values when content changes (variables added/removed)
watch(variables, (newVars) => {
  const newKeys = new Set(newVars)
  const currentKeys = Object.keys(variableValues.value)
  if (currentKeys.some(key => !newKeys.has(key))) {
    const cleaned = {}
    for (const key of currentKeys) {
      if (newKeys.has(key)) {
        cleaned[key] = variableValues.value[key]
      }
    }
    variableValues.value = cleaned
  }
})

const handleCopy = () => {
  if (renderedContent.value) {
    navigator.clipboard.writeText(renderedContent.value)
    ElMessage.success('渲染结果已复制到剪贴板')
  }
}

// Expose for parent components
defineExpose({
  variableValues,
  renderedContent
})
</script>

<style scoped>
.variable-preview-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--color-surface);
  border-left: 1px solid var(--color-border);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-4);
  background: var(--color-bg);
  border-bottom: 1px solid var(--color-border);
}

.header-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-4);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-5);
}

.section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.section-header.clickable {
  cursor: pointer;
  user-select: none;
  padding: var(--spacing-2);
  margin: calc(-1 * var(--spacing-2));
  border-radius: var(--radius-md);
  transition: background var(--transition-fast);
}

.section-header.clickable:hover {
  background: var(--color-bg);
}

.section-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.collapse-icon {
  font-size: 12px;
  color: var(--color-text-muted);
  transition: transform var(--transition-fast);
}

.collapse-icon.rotated {
  transform: rotate(90deg);
}

/* Fill Progress */
.fill-progress {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.progress-bar {
  flex: 1;
  height: 6px;
  background: var(--color-border);
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--color-primary);
  border-radius: 3px;
  transition: width var(--transition-normal);
}

.progress-fill.complete {
  background: var(--color-success);
}

.progress-text {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-muted);
  min-width: 36px;
}

/* Variable List */
.variable-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.variable-item {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
  padding: var(--spacing-2) var(--spacing-3);
  border-radius: var(--radius-md);
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  transition: all var(--transition-fast);
}

.variable-item.filled {
  border-color: var(--color-success-light);
  background: color-mix(in srgb, var(--color-success-light) 15%, var(--color-bg));
}

.var-label {
  display: flex;
  align-items: center;
  gap: 2px;
  font-size: var(--font-size-xs);
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
}

.var-marker {
  color: var(--color-primary);
  font-weight: var(--font-weight-bold);
}

.var-name {
  color: var(--color-text-primary);
  font-weight: var(--font-weight-semibold);
}

.check-icon {
  margin-left: auto;
  color: var(--color-success);
  font-size: 12px;
}

/* No Variables */
.no-variables {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--spacing-6) var(--spacing-4);
  text-align: center;
  color: var(--color-text-muted);
}

.no-var-icon {
  font-size: 28px;
  margin-bottom: var(--spacing-2);
  opacity: 0.4;
}

.no-variables p {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  margin: 0 0 var(--spacing-1);
}

.no-variables span {
  font-size: var(--font-size-xs);
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  opacity: 0.7;
}

/* Preview Box */
.preview-box {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--spacing-3);
  max-height: 200px;
  overflow-y: auto;
}

.preview-text {
  margin: 0;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: var(--font-size-sm);
  line-height: 1.7;
  color: var(--color-text-primary);
  white-space: pre-wrap;
  word-break: break-word;
}

/* Quality Cards */
.quality-section.collapsed .quality-cards {
  display: none;
}

.quality-cards {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.quality-card {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  padding: var(--spacing-2);
  background: var(--color-bg);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
}

.quality-label {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-secondary);
  min-width: 56px;
}

.quality-bar {
  flex: 1;
  height: 4px;
  background: var(--color-border);
  border-radius: 2px;
  overflow: hidden;
}

.quality-fill {
  height: 100%;
  background: var(--color-primary);
  border-radius: 2px;
}

.quality-value {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  min-width: 24px;
  text-align: right;
}
</style>
