// Command route_audit probes every API route as a user with no memberships,
// using real IDs from a victim workspace, and reports which ones succeed.
//
// It mutates the database it runs against, so point it at a disposable copy:
//
//	ROUTE_AUDIT_DATABASE_URL=postgres://.../reqmango_audit go run ./cmd/route_audit
//
// Outbound HTTP is routed to a dead proxy so no LLM or webhook call leaves
// the machine.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/config"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/router"
	"github.com/reqmango/backend/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type result struct {
	Method, Path, URL string
	Outsider, Insider int
	Verdict           string
	Body              string
}

func main() {
	wsFlag := flag.Uint64("workspace", 1, "victim workspace id")
	out := flag.String("out", "route_audit.md", "report path")
	flag.Parse()

	dsn := os.Getenv("ROUTE_AUDIT_DATABASE_URL")
	if dsn == "" || !strings.Contains(dsn, "audit") {
		log.Fatal("set ROUTE_AUDIT_DATABASE_URL to a disposable database whose name contains 'audit'")
	}
	os.Setenv("HTTP_PROXY", "http://127.0.0.1:9")
	os.Setenv("HTTPS_PROXY", "http://127.0.0.1:9")
	os.Setenv("NO_PROXY", "")

	cfg := config.Load()
	cfg.DatabaseURL = dsn
	cfg.RateLimitRequests = 1 << 30
	cfg.AIAPIKey = ""
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatal(err)
	}

	a := &auditor{db: db, ws: *wsFlag, ids: map[string]uint64{}}
	a.pickFixtures()

	auth := service.NewAuthService(db, cfg)
	outsider := model.User{Email: fmt.Sprintf("route-audit-%d@example.com", time.Now().UnixNano()), Username: fmt.Sprintf("route_audit_%d", time.Now().UnixNano()), DisplayName: "route audit"}
	if err := db.Create(&outsider).Error; err != nil {
		log.Fatal(err)
	}
	outTok, _, _ := auth.GenerateTokenForTest(outsider.ID)
	inTok, _, _ := auth.GenerateTokenForTest(a.insider)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.RedirectTrailingSlash = false
	router.SetupRoutes(r, db, cfg)
	srv := httptest.NewServer(r)
	defer srv.Close()
	client := &http.Client{Timeout: 20 * time.Second}

	routes := r.Routes()
	rank := func(m string) int {
		switch m {
		case http.MethodGet:
			return 0
		case http.MethodDelete:
			return 2
		}
		return 1
	}
	sort.SliceStable(routes, func(i, j int) bool {
		ri, rj := rank(routes[i].Method), rank(routes[j].Method)
		if ri != rj {
			return ri < rj
		}
		if ri == 2 { // delete leaves before their parents
			return strings.Count(routes[i].Path, "/") > strings.Count(routes[j].Path, "/")
		}
		return routes[i].Path < routes[j].Path
	})

	var results []result
	for _, rt := range routes {
		if skip(rt.Path) {
			continue
		}
		url, ok, missing := a.fill(rt.Path)
		res := result{Method: rt.Method, Path: rt.Path}
		if !ok {
			res.Verdict = "unresolved:" + missing
			results = append(results, res)
			continue
		}
		res.URL = url
		q := fmt.Sprintf("workspace_id=%d&project_id=%d&issue_id=%d", a.ws, a.ids["project"], a.ids["issue"])
		full := srv.URL + url + "?" + q
		var body string
		res.Outsider, body = call(client, rt.Method, full, outTok)
		if rt.Method == http.MethodGet {
			res.Insider, _ = call(client, rt.Method, full, inTok)
		}
		res.Body = body
		res.Verdict = a.classify(rt.Method, res.Outsider, res.Insider, body)
		results = append(results, res)
		fmt.Printf("%-6s %3d %3d %-10s %s\n", rt.Method, res.Outsider, res.Insider, res.Verdict, rt.Path)
	}
	writeReport(*out, a, results)
}

type auditor struct {
	db      *gorm.DB
	ws      uint64
	insider uint64
	ids     map[string]uint64 // resolved by table name, plus "project"/"issue"
	markers []string
}

func (a *auditor) pickFixtures() {
	var project struct{ ID uint64; Identifier string }
	a.db.Raw(`SELECT p.id, p.identifier FROM projects p WHERE p.workspace_id = ? AND p.deleted_at IS NULL
		ORDER BY (SELECT count(*) FROM issues i WHERE i.project_id = p.id) DESC LIMIT 1`, a.ws).Scan(&project)
	var issue uint64
	a.db.Raw(`SELECT i.id FROM issues i WHERE i.project_id = ? AND i.deleted_at IS NULL
		ORDER BY (SELECT count(*) FROM comments c WHERE c.issue_id = i.id) DESC LIMIT 1`, project.ID).Scan(&issue)
	// The control user is a regular member promoted to admin in this
	// disposable copy, so its requests exercise the normal (non-superuser) path.
	a.db.Raw(`SELECT m.user_id FROM workspace_members m JOIN users u ON u.id = m.user_id
		WHERE m.workspace_id = ? AND m.is_active AND NOT u.is_superuser
		AND EXISTS (SELECT 1 FROM project_members pm WHERE pm.user_id = m.user_id AND pm.project_id = ?)
		ORDER BY m.role DESC, m.user_id LIMIT 1`, a.ws, project.ID).Scan(&a.insider)
	a.db.Exec(`UPDATE workspace_members SET role = 20 WHERE workspace_id = ? AND user_id = ?`, a.ws, a.insider)
	a.db.Exec(`UPDATE project_members SET role = 20 WHERE project_id = ? AND user_id = ?`, project.ID, a.insider)
	var slug string
	a.db.Raw(`SELECT slug FROM workspaces WHERE id = ?`, a.ws).Scan(&slug)
	if project.ID == 0 || issue == 0 || a.insider == 0 {
		log.Fatalf("victim workspace %d lacks a project, issue or non-superuser member", a.ws)
	}
	a.ids["project"], a.ids["projects"], a.ids["issue"], a.ids["issues"] = project.ID, project.ID, issue, issue
	a.markers = []string{`"` + slug + `"`, `"` + project.Identifier + `"`}
	fmt.Printf("victim workspace=%d project=%d (%s) issue=%d insider=%d\n", a.ws, project.ID, project.Identifier, issue, a.insider)
}

var paramRe = regexp.MustCompile(`:(\w+)`)

// fill substitutes every :param in path with a victim-owned ID.
func (a *auditor) fill(path string) (string, bool, string) {
	segs := strings.Split(path, "/")
	resolved := map[string]uint64{}
	for i, s := range segs {
		if !strings.HasPrefix(s, ":") {
			continue
		}
		name := s[1:]
		prev := ""
		if i > 0 {
			prev = segs[i-1]
		}
		id, ok := a.resolve(name, prev, resolved)
		if !ok {
			return "", false, name
		}
		segs[i] = strconv.FormatUint(id, 10)
	}
	return strings.Join(segs, "/"), true, ""
}

var overrides = map[string][]string{
	"workflows":           {"agent_workflows", "workflows"},
	"transitions":         {"state_transitions"},
	"edges":               {"workflow_edges", "state_transitions"},
	"nodes":               {"workflow_nodes"},
	"runs":                {"workflow_runs", "agent_loop_runs", "pipeline_runs"},
	"agent-members":       {"project_agent_members"},
	"views":               {"saved_views"},
	"dashboards":          {"saved_dashboards"},
	"charts":              {"metric_charts"},
	"messages":            {"chat_messages", "messages"},
	"templates":           {"project_templates", "agent_templates", "work_item_templates", "page_templates"},
	"type-templates":      {"issue_type_templates", "type_templates"},
	"estimate-points":     {"estimate_points"},
	"estimate-categories": {"estimate_categories"},
	"reports":             {"saved_reports"},
	"tasks":               {"agent_tasks"},
	"configs":             {"agent_configs", "mcp_configs"},
	"mcp":                 {"mcp_configs"},
	"sessions":            {"agent_sessions"},
	"memories":            {"memories", "agent_memories"},
	"jobs":                {"cicd_jobs", "jobs"},
	"builds":              {"cicd_builds", "builds"},
	"tabs":                {"project_page_tabs"},
	"page-tabs":           {"project_page_tabs"},
	"versions":            {"page_versions"},
	"rules":               {"automation_rules"},
	"automations":         {"automation_rules"},
	"squads":              {"squads"},
	"options":             {"custom_field_options"},
	"widgets":             {"dashboard_widgets"},
}

var snakeRe = regexp.MustCompile(`([a-z0-9])([A-Z])`)

func (a *auditor) resolve(name, prev string, resolved map[string]uint64) (uint64, bool) {
	switch name {
	case "wsParam":
		return a.ws, true
	case "projectId":
		resolved["projects"] = a.ids["project"]
		return a.ids["project"], true
	case "issueId":
		resolved["issues"] = a.ids["issue"]
		return a.ids["issue"], true
	case "userId":
		var uid uint64
		a.db.Raw(`SELECT user_id FROM workspace_members WHERE workspace_id = ? AND user_id <> ? ORDER BY user_id LIMIT 1`, a.ws, a.insider).Scan(&uid)
		return uid, uid != 0
	}
	var cands []string
	cands = append(cands, overrides[prev]...)
	if prev != "" && !strings.HasPrefix(prev, ":") {
		cands = append(cands, strings.ReplaceAll(prev, "-", "_"))
	}
	if name != "id" {
		base := strings.ToLower(snakeRe.ReplaceAllString(strings.TrimSuffix(name, "Id"), "${1}_${2}"))
		cands = append(cands, overrides[base+"s"]...)
		cands = append(cands, base+"s", base, base+"es")
	}
	for _, table := range cands {
		if id, ok := a.victimRow(table, resolved); ok {
			resolved[table] = id
			return id, true
		}
	}
	return 0, false
}

// victimRow finds a row of table owned by the victim workspace, preferring
// rows under the parents already resolved for this path.
func (a *auditor) victimRow(table string, resolved map[string]uint64) (uint64, bool) {
	m := a.db.Migrator()
	if !m.HasTable(table) || !m.HasColumn(table, "id") {
		return 0, false
	}
	type link struct{ col, parentTable, owner string }
	links := []link{
		{"workspace_id", "", "?"},
		{"project_id", "projects", "SELECT id FROM projects WHERE workspace_id = ?"},
		{"issue_id", "issues", "SELECT id FROM issues WHERE workspace_id = ?"},
		{"page_id", "pages", "SELECT id FROM pages WHERE workspace_id = ?"},
		{"agent_id", "agents", "SELECT id FROM agents WHERE workspace_id = ?"},
		{"workflow_id", "agent_workflows", "SELECT id FROM agent_workflows WHERE workspace_id = ? UNION SELECT id FROM workflows WHERE workspace_id = ?"},
		{"workflow_run_id", "workflow_runs", "SELECT r.id FROM workflow_runs r JOIN agent_workflows w ON w.id = r.workflow_id WHERE w.workspace_id = ?"},
		{"chat_id", "chats", "SELECT id FROM chats WHERE workspace_id = ?"},
		{"dashboard_id", "saved_dashboards", "SELECT id FROM saved_dashboards WHERE project_id IN (SELECT id FROM projects WHERE workspace_id = ?)"},
		{"custom_field_id", "custom_fields", "SELECT id FROM custom_fields WHERE workspace_id = ?"},
	}
	soft := ""
	if m.HasColumn(table, "deleted_at") {
		soft = " AND deleted_at IS NULL"
	}
	for _, l := range links {
		if !m.HasColumn(table, l.col) {
			continue
		}
		var id uint64
		if l.owner == "?" {
			a.db.Raw(fmt.Sprintf(`SELECT id FROM %s WHERE workspace_id = ?%s ORDER BY id LIMIT 1`, table, soft), a.ws).Scan(&id)
		} else {
			if pid, ok := resolved[l.parentTable]; ok {
				a.db.Raw(fmt.Sprintf(`SELECT id FROM %s WHERE %s = ?%s ORDER BY id LIMIT 1`, table, l.col, soft), pid).Scan(&id)
			}
			if id == 0 {
				args := []interface{}{}
				for range strings.Split(l.owner, "?")[1:] {
					args = append(args, a.ws)
				}
				a.db.Raw(fmt.Sprintf(`SELECT id FROM %s WHERE %s IN (%s)%s ORDER BY id LIMIT 1`, table, l.col, l.owner, soft), args...).Scan(&id)
			}
		}
		if id != 0 {
			return id, true
		}
	}
	return 0, false
}

func skip(path string) bool {
	for _, p := range []string{"/api/v1/auth", "/api/v1/sse", "/api/v1/intake/", "/api/v1/intake-channels", "/health", "/api/v1/health"} {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return strings.Contains(path, ":token") || strings.Contains(path, "webhook") && !strings.Contains(path, "webhooks/:")
}

func call(c *http.Client, method, url, token string) (int, string) {
	var body io.Reader
	if method != http.MethodGet && method != http.MethodDelete {
		body = bytes.NewBufferString(`{}`)
	}
	req, _ := http.NewRequest(method, url, body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return resp.StatusCode, string(b)
}

func (a *auditor) classify(method string, outsider, insider int, body string) string {
	if outsider < 200 || outsider >= 300 {
		switch {
		case outsider == 401 || outsider == 403 || outsider == 404:
			return "denied"
		case outsider == 400 || outsider == 422 || outsider == 409:
			return "invalid" // rejected before or after the access check; inconclusive
		default:
			return fmt.Sprintf("status%d", outsider)
		}
	}
	if method != http.MethodGet {
		return "WRITE"
	}
	compact := strings.Join(strings.Fields(body), "")
	for _, m := range a.markers {
		if strings.Contains(body, m) {
			return "LEAK"
		}
	}
	for _, empty := range []string{"[]", "{}", "null", `{"data":[]}`, `{"data":null}`, `{"items":[]`, `"total":0`} {
		if compact == empty || strings.HasPrefix(compact, empty) || strings.Contains(compact, empty) && len(compact) < 120 {
			return "ok-empty"
		}
	}
	if len(compact) > 2 {
		return "READ"
	}
	return "ok-empty"
}

func writeReport(path string, a *auditor, results []result) {
	var b strings.Builder
	counts := map[string]int{}
	for _, r := range results {
		k := r.Verdict
		if strings.HasPrefix(k, "unresolved") {
			k = "unresolved"
		}
		counts[k]++
	}
	fmt.Fprintf(&b, "# Route audit (workspace %d)\n\n", a.ws)
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, "- %s: %d\n", k, counts[k])
	}
	for _, section := range []string{"WRITE", "LEAK", "READ", "status500", "unresolved"} {
		fmt.Fprintf(&b, "\n## %s\n\n", section)
		for _, r := range results {
			if r.Verdict == section || section == "unresolved" && strings.HasPrefix(r.Verdict, "unresolved") ||
				section == "status500" && r.Verdict == "status500" {
				snippet := strings.ReplaceAll(r.Body, "\n", " ")
				if len(snippet) > 200 {
					snippet = snippet[:200]
				}
				if section == "unresolved" {
					fmt.Fprintf(&b, "- `%s %s` (%s)\n", r.Method, r.Path, strings.TrimPrefix(r.Verdict, "unresolved:"))
				} else {
					fmt.Fprintf(&b, "- `%s %s` outsider=%d insider=%d `%s`\n", r.Method, r.Path, r.Outsider, r.Insider, snippet)
				}
			}
		}
	}
	raw, _ := json.MarshalIndent(results, "", " ")
	_ = os.WriteFile(strings.TrimSuffix(path, ".md")+".json", raw, 0o644)
	_ = os.WriteFile(path, []byte(b.String()), 0o644)
	fmt.Printf("\nreport: %s\n", path)
	for _, k := range keys {
		fmt.Printf("  %s: %d\n", k, counts[k])
	}
}
