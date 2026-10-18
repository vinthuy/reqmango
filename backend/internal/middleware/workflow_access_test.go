package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestWorkflowAndPageAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Workspace{}, &model.WorkspaceMember{}, &model.Project{}, &model.ProjectMember{},
		&model.AgentWorkflow{}, &model.WorkflowNode{}, &model.WorkflowEdge{}, &model.WorkflowRun{},
		&model.Workflow{}, &model.StateTransition{}, &model.Page{}))
	mk := func(v interface{}) { require.NoError(t, db.Create(v).Error) }

	mk(&model.Workspace{BaseModel: model.BaseModel{ID: 10}, Name: "acme", Slug: "acme"})
	const alice, bob, carol = 1, 2, 3 // member, admin, outsider
	mk(&model.WorkspaceMember{WorkspaceID: 10, UserID: alice, Role: 15, IsActive: true})
	mk(&model.WorkspaceMember{WorkspaceID: 10, UserID: bob, Role: 20, IsActive: true})
	own := model.Project{Name: "own", Identifier: "OWN", WorkspaceID: 10}
	foreign := model.Project{Name: "foreign", Identifier: "FOR", WorkspaceID: 20}
	mk(&own)
	mk(&foreign)

	awOwn := model.AgentWorkflow{Name: "a", ProjectID: own.ID, WorkspaceID: 10}
	awForeign := model.AgentWorkflow{Name: "b", ProjectID: foreign.ID, WorkspaceID: 20}
	mk(&awOwn)
	mk(&awForeign)
	nodeForeign := model.WorkflowNode{WorkflowID: awForeign.ID, AgentID: 1, Name: "n"}
	mk(&nodeForeign)
	runForeign := model.WorkflowRun{WorkflowID: awForeign.ID}
	mk(&runForeign)

	ownPID := own.ID
	// State workflows live in a separate table whose ids overlap agent
	// workflows, so pin them to ids no agent workflow uses.
	swOwn := model.Workflow{BaseModel: model.BaseModel{ID: 200}, Name: "s1", ProjectID: &ownPID, WorkspaceID: 10}
	swWs := model.Workflow{BaseModel: model.BaseModel{ID: 201}, Name: "s2", WorkspaceID: 10}
	swForeign := model.Workflow{BaseModel: model.BaseModel{ID: 202}, Name: "s3", WorkspaceID: 20}
	mk(&swOwn)
	mk(&swWs)
	mk(&swForeign)
	awNew := model.AgentWorkflow{BaseModel: model.BaseModel{ID: 100}, Name: "new", ProjectID: own.ID, WorkspaceID: 10}
	awNewForeign := model.AgentWorkflow{BaseModel: model.BaseModel{ID: 101}, Name: "new foreign", ProjectID: foreign.ID, WorkspaceID: 20}
	awDeleted := model.AgentWorkflow{BaseModel: model.BaseModel{ID: 102}, Name: "deleted", ProjectID: own.ID, WorkspaceID: 10}
	mk(&awNew)
	mk(&awNewForeign)
	mk(&awDeleted)
	require.NoError(t, db.Delete(&awDeleted).Error)
	tNew := model.StateTransition{Name: "new transition", WorkflowID: awNew.ID, WorkflowKind: model.TransitionKindAgent, SourceStateID: 1, TargetStateID: 2}
	tNewForeign := model.StateTransition{Name: "foreign transition", WorkflowID: awNewForeign.ID, WorkflowKind: model.TransitionKindAgent, SourceStateID: 1, TargetStateID: 2}
	mk(&tNew)
	mk(&tNewForeign)
	// A state-machine transition whose workflow_id happens to equal the agent
	// workflow's id: the kind must keep it out of the agent workflow's routes.
	tColliding := model.StateTransition{Name: "state transition with colliding id", WorkflowID: awNew.ID, WorkflowKind: model.TransitionKindState, SourceStateID: 1, TargetStateID: 2}
	mk(&tColliding)
	tOwn := model.StateTransition{Name: "t", WorkflowID: swOwn.ID, WorkflowKind: model.TransitionKindState, SourceStateID: 1, TargetStateID: 2}
	tForeign := model.StateTransition{Name: "t", WorkflowID: swForeign.ID, WorkflowKind: model.TransitionKindState, SourceStateID: 1, TargetStateID: 2}
	tWs := model.StateTransition{Name: "t ws", WorkflowID: swWs.ID, WorkflowKind: model.TransitionKindState, SourceStateID: 1, TargetStateID: 2}
	mk(&tOwn)
	mk(&tForeign)
	mk(&tWs)

	ownPage := model.Page{Title: "p", ProjectID: own.ID, WorkspaceID: 10}
	foreignPage := model.Page{Title: "p", ProjectID: foreign.ID, WorkspaceID: 20}
	mk(&ownPage)
	mk(&foreignPage)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		ids := map[string]uint64{"alice": alice, "bob": bob, "carol": carol}
		c.Set("currentUser", &model.User{BaseModel: model.BaseModel{ID: ids[c.GetHeader("X-User")]}})
	})
	ok := func(c *gin.Context) { c.Status(http.StatusOK) }
	pw := r.Group("/projects/:projectId/workflows", RequireProjectWorkflow(db))
	pw.GET("/:workflowId", ok)
	pw.PUT("/:workflowId/nodes/:nodeId", ok)
	pw.GET("/:workflowId/runs/:runId", ok)
	pw.GET("/:workflowId/transitions", ok)
	pw.POST("/:workflowId/transitions", ok)
	pw.PUT("/:workflowId/transitions/:transitionId", ok)
	ww := r.Group("/workspaces/:wsParam/workflows", RequireWorkspaceWorkflow(db))
	ww.GET("", ok)
	ww.GET("/:workflowId", ok)
	ww.GET("/:workflowId/transitions", ok)
	ww.POST("/:workflowId/transitions", ok)
	ww.DELETE("/:workflowId/transitions/:transitionId", ok)
	r.POST("/pages/:pageId/ai", RequirePageAccess(db), ok)

	status := func(user, method, path string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("X-User", user)
		r.ServeHTTP(w, req)
		return w.Code
	}
	p := "/projects/" + uitoa(own.ID) + "/workflows/"

	assert.Equal(t, 200, status("alice", "GET", p+uitoa(awOwn.ID)))
	assert.Equal(t, 404, status("alice", "GET", p+uitoa(awForeign.ID)), "a foreign workflow is not reachable through our project")
	assert.Equal(t, 403, status("carol", "GET", p+uitoa(awOwn.ID)))
	assert.Equal(t, 404, status("alice", "PUT", p+uitoa(awOwn.ID)+"/nodes/"+uitoa(nodeForeign.ID)))
	assert.Equal(t, 404, status("alice", "GET", p+uitoa(awOwn.ID)+"/runs/"+uitoa(runForeign.ID)))

	assert.Equal(t, 200, status("alice", "GET", p+uitoa(awNew.ID)+"/transitions"))
	assert.Equal(t, 200, status("bob", "PUT", p+uitoa(awNew.ID)+"/transitions/"+uitoa(tNew.ID)))
	assert.Equal(t, 404, status("bob", "PUT", p+uitoa(awNew.ID)+"/transitions/"+uitoa(tNewForeign.ID)))
	assert.Equal(t, 404, status("bob", "PUT", p+uitoa(awNew.ID)+"/transitions/"+uitoa(tColliding.ID)), "a state transition is not addressable through an agent workflow")
	assert.Equal(t, 404, status("bob", "POST", p+uitoa(awNewForeign.ID)+"/transitions"))
	assert.Equal(t, 404, status("bob", "POST", p+uitoa(awDeleted.ID)+"/transitions"))
	assert.Equal(t, 403, status("carol", "POST", p+uitoa(awNew.ID)+"/transitions"))
	// The two id spaces overlap; a legacy workflow with the same id must not
	// shadow the agent workflow these routes actually address.
	mk(&model.Workflow{BaseModel: model.BaseModel{ID: awNew.ID}, Name: "colliding foreign state workflow", ProjectID: &foreign.ID, WorkspaceID: 20})
	assert.Equal(t, 200, status("bob", "POST", p+uitoa(awNew.ID)+"/transitions"), "the agent workflow wins, not the colliding legacy row")

	assert.Equal(t, 200, status("alice", "GET", "/workspaces/10/workflows"))
	assert.Equal(t, 200, status("alice", "GET", "/workspaces/acme/workflows"), "slugs resolve")
	assert.Equal(t, 404, status("alice", "GET", "/workspaces/nope/workflows"))
	assert.Equal(t, 403, status("carol", "GET", "/workspaces/10/workflows"))
	assert.Equal(t, 200, status("alice", "GET", "/workspaces/10/workflows/"+uitoa(swWs.ID)), "workspace routes address state workflows")
	assert.Equal(t, 404, status("alice", "GET", "/workspaces/10/workflows/"+uitoa(awNew.ID)), "an agent workflow id is not a state workflow")
	assert.Equal(t, 404, status("alice", "GET", "/workspaces/10/workflows/"+uitoa(swForeign.ID)), "a foreign workspace workflow is unreachable")
	assert.Equal(t, 200, status("alice", "GET", "/workspaces/10/workflows/"+uitoa(swWs.ID)+"/transitions"))
	assert.Equal(t, 404, status("alice", "GET", "/workspaces/10/workflows/"+uitoa(swForeign.ID)+"/transitions"))
	assert.Equal(t, 403, status("alice", "POST", "/workspaces/10/workflows/"+uitoa(swWs.ID)+"/transitions"), "members cannot change workflows")
	assert.Equal(t, 200, status("bob", "POST", "/workspaces/10/workflows/"+uitoa(swWs.ID)+"/transitions"))
	assert.Equal(t, 404, status("bob", "POST", "/workspaces/10/workflows/"+uitoa(awNewForeign.ID)+"/transitions"))
	assert.Equal(t, 404, status("bob", "POST", "/workspaces/10/workflows/"+uitoa(awNew.ID)+"/transitions"), "an agent workflow id is not a state workflow")
	assert.Equal(t, 200, status("bob", "DELETE", "/workspaces/10/workflows/"+uitoa(swWs.ID)+"/transitions/"+uitoa(tWs.ID)))
	assert.Equal(t, 404, status("bob", "DELETE", "/workspaces/10/workflows/"+uitoa(swWs.ID)+"/transitions/"+uitoa(tForeign.ID)), "a transition of another workflow is unreachable")
	assert.Equal(t, 404, status("alice", "GET", p+uitoa(swOwn.ID)), "a state workflow id is not an agent workflow")

	assert.Equal(t, 200, status("alice", "POST", "/pages/"+uitoa(ownPage.ID)+"/ai"))
	assert.Equal(t, 403, status("alice", "POST", "/pages/"+uitoa(foreignPage.ID)+"/ai"))
	assert.Equal(t, 404, status("alice", "POST", "/pages/999/ai"))
}
