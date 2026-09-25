# 发帖成稿（国内平台）

> 主推 **GitCode**；GitHub 只作国际镜像一句带过。  
> **不投 Linux.do。** 优先国内开发者常去的平台。  
> 对外可用 **Jira 作大众参照**；**不要在公开帖里点名 Plane / 其它同类品牌**（商标与「对标叙事」风险，仓库里已 scrub）。  
> 对内可借鉴其推广结构（见下方「借鉴要点」），措辞改写成 Reqmango 自己的话。

## 渠道状态

| 平台 | 状态 | 入口 |
|------|------|------|
| V2EX · 分享创造 | **已发** + 楼主补充（Jira/类型/工作流/自动化 + 掘金链接） | https://www.v2ex.com/t/1244811 |
| 掘金 | **已发（审核中）** | https://juejin.cn/spost/7689019584239009811 |
| 开源中国 | 待登录后发 | https://www.oschina.net |
| 思否 SegmentFault | 待登录后发 | https://segmentfault.com/write |
| GitCode 项目动态 / README | 持续 | https://gitcode.com/yongfeng9m-/reqmanpy |
| 博客园 / CSDN | 可选（转载同文 + 原文链接） | — |

---

## 借鉴要点（内部，勿原样贴到帖子）

同类「现代自建 PM + AI」常见有效说法，映射到我们：

| 他们常强调的 | 我们对外怎么说 |
|--------------|----------------|
| AI 嵌在工作里，不是外挂聊天窗 | AI 嵌在 Issue / Intake / Cycle；分诊与分析走工作项路径 |
| Your AI / Your infrastructure | 数据在自己机器；自备 `AI_API_KEY`，不经过我们的云 |
| Docker 几分钟起 | 一键 `docker compose up` + demo 账号 |
| 用 Jira 作迁移心智 | 「熟悉类型 / 工作流 / 自动化的人」；缺能力请开 Issue 共建 |
| 可定制 PM 内核 | 自定义工作项类型、工作流、自动化规则（门禁能力） |

**不要写：**「对标某某」「某某平替」「clone of …」。内部对标可以，公开只讲自己的能力与共建。

---

## 标题（推荐）

用过 Jira？试试自建这一套：先分诊，再进待办

备选：
2. 自建需求系统：进来先分诊，拿不准的才进待办  
3. 开源项目管理（类型 / 工作流 / 自动化可配）：Docker 一键起，求共建  

---

## 正文（可直接粘贴）

很多团队绕不开 Jira：工作项类型、工作流、自动化都熟。但也会卡在几件事——云端/许可证成本、数据要放自己机器、入口需求还是靠人肉分诊。

Reqmango 是开源的**自建项目管理**，能力上尽量对齐大家熟悉的那套「可定制 PM」：

- **自定义工作项类型**：需求 / 缺陷 / 任务等按项目配字段与结构  
- **自定义工作流**：状态流转、权限、审批按项目配  
- **自动化规则**：触发器 → 条件 → 动作  

入口多了一层 **Intake 分诊**：新需求先给类型 / 优先级 / 疑似重复建议，拿不准的再人工进 backlog。AI **嵌在工作路径里**（Issue / Intake / Cycle），不是另开一个聊天台；自备模型 Key，**数据与推理都在你自己的机器上**。

不是要你「扔掉 Jira 信仰」，而是：想自建、想 Docker 拉起来就能试、缺哪块能力就开 Issue 一起补——**用问题驱动共建**。

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

**一分钟走查：** 登录 → 项目设置 → 分诊 →「入口表单链接」提交需求 → 队列接受/拒绝。同页侧栏看 **工作项类型 / 工作流 / 自动化**。

**特别欢迎这类反馈（请直接开 Issue）：**

- 「我们在 Jira 里这样配工作流 / 字段，这里还缺什么」  
- 「自动化想做成 XX 触发，现在做不到」  
- 「从 Jira 迁数据 / 映射类型时卡在哪」  

国内主场（Issue / PR 优先）：https://gitcode.com/yongfeng9m-/reqmanpy  
国际镜像：https://github.com/vinthuy/reqmango  
欢迎总览：https://github.com/vinthuy/reqmango/issues/13  
贡献指南：https://gitcode.com/yongfeng9m-/reqmanpy/blob/master/CONTRIBUTING-zh.md  

MIT。小步 PR 最欢迎。维护者会自己啃清单，不是只丢任务。

---

## 发帖检查

- [x] 动图已上传（README + raw 链接）
- [x] V2EX 已发（旧文；新渠道用本版）
- [ ] 下一篇优先 **掘金**（标签：开源 / Jira / 项目管理 / Docker）
- [ ] 克隆地址写 GitCode
- [ ] 提 Jira 只作参照，不贬损、不写 clone
- [ ] 明确邀请：缺能力 → 开 Issue 共建
