# KB Changelog（知识库变更日志）

---

## 2026-06-26 — AI 智能助手实现 + 知识库全量同步

**类型**: Feature + Documentation

**变更内容**:
- Task 20-23: AI Infrastructure (LLM Client + AIService + AIHandler + 前端)
- LLM Client 双协议支持 (Anthropic + OpenAI-compatible/DeepSeek)
- 17 个 AI Tool Functions 映射到已有 API
- AIChatSidebar (SSE 流式对话) + AICreateDialog (NL→工作项)
- DeepSeek 为默认 Provider
- 知识库全量同步：架构文档、数据模型、前端架构更新至实际状态
- 清理已过时的 "Python 后端独有模型" 列表
- 清理已过时的模块状态表

**影响文件**:
- 新增 10 个 AI 文件 (model/ai_config.go, service/llm_client.go, service/ai_service.go, handler/ai_handler.go, types/ai.ts, api/ai.ts, composables/useAI.ts, components/AIChatSidebar.vue, components/AICreateDialog.vue)
- 修改 4 个现有文件 (config.go, router.go, main.go, Project.vue)
- KB 全面更新: architecture/README.md, data-model.md, frontend.md

---

## 2026-06-25 — 工作项管理关键功能补全 (5 Features)

**类型**: Feature

**变更内容**:
- Feature G: 统一 Project Settings (4→12 tabs)
- Feature F: Notification 后端
- Feature A: Saved Views 保存视图
- Feature C: 项目级 Issue Type 配置
- Feature B: Pages/Wiki 文档系统

**KB 更新**: architecture/README.md, data-model.md, frontend.md + 3 新设计文档

---

## 2026-06-25 — ProjectTemplateManager 集成修复

**类型**: Fix

将孤儿组件 ProjectTemplateManager.vue 集成到 WorkspaceSettings。

---

## 2026-06-21 — 文档体系重组

**类型**: 文档重组

建立 kb/dev/superseded 三层架构，从 TECH_ARCHITECTURE.md 拆分 7 个专题文档。
