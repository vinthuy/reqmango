package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/config"
	"github.com/reqmanpy/backend-go/internal/model"
	"github.com/reqmanpy/backend-go/internal/router"
	"github.com/reqmanpy/backend-go/internal/seed"
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
	); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}
	fmt.Println("Database migration completed")

	seed.SeedAll(db)

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	router.SetupRoutes(r, db, cfg)

	addr := fmt.Sprintf(":%s", cfg.Port)
	fmt.Printf("Server starting on %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}