package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRound2(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{"zero", 0, 0},
		{"exact two decimals", 12.34, 12.34},
		{"rounds third decimal up", 12.345, 12.35},
		{"rounds third decimal down", 12.344, 12.34},
		{"large value", 99999.999, 100000.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, round2(tt.in))
		})
	}
}

func TestNormalizeBucket(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", "day"},
		{"day", "day"},
		{"week", "week"},
		{"month", "month"},
		{"YEAR", "day"},       // unsupported falls back to day
		{"hour", "day"},       // unsupported falls back to day
	}
	for _, tt := range tests {
		t.Run("bucket_"+tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeBucket(tt.in))
		})
	}
}

func TestBucketEnd(t *testing.T) {
	start := time.Date(2026, 8, 7, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, start.AddDate(0, 0, 1), bucketEnd(start, "day"))
	assert.Equal(t, start.AddDate(0, 0, 7), bucketEnd(start, "week"))
	assert.Equal(t, start.AddDate(0, 1, 0), bucketEnd(start, "month"))
	// Unknown bucket defaults to daily end.
	assert.Equal(t, start.AddDate(0, 0, 1), bucketEnd(start, "century"))
}

func TestParsePeriodRange(t *testing.T) {
	t.Run("both empty returns nil nil nil", func(t *testing.T) {
		from, to, err := parsePeriodRange("", "")
		assert.NoError(t, err)
		assert.Nil(t, from)
		assert.Nil(t, to)
	})

	t.Run("valid bounds", func(t *testing.T) {
		fromStr := "2026-01-01T00:00:00Z"
		toStr := "2026-02-01T00:00:00Z"
		from, to, err := parsePeriodRange(fromStr, toStr)
		assert.NoError(t, err)
		assert.NotNil(t, from)
		assert.NotNil(t, to)
		assert.True(t, from.Before(*to))
	})

	t.Run("invalid from format", func(t *testing.T) {
		_, _, err := parsePeriodRange("not-a-date", "")
		assert.Error(t, err)
	})

	t.Run("invalid to format", func(t *testing.T) {
		_, _, err := parsePeriodRange("", "2026-13-99")
		assert.Error(t, err)
	})

	t.Run("from after to returns error", func(t *testing.T) {
		fromStr := "2026-02-01T00:00:00Z"
		toStr := "2026-01-01T00:00:00Z"
		_, _, err := parsePeriodRange(fromStr, toStr)
		assert.Error(t, err)
	})
}
