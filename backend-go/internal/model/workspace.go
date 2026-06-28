package model

// Workspace represents a team workspace.
type Workspace struct {
	BaseModel

	Name             string  `gorm:"size:255;not null" json:"name"`
	Slug             string  `gorm:"uniqueIndex;size:50;not null" json:"slug"`
	LogoURL          *string `gorm:"size:800" json:"logo_url"`
	OrganizationSize *string `gorm:"size:50" json:"organization_size"`
	Timezone         string  `gorm:"size:255;default:UTC" json:"timezone"`
	OwnerID          uint64  `gorm:"not null" json:"owner_id"`

	// Relationships
	Owner    User              `gorm:"foreignKey:OwnerID" json:"-"`
	Members  []WorkspaceMember `gorm:"foreignKey:WorkspaceID" json:"-"`
	Projects []Project         `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (Workspace) TableName() string {
	return "workspaces"
}

// WorkspaceMember represents a user's membership in a workspace.
type WorkspaceMember struct {
	BaseModel

	WorkspaceID uint64 `gorm:"not null;uniqueIndex:idx_ws_member_user" json:"workspace_id"`
	UserID      uint64 `gorm:"not null;uniqueIndex:idx_ws_member_user" json:"user_id"`
	Role        int    `gorm:"default:15" json:"role"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`

	// Relationships
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (WorkspaceMember) TableName() string {
	return "workspace_members"
}
