package model

// ConditionalField 定义条件字段规则
type ConditionalField struct {
	BaseModel

	// 关联的工作空间
	WorkspaceID uint64 `gorm:"not null;index" json:"workspace_id"`

	// 关联的自定义字段
	FieldID uint64 `gorm:"not null;index" json:"field_id"`

	// 条件类型: type, state, priority, assignee, label
	ConditionType string `gorm:"size:50;not null" json:"condition_type"`

	// 条件操作符: equals, not_equals, contains, not_contains, is_empty, is_not_empty
	Operator string `gorm:"size:50;not null" json:"operator"`

	// 条件值（JSON数组，存储多个可能的值）
	ConditionValues string `gorm:"type:text" json:"condition_values"`

	// 是否启用
	IsEnabled bool `gorm:"default:true" json:"is_enabled"`

	// 优先级（用于排序）
	Priority int `gorm:"default:0" json:"priority"`

	// 描述
	Description string `gorm:"size:255" json:"description"`
}

func (ConditionalField) TableName() string {
	return "conditional_fields"
}

// FieldCondition 保存字段的条件值
type FieldCondition struct {
	Values []string `json:"values"`
}
