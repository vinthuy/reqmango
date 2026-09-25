import fs from 'node:fs'

function findDuplicateTopKeys(file) {
  const text = fs.readFileSync(file, 'utf8')
  const keys = []
  // top-level keys: lines that start with 2 spaces + "key":
  for (const line of text.split(/\n/)) {
    const m = line.match(/^  "([^"]+)":\s*(\{|\[|"|-?\d|true|false|null)/)
    if (m) keys.push(m[1])
  }
  const seen = new Map()
  const dups = []
  keys.forEach((k, i) => {
    if (seen.has(k)) dups.push({ k, first: seen.get(k), later: i + 1 })
    else seen.set(k, i + 1)
  })
  return dups
}

for (const file of ['src/locales/zh-CN.json', 'src/locales/en-US.json']) {
  const dups = findDuplicateTopKeys(file)
  console.log('\n' + file, 'dups', dups.length)
  for (const d of dups) console.log(`  ${d.k}`)
}
