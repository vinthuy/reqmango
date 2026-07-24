package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/ai/common"
	"github.com/reqmango/backend/internal/ai/llm"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// AIService provides AI-powered features on top of the existing services.
type AIService struct {
	db  *gorm.DB
	llm *llm.LLMClient
}

// NewAIService creates an AIService.
func NewAIService(db *gorm.DB, llmClient *llm.LLMClient) *AIService {
	return &AIService{db: db, llm: llmClient}
}

// ==================== Permission Check ====================

func (s *AIService) isProjectMember(projectID, userID uint64) bool {
	if userID == 0 {
		return false
	}
	var count int64
	s.db.Model(&model.ProjectMember{}).
		Where("project_id = ? AND user_id = ? AND is_active = ?", projectID, userID, true).
		Count(&count)
	return count > 0
}

// ==================== System Prompt ====================

func (s *AIService) buildSystemPrompt(ctx *AIContext) string {
	parts := []string{
		"你是 reqmango AI 助手，一个项目管理专家。你帮助用户管理需求、任务和 Bug。",
		fmt.Sprintf("当前上下文：工作空间「%s」> 项目「%s」", ctx.WorkspaceName, ctx.ProjectName),
	}
	if ctx.PageTitle != "" {
		parts = append(parts, fmt.Sprintf("当前页面：「%s」", ctx.PageTitle))
	}
	if ctx.IssueID > 0 {
		parts = append(parts, fmt.Sprintf("当前工作项：%s-#%d「%s」", ctx.ProjectIdentifier, ctx.IssueSequenceID, ctx.IssueName))
	}
	parts = append(parts, "",
		"你有两个模式：",
		"- **Ask 模式**：回答用户关于项目数据的问题。直接查询数据并回答。",
		"- **Build 模式**：执行操作前先展示计划，用户确认后执行。创建/修改必须预览。",
		"",
		"规则：",
		"1. 回答必须基于实际数据（通过工具查询），绝不编造",
		"2. 创建/修改/删除操作：先在对话中展示将要做什么，等用户说「确认」或「执行」后才调用工具",
		"3. 用中文回复",
		"4. 涉及具体工作项时使用项目标识符格式（如 DEMO-42）",
		"5. 如果用户问的问题超出你的工具能力范围，诚实告知",
	)
	return strings.Join(parts, "\n")
}

// ==================== Tool Definitions ====================

// GetTools returns all available AI tool definitions (exported for reuse by AgentService).
func (s *AIService) GetTools() []llm.Tool {
	return s.getTools()
}

func (s *AIService) getTools() []llm.Tool {
	return []llm.Tool{
		{
			Name:        "search_issues",
			Description: "搜索工作项。支持按标题关键词、优先级、状态、类型、负责人等条件筛选。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"project_id":  {Type: "integer", Description: "项目 ID"},
					"query":       {Type: "string", Description: "搜索关键词（标题模糊匹配）"},
					"priority":    {Type: "string", Enum: []string{"urgent", "high", "medium", "low", "none"}, Description: "优先级"},
					"state_id":    {Type: "integer", Description: "状态 ID"},
					"assignee_id": {Type: "integer", Description: "负责人用户 ID"},
					"type_id":     {Type: "integer", Description: "工作项类型 ID"},
					"limit":       {Type: "integer", Description: "返回数量上限，默认 20"},
				},
			},
		},
		{
			Name:        "create_issue",
			Description: "创建一个新的工作项。在 Build 模式下必须先展示预览等用户确认。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"project_id":   {Type: "integer", Description: "项目 ID"},
					"workspace_id": {Type: "integer", Description: "工作空间 ID"},
					"name":         {Type: "string", Description: "工作项标题"},
					"description":  {Type: "string", Description: "描述（支持 Markdown）"},
					"priority":     {Type: "string", Enum: []string{"urgent", "high", "medium", "low", "none"}, Description: "优先级"},
					"type_id":      {Type: "integer", Description: "工作项类型 ID"},
					"state_id":     {Type: "integer", Description: "状态 ID"},
					"assignee_ids": {Type: "array", Description: "负责人用户 ID 列表"},
					"label_ids":    {Type: "array", Description: "标签 ID 列表"},
					"parent_id":    {Type: "integer", Description: "父工作项 ID（用于子任务）"},
					"target_date":  {Type: "string", Description: "截止日期，格式 YYYY-MM-DD"},
				},
				Required: []string{"project_id", "workspace_id", "name"},
			},
		},
		{
			Name:        "get_issue",
			Description: "获取单个工作项的详细信息。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"issue_id": {Type: "integer", Description: "工作项 ID"},
				},
				Required: []string{"issue_id"},
			},
		},
		{
			Name:        "update_issue",
			Description: "更新工作项的属性（优先级、状态、负责人等）。在 Build 模式下必须先展示变更预览。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"issue_id":    {Type: "integer", Description: "工作项 ID"},
					"name":        {Type: "string", Description: "新标题"},
					"priority":    {Type: "string", Enum: []string{"urgent", "high", "medium", "low", "none"}},
					"state_id":    {Type: "integer", Description: "新状态 ID"},
					"description": {Type: "string", Description: "新描述"},
				},
				Required: []string{"issue_id"},
			},
		},
		{
			Name:        "get_project_stats",
			Description: "获取项目的统计信息：总工作项数、完成数、进度、状态分布、活跃成员数。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"project_id": {Type: "integer", Description: "项目 ID"},
				},
				Required: []string{"project_id"},
			},
		},
		{
			Name:        "get_issues_summary",
			Description: "获取项目工作项概要：待办/进行中/已完成/已取消的数量。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"project_id": {Type: "integer", Description: "项目 ID"},
				},
				Required: []string{"project_id"},
			},
		},
		{
			Name:        "list_members",
			Description: "列出项目成员及其角色。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"project_id": {Type: "integer", Description: "项目 ID"},
				},
				Required: []string{"project_id"},
			},
		},
		{
			Name:        "list_issue_types",
			Description: "列出可用的工作项类型（如 Bug、Task、Feature、Epic 等）。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"workspace_id": {Type: "integer", Description: "工作空间 ID"},
				},
				Required: []string{"workspace_id"},
			},
		},
		{
			Name:        "list_states",
			Description: "列出项目的所有工作项状态。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"project_id": {Type: "integer", Description: "项目 ID"},
				},
				Required: []string{"project_id"},
			},
		},
		{
			Name:        "list_labels",
			Description: "列出项目的所有标签。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"project_id": {Type: "integer", Description: "项目 ID"},
				},
				Required: []string{"project_id"},
			},
		},
		{
			Name:        "list_cycles",
			Description: "列出项目的迭代周期。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"project_id": {Type: "integer", Description: "项目 ID"},
				},
				Required: []string{"project_id"},
			},
		},
		{
			Name:        "get_cycle_progress",
			Description: "获取迭代周期的进度信息。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"cycle_id": {Type: "integer", Description: "周期 ID"},
				},
				Required: []string{"cycle_id"},
			},
		},
		{
			Name:        "list_modules",
			Description: "列出项目的功能模块。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"project_id": {Type: "integer", Description: "项目 ID"},
				},
				Required: []string{"project_id"},
			},
		},
		{
			Name:        "get_issue_activities",
			Description: "获取工作项的活动日志（变更历史）。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"issue_id": {Type: "integer", Description: "工作项 ID"},
				},
				Required: []string{"issue_id"},
			},
		},
		{
			Name:        "add_comment",
			Description: "给工作项添加评论。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"issue_id": {Type: "integer", Description: "工作项 ID"},
					"body":     {Type: "string", Description: "评论内容"},
				},
				Required: []string{"issue_id", "body"},
			},
		},
		{
			Name:        "list_releases",
			Description: "列出项目的发布版本。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"project_id": {Type: "integer", Description: "项目 ID"},
				},
				Required: []string{"project_id"},
			},
		},
		{
			Name:        "list_pages",
			Description: "列出项目的文档页面。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"project_id": {Type: "integer", Description: "项目 ID"},
				},
				Required: []string{"project_id"},
			},
		},
		{
			Name:        "web_search",
			Description: "在互联网上搜索最新信息。当用户询问的知识超出你的训练数据范围，或需要实时信息时使用。返回标题、URL 和摘要。",
			InputSchema: &llm.ToolSchema{
				Type: "object",
				Properties: map[string]llm.SchemaProp{
					"query": {Type: "string", Description: "搜索查询字符串"},
					"num":   {Type: "integer", Description: "返回结果数量，默认 5，最大 10"},
				},
				Required: []string{"query"},
			},
		},
	}
}

// ==================== Context ====================

// AIContext holds the current user/project context for AI requests.
type AIContext struct {
	WorkspaceID       uint64
	WorkspaceName     string
	ProjectID         uint64
	ProjectName       string
	ProjectIdentifier string
	PageTitle         string
	IssueID           uint64
	IssueSequenceID   int
	IssueName         string
	Mode              string // "ask" | "build"
	UserID            uint64
	Token             string `json:"-"` // JWT token for backend API calls
}

// ==================== Chat ====================

// AIChatRequest is the request for the chat endpoint.
type AIChatRequest struct {
	Message  string `json:"message"`
	ThreadID uint64 `json:"thread_id"`
	Mode     string `json:"mode"` // "ask" | "build"
}

// Chat handles a conversational AI request with SSE streaming.
func (s *AIService) Chat(ctx context.Context, req *AIChatRequest, actx *AIContext) (<-chan llm.StreamEvent, error) {
	systemPrompt := s.buildSystemPrompt(actx)
	if req.Mode != "" {
		actx.Mode = req.Mode
	}
	if actx.Mode == "build" {
		systemPrompt += "\n\n当前为 Build 模式：执行操作前展示预览，等用户说「确认」再执行。"
	} else {
		systemPrompt += "\n\n当前为 Ask 模式：回答用户问题，展示数据。如需操作请建议用户切换到 Build 模式。"
	}

	messages := []llm.Message{
		{Role: "user", Content: req.Message},
	}

	if req.ThreadID > 0 {
		var dbMessages []model.AIMessage
		if err := s.db.Where("thread_id = ?", req.ThreadID).Order("created_at ASC").Find(&dbMessages).Error; err == nil {
			var historyMsgs []llm.Message
			for _, m := range dbMessages {
				historyMsgs = append(historyMsgs, llm.Message{Role: m.Role, Content: m.Content})
			}
			messages = append(historyMsgs, messages...)
		}
	}

	tools := s.getTools()
	streamCh, err := s.llm.ChatStream(ctx, systemPrompt, messages, tools)
	if err != nil {
		return nil, err
	}

	outCh := make(chan llm.StreamEvent, 64)
	go func() {
		defer close(outCh)
		emit := func(evt llm.StreamEvent) bool {
			select {
			case outCh <- evt:
				return true
			case <-ctx.Done():
				return false
			}
		}
		var toolResults []llm.ToolResult
		hasToolCalls := false

		for evt := range streamCh {
			switch evt.Type {
			case "tool_call":
				hasToolCalls = true
				if !emit(evt) {
					return
				}
				if evt.ToolCall != nil {
					result, execErr := s.ExecuteTool(evt.ToolCall.Name, evt.ToolCall.Input, actx)
					content := ""
					if execErr != nil {
						content = fmt.Sprintf(`{"error":"%s"}`, execErr.Error())
					} else {
						b, _ := json.Marshal(result)
						content = string(b)
					}
					toolResults = append(toolResults, llm.ToolResult{
						ToolCallID: evt.ToolCall.ID,
						Content:    content,
					})
					if !emit(llm.StreamEvent{
						Type:       "tool_result",
						ToolResult: &toolResults[len(toolResults)-1],
					}) {
						return
					}
				}
			case "done":
				if hasToolCalls && len(toolResults) > 0 {
					var sb strings.Builder
					sb.WriteString("以下是查询获取的项目数据：\n<data>\n")
					for i, tr := range toolResults {
						sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, tr.Content))
					}
					sb.WriteString("</data>\n\n")
					sb.WriteString(fmt.Sprintf("用户问题：「%s」\n", messages[0].Content))
					sb.WriteString("请基于以上真实数据，用中文给出分析总结。提炼关键数字，结构化呈现，给出行动建议。")

					synthPrompt := systemPrompt + "\n\n你已获得项目数据。请基于实际数据进行分析汇报。像PM做Sprint Review：先说结论，再分要点，最后给下一步建议。"

					synthCh, synthErr := s.llm.ChatStream(ctx, synthPrompt, []llm.Message{
						{Role: "user", Content: sb.String()},
					}, nil)
					if synthErr == nil {
						for se := range synthCh {
							if !emit(se) {
								return
							}
						}
						continue
					}
				}
				if !emit(evt) {
					return
				}
			default:
				if !emit(evt) {
					return
				}
			}
		}
	}()

	return outCh, nil
}

// ==================== AI Chart Generation ====================

type AIChartRequest struct {
	Query string `json:"query"`
}

type AIChartResponse struct {
	ChartType string           `json:"chart_type"`
	Title     string           `json:"title"`
	Labels    []string         `json:"labels"`
	Datasets  []AIChartDataset `json:"datasets"`
	Options   *AIChartOptions  `json:"options,omitempty"`
}

type AIChartDataset struct {
	Label           string    `json:"label"`
	Data            []float64 `json:"data"`
	BackgroundColor []string  `json:"backgroundColor,omitempty"`
	BorderColor     []string  `json:"borderColor,omitempty"`
	Fill            *bool     `json:"fill,omitempty"`
	Tension         float64   `json:"tension,omitempty"`
}

type AIChartOptions struct {
	IndexAxis  string `json:"indexAxis,omitempty"`
	Stacked    bool   `json:"stacked,omitempty"`
	ShowLegend bool   `json:"showLegend,omitempty"`
}

type projectStatsData struct {
	ProjectID       uint64         `json:"project_id"`
	ProjectName     string         `json:"project_name"`
	TotalIssues     int64          `json:"total_issues"`
	CompletedIssues int64          `json:"completed_issues"`
	ActiveMembers   int64          `json:"active_members"`
	States          map[string]int `json:"states"`
	Priorities      map[string]int `json:"priorities"`
}

type issuesSummaryData struct {
	ProjectID   uint64         `json:"project_id"`
	ProjectName string         `json:"project_name"`
	Issues      map[string]int `json:"issues"`
}

func (s *AIService) getProjectStats(projectID uint64) (*projectStatsData, error) {
	var project model.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return nil, common.NotFound("Project not found")
	}

	var totalIssues int64
	s.db.Model(&model.Issue{}).Where("project_id = ?", projectID).Count(&totalIssues)

	var completedIssues int64
	s.db.Model(&model.Issue{}).
		Joins("JOIN states ON states.id = issues.state_id").
		Where("issues.project_id = ? AND states.group = ?", projectID, common.StateGroupCompleted).
		Count(&completedIssues)

	var activeMembers int64
	s.db.Model(&model.ProjectMember{}).
		Where("project_id = ? AND is_active = ?", projectID, true).
		Count(&activeMembers)

	stateCounts := make(map[string]int)
	var states []model.State
	s.db.Where("project_id = ? AND is_active = ?", projectID, true).Find(&states)
	for _, state := range states {
		var count int64
		s.db.Model(&model.Issue{}).Where("project_id = ? AND state_id = ?", projectID, state.ID).Count(&count)
		stateCounts[state.Name] = int(count)
	}

	priorityCounts := make(map[string]int)
	for _, p := range []string{"urgent", "high", "medium", "low", "none"} {
		var count int64
		s.db.Model(&model.Issue{}).Where("project_id = ? AND priority = ?", projectID, p).Count(&count)
		priorityCounts[p] = int(count)
	}

	return &projectStatsData{
		ProjectID:       project.ID,
		ProjectName:     project.Name,
		TotalIssues:     totalIssues,
		CompletedIssues: completedIssues,
		ActiveMembers:   activeMembers,
		States:          stateCounts,
		Priorities:      priorityCounts,
	}, nil
}

func (s *AIService) getIssuesSummaryData(projectID uint64) (*issuesSummaryData, error) {
	var project model.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return nil, common.NotFound("Project not found")
	}

	summary := &issuesSummaryData{
		ProjectID:   project.ID,
		ProjectName: project.Name,
		Issues:      make(map[string]int),
	}

	groups := []string{common.StateGroupUnstarted, common.StateGroupStarted, common.StateGroupCompleted, common.StateGroupCancelled}
	for _, group := range groups {
		var count int64
		s.db.Model(&model.Issue{}).
			Joins("JOIN states ON states.id = issues.state_id").
			Where("issues.project_id = ? AND states.group = ?", projectID, group).
			Count(&count)
		summary.Issues[group] = int(count)
	}

	summary.Issues["todo"] = summary.Issues[common.StateGroupUnstarted]
	summary.Issues["in_progress"] = summary.Issues[common.StateGroupStarted]
	summary.Issues["done"] = summary.Issues[common.StateGroupCompleted]
	summary.Issues["cancelled"] = summary.Issues[common.StateGroupCancelled]

	return summary, nil
}

// GenerateChart generates a chart from natural language query.
func (s *AIService) GenerateChart(ctx context.Context, req *AIChartRequest, actx *AIContext) (*AIChartResponse, error) {
	stats, _ := s.getProjectStats(actx.ProjectID)
	summary, _ := s.getIssuesSummaryData(actx.ProjectID)

	started := summary.Issues["started"]
	todo := summary.Issues["todo"]
	cancelled := summary.Issues["cancelled"]
	inProgress := started

	dataCtx := fmt.Sprintf(`项目数据:
- 工作项总数: %d
- 已完成: %d, 进行中: %d, 未开始: %d, 已取消: %d
- 成员数: %d
- 按状态: %v
- 按优先级: %v`,
		stats.TotalIssues, stats.CompletedIssues,
		inProgress, todo, cancelled,
		stats.ActiveMembers,
		stats.States, stats.Priorities,
	)

	prompt := fmt.Sprintf(`你是一个数据可视化专家。根据用户请求和数据上下文，生成图表配置。

%s

用户请求: %s

请直接输出纯 JSON（不要 markdown 标记），格式如下:
{
  "chart_type": "bar|pie|doughnut|line|polarArea",
  "title": "图表标题",
  "labels": ["标签1", "标签2"],
  "datasets": [{"label": "数据集标签", "data": [10, 20]}]
}

规则:
1. 聚合类（按状态/优先级/类型/负责人统计数量）用 pie/doughnut/bar
2. 趋势类（随时间变化）用 line
3. 对比类用 bar（可为 horizontal）
4. data 必须是真实统计数据，如果数据不足用 0 填充
5. labels 用中文
6. 不要编造数据，基于上面提供的数据上下文推算`, dataCtx, req.Query)

	content, err := s.llm.Complete(ctx, "你是一个数据可视化专家。输出纯 JSON，不要有任何其他内容。", prompt)
	if err != nil {
		return nil, fmt.Errorf("AI chart generation failed: %w", err)
	}

	var chartResp AIChartResponse
	if err := json.Unmarshal([]byte(content), &chartResp); err != nil {
		return nil, fmt.Errorf("parse chart response: %w (raw: %s)", err, content[:minX(len(content), 200)])
	}

	n := len(chartResp.Labels)
	if n == 0 && len(chartResp.Datasets) > 0 {
		n = len(chartResp.Datasets[0].Data)
	}

	colors8 := []string{
		"#6366f1", "#8b5cf6", "#ec4899", "#f43f5e",
		"#f97316", "#eab308", "#22c55e", "#06b6d4",
	}
	bgColors := make([]string, n)
	borderColors := make([]string, n)
	for i := 0; i < n; i++ {
		bgColors[i] = colors8[i%len(colors8)]
		borderColors[i] = colors8[i%len(colors8)]
	}

	for i := range chartResp.Datasets {
		if chartResp.Datasets[i].BackgroundColor == nil {
			chartResp.Datasets[i].BackgroundColor = bgColors
		}
		if chartResp.Datasets[i].BorderColor == nil {
			chartResp.Datasets[i].BorderColor = borderColors
		}
		if chartResp.Datasets[i].Tension == 0 {
			chartResp.Datasets[i].Tension = 0.3
		}
	}

	s.enrichChartData(&chartResp, actx)

	return &chartResp, nil
}

func (s *AIService) enrichChartData(chart *AIChartResponse, actx *AIContext) {
	labelsLower := make([]string, len(chart.Labels))
	for i, l := range chart.Labels {
		labelsLower[i] = strings.ToLower(l)
	}

	var states []model.State
	s.db.Where("project_id = ?", actx.ProjectID).Find(&states)

	hasStateLabels := false
	for _, l := range labelsLower {
		for _, st := range states {
			if strings.Contains(l, strings.ToLower(st.Name)) || strings.Contains(strings.ToLower(st.Name), l) {
				hasStateLabels = true
				break
			}
		}
	}

	if hasStateLabels && len(chart.Datasets) > 0 {
		newLabels := make([]string, 0, len(states))
		newData := make([]float64, 0, len(states))
		for _, st := range states {
			var count int64
			s.db.Model(&model.Issue{}).Where("project_id = ? AND state_id = ?", actx.ProjectID, st.ID).Count(&count)
			newLabels = append(newLabels, st.Name)
			newData = append(newData, float64(count))
		}
		chart.Labels = newLabels
		chart.Datasets[0].Data = newData
	}
}

// ==================== AI Search ====================

type AISearchRequest struct {
	Query string `json:"query"`
}

type AISearchResponse struct {
	RQL         string                   `json:"rql"`
	Explanation string                   `json:"explanation"`
	Issues      []map[string]interface{} `json:"issues"`
}

// Search translates natural language to a search query and executes it.
func (s *AIService) Search(ctx context.Context, req *AISearchRequest, actx *AIContext) (*AISearchResponse, error) {
	translatePrompt := fmt.Sprintf(`你是一个查询翻译器。将用户的自然语言需求翻译为 reqmango 查询，然后搜索并返回结果。

项目信息：
- 项目 ID: %d
- 项目标识符: %s
- 可用状态: 通过 list_states 工具查询
- 可用类型: 通过 list_issue_types 工具查询

用户需求: %s

请先用工具查询匹配的工作项，然后用中文解释搜索结果。`,
		actx.ProjectID, actx.ProjectIdentifier, req.Query)

	tools := s.getTools()
	readOnlyTools := filterReadOnlyTools(tools)

	var issues []map[string]interface{}

	resp, err := s.llm.ChatSyncWithTools(ctx, translatePrompt, []llm.Message{{Role: "user", Content: req.Query}}, readOnlyTools, func(name string, input json.RawMessage) (string, error) {
		result, execErr := s.ExecuteTool(name, input, actx)
		if execErr != nil {
			return "", execErr
		}
		if name == "search_issues" {
			if arr, ok := result.([]map[string]interface{}); ok {
				issues = arr
			}
		}
		b, _ := json.Marshal(result)
		return string(b), nil
	})
	if err != nil {
		return nil, err
	}

	return &AISearchResponse{
		RQL:         "",
		Explanation: resp.Content,
		Issues:      issues,
	}, nil
}

// ==================== Create Preview ====================

type AICreateRequest struct {
	Description string `json:"description"`
}

type AICreateResponse struct {
	Preview     map[string]interface{} `json:"preview"`
	Explanation string                 `json:"explanation"`
}

// CreatePreview parses a natural language description into a structured issue preview.
func (s *AIService) CreatePreview(ctx context.Context, req *AICreateRequest, actx *AIContext) (*AICreateResponse, error) {
	var states []model.State
	s.db.Where("project_id = ? AND is_active = ?", actx.ProjectID, true).Order("sequence").Find(&states)
	stateInfo := make([]string, len(states))
	for i, st := range states {
		stateInfo[i] = fmt.Sprintf("%d:%s(%s)", st.ID, st.Name, st.Group)
	}

	var types []model.IssueType
	s.db.Where("workspace_id = ? AND is_active = ? AND project_id IS NULL", actx.WorkspaceID, true).Find(&types)
	typeInfo := make([]string, len(types))
	for i, t := range types {
		typeInfo[i] = fmt.Sprintf("%d:%s", t.ID, t.Name)
	}

	var members []model.ProjectMember
	s.db.Preload("User").Where("project_id = ?", actx.ProjectID).Find(&members)
	memberInfo := make([]string, len(members))
	for i, m := range members {
		memberInfo[i] = fmt.Sprintf("%d:%s", m.UserID, m.User.DisplayName)
	}

	createPrompt := fmt.Sprintf(`你是一个工作项创建助手。将用户的自然语言描述解析为结构化的工作项预览。

当前项目：%s（标识符: %s）

可用状态（ID:名称(分组)）：%s
可用类型（ID:名称）：%s
可用成员（ID:显示名）：%s

解析规则：
1. 提取标题：问题的简短描述（不要包含"创建"等动词前缀）
2. 识别类型：根据关键词匹配上述可用类型，选最接近的。不确定就留空。
3. 识别优先级：P0/紧急/crash=urgent, P1/高=high, P2/中=medium, P3/低=low, 未提及=none
4. 识别人名→匹配上述成员ID。只使用列出的成员ID。
5. 识别日期→格式 YYYY-MM-DD
6. 生成详细描述（Markdown 格式，包含复现步骤、期望行为等）

只输出纯JSON，不要markdown标记，不要其他文字：
{
  "name": "标题",
  "description": "详细描述",
  "priority": "high",
  "type_id": 4,
  "state_id": 1,
  "assignee_ids": [3],
  "label_ids": [],
  "target_date": "2026-07-03"
}`,
		actx.ProjectName, actx.ProjectIdentifier,
		strings.Join(stateInfo, ", "),
		strings.Join(typeInfo, ", "),
		strings.Join(memberInfo, ", "))

	content, err := s.llm.Complete(ctx, "你是一个工作项创建助手。只输出纯JSON，不要任何markdown标记或额外文字。", createPrompt)
	if err != nil {
		return nil, err
	}

	preview := extractJSON(content)
	if preview == nil {
		preview = map[string]interface{}{
			"name":        req.Description,
			"description": "",
			"priority":    "none",
		}
	}

	explanation := "已根据你的描述生成工作项预览"
	if name, ok := preview["name"].(string); ok && name != "" {
		explanation = fmt.Sprintf("已生成预览: %s", name)
	}

	return &AICreateResponse{
		Preview:     preview,
		Explanation: explanation,
	}, nil
}

// ==================== Analyze ====================

type AIAnalyzeResponse struct {
	Summary     string                 `json:"summary"`
	Insights    []string               `json:"insights"`
	Bottlenecks []AIBottleneck         `json:"bottlenecks,omitempty"`
	Stats       map[string]interface{} `json:"stats"`
}

type AIBottleneck struct {
	IssueID     uint64 `json:"issue_id"`
	IssueName   string `json:"issue_name"`
	DaysInState int    `json:"days_in_state"`
	StateName   string `json:"state_name"`
}

// Analyze generates an AI-powered analysis of the project.
func (s *AIService) Analyze(ctx context.Context, actx *AIContext) (*AIAnalyzeResponse, error) {
	stats, err := s.getProjectStats(actx.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	summary, err := s.getIssuesSummaryData(actx.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("get summary: %w", err)
	}

	var bottlenecks []AIBottleneck
	var stuckIssues []model.Issue
	s.db.Where("project_id = ? AND state_id IN (SELECT id FROM states WHERE \"group\" = 'started')", actx.ProjectID).
		Order("updated_at ASC").Limit(20).Find(&stuckIssues)
	for _, iss := range stuckIssues {
		days := int(time.Since(iss.UpdatedAt).Hours() / 24)
		if days > 3 {
			var state model.State
			if s.db.First(&state, iss.StateID).Error == nil {
				bottlenecks = append(bottlenecks, AIBottleneck{
					IssueID: iss.ID, IssueName: iss.Name,
					DaysInState: days, StateName: state.Name,
				})
			}
		}
	}

	statsJSON, _ := json.MarshalIndent(stats, "", "  ")
	summaryJSON, _ := json.MarshalIndent(summary, "", "  ")
	bottleneckJSON, _ := json.MarshalIndent(bottlenecks, "", "  ")

	analyzePrompt := fmt.Sprintf(`你是一个高级项目分析师。根据以下真实数据，对项目「%s」进行专业分析。

## 项目统计
%s

## 工作项概要
%s

## 瓶颈检测 (停滞超过3天的进行中任务)
%s

请用中文输出 JSON 格式的分析报告：
{
  "summary": "一段简洁的项目健康度概述 (100字以内)",
  "insights": ["洞察1", "洞察2", "洞察3"],
  "recommendations": ["建议1", "建议2"]
}

关注点：
1. 完成率是否健康
2. 是否有任务积压
3. 瓶颈在哪里
4. 需要立即关注的事项`,
		actx.ProjectName, statsJSON, summaryJSON, bottleneckJSON)

	content, err := s.llm.Complete(ctx, "你是一个高级项目分析师。请输出严格的 JSON 格式，不要添加任何 markdown 标记。", analyzePrompt)
	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	var result AIAnalyzeResponse
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		result = AIAnalyzeResponse{
			Summary:  content[:minX(len(content), 200)],
			Insights: []string{},
			Stats:    map[string]interface{}{"raw_analysis": content},
		}
	}

	result.Bottlenecks = bottlenecks
	progressPct := float64(0)
	if stats.TotalIssues > 0 {
		progressPct = float64(stats.CompletedIssues) / float64(stats.TotalIssues) * 100
	}
	result.Stats = map[string]interface{}{
		"total_issues":   stats.TotalIssues,
		"completed":      stats.CompletedIssues,
		"progress_pct":   int(progressPct),
		"active_members": stats.ActiveMembers,
		"issues_summary": summary,
	}
	return &result, nil
}

func minX(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ==================== AI Labels ====================

type AILabelResponse struct {
	SuggestedLabels []AILabelSuggestion `json:"suggested_labels"`
}

type AILabelSuggestion struct {
	LabelID    uint64  `json:"label_id"`
	LabelName  string  `json:"label_name"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

// SuggestLabels uses AI to suggest labels for an issue.
func (s *AIService) SuggestLabels(ctx context.Context, projectID uint64, name, description string) (*AILabelResponse, error) {
	var labels []model.Label
	s.db.Where("project_id = ?", projectID).Find(&labels)
	if len(labels) == 0 {
		return &AILabelResponse{}, nil
	}

	labelList := make([]string, len(labels))
	for i, l := range labels {
		labelList[i] = fmt.Sprintf("%d:%s", l.ID, l.Name)
	}

	prompt := fmt.Sprintf(`根据工作项标题和描述，从以下标签中选择最合适的1-3个标签。
可用标签: %s
标题: %s
描述: %s
输出JSON: {"labels":[{"id":1,"reason":"简短理由"}]}`,
		strings.Join(labelList, ", "), name, description[:minX(len(description), 300)])

	result, err := s.llm.Complete(ctx, "你是标签分类专家。只输出JSON。", prompt)
	if err != nil {
		return s.ruleBasedLabels(labels, name, description), nil
	}

	var parsed struct {
		Labels []struct {
			ID     uint64 `json:"id"`
			Reason string `json:"reason"`
		} `json:"labels"`
	}
	jsonStr := result
	if start, end := strings.Index(result, "{"), strings.LastIndex(result, "}"); start >= 0 && end > start {
		jsonStr = result[start : end+1]
	}
	if json.Unmarshal([]byte(jsonStr), &parsed) != nil {
		return s.ruleBasedLabels(labels, name, description), nil
	}

	suggestions := make([]AILabelSuggestion, 0, len(parsed.Labels))
	for _, p := range parsed.Labels {
		for _, l := range labels {
			if l.ID == p.ID {
				suggestions = append(suggestions, AILabelSuggestion{LabelID: l.ID, LabelName: l.Name, Confidence: 0.85, Reason: p.Reason})
			}
		}
	}
	return &AILabelResponse{SuggestedLabels: suggestions}, nil
}

// ==================== AI Comments ====================

type AICommentRequest struct {
	Action   string `json:"action"`
	Comments string `json:"comments"`
}

type AICommentResponse struct {
	Result string `json:"result"`
}

// AssistComment generates AI assistance for comments.
func (s *AIService) AssistComment(ctx context.Context, req *AICommentRequest) (*AICommentResponse, error) {
	var prompt string
	switch req.Action {
	case "summarize":
		prompt = fmt.Sprintf("请用中文总结以下讨论的要点（3-5条）：\n\n%s", req.Comments)
	case "weekly_report":
		prompt = fmt.Sprintf("根据以下工作项讨论，生成一份简短的周报摘要：\n\n%s", req.Comments)
	case "improve":
		prompt = fmt.Sprintf("请改进以下评论文本，使其更清晰专业，保持原意：\n\n%s", req.Comments)
	default:
		prompt = fmt.Sprintf("请总结以下内容：\n\n%s", req.Comments)
	}

	result, err := s.llm.Complete(ctx, "你是技术支持助手。直接输出结果，不加解释。", prompt)
	if err != nil {
		return nil, fmt.Errorf("AI comment failed: %w", err)
	}
	return &AICommentResponse{Result: strings.TrimSpace(result)}, nil
}

// ==================== AI Sprint Planning ====================

type AISprintPlanResponse struct {
	RecommendedCapacity int      `json:"recommended_capacity"`
	SuggestedIssues     []uint64 `json:"suggested_issues"`
	Reasoning           string   `json:"reasoning"`
	Risks               []string `json:"risks"`
}

// SprintPlan generates AI sprint planning recommendations.
func (s *AIService) SprintPlan(ctx context.Context, projectID uint64) (*AISprintPlanResponse, error) {
	var completedCycles []model.Cycle
	s.db.Where("project_id = ? AND completed_at IS NOT NULL", projectID).Order("completed_at DESC").Limit(5).Find(&completedCycles)

	var backlogIssues []model.Issue
	var backlogState model.State
	s.db.Where("project_id = ? AND \"group\" = ?", projectID, "backlog").First(&backlogState)
	s.db.Where("project_id = ? AND state_id = ? AND priority IN (?)", projectID, backlogState.ID, []string{"urgent", "high"}).
		Order("priority, sequence_id").Limit(20).Find(&backlogIssues)

	cycleSummary := ""
	for _, c := range completedCycles {
		var count int64
		s.db.Model(&model.IssueCycle{}).Where("cycle_id = ?", c.ID).Count(&count)
		cycleSummary += fmt.Sprintf("- %s: %d issues, %s → %s\n", c.Name, count, c.StartDate.Format("01-02"), c.EndDate.Format("01-02"))
	}

	issueList := ""
	for _, i := range backlogIssues {
		issueList += fmt.Sprintf("- #%d %s [%s]\n", i.SequenceID, i.Name, i.Priority)
	}

	prompt := fmt.Sprintf(`你是敏捷教练。根据以下数据给出Sprint规划建议。

历史Sprint数据:
%s

待办高优先级工作项:
%s

请输出JSON（不要markdown标记）:
{
  "recommended_capacity": 建议纳入的工作项数量,
  "reasoning": "简短中文分析（50字内）",
  "risks": ["风险1", "风险2"]
}`,
		cycleSummary, issueList)

	result, err := s.llm.Complete(ctx, "你是敏捷规划专家。输出纯JSON。", prompt)
	if err != nil {
		avg := 10
		if len(completedCycles) > 0 {
			avg = 10
		}
		ids := make([]uint64, 0)
		for i, iss := range backlogIssues {
			if i >= avg {
				break
			}
			ids = append(ids, iss.ID)
		}
		return &AISprintPlanResponse{
			RecommendedCapacity: avg,
			SuggestedIssues:     ids,
			Reasoning:           "基于历史Sprint平均容量建议",
			Risks:               []string{"建议人工审核优先级"},
		}, nil
	}

	var plan AISprintPlanResponse
	if json.Unmarshal([]byte(result), &plan) != nil {
		plan.Reasoning = result
	}
	ids := make([]uint64, 0)
	for i, iss := range backlogIssues {
		if i >= plan.RecommendedCapacity {
			break
		}
		ids = append(ids, iss.ID)
	}
	plan.SuggestedIssues = ids
	return &plan, nil
}

func (s *AIService) ruleBasedLabels(labels []model.Label, name, description string) *AILabelResponse {
	text := strings.ToLower(name + " " + description)
	var suggestions []AILabelSuggestion
	for _, l := range labels {
		if strings.Contains(text, strings.ToLower(l.Name)) {
			suggestions = append(suggestions, AILabelSuggestion{LabelID: l.ID, LabelName: l.Name, Confidence: 0.7, Reason: "关键词匹配"})
		}
	}
	if len(suggestions) == 0 {
		if strings.Contains(text, "bug") || strings.Contains(text, "错误") || strings.Contains(text, "修复") {
			for _, l := range labels {
				if strings.Contains(strings.ToLower(l.Name), "bug") {
					suggestions = append(suggestions, AILabelSuggestion{LabelID: l.ID, LabelName: l.Name, Confidence: 0.6, Reason: "Bug关键词"})
					break
				}
			}
		}
	}
	return &AILabelResponse{SuggestedLabels: suggestions}
}

// ==================== AI Triage ====================

type AITriageResponse struct {
	SuggestedType     string   `json:"suggested_type"`
	SuggestedPriority string   `json:"suggested_priority"`
	HasDuplicates     bool     `json:"has_duplicates"`
	DuplicateIDs      []uint64 `json:"duplicate_ids"`
	Summary           string   `json:"summary"`
}

// TriageAnalyze runs AI analysis on an intake issue.
func (s *AIService) TriageAnalyze(ctx context.Context, issueID uint64, projectID uint64) (*AITriageResponse, error) {
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil {
		return nil, fmt.Errorf("issue not found")
	}

	var dupes []model.Issue
	s.db.Where("project_id = ? AND id != ? AND name ILIKE ?", projectID, issueID, "%"+issue.Name[:minX(len(issue.Name), 10)]+"%").
		Limit(5).Find(&dupes)

	dupIDs := make([]uint64, len(dupes))
	for i, d := range dupes {
		dupIDs[i] = d.ID
	}

	prompt := fmt.Sprintf(`分析以下新提交的工作项，给出分类建议。

标题: %s
描述: %s
提交者: %s
可能重复: %d 个相似工作项

请输出JSON格式（不要markdown标记）:
{
  "suggested_type": "Bug|Feature|Task|Improvement",
  "suggested_priority": "urgent|high|medium|low",
  "summary": "一句话总结（中文，30字内）"
}`,
		issue.Name, issue.DescriptionHTML, "外部用户", len(dupes))

	result, err := s.llm.Complete(ctx, "你是项目分诊专家。只输出JSON，不加解释。", prompt)
	if err != nil {
		name := strings.ToLower(issue.Name)
		st := "Task"
		sp := "medium"
		if strings.Contains(name, "bug") || strings.Contains(name, "错误") || strings.Contains(name, "崩溃") {
			st = "Bug"
			sp = "high"
		}
		if strings.Contains(name, "登录") || strings.Contains(name, "支付") {
			sp = "urgent"
		}
		return &AITriageResponse{
			SuggestedType: st, SuggestedPriority: sp,
			HasDuplicates: len(dupes) > 0, DuplicateIDs: dupIDs,
			Summary: fmt.Sprintf("[离线分析] 建议类型: %s, 优先级: %s, 相似工作项: %d个", st, sp, len(dupes)),
		}, nil
	}

	var resp AITriageResponse
	if json.Unmarshal([]byte(result), &resp) != nil {
		resp = AITriageResponse{Summary: result}
	}
	resp.HasDuplicates = len(dupes) > 0
	resp.DuplicateIDs = dupIDs
	return &resp, nil
}

// ==================== Page AI ====================

type PageAIRequest struct {
	Action  string `json:"action"`
	Content string `json:"content"`
	Context string `json:"context"`
}

type PageAIResponse struct {
	Result string `json:"result"`
}

// PageAI performs AI operations on page content.
func (s *AIService) PageAI(ctx context.Context, req *PageAIRequest) (*PageAIResponse, error) {
	var taskPrompt string
	switch req.Action {
	case "summarize":
		taskPrompt = fmt.Sprintf("请用中文总结以下内容的关键要点，用3-5个要点呈现：\n\n%s", req.Content)
	case "improve":
		taskPrompt = fmt.Sprintf("请改进以下文本的写作质量，修正语法错误，使表达更清晰专业，保持原意不变：\n\n%s", req.Content)
	case "translate":
		taskPrompt = fmt.Sprintf("请将以下内容翻译为中文，如果已经是中文则翻译为英文：\n\n%s", req.Content)
	default:
		taskPrompt = fmt.Sprintf("根据以下上下文和内容，生成专业的文档内容。上下文：%s\n\n参考内容：\n%s", req.Context, req.Content)
	}

	result, err := s.llm.Complete(ctx, "你是一个专业的技术文档撰写助手。只输出结果，不要添加解释或标记。", taskPrompt)
	if err != nil {
		return nil, fmt.Errorf("page AI failed: %w", err)
	}
	return &PageAIResponse{Result: strings.TrimSpace(result)}, nil
}

// ==================== Tool Execution ====================

func (s *AIService) ExecuteTool(name string, rawInput json.RawMessage, actx *AIContext) (any, error) {
	var args map[string]interface{}
	if err := json.Unmarshal(rawInput, &args); err != nil {
		args = map[string]interface{}{}
	}

	switch name {
	case "search_issues":
		return s.toolSearchIssues(args, actx)
	case "create_issue":
		return s.toolCreateIssue(args, actx)
	case "get_issue":
		return s.toolGetIssue(args)
	case "update_issue":
		return s.toolUpdateIssue(args, actx)
	case "get_project_stats":
		return s.toolGetProjectStats(args, actx)
	case "get_issues_summary":
		return s.toolGetIssuesSummary(args, actx)
	case "list_members":
		return s.toolListMembers(args, actx)
	case "list_issue_types":
		return s.toolListIssueTypes(args, actx)
	case "list_states":
		return s.toolListStates(args, actx)
	case "list_labels":
		return s.toolListLabels(args, actx)
	case "list_cycles":
		return s.toolListCycles(args, actx)
	case "get_cycle_progress":
		return s.toolGetCycleProgress(args)
	case "list_modules":
		return s.toolListModules(args, actx)
	case "get_issue_activities":
		return s.toolGetIssueActivities(args)
	case "add_comment":
		return s.toolAddComment(args, actx)
	case "list_releases":
		return s.toolListReleases(args, actx)
	case "list_pages":
		return s.toolListPages(args, actx)
	case "web_search":
		return s.toolWebSearch(args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

// ==================== Tool Implementations ====================

func (s *AIService) toolSearchIssues(args map[string]interface{}, actx *AIContext) (any, error) {
	projectID := getUintArg(args, "project_id", actx.ProjectID)
	query := getStrArg(args, "query", "")
	priority := getStrArg(args, "priority", "")
	limit := getIntArg(args, "limit", 20)

	dbQuery := s.db.Model(&model.Issue{}).Where("project_id = ?", projectID)
	if query != "" {
		dbQuery = dbQuery.Where("name ILIKE ?", "%"+query+"%")
	}
	if priority != "" {
		dbQuery = dbQuery.Where("priority = ?", priority)
	}
	if stateID := getUintArg(args, "state_id", 0); stateID > 0 {
		dbQuery = dbQuery.Where("state_id = ?", stateID)
	}
	if assigneeID := getUintArg(args, "assignee_id", 0); assigneeID > 0 {
		dbQuery = dbQuery.Where("id IN (SELECT issue_id FROM issue_assignees WHERE user_id = ?)", assigneeID)
	}
	if typeID := getUintArg(args, "type_id", 0); typeID > 0 {
		dbQuery = dbQuery.Where("issue_type_id = ?", typeID)
	}

	var issues []model.Issue
	if err := dbQuery.Order("sequence_id DESC").Limit(limit).Find(&issues).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(issues))
	for i, iss := range issues {
		result[i] = map[string]interface{}{
			"id":          iss.ID,
			"name":        iss.Name,
			"sequence_id": iss.SequenceID,
			"priority":    iss.Priority,
			"state_id":    iss.StateID,
		}
	}
	return result, nil
}

func (s *AIService) toolCreateIssue(args map[string]interface{}, actx *AIContext) (any, error) {
	projectID := getUintArg(args, "project_id", actx.ProjectID)
	workspaceID := getUintArg(args, "workspace_id", actx.WorkspaceID)
	name := getStrArg(args, "name", "")

	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	if !s.isProjectMember(projectID, actx.UserID) {
		return nil, fmt.Errorf("no permission to create issues in this project")
	}

	iss := &model.Issue{
		Name:            name,
		DescriptionHTML: getStrArg(args, "description", ""),
		Priority:        getStrArg(args, "priority", "none"),
		ProjectID:       projectID,
		WorkspaceID:     workspaceID,
	}
	iss.CreatedByID = &actx.UserID
	if typeID := getUintArg(args, "type_id", 0); typeID > 0 {
		iss.IssueTypeID = &typeID
	}
	if stateID := getUintArg(args, "state_id", 0); stateID > 0 {
		iss.StateID = stateID
	}

	if err := s.db.Create(iss).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"id": iss.ID, "name": iss.Name, "sequence_id": iss.SequenceID}, nil
}

func (s *AIService) toolGetIssue(args map[string]interface{}) (any, error) {
	id := getUintArg(args, "issue_id", 0)
	var iss model.Issue
	if err := s.db.Preload("State").Preload("IssueType").First(&iss, id).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id": iss.ID, "name": iss.Name, "sequence_id": iss.SequenceID,
		"priority": iss.Priority, "state": iss.State.Name,
		"type": iss.IssueType.Name,
	}, nil
}

func (s *AIService) toolUpdateIssue(args map[string]interface{}, actx *AIContext) (any, error) {
	id := getUintArg(args, "issue_id", 0)

	var iss model.Issue
	if err := s.db.First(&iss, id).Error; err != nil {
		return nil, err
	}

	if !s.isProjectMember(iss.ProjectID, actx.UserID) {
		return nil, fmt.Errorf("no permission to update issues in this project")
	}

	updates := map[string]interface{}{}
	if v := getStrArg(args, "name", ""); v != "" {
		updates["name"] = v
	}
	if v := getStrArg(args, "priority", ""); v != "" {
		updates["priority"] = v
	}
	if v := getUintArg(args, "state_id", 0); v > 0 {
		updates["state_id"] = v
	}
	if v := getStrArg(args, "description", ""); v != "" {
		updates["description_html"] = v
	}
	if len(updates) == 0 {
		return nil, fmt.Errorf("no updates provided")
	}
	if err := s.db.Model(&model.Issue{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"updated": true, "issue_id": id}, nil
}

func (s *AIService) toolGetProjectStats(args map[string]interface{}, actx *AIContext) (any, error) {
	pid := getUintArg(args, "project_id", actx.ProjectID)
	stats, err := s.getProjectStats(pid)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (s *AIService) toolGetIssuesSummary(args map[string]interface{}, actx *AIContext) (any, error) {
	pid := getUintArg(args, "project_id", actx.ProjectID)
	summary, err := s.getIssuesSummaryData(pid)
	if err != nil {
		return nil, err
	}
	return summary, nil
}

func (s *AIService) toolListMembers(args map[string]interface{}, actx *AIContext) (any, error) {
	pid := getUintArg(args, "project_id", actx.ProjectID)
	var members []model.ProjectMember
	s.db.Preload("User").Where("project_id = ?", pid).Find(&members)
	result := make([]map[string]interface{}, len(members))
	for i, m := range members {
		result[i] = map[string]interface{}{"user_id": m.UserID, "display_name": m.User.DisplayName, "username": m.User.Username, "role": m.Role}
	}
	return result, nil
}

func (s *AIService) toolListIssueTypes(args map[string]interface{}, actx *AIContext) (any, error) {
	wid := getUintArg(args, "workspace_id", actx.WorkspaceID)
	var types []model.IssueType
	s.db.Where("workspace_id = ? AND is_active = ? AND project_id IS NULL", wid, true).Find(&types)
	result := make([]map[string]interface{}, len(types))
	for i, t := range types {
		result[i] = map[string]interface{}{"id": t.ID, "name": t.Name, "color": t.Color, "level": t.Level}
	}
	return result, nil
}

func (s *AIService) toolListStates(args map[string]interface{}, actx *AIContext) (any, error) {
	pid := getUintArg(args, "project_id", actx.ProjectID)
	var states []model.State
	s.db.Where("project_id = ? AND is_active = ?", pid, true).Order("sequence").Find(&states)
	result := make([]map[string]interface{}, len(states))
	for i, st := range states {
		result[i] = map[string]interface{}{"id": st.ID, "name": st.Name, "color": st.Color, "group": st.Group}
	}
	return result, nil
}

func (s *AIService) toolListLabels(args map[string]interface{}, actx *AIContext) (any, error) {
	pid := getUintArg(args, "project_id", actx.ProjectID)
	var labels []model.Label
	s.db.Where("project_id = ?", pid).Find(&labels)
	result := make([]map[string]interface{}, len(labels))
	for i, l := range labels {
		result[i] = map[string]interface{}{"id": l.ID, "name": l.Name, "color": l.Color}
	}
	return result, nil
}

func (s *AIService) toolListCycles(args map[string]interface{}, actx *AIContext) (any, error) {
	pid := getUintArg(args, "project_id", actx.ProjectID)
	var cycles []model.Cycle
	s.db.Where("project_id = ?", pid).Order("start_date DESC").Limit(10).Find(&cycles)
	result := make([]map[string]interface{}, len(cycles))
	for i, c := range cycles {
		result[i] = map[string]interface{}{"id": c.ID, "name": c.Name, "start_date": c.StartDate, "end_date": c.EndDate}
	}
	return result, nil
}

func (s *AIService) toolGetCycleProgress(args map[string]interface{}) (any, error) {
	id := getUintArg(args, "cycle_id", 0)
	var total, done int64
	s.db.Model(&model.IssueCycle{}).Where("cycle_id = ?", id).Count(&total)
	s.db.Model(&model.IssueCycle{}).Joins("JOIN issues ON issue_cycles.issue_id = issues.id").
		Where("issue_cycles.cycle_id = ? AND issues.state_id IN (SELECT id FROM states WHERE \"group\" = 'completed')", id).Count(&done)
	return map[string]interface{}{"cycle_id": id, "total_issues": total, "completed_issues": done}, nil
}

func (s *AIService) toolListModules(args map[string]interface{}, actx *AIContext) (any, error) {
	pid := getUintArg(args, "project_id", actx.ProjectID)
	var modules []model.Module
	s.db.Where("project_id = ?", pid).Find(&modules)
	result := make([]map[string]interface{}, len(modules))
	for i, m := range modules {
		result[i] = map[string]interface{}{"id": m.ID, "name": m.Name}
	}
	return result, nil
}

func (s *AIService) toolGetIssueActivities(args map[string]interface{}) (any, error) {
	id := getUintArg(args, "issue_id", 0)
	var activities []model.IssueActivity
	s.db.Where("issue_id = ?", id).Order("created_at DESC").Limit(20).Find(&activities)
	result := make([]map[string]interface{}, len(activities))
	for i, a := range activities {
		result[i] = map[string]interface{}{"verb": a.Verb, "field": a.Field, "old_value": a.OldValue, "new_value": a.NewValue, "created_at": a.CreatedAt}
	}
	return result, nil
}

func (s *AIService) toolAddComment(args map[string]interface{}, actx *AIContext) (any, error) {
	issueID := getUintArg(args, "issue_id", 0)
	body := getStrArg(args, "body", "")
	if issueID == 0 || body == "" {
		return nil, fmt.Errorf("issue_id and body are required")
	}

	var iss model.Issue
	if err := s.db.First(&iss, issueID).Error; err != nil {
		return nil, err
	}
	if !s.isProjectMember(iss.ProjectID, actx.UserID) {
		return nil, fmt.Errorf("no permission to add comments in this project")
	}

	comment := &model.Comment{IssueID: issueID, Body: body}
	comment.CreatedByID = &actx.UserID
	if err := s.db.Create(comment).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"id": comment.ID, "created": true}, nil
}

func (s *AIService) toolListReleases(args map[string]interface{}, actx *AIContext) (any, error) {
	pid := getUintArg(args, "project_id", actx.ProjectID)
	var releases []model.Release
	s.db.Where("project_id = ?", pid).Find(&releases)
	result := make([]map[string]interface{}, len(releases))
	for i, r := range releases {
		result[i] = map[string]interface{}{"id": r.ID, "name": r.Name, "version": r.Version, "status": r.Status}
	}
	return result, nil
}

func (s *AIService) toolListPages(args map[string]interface{}, actx *AIContext) (any, error) {
	pid := getUintArg(args, "project_id", actx.ProjectID)
	var pages []model.Page
	s.db.Where("project_id = ? AND archived_at IS NULL", pid).Find(&pages)
	result := make([]map[string]interface{}, len(pages))
	for i, p := range pages {
		result[i] = map[string]interface{}{"id": p.ID, "title": p.Title, "depth": p.Depth}
	}
	return result, nil
}

// ==================== Web Search Tool ====================

type searchResultData struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

var (
	ddgLinkRe    = regexp.MustCompile(`<a[^>]*href="([^"]*)"[^>]*>([^<]+)</a>`)
	ddgSnippetRe = regexp.MustCompile(`<td[^>]*class="result-snippet"[^>]*>([\s\S]*?)</td>`)
	ddgResultRe  = regexp.MustCompile(`<tr[^>]*class="result[^"]*"[\s\S]*?</tr>`)
)

func (s *AIService) toolWebSearch(args map[string]interface{}) (any, error) {
	query := getStrArg(args, "query", "")
	if query == "" {
		return nil, fmt.Errorf("query is required for web_search")
	}
	num := getIntArg(args, "num", 5)
	if num > 10 {
		num = 10
	}

	searchURL := fmt.Sprintf("https://lite.duckduckgo.com/lite/?q=%s", url.QueryEscape(query))

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("web_search request failed: %w", err)
	}
	req.Header.Set("User-Agent", "reqmango/1.0 (AI Assistant Bot; +https://github.com/reqmango)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("web_search failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return nil, fmt.Errorf("web_search read failed: %w", err)
	}

	results := parseDuckDuckGoLite(string(body), num)

	return map[string]interface{}{
		"query":   query,
		"results": results,
	}, nil
}

func parseDuckDuckGoLite(html string, maxResults int) []searchResultData {
	var results []searchResultData

	rows := ddgResultRe.FindAllString(html, -1)
	for _, row := range rows {
		if len(results) >= maxResults {
			break
		}

		var result searchResultData

		links := ddgLinkRe.FindAllStringSubmatch(row, -1)
		for _, m := range links {
			if len(m) >= 3 {
				href := strings.TrimSpace(m[1])
				title := strings.TrimSpace(m[2])
				if title != "" && strings.HasPrefix(href, "http") && !strings.Contains(href, "duckduckgo.com") {
					result.Title = stripHTMLX(title)
					result.URL = href
					break
				}
			}
		}

		snippets := ddgSnippetRe.FindAllStringSubmatch(row, -1)
		if len(snippets) > 0 && len(snippets[0]) >= 2 {
			result.Snippet = strings.TrimSpace(stripHTMLX(snippets[0][1]))
		}

		if result.Title != "" {
			results = append(results, result)
		}
	}

	if len(results) == 0 {
		lines := strings.Split(html, "\n")
		for _, line := range lines {
			if len(results) >= maxResults {
				break
			}
			matches := ddgLinkRe.FindStringSubmatch(line)
			if len(matches) >= 3 {
				href := strings.TrimSpace(matches[1])
				title := strings.TrimSpace(matches[2])
				if strings.HasPrefix(href, "http") && !strings.Contains(href, "duckduckgo.com") {
					results = append(results, searchResultData{
						Title:   stripHTMLX(title),
						URL:     href,
						Snippet: "",
					})
				}
			}
		}
	}

	return results
}

func stripHTMLX(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return strings.TrimSpace(re.ReplaceAllString(s, ""))
}

// ==================== Helpers ====================

func getUintArg(args map[string]interface{}, key string, defaultVal uint64) uint64 {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return uint64(n)
		case json.Number:
			val, _ := n.Int64()
			return uint64(val)
		}
	}
	return defaultVal
}

func getStrArg(args map[string]interface{}, key, defaultVal string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

func getIntArg(args map[string]interface{}, key string, defaultVal int) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case json.Number:
			val, _ := n.Int64()
			return int(val)
		}
	}
	return defaultVal
}

func filterReadOnlyTools(tools []llm.Tool) []llm.Tool {
	readOnlyNames := map[string]bool{
		"search_issues": true, "get_issue": true, "get_project_stats": true,
		"get_issues_summary": true, "list_members": true, "list_issue_types": true,
		"list_states": true, "list_labels": true, "list_cycles": true,
		"get_cycle_progress": true, "list_modules": true,
		"get_issue_activities": true, "list_releases": true, "list_pages": true,
	}
	filtered := make([]llm.Tool, 0)
	for _, t := range tools {
		if readOnlyNames[t.Name] {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func extractJSON(text string) map[string]interface{} {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		var result map[string]interface{}
		if json.Unmarshal([]byte(text[start:end+1]), &result) == nil {
			return result
		}
	}
	return nil
}
