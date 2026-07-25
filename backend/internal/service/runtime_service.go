package service

import (
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type RuntimeService struct{ db *gorm.DB }

func NewRuntimeService(db *gorm.DB) *RuntimeService {
	return &RuntimeService{db: db}
}

func (s *RuntimeService) Create(wid uint64, req request.RuntimeCreate) (*response.RuntimeResponse, error) {
	runtime := model.Runtime{
		Name:        req.Name,
		RuntimeType: req.RuntimeType,
		RuntimeMode: req.RuntimeMode,
		Endpoint:    req.Endpoint,
		Status:      "offline",
		Capacity:    req.Capacity,
		CurrentLoad: 0,
		Metadata:    req.Metadata,
		WorkspaceID: wid,
	}

	if err := s.db.Create(&runtime).Error; err != nil {
		return nil, common.Internal("Failed to create runtime")
	}

	return s.Get(runtime.ID)
}

func (s *RuntimeService) Get(id uint64) (*response.RuntimeResponse, error) {
	var runtime model.Runtime
	if err := s.db.First(&runtime, id).Error; err != nil {
		return nil, common.NotFound("Runtime not found")
	}

	return s.toResponse(&runtime), nil
}

func (s *RuntimeService) List(wid uint64) ([]response.RuntimeResponse, error) {
	var runtimes []model.Runtime
	s.db.Where("workspace_id = ?", wid).Find(&runtimes)

	res := make([]response.RuntimeResponse, 0, len(runtimes))
	for _, r := range runtimes {
		res = append(res, *s.toResponse(&r))
	}

	return res, nil
}

func (s *RuntimeService) Update(id uint64, req request.RuntimeUpdate) (*response.RuntimeResponse, error) {
	var runtime model.Runtime
	if err := s.db.First(&runtime, id).Error; err != nil {
		return nil, common.NotFound("Runtime not found")
	}

	if req.Name != nil {
		runtime.Name = *req.Name
	}
	if req.RuntimeType != nil {
		runtime.RuntimeType = *req.RuntimeType
	}
	if req.RuntimeMode != nil {
		runtime.RuntimeMode = *req.RuntimeMode
	}
	if req.Endpoint != nil {
		runtime.Endpoint = req.Endpoint
	}
	if req.Capacity != nil {
		runtime.Capacity = *req.Capacity
	}
	if req.Metadata != nil {
		runtime.Metadata = *req.Metadata
	}

	if err := s.db.Save(&runtime).Error; err != nil {
		return nil, common.Internal("Failed to update runtime")
	}

	return s.Get(id)
}

func (s *RuntimeService) Delete(id uint64) error {
	var runtime model.Runtime
	if err := s.db.First(&runtime, id).Error; err != nil {
		return common.NotFound("Runtime not found")
	}

	return s.db.Delete(&runtime).Error
}

func (s *RuntimeService) Heartbeat(id uint64, req request.RuntimeHeartbeat) (*response.RuntimeResponse, error) {
	var runtime model.Runtime
	if err := s.db.First(&runtime, id).Error; err != nil {
		return nil, common.NotFound("Runtime not found")
	}

	now := time.Now()
	runtime.Status = "online"
	runtime.Version = req.Version
	runtime.HostInfo = req.HostInfo
	runtime.CurrentLoad = req.CurrentLoad
	runtime.LastHeartbeat = &now

	if err := s.db.Save(&runtime).Error; err != nil {
		return nil, common.Internal("Failed to update heartbeat")
	}

	return s.Get(id)
}

func (s *RuntimeService) Register(wid uint64, req request.RuntimeCreate) (*response.RuntimeResponse, error) {
	var existing model.Runtime
	if err := s.db.Where("name = ? AND workspace_id = ?", req.Name, wid).First(&existing).Error; err == nil {
		return s.Heartbeat(existing.ID, request.RuntimeHeartbeat{
			Version: req.Name,
		})
	}

	return s.Create(wid, req)
}

func (s *RuntimeService) FindAvailable(wid uint64) (*response.RuntimeResponse, error) {
	var runtime model.Runtime
	if err := s.db.Where("workspace_id = ? AND status = ? AND current_load < capacity", wid, "online").
		Order("current_load ASC").First(&runtime).Error; err != nil {
		return nil, common.NotFound("No available runtime")
	}

	return s.toResponse(&runtime), nil
}

func (s *RuntimeService) toResponse(r *model.Runtime) *response.RuntimeResponse {
	return &response.RuntimeResponse{
		ID:            r.ID,
		Name:          r.Name,
		RuntimeType:   r.RuntimeType,
		RuntimeMode:   r.RuntimeMode,
		Endpoint:      r.Endpoint,
		Status:        r.Status,
		Capacity:      r.Capacity,
		CurrentLoad:   r.CurrentLoad,
		Version:       r.Version,
		HostInfo:      r.HostInfo,
		LastHeartbeat: r.LastHeartbeat,
		Metadata:      r.Metadata,
		WorkspaceID:   r.WorkspaceID,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}
