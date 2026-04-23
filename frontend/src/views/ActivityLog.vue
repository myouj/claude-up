<template>
  <div class="activity-log">
    <BreadcrumbNav :items="[{ name: '活动日志' }]" />
    <el-header>
      <div class="header-content">
        <div class="right">
          <el-select v-model="filterEntityType" placeholder="全部类型" clearable class="filter-select">
            <el-option label="提示词" value="prompt" />
            <el-option label="Skill" value="skill" />
            <el-option label="Agent" value="agent" />
            <el-option label="版本" value="version" />
            <el-option label="测试" value="test" />
          </el-select>
          <el-select v-model="filterAction" placeholder="全部操作" clearable class="filter-select">
            <el-option label="创建" value="created" />
            <el-option label="更新" value="updated" />
            <el-option label="删除" value="deleted" />
            <el-option label="克隆" value="cloned" />
            <el-option label="测试" value="tested" />
            <el-option label="优化" value="optimized" />
            <el-option label="翻译" value="translated" />
            <el-option label="收藏" value="favorited" />
          </el-select>
          <el-button @click="fetchLogs">
            <el-icon><Refresh /></el-icon>
          </el-button>
        </div>
      </div>
    </el-header>

    <el-main>
      <div v-if="loading" class="loading-state">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>加载日志...</span>
      </div>

      <div v-else-if="logs.length === 0" class="empty-state">
        <div class="empty-icon">
          <svg width="80" height="80" viewBox="0 0 80 80" fill="none">
            <rect x="15" y="15" width="50" height="50" rx="8" stroke="var(--color-border)" stroke-width="2"/>
            <path d="M25 30h30M25 40h20M25 50h25" stroke="var(--color-border)" stroke-width="2" stroke-linecap="round"/>
          </svg>
        </div>
        <p>暂无活动记录</p>
      </div>

      <el-card v-else class="log-table-card">
        <el-table :data="logs" stripe class="log-table" v-loading="loading">
          <el-table-column label="时间" width="160" prop="created_at" sortable />
          <el-table-column label="操作" width="110">
            <template #default="{ row }">
              <el-tag size="small" :type="actionTagType(row.action)">
                {{ actionLabel(row.action) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="实体类型" width="110">
            <template #default="{ row }">
              <el-tag size="small" type="info">{{ entityLabel(row.entity_type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="实体ID" width="90">
            <template #default="{ row }">
              <span class="entity-id">#{{ row.entity_id }}</span>
            </template>
          </el-table-column>
          <el-table-column label="详情" min-width="200">
            <template #default="{ row }">
              <span class="details-text">{{ row.details || '-' }}</span>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination-wrapper">
          <el-pagination
            v-model:current-page="currentPage"
            :page-size="pageSize"
            :total="totalLogs"
            layout="prev, pager, next, total"
            background
            @current-change="fetchLogs"
          />
        </div>
      </el-card>
    </el-main>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import BreadcrumbNav from '../components/BreadcrumbNav.vue'

const router = useRouter()

const logs = ref([])
const loading = ref(true)
const currentPage = ref(1)
const pageSize = ref(20)
const totalLogs = ref(0)
const filterEntityType = ref('')
const filterAction = ref('')

const actionMap = {
  created: '创建',
  updated: '更新',
  deleted: '删除',
  cloned: '克隆',
  tested: '测试',
  optimized: '优化',
  translated: '翻译',
  favorited: '收藏'
}

const actionTagMap = {
  created: 'success',
  updated: 'primary',
  deleted: 'danger',
  cloned: 'info',
  tested: 'warning',
  optimized: 'success',
  translated: 'info',
  favorited: 'warning'
}

const entityMap = {
  prompt: '提示词',
  skill: 'Skill',
  agent: 'Agent',
  version: '版本',
  test: '测试'
}

const actionLabel = (action) => actionMap[action] || action
const actionTagType = (action) => actionTagMap[action] || 'info'
const entityLabel = (type) => entityMap[type] || type

const fetchLogs = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      limit: pageSize.value
    }
    if (filterEntityType.value) params.entity_type = filterEntityType.value
    if (filterAction.value) params.action = filterAction.value

    const res = await axios.get('/api/activity-logs', { params })
    if (res.data.success) {
      logs.value = res.data.data
      if (res.data.meta) {
        totalLogs.value = res.data.meta.total
      }
    }
  } catch (err) {
    ElMessage.error('加载日志失败')
  } finally {
    loading.value = false
  }
}

watch(filterEntityType, () => {
  currentPage.value = 1
  fetchLogs()
})

watch(filterAction, () => {
  currentPage.value = 1
  fetchLogs()
})

onMounted(fetchLogs)
</script>

<style scoped>
.activity-log {
  height: 100vh;
  background: var(--color-bg);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.el-header {
  background: var(--color-bg);
  box-shadow: var(--shadow-border);
  display: flex;
  align-items: center;
  padding: 0 var(--spacing-6);
  height: 64px;
  flex-shrink: 0;
}

.header-content {
  width: 100%;
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: var(--spacing-4);
}

.right {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.filter-select {
  width: 130px;
}

.el-main {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-5);
}

.loading-state,
.empty-state {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-3);
  color: var(--color-text-muted);
}

.empty-icon {
  opacity: 0.5;
}

.empty-state p {
  font-size: var(--font-size-md);
  color: var(--color-text-secondary);
  margin: 0;
}

.log-table-card :deep(.el-card__body) {
  padding: 0;
}

.log-table {
  font-size: var(--font-size-sm);
}

.entity-id {
  font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.details-text {
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
  max-width: 300px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  padding: var(--spacing-4);
  border-top: 1px solid var(--color-border);
}

/* Responsive - Mobile */
@media (max-width: 768px) {
  .header-content {
    flex-wrap: wrap;
  }

  .left h1 {
    font-size: var(--font-size-md);
  }

  .right {
    width: 100%;
    overflow-x: auto;
    flex-wrap: nowrap;
  }

  .filter-select {
    min-width: 100px;
  }

  .el-main {
    padding: var(--spacing-3);
  }

  .details-text {
    max-width: 150px;
  }
}
</style>
