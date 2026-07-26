package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
)

type SquadHandler struct{ svc *service.SquadService }

func NewSquadHandler(svc *service.SquadService) *SquadHandler {
	return &SquadHandler{svc: svc}
}

func (h *SquadHandler) respond(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if ae, ok := err.(*common.AppError); ok {
		c.JSON(ae.Code, gin.H{"message": ae.Message})
		return true
	}
	c.JSON(500, gin.H{"message": "Internal server error"})
	return true
}

func (h *SquadHandler) parseWorkspaceID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("wsParam"), 10, 64)
}

func (h *SquadHandler) CreateSquad(c *gin.Context) {
	wid, err := h.parseWorkspaceID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workspace ID"})
		return
	}
	var req request.SquadCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.Create(wid, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(201, resp)
}

func (h *SquadHandler) GetSquad(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("squadId"), 10, 64)
	resp, e := h.svc.Get(id)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *SquadHandler) ListSquads(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	projectID := c.Query("project_id")
	var pid *uint64
	if projectID != "" {
		id, _ := strconv.ParseUint(projectID, 10, 64)
		pid = &id
	}
	resp, e := h.svc.List(wid, pid)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *SquadHandler) UpdateSquad(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("squadId"), 10, 64)
	var req request.SquadUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.Update(id, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *SquadHandler) DeleteSquad(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("squadId"), 10, 64)
	if h.respond(c, h.svc.Delete(id)) {
		return
	}
	c.JSON(200, gin.H{"message": "Deleted"})
}

func (h *SquadHandler) AddMember(c *gin.Context) {
	squadID, _ := strconv.ParseUint(c.Param("squadId"), 10, 64)
	var req request.SquadMemberAdd
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.AddMember(squadID, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *SquadHandler) RemoveMember(c *gin.Context) {
	squadID, _ := strconv.ParseUint(c.Param("squadId"), 10, 64)
	memberID, _ := strconv.ParseUint(c.Param("memberId"), 10, 64)
	if h.respond(c, h.svc.RemoveMember(squadID, memberID)) {
		return
	}
	c.JSON(200, gin.H{"message": "Member removed"})
}

func (h *SquadHandler) StartExecution(c *gin.Context) {
	squadID, _ := strconv.ParseUint(c.Param("squadId"), 10, 64)
	var req request.SquadExecutionStart
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	
	// Get current user from context
	currentUser, exists := c.Get("currentUser")
	if exists {
		if user, ok := currentUser.(*model.User); ok {
			req.UserID = user.ID
		}
	}
	
	resp, e := h.svc.StartExecution(squadID, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *SquadHandler) GetExecution(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("executionId"), 10, 64)
	resp, e := h.svc.GetExecution(id)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *SquadHandler) ListExecutions(c *gin.Context) {
	squadID, _ := strconv.ParseUint(c.Param("squadId"), 10, 64)
	resp, e := h.svc.ListExecutions(squadID)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}
