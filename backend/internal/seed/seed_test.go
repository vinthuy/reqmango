package seed

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/reqmango/backend/internal/testutil"
	"github.com/stretchr/testify/require"
)

// The bootstrap admin is the only account a production install has, so the
// properties worth pinning are: it never runs twice, it never promotes a
// squatter, and it does create one admin on an empty database.

func TestEnsureBootstrapAdmin_NoOpWhenSuperuserExists(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer func() { _ = sqlDB.Close() }()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	created := EnsureBootstrapAdmin(db, "admin@reqmango.com", "secret")

	require.False(t, created, "an existing superuser must not trigger a second bootstrap")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEnsureBootstrapAdmin_RefusesToPromoteExistingUser(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer func() { _ = sqlDB.Close() }()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "is_superuser"}).
			AddRow(1, "admin@reqmango.com", false))

	created := EnsureBootstrapAdmin(db, "admin@reqmango.com", "secret")

	require.False(t, created, "a pre-registered address must not be promoted to superuser")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEnsureBootstrapAdmin_CreatesAdminOnEmptyDatabase(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer func() { _ = sqlDB.Close() }()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "is_superuser"}))
	// GORM returns the generated key, so the insert arrives as a query.
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	created := EnsureBootstrapAdmin(db, "ops@example.com", "a-configured-password")

	require.True(t, created)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRandomPassword_IsURLSafeAndLongEnough(t *testing.T) {
	seen := make(map[string]struct{}, 16)
	for i := 0; i < 16; i++ {
		pw := randomPassword()
		require.Len(t, pw, 24)
		for _, r := range pw {
			require.True(t,
				r == '-' || r == '_' || (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'),
				"unexpected character %q in generated password", r)
		}
		seen[pw] = struct{}{}
	}
	require.Len(t, seen, 16, "generated passwords must not repeat")
}
