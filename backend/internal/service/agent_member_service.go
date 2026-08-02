package service

import (
	"errors"

	"gorm.io/gorm"
)

// AgentMemberService manages agent membership in projects.
type AgentMemberService struct {
	db *gorm.DB
}

// NewAgentMemberService creates a new AgentMemberService.
func NewAgentMemberService(db *gorm.DB) *AgentMemberService {
	return &AgentMemberService{db: db}
}

// ProjectAgentMember represents an agent's membership in a project.
type ProjectAgentMember struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID uint64 `gorm:"not null;uniqueIndex:idx_proj_agent_member" json:"project_id"`
	AgentID   uint64 `gorm:"not null;uniqueIndex:idx_proj_agent_member" json:"agent_id"`
	Role      string `gorm:"size:20;default:member" json:"role"` // observer, member, admin
	IsActive  bool   `gorm:"default:true" json:"is_active"`
}

func (ProjectAgentMember) TableName() string {
	return "project_agent_members"
}

// AgentMemberResponse represents an agent member with agent details.
type AgentMemberResponse struct {
	ID        uint64 `json:"id"`
	ProjectID uint64 `json:"project_id"`
	AgentID   uint64 `json:"agent_id"`
	AgentName string `json:"agent_name"`
	AgentType string `json:"agent_type"`
	Avatar    string `json:"avatar"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
}

// ListByProject returns all agent members for a project.
func (s *AgentMemberService) ListByProject(projectID uint64) ([]AgentMemberResponse, error) {
	var members []ProjectAgentMember
	if err := s.db.Where("project_id = ?", projectID).Find(&members).Error; err != nil {
		return nil, err
	}

	// Build response with agent details
	var result []AgentMemberResponse
	for _, m := range members {
		resp := AgentMemberResponse{
			ID:        m.ID,
			ProjectID: m.ProjectID,
			AgentID:   m.AgentID,
			Role:      m.Role,
			IsActive:  m.IsActive,
		}

		// Get agent details
		var agent struct {
			Name      string `json:"name"`
			AgentType string `json:"agent_type"`
			Avatar    string `json:"avatar"`
		}
		if err := s.db.Raw("SELECT name, agent_type, avatar FROM agents WHERE id = ?", m.AgentID).Scan(&agent).Error; err == nil {
			resp.AgentName = agent.Name
			resp.AgentType = agent.AgentType
			resp.Avatar = agent.Avatar
		}

		result = append(result, resp)
	}

	if result == nil {
		result = []AgentMemberResponse{}
	}

	return result, nil
}

// Add adds an agent to a project.
func (s *AgentMemberService) Add(projectID, agentID uint64, role string) (*ProjectAgentMember, error) {
	if role == "" {
		role = "member"
	}

	// Check if already exists
	var existing ProjectAgentMember
	if err := s.db.Where("project_id = ? AND agent_id = ?", projectID, agentID).First(&existing).Error; err == nil {
		return &existing, nil // already exists, return existing
	}

	member := &ProjectAgentMember{
		ProjectID: projectID,
		AgentID:   agentID,
		Role:      role,
		IsActive:  true,
	}

	if err := s.db.Create(member).Error; err != nil {
		return nil, err
	}

	return member, nil
}

// UpdateRole updates an agent's role in a project.
func (s *AgentMemberService) UpdateRole(projectID, agentID uint64, role string) error {
	result := s.db.Model(&ProjectAgentMember{}).
		Where("project_id = ? AND agent_id = ?", projectID, agentID).
		Update("role", role)

	if result.RowsAffected == 0 {
		return errors.New("agent member not found")
	}

	return result.Error
}

// Remove removes an agent from a project.
func (s *AgentMemberService) Remove(projectID, agentID uint64) error {
	result := s.db.Where("project_id = ? AND agent_id = ?", projectID, agentID).Delete(&ProjectAgentMember{})
	if result.RowsAffected == 0 {
		return errors.New("agent member not found")
	}
	return result.Error
}

// UpdateRoleByMemberID updates a member's role by member ID.
func (s *AgentMemberService) UpdateRoleByMemberID(projectID, memberID uint64, role string) error {
	result := s.db.Model(&ProjectAgentMember{}).
		Where("id = ? AND project_id = ?", memberID, projectID).
		Update("role", role)

	if result.RowsAffected == 0 {
		return errors.New("agent member not found")
	}

	return result.Error
}

// RemoveByMemberID removes a member by member ID.
func (s *AgentMemberService) RemoveByMemberID(projectID, memberID uint64) error {
	result := s.db.Where("id = ? AND project_id = ?", memberID, projectID).Delete(&ProjectAgentMember{})
	if result.RowsAffected == 0 {
		return errors.New("agent member not found")
	}
	return result.Error
}

// IsMember checks if an agent is a member of a project.
func (s *AgentMemberService) IsMember(projectID, agentID uint64) (bool, error) {
	var count int64
	err := s.db.Model(&ProjectAgentMember{}).
		Where("project_id = ? AND agent_id = ? AND is_active = ?", projectID, agentID, true).
		Count(&count).Error
	return count > 0, err
}
