<template>
  <div class="legal-page">
    <header class="legal-topbar">
      <router-link to="/" class="legal-brand">
        <img :src="logoIcon" alt="Listmak" class="legal-logo" />
        <span>Listmak</span>
      </router-link>
      <router-link to="/" class="legal-back">
        <i class="pi pi-arrow-left"></i>
        <span>Beranda</span>
      </router-link>
    </header>

    <article class="legal-body">
      <p class="legal-eyebrow">Dokumen Hukum</p>
      <h1 class="legal-title">{{ title }}</h1>
      <p class="legal-updated">Terakhir diperbarui: {{ updated }}</p>

      <div class="legal-content">
        <slot />
      </div>
    </article>

    <footer class="legal-footer">
      <p>
        Listmak dibuat oleh
        <a href="https://enambelas.dev" target="_blank" rel="noopener">enambelas.dev</a>.
        Pertanyaan? <a href="mailto:hello@enambelas.dev">hello@enambelas.dev</a>
      </p>
      <nav class="legal-footer-links">
        <router-link to="/privacy">Kebijakan Privasi</router-link>
        <span aria-hidden="true">·</span>
        <router-link to="/terms">Syarat &amp; Ketentuan</router-link>
      </nav>
    </footer>
  </div>
</template>

<script>
import logoIcon from '../assets/logo/webp/listmak-icon-128.webp'

export default {
  name: 'LegalLayout',
  props: {
    title: { type: String, required: true },
    updated: { type: String, default: '27 Juni 2026' }
  },
  data() {
    return { logoIcon }
  }
}
</script>

<style scoped>
.legal-page {
  --ink: #0f172a;
  --panel: #1e293b;
  --line: rgba(148, 163, 184, 0.16);
  --text: #e2e8f0;
  --muted: #94a3b8;
  --accent: #3b82f6;
  min-height: 100vh;
  min-height: 100dvh;
  background: var(--ink);
  color: var(--text);
  font-family: 'Plus Jakarta Sans', system-ui, sans-serif;
  display: flex;
  flex-direction: column;
}

.legal-topbar {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  background: rgba(15, 23, 42, 0.85);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--line);
}

.legal-brand {
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
  font-weight: 700;
  color: var(--text);
  text-decoration: none;
}

.legal-logo {
  width: 28px;
  height: 28px;
}

.legal-back {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.875rem;
  color: var(--muted);
  text-decoration: none;
  transition: color 0.15s ease;
}

.legal-back:hover { color: var(--accent); }

.legal-body {
  width: 100%;
  max-width: 720px;
  margin: 0 auto;
  padding: 2.5rem 1.25rem 1rem;
  flex: 1;
}

.legal-eyebrow {
  font-family: 'Space Mono', monospace;
  font-size: 0.7rem;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--accent);
  margin-bottom: 0.75rem;
}

.legal-title {
  font-family: 'Bricolage Grotesque', system-ui, sans-serif;
  font-size: clamp(1.9rem, 5vw, 2.8rem);
  font-weight: 800;
  line-height: 1.05;
  letter-spacing: -0.02em;
  color: #f8fafc;
}

.legal-updated {
  margin-top: 0.75rem;
  font-size: 0.85rem;
  color: var(--muted);
}

.legal-content {
  margin-top: 2.25rem;
  font-size: 0.975rem;
  line-height: 1.75;
  color: #cbd5e1;
}

.legal-content :deep(h2) {
  font-family: 'Bricolage Grotesque', system-ui, sans-serif;
  font-size: 1.25rem;
  font-weight: 700;
  color: #f1f5f9;
  margin: 2.25rem 0 0.5rem;
  padding-top: 1.25rem;
  border-top: 1px solid var(--line);
}

.legal-content :deep(h2:first-child) {
  border-top: none;
  padding-top: 0;
  margin-top: 0;
}

.legal-content :deep(p) { margin: 0.75rem 0; }

.legal-content :deep(ul) {
  margin: 0.75rem 0;
  padding-left: 1.1rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.legal-content :deep(li) { padding-left: 0.25rem; }

.legal-content :deep(a) {
  color: var(--accent);
  text-decoration: underline;
  text-underline-offset: 2px;
}

.legal-content :deep(strong) { color: #f1f5f9; }

.legal-footer {
  max-width: 720px;
  width: 100%;
  margin: 0 auto;
  padding: 2rem 1.25rem 2.5rem;
  border-top: 1px solid var(--line);
  margin-top: 2rem;
  font-size: 0.85rem;
  color: var(--muted);
}

.legal-footer a { color: var(--accent); text-decoration: none; }
.legal-footer a:hover { text-decoration: underline; }

.legal-footer-links {
  margin-top: 0.6rem;
  display: flex;
  gap: 0.6rem;
  align-items: center;
}
</style>
