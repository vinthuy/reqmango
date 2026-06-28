package service

import (
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/reqmanpy/backend/internal/common"
	"github.com/reqmanpy/backend/internal/model"
	"gorm.io/gorm"
)

type AttachmentService struct {
	db *gorm.DB
}

func NewAttachmentService(db *gorm.DB) *AttachmentService {
	return &AttachmentService{db: db}
}

func (s *AttachmentService) ListByIssue(issueID uint64) ([]model.Attachment, error) {
	var attachments []model.Attachment
	err := s.db.Where("issue_id = ?", issueID).Order("created_at DESC").Find(&attachments).Error
	return attachments, err
}

func (s *AttachmentService) Get(attachmentID uint64) (*model.Attachment, error) {
	var attachment model.Attachment
	err := s.db.First(&attachment, attachmentID).Error
	if err != nil {
		return nil, common.NotFound("Attachment not found")
	}
	return &attachment, nil
}

func (s *AttachmentService) Create(issueID, uploaderID uint64, file io.Reader, fileName, mimeType string, fileSize int64) (*model.Attachment, error) {
	fileID := uuid.New().String()
	ext := filepath.Ext(fileName)
	newFileName := fileID + ext
	filePath := filepath.Join("uploads", newFileName)

	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		return nil, common.Internal("Failed to create uploads directory")
	}

	destFile, err := os.Create(filePath)
	if err != nil {
		return nil, common.Internal("Failed to create file")
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, file); err != nil {
		return nil, common.Internal("Failed to save file")
	}

	attachment := model.Attachment{
		Name:       fileName,
		FilePath:   filePath,
		FileSize:   fileSize,
		MimeType:   mimeType,
		IssueID:    issueID,
		UploaderID: &uploaderID,
	}

	err = s.db.Create(&attachment).Error
	if err != nil {
		return nil, common.Internal("Failed to save attachment")
	}

	return &attachment, nil
}

func (s *AttachmentService) Delete(attachmentID uint64) error {
	var attachment model.Attachment
	if err := s.db.First(&attachment, attachmentID).Error; err != nil {
		return common.NotFound("Attachment not found")
	}

	if err := os.Remove(attachment.FilePath); err != nil && !os.IsNotExist(err) {
		return common.Internal("Failed to delete file")
	}

	return s.db.Delete(&attachment).Error
}