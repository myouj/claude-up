<template>
  <div class="ab-compare">
    <!-- Variant Cards -->
    <div class="variants-grid">
      <div
        v-for="variant in variants"
        :key="variant.id"
        class="variant-card"
        :class="{ winner: variant.id === winner }"
      >
        <div class="variant-header">
          <div class="variant-info">
            <el-tag
              v-if="variant.id === winner"
              type="success"
              size="small"
              effect="dark"
            >
              <el-icon><Trophy /></el-icon>
              胜出
            </el-tag>
            <span class="variant-name">{{ variant.name }}</span>
            <span class="variant-desc">{{ variant.description }}</span>
          </div>
          <div class="variant-score">
            <div class="score-value">{{ variant.metrics.avg_score.toFixed(1) }}</div>
            <div class="score-stars">
              <el-rate
                :model-value="variant.metrics.avg_score"
                disabled
                show-score
                :score-template="`${variant.metrics.avg_score.toFixed(1)}`"
              />
            </div>
          </div>
        </div>

        <!-- Metrics Bar -->
        <div class="metrics-row">
          <div class="metric">
            <span class="metric-label">运行次数</span>
            <span class="metric-value">{{ variant.runs }}</span>
          </div>
          <div class="metric">
            <span class="metric-label">平均延迟</span>
            <span class="metric-value">{{ variant.metrics.avg_latency }}ms</span>
          </div>
          <div class="metric">
            <span class="metric-label">Token 消耗</span>
            <span class="metric-value">{{ variant.metrics.token_usage }}</span>
          </div>
        </div>

        <!-- Content Preview -->
        <div class="content-preview">
          <div class="preview-label">Prompt 内容</div>
          <pre class="preview-text">{{ variant.content }}</pre>
        </div>

        <!-- Test Records -->
        <div v-if="variant.test_records && variant.test_records.length > 0" class="test-records">
          <div class="records-label">测试记录</div>
          <div
            v-for="record in variant.test_records"
            :key="record.id"
            class="record-item"
            @click="selectRecord(record, variant.id)"
            :class="{ selected: selectedRecords[variant.id]?.id === record.id }"
          >
            <div class="record-header">
              <span class="record-input">{{ truncate(record.input, 50) }}</span>
              <div class="record-meta">
                <el-rate :model-value="record.score" disabled size="small" />
                <span class="record-latency">{{ record.latency }}ms</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Selected Record Detail -->
        <div v-if="selectedRecords[variant.id]" class="record-detail">
          <div class="detail-header">
            <span>输入</span>
            <el-button size="small" text @click.stop="copyText(selectedRecords[variant.id].input)">
              <el-icon><CopyDocument /></el-icon>
            </el-button>
          </div>
          <pre class="detail-content input">{{ selectedRecords[variant.id].input }}</pre>

          <div class="detail-header">
            <span>AI 回复</span>
            <el-button size="small" text @click.stop="copyText(selectedRecords[variant.id].response)">
              <el-icon><CopyDocument /></el-icon>
            </el-button>
          </div>
          <pre class="detail-content response">{{ selectedRecords[variant.id].response }}</pre>
        </div>

        <!-- Scoring -->
        <div class="scoring-section">
          <div class="scoring-label">评分</div>
          <div class="scoring-controls">
            <el-rate v-model="scores[variant.id]" :max="5" show-text>
              <template #text-1>1 星</template>
              <template #text-2>2 星</template>
              <template #text-3>3 星</template>
              <template #text-4>4 星</template>
              <template #text-5>5 星</template>
            </el-rate>
            <el-input
              v-model="notes[variant.id]"
              type="textarea"
              :rows="2"
              placeholder="添加备注..."
              size="small"
            />
            <el-button
              type="primary"
              size="small"
              @click="submitScore(variant.id)"
              :disabled="!scores[variant.id]"
            >
              提交评分
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- Summary Stats -->
    <div class="summary-section">
      <div class="summary-title">
        <el-icon><DataAnalysis /></el-icon>
        统计汇总
      </div>
      <div class="summary-grid">
        <div class="summary-card">
          <div class="summary-label">总运行次数</div>
          <div class="summary-value">{{ totalRuns }}</div>
        </div>
        <div class="summary-card">
          <div class="summary-label">平均分差</div>
          <div class="summary-value">{{ scoreDiff }}</div>
        </div>
        <div class="summary-card">
          <div class="summary-label">胜出 Variant</div>
          <div class="summary-value">{{ winnerLabel }}</div>
        </div>
        <div class="summary-card">
          <div class="summary-label">胜出概率</div>
          <div class="summary-value">{{ winProbability }}%</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'

const props = defineProps({
  variants: {
    type: Array,
    required: true
  },
  winner: {
    type: String,
    default: null
  }
})

const scores = ref({})
const notes = ref({})
const selectedRecords = ref({})

const truncate = (text, max) => {
  if (!text) return ''
  return text.length > max ? text.substring(0, max) + '...' : text
}

const totalRuns = computed(() => {
  return props.variants.reduce((sum, v) => sum + v.runs, 0)
})

const scoreDiff = computed(() => {
  if (props.variants.length < 2) return '0.0'
  const scores = props.variants.map(v => v.metrics.avg_score)
  const diff = Math.abs(scores[0] - scores[1]).toFixed(1)
  return `+${diff}`
})

const winnerLabel = computed(() => {
  if (!props.winner) return '未确定'
  const v = props.variants.find(v => v.id === props.winner)
  return v ? v.name : '未知'
})

const winProbability = computed(() => {
  if (!props.winner || props.variants.length < 2) return '—'
  const winnerVariant = props.variants.find(v => v.id === props.winner)
  if (!winnerVariant) return '—'
  return Math.round((winnerVariant.runs / totalRuns.value) * 100)
})

const selectRecord = (record, variantId) => {
  if (selectedRecords.value[variantId]?.id === record.id) {
    delete selectedRecords.value[variantId]
  } else {
    selectedRecords.value[variantId] = record
  }
}

const copyText = (text) => {
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制到剪贴板')
}

const submitScore = (variantId) => {
  if (!scores.value[variantId]) return
  ElMessage.success(`Variant ${variantId.toUpperCase()} 评分已提交：${scores.value[variantId]} 星`)
}
</script>

<style scoped>
.ab-compare {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-6);
}

.variants-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: var(--spacing-5);
}

.variant-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-5);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4);
  transition: all var(--transition-normal);
}

.variant-card.winner {
  border-color: var(--color-success);
  box-shadow: 0 0 0 1px var(--color-success-light);
}

.variant-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--spacing-4);
}

.variant-info {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
}

.variant-name {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.variant-desc {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
}

.variant-score {
  text-align: right;
}

.score-value {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-primary);
  line-height: 1;
}

.variant-card.winner .score-value {
  color: var(--color-success);
}

.metrics-row {
  display: flex;
  gap: var(--spacing-4);
  padding: var(--spacing-3);
  background: var(--color-bg);
  border-radius: var(--radius-md);
}

.metric {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
}

.metric-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  margin-bottom: 2px;
}

.metric-value {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.content-preview {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.preview-label,
.records-label {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.preview-text {
  margin: 0;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: var(--font-size-xs);
  line-height: 1.6;
  color: var(--color-text-secondary);
  background: var(--color-bg);
  border-radius: var(--radius-md);
  padding: var(--spacing-3);
  max-height: 120px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-word;
}

.test-records {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.record-item {
  padding: var(--spacing-2) var(--spacing-3);
  background: var(--color-bg);
  border-radius: var(--radius-md);
  cursor: pointer;
  border: 1px solid transparent;
  transition: all var(--transition-fast);
}

.record-item:hover {
  border-color: var(--color-border-hover);
}

.record-item.selected {
  border-color: var(--color-primary);
  background: var(--color-primary-light);
}

.record-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-2);
}

.record-input {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.record-meta {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  flex-shrink: 0;
}

.record-latency {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.record-detail {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  padding: var(--spacing-3);
  background: var(--color-bg);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-muted);
}

.detail-content {
  margin: 0;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: var(--font-size-xs);
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  padding: var(--spacing-2);
  border-radius: var(--radius-sm);
}

.detail-content.input {
  background: color-mix(in srgb, var(--color-primary) 5%, var(--color-bg));
  color: var(--color-text-primary);
}

.detail-content.response {
  background: color-mix(in srgb, var(--color-success) 5%, var(--color-bg));
  color: var(--color-text-primary);
}

.scoring-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
  padding-top: var(--spacing-3);
  border-top: 1px solid var(--color-border);
}

.scoring-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.scoring-controls {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

/* Summary Section */
.summary-section {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-5);
}

.summary-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-4);
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-4);
}

.summary-card {
  text-align: center;
  padding: var(--spacing-3);
  background: var(--color-bg);
  border-radius: var(--radius-md);
}

.summary-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  margin-bottom: var(--spacing-1);
}

.summary-value {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-primary);
}

@media (max-width: 768px) {
  .variants-grid {
    grid-template-columns: 1fr;
  }

  .summary-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .metrics-row {
    gap: var(--spacing-2);
  }
}
</style>
