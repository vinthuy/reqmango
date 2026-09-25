import fs from 'node:fs'
import path from 'node:path'

function walk(d, acc = []) {
  for (const e of fs.readdirSync(d, { withFileTypes: true })) {
    const p = path.join(d, e.name)
    if (e.isDirectory()) {
      if (!['node_modules', 'dist', '__tests__'].includes(e.name)) walk(p, acc)
    } else if (/\.(vue|ts)$/.test(e.name) && !/\.(test|spec)\./.test(e.name)) {
      acc.push(p)
    }
  }
  return acc
}

const files = walk('src')
const re = /(['"`])([^'"`\n]*[\u4e00-\u9fff][^'"`\n]*)\1/g
const hits = []

for (const f of files) {
  const lines = fs.readFileSync(f, 'utf8').split(/\n/)
  lines.forEach((line, i) => {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('//') || trimmed.startsWith('*') || trimmed.startsWith('<!--')) return
    if (!/[\u4e00-\u9fff]/.test(line)) return
    // skip pure comments mid-line after //
    const code = line.replace(/\/\/.*$/, '').replace(/<!--.*?-->/g, '')
    if (!/[\u4e00-\u9fff]/.test(code)) return
    const matches = [...code.matchAll(re)].map((m) => m[2])
    if (!matches.length) return
    hits.push({
      f: f.replaceAll('\\', '/'),
      i: i + 1,
      s: matches.join(' | '),
    })
  })
}

console.log(`hits ${hits.length}`)
for (const h of hits) console.log(`${h.f}:${h.i}\t${h.s}`)
