<template>
  <div class="test-analytics">
    <BreadcrumbNav :items="[{ name: '提示词', path: '/prompts' }, { name: '测试分析' }]" />
    <el-header>
      <div class="header-content">
        <div class="left">
          <el-button class="back-btn" @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <h2 class="page-title">测试分析</h2>
          <span class="prompt-title">{{ promptTitle }}</span>
        </div>
        <div class="right">
          <el-select v-model="selectedModel" placeholder="全部模型" class="model-select" clearable>
            <el-option label="全部模型" :value="null" />
            <el-option label="MiniMax" value="MiniMax-M2.7" />
            <el-option label="阿里百炼 (Qwen)" value="qwen3.5-plus" />
          </el-select>
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            class="date-picker"
            clearable
          />
        </div>
      </div>
    </el-header>

    <el-main>
      <div v-if="loading" class="loading-state">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>加载分析数据...</span>
      </div>

      <div v-else-if="testRecords.length === 0" class="empty-state">
        <div class="empty-icon">
          <svg width="80" height="80" viewBox="0 0 80 80" fill="none">
            <rect x="10" y="20" width="60" height="40" rx="6" stroke="var(--color-border)" stroke-width="2"/>
            <path d="M20 35l15 10 15-15 15 20" stroke="var(--color-border)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <p>暂无测试数据</p>
        <span>运行测试后即可查看分析图表</span>
        <el-button type="primary" @click="goToTest">去测试</el-button>
      </div>

      <div v-else class="analytics-content">
        <!-- Summary Cards -->
        <div class="stats-grid">
          <el-card class="stat-card">
            <div class="stat-value">{{ summary.total }}</div>
            <div class="stat-label">总测试次数</div>
          </el-card>
          <el-card class="stat-card">
            <div class="stat-value">{{ summary.avgTokens }}</div>
            <div class="stat-label">平均 Tokens</div>
          </el-card>
          <el-card class="stat-card">
            <div class="stat-value">{{ summary.avgLatency }}s</div>
            <div class="stat-label">平均延迟</div>
          </el-card>
          <el-card class="stat-card">
            <div class="stat-value" :class="summary.successRate >= 80 ? 'good' : summary.successRate >= 50 ? 'medium' : 'poor'">
              {{ summary.successRate }}%
            </div>
            <div class="stat-label">成功率</div>
          </el-card>
        </div>

        <!-- Charts Row 1 -->
        <div class="charts-row">
          <el-card class="chart-card tokens-chart">
            <template #header>
              <div class="chart-header">
                <span>Token 消耗趋势</span>
              </div>
            </template>
            <div class="chart-container">
              <Line :data="tokensChartData" :options="lineChartOptions" />
            </div>
          </el-card>

          <el-card class="chart-card latency-chart">
            <template #header>
              <div class="chart-header">
                <span>响应延迟趋势</span>
              </div>
            </template>
            <div class="chart-container">
              <Line :data="latencyChartData" :options="lineChartOptions" />
            </div>
          </el-card>
        </div>

        <!-- Charts Row 2 -->
        <div class="charts-row">
          <el-card class="chart-card">
            <template #header>
              <div class="chart-header">
                <span>测试成功率</span>
              </div>
            </template>
            <div class="chart-container">
              <Bar :data="successRateData" :options="barChartOptions" />
            </div>
          </el-card>

          <el-card class="chart-card">
            <template #header>
              <div class="chart-header">
                <span>各模型使用分布</span>
              </div>
            </template>
            <div class="chart-container">
              <Doughnut :data="modelDistributionData" :options="doughnutOptions" />
            </div>
          </el-card>
        </div>

        <!-- Recent Tests Table -->
        <el-card class="table-card">
          <template #header>
            <div class="chart-header">
              <span>最近测试记录</span>
            </div>
          </template>
          <el-table :data="recentRecords" stripe class="analytics-table">
            <el-table-column label="时间" width="160">
              <template #default="{ row }">
                {{ formatDateTime(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column label="模型" width="120">
              <template #default="{ row }">
                <el-tag size="small" type="info">{{ row.model }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="输入" prop="prompt_text">
              <template #default="{ row }">
                <span class="truncate-text">{{ row.prompt_text?.substring(0, 80) }}...</span>
              </template>
            </el-table-column>
            <el-table-column label="Tokens" width="90">
              <template #default="{ row }">
                {{ row.tokens_used || '-' }}
              </template>
            </el-table-column>
            <el-table-column label="延迟" width="80">
              <template #default="{ row }">
                {{ row.latency ? row.latency + 's' : '-' }}
              </template>
            </el-table-column>
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag size="small" :type="row.success ? 'success' : 'danger'">
                  {{ row.success ? '成功' : '失败' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </div>
    </el-main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from 'axios'
import BreadcrumbNav from '../components/BreadcrumbNav.vue'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line, Bar, Doughnut } from 'vue-chartjs'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const router = useRouter()
const route = useRoute()

const testRecords = ref([])
const loading = ref(true)
const promptTitle = ref('')
const selectedModel = ref(null)
const dateRange = ref(null)

const fetchData = async () => {
  loading.value = true
  try {
    const [promptRes, testsRes] = await Promise.all([
      axios.get(`/api/prompts/${route.params.id}`),
      axios.get(`/api/prompts/${route.params.id}/tests?limit=500`)
    ])
    if (promptRes.data.success) {
      promptTitle.value = promptRes.data.data.title
    }
    if (testsRes.data.success) {
      testRecords.value = testsRes.data.data || []
    }
  } catch (err) {
    console.error('Failed to fetch analytics data:', err)
  } finally {
    loading.value = false
  }
}

const filteredRecords = computed(() => {
  let records = testRecords.value
  if (selectedModel.value) {
    records = records.filter(r => r.model === selectedModel.value)
  }
  if (dateRange.value && dateRange.value.length === 2) {
    const [start, end] = dateRange.value
    records = records.filter(r => {
      const date = new Date(r.created_at).toISOString().split('T')[0]
      return date >= start && date <= end
    })
  }
  return records
})

const summary = computed(() => {
  const records = filteredRecords.value
  if (records.length === 0) {
    return { total: 0, avgTokens: 0, avgLatency: '0.00', successRate: 0 }
  }
  const total = records.length
  const tokensRecords = records.filter(r => r.tokens_used > 0)
  const avgTokens = tokensRecords.length > 0
    ? Math.round(tokensRecords.reduce((sum, r) => sum + r.tokens_used, 0) / tokensRecords.length)
    : 0
  const latencyRecords = records.filter(r => r.latency > 0)
  const avgLatency = latencyRecords.length > 0
    ? (latencyRecords.reduce((sum, r) => sum + parseFloat(r.latency), 0) / latencyRecords.length).toFixed(2)
    : '0.00'
  const successCount = records.filter(r => r.success).length
  const successRate = Math.round((successCount / total) * 100)
  return { total, avgTokens, avgLatency, successRate }
})

// Group records by date for line charts
const groupedByDate = computed(() => {
  const groups = {}
  for (const r of filteredRecords.value) {
    const date = new Date(r.created_at).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
    if (!groups[date]) groups[date] = []
    groups[date].push(r)
  }
  return groups
})

const chartLabels = computed(() => {
  return Object.keys(groupedByDate.value).sort((a, b) => {
    const dateA = new Date(filteredRecords.value.find(r => new Date(r.created_at).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' }) === a)?.created_at || 0)
    const dateB = new Date(filteredRecords.value.find(r => new Date(r.created_at).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' }) === b)?.created_at || 0)
    return dateA - dateB
  })
})

const tokensChartData = computed(() => {
  const labels = chartLabels.value
  const data = labels.map(date => {
    const recs = groupedByDate.value[date].filter(r => r.tokens_used > 0)
    return recs.length > 0 ? Math.round(recs.reduce((s, r) => s + r.tokens_used, 0) / recs.length) : 0
  })
  return {
    labels,
    datasets: [{
      label: '平均 Tokens',
      data,
      borderColor: '#2563EB',
      backgroundColor: 'rgba(37, 99, 235, 0.1)',
      fill: true,
      tension: 0.4,
      pointRadius: 4,
      pointHoverRadius: 6
    }]
  }
})

const latencyChartData = computed(() => {
  const labels = chartLabels.value
  const data = labels.map(date => {
    const recs = groupedByDate.value[date].filter(r => r.latency > 0)
    return recs.length > 0 ? parseFloat((recs.reduce((s, r) => s + parseFloat(r.latency), 0) / recs.length).toFixed(2)) : 0
  })
  return {
    labels,
    datasets: [{
      label: '平均延迟 (s)',
      data,
      borderColor: '#F97316',
      backgroundColor: 'rgba(249, 115, 22, 0.1)',
      fill: true,
      tension: 0.4,
      pointRadius: 4,
      pointHoverRadius: 6
    }]
  }
})

const successRateData = computed(() => {
  const labels = chartLabels.value
  const data = labels.map(date => {
    const recs = groupedByDate.value[date]
    const success = recs.filter(r => r.success).length
    return Math.round((success / recs.length) * 100)
  })
  return {
    labels,
    datasets: [{
      label: '成功率 (%)',
      data,
      backgroundColor: data.map(v => v >= 80 ? '#10B981' : v >= 50 ? '#F59E0B' : '#EF4444'),
      borderRadius: 4
    }]
  }
})

const modelDistributionData = computed(() => {
  const counts = {}
  for (const r of filteredRecords.value) {
    counts[r.model] = (counts[r.model] || 0) + 1
  }
  const models = Object.keys(counts)
  const colors = ['#2563EB', '#F97316', '#10B981', '#8B5CF6', '#EF4444']
  return {
    labels: models,
    datasets: [{
      data: models.map(m => counts[m]),
      backgroundColor: models.map((_, i) => colors[i % colors.length]),
      borderWidth: 0
    }]
  }
})

const recentRecords = computed(() => filteredRecords.value.slice(0, 20))

const lineChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: { mode: 'index', intersect: false }
  },
  scales: {
    x: {
      grid: { display: false },
      ticks: { font: { size: 11 }, color: '#94A3B8' }
    },
    y: {
      grid: { color: '#E2E8F0' },
      ticks: { font: { size: 11 }, color: '#64748B' }
    }
  }
}

const barChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: { callbacks: { label: ctx => `${ctx.raw}%` } }
  },
  scales: {
    x: {
      grid: { display: false },
      ticks: { font: { size: 11 }, color: '#94A3B8' }
    },
    y: {
      min: 0,
      max: 100,
      grid: { color: '#E2E8F0' },
      ticks: {
        font: { size: 11 },
        color: '#64748B',
        callback: v => v + '%'
      }
    }
  }
}

const doughnutOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'bottom',
      labels: { font: { size: 12 }, color: '#64748B', padding: 16 }
    }
  }
}

const formatDateTime = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

const goBack = () => router.back()
const goToTest = () => router.push(`/prompts/${route.params.id}/test`)

onMounted(fetchData)
</script>

<style scoped>
.test-analytics {
  height: 100vh;
  background: var(--color-bg);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.el-header {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  padding: 0 var(--spacing-6);
  height: 64px;
  flex-shrink: 0;
}

.header-content {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-4);
}

.left {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  min-width: 0;
}

.back-btn {
  padding: var(--spacing-2);
  flex-shrink: 0;
}

.page-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  flex-shrink: 0;
}

.prompt-title {
  color: var(--color-text-muted);
  font-size: var(--font-size-sm);
  padding-left: var(--spacing-3);
  border-left: 1px solid var(--color-border);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.right {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  flex-shrink: 0;
}

.model-select {
  width: 130px;
}

.date-picker {
  width: 240px;
}

.el-main {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-5);
}

.loading-state {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-3);
  color: var(--color-text-muted);
}

.empty-state {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-3);
  color: var(--color-text-muted);
  text-align: center;
}

.empty-icon {
  opacity: 0.5;
}

.empty-state p {
  font-size: var(--font-size-md);
  color: var(--color-text-secondary);
  margin: 0;
}

.empty-state span {
  font-size: var(--font-size-sm);
}

.analytics-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-5);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-4);
}

.stat-card {
  text-align: center;
  padding: var(--spacing-3);
}

.stat-value {
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  line-height: 1;
}

.stat-value.good { color: var(--color-success); }
.stat-value.medium { color: var(--color-warning); }
.stat-value.poor { color: var(--color-danger); }

.stat-label {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
  margin-top: var(--spacing-2);
}

.charts-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-4);
}

.chart-card :deep(.el-card__header) {
  padding: var(--spacing-3) var(--spacing-4);
  background: var(--color-bg);
  border-bottom: 1px solid var(--color-border);
}

.chart-header {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.chart-container {
  height: 220px;
  padding: var(--spacing-3);
}

.table-card :deep(.el-card__header) {
  padding: var(--spacing-3) var(--spacing-4);
  background: var(--color-bg);
  border-bottom: 1px solid var(--color-border);
}

.analytics-table {
  font-size: var(--font-size-sm);
}

.truncate-text {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 300px;
}

/* Responsive - Tablet */
@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .charts-row {
    grid-template-columns: 1fr;
  }
}

/* Responsive - Mobile */
@media (max-width: 768px) {
  .header-content {
    flex-wrap: wrap;
  }

  .left {
    gap: var(--spacing-2);
  }

  .page-title {
    font-size: var(--font-size-md);
  }

  .prompt-title {
    display: none;
  }

  .right {
    width: 100%;
    overflow-x: auto;
    flex-wrap: nowrap;
  }

  .model-select {
    min-width: 100px;
  }

  .date-picker {
    min-width: 200px;
  }

  .el-main {
    padding: var(--spacing-3);
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .stat-value {
    font-size: var(--font-size-2xl);
  }

  .charts-row {
    gap: var(--spacing-3);
  }

  .chart-container {
    height: 180px;
  }

  .truncate-text {
    max-width: 150px;
  }
}
</style>
