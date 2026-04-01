<template>
  <div class="top-nav-wrapper">
    <div class="top-nav-container">
      <!-- Logo / Brand -->
      <div class="nav-brand" @click="navigateTo('/')">
        <svg width="28" height="28" viewBox="0 0 28 28" fill="none" xmlns="http://www.w3.org/2000/svg">
          <rect width="28" height="28" rx="6" fill="var(--color-primary)"/>
          <path d="M8 10h12M8 14h8M8 18h10" stroke="white" stroke-width="2" stroke-linecap="round"/>
        </svg>
        <span class="brand-name">PromptVault</span>
      </div>

      <!-- Main Navigation Menu -->
      <el-menu
        :default-active="activeIndex"
        mode="horizontal"
        :ellipsis="false"
        class="nav-menu"
        @select="handleSelect"
      >
        <el-menu-item index="/">
          <el-icon><House /></el-icon>
          <span>Dashboard</span>
        </el-menu-item>
        <el-menu-item index="/prompts">
          <el-icon><Document /></el-icon>
          <span>Prompts</span>
        </el-menu-item>
        <el-menu-item index="/skills">
          <el-icon><MagicStick /></el-icon>
          <span>Skills</span>
        </el-menu-item>
        <el-menu-item index="/agents">
          <el-icon><User /></el-icon>
          <span>Agents</span>
        </el-menu-item>
        <el-menu-item index="/teams">
          <el-icon><Avatar /></el-icon>
          <span>Teams</span>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <span>Settings</span>
        </el-menu-item>
      </el-menu>

      <!-- Right Side Actions -->
      <div class="nav-actions">
        <el-button text @click="navigateTo('/api-docs')" title="API Docs">
          <el-icon><Collection /></el-icon>
        </el-button>
        <el-button text @click="navigateTo('/activity')" title="Activity">
          <el-icon><Clock /></el-icon>
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  House,
  Document,
  MagicStick,
  User,
  Avatar,
  Collection,
  Clock,
  Setting
} from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()

const activeIndex = computed(() => {
  const path = route.path
  // Highlight parent route for sub-routes
  if (path.startsWith('/prompts')) return '/prompts'
  if (path.startsWith('/skills')) return '/skills'
  if (path.startsWith('/agents')) return '/agents'
  if (path.startsWith('/teams')) return '/teams'
  if (path.startsWith('/ab-tests')) return '/ab-tests'
  return path
})

const handleSelect = (index) => {
  router.push(index)
}

const navigateTo = (path) => {
  router.push(path)
}
</script>

<style scoped>
.top-nav-wrapper {
  position: sticky;
  top: 0;
  z-index: var(--z-dropdown);
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  box-shadow: var(--shadow-sm);
}

.top-nav-container {
  display: flex;
  align-items: center;
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 var(--spacing-6);
  height: 56px;
  gap: var(--spacing-4);
}

.nav-brand {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  cursor: pointer;
  padding: var(--spacing-1) var(--spacing-2);
  border-radius: var(--radius-md);
  transition: background var(--transition-fast);
  flex-shrink: 0;
}

.nav-brand:hover {
  background: var(--color-bg);
}

.brand-name {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  letter-spacing: -0.02em;
}

.nav-menu {
  flex: 1;
  border-bottom: none;
  background: transparent;
  --el-menu-hover-bg-color: var(--color-primary-light);
  --el-menu-hover-text-color: var(--color-primary);
  --el-menu-active-color: var(--color-primary);
}

.nav-menu :deep(.el-menu-item) {
  border-radius: var(--radius-md);
  margin: 0 var(--spacing-1);
  padding: 0 var(--spacing-3);
  height: 36px;
  line-height: 36px;
  border-bottom: none;
  font-weight: var(--font-weight-medium);
  font-size: var(--font-size-base);
}

.nav-menu :deep(.el-menu-item.is-active) {
  background: var(--color-primary-light);
  color: var(--color-primary);
}

.nav-menu :deep(.el-menu-item .el-icon) {
  margin-right: 6px;
}

.nav-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  flex-shrink: 0;
  margin-left: auto;
}

.nav-actions .el-button {
  color: var(--color-text-secondary);
  font-size: var(--font-size-lg);
  padding: var(--spacing-2);
}

.nav-actions .el-button:hover {
  color: var(--color-primary);
  background: var(--color-primary-light);
}
</style>
