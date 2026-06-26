#!/usr/bin/env node
/**
 * Usage: node scripts/gen-changelog.js <version>
 * Example: node scripts/gen-changelog.js 1.2.0
 *
 * Reads git log since last changelog entry, groups by conventional commit type,
 * and prepends a draft entry to frontend/src/data/changelog.json.
 * Edit the file before committing.
 */

const { execSync } = require('child_process')
const fs = require('fs')
const path = require('path')

const CHANGELOG_PATH = path.join(__dirname, '../frontend/src/data/changelog.json')

const TYPE_MAP = {
  feat:     { label: 'feat',     display: 'Fitur Baru' },
  fix:      { label: 'fix',      display: 'Perbaikan' },
  security: { label: 'security', display: 'Keamanan' },
  perf:     { label: 'perf',     display: 'Performa' },
  refactor: { label: 'refactor', display: 'Peningkatan' },
  docs:     { label: 'docs',     display: 'Dokumentasi' },
  chore:    { label: 'chore',    display: 'Pemeliharaan' },
}

const version = process.argv[2]
if (!version) {
  console.error('Usage: node scripts/gen-changelog.js <version>')
  process.exit(1)
}

const changelog = JSON.parse(fs.readFileSync(CHANGELOG_PATH, 'utf-8'))
const lastDate = changelog[0]?.date ?? '2020-01-01'

const log = execSync(`git log --oneline --since="${lastDate}T00:00:00" --no-merges`, { encoding: 'utf-8' })
  .trim()
  .split('\n')
  .filter(Boolean)

const entries = []
for (const line of log) {
  const match = line.match(/^[a-f0-9]+ (?:(\w+)(?:\([^)]+\))?!?: )?(.+)$/)
  if (!match) continue
  const [, rawType, message] = match
  const type = TYPE_MAP[rawType?.toLowerCase()] ? rawType.toLowerCase() : 'chore'
  entries.push({ type, message })
}

if (entries.length === 0) {
  console.log('No new commits since last entry. Nothing added.')
  process.exit(0)
}

const newEntry = {
  version,
  date: new Date().toISOString().slice(0, 10),
  entries,
}

changelog.unshift(newEntry)
fs.writeFileSync(CHANGELOG_PATH, JSON.stringify(changelog, null, 2) + '\n')
console.log(`Added ${entries.length} entries for v${version}. Edit ${CHANGELOG_PATH} before committing.`)
