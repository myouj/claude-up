<template>
  <div class="prompt-tester">
    <BreadcrumbNav :items="[{ name: '提示词', path: '/prompts' }, { name: '测试' }]" />
    <el-header>
      <div class="header-content">
        <div class="left">
          <el-button class="mobile-menu-btn" @click="showSidebar = true">
            <el-icon><Menu /></el-icon>
          </el-button>
          <el-button class="back-btn" @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <h2 class="page-title">测试预览</h2>
        </div>
        <div class="right">
          <el-select v-model="selectedModel" placeholder="选择模型" class="model-select">
            <el-option label="MiniMax" value="MiniMax-M2.7">
              <div class="model-option">
                <span>MiniMax</span>
                <span class="model-desc">高性价比</span>
              </div>
            </el-option>
            <el-option label="阿里百炼 (Qwen)" value="qwen3.5-plus">
              <div class="model-option">
                <span>阿里百炼</span>
                <span class="model-desc">通义千问</span>
              </div>
            </el-option>
          </el-select>
          <el-button type="primary" @click="sendTest" :loading="loading" :disabled="!userMessage.trim()">
            <el-icon><Promotion /></el-icon>
            <span class="btn-text">发送</span>
          </el-button>
          <el-button @click="goToCompare">
            <el-icon><Connection /></el-icon>
            <span class="btn-text">对比测试</span>
          </el-button>
        </div>
      </div>
    </el-header>

    <el-container>
      <el-aside width="360px" class="sidebar">
        <div class="prompt-section">
          <h3>
            <el-icon><Document /></el-icon>
            当前提示词
          </h3>
          <el-input
            v-model="promptContent"
            type="textarea"
            :rows="8"
            placeholder="提示词内容..."
            class="prompt-input"
          />
        </div>

        <VariablePreviewer
          ref="variablePreviewerRef"
          :content="promptContent"
        />

        <el-divider />

        <div class="history-section">
          <h3>
            <el-icon><Clock /></el-icon>
            测试历史
          </h3>
          <div v-if="testHistory.length > 0" class="history-list">
            <div
              v-for="record in testHistory"
              :key="record.id"
              class="history-item"
              @click="loadRecord(record)"
            >
              <div class="history-header">
                <el-tag size="small" type="info">{{ record.model }}</el-tag>
                <span class="time">{{ formatTime(record.created_at) }}</span>
              </div>
              <p class="history-preview">{{ record.response?.substring(0, 80) }}...</p>
            </div>
          </div>
          <el-empty v-else description="暂无测试记录" :image-size="60" />
        </div>
      </el-aside>

      <!-- Mobile sidebar drawer -->
      <el-drawer v-model="showSidebar" title="提示词 & 历史" size="320px" direction="ltr">
        <div class="prompt-section">
          <h3>
            <el-icon><Document /></el-icon>
            当前提示词
          </h3>
          <el-input
            v-model="promptContent"
            type="textarea"
            :rows="6"
            placeholder="提示词内容..."
            class="prompt-input"
          />
        </div>
        <VariablePreviewer
          ref="variablePreviewerRef"
          :content="promptContent"
        />
        <el-divider />
        <div class="history-section">
          <h3>
            <el-icon><Clock /></el-icon>
            测试历史
          </h3>
          <div v-if="testHistory.length > 0" class="history-list">
            <div
              v-for="record in testHistory"
              :key="record.id"
              class="history-item"
              @click="() => { loadRecord(record); showSidebar = false }"
            >
              <div class="history-header">
                <el-tag size="small" type="info">{{ record.model }}</el-tag>
                <span class="time">{{ formatTime(record.created_at) }}</span>
              </div>
              <p class="history-preview">{{ record.response?.substring(0, 60) }}...</p>
            </div>
          </div>
          <el-empty v-else description="暂无测试记录" :image-size="60" />
        </div>
      </el-drawer>

      <el-main>
        <div class="chat-container">
          <div class="messages" ref="messagesRef">
            <div v-if="messages.length === 0" class="empty-chat">
              <div class="empty-icon">
                <svg width="64" height="64" viewBox="0 0 64 64" fill="none">
                  <circle cx="32" cy="32" r="28" stroke="var(--color-border)" stroke-width="2"/>
                  <path d="M20 26h24M20 32h16M20 38h20" stroke="var(--color-border)" stroke-width="2" stroke-linecap="round"/>
                </svg>
              </div>
              <p>输入消息开始测试</p>
              <span>Ctrl + Enter 快捷发送</span>
            </div>

            <div
              v-for="(msg, idx) in messages"
              :key="idx"
              :class="['message', msg.role]"
            >
              <div class="message-avatar">
                <el-icon v-if="msg.role === 'user'"><User /></el-icon>
                <el-icon v-else><ChatDotRound /></el-icon>
              </div>
              <div class="message-bubble">
                <pre>{{ msg.content }}</pre>
              </div>
            </div>
          </div>

          <div class="input-area">
            <el-input
              v-model="userMessage"
              type="textarea"
              :rows="3"
              placeholder="输入测试消息，按 Ctrl+Enter 发送..."
              @keyup.ctrl.enter="sendTest"
              class="message-input"
            />
            <div class="input-actions">
              <el-button @click="clearMessages" :disabled="messages.length === 0">
                清空对话
              </el-button>
              <el-button type="primary" @click="sendTest" :loading="loading" :disabled="!userMessage.trim()">
                发送
              </el-button>
            </div>
          </div>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { Menu } from '@element-plus/icons-vue'
import BreadcrumbNav from '../components/BreadcrumbNav.vue'
import VariablePreviewer from '../components/VariablePreviewer.vue'

const router = useRouter()
const route = useRoute()

const promptContent = ref('')
const selectedModel = ref('MiniMax-M2.7')
const userMessage = ref('')
const messages = ref([])
const loading = ref(false)
const testHistory = ref([])
const messagesRef = ref(null)
const showSidebar = ref(false)
const variablePreviewerRef = ref(null)

const fetchPrompt = async () => {
  try {
    const res = await axios.get(`/api/prompts/${route.params.id}`)
    if (res.data.success) {
      promptContent.value = res.data.data.content
    }
  } catch (err) {
    ElMessage.error('获取提示词失败')
  }
}

const fetchTestHistory = async () => {
  try {
    const res = await axios.get(`/api/prompts/${route.params.id}/tests`)
    if (res.data.success) {
      testHistory.value = res.data.data
    }
  } catch (err) {
    console.error('Failed to fetch test history:', err)
  }
}

const sendTest = async () => {
  if (!userMessage.value.trim()) return

  // 将用户消息添加到对话
  messages.value.push({
    role: 'user',
    content: userMessage.value
  })

  const userInput = userMessage.value
  userMessage.value = ''
  loading.value = true

  // 添加 AI 占位消息
  const aiMsgIndex = messages.value.length
  messages.value.push({
    role: 'assistant',
    content: '正在思考...'
  })

  try {
    // 替换变量
    const finalPrompt = variablePreviewerRef.value
      ? variablePreviewerRef.value.renderedContent.value
      : promptContent.value

    const fullPrompt = `${finalPrompt}\n\nUser: ${userInput}`
    const res = await axios.post(`/api/prompts/${route.params.id}/test`, {
      content: fullPrompt,
      model: selectedModel.value,
      messages: messages.value.slice(0, -1).map(m => ({
        role: m.role,
        content: m.content
      }))
    })

    if (res.data.success) {
      messages.value[aiMsgIndex].content = res.data.data.response
      fetchTestHistory()
    }
  } catch (err) {
    messages.value[aiMsgIndex].content = '请求失败，请重试'
    ElMessage.error('测试请求失败')
  } finally {
    loading.value = false
    await nextTick()
    scrollToBottom()
  }
}

const loadRecord = (record) => {
  messages.value = [
    { role: 'user', content: record.prompt_text },
    { role: 'assistant', content: record.response }
  ]
}

const clearMessages = () => {
  messages.value = []
}

const scrollToBottom = () => {
  if (messagesRef.value) {
    messagesRef.value.scrollTop = messagesRef.value.scrollHeight
  }
}

const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  return `${date.getMonth() + 1}/${date.getDate()} ${date.getHours()}:${String(date.getMinutes()).padStart(2, '0')}`
}

const goBack = () => router.back()
const goToCompare = () => router.push(`/prompts/${route.params.id}/test-compare`)

onMounted(() => {
  fetchPrompt()
  fetchTestHistory()
})
</script>

<style scoped>
.prompt-tester {
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

.back-btn {
  padding: var(--spacing-2);
}

.right {
  display: flex;
  gap: var(--spacing-3);
}

.model-select {
  width: 140px;
}

.model-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.model-desc {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.sidebar {
  background: var(--color-surface);
  padding: var(--spacing-5);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4);
  overflow: hidden;
}

.prompt-section h3,
.history-section h3 {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-secondary);
  margin-bottom: var(--spacing-3);
}

.prompt-input :deep(.el-textarea__inner) {
  font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
  font-size: var(--font-size-sm);
  line-height: 1.6;
}

.el-divider {
  margin: 0;
}

.history-section {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.history-list {
  flex: 1;
  overflow-y: auto;
  margin-top: var(--spacing-2);
}

.history-item {
  padding: var(--spacing-3);
  background: var(--color-bg);
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-2);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.history-item:hover {
  background: var(--color-primary-light);
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-2);
}

.time {
  color: var(--color-text-muted);
  font-size: var(--font-size-xs);
}

.history-preview {
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin: 0;
}

.el-main {
  padding: 0;
  background: var(--color-bg);
}

.chat-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.messages {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-5);
}

.empty-chat {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  text-align: center;
}

.empty-icon {
  margin-bottom: var(--spacing-4);
  opacity: 0.5;
}

.empty-chat p {
  font-size: var(--font-size-md);
  margin-bottom: var(--spacing-1);
}

.empty-chat span {
  font-size: var(--font-size-xs);
}

.message {
  display: flex;
  gap: var(--spacing-3);
  margin-bottom: var(--spacing-4);
  animation: fadeIn var(--transition-normal);
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.message.user {
  flex-direction: row-reverse;
}

.message-avatar {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  background: var(--color-primary);
  color: var(--color-surface);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.message.assistant .message-avatar {
  background: var(--color-success);
}

.message-bubble {
  max-width: 70%;
  padding: var(--spacing-3) var(--spacing-4);
  border-radius: var(--radius-lg);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
}

.message.user .message-bubble {
  background: var(--color-primary);
  border-color: var(--color-primary);
  color: var(--color-surface);
}

.message-bubble pre {
  margin: 0;
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  line-height: var(--line-height-relaxed);
  white-space: pre-wrap;
  word-break: break-word;
}

.input-area {
  padding: var(--spacing-4);
  background: var(--color-surface);
  border-top: 1px solid var(--color-border);
}

.message-input :deep(.el-textarea__inner) {
  border-radius: var(--radius-lg);
  padding: var(--spacing-3);
  font-size: var(--font-size-sm);
}

.input-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-2);
  margin-top: var(--spacing-3);
}

/* Responsive - Mobile */
@media (max-width: 768px) {
  .mobile-menu-btn {
    display: flex !important;
  }

  .sidebar {
    display: none;
  }

  .left {
    gap: var(--spacing-2);
  }

  .page-title {
    font-size: var(--font-size-md);
  }

  .right {
    gap: var(--spacing-1);
  }

  .btn-text {
    display: none;
  }

  .model-select {
    width: 100px;
  }

  .el-main {
    padding: var(--spacing-2);
  }

  .message-bubble {
    max-width: 85%;
  }
}
</style>
