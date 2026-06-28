package request

type TypeTemplateCreate struct {
	Name         string  `json:"name" binding:"required"`
	Color        string  `json:"color"`
	Icon         string  `json:"icon"`
	Description  string  `json:"description"`
	Level        int     `json:"level"`
	ParentTypeID *uint64 `json:"parent_type_id"`
}

type TypeTemplateUpdate struct {
	Name         *string `json:"name"`
	Color        *string `json:"color"`
	Icon         *string `json:"icon"`
	Description  *string `json:"description"`
	Level        *int    `json:"level"`
	ParentTypeID *uint64 `json:"parent_type_id"`
}

type TypeTemplateFieldBind struct {
	FieldID    uint64 `json:"field_id" binding:"required"`
	IsRequired bool   `json:"is_required"`
	Sequence   int    `json:"sequence"`
}
