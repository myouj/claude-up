<template>
  <div class="empty-state">
    <div class="empty-icon">
      <slot name="icon">
        <svg width="80" height="80" viewBox="0 0 80 80" fill="none">
          <rect x="15" y="20" width="50" height="40" rx="6" stroke="var(--color-border)" stroke-width="2"/>
          <path d="M25 35h30M25 45h20M25 55h25" stroke="var(--color-border)" stroke-width="2" stroke-linecap="round"/>
        </svg>
      </slot>
    </div>
    <p class="empty-title">{{ title }}</p>
    <span v-if="description" class="empty-description">{{ description }}</span>
    <slot name="action" />
  </div>
</template>

<script setup>
defineProps({
  title: {
    type: String,
    default: '暂无数据'
  },
  description: {
    type: String,
    default: ''
  }
})
</script>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-12) var(--spacing-4);
  text-align: center;
}

.empty-icon {
  opacity: 0.5;
  margin-bottom: var(--spacing-4);
}

.empty-title {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-secondary);
  margin: 0 0 var(--spacing-2) 0;
}

.empty-description {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
  margin-bottom: var(--spacing-4);
  max-width: 300px;
}
</style>
