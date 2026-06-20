// API Base URL configuration
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

/**
 * Generic fetch wrapper with error handling and default options
 */
async function apiCall(endpoint, options = {}) {
  const url = `${API_BASE_URL}${endpoint}`
  
  const defaultHeaders = {
    'Content-Type': 'application/json'
  }

  const config = {
    ...options,
    credentials: 'include',
    headers: {
      ...defaultHeaders,
      ...options.headers
    }
  }

  if (config.body && typeof config.body === 'object') {
    config.body = JSON.stringify(config.body)
  }

  try {
    const response = await fetch(url, config)
    const data = await response.json()

    // 410 Gone for expired links handles as error but might need specific check in UI
    if (!response.ok) {
        // Return error object if available
        if (data && data.error) {
            const error = new Error(data.message || 'API request failed')
            error.code = data.error
            error.data = data.data
            throw error
        }
      throw new Error(data.message || 'API request failed')
    }

    return data
  } catch (error) {
    console.error(`API Error (${endpoint}):`, error)
    throw error
  }
}

export default {
  // --- SHARE LINKS (Input Links) ---
  
  createShareLink(data) {
    return apiCall('/share-links', {
      method: 'POST',
      body: data
    })
  },

  // Public access
  getShareLink(shareId) {
    return apiCall(`/share-links/${shareId}`, {
      method: 'GET'
    })
  },

  // Public access
  getShareOrders(shareId) {
    return apiCall(`/share-links/${shareId}/orders`, {
      method: 'GET'
    })
  },

  // Public access
  submitShareOrder(shareId, data) {
    return apiCall(`/share-links/${shareId}/orders`, {
      method: 'POST',
      body: data
    })
  },

  deleteShareLink(id) {
    return apiCall(`/share-links/${id}`, {
      method: 'DELETE'
    })
  },

  getFoodSuggestions(shareId, query = '') {
    const params = query ? `?q=${encodeURIComponent(query)}` : ''
    return apiCall(`/share-links/${shareId}/food-suggestions${params}`, { method: 'GET' })
  },

  getActiveShares(listmakId) {
    return apiCall(`/listmaks/${listmakId}/active-shares`)
  },

  // --- VIEW SHARES (Read-only Links) ---

  createViewShare(data) {
    return apiCall('/view-shares', {
      method: 'POST',
      body: data
    })
  },

  // Public access
  getViewShare(viewId) {
    return apiCall(`/view-shares/${viewId}`, {
      method: 'GET'
    })
  }
}
