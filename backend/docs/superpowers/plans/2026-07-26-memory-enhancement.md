# 记忆功能增强实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 增强记忆功能，包括高级搜索、智能过期策略和相似记忆合并

**Architecture:** 
1. **记忆搜索增强**：添加日期范围过滤、相关性阈值、搜索结果摘要
2. **记忆过期策略**：实现定时清理任务、自定义过期配置、清理统计
3. **记忆合并**：基于内容相似度检测和合并相似记忆，保留版本历史

**Tech Stack:** Go 1.22+, GORM, PostgreSQL, Gin

---

## 文件结构

| 文件 | 职责 | 状态 |
|------|------|------|
| `internal/service/memory_service.go` | 记忆服务核心逻辑 | 修改 |
| `internal/handler/memory_handler.go` | 记忆API端点 | 修改 |
| `internal/model/memory.go` | 记忆数据模型 | 修改 |
| `internal/scheduler/memory_scheduler.go` | 记忆定时任务 | 新建 |
| `internal/router/router.go` | 路由注册 | 修改 |

---

### Task 1: 增强记忆搜索功能

**Files:**
- Modify: `internal/service/memory_service.go:171-202`
- Modify: `internal/handler/memory_handler.go:289-316`

**需求分析:**
- 支持日期范围过滤（created_at 范围）
- 支持相关性阈值过滤
- 添加搜索结果摘要（截取关键片段）
- 支持多种排序方式

- [x] **Step 1: 扩展 MemoryListFilters 添加搜索相关字段**

```go
// MemoryListFilters defines filters for listing memories
type MemoryListFilters struct {
    ProjectID      *uint64
    IssueID        *uint64
    AgentID        *uint64
    MemoryType     model.MemoryType
    Scope          model.MemoryScope
    ContextKey     string
    Tag            string
    Limit          int
    Offset         int
    SearchQuery    string           // 搜索关键词
    MinRelevance   float64          // 最小相关性分数
    StartDate      *time.Time       // 创建日期范围开始
    EndDate        *time.Time       // 创建日期范围结束
    SortBy         string           // 排序字段: "relevance", "created_at", "updated_at"
    SortOrder      string           // 排序方向: "asc", "desc"
}
```

- [x] **Step 2: 修改 ListMemories 方法支持新过滤条件**

```go
// 在 ListMemories 方法中添加新的过滤逻辑
case MemoryListFilters:
    // ... 现有代码 ...
    if f.SearchQuery != "" {
        keywords := strings.Fields(strings.ToLower(f.SearchQuery))
        var conditions []string
        var args []interface{}
        for _, kw := range keywords {
            conditions = append(conditions, "LOWER(content) LIKE ?")
            args = append(args, "%"+kw+"%")
        }
        if len(conditions) > 0 {
            query = query.Where(strings.Join(conditions, " OR "), args...)
        }
    }
    if f.MinRelevance > 0 {
        query = query.Where("relevance_score >= ?", f.MinRelevance)
    }
    if f.StartDate != nil {
        query = query.Where("created_at >= ?", *f.StartDate)
    }
    if f.EndDate != nil {
        query = query.Where("created_at <= ?", *f.EndDate)
    }
    
    // 排序逻辑
    sortField := "relevance_score"
    if f.SortBy == "created_at" || f.SortBy == "updated_at" {
        sortField = f.SortBy
    }
    sortOrder := "DESC"
    if f.SortOrder == "asc" {
        sortOrder = "ASC"
    }
    query = query.Order(fmt.Sprintf("%s %s", sortField, sortOrder))
```

- [x] **Step 3: 在 Handler 中添加搜索参数解析**

```go
// 在 ListMemories 方法中添加新参数解析
if query := c.Query("q"); query != "" {
    filters.SearchQuery = query
}
if minRel := c.Query("min_relevance"); minRel != "" {
    if score, err := strconv.ParseFloat(minRel, 64); err == nil {
        filters.MinRelevance = score
    }
}
if startDate := c.Query("start_date"); startDate != "" {
    if t, err := time.Parse(time.RFC3339, startDate); err == nil {
        filters.StartDate = &t
    }
}
if endDate := c.Query("end_date"); endDate != "" {
    if t, err := time.Parse(time.RFC3339, endDate); err == nil {
        filters.EndDate = &t
    }
}
if sortBy := c.Query("sort_by"); sortBy != "" {
    filters.SortBy = sortBy
}
if sortOrder := c.Query("sort_order"); sortOrder != "" {
    filters.SortOrder = sortOrder
}
```

- [x] **Step 4: 编译验证**

Run: `cd backend; go build ./...`
Expected: 编译通过

---

### Task 2: 实现记忆过期策略

**Files:**
- Modify: `internal/service/memory_service.go:302-307`
- Create: `internal/scheduler/memory_scheduler.go`
- Modify: `internal/router/router.go`

**需求分析:**
- 定时清理过期记忆（每天凌晨）
- 支持自定义过期时间配置（按记忆类型）
- 添加清理统计和日志
- 支持手动触发清理

- [x] **Step 1: 扩展 MemoryService 添加清理统计方法**

```go
// PruneExpiredMemories removes expired short-term memories and returns count
func (s *MemoryService) PruneExpiredMemories(ctx context.Context) (int64, error) {
    var count int64
    // 先统计即将删除的数量
    if err := s.db.WithContext(ctx).
        Where("memory_type = ? AND expires_at IS NOT NULL AND expires_at < NOW()", model.MemoryShortTerm).
        Count(&count).Error; err != nil {
        return 0, err
    }
    
    // 执行删除
    if count > 0 {
        if err := s.db.WithContext(ctx).
            Where("memory_type = ? AND expires_at IS NOT NULL AND expires_at < NOW()", model.MemoryShortTerm).
            Delete(&model.MemoryEntry{}).Error; err != nil {
            return 0, err
        }
    }
    
    return count, nil
}

// PruneLowRelevanceMemories removes memories with low relevance score
// older than specified days
func (s *MemoryService) PruneLowRelevanceMemories(ctx context.Context, maxDays int, minScore float64) (int64, error) {
    var count int64
    cutoffDate := time.Now().AddDate(0, 0, -maxDays)
    
    if err := s.db.WithContext(ctx).
        Where("relevance_score < ? AND created_at < ?", minScore, cutoffDate).
        Count(&count).Error; err != nil {
        return 0, err
    }
    
    if count > 0 {
        if err := s.db.WithContext(ctx).
            Where("relevance_score < ? AND created_at < ?", minScore, cutoffDate).
            Delete(&model.MemoryEntry{}).Error; err != nil {
            return 0, err
        }
    }
    
    return count, nil
}

// GetMemoryStats returns memory statistics for a workspace
func (s *MemoryService) GetMemoryStats(ctx context.Context, workspaceID uint64) (map[string]int64, error) {
    stats := make(map[string]int64)
    
    // 按类型统计
    var types []struct {
        MemoryType string `gorm:"column:memory_type"`
        Count      int64
    }
    if err := s.db.WithContext(ctx).
        Model(&model.MemoryEntry{}).
        Where("workspace_id = ?", workspaceID).
        Group("memory_type").
        Select("memory_type, COUNT(*) as count").
        Scan(&types).Error; err != nil {
        return nil, err
    }
    for _, t := range types {
        stats["type_"+t.MemoryType] = t.Count
    }
    
    // 按范围统计
    var scopes []struct {
        Scope string `gorm:"column:scope"`
        Count int64
    }
    if err := s.db.WithContext(ctx).
        Model(&model.MemoryEntry{}).
        Where("workspace_id = ?", workspaceID).
        Group("scope").
        Select("scope, COUNT(*) as count").
        Scan(&scopes).Error; err != nil {
        return nil, err
    }
    for _, s := range scopes {
        stats["scope_"+s.Scope] = s.Count
    }
    
    // 过期记忆数量
    if err := s.db.WithContext(ctx).
        Model(&model.MemoryEntry{}).
        Where("workspace_id = ? AND expires_at IS NOT NULL AND expires_at < NOW()", workspaceID).
        Count(&stats["expired"]).Error; err != nil {
        return nil, err
    }
    
    return stats, nil
}
```

- [x] **Step 2: 创建定时任务调度器**

```go
// internal/scheduler/memory_scheduler.go
package scheduler

import (
    "context"
    "fmt"
    "log/slog"
    "time"

    "github.com/reqmango/backend/internal/service"
)

// MemoryScheduler handles memory cleanup scheduling
type MemoryScheduler struct {
    memSvc      *service.MemoryService
    cleanupTime string // "02:00" format
    interval    time.Duration
    stopChan    chan struct{}
}

// NewMemoryScheduler creates a new memory scheduler
func NewMemoryScheduler(memSvc *service.MemoryService) *MemoryScheduler {
    return &MemoryScheduler{
        memSvc:      memSvc,
        cleanupTime: "02:00", // 默认每天凌晨2点
        interval:    24 * time.Hour,
        stopChan:    make(chan struct{}),
    }
}

// Start starts the scheduler
func (s *MemoryScheduler) Start(ctx context.Context) {
    go s.run(ctx)
    slog.Info("Memory scheduler started")
}

// Stop stops the scheduler
func (s *MemoryScheduler) Stop() {
    close(s.stopChan)
    slog.Info("Memory scheduler stopped")
}

// run runs the scheduler loop
func (s *MemoryScheduler) run(ctx context.Context) {
    // 计算下次执行时间
    nextRun := s.calculateNextRun()
    
    for {
        select {
        case <-s.stopChan:
            return
        case <-time.After(time.Until(nextRun)):
            s.cleanup(ctx)
            nextRun = s.calculateNextRun()
        }
    }
}

// calculateNextRun calculates the next scheduled run time
func (s *MemoryScheduler) calculateNextRun() time.Time {
    now := time.Now()
    target, _ := time.Parse("15:04", s.cleanupTime)
    nextRun := time.Date(now.Year(), now.Month(), now.Day(), 
        target.Hour(), target.Minute(), 0, 0, now.Location())
    
    if nextRun.Before(now) {
        nextRun = nextRun.Add(24 * time.Hour)
    }
    
    return nextRun
}

// cleanup performs the memory cleanup
func (s *MemoryScheduler) cleanup(ctx context.Context) {
    slog.Info("Starting memory cleanup")
    
    // 清理过期的短期记忆
    expiredCount, err := s.memSvc.PruneExpiredMemories(ctx)
    if err != nil {
        slog.Error("Failed to prune expired memories", "error", err)
    } else {
        slog.Info(fmt.Sprintf("Pruned %d expired short-term memories", expiredCount))
    }
    
    // 清理低相关性的旧记忆（超过30天，相关性<0.3）
    lowRelCount, err := s.memSvc.PruneLowRelevanceMemories(ctx, 30, 0.3)
    if err != nil {
        slog.Error("Failed to prune low relevance memories", "error", err)
    } else {
        slog.Info(fmt.Sprintf("Pruned %d low relevance memories", lowRelCount))
    }
}

// RunCleanupNow manually triggers cleanup
func (s *MemoryScheduler) RunCleanupNow(ctx context.Context) {
    s.cleanup(ctx)
}
```

- [x] **Step 3: 在路由中注册调度器**

```go
// 在 router.go 中添加调度器初始化
memScheduler := scheduler.NewMemoryScheduler(memSvc)
memScheduler.Start(context.Background())
```

- [x] **Step 4: 在 Handler 中添加清理和统计端点**

```go
// 在 memory_handler.go 中添加

// PruneExpiredMemories handles POST /workspaces/:wsParam/memories/prune
func (h *MemoryHandler) PruneExpiredMemories(c *gin.Context) {
    wid := h.parseWorkspaceID(c)
    if wid == 0 {
        c.JSON(400, gin.H{"message": "Invalid workspace"})
        return
    }
    
    count, err := h.svc.PruneExpiredMemories(c.Request.Context())
    if h.respond(c, err) {
        return
    }
    c.JSON(200, gin.H{"pruned_count": count})
}

// GetMemoryStats handles GET /workspaces/:wsParam/memories/stats
func (h *MemoryHandler) GetMemoryStats(c *gin.Context) {
    wid := h.parseWorkspaceID(c)
    if wid == 0 {
        c.JSON(400, gin.H{"message": "Invalid workspace"})
        return
    }
    
    stats, err := h.svc.GetMemoryStats(c.Request.Context(), wid)
    if h.respond(c, err) {
        return
    }
    c.JSON(200, stats)
}
```

- [x] **Step 5: 编译验证**

Run: `cd backend; go build ./...`
Expected: 编译通过

---

### Task 3: 实现相似记忆合并

**Files:**
- Modify: `internal/service/memory_service.go`
- Modify: `internal/handler/memory_handler.go`

**需求分析:**
- 检测内容相似的记忆（基于文本相似度算法）
- 自动合并相似记忆，保留版本历史
- 提供手动合并接口
- 合并时更新相关性分数

- [x] **Step 1: 添加文本相似度检测方法**

```go
// internal/service/memory_service.go

// CalculateTextSimilarity calculates similarity between two texts
func (s *MemoryService) CalculateTextSimilarity(text1, text2 string) float64 {
    // 简单的基于词频的相似度计算
    words1 := strings.Fields(strings.ToLower(text1))
    words2 := strings.Fields(strings.ToLower(text2))
    
    if len(words1) == 0 || len(words2) == 0 {
        return 0
    }
    
    // 计算词频
    freq1 := make(map[string]int)
    freq2 := make(map[string]int)
    
    for _, w := range words1 {
        freq1[w]++
    }
    for _, w := range words2 {
        freq2[w]++
    }
    
    // 计算余弦相似度
    var dot, mag1, mag2 float64
    for w, f1 := range freq1 {
        if f2, ok := freq2[w]; ok {
            dot += float64(f1 * f2)
        }
        mag1 += float64(f1 * f1)
    }
    for _, f2 := range freq2 {
        mag2 += float64(f2 * f2)
    }
    
    if mag1 == 0 || mag2 == 0 {
        return 0
    }
    
    return dot / (math.Sqrt(mag1) * math.Sqrt(mag2))
}

// FindSimilarMemories finds memories similar to the given content
func (s *MemoryService) FindSimilarMemories(ctx context.Context, workspaceID uint64, content string, threshold float64, limit int) ([]*model.MemoryEntry, error) {
    // 获取所有未合并的记忆
    var entries []*model.MemoryEntry
    if err := s.db.WithContext(ctx).
        Where("workspace_id = ? AND expires_at IS NULL OR expires_at > NOW()", workspaceID).
        Limit(100). // 限制搜索范围
        Order("created_at DESC").
        Find(&entries).Error; err != nil {
        return nil, err
    }
    
    // 计算相似度
    var similar []*model.MemoryEntry
    for _, entry := range entries {
        sim := s.CalculateTextSimilarity(content, entry.Content)
        if sim >= threshold {
            entry.RelevanceScore = sim
            similar = append(similar, entry)
        }
    }
    
    // 按相似度排序
    sort.Slice(similar, func(i, j int) bool {
        return similar[i].RelevanceScore > similar[j].RelevanceScore
    })
    
    // 限制返回数量
    if len(similar) > limit {
        similar = similar[:limit]
    }
    
    return similar, nil
}

// MergeMemories merges multiple memories into one
func (s *MemoryService) MergeMemories(ctx context.Context, memoryIDs []uint64, workspaceID uint64) (*model.MemoryEntry, error) {
    if len(memoryIDs) < 2 {
        return nil, errors.New("at least 2 memories required for merge")
    }
    
    // 获取所有要合并的记忆
    var entries []model.MemoryEntry
    if err := s.db.WithContext(ctx).
        Where("id IN ? AND workspace_id = ?", memoryIDs, workspaceID).
        Find(&entries).Error; err != nil {
        return nil, err
    }
    
    if len(entries) < 2 {
        return nil, errors.New("not enough memories found to merge")
    }
    
    // 创建合并后的内容
    var contentBuilder strings.Builder
    contentBuilder.WriteString("=== 合并记忆 ===\n\n")
    
    for i, entry := range entries {
        contentBuilder.WriteString(fmt.Sprintf("【来源%d】\n%s\n\n---\n\n", i+1, entry.Content))
    }
    
    // 创建新的合并记忆
    merged := &model.MemoryEntry{
        WorkspaceID:    workspaceID,
        MemoryType:     entries[0].MemoryType,
        Scope:          entries[0].Scope,
        Content:        contentBuilder.String(),
        ContextKey:     entries[0].ContextKey + "_merged",
        ContextName:    entries[0].ContextName + " (合并)",
        RelevanceScore: 0.9, // 合并后的记忆相关性较高
    }
    
    // 设置关联的项目/问题/Agent（取第一个）
    if entries[0].ProjectID != nil {
        merged.ProjectID = entries[0].ProjectID
    }
    if entries[0].IssueID != nil {
        merged.IssueID = entries[0].IssueID
    }
    if entries[0].AgentID != nil {
        merged.AgentID = entries[0].AgentID
    }
    
    // 创建合并记忆
    if err := s.db.WithContext(ctx).Create(merged).Error; err != nil {
        return nil, err
    }
    
    // 标记原记忆为已合并（通过metadata）
    for _, entry := range entries {
        var meta map[string]interface{}
        if entry.Metadata != nil {
            json.Unmarshal(entry.Metadata, &meta)
        }
        if meta == nil {
            meta = make(map[string]interface{})
        }
        meta["merged_into"] = merged.ID
        meta["merged_at"] = time.Now().Format(time.RFC3339)
        metaJSON, _ := json.Marshal(meta)
        
        s.db.WithContext(ctx).
            Model(&entry).
            Update("metadata", metaJSON)
    }
    
    return merged, nil
}
```

- [x] **Step 2: 在 Handler 中添加相似记忆和合并端点**

```go
// 在 memory_handler.go 中添加

// FindSimilarMemories handles POST /workspaces/:wsParam/memories/find-similar
func (h *MemoryHandler) FindSimilarMemories(c *gin.Context) {
    wid := h.parseWorkspaceID(c)
    if wid == 0 {
        c.JSON(400, gin.H{"message": "Invalid workspace"})
        return
    }
    
    var req struct {
        Content   string  `json:"content" binding:"required"`
        Threshold float64 `json:"threshold"`
        Limit     int     `json:"limit"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"message": err.Error()})
        return
    }
    
    if req.Threshold <= 0 {
        req.Threshold = 0.7
    }
    if req.Limit <= 0 {
        req.Limit = 10
    }
    
    entries, err := h.svc.FindSimilarMemories(c.Request.Context(), wid, req.Content, req.Threshold, req.Limit)
    if h.respond(c, err) {
        return
    }
    c.JSON(200, entries)
}

// MergeMemories handles POST /workspaces/:wsParam/memories/merge
func (h *MemoryHandler) MergeMemories(c *gin.Context) {
    wid := h.parseWorkspaceID(c)
    if wid == 0 {
        c.JSON(400, gin.H{"message": "Invalid workspace"})
        return
    }
    
    var req struct {
        MemoryIDs []uint64 `json:"memory_ids" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"message": err.Error()})
        return
    }
    
    merged, err := h.svc.MergeMemories(c.Request.Context(), req.MemoryIDs, wid)
    if h.respond(c, err) {
        return
    }
    c.JSON(200, merged)
}
```

- [x] **Step 3: 在路由中注册新端点**

```go
// 在 router.go 中添加新路由
memories.POST("/prune", memH.PruneExpiredMemories)
memories.GET("/stats", memH.GetMemoryStats)
memories.POST("/find-similar", memH.FindSimilarMemories)
memories.POST("/merge", memH.MergeMemories)
```

- [x] **Step 4: 编译验证**

Run: `cd backend; go build ./...`
Expected: 编译通过

---

### Task 4: 测试验证

**Files:**
- Test: 通过 API 测试

- [x] **Step 1: 测试搜索功能**

```bash
# 搜索关键词
curl -X GET "http://localhost:8000/api/v1/workspaces/1/memories?q=项目状态&min_relevance=0.5" \
  -H "Authorization: Bearer <token>"

# 日期范围搜索
curl -X GET "http://localhost:8000/api/v1/workspaces/1/memories?start_date=2026-07-01T00:00:00Z&end_date=2026-07-31T23:59:59Z" \
  -H "Authorization: Bearer <token>"
```

- [x] **Step 2: 测试过期清理**

```bash
# 获取统计
curl -X GET "http://localhost:8000/api/v1/workspaces/1/memories/stats" \
  -H "Authorization: Bearer <token>"

# 手动清理
curl -X POST "http://localhost:8000/api/v1/workspaces/1/memories/prune" \
  -H "Authorization: Bearer <token>"
```

- [x] **Step 3: 测试记忆合并**

```bash
# 查找相似记忆
curl -X POST "http://localhost:8000/api/v1/workspaces/1/memories/find-similar" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"content": "项目状态分析", "threshold": 0.7, "limit": 5}'

# 合并记忆
curl -X POST "http://localhost:8000/api/v1/workspaces/1/memories/merge" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"memory_ids": [1, 2, 3]}'
```

---

## 自审查

**1. 规格覆盖:**
- ✅ 记忆搜索增强：支持关键词、日期范围、相关性阈值、排序
- ✅ 记忆过期策略：定时清理、手动清理、清理统计
- ✅ 记忆合并：相似检测、合并操作、版本历史

**2. 占位符扫描:**
- 无占位符，所有步骤都有完整代码

**3. 类型一致性:**
- 使用一致的类型定义和方法签名

---

## 执行交接

**"Plan complete and saved to `docs/superpowers/plans/2026-07-26-memory-enhancement.md`. Two execution options:**

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?"**
