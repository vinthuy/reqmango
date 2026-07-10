package service

import (
	"encoding/json"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// DashboardService handles dashboard business logic.
type DashboardService struct {
	db              *gorm.DB
	// Lazily initialized services for widget data rendering
	reportSvc       *ReportService
	savedReportSvc  *SavedReportService
	cycleSvc        *CycleService
	issueSvc        *IssueService
}

// NewDashboardService creates a new DashboardService.
func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{
		db:             db,
		reportSvc:      NewReportService(db),
		savedReportSvc: NewSavedReportService(db),
		cycleSvc:       NewCycleService(db, nil),
		issueSvc:       nil, // issueSvc needs notificationSvc, so we leave it nil for now
	}
}

// SetIssueService sets the issue service for widget data rendering.
func (s *DashboardService) SetIssueService(svc *IssueService) {
	s.issueSvc = svc
}

// ==================== Dashboard CRUD ====================

// List returns all dashboards for a project accessible to a user.
func (s *DashboardService) List(projectID, userID uint64) ([]response.DashboardResponse, error) {
	var dashboards []model.SavedDashboard
	if err := s.db.Where("project_id = ? AND (owner_id = ? OR is_shared = ?)", projectID, userID, true).
		Preload("Widgets", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Order("is_default DESC, created_at ASC").
		Find(&dashboards).Error; err != nil {
		return nil, common.Internal("Failed to fetch dashboards")
	}

	resps := make([]response.DashboardResponse, len(dashboards))
	for i, d := range dashboards {
		resps[i] = dashboardToResponse(&d)
	}
	return resps, nil
}

// Get returns a single dashboard with widgets.
func (s *DashboardService) Get(id, projectID, userID uint64) (*response.DashboardResponse, error) {
	var d model.SavedDashboard
	if err := s.db.Where("id = ? AND project_id = ? AND (owner_id = ? OR is_shared = ?)", id, projectID, userID, true).
		Preload("Widgets", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.DashboardNotFound()
		}
		return nil, common.Internal("Failed to fetch dashboard")
	}
	resp := dashboardToResponse(&d)
	return &resp, nil
}

// Create creates a new dashboard with optional widgets.
func (s *DashboardService) Create(req *request.CreateDashboardRequest, projectID, userID uint64) (*response.DashboardResponse, error) {
	d := &model.SavedDashboard{
		Name:        req.Name,
		Description: req.Description,
		IsShared:    req.IsShared,
		OwnerID:     userID,
		ProjectID:   projectID,
		DateFrom:    req.DateFrom,
		DateTo:      req.DateTo,
		Columns:     req.Columns,
	}
	if d.Columns == 0 {
		d.Columns = 12
	}

	// If set as default, unset previous defaults for this user+project
	if req.IsDefault {
		s.db.Model(&model.SavedDashboard{}).
			Where("project_id = ? AND owner_id = ? AND is_default = ?", projectID, userID, true).
			Update("is_default", false)
		d.IsDefault = true
	}

	if err := s.db.Create(d).Error; err != nil {
		return nil, common.Internal("Failed to create dashboard")
	}

	// Create widgets if provided
	if len(req.Widgets) > 0 {
		for i, w := range req.Widgets {
			widget := &model.DashboardWidget{
				DashboardID: d.ID,
				WidgetType:  w.WidgetType,
				Title:       w.Title,
				Description: w.Description,
				Config:      normalizeJSON(w.Config),
				Position:    normalizeJSON(w.Position),
				SortOrder:   w.SortOrder,
			}
			if widget.SortOrder == 0 {
				widget.SortOrder = i
			}
			if err := s.db.Create(widget).Error; err != nil {
				return nil, common.Internal("Failed to create widget")
			}
		}
		// Reload with widgets
		s.db.Preload("Widgets", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).First(d, d.ID)
	}

	resp := dashboardToResponse(d)
	return &resp, nil
}

// Update updates a dashboard's metadata.
func (s *DashboardService) Update(id, projectID, userID uint64, req *request.UpdateDashboardRequest) (*response.DashboardResponse, error) {
	var d model.SavedDashboard
	if err := s.db.Where("id = ? AND project_id = ? AND owner_id = ?", id, projectID, userID).First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.DashboardNotFound()
		}
		return nil, common.Internal("Failed to fetch dashboard")
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = req.Description
	}
	if req.IsShared != nil {
		updates["is_shared"] = *req.IsShared
	}
	if req.IsDefault != nil && *req.IsDefault {
		// Unset previous defaults for this user+project
		s.db.Model(&model.SavedDashboard{}).
			Where("project_id = ? AND owner_id = ? AND is_default = ?", projectID, userID, true).
			Update("is_default", false)
		updates["is_default"] = true
	}
	if req.DateFrom != nil {
		updates["date_from"] = req.DateFrom
	}
	if req.DateTo != nil {
		updates["date_to"] = req.DateTo
	}
	if req.Columns != nil {
		updates["columns"] = *req.Columns
	}

	if len(updates) > 0 {
		if err := s.db.Model(&d).Updates(updates).Error; err != nil {
			return nil, common.Internal("Failed to update dashboard")
		}
		s.db.First(&d, d.ID)
	}

	// Reload with widgets
	s.db.Preload("Widgets", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).First(&d, d.ID)

	resp := dashboardToResponse(&d)
	return &resp, nil
}

// Delete soft-deletes a dashboard and its widgets.
func (s *DashboardService) Delete(id, projectID, userID uint64) error {
	result := s.db.Where("id = ? AND project_id = ? AND owner_id = ?", id, projectID, userID).Delete(&model.SavedDashboard{})
	if result.Error != nil {
		return common.Internal("Failed to delete dashboard")
	}
	if result.RowsAffected == 0 {
		return common.DashboardNotFound()
	}
	// Soft-delete all associated widgets
	s.db.Where("dashboard_id = ?", id).Delete(&model.DashboardWidget{})
	return nil
}

// SetDefault sets a dashboard as the default for the user in this project.
func (s *DashboardService) SetDefault(id, projectID, userID uint64) (*response.DashboardResponse, error) {
	var d model.SavedDashboard
	if err := s.db.Where("id = ? AND project_id = ? AND owner_id = ?", id, projectID, userID).First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.DashboardNotFound()
		}
		return nil, common.Internal("Failed to fetch dashboard")
	}

	// Unset previous defaults
	s.db.Model(&model.SavedDashboard{}).
		Where("project_id = ? AND owner_id = ? AND is_default = ?", projectID, userID, true).
		Update("is_default", false)

	// Set new default
	s.db.Model(&d).Update("is_default", true)
	d.IsDefault = true
	resp := dashboardToResponse(&d)
	return &resp, nil
}

// Duplicate duplicates a dashboard with all its widgets.
func (s *DashboardService) Duplicate(id, projectID, userID uint64) (*response.DashboardResponse, error) {
	var src model.SavedDashboard
	if err := s.db.Where("id = ? AND project_id = ? AND (owner_id = ? OR is_shared = ?)", id, projectID, userID, true).
		Preload("Widgets", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		First(&src).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.DashboardNotFound()
		}
		return nil, common.Internal("Failed to fetch dashboard")
	}

	clone := model.SavedDashboard{
		Name:        src.Name + " (Copy)",
		Description: src.Description,
		IsDefault:   false,
		IsShared:    false,
		OwnerID:     userID,
		ProjectID:   projectID,
		DateFrom:    src.DateFrom,
		DateTo:      src.DateTo,
		Columns:     src.Columns,
	}
	if err := s.db.Create(&clone).Error; err != nil {
		return nil, common.Internal("Failed to duplicate dashboard")
	}

	// Duplicate widgets
	for _, w := range src.Widgets {
		newWidget := model.DashboardWidget{
			DashboardID: clone.ID,
			WidgetType:  w.WidgetType,
			Title:       w.Title,
			Description: w.Description,
			Config:      w.Config,
			Position:    w.Position,
			SortOrder:   w.SortOrder,
		}
		s.db.Create(&newWidget)
	}

	// Reload
	s.db.Preload("Widgets", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).First(&clone, clone.ID)

	resp := dashboardToResponse(&clone)
	return &resp, nil
}

// ==================== Widget CRUD ====================

// AddWidget adds a widget to an existing dashboard.
func (s *DashboardService) AddWidget(dashboardID, projectID, userID uint64, req *request.CreateWidgetRequest) (*response.WidgetResponse, error) {
	// Verify dashboard ownership
	var d model.SavedDashboard
	if err := s.db.Where("id = ? AND project_id = ? AND owner_id = ?", dashboardID, projectID, userID).First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.DashboardNotFound()
		}
		return nil, common.Internal("Failed to fetch dashboard")
	}

	widget := &model.DashboardWidget{
		DashboardID: dashboardID,
		WidgetType:  req.WidgetType,
		Title:       req.Title,
		Description: req.Description,
		Config:      normalizeJSON(req.Config),
		Position:    normalizeJSON(req.Position),
		SortOrder:   req.SortOrder,
	}
	if err := s.db.Create(widget).Error; err != nil {
		return nil, common.Internal("Failed to add widget")
	}

	resp := widgetToResponse(widget)
	return &resp, nil
}

// UpdateWidget updates a widget's config/position/title.
func (s *DashboardService) UpdateWidget(dashboardID, widgetID, projectID, userID uint64, req *request.UpdateWidgetRequest) (*response.WidgetResponse, error) {
	// Verify dashboard ownership
	var d model.SavedDashboard
	if err := s.db.Where("id = ? AND project_id = ? AND owner_id = ?", dashboardID, projectID, userID).First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.DashboardNotFound()
		}
		return nil, common.Internal("Failed to fetch dashboard")
	}

	var w model.DashboardWidget
	if err := s.db.Where("id = ? AND dashboard_id = ?", widgetID, dashboardID).First(&w).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Widget not found")
		}
		return nil, common.Internal("Failed to fetch widget")
	}

	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = req.Description
	}
	if req.Config != nil {
		updates["config"] = normalizeJSON(*req.Config)
	}
	if req.Position != nil {
		updates["position"] = normalizeJSON(*req.Position)
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}

	if len(updates) > 0 {
		if err := s.db.Model(&w).Updates(updates).Error; err != nil {
			return nil, common.Internal("Failed to update widget")
		}
		s.db.First(&w, w.ID)
	}

	resp := widgetToResponse(&w)
	return &resp, nil
}

// DeleteWidget removes a widget from a dashboard.
func (s *DashboardService) DeleteWidget(dashboardID, widgetID, projectID, userID uint64) error {
	// Verify dashboard ownership
	var d model.SavedDashboard
	if err := s.db.Where("id = ? AND project_id = ? AND owner_id = ?", dashboardID, projectID, userID).First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.DashboardNotFound()
		}
		return common.Internal("Failed to fetch dashboard")
	}

	result := s.db.Where("id = ? AND dashboard_id = ?", widgetID, dashboardID).Delete(&model.DashboardWidget{})
	if result.Error != nil {
		return common.Internal("Failed to delete widget")
	}
	if result.RowsAffected == 0 {
		return common.NotFound("Widget not found")
	}
	return nil
}

// ReorderWidgets bulk-updates sort_order for widgets.
func (s *DashboardService) ReorderWidgets(dashboardID, projectID, userID uint64, req *request.ReorderWidgetsRequest) error {
	// Verify dashboard ownership
	var d model.SavedDashboard
	if err := s.db.Where("id = ? AND project_id = ? AND owner_id = ?", dashboardID, projectID, userID).First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.DashboardNotFound()
		}
		return common.Internal("Failed to fetch dashboard")
	}

	for i, widgetID := range req.WidgetIDs {
		s.db.Model(&model.DashboardWidget{}).Where("id = ? AND dashboard_id = ?", widgetID, dashboardID).
			Update("sort_order", i)
	}
	return nil
}

// ==================== Full Data ====================

// GetFull returns dashboard metadata plus rendered widget data.
func (s *DashboardService) GetFull(id, projectID, userID uint64) (*response.DashboardFullResponse, error) {
	var d model.SavedDashboard
	if err := s.db.Where("id = ? AND project_id = ? AND (owner_id = ? OR is_shared = ?)", id, projectID, userID, true).
		Preload("Widgets", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.DashboardNotFound()
		}
		return nil, common.Internal("Failed to fetch dashboard")
	}

	// Build widget data in parallel (simple loop for now; can optimize with goroutines later)
	var widgetData []response.WidgetDataResponse
	for _, w := range d.Widgets {
		data, err := s.renderWidget(projectID, &w, &d)
		if err != nil {
			data = json.RawMessage(`{"error":"` + err.Error() + `"}`)
		}
		widgetData = append(widgetData, response.WidgetDataResponse{
			WidgetID: w.ID,
			Data:     data,
		})
	}

	fullResp := &response.DashboardFullResponse{
		Dashboard:  dashboardToResponse(&d),
		WidgetData: widgetData,
	}
	return fullResp, nil
}

// renderWidget renders a single widget's data based on its type.
func (s *DashboardService) renderWidget(projectID uint64, w *model.DashboardWidget, d *model.SavedDashboard) (json.RawMessage, error) {
	switch w.WidgetType {
	case "number_card":
		return s.renderNumberCard(projectID, w)
	case "bar_chart", "pie_chart", "doughnut_chart", "line_chart", "table":
		return s.renderChart(projectID, w, d)
	case "saved_report":
		return s.renderSavedReportWidget(projectID, w, d)
	case "burndown":
		return s.renderBurndown(w)
	case "recent_list":
		return s.renderRecentList(projectID, w)
	default:
		return json.RawMessage("{}"), nil
	}
}

func (s *DashboardService) renderNumberCard(projectID uint64, w *model.DashboardWidget) (json.RawMessage, error) {
	// Extract metric from config
	config := struct {
		Metric string `json:"metric"`
		Label  string `json:"label"`
	}{}
	if err := json.Unmarshal(w.Config, &config); err != nil {
		return nil, err
	}

	var value int
	var err error

	switch config.Metric {
	case "total":
		var cnt int64
		s.db.Model(&model.Issue{}).Where("project_id = ? AND archived_at IS NULL", projectID).Count(&cnt)
		value = int(cnt)
	case "completed":
		type stateRow struct{ ID uint64 }
		var doneStates []stateRow
		s.db.Model(&model.State{}).Where("group = ?", "completed").Find(&doneStates)
		var cnt int64
		if len(doneStates) > 0 {
			ids := make([]uint64, len(doneStates))
			for i, s := range doneStates {
				ids[i] = s.ID
			}
			s.db.Model(&model.Issue{}).Where("project_id = ? AND archived_at IS NULL AND state_id IN ?", projectID, ids).Count(&cnt)
		}
		value = int(cnt)
	case "in_progress":
		type stateRow struct{ ID uint64 }
		var progStates []stateRow
		s.db.Model(&model.State{}).Where("group = ?", "started").Find(&progStates)
		var cnt int64
		if len(progStates) > 0 {
			ids := make([]uint64, len(progStates))
			for i, s := range progStates {
				ids[i] = s.ID
			}
			s.db.Model(&model.Issue{}).Where("project_id = ? AND archived_at IS NULL AND state_id IN ?", projectID, ids).Count(&cnt)
		}
		value = int(cnt)
	case "overdue":
		var cnt int64
		s.db.Model(&model.Issue{}).
			Joins("JOIN states ON states.id = issues.state_id").
			Where("issues.project_id = ? AND issues.archived_at IS NULL AND issues.target_date < NOW() AND states.group NOT IN ?",
				projectID, []string{common.StateGroupCompleted, common.StateGroupCancelled}).
			Count(&cnt)
		value = int(cnt)
	default:
		value = 0
	}
	_ = err

	result := map[string]interface{}{
		"metric": config.Metric,
		"label":  config.Label,
		"value":  value,
	}
	data, _ := json.Marshal(result)
	return json.RawMessage(data), nil
}

func (s *DashboardService) renderChart(projectID uint64, w *model.DashboardWidget, d *model.SavedDashboard) (json.RawMessage, error) {
	config := struct {
		ReportType string `json:"report_type"`
		GroupBy    string `json:"group_by"`
		ChartType  string `json:"chart_type"`
		RQL        string `json:"rql"`
		Interval   string `json:"interval"`
		DateFrom   string `json:"date_from"`
		DateTo     string `json:"date_to"`
	}{}
	if err := json.Unmarshal(w.Config, &config); err != nil {
		return nil, err
	}

	// Override date range with dashboard global filter if set
	dateFrom := config.DateFrom
	dateTo := config.DateTo
	if d.DateFrom != nil && *d.DateFrom != "" {
		dateFrom = *d.DateFrom
	}
	if d.DateTo != nil && *d.DateTo != "" {
		dateTo = *d.DateTo
	}

	reportReq := &ReportRequest{
		RQL:        config.RQL,
		ReportType: config.ReportType,
		GroupBy:    config.GroupBy,
		Chart:      config.ChartType,
		Interval:   config.Interval,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
	}
	if reportReq.ReportType == "" {
		reportReq.ReportType = "distribution"
	}
	if reportReq.GroupBy == "" {
		reportReq.GroupBy = "state"
	}
	if reportReq.Chart == "" {
		reportReq.Chart = w.WidgetType
		// Map widget_type to chart type
		switch w.WidgetType {
		case "bar_chart":
			reportReq.Chart = "bar"
		case "pie_chart":
			reportReq.Chart = "pie"
		case "doughnut_chart":
			reportReq.Chart = "doughnut"
		case "line_chart":
			reportReq.Chart = "line"
		case "table":
			reportReq.Chart = "table"
		default:
			reportReq.Chart = "bar"
		}
	}

	resp, err := s.reportSvc.Generate(projectID, reportReq)
	if err != nil {
		return nil, err
	}

	// Attach chart type to response
	type enrichedReport struct {
		ChartType string `json:"chart_type"`
		*ReportResponse
	}
	data, _ := json.Marshal(enrichedReport{ChartType: reportReq.Chart, ReportResponse: resp})
	return json.RawMessage(data), nil
}

func (s *DashboardService) renderBurndown(w *model.DashboardWidget) (json.RawMessage, error) {
	config := struct {
		CycleID uint64 `json:"cycle_id"`
	}{}
	if err := json.Unmarshal(w.Config, &config); err != nil {
		return nil, err
	}
	if config.CycleID == 0 {
		return json.RawMessage(`{"error":"No cycle selected"}`), nil
	}

	burndown, err := s.cycleSvc.GetBurndown(config.CycleID)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(burndown)
	return json.RawMessage(data), nil
}

func (s *DashboardService) renderRecentList(projectID uint64, w *model.DashboardWidget) (json.RawMessage, error) {
	config := struct {
		Limit int `json:"limit"`
	}{Limit: 10}
	if err := json.Unmarshal(w.Config, &config); err != nil {
		return nil, err
	}
	if config.Limit == 0 {
		config.Limit = 10
	}

	var issues []model.Issue
	if err := s.db.Where("project_id = ? AND archived_at IS NULL", projectID).
		Preload("State").Preload("IssueType").
		Order("updated_at DESC").
		Limit(config.Limit).
		Find(&issues).Error; err != nil {
		return nil, err
	}

	type recentIssue struct {
		ID         uint64 `json:"id"`
		SequenceID int    `json:"sequence_id"`
		Name       string `json:"name"`
		StateName  string `json:"state_name"`
		StateColor string `json:"state_color"`
		TypeName   string `json:"type_name"`
		UpdatedAt  string `json:"updated_at"`
	}
	items := make([]recentIssue, len(issues))
	for i, issue := range issues {
		stateName := issue.State.Name
		stateColor := issue.State.Color
		typeName := ""
		if issue.IssueTypeID != nil {
			typeName = issue.IssueType.Name
		}
		items[i] = recentIssue{
			ID:         issue.ID,
			SequenceID: issue.SequenceID,
			Name:       issue.Name,
			StateName:  stateName,
			StateColor: stateColor,
			TypeName:   typeName,
			UpdatedAt:  issue.UpdatedAt.Format("2006-01-02 15:04"),
		}
	}

	data, _ := json.Marshal(items)
	return json.RawMessage(data), nil
}

// ==================== Saved Report Widget ====================

func (s *DashboardService) renderSavedReportWidget(projectID uint64, w *model.DashboardWidget, d *model.SavedDashboard) (json.RawMessage, error) {
	config := struct {
		SavedReportID uint64 `json:"saved_report_id"`
	}{}
	if err := json.Unmarshal(w.Config, &config); err != nil {
		return nil, err
	}

	if config.SavedReportID == 0 {
		return json.RawMessage(`{"error":"No saved report selected"}`), nil
	}

	// Load the saved report
	savedReport, err := s.savedReportSvc.Get(config.SavedReportID, projectID)
	if err != nil {
		return nil, err
	}

	// Override date range with dashboard global filter if set
	dateFrom := savedReport.DateFrom
	dateTo := savedReport.DateTo
	if d.DateFrom != nil && *d.DateFrom != "" {
		dateFrom = *d.DateFrom
	}
	if d.DateTo != nil && *d.DateTo != "" {
		dateTo = *d.DateTo
	}

	// Build report request from saved report
	reportReq := &ReportRequest{
		RQL:        savedReport.RQL,
		ReportType: savedReport.ReportType,
		GroupBy:    savedReport.GroupBy,
		Chart:      savedReport.ChartType,
		Interval:   savedReport.Interval,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
	}
	if reportReq.ReportType == "" {
		reportReq.ReportType = "distribution"
	}
	if reportReq.Chart == "" {
		reportReq.Chart = "bar"
	}

	resp, err := s.reportSvc.Generate(projectID, reportReq)
	if err != nil {
		return nil, err
	}

	// Enrich with metadata
	type enrichedSavedReport struct {
		ChartType      string `json:"chart_type"`
		SavedReportID  uint64 `json:"saved_report_id"`
		SavedReportName string `json:"saved_report_name"`
		*ReportResponse
	}
	data, _ := json.Marshal(enrichedSavedReport{
		ChartType:       reportReq.Chart,
		SavedReportID:   config.SavedReportID,
		SavedReportName: savedReport.Name,
		ReportResponse:  resp,
	})
	return json.RawMessage(data), nil
}

// ==================== Helpers ====================

func dashboardToResponse(d *model.SavedDashboard) response.DashboardResponse {
	resp := response.DashboardResponse{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		IsDefault:   d.IsDefault,
		IsShared:    d.IsShared,
		OwnerID:     d.OwnerID,
		ProjectID:   d.ProjectID,
		DateFrom:    d.DateFrom,
		DateTo:      d.DateTo,
		Columns:     d.Columns,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
	if d.Widgets != nil {
		resp.Widgets = make([]response.WidgetResponse, len(d.Widgets))
		for i, w := range d.Widgets {
			resp.Widgets[i] = widgetToResponse(&w)
		}
	}
	return resp
}

func widgetToResponse(w *model.DashboardWidget) response.WidgetResponse {
	return response.WidgetResponse{
		ID:          w.ID,
		DashboardID: w.DashboardID,
		WidgetType:  w.WidgetType,
		Title:       w.Title,
		Description: w.Description,
		Config:      w.Config,
		Position:    w.Position,
		SortOrder:   w.SortOrder,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
}
