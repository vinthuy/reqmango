package service

import (
	"encoding/json"
	"sort"

	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
)

type suggestionCommentRow struct {
	AgentID            uint64
	AgentName          string
	Suggestions        []byte
	SuggestionOutcomes []byte
	Settled            bool
}

// SuggestionAcceptance reports, per agent, how users acted on the proposals
// attached to its comments in the workspace. Period bounds filter on when the
// comment was written.
func (s *AgentPerformanceService) SuggestionAcceptance(workspaceID uint64, from, to string) ([]response.AgentSuggestionStats, error) {
	fromT, toT, err := parsePeriodRange(from, to)
	if err != nil {
		return nil, err
	}
	q := s.db.Table("comments AS c").
		Select("c.agent_id AS agent_id, COALESCE(a.name, '') AS agent_name, c.suggestions AS suggestions, c.suggestion_outcomes AS suggestion_outcomes, c.suggestions_applied_at IS NOT NULL AS settled").
		Joins("JOIN issues i ON i.id = c.issue_id").
		Joins("LEFT JOIN agents a ON a.id = c.agent_id").
		Where("i.workspace_id = ? AND c.agent_id IS NOT NULL AND c.suggestions IS NOT NULL AND c.deleted_at IS NULL", workspaceID)
	q = applyPeriodFilter(q, fromT, toT, "c.")

	var rows []suggestionCommentRow
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	byAgent := map[uint64]*response.AgentSuggestionStats{}
	for _, r := range rows {
		var suggestions []model.IssueSuggestion
		if json.Unmarshal(r.Suggestions, &suggestions) != nil {
			continue
		}
		outcomes := DecodeSuggestionOutcomes(r.SuggestionOutcomes)
		// Settled before outcomes were recorded: which fields were adopted is unknown.
		if r.Settled && len(outcomes) == 0 {
			continue
		}
		st := byAgent[r.AgentID]
		if st == nil {
			st = &response.AgentSuggestionStats{AgentID: r.AgentID, AgentName: r.AgentName, ByField: map[string]response.SuggestionCounts{}}
			byAgent[r.AgentID] = st
		}
		for _, sg := range suggestions {
			if sg.NoOp {
				continue
			}
			field := st.ByField[sg.Field]
			tallySuggestion(&st.SuggestionCounts, outcomes[sg.Field].Status)
			tallySuggestion(&field, outcomes[sg.Field].Status)
			st.ByField[sg.Field] = field
		}
	}

	out := make([]response.AgentSuggestionStats, 0, len(byAgent))
	for _, st := range byAgent {
		if st.Proposed == 0 {
			continue
		}
		st.AcceptanceRate = acceptanceRate(st.SuggestionCounts)
		for f, c := range st.ByField {
			c.AcceptanceRate = acceptanceRate(c)
			st.ByField[f] = c
		}
		out = append(out, *st)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Proposed > out[j].Proposed })
	return out, nil
}

func tallySuggestion(c *response.SuggestionCounts, status string) {
	c.Proposed++
	switch status {
	case model.SuggestionApplied:
		c.Applied++
	case model.SuggestionDismissed:
		c.Dismissed++
	case model.SuggestionFailed:
		c.Failed++
	default:
		c.Pending++
	}
}

func acceptanceRate(c response.SuggestionCounts) float64 {
	decided := c.Applied + c.Dismissed
	if decided == 0 {
		return 0
	}
	return round2(float64(c.Applied) / float64(decided) * 100)
}
