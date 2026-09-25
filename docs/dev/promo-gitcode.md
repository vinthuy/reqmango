# GitCode 主场推广（可粘贴）

> 承接另一会话策略：**GitCode = 国内共建主场**；外站已发的不再追；这里只做仓库内可见动作。  
> 仓库：https://gitcode.com/yongfeng9m-/reqmanpy

## 本周在 GitCode 上做这 4 件

| # | 动作 | 材料 |
|---|------|------|
| 1 | **About 简介**（设置页） | 见下方「About」 |
| 2 | **置顶欢迎 Issue** | 正文：`docs/dev/gitcode-welcome-issue.md` |
| 3 | **批量开 GFI Issue** | `scripts/create-gitcode-gfi-issues.mjs`（需 PAT） |
| 4 | **发一条仓库动态** | 见下方「动态」 |

## About（设置 → 仓库简介）

```
自建项目管理：新需求先 Intake 分诊再进待办；可自定义工作项类型 / 工作流 / 自动化；AI 嵌在 Issue·Intake·Cycle，数据在自己机器上。
```

Topics（若支持）：`self-hosted` `project-management` `vue` `golang` `ai` `docker`

展示名尽量用 **Reqmango**（路径仍可为 `reqmanpy`）。

## 动态（粘贴到 GitCode 动态 / 讨论）

**标题：** 自建 PM 开源共建：先分诊再进 backlog，类型 / 工作流 / 自动化可配

**正文：**

```
很多团队熟悉「可定制项目管理」（工作项类型、工作流、自动化），也会卡在云端成本和数据驻留，以及入口需求靠人肉分诊。

Reqmango 开源自建：
· Intake 分诊：类型 / 优先级 / 疑似重复建议，拿不准再进 backlog
· 自定义工作项类型 · 工作流 · 自动化
· AI 嵌在 Issue / Intake / Cycle；自备 Key，数据在自己机器上

一键试用：
git clone https://gitcode.com/yongfeng9m-/reqmanpy.git reqmango
cd reqmango && cp .env.example .env && docker compose up --build
→ http://localhost  · demo@example.com / demo1234

欢迎提：
「Jira 里这样配，这里还缺什么」——直接开 Issue。
新手任务：docs/dev/good-first-issues.md
贡献指南：CONTRIBUTING-zh.md
国际镜像：https://github.com/vinthuy/reqmango
```

## 批量开 Issue（PAT）

GitCode → 个人设置 → Access Tokens（scopes: **projects** + **issues**）：

```powershell
$env:GITCODE_TOKEN = '<你的 PAT>'
node scripts/create-gitcode-gfi-issues.mjs
```

脚本会创建欢迎 Issue + GFI-01～（与清单对齐），已存在同名标题会跳过。

## 渠道边界（避免分心）

| 平台 | 状态 |
|------|------|
| GitCode | **主场，持续** |
| 开源中国「投递软件」 | **已提交待审核**（成稿 `promo-oschina.md`） |
| V2EX / 掘金 | 已发，不追加 |
| 思否 / Linux.do | 暂缓 |

## 验收（你这边点开能看到）

- [ ] About 已是产品一句话  
- [ ] 有一条置顶/显眼的「欢迎共建」Issue  
- [ ] 至少 5 个带 `good first issue` 的 Issue  
- [ ] 近 7 天有一条动态或讨论指向试用命令  
