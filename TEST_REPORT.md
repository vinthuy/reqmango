# ReqManPy — 全面测试报告

**测试日期**: 2026-06-28  
**测试工程师**: AI Test Engineer  
**分支**: master  
**提交**: 1889680

---

## 执行摘要

对 ReqManPy 项目进行了全面测试。**整体质量良好**，API 通过率 96%，前后端编译均无错误。发现了 **1 个严重 Bug** 和 **3 个配置问题**，其中严重 Bug 已修复。

| 测试层级 | 测试数 | 通过 | 失败 | 通过率 |
|---|---|---|---|---|
| Go 后端单元测试 (common) | 60+ | 60+ | 0 | 100% |
| Go 后端单元测试 (i18n) | 12 | 12 | 0 | 100% |
| Go 后端单元测试 (rql) | 17 | 17 | 0 | 100% |
| 前端 Vitest 单元测试 | 25 | 25 | 0 | 100% |
| API 集成冒烟测试 | 53 | 51 | 2 | 96% |
| TypeScript 编译检查 | 1 | 1 | 0 | 100% |
| Go vet 代码检查 | 1 | 1 | 0 | 100% |
| Vite 生产构建 | 1 | 1 | 0 | 100% |
| **总计** | **170+** | **168+** | **2** | **98.8%** |

---

## 1. 发现的 Bug

### Bug #1 🔴 严重 — ErrAlreadyAssigned / ErrAlreadyLabelled 返回错误 HTTP 状态码

- **文件**: [error_codes.go](backend-go/internal/common/error_codes.go#L71)
- **问题**: `ErrAlreadyAssigned` ("ALREADY_ASSIGNED") 和 `ErrAlreadyLabelled` ("ALREADY_LABELLED") 返回 HTTP 500，但语义上应该是 409 Conflict（资源已存在/已分配）。
- **根因**: `HTTPStatus()` 方法中 409 分支未包含这两个错误码。
- **修复**: 已在 [error_codes.go](backend-go/internal/common/error_codes.go) 中添加 `e == ErrAlreadyAssigned || e == ErrAlreadyLabelled` 到 409 分支。
- **状态**: ✅ 已修复

### Bug #2 🟡 中 — Estimate Settings 返回 404 而非自动创建

- **API**: `GET /projects/:projectId/estimate-points/settings`
- **问题**: 当项目的估算设置不存在时，返回 404 "Estimate settings not found"，但应该自动创建默认设置。
- **影响**: 新项目首次访问估算设置时会报错。
- **状态**: ⚠️ 待修复 (未修改，需确认业务逻辑)

### Bug #3 🟡 中 — Vitest 配置包含 Playwright 测试文件

- **文件**: [vitest.config.ts](frontend/vitest.config.ts)
- **问题**: Vitest 配置 `include: ['tests/**/*.test.ts']` 将 Playwright 测试文件当作 Vitest 测试运行，导致 6 个测试文件全部失败。
- **根因**: `frontend/tests/` 目录中的 `.test.ts` 文件使用 `@playwright/test` 的 `test.describe()`，不兼容 Vitest。
- **修复**: 已将 Vitest 配置改为 `include: ['src/**/*.test.ts', 'src/**/*.spec.ts']`，并配置 jsdom 环境。
- **状态**: ✅ 已修复

### Bug #4 🟡 中 — `tests/` 目录与 `e2e/` 目录测试重复

- **文件**: `frontend/tests/` (6 个文件) vs `frontend/e2e/` (5 个文件)
- **问题**: 两个目录都包含 Playwright E2E 测试，结构重复。`tests/` 中的文件被 Vitest 误配置引用，但实际应通过 Playwright 运行。
- **建议**: 将 `tests/` 中的 Playwright 测试合并到 `e2e/` 目录，保留 `tests/` 给纯 API 集成测试。
- **状态**: ⚠️ 待清理

---

## 2. 测试覆盖详情

### 2.1 后端 Go 单元测试

#### common 包 — 100% 语句覆盖 ✅
| 测试文件 | 测试数 | 测试内容 |
|---|---|---|
| `error_codes_test.go` | 45 | 45 个错误码的 HTTPStatus 映射、NOT_FOUND 后缀检测、错误码唯一性 |
| `errors_test.go` | 32 | 错误构造器、JSON 序列化、RespondError/RespondOK/RespondCreated、语言回退 |
| `pagination_test.go` | 12 | 默认值、边界值、负数、非数字、上限、不同配置 |
| `constants_test.go` | 6 | 角色层次、优先级唯一性、状态组、DefaultStates 完整性、颜色格式 |

#### i18n 包 — 94.1% 语句覆盖 ✅
| 测试文件 | 测试数 | 测试内容 |
|---|---|---|
| `i18n_test.go` | 12 | T() 中文/英文翻译、未知 key 回退、未知语言回退、DetectLanguage 解析、所有错误码翻译完整度 |

#### rql 包 — 36.3% 语句覆盖 ✅
| 测试文件 | 测试数 | 测试内容 |
|---|---|---|
| `lexer_test.go` | 5 | Token 化: 比较/AND/OR/LIKE/IN/括号 |
| `parser_test.go` | 6 | AST 构建: 比较/AND/OR/LIKE/IN/嵌套 |
| `builder_test.go` | 6 | SQL 构建: 简单/Multiple/嵌套/Nil |

#### 未覆盖的包（需要数据库 Mock）
| 包 | 文件数 | 状态 | 原因 |
|---|---|---|---|
| `config` | 1 | 0% | 需要环境变量 mock |
| `handler` | 32 | 0% | 需要 Gin + DB mock |
| `service` | 40 | 0% | 需要 DB mock |
| `model` | 1 | 0% | GORM 模型，无需独立测试 |
| `middleware` | 4 | 0% | 需要 Gin + DB mock |
| `router` | 1 | 0% | 集成测试覆盖 |
| `seed` | 1 | 0% | 数据种子脚本 |

### 2.2 前端 TypeScript 单元测试

#### Auth Store — 11 测试 ✅
| 测试组 | 测试内容 |
|---|---|
| 初始状态 (4) | 空用户、空 token、未登录、token 持久化读取 |
| 登录 (2) | 成功登录设置状态、失败抛出异常 |
| 注册 (1) | 调用 API 但不自动登录 |
| 登出 (1) | 清除用户/token/localStorage |
| 获取用户 (3) | token 存在时获取、获取失败时登出、无 token 时跳过 |

#### API 模块 — 14 测试 ✅
| 测试组 | 测试内容 |
|---|---|
| 基础配置 (2) | baseURL、Content-Type header |
| Auth 拦截器 (2) | token 存在时添加 Bearer、无 token 时不添加 |
| 响应拦截器 (2) | 401 清除 token、307 重定向重试 |
| 模块结构 (8) | auth/issue/project/workspace/AI/cycle/notification/comment 模块导出完整性 |

### 2.3 API 集成冒烟测试

**53 个端点测试，51 通过 (96%)**

<details>
<summary>失败详情</summary>

| # | 端点 | 状态码 | 错误 |
|---|---|---|---|
| 1 | `GET /projects/:id/estimate-points/settings` | 404 | Estimate settings not found |
| 2 | `POST /projects/:id/ai/chart` | 500 | AI API key 无效 (DeepSeek 401) |

失败 #2 是预期行为 — `.env` 中 `DEEPSEEK_API_KEY=sk-test-key-for-infrastructure-testing` 不是有效密钥。

</details>

### 2.4 前端编译检查

| 检查项 | 结果 |
|---|---|
| TypeScript 类型检查 (`vue-tsc --noEmit`) | ✅ 通过，零错误 |
| Vite 生产构建 | ✅ 通过，318 模块，47 输出文件 |
| Go vet 代码检查 | ✅ 通过，零警告 |

### 2.5 Playwright E2E 测试

- **状态**: ⚠️ 未执行（需要 Docker 运行完整应用栈）
- **测试文件**: 5 个 spec 文件在 `e2e/` 目录下，覆盖认证、工作空间/项目、用户流程、AI、暗色模式/响应式
- **需要的命令**: `npx playwright test`（需要先启动后端 + 前端）

---

## 3. 代码质量评估

### 优势
- ✅ 前后端分离架构清晰，Go + Vue 3 技术栈成熟
- ✅ API 覆盖全面 (200+ 端点)，功能完整度极高
- ✅ TypeScript 类型系统完整，23 个类型定义文件
- ✅ 错误码体系完善 (45 个业务错误码)，支持中英文 i18n
- ✅ 前后端编译均无错误，代码质量良好
- ✅ API 集成测试通过率 96%

### 需改进
- ⚠️ 后端测试覆盖率严重不足：service 层 (40 个文件)、handler 层 (32 个文件) 完全没有单元测试
- ⚠️ 前端组件测试为零：70+ Vue 组件全部没有单元测试
- ⚠️ `tests/` vs `e2e/` 目录混乱，测试文件分类不清
- ⚠️ 缺少测试基础设施：数据库 mock、HTTP handler mock 未建立
- ⚠️ CI/CD 流水线缺失，无法自动运行测试
- ⚠️ API 集成测试使用固定 API key，AI 相关测试必然失败

---

## 4. 本次新增的测试文件

### 后端 (5 个新测试文件)
```
backend-go/internal/common/error_codes_test.go   — 45 测试
backend-go/internal/common/errors_test.go         — 32 测试
backend-go/internal/common/pagination_test.go     — 12 测试
backend-go/internal/common/constants_test.go       — 6 测试
backend-go/internal/i18n/i18n_test.go              — 12 测试
```

### 前端 (2 个新测试文件)
```
frontend/src/stores/auth.test.ts   — 11 测试
frontend/src/api/api.test.ts       — 14 测试
```

### 配置修复
```
frontend/vitest.config.ts   — 修复 include 路径，配置 jsdom 环境
```

---

## 5. 建议优先级

| 优先级 | 建议 | 预估工作量 |
|---|---|---|
| P0 | 修复 estimate settings 404 Bug | 30 分钟 |
| P0 | 清理 `tests/` 与 `e2e/` 目录重复 | 30 分钟 |
| P1 | 建立 handler/service 层测试基础设施 (mock DB + Gin) | 2-3 天 |
| P1 | 为核心 service (issue/project/cycle) 编写单元测试 | 3-5 天 |
| P1 | 建立 CI/CD 流水线 (GitHub Actions) | 1 天 |
| P2 | 为关键 Vue 组件编写单元测试 (@vue/test-utils) | 3-5 天 |
| P2 | 修复 AI API key 配置 (使用环境变量或测试专用 key) | 15 分钟 |
| P3 | 为剩余后端包 (middleware/config) 添加测试 | 1-2 天 |

---

*报告由 Claude Code 测试工程师生成*
