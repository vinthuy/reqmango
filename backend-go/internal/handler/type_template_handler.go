package handler

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/middleware"
	"github.com/reqmanpy/backend-go/internal/service"
)

type TypeTemplateHandler struct{ svc *service.TypeTemplateService }

func NewTypeTemplateHandler(svc *service.TypeTemplateService) *TypeTemplateHandler {
	return &TypeTemplateHandler{svc: svc}
}

func (h *TypeTemplateHandler) parseID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("id"), 10, 64)
}

func (h *TypeTemplateHandler) respond(c *gin.Context, err error) bool {
	if err == nil { return false }
	if ae, ok := err.(*common.AppError); ok {
		c.JSON(ae.Code, gin.H{"message": ae.Message}); return true
	}
	c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"}); return true
}

func (h *TypeTemplateHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	wid, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil { c.JSON(400, gin.H{"message":"Invalid workspace_id"}); return }
	var req request.TypeTemplateCreate
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":"Invalid body"}); return }
	resp, e := h.svc.Create(wid, user.ID, req)
	if h.respond(c, e) { return }
	c.JSON(201, resp)
}

func (h *TypeTemplateHandler) List(c *gin.Context) {
	wid, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil { c.JSON(400, gin.H{"message":"Invalid workspace_id"}); return }
	resp, e := h.svc.List(wid)
	if h.respond(c, e) { return }
	c.JSON(200, resp)
}

func (h *TypeTemplateHandler) Get(c *gin.Context) {
	id, err := h.parseID(c)
	if err != nil { c.JSON(400, gin.H{"message":"Invalid ID"}); return }
	resp, e := h.svc.Get(id)
	if h.respond(c, e) { return }
	c.JSON(200, resp)
}

func (h *TypeTemplateHandler) Update(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	id, err := h.parseID(c)
	if err != nil { c.JSON(400, gin.H{"message":"Invalid ID"}); return }
	var req request.TypeTemplateUpdate
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":"Invalid body"}); return }
	resp, e := h.svc.Update(id, user.ID, req)
	if h.respond(c, e) { return }
	c.JSON(200, resp)
}

func (h *TypeTemplateHandler) Delete(c *gin.Context) {
	id, err := h.parseID(c)
	if err != nil { c.JSON(400, gin.H{"message":"Invalid ID"}); return }
	if h.respond(c, h.svc.Delete(id)) { return }
	c.JSON(200, gin.H{"message":"Deleted"})
}

func (h *TypeTemplateHandler) BindField(c *gin.Context) {
	id, _ := h.parseID(c)
	var req request.TypeTemplateFieldBind
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"message":"Invalid body"}); return }
	if req.Sequence == 0 { req.Sequence = 1 }
	resp, e := h.svc.BindField(id, req.FieldID, req.IsRequired, req.Sequence)
	if h.respond(c, e) { return }
	c.JSON(201, resp)
}

func (h *TypeTemplateHandler) UnbindField(c *gin.Context) {
	id, _ := h.parseID(c)
	fid, err := strconv.ParseUint(c.Param("fieldId"), 10, 64)
	if err != nil { c.JSON(400, gin.H{"message":"Invalid field ID"}); return }
	if h.respond(c, h.svc.UnbindField(id, fid)) { return }
	c.JSON(200, gin.H{"message":"Field unbound"})
}
