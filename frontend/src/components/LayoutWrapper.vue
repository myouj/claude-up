<template>
  <div class="layout-wrapper" :class="{ 'has-sidebar': hasSidebar }">
    <!-- Header -->
    <el-header class="app-header">
      <div class="header-content">
        <div class="header-left">
          <el-button
            v-if="hasSidebar"
            class="mobile-menu-btn"
            @click="drawerVisible = true"
          >
            <el-icon><Menu /></el-icon>
          </el-button>
          <el-button
            v-if="showBackBtn"
            class="back-btn"
            @click="handleBack"
          >
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <div class="brand" @click="$router.push('/')">
            <div class="logo">
              <svg width="28" height="28" viewBox="0 0 28 28" fill="none">
                <rect width="28" height="28" rx="8" fill="var(--color-primary)"/>
                <path d="M8 10h12M8 14h8M8 18h10" stroke="white" stroke-width="2" stroke-linecap="round"/>
              </svg>
            </div>
            <span class="brand-name">{{ brandName }}</span>
          </div>
          <span v-if="pageTitle" class="page-title-divider">/</span>
          <h1 v-if="pageTitle" class="page-title">{{ pageTitle }}</h1>
          <el-tag v-if="badge" :type="badgeType" size="small" class="header-badge">
            {{ badge }}
          </el-tag>
        </div>

        <div class="header-right">
          <slot name="header-actions" />
        </div>
      </div>
    </el-header>

    <!-- Desktop Sidebar -->
    <el-aside v-if="hasSidebar" width="240px" class="desktop-sidebar">
      <slot name="sidebar" />
    </el-aside>

    <!-- Mobile Drawer Sidebar -->
    <el-drawer
      v-if="hasSidebar"
      v-model="drawerVisible"
      :title="sidebarTitle"
      size="280px"
      direction="ltr"
      class="mobile-sidebar-drawer"
    >
      <slot name="sidebar" />
    </el-drawer>

    <!-- Main Content -->
    <el-main class="app-main">
      <slot />
    </el-main>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Menu, ArrowLeft } from '@element-plus/icons-vue'

const props = defineProps({
  // Brand name shown in header
  brandName: {
    type: String,
    default: 'PromptVault'
  },
  // Page title shown after brand
  pageTitle: {
    type: String,
    default: ''
  },
  // Badge text next to page title
  badge: {
    type: String,
    default: ''
  },
  // Badge Element Plus type
  badgeType: {
    type: String,
    default: 'success'
  },
  // Whether to show sidebar
  hasSidebar: {
    type: Boolean,
    default: false
  },
  // Sidebar drawer title (for mobile)
  sidebarTitle: {
    type: String,
    default: '筛选'
  },
  // Show back button in header
  showBackBtn: {
    type: Boolean,
    default: false
  }
})

const router = useRouter()
const drawerVisible = ref(false)

const handleBack = () => {
  router.back()
}
</script>

<style scoped>
.layout-wrapper {
  height: 100vh;
  background: var(--color-bg);
  display: flex;
  flex-direction: column;
}

/* Header - 统一高度 64px */
.app-header {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  padding: 0 var(--spacing-6);
  height: 64px;
  flex-shrink: 0;
  z-index: 100;
}

.header-content {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-4);
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  min-width: 0;
  flex: 1;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  flex-shrink: 0;
}

/* Brand */
.brand {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  cursor: pointer;
  flex-shrink: 0;
}

.brand-name {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.page-title-divider {
  color: var(--color-text-muted);
  font-size: var(--font-size-lg);
  flex-shrink: 0;
}

.page-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex-shrink: 1;
}

.header-badge {
  flex-shrink: 0;
}

/* Buttons */
.mobile-menu-btn,
.back-btn {
  padding: var(--spacing-2);
  display: none;
}

/* Desktop Sidebar */
.desktop-sidebar {
  background: var(--color-surface);
  border-right: 1px solid var(--color-border);
  overflow-y: auto;
}

/* Main Content */
.app-main {
  padding: var(--spacing-6);
  background: var(--color-bg);
  flex: 1;
  overflow-y: auto;
}

/* Layout with sidebar */
.layout-wrapper.has-sidebar {
  flex-direction: row;
  flex-wrap: wrap;
}

.layout-wrapper.has-sidebar .app-header {
  width: 100%;
  flex-shrink: 0;
}

.layout-wrapper.has-sidebar .desktop-sidebar {
  position: sticky;
  top: 0;
  height: calc(100vh - 64px);
  flex-shrink: 0;
}

.layout-wrapper.has-sidebar .app-main {
  flex: 1;
  min-width: 0;
}

/* Tablet */
@media (max-width: 1024px) {
  .app-main {
    padding: var(--spacing-4);
  }
}

/* Mobile */
@media (max-width: 768px) {
  .app-header {
    padding: 0 var(--spacing-4);
    height: 56px;
  }

  .mobile-menu-btn {
    display: flex;
  }

  .desktop-sidebar {
    display: none;
  }

  .brand-name {
    display: none;
  }

  .page-title-divider {
    display: none;
  }

  .page-title {
    font-size: var(--font-size-md);
  }

  .app-main {
    padding: var(--spacing-3);
  }
}
</style>
