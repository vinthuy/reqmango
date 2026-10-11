package rql

import (
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
		"milestone":   "release_issues",
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

// Two state_group conditions each need the states join; Postgres rejects the
// query if the same table is joined twice.
func TestExecute_JoinsStatesOnceForRepeatedStateGroup(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	node, err := ParseStrict(`state_group != "completed" AND state_group != "cancelled"`)
	require.NoError(t, err)

	dry := db.Session(&gorm.Session{DryRun: true}).Model(&model.Issue{})
	q, err := NewGORMExecutor().Execute(dry, node, NewIssueQueryContext(dry, 1))
	require.NoError(t, err)
	sql := q.Find(&[]model.Issue{}).Statement.SQL.String()

	assert.Equal(t, 1, strings.Count(sql, "JOIN states"), sql)
}
