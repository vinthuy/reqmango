package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// IntakeSpecDraft is a structured requirement spec drafted from an intake request.
type IntakeSpecDraft struct {
	Summary            string   `json:"summary"`
	Background         string   `json:"background"`
	UserStories        []string `json:"user_stories"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	InScope            []string `json:"in_scope"`
	OutOfScope         []string `json:"out_of_scope"`
	OpenQuestions      []string `json:"open_questions"`
	SuggestedPriority  string   `json:"suggested_priority"`
	// Source is "ai" when produced by the model, "template" for the offline fallback.
	Source string `json:"source"`
}

// DraftIntakeSpec turns a raw intake request into a structured spec. It never
// fails because the model is unavailable: callers get a template draft instead.
func (s *AIService) DraftIntakeSpec(ctx context.Context, title, description, submitter string) *IntakeSpecDraft {
	prompt := fmt.Sprintf(`把下面这条外部提交的需求整理成结构化需求规格。

标题: %s
提交人: %s
原始描述:
%s

只输出 JSON（不要 markdown 代码块），字段:
{
  "summary": "一句话概括要解决的问题（中文，40字内）",
  "background": "背景与动机，2-3 句",
  "user_stories": ["作为<角色>，我希望<能力>，以便<价值>"],
  "acceptance_criteria": ["Given … When … Then …"],
  "in_scope": ["本次要做的"],
  "out_of_scope": ["明确不做的"],
  "open_questions": ["需要与提交人或团队确认的问题"],
  "suggested_priority": "urgent|high|medium|low"
}
user_stories 1-3 条，acceptance_criteria 3-6 条，其它列表 0-4 条。`, title, orDash(submitter), description)

	if s.llm != nil {
		if out, err := s.llm.Complete(ctx, "你是资深产品经理，擅长把模糊需求写成可验收的规格。只输出JSON。", prompt); err == nil {
			if d := parseIntakeSpec(out); d != nil {
				d.Source = "ai"
				return d
			}
		}
	}
	return templateIntakeSpec(title, description)
}

func parseIntakeSpec(text string) *IntakeSpecDraft {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return nil
	}
	var d IntakeSpecDraft
	if json.Unmarshal([]byte(text[start:end+1]), &d) != nil {
		return nil
	}
	if strings.TrimSpace(d.Summary) == "" || len(d.AcceptanceCriteria) == 0 {
		return nil
	}
	switch d.SuggestedPriority {
	case "urgent", "high", "medium", "low":
	default:
		d.SuggestedPriority = ""
	}
	return &d
}

func templateIntakeSpec(title, description string) *IntakeSpecDraft {
	desc := strings.TrimSpace(description)
	if desc == "" {
		desc = title
	}
	lower := strings.ToLower(title + " " + desc)
	priority := "medium"
	for _, kw := range []string{"崩溃", "无法", "失败", "报错", "crash", "fail", "error", "乱码"} {
		if strings.Contains(lower, kw) {
			priority = "high"
			break
		}
	}
	return &IntakeSpecDraft{
		Summary:    title,
		Background: desc,
		UserStories: []string{
			fmt.Sprintf("作为提交该需求的用户，我希望「%s」得到解决，以便顺利完成工作。", title),
		},
		AcceptanceCriteria: []string{
			fmt.Sprintf("Given 用户处于该需求描述的场景 When 执行相关操作 Then「%s」所述问题不再出现 / 能力可用", title),
			"Given 异常或边界输入 When 执行相同操作 Then 给出明确的提示而不是静默失败",
			"相关改动有自动化测试覆盖，并在发布说明中记录",
		},
		InScope:           []string{title},
		OutOfScope:        []string{},
		OpenQuestions:     []string{"影响范围与频率？是否有截图或复现步骤？", "期望的上线时间？"},
		SuggestedPriority: priority,
		Source:            "template",
	}
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}
