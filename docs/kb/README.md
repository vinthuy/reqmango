# Knowledge Base（全量知识库）

KB 是 ReqManPy 系统的**唯一真相来源**，始终描述系统当前的实际状态。

## 维护原则

- **时效性**：当功能实现完成，对应的 KB 文档必须同步更新
- **准确性**：KB 描述的是代码实际做了什么，不是设计时打算做什么
- **简洁性**：KB 是参考文档，不是开发日志。保留核心信息，去除过程性内容
- **关联性**：KB 文档之间通过链接相互引用

## KB 更新时机

| 触发事件 | 更新范围 |
|----------|----------|
| 新功能完成 | 新增/更新 `kb/architecture/` 相关文档 |
| API 变更 | 更新 `api-conventions.md` |
| 数据模型变更 | 更新 `data-model.md` |
| 技术栈升级 | 更新 `tech-stack.md` |
| 目录结构调整 | 更新 `project-layout.md` |

## KB 文档清单

### 产品层
- [PRD.md](PRD.md) — 产品需求文档，描述产品功能和用户故事

### 架构层
- [architecture/README.md](architecture/README.md) — 架构文档索引
- [architecture/plane-enterprise-info-architecture.md](architecture/plane-enterprise-info-architecture.md) — Plane Enterprise 信息架构（对标参考）
- [architecture/tech-stack.md](architecture/tech-stack.md) — 技术栈总览
- [architecture/backend-go.md](architecture/backend-go.md) — Go 后端（主）
- [architecture/backend-python.md](architecture/backend-python.md) — Python 后端（遗留）
- [architecture/frontend.md](architecture/frontend.md) — Vue 3 前端
- [architecture/data-model.md](architecture/data-model.md) — 数据模型 ERD
- [architecture/api-conventions.md](architecture/api-conventions.md) — API 设计约定
- [architecture/project-layout.md](architecture/project-layout.md) — 项目目录结构

### 变更日志
- [changelog/README.md](changelog/README.md) — 按时间记录每次 KB 更新

## 对于 AI Agent

在开始任何开发工作前，先阅读相关 KB 文档以了解：
1. 现有模块的实现模式（handler → service → repository → model）
2. API 约定（URL 格式、错误处理、分页标准）
3. 前端组件和 store 的组织方式
