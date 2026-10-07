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
