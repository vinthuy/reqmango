package service

import (
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/testutil"
)

func TestPATService_Create(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewPATService(db)
	mock.ExpectQuery(`INSERT INTO "personal_access_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	token, pat, err := svc.Create(1, "cli-macbook", nil)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(token, PATPrefix))
	assert.Equal(t, "cli-macbook", pat.Name)
	assert.Equal(t, HashToken(token), pat.TokenHash)
	assert.Equal(t, token[:len(PATPrefix)+4], pat.TokenPrefix)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPATService_List(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewPATService(db)
	mock.ExpectQuery(`SELECT \* FROM "personal_access_tokens" WHERE`).
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "cli-macbook"))

	pats, err := svc.List(1)
	require.NoError(t, err)
	assert.Len(t, pats, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPATService_Revoke_NotFound(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewPATService(db)
	mock.ExpectExec(`UPDATE "personal_access_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := svc.Revoke(1, 999)
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPATService_Authenticate_Valid(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewPATService(db)
	mock.ExpectQuery(`SELECT \* FROM "personal_access_tokens" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "revoked_at"}).
			AddRow(1, 2, HashToken("reqmango_pat_abc"), nil, nil))
	mock.ExpectQuery(`SELECT \* FROM "users" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "is_active"}).AddRow(2, true))
	mock.ExpectExec(`UPDATE "personal_access_tokens" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	user, err := svc.Authenticate("reqmango_pat_abc")
	require.NoError(t, err)
	assert.Equal(t, uint64(2), user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPATService_Authenticate_Revoked(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewPATService(db)
	// A revoked token never matches the service's `revoked_at IS NULL` filter,
	// so the lookup returns no rows and Authenticate must reject it.
	mock.ExpectQuery(`SELECT \* FROM "personal_access_tokens" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "revoked_at"}))

	_, err := svc.Authenticate("reqmango_pat_abc")
	require.Error(t, err)
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, common.ErrUnauthorized, appErr.ErrorCode)
	assert.NoError(t, mock.ExpectationsWereMet())
}
