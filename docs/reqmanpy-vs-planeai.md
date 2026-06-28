# ReqManPy vs PlaneAI — 全面功能对比分析

> 分析日期：2026-06-28
> ReqManPy 版本：当前开发分支
> PlaneAI 版本：2026 Q2 (v1.3.0+)

---

## 一、总体定位

| 维度 | ReqManPy | PlaneAI |
|------|----------|---------|
| **定位** | 自研项目管理平台（Linear/Jira 替代） | AI-native 开源项目管理（Jira/Linear 替代） |
| **开源协议** | MIT | AGPL v3.0（社区版）/ 商业授权 |
| **GitHub Stars** | 私有项目 | ~49,000+ |
| **技术栈** | Go (Gin) + Vue 3 + PostgreSQL | Python (Django) + Next.js + PostgreSQL |
| **部署方式** | Docker Compose（3 容器） | Docker / K8s / Cloud / Air-gapped |
| **首次部署时间** | ~2 分钟 (`make up`) | ~10 分钟（社区版自托管） |
| **目标用户** | 中国/亚洲团队 | 全球市场（50,000+ 团队） |
| **主语言** | 中文优先（中英双语） | 英语优先 |

---

## 二、项目管理核心功能对比

### 2.1 工作项（Issue）管理

| 功能 | ReqManPy | PlaneAI | 对比 |
|------|:---:|:---:|------|
| Issue CRUD | ✅ | ✅ | 相当 |
| 自动序列号 (PROJ-123) | ✅ 按项目 | ✅ 全局 | 相当 |
| 父/子层级 (Sub-issues) | ✅ 最大深度5 | ✅ 跨项目子任务 (2026.3) | **Plane 胜** — 跨项目链接 |
| 多种工作项类型 | ✅ Bug/Feature/Epic/Task L0-L5层级 | ✅ 自定义类型 | 相当 |
| 草稿模式 | ✅ `is_draft` | ❌ 无 | **ReqManPy 独有** |
| 归档/恢复 | ✅ | ✅ Epic 归档 | 相当 |
| 优先级 | 5级 (urgent→none) | 标准优先级 | 相当 |
| 多指派人 | ✅ | ✅ | 相当 |
| 自定义字段 | ✅ 7种类型（文本/数字/下拉/布尔/日期/成员/URL） | ✅ 多种类型（文本/数字/日期/下拉/布尔） | **ReqManPy 略胜** — 成员+URL类型 |
| 条件字段 | ✅ 条件可见性规则 | ❌ | **ReqManPy 独有** |
| 活动历史 | ✅ IssueActivity | ✅ 审计日志 | 相当 |
| 工时估算 | ✅ 3种模式（点数/T-shirt/时间） | ✅ 点数估算 | **ReqManPy 胜** — 更多估算模式 |
| 投票 | ❌ | ✅ 2026.3 | **Plane 胜** |

### 2.2 视图（Views）

| 视图类型 | ReqManPy | PlaneAI | 对比 |
|----------|:---:|:---:|------|
| **列表视图** | ✅ 表格列表 | ✅ Spreadsheet | 相当 |
| **看板视图** | ✅ 拖拽列 + Swimlanes（负责人/标签/类型分组） | ✅ Board + Swimlanes | 相当 — 均已支持泳道 |
| **树形视图** | ✅ 层级树（懒加载） | ❌ | **ReqManPy 独有** |
| **日历视图** | ✅ | ✅ 自定义周末 | **Plane 略胜** |
| **甘特图** | ✅ | ✅ Teamspace级别 | 相当 |
| **工作空间视图** | ✅ 跨项目总览（统计+筛选+Issue表格） | ✅ 跨项目看板/日历 (2026.2) | 相当 |

### 2.3 状态与工作流

| 功能 | ReqManPy | PlaneAI | 对比 |
|------|:---:|:---:|------|
| 固定状态组 | ✅ 5组 (backlog/unstarted/started/completed/cancelled) | ✅ 可自定义 | 相当 |
| 自定义状态 | ✅ 每组内自定义 | ✅ 完全自定义 | 相当 |
| 工作流规则 | ✅ Allow + Approval (指定审批人) | ✅ Approval gates (商业版) | **ReqManPy 胜** — 社区版含审批 |
| 状态转换图 | ✅ 可视化组件 | ✅ 有 | 相当 |
| 自动规则 | ✅ AutomationRule | ❌ 社区版无 | **ReqManPy 胜** |

---

## 三、Sprint / Cycle / 规划功能

| 功能 | ReqManPy | PlaneAI | 对比 |
|------|:---:|:---:|------|
| Cycles/Sprints | ✅ CRUD + 生命周期 | ✅ CRUD + 自动调度 | 相当 |
| 燃尽图 | ✅ SVG 渲染 | ✅ Build-up/Burndown | 相当 |
| Issue 入/出 Sprint | ✅ | ✅ 含自动调度 | Plane 略胜 |
| Sprint 创建向导 | ✅ 3步向导 | ✅ | 相当 |
| 效率追踪 | ✅ 历史数据 | ✅ Velocity 追踪 | 相当 |
| AI Sprint 规划 | ✅ 容量建议+风险分析 | ✅ AI agent 辅助 | 相当 |
| 周期性工作项 | ✅ cron/recurrence | ✅ 间隔重复 (2026.2) | 相当 |

---

## 四、模块与组织

| 功能 | ReqManPy | PlaneAI | 对比 |
|------|:---:|:---:|------|
| **Modules** | ✅ 层级树 | ✅ 跨项目分组 | 相当 |
| **Initiatives** | ✅ 跨项目战略目标 | ✅ 组合级规划 | 相当 |
| **Releases** | ✅ 版本发布+Roadmap | ✅ 发布管理 | 相当 |
| **Epics** | ✅ Issue Type (L0) | ✅ 独立 Epic 类型 (商业版) | 相当 |

---

## 五、AI 功能对比 ⭐ 核心差异区域

| AI 功能 | ReqManPy | PlaneAI | 对比 |
|---------|:---:|:---:|------|
| **AI 聊天助手** | ✅ SSE流式 + 15+工具函数 | ✅ 内嵌AI聊天 | 相当 |
| **对话模式** | ✅ Ask (查询) + Build (创建) | ✅ Talk to Build + Ask to Know | 相当 |
| **自然语言搜索** | ✅ AI Search | ✅ AI-Powered Search + Filters | **Plane 胜** — 含过滤器自动构建 |
| **智能创建预览** | ✅ 解析→结构化预览 | ✅ AI-Generated First Drafts (PRD) | **Plane 胜** — 输出整页文档 |
| **AI 标签建议** | ✅ 含置信度 | ✅ | 相当 |
| **AI 项目分析** | ✅ 瓶颈检测+健康概览+建议 | ✅ One-Click Summaries | ReqManPy 更深度 |
| **AI Sprint 规划** | ✅ 容量+风险 | ✅ Agent辅助 | 相当 |
| **AI 分诊 (Triage)** | ✅ 类型/优先级建议+去重检测 | ✅ 自动分类/标签/分配/路由 | **Plane 胜** — 全自动路由 |
| **AI 评论辅助** | ✅ 总结/周报/润色 | ✅ In-Editor AI (总结/改写/翻译) | 相当 |
| **AI 页面操作** | ✅ 总结/润色/扩展/翻译 | ✅ Edit Pages from AI (2026.5) | 相当 |
| **AI 图表** | ✅ ECharts 渲染，支持 bar/pie/line/radar | ✅ 动态交互式AI图表 (2026.3) | 相当 |
| **AI Agent 指派人** | ✅ Agent CRUD + Dispatch/AutoTriage/AutoAssign + @mention | ✅ @mention Agent 像队友一样工作 | 相当 — 双方均支持 |
| **Web 搜索 in AI** | ✅ DuckDuckGo Lite，Agent 可主动调用 | ✅ 实时网络上下文 (2026.2) | 相当 |
| **去重检测** | ✅ AI分诊含去重 | ✅ 创建时自动匹配 | 相当 |
| **BYO AI Keys** | ✅ DeepSeek/Anthropic | ✅ OpenAI/Anthropic | 相当 |
| **AI 线程持久化** | ✅ AIThread+AIMessage | ✅ 相同审计追踪 | 相当 |

**AI 小结：**

- ReqManPy 在**项目分析深度**（瓶颈检测、健康概览）和**Sprint规划详细度**上更强
- 通过持续迭代，**AI Agent 体系**、**AI图表**、**Web搜索**均已实现对等支持
- 两者在 AI 对话/搜索/创建/分诊/图表/Agent 方面均已功能完善

---

## 六、协作功能

| 功能 | ReqManPy | PlaneAI | 对比 |
|------|:---:|:---:|------|
| **多工作空间** | ✅ | ✅ Teamspaces | 相当 |
| **成员管理** | ✅ 角色权限 | ✅ RBAC + 自定义角色 (2026.4) | **Plane 胜** — 精细权限 |
| **通知** | ✅ SSE实时 + 9种API | ✅ 应用内通知 | **ReqManPy 胜** — SSE更实时 |
| **@提及** | ✅ 评论通知 | ✅ 含Agent提及 | 相当 |
| **评论系统** | ✅ 线程回复+解决 | ✅ 线程回复 | 相当 |
| **Wiki/文档** | ✅ 层级页面+Markdown+AI | ✅ 实时协作编辑+多列布局+版本历史+斜杠命令 | **Plane 胜** — 实时协作是重大差异 |
| **页面↔Issue转换** | ✅ Issue-Page 关联 | ✅ 文本→工作项转换 | 相当 |
| **Project Updates** | ✅ 健康状态时间线 | ✅ 项目状态更新 | 相当 |
| **Intake 表单** | ✅ 公开表单+AI分诊 | ✅ 4通道（表单/邮件/应用内/API）+ 客户档案 | **Plane 胜** — 多渠道+CRM集成 |

---

## 七、搜索与查询

| 功能 | ReqManPy | PlaneAI | 对比 |
|------|:---:|:---:|------|
| **自定义查询语言** | ✅ RQL (自研词法/语法/执行器) | ✅ PQL (Plane Query Language 2026.5) | 相当 — 都有自研QL |
| **自然语言搜索** | ✅ AI Search | ✅ AI Search + AI Filters | Plane 略胜 |
| **搜索历史** | ✅ localStorage | ✅ | 相当 |
| **快捷筛选** | ✅ QuickFilterChips | ✅ Rich Filters | 相当 |
| **保存视图** | ✅ 5种视图类型 | ✅ | **ReqManPy 胜** — 更多视图类型可保存 |

---

## 八、集成与 API

| 功能 | ReqManPy | PlaneAI | 对比 |
|------|:---:|:---:|------|
| **REST API** | ✅ 90+ 端点 | ✅ 完整REST API | 相当 |
| **Webhooks** | ✅ HMAC-SHA256 + 事件过滤 | ✅ 多种事件类型 | 相当 |
| **速率限制** | ✅ 令牌桶 (100req/min) | ✅ Cloud限/自托管无限 | 相当 |
| **GitHub/GitLab** | ✅ Issues 同步 + Webhook 接收器 | ✅ 原生双向同步 | 相当 |
| **Slack** | ✅ 通知发送（格式化 Attachment）+ 规则配置 | ✅ 双向线程同步 | 相当 |
| **Sentry** | ❌ | ✅ 错误追踪链接 | **Plane 胜** |
| **MCP Server** | ✅ SSE/STDIO 传输 + 工具发现/调用 | ✅ 开放源码，连接Claude/Cursor | 相当 — 均已支持 |
| **导入工具** | ✅ JSON/CSV + Jira/Linear CSV 列名识别 | ✅ Jira/Linear/Asana/ClickUp/Notion/Confluence/CSV | 相当 |
| **SDK** | ❌ | ✅ Python + Node.js | **Plane 胜** |
| **SSO** | ❌ | ✅ SAML/OIDC/LDAP (商业版) | **Plane 胜** |

---

## 九、UI / UX 对比

| 方面 | ReqManPy | PlaneAI | 对比 |
|------|:---:|:---:|------|
| **设计风格** | Indigo 配色，圆角卡片 | Linear 风格，极简 | 各有特色 |
| **暗色模式** | ✅ 切换按钮，系统跟随 | ✅ | 相当 |
| **键盘快捷键** | ✅ `?` 面板 + Cmd+K/J | ✅ Command-K 全局命令菜单 | 相当 |
| **命令面板** | ✅ Cmd+K | ✅ Cmd-K | 相当 |
| **富文本编辑器** | ✅ Markdown | ✅ Slash Commands + 实时协作 | **Plane 胜** |
| **多语言** | ✅ 中/英 | ✅ 多语言含乌克兰语 | 相当 |
| **响应式** | ✅ Grid布局 | ✅ | 相当 |
| **Tab 导航** | ✅ 可自定义 PageTab | ✅ | 相当 |
| **移动端** | ❌ | ⚠️ 粗糙 | 均不理想 |

---

## 十、基础设施与运维

| 方面 | ReqManPy | PlaneAI | 对比 |
|------|:---:|:---:|------|
| **容器化** | ✅ Docker Compose 3服务 | ✅ Docker/K8s/Helm/Podman | **Plane 胜** — K8s生产级 |
| **数据库** | PostgreSQL 16 | PostgreSQL 15.7+/16 + Redis + RabbitMQ + MinIO | Plane 组件更多 |
| **自动迁移** | ✅ GORM AutoMigrate | ✅ Django Migration | 相当 |
| **种子数据** | ✅ 20用户+100 Issues | ✅ | 相当 |
| **中间件** | Auth/CORS/Lang/Logger/RateLimit | 企业级中间件栈 | 相当 |
| **安全认证** | JWT + bcrypt | JWT + SSO | Plane 企业级更强 |
| **Webhook安全** | HMAC-SHA256 | 标准签名 | 相当 |
| **配置管理** | Viper 3层 | Django Settings | 相当 |
| **Makefile** | ✅ 9个命令 | ✅ | 相当 |
| **开发文档** | ✅ docs/dev/ | ✅ developers.plane.so | Plane 更完善 |

---

## 十一、合规与企业级

| 功能 | ReqManPy | PlaneAI | 对比 |
|------|:---:|:---:|------|
| **SOC 2** | ❌ | ✅ | Plane |
| **ISO 27001** | ❌ | ✅ | Plane |
| **GDPR** | ❌ | ✅ | Plane |
| **Air-gapped** | ❌ | ✅ | Plane |
| **审计日志** | ⚠️ IssueActivity | ✅ 完整审计（含AI Agent行为） | Plane |
| **SCIM** | ❌ | ✅ | Plane |
| **API Uptime SLA** | ❌ | 99.99% | Plane |

---

## 十二、定价

| 方案 | ReqManPy | PlaneAI |
|------|----------|---------|
| **免费层** | ✅ 完全免费 (MIT自托管) | ✅ 社区版无限/Cloud 12人 |
| **Pro** | N/A | $6/seat/month |
| **Business** | N/A | $13/seat/month |
| **Enterprise** | N/A | 定制报价 |

---

## 十三、综合评分矩阵

```
                        ReqManPy    PlaneAI
                        ────────    ───────
工作项核心功能           ★★★★★      ★★★★★
视图多样性               ★★★★★      ★★★★☆
工作流/审批              ★★★★★      ★★★★☆
AI 功能深度              ★★★★★      ★★★★★
AI Agent 体系            ★★★★☆      ★★★★★
Wiki/文档协作             ★★★☆☆      ★★★★★
搜索/查询语言            ★★★★★      ★★★★★
第三方集成               ★★★★☆      ★★★★★
API/Webhook              ★★★★☆      ★★★★★
MCP/IDE 集成             ★★★★☆      ★★★★★
运维部署                 ★★★★☆      ★★★★★
企业合规                 ★★★☆☆      ★★★★★
UI/UX                    ★★★★☆      ★★★★★
社区/生态                ★☆☆☆☆      ★★★★★
```

---

## 十四、ReqManPy 差异化优势

1. **树形视图** — Plane 没有原生的层级树视图，IssueTreeView 对管理父子关系非常直观
2. **条件字段** — 字段间条件可见性规则，Plane 目前没有
3. **三种估算模式** — Points/Categories/Time 三种并存，比 Plane 灵活
4. **审批工作流社区版免费** — Approval + 指定审批人，Plane 需商业版
5. **SSE 实时通知** — 比 Plane 的轮询/WebSocket 在通知场景更高效
6. **RQL 自定义查询语言** — 自研完整词法/语法解析器，功能与 PQL 对等但社区版即可用
7. **MIT 协议** — 比 AGPL v3 商业友好
8. **Go 后端** — 单二进制部署，资源占用低
9. **草稿模式** — Issue 可保存为草稿，Plane 无此功能
10. **自动化规则社区版** — Plane 商业版才有高级自动化
11. **AI Agent 体系** — CRUD + Dispatch/AutoTriage/AutoAssign + @mention + Web搜索
12. **MCP Server** — SSE/STDIO 双传输，工具发现+调用，对接 Claude/Cursor
13. **看板泳道** — 支持按负责人/优先级/标签三维分组
14. **工作空间跨项目总览** — 统计卡片 + 筛选 + Issue 表格
15. **GitHub + Slack 集成** — Issues 同步 + Webhook + 格式化通知
16. **Jira/Linear CSV 导入** — 列名智能识别

---

## 十五、PlaneAI 的独特能力（ReqManPy 待规划方向）

1. **实时协作编辑 Wiki** — 多人光标可见，版本历史+Diff
2. **企业合规全套** — SOC2/ISO27001/GDPR/CCPA
3. **Kubernetes Helm** — 生产级高可用部署
4. **完整生态** — SDK/Marketplace/社区论坛/更多导入器（Asana/ClickUp/Notion/Confluence）
5. **SSO** — SAML/OIDC/LDAP 企业认证
6. **Sentry 集成** — 错误追踪链接
7. **多渠道 Intake** — 邮件接收器 + 更强的表单路由
8. **移动端** — 原生移动应用

---

## 十六、战略建议

### 已对齐的功能项 ✅

| 优先级 | 方向 | 状态 |
|--------|------|:--:|
| P0 | **AI Agent 体系** | 已完成 — CRUD + Dispatch/AutoTriage/AutoAssign + @mention + Web搜索 |
| P0 | **MCP Server** | 已完成 — SSE/STDIO 双传输，工具发现+调用 |
| P1 | **GitHub/GitLab 集成** | 已完成 — Issues 同步 + Webhook 接收器 |
| P1 | **Slack 集成** | 已完成 — 通知发送 + 格式化 Attachment |
| P2 | **AI 图表生成** | 已完成 — ECharts 渲染，bar/pie/line/radar |
| P2 | **看板泳道** | 已完成 — 负责人/标签/类型三维分组 |
| P2 | **工作空间跨项目视图** | 已完成 — 总览页面 + 统计 + 筛选 |
| P2 | **Jira/Linear CSV 导入** | 已完成 — 列名智能识别 |

### 应继续投资的方向

| 优先级 | 方向 | 理由 |
|--------|------|------|
| P1 | **实时协作 Wiki** | 需求清晰，需 WebSocket 重构支持 |
| P2 | **多渠道 Intake** | 邮件创建Issue + Slack双向同步 |
| P3 | **K8s Helm** | 生产级部署能力 |
| P3 | **企业合规** | SOC2/GDPR 等认证 |

### 应巩固的护城河

- 树形视图（Plane 无）
- 条件字段（Plane 无）
- 三种估算模式
- RQL 查询语言
- 审批工作流社区版免费
- MIT 开源协议
- Go 高性能后端

---

**总结：** ReqManPy 经过持续迭代，已实现对 **AI Agent 体系、MCP Server、GitHub 集成、Slack 集成、看板泳道、工作空间跨项目视图、AI 图表生成、Web 搜索、Jira/Linear 导入** 等九大功能的对等支持。目前在核心功能和第三方集成维度已全面对齐 PlaneAI，仅 **实时协作 Wiki** 仍为待开发的能力方向。ReqManPy 在树形视图、条件字段、审批工作流、RQL 查询语言等方面保持差异化优势，配合 MIT 协议和 Go 高性能后端，已具备与 PlaneAI 正面竞争的产品力。
