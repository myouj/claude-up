<template>
  <div class="team-member-list">
    <div class="header">
      <div class="header-left">
        <el-button class="back-btn" @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <div class="title-info">
          <h2>成员管理</h2>
          <span class="team-name">{{ team?.name }}</span>
        </div>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="showInvite = true">
          <el-icon><Plus /></el-icon>
          邀请成员
        </el-button>
      </div>
    </div>

    <div class="content">
      <!-- Member Table -->
      <div class="member-table">
        <div class="table-header">
          <span class="col-member">成员</span>
          <span class="col-role">角色</span>
          <span class="col-joined">加入时间</span>
          <span class="col-status">状态</span>
          <span class="col-actions">操作</span>
        </div>

        <div
          v-for="member in members"
          :key="member.id"
          class="table-row"
        >
          <div class="col-member">
            <el-avatar :size="36" class="member-avatar" :class="member.role">
              {{ member.avatar }}
            </el-avatar>
            <div class="member-info">
              <span class="member-name">{{ member.name }}</span>
              <span class="member-email">{{ member.email }}</span>
            </div>
          </div>

          <div class="col-role">
            <el-tag
              v-if="member.role === 'owner'"
              type="warning"
              size="small"
              effect="dark"
            >Owner</el-tag>
            <el-tag
              v-else-if="member.role === 'admin'"
              type="primary"
              size="small"
            >Admin</el-tag>
            <el-tag
              v-else-if="member.role === 'member'"
              type="info"
              size="small"
            >Member</el-tag>
            <el-tag
              v-else
              type="info"
              size="small"
              effect="plain"
            >Viewer</el-tag>
          </div>

          <div class="col-joined">
            <span class="join-date">{{ formatDate(member.joined_at) }}</span>
          </div>

          <div class="col-status">
            <el-tag
              v-if="member.status === 'active'"
              type="success"
              size="small"
              effect="plain"
            >
              <el-icon><CircleCheck /></el-icon>
              活跃
            </el-tag>
            <el-tag
              v-else
              type="warning"
              size="small"
              effect="plain"
            >
              <el-icon><Clock /></el-icon>
              待确认
            </el-tag>
          </div>

          <div class="col-actions">
            <template v-if="member.role !== 'owner' && canManage">
              <el-dropdown trigger="click" @command="(cmd) => handleAction(cmd, member)">
                <el-button size="small" text>
                  <el-icon><MoreFilled /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="invite" :disabled="member.status !== 'pending'">
                      <el-icon><Message /></el-icon>
                      重发邀请
                    </el-dropdown-item>
                    <el-dropdown-item
                      v-for="role in editableRoles"
                      :key="role.value"
                      :command="`role:${role.value}`"
                      :disabled="member.role === role.value"
                    >
                      <el-icon><User /></el-icon>
                      变更为 {{ role.label }}
                    </el-dropdown-item>
                    <el-dropdown-item
                      command="remove"
                      divided
                      style="color: var(--color-danger)"
                    >
                      <el-icon><Delete /></el-icon>
                      移除成员
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
          </div>
        </div>
      </div>

      <!-- Role Legend -->
      <div class="role-legend">
        <h4>角色说明</h4>
        <div class="legend-grid">
          <div v-for="role in roleOptions" :key="role.value" class="legend-item">
            <el-tag :type="roleTagType(role.value)" size="small">
              {{ role.label }}
            </el-tag>
            <span class="legend-desc">{{ role.description }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Role Change Dialog -->
    <el-dialog v-model="showRoleDialog" title="变更角色" width="400px">
      <div class="role-change-content">
        <p>将 <strong>{{ selectedMember?.name }}</strong> 的角色变更为：</p>
        <el-radio-group v-model="newRole" class="role-radio-group">
          <el-radio
            v-for="role in editableRoles"
            :key="role.value"
            :value="role.value"
          >
            {{ role.label }}
            <span class="role-desc-inline">{{ role.description }}</span>
          </el-radio>
        </el-radio-group>
      </div>
      <template #footer>
        <el-button @click="showRoleDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmRoleChange">确认变更</el-button>
      </template>
    </el-dialog>

    <InviteModal v-model="showInvite" :team-id="teamId" @invited="handleInvited" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { mockTeams, currentUser, roleOptions } from '../composables/useTeam'
import InviteModal from '../components/InviteModal.vue'

const props = defineProps({
  teamId: {
    type: Number,
    required: true
  }
})

const router = useRouter()

const team = computed(() => mockTeams.value.find(t => t.id === props.teamId))
const members = computed(() => team.value?.members || [])

const canManage = computed(() => {
  const user = currentUser.value
  return user.role === 'owner' || user.role === 'admin'
})

const editableRoles = computed(() => roleOptions.filter(r => r.value !== 'owner'))

const showInvite = ref(false)
const showRoleDialog = ref(false)
const selectedMember = ref(null)
const newRole = ref('member')

const roleTagType = (role) => {
  const map = { owner: 'warning', admin: 'primary', member: 'info', viewer: 'info' }
  return map[role] || 'info'
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return `${date.getFullYear()}/${date.getMonth() + 1}/${date.getDate()}`
}

const handleAction = (cmd, member) => {
  if (cmd === 'remove') {
    handleRemove(member)
  } else if (cmd === 'invite') {
    ElMessage.success(`已重发邀请到 ${member.email}`)
  } else if (cmd.startsWith('role:')) {
    selectedMember.value = member
    newRole.value = cmd.replace('role:', '')
    showRoleDialog.value = true
  }
}

const confirmRoleChange = () => {
  if (!selectedMember.value) return
  selectedMember.value.role = newRole.value
  ElMessage.success(`${selectedMember.value.name} 的角色已更新为 ${newRole.value}`)
  showRoleDialog.value = false
}

const handleRemove = async (member) => {
  try {
    await ElMessageBox.confirm(
      `确定要将 "${member.name}" 从团队中移除吗？`,
      '移除成员',
      { confirmButtonText: '移除', cancelButtonText: '取消', type: 'warning' }
    )
    if (team.value) {
      team.value.members = team.value.members.filter(m => m.id !== member.id)
      team.value.member_count--
    }
    ElMessage.success('成员已移除')
  } catch {
    // cancelled
  }
}

const handleInvited = (data) => {
  if (team.value && data.email) {
    const newMember = {
      id: Date.now(),
      name: data.email.split('@')[0],
      email: data.email,
      avatar: data.email.charAt(0).toUpperCase(),
      role: data.role,
      joined_at: new Date().toISOString(),
      status: 'pending'
    }
    team.value.members.push(newMember)
    team.value.member_count++
  }
}

const goBack = () => router.back()
</script>

<style scoped>
.team-member-list {
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

.header-right {
  display: flex;
  gap: var(--spacing-2);
}

.content {
  flex: 1;
  padding: var(--spacing-6);
  overflow-y: auto;
}

.member-table {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.table-header {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr 80px;
  gap: var(--spacing-4);
  padding: var(--spacing-3) var(--spacing-5);
  background: var(--color-bg);
  border-bottom: 1px solid var(--color-border);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.table-row {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr 80px;
  gap: var(--spacing-4);
  padding: var(--spacing-4) var(--spacing-5);
  align-items: center;
  border-bottom: 1px solid var(--color-border);
  transition: background var(--transition-fast);
}

.table-row:last-child {
  border-bottom: none;
}

.table-row:hover {
  background: var(--color-bg);
}

.col-member {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.member-avatar {
  background: var(--color-primary-light);
  color: var(--color-primary);
  font-weight: var(--font-weight-semibold);
  flex-shrink: 0;
}

.member-avatar.owner {
  background: var(--color-warning-light);
  color: var(--color-warning);
}

.member-avatar.admin {
  background: color-mix(in srgb, var(--color-primary) 15%, var(--color-surface));
  color: var(--color-primary);
}

.member-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.member-name {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.member-email {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.join-date {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

.col-actions {
  display: flex;
  justify-content: center;
}

/* Role Legend */
.role-legend {
  margin-top: var(--spacing-6);
  padding: var(--spacing-5);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.role-legend h4 {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-4);
}

.legend-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-3);
}

.legend-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.legend-desc {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

/* Role Change Dialog */
.role-change-content p {
  margin: 0 0 var(--spacing-4);
  color: var(--color-text-secondary);
}

.role-radio-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.role-radio-group :deep(.el-radio) {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-2);
  line-height: 1.5;
}

.role-desc-inline {
  display: block;
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  font-weight: var(--font-weight-normal);
  margin-left: 20px;
}

@media (max-width: 768px) {
  .table-header {
    display: none;
  }

  .table-row {
    grid-template-columns: 1fr auto;
    gap: var(--spacing-2);
  }

  .col-role,
  .col-joined,
  .col-status {
    display: none;
  }

  .legend-grid {
    grid-template-columns: 1fr;
  }

  .el-main {
    padding: var(--spacing-3);
  }
}
</style>
