<template>
  <div class="dashboard">
    <el-header>
      <div class="header-content">
        <div class="brand" @click="$router.push('/')">
          <div class="logo">
            <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
              <rect width="32" height="32" rx="8" fill="var(--color-primary)"/>
              <path d="M8 10h16M8 16h10M8 22h12" stroke="white" stroke-width="2.5" stroke-linecap="round"/>
            </svg>
          </div>
          <div class="brand-text">
            <h1>AI Hub</h1>
            <span class="subtitle">AI 周边服务管理</span>
          </div>
        </div>
      </div>
    </el-header>

    <el-main>
      <div class="stats-grid">
        <el-card class="stat-card prompts" @click="goToPrompts">
          <div class="stat-icon">
            <el-icon><Document /></el-icon>
          </div>
          <div class="stat-content">
            <span class="stat-value">{{ stats.prompts }}</span>
            <span class="stat-label">提示词</span>
          </div>
        </el-card>

        <el-card class="stat-card skills" @click="goToSkills">
          <div class="stat-icon">
            <el-icon><Timer /></el-icon>
          </div>
          <div class="stat-content">
            <span class="stat-value">{{ stats.skills }}</span>
            <span class="stat-label">Skills</span>
          </div>
        </el-card>

        <el-card class="stat-card agents" @click="goToAgents">
          <div class="stat-icon">
            <el-icon><User /></el-icon>
          </div>
          <div class="stat-content">
            <span class="stat-value">{{ stats.agents }}</span>
            <span class="stat-label">Agents</span>
          </div>
        </el-card>
      </div>

      <div class="quick-actions">
        <h2>快速操作</h2>
        <div class="actions-grid">
          <el-card class="action-card" @click="goToPrompts">
            <el-icon class="action-icon"><Document /></el-icon>
            <span>管理提示词</span>
          </el-card>
          <el-card class="action-card" @click="goToSkills">
            <el-icon class="action-icon"><Timer /></el-icon>
            <span>管理 Skills</span>
          </el-card>
          <el-card class="action-card" @click="goToAgents">
            <el-icon class="action-icon"><User /></el-icon>
            <span>管理 Agents</span>
          </el-card>
          <el-card class="action-card" @click="goToAnalytics">
            <el-icon class="action-icon"><DataAnalysis /></el-icon>
            <span>测试分析</span>
          </el-card>
          <el-card class="action-card" @click="goToActivity">
            <el-icon class="action-icon"><Clock /></el-icon>
            <span>活动日志</span>
          </el-card>
          <el-card class="action-card" @click="goToSettings">
            <el-icon class="action-icon"><Setting /></el-icon>
            <span>设置</span>
          </el-card>
          <el-card class="action-card" @click="goToApiDocs">
            <el-icon class="action-icon"><Collection /></el-icon>
            <span>API 文档</span>
          </el-card>
          <el-card class="action-card" @click="goToABTests">
            <el-icon class="action-icon"><DataLine /></el-icon>
            <span>A/B 测试</span>
          </el-card>
          <el-card class="action-card" @click="goToTeams">
            <el-icon class="action-icon"><Users /></el-icon>
            <span>团队协作</span>
          </el-card>
          <el-card class="action-card" @click="goToTemplates">
            <el-icon class="action-icon"><Shop /></el-icon>
            <span>模板市场</span>
          </el-card>
        </div>
      </div>
    </el-main>
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
}

.brand {
  display: flex;
  align-items: center;
  gap: var(--spacing-4);
}

.brand-text h1 {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  margin: 0;
}

.subtitle {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
}

.el-main {
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
  cursor: pointer;
  transition: all var(--transition-normal);
}

.stat-card:hover {
  transform: translateY(-4px);
}

.stat-icon {
  width: 64px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-lg);
  font-size: 28px;
}

.stat-card.prompts .stat-icon {
  background: var(--color-primary-light);
  color: var(--color-primary);
}

.stat-card.skills .stat-icon {
  background: var(--color-success-light);
  color: var(--color-success);
}

.stat-card.agents .stat-icon {
  background: var(--color-warning-light);
  color: var(--color-warning);
}

.stat-content {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  line-height: 1;
}

.stat-label {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  margin-top: var(--spacing-1);
}

.quick-actions h2 {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
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
  cursor: pointer;
  transition: all var(--transition-normal);
  text-align: center;
}

.action-card:hover {
  transform: translateY(-2px);
}

.action-icon {
  font-size: 48px;
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
  .el-header {
    padding: 0 var(--spacing-4);
    height: 56px;
  }
  .subtitle {
    display: none;
  }
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--spacing-3);
  }
  .actions-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .el-main {
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
