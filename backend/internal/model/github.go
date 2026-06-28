package model

type GitHubConnection struct {
	BaseModel
	WorkspaceID uint64 `gorm:"not null;index" json:"workspace_id"`
	ProjectID   uint64 `gorm:"not null;index" json:"project_id"`
	RepoOwner   string `gorm:"size:255;not null" json:"repo_owner"`
	RepoName    string `gorm:"size:255;not null" json:"repo_name"`
	AccessToken string `gorm:"size:500" json:"access_token"`
	WebhookSecret string `gorm:"size:500" json:"webhook_secret"`
	IsEnabled     bool   `gorm:"default:true" json:"is_enabled"`
	SyncIssues    bool   `gorm:"default:true" json:"sync_issues"`
	SyncPRs       bool   `gorm:"default:true" json:"sync_prs"`
	LastSyncAt    *string `gorm:"size:30" json:"last_sync_at"`
	WebhookID     *uint64 `json:"webhook_id"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Project   Project   `gorm:"foreignKey:ProjectID" json:"-"`
}

func (GitHubConnection) TableName() string {
	return "github_connections"
}
