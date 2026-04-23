<template>
  <div class="team-list">
    <el-header>
      <div class="header-content">
        <div class="left-group">
          <el-button class="mobile-menu-btn" @click="showSidebar = true">
            <el-icon><Menu /></el-icon>
          </el-button>
        </div>
        <div class="actions-group">
          <el-button type="primary" @click="showCreateDialog = true">
            <el-icon><Plus /></el-icon>
            <span class="btn-text">新建团队</span>
          </el-button>
        </div>
      </div>
    </el-header>

    <el-main>
      <!-- Team Grid -->
      <div v-if="teams.length > 0" class="team-grid">
        <el-card
          v-for="team in teams"
          :key="team.id"
          class="team-card"
          @click="goToTeam(team)"
        >
          <template #header>
            <div class="card-header">
              <div class="team-avatar" :style="{ background: avatarColor(team.name) }">
                {{ team.avatar }}
              </div>
              <div class="team-info">
                <span class="team-name">{{ team.name }}</span>
                <span class="team-desc">{{ team.description }}</span>
              </div>
            </div>
          </template>

          <div class="card-body">
            <div class="member-avatars">
              <el-avatar
                v-for="member in team.members.slice(0, 5)"
                :key="member.id"
                :size="28"
                class="member-thumb"
              >
                {{ member.avatar }}
              </el-avatar>
              <el-avatar
                v-if="team.member_count > 5"
                :size="28"
                class="member-thumb more"
              >
                +{{ team.member_count - 5 }}
              </el-avatar>
            </div>
            <div class="team-meta">
              <span class="meta-item">
                <el-icon><User /></el-icon>
                {{ team.member_count }} 成员
              </span>
              <span class="meta-item">
                <el-icon><Timer /></el-icon>
                {{ formatDate(team.created_at) }}
              </span>
            </div>
          </div>

          <template #footer>
            <div class="card-footer">
              <el-button size="small" @click.stop="goToSettings(team.id)">
                <el-icon><Setting /></el-icon>
                设置
              </el-button>
              <el-button size="small" @click.stop="goToMembers(team.id)">
                <el-icon><User /></el-icon>
                成员
              </el-button>
              <el-button
                v-if="canDelete(team)"
                size="small"
                type="danger"
                text
                @click.stop="handleDelete(team)"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </template>
        </el-card>
      </div>

      <el-empty v-else description="暂无团队">
        <el-button type="primary" @click="showCreateDialog = true">
          创建第一个团队
        </el-button>
      </el-empty>
    </el-main>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreateDialog" title="新建团队" width="480px">
      <el-form :model="newTeam" label-position="top">
        <el-form-item label="团队名称" required>
          <el-input
            v-model="newTeam.name"
            placeholder="例如：AI Lab"
            :prefix-icon="Collection"
          />
        </el-form-item>
        <el-form-item label="团队描述">
          <el-input
            v-model="newTeam.description"
            type="textarea"
            :rows="3"
            placeholder="简短描述团队..."
          />
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
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { mockTeams, currentUser } from '../composables/useTeam'

const router = useRouter()
const teams = ref([])
const showCreateDialog = ref(false)
const showSidebar = ref(false)

const newTeam = ref({
  name: '',
  description: ''
})

onMounted(() => {
  teams.value = mockTeams.value
})

const avatarColors = [
  '#2563EB', '#059669', '#D97706', '#DC2626',
  '#7C3AED', '#0891B2', '#DB2777', '#65A30D'
]

const avatarColor = (name) => {
  const idx = name.charCodeAt(0) % avatarColors.length
  return avatarColors[idx]
}

const canDelete = (team) => {
  return currentUser.value.role === 'owner'
}

const goToTeam = (team) => router.push(`/teams/${team.id}/members`)
const goToSettings = (id) => router.push(`/teams/${id}/settings`)
const goToMembers = (id) => router.push(`/teams/${id}/members`)

const handleCreate = () => {
  if (!newTeam.value.name.trim()) {
    ElMessage.warning('请输入团队名称')
    return
  }
  const id = Date.now()
  const newEntry = {
    id,
    name: newTeam.value.name,
    description: newTeam.value.description,
    avatar: newTeam.value.name.substring(0, 2).toUpperCase(),
    created_at: new Date().toISOString(),
    member_count: 1,
    owner_id: currentUser.value.id,
    members: [{
      ...currentUser.value,
      role: 'owner',
      joined_at: new Date().toISOString(),
      status: 'active'
    }]
  }
  teams.value.push(newEntry)
  mockTeams.value.push(newEntry)
  ElMessage.success('团队创建成功')
  showCreateDialog.value = false
  newTeam.value = { name: '', description: '' }
}

const handleDelete = async (team) => {
  try {
    await ElMessageBox.confirm(
      `确定删除团队 "${team.name}" 吗？`,
      '删除确认',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
    )
    teams.value = teams.value.filter(t => t.id !== team.id)
    mockTeams.value = mockTeams.value.filter(t => t.id !== team.id)
    ElMessage.success('团队已删除')
  } catch {
    // cancelled
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return `${date.getMonth() + 1}/${date.getDate()}`
}
</script>

<style scoped>
.team-list {
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

.mobile-menu-btn {
  display: none;
  padding: var(--spacing-2);
}

.el-main {
  padding: var(--spacing-6);
}

.team-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: var(--spacing-5);
}

.team-card {
  cursor: pointer;
  transition: all var(--transition-normal);
}

.team-card:hover {
  transform: translateY(-2px);
  border-color: var(--color-border-hover);
}

.card-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.team-avatar {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: var(--font-weight-bold);
  font-size: var(--font-size-sm);
  flex-shrink: 0;
}

.team-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.team-name {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.team-desc {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.member-avatars {
  display: flex;
  gap: -8px;
}

.member-thumb {
  border: 2px solid var(--color-surface);
  margin-left: -8px;
  background: var(--color-primary-light);
  color: var(--color-primary);
  font-size: 10px;
  font-weight: var(--font-weight-semibold);
}

.member-thumb:first-child {
  margin-left: 0;
}

.member-thumb.more {
  background: var(--color-bg);
  color: var(--color-text-muted);
}

.team-meta {
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

.card-footer {
  display: flex;
  justify-content: space-between;
}

@media (max-width: 768px) {
  .team-grid {
    grid-template-columns: 1fr;
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
