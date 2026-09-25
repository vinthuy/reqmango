# 推广策略（当前）

> **2026-09-26 起：主推 GitCode，其它渠道暂不大规模铺开。**  
> 已发的 V2EX / 掘金保留，不再主动加码；开源中国 / 思否等先不做。

## 主场

| | |
|---|---|
| **仓库** | https://gitcode.com/yongfeng9m-/reqmanpy |
| **产品句** | 自建项目管理：新需求先分诊再进待办；可自定义工作项类型 / 工作流 / 自动化；AI 嵌在 Issue / Intake / Cycle，数据在自己的机器上。 |
| **共建入口** | Issue + PR（中文优先）；欢迎懂 Jira 流程的朋友提缺口 |

## GitCode 上要持续做的事

1. **About / 简介** 与上表产品句一致（含类型 · 工作流 · 自动化）  
2. **置顶 / 欢迎 Issue**：说明试用方式 + 求共建问题类型  
3. **Good first issues**：与 `docs/dev/good-first-issues.md` 对齐（有 PAT 时跑 `scripts/create-gitcode-gfi-issues.mjs`）  
4. **有外部 PR / 有价值 Issue 时及时回复**，把对话留在本仓  

## 已发（存量，不追加轰炸）

| 平台 | 链接 |
|------|------|
| V2EX | https://www.v2ex.com/t/1244811 |
| 掘金（审核中） | https://juejin.cn/spost/7689019584239009811 |
| GitHub 欢迎 Issue | https://github.com/vinthuy/reqmango/issues/13 |

## 成稿

对外粘贴仍用 [promo-paste-v2ex.txt](./promo-paste-v2ex.txt) / [promo-juejin-body.md](./promo-juejin-body.md)，但**默认只在 GitCode 动态、README、Issue 里用**。大规模外站推广等你根据星标/Issue 反馈再开。

## GitCode PAT

创建 Issue 脚本：`scripts/create-gitcode-gfi-issues.mjs`  

```powershell
$env:GITCODE_TOKEN = '<Personal Settings → Access Tokens，scopes: projects + issues>'
node scripts/create-gitcode-gfi-issues.mjs
```
