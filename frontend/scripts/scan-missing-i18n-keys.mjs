import fs from 'node:fs'
import path from 'node:path'
const zh = JSON.parse(fs.readFileSync(new URL('../src/locales/zh-CN.json', import.meta.url), 'utf8'))
const en = JSON.parse(fs.readFileSync(new URL('../src/locales/en-US.json', import.meta.url), 'utf8'))

function flatten(o, p = '') {
  const r = {}
  for (const [k, v] of Object.entries(o || {})) {
    const nk = p ? `${p}.${k}` : k
    if (v && typeof v === 'object' && !Array.isArray(v)) Object.assign(r, flatten(v, nk))
    else r[nk] = v
  }
  return r
}

function walk(d, acc = []) {
  for (const x of fs.readdirSync(d, { withFileTypes: true })) {
    const p = path.join(d, x.name)
    if (x.isDirectory()) {
      if (!['node_modules', 'dist'].includes(x.name)) walk(p, acc)
    } else if (/\.(vue|ts)$/.test(x.name) && !/\.(test|spec)\./.test(x.name)) acc.push(p)
  }
  return acc
}

const z = flatten(zh)
const e = flatten(en)
const files = walk('./src')
const keyRe = /\bt\(\s*(['"])([a-zA-Z0-9_.]+)\1\s*(\+)?/g
const used = new Map()

for (const f of files) {
  const text = fs.readFileSync(f, 'utf8')
  let m
  while ((m = keyRe.exec(text))) {
    if (m[3]) continue // Concatenated keys are dynamic, not missing literal keys.
    const k = m[2]
    if (!used.has(k)) used.set(k, [])
    used.get(k).push(f.replaceAll('\\', '/'))
  }
}

const missing = []
for (const [k, refs] of used) {
  const inZ = k in z
  const inE = k in e
  if (!inZ || !inE) missing.push({ k, inZ, inE, sample: refs[0] })
}
missing.sort((a, b) => a.k.localeCompare(b.k))
console.log(`used=${used.size} missing=${missing.length}`)
for (const x of missing) {
  console.log(`${!x.inZ && !x.inE ? 'BOTH' : !x.inZ ? 'ZH' : 'EN'}\t${x.k}\t${x.sample}`)
}
if (missing.length) process.exitCode = 1
