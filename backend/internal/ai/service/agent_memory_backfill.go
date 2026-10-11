package service

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	reIssueIDEquals = regexp.MustCompile(`(?i)issue[_ ]?id\s*[=:]\s*(\d+)`)
	reIssueKey      = regexp.MustCompile(`\b([A-Za-z][A-Za-z0-9]{0,9})-(\d+)\b`)
	reHashSeq       = regexp.MustCompile(`(?:^|[^\w-])#(\d+)\b`)
)

// memoryIssueHint is a clue mined from an old agent-task memory that never
// recorded issue_id.
type memoryIssueHint struct {
	IssueID    uint64
	SequenceID uint64
	Identifier string
}

// extractMemoryIssueHint pulls the strongest issue reference from memory text.
// Explicit issue_id wins; otherwise a PROJECT-SEQ key; finally a bare #SEQ.
func extractMemoryIssueHint(content string) memoryIssueHint {
	if m := reIssueIDEquals.FindStringSubmatch(content); len(m) == 2 {
		if id, err := strconv.ParseUint(m[1], 10, 64); err == nil && id > 0 {
			return memoryIssueHint{IssueID: id}
		}
	}
	if keys := reIssueKey.FindAllStringSubmatch(content, -1); len(keys) > 0 {
		// Prefer the last key — agents often cite duplicates then the subject.
		last := keys[len(keys)-1]
		seq, _ := strconv.ParseUint(last[2], 10, 64)
		return memoryIssueHint{Identifier: strings.ToUpper(last[1]), SequenceID: seq}
	}
	if m := reHashSeq.FindStringSubmatch(content); len(m) == 2 {
		seq, _ := strconv.ParseUint(m[1], 10, 64)
		return memoryIssueHint{SequenceID: seq}
	}
	return memoryIssueHint{}
}

const (
	memoryBackfillBatch   = 200
	memoryOrphanExpireAge = 30 * 24 * time.Hour
)

// BackfillAgentMemoryIssueIDs repairs agent task memories that were saved
// before issue_id was recorded. Resolvable rows get issue_id; unresolvable
// rows older than 30 days get expires_at so they stop polluting retrieval.
// workspaceID 0 covers every workspace.
func (s *AgentService) BackfillAgentMemoryIssueIDs(workspaceID uint64) (filled, expired int, err error) {
	now := time.Now()
	var lastID uint64
	for {
		var entries []model.MemoryEntry
		q := s.db.Where("agent_id IS NOT NULL AND issue_id IS NULL AND deleted_at IS NULL").
			Where("expires_at IS NULL OR expires_at > ?", now).
			Where("id > ?", lastID).
			Order("id ASC").Limit(memoryBackfillBatch)
		if workspaceID > 0 {
			q = q.Where("workspace_id = ?", workspaceID)
		}
		if err = q.Find(&entries).Error; err != nil {
			return filled, expired, err
		}

		for i := range entries {
			e := &entries[i]
			lastID = e.ID
			hint := extractMemoryIssueHint(e.Content)
			if id := s.resolveMemoryIssueHint(e, hint); id > 0 {
				if err := s.db.Model(e).Update("issue_id", id).Error; err != nil {
					return filled, expired, err
				}
				filled++
				continue
			}
			if e.CreatedAt.IsZero() || now.Sub(e.CreatedAt) < memoryOrphanExpireAge {
				continue
			}
			if err := s.db.Model(e).Update("expires_at", now).Error; err != nil {
				return filled, expired, err
			}
			expired++
		}
		if len(entries) < memoryBackfillBatch {
			return filled, expired, nil
		}
	}
}

const memoryIssueIDBackfillMigration = "agent_memory_issue_id_backfill_v1"

// RunAgentMemoryBackfillOnce runs the issue_id backfill across all workspaces
// the first time it is called against a database. The marker is claimed before
// running so concurrent instances do not both run it, and released on failure
// so the next startup retries.
func RunAgentMemoryBackfillOnce(db *gorm.DB) (ran bool, filled, expired int, err error) {
	claim := db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&model.DataMigration{Name: memoryIssueIDBackfillMigration, AppliedAt: time.Now()})
	if claim.Error != nil {
		return false, 0, 0, claim.Error
	}
	if claim.RowsAffected == 0 {
		return false, 0, 0, nil
	}
	filled, expired, err = (&AgentService{db: db}).BackfillAgentMemoryIssueIDs(0)
	if err != nil {
		db.Where("name = ?", memoryIssueIDBackfillMigration).Delete(&model.DataMigration{})
		return true, filled, expired, err
	}
	return true, filled, expired, nil
}

func (s *AgentService) resolveMemoryIssueHint(e *model.MemoryEntry, hint memoryIssueHint) uint64 {
	if hint.IssueID > 0 {
		var issue model.Issue
		q := s.db.Select("id").Where("id = ?", hint.IssueID)
		if e.WorkspaceID > 0 {
			q = q.Where("workspace_id = ?", e.WorkspaceID)
		}
		if err := q.First(&issue).Error; err == nil {
			return issue.ID
		}
	}
	if hint.SequenceID == 0 {
		return 0
	}
	var issue model.Issue
	q := s.db.Select("issues.id").Joins("JOIN projects ON projects.id = issues.project_id").
		Where("issues.sequence_id = ?", hint.SequenceID)
	if e.ProjectID != nil && *e.ProjectID > 0 {
		q = q.Where("issues.project_id = ?", *e.ProjectID)
	} else if e.WorkspaceID > 0 {
		q = q.Where("issues.workspace_id = ?", e.WorkspaceID)
	}
	if hint.Identifier != "" {
		q = q.Where("UPPER(projects.identifier) = ?", hint.Identifier)
	}
	if err := q.First(&issue).Error; err != nil {
		return 0
	}
	return issue.ID
}
