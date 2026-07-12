# reqmango vs 同类产品 — 功能对比分析

> 分析日期：2026-06-28
> reqmango 版本：当前开发分支

---

## 一、总体定位

| 维度 | reqmango | 同类产品 |
|------|----------|---------|
| **定位** | 自研项目管理平台（Linear/Jira 替代） | AI-native 开源项目管理（Jira/Linear 替代） |
| **开源协议** | MIT | 开源协议（社区版）/ 商业授权 |
| **技术栈** | Go (Gin) + Vue 3 + PostgreSQL | Python + Next.js + PostgreSQL |
| **部署方式** | Docker Compose（3 容器） | Docker / K8s / Cloud / Air-gapped |
| **目标用户** | 中国/亚洲团队 | 全球市场 |
| **主语言** | 中文优先（中英双语） | 英语优先 |

---

## 二、项目管理核心功能对比

### 2.1 工作项（Issue）管理

| 功能 | reqmango | 对比 |
|------|:---:|------|
| Issue CRUD | ✅ | 相当 |
| 自动序列号 (PROJ-123) | ✅ 按项目 | 相当 |
| 父/子层级 (Sub-issues) | ✅ 最大深度5 | 相当 |
| 多种工作项类型 | ✅ Bug/Feature/Epic/Task L0-L5层级 | 相当 |
| 草稿模式 | ✅ `is_draft` | reqmango 独有 |
| 归档/恢复 | ✅ | 相当 |
| 优先级 | 5级 (urgent→none) | 相当 |
| 多指派人 | ✅ | 相当 |
| 自定义字段 | ✅ 7种类型（文本/数字/下拉/布尔/日期/成员/URL） | reqmango 略胜 — 成员+URL类型 |
| 条件字段 | ✅ 条件可见性规则 | reqmango 独有 |
| 活动历史 | ✅ IssueActivity | 相当 |
| 工时估算 | ✅ 3种模式（点数/T-shirt/时间） | reqmango 胜 — 更多估算模式 |

### 2.2 视图（Views）

| 视图类型 | reqmango | 对比 |
|----------|:---:|------|
| **列表视图** | ✅ 表格列表 | 相当 |
| **看板视图** | ✅ 拖拽列 + Swimlanes（负责人/标签/类型分组） | 相当 |
| **树形视图** | ✅ 层级树（懒加载） | reqmango 独有 |
| **日历视图** | ✅ | 相当 |
| **甘特图** | ✅ | 相当 |
| **工作空间视图** | ✅ 跨项目总览（统计+筛选+Issue表格） | 相当 |

### 2.3 状态与工作流

| 功能 | reqmango | 对比 |
|------|:---:|------|
| 固定状态组 | ✅ 5组 (backlog/unstarted/started/completed/cancelled) | 相当 |
| 自定义状态 | ✅ 每组内自定义 | 相当 |
| 工作流规则 | ✅ Allow + Approval (指定审批人) | reqmango 胜 — 社区版含审批 |
| 状态转换图 | ✅ 可视化组件 | 相当 |
| 自动规则 | ✅ AutomationRule | reqmango 胜 — 社区版即有 |

---

## 三、Sprint / Cycle / 规划功能

| 功能 | reqmango | 对比 |
|------|:---:|------|
| Cycles/Sprints | ✅ CRUD + 生命周期 | 相当 |
| 燃尽图 | ✅ SVG 渲染 | 相当 |
| Issue 入/出 Sprint | ✅ | 相当 |
| Sprint 创建向导 | ✅ 3步向导 | 相当 |
| 效率追踪 | ✅ 历史数据 | 相当 |
| AI Sprint 规划 | ✅ 容量建议+风险分析 | 相当 |
| 周期性工作项 | ✅ cron/recurrence | 相当 |

---

## 四、模块与组织

| 功能 | reqmango | 对比 |
|------|:---:|------|
| **Modules** | ✅ 层级树 | 相当 |
| **Initiatives** | ✅ 跨项目战略目标 | 相当 |
| **Releases** | ✅ 版本发布+Roadmap | 相当 |
| **Epics** | ✅ Issue Type (L0) | 相当 |

---

## 五、AI 功能对比

| AI 功能 | reqmango | 对比 |
|---------|:---:|------|
| **AI 聊天助手** | ✅ SSE流式 + 15+工具函数 | 相当 |
| **对话模式** | ✅ Ask (查询) + Build (创建) | 相当 |
| **自然语言搜索** | ✅ AI Search | 相当 |
| **智能创建预览** | ✅ 解析→结构化预览 | 相当 |
| **AI 标签建议** | ✅ 含置信度 | 相当 |
| **AI 项目分析** | ✅ 瓶颈检测+健康概览+建议 | reqmango 更深度 |
| **AI Sprint 规划** | ✅ 容量+风险 | 相当 |
| **AI 分诊 (Triage)** | ✅ 类型/优先级建议+去重检测 | 相当 |
| **AI 评论辅助** | ✅ 总结/周报/润色 | 相当 |
| **AI 页面操作** | ✅ 总结/润色/扩展/翻译 | 相当 |
| **AI 图表** | ✅ ECharts 渲染，支持 bar/pie/line/radar | 相当 |
| **AI Agent 指派人** | ✅ Agent CRUD + Dispatch/AutoTriage/AutoAssign + @mention | 相当 |
| **Web 搜索 in AI** | ✅ DuckDuckGo Lite，Agent 可主动调用 | 相当 |
| **去重检测** | ✅ AI分诊含去重 | 相当 |
| **BYO AI Keys** | ✅ DeepSeek/Anthropic | 相当 |
| **AI 线程持久化** | ✅ AIThread+AIMessage | 相当 |

---

## 六、协作功能

| 功能 | reqmango | 对比 |
|------|:---:|------|
| **多工作空间** | ✅ | 相当 |
| **成员管理** | ✅ 角色权限 | 相当 |
| **通知** | ✅ SSE实时 + 9种API | reqmango 胜 — SSE更实时 |
| **@提及** | ✅ 评论通知 | 相当 |
| **评论系统** | ✅ 线程回复+解决 | 相当 |
| **Wiki/文档** | ✅ 层级页面+Markdown+AI | 相当 |
| **页面↔Issue转换** | ✅ Issue-Page 关联 | 相当 |
| **Project Updates** | ✅ 健康状态时间线 | 相当 |
| **Intake 表单** | ✅ 公开表单+AI分诊 | 相当 |

---

## 七、搜索与查询

| 功能 | reqmango | 对比 |
|------|:---:|------|
| **自定义查询语言** | ✅ RQL (自研词法/语法/执行器) | 相当 |
| **自然语言搜索** | ✅ AI Search | 相当 |
| **搜索历史** | ✅ localStorage | 相当 |
| **快捷筛选** | ✅ QuickFilterChips | 相当 |
| **保存视图** | ✅ 5种视图类型 | reqmango 胜 — 更多视图类型可保存 |

---

## 八、集成与 API

| 功能 | reqmango | 对比 |
|------|:---:|------|
| **REST API** | ✅ 90+ 端点 | 相当 |
| **Webhooks** | ✅ HMAC-SHA256 + 事件过滤 | 相当 |
| **速率限制** | ✅ 令牌桶 (100req/min) | 相当 |
| **GitHub/GitLab** | ✅ Issues 同步 + Webhook 接收器 | 相当 |
| **Slack** | ✅ 通知发送（格式化 Attachment）+ 规则配置 | 相当 |
| **MCP Server** | ✅ SSE/STDIO 传输 + 工具发现/调用 | 相当 |
| **导入工具** | ✅ JSON/CSV + Jira/Linear CSV 列名识别 | 相当 |

---

## 九、UI / UX 对比

| 方面 | reqmango | 对比 |
|------|:---:|------|
| **设计风格** | Indigo 配色，圆角卡片 | 各有特色 |
| **暗色模式** | ✅ 切换按钮，系统跟随 | 相当 |
| **键盘快捷键** | ✅ `?` 面板 + Cmd+K/J | 相当 |
| **命令面板** | ✅ Cmd+K | 相当 |
| **富文本编辑器** | ✅ Markdown | 相当 |
| **多语言** | ✅ 中/英 | 相当 |
| **响应式** | ✅ Grid布局 | 相当 |
| **Tab 导航** | ✅ 可自定义 PageTab | 相当 |

---

## 十、基础设施与运维

| 方面 | reqmango | 对比 |
|------|:---:|------|
| **容器化** | ✅ Docker Compose 3服务 | 相当 |
| **数据库** | PostgreSQL 16 | 相当 |
| **自动迁移** | ✅ GORM AutoMigrate | 相当 |
| **种子数据** | ✅ 20用户+100 Issues | 相当 |
| **中间件** | Auth/CORS/Lang/Logger/RateLimit | 相当 |
| **安全认证** | JWT + bcrypt | 相当 |
| **Webhook安全** | HMAC-SHA256 | 相当 |
| **配置管理** | Viper 3层 | 相当 |
| **Makefile** | ✅ 9个命令 | 相当 |

---

## 十一、reqmango 差异化优势

1. **树形视图** — 原生的层级树视图，IssueTreeView 对管理父子关系非常直观
2. **条件字段** — 字段间条件可见性规则
3. **三种估算模式** — Points/Categories/Time 三种并存，更灵活
4. **审批工作流社区版免费** — Approval + 指定审批人
5. **SSE 实时通知** — 比轮询/WebSocket 在通知场景更高效
6. **RQL 自定义查询语言** — 自研完整词法/语法解析器
7. **MIT 协议** — 商业友好
8. **Go 后端** — 单二进制部署，资源占用低
9. **草稿模式** — Issue 可保存为草稿
10. **自动化规则社区版** — 高级自动化社区版即可用
11. **AI Agent 体系** — CRUD + Dispatch/AutoTriage/AutoAssign + @mention + Web搜索
12. **MCP Server** — SSE/STDIO 双传输，工具发现+调用
13. **看板泳道** — 支持按负责人/优先级/标签三维分组
14. **工作空间跨项目总览** — 统计卡片 + 筛选 + Issue 表格
15. **GitHub + Slack 集成** — Issues 同步 + Webhook + 格式化通知
16. **Jira/Linear CSV 导入** — 列名智能识别

---

**总结：** reqmango 经过持续迭代，已实现核心功能和第三方集成的全面对齐。目前在树形视图、条件字段、审批工作流、RQL 查询语言等方面保持差异化优势，配合 MIT 协议和 Go 高性能后端，已具备与同类产品竞争的产品力。