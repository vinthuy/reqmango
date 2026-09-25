# 发帖成稿（V2EX / Linux.do）

> 主推 **GitCode**；GitHub 只作镜像一句带过。`docs/assets/demo.gif` 已就位，可直接贴走查 + 仓库链接（论坛若支持外链图，用 GitHub raw）。

---

## 标题（推荐）

自建需求系统：进来先分诊，拿不准的才进待办

备选：
2. 开源了一套自建项目管理：Docker 一键起，欢迎领 good first issue  
3. 需求池太乱？我们做了「先分诊再进 backlog」的自建工具（求共建）

---

## 正文（可直接粘贴）

团队里最烦的不是「缺看板」，是需求从群里/邮件/口头涌进来时，类型、优先级、是否重复全靠人肉分。

Reqmango 是开源的**自建项目管理**：新需求先走 Intake 分诊（类型 / 优先级 / 疑似重复建议），拿不准的再人工处理；AI 嵌在 Issue / Intake / Cycle 里，数据在自己的机器上。

**一键试用：**

```bash
git clone https://gitcode.com/yongfeng9m-/reqmanpy.git reqmango
cd reqmango
cp .env.example .env
docker compose up --build
```

浏览器打开 http://localhost  
账号：`demo@example.com` / `demo1234`  
（可选）`.env` 里配 `AI_API_KEY` 后重启，才能看到 AI 分诊建议。

**演示：** Intake 提交 → 分诊队列接受/拒绝  
https://raw.githubusercontent.com/vinthuy/reqmango/master/docs/assets/demo.gif

**一分钟走查：** 登录 → 项目设置 → 分诊 →「入口表单链接」提交一条需求 → 回到队列点接受/拒绝。

**欢迎共建（新手任务已开好）：**

国内主场（Issue / PR 优先）：https://gitcode.com/yongfeng9m-/reqmanpy  
国际镜像：https://github.com/vinthuy/reqmango  

GitHub 上已挂 good first issue，例如：

- 筛选芯片 i18n：https://github.com/vinthuy/reqmango/issues/3  
- Compose 首启说明：https://github.com/vinthuy/reqmango/issues/9  
- 欢迎总览（置顶）：https://github.com/vinthuy/reqmango/issues/13  

完整清单：https://github.com/vinthuy/reqmango/blob/master/docs/dev/good-first-issues.md  
贡献指南：https://gitcode.com/yongfeng9m-/reqmanpy/blob/master/CONTRIBUTING-zh.md  

MIT。小步 PR 最欢迎；Agent 控制台大重构请先开 Discussion，默认产品叙事不是那个。

维护者会自己啃清单（例如 TriagePanel 中英文已合进 master），不是只丢任务。

---

## 发帖检查

- [x] 动图已上传（README + raw 链接）
- [ ] 克隆地址是 GitCode
- [ ] 只贴 2～3 个 Issue，不要贴整表
- [ ] 不求 star、不对比喷竞品
- [ ] 发完后在 GitCode 钉一条「欢迎领 GFI」（token 就绪后跑 `scripts/create-gitcode-gfi-issues.mjs`）
