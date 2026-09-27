package service

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
)

// SkippedSuggestion explains why one proposal was left out of an apply.
//
// An agent reasons from a stale snapshot: between its reply and the click, the
// project's type list or workflow may have moved. Those proposals are dropped
// individually — otherwise one stale field would fail the whole batch and the
// button would appear broken — but they are reported rather than hidden, since
// the user still needs to make that change by hand.
type SkippedSuggestion struct {
	Field  string `json:"field"`
	Label  string `json:"label,omitempty"`
	Reason string `json:"reason"`
}

// ApplySuggestionsResult is returned after adopting an agent's proposals.
type ApplySuggestionsResult struct {
	Applied []string                `json:"applied"` // field names that were written
	Skipped []SkippedSuggestion     `json:"skipped"` // proposals that could not be written
	Issue   *response.IssueResponse `json:"issue"`
}

// ApplyIssueSuggestions adopts the structured suggestions an agent attached to
// one of its comments, reusing the normal issue update path so permissions,
// validation, activity trail and automations all behave exactly as a manual edit.
//
// selected restricts the apply to a subset of field names; an empty selection
// means "apply everything the agent proposed".
func (s *IssueService) ApplyIssueSuggestions(issueID, commentID, userID uint64, selected []string) (*ApplySuggestionsResult, error) {
	var comment model.Comment
	if err := s.db.First(&comment, commentID).Error; err != nil {
		return nil, common.NotFound("Comment not found")
	}
	if comment.IssueID != issueID {
		return nil, common.BadRequest("Comment does not belong to this work item")
	}
	if comment.SuggestionsAppliedAt != nil {
		return nil, common.Conflict("Suggestions on this comment have already been applied")
	}
	var suggestions []model.IssueSuggestion
	if len(comment.Suggestions) == 0 || json.Unmarshal(comment.Suggestions, &suggestions) != nil || len(suggestions) == 0 {
		return nil, common.BadRequest("This comment has no suggestions to apply")
	}

	// Needed to tell whether an id in the proposals is still usable here; the
	// earlier checks already ran, so a missing issue is still reported as such
	// by Update below.
	var target model.Issue
	if err := s.db.First(&target, issueID).Error; err != nil {
		return nil, common.NotFound("Work item not found")
	}

	only := make(map[string]bool, len(selected))
	for _, f := range selected {
		only[strings.TrimSpace(f)] = true
	}

	req := &request.IssueUpdateRequest{}
	applied := make([]string, 0, len(suggestions))
	skipped := make([]SkippedSuggestion, 0)
	for _, sg := range selectSuggestions(suggestions, only) {
		if reason := s.suggestionBlockedReason(sg, &target, userID); reason != "" {
			skipped = append(skipped, SkippedSuggestion{Field: sg.Field, Label: sg.Label, Reason: reason})
			continue
		}
		if !applySuggestionToRequest(req, sg) {
			skipped = append(skipped, SkippedSuggestion{
				Field: sg.Field, Label: sg.Label, Reason: "建议的值不可用",
			})
			continue
		}
		applied = append(applied, sg.Field)
	}
	if len(applied) == 0 {
		if len(skipped) > 0 {
			return nil, common.BadRequest(skipped[0].Reason)
		}
		return nil, common.BadRequest("None of the selected suggestions contain a usable value")
	}

	issue, err := s.Update(issueID, req, userID)
	if err != nil {
		return nil, err
	}

	// Mark applied so the action buttons disappear; the update is already
	// committed, so a failure here must not fail the request.
	now := time.Now()
	s.db.Model(&model.Comment{}).Where("id = ?", commentID).Update("suggestions_applied_at", now)

	return &ApplySuggestionsResult{Applied: applied, Skipped: skipped, Issue: issue}, nil
}

// suggestionBlockedReason reports why a proposal cannot be written right now, or
// "" when it is safe to include in the update. Both checks mirror a constraint
// the update path enforces anyway; running them here lets the batch survive a
// stale proposal instead of failing as a whole.
func (s *IssueService) suggestionBlockedReason(sg model.IssueSuggestion, issue *model.Issue, userID uint64) string {
	switch sg.Field {
	case "type":
		// Type ids are project-scoped: a workspace-level id the agent read is
		// rejected unless this project imported it.
		if !s.typeUsableForIssue(suggestionUint(sg.Value), issue) {
			return "该类型不在本项目可用类型中，请先导入或改用项目类型"
		}
	case "state":
		targetStateID := suggestionUint(sg.Value)
		if targetStateID == 0 || targetStateID == issue.StateID {
			return ""
		}
		switch err := s.validateStateTransition(s.db, issue.ProjectID, issue.ID, issue.StateID, targetStateID, userID); {
		case err == nil:
			return ""
		default:
			if appErr, ok := err.(*common.AppError); ok {
				return appErr.Message
			}
			return "该状态流转不被当前工作流允许"
		}
	}
	return ""
}

// typeUsableForIssue reports whether the type id can be written on this issue.
// Type ids are project-scoped: the workspace-level id an agent may have read is
// rejected unless the project imported it.
func (s *IssueService) typeUsableForIssue(typeID uint64, issue *model.Issue) bool {
	if typeID == 0 {
		return false
	}
	return s.validateTypeVisibleInProject(typeID, issue.ProjectID, issue.WorkspaceID) == nil
}

// selectSuggestions narrows the agent's proposals down to the ones this apply
// should act on. only restricts to an explicit field selection (empty = all).
//
// A no-op proposal — the agent recommending the current value be kept — is
// carried for its reasoning alone, so "apply all" must not write it. Naming it
// explicitly still wins, which keeps the client's "apply this single row" path
// honest if a row is ever rendered for a no-op.
func selectSuggestions(suggestions []model.IssueSuggestion, only map[string]bool) []model.IssueSuggestion {
	out := make([]model.IssueSuggestion, 0, len(suggestions))
	for _, sg := range suggestions {
		if len(only) > 0 && !only[sg.Field] {
			continue
		}
		if sg.NoOp && !only[sg.Field] {
			continue
		}
		out = append(out, sg)
	}
	return out
}

// applySuggestionToRequest folds one suggestion into the update request.
// It reports false when the field is unknown or its value is unusable, so the
// caller can drop it instead of failing the whole apply.
func applySuggestionToRequest(req *request.IssueUpdateRequest, sg model.IssueSuggestion) bool {
	switch sg.Field {
	case "title":
		v, ok := sg.Value.(string)
		if !ok || strings.TrimSpace(v) == "" {
			return false
		}
		v = strings.TrimSpace(v)
		req.Name = &v
		return true

	case "description":
		v, ok := sg.Value.(string)
		if !ok || strings.TrimSpace(v) == "" {
			return false
		}
		req.DescriptionHTML = &v
		return true

	case "priority":
		v, ok := sg.Value.(string)
		if !ok || strings.TrimSpace(v) == "" {
			return false
		}
		v = strings.ToLower(strings.TrimSpace(v))
		req.Priority = &v
		return true

	case "type":
		id := suggestionUint(sg.Value)
		if id == 0 {
			return false
		}
		req.TypeID = &id
		return true

	case "state":
		id := suggestionUint(sg.Value)
		if id == 0 {
			return false
		}
		req.StateID = &id
		return true

	case "assignee":
		id := suggestionUint(sg.Value)
		if id == 0 {
			return false
		}
		// AssigneeIDs replaces the whole assignee set, which matches the
		// "改派给 X" wording agents use for this suggestion.
		req.AssigneeIDs = []uint64{id}
		return true
	}
	return false
}

// suggestionUint accepts the numeric shapes JSON unmarshalling can produce for
// an id: float64 from a JSON number, or a numeric string.
func suggestionUint(value interface{}) uint64 {
	switch v := value.(type) {
	case float64:
		if v > 0 {
			return uint64(v)
		}
	case int:
		if v > 0 {
			return uint64(v)
		}
	case uint64:
		return v
	case string:
		var id uint64
		for _, r := range v {
			if r < '0' || r > '9' {
				return 0
			}
			id = id*10 + uint64(r-'0')
		}
		return id
	}
	return 0
}
