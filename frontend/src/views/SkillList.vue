<template>
  <div class="skill-list">
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
            <span class="btn-text">新建 Skill</span>
          </el-button>
          <el-button @click="handleExport">
            <el-icon><Download /></el-icon>
            <span class="btn-text">导出</span>
          </el-button>
          <el-button @click="showImportDialog = true">
            <el-icon><Upload /></el-icon>
            <span class="btn-text">导入</span>
          </el-button>
        </div>
      </div>
    </el-header>

    <el-drawer v-model="showSidebar" title="筛选" size="280px" direction="ltr">
      <div class="drawer-content">
        <el-button type="primary" @click="showCreateDialog = true; showSidebar = false">
          <el-icon><Plus /></el-icon>
          新建 Skill
        </el-button>
        <el-button @click="handleExport">
          <el-icon><Download /></el-icon>
          导出
        </el-button>
        <el-button @click="showImportDialog = true; showSidebar = false">
          <el-icon><Upload /></el-icon>
          导入
        </el-button>
      </div>
    </el-drawer>

    <el-main>
      <div v-if="skills.length > 0" class="skill-grid">
        <el-card
          v-for="skill in paginatedSkills"
          :key="skill.id"
          class="skill-card"
          :class="{ builtin: skill.source === 'builtin' }"
          @click="goToEditor(skill.id)"
        >
          <template #header>
            <div class="card-header">
              <div class="title-row">
                <el-tag v-if="skill.source === 'builtin'" type="success" size="small">内置</el-tag>
                <el-tag v-else type="info" size="small">自定义</el-tag>
                <span class="name">{{ skill.name }}</span>
              </div>
            </div>
          </template>

          <div class="card-body">
            <p class="description">{{ skill.description || '暂无描述' }}</p>
            <div class="meta">
              <el-tag v-if="skill.category" size="small" type="info">{{ skill.category }}</el-tag>
              <span v-if="skill.content_cn" class="translated-badge">
                <el-icon><Check /></el-icon>
                已翻译
              </span>
            </div>
          </div>

          <template #footer>
            <div class="card-footer">
              <div class="footer-left">
                <el-button
                  size="small"
                  text
                  @click.stop="handleClone(skill)"
                  class="clone-btn"
                  title="克隆"
                  :disabled="skill.source === 'builtin'"
                >
                  <el-icon><CopyDocument /></el-icon>
                </el-button>
                <el-button size="small" @click.stop="goToTranslate(skill.id)">
                  <el-icon><Translate /></el-icon>
                  翻译
                </el-button>
              </div>
              <el-button
                v-if="skill.source !== 'builtin'"
                size="small"
                type="danger"
                @click.stop="handleDelete(skill)"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </template>
        </el-card>
      </div>

      <el-empty v-else description="暂无 Skills">
        <el-button type="primary" @click="showCreateDialog = true">
          创建第一个 Skill
        </el-button>
      </el-empty>

      <div v-if="totalSkills > pageSize" class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="totalSkills"
          layout="prev, pager, next"
          background
        />
      </div>
    </el-main>

    <el-dialog v-model="showCreateDialog" title="新建 Skill" width="520px">
      <el-form :model="newSkill" label-position="top">
        <el-form-item label="名称" required>
          <el-input v-model="newSkill.name" placeholder="如: /commit" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="newSkill.description" type="textarea" :rows="2" placeholder="简短描述" />
        </el-form-item>
        <el-form-item label="内容" required>
          <el-input v-model="newSkill.content" type="textarea" :rows="6" placeholder="Skill 内容..." />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="newSkill.category" placeholder="选择分类" allow-create filterable clearable>
            <el-option label="git" value="git" />
            <el-option label="code" value="code" />
            <el-option label="docs" value="docs" />
            <el-option label="security" value="security" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showImportDialog" title="导入 Skills" width="560px">
      <el-form :model="{ importType, importText }" label-position="top">
        <el-form-item label="导入格式">
          <el-radio-group v-model="importType">
            <el-radio label="json">JSON</el-radio>
            <el-radio label="md">Markdown</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="导入内容">
          <el-input
            v-model="importText"
            type="textarea"
            :rows="8"
            placeholder="粘贴 JSON 数据..."
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showImportDialog = false">取消</el-button>
        <el-button type="primary" @click="handleImport">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Menu } from '@element-plus/icons-vue'

const router = useRouter()
const skills = ref([])
const showCreateDialog = ref(false)
const showImportDialog = ref(false)
const showSidebar = ref(false)
const currentPage = ref(1)
const pageSize = ref(12)
const totalSkills = ref(0)
const importType = ref('json')
const importText = ref('')

const newSkill = ref({
  name: '',
  description: '',
  content: '',
  category: ''
})

const fetchSkills = async () => {
  try {
    const res = await axios.get('/api/skills', {
      params: { page: currentPage.value, limit: pageSize.value }
    })
    if (res.data.success) {
      skills.value = res.data.data
      if (res.data.meta) {
        totalSkills.value = res.data.meta.total
      }
    }
  } catch (err) {
    console.error('Failed to fetch skills:', err)
  }
}

const paginatedSkills = computed(() => skills.value)

const handleClone = async (skill) => {
  try {
    const res = await axios.post(`/api/skills/${skill.id}/clone`)
    if (res.data.success) {
      ElMessage.success('克隆成功')
      fetchSkills()
    }
  } catch (err) {
    ElMessage.error('克隆失败')
  }
}

const handleExport = async () => {
  try {
    const res = await axios.get('/api/skills/export')
    if (res.data.success) {
      const content = JSON.stringify(res.data.data, null, 2)
      const blob = new Blob([content], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'skills.json'
      a.click()
      URL.revokeObjectURL(url)
      ElMessage.success('导出成功')
    }
  } catch (err) {
    ElMessage.error('导出失败')
  }
}

const handleImport = () => {
  if (!importText.value.trim()) {
    ElMessage.warning('请输入要导入的内容')
    return
  }
  try {
    const payload = JSON.parse(importText.value)
    axios.post('/api/skills/import', { skills: payload.skills || payload })
      .then(res => {
        if (res.data.success) {
          ElMessage.success(`成功导入 ${res.data.imported} 条 Skills`)
          showImportDialog.value = false
          importText.value = ''
          currentPage.value = 1
          fetchSkills()
        }
      })
      .catch(() => ElMessage.error('导入失败'))
  } catch (err) {
    ElMessage.error('JSON 解析失败，请检查格式')
  }
}

const handleCreate = async () => {
  if (!newSkill.value.name || !newSkill.value.content) {
    ElMessage.warning('请填写名称和内容')
    return
  }
  try {
    await axios.post('/api/skills', newSkill.value)
    ElMessage.success('创建成功')
    showCreateDialog.value = false
    newSkill.value = { name: '', description: '', content: '', category: '' }
    fetchSkills()
  } catch (err) {
    ElMessage.error('创建失败')
  }
}

const handleDelete = async (skill) => {
  try {
    await ElMessageBox.confirm(`确定删除 "${skill.name}" 吗？`, '删除确认', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await axios.delete(`/api/skills/${skill.id}`)
    ElMessage.success('删除成功')
    fetchSkills()
  } catch (err) {
    if (err !== 'cancel') ElMessage.error('删除失败')
  }
}

const goToEditor = (id) => router.push(`/skills/${id}`)
const goToTranslate = (id) => router.push(`/skills/${id}/translate`)

onMounted(fetchSkills)

watch(currentPage, () => fetchSkills())
</script>

<style scoped>
.skill-list {
  height: 100vh;
  background: var(--color-bg);
}

.el-header {
  background: var(--color-bg);
  box-shadow: var(--shadow-border);
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

.el-main {
  padding: var(--spacing-6);
}

.skill-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--spacing-5);
}

.skill-card {
  cursor: pointer;
  transition: box-shadow var(--transition-normal);
  background: var(--color-bg);
  box-shadow: var(--shadow-card);
}

.skill-card:hover {
  box-shadow: var(--shadow-card-hover);
}

.skill-card.builtin {
  box-shadow: var(--shadow-card-hover);
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
}

.name {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.card-body .description {
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  margin-bottom: var(--spacing-3);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.meta {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.translated-badge {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  font-size: var(--font-size-xs);
  color: var(--color-success);
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.footer-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
}

.clone-btn {
  padding: 2px 4px;
  color: var(--color-text-muted);
}

.clone-btn:hover {
  color: var(--color-primary);
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: var(--spacing-6);
  padding-bottom: var(--spacing-4);
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

@media (max-width: 1024px) {
  .skill-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .left-group {
    flex-wrap: nowrap;
    min-width: 0;
  }

  .mobile-menu-btn {
    display: flex;
  }

  .btn-text {
    display: none;
  }

  .skill-grid {
    grid-template-columns: 1fr;
  }

  .el-main {
    padding: var(--spacing-3);
  }

  .header-content :deep(.el-form-item) {
    flex-direction: column;
  }
}
</style>
