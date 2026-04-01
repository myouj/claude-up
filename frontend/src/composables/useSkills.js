import api from './useApi.js'

export const skillsApi = {
  list: (params = {}) => api.get('/skills', { params }),
  get: (id) => api.get(`/skills/${id}`),
  create: (data) => api.post('/skills', data),
  update: (id, data) => api.put(`/skills/${id}`, data),
  delete: (id) => api.delete(`/skills/${id}`),
  clone: (id) => api.post(`/skills/${id}/clone`),
  export: () => api.get('/skills/export'),
  import: (data) => api.post('/skills/import', data)
}

export function useSkills() {
  return {
    skillsApi
  }
}
