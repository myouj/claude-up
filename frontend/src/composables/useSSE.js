import { ref, onUnmounted } from 'vue'

/**
 * SSE connection manager with auto-reconnect using exponential backoff
 * @param {Object} options
 * @param {string} options.url - SSE endpoint URL
 * @param {Function} options.onMessage - Callback for messages
 * @param {Function} options.onError - Callback for errors
 * @param {Function} options.onOpen - Callback when connection opens
 * @param {number} options.maxRetries - Max reconnection attempts (default: 5)
 * @param {number} options.baseDelay - Base delay in ms for backoff (default: 1000)
 */
export function useSSE(options = {}) {
  const {
    url = '',
    onMessage = () => {},
    onError = () => {},
    onOpen = () => {},
    maxRetries = 5,
    baseDelay = 1000
  } = options

  const connected = ref(false)
  const reconnectAttempt = ref(0)
  let eventSource = null
  let reconnectTimeout = null

  const calculateDelay = () => {
    // Exponential backoff: baseDelay * 2^attempt, max 30 seconds
    const delay = Math.min(baseDelay * Math.pow(2, reconnectAttempt.value), 30000)
    // Add jitter (±20%)
    const jitter = delay * 0.2 * (Math.random() - 0.5)
    return Math.round(delay + jitter)
  }

  const connect = () => {
    if (!url) {
      onError(new Error('SSE URL is required'))
      return
    }

    // Close existing connection
    disconnect()

    try {
      eventSource = new EventSource(url)

      eventSource.onopen = () => {
        connected.value = true
        reconnectAttempt.value = 0
        onOpen()
      }

      eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          onMessage(data)
        } catch {
          // Plain text message
          onMessage(event.data)
        }
      }

      eventSource.onerror = (error) => {
        connected.value = false
        onError(error)
        scheduleReconnect()
      }
    } catch (error) {
      onError(error)
      scheduleReconnect()
    }
  }

  const scheduleReconnect = () => {
    if (reconnectAttempt.value >= maxRetries) {
      console.warn(`SSE: Max retries (${maxRetries}) reached. Giving up.`)
      return
    }

    const delay = calculateDelay()
    reconnectAttempt.value++

    console.log(`SSE: Reconnecting in ${delay}ms (attempt ${reconnectAttempt.value}/${maxRetries})`)

    reconnectTimeout = setTimeout(() => {
      connect()
    }, delay)
  }

  const disconnect = () => {
    if (reconnectTimeout) {
      clearTimeout(reconnectTimeout)
      reconnectTimeout = null
    }

    if (eventSource) {
      eventSource.close()
      eventSource = null
    }

    connected.value = false
  }

  const send = (data) => {
    // SSE is receive-only; use fetch for sending data to server
    console.warn('SSE does not support sending data. Use fetch/XHR for server communication.')
  }

  // Cleanup on unmount
  onUnmounted(() => {
    disconnect()
  })

  return {
    connected,
    reconnectAttempt,
    connect,
    disconnect,
    send
  }
}

export default useSSE
