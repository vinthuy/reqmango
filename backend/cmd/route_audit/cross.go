package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/model"
)

var paramRe = regexp.MustCompile(`:(\w+)`)

// ownParams are path params that -cross mode fills with the attacker's own
// IDs; any other param carries a victim ID.
var ownParams = map[string]bool{"wsParam": true, "projectId": true, "issueId": true, "userId": true}

func hasChildParam(path string) bool {
	for _, m := range paramRe.FindAllStringSubmatch(path, -1) {
		if !ownParams[m[1]] {
			return true
		}
	}
	return false
}

// setupAttacker makes user the admin of a fresh workspace with one project
// and issue, so its requests pass every membership check for its own tenant.
func (a *auditor) setupAttacker(userID uint64) {
	stamp := time.Now().UnixNano() % 1e9
	ws := model.Workspace{Name: "route audit attacker", Slug: fmt.Sprintf("ra-attacker-%d", stamp), OwnerID: userID}
	must(a.db.Create(&ws).Error, "attacker workspace")
	must(a.db.Create(&model.WorkspaceMember{WorkspaceID: ws.ID, UserID: userID, Role: 20, IsActive: true}).Error, "attacker workspace member")
	project := model.Project{Name: "attacker", Identifier: fmt.Sprintf("RA%d", stamp%1e6), WorkspaceID: ws.ID}
	must(a.db.Create(&project).Error, "attacker project")
	must(a.db.Create(&model.ProjectMember{ProjectID: project.ID, UserID: userID, Role: 20, IsActive: true}).Error, "attacker project member")
	state := model.State{Name: "Todo", ProjectID: &project.ID, WorkspaceID: ws.ID, IsDefault: true}
	must(a.db.Create(&state).Error, "attacker state")
	issue := model.Issue{Name: "attacker issue", ProjectID: project.ID, WorkspaceID: ws.ID, StateID: state.ID}
	must(a.db.Create(&issue).Error, "attacker issue")
	a.attacker = &tenant{ws: ws.ID, project: project.ID, issue: issue.ID}
	fmt.Printf("attacker workspace=%d project=%d issue=%d user=%d\n", ws.ID, project.ID, issue.ID, userID)
}

func must(err error, what string) {
	if err != nil {
		log.Fatalf("create %s: %v", what, err)
	}
}

// classifyCross flags any success on a route that addressed a victim child
// resource through the attacker's own workspace/project/issue. An empty body
// means the handler found nothing in the attacker's tenant, so no data crossed.
func classifyCross(code int, body string) string {
	switch {
	case code >= 200 && code < 300:
		compact := strings.Join(strings.Fields(body), "")
		if compact == "" || compact == "[]" || compact == "{}" || compact == "null" ||
			strings.HasPrefix(compact, `{"data":[]`) || strings.HasPrefix(compact, `{"items":[]`) {
			return "ok-empty"
		}
		return "XTENANT"
	case code == 401 || code == 403 || code == 404:
		return "denied"
	case code == 400 || code == 422 || code == 409:
		return "invalid"
	}
	return fmt.Sprintf("status%d", code)
}