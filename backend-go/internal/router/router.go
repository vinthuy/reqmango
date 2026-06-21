package router

import (
	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/config"
	"github.com/reqmanpy/backend-go/internal/handler"
	"github.com/reqmanpy/backend-go/internal/middleware"
	"github.com/reqmanpy/backend-go/internal/service"
	"gorm.io/gorm"
)

// SetupRoutes initializes all services, handlers, and routes.
func SetupRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// Initialize services
	authSvc := service.NewAuthService(db, cfg)
	workspaceSvc := service.NewWorkspaceService(db)
	projectSvc := service.NewProjectService(db)
	settingsSvc := service.NewProjectSettingsService(db)
	issueSvc := service.NewIssueService(db)
	cycleSvc := service.NewCycleService(db)
	moduleSvc := service.NewModuleService(db)

	// Initialize handlers
	authH := handler.NewAuthHandler(authSvc)
	workspaceH := handler.NewWorkspaceHandler(workspaceSvc)
	projectH := handler.NewProjectHandler(projectSvc)
	settingsH := handler.NewProjectSettingsHandler(settingsSvc)
	issueH := handler.NewIssueHandler(issueSvc)
	cycleH := handler.NewCycleHandler(cycleSvc)
	moduleH := handler.NewModuleHandler(moduleSvc)

	// JWT middleware
	authMiddleware := middleware.AuthMiddleware(db, cfg.SecretKey)

	// ==================== API v1 ====================
	v1 := r.Group("/api/v1")
	{
		// ---- Auth (public + protected) ----
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
			auth.GET("/me", authMiddleware, authH.GetCurrentUser)
		}

		// ---- Workspaces (protected) ----
		workspaces := v1.Group("/workspaces", authMiddleware)
		{
			workspaces.GET("", workspaceH.List)
			workspaces.POST("", workspaceH.Create)
			workspaces.GET("/:wsParam", workspaceH.Get)     // slug or numeric ID
			workspaces.PATCH("/:id", workspaceH.Update)      // numeric ID
			workspaces.DELETE("/:id", workspaceH.Delete)     // numeric ID
			workspaces.GET("/:wsParam/members", workspaceH.ListMembers)
		}

		// ---- Projects (protected) ----
		projects := v1.Group("/projects", authMiddleware)
		{
			projects.POST("", projectH.Create)                        // ?workspace_id=
			projects.GET("", projectH.List)                          // ?workspace_id=
			projects.GET("/:projectId", projectH.Get)
			projects.PATCH("/:projectId", projectH.Update)
			projects.DELETE("/:projectId", projectH.Delete)
			projects.POST("/:projectId/archive", projectH.Archive)
			projects.POST("/:projectId/restore", projectH.Restore)
			projects.GET("/:projectId/members", projectH.ListMembers)       // ?only_active=
			projects.POST("/:projectId/members", projectH.AddMember)        // ?user_id=&role=
			projects.PATCH("/:projectId/members/:userId", projectH.UpdateMember) // ?role=
			projects.DELETE("/:projectId/members/:userId", projectH.RemoveMember)
			projects.GET("/:projectId/statistics", projectH.GetStatistics)
			projects.GET("/:projectId/issues-summary", projectH.GetIssuesSummary)
			projects.GET("/:projectId/cycles", cycleH.List)

			// ---- Project Settings: States + Labels ----
			settings := projects.Group("/:projectId/settings", authMiddleware)
			{
				// States
				settings.POST("/states", settingsH.CreateState)              // ?workspace_id=
				settings.GET("/states", settingsH.ListStates)               // ?include_inactive=
				settings.POST("/states/default", settingsH.CreateDefaultStates) // ?workspace_id=
				settings.GET("/states/:stateId", settingsH.GetState)
				settings.PUT("/states/:stateId", settingsH.UpdateState)
				settings.DELETE("/states/:stateId", settingsH.DeleteState)

				// Labels
				settings.POST("/labels", settingsH.CreateLabel)             // ?workspace_id=
				settings.GET("/labels", settingsH.ListLabels)
				settings.GET("/labels/:labelId", settingsH.GetLabel)
				settings.PUT("/labels/:labelId", settingsH.UpdateLabel)
				settings.DELETE("/labels/:labelId", settingsH.DeleteLabel)
			}
		}

		// ---- Issues (protected) ----
		issues := v1.Group("/issues", authMiddleware)
		{
			// CRUD
			issues.POST("", issueH.Create)                       // ?project_id=&workspace_id=
			issues.GET("", issueH.List)                          // ?project_id=&filters...
			issues.GET("/statistics", issueH.GetStatistics)       // ?project_id=
			issues.GET("/search", issueH.Search)                  // ?workspace_id=&query=
			issues.POST("/bulk/update", issueH.BulkUpdate)        // ?project_id=
			issues.POST("/bulk/delete", issueH.BulkDelete)

			// Single issue
			issues.GET("/:issueId", issueH.Get)
			issues.PUT("/:issueId", issueH.Update)
			issues.DELETE("/:issueId", issueH.Delete)
			issues.POST("/:issueId/archive", issueH.Archive)
			issues.POST("/:issueId/restore", issueH.Restore)
			issues.GET("/:issueId/activities", issueH.GetActivities) // ?limit=&offset=

			// Assignees
			issues.POST("/:issueId/assignees", issueH.AddAssignee)          // ?user_id=
			issues.DELETE("/:issueId/assignees/:userId", issueH.RemoveAssignee)

			// Labels
			issues.POST("/:issueId/labels", issueH.AddLabel)               // ?label_id=
			issues.DELETE("/:issueId/labels/:labelId", issueH.RemoveLabel)

			// Cycle
			issues.POST("/:issueId/cycle", issueH.SetCycle)                // ?cycle_id=
			issues.DELETE("/:issueId/cycle", issueH.RemoveCycle)
		}

		// ---- Modules (protected) ----
		modules := v1.Group("/modules", authMiddleware)
		{
			// CRUD
			modules.POST("", moduleH.Create)                      // ?workspace_id=
			modules.GET("", moduleH.List)                         // ?project_id=&workspace_id=&include_archived=
			modules.GET("/:moduleId", moduleH.Get)
			modules.PUT("/:moduleId", moduleH.Update)
			modules.DELETE("/:moduleId", moduleH.Delete)
		}
	}
}
