package service

import (
	"fmt"

	"github.com/reqmango/backend/internal/model"
)

// checkToolScope rejects ids the model supplied that point outside the
// caller's workspace or at a project the caller is not a member of. Tools
// trust their arguments, so without this a prompt can read any tenant's data,
// and a guessed id (e.g. workspace_id passed as project_id) silently returns
// empty results instead of an error the model can correct.
//
// Ids equal to the context's own are always allowed; those come from the
// server, not the model.
func (s *AIService) checkToolScope(args map[string]interface{}, actx *AIContext) error {
	if actx == nil || actx.WorkspaceID == 0 {
		return nil
	}
	outOfScope := func(key string, id uint64) error {
		return fmt.Errorf("%s=%d 超出当前范围（当前 workspace_id=%d, project_id=%d），请使用当前上下文中的 ID 或先用工具查询",
			key, id, actx.WorkspaceID, actx.ProjectID)
	}

	if wid := getUintArg(args, "workspace_id", 0); wid != 0 && wid != actx.WorkspaceID {
		return outOfScope("workspace_id", wid)
	}
	if pid := getUintArg(args, "project_id", 0); pid != 0 && pid != actx.ProjectID && !s.projectInScope(pid, actx) {
		return outOfScope("project_id", pid)
	}
	if iid := getUintArg(args, "issue_id", 0); iid != 0 && iid != actx.IssueID {
		var issue model.Issue
		if err := s.db.Select("id", "project_id").First(&issue, iid).Error; err == nil &&
			issue.ProjectID != actx.ProjectID && !s.projectInScope(issue.ProjectID, actx) {
			return outOfScope("issue_id", iid)
		}
	}
	if cid := getUintArg(args, "cycle_id", 0); cid != 0 {
		var cycle model.Cycle
		if err := s.db.Select("id", "project_id").First(&cycle, cid).Error; err == nil &&
			cycle.ProjectID != actx.ProjectID && !s.projectInScope(cycle.ProjectID, actx) {
			return outOfScope("cycle_id", cid)
		}
	}
	return nil
}

// projectInScope reports whether a project other than the context's own is
// one the caller may read: same workspace, and a project or workspace
// membership (the HTTP layer grants access by the higher of the two roles).
func (s *AIService) projectInScope(projectID uint64, actx *AIContext) bool {
	var project model.Project
	if err := s.db.Select("id", "workspace_id").First(&project, projectID).Error; err != nil {
		return false
	}
	if project.WorkspaceID != actx.WorkspaceID {
		return false
	}
	if s.isProjectMember(projectID, actx.UserID) {
		return true
	}
	var count int64
	s.db.Model(&model.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ? AND is_active = ?", actx.WorkspaceID, actx.UserID, true).
		Count(&count)
	return count > 0
}
