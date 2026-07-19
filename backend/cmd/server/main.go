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

	// Purge legacy workspace-level labels so project_id can become NOT NULL (project-only labels, aligned with Plane)
	db.Exec(`DELETE FROM issue_labels WHERE label_id IN (SELECT id FROM labels WHERE project_id IS NULL)`)
	db.Exec(`DELETE FROM labels WHERE project_id IS NULL`)
	// GORM AutoMigrate does not tighten nullability on existing columns; enforce explicitly
	db.Exec(`ALTER TABLE labels ALTER COLUMN project_id SET NOT NULL`)

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

	// Migrate existing workspace-level issue types to imported state for all projects
	// (Plane v3-style Import model: project references workspace type by link).
	// After this migration, every project that can see a workspace-level type will
	// have an explicit IssueTypeImport record, so the IsImported flag is set and
	// attached custom fields become visible without separate enrollment.
	db.Exec(`INSERT INTO issue_type_imports (project_id, workspace_type_id, workspace_id, created_at, updated_at)
		SELECT DISTINCT p.id, t.id, t.workspace_id, NOW(), NOW()
		FROM projects p
		JOIN issue_types t ON t.workspace_id = p.workspace_id AND t.project_id IS NULL
		WHERE NOT EXISTS (
			SELECT 1 FROM issue_type_imports iti
			WHERE iti.project_id = p.id AND iti.workspace_type_id = t.id
		)
		ON CONFLICT DO NOTHING`)
	fmt.Println("Issue type import migration completed")

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
