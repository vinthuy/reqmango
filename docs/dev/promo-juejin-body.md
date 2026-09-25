很多团队绕不开 Jira：工作项类型、工作流、自动化都熟。但也会卡在几件事——云端/许可证成本、数据要放自己机器、入口需求还是靠人肉分诊。

Reqmango 是开源的**自建项目管理**，能力上尽量对齐大家熟悉的那套「可定制 PM」：

- **自定义工作项类型**：需求 / 缺陷 / 任务等按项目配字段与结构
- **自定义工作流**：状态流转、权限、审批按项目配
- **自动化规则**：触发器 → 条件 → 动作

入口多了一层 **Intake 分诊**：新需求先给类型 / 优先级 / 疑似重复建议，拿不准的再人工进 backlog。AI **嵌在工作路径里**（Issue / Intake / Cycle），不是另开一个聊天台；自备模型 Key，**数据与推理都在你自己的机器上**。

不是要你「扔掉 Jira 信仰」，而是：想自建、想 Docker 拉起来就能试、缺哪块能力就开 Issue 一起补——**用问题驱动共建**。

## 一键试用

```bash
git clone https://gitcode.com/yongfeng9m-/reqmanpy.git reqmango
cd reqmango
cp .env.example .env
docker compose up --build
```

浏览器打开 http://localhost  
账号：`demo@example.com` / `demo1234`  
（可选）.env 里配 `AI_API_KEY` 后重启，才能看到 AI 分诊建议。

## 演示

![Intake 提交 → 分诊队列](https://raw.githubusercontent.com/vinthuy/reqmango/master/docs/assets/demo.gif)

**一分钟走查：** 登录 → 项目设置 → 分诊 →「入口表单链接」提交需求 → 队列接受/拒绝。同页侧栏看 **工作项类型 / 工作流 / 自动化**。

## 欢迎这类反馈（请直接开 Issue）

- 「我们在 Jira 里这样配工作流 / 字段，这里还缺什么」
- 「自动化想做成 XX 触发，现在做不到」
- 「从 Jira 迁数据 / 映射类型时卡在哪」

国内主场：https://gitcode.com/yongfeng9m-/reqmanpy  
国际镜像：https://github.com/vinthuy/reqmango  
欢迎总览：https://github.com/vinthuy/reqmango/issues/13  
贡献指南：https://gitcode.com/yongfeng9m-/reqmanpy/blob/master/CONTRIBUTING-zh.md

MIT。小步 PR 最欢迎。
