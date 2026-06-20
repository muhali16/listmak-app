const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

async function apiCall(endpoint, options = {}) {
  const url = `${API_BASE_URL}${endpoint}`
  const config = {
    ...options,
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...options.headers }
  }
  if (config.body && typeof config.body === 'object') {
    config.body = JSON.stringify(config.body)
  }
  const response = await fetch(url, config)
  const data = await response.json()
  if (!response.ok) throw new Error(data.message || 'API request failed')
  return data
}

export default {
  getAILogs(page = 1) {
    return apiCall(`/admin/ai-logs?page=${page}`)
  },
  getSystemLogs(query = 'page=1') {
    return apiCall(`/admin/system-logs?${query}`)
  },
  updateUserRole(userId, role) {
    return apiCall(`/admin/users/${userId}/role`, {
      method: 'PATCH',
      body: { role }
    })
  }
}
