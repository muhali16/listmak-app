<template>
  <div class="lp">
    <!-- Nav -->
    <header class="lp-nav">
      <a href="#top" class="lp-brand">
        <img :src="logoIcon" alt="Listmak" class="lp-brand-logo" />
        <span>Listmak</span>
      </a>
      <div class="lp-nav-actions">
        <a href="#cara-kerja" class="lp-nav-link">Cara kerja</a>
        <a href="#fitur" class="lp-nav-link">Fitur</a>
        <Button v-if="loggedIn" class="lp-nav-cta" @click="goApp">Buka aplikasi</Button>
        <Button v-else class="lp-nav-cta" :loading="isLoading" @click="login">Masuk</Button>
      </div>
    </header>

    <!-- Hero -->
    <section id="top" class="lp-hero">
      <div class="lp-hero-copy">
        <p class="lp-eyebrow">Buat tim kantor yang patungan makan</p>
        <h1 class="lp-h1">
          Pesan makan bareng,<br />
          beres tanpa <span class="lp-accent">ribet rekap.</span>
        </h1>
        <p class="lp-sub">
          Kumpulkan pesanan satu tim, total kehitung otomatis, lalu bagikan link
          biar semua isi sendiri. Nggak ada lagi chat <em>“punyaku tadi berapa ya?”</em>
        </p>
        <div class="lp-cta-row">
          <Button v-if="loggedIn" class="lp-cta-primary" @click="goApp">
            <i class="pi pi-arrow-right"></i>
            <span>Buka aplikasi</span>
          </Button>
          <Button v-else class="lp-cta-primary" :loading="isLoading" @click="login">
            <svg class="lp-g" viewBox="0 0 24 24" width="18" height="18">
              <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
              <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
              <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
              <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
            </svg>
            <span>Mulai dengan Google</span>
          </Button>
          <a href="#cara-kerja" class="lp-cta-ghost">Lihat cara kerja</a>
        </div>
        <p class="lp-trust">Gratis · Tanpa kartu kredit · Masuk pakai akun Google</p>
      </div>

      <!-- Signature: warung-style receipt tally -->
      <div class="lp-receipt-wrap">
        <div class="lp-receipt" role="img" aria-label="Contoh rekap pesanan otomatis Listmak">
          <div class="lp-receipt-head">
            <span class="lp-receipt-store">LISTMAK · MAKAN SIANG</span>
            <span class="lp-receipt-date">Jum, 27 Jun · Tim Produk</span>
          </div>
          <div class="lp-receipt-rule"></div>
          <ul class="lp-receipt-lines">
            <li v-for="(o, i) in orders" :key="i" class="lp-line" :style="{ animationDelay: revealDelay(i) }">
              <span class="lp-line-name">{{ o.name }}</span>
              <span class="lp-line-item">{{ o.item }}</span>
              <span class="lp-line-price">{{ rupiah(o.price) }}</span>
            </li>
          </ul>
          <div class="lp-receipt-rule lp-receipt-rule--dashed"></div>
          <div class="lp-receipt-total">
            <span class="lp-total-label">TOTAL</span>
            <span class="lp-total-value">{{ rupiah(displayTotal) }}</span>
          </div>
          <div class="lp-receipt-meta">
            <span>{{ orders.length }} orang</span>
            <span class="lp-stamp">REKAP OTOMATIS</span>
          </div>
        </div>
        <div class="lp-receipt-glow" aria-hidden="true"></div>
      </div>
    </section>

    <!-- How it works (real 3-step sequence → numbered) -->
    <section id="cara-kerja" class="lp-section">
      <p class="lp-section-eyebrow">Cara kerja</p>
      <h2 class="lp-h2">Tiga langkah, selesai sebelum makan datang.</h2>
      <ol class="lp-steps">
        <li v-for="(s, i) in steps" :key="i" class="lp-step">
          <span class="lp-step-num">{{ String(i + 1).padStart(2, '0') }}</span>
          <h3 class="lp-step-title">{{ s.title }}</h3>
          <p class="lp-step-desc">{{ s.desc }}</p>
        </li>
      </ol>
    </section>

    <!-- Features -->
    <section id="fitur" class="lp-section lp-section--alt">
      <p class="lp-section-eyebrow">Fitur</p>
      <h2 class="lp-h2">Semua yang dibutuhin si tukang rekap.</h2>
      <div class="lp-features">
        <div v-for="(f, i) in features" :key="i" class="lp-feature">
          <div class="lp-feature-ico"><i :class="f.icon"></i></div>
          <h3 class="lp-feature-title">{{ f.title }}</h3>
          <p class="lp-feature-desc">{{ f.desc }}</p>
        </div>
      </div>
    </section>

    <!-- Final CTA -->
    <section class="lp-final">
      <h2 class="lp-final-title">Stop jadi kalkulator grup.</h2>
      <p class="lp-final-sub">Bikin listmak pertama kamu hari ini.</p>
      <Button v-if="loggedIn" class="lp-cta-primary" @click="goApp">
        <i class="pi pi-arrow-right"></i><span>Buka aplikasi</span>
      </Button>
      <Button v-else class="lp-cta-primary" :loading="isLoading" @click="login">
        <svg class="lp-g" viewBox="0 0 24 24" width="18" height="18">
          <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
          <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
          <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
          <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
        </svg>
        <span>Mulai dengan Google</span>
      </Button>
    </section>

    <!-- Footer -->
    <footer class="lp-footer">
      <div class="lp-footer-brand">
        <img :src="logoIcon" alt="Listmak" class="lp-brand-logo" />
        <span>Listmak</span>
      </div>
      <p class="lp-footer-by">
        Dibuat oleh
        <a href="https://enambelas.dev" target="_blank" rel="noopener">enambelas.dev</a>
        · <a href="mailto:hello@enambelas.dev">hello@enambelas.dev</a>
      </p>
      <nav class="lp-footer-links">
        <router-link to="/privacy">Kebijakan Privasi</router-link>
        <router-link to="/terms">Syarat &amp; Ketentuan</router-link>
      </nav>
      <p class="lp-footer-copy">&copy; {{ year }} enambelas.dev. Semua hak dilindungi.</p>
    </footer>
  </div>
</template>

<script>
import Button from 'primevue/button'
import logoIcon from '../assets/logo/webp/listmak-icon-128.webp'
import { loginWithGoogle, saveUser, isAuthenticated } from '../api/auth'

export default {
  name: 'LandingView',
  components: { Button },
  data() {
    return {
      logoIcon,
      isLoading: false,
      loggedIn: isAuthenticated(),
      year: new Date().getFullYear(),
      displayTotal: 0,
      raf: null,
      orders: [
        { name: 'Budi', item: 'Nasi Padang', price: 22000 },
        { name: 'Sari', item: 'Ayam Geprek', price: 18000 },
        { name: 'Dewi', item: 'Es Teh ×2', price: 8000 },
        { name: 'Rian', item: 'Mie Ayam', price: 17000 }
      ],
      steps: [
        { title: 'Buat listmak hari ini', desc: 'Mulai daftar baru untuk makan siang tim. Tambah item secepat ngetik chat.' },
        { title: 'Bagikan link ke tim', desc: 'Kirim satu tautan. Semua orang isi pesanannya sendiri — tanpa perlu login.' },
        { title: 'Total kehitung otomatis', desc: 'Lihat rekap per orang dan total keseluruhan. Tinggal setor, beres.' }
      ],
      features: [
        { icon: 'pi pi-list', title: 'Catat pesanan kilat', desc: 'Tambah item dan harga dalam hitungan detik.' },
        { icon: 'pi pi-calculator', title: 'Hitung total otomatis', desc: 'Per orang dan total tim langsung terjumlah, tanpa kalkulator.' },
        { icon: 'pi pi-share-alt', title: 'Link tanpa login', desc: 'Bagikan tautan, tim isi pesanan sendiri dari HP masing-masing.' },
        { icon: 'pi pi-sparkles', title: 'Tempel chat jadi pesanan', desc: 'Salin pesanan dari grup, biarkan AI yang merapikan.' },
        { icon: 'pi pi-users', title: 'Kelola kontak tim', desc: 'Simpan anggota tim biar nggak ngetik nama berulang.' },
        { icon: 'pi pi-calendar', title: 'Rekap harian', desc: 'Riwayat pesanan tiap hari tersimpan rapi dan bisa dibuka lagi.' }
      ]
    }
  },
  computed: {
    grandTotal() {
      return this.orders.reduce((s, o) => s + o.price, 0)
    }
  },
  mounted() {
    this.handleCallback()
    this.animateTotal()
  },
  beforeUnmount() {
    if (this.raf) cancelAnimationFrame(this.raf)
  },
  methods: {
    rupiah(n) {
      return 'Rp' + Math.round(n).toLocaleString('id-ID')
    },
    revealDelay(i) {
      return `${0.15 + i * 0.12}s`
    },
    login() {
      this.isLoading = true
      loginWithGoogle()
    },
    goApp() {
      this.$router.push('/today')
    },
    animateTotal() {
      const reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches
      if (reduce) {
        this.displayTotal = this.grandTotal
        return
      }
      const target = this.grandTotal
      const duration = 950
      const start = performance.now() + 300 // brief beat before counting
      const tick = (now) => {
        const t = Math.min(1, Math.max(0, (now - start) / duration))
        // easeOutCubic
        const eased = 1 - Math.pow(1 - t, 3)
        this.displayTotal = Math.round(target * eased)
        if (t < 1) this.raf = requestAnimationFrame(tick)
      }
      this.raf = requestAnimationFrame(tick)
    },
    // Backend OAuth redirects to "/" with ?user= or ?error=
    handleCallback() {
      const params = new URLSearchParams(window.location.search)
      const userParam = params.get('user')
      if (userParam) {
        try {
          const user = JSON.parse(decodeURIComponent(userParam))
          saveUser(user)
          window.history.replaceState({}, document.title, '/')
          this.$router.push('/today')
        } catch (e) {
          console.error('Failed to parse user data:', e)
        }
        return
      }
      const error = params.get('error')
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
.lp {
  --ink: #0b1220;
  --ink2: #0f172a;
  --panel: #1e293b;
  --line: rgba(148, 163, 184, 0.14);
  --text: #f1f5f9;
  --muted: #94a3b8;
  --faint: #64748b;
  --accent: #3b82f6;
  --accent-2: #2563eb;
  --teal: #1a8a8a;
  --receipt: #f4efe2;
  --receipt-ink: #2a2417;
  background: var(--ink);
  color: var(--text);
  font-family: 'Plus Jakarta Sans', system-ui, sans-serif;
  overflow-x: hidden;
}

.lp :deep(.p-button) {
  border: none;
  font-family: inherit;
}

/* Nav */
.lp-nav {
  position: sticky;
  top: 0;
  z-index: 20;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.9rem 1.25rem;
  background: rgba(11, 18, 32, 0.8);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--line);
}

.lp-brand {
  display: inline-flex;
  align-items: center;
  gap: 0.55rem;
  font-weight: 700;
  font-size: 1.05rem;
  color: var(--text);
  text-decoration: none;
}

.lp-brand-logo { width: 28px; height: 28px; }

.lp-nav-actions { display: flex; align-items: center; gap: 1.25rem; }

.lp-nav-link {
  color: var(--muted);
  text-decoration: none;
  font-size: 0.9rem;
  transition: color 0.15s ease;
}
.lp-nav-link:hover { color: var(--text); }

.lp-nav-cta {
  background: var(--accent) !important;
  color: #fff !important;
  padding: 0.5rem 1rem !important;
  border-radius: 0.6rem !important;
  font-size: 0.875rem !important;
  font-weight: 600 !important;
}
.lp-nav-cta:hover { background: var(--accent-2) !important; }

@media (max-width: 600px) {
  .lp-nav-link { display: none; }
}

/* Hero */
.lp-hero {
  max-width: 1140px;
  margin: 0 auto;
  padding: 4rem 1.25rem 3rem;
  display: grid;
  grid-template-columns: 1.05fr 0.95fr;
  gap: 3rem;
  align-items: center;
}

.lp-eyebrow {
  font-family: 'Space Mono', monospace;
  font-size: 0.72rem;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--accent);
  margin-bottom: 1.1rem;
}

.lp-h1 {
  font-family: 'Bricolage Grotesque', system-ui, sans-serif;
  font-size: clamp(2.3rem, 6vw, 4rem);
  font-weight: 800;
  line-height: 1.02;
  letter-spacing: -0.03em;
  color: #f8fafc;
}

.lp-accent {
  color: var(--accent);
  position: relative;
  white-space: nowrap;
}

.lp-sub {
  margin-top: 1.25rem;
  font-size: 1.075rem;
  line-height: 1.6;
  color: var(--muted);
  max-width: 30rem;
}
.lp-sub em { color: #cbd5e1; font-style: italic; }

.lp-cta-row {
  margin-top: 2rem;
  display: flex;
  align-items: center;
  gap: 1.25rem;
  flex-wrap: wrap;
}

.lp-cta-primary {
  background: var(--accent) !important;
  color: #fff !important;
  padding: 0.85rem 1.4rem !important;
  border-radius: 0.75rem !important;
  font-weight: 600 !important;
  font-size: 0.975rem !important;
  display: inline-flex !important;
  align-items: center;
  gap: 0.6rem;
  box-shadow: 0 10px 30px -10px rgba(59, 130, 246, 0.6);
  transition: transform 0.15s ease, background 0.15s ease;
}
.lp-cta-primary:hover { background: var(--accent-2) !important; transform: translateY(-2px); }
.lp-g { flex-shrink: 0; background: #fff; border-radius: 3px; padding: 1px; }

.lp-cta-ghost {
  color: var(--text);
  text-decoration: none;
  font-weight: 600;
  font-size: 0.95rem;
  border-bottom: 1px solid var(--faint);
  padding-bottom: 2px;
  transition: border-color 0.15s ease;
}
.lp-cta-ghost:hover { border-color: var(--accent); }

.lp-trust {
  margin-top: 1.5rem;
  font-size: 0.8rem;
  color: var(--faint);
}

/* Receipt signature */
.lp-receipt-wrap { position: relative; display: flex; justify-content: center; }

.lp-receipt {
  position: relative;
  z-index: 2;
  width: 100%;
  max-width: 340px;
  background: var(--receipt);
  color: var(--receipt-ink);
  border-radius: 6px;
  padding: 1.5rem 1.4rem 1.3rem;
  font-family: 'Space Mono', monospace;
  box-shadow: 0 30px 60px -20px rgba(0, 0, 0, 0.7);
  transform: rotate(-1.5deg);
  /* torn-edge top/bottom */
  --tooth: 9px;
  -webkit-mask:
    radial-gradient(var(--tooth) at 50% var(--tooth), #0000 98%, #000) 0 0/calc(var(--tooth)*2) 100% repeat-x,
    radial-gradient(var(--tooth) at 50% calc(100% - var(--tooth)), #0000 98%, #000) 0 100%/calc(var(--tooth)*2) 100% repeat-x;
}

.lp-receipt-head {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  text-align: center;
  margin-bottom: 0.9rem;
}
.lp-receipt-store { font-weight: 700; font-size: 0.82rem; letter-spacing: 0.06em; }
.lp-receipt-date { font-size: 0.68rem; color: #6b6450; }

.lp-receipt-rule { border-top: 2px dashed #cbbf9d; margin: 0.4rem 0; }
.lp-receipt-rule--dashed { border-top-style: dashed; }

.lp-receipt-lines { display: flex; flex-direction: column; gap: 0.55rem; margin: 0.7rem 0; }

.lp-line {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 0.5rem;
  align-items: baseline;
  font-size: 0.78rem;
  animation: lineIn 0.5s ease both;
}
.lp-line-name { font-weight: 700; }
.lp-line-item { color: #6b6450; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.lp-line-price { font-weight: 700; }

.lp-receipt-total {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-top: 0.5rem;
}
.lp-total-label { font-weight: 700; font-size: 0.9rem; }
.lp-total-value { font-weight: 700; font-size: 1.4rem; letter-spacing: -0.02em; }

.lp-receipt-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 0.9rem;
  font-size: 0.66rem;
  color: #6b6450;
}
.lp-stamp {
  border: 2px solid var(--accent);
  color: var(--accent);
  padding: 0.18rem 0.5rem;
  border-radius: 4px;
  font-weight: 700;
  letter-spacing: 0.05em;
  transform: rotate(-4deg);
}

.lp-receipt-glow {
  position: absolute;
  inset: -10% -5%;
  background: radial-gradient(circle at 60% 40%, rgba(59, 130, 246, 0.28), transparent 60%);
  filter: blur(40px);
  z-index: 1;
}

@keyframes lineIn {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Sections */
.lp-section {
  max-width: 1140px;
  margin: 0 auto;
  padding: 4.5rem 1.25rem;
}
.lp-section--alt {
  background: linear-gradient(180deg, transparent, rgba(30, 41, 59, 0.35), transparent);
  max-width: none;
}
.lp-section--alt > * { max-width: 1140px; margin-left: auto; margin-right: auto; }

.lp-section-eyebrow {
  font-family: 'Space Mono', monospace;
  font-size: 0.72rem;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--accent);
  margin-bottom: 0.75rem;
}

.lp-h2 {
  font-family: 'Bricolage Grotesque', system-ui, sans-serif;
  font-size: clamp(1.6rem, 4vw, 2.4rem);
  font-weight: 700;
  letter-spacing: -0.02em;
  color: #f8fafc;
  max-width: 22ch;
}

/* Steps */
.lp-steps {
  margin-top: 2.5rem;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1.5rem;
  list-style: none;
  padding: 0;
  counter-reset: none;
}
.lp-step {
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid var(--line);
  border-radius: 1rem;
  padding: 1.5rem;
}
.lp-step-num {
  font-family: 'Space Mono', monospace;
  font-size: 1.6rem;
  font-weight: 700;
  color: var(--accent);
  display: block;
  margin-bottom: 0.75rem;
}
.lp-step-title {
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 0.4rem;
}
.lp-step-desc { font-size: 0.9rem; line-height: 1.55; color: var(--muted); }

/* Features */
.lp-features {
  margin-top: 2.5rem;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}
.lp-feature {
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid var(--line);
  border-radius: 1rem;
  padding: 1.5rem;
  transition: border-color 0.2s ease, transform 0.2s ease;
}
.lp-feature:hover { border-color: rgba(59, 130, 246, 0.4); transform: translateY(-3px); }
.lp-feature-ico {
  width: 42px;
  height: 42px;
  border-radius: 0.6rem;
  background: rgba(59, 130, 246, 0.12);
  color: var(--accent);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.15rem;
  margin-bottom: 1rem;
}
.lp-feature-title { font-size: 1rem; font-weight: 700; color: var(--text); margin-bottom: 0.35rem; }
.lp-feature-desc { font-size: 0.875rem; line-height: 1.55; color: var(--muted); }

/* Final CTA */
.lp-final {
  max-width: 1140px;
  margin: 0 auto;
  padding: 5rem 1.25rem;
  text-align: center;
}
.lp-final-title {
  font-family: 'Bricolage Grotesque', system-ui, sans-serif;
  font-size: clamp(1.8rem, 5vw, 3rem);
  font-weight: 800;
  letter-spacing: -0.02em;
  color: #f8fafc;
}
.lp-final-sub { margin: 0.75rem 0 1.75rem; color: var(--muted); font-size: 1.05rem; }
.lp-final .lp-cta-primary { margin: 0 auto; }

/* Footer */
.lp-footer {
  border-top: 1px solid var(--line);
  padding: 2.5rem 1.25rem;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
}
.lp-footer-brand {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 700;
  color: var(--text);
}
.lp-footer-by { font-size: 0.875rem; color: var(--muted); }
.lp-footer-by a { color: var(--accent); text-decoration: none; }
.lp-footer-by a:hover { text-decoration: underline; }
.lp-footer-links { display: flex; gap: 1.25rem; }
.lp-footer-links a { color: var(--muted); text-decoration: none; font-size: 0.85rem; }
.lp-footer-links a:hover { color: var(--accent); }
.lp-footer-copy { font-size: 0.78rem; color: var(--faint); margin-top: 0.25rem; }

/* Responsive */
@media (max-width: 860px) {
  .lp-hero { grid-template-columns: 1fr; gap: 2.5rem; padding-top: 2.5rem; }
  .lp-receipt-wrap { order: 2; }
  .lp-steps, .lp-features { grid-template-columns: 1fr; }
}

@media (prefers-reduced-motion: reduce) {
  .lp-line, .lp-cta-primary, .lp-feature { animation: none !important; transition: none !important; }
}
</style>
