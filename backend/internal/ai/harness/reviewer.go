package harness

import (
	"context"
	"fmt"
	"sync"
)

// ReviewVerdict is the output of a single reviewer agent.
type ReviewVerdict struct {
	Pass       bool          `json:"pass"`
	Issues     []ReviewIssue `json:"issues"`
	Confidence float64       `json:"confidence"`
	ReviewerID int           `json:"reviewer_id"`
}

// ReviewIssue describes a problem found during review.
type ReviewIssue struct {
	Severity    string `json:"severity"` // critical, major, minor
	Description string `json:"description"`
}

// AdversarialReviewer runs the 3-vote adversarial verification protocol.
type AdversarialReviewer struct {
	caller  AgentCaller
	agentID uint64
	model   string
}

// NewAdversarialReviewer creates a new adversarial reviewer.
func NewAdversarialReviewer(caller AgentCaller, agentID uint64, model string) *AdversarialReviewer {
	return &AdversarialReviewer{caller: caller, agentID: agentID, model: model}
}

// Review runs 3 independent reviewers against the artifact. Returns the verdicts and whether the artifact passes.
func (r *AdversarialReviewer) Review(ctx context.Context, spec string, artifact string) ([]ReviewVerdict, bool, error) {
	const numReviewers = 3
	var wg sync.WaitGroup
	var mu sync.Mutex
	verdicts := make([]ReviewVerdict, numReviewers)

	for i := 0; i < numReviewers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			prompt := buildRefutePrompt(spec, artifact)
			result, _, _, err := r.caller.CallAgent(ctx, r.agentID, r.model, "", prompt, nil)
			verdict := ReviewVerdict{ReviewerID: idx + 1}
			if err != nil {
				verdict.Pass = false
				verdict.Issues = []ReviewIssue{{Severity: "critical", Description: fmt.Sprintf("reviewer error: %v", err)}}
				verdict.Confidence = 0
			} else {
				verdict = parseVerdict(result, idx+1)
			}
			mu.Lock()
			verdicts[idx] = verdict
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	// Count passes -- need >= 2 to approve
	passes := 0
	for _, v := range verdicts {
		if v.Pass {
			passes++
		}
	}

	return verdicts, passes >= 2, nil
}

// buildRefutePrompt creates the adversarial review prompt.
func buildRefutePrompt(spec string, artifact string) string {
	return fmt.Sprintf(`You are a QA Reviewer. Your job is to find problems.

CRITICAL INSTRUCTION: Assume the following output contains errors.
Your task is to FIND them. Default to REFUTED if you are uncertain.

=== SPECIFICATION ===
%s

=== OUTPUT TO REVIEW ===
%s

=== REVIEW CHECKLIST ===
1. Does the output fully satisfy the spec?
2. Are there logical errors or omissions?
3. Are there any edge cases not handled?
4. Is the quality acceptable for production?
5. Are there security or performance concerns?

Respond in JSON format:
{
  "pass": true/false,
  "issues": [{"severity": "critical|major|minor", "description": "..."}],
  "confidence": 0.0-1.0
}`, spec, artifact)
}

// parseVerdict extracts a ReviewVerdict from an LLM response.
func parseVerdict(response string, reviewerID int) ReviewVerdict {
	// Simple heuristic: look for "pass": true/false in the response
	// In production, use proper JSON parsing
	v := ReviewVerdict{ReviewerID: reviewerID, Confidence: 0.5}

	// Try to find pass/fail indicator
	passIndicators := []string{"\"pass\": true", "\"pass\":true", "PASS", "pass: true"}
	failIndicators := []string{"\"pass\": false", "\"pass\":false", "FAIL", "pass: false"}

	for _, ind := range passIndicators {
		if contains(response, ind) {
			v.Pass = true
			break
		}
	}
	for _, ind := range failIndicators {
		if contains(response, ind) {
			v.Pass = false
			break
		}
	}

	// Default: if no clear indicator, treat as fail (refute-first principle)
	if !v.Pass && !containsAny(response, passIndicators) && !containsAny(response, failIndicators) {
		v.Pass = false
		v.Issues = []ReviewIssue{{Severity: "major", Description: "Reviewer could not determine pass/fail -- defaulting to refuted"}}
	}

	return v
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if contains(s, sub) {
			return true
		}
	}
	return false
}
