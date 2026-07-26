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

// ======== Health Check & Auto Scheduling ========

// HealthCheckInterval defines the interval for health checks (default 30 seconds).
const HealthCheckInterval = 30 * time.Second

// HeartbeatTimeout defines the timeout for considering a runtime offline (default 60 seconds).
const HeartbeatTimeout = 60 * time.Second

// HeartbeatWarningThreshold defines the threshold for marking runtime as recently_lost (default 30 seconds).
const HeartbeatWarningThreshold = 30 * time.Second

// PerformHealthCheck performs health check for all runtimes in a workspace.
// Updates status based on last heartbeat time with three-level health status:
// - online: heartbeat within warning threshold
// - recently_lost: heartbeat between warning threshold and timeout
// - offline: heartbeat beyond timeout
// Also supports status recovery: recently_lost/offline -> online when heartbeat resumes
func (s *RuntimeService) PerformHealthCheck(workspaceID uint64) error {
	now := time.Now()
	timeoutThreshold := now.Add(-HeartbeatTimeout)
	warningThreshold := now.Add(-HeartbeatWarningThreshold)

	// Step 1: Mark runtimes with heartbeat beyond timeout as offline
	if err := s.db.Model(&model.Runtime{}).
		Where("workspace_id = ? AND last_heartbeat IS NOT NULL AND last_heartbeat < ?",
			workspaceID, timeoutThreshold).
		Updates(map[string]interface{}{
			"status": "offline",
			"health": model.RuntimeHealthOffline,
		}).Error; err != nil {
		return common.Internal("Failed to update offline runtimes")
	}

	// Step 2: Mark runtimes with heartbeat between warning and timeout as recently_lost
	if err := s.db.Model(&model.Runtime{}).
		Where("workspace_id = ? AND last_heartbeat IS NOT NULL AND last_heartbeat >= ? AND last_heartbeat < ?",
			workspaceID, warningThreshold, timeoutThreshold).
		Updates(map[string]interface{}{
			"status": "online",
			"health": model.RuntimeHealthRecentlyLost,
		}).Error; err != nil {
		return common.Internal("Failed to update recently_lost runtimes")
	}

	// Step 3: Mark runtimes with recent heartbeat as online health
	if err := s.db.Model(&model.Runtime{}).
		Where("workspace_id = ? AND (last_heartbeat IS NULL OR last_heartbeat >= ?)",
			workspaceID, warningThreshold).
		Updates(map[string]interface{}{
			"status": "online",
			"health": model.RuntimeHealthOnline,
		}).Error; err != nil {
		return common.Internal("Failed to update online runtime health status")
	}

	return nil
}

// PerformGlobalHealthCheck performs health check for all runtimes across all workspaces.
func (s *RuntimeService) PerformGlobalHealthCheck() error {
	now := time.Now()
	timeoutThreshold := now.Add(-HeartbeatTimeout)
	warningThreshold := now.Add(-HeartbeatWarningThreshold)

	// Step 1: Mark runtimes with heartbeat beyond timeout as offline
	if err := s.db.Model(&model.Runtime{}).
		Where("last_heartbeat IS NOT NULL AND last_heartbeat < ?",
			timeoutThreshold).
		Updates(map[string]interface{}{
			"status": "offline",
			"health": model.RuntimeHealthOffline,
		}).Error; err != nil {
		return common.Internal("Failed to perform global health check")
	}

	// Step 2: Mark runtimes with heartbeat between warning and timeout as recently_lost
	if err := s.db.Model(&model.Runtime{}).
		Where("last_heartbeat IS NOT NULL AND last_heartbeat >= ? AND last_heartbeat < ?",
			warningThreshold, timeoutThreshold).
		Updates(map[string]interface{}{
			"status": "online",
			"health": model.RuntimeHealthRecentlyLost,
		}).Error; err != nil {
		return common.Internal("Failed to update global recently_lost runtimes")
	}

	// Step 3: Mark runtimes with recent heartbeat as online health
	if err := s.db.Model(&model.Runtime{}).
		Where("last_heartbeat IS NULL OR last_heartbeat >= ?",
			warningThreshold).
		Updates(map[string]interface{}{
			"status": "online",
			"health": model.RuntimeHealthOnline,
		}).Error; err != nil {
		return common.Internal("Failed to update global online runtime health status")
	}

	return nil
}

// ScheduleTask finds the best available runtime for a task and increments its load.
// Returns the runtime ID or error if no available runtime.
// Uses atomic update to prevent concurrent over-allocation.
func (s *RuntimeService) ScheduleTask(workspaceID uint64) (uint64, error) {
	var runtime model.Runtime

	// Find runtime with lowest current load among available runtimes
	if err := s.db.Where("workspace_id = ? AND status = ? AND current_load < capacity",
		workspaceID, "online").
		Order("current_load ASC, created_at ASC").
		First(&runtime).Error; err != nil {
		return 0, common.NotFound("No available runtime for task scheduling")
	}

	// Use atomic update to prevent concurrent over-allocation
	result := s.db.Model(&model.Runtime{}).
		Where("id = ? AND current_load < capacity", runtime.ID).
		Update("current_load", gorm.Expr("current_load + ?", 1))
	if result.Error != nil {
		return 0, common.Internal("Failed to update runtime load")
	}
	if result.RowsAffected == 0 {
		return 0, common.NotFound("Runtime became unavailable during scheduling")
	}

	// Refresh the runtime to get updated current_load
	if err := s.db.First(&runtime, runtime.ID).Error; err != nil {
		return 0, common.Internal("Failed to refresh runtime")
	}

	// Mark as busy if load reaches capacity
	if runtime.CurrentLoad >= runtime.Capacity {
		if err := s.db.Model(&model.Runtime{}).
			Where("id = ?", runtime.ID).
			Update("status", "busy").Error; err != nil {
			return 0, common.Internal("Failed to mark runtime as busy")
		}
	}

	return runtime.ID, nil
}

// ReleaseTask decrements the load on a runtime when a task completes.
func (s *RuntimeService) ReleaseTask(runtimeID uint64) error {
	var runtime model.Runtime
	if err := s.db.First(&runtime, runtimeID).Error; err != nil {
		return common.NotFound("Runtime not found")
	}

	if runtime.CurrentLoad > 0 {
		// Use atomic update to prevent negative load
		result := s.db.Model(&model.Runtime{}).
			Where("id = ? AND current_load > 0", runtime.ID).
			Update("current_load", gorm.Expr("current_load - ?", 1))
		if result.Error != nil {
			return common.Internal("Failed to release runtime load")
		}
		if result.RowsAffected == 0 {
			return common.Internal("Runtime load was already 0")
		}

		// Refresh the runtime to get updated current_load
		if err := s.db.First(&runtime, runtimeID).Error; err != nil {
			return common.Internal("Failed to refresh runtime")
		}

		// Restore to online if load drops to 0
		if runtime.CurrentLoad == 0 {
			if err := s.db.Model(&model.Runtime{}).
				Where("id = ?", runtime.ID).
				Update("status", "online").Error; err != nil {
				return common.Internal("Failed to restore runtime status")
			}
		}
	}

	return nil
}

// GetRuntimeStats returns runtime statistics for a workspace.
func (s *RuntimeService) GetRuntimeStats(workspaceID uint64) (*response.RuntimeStatsResponse, error) {
	var stats response.RuntimeStatsResponse

	// Total runtimes
	s.db.Model(&model.Runtime{}).Where("workspace_id = ?", workspaceID).Count(&stats.Total)

	// By status
	s.db.Model(&model.Runtime{}).Where("workspace_id = ? AND status = ?", workspaceID, "online").Count(&stats.Online)
	s.db.Model(&model.Runtime{}).Where("workspace_id = ? AND status = ?", workspaceID, "offline").Count(&stats.Offline)
	s.db.Model(&model.Runtime{}).Where("workspace_id = ? AND status = ?", workspaceID, "busy").Count(&stats.Busy)

	// Total capacity and current load
	var capacities []model.Runtime
	s.db.Where("workspace_id = ?", workspaceID).Find(&capacities)
	for _, r := range capacities {
		stats.TotalCapacity += r.Capacity
		stats.TotalCurrentLoad += r.CurrentLoad
	}

	// Calculate available capacity
	stats.AvailableCapacity = stats.TotalCapacity - stats.TotalCurrentLoad

	return &stats, nil
}
