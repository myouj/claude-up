import api from './useApi.js'

export const settingsApi = {
  list: () => api.get('/settings'),
  get: (key) => api.get(`/settings/${key}`),
  update: (key, data) => api.put(`/settings/${key}`, data)
}

export const activityApi = {
  list: (params = {}) => api.get('/activity-logs', { params })
}

export const translateApi = {
  text: (data) => api.post('/translate', data),
  entity: (type, id, data) => api.post(`/translate/${type}/${id}`, data)
}

export const statsApi = {
  get: () => api.get('/stats')
}

export function useSettings() {
  return {
    settingsApi,
    activityApi,
    translateApi,
    statsApi
  }
}
