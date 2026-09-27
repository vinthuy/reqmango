package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/model"
)

const (
	IssueAnalysisSummary   = "summary"
	IssueAnalysisRisk      = "risk"
	IssueAnalysisNextSteps = "next_steps"
)

const (
	issueStaleDays      = 7
	issueDueSoonDays    = 3
	issueDescMaxRunes   = 2000
	issueCommentsLimit  = 5
	issueCommentMaxRune = 200
)

// NormalizeIssueAnalysisMode maps unknown or empty modes to summary.
func NormalizeIssueAnalysisMode(mode string) string {
	switch mode {
	case IssueAnalysisRisk, IssueAnalysisNextSteps:
		return mode
	default:
		return IssueAnalysisSummary
	}
}

type AIRisk struct {
	Level  string `json:"level"` // high | medium | low
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// IssueAnalysisContext is the factual snapshot of one issue fed to the model.
type IssueAnalysisContext struct {
	ID              uint64
	Identifier      string
	Name            string
	Priority        string
	State           string
	StateGroup      string
	Assignees       []string
	AgentAssignee   string
	StartDate       *time.Time
	TargetDate      *time.Time
	DaysSinceUpdate int
	SubIssuesTotal  int
	SubIssuesDone   int
	Relations       []string
	RecentComments  []string
	Description     string
	Now             time.Time
}

func (ic IssueAnalysisContext) isClosed() bool {
	return ic.StateGroup == "completed" || ic.StateGroup == "cancelled"
}

func daysBetween(from, to time.Time) int {
	y1, m1, d1 := from.Date()
	y2, m2, d2 := to.Date()
	a := time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC)
	b := time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC)
	return int(b.Sub(a).Hours() / 24)
}

func isBlockingRelation(desc string) bool {
	lower := strings.ToLower(desc)
	for _, kw := range []string{"阻塞", "block", "依赖", "depend"} {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// issueRiskSignals derives rule-based risks from facts, so the model is anchored
// on real data and risk mode still returns something useful if the model fails.
func issueRiskSignals(ic IssueAnalysisContext) []AIRisk {
	if ic.isClosed() {
		return nil
	}
	var out []AIRisk
	if ic.TargetDate != nil {
		left := daysBetween(ic.Now, *ic.TargetDate)
		switch {
		case left < 0:
			out = append(out, AIRisk{Level: "high", Title: fmt.Sprintf("已逾期 %d 天", -left),
				Detail: fmt.Sprintf("截止日期 %s 已过，仍处于「%s」", ic.TargetDate.Format("2006-01-02"), ic.State)})
		case left <= issueDueSoonDays:
			out = append(out, AIRisk{Level: "medium", Title: fmt.Sprintf("%d 天内到期", left),
				Detail: fmt.Sprintf("截止日期 %s", ic.TargetDate.Format("2006-01-02"))})
		}
	}
	for _, r := range ic.Relations {
		if isBlockingRelation(r) {
			out = append(out, AIRisk{Level: "high", Title: "存在阻塞依赖", Detail: r})
			break
		}
	}
	if len(ic.Assignees) == 0 && ic.AgentAssignee == "" {
		level := "medium"
		if ic.Priority == "urgent" || ic.Priority == "high" {
			level = "high"
		}
		out = append(out, AIRisk{Level: level, Title: "无负责人", Detail: "没有人或 Agent 对该工作项负责"})
	}
	if ic.DaysSinceUpdate > issueStaleDays {
		out = append(out, AIRisk{Level: "medium", Title: fmt.Sprintf("停滞 %d 天未更新", ic.DaysSinceUpdate),
			Detail: fmt.Sprintf("最近一次更新在 %d 天前", ic.DaysSinceUpdate)})
	}
	return out
}

func formatDate(t *time.Time) string {
	if t == nil {
		return "未设置"
	}
	return t.Format("2006-01-02")
}

func bulletList(items []string, empty string) string {
	if len(items) == 0 {
		return "- " + empty
	}
	return "- " + strings.Join(items, "\n- ")
}

func issueFactsBlock(ic IssueAnalysisContext) string {
	assignees := strings.Join(ic.Assignees, "、")
	if assignees == "" {
		assignees = "无"
	}
	agent := ic.AgentAssignee
	if agent == "" {
		agent = "无"
	}
	subIssues := "无子项"
	if ic.SubIssuesTotal > 0 {
		subIssues = fmt.Sprintf("已完成 %d/%d", ic.SubIssuesDone, ic.SubIssuesTotal)
	}
	desc := strings.TrimSpace(ic.Description)
	if desc == "" {
		desc = "（无描述）"
	}
	return fmt.Sprintf(`## 工作项 %s
- 标题: %s
- 优先级: %s
- 状态: %s
- 负责人: %s
- 指派 Agent: %s
- 开始日期: %s
- 截止日期: %s
- 今天: %s
- 距上次更新: %d 天
- 子项: %s

## 关联关系
%s

## 最近评论（新→旧）
%s

## 描述
%s`,
		ic.Identifier, ic.Name, ic.Priority, ic.State, assignees, agent,
		formatDate(ic.StartDate), formatDate(ic.TargetDate), ic.Now.Format("2006-01-02"),
		ic.DaysSinceUpdate, subIssues,
		bulletList(ic.Relations, "无"),
		bulletList(ic.RecentComments, "无"),
		desc)
}

func issueAnalysisPrompt(mode string, ic IssueAnalysisContext) string {
	facts := issueFactsBlock(ic)
	switch mode {
	case IssueAnalysisRisk:
		signals := issueRiskSignals(ic)
		lines := make([]string, 0, len(signals))
		for _, s := range signals {
			lines = append(lines, fmt.Sprintf("[%s] %s：%s", s.Level, s.Title, s.Detail))
		}
		return fmt.Sprintf(`你是交付风险分析师。只针对下面这一条工作项评估交付风险，必须以给出的事实为依据，不要编造数据。

%s

## 系统已检测到的信号
%s

请用中文输出 JSON：
{
  "summary": "一句话风险结论（40字以内）",
  "risks": [{"level": "high|medium|low", "title": "风险名（10字以内）", "detail": "依据事实说明原因与影响"}]
}
按严重程度从高到低排列，最多 5 条；确实没有明显风险时 risks 返回空数组并在 summary 说明。`,
			facts, bulletList(lines, "无"))
	case IssueAnalysisNextSteps:
		return fmt.Sprintf(`你是项目经理助手。针对下面这一条工作项，给出接下来最该做的具体行动。

%s

请用中文输出 JSON：
{
  "summary": "一句话说明当前最关键的推进点（40字以内）",
  "next_steps": ["以动词开头的可执行动作，能指明负责人时写明"]
}
给 3–5 条，按先后顺序；不要泛泛而谈（如"加强沟通"），要能直接照做。`, facts)
	default:
		return fmt.Sprintf(`你是项目经理助手。请用几句话讲清楚下面这一条工作项的现状，让没跟进的人 30 秒看懂。

%s

请用中文输出 JSON：
{
  "summary": "现状摘要：要做什么、进展到哪、谁在负责（120字以内）",
  "insights": ["值得注意的关键事实，2–4 条"]
}
只陈述事实与现状，不做风险评估，也不给行动建议。`, facts)
	}
}

func normalizeRiskLevel(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "high", "高":
		return "high"
	case "low", "低":
		return "low"
	default:
		return "medium"
	}
}

func parseIssueAnalysis(mode string, content string, ic IssueAnalysisContext) *AIAnalyzeResponse {
	out := &AIAnalyzeResponse{Mode: mode, Insights: []string{}}
	var parsed struct {
		Summary   string   `json:"summary"`
		Insights  []string `json:"insights"`
		NextSteps []string `json:"next_steps"`
		Risks     []AIRisk `json:"risks"`
	}
	jsonStr := content
	if start, end := strings.Index(content, "{"), strings.LastIndex(content, "}"); start >= 0 && end > start {
		jsonStr = content[start : end+1]
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		runes := []rune(strings.TrimSpace(content))
		out.Summary = string(runes[:minX(len(runes), 200)])
		if mode == IssueAnalysisRisk {
			out.Risks = issueRiskSignals(ic)
		}
		return out
	}

	out.Summary = parsed.Summary
	switch mode {
	case IssueAnalysisRisk:
		out.Risks = make([]AIRisk, 0, len(parsed.Risks))
		for _, r := range parsed.Risks {
			r.Level = normalizeRiskLevel(r.Level)
			out.Risks = append(out.Risks, r)
		}
	case IssueAnalysisNextSteps:
		out.NextSteps = append([]string{}, parsed.NextSteps...)
	default:
		out.Insights = append(out.Insights, parsed.Insights...)
	}
	return out
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}

func (s *AIService) loadIssueAnalysisContext(issue *model.Issue, identifier string) IssueAnalysisContext {
	ic := IssueAnalysisContext{
		ID:              issue.ID,
		Identifier:      fmt.Sprintf("%s-%d", identifier, issue.SequenceID),
		Name:            issue.Name,
		Priority:        issue.Priority,
		StartDate:       issue.StartDate,
		TargetDate:      issue.TargetDate,
		Now:             time.Now(),
		DaysSinceUpdate: daysBetween(issue.UpdatedAt, time.Now()),
	}
	if issue.DescriptionStripped != nil && *issue.DescriptionStripped != "" {
		ic.Description = truncateRunes(*issue.DescriptionStripped, issueDescMaxRunes)
	} else {
		ic.Description = truncateRunes(stripHTMLX(issue.DescriptionHTML), issueDescMaxRunes)
	}

	var state model.State
	if issue.StateID > 0 && s.db.First(&state, issue.StateID).Error == nil {
		ic.State, ic.StateGroup = state.Name, state.Group
	}

	s.db.Table("users").
		Joins("JOIN issue_assignees ia ON ia.user_id = users.id").
		Where("ia.issue_id = ?", issue.ID).
		Pluck("users.display_name", &ic.Assignees)

	if issue.AgentAssigneeID != nil {
		var agent model.Agent
		if s.db.Select("name").First(&agent, *issue.AgentAssigneeID).Error == nil {
			ic.AgentAssignee = agent.Name
		}
	}

	var subs []struct{ Group string }
	s.db.Table("issues").Select("states.\"group\" AS \"group\"").
		Joins("JOIN states ON states.id = issues.state_id").
		Where("issues.parent_id = ? AND issues.deleted_at IS NULL", issue.ID).
		Scan(&subs)
	ic.SubIssuesTotal = len(subs)
	for _, sub := range subs {
		if sub.Group == "completed" {
			ic.SubIssuesDone++
		}
	}

	ic.Relations = s.describeIssueRelations(issue.ID, identifier)

	var comments []model.Comment
	s.db.Preload("Author").Where("issue_id = ?", issue.ID).
		Order("created_at DESC").Limit(issueCommentsLimit).Find(&comments)
	for _, c := range comments {
		author := "匿名"
		if c.Author != nil && c.Author.DisplayName != "" {
			author = c.Author.DisplayName
		}
		body := truncateRunes(stripHTMLX(c.Body), issueCommentMaxRune)
		if body != "" {
			ic.RecentComments = append(ic.RecentComments, fmt.Sprintf("%s：%s", author, body))
		}
	}
	return ic
}

func (s *AIService) describeIssueRelations(issueID uint64, identifier string) []string {
	var rels []model.IssueRelation
	s.db.Preload("RelationType").Preload("Issue").Preload("RelatedIssue").
		Where("issue_id = ? OR related_issue_id = ?", issueID, issueID).
		Limit(20).Find(&rels)

	var out []string
	for _, r := range rels {
		verb, other := r.RelationType.OutwardName, r.RelatedIssue
		if r.RelatedIssueID == issueID {
			verb, other = r.RelationType.InwardName, r.Issue
		}
		if verb == "" {
			verb = r.RelationType.Name
		}
		var otherState model.State
		stateName := ""
		if other.StateID > 0 && s.db.Select("name").First(&otherState, other.StateID).Error == nil {
			stateName = "（" + otherState.Name + "）"
		}
		out = append(out, fmt.Sprintf("%s %s-%d %s%s", verb, identifier, other.SequenceID, other.Name, stateName))
	}
	return out
}

func (s *AIService) analyzeIssue(ctx context.Context, actx *AIContext) (*AIAnalyzeResponse, error) {
	var issue model.Issue
	q := s.db.Where("id = ?", actx.IssueID)
	if actx.ProjectID > 0 {
		q = q.Where("project_id = ?", actx.ProjectID)
	}
	if err := q.First(&issue).Error; err != nil {
		return nil, fmt.Errorf("get issue: %w", err)
	}

	identifier := actx.ProjectIdentifier
	if identifier == "" {
		var project model.Project
		if s.db.Select("identifier").First(&project, issue.ProjectID).Error == nil {
			identifier = project.Identifier
		}
	}

	mode := NormalizeIssueAnalysisMode(actx.AnalysisMode)
	ic := s.loadIssueAnalysisContext(&issue, identifier)
	content, err := s.llm.Complete(ctx,
		"你是一个资深项目经理助手。请输出严格的 JSON 格式，不要添加任何 markdown 标记。",
		issueAnalysisPrompt(mode, ic))
	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	result := parseIssueAnalysis(mode, content, ic)
	result.Stats = map[string]interface{}{
		"issue_id":   issue.ID,
		"issue_name": issue.Name,
		"priority":   issue.Priority,
		"state":      ic.State,
		"scope":      "issue",
	}
	return result, nil
}
