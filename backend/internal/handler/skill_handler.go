package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/service"
)

type SkillHandler struct{ svc *service.SkillService }

func NewSkillHandler(svc *service.SkillService) *SkillHandler {
	return &SkillHandler{svc: svc}
}

func (h *SkillHandler) respond(c *gin.Context, err error) bool {
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

func (h *SkillHandler) parseWorkspaceID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("wsParam"), 10, 64)
}

func (h *SkillHandler) CreateSkill(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	var req request.SkillCreate
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

func (h *SkillHandler) GetSkill(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("skillId"), 10, 64)
	resp, e := h.svc.Get(id)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *SkillHandler) ListSkills(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	resp, e := h.svc.List(wid)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *SkillHandler) UpdateSkill(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("skillId"), 10, 64)
	var req request.SkillUpdate
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

func (h *SkillHandler) DeleteSkill(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("skillId"), 10, 64)
	if h.respond(c, h.svc.Delete(id)) {
		return
	}
	c.JSON(200, gin.H{"message": "Deleted"})
}

func (h *SkillHandler) ExecuteSkill(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("skillId"), 10, 64)
	var req request.SkillExecute
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.Execute(c.Request.Context(), id, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// ListSkillExecutionLogs retrieves skill execution logs with filtering and pagination.
func (h *SkillHandler) ListSkillExecutionLogs(c *gin.Context) {
	// Parse workspace ID
	wid, _ := h.parseWorkspaceID(c)

	// Parse skill ID (optional, from query or URL param)
	var skillID uint64
	if skillIdStr := c.Query("skill_id"); skillIdStr != "" {
		skillID, _ = strconv.ParseUint(skillIdStr, 10, 64)
	} else if skillIdStr := c.Param("skillId"); skillIdStr != "" {
		skillID, _ = strconv.ParseUint(skillIdStr, 10, 64)
	}

	// Parse status filter
	status := c.Query("status")

	// Parse date range
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		}
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	// Parse sorting
	sortBy := c.Query("sort_by")
	sortOrder := c.Query("sort_order")

	// Build query params
	params := service.SkillExecutionLogQueryParams{
		SkillID:     skillID,
		WorkspaceID: wid,
		Status:      status,
		StartDate:   startDate,
		EndDate:     endDate,
		Page:        page,
		PageSize:    pageSize,
		SortBy:      sortBy,
		SortOrder:   sortOrder,
	}

	// Execute query
	logs, total, err := h.svc.ListExecutionLogs(params)
	if h.respond(c, err) {
		return
	}

	// Calculate total pages
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	// Build paginated response
	c.JSON(200, response.PaginationResponse{
		Items:      logs,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	})
}

// ListPresetSkills retrieves preset skills for a workspace.
func (h *SkillHandler) ListPresetSkills(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	resp, e := h.svc.ListPresetSkills(wid)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// InitializePresetSkills initializes preset skills for a workspace.
func (h *SkillHandler) InitializePresetSkills(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	if e := h.svc.InitializePresetSkills(wid); h.respond(c, e) {
		return
	}
	c.JSON(200, gin.H{"message": "Preset skills initialized"})
}
