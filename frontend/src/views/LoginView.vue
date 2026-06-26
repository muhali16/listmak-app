<template>
  <div class="login-container">
    <div class="login-card">
      <!-- App Logo -->
      <div class="login-header">
        <div class="logo-wrapper">
          <img src="../assets/logo/webp/listmak-icon-128.webp" alt="Listmak" class="logo-icon-img" />
        </div>
        <h1 class="app-title">Listmak</h1>
        <p class="app-subtitle">Kelola pesanan makanan kantor dengan mudah</p>
      </div>

      <!-- Google Login Button -->
      <div class="login-form">
        <Button 
          @click="handleGoogleLogin"
          class="google-btn"
          :loading="isLoading"
        >
          <template #icon>
            <svg class="google-icon" viewBox="0 0 24 24" width="20" height="20">
              <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
              <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
              <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
              <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
            </svg>
          </template>
          <span class="btn-text">Masuk dengan Google</span>
        </Button>

        <p class="login-terms">
          Dengan masuk, Anda menyetujui
          <router-link to="/terms" class="login-link">Syarat &amp; Ketentuan</router-link>
          dan
          <router-link to="/privacy" class="login-link">Kebijakan Privasi</router-link>
          kami.
        </p>
      </div>

      <!-- Features Preview -->
      <div class="features-preview">
        <div class="feature-item">
          <i class="pi pi-list"></i>
          <span>Catat Pesanan</span>
        </div>
        <div class="feature-item">
          <i class="pi pi-calculator"></i>
          <span>Hitung Otomatis</span>
        </div>
        <div class="feature-item">
          <i class="pi pi-users"></i>
          <span>Kelola Kontak</span>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <p class="login-footer">
      &copy; 2024 ListMak. All rights reserved.
    </p>
  </div>
</template>

<script>
import Button from 'primevue/button'
import { loginWithGoogle, saveUser } from '../api/auth'

export default {
  name: 'LoginView',
  components: {
    Button
  },
  data() {
    return {
      isLoading: false
    }
  },
  mounted() {
    // Check for callback params
    this.handleCallback()
  },
  methods: {
    handleGoogleLogin() {
      this.isLoading = true
      loginWithGoogle()
    },
    handleCallback() {
      // Check URL for user data from callback
      const urlParams = new URLSearchParams(window.location.search)
      const userParam = urlParams.get('user')
      
      if (userParam) {
        try {
          const user = JSON.parse(decodeURIComponent(userParam))
          saveUser(user)
          this.$toast.add({
            severity: 'success',
            summary: 'Login Berhasil',
            detail: `Selamat datang, ${user.name}!`,
            life: 3000
          })
          // Clear URL params and redirect
          window.history.replaceState({}, document.title, '/')
          this.$router.push('/today')
        } catch (e) {
          console.error('Failed to parse user data:', e)
        }
      }
      
      // Check for error
      const error = urlParams.get('error')
      if (error) {
        this.$toast.add({
          severity: 'error',
          summary: 'Login Gagal',
          detail: decodeURIComponent(error),
          life: 5000
        })
        window.history.replaceState({}, document.title, '/')
      }
    }
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: linear-gradient(180deg, #0f172a 0%, #1e293b 100%);
}

.login-card {
  width: 100%;
  max-width: 360px;
  background: rgba(30, 41, 59, 0.6);
  backdrop-filter: blur(10px);
  border-radius: 1.5rem;
  padding: 2rem 1.5rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.4);
}

.login-header {
  text-align: center;
  margin-bottom: 2rem;
}

.logo-wrapper {
  width: 80px;
  height: 80px;
  margin: 0 auto 1rem;
}

.logo-icon-img {
  width: 80px;
  height: 80px;
}

.app-title {
  font-size: 1.75rem;
  font-weight: 700;
  color: #f1f5f9;
  margin-bottom: 0.25rem;
}

.app-subtitle {
  color: #94a3b8;
  font-size: 0.875rem;
  line-height: 1.4;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.google-btn {
  width: 100%;
  padding: 0.875rem 1rem;
  background: #ffffff !important;
  color: #374151 !important;
  border: none !important;
  border-radius: 0.75rem;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  transition: all 0.2s ease;
}

.google-btn:hover {
  background: #f3f4f6 !important;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.google-icon {
  flex-shrink: 0;
}

.btn-text {
  font-size: 0.9375rem;
}

.login-terms {
  text-align: center;
  color: #64748b;
  font-size: 0.75rem;
  line-height: 1.5;
}

.login-link {
  color: #3b82f6;
  text-decoration: none;
}

.login-link:hover {
  text-decoration: underline;
}

.features-preview {
  display: flex;
  justify-content: space-around;
  padding-top: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.feature-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  color: #64748b;
}

.feature-item i {
  font-size: 1.25rem;
  color: #3b82f6;
}

.feature-item span {
  font-size: 0.625rem;
  text-align: center;
}

.login-footer {
  margin-top: 2rem;
  color: #475569;
  font-size: 0.75rem;
}
</style>
