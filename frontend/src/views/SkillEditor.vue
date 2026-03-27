<template>
  <div class="skill-editor">
    <el-header>
      <div class="header-content">
        <div class="left">
          <el-button @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <h2>{{ skill.name || 'Skill 详情' }}</h2>
          <el-tag v-if="skill.source === 'builtin'" type="success" size="small">内置</el-tag>
        </div>
        <div class="right">
          <el-button @click="goToTranslate">
            <el-icon><Translate /></el-icon>
            翻译对比
          </el-button>
          <el-button type="primary" @click="handleSave" :disabled="skill.source === 'builtin'">
            <el-icon><Check /></el-icon>
            保存
          </el-button>
        </div>
      </div>
    </el-header>

    <el-container>
      <el-aside width="300px" class="sidebar">
        <el-form :model="skill" label-position="top">
          <el-form-item label="名称">
            <el-input v-model="skill.name" :disabled="skill.source === 'builtin'" />
          </el-form-item>
          <el-form-item label="描述">
            <el-input v-model="skill.description" type="textarea" :rows="3" />
          </el-form-item>
          <el-form-item label="分类">
            <el-select v-model="skill.category" :disabled="skill.source === 'builtin'" allow-create filterable clearable>
              <el-option label="git" value="git" />
              <el-option label="code" value="code" />
              <el-option label="docs" value="docs" />
              <el-option label="security" value="security" />
            </el-select>
          </el-form-item>
        </el-form>
      </el-aside>

      <el-main>
        <div class="editor-container">
          <div class="editor-header">
            <span>Skill 内容</span>
          </div>
          <el-input
            v-model="skill.content"
            type="textarea"
            class="content-editor"
            :disabled="skill.source === 'builtin'"
            placeholder="Skill 内容..."
          />
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from 'axios'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()

const skill = ref({
  id: null,
  name: '',
  description: '',
  content: '',
  content_cn: '',
  category: '',
  source: 'custom'
})

const fetchSkill = async () => {
  try {
    const res = await axios.get(`/api/skills/${route.params.id}`)
    if (res.data.success) {
      skill.value = res.data.data
    }
  } catch (err) {
    ElMessage.error('获取 Skill 失败')
  }
}

const handleSave = async () => {
  try {
    await axios.put(`/api/skills/${skill.value.id}`, {
      name: skill.value.name,
      description: skill.value.description,
      content: skill.value.content,
      category: skill.value.category
    })
    ElMessage.success('保存成功')
  } catch (err) {
    ElMessage.error('保存失败')
  }
}

const goBack = () => router.push('/skills')
const goToTranslate = () => router.push(`/skills/${route.params.id}/translate`)

onMounted(fetchSkill)
</script>

<style scoped>
.skill-editor {
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

.left h2 {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.right {
  display: flex;
  gap: var(--spacing-2);
}

.sidebar {
  background: var(--color-surface);
  padding: var(--spacing-5);
  border-right: 1px solid var(--color-border);
}

.el-main {
  padding: var(--spacing-5);
  background: var(--color-bg);
}

.editor-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--color-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
  overflow: hidden;
}

.editor-header {
  padding: var(--spacing-3) var(--spacing-4);
  background: var(--color-bg);
  border-bottom: 1px solid var(--color-border);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-secondary);
}

.content-editor {
  flex: 1;
}

.content-editor :deep(.el-textarea__inner) {
  height: 100%;
  padding: var(--spacing-4);
  font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
  font-size: var(--font-size-sm);
  line-height: 1.7;
  border: none;
  border-radius: 0;
  resize: none;
}
</style>
