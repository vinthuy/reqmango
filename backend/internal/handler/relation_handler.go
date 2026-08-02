package handler

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
)

type RelationHandler struct{ svc *service.RelationService }

func NewRelationHandler(svc *service.RelationService) *RelationHandler { return &RelationHandler{svc: svc} }
func (h *RelationHandler) respond(c *gin.Context, err error) bool {
	if err == nil { return false }
	if ae, ok := err.(*common.AppError); ok { c.JSON(ae.Code, gin.H{"message": ae.Message}); return true }
	c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"}); return true
}

// ---- Relation Types ----
func (h *RelationHandler) CreateType(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	wid, _ := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	var req request.RelationTypeCreate
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":"Invalid body"}); return }
	resp, e := h.svc.CreateType(wid, user.ID, req)
	if h.respond(c, e) { return }
	c.JSON(201, resp)
}
func (h *RelationHandler) ListTypes(c *gin.Context) {
	wid, _ := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	resp, e := h.svc.ListTypes(wid)
	if h.respond(c, e) { return }
	c.JSON(200, resp)
}
func (h *RelationHandler) UpdateType(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req request.RelationTypeUpdate
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":"Invalid body"}); return }
	resp, e := h.svc.UpdateType(id, user.ID, req)
	if h.respond(c, e) { return }
	c.JSON(200, resp)
}
func (h *RelationHandler) DeleteType(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if h.respond(c, h.svc.DeleteType(id, middleware.GetCurrentUser(c).ID)) { return }
	c.JSON(200, gin.H{"message":"Deleted"})
}

// ---- Issue Relations ----
func (h *RelationHandler) CreateRelation(c *gin.Context) {
	iid, _ := strconv.ParseUint(c.Param("issueId"), 10, 64)
	var req request.IssueRelationCreate
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":"Invalid body"}); return }
	resp, e := h.svc.CreateRelation(iid, middleware.GetCurrentUser(c).ID, req)
	if h.respond(c, e) { return }
	c.JSON(201, resp)
}
func (h *RelationHandler) ListRelations(c *gin.Context) {
	iid, _ := strconv.ParseUint(c.Param("issueId"), 10, 64)
	direction := c.DefaultQuery("direction", "both")
	var resp []response.IssueRelationResponse
	var e error
	if direction == "both" {
		resp, e = h.svc.ListRelationsBidirectional(iid)
	} else {
		resp, e = h.svc.ListRelations(iid)
	}
	if h.respond(c, e) { return }
	c.JSON(200, resp)
}
func (h *RelationHandler) DeleteRelation(c *gin.Context) {
	rid, _ := strconv.ParseUint(c.Param("relationId"), 10, 64)
	if h.respond(c, h.svc.DeleteRelation(rid, middleware.GetCurrentUser(c).ID)) { return }
	c.JSON(200, gin.H{"message": "Deleted"})
}
