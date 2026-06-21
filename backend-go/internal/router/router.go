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
	issueTypeSvc := service.NewIssueTypeService(db)
	customFieldSvc := service.NewCustomFieldService(db)
	templateSvc := service.NewProjectTemplateService(db)
	typeTemplateSvc := service.NewTypeTemplateService(db)

	// Initialize handlers
	authH := handler.NewAuthHandler(authSvc)
	workspaceH := handler.NewWorkspaceHandler(workspaceSvc)
	projectH := handler.NewProjectHandler(projectSvc)
	settingsH := handler.NewProjectSettingsHandler(settingsSvc)
	issueH := handler.NewIssueHandler(issueSvc)
	cycleH := handler.NewCycleHandler(cycleSvc)
	moduleH := handler.NewModuleHandler(moduleSvc)
	issueTypeH := handler.NewIssueTypeHandler(issueTypeSvc)
	customFieldH := handler.NewCustomFieldHandler(customFieldSvc)
	templateH := handler.NewProjectTemplateHandler(templateSvc)
	typeTemplateH := handler.NewTypeTemplateHandler(typeTemplateSvc)

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
			projects.POST("/:projectId/cycles", cycleH.Create)     // ?workspace_id=

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
			
			// Issues
			modules.POST("/:moduleId/issues", moduleH.AddIssue)         // ?issue_id=
			modules.DELETE("/:moduleId/issues/:issueId", moduleH.RemoveIssue)
			modules.GET("/:moduleId/issues", moduleH.ListIssues)
			
			// Analysis
			modules.GET("/:moduleId/progress", moduleH.GetProgress)
			modules.GET("/:moduleId/statistics", moduleH.GetStatistics)
			
			// Tree
			modules.GET("/tree", moduleH.GetTree)                       // ?project_id=
		}
		// ---- Cycles (protected) ----
		cycles := v1.Group("/cycles", authMiddleware)
		{
			cycles.GET("/:cycleId", cycleH.Get)
			cycles.PUT("/:cycleId", cycleH.Update)
			cycles.DELETE("/:cycleId", cycleH.Delete)
			cycles.POST("/:cycleId/start", cycleH.Start)
			cycles.POST("/:cycleId/end", cycleH.End)
			cycles.POST("/:cycleId/cancel", cycleH.Cancel)
			cycles.POST("/:cycleId/issues", cycleH.AddIssue)
			cycles.DELETE("/:cycleId/issues/:issueId", cycleH.RemoveIssue)
			cycles.GET("/:cycleId/issues", cycleH.ListIssues)
			cycles.GET("/:cycleId/progress", cycleH.GetProgress)
			cycles.GET("/:cycleId/statistics", cycleH.GetStatistics)
			cycles.GET("/:cycleId/burndown", cycleH.GetBurndown)
			}

			// ---- Issue Types (protected) ----
			issueTypes := v1.Group("/issue-types", authMiddleware)
			{
				issueTypes.POST("", issueTypeH.Create)
				issueTypes.GET("", issueTypeH.List)
				issueTypes.GET("/:typeId", issueTypeH.Get)
				issueTypes.PUT("/:typeId", issueTypeH.Update)
				issueTypes.DELETE("/:typeId", issueTypeH.Delete)
				issueTypes.PATCH("/:typeId/disable", issueTypeH.Disable)
				issueTypes.GET("/:typeId/fields", issueTypeH.ListFields)
				issueTypes.POST("/:typeId/fields", issueTypeH.AddField)
				issueTypes.DELETE("/:typeId/fields/:fieldId", issueTypeH.RemoveField)
			}

			// ---- Custom Fields (protected) ----
			customFields := v1.Group("/custom-fields", authMiddleware)
			{
				customFields.POST("", customFieldH.Create)
				customFields.GET("", customFieldH.List)
				customFields.GET("/:fieldId", customFieldH.Get)
				customFields.PUT("/:fieldId", customFieldH.Update)
				customFields.DELETE("/:fieldId", customFieldH.Delete)
				customFields.POST("/:fieldId/options", customFieldH.CreateOption)
				customFields.PUT("/:fieldId/options/:optionId", customFieldH.UpdateOption)
				customFields.DELETE("/:fieldId/options/:optionId", customFieldH.DeleteOption)
				customFields.POST("/issues/:issueId/values", customFieldH.SetIssueValue)
				customFields.GET("/issues/:issueId/values", customFieldH.ListIssueValues)
				customFields.POST("/issues/:issueId/values/bulk", customFieldH.BulkSetIssueValues)
				customFields.PUT("/issues/:issueId/values/:fieldId", customFieldH.UpdateIssueValue)
				customFields.DELETE("/issues/:issueId/values/:fieldId", customFieldH.DeleteIssueValue)
				customFields.GET("/issues/:issueId/fields", customFieldH.GetIssueFieldsWithValues)
		}
// ---- Project Templates (protected) ----
			templates := v1.Group("/templates", authMiddleware)
			{
				templates.POST("", templateH.Create)
				templates.GET("", templateH.List)
				templates.GET("/:templateId", templateH.Get)
				templates.PUT("/:templateId", templateH.Update)
				templates.DELETE("/:templateId", templateH.Delete)
				templates.POST("/:templateId/types", templateH.AddType)
				templates.DELETE("/:templateId/types/:typeId", templateH.RemoveType)
				templates.POST("/:templateId/apply", templateH.Apply)
			}
// ---- Type Templates (Workspace-level) ----
			typeTemplates := v1.Group("/type-templates", authMiddleware)
			{
				typeTemplates.POST("", typeTemplateH.Create)
				typeTemplates.GET("", typeTemplateH.List)
				typeTemplates.GET("/:id", typeTemplateH.Get)
				typeTemplates.PUT("/:id", typeTemplateH.Update)
				typeTemplates.DELETE("/:id", typeTemplateH.Delete)
				typeTemplates.POST("/:id/fields", typeTemplateH.BindField)
				typeTemplates.DELETE("/:id/fields/:fieldId", typeTemplateH.UnbindField)
			}
	}
}
