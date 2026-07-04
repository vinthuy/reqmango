package service

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/testutil"
)

var savedReportTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

var savedReportCols = []string{
	"id", "created_at", "updated_at", "deleted_at", "created_by_id", "updated_by_id",
	"name", "report_type", "group_by", "chart_type", "rql", "interval", "date_from", "date_to", "project_id",
}

func newReportRow(id uint64, name, reportType, groupBy, chartType string, projectID uint64) *sqlmock.Rows {
	return sqlmock.NewRows(savedReportCols).
		AddRow(id, savedReportTime, savedReportTime, nil, nil, nil,
			name, reportType, groupBy, chartType, "", "", "", "", projectID)
}

// ==================== List ====================

func TestSavedReportService_List(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()
	svc := NewSavedReportService(db)

	mock.ExpectQuery(`SELECT \* FROM "saved_reports" WHERE`).
		WithArgs(uint64(1)).
		WillReturnRows(newReportRow(1, "Report1", "distribution", "state", "bar", 1))

	reports, err := svc.List(1)
	require.NoError(t, err)
	assert.Len(t, reports, 1)
	assert.Equal(t, "Report1", reports[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSavedReportService_List_Empty(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()
	svc := NewSavedReportService(db)

	mock.ExpectQuery(`SELECT \* FROM "saved_reports" WHERE`).
		WithArgs(uint64(99)).
		WillReturnRows(sqlmock.NewRows(savedReportCols))

	reports, err := svc.List(99)
	require.NoError(t, err)
	assert.Len(t, reports, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== Create ====================

func TestSavedReportService_Create(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()
	svc := NewSavedReportService(db)

	req := &request.SavedReportCreateRequest{
		Name:       "Monthly Report",
		ReportType: "distribution",
		GroupBy:    "state",
		ChartType:  "bar",
		RQL:        "priority = high",
		Interval:   "week",
		DateFrom:   "2024-01-01",
		DateTo:     "2024-06-30",
	}

	// GORM INSERT with SkipDefaultTransaction — no transaction wrapping
	mock.ExpectQuery(`INSERT INTO "saved_reports"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(101))

	report, err := svc.Create(5, req)
	require.NoError(t, err)
	assert.Equal(t, "Monthly Report", report.Name)
	assert.Equal(t, "distribution", report.ReportType)
	assert.Equal(t, uint64(101), report.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSavedReportService_Create_Validation(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()
	svc := NewSavedReportService(db)

	_, err := svc.Create(1, &request.SavedReportCreateRequest{Name: "", ReportType: "distribution"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Name and report_type are required")

	_, err = svc.Create(1, &request.SavedReportCreateRequest{Name: "R", ReportType: ""})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Name and report_type are required")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== Update ====================

func TestSavedReportService_Update(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()
	svc := NewSavedReportService(db)

	newName := "Updated Report"
	req := &request.SavedReportUpdateRequest{Name: &newName}

	// GORM First with WHERE id=$1 AND project_id=$2 LIMIT $3
	mock.ExpectQuery(`SELECT`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(newReportRow(1, "Old", "distribution", "state", "bar", 10))

	// GORM Update (in-memory apply + DB update)
	mock.ExpectExec(`UPDATE`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	report, err := svc.Update(1, 10, req)
	require.NoError(t, err)
	assert.Equal(t, "Updated Report", report.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSavedReportService_Update_NotFound(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()
	svc := NewSavedReportService(db)

	newName := "X"
	mock.ExpectQuery(`SELECT \* FROM "saved_reports" WHERE`).
		WithArgs(uint64(999), uint64(10), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := svc.Update(999, 10, &request.SavedReportUpdateRequest{Name: &newName})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== Delete ====================

func TestSavedReportService_Delete(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()
	svc := NewSavedReportService(db)

	// GORM soft-delete = UPDATE (single statement, no transaction)
	mock.ExpectExec(`UPDATE "saved_reports" SET "deleted_at"=`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.Delete(1, 5)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSavedReportService_Delete_NotFound(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()
	svc := NewSavedReportService(db)

	mock.ExpectExec(`UPDATE "saved_reports" SET "deleted_at"=`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := svc.Delete(999, 5)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== Model ====================

func TestSavedReport_TableName(t *testing.T) {
	r := model.SavedReport{}
	assert.Equal(t, "saved_reports", r.TableName())
}
