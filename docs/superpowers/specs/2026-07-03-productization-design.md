# reqmango v1.0 产品化设计文档

> 日期：2026-07-03
> 版本：1.0
> 状态：已确认

---

## 一、目标与范围

### 1.1 目标

将 reqmango 从"功能完整但工程粗糙"的内部项目提升为**可公开发布的 v1.0 产品**，系统性缩小与 Plane AI 的竞争力差距。

### 1.2 时间与团队

| 维度 | 约束 |
|------|------|
| **时间** | 8 周（~2 个月），2026-07 至 2026-08 |
| **团队** | 3 人，全栈能力 |
| **角色** | vinthuy（统筹 + 架构）、A（产品体验）、B（工程质量） |
| **总工作量** | 24 人周 |

### 1.3 精力配比

| 维度 | 占比 | 人周 | 说明 |
|------|:----:|:----:|------|
| **B 工程质量** | 40% | 9.5 | 测试、CI/CD、安全、日志、API 文档、性能 |
| **C 用户体验** | 30% | 7.0 | UI 一致性、i18n、Onboarding、响应式、编辑器 |
| **A 功能对齐** | 20% | 5.0 | 实时协作 Wiki、SSO、AI 增强、投票、Intake |
| **D 生态发布** | 10% | 2.5 | SDK、文档站、Landing Page、GitHub 公开 |

---

## 二、竞争力差距分析

### 2.1 差距总览

基于 `docs/reqmango-vs-planeai.md` 的 16 类 100+ 功能点逐项对比，识别出 **11 项需补齐的差距**：

#### 功能层面（Plane 胜出的功能）

| # | 差距项 | 严重度 | 工作量 | 对应 Sprint |
|---|--------|:------:|:------:|:-----------:|
| 1 | **实时协作 Wiki** — 多人光标、版本历史、实时同步 | 🔴 高 | 4 人周 | S2 + S3 |
| 2 | **SSO** — OIDC/SAML/LDAP 企业认证 | 🟡 中 | 2 人周 | S3 |
| 3 | **AI 搜索与过滤** — AI Filters 自动构建 | 🟡 中 | 0.5 人周 | S1 |
| 4 | **AI 分诊自动路由** — 全自动分类/分配 | 🟡 中 | 0.5 人周 | S2 |
| 5 | **投票功能** | 🟢 低 | 0.5 人周 | S3 |
| 6 | **多渠道 Intake** — 邮件创建 Issue | 🟢 低 | 1 人周 | S3 |

#### 工程层面（从"能跑"到"生产级"）

| # | 差距项 | 严重度 | 工作量 | 对应 Sprint |
|---|--------|:------:|:------:|:-----------:|
| 7 | **测试覆盖** — service 0%、handler 0%、前端组件 0% | 🔴 高 | 5 人周 | S1 + S2 |
| 8 | **CI/CD** — 无任何 CI 配置 | 🔴 高 | 1.5 人周 | S1 |
| 9 | **安全加固** — 未审计、无扫描 | 🟡 中 | 1.5 人周 | S2 |
| 10 | **API 文档** — 无 Swagger/OpenAPI | 🟢 低 | 1 人周 | S2 |

#### 生态层面

| # | 差距项 | 严重度 | 工作量 | 对应 Sprint |
|---|--------|:------:|:------:|:-----------:|
| 11 | **SDK 缺失** — Plane 有 Python + Node.js SDK | 🟢 低 | 1 人周 | S3 |

### 2.2 reqmango 的护城河（保持并强化）

以下差异化优势在本次产品化中**不做削弱、适度增强**：

- ✅ 树形视图（Plane 无）
- ✅ 条件字段（Plane 无）
- ✅ 三种估算模式（Points/Categories/Time）
- ✅ RQL 自研查询语言
- ✅ 审批工作流社区版免费（Plane 需商业版）
- ✅ MIT 开源协议（vs Plane AGPL v3.0）
- ✅ Go 高性能后端（单二进制，资源占用低）
- ✅ SSE 实时通知
- ✅ AI Agent 体系（CRUD + Dispatch + @mention + Web 搜索）
- ✅ MCP Server（SSE/STDIO 双传输）

### 2.3 是否对齐的决策

以下 Plane 能力**不纳入本次范围**（P3 或后续版本）：

- ❌ K8s Helm（P3，用户量上来再做）
- ❌ 企业合规 SOC2/ISO/GDPR（P3，需要外部审计）
- ❌ 移动端原生应用（P3，PWA 可暂代）
- ❌ Sentry 集成（P3，结构化日志后接入成本低）
- ❌ SDK 全量 API（仅核心资源，其余后续补）
- ❌ 实时协作 Wiki 完整 OT/CRDT（最小可用版本，2 人实时编辑即可）

---

## 三、执行方案：3-Sprint 并行

### 3.1 Sprint 节奏

```
Sprint 1 (W1-W3)          Sprint 2 (W4-W6)          Sprint 3 (W7-W8)
═══════════════           ═══════════════           ═══════════════
打基础                    磨体验                    亮功能
───────────────────────── ──────────────────────── ────────────────────────
vinthuy:                   vinthuy:                  vinthuy:
  架构审计 + Wiki方案         Wiki后端开发              SSO + GitHub发布
  + RQL增强 + 规范落地        + 分诊增强 + Review        + Docker + Release

A(产品):                   A(产品):                  A(产品):
  UX审计 + i18n + 筛选      UI统一 + Onboarding        Wiki前端 + Landing
  + 状态组件 + 响应式        + 暗色/键盘 + 编辑器       + 投票 + 文档站 + 细节

B(质量):                   B(质量):                  B(质量):
  测试框架 + CI/CD          安全 + 性能 + Swagger      E2E + 压测 + Bug收敛
  + DB迁移 + 测试第1批       + 日志 + 组件测试          + SDK + 安全报告
```

### 3.2 Sprint 边界机制

- **W3 末**：Sprint 1 验收 → 交叉 review → 产出可演示版本（make ci 通过）
- **W6 末**：Sprint 2 验收 → 交叉 review → 产出功能预览版（Wiki demo + Onboarding 可走通）
- **W8 末**：Sprint 3 终验 → 全量 review → git tag v1.0.0 公开发布

**交叉 review 规则**：
- A 审查 B 的测试/安全/CI 产出（产品视角验证质量）
- B 审查 A 的 UI/体验产出（工程视角验证实现）
- vinthuy 终审所有合并

---

## 四、详细任务清单

### 4.1 Sprint 1 (Week 1-3)：基础夯实

#### vinthuy — 统筹 + 架构（6 项，共 5 天）

| ID | 任务 | 时间 | 产出 |
|----|------|:----:|------|
| T1.1 | 架构审计与债务梳理 | W1 · 3d | 《架构审计报告》+ 技术债务 backlog |
| T1.2 | 实时协作 Wiki 技术方案 | W1-W2 · 3d | 选型验证 + PoC + 《实时协作技术设计文档》 |
| T1.3 | 代码规范 + Review 机制 | W1 · 1d | .golangci.yml + .eslintrc + pre-commit hook + Review 模板 |
| T1.4 | RQL AI 搜索增强 | W3 · 2d | NL→RQL 鲁棒性提升 + 搜索结果可微调 |
| T1.5 | Sprint 任务拆解 + 管理 | W1/W3 · 1d | 任务看板 + Sprint 1 验收 |
| T1.6 | make test/lint 基础设施 | W1 · 1d | `make ci` = lint + test + build |

#### A — 产品体验（5 项，共 7.5 天）

| ID | 任务 | 时间 | 产出 |
|----|------|:----:|------|
| A1.1 | 全站 UX 审计 | W1 · 3d | 逐页截图 + P0/P1/P2 分级清单 |
| A1.2 | i18n 缺失扫描与修复 | W1-W2 · 4d | 100% 双语覆盖（~200 条补全） |
| A1.3 | 筛选栏 Plane 风格 Phase 2 | W2 · 3d | QuickFilterChips + 持久化 + 快捷预设 |
| A1.4 | 状态组件标准化 | W2-W3 · 3d | EmptyState / ErrorState / LoadingSkeleton 三件套 |
| A1.5 | 响应式布局修复 | W3 · 2d | 320-1920px 全断点适配 |

#### B — 工程质量（6 项，共 8 天）

| ID | 任务 | 时间 | 产出 |
|----|------|:----:|------|
| B1.1 | Go 测试基础设施 | W1 · 3d | testify + sqlmock + testutil 包 + 2 个样板测试 |
| B1.2 | 前端测试基础设施 | W1 · 2d | @vue/test-utils + 2 个样板测试 |
| B1.3 | CI/CD Pipeline | W2 · 3d | GitHub Actions: lint → test → build → docker |
| B1.4 | DB Migration 版本化 | W2 · 2d | golang-migrate + 初始 schema + CI 检查 |
| B1.5 | Service 测试第 1 批 | W2-W3 · 4d | 8 个测试文件（issue/project/cycle/workspace/auth/workflow/notification/module） |
| B1.6 | Handler 集成测试框架 | W3 · 2d | httptest 框架 + auth/workspace 样板 |

**Sprint 1 检查点：**
- make ci 一键通过 ✅
- service 层覆盖率 ≥30%
- UX 审计报告已分级 ✅
- CI badge: passing ✅

---

### 4.2 Sprint 2 (Week 4-6)：体验 + 安全

#### vinthuy — 统筹 + 核心功能（3 项，共 6 天）

| ID | 任务 | 时间 | 产出 |
|----|------|:----:|------|
| T2.1 | 实时协作 Wiki 后端 | W4-W6 · 8d | WebSocket Hub/Room + OT 同步 + PageVersion 模型 |
| T2.2 | Code Review + Sprint 2 分配 | W4/W6 · 2d | Sprint 1 验收 + Sprint 2 拆解 |
| T2.3 | AI 分诊自动路由 | W6 · 2d | 一键执行 + 历史学习 + 可撤销 |

#### A — 产品体验（5 项，共 8 天）

| ID | 任务 | 时间 | 产出 |
|----|------|:----:|------|
| A2.1 | UI 组件库一致性 | W4-W5 · 5d | Design Token + 10 个标准组件重构 |
| A2.2 | Onboarding 流程 | W4-W5 · 4d | 3 步向导 + 空状态引导 + 示例数据 |
| A2.3 | 暗色模式 + 键盘增强 | W5 · 3d | 全站暗色审查 + Cmd+K 扩展 + `?` 面板 |
| A2.4 | 富文本编辑器优化 | W5-W6 · 3d | Slash commands + Markdown 粘贴 + @优化 + 图片拖拽 |
| A2.5 | 通知中心体验 | W6 · 1d | 分组/已读/设置面板 |

#### B — 工程质量（6 项，共 7.5 天）

| ID | 任务 | 时间 | 产出 |
|----|------|:----:|------|
| B2.1 | 安全加固 | W4-W5 · 4d | SQL注入/XSS/CSRF/JWT/API Key/密码/错误信息 |
| B2.2 | 性能优化 | W4-W5 · 3d | DB索引 + N+1消除 + 前端bundle + lazy load |
| B2.3 | API 文档 (Swagger) | W5-W6 · 3d | swaggo 集成 + 核心 50 API 注释 + /api/docs |
| B2.4 | 结构化日志 | W5 · 2d | zerolog/slog + traceID + 日志级别 + 脱敏 |
| B2.5 | 前端组件测试 | W6 · 2d | 6 个核心组件测试（FilterBar/Table/Board/Form/CmdK/Notification） |
| B2.6 | Service 测试第 2 批 | W6 · 2d | 5 个文件（ai/agent/page/automation/webhook），达 50% |

**Sprint 2 检查点：**
- UI 一致性达标 ✅
- Onboarding 可走通 ✅
- 安全扫描零高危 ✅
- /api/docs 可访问 ✅
- Wiki 协作 demo 可两人实时编辑 ✅

---

### 4.3 Sprint 3 (Week 7-8)：功能 + 发布

#### vinthuy — 统筹 + 发布（4 项，共 3.5 天）

| ID | 任务 | 时间 | 产出 |
|----|------|:----:|------|
| T3.1 | SSO OIDC 最小实现 | W7 · 3d | Google/GitHub OAuth + 用户绑定/解绑 |
| T3.2 | GitHub 公开仓库准备 | W7-W8 · 2d | README/CONTRIBUTING/SECURITY/CHANGELOG + 敏感信息审计 |
| T3.3 | Docker 部署最终化 | W8 · 1d | 一键部署验证 + 健康检查 + 部署文档 |
| T3.4 | 最终 Review + Release | W8 · 1d | 全量 review → git tag v1.0.0 |

#### A — 产品 + 生态（6 项，共 6.5 天）

| ID | 任务 | 时间 | 产出 |
|----|------|:----:|------|
| A3.1 | 实时协作 Wiki 前端 | W7 · 4d | 多人光标 + 实时同步 + 版本历史 Diff + 离线队列 |
| A3.2 | 投票功能 | W7 · 1d | Vote 模型 + UI + API |
| A3.3 | 多渠道 Intake 基础 | W7 · 2d | 邮件→Issue + 表单增强 |
| A3.4 | Landing Page | W7-W8 · 2d | 5 区块单页 + 中英双语 + 响应式 |
| A3.5 | 产品文档站 | W8 · 2d | VitePress + 快速开始 + 功能指南 + AI 文档 |
| A3.6 | 交互细节打磨 | W8 · 2d | 微动效 + Skeleton 替换 + 移动端手势 |

#### B — 质量 + 生态（6 项，共 5.5 天）

| ID | 任务 | 时间 | 产出 |
|----|------|:----:|------|
| B3.1 | E2E 测试 (Playwright) | W7 · 3d | 8 个核心流程 + CI 集成 |
| B3.2 | 压力测试 + 性能基线 | W7 · 2d | k6 压测 + TPS/P50/P95/P99 基线报告 |
| B3.3 | Bug 收敛 | W7-W8 · 3d | P0 清零 + P1 ≥80% |
| B3.4 | 安全测试报告 | W8 · 1d | OWASP + 依赖扫描 + Docker 扫描 |
| B3.5 | Go SDK 初版 | W8 · 1.5d | 核心 API 封装 + GoDoc + examples |
| B3.6 | TypeScript SDK 初版 | W8 · 1d | npm 包 + 完整类型 + README |

**Sprint 3 检查点（v1.0 发布标准）：**
- v1.0.0 git tag ✅
- CI 全绿 + E2E 全部通过 ✅
- Wiki 协作可用 ✅
- 文档站上线 ✅
- Landing Page 上线 ✅
- SDK (Go + TS) 发布 ✅
- Docker 一键部署验证通过 ✅

---

## 五、工作量配比验证

| 维度 | 目标占比 | 实际人周 | 实际占比 | 偏差 |
|------|:--------:|:--------:|:--------:|:----:|
| B 工程质量 | 40% | 9.5 | 39.6% | ✓ |
| C 用户体验 | 30% | 7.0 | 29.2% | ✓ |
| A 功能对齐 | 20% | 5.0 | 20.8% | ✓ |
| D 生态发布 | 10% | 2.5 | 10.4% | ✓ |
| **合计** | **100%** | **24.0** | **100%** | — |

---

## 六、风险与缓解

| 风险 | 概率 | 影响 | 缓解措施 |
|------|:----:|:----:|----------|
| **实时 Wiki 复杂度超预期** | 中 | 高（延期 1-2 周） | OT 最小实现，不做完整 CRDT；Sprint 2 仅后端，Sprint 3 前端可降级 |
| **安全审计发现严重漏洞** | 低 | 高（需额外修复时间） | Sprint 2 尽早开始安全扫描，预留 buffer |
| **新手磨合成本** | 中 | 中（前期效率低） | Sprint 1 前 2 天用作 ramp-up（搭建环境 + 读代码 + 跑通样板任务） |
| **功能蠕变** | 中 | 中（挤占质量时间） | vinthuy 作为 gatekeeper 审核所有"额外需求"；严格按优先级执行 |
| **E2E 环境不稳定** | 中 | 中（CI 误报） | 使用 Docker Compose 启动测试环境；E2E 失败自动 retry 2 次 |
| **依赖库安全漏洞** | 低 | 中 | Sprint 2/3 各执行一次 npm audit + go mod tidy -v |

---

## 七、不纳入范围（P3/后续）

以下项目明确**不纳入**本次 8 周产品化周期：

- K8s Helm Chart（等用户规模起来再投入）
- SOC2 / ISO 27001 / GDPR 合规认证
- 移动端原生 App（可用 PWA 响应式作为过渡）
- Sentry 集成（结构化日志后可快速接入，P3）
- SAML/LDAP（仅做 OIDC Google/GitHub）
- SDK 全量 API 覆盖（仅核心资源：Issue/Project/Cycle/Page）
- 实时协作 Wiki 的离线优先、冲突解决高级策略
- Marketplace / 插件系统

---

## 八、成功标准

### 8.1 v1.0 发布标准（MUST）

- [ ] `make ci` 一键通过（lint + test + build）
- [ ] Service 层测试覆盖率 ≥50%
- [ ] 8 个 E2E 核心流程全部通过
- [ ] CI/CD Pipeline 跑通（GitHub Actions）
- [ ] Swagger API 文档可访问（核心 50+ 端点）
- [ ] 安全扫描零高危漏洞
- [ ] i18n 中英双语 100% 覆盖
- [ ] Docker Compose 一键部署验证通过
- [ ] GitHub 仓库公开（README/Contributing/Security/Changelog 齐全）
- [ ] Landing Page + 文档站上线
- [ ] Go SDK + TypeScript SDK 初版发布

### 8.2 v1.0 体验标准（SHOULD）

- [ ] Onboarding 流程可完整走通（新用户 5 分钟内创建第一个 Issue）
- [ ] UI 组件一致性达标（10 个标准组件统一）
- [ ] 响应式布局 320-1920px 无严重断裂
- [ ] 键盘快捷键面板 + Cmd+K 命令面板完善
- [ ] Wiki 两人实时协作 demo 可用

### 8.3 竞争力对齐标准（NICE TO HAVE）

- [ ] SSO (Google/GitHub OAuth) 可用
- [ ] AI 分诊一键自动路由
- [ ] 投票功能
- [ ] 邮件→Issue 接收

---

## 九、附录

### A. 相关文档

- [reqmango vs PlaneAI 全面功能对比](../../reqmango-vs-planeai.md)
- [项目 PRD](../../kb/PRD.md)
- [后端架构](../../kb/architecture-backend.md)
- [开发管线](../../dev/)

### B. 当前技术状态基线

| 指标 | 当前值 |
|------|--------|
| Go 文件数 | 195 |
| Vue 文件数 | 100 |
| TypeScript 文件数 | 82 |
| API 端点 | 200+ |
| 数据库模型 | ~50 |
| Go 测试文件 | 8（仅 common/rql） |
| 测试覆盖率 | 0%（service/handler） |
| CI/CD | 无 |
| WebSocket | 无 |
| Swagger | 无 |
| 结构化日志 | 无（使用 stdlib log） |
