<template>
  <div class="dashboard">
    <main class="dashboard-main">
      <div class="stats-grid">
        <div class="stat-card prompts clickable" @click="goToPrompts">
          <div class="stat-icon">
            <el-icon><Document /></el-icon>
          </div>
          <div class="stat-content">
            <span class="stat-value">{{ stats.prompts }}</span>
            <span class="stat-label">提示词</span>
          </div>
        </div>

        <div class="stat-card skills clickable" @click="goToSkills">
          <div class="stat-icon">
            <el-icon><Timer /></el-icon>
          </div>
          <div class="stat-content">
            <span class="stat-value">{{ stats.skills }}</span>
            <span class="stat-label">Skills</span>
          </div>
        </div>

        <div class="stat-card agents clickable" @click="goToAgents">
          <div class="stat-icon">
            <el-icon><User /></el-icon>
          </div>
          <div class="stat-content">
            <span class="stat-value">{{ stats.agents }}</span>
            <span class="stat-label">Agents</span>
          </div>
        </div>
      </div>

      <div class="quick-actions">
        <h2 class="section-title">快速操作</h2>
        <div class="actions-grid">
          <div class="action-card clickable" @click="goToPrompts">
            <el-icon class="action-icon"><Document /></el-icon>
            <span>管理提示词</span>
          </div>
          <div class="action-card clickable" @click="goToSkills">
            <el-icon class="action-icon"><Timer /></el-icon>
            <span>管理 Skills</span>
          </div>
          <div class="action-card clickable" @click="goToAgents">
            <el-icon class="action-icon"><User /></el-icon>
            <span>管理 Agents</span>
          </div>
          <div class="action-card clickable" @click="goToAnalytics">
            <el-icon class="action-icon"><DataAnalysis /></el-icon>
            <span>测试分析</span>
          </div>
          <div class="action-card clickable" @click="goToActivity">
            <el-icon class="action-icon"><Clock /></el-icon>
            <span>活动日志</span>
          </div>
          <div class="action-card clickable" @click="goToSettings">
            <el-icon class="action-icon"><Setting /></el-icon>
            <span>设置</span>
          </div>
          <div class="action-card clickable" @click="goToApiDocs">
            <el-icon class="action-icon"><Collection /></el-icon>
            <span>API 文档</span>
          </div>
          <div class="action-card clickable" @click="goToABTests">
            <el-icon class="action-icon"><DataLine /></el-icon>
            <span>A/B 测试</span>
          </div>
          <div class="action-card clickable" @click="goToTeams">
            <el-icon class="action-icon"><Users /></el-icon>
            <span>团队协作</span>
          </div>
          <div class="action-card clickable" @click="goToTemplates">
            <el-icon class="action-icon"><Shop /></el-icon>
            <span>模板市场</span>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'

const router = useRouter()
const stats = ref({
  prompts: 0,
  skills: 0,
  agents: 0
})

const fetchStats = async () => {
  try {
    const res = await axios.get('/api/stats')
    if (res.data.success) {
      stats.value = res.data.data
    }
  } catch (err) {
    console.error('Failed to fetch stats:', err)
  }
}

const goToPrompts = () => router.push('/prompts')
const goToSkills = () => router.push('/skills')
const goToAgents = () => router.push('/agents')
const goToActivity = () => router.push('/activity')
const goToSettings = () => router.push('/settings')
const goToApiDocs = () => router.push('/api-docs')
const goToABTests = () => router.push('/ab-tests')
const goToTeams = () => router.push('/teams')
const goToTemplates = () => router.push('/templates')

onMounted(fetchStats)
</script>

<style scoped>
.dashboard {
  min-height: 100vh;
  background: var(--color-bg);
}

.dashboard-main {
  padding: var(--spacing-8);
  max-width: 1200px;
  margin: 0 auto;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--spacing-6);
  margin-bottom: var(--spacing-8);
}

.stat-card {
  display: flex;
  align-items: center;
  gap: var(--spacing-5);
  padding: var(--spacing-5);
  background: var(--color-bg);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
  transition: box-shadow var(--transition-normal);
}

.stat-card:hover {
  box-shadow: var(--shadow-card-hover);
}

.stat-icon {
  width: 64px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-lg);
  font-size: 28px;
  flex-shrink: 0;
}

.stat-card.prompts .stat-icon {
  background: #f3f4f6;
  color: var(--color-primary);
}

.stat-card.skills .stat-icon {
  background: #ecfdf5;
  color: var(--color-success);
}

.stat-card.agents .stat-icon {
  background: #fffbeb;
  color: var(--color-warning);
}

.stat-content {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-semibold);
  letter-spacing: var(--tracking-tight);
  color: var(--color-text-primary);
  line-height: var(--line-height-tight);
}

.stat-label {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  margin-top: var(--spacing-1);
}

.section-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  letter-spacing: var(--tracking-tight);
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-4);
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--spacing-4);
}

.action-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-8);
  background: var(--color-bg);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
  transition: box-shadow var(--transition-normal);
  text-align: center;
}

.action-card:hover {
  box-shadow: var(--shadow-card-hover);
}

.action-icon {
  font-size: 32px;
  color: var(--color-primary);
  margin-bottom: var(--spacing-3);
}

.action-card span {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
}

@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .actions-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--spacing-3);
  }
  .actions-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .dashboard-main {
    padding: var(--spacing-4);
  }
  .stat-card {
    padding: var(--spacing-3);
    gap: var(--spacing-3);
  }
  .stat-icon {
    width: 44px;
    height: 44px;
    font-size: 22px;
  }
  .stat-value {
    font-size: var(--font-size-2xl);
  }
  .action-card {
    padding: var(--spacing-4);
  }
  .action-icon {
    font-size: 24px;
    margin-bottom: var(--spacing-2);
  }
}
</style>
