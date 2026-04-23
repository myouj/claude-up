<template>
  <div id="app">
    <TopNav />
    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import TopNav from './components/TopNav.vue'
</script>

<style scoped>
.main-content {
  max-width: 1400px;
  margin: 0 auto;
  padding: var(--spacing-6);
  min-height: calc(100vh - 56px);
}
</style>

<style>
/* ========================================
   PromptVault Design System
   Inspired by Vercel Geist Design
   ======================================== */

/* Google Fonts - Inter + JetBrains Mono */
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=JetBrains+Mono:wght@400;500&display=swap');

/* CSS Variables - Design Tokens */
:root {
  /* Colors - Vercel Black System */
  --color-primary: #171717;
  --color-bg: #ffffff;
  --color-surface: #ffffff;
  --color-border: rgba(0, 0, 0, 0.08);
  --color-border-hover: rgba(0, 0, 0, 0.16);
  --color-text-primary: #171717;
  --color-text-secondary: #4d4d4d;
  --color-text-muted: #808080;

  /* Shadows - Shadow-as-border */
  --shadow-border: 0 0 0 1px rgba(0, 0, 0, 0.08);
  --shadow-border-hover: 0 0 0 1px rgba(0, 0, 0, 0.16);
  --shadow-card: 0 0 0 1px rgba(0, 0, 0, 0.08), 0 2px 2px rgba(0, 0, 0, 0.04);
  --shadow-card-hover: 0 0 0 1px rgba(0, 0, 0, 0.12), 0 4px 4px rgba(0, 0, 0, 0.06);
  --shadow-full-card: 0 0 0 1px rgba(0, 0, 0, 0.08), 0 2px 2px rgba(0, 0, 0, 0.04), 0 8px 8px -8px rgba(0, 0, 0, 0.04), #fafafa 0 0 0 1px;

  /* Typography */
  --font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  --font-mono: 'JetBrains Mono', ui-monospace, SFMono-Regular, monospace;
  --tracking-display: -2.4px;
  --tracking-heading: -1.28px;
  --tracking-tight: -0.96px;
  --tracking-normal: normal;

  /* Font Sizes */
  --font-size-xs: 12px;
  --font-size-sm: 13px;
  --font-size-base: 14px;
  --font-size-md: 15px;
  --font-size-lg: 16px;
  --font-size-xl: 18px;
  --font-size-2xl: 20px;
  --font-size-3xl: 24px;
  --font-size-display: 48px;

  /* Font Weights */
  --font-weight-regular: 400;
  --font-weight-medium: 500;
  --font-weight-semibold: 600;

  /* Line Heights */
  --line-height-tight: 1.0;
  --line-height-snug: 1.25;
  --line-height-normal: 1.5;
  --line-height-relaxed: 1.625;

  /* Spacing */
  --spacing-1: 4px;
  --spacing-2: 8px;
  --spacing-3: 12px;
  --spacing-4: 16px;
  --spacing-5: 20px;
  --spacing-6: 24px;
  --spacing-8: 32px;
  --spacing-10: 40px;
  --spacing-12: 48px;

  /* Border Radius */
  --radius-sm: 4px;
  --radius-md: 6px;
  --radius-lg: 8px;
  --radius-xl: 12px;
  --radius-full: 9999px;

  /* Transitions */
  --transition-fast: 150ms ease;
  --transition-normal: 200ms ease;
  --transition-slow: 300ms ease;

  /* Workflow Accent Colors */
  --color-ship: #ff5b4f;
  --color-preview: #de1d8d;
  --color-develop: #0a72ef;

  /* Semantic Colors */
  --color-success: #10b981;
  --color-warning: #f59e0b;
  --color-danger: #ef4444;
  --color-info: #3b82f6;

  /* Z-index scale */
  --z-dropdown: 10;
  --z-modal: 100;
  --z-tooltip: 1000;
}

/* Reset & Base */
*, *::before, *::after {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html {
  font-size: 16px;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

body {
  font-family: var(--font-family);
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-regular);
  line-height: var(--line-height-normal);
  color: var(--color-text-primary);
  background: var(--color-bg);
  min-height: 100vh;
}

#app {
  min-height: 100vh;
}

/* Typography Classes */
.text-display {
  font-size: var(--font-size-display);
  font-weight: var(--font-weight-semibold);
  letter-spacing: var(--tracking-display);
  line-height: var(--line-height-tight);
}

.text-heading {
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-semibold);
  letter-spacing: var(--tracking-heading);
  line-height: var(--line-height-snug);
}

.text-subheading {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
  letter-spacing: var(--tracking-tight);
}

.text-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  letter-spacing: var(--tracking-tight);
}

.text-body {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-regular);
  line-height: var(--line-height-relaxed);
}

.text-small {
  font-size: var(--font-size-sm);
}

.text-caption {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

/* Cursor */
a, button, [role="button"], .clickable {
  cursor: pointer;
}

/* Focus States - Accessibility */
:focus-visible {
  outline: 2px solid var(--color-develop);
  outline-offset: 2px;
}

/* Selection */
::selection {
  background: rgba(10, 114, 239, 0.15);
  color: var(--color-primary);
}

/* Scrollbar */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: var(--color-border);
  border-radius: var(--radius-full);
}

::-webkit-scrollbar-thumb:hover {
  background: var(--color-border-hover);
}

/* ========================================
   Element Plus Overrides
   ======================================== */

/* Buttons */
.el-button {
  font-family: var(--font-family);
  font-weight: var(--font-weight-medium);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
  border: none;
  box-shadow: var(--shadow-border);
  background: #ffffff;
  color: var(--color-primary);
  padding: 8px 16px;
  height: auto;
  line-height: 1;
}

.el-button:hover {
  background: var(--color-bg);
  box-shadow: var(--shadow-border-hover);
}

.el-button--primary {
  background: var(--color-primary);
  color: #ffffff;
  box-shadow: none;
}

.el-button--primary:hover {
  background: #333333;
  box-shadow: none;
}

.el-button--primary:active {
  background: #000000;
}

.el-button--small {
  padding: 6px 12px;
  font-size: var(--font-size-sm);
}

.el-button--large {
  padding: 12px 24px;
  font-size: var(--font-size-lg);
}

.el-button.is-text {
  border: none;
  box-shadow: none;
  background: transparent;
}

.el-button.is-text:hover {
  background: var(--color-bg);
}

/* Inputs */
.el-input__wrapper {
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
  box-shadow: var(--shadow-border) !important;
  background: #ffffff;
}

.el-input__wrapper:hover {
  box-shadow: var(--shadow-border-hover) !important;
}

.el-input__wrapper.is-focus {
  box-shadow: var(--shadow-border-hover), 0 0 0 2px rgba(10, 114, 239, 0.2) !important;
}

.el-textarea__inner {
  font-family: var(--font-family);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-border) !important;
}

.el-textarea__inner:hover {
  box-shadow: var(--shadow-border-hover) !important;
}

.el-textarea__inner:focus {
  box-shadow: var(--shadow-border-hover), 0 0 0 2px rgba(10, 114, 239, 0.2) !important;
}

/* Cards */
.el-card {
  border-radius: var(--radius-lg);
  border: none;
  box-shadow: var(--shadow-card);
  transition: all var(--transition-normal);
  background: #ffffff;
}

.el-card:hover {
  box-shadow: var(--shadow-card-hover);
}

/* Tags */
.el-tag {
  border-radius: var(--radius-sm);
  font-weight: var(--font-weight-medium);
  font-size: var(--font-size-xs);
  padding: 0 8px;
  height: 24px;
  line-height: 24px;
}

.el-tag--success {
  background: #ecfdf5;
  color: #059669;
  border-color: transparent;
}

.el-tag--warning {
  background: #fffbeb;
  color: #d97706;
  border-color: transparent;
}

.el-tag--danger {
  background: #fef2f2;
  color: #dc2626;
  border-color: transparent;
}

.el-tag--info {
  background: #f3f4f6;
  color: #4b5563;
  border-color: transparent;
}

/* Pill Tags / Badges */
.el-tag--pill {
  border-radius: var(--radius-full);
  padding: 0 12px;
}

/* Menu */
.el-menu {
  border: none;
  background: transparent;
}

.el-menu-item {
  border-radius: var(--radius-md);
  margin-bottom: 2px;
  transition: all var(--transition-fast);
}

.el-menu-item:hover {
  background: rgba(0, 0, 0, 0.04);
}

.el-menu-item.is-active {
  background: rgba(0, 0, 0, 0.06);
  color: var(--color-primary);
}

/* Dialogs */
.el-dialog {
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-full-card);
}

.el-dialog__header {
  border-bottom: 1px solid var(--color-border);
  padding: var(--spacing-5) var(--spacing-6);
}

.el-dialog__title {
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-lg);
}

.el-dialog__body {
  padding: var(--spacing-6);
}

.el-dialog__footer {
  border-top: 1px solid var(--color-border);
  padding: var(--spacing-4) var(--spacing-6);
}

/* Drawer */
.el-drawer {
  box-shadow: var(--shadow-full-card);
}

.el-drawer__header {
  border-bottom: 1px solid var(--color-border);
  padding: var(--spacing-5) var(--spacing-6);
  margin-bottom: 0;
}

/* Form */
.el-form-item__label {
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
}

/* Dropdown */
.el-dropdown-menu {
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card-hover);
  padding: var(--spacing-2);
  border: none;
}

.el-dropdown-menu__item {
  border-radius: var(--radius-md);
  padding: 8px 12px;
  font-size: var(--font-size-sm);
}

.el-dropdown-menu__item:hover {
  background: rgba(0, 0, 0, 0.04);
  color: var(--color-primary);
}

/* Pagination */
.el-pagination {
  font-weight: var(--font-weight-medium);
}

.el-pagination.is-background .el-pager li {
  border-radius: var(--radius-md);
  font-weight: var(--font-weight-medium);
}

.el-pagination.is-background .el-pager li:not(.is-disabled).is-active {
  background: var(--color-primary);
  color: #ffffff;
}

/* Message Box */
.el-message-box {
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-full-card);
  padding: var(--spacing-6);
}

/* Breadcrumb */
.el-breadcrumb {
  font-size: var(--font-size-sm);
}

.el-breadcrumb__item:last-child .el-breadcrumb__inner {
  color: var(--color-text-primary);
  font-weight: var(--font-weight-medium);
}

/* Empty State */
.el-empty__description {
  color: var(--color-text-muted);
  font-size: var(--font-size-sm);
}

/* Loading */
.el-loading-mask {
  background: rgba(255, 255, 255, 0.9);
}

.el-loading-spinner .circular {
  stroke: var(--color-primary);
}

/* Avatar */
.el-avatar {
  box-shadow: var(--shadow-border);
}

/* Badge */
.el-badge__content {
  border-radius: var(--radius-full);
  font-weight: var(--font-weight-medium);
}

/* Tooltip */
.el-tooltip__popper {
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-card-hover);
}

/* Switch */
.el-switch.is-checked .el-switch__core {
  background: var(--color-primary);
  border-color: var(--color-primary);
}

/* Checkbox */
.el-checkbox__input.is-checked .el-checkbox__inner {
  background: var(--color-primary);
  border-color: var(--color-primary);
}

/* Radio */
.el-radio__input.is-checked .el-radio__inner {
  background: var(--color-primary);
  border-color: var(--color-primary);
}

/* Select */
.el-select .el-input.is-focus .el-input__wrapper {
  box-shadow: var(--shadow-border-hover), 0 0 0 2px rgba(10, 114, 239, 0.2) !important;
}

.el-select-dropdown__item.is-selected {
  color: var(--color-primary);
  font-weight: var(--font-weight-medium);
}

/* Tabs */
.el-tabs__item.is-active {
  color: var(--color-primary);
  font-weight: var(--font-weight-medium);
}

.el-tabs__active-bar {
  background: var(--color-primary);
}

/* Progress */
.el-progress-bar__outer {
  border-radius: var(--radius-full);
}

/* Alert */
.el-alert {
  border-radius: var(--radius-lg);
}

/* Table */
.el-table {
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
}

.el-table th.el-table__cell {
  background: var(--color-bg);
  font-weight: var(--font-weight-semibold);
}
</style>
