# {Feature Name} �?Completion Review

**完成日期**: {YYYY-MM-DD}
**分支**: {branch name}
**设计文档**: [design.md](design.md)
**实施计划**: [plan.md](plan.md)

---

## 验收标准 Checklist

- [ ] AC1: {验收条件}
- [ ] AC2: {验收条件}
- [ ] AC3: {验收条件}

## 技术验�?
- [ ] 后端所�?API 可正常调�?- [ ] 前端编译�?TypeScript 错误
- [ ] 手动冒烟测试通过
- [ ] 关键路径 curl 测试通过

## 偏离设计记录

> 记录实施过程中与原始设计的偏离及原因�?
| 偏离�?| 原因 | 影响 |
|--------|------|------|
| {描述} | {原因} | {影响} |

如无偏离，标注：**无偏离，按设计实现�?*

## 已知问题

- {问题描述}（如无则标注 "�?�?
---

## KB 更新清单

实施完成后，以下 KB 文档需要更新：

- [ ] `kb/architecture/data-model.md` �?新增 `{table_name}` �?- [ ] `kb/architecture/api-conventions.md` �?（如 API 约定有变�?- [ ] `kb/architecture/backend.md` �?（如架构有变化）
- [ ] `kb/architecture/frontend.md` �?（如前端架构有变化）
- [ ] `kb/changelog/README.md` �?添加变更条目

---

## 归档

- [ ] �?`dev/features/{date}-{slug}/` 移动�?`dev/archive/{date}-{slug}/`
- [ ] 更新 `dev/pipeline-status.md`
