<template>
  <div class="variable-previewer" :class="{ collapsed: !expanded }">
    <div class="previewer-header" @click="expanded = !expanded">
      <div class="header-left">
        <el-icon class="collapse-icon" :class="{ rotated: expanded }">
          <ArrowRight />
        </el-icon>
        <el-icon><Edit /></el-icon>
        <span>变量预览</span>
        <el-tag v-if="variables.length > 0" size="small" type="info">
          {{ variables.length }} 个
        </el-tag>
        <div v-if="variables.length > 0 && expanded" class="fill-indicator">
          <div class="fill-bar">
            <div class="fill-progress" :style="{ width: fillRate + '%' }"></div>
          </div>
          <span class="fill-text">{{ fillRate }}%</span>
        </div>
      </div>
      <div class="header-right" @click.stop>
        <el-button
          v-if="variables.length > 0"
          size="small"
          :type="allFilled ? 'success' : 'default'"
          :disabled="!allFilled"
          @click="handleCopy"
        >
          <el-icon><CopyDocument /></el-icon>
          <span class="btn-text">复制渲染结果</span>
        </el-button>
        <el-button
          v-if="variables.length > 0"
          size="small"
          @click="clearValues"
        >
          <el-icon><RefreshRight /></el-icon>
          <span class="btn-text">清空</span>
        </el-button>
      </div>
    </div>

    <div v-show="expanded" class="previewer-body">
      <!-- Variable Inputs -->
      <div v-if="variables.length > 0" class="variable-inputs">
        <div
          v-for="v in variables"
          :key="v"
          class="variable-row"
          :class="{ filled: hasValue(v) }"
        >
          <label class="var-label">
            <span class="var-marker">&#123;&#123;</span>
            <span class="var-name">{{ v }}</span>
            <span class="var-marker">&#125;&#125;</span>
          </label>
          <el-input
            v-model="variableValues[v]"
            :placeholder="`输入 ${v} 的值...`"
            size="small"
            clearable
          >
            <template #suffix>
              <el-icon v-if="hasValue(v)" class="check-icon"><Check /></el-icon>
            </template>
          </el-input>
        </div>
      </div>

      <!-- Rendered Preview -->
      <div v-if="variables.length > 0" class="preview-section">
        <div class="preview-label">
          <el-icon><View /></el-icon>
          <span>渲染预览</span>
          <span v-if="!allFilled" class="preview-hint">(未填完)</span>
        </div>
        <div class="preview-content">
          <pre class="preview-text">{{ renderedContent || '无内容' }}</pre>
        </div>
      </div>

      <!-- No variables state -->
      <div v-else class="no-variables">
        <el-icon class="no-var-icon"><CircleCheck /></el-icon>
        <p>当前内容无变量</p>
        <span>使用 &#123;&#123;变量名&#125;&#125; 语法定义变量</span>
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

// Internal content ref that stays reactive
const internalContent = ref(props.content)

// Sync when prop changes
watch(() => props.content, (val) => {
  internalContent.value = val
})

const variableValues = ref({})

// Extract unique variable names from content
const variables = computed(() => {
  const text = internalContent.value || ''
  const regex = /\{\{([^}]+)\}\}/g
  const vars = new Set()
  let match
  while ((match = regex.exec(text)) !== null) {
    vars.add(match[1].trim())
  }
  return Array.from(vars)
})

// Replace all {{var}} with their values
const renderedContent = computed(() => {
  let result = internalContent.value || ''
  for (const [key, value] of Object.entries(variableValues.value)) {
    if (value) {
      result = result.replace(new RegExp(`\\{\\{${key}\\}\\}`, 'g'), value)
    }
  }
  return result
})

const hasValue = (varName) => !!variableValues.value[varName]

const allFilled = computed(() => {
  return variables.value.length > 0 &&
    variables.value.every(v => !!variableValues.value[v])
})

const fillRate = computed(() => {
  if (variables.value.length === 0) return 100
  const filled = variables.value.filter(v => !!variableValues.value[v]).length
  return Math.round((filled / variables.value.length) * 100)
})

const clearValues = () => {
  variableValues.value = {}
}

const expanded = ref(true)

const handleCopy = () => {
  if (renderedContent.value) {
    navigator.clipboard.writeText(renderedContent.value)
    ElMessage.success('渲染结果已复制到剪贴板')
  }
}

// Sync variable values when content changes (variables added/removed)
watch(variables, (newVars) => {
  const newKeys = new Set(newVars)
  const currentKeys = Object.keys(variableValues.value)
  for (const key of currentKeys) {
    if (!newKeys.has(key)) {
      delete variableValues.value[key]
    }
  }
})

// Expose for parent components
defineExpose({
  variableValues,
  renderedContent,
  clearValues
})
</script>

<style scoped>
.variable-previewer {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition: all var(--transition-normal);
}

.previewer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-3) var(--spacing-4);
  background: var(--color-bg);
  cursor: pointer;
  user-select: none;
  border-bottom: 1px solid transparent;
}

.variable-previewer:not(.collapsed) .previewer-header {
  border-bottom-color: var(--color-border);
}

.previewer-header:hover {
  background: var(--color-surface);
}

.header-left {
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

.fill-indicator {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin-left: var(--spacing-2);
}

.fill-bar {
  width: 60px;
  height: 4px;
  background: var(--color-border);
  border-radius: 2px;
  overflow: hidden;
}

.fill-progress {
  height: 100%;
  background: var(--color-primary);
  border-radius: 2px;
  transition: width var(--transition-normal);
}

.fill-text {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  font-weight: var(--font-weight-medium);
}

.header-right {
  display: flex;
  gap: var(--spacing-2);
}

.previewer-body {
  padding: var(--spacing-4);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4);
}

.variable-inputs {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.variable-row {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
  padding: var(--spacing-2);
  border-radius: var(--radius-md);
  background: var(--color-bg);
  border: 1px solid transparent;
  transition: all var(--transition-fast);
}

.variable-row.filled {
  border-color: var(--color-success-light);
  background: color-mix(in srgb, var(--color-success-light) 30%, var(--color-bg));
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
  color: var(--color-success);
}

.preview-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.preview-label {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-secondary);
}

.preview-hint {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-normal);
  color: var(--color-warning);
}

.preview-content {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--spacing-3);
  max-height: 300px;
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

.no-variables {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--spacing-6) var(--spacing-4);
  text-align: center;
  color: var(--color-text-muted);
}

.no-var-icon {
  font-size: 32px;
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

@media (max-width: 768px) {
  .previewer-header {
    padding: var(--spacing-2) var(--spacing-3);
  }

  .fill-indicator {
    display: none;
  }

  .header-right .btn-text {
    display: none;
  }

  .previewer-body {
    padding: var(--spacing-3);
  }
}
</style>
