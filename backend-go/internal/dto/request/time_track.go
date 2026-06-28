package request

// TimeTrackStartRequest starts a timer.
type TimeTrackStartRequest struct {
	Description *string `json:"description"`
}

// TimeTrackStopRequest stops a timer.
type TimeTrackStopRequest struct {
	Description *string `json:"description"`
}
