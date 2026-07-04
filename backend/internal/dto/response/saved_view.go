package response

import (
	"encoding/json"
	"time"
)

// SavedViewResponse is the API response for a saved view.
type SavedViewResponse struct {
	ID          uint64          `json:"id"`
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	ViewType    string          `json:"view_type"`
	Filters     json.RawMessage `json:"filters"`
	RQL         string          `json:"rql"`
	SortConfig  json.RawMessage `json:"sort_config"`
	Columns     json.RawMessage `json:"columns"`
	GroupBy     *string         `json:"group_by"`
	SubGroupBy  *string         `json:"sub_group_by"`
	IsDefault   bool            `json:"is_default"`
	IsShared    bool            `json:"is_shared"`
	OwnerID     uint64          `json:"owner_id"`
	ProjectID   uint64          `json:"project_id"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
