# Knowledge Base（全量知识库）

KB 是 ReqManPy 系统的**唯一真相来源**，始终描述系统当前的实际状态。

**最后更新**: 2026-06-28

## 维护原则

- **时效性**：当功能实现完成，对应的 KB 文档必须同步更新
- **准确性**：KB 描述的是代码实际做了什么，不是设计时打算做什么
- **简洁性**：KB 是参考文档，不是开发日志。保留核心信息，去除过程性内容

## KB 更新时机

| 触发事件 | 更新范围 |
|----------|----------|
| 新功能完成 | 更新 `architecture/README.md` 模块表 + `data-model.md` 表清单 + `frontend.md` |
| API 变更 | 更新 `api-conventions.md` |
| 数据模型变更 | 更新 `data-model.md` |
| 技术栈升级 | 更新 `architecture/README.md` |

## KB 文档清单

### 产品层
- [PRD.md](PRD.md) — 产品需求文档

### 架构层
- [architecture/README.md](architecture/README.md) — 架构总览 + 模块状态
- [architecture/project-layout.md](architecture/project-layout.md) — 项目目录结构
- [architecture/backend-go.md](architecture/backend-go.md) — Go 后端架构（36 Model / 36 Service / 38 Handler / 85+ 端点，含 RBAC）
- [architecture/frontend.md](architecture/frontend.md) — 前端架构（18 views / 78 components / 35 API 模块）
- [architecture/data-model.md](architecture/data-model.md) — 数据模型（37 张表，含 RBAC）
- [architecture/api-conventions.md](architecture/api-conventions.md) — API 设计约定
- [architecture/tech-stack.md](architecture/tech-stack.md) — 技术栈详情

### 功能设计文档
- [architecture/saved-views-design.md](architecture/saved-views-design.md) — 保存视图
- [architecture/pages-design.md](architecture/pages-design.md) — 页面文档/Wiki
- [architecture/notification-design.md](architecture/notification-design.md) — 通知系统
- [architecture/type-hierarchy-template-design.md](architecture/type-hierarchy-template-design.md) — 类型层级与模板
- [architecture/relation-system-design.md](architecture/relation-system-design.md) — 关联类型系统

### 遗留
- [architecture/backend-python.md](architecture/backend-python.md) — Python 后端（已淘汰）

### 变更日志
- [changelog/README.md](changelog/README.md) — KB 更新历史
