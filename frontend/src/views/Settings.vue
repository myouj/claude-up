<template>
  <div class="settings-page">
    <BreadcrumbNav :items="[{ name: '设置' }]" />
    <el-header>
      <div class="header-content">
        <div class="left">
          <el-button class="back-btn" @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <h1>设置</h1>
        </div>
      </div>
    </el-header>

    <el-main>
      <div v-if="loading" class="loading-state">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>加载设置...</span>
      </div>

      <div v-else class="settings-content">
        <!-- API Keys Section -->
        <el-card class="settings-card">
          <template #header>
            <div class="card-header">
              <span>API Keys</span>
            </div>
          </template>

          <div class="api-keys-info">
            <el-icon><InfoFilled /></el-icon>
            <span>API Key 将加密存储在数据库中。环境变量优先级高于数据库存储。</span>
          </div>

          <el-form :model="apiSettings" label-position="top" class="settings-form">
            <el-form-item label="OpenAI API Key">
              <div class="key-input-row">
                <el-input
                  v-model="apiSettings.openai_api_key"
                  :type="showOpenAI ? 'text' : 'password'"
                  placeholder="sk-..."
                  clearable
                >
                  <template #suffix>
                    <el-button text @click="showOpenAI = !showOpenAI">
                      <el-icon>
                        <View v-if="!showOpenAI" />
                        <Hide v-else />
                      </el-icon>
                    </el-button>
                  </template>
                </el-input>
                <el-button type="primary" @click="saveSetting('openai_api_key')" :loading="savingKey === 'openai_api_key'">
                  保存
                </el-button>
              </div>
            </el-form-item>

            <el-form-item label="Anthropic (Claude) API Key">
              <div class="key-input-row">
                <el-input
                  v-model="apiSettings.anthropic_api_key"
                  :type="showClaude ? 'text' : 'password'"
                  placeholder="sk-ant-..."
                  clearable
                >
                  <template #suffix>
                    <el-button text @click="showClaude = !showClaude">
                      <el-icon>
                        <View v-if="!showClaude" />
                        <Hide v-else />
                      </el-icon>
                    </el-button>
                  </template>
                </el-input>
                <el-button type="primary" @click="saveSetting('anthropic_api_key')" :loading="savingKey === 'anthropic_api_key'">
                  保存
                </el-button>
              </div>
            </el-form-item>

            <el-form-item label="Google Gemini API Key">
              <div class="key-input-row">
                <el-input
                  v-model="apiSettings.gemini_api_key"
                  :type="showGemini ? 'text' : 'password'"
                  placeholder="AIza..."
                  clearable
                >
                  <template #suffix>
                    <el-button text @click="showGemini = !showGemini">
                      <el-icon>
                        <View v-if="!showGemini" />
                        <Hide v-else />
                      </el-icon>
                    </el-button>
                  </template>
                </el-input>
                <el-button type="primary" @click="saveSetting('gemini_api_key')" :loading="savingKey === 'gemini_api_key'">
                  保存
                </el-button>
              </div>
            </el-form-item>

            <el-form-item label="MiniMax API Key">
              <div class="key-input-row">
                <el-input
                  v-model="apiSettings.minimax_api_key"
                  :type="showMiniMax ? 'text' : 'password'"
                  placeholder="eyJ..."
                  clearable
                >
                  <template #suffix>
                    <el-button text @click="showMiniMax = !showMiniMax">
                      <el-icon>
                        <View v-if="!showMiniMax" />
                        <Hide v-else />
                      </el-icon>
                    </el-button>
                  </template>
                </el-input>
                <el-button type="primary" @click="saveSetting('minimax_api_key')" :loading="savingKey === 'minimax_api_key'">
                  保存
                </el-button>
              </div>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- Encryption Info -->
        <el-card class="settings-card">
          <template #header>
            <div class="card-header">
              <span>安全设置</span>
            </div>
          </template>
          <div class="encryption-info">
            <el-alert type="warning" :closable="false" show-icon>
              <template #title>
                <div>ENCRYPTION_KEY 环境变量</div>
              </template>
              <div class="alert-content">
                API Key 使用 AES-256-GCM 加密存储。需要设置 <code>ENCRYPTION_KEY</code> 环境变量（32 字节）以确保安全。
                <br />
                <br />
                示例：<code>ENCRYPTION_KEY=your-32-byte-secret-key-here</code>
                <br />
                如未设置，将使用默认密钥（仅用于开发环境）。
              </div>
            </el-alert>
          </div>
        </el-card>

        <!-- Other Settings -->
        <el-card class="settings-card">
          <template #header>
            <div class="card-header">
              <span>其他设置</span>
            </div>
          </template>

          <div v-if="otherSettings.length === 0" class="no-settings">
            暂无其他设置项
          </div>

          <el-form v-else :model="otherSettings" label-position="top" class="settings-form">
            <el-form-item
              v-for="setting in otherSettings"
              :key="setting.key"
              :label="setting.key"
            >
              <div class="key-input-row">
                <el-input
                  v-model="otherSettingValues[setting.key]"
                  :type="setting.is_secret && !visibleKeys.includes(setting.key) ? 'password' : 'text'"
                  clearable
                >
                  <template #suffix v-if="setting.is_secret">
                    <el-button text @click="toggleVisible(setting.key)">
                      <el-icon>
                        <View v-if="!visibleKeys.includes(setting.key)" />
                        <Hide v-else />
                      </el-icon>
                    </el-button>
                  </template>
                </el-input>
                <el-button type="primary" @click="saveOtherSetting(setting.key)">
                  保存
                </el-button>
              </div>
            </el-form-item>
          </el-form>
        </el-card>
      </div>
    </el-main>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import BreadcrumbNav from '../components/BreadcrumbNav.vue'

const router = useRouter()

const loading = ref(true)
const apiSettings = reactive({
  openai_api_key: '',
  anthropic_api_key: '',
  gemini_api_key: '',
  minimax_api_key: ''
})
const otherSettings = ref([])
const otherSettingValues = reactive({})
const visibleKeys = ref([])
const showOpenAI = ref(false)
const showClaude = ref(false)
const showGemini = ref(false)
const showMiniMax = ref(false)
const savingKey = ref('')

const apiKeyLabels = {
  openai_api_key: 'OpenAI API Key',
  anthropic_api_key: 'Anthropic API Key',
  gemini_api_key: 'Google Gemini API Key',
  minimax_api_key: 'MiniMax API Key'
}

const fetchSettings = async () => {
  loading.value = true
  try {
    const res = await axios.get('/api/settings')
    if (res.data.success) {
      const settings = res.data.data || []
      for (const s of settings) {
        if (s.key in apiKeyLabels) {
          apiSettings[s.key] = s.value && s.value !== '********' ? s.value : ''
        } else {
          otherSettings.value.push(s)
          otherSettingValues[s.key] = s.value
        }
      }
    }
  } catch (err) {
    ElMessage.error('加载设置失败')
  } finally {
    loading.value = false
  }
}

const saveSetting = async (key) => {
  savingKey.value = key
  try {
    const res = await axios.put(`/api/settings/${key}`, {
      value: apiSettings[key] || '',
      is_secret: true
    })
    if (res.data.success) {
      ElMessage.success(`${apiKeyLabels[key]} 已保存`)
      apiSettings[key] = ''
    }
  } catch (err) {
    ElMessage.error('保存失败')
  } finally {
    savingKey.value = ''
  }
}

const saveOtherSetting = async (key) => {
  try {
    const res = await axios.put(`/api/settings/${key}`, {
      value: otherSettingValues[key] || '',
      is_secret: otherSettings.value.find(s => s.key === key)?.is_secret || false
    })
    if (res.data.success) {
      ElMessage.success('设置已保存')
    }
  } catch (err) {
    ElMessage.error('保存失败')
  }
}

const toggleVisible = (key) => {
  const idx = visibleKeys.value.indexOf(key)
  if (idx >= 0) {
    visibleKeys.value.splice(idx, 1)
  } else {
    visibleKeys.value.push(key)
  }
}

const goBack = () => router.back()

onMounted(fetchSettings)
</script>

<style scoped>
.settings-page {
  height: 100vh;
  background: var(--color-bg);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.el-header {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  padding: 0 var(--spacing-6);
  height: 64px;
  flex-shrink: 0;
}

.header-content {
  width: 100%;
  display: flex;
  align-items: center;
}

.left {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.left h1 {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0;
}

.back-btn {
  padding: var(--spacing-2);
}

.el-main {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-5);
}

.loading-state {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-3);
  color: var(--color-text-muted);
}

.settings-content {
  max-width: 720px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-5);
}

.settings-card :deep(.el-card__header) {
  background: var(--color-bg);
  border-bottom: 1px solid var(--color-border);
  padding: var(--spacing-3) var(--spacing-4);
}

.card-header {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.api-keys-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  background: var(--color-info-light);
  padding: var(--spacing-3);
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-4);
}

.settings-form :deep(.el-form-item__label) {
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
}

.settings-form :deep(.el-form-item) {
  margin-bottom: var(--spacing-5);
}

.key-input-row {
  display: flex;
  gap: var(--spacing-3);
  align-items: center;
}

.key-input-row .el-input {
  flex: 1;
}

.encryption-info {
  margin-bottom: var(--spacing-3);
}

.alert-content {
  font-size: var(--font-size-sm);
  line-height: 1.7;
  margin-top: var(--spacing-2);
}

.alert-content code {
  background: rgba(0, 0, 0, 0.06);
  padding: 1px 6px;
  border-radius: var(--radius-sm);
  font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
  font-size: var(--font-size-xs);
}

.no-settings {
  color: var(--color-text-muted);
  font-size: var(--font-size-sm);
  text-align: center;
  padding: var(--spacing-4);
}

/* Responsive - Mobile */
@media (max-width: 768px) {
  .left h1 {
    font-size: var(--font-size-md);
  }

  .el-main {
    padding: var(--spacing-3);
  }

  .settings-content {
    gap: var(--spacing-3);
  }

  .key-input-row {
    flex-direction: column;
    align-items: stretch;
  }

  .key-input-row .el-input {
    width: 100%;
  }
}
</style>
