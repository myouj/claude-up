<template>
  <div class="cost-center">
    <div class="cost-header">
      <h3 class="cost-title">
        <el-icon><Coin /></el-icon>
        <span>配额使用</span>
      </h3>
      <el-button
        v-if="refreshable"
        size="small"
        text
        @click="$emit('refresh')"
      >
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </div>

    <div class="provider-list">
      <div
        v-for="provider in providers"
        :key="provider.id"
        class="provider-item"
      >
        <div class="provider-info">
          <span class="provider-name">{{ provider.name }}</span>
          <span class="provider-usage">
            {{ formatNumber(provider.used) }} / {{ formatNumber(provider.total) }}
          </span>
        </div>
        <div class="progress-bar">
          <div
            class="progress-fill"
            :class="getProgressClass(provider)"
            :style="{ width: getProgressPercent(provider) + '%' }"
          ></div>
        </div>
        <div class="provider-meta">
          <span class="remaining">{{ formatNumber(getRemaining(provider)) }} 剩余</span>
          <span class="percentage">{{ getProgressPercent(provider) }}%</span>
        </div>
      </div>

      <div v-if="providers.length === 0" class="empty-state">
        <el-icon class="empty-icon"><Wallet /></el-icon>
        <p>暂无配额数据</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  providers: {
    type: Array,
    default: () => []
  },
  refreshable: {
    type: Boolean,
    default: true
  }
})

defineEmits(['refresh'])

// Calculate progress percentage
const getProgressPercent = (provider) => {
  if (!provider.total || provider.total === 0) return 0
  return Math.min(Math.round((provider.used / provider.total) * 100), 100)
}

// Calculate remaining quota
const getRemaining = (provider) => {
  return Math.max(0, provider.total - provider.used)
}

// Get progress bar color class based on usage
const getProgressClass = (provider) => {
  const percent = getProgressPercent(provider)
  if (percent >= 90) return 'danger'
  if (percent >= 70) return 'warning'
  return 'success'
}

// Format large numbers with K/M suffixes
const formatNumber = (num) => {
  if (num === undefined || num === null) return '0'
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
  if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
  return num.toString()
}
</script>

<style scoped>
.cost-center {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-4);
}

.cost-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-4);
}

.cost-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0;
}

.provider-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4);
}

.provider-item {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.provider-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.provider-name {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
}

.provider-usage {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
}

.progress-bar {
  height: 8px;
  background: var(--color-border);
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  border-radius: 4px;
  transition: width var(--transition-normal), background-color var(--transition-fast);
}

.progress-fill.success {
  background: var(--color-success);
}

.progress-fill.warning {
  background: var(--color-warning);
}

.progress-fill.danger {
  background: var(--color-danger);
}

.provider-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.remaining {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.percentage {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-secondary);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--spacing-6);
  color: var(--color-text-muted);
}

.empty-icon {
  font-size: 32px;
  margin-bottom: var(--spacing-2);
  opacity: 0.4;
}

.empty-state p {
  font-size: var(--font-size-sm);
  margin: 0;
}
</style>
