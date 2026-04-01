<template>
  <div class="template-detail">
    <BreadcrumbNav :items="[{ name: '模板市场', path: '/templates' }, { name: template?.name || '模板详情' }]" />
    <el-header>
      <div class="header-content">
        <div class="left">
          <el-button class="back-btn" @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <h2>{{ template?.name }}</h2>
          <el-tag v-if="isInstalled" type="success" size="small">已安装</el-tag>
        </div>
        <div class="right">
          <el-button @click="handleInstall" :type="isInstalled ? 'default' : 'primary'">
            <el-icon>
              <component :is="isInstalled ? 'CircleCheck' : 'Download'" />
            </el-icon>
            {{ isInstalled ? '已安装' : '安装' }}
          </el-button>
        </div>
      </div>
    </el-header>

    <el-main v-if="template">
      <div class="detail-layout">
        <!-- Main Content -->
        <div class="main-content">
          <!-- Stats Bar -->
          <div class="stats-bar">
            <div class="stat">
              <el-rate :model-value="template.score" disabled show-score />
              <span class="stat-label">{{ template.score.toFixed(1) }}</span>
            </div>
            <div class="stat">
              <el-icon><Download /></el-icon>
              <span class="stat-value">{{ template.installs }}</span>
              <span class="stat-label">安装</span>
            </div>
            <div class="stat">
              <el-icon><ChatDotRound /></el-icon>
              <span class="stat-value">{{ template.comments.length }}</span>
              <span class="stat-label">评论</span>
            </div>
            <div class="stat">
              <el-icon><Timer /></el-icon>
              <span class="stat-label">{{ formatDate(template.created_at) }}</span>
            </div>
          </div>

          <!-- Description -->
          <el-card class="section-card">
            <template #header>
              <div class="card-title">
                <el-icon><InfoFilled /></el-icon>
                <span>描述</span>
              </div>
            </template>
            <p class="description">{{ template.description }}</p>
            <div class="tags">
              <el-tag v-for="tag in template.tags" :key="tag" size="small" effect="plain">
                {{ tag }}
              </el-tag>
            </div>
          </el-card>

          <!-- Preview -->
          <el-card class="section-card">
            <template #header>
              <div class="card-title">
                <el-icon><View /></el-icon>
                <span>模板内容</span>
                <el-button size="small" text @click="copyContent">
                  <el-icon><CopyDocument /></el-icon>
                  复制
                </el-button>
              </div>
            </template>
            <pre class="content-preview">{{ template.content }}</pre>
          </el-card>

          <!-- Comments -->
          <el-card class="section-card">
            <template #header>
              <div class="card-title">
                <el-icon><ChatDotRound /></el-icon>
                <span>用户评论 ({{ template.comments.length }})</span>
              </div>
            </template>

            <!-- Add Comment -->
            <div class="add-comment">
              <el-input
                v-model="newComment"
                type="textarea"
                :rows="2"
                placeholder="分享你的使用体验..."
              />
              <div class="comment-actions">
                <el-rate v-model="newCommentScore" />
                <el-button type="primary" size="small" @click="submitComment" :disabled="!newComment.trim()">
                  发表评论
                </el-button>
              </div>
            </div>

            <!-- Comments List -->
            <div v-if="template.comments.length > 0" class="comments-list">
              <div
                v-for="comment in template.comments"
                :key="comment.id"
                class="comment-item"
              >
                <div class="comment-header">
                  <el-avatar :size="32" class="comment-avatar">
                    {{ comment.author.avatar }}
                  </el-avatar>
                  <div class="comment-info">
                    <span class="comment-author">{{ comment.author.name }}</span>
                    <el-rate :model-value="comment.score" disabled size="small" />
                  </div>
                  <span class="comment-date">{{ formatDate(comment.created_at) }}</span>
                </div>
                <p class="comment-content">{{ comment.content }}</p>
              </div>
            </div>

            <el-empty v-else description="暂无评论，成为第一个评论者！" />
          </el-card>
        </div>

        <!-- Sidebar -->
        <div class="sidebar">
          <!-- Author Card -->
          <el-card class="author-card">
            <div class="author-info">
              <el-avatar :size="48" class="author-avatar">
                {{ template.author.avatar }}
              </el-avatar>
              <div class="author-text">
                <span class="author-name">{{ template.author.name }}</span>
                <span class="author-label">作者</span>
              </div>
            </div>
          </el-card>

          <!-- Related Templates -->
          <el-card class="related-card">
            <template #header>
              <div class="card-title">
                <el-icon><Collection /></el-icon>
                <span>相关模板</span>
              </div>
            </template>
            <div class="related-list">
              <div
                v-for="related in relatedTemplates"
                :key="related.id"
                class="related-item"
                @click="goToDetail(related.id)"
              >
                <div class="related-name">{{ related.name }}</div>
                <div class="related-meta">
                  <el-icon><Star /></el-icon>
                  {{ related.score.toFixed(1) }}
                </div>
              </div>
            </div>
          </el-card>
        </div>
      </div>
    </el-main>

    <el-main v-else>
      <el-empty description="模板不存在" />
    </el-main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  mockTemplates,
  installedTemplates
} from '../composables/useTemplate'
import BreadcrumbNav from '../components/BreadcrumbNav.vue'

const router = useRouter()
const route = useRoute()

const template = ref(null)
const newComment = ref('')
const newCommentScore = ref(5)

const isInstalled = computed(() =>
  template.value && installedTemplates.value.has(template.value.id)
)

const relatedTemplates = computed(() => {
  if (!template.value) return []
  return mockTemplates.value
    .filter(t => t.id !== template.value.id && t.category === template.value.category)
    .slice(0, 3)
})

const fetchTemplate = () => {
  const id = parseInt(route.params.id)
  template.value = mockTemplates.value.find(t => t.id === id) || null
}

const goBack = () => router.push('/templates')
const goToDetail = (id) => router.push(`/templates/${id}`)

const handleInstall = () => {
  if (!template.value) return
  if (isInstalled.value) {
    installedTemplates.value.delete(template.value.id)
    ElMessage.success('已卸载模板')
  } else {
    installedTemplates.value.add(template.value.id)
    ElMessage.success('模板安装成功！')
  }
}

const copyContent = () => {
  if (template.value) {
    navigator.clipboard.writeText(template.value.content)
    ElMessage.success('内容已复制到剪贴板')
  }
}

const submitComment = () => {
  if (!newComment.value.trim() || !template.value) return
  template.value.comments.push({
    id: Date.now(),
    author: { name: '当前用户', avatar: '我' },
    content: newComment.value,
    score: newCommentScore.value,
    created_at: new Date().toISOString()
  })
  newComment.value = ''
  newCommentScore.value = 5
  ElMessage.success('评论已发表')
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return `${date.getFullYear()}/${date.getMonth() + 1}/${date.getDate()}`
}

onMounted(fetchTemplate)
</script>

<style scoped>
.template-detail {
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

.back-btn {
  padding: var(--spacing-2);
}

.left h2 {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0;
}

.right {
  display: flex;
  gap: var(--spacing-2);
  flex-shrink: 0;
}

.el-main {
  padding: var(--spacing-6);
}

.detail-layout {
  display: grid;
  grid-template-columns: 1fr 300px;
  gap: var(--spacing-6);
  align-items: start;
}

.main-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-5);
}

.stats-bar {
  display: flex;
  gap: var(--spacing-6);
  padding: var(--spacing-4);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.stat {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.stat-label {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
}

.stat-value {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.section-card {
  border-radius: var(--radius-lg);
}

.card-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.card-title .el-button {
  margin-left: auto;
}

.description {
  margin: 0 0 var(--spacing-4);
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  line-height: 1.6;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-2);
}

.content-preview {
  margin: 0;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: var(--font-size-sm);
  line-height: 1.7;
  color: var(--color-text-primary);
  background: var(--color-bg);
  border-radius: var(--radius-md);
  padding: var(--spacing-4);
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 400px;
  overflow-y: auto;
}

/* Comments */
.add-comment {
  margin-bottom: var(--spacing-5);
  padding-bottom: var(--spacing-4);
  border-bottom: 1px solid var(--color-border);
}

.comment-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: var(--spacing-2);
}

.comments-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4);
}

.comment-item {
  padding-bottom: var(--spacing-4);
  border-bottom: 1px solid var(--color-border);
}

.comment-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.comment-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  margin-bottom: var(--spacing-2);
}

.comment-avatar {
  background: var(--color-primary-light);
  color: var(--color-primary);
  font-weight: var(--font-weight-semibold);
}

.comment-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
}

.comment-author {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.comment-date {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.comment-content {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  line-height: 1.6;
  padding-left: var(--spacing-10);
}

/* Sidebar */
.sidebar {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4);
  position: sticky;
  top: var(--spacing-6);
}

.author-card {
  border-radius: var(--radius-lg);
}

.author-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.author-avatar {
  background: var(--color-warning-light);
  color: var(--color-warning);
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-md);
}

.author-text {
  display: flex;
  flex-direction: column;
}

.author-name {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.author-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.related-card {
  border-radius: var(--radius-lg);
}

.related-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.related-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-2) var(--spacing-3);
  background: var(--color-bg);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.related-item:hover {
  background: var(--color-primary-light);
}

.related-name {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
}

.related-meta {
  display: flex;
  align-items: center;
  gap: 2px;
  font-size: var(--font-size-xs);
  color: var(--color-warning);
}

@media (max-width: 1024px) {
  .detail-layout {
    grid-template-columns: 1fr;
  }

  .sidebar {
    position: static;
    flex-direction: row;
    flex-wrap: wrap;
  }

  .author-card,
  .related-card {
    flex: 1;
    min-width: 240px;
  }
}

@media (max-width: 768px) {
  .stats-bar {
    flex-wrap: wrap;
    gap: var(--spacing-3);
  }

  .comment-content {
    padding-left: 0;
  }

  .el-main {
    padding: var(--spacing-3);
  }
}
</style>
