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
  getAILogs(page = 1, status = '', search = '') {
    const params = new URLSearchParams({ page })
    if (status) params.set('status', status)
    if (search) params.set('search', search)
    return apiCall(`/admin/ai-logs?${params}`)
  },
  getSystemLogs(query = 'page=1') {
    return apiCall(`/admin/system-logs?${query}`)
  },
  updateUserRole(userId, role) {
    return apiCall(`/admin/users/${userId}/role`, { method: 'PATCH', body: { role } })
  },

  // Listmaks (all users)
  getAllListmaks(params = {}) {
    const qs = new URLSearchParams(params).toString()
    return apiCall(`/admin/listmaks?${qs}`)
  },
  updateListmakStatus(id, status) {
    return apiCall(`/listmaks/${id}`, { method: 'PUT', body: { status } })
  },
  deleteListmak(id) {
    return apiCall(`/listmaks/${id}`, { method: 'DELETE' })
  },
  getActiveShares(listmakId) {
    return apiCall(`/listmaks/${listmakId}/active-shares`)
  },
  deleteShareLink(id) {
    return apiCall(`/share-links/${id}`, { method: 'DELETE' })
  },
  deleteViewShare(id) {
    return apiCall(`/admin/view-shares/${id}`, { method: 'DELETE' })
  },
  getListmakSummary(listmakId) {
    return apiCall(`/listmaks/${listmakId}/summary`)
  },
  deleteSummary(listmakId) {
    return apiCall(`/admin/summaries/listmak/${listmakId}`, { method: 'DELETE' })
  },

  // Price Catalog
  getPriceCatalog() {
    return apiCall('/admin/price-catalog')
  },
  upsertPriceCatalog(entries) {
    return apiCall('/admin/price-catalog', { method: 'POST', body: entries })
  },
  deletePriceCatalog(id) {
    return apiCall(`/admin/price-catalog/${id}`, { method: 'DELETE' })
  }
}
