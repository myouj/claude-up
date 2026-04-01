<template>
  <el-card
    class="template-card"
    :class="{ installed: isInstalled }"
    @click="$emit('click', template)"
  >
    <!-- Preview Image -->
    <div class="card-preview">
      <div class="preview-bg" :style="{ background: previewColor }">
        <div class="preview-icon">
          <el-icon><component :is="previewIcon" /></el-icon>
        </div>
      </div>
      <el-tag
        v-if="isInstalled"
        class="installed-badge"
        type="success"
        size="small"
        effect="dark"
      >
        <el-icon><CircleCheck /></el-icon>
        已安装
      </el-tag>
    </div>

    <template #header>
      <div class="card-header">
        <span class="template-name">{{ template.name }}</span>
        <el-tag size="small" type="info">{{ categoryLabel }}</el-tag>
      </div>
    </template>

    <div class="card-body">
      <p class="template-desc">{{ template.description }}</p>

      <div class="template-tags">
        <el-tag
          v-for="tag in template.tags.slice(0, 3)"
          :key="tag"
          size="small"
          effect="plain"
          class="tag"
        >
          {{ tag }}
        </el-tag>
      </div>

      <div class="template-meta">
        <div class="meta-item score">
          <el-icon><Star /></el-icon>
          <span>{{ template.score.toFixed(1) }}</span>
        </div>
        <div class="meta-item installs">
          <el-icon><Download /></el-icon>
          <span>{{ formatInstalls(template.installs) }}</span>
        </div>
        <div class="meta-item author">
          <el-avatar :size="16">{{ template.author.avatar }}</el-avatar>
          <span>{{ template.author.name }}</span>
        </div>
      </div>
    </div>
  </el-card>
</template>

<script setup>
import { computed } from 'vue'
import { templateCategories } from '../composables/useTemplate'

const props = defineProps({
  template: {
    type: Object,
    required: true
  },
  isInstalled: {
    type: Boolean,
    default: false
  }
})

defineEmits(['click'])

const categoryLabel = computed(() => {
  const cat = templateCategories.value.find(c => c.value === props.template.category)
  return cat?.label || props.template.category
})

const previewColor = computed(() => {
  const colors = {
    development: '#EFF6FF',
    data: '#F0FDF4',
    docs: '#FEF3C7',
    debug: '#FEF2F2',
    product: '#F3E8FF',
    git: '#FCE7F3',
    testing: '#ECFEFF',
    translation: '#F0FDFA'
  }
  return colors[props.template.category] || '#F3F4F6'
})

const previewIcon = computed(() => {
  const icons = {
    development: 'Code',
    data: 'DataAnalysis',
    docs: 'Document',
    debug: 'Bug',
    product: 'Goods',
    git: 'Branch',
    testing: 'CircleCheck',
    translation: 'Translate'
  }
  return icons[props.template.category] || 'Document'
})

const formatInstalls = (num) => {
  if (num >= 1000) return (num / 1000).toFixed(1) + 'k'
  return num.toString()
}
</script>

<style scoped>
.template-card {
  cursor: pointer;
  transition: all var(--transition-normal);
}

.template-card:hover {
  transform: translateY(-2px);
  border-color: var(--color-border-hover);
}

.template-card.installed {
  border-left: 3px solid var(--color-success);
}

.card-preview {
  position: relative;
  height: 100px;
  border-radius: var(--radius-md);
  overflow: hidden;
  margin-bottom: var(--spacing-3);
}

.preview-bg {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-icon {
  font-size: 40px;
  opacity: 0.6;
  color: var(--color-text-muted);
}

.installed-badge {
  position: absolute;
  top: var(--spacing-2);
  right: var(--spacing-2);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-2);
}

.template-name {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.template-desc {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.5;
}

.template-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-1);
}

.tag {
  font-size: var(--font-size-xs);
}

.template-meta {
  display: flex;
  align-items: center;
  gap: var(--spacing-4);
  padding-top: var(--spacing-2);
  border-top: 1px solid var(--color-border);
}

.meta-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.meta-item.score {
  color: var(--color-warning);
  font-weight: var(--font-weight-semibold);
}

.meta-item.author span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 60px;
}
</style>
