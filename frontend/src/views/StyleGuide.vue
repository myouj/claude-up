<template>
  <div class="style-guide">
    <div class="guide-header">
      <h1>Design System</h1>
      <p>PromptVault 样式系统 - Vercel Geist 风格</p>
    </div>

    <!-- Colors Section -->
    <section class="guide-section">
      <h2>色彩系统</h2>
      <div class="section-grid">
        <!-- Primary Colors -->
        <div class="color-group">
          <h3>主色 / Primary</h3>
          <div class="color-swatches">
            <div class="swatch">
              <div class="swatch-preview" style="background: #171717"></div>
              <div class="swatch-info">
                <span class="swatch-name">Vercel Black</span>
                <code>#171717</code>
              </div>
            </div>
            <div class="swatch">
              <div class="swatch-preview" style="background: #ffffff; border: 1px solid #e2e8f0"></div>
              <div class="swatch-info">
                <span class="swatch-name">Pure White</span>
                <code>#ffffff</code>
              </div>
            </div>
          </div>
        </div>

        <!-- Neutral Scale -->
        <div class="color-group">
          <h3>中性色 / Neutral</h3>
          <div class="color-swatches horizontal">
            <div class="swatch-compact" style="background: #171717"><span>900</span></div>
            <div class="swatch-compact" style="background: #4d4d4d"><span>600</span></div>
            <div class="swatch-compact" style="background: #666666"><span>500</span></div>
            <div class="swatch-compact" style="background: #808080"><span>400</span></div>
            <div class="swatch-compact" style="background: #ebebeb"><span>100</span></div>
            <div class="swatch-compact" style="background: #fafafa"><span>50</span></div>
          </div>
        </div>

        <!-- Workflow Accents -->
        <div class="color-group">
          <h3>工作流三色 / Workflow</h3>
          <div class="color-swatches">
            <div class="swatch">
              <div class="swatch-preview" style="background: #ff5b4f"></div>
              <div class="swatch-info">
                <span class="swatch-name">Ship Red</span>
                <code>#ff5b4f</code>
              </div>
            </div>
            <div class="swatch">
              <div class="swatch-preview" style="background: #de1d8d"></div>
              <div class="swatch-info">
                <span class="swatch-name">Preview Pink</span>
                <code>#de1d8d</code>
              </div>
            </div>
            <div class="swatch">
              <div class="swatch-preview" style="background: #0a72ef"></div>
              <div class="swatch-info">
                <span class="swatch-name">Develop Blue</span>
                <code>#0a72ef</code>
              </div>
            </div>
          </div>
        </div>

        <!-- Semantic Colors -->
        <div class="color-group">
          <h3>语义色 / Semantic</h3>
          <div class="color-swatches">
            <div class="swatch">
              <div class="swatch-preview" style="background: #10b981"></div>
              <div class="swatch-info">
                <span class="swatch-name">Success</span>
                <code>#10b981</code>
              </div>
            </div>
            <div class="swatch">
              <div class="swatch-preview" style="background: #f59e0b"></div>
              <div class="swatch-info">
                <span class="swatch-name">Warning</span>
                <code>#f59e0b</code>
              </div>
            </div>
            <div class="swatch">
              <div class="swatch-preview" style="background: #ef4444"></div>
              <div class="swatch-info">
                <span class="swatch-name">Danger</span>
                <code>#ef4444</code>
              </div>
            </div>
            <div class="swatch">
              <div class="swatch-preview" style="background: #3b82f6"></div>
              <div class="swatch-info">
                <span class="swatch-name">Info</span>
                <code>#3b82f6</code>
              </div>
            </div>
          </div>
        </div>

        <!-- Shadow Border Demo -->
        <div class="color-group full-width">
          <h3>阴影边框 / Shadow Border</h3>
          <div class="shadow-demo-grid">
            <div class="shadow-box level-1">Level 1 Ring</div>
            <div class="shadow-box level-2">Level 2 Card</div>
            <div class="shadow-box level-3">Level 3 Full</div>
          </div>
        </div>
      </div>
    </section>

    <!-- Typography Section -->
    <section class="guide-section">
      <h2>字体层级</h2>
      <div class="type-scale">
        <div class="type-row">
          <span class="type-label">Display</span>
          <span class="type-sample text-display">PromptVault</span>
          <code>48px / -2.4px tracking / 600</code>
        </div>
        <div class="type-row">
          <span class="type-label">Heading</span>
          <span class="type-sample text-heading">PromptVault</span>
          <code>24px / -1.28px tracking / 600</code>
        </div>
        <div class="type-row">
          <span class="type-label">Title</span>
          <span class="type-sample text-title">PromptVault</span>
          <code>18px / -0.96px tracking / 600</code>
        </div>
        <div class="type-row">
          <span class="type-label">Body</span>
          <span class="type-sample text-body">The quick brown fox jumps over the lazy dog.</span>
          <code>16px / normal / 400</code>
        </div>
        <div class="type-row">
          <span class="type-label">Small</span>
          <span class="type-sample text-small">The quick brown fox jumps over the lazy dog.</span>
          <code>13px / normal / 400</code>
        </div>
        <div class="type-row">
          <span class="type-label">Caption</span>
          <span class="type-sample text-caption">The quick brown fox jumps over the lazy dog.</span>
          <code>12px / normal / 400</code>
        </div>
      </div>
    </section>

    <!-- Buttons Section -->
    <section class="guide-section">
      <h2>按钮 / Buttons</h2>
      <div class="button-showcase">
        <div class="button-group">
          <h3>Primary CTA</h3>
          <div class="button-row">
            <el-button type="primary">Primary Button</el-button>
            <el-button type="primary" disabled>Disabled</el-button>
          </div>
        </div>
        <div class="button-group">
          <h3>Ghost Buttons</h3>
          <div class="button-row">
            <el-button>Default</el-button>
            <el-button disabled>Disabled</el-button>
            <el-button text>Text Button</el-button>
          </div>
        </div>
        <div class="button-group">
          <h3>Button Sizes</h3>
          <div class="button-row">
            <el-button type="primary" size="large">Large</el-button>
            <el-button type="primary">Default</el-button>
            <el-button type="primary" size="small">Small</el-button>
          </div>
        </div>
        <div class="button-group">
          <h3>Icon Buttons</h3>
          <div class="button-row">
            <el-button :icon="Plus">New</el-button>
            <el-button :icon="Download">Export</el-button>
            <el-button :icon="Setting"></el-button>
            <el-button text :icon="Delete"></el-button>
          </div>
        </div>
        <div class="button-group">
          <h3>Pill Badges / Tags</h3>
          <div class="button-row">
            <span class="pill-badge primary">Prompts</span>
            <span class="pill-badge success">Active</span>
            <span class="pill-badge warning">Pending</span>
            <span class="pill-badge danger">Error</span>
            <span class="pill-badge info">Info</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Cards Section -->
    <section class="guide-section">
      <h2>卡片 / Cards</h2>
      <div class="card-showcase">
        <el-card class="demo-card">
          <template #header>
            <div class="card-header-demo">
              <span class="card-title">Card Title</span>
              <el-tag type="success" size="small">New</el-tag>
            </div>
          </template>
          <p class="card-description">
            This is a standard card with shadow-border styling. Cards use multiple shadow layers for depth.
          </p>
          <div class="card-tags">
            <el-tag size="small">Tag 1</el-tag>
            <el-tag size="small">Tag 2</el-tag>
          </div>
          <template #footer>
            <div class="card-footer-demo">
              <span class="card-meta">Updated 2 hours ago</span>
              <el-button size="small">Action</el-button>
            </div>
          </template>
        </el-card>

        <el-card class="demo-card featured">
          <template #header>
            <div class="card-header-demo">
              <span class="card-title">Featured Card</span>
              <span class="featured-badge">Featured</span>
            </div>
          </template>
          <p class="card-description">
            This card uses the Level 3 shadow stack with an inner highlight for a more prominent appearance.
          </p>
          <div class="card-tags">
            <el-tag size="small" type="primary">Promoted</el-tag>
          </div>
        </el-card>
      </div>
    </section>

    <!-- Form Elements Section -->
    <section class="guide-section">
      <h2>表单元素 / Forms</h2>
      <div class="form-showcase">
        <div class="form-group">
          <label>Text Input</label>
          <el-input placeholder="Enter text..." />
        </div>
        <div class="form-group">
          <label>Textarea</label>
          <el-input type="textarea" :rows="3" placeholder="Enter description..." />
        </div>
        <div class="form-group">
          <label>Select</label>
          <el-select placeholder="Select option" style="width: 100%">
            <el-option label="Option 1" value="1" />
            <el-option label="Option 2" value="2" />
            <el-option label="Option 3" value="3" />
          </el-select>
        </div>
        <div class="form-group">
          <label>Checkbox & Radio</label>
          <div class="checkbox-group">
            <el-checkbox v-model="checked1">Enable notifications</el-checkbox>
            <el-checkbox v-model="checked2">Dark mode</el-checkbox>
          </div>
        </div>
      </div>
    </section>

    <!-- Shadows Detail Section -->
    <section class="guide-section">
      <h2>阴影系统详解</h2>
      <div class="shadow-reference">
        <div class="shadow-code">
          <h3>Shadow Border (Level 1)</h3>
          <pre><code>box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.08);</code></pre>
          <p>Used for: buttons, inputs, badges, subtle containers</p>
        </div>
        <div class="shadow-code">
          <h3>Card Shadow (Level 2)</h3>
          <pre><code>box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.08),
              0 2px 2px rgba(0, 0, 0, 0.04);</code></pre>
          <p>Used for: standard cards, dropdowns</p>
        </div>
        <div class="shadow-code">
          <h3>Full Card Shadow (Level 3)</h3>
          <pre><code>box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.08),
              0 2px 2px rgba(0, 0, 0, 0.04),
              0 8px 8px -8px rgba(0, 0, 0, 0.04),
              #fafafa 0 0 0 1px;</code></pre>
          <p>Used for: featured cards, modals, drawers</p>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Plus, Download, Setting, Delete } from '@element-plus/icons-vue'

const checked1 = ref(true)
const checked2 = ref(false)
</script>

<style scoped>
.style-guide {
  max-width: 900px;
  margin: 0 auto;
  padding: var(--spacing-8);
}

.guide-header {
  margin-bottom: var(--spacing-10);
  padding-bottom: var(--spacing-6);
  border-bottom: 1px solid var(--color-border);
}

.guide-header h1 {
  font-size: var(--font-size-display);
  font-weight: var(--font-weight-semibold);
  letter-spacing: var(--tracking-display);
  color: var(--color-primary);
  margin-bottom: var(--spacing-2);
}

.guide-header p {
  color: var(--color-text-secondary);
  font-size: var(--font-size-lg);
}

.guide-section {
  margin-bottom: var(--spacing-12);
}

.guide-section h2 {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-primary);
  margin-bottom: var(--spacing-6);
  padding-bottom: var(--spacing-3);
  border-bottom: 1px solid var(--color-border);
}

.section-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-6);
}

.full-width {
  grid-column: 1 / -1;
}

/* Color Swatches */
.color-group h3 {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-secondary);
  margin-bottom: var(--spacing-3);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.color-swatches {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.color-swatches.horizontal {
  flex-direction: row;
  flex-wrap: wrap;
  gap: var(--spacing-2);
}

.swatch {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.swatch-preview {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  flex-shrink: 0;
}

.swatch-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.swatch-name {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
}

.swatch-info code {
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.swatch-compact {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: flex-end;
  justify-content: center;
  padding-bottom: 4px;
  font-size: 10px;
  color: white;
  font-weight: var(--font-weight-medium);
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
}

/* Shadow Demo */
.shadow-demo-grid {
  display: flex;
  gap: var(--spacing-4);
  flex-wrap: wrap;
}

.shadow-box {
  width: 140px;
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-secondary);
  border-radius: var(--radius-lg);
}

.shadow-box.level-1 {
  box-shadow: var(--shadow-border);
  background: white;
}

.shadow-box.level-2 {
  box-shadow: var(--shadow-card);
  background: white;
}

.shadow-box.level-3 {
  box-shadow: var(--shadow-full-card);
  background: white;
}

/* Typography Scale */
.type-scale {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4);
}

.type-row {
  display: grid;
  grid-template-columns: 80px 1fr 200px;
  align-items: baseline;
  gap: var(--spacing-4);
  padding: var(--spacing-3) 0;
  border-bottom: 1px solid var(--color-border);
}

.type-label {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.type-sample {
  color: var(--color-primary);
}

.type-row code {
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

/* Buttons */
.button-showcase {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-6);
}

.button-group h3 {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-secondary);
  margin-bottom: var(--spacing-3);
}

.button-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  flex-wrap: wrap;
}

.pill-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  border-radius: var(--radius-full);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
}

.pill-badge.primary {
  background: rgba(23, 23, 23, 0.08);
  color: var(--color-primary);
}

.pill-badge.success {
  background: #ecfdf5;
  color: #059669;
}

.pill-badge.warning {
  background: #fffbeb;
  color: #d97706;
}

.pill-badge.danger {
  background: #fef2f2;
  color: #dc2626;
}

.pill-badge.info {
  background: #eff6ff;
  color: #2563eb;
}

/* Cards */
.card-showcase {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-4);
}

.demo-card {
  cursor: default;
}

.card-header-demo {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-primary);
}

.card-description {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  line-height: var(--line-height-relaxed);
  margin-bottom: var(--spacing-3);
}

.card-tags {
  display: flex;
  gap: var(--spacing-2);
}

.card-footer-demo {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-meta {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.featured-badge {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  color: var(--color-preview);
  background: rgba(222, 29, 141, 0.1);
  padding: 2px 8px;
  border-radius: var(--radius-full);
}

.demo-card.featured {
  box-shadow: var(--shadow-full-card);
}

/* Forms */
.form-showcase {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-5);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.form-group label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
}

.form-group .el-select {
  width: 100%;
}

.checkbox-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

/* Shadow Reference */
.shadow-reference {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-5);
}

.shadow-code h3 {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-primary);
  margin-bottom: var(--spacing-2);
}

.shadow-code pre {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--spacing-3);
  margin-bottom: var(--spacing-2);
  overflow-x: auto;
}

.shadow-code code {
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
}

.shadow-code p {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
}

/* Responsive */
@media (max-width: 768px) {
  .style-guide {
    padding: var(--spacing-4);
  }

  .section-grid {
    grid-template-columns: 1fr;
  }

  .type-row {
    grid-template-columns: 60px 1fr;
    grid-template-rows: auto auto;
  }

  .type-row code {
    grid-column: 2;
    grid-row: 2;
  }

  .card-showcase,
  .form-showcase {
    grid-template-columns: 1fr;
  }
}
</style>
