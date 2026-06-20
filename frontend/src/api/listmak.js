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
    credentials: 'include', // Important for cookies/session
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

    if (!response.ok) {
      throw new Error(data.message || 'API request failed')
    }

    return data
  } catch (error) {
    console.error(`API Error (${endpoint}):`, error)
    throw error
  }
}

export default {
  // --- LISTMAKS ---
  
  getAllListMaks(params = {}) {
    const query = new URLSearchParams(params).toString()
    return apiCall(`/listmaks?${query}`)
  },

  getListMakById(id) {
    return apiCall(`/listmaks/${id}`)
  },

  getListMakByDate(date) {
    return apiCall(`/listmaks/date/${date}`)
  },

  createListMak(data) {
    return apiCall('/listmaks', {
      method: 'POST',
      body: data
    })
  },

  updateListMak(id, data) {
    return apiCall(`/listmaks/${id}`, {
      method: 'PUT',
      body: data
    })
  },

  deleteListMak(id) {
    return apiCall(`/listmaks/${id}`, {
      method: 'DELETE'
    })
  },

  // --- ORDERS ---

  getOrders(listmakId, params = {}) {
    const query = new URLSearchParams(params).toString()
    return apiCall(`/listmaks/${listmakId}/orders?${query}`)
  },

  addOrder(listmakId, data) {
    return apiCall(`/listmaks/${listmakId}/orders`, {
      method: 'POST',
      body: data
    })
  },

  addBulkOrders(listmakId, orders) {
    return apiCall(`/listmaks/${listmakId}/orders/bulk`, {
      method: 'POST',
      body: orders
    })
  },

  updateOrder(id, data) {
    return apiCall(`/orders/${id}`, {
      method: 'PUT',
      body: data
    })
  },

  updateOrderPaidStatus(id, isPaid) {
    return apiCall(`/orders/${id}/paid`, {
      method: 'PATCH',
      body: { is_paid: isPaid }
    })
  },

  deleteOrder(id) {
    return apiCall(`/orders/${id}`, {
      method: 'DELETE'
    })
  },

  updateOrdersPaidByName(listmakId, name, isPaid) {
    return apiCall(`/listmaks/${listmakId}/orders/paid`, {
      method: 'PATCH',
      body: { name, is_paid: isPaid }
    })
  }
}
