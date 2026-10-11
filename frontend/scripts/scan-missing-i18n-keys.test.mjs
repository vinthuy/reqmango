import { test } from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { spawnSync } from 'node:child_process'

function scan(source, zh, en) {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'reqmango-i18n-'))
  try {
    fs.mkdirSync(path.join(root, 'scripts'))
    fs.mkdirSync(path.join(root, 'src/locales'), { recursive: true })
    fs.copyFileSync(new URL('./scan-missing-i18n-keys.mjs', import.meta.url), path.join(root, 'scripts/scan.mjs'))
    fs.writeFileSync(path.join(root, 'src/example.vue'), source)
    fs.writeFileSync(path.join(root, 'src/locales/zh-CN.json'), JSON.stringify(zh))
    fs.writeFileSync(path.join(root, 'src/locales/en-US.json'), JSON.stringify(en))
    return spawnSync(process.execPath, ['scripts/scan.mjs'], { cwd: root, encoding: 'utf8' })
  } finally {
    fs.rmSync(root, { recursive: true, force: true })
  }
}

test('accepts literal keys present in both locales', () => {
  const result = scan("t('common.refresh')", { common: { refresh: '刷新' } }, { common: { refresh: 'Refresh' } })
  assert.equal(result.status, 0)
  assert.match(result.stdout, /missing=0/)
})

test('fails when a literal key is missing in either locale', () => {
  for (const [zh, en] of [[{ common: { refresh: '刷新' } }, {}], [{}, { common: { refresh: 'Refresh' } }], [{}, {}]]) {
    const result = scan('t("common.refresh")', zh, en)
    assert.equal(result.status, 1)
    assert.match(result.stdout, /common.refresh/)
  }
})

test('ignores dynamic concatenations, but not missing literals ending with a dot', () => {
  assert.equal(scan("t('pageTab.builtInTabs.' + tab.tab_type); t('plugin.' + status)", {}, {}).status, 0)
  assert.equal(scan("t('missing.')", {}, {}).status, 1)
})
