# 发帖成稿（V2EX / Linux.do）

> 主推 **GitCode**；GitHub 只作镜像一句带过。发前请先有 `docs/assets/demo.gif`，没有动图就把「动图」改成「一分钟走查步骤」。

---

## 标题（三选一）

1. 自建需求系统：进来先分诊，拿不准的才进待办  
2. 开源了一套自建项目管理：Docker 一键起，欢迎领 good first issue  
3. 需求池太乱？我们做了「先分诊再进 backlog」的自建工具（求共建）

推荐用 **1**。

---

## 正文

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

**演示：**（此处插入 `demo.gif` 或贴走查：项目设置 → 请求分诊 → 提交 Intake → 队列里接受/拒绝）

**欢迎共建（新手任务已开好）：**

国内主场（Issue / PR 优先）：https://gitcode.com/yongfeng9m-/reqmanpy  
国际镜像：https://github.com/vinthuy/reqmango  

GitHub 上已挂 good first issue，例如：

- 录 Intake 演示 GIF：https://github.com/vinthuy/reqmango/issues/1  
- 筛选芯片 i18n：https://github.com/vinthuy/reqmango/issues/3  
- Compose 首启说明：https://github.com/vinthuy/reqmango/issues/9  

完整清单：https://github.com/vinthuy/reqmango/blob/master/docs/dev/good-first-issues.md  
贡献指南：https://gitcode.com/yongfeng9m-/reqmanpy/blob/master/CONTRIBUTING-zh.md  

MIT。小步 PR 最欢迎；Agent 控制台大重构请先开 Discussion，默认产品叙事不是那个。

维护者会自己啃清单（例如 TriagePanel 中英文已合进 master），不是只丢任务。

---

## 发帖检查

- [ ] 动图已上传或走查写清  
- [ ] 克隆地址是 GitCode  
- [ ] 只贴 2～3 个 Issue，不要贴整表  
- [ ] 不求 star、不对比喷竞品  
- [ ] 发完后在 GitCode/GitHub 各钉一条「欢迎领 GFI」的 Discussion 或置顶 Issue（可选）
