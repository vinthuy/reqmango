package harness

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// PipelineDSL is the top-level YAML structure for pipeline definitions.
type PipelineDSL struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Version     string            `yaml:"version,omitempty"`
	Trigger     TriggerDSL        `yaml:"trigger,omitempty"`
	Pipeline    PipelineStagesDSL `yaml:"pipeline"`
	Retry       RetryDSL          `yaml:"retry,omitempty"`
	Budget      BudgetDSL         `yaml:"budget,omitempty"`
}

// TriggerDSL defines how the pipeline is activated.
type TriggerDSL struct {
	Type     string `yaml:"type"`               // cron, event, manual, webhook
	Schedule string `yaml:"schedule,omitempty"` // cron expression
	Event    string `yaml:"event,omitempty"`    // event type
}

// PipelineStagesDSL defines the stages of the pipeline.
type PipelineStagesDSL struct {
	Mode   string     `yaml:"mode,omitempty"` // sequential, fan_out, tournament, classify
	Stages []StageDSL `yaml:"stages"`
}

// StageDSL is a single stage in the YAML definition.
type StageDSL struct {
	Name       string      `yaml:"name"`
	Agent      string      `yaml:"agent"`
	StageType  string      `yaml:"type"` // planner, executor, reviewer, judge
	Model      string      `yaml:"model,omitempty"`
	Effort     string      `yaml:"effort,omitempty"`
	Mode       string      `yaml:"mode,omitempty"` // "adversarial" for reviewer stages
	InputFrom  string      `yaml:"input_from,omitempty"`
	OnComplete []ActionDSL `yaml:"on_complete,omitempty"`
}

// ActionDSL defines a post-stage action.
type ActionDSL struct {
	Action string                 `yaml:"action"`
	Params map[string]interface{} `yaml:"params,omitempty"`
}

// RetryDSL defines retry behavior.
type RetryDSL struct {
	MaxAttempts int    `yaml:"max_attempts,omitempty"`
	Backoff     string `yaml:"backoff,omitempty"` // exponential, linear
}

// BudgetDSL defines pipeline budget constraints.
type BudgetDSL struct {
	MaxTokens   int    `yaml:"max_tokens,omitempty"`
	OnExhausted string `yaml:"on_exhausted,omitempty"` // escalate_to_human, stop
}

// ParsePipelineDSL parses a YAML byte slice into a PipelineDSL.
func ParsePipelineDSL(yamlBytes []byte) (*PipelineDSL, error) {
	var dsl PipelineDSL
	if err := yaml.Unmarshal(yamlBytes, &dsl); err != nil {
		return nil, fmt.Errorf("failed to parse pipeline DSL: %w", err)
	}
	if err := dsl.Validate(); err != nil {
		return nil, err
	}
	return &dsl, nil
}

// Validate checks the DSL for required fields and valid values.
func (d *PipelineDSL) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("pipeline name is required")
	}
	if len(d.Pipeline.Stages) == 0 {
		return fmt.Errorf("pipeline must have at least one stage")
	}

	validTypes := map[string]bool{"planner": true, "executor": true, "reviewer": true, "judge": true}
	for _, stage := range d.Pipeline.Stages {
		if stage.Name == "" {
			return fmt.Errorf("stage name is required")
		}
		if !validTypes[stage.StageType] && stage.StageType != "" {
			return fmt.Errorf("invalid stage type '%s' for stage '%s'", stage.StageType, stage.Name)
		}
	}

	validModes := map[string]bool{"sequential": true, "fan_out": true, "tournament": true, "classify": true, "": true}
	if !validModes[d.Pipeline.Mode] {
		return fmt.Errorf("invalid pipeline mode: %s", d.Pipeline.Mode)
	}

	return nil
}

// ToPipelineConfig converts the DSL to a PipelineConfig for execution.
func (d *PipelineDSL) ToPipelineConfig() (*PipelineConfig, error) {
	mode := PipelineMode(d.Pipeline.Mode)
	if mode == "" {
		mode = ModeSequential
	}

	stages := make([]StageConfig, len(d.Pipeline.Stages))
	for i, s := range d.Pipeline.Stages {
		stages[i] = StageConfig{
			Name:        s.Name,
			StageType:   StageType(s.StageType),
			Model:       s.Model,
			Effort:      s.Effort,
			Adversarial: s.Mode == "adversarial",
			InputFrom:   s.InputFrom,
		}
	}

	maxRetries := d.Retry.MaxAttempts
	if maxRetries == 0 {
		maxRetries = 3
	}

	return &PipelineConfig{
		Name:   d.Name,
		Mode:   mode,
		Stages: stages,
		Retry: BudgetConfig{
			MaxTokens:  d.Budget.MaxTokens,
			MaxRetries: maxRetries,
		},
	}, nil
}
