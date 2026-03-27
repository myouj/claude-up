<template>
  <div class="skill-list">
    <el-header>
      <div class="header-content">
        <div class="left">
          <el-button class="back-btn" @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <h1>Skills</h1>
        </div>
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          新建 Skill
        </el-button>
      </div>
    </el-header>

    <el-main>
      <div v-if="skills.length > 0" class="skill-grid">
        <el-card
          v-for="skill in skills"
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
              <el-button size="small" @click.stop="goToTranslate(skill.id)">
                <el-icon><Translate /></el-icon>
                翻译
              </el-button>
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
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'

const router = useRouter()
const skills = ref([])
const showCreateDialog = ref(false)

const newSkill = ref({
  name: '',
  description: '',
  content: '',
  category: ''
})

const fetchSkills = async () => {
  try {
    const res = await axios.get('/api/skills')
    if (res.data.success) {
      skills.value = res.data.data
    }
  } catch (err) {
    console.error('Failed to fetch skills:', err)
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

const goBack = () => router.push('/')
const goToEditor = (id) => router.push(`/skills/${id}`)
const goToTranslate = (id) => router.push(`/skills/${id}/translate`)

onMounted(fetchSkills)
</script>

<style scoped>
.skill-list {
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

.left h1 {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.back-btn {
  padding: var(--spacing-2);
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
  transition: all var(--transition-normal);
}

.skill-card.builtin {
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
}
</style>
