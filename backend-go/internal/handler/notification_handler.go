package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/model"
	"github.com/reqmanpy/backend-go/internal/service"
)

type NotificationHandler struct{ svc *service.NotificationService }

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) getCurrentUserID(c *gin.Context) uint64 {
	u, _ := c.Get("currentUser")
	if u, ok := u.(*model.User); ok { return u.ID }
	return 0
}

func (h *NotificationHandler) List(c *gin.Context) {
	uid := h.getCurrentUserID(c)
	unreadOnly := c.Query("unread_only") == "true"
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	r, err := h.svc.List(uid, unreadOnly, limit, offset)
	if err != nil { common.RespondError(c, err); return }
	common.RespondOK(c, r)
}

func (h *NotificationHandler) GetSummary(c *gin.Context) {
	r, err := h.svc.GetSummary(h.getCurrentUserID(c))
	if err != nil { common.RespondError(c, err); return }
	common.RespondOK(c, r)
}

func (h *NotificationHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil { common.RespondError(c, common.BadRequest("Invalid notification ID")); return }
	r, svcErr := h.svc.Get(id, h.getCurrentUserID(c))
	if svcErr != nil { common.RespondError(c, svcErr); return }
	common.RespondOK(c, r)
}

func (h *NotificationHandler) Create(c *gin.Context) {
	var req request.NotificationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil { common.RespondError(c, common.BadRequest(err.Error())); return }
	r, err := h.svc.Create(&req)
	if err != nil { common.RespondError(c, err); return }
	common.RespondCreated(c, r)
}

func (h *NotificationHandler) CreateBulk(c *gin.Context) {
	var req request.NotificationBulkCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil { common.RespondError(c, common.BadRequest(err.Error())); return }
	r, err := h.svc.CreateBulk(&req)
	if err != nil { common.RespondError(c, err); return }
	common.RespondCreated(c, r)
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil { common.RespondError(c, common.BadRequest("Invalid notification ID")); return }
	r, svcErr := h.svc.MarkRead(id, h.getCurrentUserID(c))
	if svcErr != nil { common.RespondError(c, svcErr); return }
	common.RespondOK(c, r)
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	r, err := h.svc.MarkAllRead(h.getCurrentUserID(c))
	if err != nil { common.RespondError(c, err); return }
	common.RespondOK(c, r)
}

func (h *NotificationHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil { common.RespondError(c, common.BadRequest("Invalid notification ID")); return }
	if err := h.svc.Delete(id, h.getCurrentUserID(c)); err != nil { common.RespondError(c, err); return }
	common.RespondOK(c, gin.H{"message": "Notification deleted"})
}
