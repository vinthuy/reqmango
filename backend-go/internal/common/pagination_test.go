package common

import (
	"testing"
)

func TestParsePagination_Defaults(t *testing.T) {
	p := ParsePagination("", "", 20, 100)
	if p.Limit != 20 {
		t.Errorf("default limit = %d, want 20", p.Limit)
	}
	if p.Offset != 0 {
		t.Errorf("default offset = %d, want 0", p.Offset)
	}
}

func TestParsePagination_ValidValues(t *testing.T) {
	p := ParsePagination("10", "5", 20, 100)
	if p.Limit != 10 {
		t.Errorf("limit = %d, want 10", p.Limit)
	}
	if p.Offset != 5 {
		t.Errorf("offset = %d, want 5", p.Offset)
	}
}

func TestParsePagination_LimitExceedsMax(t *testing.T) {
	p := ParsePagination("200", "0", 20, 100)
	if p.Limit != 100 {
		t.Errorf("limit = %d, want 100 (capped at max)", p.Limit)
	}
}

func TestParsePagination_NegativeLimit(t *testing.T) {
	p := ParsePagination("-5", "0", 20, 100)
	if p.Limit != 20 {
		t.Errorf("limit = %d, want 20 (default for invalid)", p.Limit)
	}
}

func TestParsePagination_NegativeOffset(t *testing.T) {
	p := ParsePagination("10", "-5", 20, 100)
	if p.Offset != 0 {
		t.Errorf("offset = %d, want 0 (default for invalid)", p.Offset)
	}
}

func TestParsePagination_ZeroLimit(t *testing.T) {
	p := ParsePagination("0", "5", 20, 100)
	if p.Limit != 20 {
		t.Errorf("limit = %d, want 20 (default for 0)", p.Limit)
	}
}

func TestParsePagination_NonNumericInput(t *testing.T) {
	// Non-numeric limit should fall back to default.
	p := ParsePagination("abc", "xyz", 20, 100)
	if p.Limit != 20 {
		t.Errorf("limit = %d, want 20 (default)", p.Limit)
	}
	if p.Offset != 0 {
		t.Errorf("offset = %d, want 0 (default)", p.Offset)
	}
}

func TestParsePagination_MaxLimitEdge(t *testing.T) {
	// Exactly max limit
	p := ParsePagination("100", "0", 20, 100)
	if p.Limit != 100 {
		t.Errorf("limit = %d, want 100", p.Limit)
	}
}

func TestParsePagination_OnePastMax(t *testing.T) {
	p := ParsePagination("101", "0", 20, 100)
	if p.Limit != 100 {
		t.Errorf("limit = %d, want 100 (capped)", p.Limit)
	}
}

func TestParsePagination_DifferentDefaults(t *testing.T) {
	tests := []struct {
		name         string
		limitStr     string
		offsetStr    string
		defaultLimit int
		maxLimit     int
		wantLimit    int
		wantOffset   int
	}{
		{"default 50 max 200", "", "", 50, 200, 50, 0},
		{"default 10 max 50 with valid", "15", "3", 10, 50, 15, 3},
		{"exceed max 50", "100", "0", 10, 50, 50, 0},
		{"large offset", "10", "1000", 20, 100, 10, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ParsePagination(tt.limitStr, tt.offsetStr, tt.defaultLimit, tt.maxLimit)
			if p.Limit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", p.Limit, tt.wantLimit)
			}
			if p.Offset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", p.Offset, tt.wantOffset)
			}
		})
	}
}

func TestPaginationStruct(t *testing.T) {
	p := Pagination{Limit: 25, Offset: 50}
	if p.Limit != 25 {
		t.Errorf("Limit = %d, want 25", p.Limit)
	}
	if p.Offset != 50 {
		t.Errorf("Offset = %d, want 50", p.Offset)
	}
}
