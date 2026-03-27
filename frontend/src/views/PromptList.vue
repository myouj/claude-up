<template>
  <div class="prompt-list">
    <el-container>
      <el-header>
        <div class="header-content">
          <div class="brand">
            <div class="logo">
              <svg width="28" height="28" viewBox="0 0 28 28" fill="none">
                <rect width="28" height="28" rx="8" fill="var(--color-primary)"/>
                <path d="M8 10h12M8 14h8M8 18h10" stroke="white" stroke-width="2" stroke-linecap="round"/>
              </svg>
            </div>
            <h1>PromptVault</h1>
          </div>
          <el-button type="primary" class="create-btn" @click="showCreateDialog = true">
            <el-icon><Plus /></el-icon>
            新建提示词
          </el-button>
        </div>
      </el-header>

      <el-container>
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
                  <span class="version-badge">
                    <el-icon><Clock /></el-icon>
                    v{{ prompt.version_count }}
                  </span>
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
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'

const router = useRouter()
const prompts = ref([])
const searchKeyword = ref('')
const activeCategory = ref('')
const showCreateDialog = ref(false)

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

const filteredPrompts = computed(() => {
  let result = [...prompts.value]

  if (searchKeyword.value) {
    const kw = searchKeyword.value.toLowerCase()
    result = result.filter(p =>
      p.title.toLowerCase().includes(kw) ||
      p.content.toLowerCase().includes(kw) ||
      p.description?.toLowerCase().includes(kw)
    )
  }

  if (activeCategory.value === 'favorite') {
    result = result.filter(p => p.is_favorite)
  } else if (activeCategory.value) {
    result = result.filter(p => p.category === activeCategory.value)
  }

  // 置顶的排在前面
  return result.sort((a, b) => (b.is_pinned ? 1 : 0) - (a.is_pinned ? 1 : 0))
})

const fetchPrompts = async () => {
  try {
    const res = await axios.get('/api/prompts')
    if (res.data.success) {
      prompts.value = res.data.data
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
</style>
