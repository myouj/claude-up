import { ref, computed } from 'vue'

/**
 * Task state management composable
 * Supports mock mode for development/testing
 *
 * States: idle | pending | running | done | failed | cancelled
 */

// Task states
export const TaskState = {
  IDLE: 'idle',
  PENDING: 'pending',
  RUNNING: 'running',
  DONE: 'done',
  FAILED: 'failed',
  CANCELLED: 'cancelled'
}

export function useTask(options = {}) {
  const { mockMode = true } = options

  // Task state
  const state = ref(TaskState.IDLE)
  const progress = ref(0)
  const error = ref(null)
  const result = ref(null)
  const taskId = ref(null)

  // Mock interval reference
  let mockInterval = null

  // Computed properties
  const isIdle = computed(() => state.value === TaskState.IDLE)
  const isPending = computed(() => state.value === TaskState.PENDING)
  const isRunning = computed(() => state.value === TaskState.RUNNING)
  const isDone = computed(() => state.value === TaskState.DONE)
  const isFailed = computed(() => state.value === TaskState.FAILED)
  const isCancelled = computed(() => state.value === TaskState.CANCELLED)
  const isFinished = computed(() => [TaskState.DONE, TaskState.FAILED, TaskState.CANCELLED].includes(state.value))

  // Reset task state
  const reset = () => {
    state.value = TaskState.IDLE
    progress.value = 0
    error.value = null
    result.value = null
    taskId.value = null
    if (mockInterval) {
      clearInterval(mockInterval)
      mockInterval = null
    }
  }

  /**
   * Start a new task
   * @param {Object} taskConfig
   * @param {string} taskConfig.id - Task identifier
   * @param {Function} taskConfig.handler - Actual task handler (non-mock mode)
   * @param {Object} taskConfig.mockConfig - Mock configuration
   * @param {number} taskConfig.mockConfig.duration - Total duration in ms (default: 5000)
   * @param {number} taskConfig.mockConfig.steps - Number of progress steps (default: 10)
   * @param {boolean} taskConfig.mockConfig.failOnLastStep - Fail at 90% progress (default: false)
   */
  const startTask = async (taskConfig = {}) => {
    const {
      id = `task_${Date.now()}`,
      handler = null,
      mockConfig = {}
    } = taskConfig

    const {
      duration = 5000,
      steps = 10,
      failOnLastStep = false
    } = mockConfig

    // Reset previous task state
    reset()

    taskId.value = id
    state.value = TaskState.PENDING
    progress.value = 0

    if (mockMode) {
      // Mock mode: simulate progress
      return new Promise((resolve, reject) => {
        state.value = TaskState.RUNNING
        const stepDuration = duration / steps
        let currentStep = 0

        mockInterval = setInterval(() => {
          currentStep++

          if (currentStep >= steps) {
            // Task complete
            clearInterval(mockInterval)
            mockInterval = null
            progress.value = 100
            state.value = TaskState.DONE
            result.value = {
              id: taskId.value,
              success: true,
              message: 'Task completed successfully',
              timestamp: new Date().toISOString()
            }
            resolve(result.value)
          } else if (failOnLastStep && currentStep === steps - 1) {
            // Simulate failure at 90%
            clearInterval(mockInterval)
            mockInterval = null
            progress.value = 90
            state.value = TaskState.FAILED
            error.value = 'Mock task failed at final step'
            result.value = {
              id: taskId.value,
              success: false,
              error: error.value,
              timestamp: new Date().toISOString()
            }
            reject(result.value)
          } else {
            // Update progress
            progress.value = Math.round((currentStep / steps) * 100)
          }
        }, stepDuration)
      })
    } else {
      // Real mode: use provided handler
      state.value = TaskState.RUNNING
      try {
        result.value = await handler()
        progress.value = 100
        state.value = TaskState.DONE
        return result.value
      } catch (err) {
        error.value = err.message || 'Task failed'
        state.value = TaskState.FAILED
        result.value = {
          id: taskId.value,
          success: false,
          error: error.value,
          timestamp: new Date().toISOString()
        }
        throw result.value
      }
    }
  }

  /**
   * Cancel running task
   */
  const cancelTask = () => {
    if (state.value === TaskState.RUNNING || state.value === TaskState.PENDING) {
      if (mockInterval) {
        clearInterval(mockInterval)
        mockInterval = null
      }
      state.value = TaskState.CANCELLED
      progress.value = 0
      result.value = {
        id: taskId.value,
        success: false,
        cancelled: true,
        timestamp: new Date().toISOString()
      }
    }
  }

  return {
    // State
    state,
    progress,
    error,
    result,
    taskId,
    // Computed
    isIdle,
    isPending,
    isRunning,
    isDone,
    isFailed,
    isCancelled,
    isFinished,
    // Methods
    startTask,
    cancelTask,
    reset,
    // Constants
    TaskState
  }
}

export default useTask
