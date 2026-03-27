<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script setup>
</script>

<style>
/* ========================================
   PromptVault Design System
   ======================================== */

/* Google Fonts */
@import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@300;400;500;600;700&display=swap');

/* CSS Variables - Design Tokens */
:root {
  /* Colors */
  --color-primary: #2563EB;
  --color-primary-hover: #1D4ED8;
  --color-primary-light: #EFF6FF;
  --color-secondary: #3B82F6;
  --color-cta: #F97316;
  --color-cta-hover: #EA580C;

  --color-bg: #F8FAFC;
  --color-surface: #FFFFFF;
  --color-border: #E2E8F0;
  --color-border-hover: #CBD5E1;

  --color-text-primary: #1E293B;
  --color-text-secondary: #64748B;
  --color-text-muted: #94A3B8;

  --color-success: #10B981;
  --color-success-light: #ECFDF5;
  --color-warning: #F59E0B;
  --color-warning-light: #FFFBEB;
  --color-danger: #EF4444;
  --color-danger-light: #FEF2F2;
  --color-info: #3B82F6;
  --color-info-light: #EFF6FF;

  --color-star: #FBBF24;
  --color-pin: #2563EB;

  /* Typography */
  --font-family: 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  --font-size-xs: 12px;
  --font-size-sm: 13px;
  --font-size-base: 14px;
  --font-size-md: 15px;
  --font-size-lg: 16px;
  --font-size-xl: 18px;
  --font-size-2xl: 20px;
  --font-size-3xl: 24px;

  --font-weight-regular: 400;
  --font-weight-medium: 500;
  --font-weight-semibold: 600;
  --font-weight-bold: 700;

  --line-height-tight: 1.25;
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

  /* Shadows - Minimal for flat design */
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.04);
  --shadow-md: 0 2px 4px rgba(0, 0, 0, 0.06);
  --shadow-hover: 0 4px 12px rgba(0, 0, 0, 0.08);

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

/* Cursor */
a, button, [role="button"], .clickable {
  cursor: pointer;
}

/* Focus States - Accessibility */
:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
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

/* Element Plus Overrides */
.el-button {
  font-family: var(--font-family);
  font-weight: var(--font-weight-medium);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.el-button--primary {
  background: var(--color-primary);
  border-color: var(--color-primary);
}

.el-button--primary:hover {
  background: var(--color-primary-hover);
  border-color: var(--color-primary-hover);
}

.el-input__wrapper {
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.el-textarea__inner {
  font-family: var(--font-family);
  border-radius: var(--radius-md);
}

.el-card {
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
  box-shadow: var(--shadow-sm);
  transition: all var(--transition-normal);
}

.el-card:hover {
  box-shadow: var(--shadow-hover);
}

.el-tag {
  border-radius: var(--radius-sm);
  font-weight: var(--font-weight-medium);
}

.el-menu-item {
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-1);
}

.el-dialog {
  border-radius: var(--radius-xl);
}

.el-message-box {
  border-radius: var(--radius-lg);
}

/* Selection */
::selection {
  background: var(--color-primary-light);
  color: var(--color-primary);
}
</style>
