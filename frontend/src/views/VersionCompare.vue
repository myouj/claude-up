<template>
  <div class="version-compare">
    <BreadcrumbNav :items="[{ name: '提示词', path: '/prompts' }, { name: '版本对比' }]" />
    <el-header>
      <div class="header-content">
        <div class="left">
          <el-button class="back-btn" @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <h2>版本对比</h2>
        </div>
        <div class="right">
          <el-select v-model="version1Id" placeholder="选择版本1" class="version-select">
            <el-option
              v-for="v in versions"
              :key="v.id"
              :label="`v${v.version}`"
              :value="v.id"
            />
          </el-select>
          <span class="vs-badge">VS</span>
          <el-select v-model="version2Id" placeholder="选择版本2" class="version-select">
            <el-option
              v-for="v in versions"
              :key="v.id"
              :label="`v${v.version}`"
              :value="v.id"
            />
          </el-select>
        </div>
      </div>
    </el-header>

    <el-main>
      <div class="diff-container">
        <div class="diff-header">
          <div class="diff-info left">
            <span class="version-badge old">v{{ version1?.version || '-' }}</span>
            <span class="version-time">{{ version1?.created_at || '' }}</span>
          </div>
          <div class="diff-info right">
            <span class="version-badge new">v{{ version2?.version || '-' }}</span>
            <span class="version-time">{{ version2?.created_at || '' }}</span>
          </div>
        </div>

        <div class="diff-content">
          <div class="diff-panel left">
            <pre v-html="diffLeft"></pre>
          </div>
          <div class="diff-panel right">
            <pre v-html="diffRight"></pre>
          </div>
        </div>

        <div class="diff-footer">
          <div class="diff-stats">
            <span class="stat-item added">
              <el-icon><Plus /></el-icon>
              新增 {{ addedLines }} 行
            </span>
            <span class="stat-item removed">
              <el-icon><Minus /></el-icon>
              删除 {{ removedLines }} 行
            </span>
          </div>
        </div>
      </div>
    </el-main>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from 'axios'
import * as Diff from 'diff'
import BreadcrumbNav from '../components/BreadcrumbNav.vue'

const router = useRouter()
const route = useRoute()

const versions = ref([])
const version1Id = ref(null)
const version2Id = ref(null)

const version1 = computed(() => versions.value.find(v => v.id === version1Id.value))
const version2 = computed(() => versions.value.find(v => v.id === version2Id.value))

const diffLeft = ref('')
const diffRight = ref('')
const addedLines = ref(0)
const removedLines = ref(0)
const modifiedLines = ref(0)

const fetchVersions = async () => {
  try {
    const res = await axios.get(`/api/prompts/${route.params.id}/versions`)
    if (res.data.success) {
      versions.value = res.data.data

      // 如果 URL 中有版本参数，设置选中
      if (route.query.v1 && route.query.v2) {
        version1Id.value = parseInt(route.query.v1)
        version2Id.value = parseInt(route.query.v2)
      } else if (versions.value.length >= 2) {
        version1Id.value = versions.value[1]?.id
        version2Id.value = versions.value[0]?.id
      }
    }
  } catch (err) {
    console.error('Failed to fetch versions:', err)
  }
}

const computeDiff = () => {
  if (!version1.value || !version2.value) return

  const oldText = version1.value.content || ''
  const newText = version2.value.content || ''

  const diff = Diff.diffLines(oldText, newText)

  let leftHtml = ''
  let rightHtml = ''
  let added = 0
  let removed = 0

  diff.forEach(part => {
    if (part.added) {
      rightHtml += `<div class="diff-line added">${escapeHtml(part.value)}</div>`
      added += part.count || 1
    } else if (part.removed) {
      leftHtml += `<div class="diff-line removed">${escapeHtml(part.value)}</div>`
      removed += part.count || 1
    } else {
      const lines = part.value.split('\n').filter(l => l || part.value === '\n')
      lines.forEach(line => {
        leftHtml += `<div class="diff-line">${escapeHtml(line)}</div>`
        rightHtml += `<div class="diff-line">${escapeHtml(line)}</div>`
      })
    }
  })

  diffLeft.value = leftHtml
  diffRight.value = rightHtml
  addedLines.value = added
  removedLines.value = removed
  modifiedLines.value = Math.min(added, removed)
}

const escapeHtml = (str) => {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

watch([version1Id, version2Id], computeDiff)

const goBack = () => router.back()

onMounted(fetchVersions)
</script>

<style scoped>
.version-compare {
  height: 100vh;
  background: var(--color-bg);
}

.el-header {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  padding: 0 var(--spacing-6);
  height: 64px;
}

.header-content {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.left {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.left h2 {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.back-btn {
  padding: var(--spacing-2);
}

.right {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.version-select {
  width: 120px;
}

.vs-badge {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-muted);
  background: var(--color-bg);
  padding: var(--spacing-1) var(--spacing-3);
  border-radius: var(--radius-full);
}

.el-main {
  padding: 0;
  background: var(--color-bg);
  height: calc(100vh - 64px);
}

.diff-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--color-surface);
}

.diff-header {
  display: flex;
  background: var(--color-bg);
  border-bottom: 1px solid var(--color-border);
}

.diff-info {
  flex: 1;
  padding: var(--spacing-3) var(--spacing-4);
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.diff-info.left {
  border-right: 1px solid var(--color-border);
}

.version-badge {
  padding: var(--spacing-1) var(--spacing-3);
  border-radius: var(--radius-sm);
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-sm);
}

.version-badge.old {
  background: var(--color-danger-light);
  color: var(--color-danger);
}

.version-badge.new {
  background: var(--color-success-light);
  color: var(--color-success);
}

.version-time {
  color: var(--color-text-muted);
  font-size: var(--font-size-xs);
}

.diff-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.diff-panel {
  flex: 1;
  overflow-y: auto;
  background: var(--color-surface);
}

.diff-panel.left {
  border-right: 1px solid var(--color-border);
}

.diff-panel pre {
  margin: 0;
  padding: var(--spacing-4);
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: var(--font-size-sm);
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-all;
}

.diff-footer {
  padding: var(--spacing-3) var(--spacing-4);
  background: var(--color-bg);
  border-top: 1px solid var(--color-border);
}

.diff-stats {
  display: flex;
  gap: var(--spacing-4);
}

.stat-item {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-1);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
}

.stat-item.added {
  color: var(--color-success);
}

.stat-item.removed {
  color: var(--color-danger);
}

:deep(.diff-line) {
  padding: 2px 0;
  border-left: 3px solid transparent;
}

:deep(.diff-line.added) {
  background: var(--color-success-light);
  border-left-color: var(--color-success);
  color: var(--color-success);
}

:deep(.diff-line.removed) {
  background: var(--color-danger-light);
  border-left-color: var(--color-danger);
  color: var(--color-danger);
}
</style>
