package common

import "strconv"

// Pagination holds standard pagination parameters.
type Pagination struct {
	Limit  int
	Offset int
}

// ParsePagination extracts limit and offset from query params with defaults.
func ParsePagination(limitStr, offsetStr string, defaultLimit, maxLimit int) Pagination {
	limit := defaultLimit
	offset := 0

	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
		if limit > maxLimit {
			limit = maxLimit
		}
	}

	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	return Pagination{Limit: limit, Offset: offset}
}
