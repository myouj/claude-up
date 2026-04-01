import api from './useApi.js'

export const promptsApi = {
  list: (params = {}) => api.get('/prompts', { params }),
  get: (id) => api.get(`/prompts/${id}`),
  create: (data) => api.post('/prompts', data),
  update: (id, data) => api.put(`/prompts/${id}`, data),
  delete: (id) => api.delete(`/prompts/${id}`),
  clone: (id) => api.post(`/prompts/${id}/clone`),
  test: (id, data) => api.post(`/prompts/${id}/test`, data),
  optimize: (id, data) => api.post(`/prompts/${id}/optimize`, data),
  versions: (id) => api.get(`/prompts/${id}/versions`),
  createVersion: (id, data) => api.post(`/prompts/${id}/versions`, data),
  tests: (id, params = {}) => api.get(`/prompts/${id}/tests`, { params }),
  export: () => api.get('/prompts/export'),
  import: (data) => api.post('/prompts/import', data)
}

export function usePrompts() {
  return {
    promptsApi
  }
}
