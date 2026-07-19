package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/config"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/router"
	"github.com/reqmango/backend/internal/seed"
	searchtemplate "github.com/reqmango/backend/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	fmt.Println("Database connected")

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
		&model.IssueAssignee{},
		&model.IssueLabel{},
		&model.IssueCycle{},
		&model.IssueActivity{},
		&model.Cycle{},
		&model.Module{},
		&model.ModuleIssue{},
		&model.IssueType{},
		&model.IssueTypeField{},
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
		&model.WorkItemTemplate{},
		&model.Release{},
		&model.ReleaseIssue{},
		&model.EstimatePoint{},
		&model.EstimateCategory{},
		&model.EstimateTime{},
		&model.ProjectEstimateSettings{},
		&model.IssuePage{},
		&model.Attachment{},
		&model.AIConfig{},
		&model.AIThread{},
		&model.AIMessage{},
		&model.ConditionalField{},
		&model.TimeTrack{},
		&model.RecurrenceRule{},
		&model.Webhook{},
		&model.Initiative{},
		&model.InitiativeProject{},
		&model.ProjectUpdate{},
		&model.ProjectPageTab{},
		&model.Agent{},
		&model.AgentActivity{},
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
	); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}
	fmt.Println("Database migration completed")

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

	seed.SeedAll(db)

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
