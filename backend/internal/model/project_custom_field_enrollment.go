package model

// ProjectCustomFieldEnrollment represents the enrollment state of a workspace-level custom field for a project.
// When a workspace-level field is created, it's not automatically available to projects.
// Projects must explicitly enroll to use the field.
type ProjectCustomFieldEnrollment struct {
	ProjectID  uint64 `gorm:"primaryKey;autoIncrement:false" json:"project_id"`
	FieldID    uint64 `gorm:"primaryKey;autoIncrement:false" json:"field_id"`
	IsEnabled  bool   `gorm:"default:true" json:"is_enabled"`
	Sequence   int    `gorm:"default:0" json:"sequence"`

	Project      Project      `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"-"`
	CustomField  CustomField  `gorm:"foreignKey:FieldID;constraint:OnDelete:CASCADE" json:"-"`
}

func (ProjectCustomFieldEnrollment) TableName() string {
	return "project_custom_field_enrollments"
}