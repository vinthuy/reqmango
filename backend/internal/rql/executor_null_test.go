package rql

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The *_id fields are join-table backed; an emptiness check must not reference
// a column issues does not have.
func TestBuildNullRaw_JoinTableIDFieldsUseExists(t *testing.T) {
	ctx := NewIssueQueryContext(nil, 1)
	e := &GORMExecutor{}
	cases := map[string]string{
		"assignee_id": "issue_assignees",
		"label_id":    "issue_labels",
		"cycle_id":    "issue_cycles",
		"module_id":   "module_issues",
	}
	for field, table := range cases {
		isNull, err := e.buildNullRaw(&NullCheck{Field: field, Operator: "IS NULL"}, ctx)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(isNull.SQL, "NOT EXISTS"), field)
		assert.Contains(t, isNull.SQL, table, field)

		notNull, err := e.buildNullRaw(&NullCheck{Field: field, Operator: "IS NOT NULL"}, ctx)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(notNull.SQL, "EXISTS"), field)
	}

	dates, err := e.buildNullRaw(&NullCheck{Field: "target_date", Operator: "IS NULL"}, ctx)
	require.NoError(t, err)
	assert.Equal(t, "issues.target_date IS NULL", dates.SQL)
}
