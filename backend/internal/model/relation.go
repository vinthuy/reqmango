package model

// RelationType defines a workspace-level custom relation between work items.
// e.g., "Blocks" with inward "blocked by" and outward "blocks".
type RelationType struct {
	BaseModel
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	InwardName  string `gorm:"type:varchar(100);not null" json:"inward_name"`  // how it reads from source
	OutwardName string `gorm:"type:varchar(100);not null" json:"outward_name"` // how it reads from target
	WorkspaceID uint64 `gorm:"not null;index" json:"workspace_id"`
}

func (RelationType) TableName() string { return "relation_types" }

// IssueRelation links two issues with a relation type.
type IssueRelation struct {
	BaseModel
	IssueID        uint64  `gorm:"not null;index" json:"issue_id"`
	RelatedIssueID uint64  `gorm:"not null;index" json:"related_issue_id"`
	RelationTypeID uint64  `gorm:"not null" json:"relation_type_id"`
	Comment        *string `gorm:"type:text" json:"comment"`

	Issue        Issue        `gorm:"foreignKey:IssueID" json:"-"`
	RelatedIssue Issue        `gorm:"foreignKey:RelatedIssueID" json:"-"`
	RelationType RelationType `gorm:"foreignKey:RelationTypeID" json:"-"`
}

func (IssueRelation) TableName() string { return "issue_relations" }
