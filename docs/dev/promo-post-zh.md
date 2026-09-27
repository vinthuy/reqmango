# 推广策略（当前）

> **主推 GitCode**；开源中国做一次「添加软件」收录；V2EX/掘金已发不追加；思否等先不大规模铺。  
> 对外可用 **Jira 作大众参照**；公开帖不点名其它同类品牌。

## 渠道状态

| 平台 | 状态 | 入口 |
|------|------|------|
| GitCode 主场 | **持续** | https://gitcode.com/yongfeng9m-/reqmanpy · 操作清单：`docs/dev/promo-gitcode.md` |
| 开源中国 | **已投递软件，待审核** | 入口：`my.oschina.net/.../admin/publish` → 投递软件；成稿：`docs/dev/promo-oschina.md` |
| V2EX | 已发（存量）· 回帖已发 ✅ | https://www.v2ex.com/t/1244811 · 回帖草稿：`docs/dev/promo-reply-drafts.md` |
| 掘金 | 已发审核中（存量） | https://juejin.cn/spost/7689019584239009811 |
| 思否等 | 暂缓 | — |

## 里程碑（2026-09-27）

- **第一个外部贡献者 xiapdo**：完成 GFI-11「Compose 首启说明」（MR !1，修 Issue #9），已合并并同步两端；已在欢迎 Issue #10 公开致谢（@xiapdo）。
- **demo 登录修复**：种子数据新增 `SeedDemoLoginUser`，`demo@example.com / demo1234` 每次启动自动创建（新装 + 已部署库都可用）；README 同步更正了工作区/项目名。
- 回帖草稿统一放在：`docs/dev/promo-reply-drafts.md`。

## 成稿

对外粘贴仍用 [promo-paste-v2ex.txt](./promo-paste-v2ex.txt) / [promo-juejin-body.md](./promo-juejin-body.md)，但**默认只在 GitCode 动态、README、Issue 里用**。大规模外站推广等你根据星标/Issue 反馈再开。

## GitCode PAT

创建 Issue 脚本：`scripts/create-gitcode-gfi-issues.mjs`  

```powershell
$env:GITCODE_TOKEN = '<Personal Settings → Access Tokens，scopes: projects + issues>'
node scripts/create-gitcode-gfi-issues.mjs
```
