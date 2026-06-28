package model

// ProjectPageTab is a user-customizable tab on the project detail page.
// Supports both built-in tabs (toggle on/off) and custom tabs (user-defined links to views).
type ProjectPageTab struct {
	BaseModel

	ProjectID uint64 `gorm:"not null;index" json:"project_id"`
	OwnerID   uint64 `gorm:"not null;index" json:"owner_id"`

	// Display name shown on the tab
	Name string `gorm:"size:50;not null" json:"name"`

	// Icon identifier (emoji or SVG path key)
	Icon string `gorm:"size:30" json:"icon"`

	// Tab type: "issues"|"cycles"|"modules"|"updates"|"pages"|"settings"|"custom"
	// Standard types use pre-built components; "custom" links to a saved view or URL
	TabType string `gorm:"size:30;not null;default:'custom'" json:"tab_type"`

	// For standard tabs: internal route key used by the frontend
	RouteKey string `gorm:"size:50" json:"route_key"`

	// For custom tabs: optional target (saved_view_id or external URL)
	TargetType string `gorm:"size:20" json:"target_type"` // "saved_view" | "url" | ""
	TargetID   *uint64 `json:"target_id"`                  // saved_view_id when target_type=saved_view
	TargetURL  string  `gorm:"size:500" json:"target_url"` // external URL when target_type=url

	// Whether this tab is visible; built-in tabs default false unless explicitly enabled
	Visible bool `gorm:"default:true" json:"visible"`

	// Display ordering (0-based ascending)
	SortOrder int `gorm:"default:0" json:"sort_order"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID" json:"-"`
}
