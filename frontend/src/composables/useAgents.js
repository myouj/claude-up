import api from './useApi.js'

export const agentsApi = {
  list: (params = {}) => api.get('/agents', { params }),
  get: (id) => api.get(`/agents/${id}`),
  create: (data) => api.post('/agents', data),
  update: (id, data) => api.put(`/agents/${id}`, data),
  delete: (id) => api.delete(`/agents/${id}`),
  clone: (id) => api.post(`/agents/${id}/clone`),
  export: () => api.get('/agents/export'),
  import: (data) => api.post('/agents/import', data)
}

export function useAgents() {
  return {
    agentsApi
  }
}
