# Pipeline Status（功能管线状态）

最后更新：2026-06-21

---

## 活跃功能

| 功能 | 后端 | Spec | Design | Plan | Implement | Review | KB | 备注 |
|------|------|------|--------|------|-----------|--------|----|------|
| Cycle（周期） | Go | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 基础 CRUD + 状态流转 + 进度 + 燃尽图 |
| Module（模块） | Go | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | CRUD + 树形 + Issue 关联 + 统计 |
| CustomField | Go | ⏳ | - | - | - | - | - | 规划中 |
| Workflow | Go | ⏳ | - | - | - | - | - | 规划中 |

---

## Backlog

| 优先级 | 功能 | 描述 |
|--------|------|------|
| P1 | CustomField（自定义字段） | 7 种字段类型 + 选项 + 值历史 |
| P1 | Workflow + Automation | 状态流转 + 触发器/条件/动作 |
| P1 | IssueType（工作项类型） | 类型 CRUD + 图标/颜色/默认 |
| P2 | AI Assistant | 自然语言交互 + PQL 生成 |
| P2 | EstimatePoint | 估算点管理 |
| P2 | Comments | 评论（支持回复） |
| P2 | Notifications | 通知系统 |
| P2 | Attachments | 附件上传/下载 |

---

## 图例

| 标记 | 含义 |
|------|------|
| ✅ | 完成 |
| 🔄 | 进行中 |
| ⏳ | 待开始 |
| ❌ | 取消 |
| - | 不适用 |
