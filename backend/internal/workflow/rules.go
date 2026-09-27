// Package workflow resolves the state transitions a project's active workflows
// allow for a given issue type.
//
// It exists because the rule has more than one reader: the update path must
// reject illegal moves, and the AI tools must tell an agent which moves are
// legal so it stops proposing illegal ones. Two copies of the rule drifted
// apart once already, so both now read it from here.
package workflow

import (
	"encoding/json"
	"strings"

	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// Outcome is the decision for a single from→to move.
type Outcome int

const (
	// Denied means no governing workflow permits the move.
	Denied Outcome = iota
	// Allowed means a workflow permits the move outright.
	Allowed
	// ApprovalRequired means the move is permitted only via an approval.
	ApprovalRequired
)

// Target is one legal destination state.
type Target struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	// RequiresApproval is set when the move is configured to go through approval
	// rather than applying immediately.
	RequiresApproval bool `json:"requires_approval,omitempty"`
}

// Verdict is the decision for one from→to move, carrying the edge that decided
// it so callers can build an approval request without re-querying.
type Verdict struct {
	Outcome      Outcome
	TransitionID uint64
	WorkflowID   uint64
	WorkflowName string
}

// edge is one configured transition, with its endpoints resolved so a rule
// written for a state in one project also governs an equivalent state in
// another (workflows are often shared across a workspace).
type edge struct {
	transition       model.StateTransition
	workflow         model.Workflow
	source           model.State
	target           model.State
	requiresApproval bool
}

// Rules is the transition rule set governing one issue type in one project.
type Rules struct {
	// Restricts reports whether any governing workflow actually defines
	// transitions. When false the project constrains nothing and every move is
	// allowed, which is also the case before a project configures a workflow.
	Restricts bool
	edges     []edge
}

// Load reads the active workflows governing projectID, scoped to issueTypeID
// (nil meaning an untyped issue), together with their transitions.
func Load(db *gorm.DB, projectID uint64, issueTypeID *uint64) (*Rules, error) {
	// Workflows are shared per workspace, so a project inherits the
	// workspace-level ones in addition to its own.
	var workspaceID uint64
	if err := db.Raw("SELECT workspace_id FROM projects WHERE id = ?", projectID).Scan(&workspaceID).Error; err != nil {
		return nil, err
	}

	var candidates []model.Workflow
	// A superset of the governing workflows: everything for this project plus
	// the workspace-level defaults. Issue-type scoping is applied in Go below
	// because issue_type_ids is a JSON list, and evaluating membership in SQL
	// pinned this rule to Postgres.
	if err := db.Where("is_active = ?", true).
		Where("(project_id = ?) OR (workspace_id = ? AND project_id IS NULL)", projectID, workspaceID).
		Find(&candidates).Error; err != nil {
		return nil, err
	}

	rules := &Rules{}
	for _, wf := range candidates {
		if !governsIssueType(wf, issueTypeID) {
			continue
		}
		var transitions []model.StateTransition
		if err := db.Where("workflow_id = ?", wf.ID).Find(&transitions).Error; err != nil {
			return nil, err
		}
		if len(transitions) > 0 {
			rules.Restricts = true
		}
		for _, tr := range transitions {
			// Rule types other than allow/approval are not yet implemented; they
			// constrain nothing either way.
			requiresApproval := tr.RuleType == "approval"
			if tr.RuleType != "allow" && tr.RuleType != "" && !requiresApproval {
				continue
			}
			src, ok := stateByID(db, tr.SourceStateID)
			if !ok {
				continue
			}
			tgt, ok := stateByID(db, tr.TargetStateID)
			if !ok {
				continue
			}
			rules.edges = append(rules.edges, edge{
				transition:       tr,
				workflow:         wf,
				source:           src,
				target:           tgt,
				requiresApproval: requiresApproval,
			})
		}
	}
	return rules, nil
}

// Check decides whether from may move to to. The precedence mirrors the update
// path exactly: an approval edge beats an allow edge for the same pair, and a
// project that defines no transitions at all constrains nothing.
func (r *Rules) Check(from, to model.State) Verdict {
	var approval *edge
	matchedAllow := false
	for i := range r.edges {
		e := &r.edges[i]
		if !statesEquivalent(&e.source, &from) || !statesEquivalent(&e.target, &to) {
			continue
		}
		if e.requiresApproval {
			approval = e
			continue
		}
		matchedAllow = true
	}
	if approval != nil {
		return Verdict{
			Outcome:      ApprovalRequired,
			TransitionID: approval.transition.ID,
			WorkflowID:   approval.workflow.ID,
			WorkflowName: approval.workflow.Name,
		}
	}
	if matchedAllow || !r.Restricts {
		return Verdict{Outcome: Allowed}
	}
	return Verdict{Outcome: Denied}
}

// Targets lists the states `from` may move to, drawn from the given project
// states so the result only ever names states that exist here. Ordering follows
// `states`, and an unrestricted project offers every other state.
func (r *Rules) Targets(from model.State, states []model.State) []Target {
	targets := make([]Target, 0, len(states))
	for i := range states {
		to := states[i]
		if to.ID == from.ID {
			continue
		}
		verdict := r.Check(from, to)
		if verdict.Outcome == Denied {
			continue
		}
		targets = append(targets, Target{
			ID:               to.ID,
			Name:             to.Name,
			RequiresApproval: verdict.Outcome == ApprovalRequired,
		})
	}
	return targets
}

// governsIssueType reports whether a workflow applies to an issue of this type,
// mirroring what the SQL used to express: a workflow with no type applies to
// everything, one with a type applies to that type, and the issue_type_ids list
// names further types it covers.
func governsIssueType(wf model.Workflow, issueTypeID *uint64) bool {
	if issueTypeID == nil {
		// An untyped issue only sees workflows that are not type-specific.
		return wf.IssueTypeID == nil
	}
	if wf.IssueTypeID == nil {
		return true
	}
	if *wf.IssueTypeID == *issueTypeID {
		return true
	}
	return containsTypeID(wf.IssueTypeIDs, *issueTypeID)
}

// containsTypeID reports whether a JSON list of type ids includes id.
func containsTypeID(raw json.RawMessage, id uint64) bool {
	if len(raw) == 0 || string(raw) == "null" {
		return false
	}
	var ids []uint64
	if err := json.Unmarshal(raw, &ids); err != nil {
		return false
	}
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

func stateByID(db *gorm.DB, id uint64) (model.State, bool) {
	var st model.State
	if err := db.First(&st, id).Error; err != nil {
		return model.State{}, false
	}
	return st, true
}

// statesEquivalent reports whether two states denote the same step. Workflows
// are shared across a workspace, so an edge written against one project's
// "Backlog" also governs another project's state with the same name.
func statesEquivalent(a, b *model.State) bool {
	if a == nil || b == nil {
		return false
	}
	if a.ID != 0 && a.ID == b.ID {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(a.Name), strings.TrimSpace(b.Name))
}
