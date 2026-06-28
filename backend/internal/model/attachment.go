package model

type Attachment struct {
	BaseModel
	Name      string   `gorm:"size:255;not null" json:"name"`
	FilePath  string   `gorm:"size:500;not null" json:"file_path"`
	FileSize  int64    `json:"file_size"`
	MimeType  string   `gorm:"size:100" json:"mime_type"`
	IssueID   uint64   `gorm:"not null;index" json:"issue_id"`
	UploaderID *uint64 `json:"uploader_id"`
}

func (Attachment) TableName() string {
	return "attachments"
}