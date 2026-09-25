/**
 * Mirror curated GFI + welcome issues onto GitCode (CN hub).
 *
 * Usage (PowerShell):
 *   $env:GITCODE_TOKEN = '<PAT from GitCode → Personal Settings → Access Tokens>'
 *   node scripts/create-gitcode-gfi-issues.mjs
 *
 * Needs scopes: projects + issues. Token is read from env only — never commit it.
 */
import { execFileSync } from 'node:child_process'

const owner = 'yongfeng9m-'
// GitCode path may still be historical `reqmanpy`; UI display name is Reqmango.
const repo = process.env.GITCODE_REPO || 'reqmanpy'
const token = process.env.GITCODE_TOKEN || process.env.GITCODE_ACCESS_TOKEN || ''
if (!token) {
  console.error('Set GITCODE_TOKEN to a GitCode personal access token.')
  process.exit(1)
}

const base = 'https://api.gitcode.com/api/v5'

function api(method, path, form = null) {
  const url = `${base}${path}`
  const args = ['-sS', '-X', method, '-H', `Authorization: Bearer ${token}`, '-H', `private-token: ${token}`, '-w', '\n%{http_code}']
  if (form) {
    for (const [k, v] of Object.entries(form)) {
      if (v == null) continue
      args.push('-F', `${k}=${v}`)
    }
  }
  args.push(url)
  const out = execFileSync('curl.exe', args, { encoding: 'utf8' })
  const nl = out.lastIndexOf('\n')
  const body = out.slice(0, nl)
  const code = Number(out.slice(nl + 1).trim())
  let json = null
  try {
    json = JSON.parse(body)
  } catch {
    json = { raw: body }
  }
  return { code, json }
}

function listIssues() {
  const { code, json } = api('GET', `/repos/${owner}/${repo}/issues?state=all&per_page=100`)
  if (code >= 400) {
    console.error('List issues failed', code, json)
    process.exit(1)
  }
  return Array.isArray(json) ? json : []
}

function createIssue({ title, body, labels }) {
  const form = { repo, title, body }
  if (labels?.length) form.labels = labels.join(',')
  const { code, json } = api('POST', `/repos/${owner}/issues`, form)
  if (code >= 400) {
    console.error('Create failed', title, code, json)
    return null
  }
  const num = json.number ?? json.id
  const html = json.html_url ?? `https://gitcode.com/${owner}/${repo}/issues/${num}`
  console.log('created', title, html)
  return json
}

const items = [
  {
    id: 'GFI-01',
    title: '[good first] GFI-01 Record and add Intake demo GIF',
    labels: ['good first issue', 'help wanted'],
    body: `## 任务
录制 15–30 秒：提交 Intake → 分诊队列（类型 / 优先级 / 疑似重复；有 AI key 时更完整）。控制在约 5MB，挂到两边 README 首屏。

## 文件
- \`docs/assets/demo.gif\`（新建）
- \`docs/assets/README.md\`
- \`README.md\` / \`README-zh.md\`

## 完成标准
- [ ] README 能渲染 GIF
- [ ] 无 AI key 时文字走查仍可用

清单：[docs/dev/good-first-issues.md](https://gitcode.com/${owner}/${repo}/blob/master/docs/dev/good-first-issues.md)
镜像参考：https://github.com/vinthuy/reqmango/issues/1`,
  },
  {
    id: 'GFI-04',
    title: '[good first] GFI-04 PageTabConfig built-in names via locale',
    labels: ['good first issue', 'help wanted'],
    body: `## 任务
\`builtInTabs[].name\` 不要存中文字面量；开关/保存时用 \`pageTab.builtInTabs.*\`（key 已存在）。展示已走 \`getTabDisplayName\`。

## 文件
- \`frontend/src/components/PageTabConfig.vue\`

镜像：https://github.com/vinthuy/reqmango/issues/2`,
  },
  {
    id: 'GFI-05',
    title: '[good first] GFI-05 useIssueFilters chip labels i18n',
    labels: ['good first issue', 'help wanted'],
    body: `## 任务
硬编码「状态 / 优先级 / 紧急…」改为 \`t()\`。

## 文件
- \`frontend/src/composables/useIssueFilters.ts\`
- locales

镜像：https://github.com/vinthuy/reqmango/issues/3`,
  },
  {
    id: 'GFI-06',
    title: '[good first] GFI-06 Priority / state group label helpers i18n',
    labels: ['good first issue', 'help wanted'],
    body: `## 任务
\`getPriorityName\` / \`getStateGroupName\` 跟语言走。

## 文件
- \`frontend/src/types/issue.ts\` 及调用点

镜像：https://github.com/vinthuy/reqmango/issues/4`,
  },
  {
    id: 'GFI-07',
    title: '[good first] GFI-07 ProjectSettings toast errors i18n',
    labels: ['good first issue', 'help wanted'],
    body: `## 任务
英文 toast 兜底改为 \`t('…')\`。

## 文件
- \`frontend/src/views/ProjectSettings.vue\`
- locales

镜像：https://github.com/vinthuy/reqmango/issues/5`,
  },
  {
    id: 'GFI-08',
    title: '[good first] GFI-08 Router meta titles i18n',
    labels: ['good first issue', 'help wanted'],
    body: `## 任务
路由 \`meta.title\` 中文硬编码改为 \`t()\` 驱动。

## 文件
- \`frontend/src/router/index.ts\`

镜像：https://github.com/vinthuy/reqmango/issues/6`,
  },
  {
    id: 'GFI-09',
    title: '[good first] GFI-09 Add npm script i18n:check',
    labels: ['good first issue', 'help wanted'],
    body: `## 任务
增加 \`i18n:check\`，跑 \`frontend/scripts/scan-missing-i18n-keys.mjs\`，有真实缺口则非零退出；写入 CONTRIBUTING。

镜像：https://github.com/vinthuy/reqmango/issues/7`,
  },
  {
    id: 'GFI-10',
    title: '[good first] GFI-10 Improve .env.example contributor comments',
    labels: ['good first issue', 'help wanted'],
    body: `## 任务
补全端口、demo 账号、可选 \`AI_*\`、勿提交真 key 的注释。

## 文件
- \`.env.example\`

镜像：https://github.com/vinthuy/reqmango/issues/8`,
  },
  {
    id: 'GFI-11',
    title: '[good first] GFI-11 Compose first-boot guidance',
    labels: ['good first issue', 'help wanted'],
    body: `## 任务
首启等待说明：首次 build 可能几分钟；避免静默白屏。

## 文件
- \`docker-compose.yml\` / README

镜像：https://github.com/vinthuy/reqmango/issues/9`,
  },
  {
    id: 'WELCOME',
    title: '欢迎共建：Good First Issues（本仓为国内主场）',
    labels: ['help wanted'],
    body: `## 欢迎

Reqmango 自建项目管理：新需求先 Intake 分诊再进 backlog；AI 嵌在 Issue / Intake / Cycle。

- 贡献指南：https://gitcode.com/${owner}/${repo}/blob/master/CONTRIBUTING-zh.md
- 任务清单：https://gitcode.com/${owner}/${repo}/blob/master/docs/dev/good-first-issues.md
- 国际镜像：https://github.com/vinthuy/reqmango（欢迎 Issue 见 https://github.com/vinthuy/reqmango/issues/13）

认领：在对应 Issue 评论「我来做」，一次一个。小步 PR 最欢迎。`,
  },
]

const existing = listIssues()
const titles = new Set(existing.map((i) => String(i.title || '')))

for (const item of items) {
  if ([...titles].some((t) => t.includes(item.id) || t === item.title)) {
    console.log('skip existing', item.id)
    continue
  }
  createIssue(item)
}

console.log('done')
