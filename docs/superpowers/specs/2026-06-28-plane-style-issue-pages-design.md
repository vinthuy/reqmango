# Issue 页面设计优化

> 日期: 2026-06-28
> 范围: IssueCreate, IssueDetail, IssueList, IssueDetailPanel

---

## 设计目标

将 4 个核心 Issue 页面的视觉风格对齐现代项目管理平台设计趋势，遵循：
- **"结构可感知，不可见"** — 减少视觉分隔线，用留白代替硬边框
- **"不争夺注意力"** — 导航和容器后退，内容成为焦点
- **中性色调** — 从 indigo-600 重型配色转向灰色中性调，以蓝色 `#3b82f1` 为微妙强调色
- **扁平化表面** — 减少白色卡片 + 阴影模式，使用统一的表面和细分割线

---

## 变更概览

### 1. IssueList.vue — 列表/表格视图

| 元素 | 当前 | 变更后 |
|------|------|--------|
| 容器 | `rounded-lg border border-gray-200` | `rounded-xl border border-gray-100` |
| 工具栏按钮 | `border-gray-200 bg-indigo-600` (新建) | `bg-neutral-900` 主按钮，其他为 outline |
| 搜索栏 | `bg-gray-50 border-gray-200 focus:ring-indigo-500` | `bg-transparent border-gray-100 focus:ring-blue-500` |
| 表头 | `bg-gray-50 border-gray-200` | `bg-transparent border-gray-100`，更小字体 |
| 数据行 | `hover:bg-gray-50` | `hover:bg-gray-50/50` |
| 筛选芯片 | `bg-white border-gray-200` | `bg-gray-50 border-gray-100` |
| 批量操作栏 | `bg-indigo-50 border-indigo-200` | `bg-gray-50 border-gray-100` |
| 分页按钮 | `bg-indigo-600` active | `bg-neutral-900` active |
| 空状态 | 通用 | 更简洁的排版 |

### 2. IssueCreate.vue — 新建表单

| 元素 | 当前 | 变更后 |
|------|------|--------|
| 页面背景 | `bg-gray-50` | `bg-white` (统一表面) |
| 头部 | `bg-white border-b` 白色头 | 更紧凑的头部，更小的标题 |
| 表单区域 | `bg-white rounded-lg border p-6` | `border-r border-gray-100` 左右分割（扁平） |
| 属性面板 | `w-72 bg-white rounded-lg border p-6` | `w-64` 扁平右侧面板，section 间用分割线 |
| 类型选择器 | 硬边框、indigo-50 active | 灰色微妙边框、`bg-gray-100` active |
| 输入框 focus | `focus:ring-2 focus:ring-indigo-500` | `focus:ring-1 focus:ring-blue-400` |
| 底部按钮 | `bg-indigo-600` | `bg-neutral-900` 主按钮 |
| 按钮 | `rounded-lg` | `rounded-lg` (保持) |

### 3. IssueDetail.vue — 详情页

| 元素 | 当前 | 变更后 |
|------|------|--------|
| 页面背景 | `bg-gray-50` | `bg-white` |
| 头部 | 白色条 + 粗边框底部 | 紧凑头部，细边框 |
| 标题输入 | 边框0、白色卡片 | 无边框、无卡片包装 |
| 左侧面板 | 多个 `bg-white rounded-lg shadow-sm p-6` 卡片 | 统一表面，各部分间用 `border-b border-gray-100` 分隔 |
| 右侧属性 | 多个 `bg-white rounded-lg shadow-sm p-4` 卡片 | 单个面板，各部分间用细微分割线 |
| 保存按钮 | `bg-indigo-600` | `bg-neutral-900` |
| 分区标签 | `text-gray-700 font-medium` | `text-gray-400 font-medium text-xs uppercase` |

### 4. IssueDetailPanel.vue — 侧滑面板

| 元素 | 当前 | 变更后 |
|------|------|--------|
| 遮罩 | `bg-black/30` | `bg-black/20` (更微妙) |
| 头部编辑按钮 | `text-indigo-600 border-indigo-300` | `text-blue-600 border-blue-200` |
| 状态/优先级查看标签 | indigo 色 | 蓝色/灰色中性色 |
| 底部保存按钮 | `bg-indigo-600` | `bg-neutral-900` |
| 删除按钮 | `text-red-600 border-red-300` | `text-red-500 border-red-200` |

---

## 调色板

保持支持暗色模式：

```
亮色:
  primary: neutral-900 → gray-800
  accent:  blue-500 (#3b82f1)
  surface: white
  border:  gray-100 → gray-200
  muted:   gray-400 → gray-500

暗色:
  primary: gray-100
  accent:  blue-400
  surface: gray-900, gray-950
  border:  gray-800
  muted:   gray-500
```

---

## 不涉及的变更

- 不修改功能逻辑或 API 调用
- 不修改路由
- 不修改组件 props/emits 接口
- 不影响其他页面（Cycle、Module、Workspace 等）
- 暗色模式继续使用现有全局覆盖，仅调整被覆盖的 border/text 类

---

## 自审清单

1. **Placeholder 扫描**: 无 TBD/TODO
2. **内部一致性**: 所有 4 个文件遵循相同的色调和间距规则
3. **范围检查**: 仅限 4 个文件，无级联依赖
4. **歧义检查**: 所有颜色/边框/spacing 变更有明确的新旧对照