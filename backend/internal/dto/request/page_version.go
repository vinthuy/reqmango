package request

// PageVersionRestoreRequest restores a page to a previous version.
type PageVersionRestoreRequest struct {
	VersionNumber int `json:"version_number" binding:"required"`
}
