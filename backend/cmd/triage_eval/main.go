// Command triage_eval measures auto-triage quality against a real model and a
// real Postgres: for each case it creates a work item, runs
// AgentService.AutoTriage, and checks the agent's structured suggestions.
//
// Everything runs inside one transaction that is rolled back, so the database
// is left untouched.
//
//	go run ./cmd/triage_eval              # uses DATABASE_URL / AI_API_KEY / AI_MODEL / AI_BASE_URL
//	go run ./cmd/triage_eval -run 支付 -v
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/ai/llm"
	aiservice "github.com/reqmango/backend/internal/ai/service"
	"github.com/reqmango/backend/internal/config"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type triageCase struct {
	name, description string
	// Preset values model an already-triaged item; empty means unset.
	presetType, presetPriority string

	// Acceptable final values (suggested, else the current one). Nil skips the check.
	wantType     []string
	wantPriority []string
	// dupOf is the sequence id of an existing item the reply must point at.
	dupOf int
	// quiet cases may legitimately end with no actionable suggestion.
	quiet bool
}

var cases = []triageCase{
	{name: "支付页点击提交后一直转圈，线上大量用户无法付款",
		wantType: []string{"缺陷"}, wantPriority: []string{"urgent"}},
	{name: "登录接口偶发 500，约 5% 的请求失败", description: "监控显示从昨晚 22 点开始，/api/login 错误率约 5%。",
		wantType: []string{"缺陷"}, wantPriority: []string{"urgent", "high"}},
	{name: "生产数据库磁盘使用率已到 95%，预计两天内写满",
		wantType: []string{"任务", "缺陷"}, wantPriority: []string{"urgent", "high"}},
	{name: "首页加载需要 8 秒，用户投诉较多",
		wantType: []string{"缺陷", "优化"}, wantPriority: []string{"high", "medium"}},
	{name: "设置页按钮文案「保存」错写成「保寸」",
		wantType: []string{"缺陷"}, wantPriority: []string{"low", "medium"}},
	{name: "希望支持把工作项列表导出为 Excel", description: "运营同学每周需要手动整理工作项统计。",
		wantType: []string{"需求"}, wantPriority: []string{"low", "medium"}},
	{name: "新增深色模式",
		wantType: []string{"需求"}, wantPriority: []string{"low", "medium"}},
	{name: "为用户列表接口补充单元测试",
		wantType: []string{"任务"}, wantPriority: []string{"low", "medium"}},
	{name: "整理 Q4 迭代复盘文档",
		wantType: []string{"任务"}, wantPriority: []string{"low", "medium"}},
	{name: "导出报表时报 500 错误", description: "点击「导出」按钮后接口返回 500。",
		wantType: []string{"缺陷"}, dupOf: 1},
	{name: "搜索建议出来太慢", description: "输入关键字后要等两三秒才出现联想。",
		wantType: []string{"优化", "缺陷"}, dupOf: 2},
	{name: "移动端上传图片失败", description: "iOS 17 上选择图片后提示上传失败，安卓正常。",
		presetType: "缺陷", presetPriority: "high",
		wantType: []string{"缺陷"}, wantPriority: []string{"high", "urgent"}, quiet: true},
	{name: "有问题",
		wantPriority: []string{"none", "low", "medium"}, quiet: true},
}

// Existing items the agent can find while checking for duplicates.
var backlog = []struct{ name, description string }{
	{"导出报表时提示服务器错误", "在报表页点击导出，弹出「服务器错误」，复现率 100%。"},
	{"优化搜索框的联想速度", ""},
	{"增加工作项批量编辑", ""},
}

type result struct {
	checks map[string]bool
	notes  []string
}

func main() {
	runPattern := flag.String("run", "", "only run cases whose title matches this regexp")
	verbose := flag.Bool("v", false, "print the agent reply and suggestions for every case")
	minPass := flag.Float64("min-pass", 0, "exit 1 when the case pass rate is below this fraction (0-1)")
	flag.Parse()

	cfg := config.Load()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	client, err := llmClient(db, cfg)
	if err != nil {
		log.Fatal(err)
	}
	var filter *regexp.Regexp
	if *runPattern != "" {
		filter = regexp.MustCompile(*runPattern)
	}

	tx := db.Begin()
	defer tx.Rollback()
	// The server migrates on start; a dev database may lag behind the model.
	if err := tx.AutoMigrate(&model.Comment{}, &model.AgentActivity{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	fx, err := seed(tx)
	if err != nil {
		log.Fatalf("seed fixture: %v", err)
	}
	agents := aiservice.NewAgentService(tx, client, aiservice.NewAIService(tx, client))
	if _, err := agents.EnsurePMAgents(fx.workspaceID, fx.userID); err != nil {
		log.Fatalf("ensure agents: %v", err)
	}

	dims := []string{"ran", "actionable", "type", "priority", "duplicate"}
	passedDim, totalDim := map[string]int{}, map[string]int{}
	var total, passed int
	for i, c := range cases {
		if filter != nil && !filter.MatchString(c.name) {
			continue
		}
		total++
		tx.SavePoint("triage_case")
		start := time.Now()
		res, reply, sugs := runCase(tx, agents, fx, c, 100+i)
		tx.RollbackTo("triage_case")

		ok := true
		var marks []string
		for _, d := range dims {
			v, checked := res.checks[d]
			if !checked {
				continue
			}
			totalDim[d]++
			if v {
				passedDim[d]++
				marks = append(marks, d)
			} else {
				ok = false
				marks = append(marks, "✗"+d)
			}
		}
		mark := "FAIL"
		if ok {
			mark = "PASS"
			passed++
		}
		fmt.Printf("%s  %-30s %5s  %s\n", mark, c.name, time.Since(start).Round(time.Second), strings.Join(marks, " "))
		for _, n := range res.notes {
			fmt.Printf("      %s\n", n)
		}
		if *verbose || !ok {
			fmt.Printf("      suggestions: %s\n      reply: %s\n", sugs, oneLine(reply))
		}
	}

	if total == 0 {
		fmt.Println("no cases matched")
		return
	}
	fmt.Println()
	for _, d := range dims {
		if totalDim[d] > 0 {
			fmt.Printf("  %-10s %d/%d\n", d, passedDim[d], totalDim[d])
		}
	}
	rate := float64(passed) / float64(total)
	fmt.Printf("passed %d/%d (%.0f%%)\n", passed, total, rate*100)
	if rate < *minPass {
		os.Exit(1)
	}
}

func runCase(tx *gorm.DB, agents *aiservice.AgentService, fx *fixture, c triageCase, seq int) (result, string, string) {
	res := result{checks: map[string]bool{}}
	issue := model.Issue{
		Name: c.name, DescriptionHTML: c.description, Priority: "none", SequenceID: seq,
		ProjectID: fx.projectID, WorkspaceID: fx.workspaceID, StateID: fx.backlogState,
	}
	if c.presetPriority != "" {
		issue.Priority = c.presetPriority
	}
	if c.presetType != "" {
		id := fx.typeIDs[c.presetType]
		issue.IssueTypeID = &id
	}
	issue.CreatedByID = &fx.userID
	if err := tx.Create(&issue).Error; err != nil {
		res.checks["ran"] = false
		res.notes = append(res.notes, "create issue: "+err.Error())
		return res, "", ""
	}

	_, err := agents.AutoTriage(issue.ID, fx.userID)
	var comment model.Comment
	found := tx.Where("issue_id = ? AND agent_id IS NOT NULL", issue.ID).Order("id DESC").First(&comment).Error == nil
	res.checks["ran"] = err == nil && found
	if err != nil {
		res.notes = append(res.notes, "auto-triage: "+err.Error())
	}
	if !found {
		return res, "", ""
	}

	sugs, _ := aiservice.ParseIssueSuggestions(comment.Suggestions)
	finalType := c.presetType
	finalPriority := issue.Priority
	actionable := false
	for _, s := range sugs {
		if s.NoOp {
			continue
		}
		actionable = true
		switch s.Field {
		case "type":
			finalType = fx.typeNames[toUint(s.Value)]
		case "priority":
			finalPriority = fmt.Sprint(s.Value)
		}
	}

	if !c.quiet {
		res.checks["actionable"] = actionable
	}
	if c.wantType != nil {
		res.checks["type"] = contains(c.wantType, finalType)
		if !res.checks["type"] {
			res.notes = append(res.notes, fmt.Sprintf("type=%q want %v", finalType, c.wantType))
		}
	}
	if c.wantPriority != nil {
		res.checks["priority"] = contains(c.wantPriority, finalPriority)
		if !res.checks["priority"] {
			res.notes = append(res.notes, fmt.Sprintf("priority=%q want %v", finalPriority, c.wantPriority))
		}
	}
	if c.dupOf != 0 {
		ref := regexp.MustCompile(fmt.Sprintf(`(#|%s-)%d\b`, fx.identifier, c.dupOf))
		res.checks["duplicate"] = ref.MatchString(comment.Body)
	}
	return res, comment.Body, string(comment.Suggestions)
}

type fixture struct {
	workspaceID, projectID, userID, backlogState uint64
	identifier                                   string
	typeIDs                                      map[string]uint64
	typeNames                                    map[uint64]string
}

func seed(tx *gorm.DB) (*fixture, error) {
	suffix := fmt.Sprintf("%d", time.Now().UnixNano()%1_000_000)
	fx := &fixture{identifier: "TRI", typeIDs: map[string]uint64{}, typeNames: map[uint64]string{}}

	alice := model.User{Email: "triage-alice-" + suffix + "@example.test", Username: "triage_alice_" + suffix, DisplayName: "爱丽丝", PasswordHash: "x"}
	bob := model.User{Email: "triage-bob-" + suffix + "@example.test", Username: "triage_bob_" + suffix, DisplayName: "鲍勃", PasswordHash: "x"}
	for _, u := range []*model.User{&alice, &bob} {
		if err := tx.Create(u).Error; err != nil {
			return nil, err
		}
	}
	fx.userID = alice.ID
	ws := model.Workspace{Name: "分诊评测", Slug: "triage-eval-" + suffix, OwnerID: alice.ID}
	if err := tx.Create(&ws).Error; err != nil {
		return nil, err
	}
	fx.workspaceID = ws.ID
	project := model.Project{Name: "分诊评测项目", Identifier: fx.identifier, WorkspaceID: ws.ID}
	if err := tx.Create(&project).Error; err != nil {
		return nil, err
	}
	fx.projectID = project.ID
	for _, u := range []model.User{alice, bob} {
		if err := tx.Create(&model.WorkspaceMember{WorkspaceID: ws.ID, UserID: u.ID, Role: 20, IsActive: true}).Error; err != nil {
			return nil, err
		}
		if err := tx.Create(&model.ProjectMember{ProjectID: project.ID, UserID: u.ID, IsActive: true}).Error; err != nil {
			return nil, err
		}
	}

	for i, st := range []struct{ name, group string }{
		{"待办", "backlog"}, {"未开始", "unstarted"}, {"进行中", "started"}, {"已完成", "completed"}, {"已取消", "cancelled"},
	} {
		s := model.State{Name: st.name, Group: st.group, Sequence: i + 1, IsActive: true, ProjectID: &project.ID, WorkspaceID: ws.ID}
		if err := tx.Create(&s).Error; err != nil {
			return nil, err
		}
		if st.group == "backlog" {
			fx.backlogState = s.ID
		}
	}
	for i, name := range []string{"缺陷", "需求", "任务", "优化"} {
		t := model.IssueType{Name: name, IsActive: true, Sequence: i + 1, ProjectID: &project.ID, WorkspaceID: ws.ID}
		if err := tx.Create(&t).Error; err != nil {
			return nil, err
		}
		fx.typeIDs[name] = t.ID
		fx.typeNames[t.ID] = name
	}
	for _, name := range []string{"前端", "后端", "支付", "性能", "文档"} {
		if err := tx.Create(&model.Label{Name: name, ProjectID: project.ID, WorkspaceID: ws.ID}).Error; err != nil {
			return nil, err
		}
	}
	for i, b := range backlog {
		iss := model.Issue{Name: b.name, DescriptionHTML: b.description, Priority: "medium", SequenceID: i + 1,
			ProjectID: project.ID, WorkspaceID: ws.ID, StateID: fx.backlogState}
		iss.CreatedByID = &alice.ID
		if err := tx.Create(&iss).Error; err != nil {
			return nil, err
		}
	}
	return fx, nil
}

// llmClient prefers the server's env config and falls back to the first active
// workspace AI config, matching how the app resolves a key.
func llmClient(db *gorm.DB, cfg *config.Config) (*llm.LLMClient, error) {
	if cfg.AIAPIKey != "" {
		return llm.NewLLMClient(cfg.AIAPIKey, cfg.AIModel, cfg.AIBaseURL, cfg.AIProvider), nil
	}
	var ac model.AIConfig
	if err := db.Where("is_active = ? AND api_key <> ''", true).Order("id").First(&ac).Error; err != nil {
		return nil, fmt.Errorf("no AI key: set AI_API_KEY or configure a workspace AI key")
	}
	return llm.NewLLMClient(ac.APIKey, ac.Model, cfg.AIBaseURL, ac.Provider), nil
}

func toUint(v interface{}) uint64 {
	switch n := v.(type) {
	case uint64:
		return n
	case float64:
		return uint64(n)
	case int:
		return uint64(n)
	}
	return 0
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func oneLine(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > 600 {
		return string(r[:600]) + "…"
	}
	return s
}
