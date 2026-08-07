package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// CICDService implements PRD P4-004: CI/CD 配置、触发构建、查看状态 API.
//
// The service is split into two halves:
//
//   - Config CRUD: manage CICDConfig rows (workspace-scoped, optionally
//     project-scoped). Requires workspace admin.
//   - Build orchestration: TriggerBuild creates a BuildRecord and runs an
//     async workflow that simulates the provider pipeline. Real provider
//     calls are abstracted behind the CICDProvider interface so tests can
//     substitute a stub without external dependencies.
//
// Status updates are broadcast via SSE as cicd_build.* events so the UI can
// render live progress without polling.
type CICDService struct {
	db       *gorm.DB
	provider CICDProvider
}

// CICDProvider abstracts the external CI/CD platform interaction so the
// service can be unit-tested without real API calls.
type CICDProvider interface {
	// Trigger starts a build on the provider and returns the external
	// build id and a URL where logs/status can be viewed.
	Trigger(ctx context.Context, cfg *model.CICDConfig, req CICDTriggerRequest) (externalID, buildURL string, err error)

	// GetStatus polls the provider for the current status of a build.
	GetStatus(ctx context.Context, cfg *model.CICDConfig, externalID string) (CICDProviderStatus, error)

	// Cancel attempts to abort a running build on the provider.
	Cancel(ctx context.Context, cfg *model.CICDConfig, externalID string) error
}

// CICDTriggerRequest bundles the inputs handed to a provider Trigger call.
type CICDTriggerRequest struct {
	Branch    string
	CommitSHA string
	Trigger   model.BuildTrigger
}

// CICDProviderStatus is the normalized status returned by a provider poll.
type CICDProviderStatus struct {
	Status       model.BuildStatus
	Stage        string // current stage name, if running
	Progress     int    // 0-100
	Stages       []CICDStage
	ErrorMessage string
	BuildURL     string
}

// CICDStage describes one stage of a pipeline.
type CICDStage struct {
	Name        string         `json:"name"`
	Status      string         `json:"status"` // pending | running | success | failed | skipped
	DurationMs  int64          `json:"duration_ms"`
	StartedAt   *time.Time     `json:"started_at,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	LogURL      string         `json:"log_url,omitempty"`
}

// NewCICDService creates a new CICDService. When provider is nil a stub
// provider is used so the workflow remains end-to-end testable without
// external dependencies.
func NewCICDService(db *gorm.DB) *CICDService {
	svc := &CICDService{db: db}
	svc.provider = &stubCICDProvider{}
	return svc
}

// SetCICDProvider overrides the default provider. Used in tests.
func (s *CICDService) SetCICDProvider(p CICDProvider) {
	if p != nil {
		s.provider = p
	}
}

// checkWorkspaceAdmin mirrors the guard used by other workspace services.
func (s *CICDService) checkWorkspaceAdmin(workspaceID, callerID uint64) error {
	var member model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ? AND is_active = ?", workspaceID, callerID, true).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.Forbidden("You must be a workspace admin to manage CI/CD resources")
		}
		return common.Internal("Database error")
	}
	if member.Role < common.RoleAdmin {
		return common.Forbidden("You must be a workspace admin to manage CI/CD resources")
	}
	return nil
}

// ======== Config request / response types ========

// CICDConfigCreate captures the inputs for a new CICDConfig.
type CICDConfigCreate struct {
	Name          string          `json:"name" binding:"required"`
	Provider      string          `json:"provider"`
	APIEndpoint   string          `json:"api_endpoint"`
	ProjectSlug   string          `json:"project_slug"`
	DefaultBranch string          `json:"default_branch"`
	AuthTokenRef  string          `json:"auth_token_ref"`
	TriggerEvents []string        `json:"trigger_events"`
	ExtraConfig   json.RawMessage `json:"extra_config"`
	ProjectID     *uint64         `json:"project_id"`
	Enabled       *bool           `json:"enabled"`
}

// CICDConfigUpdate captures the partial-update inputs for a CICDConfig.
type CICDConfigUpdate struct {
	Name          *string         `json:"name"`
	Provider      *string         `json:"provider"`
	APIEndpoint   *string         `json:"api_endpoint"`
	ProjectSlug   *string         `json:"project_slug"`
	DefaultBranch *string         `json:"default_branch"`
	AuthTokenRef  *string         `json:"auth_token_ref"`
	TriggerEvents []string        `json:"trigger_events"`
	ExtraConfig   json.RawMessage `json:"extra_config"`
	Enabled       *bool           `json:"enabled"`
}

// CICDConfigResponse is the API representation of a CICDConfig.
type CICDConfigResponse struct {
	ID            uint64          `json:"id"`
	WorkspaceID   uint64          `json:"workspace_id"`
	ProjectID     *uint64         `json:"project_id,omitempty"`
	Name          string          `json:"name"`
	Provider      string          `json:"provider"`
	APIEndpoint   string          `json:"api_endpoint"`
	ProjectSlug   string          `json:"project_slug"`
	DefaultBranch string          `json:"default_branch"`
	AuthTokenRef  string          `json:"auth_token_ref"`
	TriggerEvents []string        `json:"trigger_events"`
	ExtraConfig   json.RawMessage `json:"extra_config"`
	Enabled       bool            `json:"enabled"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// ======== Build request / response types ========

// BuildTriggerRequest captures the inputs for triggering a new build.
type BuildTriggerRequest struct {
	CICDConfigID uint64          `json:"cicd_config_id" binding:"required"`
	Branch       string          `json:"branch"`
	CommitSHA    string          `json:"commit_sha"`
	Trigger      string          `json:"trigger"` // defaults to "manual"
	ProjectID    *uint64         `json:"project_id"`
	IssueID      *uint64         `json:"issue_id"`
	AgentTaskID  *uint64         `json:"agent_task_id"`
	Extra        json.RawMessage `json:"extra"`
}

// BuildRecordResponse is the API representation of a BuildRecord.
type BuildRecordResponse struct {
	ID             uint64          `json:"id"`
	WorkspaceID    uint64          `json:"workspace_id"`
	ProjectID      *uint64         `json:"project_id,omitempty"`
	CICDConfigID   uint64          `json:"cicd_config_id"`
	CICDConfigName string          `json:"cicd_config_name,omitempty"`
	Trigger        string          `json:"trigger"`
	Branch         string          `json:"branch"`
	CommitSHA      string          `json:"commit_sha"`
	IssueID        *uint64         `json:"issue_id,omitempty"`
	AgentTaskID    *uint64         `json:"agent_task_id,omitempty"`
	TriggeredByID  uint64          `json:"triggered_by_id"`
	ExternalBuildID string        `json:"external_build_id"`
	BuildURL       string          `json:"build_url"`
	Stages         json.RawMessage `json:"stages"`
	Status         string          `json:"status"`
	Progress       int             `json:"progress"`
	CurrentStage   *string         `json:"current_stage,omitempty"`
	ErrorMessage   *string         `json:"error_message,omitempty"`
	StartedAt      *time.Time      `json:"started_at,omitempty"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
	CancelledAt    *time.Time      `json:"cancelled_at,omitempty"`
	DurationMs     int64           `json:"duration_ms"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// ======== Config CRUD ========

// CreateConfig persists a new CICDConfig.
func (s *CICDService) CreateConfig(wid, callerID uint64, req CICDConfigCreate) (*CICDConfigResponse, error) {
	if err := s.checkWorkspaceAdmin(wid, callerID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, common.BadRequest("Name is required")
	}
	provider := model.CICDProvider(req.Provider)
	if provider == "" {
		provider = model.CICDProviderGeneric
	}
	if !isValidProvider(provider) {
		return nil, common.BadRequest("Unsupported provider: " + req.Provider)
	}
	defaultBranch := req.DefaultBranch
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	cfg := model.CICDConfig{
		WorkspaceID:  wid,
		ProjectID:    req.ProjectID,
		Name:         req.Name,
		Provider:     provider,
		APIEndpoint:  req.APIEndpoint,
		ProjectSlug:  req.ProjectSlug,
		DefaultBranch: defaultBranch,
		AuthTokenRef: req.AuthTokenRef,
		TriggerEvents: normalizeStringArray(req.TriggerEvents),
		ExtraConfig:   normalizeCICDJSON(req.ExtraConfig, "{}"),
		Enabled:       true,
	}
	if req.Enabled != nil {
		cfg.Enabled = *req.Enabled
	}
	if err := s.db.Create(&cfg).Error; err != nil {
		return nil, common.Internal("Failed to create CI/CD config")
	}
	resp := s.configToResponse(&cfg)
	s.pushEvent("cicd_config.created", resp)
	return resp, nil
}

// GetConfig returns a single CICDConfig by ID.
func (s *CICDService) GetConfig(id uint64) (*CICDConfigResponse, error) {
	var cfg model.CICDConfig
	if err := s.db.First(&cfg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("CI/CD config not found")
		}
		return nil, common.Internal("Failed to get CI/CD config")
	}
	return s.configToResponse(&cfg), nil
}

// ListConfigs returns CICDConfigs for a workspace.
// projectId filter is optional; when supplied only configs for that project
// (plus workspace-wide configs with no project_id) are returned.
func (s *CICDService) ListConfigs(wid uint64, projectID *uint64) ([]CICDConfigResponse, error) {
	q := s.db.Where("workspace_id = ?", wid)
	if projectID != nil && *projectID != 0 {
		q = q.Where("project_id IS NULL OR project_id = ?", *projectID)
	}
	var cfgs []model.CICDConfig
	if err := q.Order("created_at DESC").Find(&cfgs).Error; err != nil {
		return nil, common.Internal("Failed to list CI/CD configs")
	}
	out := make([]CICDConfigResponse, 0, len(cfgs))
	for i := range cfgs {
		out = append(out, *s.configToResponse(&cfgs[i]))
	}
	return out, nil
}

// UpdateConfig applies a partial update to a CICDConfig.
func (s *CICDService) UpdateConfig(id, callerID uint64, req CICDConfigUpdate) (*CICDConfigResponse, error) {
	var cfg model.CICDConfig
	if err := s.db.First(&cfg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("CI/CD config not found")
		}
		return nil, common.Internal("Failed to get CI/CD config")
	}
	if err := s.checkWorkspaceAdmin(cfg.WorkspaceID, callerID); err != nil {
		return nil, err
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return nil, common.BadRequest("Name cannot be empty")
		}
		cfg.Name = *req.Name
	}
	if req.Provider != nil {
		p := model.CICDProvider(*req.Provider)
		if !isValidProvider(p) {
			return nil, common.BadRequest("Unsupported provider: " + *req.Provider)
		}
		cfg.Provider = p
	}
	if req.APIEndpoint != nil {
		cfg.APIEndpoint = *req.APIEndpoint
	}
	if req.ProjectSlug != nil {
		cfg.ProjectSlug = *req.ProjectSlug
	}
	if req.DefaultBranch != nil {
		cfg.DefaultBranch = *req.DefaultBranch
	}
	if req.AuthTokenRef != nil {
		cfg.AuthTokenRef = *req.AuthTokenRef
	}
	if req.TriggerEvents != nil {
		cfg.TriggerEvents = normalizeStringArray(req.TriggerEvents)
	}
	if len(req.ExtraConfig) > 0 && string(req.ExtraConfig) != "null" {
		cfg.ExtraConfig = normalizeCICDJSON(req.ExtraConfig, "{}")
	}
	if req.Enabled != nil {
		cfg.Enabled = *req.Enabled
	}
	if err := s.db.Save(&cfg).Error; err != nil {
		return nil, common.Internal("Failed to update CI/CD config")
	}
	resp := s.configToResponse(&cfg)
	s.pushEvent("cicd_config.updated", resp)
	return resp, nil
}

// DeleteConfig removes a CICDConfig (soft-delete via BaseModel).
func (s *CICDService) DeleteConfig(id, callerID uint64) error {
	var cfg model.CICDConfig
	if err := s.db.First(&cfg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NotFound("CI/CD config not found")
		}
		return common.Internal("Failed to get CI/CD config")
	}
	if err := s.checkWorkspaceAdmin(cfg.WorkspaceID, callerID); err != nil {
		return err
	}
	return s.db.Delete(&cfg).Error
}

// ======== Build orchestration ========

// TriggerBuild starts a new build for the given config.
//
// The build record is persisted synchronously (status=pending) and the
// workflow runs asynchronously so the API can return immediately. Callers
// poll the build status or subscribe to SSE events (cicd_build.updated) for
// completion.
func (s *CICDService) TriggerBuild(wid, callerID uint64, req BuildTriggerRequest) (*BuildRecordResponse, error) {
	if err := s.checkWorkspaceAdmin(wid, callerID); err != nil {
		return nil, err
	}
	if req.CICDConfigID == 0 {
		return nil, common.BadRequest("cicd_config_id is required")
	}
	var cfg model.CICDConfig
	if err := s.db.First(&cfg, req.CICDConfigID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("CI/CD config not found")
		}
		return nil, common.Internal("Failed to get CI/CD config")
	}
	if cfg.WorkspaceID != wid {
		return nil, common.NotFound("CI/CD config not found")
	}
	if !cfg.Enabled {
		return nil, common.BadRequest("CI/CD config is disabled")
	}
	trigger := model.BuildTrigger(req.Trigger)
	if trigger == "" {
		trigger = model.BuildTriggerManual
	}
	if !isValidTrigger(trigger) {
		return nil, common.BadRequest("Unsupported trigger: " + req.Trigger)
	}
	branch := req.Branch
	if branch == "" {
		branch = cfg.DefaultBranch
		if branch == "" {
			branch = "main"
		}
	}
	build := model.BuildRecord{
		WorkspaceID:    wid,
		ProjectID:      req.ProjectID,
		CICDConfigID:   cfg.ID,
		Trigger:        trigger,
		Branch:         branch,
		CommitSHA:      req.CommitSHA,
		IssueID:        req.IssueID,
		AgentTaskID:    req.AgentTaskID,
		TriggeredByID:  callerID,
		Stages:         json.RawMessage("[]"),
		Status:         model.BuildPending,
	}
	if err := s.db.Create(&build).Error; err != nil {
		return nil, common.Internal("Failed to create build record")
	}
	go s.runBuildWorkflow(build.ID, &cfg, trigger)
	resp := s.buildToResponse(&build, &cfg)
	s.pushEvent("cicd_build.created", resp)
	return resp, nil
}

// GetBuild returns a single BuildRecord by ID.
func (s *CICDService) GetBuild(id uint64) (*BuildRecordResponse, error) {
	var build model.BuildRecord
	if err := s.db.First(&build, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Build not found")
		}
		return nil, common.Internal("Failed to get build")
	}
	var cfg model.CICDConfig
	_ = s.db.First(&cfg, build.CICDConfigID).Error
	return s.buildToResponse(&build, &cfg), nil
}

// ListBuilds returns BuildRecords for a workspace, newest first.
// Optional filters: configId, status, projectId.
func (s *CICDService) ListBuilds(wid uint64, configID *uint64, status string, projectID *uint64, limit int) ([]BuildRecordResponse, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := s.db.Where("workspace_id = ?", wid)
	if configID != nil && *configID != 0 {
		q = q.Where("cicd_config_id = ?", *configID)
	}
	if projectID != nil && *projectID != 0 {
		q = q.Where("project_id = ?", *projectID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var builds []model.BuildRecord
	if err := q.Order("created_at DESC").Limit(limit).Find(&builds).Error; err != nil {
		return nil, common.Internal("Failed to list builds")
	}
	// Preload config names in one query to avoid N+1.
	cfgNames := make(map[uint64]string)
	if len(builds) > 0 {
		ids := make([]uint64, 0, len(builds))
		for _, b := range builds {
			if _, ok := cfgNames[b.CICDConfigID]; !ok {
				ids = append(ids, b.CICDConfigID)
			}
		}
		var cfgs []model.CICDConfig
		if err := s.db.Where("id IN ?", ids).Find(&cfgs).Error; err == nil {
			for _, c := range cfgs {
				cfgNames[c.ID] = c.Name
			}
		}
	}
	out := make([]BuildRecordResponse, 0, len(builds))
	for i := range builds {
		r := s.buildToResponse(&builds[i], nil)
		if name, ok := cfgNames[builds[i].CICDConfigID]; ok {
			r.CICDConfigName = name
		}
		out = append(out, *r)
	}
	return out, nil
}

// CancelBuild marks a pending/running build as cancelled and best-effort
// cancels it on the provider side.
func (s *CICDService) CancelBuild(id, callerID uint64) (*BuildRecordResponse, error) {
	var build model.BuildRecord
	if err := s.db.First(&build, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Build not found")
		}
		return nil, common.Internal("Failed to get build")
	}
	if err := s.checkWorkspaceAdmin(build.WorkspaceID, callerID); err != nil {
		return nil, err
	}
	if build.Status == model.BuildSuccess || build.Status == model.BuildFailed {
		return nil, common.BadRequest("Cannot cancel a terminal build")
	}
	// Best-effort cancel on the provider side.
	if build.ExternalBuildID != "" {
		var cfg model.CICDConfig
		if err := s.db.First(&cfg, build.CICDConfigID).Error; err == nil {
			_ = s.provider.Cancel(context.Background(), &cfg, build.ExternalBuildID)
		}
	}
	now := time.Now()
	build.Status = model.BuildCancelled
	build.CancelledAt = &now
	if build.CompletedAt == nil {
		build.CompletedAt = &now
	}
	if err := s.db.Save(&build).Error; err != nil {
		return nil, common.Internal("Failed to cancel build")
	}
	var cfg model.CICDConfig
	_ = s.db.First(&cfg, build.CICDConfigID).Error
	resp := s.buildToResponse(&build, &cfg)
	s.pushEvent("cicd_build.cancelled", resp)
	return resp, nil
}

// DeleteBuild removes a BuildRecord (soft-delete via BaseModel).
func (s *CICDService) DeleteBuild(id, callerID uint64) error {
	var build model.BuildRecord
	if err := s.db.First(&build, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NotFound("Build not found")
		}
		return common.Internal("Failed to get build")
	}
	if err := s.checkWorkspaceAdmin(build.WorkspaceID, callerID); err != nil {
		return err
	}
	return s.db.Delete(&build).Error
}

// ======== Async workflow ========

// runBuildWorkflow simulates a CI/CD pipeline run.
//
// Steps:
//  1. queued   — invoke provider.Trigger to obtain external build id
//  2. running  — poll provider.GetStatus, advancing through stages
//  3. terminal — success | failed (based on provider status)
//
// Any error transitions the build to "failed" with the error message.
// The stub provider simulates a deterministic success path so the workflow
// is observable end-to-end without external dependencies.
func (s *CICDService) runBuildWorkflow(buildID uint64, cfg *model.CICDConfig, trigger model.BuildTrigger) {
	var build model.BuildRecord
	if err := s.db.First(&build, buildID).Error; err != nil {
		return
	}
	if build.Status == model.BuildCancelled {
		return
	}

	// ---- Step 1: queued + trigger ----
	now := time.Now()
	build.Status = model.BuildQueued
	build.StartedAt = &now
	build.Progress = 5
	stage := "queued"
	build.CurrentStage = &stage
	s.db.Save(&build)
	s.pushEvent("cicd_build.updated", s.buildToResponse(&build, cfg))

	externalID, buildURL, trigErr := s.provider.Trigger(context.Background(), cfg, CICDTriggerRequest{
		Branch:    build.Branch,
		CommitSHA: build.CommitSHA,
		Trigger:   trigger,
	})
	if trigErr != nil {
		s.failBuild(&build, cfg, fmt.Sprintf("Provider trigger failed: %v", trigErr))
		return
	}
	build.ExternalBuildID = externalID
	if buildURL != "" {
		build.BuildURL = buildURL
	}
	s.db.Save(&build)
	s.pushEvent("cicd_build.updated", s.buildToResponse(&build, cfg))

	// ---- Step 2: running + poll ----
	build.Status = model.BuildRunning
	build.Progress = 20
	runningStage := "running"
	build.CurrentStage = &runningStage
	s.db.Save(&build)
	s.pushEvent("cicd_build.updated", s.buildToResponse(&build, cfg))

	// Poll loop: cap at ~30s of simulated time so a stuck provider does not
	// block forever. The stub provider completes within a few iterations.
	deadline := time.Now().Add(60 * time.Second)
	lastStatus := ""
	for {
		if time.Now().After(deadline) {
			s.failBuild(&build, cfg, "Build polling timed out")
			return
		}
		// Re-fetch to honor in-flight cancellation.
		if err := s.db.First(&build, buildID).Error; err != nil {
			return
		}
		if build.Status == model.BuildCancelled {
			return
		}
		time.Sleep(1 * time.Second)

		status, pollErr := s.provider.GetStatus(context.Background(), cfg, externalID)
		if pollErr != nil {
			// Transient poll errors should not fail the build immediately;
			// retry until the deadline.
			continue
		}
		if status.Progress > build.Progress {
			build.Progress = status.Progress
		}
		if status.Stage != "" {
			build.CurrentStage = &status.Stage
		}
		if len(status.Stages) > 0 {
			stagesJSON, _ := json.Marshal(status.Stages)
			build.Stages = stagesJSON
		}
		if status.BuildURL != "" && build.BuildURL == "" {
			build.BuildURL = status.BuildURL
		}
		curStatus := string(status.Status)
		if curStatus != lastStatus {
			build.Status = status.Status
			lastStatus = curStatus
		}
		s.db.Save(&build)
		s.pushEvent("cicd_build.updated", s.buildToResponse(&build, cfg))

		switch status.Status {
		case model.BuildSuccess:
			s.completeBuild(&build, cfg, status)
			return
		case model.BuildFailed:
			msg := status.ErrorMessage
			if msg == "" {
				msg = "Build failed on the provider"
			}
			s.failBuild(&build, cfg, msg)
			return
		case model.BuildCancelled:
			now := time.Now()
			build.CancelledAt = &now
			s.db.Save(&build)
			s.pushEvent("cicd_build.cancelled", s.buildToResponse(&build, cfg))
			return
		}
	}
}

// completeBuild finalizes a successful build.
func (s *CICDService) completeBuild(build *model.BuildRecord, cfg *model.CICDConfig, status CICDProviderStatus) {
	now := time.Now()
	build.Status = model.BuildSuccess
	build.Progress = 100
	completedStage := "completed"
	build.CurrentStage = &completedStage
	build.CompletedAt = &now
	if build.StartedAt != nil {
		build.DurationMs = now.Sub(*build.StartedAt).Milliseconds()
	}
	s.db.Save(build)
	s.pushEvent("cicd_build.completed", s.buildToResponse(build, cfg))
}

// failBuild marks a build as failed and emits an SSE event.
func (s *CICDService) failBuild(build *model.BuildRecord, cfg *model.CICDConfig, message string) {
	now := time.Now()
	build.Status = model.BuildFailed
	build.ErrorMessage = &message
	build.CompletedAt = &now
	stage := "failed"
	build.CurrentStage = &stage
	if build.StartedAt != nil {
		build.DurationMs = now.Sub(*build.StartedAt).Milliseconds()
	}
	s.db.Save(build)
	s.pushEvent("cicd_build.failed", s.buildToResponse(build, cfg))
}

// ======== Response builders ========

// configToResponse converts a model.CICDConfig to its API representation.
func (s *CICDService) configToResponse(cfg *model.CICDConfig) *CICDConfigResponse {
	events := cfg.TriggerEvents
	if len(events) == 0 {
		events = json.RawMessage("[]")
	}
	extra := cfg.ExtraConfig
	if len(extra) == 0 {
		extra = json.RawMessage("{}")
	}
	var parsedEvents []string
	_ = json.Unmarshal(events, &parsedEvents)
	if parsedEvents == nil {
		parsedEvents = []string{}
	}
	return &CICDConfigResponse{
		ID:            cfg.ID,
		WorkspaceID:   cfg.WorkspaceID,
		ProjectID:     cfg.ProjectID,
		Name:          cfg.Name,
		Provider:      string(cfg.Provider),
		APIEndpoint:   cfg.APIEndpoint,
		ProjectSlug:   cfg.ProjectSlug,
		DefaultBranch: cfg.DefaultBranch,
		AuthTokenRef:  cfg.AuthTokenRef,
		TriggerEvents: parsedEvents,
		ExtraConfig:   extra,
		Enabled:       cfg.Enabled,
		CreatedAt:     cfg.CreatedAt,
		UpdatedAt:     cfg.UpdatedAt,
	}
}

// buildToResponse converts a model.BuildRecord to its API representation.
func (s *CICDService) buildToResponse(b *model.BuildRecord, cfg *model.CICDConfig) *BuildRecordResponse {
	stages := b.Stages
	if len(stages) == 0 {
		stages = json.RawMessage("[]")
	}
	resp := &BuildRecordResponse{
		ID:              b.ID,
		WorkspaceID:     b.WorkspaceID,
		ProjectID:       b.ProjectID,
		CICDConfigID:    b.CICDConfigID,
		Trigger:         string(b.Trigger),
		Branch:          b.Branch,
		CommitSHA:       b.CommitSHA,
		IssueID:         b.IssueID,
		AgentTaskID:     b.AgentTaskID,
		TriggeredByID:   b.TriggeredByID,
		ExternalBuildID: b.ExternalBuildID,
		BuildURL:        b.BuildURL,
		Stages:          stages,
		Status:          string(b.Status),
		Progress:        b.Progress,
		CurrentStage:    b.CurrentStage,
		ErrorMessage:    b.ErrorMessage,
		StartedAt:       b.StartedAt,
		CompletedAt:     b.CompletedAt,
		CancelledAt:     b.CancelledAt,
		DurationMs:      b.DurationMs,
		CreatedAt:       b.CreatedAt,
		UpdatedAt:       b.UpdatedAt,
	}
	if cfg != nil {
		resp.CICDConfigName = cfg.Name
	}
	return resp
}

// pushEvent broadcasts a cicd_* SSE event to all clients.
func (s *CICDService) pushEvent(event string, resp interface{}) {
	data, _ := json.Marshal(resp)
	SSE.BroadcastEvent(event, json.RawMessage(data))
}

// ======== Helpers ========

// isValidProvider returns true when p is one of the supported providers.
func isValidProvider(p model.CICDProvider) bool {
	switch p {
	case model.CICDProviderGitHubActions,
		model.CICDProviderGitLabCI,
		model.CICDProviderJenkins,
		model.CICDProviderGeneric:
		return true
	}
	return false
}

// isValidTrigger returns true when t is one of the supported triggers.
func isValidTrigger(t model.BuildTrigger) bool {
	switch t {
	case model.BuildTriggerManual,
		model.BuildTriggerPush,
		model.BuildTriggerPull,
		model.BuildTriggerSchedule,
		model.BuildTriggerAgent,
		model.BuildTriggerWebhook:
		return true
	}
	return false
}

// normalizeStringArray marshals a string slice to JSONB, treating nil/empty
// as an empty array so the column default stays consistent.
func normalizeStringArray(in []string) json.RawMessage {
	if in == nil {
		return json.RawMessage("[]")
	}
	b, _ := json.Marshal(in)
	return b
}

// normalizeCICDJSON ensures empty/nil JSON is stored as fallback rather than null.
func normalizeCICDJSON(raw json.RawMessage, fallback string) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return json.RawMessage(fallback)
	}
	return raw
}

// ======== Stub CI/CD provider ========

// stubCICDProvider simulates a CI/CD pipeline without external calls. It
// advances through a deterministic set of stages (build → test → deploy)
// over a few poll iterations and ends in success. Used when no real
// provider is configured so the workflow remains observable end-to-end.
type stubCICDProvider struct {
	// callCount tracks poll iterations per external build id so the stub
	// can advance through stages. Tests rely on the deterministic flow.
	callCount map[string]int
}

// stubCICDStages is the canonical stage list used by the stub provider.
var stubCICDStages = []string{"build", "test", "deploy"}

func (p *stubCICDProvider) Trigger(ctx context.Context, cfg *model.CICDConfig, req CICDTriggerRequest) (string, string, error) {
	if cfg == nil {
		return "", "", errors.New("config is required")
	}
	if cfg.APIEndpoint == "" {
		// Without an endpoint the stub still produces a fake build id so
		// the workflow remains observable.
	}
	if p.callCount == nil {
		p.callCount = make(map[string]int)
	}
	externalID := fmt.Sprintf("stub-%d-%d", cfg.ID, time.Now().UnixNano())
	buildURL := ""
	if cfg.APIEndpoint != "" {
		buildURL = strings.TrimRight(cfg.APIEndpoint, "/") + "/builds/" + externalID
	}
	p.callCount[externalID] = 0
	return externalID, buildURL, nil
}

func (p *stubCICDProvider) GetStatus(ctx context.Context, cfg *model.CICDConfig, externalID string) (CICDProviderStatus, error) {
	if p.callCount == nil {
		p.callCount = make(map[string]int)
	}
	p.callCount[externalID]++
	n := p.callCount[externalID]

	// Simulate ~3 poll iterations to walk through the stages.
	stages := make([]CICDStage, 0, len(stubCICDStages))
	now := time.Now()
	for i, name := range stubCICDStages {
		st := "pending"
		var started, completed *time.Time
		// Stage index that should currently be running.
		activeIdx := n - 1
		if i < activeIdx {
			st = "success"
			s := now.Add(-time.Duration(len(stubCICDStages)-i) * time.Second)
			c := now.Add(-time.Duration(len(stubCICDStages)-i-1) * time.Second)
			started = &s
			completed = &c
		} else if i == activeIdx {
			st = "running"
			s := now
			started = &s
		}
		stages = append(stages, CICDStage{
			Name:        name,
			Status:      st,
			StartedAt:   started,
			CompletedAt: completed,
			DurationMs:  int64(800 + i*200),
		})
	}

	status := CICDProviderStatus{Stages: stages}
	switch {
	case n >= len(stubCICDStages)+1:
		// All stages done → success.
		status.Status = model.BuildSuccess
		status.Progress = 100
		status.Stage = "completed"
		// Mark every stage as success in the final snapshot.
		for i := range stages {
			stages[i].Status = "success"
			if stages[i].CompletedAt == nil {
				c := now
				stages[i].CompletedAt = &c
			}
		}
		status.Stages = stages
	case n == 1:
		status.Status = model.BuildRunning
		status.Progress = 25
		status.Stage = stubCICDStages[0]
	default:
		status.Status = model.BuildRunning
		idx := n - 1
		if idx >= len(stubCICDStages) {
			idx = len(stubCICDStages) - 1
		}
		status.Stage = stubCICDStages[idx]
		status.Progress = 25 + idx*25
	}
	return status, nil
}

func (p *stubCICDProvider) Cancel(ctx context.Context, cfg *model.CICDConfig, externalID string) error {
	// Stub: no-op, the service marks the record as cancelled.
	return nil
}

// Compile-time interface check.
var _ CICDProvider = (*stubCICDProvider)(nil)
