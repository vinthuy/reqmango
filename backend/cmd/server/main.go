package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	airegistry "github.com/reqmango/backend/internal/ai/registry"
	aiservice "github.com/reqmango/backend/internal/ai/service"
	"github.com/reqmango/backend/internal/config"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/router"
	"github.com/reqmango/backend/internal/seed"
	"github.com/reqmango/backend/internal/service"
	searchtemplate "github.com/reqmango/backend/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg := config.Load()

	// GORM's Info level logs every statement *with its parameters*, which is handy
	// while debugging but produced ~80 MB of log in a 20 minute E2E run and wrote
	// user data to disk. Verbose output is therefore opt-in: DEBUG=true (or an
	// explicit DB_LOG_LEVEL), otherwise only warnings and errors are emitted.
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(cfg.DBLogLevel()),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	fmt.Println("Database connected")

	// Purge legacy workspace-level labels so project_id can become NOT NULL (project-only labels, project-scoped).
	// On a fresh DB these tables do not exist yet — guard so AutoMigrate can create them cleanly.
	if db.Migrator().HasTable("labels") {
		if db.Migrator().HasTable("issue_labels") {
			db.Exec(`DELETE FROM issue_labels WHERE label_id IN (SELECT id FROM labels WHERE project_id IS NULL)`)
		}
		db.Exec(`DELETE FROM labels WHERE project_id IS NULL`)
		// GORM AutoMigrate does not tighten nullability on existing columns; enforce explicitly
		db.Exec(`ALTER TABLE labels ALTER COLUMN project_id SET NOT NULL`)
	}

	// Drop old foreign key constraint on automation_rules BEFORE AutoMigrate
	// (workspace rules have project_id=0, GORM may validate FK during migration)
	db.Exec(`DO $$ BEGIN
		IF EXISTS (SELECT 1 FROM information_schema.table_constraints
			WHERE table_name = 'automation_rules' AND constraint_name = 'fk_automation_rules_project') THEN
			ALTER TABLE automation_rules DROP CONSTRAINT fk_automation_rules_project;
		END IF;
	END $$`)
	fmt.Println("Automation rules FK cleanup completed")

	// Retire the state_transitions.workflow_kind discriminator before
	// AutoMigrate restores the foreign key to `workflows`: a leftover agent
	// transition would make that constraint impossible to create. Migration
	// 000002 does the same for deployments that run cmd/migrate; this keeps the
	// AutoMigrate-only path (the Docker image) converging on the same schema.
	if err := retireStateTransitionKind(db); err != nil {
		log.Fatalf("Failed to retire state_transitions.workflow_kind: %v", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Workspace{},
		&model.WorkspaceMember{},
		&model.Project{},
		&model.ProjectMember{},
		&model.State{},
		&model.StateTransition{},
		&model.Label{},
		&model.Issue{},
		&model.ProjectIntakeSetting{},
		&model.IssueAssignee{},
		&model.IssueLabel{},
		&model.IssueCycle{},
		&model.IssueActivity{},
		&model.IssueWatcher{},
		&model.Cycle{},
		&model.Module{},
		&model.ModuleIssue{},
		&model.ModuleInheritanceOverride{},
		&model.IssueType{},
		&model.IssueTypeField{},
		&model.IssueTypeImport{},
		&model.CustomField{},
		&model.CustomFieldOption{},
		&model.IssueCustomFieldValue{},
		&model.ProjectSubscriber{},
		&model.ProjectTemplate{},
		&model.ProjectTemplateType{},
		&model.IssueTypeTemplate{},
		&model.IssueTypeTemplateField{},
		&model.RelationType{},
		&model.IssueRelation{},
		&model.Workflow{},
		&model.AutomationRule{},
		&model.AutomationExecution{},
		&model.Comment{},
		&model.Notification{},
		&model.SavedView{},
		&model.Page{},
		&model.PageTemplate{},
		&model.PageVersion{},
		&model.WorkItemTemplate{},
		&model.Release{},
		&model.ReleaseIssue{},
		&model.EstimatePoint{},
		&model.EstimateCategory{},
		&model.EstimateTime{},
		&model.ProjectEstimateSettings{},
		&model.IssuePage{},
		&model.Attachment{},
		&model.ConditionalField{},
		&model.TimeTrack{},
		&model.RecurrenceRule{},
		&model.Webhook{},
		&model.Initiative{},
		&model.InitiativeProject{},
		&model.ProjectUpdate{},
		&model.ProjectPageTab{},
		&model.MCPConfig{},
		&model.GitHubConnection{},
		&model.SlackConnection{},
		&model.Role{},
		&model.Permission{},
		&model.RolePermission{},
		&model.FieldPermission{},
		&model.SavedReport{},
		&model.ProjectCustomFieldEnrollment{},
		&model.SearchTemplate{},
		&model.SavedDashboard{},
		&model.DashboardWidget{},
		&model.MetricChart{},
		&model.Plugin{},
		&model.PluginEventLog{},
		&model.GitIntegration{},
		&model.GitIssueLink{},
		&model.Approval{},
		&model.ApprovalRecord{},
		&model.AutomationRuleOverride{},
		// AI models (merged from agent-service)
		&model.AIConfig{},
		&model.AIThread{},
		&model.AIMessage{},
		&model.Agent{},
		&model.AgentActivity{},
		&airegistry.AgentEntry{},
		&model.Loop{},
		&model.LoopRun{},
		&model.LoopIteration{},
		&model.Pipeline{},
		&model.PipelineRun{},
		&model.AgentSession{},
		// New Agent models
		&model.AgentTemplate{},
		&model.AgentConfig{},
		&model.Skill{},
		&model.SkillExecutionLog{},
		&model.AgentTask{},
		&model.AgentTaskSnapshot{},
		&model.Runtime{},
		// Tool Calling models
		&model.Tool{},
		&model.ToolPermission{},
		&model.ToolCallLog{},
		// Memory models
		&model.MemoryEntry{},
		&model.MemorySession{},
		&model.MemoryVectorIndex{},
		// Squad models
		&model.Squad{},
		&model.SquadMember{},
		&model.SquadTask{},
		&model.SquadExecution{},
		// Autopilot models
		&model.AutopilotTask{},
		&model.AutopilotExecution{},
		&model.AutopilotWebhook{},
		// Agent-Project Integration models
		&model.AgentWorkflow{},
		&model.WorkflowNode{},
		&model.WorkflowEdge{},
		&model.WorkflowRun{},
		&model.WorkflowNodeRun{},
		// Chat & Messages
		&model.Chat{},
		&model.Message{},
		&model.MessageReaction{},
		// Phase 4 - Developer Agent (P4-001)
		&model.DeveloperJob{},
		// Phase 4 - Tester Agent (P4-002)
		&model.TesterJob{},
		// Phase 4 - CI/CD integration (P4-003)
		&model.CICDConfig{},
		&model.BuildRecord{},
		// Phase 4 - SDLC orchestration (P4-006)
		&model.SDLCWorkflow{},
		&model.SDLCStage{},
		&model.DataMigration{},
		// Models that only had SQL migrations. A deployment that relies on
		// AutoMigrate alone (the Docker image) got 500s from every handler
		// reading these tables, so they belong in the same list.
		&model.ProjectAgentMember{},
		&model.AgentDecision{},
		&model.AgentCostBudget{},
		&model.AgentSLA{},
		&model.TaskLog{},
		&model.PersonalAccessToken{},
	); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}
	fmt.Println("Database migration completed")

	// Register built-in tools
	toolSvc := service.NewToolService(db)
	if err := toolSvc.RegisterBuiltinTools(0); err != nil {
		log.Fatalf("Failed to register built-in tools: %v", err)
	}
	fmt.Println("Built-in tools registered")

	// Drop old foreign key constraint on automation_executions (if exists)
	db.Exec(`DO $$ BEGIN
		IF EXISTS (SELECT 1 FROM information_schema.table_constraints
			WHERE table_name = 'automation_executions' AND constraint_name = 'fk_automation_executions_rule') THEN
			ALTER TABLE automation_executions DROP CONSTRAINT fk_automation_executions_rule;
		END IF;
		IF EXISTS (SELECT 1 FROM information_schema.table_constraints
			WHERE table_name = 'automation_executions' AND constraint_name = 'fk_automation_executions_issue') THEN
			ALTER TABLE automation_executions DROP CONSTRAINT fk_automation_executions_issue;
		END IF;
	END $$`)
	fmt.Println("Foreign key cleanup completed")

	// Migrate existing workspace-level fields to enrolled state for all projects (backward compatibility)
	db.Exec(`INSERT INTO project_custom_field_enrollments (project_id, field_id, is_enabled)
		SELECT p.id, cf.id, true
		FROM custom_fields cf
		JOIN projects p ON cf.workspace_id = p.workspace_id
		WHERE cf.project_id IS NULL
		ON CONFLICT (project_id, field_id) DO NOTHING`)
	fmt.Println("Custom field enrollment migration completed")

	// Seed default agents in registry

	// Create full-text search index for issues
	db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_issues_search ON issues 
		USING gin(to_tsvector('english', COALESCE(name, '') || ' ' || COALESCE(description_stripped, '')))
	`)
	fmt.Println("Full-text search index created")

	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_labels_project_name ON labels (project_id, name) WHERE deleted_at IS NULL`).Error; err != nil {
		log.Printf("WARNING: failed to create labels unique index: %v", err)
	}

	seed.SeedAll(db, seed.Options{Demo: cfg.SeedDemo, E2E: cfg.SeedE2E})
	seed.EnsureBootstrapAdmin(db, cfg.AdminEmail, cfg.AdminPassword)

	go func() {
		ran, filled, expired, err := aiservice.RunAgentMemoryBackfillOnce(db)
		if err != nil {
			log.Printf("WARNING: agent memory backfill failed (will retry next start): %v", err)
		} else if ran {
			fmt.Printf("Agent memory backfill completed: %d linked, %d expired\n", filled, expired)
		}
	}()

	// Update built-in search templates for all projects (fixes RQL / description)
	var projects []model.Project
	db.Find(&projects)
	sts := searchtemplate.NewSearchTemplateService(db)
	for _, p := range projects {
		if err := sts.InitializeBuiltInTemplates(p.ID); err != nil {
			fmt.Printf("WARNING: Failed to update built-in templates for project %d: %v\n", p.ID, err)
		}
	}

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.RedirectFixedPath = false
	r.RedirectTrailingSlash = false
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	router.SetupRoutes(r, db, cfg)

	addr := fmt.Sprintf(":%s", cfg.Port)
	fmt.Printf("Server starting on %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// retireStateTransitionKind converges a database that predates migration 000002
// onto the single-owner schema AutoMigrate expects. It is a no-op once the
// column is gone, so it costs one catalogue query per start.
func retireStateTransitionKind(db *gorm.DB) error {
	var hasColumn bool
	if err := db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_name = 'state_transitions'
		  AND column_name = 'workflow_kind')`).Scan(&hasColumn).Error; err != nil {
		return err
	}
	if !hasColumn {
		return nil
	}

	// Rows are archived, never dropped: agent transitions and transitions whose
	// workflow row is gone cannot satisfy the restored foreign key.
	steps := []string{
		`CREATE TABLE IF NOT EXISTS state_transitions_agent_archive
			(LIKE state_transitions INCLUDING DEFAULTS)`,
		`ALTER TABLE state_transitions_agent_archive
			ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ DEFAULT now()`,
		`INSERT INTO state_transitions_agent_archive
			SELECT t.*, now() FROM state_transitions t
			WHERE t.workflow_kind = 'agent'
			  AND NOT EXISTS (SELECT 1 FROM state_transitions_agent_archive a WHERE a.id = t.id)`,
		`INSERT INTO state_transitions_agent_archive
			SELECT t.*, now() FROM state_transitions t
			WHERE t.workflow_kind <> 'agent'
			  AND NOT EXISTS (SELECT 1 FROM workflows w WHERE w.id = t.workflow_id)
			  AND NOT EXISTS (SELECT 1 FROM state_transitions_agent_archive a WHERE a.id = t.id)`,
		`DELETE FROM state_transitions t
			WHERE t.workflow_kind = 'agent'
			   OR NOT EXISTS (SELECT 1 FROM workflows w WHERE w.id = t.workflow_id)`,
		`ALTER TABLE state_transitions DROP COLUMN IF EXISTS workflow_kind`,
	}
	for _, step := range steps {
		if err := db.Exec(step).Error; err != nil {
			return err
		}
	}
	fmt.Println("Retired state_transitions.workflow_kind")
	return nil
}
