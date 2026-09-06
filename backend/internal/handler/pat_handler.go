package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
)

// PATHandler serves /auth/tokens endpoints.
type PATHandler struct {
	svc *service.PATService
}

func NewPATHandler(svc *service.PATService) *PATHandler { return &PATHandler{svc: svc} }

func toPATResponse(p model.PersonalAccessToken) response.PATResponse {
	return response.PATResponse{
		ID:          p.ID,
		Name:        p.Name,
		TokenPrefix: p.TokenPrefix,
		Scopes:      p.Scopes,
		LastUsedAt:  p.LastUsedAt,
		ExpiresAt:   p.ExpiresAt,
		RevokedAt:   p.RevokedAt,
		CreatedAt:   p.CreatedAt,
	}
}

// List handles GET /auth/tokens.
func (h *PATHandler) List(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	pats, err := h.svc.List(user.ID)
	if err != nil {
		writeAppError(c, err)
		return
	}
	out := make([]response.PATResponse, 0, len(pats))
	for _, p := range pats {
		out = append(out, toPATResponse(p))
	}
	c.JSON(http.StatusOK, out)
}

// Create handles POST /auth/tokens.
func (h *PATHandler) Create(c *gin.Context) {
	var req struct {
		Name      string     `json:"name" binding:"required,min=1,max=100"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	user := middleware.GetCurrentUser(c)
	token, pat, err := h.svc.Create(user.ID, req.Name, req.ExpiresAt)
	if err != nil {
		writeAppError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.PATCreateResponse{Token: token, PATResponse: toPATResponse(*pat)})
}

// Revoke handles DELETE /auth/tokens/:id.
func (h *PATHandler) Revoke(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid token id"})
		return
	}
	user := middleware.GetCurrentUser(c)
	if err := h.svc.Revoke(user.ID, id); err != nil {
		writeAppError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "revoked"})
}

// writeAppError maps common.AppError to the project's {"message": ...} format.
func writeAppError(c *gin.Context, err error) {
	if appErr, ok := err.(*common.AppError); ok {
		c.JSON(appErr.Code, gin.H{"message": appErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
}
