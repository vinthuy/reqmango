package service

import (
	"errors"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type RelationService struct{ db *gorm.DB }

func NewRelationService(db *gorm.DB) *RelationService { return &RelationService{db: db} }

// checkWorkspaceAdmin verifies that the caller is an active admin-level member
// of the workspace. Guards relation type mutations against privilege escalation.
func (s *RelationService) checkWorkspaceAdmin(workspaceID, callerID uint64) error {
	var member model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ? AND is_active = ?", workspaceID, callerID, true).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.Forbidden("You must be a workspace admin to manage relation types")
		}
		return common.Internal("Database error")
	}
	if member.Role < common.RoleAdmin {
		return common.Forbidden("You must be a workspace admin to manage relation types")
	}
	return nil
}

// checkProjectMembership verifies that the caller is an active member of the
// project. Issue relation mutations are member-scoped.
func (s *RelationService) checkProjectMembership(projectID, userID uint64) error {
	var count int64
	s.db.Model(&model.ProjectMember{}).
		Where("project_id = ? AND user_id = ? AND is_active = ?", projectID, userID, true).
		Count(&count)
	if count == 0 {
		return common.Forbidden("You must be a member of the project to manage issue relations")
	}
	return nil
}

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

func (s *RelationService) DeleteType(id, callerID uint64) error {
	var t model.RelationType
	if err := s.db.First(&t, id).Error; err != nil { return common.NotFound("Relation type not found") }
	if err := s.checkWorkspaceAdmin(t.WorkspaceID, callerID); err != nil { return err }
	s.db.Where("relation_type_id = ?", id).Delete(&model.IssueRelation{})
	return s.db.Delete(&t).Error
}

// ---- Issue Relations ----

func (s *RelationService) CreateRelation(issueID, callerID uint64, req request.IssueRelationCreate) (*response.IssueRelationResponse, error) {
	var issue model.Issue
	if err := s.db.First(&issue, issueID).Error; err != nil { return nil, common.NotFound("Issue not found") }
	if err := s.checkProjectMembership(issue.ProjectID, callerID); err != nil { return nil, err }
	var related model.Issue
	if err := s.db.Preload("Project").First(&related, req.RelatedIssueID).Error; err != nil { return nil, common.NotFound("Related issue not found") }
	var rt model.RelationType
	if err := s.db.First(&rt, req.RelationTypeID).Error; err != nil { return nil, common.NotFound("Relation type not found") }

	var count int64
	s.db.Model(&model.IssueRelation{}).Where("issue_id = ? AND related_issue_id = ? AND relation_type_id = ?", issueID, req.RelatedIssueID, req.RelationTypeID).Count(&count)
	if count > 0 { return nil, common.Conflict("Relation already exists") }

	r := model.IssueRelation{IssueID: issueID, RelatedIssueID: req.RelatedIssueID, RelationTypeID: req.RelationTypeID, Comment: req.Comment}
	if err := s.db.Create(&r).Error; err != nil { return nil, common.Internal("Failed to create relation") }

	// Record activity on both issues
	relName := rt.OutwardName
	s.db.Create(&model.IssueActivity{IssueID: &issueID, Verb: "relation_added", Field: strPtr("relation"), NewValue: &related.Name, Comment: &relName})
	s.db.Create(&model.IssueActivity{IssueID: &req.RelatedIssueID, Verb: "relation_added", Field: strPtr("relation"), NewValue: &issue.Name, Comment: &relName})

	resp := &response.IssueRelationResponse{
		ID: r.ID, IssueID: r.IssueID, RelatedIssueID: r.RelatedIssueID, RelationTypeID: r.RelationTypeID, Comment: r.Comment,
		Direction: "outbound",
		RelationName: rt.Name, InwardName: rt.InwardName, OutwardName: rt.OutwardName,
		RelatedName: related.Name, RelatedSeqID: related.SequenceID, RelatedProject: related.Project.Identifier,
		RelationType: &response.RelationTypeLite{
			ID: rt.ID, Name: rt.Name, OutwardName: rt.OutwardName,
		},
	}
	return resp, nil
}

func (s *RelationService) ListRelations(issueID uint64) ([]response.IssueRelationResponse, error) {
	// Returns outbound relations only (backward compat)
	return s.listRelationsByDirection(issueID, "outbound")
}

// ListRelationsBidirectional returns both outbound and inbound relations
func (s *RelationService) ListRelationsBidirectional(issueID uint64) ([]response.IssueRelationResponse, error) {
	return s.listRelationsByDirection(issueID, "both")
}

func (s *RelationService) listRelationsByDirection(issueID uint64, direction string) ([]response.IssueRelationResponse, error) {
	var result []response.IssueRelationResponse

	// Outbound: I relate to others (issue_id = ?)
	if direction == "outbound" || direction == "both" {
		var outbound []model.IssueRelation
		if err := s.db.Preload("RelationType").
			Preload("RelatedIssue.State").
			Preload("RelatedIssue.IssueType").
			Preload("RelatedIssue.AssigneeLinks.User").
			Where("issue_id = ?", issueID).Find(&outbound).Error; err != nil {
			return nil, common.Internal("Failed to list outbound relations")
		}
		for _, r := range outbound {
			result = append(result, s.buildRelationResponse(r, "outbound"))
		}
	}

	// Inbound: others relate to me (related_issue_id = ?)
	if direction == "inbound" || direction == "both" {
		var inbound []model.IssueRelation
		if err := s.db.Preload("RelationType").
			Preload("Issue.State").       // The source issue (who relates to me)
			Preload("Issue.IssueType").
			Preload("Issue.AssigneeLinks.User").
			Where("related_issue_id = ?", issueID).Find(&inbound).Error; err != nil {
			return nil, common.Internal("Failed to list inbound relations")
		}
		for _, r := range inbound {
			// For inbound, the "RelatedIssue" is actually the source issue
			result = append(result, s.buildInboundRelationResponse(r))
		}
	}

	if result == nil {
		result = []response.IssueRelationResponse{}
	}
	return result, nil
}

func (s *RelationService) buildRelationResponse(r model.IssueRelation, direction string) response.IssueRelationResponse {
	ri := r.RelatedIssue
	item := response.IssueRelationResponse{
		ID: r.ID, IssueID: r.IssueID, RelatedIssueID: r.RelatedIssueID,
		RelationTypeID: r.RelationTypeID, Comment: r.Comment,
		Direction: direction,
		RelationType: &response.RelationTypeLite{
			ID: r.RelationType.ID, Name: r.RelationType.Name, OutwardName: r.RelationType.OutwardName,
		},
		RelationName: r.RelationType.Name, InwardName: r.RelationType.InwardName,
		OutwardName: r.RelationType.OutwardName,
		RelatedName: ri.Name, RelatedSeqID: ri.SequenceID,
	}
	related := &response.RelatedIssueLite{
		ID: ri.ID, SequenceID: ri.SequenceID, Name: ri.Name,
		StateName: ri.State.Name, StateGroup: ri.State.Group,
		Priority: ri.Priority, TargetDate: ri.TargetDate,
	}
	if ri.IssueType.ID != 0 {
		related.IssueType = &response.IssueTypeLite{
			ID:                  ri.IssueType.ID,
			Name:                ri.IssueType.Name,
			Color:               ri.IssueType.Color,
			Icon:                ri.IssueType.Icon,
			Level:               ri.IssueType.Level,
			AllowedChildTypeIDs: ri.IssueType.AllowedChildTypeIDs,
		}
	}
	for _, al := range ri.AssigneeLinks {
		u := al.User
		related.Assignees = append(related.Assignees, response.UserLite{ID: u.ID, DisplayName: u.DisplayName})
	}
	item.RelatedIssue = related
	return item
}

func (s *RelationService) buildInboundRelationResponse(r model.IssueRelation) response.IssueRelationResponse {
	// For inbound, the source issue is the one that created the relation to us
	src := r.Issue
	item := response.IssueRelationResponse{
		ID: r.ID, IssueID: r.IssueID, RelatedIssueID: r.RelatedIssueID,
		RelationTypeID: r.RelationTypeID, Comment: r.Comment,
		Direction: "inbound",
		RelationType: &response.RelationTypeLite{
			ID: r.RelationType.ID, Name: r.RelationType.Name, OutwardName: r.RelationType.OutwardName,
		},
		RelationName: r.RelationType.Name, InwardName: r.RelationType.InwardName,
		OutwardName: r.RelationType.OutwardName,
		RelatedName: src.Name, RelatedSeqID: src.SequenceID,
	}
	related := &response.RelatedIssueLite{
		ID: src.ID, SequenceID: src.SequenceID, Name: src.Name,
		StateName: src.State.Name, StateGroup: src.State.Group,
		Priority: src.Priority, TargetDate: src.TargetDate,
	}
	if src.IssueType.ID != 0 {
		related.IssueType = &response.IssueTypeLite{
			ID:                  src.IssueType.ID,
			Name:                src.IssueType.Name,
			Color:               src.IssueType.Color,
			Icon:                src.IssueType.Icon,
			Level:               src.IssueType.Level,
			AllowedChildTypeIDs: src.IssueType.AllowedChildTypeIDs,
		}
	}
	for _, al := range src.AssigneeLinks {
		u := al.User
		related.Assignees = append(related.Assignees, response.UserLite{ID: u.ID, DisplayName: u.DisplayName})
	}
	item.RelatedIssue = related
	return item
}

func (s *RelationService) DeleteRelation(relationID, callerID uint64) error {
	var rel model.IssueRelation
	if err := s.db.Preload("RelationType").Preload("Issue").Preload("RelatedIssue").First(&rel, relationID).Error; err != nil {
		return common.NotFound("Relation not found")
	}
	if err := s.checkProjectMembership(rel.Issue.ProjectID, callerID); err != nil { return err }
	result := s.db.Delete(&model.IssueRelation{}, relationID)
	if result.RowsAffected == 0 { return common.NotFound("Relation not found") }

	// Record activity on both issues
	relName := rel.RelationType.OutwardName
	s.db.Create(&model.IssueActivity{IssueID: &rel.IssueID, Verb: "relation_removed", Field: strPtr("relation"), NewValue: &rel.RelatedIssue.Name, Comment: &relName})
	s.db.Create(&model.IssueActivity{IssueID: &rel.RelatedIssueID, Verb: "relation_removed", Field: strPtr("relation"), NewValue: &rel.Issue.Name, Comment: &relName})

	return nil
}
