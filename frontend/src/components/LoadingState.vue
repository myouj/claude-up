<template>
  <div class="loading-state">
    <div v-if="type === 'spinner'" class="spinner-container">
      <el-icon class="is-loading spinner-icon">
        <Loading />
      </el-icon>
      <span v-if="text" class="loading-text">{{ text }}</span>
    </div>
    <div v-else-if="type === 'skeleton'" class="skeleton-container">
      <div v-for="i in skeletonRows" :key="i" class="skeleton-row" :style="{ width: `${60 + Math.random() * 30}%` }" />
    </div>
    <div v-else class="spinner-container">
      <el-icon class="is-loading spinner-icon">
        <Loading />
      </el-icon>
      <span v-if="text" class="loading-text">{{ text }}</span>
    </div>
  </div>
</template>

<script setup>
import { Loading } from '@element-plus/icons-vue'

defineProps({
  type: {
    type: String,
    default: 'spinner',
    validator: (v) => ['spinner', 'skeleton'].includes(v)
  },
  text: {
    type: String,
    default: ''
  },
  skeletonRows: {
    type: Number,
    default: 5
  }
})
</script>

<style scoped>
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-12) var(--spacing-4);
  min-height: 200px;
}

.spinner-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-3);
}

.spinner-icon {
  font-size: 32px;
  color: var(--color-primary);
}

.loading-text {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
}

.skeleton-container {
  width: 100%;
  max-width: 600px;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.skeleton-row {
  height: 20px;
  background: linear-gradient(90deg, var(--color-border) 25%, var(--color-bg) 50%, var(--color-border) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: var(--radius-sm);
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
</style>
