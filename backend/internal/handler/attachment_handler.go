package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/reqmanpy/backend/internal/middleware"
	"github.com/reqmanpy/backend/internal/service"
	"strconv"
)

type AttachmentHandler struct {
	attachmentService *service.AttachmentService
}

func NewAttachmentHandler(attachmentService *service.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{attachmentService: attachmentService}
}

func (h *AttachmentHandler) ListByIssue(c *gin.Context) {
	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid issue ID"})
		return
	}

	attachments, err := h.attachmentService.ListByIssue(issueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, attachments)
}

func (h *AttachmentHandler) Get(c *gin.Context) {
	attachmentID, err := strconv.ParseUint(c.Param("attachmentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attachment ID"})
		return
	}

	attachment, err := h.attachmentService.Get(attachmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, attachment)
}

func (h *AttachmentHandler) Create(c *gin.Context) {
	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid issue ID"})
		return
	}

	user := middleware.GetCurrentUser(c)
	uploaderID := user.ID

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded: " + err.Error()})
		return
	}

	srcFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer srcFile.Close()

	attachment, err := h.attachmentService.Create(issueID, uploaderID, srcFile, file.Filename, file.Header.Get("Content-Type"), file.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, attachment)
}

func (h *AttachmentHandler) Delete(c *gin.Context) {
	attachmentID, err := strconv.ParseUint(c.Param("attachmentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attachment ID"})
		return
	}

	err = h.attachmentService.Delete(attachmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Attachment deleted"})
}