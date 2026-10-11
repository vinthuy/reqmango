package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/reqmango/backend/internal/model"
)

const declinedSuggestionsLimit = 10

// declinedSuggestionsNote lists proposals users already declined (or that the
// project refused) on this issue, newest first, so an agent does not keep
// re-proposing them. Returns "" when there are none.
func (s *AgentService) declinedSuggestionsNote(issueID uint64) string {
	if s.db == nil {
		return ""
	}
	var comments []model.Comment
	if err := s.db.Select("id", "suggestions", "suggestion_outcomes").
		Where("issue_id = ? AND agent_id IS NOT NULL AND suggestion_outcomes IS NOT NULL", issueID).
		Order("id DESC").Limit(20).Find(&comments).Error; err != nil {
		return ""
	}

	lines := make([]string, 0, declinedSuggestionsLimit)
	seen := map[string]bool{}
	for _, c := range comments {
		var suggestions []model.IssueSuggestion
		outcomes := map[string]model.SuggestionOutcome{}
		if json.Unmarshal(c.Suggestions, &suggestions) != nil || json.Unmarshal(c.SuggestionOutcomes, &outcomes) != nil {
			continue
		}
		for _, sg := range suggestions {
			o, ok := outcomes[sg.Field]
			if !ok || (o.Status != model.SuggestionDismissed && o.Status != model.SuggestionFailed) {
				continue
			}
			target := sg.Label
			if target == "" {
				target = fmt.Sprint(sg.Value)
			}
			key := sg.Field + "\x00" + target
			if seen[key] {
				continue
			}
			seen[key] = true
			line := fmt.Sprintf("- %s → %s", sg.Field, target)
			if o.Status == model.SuggestionDismissed {
				line += "（用户已忽略）"
			} else {
				line += fmt.Sprintf("（无法应用：%s）", o.Reason)
			}
			lines = append(lines, line)
			if len(lines) == declinedSuggestionsLimit {
				break
			}
		}
		if len(lines) == declinedSuggestionsLimit {
			break
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return "此前在该工作项上未被采纳的建议（除非情况已变化，否则不要重复提出；若重复提出，请说明新的理由）：\n" +
		strings.Join(lines, "\n") + "\n"
}
