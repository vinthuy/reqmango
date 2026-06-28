package response

import "github.com/reqmanpy/backend/internal/model"

// RoleResponse for role API responses.
type RoleResponse struct {
	ID          uint64               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Scope       string               `json:"scope"`
	WorkspaceID *uint64              `json:"workspace_id"`
	ProjectID   *uint64              `json:"project_id"`
	IsSystem    bool                 `json:"is_system"`
	SortOrder   int                  `json:"sort_order"`
	Level       int                  `json:"level"`
	Permissions []PermissionResponse `json:"permissions"`
	CreatedAt   string               `json:"created_at"`
}

// PermissionResponse for permission API responses.
type PermissionResponse struct {
	ID          uint64 `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Scope       string `json:"scope"`
}

// ToRoleResponse converts a model.Role to response.
func ToRoleResponse(r *model.Role) RoleResponse {
	resp := RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Scope:       r.Scope,
		WorkspaceID: r.WorkspaceID,
		ProjectID:   r.ProjectID,
		IsSystem:    r.IsSystem,
		SortOrder:   r.SortOrder,
		Level:       r.Level,
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	for _, p := range r.Permissions {
		resp.Permissions = append(resp.Permissions, PermissionResponse{
			ID:          p.ID,
			Code:        p.Code,
			Name:        p.Name,
			Description: p.Description,
			Resource:    p.Resource,
			Action:      p.Action,
			Scope:       p.Scope,
		})
	}
	return resp
}

// ToPermissionResponse converts a model.Permission to response.
func ToPermissionResponse(p *model.Permission) PermissionResponse {
	return PermissionResponse{
		ID:          p.ID,
		Code:        p.Code,
		Name:        p.Name,
		Description: p.Description,
		Resource:    p.Resource,
		Action:      p.Action,
		Scope:       p.Scope,
	}
}
