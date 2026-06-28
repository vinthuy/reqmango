package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/service"
)

type WebhookHandler struct{ svc *service.WebhookService }

func NewWebhookHandler(svc *service.WebhookService) *WebhookHandler { return &WebhookHandler{svc: svc} }

func (h *WebhookHandler) List(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	r, err := h.svc.List(pid)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			common.RespondError(c, appErr)
			return
		}
		common.RespondError(c, common.Internal("Failed to list webhooks"))
		return
	}
	common.RespondOK(c, r)
}

func (h *WebhookHandler) Create(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("projectId"), 10, 64)
	wid, _ := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	var req struct{ Name, URL, Secret, Events string }
	if err := c.ShouldBindJSON(&req); err != nil { common.RespondError(c, common.BadRequest(err.Error())); return }
	r, err := h.svc.Create(pid, wid, &req)
	if err != nil { common.RespondError(c, err); return }
	common.RespondCreated(c, r)
}

func (h *WebhookHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct{ Name, URL, Secret, Events *string; IsActive *bool }
	if err := c.ShouldBindJSON(&req); err != nil { common.RespondError(c, common.BadRequest(err.Error())); return }
	r, err := h.svc.Update(id, &req)
	if err != nil { common.RespondError(c, err); return }
	common.RespondOK(c, r)
}

func (h *WebhookHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(id); err != nil { common.RespondError(c, err); return }
	common.RespondOK(c, gin.H{"message": "Webhook deleted"})
}
