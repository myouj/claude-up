<template>
  <div class="team-settings">
    <div class="header">
      <div class="header-left">
        <el-button class="back-btn" @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <div class="title-info">
          <h2>团队设置</h2>
          <span class="team-name">{{ team?.name }}</span>
        </div>
      </div>
    </div>

    <el-main v-if="team">
      <div class="settings-grid">
        <!-- Basic Info -->
        <el-card class="settings-card">
          <template #header>
            <div class="card-title">
              <el-icon><Setting /></el-icon>
              <span>基本信息</span>
            </div>
          </template>
          <el-form :model="form" label-position="top">
            <el-form-item label="团队名称">
              <el-input
                v-model="form.name"
                :disabled="!canEdit"
                placeholder="输入团队名称"
              />
            </el-form-item>
            <el-form-item label="团队描述">
              <el-input
                v-model="form.description"
                type="textarea"
                :rows="3"
                :disabled="!canEdit"
                placeholder="简短描述团队..."
              />
            </el-form-item>
            <el-form-item label="创建时间">
              <span class="info-value">{{ formatDate(team.created_at) }}</span>
            </el-form-item>
            <el-form-item v-if="canEdit">
              <el-button type="primary" @click="saveBasic">
                保存更改
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- Danger Zone -->
        <el-card class="settings-card danger-card" v-if="canEdit">
          <template #header>
            <div class="card-title danger">
              <el-icon><Warning /></el-icon>
              <span>危险区域</span>
            </div>
          </template>
          <div class="danger-actions">
            <div class="danger-item">
              <div class="danger-info">
                <h4>邀请链接</h4>
                <p>生成或撤销团队邀请链接</p>
              </div>
              <el-button @click="generateInviteLink">
                <el-icon><Link /></el-icon>
                生成邀请链接
              </el-button>
            </div>
            <div v-if="inviteLink" class="invite-link-display">
              <el-input :model-value="inviteLink" readonly>
                <template #append>
                  <el-button @click="copyLink">
                    <el-icon><CopyDocument /></el-icon>
                  </el-button>
                </template>
              </el-input>
            </div>
            <div class="danger-item">
              <div class="danger-info">
                <h4>转让团队</h4>
                <p>将团队所有权转让给其他成员</p>
              </div>
              <el-button type="warning" @click="showTransferDialog = true">
                转让所有权
              </el-button>
            </div>
            <div class="danger-item">
              <div class="danger-info">
                <h4>删除团队</h4>
                <p>永久删除此团队及所有数据。此操作不可恢复。</p>
              </div>
              <el-button type="danger" @click="handleDelete">
                删除团队
              </el-button>
            </div>
          </div>
        </el-card>

        <!-- Team Stats -->
        <el-card class="settings-card">
          <template #header>
            <div class="card-title">
              <el-icon><DataAnalysis /></el-icon>
              <span>团队统计</span>
            </div>
          </template>
          <div class="stats-grid">
            <div class="stat-item">
              <span class="stat-value">{{ team.member_count }}</span>
              <span class="stat-label">成员数</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ teamPrompts }}</span>
              <span class="stat-label">提示词</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ teamSkills }}</span>
              <span class="stat-label">Skills</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ teamAgents }}</span>
              <span class="stat-label">Agents</span>
            </div>
          </div>
        </el-card>
      </div>
    </el-main>

    <!-- Transfer Dialog -->
    <el-dialog v-model="showTransferDialog" title="转让团队所有权" width="480px">
      <p class="transfer-desc">选择要接收团队所有权的成员：</p>
      <el-select v-model="transferToId" placeholder="选择成员" class="full-width">
        <el-option
          v-for="member in transferCandidates"
          :key="member.id"
          :label="member.name"
          :value="member.id"
        >
          <div class="member-option">
            <el-avatar :size="24">{{ member.avatar }}</el-avatar>
            <span>{{ member.name }}</span>
            <span class="member-email">{{ member.email }}</span>
          </div>
        </el-option>
      </el-select>
      <template #footer>
        <el-button @click="showTransferDialog = false">取消</el-button>
        <el-button type="warning" @click="confirmTransfer" :disabled="!transferToId">
          确认转让
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { mockTeams, currentUser } from '../composables/useTeam'

const props = defineProps({
  teamId: {
    type: Number,
    required: true
  }
})

const router = useRouter()

const team = computed(() => mockTeams.value.find(t => t.id === props.teamId))
const form = computed(() => ({
  name: team.value?.name || '',
  description: team.value?.description || ''
}))

const canEdit = computed(() => {
  return currentUser.value.role === 'owner' || currentUser.value.role === 'admin'
})

const teamPrompts = ref(12)
const teamSkills = ref(5)
const teamAgents = ref(3)

const inviteLink = ref('')
const showTransferDialog = ref(false)
const transferToId = ref(null)

const transferCandidates = computed(() => {
  return team.value?.members.filter(m => m.role !== 'owner') || []
})

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return `${date.getFullYear()}/${date.getMonth() + 1}/${date.getDate()}`
}

const saveBasic = () => {
  if (team.value) {
    team.value.name = form.value.name
    team.value.description = form.value.description
  }
  ElMessage.success('设置已保存')
}

const generateInviteLink = () => {
  inviteLink.value = `https://aivault.app/join/${props.teamId}?token=${Date.now().toString(36)}`
  ElMessage.success('邀请链接已生成')
}

const copyLink = () => {
  navigator.clipboard.writeText(inviteLink.value)
  ElMessage.success('链接已复制到剪贴板')
}

const confirmTransfer = async () => {
  if (!transferToId.value) return
  const newOwner = team.value?.members.find(m => m.id === transferToId.value)
  if (!newOwner) return

  try {
    await ElMessageBox.confirm(
      `确定要将 "${team.value?.name}" 的所有权转让给 "${newOwner.name}" 吗？你将失去 Owner 权限。`,
      '确认转让',
      { confirmButtonText: '确认转让', cancelButtonText: '取消', type: 'warning' }
    )
    if (team.value) {
      const oldOwner = team.value.members.find(m => m.role === 'owner')
      if (oldOwner) oldOwner.role = 'admin'
      newOwner.role = 'owner'
      currentUser.value.role = 'admin'
    }
    ElMessage.success('团队所有权已转让')
    showTransferDialog.value = false
    transferToId.value = null
  } catch {
    // cancelled
  }
}

const handleDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要永久删除团队 "${team.value?.name}" 吗？此操作不可恢复，所有提示词、Skills、Agents 将被删除。`,
      '删除团队',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'error' }
    )
    mockTeams.value = mockTeams.value.filter(t => t.id !== props.teamId)
    ElMessage.success('团队已删除')
    router.push('/teams')
  } catch {
    // cancelled
  }
}

const goBack = () => router.back()
</script>

<style scoped>
.team-settings {
  height: 100vh;
  background: var(--color-bg);
  display: flex;
  flex-direction: column;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-4) var(--spacing-6);
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  height: 64px;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.back-btn {
  padding: var(--spacing-2);
}

.title-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.title-info h2 {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0;
}

.team-name {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
  padding-left: var(--spacing-3);
  border-left: 1px solid var(--color-border);
}

.el-main {
  padding: var(--spacing-6);
}

.settings-grid {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-5);
  max-width: 800px;
}

.settings-card {
  border-radius: var(--radius-lg);
}

.card-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.card-title.danger {
  color: var(--color-danger);
}

.info-value {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

/* Danger Zone */
.danger-card {
  border-color: var(--color-danger-light);
}

.danger-actions {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4);
}

.danger-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-4);
  padding: var(--spacing-4);
  background: var(--color-bg);
  border-radius: var(--radius-md);
}

.danger-info h4 {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-1);
}

.danger-info p {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  margin: 0;
}

.invite-link-display {
  padding: 0;
}

/* Stats */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-4);
}

.stat-item {
  text-align: center;
  padding: var(--spacing-3);
  background: var(--color-bg);
  border-radius: var(--radius-md);
}

.stat-value {
  display: block;
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-primary);
  margin-bottom: var(--spacing-1);
}

.stat-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

/* Transfer Dialog */
.transfer-desc {
  margin: 0 0 var(--spacing-4);
  color: var(--color-text-secondary);
}

.member-option {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.member-email {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  margin-left: var(--spacing-2);
}

.full-width {
  width: 100%;
}

@media (max-width: 768px) {
  .el-main {
    padding: var(--spacing-3);
  }

  .danger-item {
    flex-direction: column;
    align-items: flex-start;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
