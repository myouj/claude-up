<template>
  <div class="ab-sequential-panel">
    <!-- Variant A/B Progress Section -->
    <div class="variant-progress-section">
      <div class="section-title">
        <el-icon><DataAnalysis /></el-icon>
        变体进度
      </div>
      <div class="variants-split">
        <!-- Variant A -->
        <div class="variant-column variant-a">
          <div class="variant-label">
            <span class="variant-tag" style="background: #2563EB; color: white;">A</span>
            <span class="variant-name">Variant A</span>
          </div>
          <div class="progress-bar-container">
            <div class="progress-bar" :style="{ width: progressA + '%', background: '#2563EB' }"></div>
          </div>
          <div class="progress-stats">
            <span class="runs-count">{{ resultsA.length }} 次</span>
            <span class="avg-score">{{ avgScoreA.toFixed(1) }} 分</span>
          </div>
        </div>

        <!-- Divider -->
        <div class="split-divider">
          <div class="vs-badge">VS</div>
        </div>

        <!-- Variant B -->
        <div class="variant-column variant-b">
          <div class="variant-label">
            <span class="variant-tag" style="background: #F97316; color: white;">B</span>
            <span class="variant-name">Variant B</span>
          </div>
          <div class="progress-bar-container">
            <div class="progress-bar" :style="{ width: progressB + '%', background: '#F97316' }"></div>
          </div>
          <div class="progress-stats">
            <span class="runs-count">{{ resultsB.length }} 次</span>
            <span class="avg-score">{{ avgScoreB.toFixed(1) }} 分</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Sequential Test Progress Section -->
    <div class="sequential-progress-section">
      <div class="section-title">
        <el-icon><TrendCharts /></el-icon>
        序贯检验进度
      </div>

      <div class="sequential-stats">
        <div class="stat-item">
          <span class="stat-label">最低样本</span>
          <span class="stat-value">{{ props.minSamples || 15 }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">已完成</span>
          <span class="stat-value">{{ props.completedSamples || completedRuns }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">最大样本</span>
          <span class="stat-value">{{ props.maxSamples || 50 }}</span>
        </div>
      </div>

      <!-- Sequential Progress Bar -->
      <div class="sequential-bar-container">
        <div
          class="sequential-bar-bg"
          :style="{ '--min-pos': minPosition + '%', '--max-pos': maxPosition + '%' }"
        >
          <!-- Min threshold marker -->
          <div class="threshold-min">
            <div class="threshold-line"></div>
            <span class="threshold-label">最低</span>
          </div>
          <!-- Max threshold marker -->
          <div class="threshold-max">
            <div class="threshold-line"></div>
            <span class="threshold-label">最大</span>
          </div>
          <!-- Current progress -->
          <div
            class="sequential-bar-fill"
            :style="{ width: sequentialPercent + '%', background: significanceColor }"
          ></div>
        </div>
      </div>

      <!-- Significance Info Row -->
      <div class="significance-row">
        <div class="significance-indicator">
          <div class="significance-bar">
            <div
              class="significance-fill"
              :style="{ width: significancePercent + '%', background: significanceColor }"
            ></div>
          </div>
          <span class="significance-label" :style="{ color: significanceColor }">
            {{ significanceText }}
          </span>
        </div>
        <span class="p-value">
          p {{ pValueText }}
        </span>
      </div>
    </div>

    <!-- Confidence Interval & Stats -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon" :style="{ background: significanceBgColor }">
          <el-icon :style="{ color: significanceColor }"><TrendCharts /></el-icon>
        </div>
        <div class="stat-content">
          <span class="stat-label">统计显著性</span>
          <span class="stat-value" :style="{ color: significanceColor }">{{ significanceLabel }}</span>
        </div>
      </div>
    </div>

    <!-- Recommendation -->
    <div v-if="recommendation" class="recommendation" :class="recommendation.type">
      <el-icon><InfoFilled /></el-icon>
      <span>{{ recommendation.text }}</span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { ChatLineSquare, TrendCharts, DataAnalysis, InfoFilled } from '@element-plus/icons-vue'

const props = defineProps({
  resultsA: {
    type: Array,
    required: true
  },
  resultsB: {
    type: Array,
    required: true
  },
  minSamples: {
    type: Number,
    default: 15
  },
  maxSamples: {
    type: Number,
    default: 50
  },
  completedSamples: {
    type: Number,
    default: null
  },
  ciLower: {
    type: Number,
    default: 0
  },
  ciUpper: {
    type: Number,
    default: 0
  },
  winner: {
    type: String,
    default: 'none'
  },
  pValue: {
    type: Number,
    default: null
  }
})

// Calculate total runs
const completedRuns = computed(() => props.resultsA.length + props.resultsB.length)

// Variant progress percentages
const progressA = computed(() => (props.resultsA.length / props.maxSamples) * 100)
const progressB = computed(() => (props.resultsB.length / props.maxSamples) * 100)

// Average scores
const avgScoreA = computed(() => {
  if (props.resultsA.length === 0) return 0
  const sum = props.resultsA.reduce((acc, r) => acc + r.score, 0)
  return sum / props.resultsA.length
})

const avgScoreB = computed(() => {
  if (props.resultsB.length === 0) return 0
  const sum = props.resultsB.reduce((acc, r) => acc + r.score, 0)
  return sum / props.resultsB.length
})

// Sequential progress
const minPosition = computed(() => (props.minSamples / props.maxSamples) * 100)
const maxPosition = computed(() => 100)

const sequentialPercent = computed(() => {
  const completed = props.completedSamples ?? completedRuns.value
  return Math.min(100, (completed / props.maxSamples) * 100)
})

// Confidence interval calculations
const ciLeft = computed(() => Math.max(0, (props.ciLower + 0.5) * 100))
const ciWidth = computed(() => Math.max(0, (props.ciUpper - props.ciLower) * 100))

// Significance type based on p-value thresholds
const significanceType = computed(() => {
  const pVal = props.pValue
  if (pVal !== null && pVal < 0.05) return 'success'
  if (pVal !== null && pVal < 0.1) return 'warning'
  return 'info'
})

// Significance label based on p-value thresholds
const significanceLabel = computed(() => {
  const pVal = props.pValue
  if (pVal !== null && pVal < 0.05) return '显著'
  if (pVal !== null && pVal < 0.1) return '接近'
  return '不足'
})

const significanceColor = computed(() => {
  switch (significanceType.value) {
    case 'success': return '#10B981' // green
    case 'warning': return '#F59E0B' // yellow
    default: return '#EF4444' // red
  }
})

const significanceBgColor = computed(() => {
  switch (significanceType.value) {
    case 'success': return '#ECFDF5'
    case 'warning': return '#FFFBEB'
    default: return '#FEF2F2'
  }
})

const significanceText = computed(() => {
  switch (significanceType.value) {
    case 'success': return '显著 (p < 0.05)'
    case 'warning': return '接近显著'
    default: return '不足'
  }
})

const significanceStatus = computed(() => {
  switch (significanceType.value) {
    case 'success': return '已验证'
    case 'warning': return '进行中'
    default: return '待积累'
  }
})

const significancePercent = computed(() => {
  switch (significanceType.value) {
    case 'success': return 100
    case 'warning': return 65
    default: return 25
  }
})

const pValueText = computed(() => {
  const pVal = props.pValue
  if (pVal !== null) {
    return `= ${pVal.toFixed(4)}`
  }
  // Estimate p-value based on score difference
  const diff = Math.abs(avgScoreA.value - avgScoreB.value)
  if (diff >= 1.5) return '< 0.01'
  if (diff >= 1.0) return '< 0.05'
  if (diff >= 0.5) return '< 0.10'
  return '> 0.10'
})

const recommendation = computed(() => {
  if (significanceType.value === 'success') {
    return {
      type: 'success',
      text: `Variant ${avgScoreB.value > avgScoreA.value ? 'B' : 'A'} 表现显著优于另一方，建议采用。`
    }
  }
  if (significanceType.value === 'warning' && completedRuns.value >= props.maxSamples) {
    return {
      type: 'warning',
      text: '样本已达最大量，但差异未达统计显著水平，建议增加测试次数或接受当前结果。'
    }
  }
  if (completedRuns.value < props.minSamples) {
    return {
      type: 'info',
      text: `还需 ${props.minSamples - completedRuns.value} 次样本才能进行有效的序贯检验。`
    }
  }
  return null
})
</script>

<style scoped>
.ab-sequential-panel {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-5);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-5);
}

.section-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-3);
}

/* Variant Progress Section */
.variant-progress-section {
  padding-bottom: var(--spacing-4);
  border-bottom: 1px solid var(--color-border);
}

.variants-split {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: var(--spacing-4);
  align-items: center;
}

.variant-column {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.variant-label {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.variant-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-bold);
  border-radius: 4px;
}

.variant-name {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.progress-bar-container {
  height: 8px;
  background: var(--color-bg);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.progress-bar {
  height: 100%;
  border-radius: var(--radius-full);
  transition: width var(--transition-slow);
}

.progress-stats {
  display: flex;
  justify-content: space-between;
  font-size: var(--font-size-xs);
}

.runs-count {
  color: var(--color-text-secondary);
}

.avg-score {
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.split-divider {
  display: flex;
  align-items: center;
  justify-content: center;
}

.vs-badge {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg);
  border-radius: var(--radius-full);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-muted);
}

/* Sequential Progress Section */
.sequential-progress-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.sequential-stats {
  display: flex;
  justify-content: space-between;
  padding: var(--spacing-3);
  background: var(--color-bg);
  border-radius: var(--radius-md);
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.stat-item .stat-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.stat-item .stat-value {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
}

.sequential-bar-container {
  padding: var(--spacing-2) 0;
}

.sequential-bar-bg {
  position: relative;
  height: 12px;
  background: var(--color-bg);
  border-radius: var(--radius-full);
  overflow: visible;
}

.threshold-min,
.threshold-max {
  position: absolute;
  top: -20px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.threshold-min {
  left: var(--min-pos);
  transform: translateX(-50%);
}

.threshold-max {
  left: var(--max-pos);
  transform: translateX(-50%);
}

.threshold-line {
  width: 2px;
  height: 28px;
  background: var(--color-border);
}

.threshold-label {
  font-size: 10px;
  color: var(--color-text-muted);
  margin-top: 2px;
}

.sequential-bar-fill {
  height: 100%;
  border-radius: var(--radius-full);
  transition: width var(--transition-slow), background var(--transition-normal);
}

/* Significance Row */
.significance-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-4);
}

.significance-indicator {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  flex: 1;
}

.significance-bar {
  flex: 1;
  height: 6px;
  background: var(--color-bg);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.significance-fill {
  height: 100%;
  border-radius: var(--radius-full);
  transition: width var(--transition-slow), background var(--transition-normal);
}

.significance-label {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  white-space: nowrap;
}

.p-value {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
}

/* Stats Row */
.stats-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--spacing-3);
}

.stat-card {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  padding: var(--spacing-3);
  background: var(--color-bg);
  border-radius: var(--radius-md);
}

.stat-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  flex-shrink: 0;
}

.stat-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.stat-content .stat-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.stat-content .stat-value {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Recommendation */
.recommendation {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-3);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
}

.recommendation.success {
  background: var(--color-success-light);
  color: var(--color-success);
}

.recommendation.warning {
  background: var(--color-warning-light);
  color: var(--color-warning);
}

.recommendation.info {
  background: var(--color-info-light);
  color: var(--color-info);
}

/* Responsive */
@media (max-width: 640px) {
  .variants-split {
    grid-template-columns: 1fr;
    gap: var(--spacing-3);
  }

  .split-divider {
    transform: rotate(90deg);
    padding: var(--spacing-1) 0;
  }

  .stats-row {
    grid-template-columns: 1fr;
  }

  .sequential-stats {
    flex-wrap: wrap;
    gap: var(--spacing-2);
  }

  .stat-item {
    flex: 1;
    min-width: 80px;
  }
}
</style>
