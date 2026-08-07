package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/ai/llm"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// DeveloperAgentService implements PRD P4-001: Developer Agent.
//
// The Developer Agent takes a requirement (typically a User Story work item
// plus optional design document) and orchestrates the following pipeline:
//
//	analyze requirement → generate code → commit to git → open pull request
//
// Each execution is recorded as a DeveloperJob row so the lifecycle can be
// observed and audited. The service is intentionally decoupled from the LLM
// through the CodeGenerator interface so unit tests can substitute a stub
// generator without touching the GitHub API client.
type DeveloperAgentService struct {
	db        *gorm.DB
	gh        *GitHubService
	llm       *llm.LLMClient
	generator CodeGenerator
}

// CodeGenerator abstracts the code-generation step so it can be replaced in
// tests. Implementations should return one or more files (path + content)
// suitable for committing to the target repository.
type CodeGenerator interface {
	Generate(ctx context.Context, req CodeGenerationRequest) ([]GeneratedFile, error)
}

// CodeGenerationRequest bundles the inputs handed to the code generator.
type CodeGenerationRequest struct {
	WorkspaceID     uint64
	ProjectID       *uint64
	IssueID         *uint64
	Title           string
	RequirementText string
	DesignDocURL    string
	Language        string // optional hint, e.g. "go", "typescript"
	ExtraContext    map[string]interface{}
}

// GeneratedFile describes a single file produced by the code generator.
type GeneratedFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Mode    string `json:"mode,omitempty"`
}

// NewDeveloperAgentService creates a new DeveloperAgentService.
// llm may be nil — when no LLM is configured, code generation falls back to
// an inline stub that emits a single README.md describing the requirement.
func NewDeveloperAgentService(db *gorm.DB, gh *GitHubService, llm *llm.LLMClient) *DeveloperAgentService {
	svc := &DeveloperAgentService{db: db, gh: gh, llm: llm}
	svc.generator = &llmCodeGenerator{llm: llm}
	return svc
}

// SetCodeGenerator overrides the default code generator. Used in tests.
func (s *DeveloperAgentService) SetCodeGenerator(g CodeGenerator) {
	if g != nil {
		s.generator = g
	}
}

// checkWorkspaceAdmin mirrors the guard used by other workspace services.
func (s *DeveloperAgentService) checkWorkspaceAdmin(workspaceID, callerID uint64) error {
	var member model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ? AND is_active = ?", workspaceID, callerID, true).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.Forbidden("You must be a workspace admin to manage developer agent jobs")
		}
		return common.Internal("Database error")
	}
	if member.Role < common.RoleAdmin {
		return common.Forbidden("You must be a workspace admin to manage developer agent jobs")
	}
	return nil
}

// ======== Request / Response types ========

// DeveloperJobCreate captures the inputs for a new Developer Agent run.
type DeveloperJobCreate struct {
	Title           string          `json:"title" binding:"required"`
	RequirementText string          `json:"requirement_text"`
	DesignDocURL    *string         `json:"design_doc_url"`
	ProjectID       *uint64         `json:"project_id"`
	IssueID         *uint64         `json:"issue_id"`
	GitConnectionID *uint64         `json:"git_connection_id"`
	GitProvider     string          `json:"git_provider"`
	BranchName      string          `json:"branch_name"`
	BaseBranch      string          `json:"base_branch"`
	CommitMessage   string          `json:"commit_message"`
	PRTitle         string          `json:"pr_title"`
	PRBody          string          `json:"pr_body"`
	Language        string          `json:"language"`
	Files           []GeneratedFile `json:"files"`         // optional: pre-generated files to commit
	InputContext    json.RawMessage `json:"input_context"` // optional: extra context
}

// DeveloperJobResponse is the API representation of a DeveloperJob.
type DeveloperJobResponse struct {
	ID              uint64          `json:"id"`
	WorkspaceID     uint64          `json:"workspace_id"`
	ProjectID       *uint64         `json:"project_id,omitempty"`
	IssueID         *uint64         `json:"issue_id,omitempty"`
	AgentTaskID     *uint64         `json:"agent_task_id,omitempty"`
	GitProvider     string          `json:"git_provider"`
	GitConnectionID *uint64         `json:"git_connection_id,omitempty"`
	Title           string          `json:"title"`
	RequirementText string          `json:"requirement_text"`
	DesignDocURL    *string         `json:"design_doc_url,omitempty"`
	InputContext    json.RawMessage `json:"input_context"`
	BranchName      string          `json:"branch_name"`
	BaseBranch      string          `json:"base_branch"`
	CommitMessage   string          `json:"commit_message"`
	GeneratedFiles  json.RawMessage `json:"generated_files"`
	PRNumber        *int            `json:"pr_number,omitempty"`
	PRURL           *string         `json:"pr_url,omitempty"`
	PRTitle         *string         `json:"pr_title,omitempty"`
	CommitSHA       *string         `json:"commit_sha,omitempty"`
	Status          string          `json:"status"`
	Progress        int             `json:"progress"`
	CurrentStep     *string         `json:"current_step,omitempty"`
	ErrorMessage    *string         `json:"error_message,omitempty"`
	StartedAt       *time.Time      `json:"started_at,omitempty"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty"`
	CancelledAt     *time.Time      `json:"cancelled_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// ======== CRUD ========

// Create starts a new Developer Agent job.
//
// The job is persisted synchronously (status=pending) and the workflow runs
// asynchronously so the API can return immediately. Callers poll the job
// status or subscribe to SSE events (developer_job.updated) for completion.
func (s *DeveloperAgentService) Create(wid, callerID uint64, req DeveloperJobCreate) (*DeveloperJobResponse, error) {
	if err := s.checkWorkspaceAdmin(wid, callerID); err != nil {
		return nil, err
	}
	if req.Title == "" {
		return nil, common.BadRequest("Title is required")
	}
	if req.GitConnectionID == nil || *req.GitConnectionID == 0 {
		return nil, common.BadRequest("git_connection_id is required")
	}

	gitProvider := req.GitProvider
	if gitProvider == "" {
		gitProvider = "github"
	}
	baseBranch := req.BaseBranch
	if baseBranch == "" {
		baseBranch = "main"
	}

	// Persist initial record so the API can return an ID immediately.
	filesJSON, _ := json.Marshal(req.Files)
	if string(filesJSON) == "null" {
		filesJSON = []byte("[]")
	}

	job := model.DeveloperJob{
		WorkspaceID:     wid,
		ProjectID:       req.ProjectID,
		IssueID:         req.IssueID,
		GitProvider:     gitProvider,
		GitConnectionID: req.GitConnectionID,
		Title:           req.Title,
		RequirementText: req.RequirementText,
		DesignDocURL:    req.DesignDocURL,
		InputContext:    normalizeDeveloperJSON(req.InputContext),
		BranchName:      req.BranchName,
		BaseBranch:      baseBranch,
		CommitMessage:   req.CommitMessage,
		GeneratedFiles:  filesJSON,
		Status:          model.DeveloperJobPending,
	}

	if err := s.db.Create(&job).Error; err != nil {
		return nil, common.Internal("Failed to create developer job")
	}

	// Persist input context snapshot for the async runner.
	runCtx := developerRunContext{
		CallerID:    callerID,
		PRTitle:     req.PRTitle,
		PRBody:      req.PRBody,
		Language:    req.Language,
		PresetFiles: req.Files,
	}

	// Spawn the asynchronous workflow.
	go s.runWorkflow(job.ID, runCtx)

	resp := s.toResponse(&job)
	s.pushEvent("developer_job.created", resp)
	return resp, nil
}

// Get returns a single DeveloperJob by ID.
func (s *DeveloperAgentService) Get(id uint64) (*DeveloperJobResponse, error) {
	var job model.DeveloperJob
	if err := s.db.First(&job, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Developer job not found")
		}
		return nil, common.Internal("Failed to get developer job")
	}
	return s.toResponse(&job), nil
}

// List returns DeveloperJobs for a workspace, newest first.
// status filter is optional.
func (s *DeveloperAgentService) List(wid uint64, status string, limit int) ([]DeveloperJobResponse, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := s.db.Where("workspace_id = ?", wid)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var jobs []model.DeveloperJob
	if err := q.Order("created_at DESC").Limit(limit).Find(&jobs).Error; err != nil {
		return nil, common.Internal("Failed to list developer jobs")
	}
	out := make([]DeveloperJobResponse, 0, len(jobs))
	for i := range jobs {
		out = append(out, *s.toResponse(&jobs[i]))
	}
	return out, nil
}

// Cancel marks a pending/running job as cancelled. In-flight GitHub API
// calls are best-effort and may still complete; the result is simply not
// promoted to "completed".
func (s *DeveloperAgentService) Cancel(id uint64, callerID uint64) (*DeveloperJobResponse, error) {
	var job model.DeveloperJob
	if err := s.db.First(&job, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Developer job not found")
		}
		return nil, common.Internal("Failed to get developer job")
	}
	if err := s.checkWorkspaceAdmin(job.WorkspaceID, callerID); err != nil {
		return nil, err
	}
	if job.Status == model.DeveloperJobCompleted || job.Status == model.DeveloperJobFailed {
		return nil, common.BadRequest("Cannot cancel a terminal job")
	}
	now := time.Now()
	job.Status = model.DeveloperJobCancelled
	job.CancelledAt = &now
	if err := s.db.Save(&job).Error; err != nil {
		return nil, common.Internal("Failed to cancel developer job")
	}
	resp := s.toResponse(&job)
	s.pushEvent("developer_job.cancelled", resp)
	return resp, nil
}

// Delete removes a DeveloperJob record (soft-delete via BaseModel).
func (s *DeveloperAgentService) Delete(id uint64, callerID uint64) error {
	var job model.DeveloperJob
	if err := s.db.First(&job, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NotFound("Developer job not found")
		}
		return common.Internal("Failed to get developer job")
	}
	if err := s.checkWorkspaceAdmin(job.WorkspaceID, callerID); err != nil {
		return err
	}
	return s.db.Delete(&job).Error
}

// ======== Async workflow ========

// developerRunContext carries request-scoped data into the async goroutine.
type developerRunContext struct {
	CallerID    uint64
	PRTitle     string
	PRBody      string
	Language    string
	PresetFiles []GeneratedFile
}

// runWorkflow executes the full Developer Agent pipeline.
//
// Steps:
//  1. analyzing   — load issue / design doc, build context
//  2. generating  — invoke CodeGenerator (LLM or pre-supplied files)
//  3. committing  — create branch + commit files via GitHub API
//  4. opening_pr  — open a pull request on the head branch
//  5. completed   — record PR URL + commit SHA
//
// Any error transitions the job to "failed" with the error message.
func (s *DeveloperAgentService) runWorkflow(jobID uint64, runCtx developerRunContext) {
	// Reload the job to ensure we have the latest state.
	var job model.DeveloperJob
	if err := s.db.First(&job, jobID).Error; err != nil {
		return
	}
	if job.Status == model.DeveloperJobCancelled {
		return
	}

	now := time.Now()
	job.Status = model.DeveloperJobAnalyzing
	job.StartedAt = &now
	job.Progress = 10
	step := "analyzing requirement"
	job.CurrentStep = &step
	s.db.Save(&job)
	s.pushEvent("developer_job.updated", s.toResponse(&job))

	// Resolve the GitHub connection.
	conn, err := s.gh.GetConnection(*job.GitConnectionID)
	if err != nil {
		s.failJob(&job, fmt.Sprintf("Failed to resolve git connection: %v", err))
		return
	}
	if !conn.IsEnabled {
		s.failJob(&job, "GitHub connection is disabled")
		return
	}

	// Build the requirement context. If an issue ID is supplied, pull its
	// name + description so the LLM has structured context to work with.
	requirementText := job.RequirementText
	if job.IssueID != nil && *job.IssueID != 0 {
		var issue model.Issue
		if err := s.db.First(&issue, *job.IssueID).Error; err == nil {
			if requirementText == "" {
				requirementText = issue.Name
			}
			requirementText = fmt.Sprintf("%s\n\n%s", requirementText, issue.DescriptionHTML)
		}
	}

	// Determine branch name (auto-generate if not provided).
	branchName := job.BranchName
	if branchName == "" {
		branchName = s.generateBranchName(&job)
		job.BranchName = branchName
		s.db.Save(&job)
	}

	// ---- Step 2: generate code ----
	job.Status = model.DeveloperJobGenerating
	job.Progress = 30
	genStep := "generating code"
	job.CurrentStep = &genStep
	s.db.Save(&job)
	s.pushEvent("developer_job.updated", s.toResponse(&job))

	var files []GeneratedFile
	if len(runCtx.PresetFiles) > 0 {
		// Caller supplied files directly — skip LLM generation.
		files = runCtx.PresetFiles
	} else {
		genReq := CodeGenerationRequest{
			WorkspaceID:     job.WorkspaceID,
			ProjectID:       job.ProjectID,
			IssueID:         job.IssueID,
			Title:           job.Title,
			RequirementText: requirementText,
			Language:        runCtx.Language,
			ExtraContext: map[string]interface{}{
				"design_doc_url": job.DesignDocURL,
				"branch":         branchName,
				"base_branch":    job.BaseBranch,
			},
		}
		if job.DesignDocURL != nil {
			genReq.DesignDocURL = *job.DesignDocURL
		}
		generated, genErr := s.generator.Generate(context.Background(), genReq)
		if genErr != nil {
			s.failJob(&job, fmt.Sprintf("Code generation failed: %v", genErr))
			return
		}
		files = generated
	}
	if len(files) == 0 {
		s.failJob(&job, "Code generator produced no files")
		return
	}

	// Persist the generated files on the job record.
	filesJSON, _ := json.Marshal(files)
	job.GeneratedFiles = filesJSON
	s.db.Save(&job)

	// ---- Step 3: commit code ----
	job.Status = model.DeveloperJobCommitting
	job.Progress = 60
	commitStep := "committing code to repository"
	job.CurrentStep = &commitStep
	s.db.Save(&job)
	s.pushEvent("developer_job.updated", s.toResponse(&job))

	if _, err := s.gh.CreateBranch(conn.RepoOwner, conn.RepoName, job.BaseBranch, branchName, conn.AccessToken, ""); err != nil {
		// 422 = branch already exists; that's fine, we can still commit on it.
		if !isBranchExistsErr(err) {
			s.failJob(&job, fmt.Sprintf("Failed to create branch: %v", err))
			return
		}
	}

	commitMessage := job.CommitMessage
	if commitMessage == "" {
		commitMessage = fmt.Sprintf("feat: %s\n\nGenerated by Developer Agent (job #%d)", job.Title, job.ID)
	}

	ghFiles := make([]GitHubFileInput, 0, len(files))
	for _, f := range files {
		ghFiles = append(ghFiles, GitHubFileInput{Path: f.Path, Content: f.Content, Mode: f.Mode})
	}
	commitSHA, committedCount, commitErr := s.gh.CommitFiles(conn.RepoOwner, conn.RepoName, branchName, commitMessage, conn.AccessToken, ghFiles)
	if commitErr != nil && committedCount == 0 {
		s.failJob(&job, fmt.Sprintf("Failed to commit files: %v", commitErr))
		return
	}
	if commitSHA != "" {
		job.CommitSHA = &commitSHA
		s.db.Save(&job)
	}

	// ---- Step 4: open PR ----
	job.Status = model.DeveloperJobOpeningPR
	job.Progress = 85
	prStep := "opening pull request"
	job.CurrentStep = &prStep
	s.db.Save(&job)
	s.pushEvent("developer_job.updated", s.toResponse(&job))

	prTitle := runCtx.PRTitle
	if prTitle == "" {
		prTitle = fmt.Sprintf("feat: %s", job.Title)
	}
	prBody := runCtx.PRBody
	if prBody == "" {
		prBody = s.buildDefaultPRBody(&job, requirementText, files, committedCount)
	}

	pr, prErr := s.gh.CreatePullRequest(conn.RepoOwner, conn.RepoName, job.BaseBranch, branchName, prTitle, prBody, conn.AccessToken)
	if prErr != nil {
		s.failJob(&job, fmt.Sprintf("Failed to open pull request: %v", prErr))
		return
	}

	// ---- Step 5: complete ----
	job.PRNumber = &pr.Number
	job.PRURL = &pr.HTMLURL
	job.PRTitle = &pr.Title
	if pr.Head.SHA != "" {
		sha := pr.Head.SHA
		job.CommitSHA = &sha
	}
	completedAt := time.Now()
	job.Status = model.DeveloperJobCompleted
	job.Progress = 100
	completedStep := "completed"
	job.CurrentStep = &completedStep
	job.CompletedAt = &completedAt
	s.db.Save(&job)
	s.pushEvent("developer_job.completed", s.toResponse(&job))

	// Optionally link the issue to the PR via the existing git integration.
	if job.IssueID != nil && conn.ProjectID != 0 {
		s.linkIssueToPR(*job.IssueID, conn.ID, pr)
	}
}

// failJob marks a job as failed and emits an SSE event.
func (s *DeveloperAgentService) failJob(job *model.DeveloperJob, message string) {
	now := time.Now()
	job.Status = model.DeveloperJobFailed
	job.ErrorMessage = &message
	job.CompletedAt = &now
	step := "failed"
	job.CurrentStep = &step
	s.db.Save(job)
	s.pushEvent("developer_job.failed", s.toResponse(job))
}

// generateBranchName produces a deterministic, unique-ish branch name.
// Format: dev-agent/<job-id>-<slug-of-title>
func (s *DeveloperAgentService) generateBranchName(job *model.DeveloperJob) string {
	slug := sanitizeBranchSlug(job.Title)
	if slug == "" {
		slug = "feature"
	}
	return fmt.Sprintf("dev-agent/%d-%s", job.ID, slug)
}

// buildDefaultPRBody assembles a markdown PR description from the job context.
func (s *DeveloperAgentService) buildDefaultPRBody(job *model.DeveloperJob, requirement string, files []GeneratedFile, committed int) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## %s\n\n", job.Title))
	b.WriteString("Generated by **Developer Agent** (ReqMango P4-001).\n\n")
	b.WriteString("### Requirement\n\n")
	if requirement != "" {
		b.WriteString(requirement)
		b.WriteString("\n\n")
	} else {
		b.WriteString("_No requirement text provided._\n\n")
	}
	b.WriteString("### Changed files\n\n")
	for _, f := range files {
		b.WriteString(fmt.Sprintf("- `%s`\n", f.Path))
	}
	b.WriteString(fmt.Sprintf("\n\n_%d file(s) committed on branch `%s`._\n", committed, job.BranchName))
	return b.String()
}

// linkIssueToPR records a GitIssueLink tying the source issue to the opened PR.
func (s *DeveloperAgentService) linkIssueToPR(issueID, integrationID uint64, pr *GitHubPullRequest) {
	link := model.GitIssueLink{
		IssueID:       issueID,
		GitType:       "github_pr",
		GitID:         fmt.Sprintf("%d", pr.Number),
		GitURL:        pr.HTMLURL,
		GitTitle:      pr.Title,
		GitState:      pr.State,
		GitBranch:     pr.Head.Ref,
		IntegrationID: integrationID,
	}
	s.db.Create(&link)
}

// toResponse converts a model.DeveloperJob to its API representation.
func (s *DeveloperAgentService) toResponse(job *model.DeveloperJob) *DeveloperJobResponse {
	files := job.GeneratedFiles
	if len(files) == 0 {
		files = json.RawMessage("[]")
	}
	ctx := job.InputContext
	if len(ctx) == 0 {
		ctx = json.RawMessage("{}")
	}
	return &DeveloperJobResponse{
		ID:              job.ID,
		WorkspaceID:     job.WorkspaceID,
		ProjectID:       job.ProjectID,
		IssueID:         job.IssueID,
		AgentTaskID:     job.AgentTaskID,
		GitProvider:     job.GitProvider,
		GitConnectionID: job.GitConnectionID,
		Title:           job.Title,
		RequirementText: job.RequirementText,
		DesignDocURL:    job.DesignDocURL,
		InputContext:    ctx,
		BranchName:      job.BranchName,
		BaseBranch:      job.BaseBranch,
		CommitMessage:   job.CommitMessage,
		GeneratedFiles:  files,
		PRNumber:        job.PRNumber,
		PRURL:           job.PRURL,
		PRTitle:         job.PRTitle,
		CommitSHA:       job.CommitSHA,
		Status:          string(job.Status),
		Progress:        job.Progress,
		CurrentStep:     job.CurrentStep,
		ErrorMessage:    job.ErrorMessage,
		StartedAt:       job.StartedAt,
		CompletedAt:     job.CompletedAt,
		CancelledAt:     job.CancelledAt,
		CreatedAt:       job.CreatedAt,
		UpdatedAt:       job.UpdatedAt,
	}
}

// pushEvent broadcasts a developer_job SSE event to all clients.
func (s *DeveloperAgentService) pushEvent(event string, resp *DeveloperJobResponse) {
	data, _ := json.Marshal(resp)
	SSE.BroadcastEvent(event, json.RawMessage(data))
}

// ======== Helpers ========

// normalizeDeveloperJSON ensures empty/nil JSON is stored as "{}" rather than null.
func normalizeDeveloperJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return json.RawMessage("{}")
	}
	return raw
}

// isBranchExistsErr returns true when the GitHub API reported that the branch
// already exists (HTTP 422 with the "already exists" message).
func isBranchExistsErr(err error) bool {
	if err == nil {
		return false
	}
	if ae, ok := err.(*common.AppError); ok {
		// AppError from CreateBranch uses BadRequest for 422 responses.
		if ae.Code == 400 && strings.Contains(ae.Message, "already exists") {
			return true
		}
	}
	return strings.Contains(err.Error(), "already exists")
}

// sanitizeBranchSlug converts a free-form title into a git-safe slug.
func sanitizeBranchSlug(title string) string {
	title = strings.ToLower(strings.TrimSpace(title))
	title = strings.ReplaceAll(title, " ", "-")
	title = strings.ReplaceAll(title, "_", "-")
	var b strings.Builder
	prevDash := false
	for _, r := range title {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if r == '-' && !prevDash {
			b.WriteRune('-')
			prevDash = true
		}
	}
	s := strings.Trim(b.String(), "-")
	if len(s) > 40 {
		s = strings.Trim(s[:40], "-")
	}
	return s
}

// ======== LLM code generator ========

// llmCodeGenerator implements CodeGenerator using the workspace LLM client.
// When no LLM is configured (llm == nil), it falls back to a deterministic
// stub so the workflow remains end-to-end testable without external deps.
type llmCodeGenerator struct {
	llm *llm.LLMClient
}

func (g *llmCodeGenerator) Generate(ctx context.Context, req CodeGenerationRequest) ([]GeneratedFile, error) {
	if g.llm == nil {
		return stubGenerate(req), nil
	}

	systemPrompt := `You are a Developer Agent. Your job is to produce a minimal, production-ready code change that implements the given requirement.

Output ONLY a JSON object with the following shape, no prose, no markdown fences:
{
  "files": [
    { "path": "<repository-relative path>", "content": "<full file content>", "mode": "100644" }
  ]
}

Guidelines:
- Prefer one focused file unless the requirement clearly needs more.
- Include sensible defaults and inline comments for non-obvious logic.
- Do NOT include tests in this step — they are produced by the Tester Agent.
- If a design_doc_url is provided, treat it as authoritative.`

	userMsg := fmt.Sprintf("Requirement: %s\n\nTitle: %s\nLanguage hint: %s\nDesign doc: %s",
		req.RequirementText, req.Title, req.Language, req.DesignDocURL)

	resp, err := g.llm.ChatSync(ctx, systemPrompt, []llm.Message{
		{Role: "user", Content: userMsg},
	}, nil)
	if err != nil {
		// Fall back to stub when the LLM is not configured (e.g. missing API key).
		// Real generation failures should still propagate so the job can fail loudly.
		if strings.Contains(err.Error(), "未配置") || strings.Contains(err.Error(), "API Key") {
			return stubGenerate(req), nil
		}
		return nil, err
	}
	if resp.Content == "" {
		return nil, errors.New("LLM returned empty content")
	}

	// The LLM may wrap JSON in markdown fences; strip them defensively.
	body := stripCodeFences(resp.Content)
	var parsed struct {
		Files []GeneratedFile `json:"files"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		return nil, fmt.Errorf("LLM output was not valid JSON: %w", err)
	}
	if len(parsed.Files) == 0 {
		return nil, errors.New("LLM did not produce any files")
	}
	return parsed.Files, nil
}

// stubGenerate produces a placeholder README.md describing the requirement.
// Used when no LLM is configured so the workflow remains observable.
func stubGenerate(req CodeGenerationRequest) []GeneratedFile {
	content := fmt.Sprintf(`# %s

Generated by Developer Agent (stub generator — no LLM configured).

## Requirement

%s

## Notes

- Workspace: %d
- Language hint: %s
- Replace this stub by configuring an LLM provider (Agent Configs) to enable real code generation.
`, req.Title, req.RequirementText, req.WorkspaceID, req.Language)
	return []GeneratedFile{
		{Path: "DEVELOPER_AGENT_OUTPUT.md", Content: content, Mode: "100644"},
	}
}

// stripCodeFences removes surrounding ```json ... ``` fences if present.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	// Drop the opening fence line.
	if idx := strings.Index(s, "\n"); idx >= 0 {
		s = s[idx+1:]
	}
	// Drop the trailing fence.
	s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	return strings.TrimSpace(s)
}

// Compile-time interface check.
var _ CodeGenerator = (*llmCodeGenerator)(nil)
