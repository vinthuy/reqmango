package service

import (
	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/dto/response"
	"github.com/reqmanpy/backend-go/internal/model"
	"gorm.io/gorm"
)

type RelationService struct{ db *gorm.DB }

func NewRelationService(db *gorm.DB) *RelationService { return &RelationService{db: db} }

// ---- Relation Types ----

func (s *RelationService) CreateType(workspaceID, userID uint64, req request.RelationTypeCreate) (*response.RelationTypeResponse, error) {
	rt := model.RelationType{Name: req.Name, InwardName: req.InwardName, OutwardName: req.OutwardName, WorkspaceID: workspaceID}
	rt.CreatedByID = &userID
	if err := s.db.Create(&rt).Error; err != nil {
		return nil, common.Internal("Failed to create relation type")
	}
	return &response.RelationTypeResponse{ID: rt.ID, Name: rt.Name, InwardName: rt.InwardName, OutwardName: rt.OutwardName, WorkspaceID: rt.WorkspaceID, CreatedAt: rt.CreatedAt, UpdatedAt: rt.UpdatedAt}, nil
}

func (s *RelationService) ListTypes(workspaceID uint64) ([]response.RelationTypeResponse, error) {
	var types []model.RelationType
	if err := s.db.Where("workspace_id = ?", workspaceID).Order("created_at").Find(&types).Error; err != nil {
		return nil, common.Internal("Failed to list relation types")
	}
	result := make([]response.RelationTypeResponse, len(types))
	for i, t := range types {
		result[i] = response.RelationTypeResponse{ID: t.ID, Name: t.Name, InwardName: t.InwardName, OutwardName: t.OutwardName, WorkspaceID: t.WorkspaceID, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt}
	}
	if result == nil { result = []response.RelationTypeResponse{} }
	return result, nil
}

func (s *RelationService) UpdateType(id, userID uint64, req request.RelationTypeUpdate) (*response.RelationTypeResponse, error) {
	var t model.RelationType
	if err := s.db.First(&t, id).Error; err != nil { return nil, common.NotFound("Relation type not found") }
	if req.Name != nil { t.Name = *req.Name }
	if req.InwardName != nil { t.InwardName = *req.InwardName }
	if req.OutwardName != nil { t.OutwardName = *req.OutwardName }
	t.UpdatedByID = &userID
	s.db.Save(&t)
	return &response.RelationTypeResponse{ID: t.ID, Name: t.Name, InwardName: t.InwardName, OutwardName: t.OutwardName, WorkspaceID: t.WorkspaceID, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt}, nil
}

func (s *RelationService) DeleteType(id uint64) error {
	var t model.RelationType
	if err := s.db.First(&t, id).Error; err != nil { return common.NotFound("Relation type not found") }
	s.db.Where("relation_type_id = ?", id).Delete(&model.IssueRelation{})
	return s.db.Delete(&t).Error
}

// ---- Issue Relations ----

func (s *RelationService) CreateRelation(issueID uint64, req request.IssueRelationCreate) (*response.IssueRelationResponse, error) {
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil { return nil, common.NotFound("Issue not found") }
	var related model.Issue
	if err := s.db.Preload("Project").First(&related, req.RelatedIssueID).Error; err != nil { return nil, common.NotFound("Related issue not found") }
	var rt model.RelationType
	if err := s.db.First(&rt, req.RelationTypeID).Error; err != nil { return nil, common.NotFound("Relation type not found") }

	var count int64
	s.db.Model(&model.IssueRelation{}).Where("issue_id = ? AND related_issue_id = ? AND relation_type_id = ?", issueID, req.RelatedIssueID, req.RelationTypeID).Count(&count)
	if count > 0 { return nil, common.Conflict("Relation already exists") }

	r := model.IssueRelation{IssueID: issueID, RelatedIssueID: req.RelatedIssueID, RelationTypeID: req.RelationTypeID, Comment: req.Comment}
	if err := s.db.Create(&r).Error; err != nil { return nil, common.Internal("Failed to create relation") }
	return &response.IssueRelationResponse{
		ID: r.ID, IssueID: r.IssueID, RelatedIssueID: r.RelatedIssueID, RelationTypeID: r.RelationTypeID, Comment: r.Comment,
		RelationName: rt.Name, InwardName: rt.InwardName, OutwardName: rt.OutwardName,
		RelatedName: related.Name, RelatedSeqID: related.SequenceID, RelatedProject: related.Project.Identifier,
	}, nil
}

func (s *RelationService) ListRelations(issueID uint64) ([]response.IssueRelationResponse, error) {
	var relations []model.IssueRelation
	if err := s.db.Preload("RelationType").Preload("RelatedIssue.Project").Where("issue_id = ?", issueID).Find(&relations).Error; err != nil {
		return nil, common.Internal("Failed to list relations")
	}
	result := make([]response.IssueRelationResponse, len(relations))
	for i, r := range relations {
		result[i] = response.IssueRelationResponse{
			ID: r.ID, IssueID: r.IssueID, RelatedIssueID: r.RelatedIssueID, RelationTypeID: r.RelationTypeID, Comment: r.Comment,
			RelationName: r.RelationType.Name, InwardName: r.RelationType.InwardName, OutwardName: r.RelationType.OutwardName,
			RelatedName: r.RelatedIssue.Name, RelatedSeqID: r.RelatedIssue.SequenceID, RelatedProject: r.RelatedIssue.Project.Identifier,
		}
	}
	if result == nil { result = []response.IssueRelationResponse{} }
	return result, nil
}

func (s *RelationService) DeleteRelation(relationID uint64) error {
	result := s.db.Delete(&model.IssueRelation{}, relationID)
	if result.RowsAffected == 0 { return common.NotFound("Relation not found") }
	return nil
}
