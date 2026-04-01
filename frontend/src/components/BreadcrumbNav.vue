<template>
  <div class="breadcrumb-nav">
    <el-breadcrumb separator="/">
      <el-breadcrumb-item :to="{ path: '/' }">
        <el-icon><HomeFilled /></el-icon>
        <span>首页</span>
      </el-breadcrumb-item>
      <el-breadcrumb-item v-for="item in items" :key="item.path" :to="item.path ? { path: item.path } : undefined">
        {{ item.name }}
      </el-breadcrumb-item>
    </el-breadcrumb>
  </div>
</template>

<script setup>
import { HomeFilled } from '@element-plus/icons-vue'

defineProps({
  items: {
    type: Array,
    required: true,
    default: () => []
  }
})
</script>

<style scoped>
.breadcrumb-nav {
  display: flex;
  align-items: center;
  padding: var(--spacing-2) var(--spacing-4);
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
}

.el-breadcrumb {
  font-size: var(--font-size-sm);
}

.el-breadcrumb-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
}

.el-breadcrumb-item .el-icon {
  margin-right: 2px;
}

.el-breadcrumb-item a {
  color: var(--color-text-secondary);
  text-decoration: none;
  transition: color var(--transition-fast);
}

.el-breadcrumb-item a:hover {
  color: var(--color-primary);
}

:deep(.el-breadcrumb__separator) {
  color: var(--color-text-muted);
}
</style>
