import { ref } from 'vue'

export const mockTeams = ref([
  {
    id: 1,
    name: 'AI Lab',
    description: 'AI 应用研发团队',
    avatar: 'AI',
    created_at: '2026-01-15T10:00:00Z',
    member_count: 5,
    owner_id: 1,
    members: [
      {
        id: 1,
        name: '张明',
        email: 'zhangming@example.com',
        avatar: '张',
        role: 'owner',
        joined_at: '2026-01-15T10:00:00Z',
        status: 'active'
      },
      {
        id: 2,
        name: '李华',
        email: 'lihua@example.com',
        avatar: '李',
        role: 'admin',
        joined_at: '2026-01-20T14:30:00Z',
        status: 'active'
      },
      {
        id: 3,
        name: '王芳',
        email: 'wangfang@example.com',
        avatar: '王',
        role: 'member',
        joined_at: '2026-02-01T09:15:00Z',
        status: 'active'
      },
      {
        id: 4,
        name: '赵伟',
        email: 'zhaowei@example.com',
        avatar: '赵',
        role: 'member',
        joined_at: '2026-02-10T11:00:00Z',
        status: 'active'
      },
      {
        id: 5,
        name: '陈静',
        email: 'chenjing@example.com',
        avatar: '陈',
        role: 'viewer',
        joined_at: '2026-03-01T16:45:00Z',
        status: 'pending'
      }
    ]
  },
  {
    id: 2,
    name: 'Prompt Engineering',
    description: 'Prompt 工程化研究组',
    avatar: 'PE',
    created_at: '2026-02-01T08:00:00Z',
    member_count: 3,
    owner_id: 2,
    members: [
      {
        id: 2,
        name: '李华',
        email: 'lihua@example.com',
        avatar: '李',
        role: 'owner',
        joined_at: '2026-02-01T08:00:00Z',
        status: 'active'
      },
      {
        id: 6,
        name: '周杰',
        email: 'zhoujie@example.com',
        avatar: '周',
        role: 'admin',
        joined_at: '2026-02-05T10:00:00Z',
        status: 'active'
      },
      {
        id: 7,
        name: '吴婷',
        email: 'wuting@example.com',
        avatar: '吴',
        role: 'member',
        joined_at: '2026-02-10T15:00:00Z',
        status: 'active'
      }
    ]
  }
])

export const currentUser = ref({
  id: 1,
  name: '张明',
  email: 'zhangming@example.com',
  avatar: '张',
  role: 'owner'
})

export const roleOptions = [
  { label: 'Owner', value: 'owner', description: '完全控制权，可删除团队' },
  { label: 'Admin', value: 'admin', description: '管理成员和设置，无法删除团队' },
  { label: 'Member', value: 'member', description: '创建和编辑提示词' },
  { label: 'Viewer', value: 'viewer', description: '仅查看权限' }
]

export function useTeam() {
  return { mockTeams, currentUser, roleOptions }
}
