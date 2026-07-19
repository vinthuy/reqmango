package harness

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// PipelineMode defines the orchestration strategy.
type PipelineMode string

const (
	ModeSequential PipelineMode = "sequential"
	ModeFanOut     PipelineMode = "fan_out"
	ModeTournament PipelineMode = "tournament"
	ModeClassify   PipelineMode = "classify"
)

// StageType categorizes an agent's role in the pipeline.
type StageType string

const (
	StagePlanner  StageType = "planner"
	StageExecutor StageType = "executor"
	StageReviewer StageType = "reviewer"
	StageJudge    StageType = "judge"
)

// StageConfig defines one stage in the pipeline.
type StageConfig struct {
	Name        string    `json:"name"`
	StageType   StageType `json:"stage_type"`
	AgentID     uint64    `json:"agent_id,omitempty"`
	Model       string    `json:"model,omitempty"`
	Effort      string    `json:"effort,omitempty"`
	Adversarial bool      `json:"adversarial,omitempty"` // enable 3-vote review for reviewer stages
	InputFrom   string    `json:"input_from,omitempty"`  // name of stage providing input
}

// BudgetConfig limits pipeline resource usage.
type BudgetConfig struct {
	MaxTokens  int `json:"max_tokens"`
	MaxRetries int `json:"max_retries"`
}

// PipelineConfig is the complete pipeline definition.
type PipelineConfig struct {
	Name   string        `json:"name"`
	Mode   PipelineMode  `json:"mode"`
	Stages []StageConfig `json:"stages"`
	Retry  BudgetConfig  `json:"retry"`
}

// StageResult captures the output of a single pipeline stage.
type StageResult struct {
	StageName  string    `json:"stage_name"`
	StageType  StageType `json:"stage_type"`
	Output     string    `json:"output"`
	TokensUsed int       `json:"tokens_used"`
	Cost       float64   `json:"cost"`
	DurationMs int       `json:"duration_ms"`
	Error      string    `json:"error,omitempty"`
}

// AgentCaller is the interface for invoking agents. Implemented by the service layer.
type AgentCaller interface {
	CallAgent(ctx context.Context, agentID uint64, model string, systemPrompt string, userMessage string, context map[string]interface{}) (result string, tokens int, cost float64, err error)
}

// PipelineRunner executes a PipelineConfig.
type PipelineRunner struct {
	caller AgentCaller
}

// NewPipelineRunner creates a new PipelineRunner.
func NewPipelineRunner(caller AgentCaller) *PipelineRunner {
	return &PipelineRunner{caller: caller}
}

// Run executes the pipeline with the given input context.
func (r *PipelineRunner) Run(ctx context.Context, config PipelineConfig, input map[string]interface{}) ([]StageResult, error) {
	switch config.Mode {
	case ModeSequential:
		return r.runSequential(ctx, config, input)
	case ModeFanOut:
		return r.runFanOut(ctx, config, input)
	case ModeTournament:
		return r.runTournament(ctx, config, input)
	case ModeClassify:
		return r.runClassify(ctx, config, input)
	default:
		return nil, fmt.Errorf("unknown pipeline mode: %s", config.Mode)
	}
}

func (r *PipelineRunner) runSequential(ctx context.Context, config PipelineConfig, input map[string]interface{}) ([]StageResult, error) {
	var results []StageResult
	var previousOutput string

	for i, stage := range config.Stages {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		// Build prompt based on stage type
		systemPrompt := r.buildSystemPrompt(stage)
		userMessage := r.buildUserMessage(stage, previousOutput, input)

		start := time.Now()
		output, tokens, cost, err := r.caller.CallAgent(ctx, stage.AgentID, stage.Model, systemPrompt, userMessage, input)
		duration := int(time.Since(start).Milliseconds())

		result := StageResult{
			StageName:  stage.Name,
			StageType:  stage.StageType,
			Output:     output,
			TokensUsed: tokens,
			Cost:       cost,
			DurationMs: duration,
		}

		if err != nil {
			result.Error = err.Error()
			// Retry logic
			if i < config.Retry.MaxRetries {
				log.Printf("[Pipeline] stage %s failed, retrying (%d/%d): %v", stage.Name, i+1, config.Retry.MaxRetries, err)
				continue
			}
			results = append(results, result)
			return results, fmt.Errorf("pipeline stage %s failed after retries: %w", stage.Name, err)
		}

		results = append(results, result)
		previousOutput = output
	}

	return results, nil
}

func (r *PipelineRunner) runFanOut(ctx context.Context, config PipelineConfig, input map[string]interface{}) ([]StageResult, error) {
	var results []StageResult

	// Find planner and executor stages
	var planner *StageConfig
	var executors []StageConfig
	var reviewer *StageConfig

	for i := range config.Stages {
		switch config.Stages[i].StageType {
		case StagePlanner:
			planner = &config.Stages[i]
		case StageExecutor:
			executors = append(executors, config.Stages[i])
		case StageReviewer:
			reviewer = &config.Stages[i]
		}
	}

	if planner == nil {
		return nil, fmt.Errorf("fan_out mode requires a planner stage")
	}

	// Step 1: Run planner
	planOutput, _, _, err := r.caller.CallAgent(ctx, planner.AgentID, planner.Model,
		r.buildSystemPrompt(*planner), r.buildUserMessage(*planner, "", input), input)
	if err != nil {
		return nil, fmt.Errorf("planner failed: %w", err)
	}
	results = append(results, StageResult{StageName: planner.Name, StageType: StagePlanner, Output: planOutput})

	// Step 2: Run executors in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex
	execResults := make([]StageResult, len(executors))

	for i, exec := range executors {
		wg.Add(1)
		go func(idx int, stage StageConfig) {
			defer wg.Done()
			start := time.Now()
			output, tokens, cost, err := r.caller.CallAgent(ctx, stage.AgentID, stage.Model,
				r.buildSystemPrompt(stage), planOutput+"\n\nYour sub-task: "+r.buildUserMessage(stage, "", input), input)
			dur := int(time.Since(start).Milliseconds())
			mu.Lock()
			execResults[idx] = StageResult{
				StageName:  stage.Name,
				StageType:  StageExecutor,
				Output:     output,
				TokensUsed: tokens,
				Cost:       cost,
				DurationMs: dur,
			}
			if err != nil {
				execResults[idx].Error = err.Error()
			}
			mu.Unlock()
		}(i, exec)
	}
	wg.Wait()

	for _, r := range execResults {
		results = append(results, r)
	}

	// Step 3: Run reviewer if present
	if reviewer != nil {
		combinedOutput := planOutput
		for _, er := range execResults {
			combinedOutput += "\n---\n" + er.Output
		}
		output, tokens, cost, err := r.caller.CallAgent(ctx, reviewer.AgentID, reviewer.Model,
			r.buildSystemPrompt(*reviewer), combinedOutput, input)
		results = append(results, StageResult{
			StageName: reviewer.Name, StageType: StageReviewer,
			Output: output, TokensUsed: tokens, Cost: cost,
		})
		if err != nil {
			results[len(results)-1].Error = err.Error()
		}
	}

	return results, nil
}

func (r *PipelineRunner) runTournament(ctx context.Context, config PipelineConfig, input map[string]interface{}) ([]StageResult, error) {
	// Find executors and judge
	var executors []StageConfig
	var judge *StageConfig

	for i := range config.Stages {
		switch config.Stages[i].StageType {
		case StageExecutor:
			executors = append(executors, config.Stages[i])
		case StageJudge:
			judge = &config.Stages[i]
		}
	}

	// Run all executors in parallel (same as fan-out)
	var results []StageResult
	var wg sync.WaitGroup
	var mu sync.Mutex
	execResults := make([]StageResult, len(executors))

	for i, exec := range executors {
		wg.Add(1)
		go func(idx int, stage StageConfig) {
			defer wg.Done()
			output, tokens, cost, err := r.caller.CallAgent(ctx, stage.AgentID, stage.Model,
				r.buildSystemPrompt(stage), r.buildUserMessage(stage, "", input), input)
			mu.Lock()
			execResults[idx] = StageResult{
				StageName: stage.Name, StageType: StageExecutor,
				Output: output, TokensUsed: tokens, Cost: cost,
			}
			if err != nil {
				execResults[idx].Error = err.Error()
			}
			mu.Unlock()
		}(i, exec)
	}
	wg.Wait()
	for _, r := range execResults {
		results = append(results, r)
	}

	// Judge picks the winner
	if judge != nil {
		options := ""
		for i, er := range execResults {
			options += fmt.Sprintf("\n=== Option %d ===\n%s\n", i+1, er.Output)
		}
		prompt := fmt.Sprintf("Compare the following options and select the best one. Explain your reasoning.\n%s", options)
		output, tokens, cost, err := r.caller.CallAgent(ctx, judge.AgentID, judge.Model,
			"You are a judge. Compare options and pick the best.", prompt, input)
		results = append(results, StageResult{
			StageName: judge.Name, StageType: StageJudge,
			Output: output, TokensUsed: tokens, Cost: cost,
		})
		if err != nil {
			results[len(results)-1].Error = err.Error()
		}
	}

	return results, nil
}

func (r *PipelineRunner) runClassify(ctx context.Context, config PipelineConfig, input map[string]interface{}) ([]StageResult, error) {
	var results []StageResult

	// First stage must be a classifier (type executor acting as classifier)
	if len(config.Stages) < 2 {
		return nil, fmt.Errorf("classify mode requires at least 2 stages")
	}

	classifier := config.Stages[0]
	// Run classifier to determine which executor to use
	options := ""
	for i := 1; i < len(config.Stages); i++ {
		options += fmt.Sprintf("- %s: %s\n", config.Stages[i].Name, config.Stages[i].StageType)
	}

	classifyPrompt := fmt.Sprintf("Classify this task and respond with EXACTLY the name of the best handler:\nTask: %v\n\nOptions:\n%s\n\nRespond with just the handler name.", input, options)

	choice, tokens, cost, err := r.caller.CallAgent(ctx, classifier.AgentID, classifier.Model,
		"You are a task classifier. Respond with only the handler name.", classifyPrompt, input)
	results = append(results, StageResult{
		StageName: classifier.Name, StageType: StageExecutor,
		Output: choice, TokensUsed: tokens, Cost: cost,
	})
	if err != nil {
		return results, err
	}

	// Find and run the chosen executor
	for i := 1; i < len(config.Stages); i++ {
		if config.Stages[i].Name == choice || i == 1 { // fallback to first executor
			output, tokens, cost, err := r.caller.CallAgent(ctx, config.Stages[i].AgentID, config.Stages[i].Model,
				r.buildSystemPrompt(config.Stages[i]), r.buildUserMessage(config.Stages[i], "", input), input)
			results = append(results, StageResult{
				StageName: config.Stages[i].Name, StageType: StageExecutor,
				Output: output, TokensUsed: tokens, Cost: cost,
			})
			if err != nil {
				results[len(results)-1].Error = err.Error()
			}
			break
		}
	}

	return results, nil
}

func (r *PipelineRunner) buildSystemPrompt(stage StageConfig) string {
	switch stage.StageType {
	case StagePlanner:
		return "You are a Planner. Analyze the task and produce a clear, actionable specification. Break complex tasks into concrete steps."
	case StageReviewer:
		return "You are a QA Reviewer. CRITICAL: Assume the following output contains errors. Your task is to FIND them. Default to finding issues if uncertain."
	case StageJudge:
		return "You are a Judge. Compare the options objectively and select the best one. Explain your reasoning."
	default:
		return "You are an Executor. Follow the specification precisely and produce the requested output."
	}
}

func (r *PipelineRunner) buildUserMessage(stage StageConfig, previousOutput string, input map[string]interface{}) string {
	msg := fmt.Sprintf("Task context: %v", input)
	if previousOutput != "" {
		msg += fmt.Sprintf("\n\nInput from previous stage:\n%s", previousOutput)
	}
	return msg
}
