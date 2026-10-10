package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsIssueMaps_AcceptsTypedAndUntypedSlices(t *testing.T) {
	typed := []map[string]interface{}{{"id": uint64(1), "name": "a"}}
	assert.Equal(t, typed, asIssueMaps(typed))

	untyped := []interface{}{
		map[string]interface{}{"id": float64(2), "name": "b"},
		"skip-me",
	}
	got := asIssueMaps(untyped)
	assert.Len(t, got, 1)
	assert.Equal(t, "b", got[0]["name"])

	assert.Nil(t, asIssueMaps("nope"))
}

func TestBuildAISearchRQL(t *testing.T) {
	assert.Equal(t, "", buildAISearchRQL(nil))
	assert.Equal(t, "", buildAISearchRQL([]map[string]interface{}{{"name": "x"}}))
	assert.Equal(t, "sequence_id = 42", buildAISearchRQL([]map[string]interface{}{
		{"sequence_id": float64(42), "name": "a"},
	}))
	assert.Equal(t, "sequence_id IN (3, 7, 11)", buildAISearchRQL([]map[string]interface{}{
		{"sequence_id": float64(3)},
		{"sequence_id": uint64(7)},
		{"sequence_id": float64(3)}, // dedupe
		{"sequence_id": float64(11)},
	}))
}
