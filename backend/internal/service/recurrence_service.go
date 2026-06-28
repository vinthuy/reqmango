package service

import (
	"time"

	"github.com/reqmanpy/backend/internal/common"
	"github.com/reqmanpy/backend/internal/dto/request"
	"github.com/reqmanpy/backend/internal/dto/response"
	"github.com/reqmanpy/backend/internal/model"
	"gorm.io/gorm"
)

type RecurrenceService struct{ db *gorm.DB }

func NewRecurrenceService(db *gorm.DB) *RecurrenceService { return &RecurrenceService{db: db} }

func (s *RecurrenceService) Create(req *request.RecurrenceCreateRequest, issueID uint64) (*response.RecurrenceResponse, error) {
	// Clean up soft-deleted duplicates
	s.db.Unscoped().Where("issue_id = ? AND deleted_at IS NOT NULL", issueID).Delete(&model.RecurrenceRule{})

	var count int64
	s.db.Model(&model.RecurrenceRule{}).Where("issue_id = ?", issueID).Count(&count)
	if count > 0 { return nil, common.Conflict("Issue already has a recurrence rule") }

	nextRun, _ := time.Parse(time.RFC3339, req.NextRun)
	if nextRun.IsZero() { nextRun = s.computeNext(req.Frequency, req.Interval, time.Now()) }

	var endDate *time.Time
	if req.EndDate != nil { t, _ := time.Parse(time.RFC3339, *req.EndDate); endDate = &t }

	r := &model.RecurrenceRule{
		IssueID: issueID, Frequency: req.Frequency,
		Interval: req.Interval, CronExpr: req.CronExpr,
		NextRun: nextRun, EndDate: endDate, IsActive: true,
	}
	if r.Interval == 0 { r.Interval = 1 }
	if err := s.db.Create(r).Error; err != nil { return nil, common.Internal("Failed to create") }
	return toRecurrenceResp(r), nil
}

func (s *RecurrenceService) Get(issueID uint64) (*response.RecurrenceResponse, error) {
	var r model.RecurrenceRule
	if err := s.db.Where("issue_id = ?", issueID).First(&r).Error; err != nil {
		return nil, common.NotFound("No recurrence rule found")
	}
	return toRecurrenceResp(&r), nil
}

func (s *RecurrenceService) Update(issueID uint64, req *request.RecurrenceUpdateRequest) (*response.RecurrenceResponse, error) {
	var r model.RecurrenceRule
	if err := s.db.Where("issue_id = ?", issueID).First(&r).Error; err != nil {
		return nil, common.NotFound("No recurrence rule found")
	}
	u := map[string]interface{}{}
	if req.Frequency != nil { u["frequency"] = *req.Frequency }
	if req.Interval != nil { u["interval"] = *req.Interval }
	if req.IsActive != nil { u["is_active"] = *req.IsActive }
	if req.CronExpr != nil { u["cron_expr"] = req.CronExpr }
	if req.NextRun != nil {
		if t, err := time.Parse(time.RFC3339, *req.NextRun); err == nil { u["next_run"] = t }
	}
	if req.EndDate != nil {
		if t, err := time.Parse(time.RFC3339, *req.EndDate); err == nil { u["end_date"] = t }
	}
	s.db.Model(&r).Updates(u)
	s.db.First(&r, r.ID)
	return toRecurrenceResp(&r), nil
}

func (s *RecurrenceService) Delete(issueID uint64) error {
	r := s.db.Where("issue_id = ?", issueID).Delete(&model.RecurrenceRule{})
	if r.RowsAffected == 0 { return common.NotFound("No recurrence rule found") }
	return r.Error
}

func (s *RecurrenceService) computeNext(freq string, interval int, from time.Time) time.Time {
	if interval <= 0 { interval = 1 }
	switch freq {
	case "daily": return from.AddDate(0, 0, interval)
	case "weekly": return from.AddDate(0, 0, 7*interval)
	case "monthly": return from.AddDate(0, interval, 0)
	default: return from.AddDate(0, 0, 1)
	}
}

func toRecurrenceResp(r *model.RecurrenceRule) *response.RecurrenceResponse {
	return &response.RecurrenceResponse{
		ID: r.ID, IssueID: r.IssueID, Frequency: r.Frequency,
		Interval: r.Interval, CronExpr: r.CronExpr,
		NextRun: r.NextRun, EndDate: r.EndDate, IsActive: r.IsActive, CreatedAt: r.CreatedAt,
	}
}
