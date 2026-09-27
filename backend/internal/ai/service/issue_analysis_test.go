package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func day(s string) *time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return &t
}

func sampleIssueContext() IssueAnalysisContext {
	now, _ := time.Parse("2006-01-02", "2026-09-27")
	return IssueAnalysisContext{
		Identifier:      "CORE-704",
		Name:            "登录页 500",
		Priority:        "urgent",
		State:           "进行中",
		StateGroup:      "started",
		TargetDate:      day("2026-09-20"),
		DaysSinceUpdate: 9,
		SubIssuesTotal:  4,
		SubIssuesDone:   1,
		Relations:       []string{"被阻塞于 CORE-12 会话服务重构（进行中）"},
		RecentComments:  []string{"张三：还在等运维给日志"},
		Description:     "用户登录时偶发 500",
		Now:             now,
	}
}

func TestNormalizeIssueAnalysisMode(t *testing.T) {
	assert.Equal(t, IssueAnalysisSummary, NormalizeIssueAnalysisMode(""))
	assert.Equal(t, IssueAnalysisSummary, NormalizeIssueAnalysisMode("bogus"))
	assert.Equal(t, IssueAnalysisRisk, NormalizeIssueAnalysisMode("risk"))
	assert.Equal(t, IssueAnalysisNextSteps, NormalizeIssueAnalysisMode("next_steps"))
}

func TestIssueRiskSignals_DetectsFacts(t *testing.T) {
	signals := issueRiskSignals(sampleIssueContext())
	titles := make([]string, 0, len(signals))
	for _, s := range signals {
		titles = append(titles, s.Title)
	}
	assert.Contains(t, titles, "已逾期 7 天")
	assert.Contains(t, titles, "无负责人")
	assert.Contains(t, titles, "停滞 9 天未更新")
	assert.Contains(t, titles, "存在阻塞依赖")
	for _, s := range signals {
		if s.Title == "已逾期 7 天" {
			assert.Equal(t, "high", s.Level)
		}
	}
}

func TestIssueRiskSignals_NoneForHealthyOrClosedIssue(t *testing.T) {
	ic := sampleIssueContext()
	ic.Assignees = []string{"张三"}
	ic.TargetDate = day("2026-10-30")
	ic.DaysSinceUpdate = 1
	ic.Relations = nil
	assert.Empty(t, issueRiskSignals(ic))

	closed := sampleIssueContext()
	closed.StateGroup = "completed"
	assert.Empty(t, issueRiskSignals(closed))
}

func TestIssueAnalysisPrompt_DiffersByModeAndCarriesContext(t *testing.T) {
	ic := sampleIssueContext()
	summary := issueAnalysisPrompt(IssueAnalysisSummary, ic)
	risk := issueAnalysisPrompt(IssueAnalysisRisk, ic)
	next := issueAnalysisPrompt(IssueAnalysisNextSteps, ic)

	assert.NotEqual(t, summary, risk)
	assert.NotEqual(t, risk, next)
	for _, p := range []string{summary, risk, next} {
		assert.Contains(t, p, "CORE-704")
		assert.Contains(t, p, "2026-09-20")
		assert.Contains(t, p, "被阻塞于 CORE-12")
		assert.Contains(t, p, "还在等运维给日志")
		assert.Contains(t, p, "1/4")
	}
	assert.Contains(t, risk, `"risks"`)
	assert.Contains(t, risk, "已逾期 7 天")
	assert.Contains(t, next, `"next_steps"`)
	assert.NotContains(t, summary, `"risks"`)
}

func TestParseIssueAnalysis_Risk(t *testing.T) {
	raw := "```json\n{\"summary\":\"交付风险高\",\"risks\":[{\"level\":\"HIGH\",\"title\":\"逾期\",\"detail\":\"已过截止日\"},{\"level\":\"weird\",\"title\":\"x\",\"detail\":\"y\"}]}\n```"
	out := parseIssueAnalysis(IssueAnalysisRisk, raw, sampleIssueContext())
	assert.Equal(t, "risk", out.Mode)
	assert.Equal(t, "交付风险高", out.Summary)
	require.Len(t, out.Risks, 2)
	assert.Equal(t, "high", out.Risks[0].Level)
	assert.Equal(t, "medium", out.Risks[1].Level)
}

func TestParseIssueAnalysis_RiskFallsBackToSignalsOnBadJSON(t *testing.T) {
	out := parseIssueAnalysis(IssueAnalysisRisk, "模型胡言乱语", sampleIssueContext())
	assert.NotEmpty(t, out.Risks)
	assert.Equal(t, "high", out.Risks[0].Level)
}

func TestParseIssueAnalysis_NextSteps(t *testing.T) {
	raw := `{"summary":"先解阻塞","next_steps":["找运维要日志","补复现步骤"]}`
	out := parseIssueAnalysis(IssueAnalysisNextSteps, raw, sampleIssueContext())
	assert.Equal(t, "next_steps", out.Mode)
	assert.Equal(t, []string{"找运维要日志", "补复现步骤"}, out.NextSteps)
	assert.Empty(t, out.Risks)
}

func TestParseIssueAnalysis_Summary(t *testing.T) {
	raw := `{"summary":"登录 500 排查中","insights":["等日志","子项 1/4"]}`
	out := parseIssueAnalysis(IssueAnalysisSummary, raw, sampleIssueContext())
	assert.Equal(t, "登录 500 排查中", out.Summary)
	assert.Equal(t, []string{"等日志", "子项 1/4"}, out.Insights)
}
