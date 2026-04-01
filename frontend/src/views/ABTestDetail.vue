<template>
  <div class="ab-test-detail">
    <BreadcrumbNav :items="[{ name: 'A/B 测试', path: '/ab-tests' }, { name: test?.name || '测试详情' }]" />
    <el-header>
      <div class="header-content">
        <div class="left">
          <el-button class="back-btn" @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <div class="title-area">
            <h2>{{ test?.name || 'A/B 测试详情' }}</h2>
            <el-tag v-if="test?.status === 'completed'" type="success" size="small">已完成</el-tag>
            <el-tag v-else-if="test?.status === 'running'" type="warning" size="small">运行中</el-tag>
          </div>
        </div>
        <div class="right">
          <el-button @click="goToPrompt">
            <el-icon><Document /></el-icon>
            查看 Prompt
          </el-button>
          <el-button type="primary" @click="runNewTest">
            <el-icon><RefreshRight /></el-icon>
            重新测试
          </el-button>
        </div>
      </div>
    </el-header>

    <el-main v-if="test">
      <!-- Test Info -->
      <div class="test-info">
        <div class="info-item">
          <span class="info-label">Prompt</span>
          <span class="info-value">{{ test.prompt_title }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">创建时间</span>
          <span class="info-value">{{ formatDate(test.created_at) }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">总运行次数</span>
          <span class="info-value">{{ test.total_runs }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">胜出</span>
          <span class="info-value" :class="{ winner: test.winner }">
            {{ test.winner ? `Variant ${test.winner.toUpperCase()}` : '待定' }}
          </span>
        </div>
      </div>

      <!-- Variant Compare -->
      <ABTestCompare
        v-if="test.variants"
        :variants="test.variants"
        :winner="test.winner"
      />
    </el-main>

    <el-main v-else>
      <el-empty description="测试不存在" />
    </el-main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { mockABTests } from '../composables/useABTest'
import BreadcrumbNav from '../components/BreadcrumbNav.vue'
import ABTestCompare from '../components/ABTestCompare.vue'

const router = useRouter()
const route = useRoute()

const test = ref(null)

const fetchTest = () => {
  const id = parseInt(route.params.id)
  test.value = mockABTests.value.find(t => t.id === id) || null
}

const goBack = () => router.push('/ab-tests')
const goToPrompt = () => {
  if (test.value) router.push(`/prompts/${test.value.prompt_id}`)
}

const runNewTest = () => {
  ElMessage.info('重新测试功能待后端 API 支持后实现')
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return `${date.getFullYear()}/${date.getMonth() + 1}/${date.getDate()} ${date.getHours()}:${String(date.getMinutes()).padStart(2, '0')}`
}

onMounted(fetchTest)
</script>

<style scoped>
.ab-test-detail {
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

.back-btn {
  padding: var(--spacing-2);
}

.title-area {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.title-area h2 {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0;
}

.right {
  display: flex;
  gap: var(--spacing-2);
  flex-shrink: 0;
}

.el-main {
  padding: var(--spacing-6);
}

.test-info {
  display: flex;
  gap: var(--spacing-6);
  padding: var(--spacing-4);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  margin-bottom: var(--spacing-6);
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.info-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.info-value {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.info-value.winner {
  color: var(--color-success);
}

@media (max-width: 768px) {
  .test-info {
    flex-wrap: wrap;
    gap: var(--spacing-3);
  }

  .right .btn-text {
    display: none;
  }

  .el-main {
    padding: var(--spacing-3);
  }
}
</style>
