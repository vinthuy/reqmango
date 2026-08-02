package service

import (
	"encoding/json"

	"gorm.io/gorm"
)

// ContextPayloadService manages context payloads for workflow execution.
type ContextPayloadService struct {
	db *gorm.DB
}

// NewContextPayloadService creates a new ContextPayloadService.
func NewContextPayloadService(db *gorm.DB) *ContextPayloadService {
	return &ContextPayloadService{db: db}
}

// ContextPayload represents the context passed between workflow nodes.
type ContextPayload struct {
	IssueContext  *IssueContext  `json:"issue_context,omitempty"`
	Documents     []DocumentRef  `json:"documents,omitempty"`
	AgentOutputs  []AgentOutput  `json:"agent_outputs,omitempty"`
	SharedData    map[string]interface{} `json:"shared_data,omitempty"`
}

// IssueContext represents issue-related context.
type IssueContext struct {
	IssueID     uint64                 `json:"issue_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Priority    string                 `json:"priority"`
	Type        string                 `json:"type"`
	State       string                 `json:"state"`
	Labels      []string               `json:"labels"`
	Assignees   []string               `json:"assignees"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// DocumentRef represents a reference to a document.
type DocumentRef struct {
	ID         uint64 `json:"id"`
	Title      string `json:"title"`
	Type       string `json:"type"` // wiki, requirement, design, test
	TokenCount int    `json:"token_count"`
	Summary    string `json:"summary,omitempty"` // for compressed docs
	FullContent string `json:"full_content,omitempty"`
}

// AgentOutput represents the output from a previous agent.
type AgentOutput struct {
	AgentID    uint64 `json:"agent_id"`
	AgentName  string `json:"agent_name"`
	NodeType   string `json:"node_type"`
	Content    string `json:"content"`
	TokenCount int    `json:"token_count"`
}

// BuildInitialContext builds the initial context from an issue.
func (s *ContextPayloadService) BuildInitialContext(issueID uint64) (*ContextPayload, error) {
	ctx := &ContextPayload{
		Documents:    []DocumentRef{},
		AgentOutputs: []AgentOutput{},
		SharedData:   make(map[string]interface{}),
	}

	// Get issue details
	if issueID > 0 {
		var issue struct {
			ID          uint64 `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description_stripped"`
			Priority    string `json:"priority"`
		}

		err := s.db.Raw(`
			SELECT id, name, COALESCE(description_stripped, '') as description_stripped, priority 
			FROM issues WHERE id = ?
		`, issueID).Scan(&issue).Error

		if err == nil {
			ctx.IssueContext = &IssueContext{
				IssueID:     issue.ID,
				Name:        issue.Name,
				Description: issue.Description,
				Priority:    issue.Priority,
			}
		}
	}

	return ctx, nil
}

// BuildNodeInput builds the input context for a specific node.
func (s *ContextPayloadService) BuildNodeInput(workflowRunID, nodeID uint64, edges []interface{}) (*ContextPayload, error) {
	// Get workflow run context
	var run struct {
		Context json.RawMessage `json:"context"`
	}

	err := s.db.Raw("SELECT context FROM workflow_runs WHERE id = ?", workflowRunID).Scan(&run).Error
	if err != nil {
		return nil, err
	}

	ctx := &ContextPayload{
		Documents:    []DocumentRef{},
		AgentOutputs: []AgentOutput{},
		SharedData:   make(map[string]interface{}),
	}

	if run.Context != nil {
		json.Unmarshal(run.Context, ctx)
	}

	// Get outputs from previous nodes
	var prevOutputs []struct {
		OutputContext json.RawMessage `json:"output_context"`
	}

	s.db.Raw(`
		SELECT wr.output_context 
		FROM workflow_node_runs wr
		JOIN workflow_edges we ON we.source_node_id = wr.node_id
		WHERE wr.workflow_run_id = ? AND we.target_node_id = ? AND wr.status = 'completed'
	`, workflowRunID, nodeID).Scan(&prevOutputs)

	for _, output := range prevOutputs {
		if output.OutputContext != nil {
			var agentOutput AgentOutput
			if err := json.Unmarshal(output.OutputContext, &agentOutput); err == nil {
				ctx.AgentOutputs = append(ctx.AgentOutputs, agentOutput)
			}
		}
	}

	return ctx, nil
}

// MergeParallelOutputs merges outputs from parallel nodes.
func (s *ContextPayloadService) MergeParallelOutputs(workflowRunID uint64, nodeIDs []uint64) (*ContextPayload, error) {
	ctx := &ContextPayload{
		Documents:    []DocumentRef{},
		AgentOutputs: []AgentOutput{},
		SharedData:   make(map[string]interface{}),
	}

	for _, nodeID := range nodeIDs {
		var output struct {
			OutputContext json.RawMessage `json:"output_context"`
		}

		s.db.Raw(`
			SELECT output_context 
			FROM workflow_node_runs 
			WHERE workflow_run_id = ? AND node_id = ? AND status = 'completed'
			LIMIT 1
		`, workflowRunID, nodeID).Scan(&output)

		if output.OutputContext != nil {
			var agentOutput AgentOutput
			if err := json.Unmarshal(output.OutputContext, &agentOutput); err == nil {
				ctx.AgentOutputs = append(ctx.AgentOutputs, agentOutput)
			}
		}
	}

	return ctx, nil
}

// CompressContext compresses the context to reduce token usage.
func (s *ContextPayloadService) CompressContext(ctx *ContextPayload) (*ContextPayload, error) {
	compressed := &ContextPayload{
		IssueContext: ctx.IssueContext,
		Documents:    make([]DocumentRef, 0),
		AgentOutputs: make([]AgentOutput, 0),
		SharedData:   ctx.SharedData,
	}

	// 1. Compress documents
	for _, doc := range ctx.Documents {
		if doc.TokenCount > 2000 {
			// Replace with summary
			compressedDoc := doc
			compressedDoc.FullContent = ""
			compressedDoc.Summary = "Document compressed: " + doc.Title
			compressed.Documents = append(compressed.Documents, compressedDoc)
		} else {
			compressed.Documents = append(compressed.Documents, doc)
		}
	}

	// 2. Compress agent outputs
	if len(ctx.AgentOutputs) > 3 {
		// Keep only last 3 outputs, compress older ones to conclusions
		recent := ctx.AgentOutputs[len(ctx.AgentOutputs)-3:]
		compressed.AgentOutputs = append(compressed.AgentOutputs, AgentOutput{
			AgentID:   0,
			AgentName: "System",
			NodeType:  "summary",
			Content:   "Previous outputs compressed",
		})
		compressed.AgentOutputs = append(compressed.AgentOutputs, recent...)
	} else {
		compressed.AgentOutputs = ctx.AgentOutputs
	}

	return compressed, nil
}

// ContextToJSON converts a context payload to JSON.
func (s *ContextPayloadService) ContextToJSON(ctx *ContextPayload) (json.RawMessage, error) {
	return json.Marshal(ctx)
}

// JSONToContext converts JSON to a context payload.
func (s *ContextPayloadService) JSONToContext(data json.RawMessage) (*ContextPayload, error) {
	ctx := &ContextPayload{}
	err := json.Unmarshal(data, ctx)
	return ctx, err
}
