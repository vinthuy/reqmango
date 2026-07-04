package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/testutil"
)

var testTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

// allDashCols covers all SavedDashboard model fields (15 columns).
var allDashCols = []string{
	"id", "created_at", "updated_at", "deleted_at", "created_by_id", "updated_by_id",
	"name", "description", "is_default", "is_shared", "owner_id", "project_id", "date_from", "date_to", "columns",
}

func newDashRow(id uint64, name string, isDefault, isShared bool, ownerID, projectID uint64, cols int) *sqlmock.Rows {
	return sqlmock.NewRows(allDashCols).
		AddRow(id, testTime, testTime, nil, nil, nil, name, nil, isDefault, isShared, ownerID, projectID, nil, nil, cols)
}

func newFullDashRow(id uint64, name string, desc *string, isDefault, isShared bool, ownerID, projectID uint64, dateFrom, dateTo *string, cols int) *sqlmock.Rows {
	return sqlmock.NewRows(allDashCols).
		AddRow(id, testTime, testTime, nil, nil, nil, name, desc, isDefault, isShared, ownerID, projectID, dateFrom, dateTo, cols)
}

func allWgtCols() []string {
	return []string{"id", "created_at", "updated_at", "deleted_at", "created_by_id", "updated_by_id",
		"dashboard_id", "widget_type", "title", "description", "config", "position", "sort_order"}
}

var wgtCols = allWgtCols()

func newWgtRow(id, dashID uint64, wtype, title string, sortOrder int) *sqlmock.Rows {
	return sqlmock.NewRows(wgtCols).
		AddRow(id, testTime, testTime, nil, nil, nil, dashID, wtype, title, nil, json.RawMessage(`{}`), json.RawMessage(`{}`), sortOrder)
}

// ==================== Dashboard CRUD ====================

func TestDashboardService_Create(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewDashboardService(db)
	mock.ExpectQuery(`INSERT INTO`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	resp, err := svc.Create(&request.CreateDashboardRequest{Name: "My Dashboard", Columns: 12}, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, "My Dashboard", resp.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_Create_DefaultColumns(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewDashboardService(db)
	mock.ExpectQuery(`INSERT INTO`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	resp, err := svc.Create(&request.CreateDashboardRequest{Name: "Auto", Columns: 0}, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 12, resp.Columns)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_Create_WithIsDefault(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewDashboardService(db)
	mock.ExpectExec(`UPDATE`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`INSERT INTO`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

	resp, err := svc.Create(&request.CreateDashboardRequest{Name: "Default Dash", IsDefault: true}, 1, 2)
	assert.NoError(t, err)
	assert.True(t, resp.IsDefault)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_Get_NotFound(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewDashboardService(db)
	// GORM adds deleted_at IS NULL, so query doesn't match 15-column row
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows(allDashCols))

	resp, err := svc.Get(999, 1, 2)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_Delete(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewDashboardService(db)
	// GORM soft-delete = UPDATE "saved_dashboards" SET "deleted_at"=$1
	// WHERE id=$2 AND project_id=$3 AND owner_id=$4 AND "deleted_at" IS NULL
	mock.ExpectExec(`UPDATE`).WithArgs(sqlmock.AnyArg(), uint64(1), uint64(1), uint64(2)).WillReturnResult(sqlmock.NewResult(0, 1))
	// Also soft-deletes associated widgets
	mock.ExpectExec(`UPDATE`).WithArgs(sqlmock.AnyArg(), uint64(1)).WillReturnResult(sqlmock.NewResult(0, 0))

	err := svc.Delete(1, 1, 2)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_Delete_NotFound(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewDashboardService(db)
	mock.ExpectExec(`UPDATE`).WithArgs(sqlmock.AnyArg(), uint64(999), uint64(1), uint64(2)).WillReturnResult(sqlmock.NewResult(0, 0))

	err := svc.Delete(999, 1, 2)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_SetDefault(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewDashboardService(db)
	// GORM First = SELECT ... WHERE id=$1 AND project_id=$2 AND owner_id=$3 ... LIMIT $4
	mock.ExpectQuery(`SELECT`).
		WithArgs(uint64(1), uint64(1), uint64(2), 1).
		WillReturnRows(newDashRow(1, "Best Dash", false, false, 2, 1, 12))
	mock.ExpectExec(`UPDATE`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`UPDATE`).WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := svc.SetDefault(1, 1, 2)
	assert.NoError(t, err)
	assert.True(t, resp.IsDefault)
	assert.Equal(t, "Best Dash", resp.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_Duplicate(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewDashboardService(db)

	// find source
	mock.ExpectQuery(`SELECT`).WillReturnRows(newFullDashRow(1, "Source", nil, false, false, 2, 1, nil, nil, 6))
	// preload widgets
	mock.ExpectQuery(`SELECT`).WillReturnRows(newWgtRow(10, 1, "number_card", "Total", 0))
	// create clone
	mock.ExpectQuery(`INSERT INTO`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	// duplicate widget
	mock.ExpectQuery(`INSERT INTO`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(20))
	// reload
	mock.ExpectQuery(`SELECT`).WillReturnRows(newDashRow(2, "Source (Copy)", false, false, 2, 1, 6))
	mock.ExpectQuery(`SELECT`).WillReturnRows(newWgtRow(20, 2, "number_card", "Total", 0))

	resp, err := svc.Duplicate(1, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, "Source (Copy)", resp.Name)
	assert.Len(t, resp.Widgets, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== Widget CRUD ====================

func TestDashboardService_AddWidget(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewDashboardService(db)

	mock.ExpectQuery(`SELECT`).WillReturnRows(newDashRow(1, "My Dash", false, false, 2, 1, 12))
	mock.ExpectQuery(`INSERT INTO`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

	resp, err := svc.AddWidget(1, 1, 2, &request.CreateWidgetRequest{WidgetType: "number_card", Title: "Issue Count"})
	assert.NoError(t, err)
	assert.Equal(t, "number_card", resp.WidgetType)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_UpdateWidget(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewDashboardService(db)

	mock.ExpectQuery(`SELECT`).WillReturnRows(newDashRow(1, "My Dash", false, false, 2, 1, 12))
	mock.ExpectQuery(`SELECT`).WillReturnRows(newWgtRow(10, 1, "number_card", "Old", 0))
	mock.ExpectExec(`UPDATE`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT`).WillReturnRows(newWgtRow(10, 1, "number_card", "Updated Widget", 0))

	newTitle := "Updated Widget"
	resp, err := svc.UpdateWidget(1, 10, 1, 2, &request.UpdateWidgetRequest{Title: &newTitle})
	assert.NoError(t, err)
	assert.Equal(t, "Updated Widget", resp.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_DeleteWidget(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewDashboardService(db)

	mock.ExpectQuery(`SELECT`).WillReturnRows(newDashRow(1, "My Dash", false, false, 2, 1, 12))
	// GORM soft-delete = UPDATE "dashboard_widgets" SET "deleted_at"=$1
	// WHERE id=$2 AND dashboard_id=$3 AND "deleted_at" IS NULL
	mock.ExpectExec(`UPDATE`).WithArgs(sqlmock.AnyArg(), uint64(10), uint64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.DeleteWidget(1, 10, 1, 2)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_DeleteWidget_NotFound(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewDashboardService(db)

	mock.ExpectQuery(`SELECT`).WillReturnRows(newDashRow(1, "My Dash", false, false, 2, 1, 12))
	mock.ExpectExec(`UPDATE`).WithArgs(sqlmock.AnyArg(), uint64(999), uint64(1)).WillReturnResult(sqlmock.NewResult(0, 0))

	err := svc.DeleteWidget(1, 999, 1, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Widget not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_ReorderWidgets(t *testing.T) {
	db, mock, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()

	svc := NewDashboardService(db)

	mock.ExpectQuery(`SELECT`).WillReturnRows(newDashRow(1, "My Dash", false, false, 2, 1, 12))
	mock.ExpectExec(`UPDATE`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE`).WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.ReorderWidgets(1, 1, 2, &request.ReorderWidgetsRequest{WidgetIDs: []uint64{30, 10, 20}})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== Helpers ====================

func TestDashboardService_HelperConversions(t *testing.T) {
	db, _, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()
	_ = NewDashboardService(db)

	assert.Equal(t, json.RawMessage("{}"), normalizeJSON(nil))
	assert.Equal(t, json.RawMessage("{}"), normalizeJSON(json.RawMessage("")))
	assert.Equal(t, json.RawMessage("{}"), normalizeJSON(json.RawMessage("null")))
	assert.Equal(t, json.RawMessage(`{"key":"value"}`), normalizeJSON(json.RawMessage(`{"key":"value"}`)))
}

func TestDashboardService_ResponseConversion(t *testing.T) {
	db, _, sqlDB := testutil.NewMockDB(t)
	defer sqlDB.Close()
	_ = NewDashboardService(db)

	t.Run("dashboardToResponse", func(t *testing.T) {
		d := &model.SavedDashboard{
			BaseModel: model.BaseModel{ID: 1},
			Name:      "Test Dash",
			IsShared:  true,
			OwnerID:   2,
			ProjectID: 1,
			Columns:   6,
		}
		resp := dashboardToResponse(d)
		assert.Equal(t, "Test Dash", resp.Name)
		assert.True(t, resp.IsShared)
		assert.Equal(t, 6, resp.Columns)
	})

	t.Run("widgetToResponse", func(t *testing.T) {
		w := &model.DashboardWidget{
			BaseModel:   model.BaseModel{ID: 42},
			DashboardID: 7,
			WidgetType:  "pie_chart",
			Title:       "Distribution",
			SortOrder:   3,
		}
		resp := widgetToResponse(w)
		assert.Equal(t, uint64(42), resp.ID)
		assert.Equal(t, "pie_chart", resp.WidgetType)
		assert.Equal(t, 3, resp.SortOrder)
	})
}
