<template>
  <div class="changelog-view">
    <div class="changelog-header">
      <h1>Pembaruan Aplikasi</h1>
      <p class="changelog-subtitle">Riwayat perubahan dan peningkatan ListMak</p>
    </div>

    <div class="changelog-list">
      <div v-for="release in changelog" :key="release.version" class="release">
        <div class="release-header">
          <span class="version-badge">v{{ release.version }}</span>
          <span class="release-date">{{ formatDate(release.date) }}</span>
        </div>
        <ul class="entry-list">
          <li v-for="(entry, i) in release.entries" :key="i" class="entry">
            <span class="entry-badge" :class="entry.type">{{ typeLabel(entry.type) }}</span>
            <span class="entry-message">{{ entry.message }}</span>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script>
import changelog from '../data/changelog.json'

const TYPE_LABELS = {
  feat:     'Fitur',
  fix:      'Fix',
  security: 'Keamanan',
  perf:     'Performa',
  refactor: 'Peningkatan',
  docs:     'Docs',
  chore:    'Lainnya',
}

export default {
  name: 'ChangelogView',
  data() {
    return { changelog }
  },
  methods: {
    formatDate(dateStr) {
      return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
    },
    typeLabel(type) {
      return TYPE_LABELS[type] ?? type
    },
  },
}
</script>

<style scoped>
.changelog-view {
  max-width: 720px;
  margin: 0 auto;
  padding: 2rem 1.5rem;
}

.changelog-header {
  margin-bottom: 2rem;
}

.changelog-header h1 {
  font-size: 1.5rem;
  font-weight: 700;
  margin: 0 0 0.25rem;
}

.changelog-subtitle {
  color: var(--p-text-muted-color, #6b7280);
  margin: 0;
  font-size: 0.9rem;
}

.changelog-list {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.release {
  border-left: 2px solid var(--p-primary-color, #3b82f6);
  padding-left: 1.25rem;
}

.release-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}

.version-badge {
  font-weight: 700;
  font-size: 1rem;
  color: var(--p-primary-color, #3b82f6);
}

.release-date {
  font-size: 0.8rem;
  color: var(--p-text-muted-color, #6b7280);
}

.entry-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.entry {
  display: flex;
  align-items: flex-start;
  gap: 0.6rem;
}

.entry-badge {
  flex-shrink: 0;
  font-size: 0.7rem;
  font-weight: 600;
  padding: 0.15rem 0.5rem;
  border-radius: 9999px;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.entry-badge.feat     { background: #dbeafe; color: #1d4ed8; }
.entry-badge.fix      { background: #fef3c7; color: #92400e; }
.entry-badge.security { background: #fee2e2; color: #991b1b; }
.entry-badge.perf     { background: #d1fae5; color: #065f46; }
.entry-badge.refactor { background: #ede9fe; color: #5b21b6; }
.entry-badge.docs     { background: #f3f4f6; color: #374151; }
.entry-badge.chore    { background: #f3f4f6; color: #374151; }

.entry-message {
  font-size: 0.9rem;
  line-height: 1.4;
}

@media (prefers-color-scheme: dark) {
  .entry-badge.feat     { background: #1e3a5f; color: #93c5fd; }
  .entry-badge.fix      { background: #451a03; color: #fcd34d; }
  .entry-badge.security { background: #450a0a; color: #fca5a5; }
  .entry-badge.perf     { background: #052e16; color: #6ee7b7; }
  .entry-badge.refactor { background: #2e1065; color: #c4b5fd; }
  .entry-badge.docs     { background: #1f2937; color: #d1d5db; }
  .entry-badge.chore    { background: #1f2937; color: #d1d5db; }
}
</style>
