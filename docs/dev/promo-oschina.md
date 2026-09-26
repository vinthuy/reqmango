# OSChina 投递材料（Reqmango）

> 路径：登录后打开 `https://my.oschina.net/<uid>/admin/publish` → **投递软件**（审核通过后再投递资讯）。  
> 主场仍写 GitCode；GitHub 作镜像一句带过。  
> **状态（2026-09-26）**：Reqmango 已提交「投递软件」，表单提交后已清空，等待编辑审核（通常约 1 个工作日）。

---

## A. 添加软件（优先做这一步）

| 字段 | 建议填写 |
|------|----------|
| **软件名称** | Reqmango |
| **中文名** | 自建项目管理（Intake 分诊） |
| **授权协议** | MIT |
| **开发语言** | Go / TypeScript / Vue |
| **操作系统** | 跨平台（Docker / Linux / Windows / macOS） |
| **软件官网 / 源码** | https://gitcode.com/yongfeng9m-/reqmanpy |
| **国际镜像** | https://github.com/vinthuy/reqmango |
| **一句话简介** | 自建项目管理：新需求先分诊再进待办；可自定义工作项类型、工作流与自动化；AI 嵌在 Issue / Intake / Cycle。 |

### 软件详细介绍（可粘贴）

Reqmango 是面向团队的**自建开源项目管理**工具。

**解决什么问题**

很多团队熟悉 Jira 一类工具的工作项类型、工作流与自动化，但也会卡在云端成本、数据驻留，以及入口需求仍靠人肉分诊。Reqmango 用 Docker 一键部署，数据留在自己的机器上。

**核心能力**

1. **Intake 分诊**：新需求先给出类型 / 优先级 / 疑似重复建议，拿不准的再人工进入 backlog。  
2. **自定义工作项类型**：按项目配置字段与结构。  
3. **自定义工作流**：状态流转、权限、审批按项目配置。  
4. **自动化规则**：触发器 → 条件 → 动作。  
5. **嵌在工作路径里的 AI**：Issue / Intake / Cycle；自备模型 API Key。

**快速试用**

```bash
git clone https://gitcode.com/yongfeng9m-/reqmanpy.git reqmango
cd reqmango
cp .env.example .env
docker compose up --build
```

浏览器打开 http://localhost  
演示账号：`demo@example.com` / `demo1234`  
（可选）配置 `AI_API_KEY` 后重启以启用 AI 分诊建议。

演示动图：https://raw.githubusercontent.com/vinthuy/reqmango/master/docs/assets/demo.gif

**参与共建**

欢迎熟悉 Jira 流程的开发者直接提缺口 Issue（工作流/字段/自动化/迁移）：  
https://gitcode.com/yongfeng9m-/reqmanpy

截图建议：仓库 `docs/assets/demo-0*.png` 或 `demo.gif`（Intake 表单 → 分诊队列）。

---

## B. 投递资讯（软件审核通过后再发）

**标题：** Reqmango：自建项目管理，先分诊再进待办（类型 / 工作流 / 自动化可配）

**分类：** 综合新闻 / 开源资讯（以页面选项为准）  
**关联软件：** Reqmango（审核通过后可选）  
**出处：** https://gitcode.com/yongfeng9m-/reqmanpy  

正文可用上面「软件详细介绍」缩略版 + 试用命令 + 共建链接。

---

## C. 博客备选（若有 my.oschina 博客）

标题与正文同掘金成稿：`docs/dev/promo-juejin-body.md`  
文末注明：讨论与 Issue 请到 GitCode 主场。
