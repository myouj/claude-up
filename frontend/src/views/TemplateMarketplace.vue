<template>
  <div class="template-marketplace">
    <el-header>
      <div class="header-content">
        <div class="left-group">
          <el-button class="mobile-menu-btn" @click="showSidebar = true">
            <el-icon><Menu /></el-icon>
          </el-button>
          <div class="brand">
            <el-button class="back-btn" @click="goBack">
              <el-icon><ArrowLeft /></el-icon>
            </el-button>
            <h1>模板市场</h1>
          </div>
        </div>
        <div class="actions-group">
          <el-button @click="handleExport">
            <el-icon><Upload /></el-icon>
            <span class="btn-text">导出我的模板</span>
          </el-button>
          <el-button type="primary" @click="showCreateDialog = true">
            <el-icon><Plus /></el-icon>
            <span class="btn-text">发布模板</span>
          </el-button>
        </div>
      </div>
    </el-header>

    <el-main>
      <!-- Search & Sort Bar -->
      <div class="search-bar">
        <el-input
          v-model="searchQuery"
          placeholder="搜索模板..."
          :prefix-icon="Search"
          clearable
          class="search-input"
        />
        <el-select v-model="sortBy" class="sort-select">
          <el-option label="热门" value="popular" />
          <el-option label="最新" value="recent" />
          <el-option label="评分最高" value="rating" />
        </el-select>
      </div>

      <!-- Category Tabs -->
      <div class="category-tabs">
        <div class="tabs-scroll">
          <el-button
            v-for="cat in categories"
            :key="cat.value"
            :type="activeCategory === cat.value ? 'primary' : 'default'"
            :icon="cat.icon"
            size="small"
            class="category-btn"
            @click="activeCategory = cat.value"
          >
            {{ cat.label }}
          </el-button>
        </div>
      </div>

      <!-- Results Count -->
      <div class="results-info">
        <span>{{ filteredTemplates.length }} 个模板</span>
        <span v-if="installedCount > 0" class="installed-count">
          · {{ installedCount }} 个已安装
        </span>
      </div>

      <!-- Template Grid -->
      <div v-if="filteredTemplates.length > 0" class="template-grid">
        <TemplateCard
          v-for="template in filteredTemplates"
          :key="template.id"
          :template="template"
          :is-installed="isInstalled(template.id)"
          @click="goToDetail(template.id)"
        />
      </div>

      <el-empty v-else description="未找到匹配的模板">
        <el-button type="primary" @click="searchQuery = ''; activeCategory = 'all'">
          重置筛选
        </el-button>
      </el-empty>
    </el-main>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreateDialog" title="发布模板" width="560px">
      <el-form :model="newTemplate" label-position="top">
        <el-form-item label="模板名称" required>
          <el-input v-model="newTemplate.name" placeholder="例如：代码审查专家" />
        </el-form-item>
        <el-form-item label="分类" required>
          <el-select v-model="newTemplate.category" class="full-width">
            <el-option
              v-for="cat in categories.filter(c => c.value !== 'all')"
              :key="cat.value"
              :label="cat.label"
              :value="cat.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="newTemplate.description" type="textarea" :rows="2" placeholder="简短描述模板用途..." />
        </el-form-item>
        <el-form-item label="模板内容" required>
          <el-input v-model="newTemplate.content" type="textarea" :rows="6" placeholder="输入模板内容..." />
        </el-form-item>
        <el-form-item label="标签">
          <el-select v-model="newTemplate.tags" multiple allow-create filterable placeholder="添加标签" class="full-width">
            <el-option label="code-review" value="code-review" />
            <el-option label="security" value="security" />
            <el-option label="generator" value="generator" />
            <el-option label="documentation" value="documentation" />
            <el-option label="debug" value="debug" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">发布</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Menu } from '@element-plus/icons-vue'
import {
  mockTemplates,
  templateCategories,
  installedTemplates
} from '../composables/useTemplate'
import TemplateCard from '../components/TemplateCard.vue'

const router = useRouter()

const templates = ref([])
const searchQuery = ref('')
const sortBy = ref('popular')
const activeCategory = ref('all')
const showCreateDialog = ref(false)
const showSidebar = ref(false)
const categories = ref([])

const newTemplate = ref({
  name: '',
  category: '',
  description: '',
  content: '',
  tags: []
})

onMounted(() => {
  templates.value = mockTemplates.value
  categories.value = templateCategories.value
})

const installedCount = computed(() => {
  return templates.value.filter(t => installedTemplates.value.has(t.id)).length
})

const filteredTemplates = computed(() => {
  let result = templates.value

  if (activeCategory.value !== 'all') {
    result = result.filter(t => t.category === activeCategory.value)
  }

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(t =>
      t.name.toLowerCase().includes(q) ||
      t.description.toLowerCase().includes(q) ||
      t.tags.some(tag => tag.toLowerCase().includes(q))
    )
  }

  switch (sortBy.value) {
    case 'popular':
      return [...result].sort((a, b) => b.installs - a.installs)
    case 'recent':
      return [...result].sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
    case 'rating':
      return [...result].sort((a, b) => b.score - a.score)
    default:
      return result
  }
})

const isInstalled = (id) => installedTemplates.value.has(id)

const goBack = () => router.push('/')
const goToDetail = (id) => router.push(`/templates/${id}`)

const handleCreate = () => {
  if (!newTemplate.value.name || !newTemplate.value.content) {
    ElMessage.warning('请填写名称和内容')
    return
  }
  ElMessage.success('模板已发布（mock 模式）')
  showCreateDialog.value = false
}

const handleExport = () => {
  ElMessage.info('导出功能待后端 API 支持后实现')
}
</script>

<style scoped>
.template-marketplace {
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

.brand {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.back-btn {
  padding: var(--spacing-2);
}

.brand h1 {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0;
}

.mobile-menu-btn {
  display: none;
  padding: var(--spacing-2);
}

.el-main {
  padding: var(--spacing-5);
}

.search-bar {
  display: flex;
  gap: var(--spacing-3);
  margin-bottom: var(--spacing-4);
}

.search-input {
  flex: 1;
  max-width: 400px;
}

.sort-select {
  width: 140px;
  flex-shrink: 0;
}

.category-tabs {
  margin-bottom: var(--spacing-4);
  overflow: hidden;
}

.tabs-scroll {
  display: flex;
  gap: var(--spacing-2);
  overflow-x: auto;
  padding-bottom: var(--spacing-2);
  scrollbar-width: none;
}

.tabs-scroll::-webkit-scrollbar {
  display: none;
}

.category-btn {
  flex-shrink: 0;
}

.results-info {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
  margin-bottom: var(--spacing-4);
}

.installed-count {
  color: var(--color-success);
}

.template-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--spacing-4);
}

.full-width {
  width: 100%;
}

@media (max-width: 768px) {
  .template-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--spacing-3);
  }

  .search-bar {
    flex-direction: column;
  }

  .search-input {
    max-width: 100%;
  }

  .sort-select {
    width: 100%;
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

@media (max-width: 480px) {
  .template-grid {
    grid-template-columns: 1fr;
  }
}
</style>
