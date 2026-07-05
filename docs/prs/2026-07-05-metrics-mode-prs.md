# 产品需求规格说明书（PRS）
## ReqMango 度量模式（Metrics Mode）

| 字段 | 内容 |
|------|------|
| 文档版本 | v1.0 |
| 创建日期 | 2026-07-05 |
| 产品名称 | ReqMango 项目管理平台 |
| 功能模块 | 度量模式（Metrics Mode） |
| 文档状态 | 待评审 |
| 对标产品 | 飞书项目度量模式 |

---

## 1. 产品概述

### 1.1 背景

ReqMango 当前具备基础报表能力（快速图表 + 自定义报表），但缺乏系统化的度量分析能力。参考飞书项目度量模式的设计，需要构建一个面向全角色（开发者、PM、SM、管理者）的数据分析中心，实现"无法度量就无法改进"的数据驱动理念。

### 1.2 目标

- 替代现有 ReportBuilder，成为项目内的唯一数据分析入口
- 提供开箱即用的预置度量模板，覆盖敏捷效能、项目管理、质量分析三大场景
- 支持图表配置增强（堆叠、参考线、双 Y 轴、同比环比）
- 增强 KPI 指标卡，支持同比箭头和辅助指标
- 渐进式交付，分三个阶段实施

### 1.3 目标用户

| 角色 | 核心需求 |
|------|---------|
| 开发者 | 个人效能、任务完成情况 |
| Scrum Master | 燃尽图、速率、WIP、累积流图 |
| 项目经理 | 需求吞吐量、里程碑达成率、资源负载 |
| QA | 缺陷趋势、缺陷密度、评审通过率 |
| 技术管理者 | 团队效能大盘、研发浓度、交付周期 |

### 1.4 术语定义

| 术语 | 定义 |
|------|------|
| 度量（Metrics） | 对项目数据进行定义、收集、分析的持续性定量化过程 |
| 维度（Dimension） | 数据的属性，决定图表分组方式，通常用于 X 轴 |
| 指标（Metric） | 量化衡量标准，通常用于 Y 轴 |
| 预置模板 | 系统内置的度量图表配置，开箱即用 |
| 同比 | 与去年同期对比（如 2026-06 vs 2025-06） |
| 环比 | 与上一周期对比（如本周 vs 上周） |
| 参考线 | 图表中的辅助基准线（平均值、目标值等） |

---

## 2. 功能需求

### 2.1 度量视图（Metrics View）

#### 2.1.1 页面结构

度量视图作为项目导航栏的一级入口，替代现有"报表"Tab。

**页面布局**：
```
┌─ 项目导航栏 ─────────────────────────────────────┐
│  ... | 工作项 | 视图 | 仪表板 | 度量 | 设置 |    │
└──────────────────────────────────────────────────┘

┌─ 度量视图 ───────────────────────────────────────┐
│  [预置模板] [我的图表] [新建图表]                   │
│                                                   │
│  ┌─ 预置模板区 ─────────────────────────────────┐ │
│  │  分类标签：敏捷效能 | 项目管理 | 质量分析       │ │
│  │  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐       │ │
│  │  │燃尽图│ │速率  │ │周期  │ │缺陷  │ ...    │ │
│  │  │      │ │趋势  │ │时间  │ │趋势  │       │ │
│  │  └──────┘ └──────┘ └──────┘ └──────┘       │ │
│  └──────────────────────────────────────────────┘ │
│                                                   │
│  ┌─ 我的图表区 ─────────────────────────────────┐ │
│  │  图表网格（支持拖拽排序）                       │ │
│  │  每个图表卡片包含：                             │ │
│  │  - 标题栏：名称 | 图表类型切换 | 编辑 | 删除   │ │
│  │  - 图表区域：渲染图表                          │ │
│  │  - 底部：最后更新时间                          │ │
│  └──────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────┘
```

#### 2.1.2 核心交互流程

**流程 1：使用预置模板**
1. 用户点击预置模板卡片
2. 弹出配置面板，预填模板默认值（维度、指标、筛选）
3. 用户可调整配置，实时预览图表
4. 点击"保存"→ 输入名称 → 创建为自定义图表

**流程 2：创建自定义图表**
1. 用户点击"新建图表"按钮
2. 弹出图表配置面板
3. 选择图表类型 → 配置 X 轴维度 → 配置 Y 轴指标 → 设置筛选条件
4. 可选配置：堆叠模式、参考线、数据标签、双 Y 轴
5. 实时预览 → 保存

**流程 3：管理图表**
- 拖拽排序：在"我的图表"区域拖拽调整顺序
- 编辑：点击编辑按钮打开配置面板
- 删除：点击删除按钮，二次确认后删除
- 全屏：点击全屏按钮进入演示模式

#### 2.1.3 功能规格

| 编号 | 功能 | 优先级 | 描述 |
|------|------|--------|------|
| MV-01 | 预置模板展示 | P0 | 按分类展示预置度量模板，支持一键使用 |
| MV-02 | 自定义图表 CRUD | P0 | 创建、查看、编辑、删除自定义度量图表 |
| MV-03 | 图表拖拽排序 | P1 | 支持拖拽调整图表顺序，持久化排序 |
| MV-04 | 图表类型切换 | P0 | 已创建的图表可切换图表类型（柱状→折线等） |
| MV-05 | 图表实时预览 | P0 | 配置过程中实时预览图表效果 |
| MV-06 | 全屏演示模式 | P2 | 图表全屏展示，适用于会议演示 |
| MV-07 | 数据导出 | P1 | 支持导出 CSV 和 PNG |

---

### 2.2 预置度量模板

#### 2.2.1 敏捷效能套件

| 模板 ID | 模板名称 | 图表类型 | X 轴 | Y 轴 | 适用角色 | 描述 |
|---------|---------|---------|------|------|---------|------|
| `agile_burndown` | 迭代燃尽图 | 折线图 | 迭代天数 | 剩余工作量 | 开发/SM | 展示迭代内剩余工作量随时间递减的趋势，预测完成日期 |
| `agile_velocity` | 速率趋势 | 柱状图 | 迭代 | 完成 Story Point | SM/PM | 展示每个迭代的完成速率，识别团队产能趋势 |
| `agile_cfd` | 累积流图 | 面积图（堆叠） | 时间 | 各状态 Issue 数 | SM/PM | 展示各状态 Issue 数量随时间的累积变化，识别瓶颈 |
| `agile_cycle_time` | 周期时间分布 | 柱状图 | 时间区间 | Issue 数 | SM/管理者 | 展示 Issue 从开始到完成的时间分布 |
| `agile_lead_time` | 前置时间趋势 | 折线图 | 时间 | 平均前置时间 | PM | 展示从需求提出到交付的平均前置时间趋势 |
| `agile_wip` | WIP 限制 | 柱状图 | 状态 | 进行中 Issue 数 | SM | 展示各状态的在制品数量，辅助 WIP 限制管理 |

#### 2.2.2 项目管理套件

| 模板 ID | 模板名称 | 图表类型 | X 轴 | Y 轴 | 适用角色 | 描述 |
|---------|---------|---------|------|------|---------|------|
| `pm_throughput` | 需求吞吐量 | 柱状图 | 周/月 | 完成需求数 | PM/管理者 | 展示单位时间内完成的需求数量 |
| `pm_milestone` | 里程碑达成率 | 指标卡 | — | 达成率% | PM/管理者 | 展示里程碑按时完成的百分比 |
| `pm_defect_trend` | 缺陷趋势 | 折线图 | 时间 | 新建/解决缺陷数 | QA/管理者 | 展示缺陷新建与解决的趋势对比 |
| `pm_priority` | 优先级分布 | 饼图 | 优先级 | Issue 数 | PM | 展示各优先级 Issue 的占比分布 |
| `pm_workload` | 资源负载 | 条形图 | 成员 | 分配 Issue 数 | PM | 展示每个成员的工作负载 |

#### 2.2.3 质量分析套件

| 模板 ID | 模板名称 | 图表类型 | X 轴 | Y 轴 | 适用角色 | 描述 |
|---------|---------|---------|------|------|---------|------|
| `qa_density` | 缺陷密度 | 柱状图 | 模块 | 缺陷数/需求数 | QA | 展示各模块的缺陷密度，识别高风险模块 |
| `qa_escape` | 缺陷逃逸率 | 指标卡 | — | 逃逸率% | QA/管理者 | 展示线上缺陷占总缺陷的比例 |
| `qa_review_pass` | 评审一次通过率 | 指标卡 | — | 通过率% | QA | 展示代码/需求评审一次通过的比例 |
| `qa_state_dwell` | 状态停留时间 | 条形图 | 状态 | 平均停留天数 | SM | 展示 Issue 在各状态的平均停留时间 |

#### 2.2.4 模板数据结构

```json
{
  "id": "agile_burndown",
  "category": "agile",
  "name_key": "metrics.template.burndown",
  "description_key": "metrics.template.burndown.desc",
  "chart_type": "line",
  "default_x_axis": "sprint_day",
  "default_y_axis": "remaining_points",
  "default_filters": {},
  "default_config": {
    "reference_lines": [{"type": "ideal", "label": "理想燃尽"}],
    "show_labels": true
  },
  "configurable": true,
  "icon": "flame"
}
```

---

### 2.3 图表配置增强

#### 2.3.1 堆叠模式

| 配置项 | 选项 | 适用图表 | 说明 |
|--------|------|---------|------|
| 堆叠模式 | 无 / 堆叠 / 百分比堆叠 | 柱状图、条形图、面积图 | 控制多数据集的堆叠展示方式 |

#### 2.3.2 参考线

| 参考线类型 | 说明 | 示例 |
|-----------|------|------|
| 常量值 | 用户指定固定数值 | 目标值 = 100 |
| 最大值 | 数据集最大值 | — |
| 最小值 | 数据集最小值 | — |
| 平均值 | 数据集平均值 | — |
| 中位数 | 数据集中位数 | — |
| 总值 | 数据集总和 | — |

**配置结构**：
```json
{
  "reference_lines": [
    {"type": "constant", "value": 100, "label": "目标值", "color": "#EF4444", "style": "dashed"},
    {"type": "average", "label": "平均值", "color": "#6B7280", "style": "dotted"}
  ]
}
```

#### 2.3.3 双 Y 轴

- 支持左右双 Y 轴，适用于不同量级数据的对比展示
- 示例：左轴显示 Issue 数量（柱状图），右轴显示完成率（折线图）

#### 2.3.4 数据标签

- 显示/隐藏数值标签
- 标签位置：图表上方、内部、下方
- 标签格式：原始值、百分比、带单位

#### 2.3.5 图例配置

- 显示位置：上、下、左、右
- 对齐方式：居中、左对齐、右对齐
- 显示/隐藏

---

### 2.4 KPI 指标卡

#### 2.4.1 卡片结构

```
┌─────────────────────────────────────┐
│  指标名称                            │
│  503            ↑ 12.3% vs 上周     │
│  ▓▓▓▓▓▓▓▓▓▓░░░░░ 75.2% 完成率      │
│  ─────────────────────────────────── │
│  进行中: 117    已完成: 103          │
└─────────────────────────────────────┘
```

#### 2.4.2 组成元素

| 元素 | 必选 | 说明 |
|------|------|------|
| 指标名称 | 是 | 指标的中文名称 |
| 主数值 | 是 | 大字体显示当前值 |
| 同比/环比指示器 | 否 | 箭头 + 百分比 + 对比周期标签 |
| 进度条 | 否 | 比率类指标的可视化进度条 |
| 辅助指标 | 否 | 底部 2-3 个辅助数字 |

#### 2.4.3 同比/环比指示器

- **箭头方向**：↑ 上升 / ↓ 下降 / — 持平
- **颜色编码**：绿色 = 正向变化 / 红色 = 负向变化 / 灰色 = 中性
- **对比标签**：显示对比周期（如"vs 上周"、"vs 上月"、"vs 去年同期"）

**变化方向判定**：
| 指标类型 | 上升为好（绿色） | 下降为好（绿色） |
|---------|----------------|----------------|
| 完成率、通过率、吞吐量 | ↑ 绿色 | ↓ 红色 |
| 缺陷数、缺陷率、周期时间 | ↓ 绿色 | ↑ 红色 |

---

### 2.5 同比/环比分析

#### 2.5.1 对比维度

| 对比类型 | 周期 | 说明 |
|---------|------|------|
| 周环比 | 本周 vs 上周 | 默认对比方式 |
| 月环比 | 本月 vs 上月 | — |
| 月同比 | 本月 vs 去年同月 | — |
| 自定义 | 用户指定基准周期 | 高级选项 |

#### 2.5.2 支持的指标

- Issue 总数
- 完成数
- 新建数
- 平均周期时间
- 平均前置时间
- 完成率
- 任何自定义 X+Y 组合的图表数据

---

### 2.6 数据导出

| 导出格式 | 说明 | 触发方式 |
|---------|------|---------|
| CSV | 图表原始数据表格 | 点击导出按钮 |
| PNG | 图表截图 | 点击导出按钮 |
| PDF | 图表 + 数据报告（Phase 3） | 点击导出按钮 |

---

## 3. 非功能需求

### 3.1 性能

| 指标 | 目标值 |
|------|--------|
| 图表渲染时间 | < 2 秒（1000 条 Issue 以内） |
| 页面首次加载 | < 3 秒 |
| API 响应时间 | < 500ms（P95） |
| 同比/环比计算 | < 1 秒 |

### 3.2 可用性

- 支持主流浏览器：Chrome、Firefox、Safari、Edge
- 响应式布局：适配 1280px 以上屏幕
- 无障碍：支持键盘导航（Phase 2）

### 3.3 数据安全

- 度量数据遵循项目权限控制
- 仅项目成员可查看项目度量数据
- 图表配置归创建者所有，项目管理员可管理所有图表

---

## 4. API 设计

### 4.1 预置模板

```
GET /api/v1/projects/:projectId/metrics/templates
→ {
    "categories": [
      {
        "id": "agile",
        "name": "敏捷效能",
        "templates": [
          {
            "id": "agile_burndown",
            "name": "迭代燃尽图",
            "chart_type": "line",
            "description": "展示迭代内剩余工作量趋势",
            "icon": "flame",
            "configurable": true
          }
        ]
      }
    ]
  }
```

### 4.2 自定义图表 CRUD

```
POST /api/v1/projects/:projectId/metrics/charts
→ Request: {
    "name": "我的燃尽图",
    "template_id": "agile_burndown",  // 可选
    "chart_type": "line",
    "x_axis": "sprint_day",
    "y_axis": "remaining_points",
    "filters": {},
    "config": {
      "reference_lines": [{"type": "average", "label": "平均值"}],
      "stack_mode": "none",
      "show_labels": true
    }
  }
→ Response: { "id": 1, "name": "我的燃尽图", ... }

GET /api/v1/projects/:projectId/metrics/charts
→ Response: { "charts": [...], "total": 5 }

PUT /api/v1/projects/:projectId/metrics/charts/:chartId
→ 更新图表配置

DELETE /api/v1/projects/:projectId/metrics/charts/:chartId
→ 删除图表
```

### 4.3 图表渲染

```
POST /api/v1/projects/:projectId/metrics/charts/:chartId/render
→ Response: {
    "labels": ["周一", "周二", ...],
    "datasets": [
      {
        "label": "剩余工作量",
        "data": [100, 85, 70, ...],
        "color": "#3B82F6"
      }
    ],
    "reference_lines": [
      {"type": "average", "value": 50, "label": "平均值"}
    ],
    "total": 503,
    "compare": {
      "current": 503,
      "previous": 480,
      "change_pct": 4.8,
      "direction": "up"
    }
  }
```

### 4.4 同比/环比

```
GET /api/v1/projects/:projectId/metrics/compare
  ?metric=count
  &period=week          // week | month
  &compare=prev_week    // prev_week | prev_month | same_month_last_year | custom
  &date_from=2026-06-01 // 自定义周期时使用
  &date_to=2026-06-30
→ Response: {
    "metric": "count",
    "current": { "value": 503, "period": "2026-W27" },
    "previous": { "value": 480, "period": "2026-W26" },
    "change_pct": 4.8,
    "direction": "up",
    "is_positive": true
  }
```

### 4.5 图表排序

```
POST /api/v1/projects/:projectId/metrics/charts/reorder
→ Request: { "chart_ids": [3, 1, 5, 2, 4] }
```

---

## 5. 数据模型

### 5.1 新增表：metric_charts

```sql
CREATE TABLE metric_charts (
    id            BIGSERIAL PRIMARY KEY,
    project_id    BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    creator_id    BIGINT NOT NULL REFERENCES users(id),
    name          VARCHAR(255) NOT NULL,
    template_id   VARCHAR(100),                    -- 关联预置模板 ID（NULL = 纯自定义）
    chart_type    VARCHAR(50) NOT NULL,             -- bar/line/pie/doughnut/area/table/...
    x_axis        VARCHAR(100) NOT NULL,            -- 维度字段
    y_axis        VARCHAR(100) NOT NULL,            -- 指标字段
    filters       JSONB DEFAULT '{}',               -- 筛选条件
    config        JSONB DEFAULT '{}',               -- 高级配置（堆叠/参考线/双轴等）
    sort_order    INT DEFAULT 0,                    -- 排序序号
    is_visible    BOOLEAN DEFAULT TRUE,             -- 是否可见
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_metric_charts_project ON metric_charts(project_id);
CREATE INDEX idx_metric_charts_creator ON metric_charts(creator_id);
```

### 5.2 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `template_id` | VARCHAR(100) | 可选。关联预置模板，用于标识图表来源 |
| `chart_type` | VARCHAR(50) | 图表类型：bar, line, pie, doughnut, area, radar, scatter, bubble, mixed, table |
| `x_axis` | VARCHAR(100) | 维度字段：state, priority, assignee, type, label, cycle, module, created_day, created_week, created_month, completed_day, completed_week, completed_month, updated_day, updated_week, updated_month |
| `y_axis` | VARCHAR(100) | 指标字段：count, avg_processing_time, current_retention, created_vs_resolved |
| `filters` | JSONB | 筛选条件，格式：`{"conditions": [{"field": "state", "operator": "=", "value": "Todo"}], "rql": ""}` |
| `config` | JSONB | 高级配置，格式：`{"stack_mode": "none", "reference_lines": [], "show_labels": false, "dual_y_axis": false, "legend_position": "top"}` |

---

## 6. 技术方案

### 6.1 架构概览

```
┌─ 前端 ──────────────────────────────────────────┐
│  MetricsView.vue                                 │
│  ├── MetricsTemplateGallery (预置模板展示)         │
│  ├── MetricsChartGrid (图表网格/拖拽)             │
│  ├── MetricsChartCard (单个图表卡片)              │
│  ├── MetricsChartConfig (图表配置面板)            │
│  ├── MetricsKPICard (KPI 指标卡)                 │
│  └── MetricsCompareIndicator (同比环比指示器)      │
└──────────────────────┬──────────────────────────┘
                       │ HTTP API
┌─ 后端 ───────────────┴──────────────────────────┐
│  MetricsHandler                                  │
│  ├── ListTemplates / GetTemplate                 │
│  ├── CRUD Charts                                 │
│  ├── RenderChart                                 │
│  ├── CompareMetrics                              │
│  └── ReorderCharts                               │
│                                                  │
│  MetricsService                                  │
│  ├── TemplateEngine (模板解析与默认值填充)          │
│  ├── ChartRenderer (图表数据计算与聚合)            │
│  ├── CompareEngine (同比环比计算)                  │
│  └── ConfigValidator (配置校验)                   │
│                                                  │
│  复用现有：RQLService, ReportService(V2)          │
└──────────────────────────────────────────────────┘
```

### 6.2 前端技术选型

| 技术 | 用途 | 备注 |
|------|------|------|
| Vue 3 + Composition API | 框架 | 复用现有 |
| Chart.js | 图表渲染 | 复用现有，增强配置 |
| dnd-kit / vuedraggable | 拖拽排序 | 新增依赖 |
| TypeScript | 类型安全 | 复用现有 |

### 6.3 后端技术选型

| 技术 | 用途 | 备注 |
|------|------|------|
| Go + Gin | Web 框架 | 复用现有 |
| GORM | ORM | 复用现有 |
| PostgreSQL + JSONB | 数据存储 | 复用现有，JSONB 存储配置 |

---

## 7. 实施计划

### Phase 1：核心度量能力（MVP）

**目标**：度量视图上线，替代 ReportBuilder

| 任务 | 说明 |
|------|------|
| 新增 `metric_charts` 表 | 数据库迁移 |
| 后端 Metrics API | CRUD + 渲染端点 |
| 前端度量视图 | 替代 ReportBuilder |
| 预置模板引擎 | 模板存储与默认值填充 |
| 图表拖拽排序 | 支持排序持久化 |
| 基础图表配置增强 | 堆叠模式、数据标签 |

**交付物**：度量视图可用，16 个预置模板可一键创建，图表可自定义配置

### Phase 2：分析增强

**目标**：同比环比 + 参考线 + KPI 指标卡

| 任务 | 说明 |
|------|------|
| 同比环比 API | 后端计算引擎 |
| KPI 指标卡增强 | 同比箭头、进度条、辅助指标 |
| 参考线功能 | 6 种参考线类型 |
| 双 Y 轴支持 | 左右双轴图表 |
| 数据导出增强 | CSV + PNG 导出 |

**交付物**：完整度量分析能力，支持同比环比对比

### Phase 3：高级能力

**目标**：全屏演示 + 多主题 + 数据迁移

| 任务 | 说明 |
|------|------|
| 全屏演示模式 | 适用于会议演示 |
| 多主题切换 | 深色/浅色/自定义主题 |
| 旧报表迁移工具 | Saved Reports → Metric Charts |
| Dashboard 数据源切换 | Widget 绑定度量 API |
| PDF 导出 | 图表 + 数据报告 |

**交付物**：完整度量平台，旧报表完全迁移

---

## 8. 验收标准

### 8.1 Phase 1 验收标准

| 编号 | 验收条件 | 验证方式 |
|------|---------|---------|
| AC-01 | 度量视图在项目导航栏可见且可访问 | 手动测试 |
| AC-02 | 16 个预置模板全部可正常展示和使用 | 手动测试 |
| AC-03 | 点击预置模板可弹出配置面板，预填默认值 | 手动测试 |
| AC-04 | 自定义图表支持完整的 CRUD 操作 | API 测试 |
| AC-05 | 图表可拖拽排序，刷新后排序保持 | 手动测试 |
| AC-06 | 图表类型可在已创建后切换 | 手动测试 |
| AC-07 | 堆叠模式（普通/百分比）正常工作 | 手动测试 |
| AC-08 | 图表渲染时间 < 2 秒（1000 条 Issue） | 性能测试 |
| AC-09 | ReportBuilder 入口从主导航移除 | 手动测试 |
| AC-10 | 已保存报表可一键迁移到度量图表 | 手动测试 |

### 8.2 Phase 2 验收标准

| 编号 | 验收条件 | 验证方式 |
|------|---------|---------|
| AC-11 | 同比/环比指示器正确显示变化方向和百分比 | 手动测试 |
| AC-12 | 6 种参考线类型均可正常渲染 | 手动测试 |
| AC-13 | KPI 指标卡显示主数值、同比、进度条、辅助指标 | 手动测试 |
| AC-14 | 双 Y 轴图表左右轴数据正确 | 手动测试 |
| AC-15 | CSV 和 PNG 导出功能正常 | 手动测试 |

---

## 9. 风险与依赖

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| Chart.js 不支持瀑布图/漏斗图 | 部分预置模板无法实现 | 使用 echarts 替代或自定义 Canvas 渲染 |
| 大数据量图表渲染性能 | 图表加载缓慢 | 后端聚合 + 前端虚拟滚动 + 缓存 |
| 旧报表迁移数据丢失 | 用户历史数据不可用 | 迁移前备份 + 迁移工具校验 |
| 拖拽排序在移动端体验差 | 移动端用户无法排序 | 移动端使用列表排序替代拖拽 |

---

## 10. 附录

### 10.1 维度字段完整列表

| 字段值 | 中文名 | 类型 | 分类 |
|--------|--------|------|------|
| `state` | 状态 | 分类 | 分类维度 |
| `priority` | 优先级 | 分类 | 分类维度 |
| `assignee` | 经办人 | 分类 | 分类维度 |
| `type` | 类型 | 分类 | 分类维度 |
| `label` | 标签 | 分类 | 分类维度 |
| `cycle` | 迭代 | 分类 | 分类维度 |
| `module` | 模块 | 分类 | 分类维度 |
| `state_group` | 状态组 | 分类 | 分类维度 |
| `created_by` | 创建人 | 分类 | 分类维度 |
| `created_day` | 创建日期 | 时间 | 时间维度 |
| `created_week` | 创建周 | 时间 | 时间维度 |
| `created_month` | 创建月份 | 时间 | 时间维度 |
| `completed_day` | 完成日期 | 时间 | 时间维度 |
| `completed_week` | 完成周 | 时间 | 时间维度 |
| `completed_month` | 完成月份 | 时间 | 时间维度 |
| `updated_day` | 更新日期 | 时间 | 时间维度 |
| `updated_week` | 更新周 | 时间 | 时间维度 |
| `updated_month` | 更新月份 | 时间 | 时间维度 |

### 10.2 指标字段完整列表

| 字段值 | 中文名 | 计算方式 | 适用图表 |
|--------|--------|---------|---------|
| `count` | Issue 数量 | `COUNT(*)` | 全部 |
| `avg_processing_time` | 平均处理时长 | `AVG(completed_at - created_at)` | 全部 |
| `current_retention` | 当前滞留时长 | `AVG(NOW() - created_at)` 仅未完成 | 全部 |
| `created_vs_resolved` | 创建 vs 解决 | 双数据集对比 | 折线图/柱状图 |

### 10.3 图表类型完整列表

| 类型 | Chart.js 类型 | 说明 |
|------|--------------|------|
| `bar` | bar | 柱状图 |
| `line` | line | 折线图 |
| `pie` | pie | 饼图 |
| `doughnut` | doughnut | 环形图 |
| `area` | line (fill) | 面积图 |
| `radar` | radar | 雷达图 |
| `scatter` | scatter | 散点图 |
| `bubble` | bubble | 气泡图 |
| `mixed` | bar+line | 混合图 |
| `table` | 自定义 | 数据表格 |
| `stacked_bar` | bar (stacked) | 堆叠柱状图 |
| `percent_stacked_bar` | bar (stacked, 100%) | 百分比堆叠柱状图 |
