// Package model — Partial read/write mappings of tables owned by the main backend.
// These structs MUST NEVER be added to AutoMigrate. They exist only so that
// agent-service can read and write backend-owned tables via the shared Postgres.
package model

import (
	"time"

	"gorm.io/gorm"
)

// backendBaseModel provides common fields matching backend's model.BaseModel.
type backendBaseModel struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	CreatedByID *uint64        `gorm:"column:created_by_id" json:"created_by_id"`
	UpdatedByID *uint64        `gorm:"column:updated_by_id" json:"updated_by_id"`
}

// Issue is a partial mapping of the backend "issues" table.
type Issue struct {
	backendBaseModel

	Name            string  `gorm:"size:255;not null" json:"name"`
	DescriptionHTML string  `gorm:"type:text;default:<p></p>" json:"description_html"`
	Priority        string  `gorm:"size:30;default:none" json:"priority"`
	SequenceID      int     `gorm:"default:1" json:"sequence_id"`
	ProjectID       uint64  `gorm:"not null;index" json:"project_id"`
	WorkspaceID     uint64  `gorm:"not null" json:"workspace_id"`
	IssueTypeID     *uint64 `json:"issue_type_id"`
	StateID         uint64  `gorm:"not null;index" json:"state_id"`

	// Relationships needed for Preload
	State     State     `gorm:"foreignKey:StateID" json:"-"`
	IssueType IssueType `gorm:"foreignKey:IssueTypeID" json:"-"`
}

func (Issue) TableName() string { return "issues" }

// State is a partial mapping of the backend "states" table.
type State struct {
	backendBaseModel

	Name        string  `gorm:"size:255;not null" json:"name"`
	Color       string  `gorm:"size:50;default:#6B7280" json:"color"`
	Group       string  `gorm:"size:50;default:backlog" json:"group"`
	Sequence    int     `gorm:"default:1" json:"sequence"`
	IsActive    bool    `gorm:"default:true" json:"is_active"`
	IsDefault   bool    `gorm:"default:false" json:"is_default"`
	ProjectID   *uint64 `gorm:"index" json:"project_id"`
	WorkspaceID uint64  `gorm:"not null" json:"workspace_id"`
}

func (State) TableName() string { return "states" }

// IssueType is a partial mapping of the backend "issue_types" table.
type IssueType struct {
	backendBaseModel

	Name        string  `gorm:"type:varchar(100);not null" json:"name"`
	Color       string  `gorm:"type:varchar(20);default:'#6366F1'" json:"color"`
	Level       int     `gorm:"default:0" json:"level"`
	IsActive    bool    `gorm:"default:true" json:"is_active"`
	IsDefault   bool    `gorm:"default:false" json:"is_default"`
	Sequence    int     `gorm:"default:1" json:"sequence"`
	ProjectID   *uint64 `gorm:"index" json:"project_id"`
	WorkspaceID uint64  `gorm:"not null;index" json:"workspace_id"`
}

func (IssueType) TableName() string { return "issue_types" }

// Label is a partial mapping of the backend "labels" table.
type Label struct {
	backendBaseModel

	Name        string  `gorm:"size:255;not null" json:"name"`
	Color       string  `gorm:"size:50;default:#6B7280" json:"color"`
	Description *string `gorm:"size:255" json:"description"`
	ProjectID   *uint64 `gorm:"index" json:"project_id"`
	WorkspaceID uint64  `gorm:"not null" json:"workspace_id"`
}

func (Label) TableName() string { return "labels" }

// Cycle is a partial mapping of the backend "cycles" table.
type Cycle struct {
	backendBaseModel

	Name        string     `gorm:"size:255;not null" json:"name"`
	StartDate   time.Time  `gorm:"type:date;not null" json:"start_date"`
	EndDate     *time.Time `gorm:"type:date" json:"end_date"`
	CompletedAt *time.Time `json:"completed_at"`
	CancelledAt *time.Time `json:"cancelled_at"`
	ProjectID   uint64     `gorm:"not null;index" json:"project_id"`
	WorkspaceID uint64     `gorm:"not null" json:"workspace_id"`
}

func (Cycle) TableName() string { return "cycles" }

// IssueCycle is a partial mapping of the backend "issue_cycles" join table.
type IssueCycle struct {
	IssueID uint64 `gorm:"primaryKey;autoIncrement:false" json:"issue_id"`
	CycleID uint64 `gorm:"primaryKey;autoIncrement:false" json:"cycle_id"`
}

func (IssueCycle) TableName() string { return "issue_cycles" }

// Module is a partial mapping of the backend "modules" table.
type Module struct {
	backendBaseModel

	Name        string  `gorm:"type:varchar(100);not null" json:"name"`
	Description string  `gorm:"type:text" json:"description"`
	ProjectID   *uint64 `gorm:"index" json:"project_id"`
	WorkspaceID uint64  `gorm:"not null;index" json:"workspace_id"`
}

func (Module) TableName() string { return "modules" }

// ProjectMember is a partial mapping of the backend "project_members" table.
type ProjectMember struct {
	backendBaseModel

	ProjectID uint64 `gorm:"not null;uniqueIndex:idx_proj_member_user" json:"project_id"`
	UserID    uint64 `gorm:"not null;uniqueIndex:idx_proj_member_user" json:"user_id"`
	Role      int    `gorm:"default:15" json:"role"`
	IsActive  bool   `gorm:"default:true" json:"is_active"`

	// Relationships needed for Preload
	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (ProjectMember) TableName() string { return "project_members" }

// Project is a partial mapping of the backend "projects" table.
type Project struct {
	backendBaseModel

	Name        string  `gorm:"size:255;not null" json:"name"`
	Identifier  string  `gorm:"size:10;not null" json:"identifier"`
	Description *string `gorm:"size:1000" json:"description"`
	WorkspaceID uint64  `gorm:"not null;index" json:"workspace_id"`
}

func (Project) TableName() string { return "projects" }

// Page is a partial mapping of the backend "pages" table.
type Page struct {
	backendBaseModel

	Title       string     `gorm:"size:255;not null" json:"title"`
	Content     string     `gorm:"type:text;not null;default:''" json:"content"`
	ArchivedAt  *time.Time `json:"archived_at"`
	Depth       int        `gorm:"default:0" json:"depth"`
	ProjectID   uint64     `gorm:"not null;index" json:"project_id"`
	WorkspaceID uint64     `gorm:"not null" json:"workspace_id"`
}

func (Page) TableName() string { return "pages" }

// Release is a partial mapping of the backend "releases" table.
type Release struct {
	backendBaseModel

	Name      string `gorm:"size:100;not null" json:"name"`
	Version   string `gorm:"size:50;not null" json:"version"`
	Status    string `gorm:"size:30;default:planned" json:"status"`
	ProjectID uint64 `gorm:"not null;index" json:"project_id"`
}

func (Release) TableName() string { return "releases" }

// Comment is a partial mapping of the backend "comments" table.
type Comment struct {
	backendBaseModel

	IssueID uint64  `gorm:"not null;index" json:"issue_id"`
	Body    string  `gorm:"type:text;not null" json:"body"`
	AuthorID *uint64 `json:"author_id"`
}

func (Comment) TableName() string { return "comments" }

// IssueActivity is a partial mapping of the backend "issue_activities" table.
type IssueActivity struct {
	backendBaseModel

	IssueID  *uint64 `json:"issue_id"`
	Verb     string  `gorm:"size:255;default:created" json:"verb"`
	Field    *string `gorm:"size:255" json:"field"`
	OldValue *string `gorm:"type:text" json:"old_value"`
	NewValue *string `gorm:"type:text" json:"new_value"`
	Comment  *string `gorm:"type:text" json:"comment"`
	ActorID  *uint64 `json:"actor_id"`
}

func (IssueActivity) TableName() string { return "issue_activities" }

// User is a partial mapping of the backend "users" table.
type User struct {
	backendBaseModel

	Email       string `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Username    string `gorm:"uniqueIndex;size:128;not null" json:"username"`
	DisplayName string `gorm:"size:255;default:''" json:"display_name"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
}

func (User) TableName() string { return "users" }
