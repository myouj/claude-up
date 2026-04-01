<template>
  <div class="prompt-list">
    <el-container>
      <el-header>
        <div class="header-content">
          <div class="left-group">
            <el-button class="mobile-menu-btn" @click="showSidebar = true">
              <el-icon><Menu /></el-icon>
            </el-button>
            <div class="brand">
              <div class="logo">
                <svg width="28" height="28" viewBox="0 0 28 28" fill="none">
                  <rect width="28" height="28" rx="8" fill="var(--color-primary)"/>
                  <path d="M8 10h12M8 14h8M8 18h10" stroke="white" stroke-width="2" stroke-linecap="round"/>
                </svg>
              </div>
              <h1>PromptVault</h1>
            </div>
          </div>
          <div class="actions-group">
            <el-button type="primary" class="create-btn" @click="showCreateDialog = true">
              <el-icon><Plus /></el-icon>
              <span class="btn-text">新建提示词</span>
            </el-button>
            <el-dropdown trigger="click" @command="handleExport">
              <el-button>
                <el-icon><Download /></el-icon>
                <span class="btn-text">导出</span>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="json">导出为 JSON</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
            <el-button @click="showImportDialog = true">
              <el-icon><Upload /></el-icon>
              <span class="btn-text">导入</span>
            </el-button>
            <el-button @click="showTemplateLibrary = true">
              <el-icon><Collection /></el-icon>
              <span class="btn-text">模板库</span>
            </el-button>
          </div>
        </div>
      </el-header>

      <el-container>
        <!-- Desktop sidebar -->
        <el-aside width="240px" class="sidebar">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索提示词..."
            :prefix-icon="Search"
            clearable
            class="search-input"
          />

          <div class="nav-section">
            <el-menu :default-active="activeCategory" @select="handleCategorySelect" :ellipsis="false">
              <el-menu-item index="">
                <el-icon><Document /></el-icon>
                <span>全部提示词</span>
                <span class="count">{{ prompts.length }}</span>
              </el-menu-item>
              <el-menu-item index="favorite">
                <el-icon><Star /></el-icon>
                <span>收藏</span>
                <span class="count">{{ favoriteCount }}</span>
              </el-menu-item>
            </el-menu>
          </div>

          <div class="category-section">
            <h3>分类</h3>
            <div class="category-tags">
              <el-tag
                v-for="cat in categories"
                :key="cat"
                :type="activeCategory === cat ? 'primary' : 'info'"
                class="category-tag"
                :effect="activeCategory === cat ? 'dark' : 'light'"
                @click="handleCategorySelect(cat)"
              >
                {{ cat }}
              </el-tag>
              <el-tag
                v-if="!categories.includes('未分类')"
                type="info"
                effect="light"
                class="category-tag"
                @click="handleCategorySelect('未分类')"
              >
                未分类
              </el-tag>
            </div>
          </div>
        </el-aside>

        <!-- Mobile sidebar drawer -->
        <el-drawer v-model="showSidebar" title="筛选" size="280px" direction="ltr" class="mobile-sidebar-drawer">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索提示词..."
            :prefix-icon="Search"
            clearable
            class="search-input"
          />

          <div class="nav-section">
            <el-menu :default-active="activeCategory" @select="(key) => { handleCategorySelect(key); showSidebar = false }" :ellipsis="false">
              <el-menu-item index="">
                <el-icon><Document /></el-icon>
                <span>全部提示词</span>
                <span class="count">{{ prompts.length }}</span>
              </el-menu-item>
              <el-menu-item index="favorite">
                <el-icon><Star /></el-icon>
                <span>收藏</span>
                <span class="count">{{ favoriteCount }}</span>
              </el-menu-item>
            </el-menu>
          </div>

          <div class="category-section">
            <h3>分类</h3>
            <div class="category-tags">
              <el-tag
                v-for="cat in categories"
                :key="cat"
                :type="activeCategory === cat ? 'primary' : 'info'"
                class="category-tag"
                :effect="activeCategory === cat ? 'dark' : 'light'"
                @click="() => { handleCategorySelect(cat); showSidebar = false }"
              >
                {{ cat }}
              </el-tag>
              <el-tag
                v-if="!categories.includes('未分类')"
                type="info"
                effect="light"
                class="category-tag"
                @click="() => { handleCategorySelect('未分类'); showSidebar = false }"
              >
                未分类
              </el-tag>
            </div>
          </div>
        </el-drawer>

        <el-main>
          <div v-if="filteredPrompts.length > 0" class="prompt-grid">
            <el-card
              v-for="prompt in filteredPrompts"
              :key="prompt.id"
              class="prompt-card"
              :class="{ pinned: prompt.is_pinned }"
              shadow="hover"
              @click="goToEditor(prompt.id)"
            >
              <template #header>
                <div class="card-header">
                  <div class="title-row">
                    <el-icon v-if="prompt.is_pinned" class="pin-icon"><Pin /></el-icon>
                    <span class="title">{{ prompt.title }}</span>
                    <el-icon v-if="prompt.is_favorite" class="star-icon"><Star /></el-icon>
                  </div>
                  <div class="actions" @click.stop>
                    <el-dropdown trigger="click" :popper-options="{ strategy: 'fixed' }">
                      <button class="icon-btn">
                        <el-icon><MoreFilled /></el-icon>
                      </button>
                      <template #dropdown>
                        <el-dropdown-menu>
                          <el-dropdown-item @click="toggleFavorite(prompt)">
                            <el-icon><Star /></el-icon>
                            {{ prompt.is_favorite ? '取消收藏' : '收藏' }}
                          </el-dropdown-item>
                          <el-dropdown-item @click="togglePinned(prompt)">
                            <el-icon><Pushpin /></el-icon>
                            {{ prompt.is_pinned ? '取消置顶' : '置顶' }}
                          </el-dropdown-item>
                          <el-dropdown-item divided @click="goToVersions(prompt.id)">
                            <el-icon><Clock /></el-icon> 版本历史
                          </el-dropdown-item>
                          <el-dropdown-item @click="goToTest(prompt.id)">
                            <el-icon><ChatDotRound /></el-icon> 测试
                          </el-dropdown-item>
                          <el-dropdown-item @click="goToOptimize(prompt.id)">
                            <el-icon><MagicStick /></el-icon> AI 优化
                          </el-dropdown-item>
                          <el-dropdown-item divided @click="handleDelete(prompt)">
                            <el-icon><Delete /></el-icon> 删除
                          </el-dropdown-item>
                        </el-dropdown-menu>
                      </template>
                    </el-dropdown>
                  </div>
                </div>
              </template>

              <div class="card-body">
                <p class="description">{{ prompt.description || '暂无描述' }}</p>
                <div class="meta">
                  <el-tag v-if="prompt.category" size="small" type="info" effect="plain">
                    {{ prompt.category }}
                  </el-tag>
                  <el-tag
                    v-for="tag in prompt.tags.slice(0, 3)"
                    :key="tag"
                    size="small"
                    effect="plain"
                    class="tag"
                  >
                    {{ tag }}
                  </el-tag>
                  <span v-if="prompt.tags.length > 3" class="more-tags">+{{ prompt.tags.length - 3 }}</span>
                </div>
              </div>

              <template #footer>
                      <div class="card-footer">
                  <div class="footer-left">
                    <el-button
                      size="small"
                      text
                      @click.stop="handleClone(prompt)"
                      class="clone-btn"
                      title="克隆"
                    >
                      <el-icon><CopyDocument /></el-icon>
                    </el-button>
                    <span class="version-badge">
                      <el-icon><Clock /></el-icon>
                      v{{ prompt.version_count }}
                    </span>
                  </div>
                  <span class="date">{{ formatDate(prompt.updated_at) }}</span>
                </div>
              </template>
            </el-card>
          </div>

          <el-empty v-else description="暂无提示词" class="empty-state">
            <template #image>
              <svg width="120" height="120" viewBox="0 0 120 120" fill="none">
                <rect x="20" y="30" width="80" height="60" rx="8" stroke="var(--color-border)" stroke-width="2" fill="none"/>
                <path d="M35 50h50M35 60h30M35 70h40" stroke="var(--color-border)" stroke-width="2" stroke-linecap="round"/>
              </svg>
            </template>
            <el-button type="primary" @click="showCreateDialog = true">
              创建第一个提示词
            </el-button>
          </el-empty>

          <div v-if="totalPrompts > pageSize" class="pagination-wrapper">
            <el-pagination
              v-model:current-page="currentPage"
              :page-size="pageSize"
              :total="totalPrompts"
              layout="prev, pager, next"
              background
            />
          </div>
        </el-main>
      </el-container>
    </el-container>

    <!-- 创建对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      title="新建提示词"
      width="520px"
      :close-on-click-modal="false"
      class="create-dialog"
    >
      <el-form :model="newPrompt" label-position="top" class="create-form">
        <el-form-item label="标题" required>
          <el-input
            v-model="newPrompt.title"
            placeholder="输入提示词标题"
            maxlength="100"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="内容" required>
          <el-input
            v-model="newPrompt.content"
            type="textarea"
            :rows="6"
            placeholder="输入提示词内容..."
            maxlength="5000"
          />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="newPrompt.description"
            type="textarea"
            :rows="2"
            placeholder="简短描述这个提示词的用途"
            maxlength="500"
          />
        </el-form-item>
        <div class="form-row">
          <el-form-item label="分类" class="form-col">
            <el-select
              v-model="newPrompt.category"
              placeholder="选择分类"
              allow-create
              filterable
              clearable
              class="full-width"
            >
              <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
            </el-select>
          </el-form-item>
          <el-form-item label="标签" class="form-col">
            <el-select
              v-model="newPrompt.tags"
              multiple
              placeholder="添加标签"
              allow-create
              filterable
              clearable
              class="full-width"
            />
          </el-form-item>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 导入对话框 -->
    <el-dialog
      v-model="showImportDialog"
      title="导入提示词"
      width="560px"
      :close-on-click-modal="false"
    >
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
            :placeholder="importType === 'json' ? '粘贴 JSON 数据...' : '粘贴 Markdown 格式数据（## 标题\\n\\n内容\\n\\n---）'"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showImportDialog = false">取消</el-button>
        <el-button type="primary" @click="handleImport">导入</el-button>
      </template>
    </el-dialog>

    <!-- 模板库对话框 -->
    <el-drawer
      v-model="showTemplateLibrary"
      title="提示词模板库"
      size="600px"
      direction="rtl"
    >
      <div class="template-library">
        <p class="template-intro">选择一个模板快速创建提示词</p>

        <div class="template-grid">
          <div
            v-for="tpl in templates"
            :key="tpl.name"
            class="template-card"
            @click="useTemplate(tpl)"
          >
            <div class="tpl-header">
              <el-icon class="tpl-icon"><Document /></el-icon>
              <span class="tpl-name">{{ tpl.name }}</span>
            </div>
            <p class="tpl-desc">{{ tpl.description }}</p>
            <div class="tpl-tags">
              <el-tag
                v-for="tag in tpl.tags"
                :key="tag"
                size="small"
                type="info"
                effect="plain"
              >{{ tag }}</el-tag>
            </div>
          </div>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Collection, Menu } from '@element-plus/icons-vue'

const router = useRouter()
const prompts = ref([])
const searchKeyword = ref('')
const activeCategory = ref('')
const showCreateDialog = ref(false)
const showImportDialog = ref(false)
const showTemplateLibrary = ref(false)
const showSidebar = ref(false)
const currentPage = ref(1)
const pageSize = ref(12)
const totalPrompts = ref(0)
const importType = ref('json')
const importText = ref('')

const templates = [
  {
    name: '角色扮演',
    description: '定义 AI 扮演的角色和专业知识',
    tags: ['角色', '专业'],
    content: `## Role
You are a [role/expertise level] with extensive experience in [field].

## Context
[Provide relevant background information about the user's situation]

## Task
[Describe the specific task or question]

## Requirements
- [Requirement 1]
- [Requirement 2]

## Output Format
[Describe the expected response format]`
  },
  {
    name: '代码生成',
    description: '生成高质量、可运行的代码',
    tags: ['代码', '开发'],
    content: `## Task
Generate [language/framework] code that [description].

## Requirements
- Language: [language]
- Framework: [framework if applicable]
- Follow best practices: [specific guidelines]

## Input
\`\`\`
[paste your input here]
\`\`\`

## Output
Provide clean, well-commented code with:
- Clear function/variable names
- Error handling
- Type hints (if applicable)
- Usage examples`
  },
  {
    name: '代码审查',
    description: '分析代码并提供改进建议',
    tags: ['代码', '审查'],
    content: `## Role
You are an expert code reviewer with knowledge of:
- Software design patterns
- Security best practices
- Performance optimization
- Code readability

## Task
Review the following code and provide feedback:

\`\`\`
[code to review]
\`\`\`

## Review Criteria
Evaluate on:
1. Correctness and bugs
2. Security vulnerabilities
3. Performance issues
4. Code style and readability
5. Improvement suggestions

## Output Format
Provide a structured review with severity levels (Critical/High/Medium/Low) for each finding.`
  },
  {
    name: '文案写作',
    description: '创作吸引人的营销和商业文案',
    tags: ['文案', '营销'],
    content: `## Role
You are a professional copywriter specializing in [industry/type].

## Task
Write [type of content] for [product/service/campaign].

## Target Audience
[Describe the audience demographics, pain points, and motivations]

## Tone and Style
- Tone: [formal/casual/professional]
- Voice: [brand voice description]
- Length: [desired length]

## Key Points to Include
- [Point 1]
- [Point 2]
- [Point 3]

## Call to Action
[Desired action]

## Output
[Format specifications if any]`
  },
  {
    name: '数据解释',
    description: '分析和解释复杂的数据',
    tags: ['数据', '分析'],
    content: `## Task
Analyze the following data and provide insights:

**Data:**
\`\`\`
[data or description of data]
\`\`\`

## Analysis Goals
- [Goal 1]
- [Goal 2]

## Context
[What decisions will this analysis inform?]

## Output Format
Provide:
1. **Summary**: Key findings in 2-3 sentences
2. **Detailed Analysis**: Breakdown by [relevant dimensions]
3. **Insights**: Actionable observations
4. **Recommendations**: Next steps based on the data`
  },
  {
    name: '学习辅导',
    description: '以苏格拉底式提问引导学习',
    tags: ['教育', '学习'],
    content: `## Role
You are a patient, encouraging tutor who uses the Socratic method. You guide students to understanding through thoughtful questions rather than direct answers.

## Student Context
- Subject: [subject name]
- Current level: [beginner/intermediate/advanced]
- Topic: [specific topic]

## Guidelines
- Ask one question at a time
- Build on student's previous answers
- Use concrete examples to illustrate abstract concepts
- Encourage critical thinking
- Celebrate incremental progress

## Task
Help the student understand [topic/concept] by asking guiding questions.

## Approach
1. Assess current understanding with open questions
2. Identify misconceptions gently
3. Build toward the correct understanding
4. Connect to broader concepts

Begin by asking the student what they already know about [topic].`
  }
]

const useTemplate = (tpl) => {
  newPrompt.value = {
    title: tpl.name,
    content: tpl.content,
    description: tpl.description,
    category: tpl.tags[0] || '',
    tags: tpl.tags
  }
  showTemplateLibrary.value = false
  showCreateDialog.value = true
}

const newPrompt = ref({
  title: '',
  content: '',
  description: '',
  category: '',
  tags: []
})

const categories = computed(() => {
  const cats = new Set(prompts.value.map(p => p.category).filter(Boolean))
  return Array.from(cats).sort()
})

const favoriteCount = computed(() => prompts.value.filter(p => p.is_favorite).length)

const filteredPrompts = computed(() => prompts.value)

const fetchPrompts = async () => {
  try {
    const params = {
      page: currentPage.value,
      limit: pageSize.value
    }
    if (searchKeyword.value) params.search = searchKeyword.value
    if (activeCategory.value === 'favorite') params.favorite = 'true'
    else if (activeCategory.value) params.category = activeCategory.value

    const res = await axios.get('/api/prompts', { params })
    if (res.data.success) {
      prompts.value = res.data.data
      if (res.data.meta) {
        totalPrompts.value = res.data.meta.total
      }
    }
  } catch (err) {
    console.error('Failed to fetch prompts:', err)
  }
}

const handleCreate = async () => {
  if (!newPrompt.value.title || !newPrompt.value.content) {
    ElMessage.warning('请填写标题和内容')
    return
  }
  try {
    const res = await axios.post('/api/prompts', newPrompt.value)
    if (res.data.success) {
      ElMessage.success('创建成功')
      showCreateDialog.value = false
      newPrompt.value = { title: '', content: '', description: '', category: '', tags: [] }
      fetchPrompts()
    }
  } catch (err) {
    ElMessage.error('创建失败')
  }
}

const toggleFavorite = async (prompt) => {
  try {
    await axios.put(`/api/prompts/${prompt.id}`, { is_favorite: !prompt.is_favorite })
    prompt.is_favorite = !prompt.is_favorite
  } catch (err) {
    ElMessage.error('操作失败')
  }
}

const togglePinned = async (prompt) => {
  try {
    await axios.put(`/api/prompts/${prompt.id}`, { is_pinned: !prompt.is_pinned })
    prompt.is_pinned = !prompt.is_pinned
  } catch (err) {
    ElMessage.error('操作失败')
  }
}

const handleDelete = async (prompt) => {
  try {
    await ElMessageBox.confirm(`确定删除 "${prompt.title}" 吗？此操作不可恢复。`, '删除确认', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
      confirmButtonClass: 'el-button--danger'
    })
    await axios.delete(`/api/prompts/${prompt.id}`)
    ElMessage.success('删除成功')
    fetchPrompts()
  } catch (err) {
    if (err !== 'cancel') ElMessage.error('删除失败')
  }
}

const handleCategorySelect = (key) => {
  activeCategory.value = key
  currentPage.value = 1
}

const handleClone = async (prompt) => {
  try {
    const res = await axios.post(`/api/prompts/${prompt.id}/clone`)
    if (res.data.success) {
      ElMessage.success('克隆成功')
      fetchPrompts()
    }
  } catch (err) {
    ElMessage.error('克隆失败')
  }
}

const handleExport = async () => {
  try {
    const res = await axios.get('/api/prompts/export')
    if (res.data.success) {
      const content = JSON.stringify(res.data.data, null, 2)
      const blob = new Blob([content], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'prompts.json'
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
    axios.post('/api/prompts/import', { prompts: payload.prompts || payload })
      .then(res => {
        if (res.data.success) {
          ElMessage.success(`成功导入 ${res.data.imported} 条提示词`)
          showImportDialog.value = false
          importText.value = ''
          currentPage.value = 1
          fetchPrompts()
        }
      })
      .catch(() => ElMessage.error('导入失败'))
  } catch (err) {
    ElMessage.error('JSON 解析失败，请检查格式')
  }
}

const goToEditor = (id) => router.push(`/prompts/${id}`)
const goToVersions = (id) => router.push(`/prompts/${id}/versions`)
const goToTest = (id) => router.push(`/prompts/${id}/test`)
const goToOptimize = (id) => router.push(`/prompts/${id}/optimize`)

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return `${date.getMonth() + 1}/${date.getDate()} ${date.getHours()}:${String(date.getMinutes()).padStart(2, '0')}`
}

onMounted(fetchPrompts)

watch(currentPage, () => fetchPrompts())
watch(searchKeyword, (val) => {
  currentPage.value = 1
  fetchPrompts()
})
</script>

<style scoped>
.prompt-list {
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

.left-group {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.mobile-menu-btn {
  display: none;
  padding: var(--spacing-2);
}

.brand {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.logo {
  display: flex;
  align-items: center;
  justify-content: center;
}

.brand h1 {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.actions-group {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.create-btn {
  gap: var(--spacing-2);
}

.sidebar {
  background: var(--color-surface);
  padding: var(--spacing-4);
  border-right: 1px solid var(--color-border);
}

.search-input {
  margin-bottom: var(--spacing-4);
}

.nav-section {
  margin-bottom: var(--spacing-6);
}

.nav-section :deep(.el-menu) {
  border: none;
}

.nav-section :deep(.el-menu-item) {
  height: 40px;
  line-height: 40px;
  display: flex;
  align-items: center;
}

.nav-section :deep(.el-menu-item span) {
  flex: 1;
}

.count {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  background: var(--color-bg);
  padding: 2px 8px;
  border-radius: var(--radius-full);
}

.category-section h3 {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: var(--spacing-3);
}

.category-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-2);
}

.category-tag {
  cursor: pointer;
  transition: all var(--transition-fast);
}

.category-tag:hover {
  transform: translateY(-1px);
}

.el-main {
  padding: var(--spacing-6);
}

.prompt-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--spacing-5);
}

.prompt-card {
  cursor: pointer;
  transition: all var(--transition-normal);
  border: 1px solid var(--color-border);
}

.prompt-card:hover {
  transform: translateY(-2px);
  border-color: var(--color-border-hover);
}

.prompt-card.pinned {
  border-color: var(--color-primary);
  background: var(--color-primary-light);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--spacing-3);
}

.title-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  flex: 1;
  min-width: 0;
}

.pin-icon {
  color: var(--color-pin);
  flex-shrink: 0;
}

.title {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.star-icon {
  color: var(--color-star);
  flex-shrink: 0;
}

.actions {
  flex-shrink: 0;
}

.icon-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  border-radius: var(--radius-md);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
  transition: all var(--transition-fast);
}

.icon-btn:hover {
  background: var(--color-bg);
  color: var(--color-text-primary);
}

.card-body {
  padding: 0;
}

.description {
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  line-height: var(--line-height-relaxed);
  margin-bottom: var(--spacing-3);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--spacing-2);
}

.tag {
  font-size: var(--font-size-xs);
}

.more-tags {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: var(--spacing-3);
  border-top: 1px solid var(--color-border);
}

.version-badge {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-1);
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  background: var(--color-bg);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
}

.date {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.empty-state {
  padding: var(--spacing-12) 0;
}

.create-dialog :deep(.el-dialog__body) {
  padding: var(--spacing-6);
}

.create-form :deep(.el-form-item__label) {
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
}

.form-row {
  display: flex;
  gap: var(--spacing-4);
}

.form-col {
  flex: 1;
}

.full-width {
  width: 100%;
}

/* Dropdown menu items */
:deep(.el-dropdown-menu__item) {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-sm);
}

.footer-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
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

.template-library {
  padding: 0 var(--spacing-2);
}

.template-intro {
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  margin-bottom: var(--spacing-5);
}

.template-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-4);
}

.template-card {
  padding: var(--spacing-4);
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.template-card:hover {
  border-color: var(--color-primary);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.tpl-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-2);
}

.tpl-icon {
  font-size: 20px;
  color: var(--color-primary);
}

.tpl-name {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.tpl-desc {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  line-height: var(--line-height-relaxed);
  margin: 0 0 var(--spacing-3) 0;
  min-height: 36px;
}

.tpl-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-1);
}

/* Responsive - Tablet */
@media (max-width: 1024px) {
  .prompt-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

/* Responsive - Mobile */
@media (max-width: 768px) {
  .mobile-menu-btn {
    display: flex;
  }

  .sidebar {
    display: none;
  }

  .header-content {
    gap: var(--spacing-2);
  }

  .actions-group {
    gap: var(--spacing-1);
  }

  .btn-text {
    display: none;
  }

  .prompt-grid {
    grid-template-columns: 1fr;
  }

  .el-main {
    padding: var(--spacing-3);
  }

  .pagination-wrapper {
    margin-top: var(--spacing-4);
  }

  .template-grid {
    grid-template-columns: 1fr;
  }

  .form-row {
    flex-direction: column;
  }

  .form-col {
    width: 100%;
  }
}
</style>
