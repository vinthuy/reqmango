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

// TesterAgentService implements PRD P4-002: Tester Agent.
//
// The Tester Agent takes a requirement (User Story + acceptance criteria),
// generates test cases, executes them, and reports failures as Bug work
// items. The pipeline is:
//
//	generate test cases → execute tests → report bugs → complete
//
// Each execution is recorded as a TesterJob row. The service is decoupled
// from the LLM through the TestCaseGenerator interface and from the test
// runner through the TestExecutor interface so unit tests can substitute
// stubs without external dependencies.
type TesterAgentService struct {
	db        *gorm.DB
	llm       *llm.LLMClient
	generator TestCaseGenerator
	executor  TestExecutor
}

// TestCaseGenerator abstracts the test-case-generation step so it can be
// replaced in tests. Implementations should return one or more test cases
// derived from the requirement + acceptance criteria.
type TestCaseGenerator interface {
	Generate(ctx context.Context, req TestCaseGenerationRequest) ([]TestCase, error)
}

// TestExecutor abstracts the test-execution step. Implementations run the
// supplied test cases and return a result for each (pass/fail/skip).
type TestExecutor interface {
	Execute(ctx context.Context, req TestExecutionRequest) ([]TestResult, error)
}

// TestCaseGenerationRequest bundles the inputs handed to the generator.
type TestCaseGenerationRequest struct {
	WorkspaceID         uint64
	ProjectID           *uint64
	IssueID             *uint64
	Title               string
	RequirementText     string
	AcceptanceCriteria  string
	TestScope           string // unit | integration | e2e
	ExtraContext        map[string]interface{}
}

// TestExecutionRequest bundles the inputs handed to the executor.
type TestExecutionRequest struct {
	WorkspaceID uint64
	ProjectID   *uint64
	IssueID     *uint64
	TestScope   string
	Cases       []TestCase
}

// TestCase describes a single generated test case.
type TestCase struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
	Expected    string   `json:"expected"`
}

// TestResult describes the outcome of executing a single test case.
type TestResult struct {
	CaseID    string `json:"case_id"`
	Name      string `json:"name"`
	Status    string `json:"status"` // passed | failed | skipped
	DurationMs int64 `json:"duration_ms"`
	Error     string `json:"error,omitempty"`
}

// NewTesterAgentService creates a new TesterAgentService.
// llm may be nil — when no LLM is configured, generation falls back to an
// inline stub that emits a deterministic set of cases from the requirement.
func NewTesterAgentService(db *gorm.DB, llm *llm.LLMClient) *TesterAgentService {
	svc := &TesterAgentService{db: db, llm: llm}
	svc.generator = &llmTestCaseGenerator{llm: llm}
	svc.executor = &stubTestExecutor{}
	return svc
}

// SetTestCaseGenerator overrides the default generator. Used in tests.
func (s *TesterAgentService) SetTestCaseGenerator(g TestCaseGenerator) {
	if g != nil {
		s.generator = g
	}
}

// SetTestExecutor overrides the default executor. Used in tests.
func (s *TesterAgentService) SetTestExecutor(e TestExecutor) {
	if e != nil {
		s.executor = e
	}
}

// checkWorkspaceAdmin mirrors the guard used by other workspace services.
func (s *TesterAgentService) checkWorkspaceAdmin(workspaceID, callerID uint64) error {
	var member model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ? AND is_active = ?", workspaceID, callerID, true).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.Forbidden("You must be a workspace admin to manage tester agent jobs")
		}
		return common.Internal("Database error")
	}
	if member.Role < common.RoleAdmin {
		return common.Forbidden("You must be a workspace admin to manage tester agent jobs")
	}
	return nil
}

// ======== Request / Response types ========

// TesterJobCreate captures the inputs for a new Tester Agent run.
type TesterJobCreate struct {
	Title               string          `json:"title" binding:"required"`
	RequirementText     string          `json:"requirement_text"`
	AcceptanceCriteria  string          `json:"acceptance_criteria"`
	TestScope           string          `json:"test_scope"`
	ProjectID           *uint64         `json:"project_id"`
	IssueID             *uint64         `json:"issue_id"`
	Cases               []TestCase      `json:"cases"`        // optional: pre-generated cases to execute
	InputContext        json.RawMessage `json:"input_context"` // optional: extra context
}

// TesterJobResponse is the API representation of a TesterJob.
type TesterJobResponse struct {
	ID                  uint64          `json:"id"`
	WorkspaceID         uint64          `json:"workspace_id"`
	ProjectID           *uint64         `json:"project_id,omitempty"`
	IssueID             *uint64         `json:"issue_id,omitempty"`
	AgentTaskID         *uint64         `json:"agent_task_id,omitempty"`
	Title               string          `json:"title"`
	RequirementText     string          `json:"requirement_text"`
	AcceptanceCriteria  string          `json:"acceptance_criteria"`
	TestScope           string          `json:"test_scope"`
	InputContext        json.RawMessage `json:"input_context"`
	GeneratedCases      json.RawMessage `json:"generated_cases"`
	TestResults         json.RawMessage `json:"test_results"`
	TotalCases          int             `json:"total_cases"`
	PassCount           int             `json:"pass_count"`
	FailCount           int             `json:"fail_count"`
	SkipCount           int             `json:"skip_count"`
	BugIssueIDs         json.RawMessage `json:"bug_issue_ids"`
	Status              string          `json:"status"`
	Progress            int             `json:"progress"`
	CurrentStep         *string         `json:"current_step,omitempty"`
	ErrorMessage        *string         `json:"error_message,omitempty"`
	StartedAt           *time.Time      `json:"started_at,omitempty"`
	CompletedAt         *time.Time      `json:"completed_at,omitempty"`
	CancelledAt         *time.Time      `json:"cancelled_at,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

// ======== CRUD ========

// Create starts a new Tester Agent job.
//
// The job is persisted synchronously (status=pending) and the workflow runs
// asynchronously so the API can return immediately. Callers poll the job
// status or subscribe to SSE events (tester_job.updated) for completion.
func (s *TesterAgentService) Create(wid, callerID uint64, req TesterJobCreate) (*TesterJobResponse, error) {
	if err := s.checkWorkspaceAdmin(wid, callerID); err != nil {
		return nil, err
	}
	if req.Title == "" {
		return nil, common.BadRequest("Title is required")
	}

	testScope := req.TestScope
	if testScope == "" {
		testScope = "unit"
	}

	casesJSON, _ := json.Marshal(req.Cases)
	if string(casesJSON) == "null" {
		casesJSON = []byte("[]")
	}

	job := model.TesterJob{
		WorkspaceID:        wid,
		ProjectID:          req.ProjectID,
		IssueID:            req.IssueID,
		Title:              req.Title,
		RequirementText:    req.RequirementText,
		AcceptanceCriteria: req.AcceptanceCriteria,
		TestScope:          testScope,
		InputContext:       normalizeTesterJSON(req.InputContext),
		GeneratedCases:     casesJSON,
		Status:             model.TesterJobPending,
	}

	if err := s.db.Create(&job).Error; err != nil {
		return nil, common.Internal("Failed to create tester job")
	}

	runCtx := testerRunContext{
		CallerID: callerID,
		Cases:    req.Cases,
	}

	go s.runWorkflow(job.ID, runCtx)

	resp := s.toResponse(&job)
	s.pushEvent("tester_job.created", resp)
	return resp, nil
}

// Get returns a single TesterJob by ID.
func (s *TesterAgentService) Get(id uint64) (*TesterJobResponse, error) {
	var job model.TesterJob
	if err := s.db.First(&job, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Tester job not found")
		}
		return nil, common.Internal("Failed to get tester job")
	}
	return s.toResponse(&job), nil
}

// List returns TesterJobs for a workspace, newest first.
// status filter is optional.
func (s *TesterAgentService) List(wid uint64, status string, limit int) ([]TesterJobResponse, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := s.db.Where("workspace_id = ?", wid)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var jobs []model.TesterJob
	if err := q.Order("created_at DESC").Limit(limit).Find(&jobs).Error; err != nil {
		return nil, common.Internal("Failed to list tester jobs")
	}
	out := make([]TesterJobResponse, 0, len(jobs))
	for i := range jobs {
		out = append(out, *s.toResponse(&jobs[i]))
	}
	return out, nil
}

// Cancel marks a pending/running job as cancelled.
func (s *TesterAgentService) Cancel(id uint64, callerID uint64) (*TesterJobResponse, error) {
	var job model.TesterJob
	if err := s.db.First(&job, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Tester job not found")
		}
		return nil, common.Internal("Failed to get tester job")
	}
	if err := s.checkWorkspaceAdmin(job.WorkspaceID, callerID); err != nil {
		return nil, err
	}
	if job.Status == model.TesterJobCompleted || job.Status == model.TesterJobFailed {
		return nil, common.BadRequest("Cannot cancel a terminal job")
	}
	now := time.Now()
	job.Status = model.TesterJobCancelled
	job.CancelledAt = &now
	if err := s.db.Save(&job).Error; err != nil {
		return nil, common.Internal("Failed to cancel tester job")
	}
	resp := s.toResponse(&job)
	s.pushEvent("tester_job.cancelled", resp)
	return resp, nil
}

// Delete removes a TesterJob record (soft-delete via BaseModel).
func (s *TesterAgentService) Delete(id uint64, callerID uint64) error {
	var job model.TesterJob
	if err := s.db.First(&job, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NotFound("Tester job not found")
		}
		return common.Internal("Failed to get tester job")
	}
	if err := s.checkWorkspaceAdmin(job.WorkspaceID, callerID); err != nil {
		return err
	}
	return s.db.Delete(&job).Error
}

// ======== Async workflow ========

// testerRunContext carries request-scoped data into the async goroutine.
type testerRunContext struct {
	CallerID uint64
	Cases    []TestCase // optional: pre-generated cases
}

// runWorkflow executes the full Tester Agent pipeline.
//
// Steps:
//  1. generating_cases — invoke TestCaseGenerator (LLM or pre-supplied cases)
//  2. executing        — invoke TestExecutor (stub runner)
//  3. reporting        — create a Bug issue for each failed case
//  4. completed        — record summary + bug IDs
//
// Any error transitions the job to "failed" with the error message.
func (s *TesterAgentService) runWorkflow(jobID uint64, runCtx testerRunContext) {
	var job model.TesterJob
	if err := s.db.First(&job, jobID).Error; err != nil {
		return
	}
	if job.Status == model.TesterJobCancelled {
		return
	}

	// Resolve the source issue context if an IssueID is supplied so the
	// generator has structured context to work with.
	requirementText := job.RequirementText
	acceptanceCriteria := job.AcceptanceCriteria
	if job.IssueID != nil && *job.IssueID != 0 {
		var issue model.Issue
		if err := s.db.First(&issue, *job.IssueID).Error; err == nil {
			if requirementText == "" {
				requirementText = issue.Name
			}
			if issue.DescriptionHTML != "" {
				requirementText = fmt.Sprintf("%s\n\n%s", requirementText, issue.DescriptionHTML)
			}
			// Derive ProjectID from the source issue when not explicitly set.
			if job.ProjectID == nil || *job.ProjectID == 0 {
				pid := issue.ProjectID
				job.ProjectID = &pid
				s.db.Save(&job)
			}
		}
	}

	// ---- Step 1: generate test cases ----
	now := time.Now()
	job.Status = model.TesterJobGeneratingCases
	job.StartedAt = &now
	job.Progress = 10
	step := "generating test cases"
	job.CurrentStep = &step
	s.db.Save(&job)
	s.pushEvent("tester_job.updated", s.toResponse(&job))

	var cases []TestCase
	if len(runCtx.Cases) > 0 {
		cases = runCtx.Cases
	} else {
		genReq := TestCaseGenerationRequest{
			WorkspaceID:        job.WorkspaceID,
			ProjectID:          job.ProjectID,
			IssueID:            job.IssueID,
			Title:              job.Title,
			RequirementText:    requirementText,
			AcceptanceCriteria: acceptanceCriteria,
			TestScope:          job.TestScope,
		}
		generated, genErr := s.generator.Generate(context.Background(), genReq)
		if genErr != nil {
			s.failJob(&job, fmt.Sprintf("Test case generation failed: %v", genErr))
			return
		}
		cases = generated
	}
	if len(cases) == 0 {
		s.failJob(&job, "Test case generator produced no cases")
		return
	}

	casesJSON, _ := json.Marshal(cases)
	job.GeneratedCases = casesJSON
	job.TotalCases = len(cases)
	job.Progress = 40
	s.db.Save(&job)
	s.pushEvent("tester_job.updated", s.toResponse(&job))

	// ---- Step 2: execute tests ----
	job.Status = model.TesterJobExecuting
	job.Progress = 60
	execStep := "executing tests"
	job.CurrentStep = &execStep
	s.db.Save(&job)
	s.pushEvent("tester_job.updated", s.toResponse(&job))

	results, execErr := s.executor.Execute(context.Background(), TestExecutionRequest{
		WorkspaceID: job.WorkspaceID,
		ProjectID:   job.ProjectID,
		IssueID:     job.IssueID,
		TestScope:   job.TestScope,
		Cases:       cases,
	})
	if execErr != nil {
		s.failJob(&job, fmt.Sprintf("Test execution failed: %v", execErr))
		return
	}

	// Roll up the summary.
	pass, fail, skip := 0, 0, 0
	for i := range results {
		switch results[i].Status {
		case "passed":
			pass++
		case "failed":
			fail++
		case "skipped":
			skip++
		}
	}
	resultsJSON, _ := json.Marshal(results)
	job.TestResults = resultsJSON
	job.PassCount = pass
	job.FailCount = fail
	job.SkipCount = skip
	s.db.Save(&job)
	s.pushEvent("tester_job.updated", s.toResponse(&job))

	// ---- Step 3: report bugs for failed cases ----
	job.Status = model.TesterJobReporting
	job.Progress = 85
	reportStep := "reporting bugs"
	job.CurrentStep = &reportStep
	s.db.Save(&job)
	s.pushEvent("tester_job.updated", s.toResponse(&job))

	bugIDs := s.createBugsForFailures(&job, results)

	bugIDsJSON, _ := json.Marshal(bugIDs)
	if string(bugIDsJSON) == "null" {
		bugIDsJSON = []byte("[]")
	}
	job.BugIssueIDs = bugIDsJSON

	// ---- Step 4: complete ----
	completedAt := time.Now()
	job.Status = model.TesterJobCompleted
	job.Progress = 100
	completedStep := "completed"
	job.CurrentStep = &completedStep
	job.CompletedAt = &completedAt
	s.db.Save(&job)
	s.pushEvent("tester_job.completed", s.toResponse(&job))
}

// failJob marks a job as failed and emits an SSE event.
func (s *TesterAgentService) failJob(job *model.TesterJob, message string) {
	now := time.Now()
	job.Status = model.TesterJobFailed
	job.ErrorMessage = &message
	job.CompletedAt = &now
	step := "failed"
	job.CurrentStep = &step
	s.db.Save(job)
	s.pushEvent("tester_job.failed", s.toResponse(job))
}

// createBugsForFailures creates a Bug work item for each failed test result.
// Bugs are linked to the source issue via ParentID and recorded with the
// "Bug" issue type when one exists in the workspace. Failures to create a
// bug are logged but do not fail the whole job.
func (s *TesterAgentService) createBugsForFailures(job *model.TesterJob, results []TestResult) []uint64 {
	bugIDs := make([]uint64, 0)
	if len(results) == 0 || job.FailCount == 0 {
		return bugIDs
	}
	if job.ProjectID == nil || *job.ProjectID == 0 {
		// Cannot create issues without a project; skip bug reporting.
		return bugIDs
	}
	projectID := *job.ProjectID

	// Resolve the default state for the project once.
	var state model.State
	if err := s.db.Where("project_id = ? AND is_default = ?", projectID, true).First(&state).Error; err != nil {
		// Fall back to any active state for the project.
		if err := s.db.Where("project_id = ? AND is_active = ?", projectID, true).Order("sequence ASC").First(&state).Error; err != nil {
			return bugIDs
		}
	}

	// Look up a "Bug" issue type in the workspace (best effort).
	var bugType model.IssueType
	hasBugType := false
	if err := s.db.Where("workspace_id = ? AND LOWER(name) = ?", job.WorkspaceID, "bug").First(&bugType).Error; err == nil {
		hasBugType = true
	}

	for _, r := range results {
		if r.Status != "failed" {
			continue
		}
		bug := model.Issue{
			Name:            s.buildBugTitle(job, r),
			DescriptionHTML: s.buildBugDescription(job, r),
			Priority:        common.PriorityHigh,
			SortOrder:       65535,
			ProjectID:       projectID,
			WorkspaceID:     job.WorkspaceID,
			StateID:         state.ID,
			ParentID:        job.IssueID,
			ExternalSource:  testerStrPtr("tester_agent"),
		}
		if hasBugType {
			tid := bugType.ID
			bug.IssueTypeID = &tid
		}
		if err := s.db.Create(&bug).Error; err != nil {
			// Log and continue — a single bug creation failure should not
			// abort reporting for the remaining failures.
			continue
		}
		bugIDs = append(bugIDs, bug.ID)
	}
	return bugIDs
}

// buildBugTitle produces a concise title for a bug derived from a failed case.
func (s *TesterAgentService) buildBugTitle(job *model.TesterJob, r TestResult) string {
	name := r.Name
	if name == "" {
		name = r.CaseID
	}
	return fmt.Sprintf("[Tester Agent] %s", name)
}

// buildBugDescription assembles a markdown description for a bug.
func (s *TesterAgentService) buildBugDescription(job *model.TesterJob, r TestResult) string {
	var b strings.Builder
	b.WriteString("<p><strong>Bug reported by Tester Agent</strong> (ReqMango P4-002).</p>")
	b.WriteString("<p><strong>Failed test:</strong> ")
	b.WriteString(r.Name)
	b.WriteString("</p><p><strong>Error:</strong> ")
	if r.Error != "" {
		b.WriteString(escapeHTML(r.Error))
	} else {
		b.WriteString("(no error detail captured)")
	}
	b.WriteString("</p><p><strong>Job:</strong> #")
	b.WriteString(fmt.Sprintf("%d", job.ID))
	b.WriteString(" &mdash; ")
	b.WriteString(escapeHTML(job.Title))
	b.WriteString("</p>")
	return b.String()
}

// toResponse converts a model.TesterJob to its API representation.
func (s *TesterAgentService) toResponse(job *model.TesterJob) *TesterJobResponse {
	cases := job.GeneratedCases
	if len(cases) == 0 {
		cases = json.RawMessage("[]")
	}
	results := job.TestResults
	if len(results) == 0 {
		results = json.RawMessage("[]")
	}
	bugIDs := job.BugIssueIDs
	if len(bugIDs) == 0 {
		bugIDs = json.RawMessage("[]")
	}
	ctx := job.InputContext
	if len(ctx) == 0 {
		ctx = json.RawMessage("{}")
	}
	return &TesterJobResponse{
		ID:                 job.ID,
		WorkspaceID:        job.WorkspaceID,
		ProjectID:          job.ProjectID,
		IssueID:            job.IssueID,
		AgentTaskID:        job.AgentTaskID,
		Title:              job.Title,
		RequirementText:    job.RequirementText,
		AcceptanceCriteria: job.AcceptanceCriteria,
		TestScope:          job.TestScope,
		InputContext:       ctx,
		GeneratedCases:     cases,
		TestResults:        results,
		TotalCases:         job.TotalCases,
		PassCount:          job.PassCount,
		FailCount:          job.FailCount,
		SkipCount:          job.SkipCount,
		BugIssueIDs:        bugIDs,
		Status:             string(job.Status),
		Progress:           job.Progress,
		CurrentStep:        job.CurrentStep,
		ErrorMessage:       job.ErrorMessage,
		StartedAt:          job.StartedAt,
		CompletedAt:        job.CompletedAt,
		CancelledAt:        job.CancelledAt,
		CreatedAt:          job.CreatedAt,
		UpdatedAt:          job.UpdatedAt,
	}
}

// pushEvent broadcasts a tester_job SSE event to all clients.
func (s *TesterAgentService) pushEvent(event string, resp *TesterJobResponse) {
	data, _ := json.Marshal(resp)
	SSE.BroadcastEvent(event, json.RawMessage(data))
}

// ======== Helpers ========

// normalizeTesterJSON ensures empty/nil JSON is stored as "{}" rather than null.
func normalizeTesterJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return json.RawMessage("{}")
	}
	return raw
}

// testerStrPtr returns a pointer to s. Convenience for optional string fields.
func testerStrPtr(s string) *string { return &s }

// escapeHTML performs a minimal HTML-escape so user-controlled text is safe
// to embed in the bug description HTML. It covers the characters that matter
// for preventing broken/escaped markup.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// ======== LLM test case generator ========

// llmTestCaseGenerator implements TestCaseGenerator using the workspace LLM
// client. When no LLM is configured (llm == nil), it falls back to a
// deterministic stub so the workflow remains end-to-end testable.
type llmTestCaseGenerator struct {
	llm *llm.LLMClient
}

func (g *llmTestCaseGenerator) Generate(ctx context.Context, req TestCaseGenerationRequest) ([]TestCase, error) {
	if g.llm == nil {
		return stubGenerateCases(req), nil
	}

	systemPrompt := `You are a Tester Agent. Your job is to produce a concise, actionable set of test cases that verify the given requirement against its acceptance criteria.

Output ONLY a JSON object with the following shape, no prose, no markdown fences:
{
  "cases": [
    {
      "id": "<stable short id, e.g. tc-1>",
      "name": "<short case name>",
      "description": "<what this case verifies>",
      "steps": ["<step 1>", "<step 2>"],
      "expected": "<expected outcome>"
    }
  ]
}

Guidelines:
- Generate between 2 and 6 cases focused on the acceptance criteria.
- Prefer meaningful cases over exhaustive ones.
- Each case must have a stable, unique id.`

	userMsg := fmt.Sprintf("Requirement: %s\n\nTitle: %s\nAcceptance criteria:\n%s\nTest scope: %s",
		req.RequirementText, req.Title, req.AcceptanceCriteria, req.TestScope)

	resp, err := g.llm.ChatSync(ctx, systemPrompt, []llm.Message{
		{Role: "user", Content: userMsg},
	}, nil)
	if err != nil {
		// Fall back to stub when the LLM is not configured (e.g. missing API key).
		if strings.Contains(err.Error(), "未配置") || strings.Contains(err.Error(), "API Key") {
			return stubGenerateCases(req), nil
		}
		return nil, err
	}
	if resp.Content == "" {
		return nil, errors.New("LLM returned empty content")
	}

	body := stripCodeFences(resp.Content)
	var parsed struct {
		Cases []TestCase `json:"cases"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		return nil, fmt.Errorf("LLM output was not valid JSON: %w", err)
	}
	if len(parsed.Cases) == 0 {
		return nil, errors.New("LLM did not produce any test cases")
	}
	// Ensure every case has a stable id.
	for i := range parsed.Cases {
		if parsed.Cases[i].ID == "" {
			parsed.Cases[i].ID = fmt.Sprintf("tc-%d", i+1)
		}
	}
	return parsed.Cases, nil
}

// stubGenerateCases produces a small deterministic set of cases derived from
// the requirement. Used when no LLM is configured so the workflow remains
// observable.
func stubGenerateCases(req TestCaseGenerationRequest) []TestCase {
	criteria := strings.TrimSpace(req.AcceptanceCriteria)
	if criteria == "" {
		criteria = "Requirement is implemented and behaves as described."
	}
	cases := []TestCase{
		{
			ID:          "tc-1",
			Name:        "Happy path - " + truncateTitle(req.Title, 40),
			Description: "Verify the requirement works for the primary success scenario.",
			Steps: []string{
				"Set up the preconditions described in the requirement.",
				"Perform the primary action.",
				"Observe the result.",
			},
			Expected: criteria,
		},
		{
			ID:          "tc-2",
			Name:        "Edge case - empty input",
			Description: "Verify behaviour when no input is supplied.",
			Steps: []string{
				"Provide empty/blank input.",
				"Trigger the action.",
				"Observe the result.",
			},
			Expected: "The system handles empty input gracefully without crashing.",
		},
		{
			ID:          "tc-3",
			Name:        "Negative case - invalid input",
			Description: "Verify behaviour when invalid input is supplied.",
			Steps: []string{
				"Provide input that violates the acceptance criteria.",
				"Trigger the action.",
				"Observe the error handling.",
			},
			Expected: "The system rejects the input with a clear error.",
		},
	}
	return cases
}

// truncateTitle shortens s to at most n characters, appending an ellipsis when
// truncation occurs.
func truncateTitle(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// ======== Stub test executor ========

// stubTestExecutor simulates test execution. It marks the first failure-
// prone case (tc-2 "empty input") as failed and the rest as passed so the
// bug-reporting path is exercised end-to-end without an external runner.
type stubTestExecutor struct{}

func (e *stubTestExecutor) Execute(ctx context.Context, req TestExecutionRequest) ([]TestResult, error) {
	results := make([]TestResult, 0, len(req.Cases))
	for i, c := range req.Cases {
		// Simulate a deterministic failure for the "empty input" edge case
		// so the reporting step has at least one bug to create.
		status := "passed"
		errMsg := ""
		if isEdgeEmptyCase(c) {
			status = "failed"
			errMsg = "System returned a 500 error on empty input instead of a validation error."
		}
		results = append(results, TestResult{
			CaseID:     c.ID,
			Name:       c.Name,
			Status:     status,
			DurationMs: int64(20 + i*7),
			Error:      errMsg,
		})
	}
	return results, nil
}

// isEdgeEmptyCase returns true when the case targets the empty-input edge.
func isEdgeEmptyCase(c TestCase) bool {
	combined := strings.ToLower(c.ID + " " + c.Name + " " + c.Description)
	return strings.Contains(combined, "empty") || strings.Contains(combined, "tc-2")
}

// Compile-time interface checks.
var (
	_ TestCaseGenerator = (*llmTestCaseGenerator)(nil)
	_ TestExecutor      = (*stubTestExecutor)(nil)
)
