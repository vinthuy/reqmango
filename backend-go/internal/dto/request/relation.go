package request

type RelationTypeCreate struct {
	Name        string `json:"name" binding:"required"`
	InwardName  string `json:"inward_name" binding:"required"`
	OutwardName string `json:"outward_name" binding:"required"`
}

type RelationTypeUpdate struct {
	Name        *string `json:"name"`
	InwardName  *string `json:"inward_name"`
	OutwardName *string `json:"outward_name"`
}

type IssueRelationCreate struct {
	RelatedIssueID uint64  `json:"related_issue_id" binding:"required"`
	RelationTypeID uint64  `json:"relation_type_id" binding:"required"`
	Comment        *string `json:"comment"`
}
