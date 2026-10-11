// Command ai_search_eval measures how well AI search turns natural language
// into the right issue set, against a real model and a real Postgres.
//
// It seeds a throwaway workspace inside a transaction, runs every case through
// AIService.Search, and rolls back, so the database is left untouched.
//
//	go run ./cmd/ai_search_eval            # uses DATABASE_URL / AI_API_KEY / AI_MODEL / AI_BASE_URL
//	go run ./cmd/ai_search_eval -run 登录 -v
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
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

type fixtureIssue struct {
	name      string
	priority  string
	state     string // state group
	labels    []string
	assignee  string // "alice" | "bob" | ""
	dueInDays *int
}

func days(n int) *int { return &n }

// Alice is the current user. Sequence ids are the 1-based positions below.
var fixture = []fixtureIssue{
	{"登录页白屏", "urgent", "started", []string{"前端"}, "alice", days(-2)},
	{"支付接口超时", "high", "started", []string{"后端", "性能"}, "bob", days(3)},
	{"首页加载慢", "medium", "unstarted", []string{"前端", "性能"}, "alice", nil},
	{"导出报表失败", "high", "backlog", []string{"后端"}, "", nil},
	{"修改密码邮件未发送", "low", "completed", []string{"后端"}, "alice", days(-5)},
	{"深色模式适配", "none", "cancelled", []string{"前端"}, "", nil},
	{"登录验证码不刷新", "medium", "unstarted", []string{"前端"}, "bob", days(1)},
	{"数据库索引优化", "low", "backlog", []string{"后端", "性能"}, "alice", nil},
}

type evalCase struct {
	query string
	// Any of these sequence-id sets counts as correct (queries like "高优先级"
	// are legitimately read as high-only or high+urgent).
	want [][]int
}

var cases = []evalCase{
	{"紧急的工作项", [][]int{{1}}},
	{"高优先级的", [][]int{{2, 4}, {1, 2, 4}}},
	{"我负责的", [][]int{{1, 3, 5, 8}}},
	{"我负责的还没完成的", [][]int{{1, 3, 8}}},
	{"进行中的", [][]int{{1, 2}}},
	{"已完成的", [][]int{{5}}},
	{"标题里有登录的", [][]int{{1, 7}}},
	{"前端相关的", [][]int{{1, 3, 6, 7}}},
	{"带性能标签的", [][]int{{2, 3, 8}}},
	{"分配给鲍勃的", [][]int{{2, 7}}},
	{"没有负责人的", [][]int{{4, 6}}},
	{"积压中的后端任务", [][]int{{4, 8}}},
	{"截止日期已过但还没完成的", [][]int{{1}}},
	{"截止日期在今天之后的", [][]int{{2, 7}}},
	{"高优先级或紧急的未完成任务", [][]int{{1, 2, 4}}},
	{"低优先级的已完成工作", [][]int{{5}}},
	{"我创建的", [][]int{{1, 3, 4, 6, 8}}},
	{"鲍勃创建的", [][]int{{2, 5, 7}}},
	{"v1.0 里程碑里的", [][]int{{1, 2}}},
	{"没有关联里程碑的", [][]int{{5, 6, 7, 8}}},
}

// Bob created these sequence ids; Alice created the rest.
var createdByBob = map[int]bool{2: true, 5: true, 7: true}

// Milestone name → sequence ids.
var milestones = map[string][]int{"v1.0": {1, 2}, "v2.0": {3, 4}}

func main() {
	runPattern := flag.String("run", "", "only run cases whose query matches this regexp")
	verbose := flag.Bool("v", false, "print the generated RQL and explanation for every case")
	minPass := flag.Float64("min-pass", 0, "exit 1 when the pass rate is below this fraction (0-1)")
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
	actx, err := seed(tx)
	if err != nil {
		log.Fatalf("seed fixture: %v", err)
	}
	svc := aiservice.NewAIService(tx, client)

	var total, passed, viaRQL int
	for _, c := range cases {
		if filter != nil && !filter.MatchString(c.query) {
			continue
		}
		total++
		// A SQL error aborts the whole Postgres transaction; the savepoint keeps
		// one bad query from failing every later case.
		tx.SavePoint("eval_case")
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		start := time.Now()
		res, err := svc.Search(ctx, &aiservice.AISearchRequest{Query: c.query}, actx)
		cancel()
		tx.RollbackTo("eval_case")
		elapsed := time.Since(start).Round(100 * time.Millisecond)
		if err != nil {
			fmt.Printf("ERROR %-28s %v\n", c.query, err)
			continue
		}

		got := sequenceIDs(res.Issues)
		ok := matchesAny(got, c.want)
		source := "rql"
		switch {
		case res.RQL == "":
			source = "none"
		case strings.HasPrefix(res.RQL, "sequence_id"):
			source = "fallback"
		}
		if source == "rql" {
			viaRQL++
		}
		mark := "FAIL"
		if ok {
			mark = "PASS"
			passed++
		}
		fmt.Printf("%s  %-28s %-8s %6s  got=%v want=%v\n", mark, c.query, source, elapsed, got, c.want[0])
		if *verbose || !ok {
			fmt.Printf("      rql: %s\n      explanation: %s\n", res.RQL, oneLine(res.Explanation))
		}
	}

	if total == 0 {
		fmt.Println("no cases matched")
		return
	}
	rate := float64(passed) / float64(total)
	fmt.Printf("\npassed %d/%d (%.0f%%), real filter generated for %d/%d\n", passed, total, rate*100, viaRQL, total)
	if rate < *minPass {
		os.Exit(1)
	}
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

func seed(tx *gorm.DB) (*aiservice.AIContext, error) {
	suffix := fmt.Sprintf("%d", time.Now().UnixNano()%1_000_000)
	users := map[string]*model.User{
		"alice": {Email: "eval-alice-" + suffix + "@example.test", Username: "eval_alice_" + suffix, DisplayName: "爱丽丝", PasswordHash: "x"},
		"bob":   {Email: "eval-bob-" + suffix + "@example.test", Username: "eval_bob_" + suffix, DisplayName: "鲍勃", PasswordHash: "x"},
	}
	for _, key := range []string{"alice", "bob"} {
		if err := tx.Create(users[key]).Error; err != nil {
			return nil, err
		}
	}
	ws := model.Workspace{Name: "AI 搜索评测", Slug: "ai-search-eval-" + suffix, OwnerID: users["alice"].ID}
	if err := tx.Create(&ws).Error; err != nil {
		return nil, err
	}
	project := model.Project{Name: "评测项目", Identifier: "EVAL", WorkspaceID: ws.ID}
	if err := tx.Create(&project).Error; err != nil {
		return nil, err
	}
	for _, u := range users {
		if err := tx.Create(&model.ProjectMember{ProjectID: project.ID, UserID: u.ID, IsActive: true}).Error; err != nil {
			return nil, err
		}
	}

	states := map[string]uint64{}
	for i, st := range []struct{ name, group string }{
		{"待办", "backlog"}, {"未开始", "unstarted"}, {"进行中", "started"}, {"已完成", "completed"}, {"已取消", "cancelled"},
	} {
		s := model.State{Name: st.name, Group: st.group, Sequence: i + 1, IsActive: true, ProjectID: &project.ID, WorkspaceID: ws.ID}
		if err := tx.Create(&s).Error; err != nil {
			return nil, err
		}
		states[st.group] = s.ID
	}
	labels := map[string]uint64{}
	for _, name := range []string{"前端", "后端", "性能"} {
		l := model.Label{Name: name, ProjectID: project.ID, WorkspaceID: ws.ID}
		if err := tx.Create(&l).Error; err != nil {
			return nil, err
		}
		labels[name] = l.ID
	}

	today := time.Now().Truncate(24 * time.Hour)
	issueIDs := map[int]uint64{}
	for i, fi := range fixture {
		creator := users["alice"].ID
		if createdByBob[i+1] {
			creator = users["bob"].ID
		}
		issue := model.Issue{
			Name: fi.name, Priority: fi.priority, SequenceID: i + 1,
			ProjectID: project.ID, WorkspaceID: ws.ID, StateID: states[fi.state],
		}
		issue.CreatedByID = &creator
		if fi.dueInDays != nil {
			due := today.AddDate(0, 0, *fi.dueInDays)
			issue.TargetDate = &due
		}
		if err := tx.Create(&issue).Error; err != nil {
			return nil, err
		}
		issueIDs[i+1] = issue.ID
		for _, l := range fi.labels {
			if err := tx.Create(&model.IssueLabel{IssueID: issue.ID, LabelID: labels[l]}).Error; err != nil {
				return nil, err
			}
		}
		if u := users[fi.assignee]; u != nil {
			if err := tx.Create(&model.IssueAssignee{IssueID: issue.ID, UserID: u.ID}).Error; err != nil {
				return nil, err
			}
		}
	}

	for _, name := range []string{"v1.0", "v2.0"} {
		rel := model.Release{Name: name, Version: name, ProjectID: project.ID}
		if err := tx.Create(&rel).Error; err != nil {
			return nil, err
		}
		for _, seq := range milestones[name] {
			if err := tx.Create(&model.ReleaseIssue{ReleaseID: rel.ID, IssueID: issueIDs[seq]}).Error; err != nil {
				return nil, err
			}
		}
	}

	return &aiservice.AIContext{
		WorkspaceID: ws.ID, WorkspaceName: ws.Name,
		ProjectID: project.ID, ProjectName: project.Name, ProjectIdentifier: project.Identifier,
		UserID: users["alice"].ID,
	}, nil
}

func sequenceIDs(issues []map[string]interface{}) []int {
	out := make([]int, 0, len(issues))
	for _, iss := range issues {
		switch v := iss["sequence_id"].(type) {
		case int:
			out = append(out, v)
		case float64:
			out = append(out, int(v))
		case int64:
			out = append(out, int(v))
		}
	}
	sort.Ints(out)
	return out
}

func matchesAny(got []int, wants [][]int) bool {
	for _, want := range wants {
		w := append([]int(nil), want...)
		sort.Ints(w)
		if fmt.Sprint(got) == fmt.Sprint(w) {
			return true
		}
	}
	return false
}

func oneLine(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > 160 {
		return string(r[:160]) + "…"
	}
	return s
}
