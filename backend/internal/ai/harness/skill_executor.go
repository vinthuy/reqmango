package harness

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// SkillExecutor executes SKILL.md formatted skills.
type SkillExecutor struct {
	db           *gorm.DB
	toolExecutor ToolExecutor
}

// ToolExecutor defines the interface for executing tools.
type ToolExecutor interface {
	Execute(ctx context.Context, toolName string, input json.RawMessage) (interface{}, error)
}

func NewSkillExecutor(db *gorm.DB, toolExecutor ToolExecutor) *SkillExecutor {
	return &SkillExecutor{
		db:           db,
		toolExecutor: toolExecutor,
	}
}

// SkillStep represents a step in a skill execution.
type SkillStep struct {
	Step   int                    `json:"step"`
	Action string                 `json:"action"`
	Tool   string                 `json:"tool,omitempty"`
	Input  map[string]interface{} `json:"input,omitempty"`
	Output interface{}            `json:"output,omitempty"`
	Error  string                 `json:"error,omitempty"`
	Status string                 `json:"status"` // "pending", "running", "completed", "failed"
}

// ExecuteSkillResult contains the result of skill execution.
type ExecuteSkillResult struct {
	SkillID     uint64      `json:"skill_id"`
	SkillName   string      `json:"skill_name"`
	Steps       []SkillStep `json:"steps"`
	FinalResult string      `json:"final_result"`
	Error       string      `json:"error,omitempty"`
	TokensUsed  int         `json:"tokens_used"`
}

// ParseSkillMD parses a SKILL.md string into structured steps.
func ParseSkillMD(skillMD string) ([]SkillStep, error) {
	var steps []SkillStep

	// Split by step markers (## Step X or ### Step X). Go's RE2 engine has no
	// lookahead support, so the headings are located first and each step body is
	// sliced out up to the start of the next heading.
	headingRegex := regexp.MustCompile(`(?m)^#{2,3}[ \t]*Step[ \t]*(\d+)[^\n]*$`)
	headings := headingRegex.FindAllStringSubmatchIndex(skillMD, -1)

	for i, heading := range headings {
		stepNum, _ := strconv.Atoi(skillMD[heading[2]:heading[3]])

		contentEnd := len(skillMD)
		if i+1 < len(headings) {
			contentEnd = headings[i+1][0]
		}
		content := strings.TrimSpace(skillMD[heading[1]:contentEnd])

		step := SkillStep{
			Step:   stepNum,
			Status: "pending",
		}

		// Extract tool calls
		// The "\n**" / end-of-text terminator is consumed by a non-capturing
		// group instead of a lookahead, which leaves the captured input identical.
		toolRegex := regexp.MustCompile(`\*\*Tool:\*\*\s*(\w+)\s*\n\*\*Input:\*\*\s*([\s\S]*?)(?:\n\*\*|$)`)
		toolMatch := toolRegex.FindStringSubmatch(content)
		if len(toolMatch) > 0 {
			step.Tool = toolMatch[1]
			step.Action = content
			// Parse JSON input
			var input map[string]interface{}
			if err := json.Unmarshal([]byte(toolMatch[2]), &input); err == nil {
				step.Input = input
			}
		} else {
			step.Action = content
		}

		steps = append(steps, step)
	}

	if len(steps) == 0 {
		// If no steps found, treat the entire content as one step
		steps = append(steps, SkillStep{
			Step:   1,
			Action: skillMD,
			Status: "pending",
		})
	}

	return steps, nil
}

// Execute executes a skill with the given parameters.
func (e *SkillExecutor) Execute(ctx context.Context, skill *model.Skill, params map[string]interface{}) (*ExecuteSkillResult, error) {
	result := &ExecuteSkillResult{
		SkillID:   skill.ID,
		SkillName: skill.Name,
	}

	// Parse skill steps
	steps, err := ParseSkillMD(skill.SkillMD)
	if err != nil {
		return nil, fmt.Errorf("failed to parse skill: %w", err)
	}
	result.Steps = steps

	// Execute each step
	for i := range result.Steps {
		step := &result.Steps[i]
		step.Status = "running"

		// Merge skill parameters with step input
		if params != nil && step.Input != nil {
			for k, v := range params {
				if _, exists := step.Input[k]; !exists {
					step.Input[k] = v
				}
			}
		} else if params != nil && step.Input == nil {
			step.Input = params
		}

		if step.Tool != "" && e.toolExecutor != nil {
			// Execute tool call
			inputJSON, err := json.Marshal(step.Input)
			if err != nil {
				step.Error = fmt.Sprintf("Failed to marshal input: %v", err)
				step.Status = "failed"
				result.Error = step.Error
				break
			}

			output, err := e.toolExecutor.Execute(ctx, step.Tool, inputJSON)
			if err != nil {
				step.Error = err.Error()
				step.Status = "failed"
				result.Error = step.Error
				break
			}
			step.Output = output
		}

		step.Status = "completed"
	}

	// Generate final result summary
	if result.Error == "" {
		result.FinalResult = e.generateSummary(result.Steps)
	}

	return result, nil
}

// generateSummary generates a human-readable summary of the execution.
func (e *SkillExecutor) generateSummary(steps []SkillStep) string {
	var sb strings.Builder
	sb.WriteString("Skill execution completed successfully.\n\n")
	for _, step := range steps {
		sb.WriteString(fmt.Sprintf("Step %d: %s\n", step.Step, step.Action[:min(len(step.Action), 100)]))
		if step.Output != nil {
			outputJSON, _ := json.Marshal(step.Output)
			sb.WriteString(fmt.Sprintf("  Output: %s\n", string(outputJSON)[:min(len(string(outputJSON)), 200)]))
		}
	}
	return sb.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
