package request

import "encoding/json"

// SavedViewCreateRequest represents the request to create a saved view.
type SavedViewCreateRequest struct {
	Name        string          `json:"name" binding:"required"`
	Description *string         `json:"description"`
	ViewType    string          `json:"view_type"`
	Filters     json.RawMessage `json:"filters"`
	RQL         string          `json:"rql"`
	SortConfig  json.RawMessage `json:"sort_config"`
	Columns     json.RawMessage `json:"columns"`
	GroupBy     *string         `json:"group_by"`
	IsShared    bool            `json:"is_shared"`
}

// SavedViewUpdateRequest represents the request to update a saved view.
type SavedViewUpdateRequest struct {
	Name        *string          `json:"name"`
	Description *string          `json:"description"`
	ViewType    *string          `json:"view_type"`
	Filters     *json.RawMessage `json:"filters"`
	RQL         *string          `json:"rql"`
	SortConfig  *json.RawMessage `json:"sort_config"`
	Columns     *json.RawMessage `json:"columns"`
	GroupBy     *string          `json:"group_by"`
	IsShared    *bool            `json:"is_shared"`
}
