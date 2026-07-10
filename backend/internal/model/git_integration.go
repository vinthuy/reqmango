package model

import "time"

type GitIntegration struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID     uint64     `gorm:"index;not null" json:"project_id"`
	Provider      string     `gorm:"type:varchar(20);not null" json:"provider"`
	RepoURL       string     `gorm:"type:text" json:"repo_url"`
	RepoName      string     `gorm:"type:varchar(255)" json:"repo_name"`
	AccessToken   string     `gorm:"type:text" json:"access_token"`
	WebhookSecret string     `gorm:"type:varchar(255)" json:"webhook_secret"`
	Active        bool       `gorm:"default:true" json:"active"`
	SyncPRs       bool       `gorm:"default:true" json:"sync_prs"`
	SyncCommits   bool       `gorm:"default:true" json:"sync_commits"`
	SyncBranches  bool       `gorm:"default:false" json:"sync_branches"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (GitIntegration) TableName() string {
	return "git_integrations"
}

type GitIssueLink struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	IssueID      uint64     `gorm:"index;not null" json:"issue_id"`
	GitType      string     `gorm:"type:varchar(20);not null" json:"git_type"`
	GitID        string     `gorm:"type:varchar(255);not null" json:"git_id"`
	GitURL       string     `gorm:"type:text" json:"git_url"`
	GitTitle     string     `gorm:"type:text" json:"git_title"`
	GitState     string     `gorm:"type:varchar(20)" json:"git_state"`
	GitAuthor    string     `gorm:"type:varchar(255)" json:"git_author"`
	GitBranch    string     `gorm:"type:varchar(255)" json:"git_branch"`
	IntegrationID uint64    `gorm:"index" json:"integration_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (GitIssueLink) TableName() string {
	return "git_issue_links"
}