# 自定义关联关系

**最后更新**: 2026-10-25

关联类型在工作空间上定义。一条 `issue_relations` 把两个工作项连起来。父子层级另外存在 `issues.parent_id` 上，和关联类型是两套数据。

## 数据

`relation_types`：`name`、`inward_name`、`outward_name`、`workspace_id`。

`issue_relations`：`issue_id`、`related_issue_id`、`relation_type_id`、`comment`（可空）。

## API

路径都在 `/api/v1` 下，并且要登录。创建、修改、删除关系类型还要 `settings:manage`。

```
POST   /relations/types?workspace_id=N
GET    /relations/types?workspace_id=N
PUT    /relations/types/:id
DELETE /relations/types/:id

POST   /issues/:issueId/relations
GET    /issues/:issueId/relations
DELETE /relations/:relationId
```

## 种子数据

`SeedRelationTypesForAllWorkspaces` 只在工作空间还没有任何关系类型时插入下面五种。已经有类型的工作空间整段跳过，不会按名称补缺。

| 名称 | inward | outward |
|------|--------|---------|
| 阻塞 | 被阻塞于 | 阻塞 |
| 关联 | 关联到 | 被关联于 |
| 重复 | 重复于 | 被重复于 |
| 父子 | 子任务 | 父任务 |
| 依赖 | 依赖于 | 被依赖于 |

## 和层级的关系

工作项树用 `parent_id` 和 `depth`。父级 `depth >= 5` 时不能再挂子项，子项 `depth = 父级 + 1`，所以 depth 是 0 到 5。错误文案写成 “Maximum hierarchy depth (6 levels) exceeded”。

「父子」「依赖」也是上面的关系类型，可以单独连一条 `issue_relations`。它们不代替 `parent_id`，也没有排程约束（没有 Starts Before / Finishes After 这种时间关系）。

疑似重复的检测是另一条路径：`GET` 或 `POST /projects/:projectId/issues/duplicate-check`，以及 `POST /issues/merge`。那不是 `relation_types` 里的「重复」。
