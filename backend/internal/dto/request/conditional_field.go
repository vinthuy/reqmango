package request

// ConditionalFieldCreateRequest 创建条件字段请求
type ConditionalFieldCreateRequest struct {
	WorkspaceID     uint64   `json:"workspace_id"`
	FieldID         uint64   `json:"field_id" binding:"required"`
	ConditionType   string   `json:"condition_type" binding:"required"`   // type, state, priority, assignee, label
	Operator        string   `json:"operator" binding:"required"`        // equals, not_equals, contains, not_contains, is_empty, is_not_empty
	ConditionValues []string `json:"condition_values"`                    // 条件值数组
	IsEnabled       bool     `json:"is_enabled"`
	Priority        int      `json:"priority"`
	Description     string   `json:"description"`
}

// ConditionalFieldUpdateRequest 更新条件字段请求
type ConditionalFieldUpdateRequest struct {
	ConditionType   *string  `json:"condition_type"`
	Operator        *string `json:"operator"`
	ConditionValues []string `json:"condition_values"`
	IsEnabled       *bool   `json:"is_enabled"`
	Priority        *int    `json:"priority"`
	Description     *string `json:"description"`
}

// ConditionalFieldResponse 条件字段响应
type ConditionalFieldResponse struct {
	ID              uint64   `json:"id"`
	WorkspaceID     uint64   `json:"workspace_id"`
	FieldID         uint64   `json:"field_id"`
	ConditionType   string   `json:"condition_type"`
	Operator        string   `json:"operator"`
	ConditionValues []string `json:"condition_values"`
	IsEnabled       bool     `json:"is_enabled"`
	Priority        int      `json:"priority"`
	Description     string   `json:"description"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}
