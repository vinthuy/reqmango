# GitCode 欢迎共建

自建项目管理：**新需求先 Intake 分诊，再进 backlog**。

日常可定制能力：

- 自定义工作项类型（字段与结构）
- 自定义工作流（流转 / 权限 / 审批）
- 自动化规则（触发 → 条件 → 动作）

AI 嵌在 Issue / Intake / Cycle；自备模型 Key，数据在自己的机器上。

## 试用

```bash
git clone https://gitcode.com/yongfeng9m-/reqmanpy.git reqmango
cd reqmango
cp .env.example .env
docker compose up --build
```

浏览器 http://localhost · `demo@example.com` / `demo1234`

演示 GIF：见仓库 `docs/assets/demo.gif`

## 欢迎提这类问题（直接开 Issue）

- 「我们在 Jira 里这样配工作流 / 字段，这里还缺什么」
- 「自动化想做成 XX 触发，现在做不到」
- 「从其它工具迁数据 / 映射类型时卡在哪」

## 文档

- 贡献指南：`CONTRIBUTING-zh.md`
- Good first issues：`docs/dev/good-first-issues.md`
- 国际镜像：https://github.com/vinthuy/reqmango
