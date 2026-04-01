<template>
  <el-dialog
    v-model="visible"
    title="邀请成员"
    width="480px"
    @close="handleClose"
  >
    <el-form :model="form" label-position="top">
      <el-form-item label="邮箱地址">
        <el-input
          v-model="form.email"
          placeholder="输入成员邮箱..."
          :prefix-icon="Message"
          clearable
        />
      </el-form-item>
      <el-form-item label="邀请角色">
        <el-select v-model="form.role" placeholder="选择角色" class="full-width">
          <el-option
            v-for="role in roleOptions"
            :key="role.value"
            :label="role.label"
            :value="role.value"
          >
            <div class="role-option">
              <span class="role-label">{{ role.label }}</span>
              <span class="role-desc">{{ role.description }}</span>
            </div>
          </el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="添加消息（可选）">
        <el-input
          v-model="form.message"
          type="textarea"
          :rows="3"
          placeholder="添加邀请附言..."
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" @click="handleInvite" :disabled="!form.email">
        <el-icon><Promotion /></el-icon>
        发送邀请
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useTeam, roleOptions } from '../composables/useTeam'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  teamId: {
    type: Number,
    default: null
  }
})

const emit = defineEmits(['update:modelValue', 'invited'])

const visible = ref(props.modelValue)
watch(() => props.modelValue, (val) => { visible.value = val })
watch(visible, (val) => { emit('update:modelValue', val) })

const form = ref({
  email: '',
  role: 'member',
  message: ''
})

const handleInvite = () => {
  if (!form.value.email) return
  ElMessage.success(`邀请已发送到 ${form.value.email}`)
  emit('invited', { ...form.value, teamId: props.teamId })
  handleClose()
}

const handleClose = () => {
  form.value = { email: '', role: 'member', message: '' }
  visible.value = false
}
</script>

<style scoped>
.full-width {
  width: 100%;
}

.role-option {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 2px 0;
}

.role-label {
  font-weight: var(--font-weight-medium);
  font-size: var(--font-size-sm);
}

.role-desc {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}
</style>
