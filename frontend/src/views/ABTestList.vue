<template>
  <div class="ab-test-list">
    <el-header>
      <div class="header-content">
        <div class="left-group">
          <el-button class="mobile-menu-btn" @click="showSidebar = true">
            <el-icon><Menu /></el-icon>
          </el-button>
          <div class="brand">
            <el-button class="back-btn" @click="goBack">
              <el-icon><ArrowLeft /></el-icon>
            </el-button>
            <h1>A/B 测试</h1>
          </div>
        </div>
        <div class="actions-group">
          <el-button type="primary" @click="showCreateDialog = true">
            <el-icon><Plus /></el-icon>
            <span class="btn-text">新建测试</span>
          </el-button>
          <el-button @click="handleExport">
            <el-icon><Download /></el-icon>
            <span class="btn-text">导出</span>
          </el-button>
        </div>
      </div>
    </el-header>

    <el-main>
      <!-- Filter Tabs -->
      <div class="filter-tabs">
        <el-radio-group v-model="filterStatus" class="status-tabs">
          <el-radio-button label="all">全部</el-radio-button>
          <el-radio-button label="running">运行中</el-radio-button>
          <el-radio-button label="completed">已完成</el-radio-button>
        </el-radio-group>
        <el-input
          v-model="searchQuery"
          placeholder="搜索测试..."
          :prefix-icon="Search"
          clearable
          class="search-input"
        />
      </div>

      <!-- Test List -->
      <div v-if="filteredTests.length > 0" class="test-grid">
        <el-card
          v-for="test in filteredTests"
          :key="test.id"
          class="test-card"
          :class="{ winner: test.winner }"
          @click="goToDetail(test.id)"
        >
          <template #header>
            <div class="card-header">
              <div class="title-row">
                <el-tag
                  v-if="test.status === 'completed'"
                  type="success"
                  size="small"
                >已完成</el-tag>
                <el-tag
                  v-else-if="test.status === 'running'"
                  type="warning"
                  size="small"
                  effect="plain"
                >
                  <el-icon class="running-icon"><Loading /></el-icon>
                  运行中
                </el-tag>
                <span class="test-name">{{ test.name }}</span>
              </div>
            </div>
          </template>

          <div class="card-body">
            <p class="prompt-title">
              <el-icon><Document /></el-icon>
              {{ test.prompt_title }}
            </p>
            <div class="meta-row">
              <span class="meta-item">
                <el-icon><Timer /></el-icon>
                {{ formatDate(test.created_at) }}
              </span>
              <span class="meta-item">
                <el-icon><ChatDotRound /></el-icon>
                {{ test.total_runs }} 次运行
              </span>
            </div>

            <!-- Variant Preview -->
            <div v-if="test.variants" class="variants-preview">
              <div
                v-for="variant in test.variants"
                :key="variant.id"
                class="variant-mini"
                :class="{ winner: variant.id === test.winner }"
              >
                <span class="variant-tag">{{ variant.id.toUpperCase() }}</span>
                <el-rate
                  :model-value="variant.metrics.avg_score"
                  disabled
                  size="small"
                  :score-template="`${variant.metrics.avg_score.toFixed(1)}`"
                />
              </div>
            </div>

            <!-- Winner Badge -->
            <div v-if="test.winner" class="winner-badge">
              <el-icon><Trophy /></el-icon>
              Variant {{ test.winner.toUpperCase() }} 胜出
            </div>
          </div>

          <template #footer>
            <div class="card-footer">
              <el-button size="small" @click.stop="goToDetail(test.id)">
                查看详情
              </el-button>
              <el-button
                size="small"
                type="danger"
                text
                @click.stop="handleDelete(test)"
                :disabled="test.status === 'running'"
              >
                删除
              </el-button>
            </div>
          </template>
        </el-card>
      </div>

      <el-empty v-else description="暂无测试记录">
        <el-button type="primary" @click="showCreateDialog = true">
          创建第一个 A/B 测试
        </el-button>
      </el-empty>
    </el-main>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreateDialog" title="新建 A/B 测试" width="560px">
      <el-form :model="newTest" label-position="top">
        <el-form-item label="测试名称" required>
          <el-input v-model="newTest.name" placeholder="例如：Code Review Prompt 优化测试" />
        </el-form-item>
        <el-form-item label="选择 Prompt" required>
          <el-select v-model="newTest.prompt_id" placeholder="选择一个 Prompt" class="full-width">
            <el-option label="代码审查专家" :value="1" />
            <el-option label="SQL 生成助手" :value="2" />
            <el-option label="翻译润色助手" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="Variant A 描述">
          <el-input v-model="newTest.variantA" type="textarea" :rows="2" placeholder="Variant A 的描述..." />
        </el-form-item>
        <el-form-item label="Variant B 描述">
          <el-input v-model="newTest.variantB" type="textarea" :rows="2" placeholder="Variant B 的描述..." />
        </el-form-item>
        <el-form-item label="运行次数">
          <el-input-number v-model="newTest.runs" :min="1" :max="100" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Menu } from '@element-plus/icons-vue'
import { mockABTests } from '../composables/useABTest'
import BreadcrumbNav from '../components/BreadcrumbNav.vue'

const router = useRouter()

const tests = ref([])
const filterStatus = ref('all')
const searchQuery = ref('')
const showCreateDialog = ref(false)
const showSidebar = ref(false)

const newTest = ref({
  name: '',
  prompt_id: null,
  variantA: '',
  variantB: '',
  runs: 20
})

onMounted(() => {
  tests.value = mockABTests.value
})

const filteredTests = computed(() => {
  let result = tests.value

  if (filterStatus.value !== 'all') {
    result = result.filter(t => t.status === filterStatus.value)
  }

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(t =>
      t.name.toLowerCase().includes(q) ||
      t.prompt_title.toLowerCase().includes(q)
    )
  }

  return result
})

const goBack = () => router.push('/')
const goToDetail = (id) => router.push(`/ab-tests/${id}`)

const handleCreate = () => {
  if (!newTest.value.name || !newTest.value.prompt_id) {
    ElMessage.warning('请填写名称和选择 Prompt')
    return
  }
  ElMessage.success('A/B 测试已创建（mock 模式）')
  showCreateDialog.value = false
}

const handleDelete = async (test) => {
  try {
    await ElMessageBox.confirm(`确定删除测试 "${test.name}" 吗？`, '删除确认', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    tests.value = tests.value.filter(t => t.id !== test.id)
    ElMessage.success('删除成功')
  } catch {
    // cancelled
  }
}

const handleExport = () => {
  const content = JSON.stringify(tests.value, null, 2)
  const blob = new Blob([content], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'ab-tests.json'
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('导出成功')
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return `${date.getMonth() + 1}/${date.getDate()} ${date.getHours()}:${String(date.getMinutes()).padStart(2, '0')}`
}
</script>

<style scoped>
.ab-test-list {
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

.left-group,
.actions-group {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.brand {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.back-btn {
  padding: var(--spacing-2);
}

.brand h1 {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0;
}

.mobile-menu-btn {
  display: none;
  padding: var(--spacing-2);
}

.el-main {
  padding: var(--spacing-6);
}

.filter-tabs {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-4);
  margin-bottom: var(--spacing-5);
}

.status-tabs {
  flex-shrink: 0;
}

.search-input {
  max-width: 280px;
}

.test-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: var(--spacing-5);
}

.test-card {
  cursor: pointer;
  transition: all var(--transition-normal);
}

.test-card:hover {
  transform: translateY(-2px);
  border-color: var(--color-border-hover);
}

.test-card.winner {
  border-left: 3px solid var(--color-success);
}

.card-header {
  display: flex;
  justify-content: space-between;
}

.title-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  overflow: hidden;
}

.test-name {
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

.prompt-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  margin: 0;
}

.meta-row {
  display: flex;
  gap: var(--spacing-4);
}

.meta-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.variants-preview {
  display: flex;
  gap: var(--spacing-3);
}

.variant-mini {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-3);
  background: var(--color-bg);
  border-radius: var(--radius-md);
  flex: 1;
}

.variant-mini.winner {
  background: color-mix(in srgb, var(--color-success-light) 30%, var(--color-bg));
  border: 1px solid var(--color-success-light);
}

.variant-tag {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-bold);
  color: var(--color-primary);
  background: var(--color-primary-light);
  padding: 2px 6px;
  border-radius: 4px;
}

.variant-mini.winner .variant-tag {
  background: var(--color-success-light);
  color: var(--color-success);
}

.winner-badge {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-success);
  padding: var(--spacing-2);
  background: color-mix(in srgb, var(--color-success-light) 20%, var(--color-bg));
  border-radius: var(--radius-md);
}

.card-footer {
  display: flex;
  justify-content: space-between;
}

.running-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@media (max-width: 768px) {
  .test-grid {
    grid-template-columns: 1fr;
  }

  .filter-tabs {
    flex-direction: column;
    align-items: stretch;
  }

  .search-input {
    max-width: 100%;
  }

  .mobile-menu-btn {
    display: flex;
  }

  .btn-text {
    display: none;
  }

  .el-main {
    padding: var(--spacing-3);
  }
}
</style>
