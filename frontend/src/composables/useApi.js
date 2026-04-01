import axios from 'axios'

// Axios instance with default config
const api = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// Response interceptor for consistent error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      // Server responded with error status
      const message = error.response.data?.error || `请求失败 (${error.response.status})`
      console.error(`API Error: ${message}`, error.response.config?.url)
    } else if (error.request) {
      // Request made but no response
      console.error('API Error: 网络请求无响应', error.config?.url)
    } else {
      console.error('API Error:', error.message)
    }
    return Promise.reject(error)
  }
)

export default api
