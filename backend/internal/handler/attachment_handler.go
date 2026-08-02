package handler

import (
	"errors"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
)

const maxFileSize = 10 * 1024 * 1024

var allowedMIMETypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
	"application/pdf": true,
	"application/docx": true,
	"application/xlsx": true,
	"application/pptx": true,
	"text/plain": true,
	"text/csv":   true,
	"text/markdown": true,
	"application/json": true,
	"application/octet-stream": true,
}

func validateFile(file *multipart.FileHeader) error {
	if file.Size > maxFileSize {
		return errors.New("File size exceeds 10MB limit")
	}
	mimeType := file.Header.Get("Content-Type")
	if !allowedMIMETypes[mimeType] {
		return errors.New("File type not allowed")
	}
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		return errors.New("File must have an extension")
	}
	return nil
}

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

func (h *AttachmentHandler) Download(c *gin.Context) {
	attachmentID, err := strconv.ParseUint(c.Param("attachmentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attachment ID"})
		return
	}

	attachment, err := h.attachmentService.Get(attachmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.File(attachment.FilePath)
}

func (h *AttachmentHandler) Create(c *gin.Context) {
	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid issue ID"})
		return
	}

	// Limit request body to defend against oversized uploads before they are
	// buffered to disk/memory. c.FormFile parses the multipart form lazily, so
	// wrapping the body here aborts the upload mid-stream when it exceeds the
	// limit instead of rejecting it only after the full file has been read.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileSize+1024)

	user := middleware.GetCurrentUser(c)
	uploaderID := user.ID

	file, err := c.FormFile("file")
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "File size exceeds 10MB limit"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded: " + err.Error()})
		return
	}

	if err := validateFile(file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	user := middleware.GetCurrentUser(c)
	err = h.attachmentService.Delete(attachmentID, user.ID)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Attachment deleted"})
}