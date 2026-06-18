// API Configuration using environment variables
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

/**
 * Redirect user to Google OAuth login
 */
export function loginWithGoogle() {
  window.location.href = `${API_BASE_URL}/auth/google/login`
}

/**
 * Verify current user authentication by calling the API
 * @returns {Promise<{authenticated: boolean, user: object|null}>}
 */
export async function verifyAuth() {
  try {
    const response = await fetch(`${API_BASE_URL}/auth/user`, {
      method: 'GET',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json'
      }
    })
    
    const data = await response.json()
    
    if (data.success && data.data) {
      // User is authenticated, save to localStorage
      saveUser(data.data)
      return { authenticated: true, user: data.data }
    }
    
    // Not authenticated, clear localStorage
    localStorage.removeItem('user')
    return { authenticated: false, user: null }
  } catch (error) {
    console.error('Auth verification error:', error)
    // On network error, check localStorage as fallback
    const user = getCurrentUser()
    return { authenticated: !!user, user }
  }
}

/**
 * Logout user and clear authentication
 */
export async function logout() {
  try {
    const response = await fetch(`${API_BASE_URL}/auth/logout`, {
      method: 'GET',
      credentials: 'include'
    })
    
    const data = await response.json()
    
    // Always clear local storage on logout attempt
    localStorage.removeItem('user')
    
    if (data.success) {
      return { success: true, message: data.message }
    }
    
    return { success: true, message: 'Logged out' }
  } catch (error) {
    // Clear local storage even on network error
    localStorage.removeItem('user')
    console.error('Logout error:', error)
    return { success: true, message: 'Logged out locally' }
  }
}

/**
 * Get current user from local storage
 */
export function getCurrentUser() {
  const user = localStorage.getItem('user')
  return user ? JSON.parse(user) : null
}

/**
 * Save user to local storage
 */
export function saveUser(user) {
  localStorage.setItem('user', JSON.stringify(user))
}

/**
 * Check if user is authenticated (based on localStorage)
 */
export function isAuthenticated() {
  return !!getCurrentUser()
}

/**
 * Get API base URL
 */
export function getApiBaseUrl() {
  return API_BASE_URL
}

export default {
  loginWithGoogle,
  verifyAuth,
  logout,
  getCurrentUser,
  saveUser,
  isAuthenticated,
  getApiBaseUrl
}
