// Create curated GFI issues on GitHub mirror
// Usage: node frontend/scripts/create-gfi-issues.mjs  (requires gh auth)

import { execFileSync } from 'node:child_process'

const repo = 'vinthuy/reqmango'

const items = [
  {
    id: 'GFI-01',
    title: 'Record and add Intake demo GIF',
    labels: ['good first issue', 'help wanted', 'area:docs'],
    body: `## Task
Record a 15–30s clip: submit Intake → triage queue (type / priority / duplicate hints if AI key set). Keep under ~5 MB. Link it in both READMEs above the fold.

## Files
- \`docs/assets/demo.gif\` (new)
- \`docs/assets/README.md\`
- \`README.md\` / \`README-zh.md\`

## Done when
- [ ] GIF renders on GitHub README
- [ ] Walkthrough text still works without AI key

See [docs/dev/good-first-issues.md](https://github.com/vinthuy/reqmango/blob/master/docs/dev/good-first-issues.md).

**CN hub:** prefer discussing on [GitCode](https://gitcode.com/yongfeng9m-/reqmanpy) — same product, same task.`,
  },
  {
    id: 'GFI-04',
    title: 'PageTabConfig built-in names via locale',
    labels: ['good first issue', 'help wanted', 'area:frontend'],
    body: `## Task
Stop storing Chinese literals in \`builtInTabs[].name\`. Use \`pageTab.builtInTabs.*\` when toggling/saving (keys already exist). Display already uses \`getTabDisplayName\`.

## Files
- \`frontend/src/components/PageTabConfig.vue\`

## Done when
- [ ] Enabling a built-in tab persists a stable type
- [ ] UI label follows locale

Ref: [good-first-issues.md](https://github.com/vinthuy/reqmango/blob/master/docs/dev/good-first-issues.md)`,
  },
  {
    id: 'GFI-05',
    title: 'useIssueFilters chip labels i18n',
    labels: ['good first issue', 'help wanted', 'area:frontend'],
    body: `## Task
Replace hardcoded \`状态\` / \`优先级\` / \`紧急\`… with \`t()\` (e.g. \`common.*\` / priority keys).

## Files
- \`frontend/src/composables/useIssueFilters.ts\`
- \`frontend/src/locales/zh-CN.json\`
- \`frontend/src/locales/en-US.json\`

## Done when
- [ ] Filter chips follow locale
- [ ] Related vitest updated/passing

Ref: [good-first-issues.md](https://github.com/vinthuy/reqmango/blob/master/docs/dev/good-first-issues.md)`,
  },
  {
    id: 'GFI-06',
    title: 'Priority / state group label helpers i18n',
    labels: ['good first issue', 'help wanted', 'area:frontend'],
    body: `## Task
Make \`getPriorityName\` / \`getStateGroupName\` locale-aware (pass \`t\` or thin helper). Do not leave Chinese-only maps for UI.

## Files
- \`frontend/src/types/issue.ts\`
- Call sites as needed

## Done when
- [ ] Priority/state group names switch with locale in touched views

Ref: [good-first-issues.md](https://github.com/vinthuy/reqmango/blob/master/docs/dev/good-first-issues.md)`,
  },
  {
    id: 'GFI-07',
    title: 'ProjectSettings toast errors i18n',
    labels: ['good first issue', 'help wanted', 'area:frontend'],
    body: `## Task
Replace English toast fallbacks (\`Failed to load settings data\`, etc.) with \`t('…')\` keys.

## Files
- \`frontend/src/views/ProjectSettings.vue\`
- locales

## Done when
- [ ] Forced error path shows translated toast in zh and en

Ref: [good-first-issues.md](https://github.com/vinthuy/reqmango/blob/master/docs/dev/good-first-issues.md)`,
  },
  {
    id: 'GFI-08',
    title: 'Router meta titles i18n',
    labels: ['good first issue', 'help wanted', 'area:frontend'],
    body: `## Task
Hardcoded Chinese \`meta.title\` (\`工作流\`, \`预算与SLA\`, …). Prefer driving document title through \`t()\`.

## Files
- \`frontend/src/router/index.ts\`
- document title setter if any

## Done when
- [ ] Browser tab title follows locale for changed routes

Ref: [good-first-issues.md](https://github.com/vinthuy/reqmango/blob/master/docs/dev/good-first-issues.md)`,
  },
  {
    id: 'GFI-09',
    title: 'Add npm script i18n:check',
    labels: ['good first issue', 'help wanted', 'area:docs'],
    body: `## Task
Add npm script e.g. \`i18n:check\` that runs \`frontend/scripts/scan-missing-i18n-keys.mjs\` and exits non-zero if real gaps remain (ignore dynamic \`pageTab.builtInTabs.\` / \`plugin.\` prefixes). Document in CONTRIBUTING.

## Files
- \`frontend/package.json\`
- \`frontend/scripts/scan-missing-i18n-keys.mjs\`
- \`CONTRIBUTING.md\` / \`CONTRIBUTING-zh.md\`

## Done when
- [ ] \`npm run i18n:check\` works and is documented

Ref: [good-first-issues.md](https://github.com/vinthuy/reqmango/blob/master/docs/dev/good-first-issues.md)`,
  },
  {
    id: 'GFI-10',
    title: 'Improve .env.example contributor comments',
    labels: ['good first issue', 'help wanted', 'area:docs'],
    body: `## Task
Ensure comments cover: ports, demo login pointer, optional \`AI_*\`, and “never commit real keys”.

## Files
- \`.env.example\`

## Done when
- [ ] New contributor can configure Compose without reading backend code

Ref: [good-first-issues.md](https://github.com/vinthuy/reqmango/blob/master/docs/dev/good-first-issues.md)`,
  },
  {
    id: 'GFI-11',
    title: 'Compose first-boot guidance',
    labels: ['good first issue', 'help wanted', 'area:docs'],
    body: `## Task
Confirm frontend does not serve before backend is ready (or document retry). Improve README “first boot takes a few minutes” note.

## Files
- \`docker-compose.yml\`
- \`README.md\` / \`README-zh.md\`

## Done when
- [ ] README states expected first-build time
- [ ] No silent blank page without guidance

Ref: [good-first-issues.md](https://github.com/vinthuy/reqmango/blob/master/docs/dev/good-first-issues.md)`,
  },
]

function sh(args) {
  return execFileSync('gh', args, { encoding: 'utf8' }).trim()
}

for (const item of items) {
  const title = `[good first] ${item.id} ${item.title}`
  const existing = sh(['issue', 'list', '-R', repo, '--search', item.id, '--json', 'number,title', '--limit', '5'])
  const list = JSON.parse(existing || '[]')
  if (list.some((i) => String(i.title).includes(item.id))) {
    console.log('skip existing', item.id)
    continue
  }
  const out = sh([
    'issue',
    'create',
    '-R',
    repo,
    '--title',
    title,
    '--body',
    item.body,
    ...item.labels.flatMap((l) => ['--label', l]),
  ])
  console.log('created', item.id, out)
}

// Close already-done items as reference issues if missing
const done = [
  {
    id: 'GFI-03',
    title: 'TriagePanel i18n (done by maintainers)',
    body: `Completed on \`master\` — TriagePanel uses \`intake.*\` locale keys via \`useI18n\`.\n\nLeaving this closed issue so the GFI list stays traceable.`,
  },
  {
    id: 'GFI-02',
    title: 'GitHub About checklist doc (done)',
    body: `Completed — see \`docs/dev/github-public-profile.md\` and \`docs/dev/dual-remote.md\`.`,
  },
  {
    id: 'GFI-12',
    title: 'CONTRIBUTING links from README (done)',
    body: `Completed — README / README-zh link CONTRIBUTING and good-first list.`,
  },
]

for (const item of done) {
  const existing = sh(['issue', 'list', '-R', repo, '--state', 'all', '--search', item.id, '--json', 'number,title,state', '--limit', '5'])
  const list = JSON.parse(existing || '[]')
  if (list.some((i) => String(i.title).includes(item.id))) {
    console.log('skip existing done', item.id)
    continue
  }
  const url = sh([
    'issue',
    'create',
    '-R',
    repo,
    '--title',
    `[good first] ${item.id} ${item.title}`,
    '--body',
    item.body,
    '--label',
    'good first issue',
    '--label',
    'area:docs',
  ])
  const num = url.match(/\/issues\/(\d+)/)?.[1]
  if (num) {
    sh(['issue', 'close', num, '-R', repo, '--reason', 'completed', '--comment', 'Marked done — see master.'])
    console.log('created+closed', item.id, url)
  }
}
