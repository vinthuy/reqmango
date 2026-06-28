package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend/internal/middleware"
	"github.com/reqmanpy/backend/internal/service"
)

type CommentHandler struct{ svc *service.CommentService }

func NewCommentHandler(svc *service.CommentService) *CommentHandler { return &CommentHandler{svc: svc} }

func (h *CommentHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	var req struct {
		IssueID  uint64  `json:"issue_id" binding:"required"`
		Body     string  `json:"body" binding:"required"`
		ParentID *uint64 `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid body"})
		return
	}
	resp, err := h.svc.Create(req.IssueID, user.ID, req.Body, req.ParentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *CommentHandler) ListByIssue(c *gin.Context) {
	issueID, _ := strconv.ParseUint(c.Param("issueId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	comments, total, err := h.svc.ListByIssue(issueID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comments": comments, "total": total, "page": page, "page_size": pageSize})
}

func (h *CommentHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("commentId"), 10, 64)
	resp, err := h.svc.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *CommentHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("commentId"), 10, 64)
	var req struct {
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid body"})
		return
	}
	resp, err := h.svc.Update(id, req.Body)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *CommentHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("commentId"), 10, 64)
	if err := h.svc.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

func (h *CommentHandler) Resolve(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("commentId"), 10, 64)
	resp, err := h.svc.Resolve(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *CommentHandler) Unresolve(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("commentId"), 10, 64)
	resp, err := h.svc.Unresolve(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
