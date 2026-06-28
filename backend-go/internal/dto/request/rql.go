// backend-go/internal/dto/request/rql.go

package request

type RQLSearchRequest struct {
	Entity    string `json:"entity" binding:"required,oneof=issue cycle module"`
	ProjectID uint64 `json:"project_id" binding:"required"`
	RQL       string `json:"rql" binding:"required"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type RQLSearchResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *RQLError   `json:"error,omitempty"`
}

type RQLError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
