package response

import (
	"encoding/json"
	"time"
)

type SearchTemplateResponse struct {
	ID           uint64          `json:"id"`
	Name         string          `json:"name"`
	Description  *string         `json:"description"`
	Icon         string          `json:"icon"`
	RQLTemplate  string          `json:"rql_template"`
	ViewType     string          `json:"view_type"`
	SortConfig   json.RawMessage `json:"sort_config"`
	GroupBy      *string         `json:"group_by"`
	Columns      json.RawMessage `json:"columns"`
	IsBuiltIn    bool     `json:"is_built_in"`
	IsPublic     bool     `json:"is_public"`
	OwnerID      *uint64  `json:"owner_id"`
	ProjectID    uint64   `json:"project_id"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}