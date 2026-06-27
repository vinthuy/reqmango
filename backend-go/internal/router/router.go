package router

import (
	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/config"
	"github.com/reqmanpy/backend-go/internal/handler"
	"github.com/reqmanpy/backend-go/internal/middleware"
	"github.com/reqmanpy/backend-go/internal/rql"
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
	notificationSvc := service.NewNotificationService(db)
	webhookSvc := service.NewWebhookService(db)
	issueSvc := service.NewIssueService(db, notificationSvc, webhookSvc)
	cycleSvc := service.NewCycleService(db)
	moduleSvc := service.NewModuleService(db)
	issueTypeSvc := service.NewIssueTypeService(db)
	customFieldSvc := service.NewCustomFieldService(db)
	templateSvc := service.NewProjectTemplateService(db)
	typeTemplateSvc := service.NewTypeTemplateService(db)
	relationSvc := service.NewRelationService(db)
	workflowSvc := service.NewWorkflowService(db)
	commentSvc := service.NewCommentService(db, notificationSvc)
	savedViewSvc := service.NewSavedViewService(db)
	projectIssueTypeH := handler.NewProjectIssueTypeHandler(issueTypeSvc)
	pageSvc := service.NewPageService(db)
	witSvc := service.NewWorkItemTemplateService(db)
	conditionalFieldSvc := service.NewConditionalFieldService(db)
	releaseSvc := service.NewReleaseService(db)
	estimateSvc := service.NewEstimateService(db)
	attachmentSvc := service.NewAttachmentService(db)
	timeTrackSvc := service.NewTimeTrackService(db)
	recurrenceSvc := service.NewRecurrenceService(db)
	intakeH := handler.NewIntakeHandler(db)
	llmClient := service.NewLLMClient(cfg.AIAPIKey, cfg.AIModel, cfg.AIBaseURL, cfg.AIProvider)
	aiSvc := service.NewAIService(db, llmClient, issueSvc, projectSvc)

	// Initialize handlers
	authH := handler.NewAuthHandler(authSvc)
	workspaceH := handler.NewWorkspaceHandler(workspaceSvc)
	projectH := handler.NewProjectHandler(projectSvc, templateSvc)
	settingsH := handler.NewProjectSettingsHandler(settingsSvc)
	issueH := handler.NewIssueHandler(issueSvc)
	cycleH := handler.NewCycleHandler(cycleSvc)
	moduleH := handler.NewModuleHandler(moduleSvc)
	issueTypeH := handler.NewIssueTypeHandler(issueTypeSvc)
	customFieldH := handler.NewCustomFieldHandler(customFieldSvc)
	templateH := handler.NewProjectTemplateHandler(templateSvc)
	typeTemplateH := handler.NewTypeTemplateHandler(typeTemplateSvc)
	relationH := handler.NewRelationHandler(relationSvc)
	workflowH := handler.NewWorkflowHandler(workflowSvc)
	commentH := handler.NewCommentHandler(commentSvc)
	notificationH := handler.NewNotificationHandler(notificationSvc)
	savedViewH := handler.NewSavedViewHandler(savedViewSvc)
	pageH := handler.NewPageHandler(pageSvc)
	conditionalFieldH := handler.NewConditionalFieldHandler(conditionalFieldSvc)
	witH := handler.NewWorkItemTemplateHandler(witSvc)
	releaseH := handler.NewReleaseHandler(releaseSvc)
	estimateH := handler.NewEstimateHandler(estimateSvc)
	attachmentH := handler.NewAttachmentHandler(attachmentSvc)
	webhookH := handler.NewWebhookHandler(webhookSvc)
	timeTrackH := handler.NewTimeTrackHandler(timeTrackSvc)
	recurrenceH := handler.NewRecurrenceHandler(recurrenceSvc)
	aiH := handler.NewAIHandler(aiSvc, db)

	// JWT middleware
	authMiddleware := middleware.AuthMiddleware(db, cfg.SecretKey)

	// Rate limiter: global middleware (applied before routes)
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitWindowSec)
	r.Use(rateLimiter.Middleware())

	// ==================== API v1 ====================
	v1 := r.Group("/api/v1")
	{
		// ---- Intake (public) ----
		v1.POST("/intake/:projectId", intakeH.Submit)

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
			workspaces.PATCH("/:wsParam", workspaceH.Update)      // numeric ID or slug
			workspaces.DELETE("/:wsParam", workspaceH.Delete)     // numeric ID or slug
			workspaces.GET("/:wsParam/members", workspaceH.ListMembers)
			workspaces.POST("/:wsParam/members", workspaceH.AddMember)
			workspaces.PATCH("/:wsParam/members/:userId", workspaceH.UpdateMember)
			workspaces.DELETE("/:wsParam/members/:userId", workspaceH.RemoveMember)
			workspaces.GET("/:wsParam/ai-config", aiH.GetAIConfig)
			workspaces.PUT("/:wsParam/ai-config", aiH.UpdateAIConfig)
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
	projects.GET("/:projectId/intake", intakeH.ListPending)
	projects.GET("/:projectId/webhooks", webhookH.List)
	projects.POST("/:projectId/webhooks", webhookH.Create)
	projects.PUT("/:projectId/webhooks/:id", webhookH.Update)
	projects.DELETE("/:projectId/webhooks/:id", webhookH.Delete)
	projects.POST("/:projectId/intake/:issueId/triage", intakeH.Triage)
	projects.POST("/:projectId/intake/:issueId/ai-analyze", aiH.TriageAnalyze)
			projects.GET("/:projectId/issues-summary", projectH.GetIssuesSummary)
			projects.PATCH("/:projectId/lead", projectH.UpdateProjectLead)
			projects.GET("/:projectId/subscribers", projectH.ListSubscribers)
			projects.POST("/:projectId/subscribers", projectH.AddSubscriber)
			projects.DELETE("/:projectId/subscribers/:userId", projectH.RemoveSubscriber)
			projects.GET("/:projectId/cycles", cycleH.List)
			projects.POST("/:projectId/cycles", cycleH.Create)     // ?workspace_id=

			// ---- Pages ----
			pages := projects.Group("/:projectId/pages", authMiddleware)
			{
				pages.GET("", pageH.List)
				pages.POST("", pageH.Create)
				pages.GET("/tree", pageH.GetTree)
				pages.GET("/:pageId", pageH.Get)
				pages.PUT("/:pageId", pageH.Update)
				pages.DELETE("/:pageId", pageH.Delete)
				pages.POST("/:pageId/archive", pageH.Archive)
				pages.POST("/:pageId/restore", pageH.Restore)
				pages.POST("/:pageId/move", pageH.Move)
				pages.GET("/:pageId/children", pageH.ListChildren)
			}

			// ---- Project Issue Types ----
			projTypes := projects.Group("/:projectId/issue-types", authMiddleware)
			{
				projTypes.GET("", projectIssueTypeH.ListProjectTypes)
				projTypes.POST("", projectIssueTypeH.CreateProjectType)
				projTypes.POST("/copy-from-workspace", projectIssueTypeH.CopyFromWorkspace)
				projTypes.PATCH("/reorder", projectIssueTypeH.Reorder)
			}

			// ---- Saved Views ----
			views := projects.Group("/:projectId/views", authMiddleware)
			{
				views.GET("", savedViewH.List)
				views.POST("", savedViewH.Create)
				views.GET("/:viewId", savedViewH.Get)
				views.PUT("/:viewId", savedViewH.Update)
				views.DELETE("/:viewId", savedViewH.Delete)
				views.POST("/:viewId/set-default", savedViewH.SetDefault)
				views.POST("/:viewId/duplicate", savedViewH.Duplicate)
			}

			// ---- Work Item Templates ----
			templates := projects.Group("/:projectId/work-item-templates", authMiddleware)
			{
				templates.GET("", witH.List)
				templates.POST("", witH.Create)
				templates.GET("/:id", witH.Get)
				templates.PUT("/:id", witH.Update)
				templates.DELETE("/:id", witH.Delete)
			}

			// ---- Releases ----
			releases := projects.Group("/:projectId/releases", authMiddleware)
			{
				releases.GET("", releaseH.List)
				releases.POST("", releaseH.Create)
				releases.GET("/:releaseId", releaseH.Get)
				releases.PUT("/:releaseId", releaseH.Update)
				releases.DELETE("/:releaseId", releaseH.Delete)
				releases.POST("/:releaseId/issues", releaseH.AddIssues)
				releases.DELETE("/:releaseId/issues", releaseH.RemoveIssues)
				releases.GET("/:releaseId/progress", releaseH.GetProgress)
			}

			// ---- Estimates ----
			estimates := projects.Group("/:projectId/estimate-points", authMiddleware)
			{
				estimates.GET("/settings", estimateH.GetSettings)
				estimates.PUT("/settings", estimateH.UpdateSettings)
				estimates.GET("", estimateH.ListPoints)
				estimates.POST("", estimateH.CreatePoint)
				estimates.GET("/:pointId", estimateH.GetPoint)
				estimates.PATCH("/:pointId", estimateH.UpdatePoint)
				estimates.DELETE("/:pointId", estimateH.DeletePoint)
				estimates.POST("/reorder", estimateH.ReorderPoints)
				estimates.POST("/bulk", estimateH.BulkCreatePoints)
				estimates.POST("/defaults", estimateH.CreateDefaultPoints)
			}

			estimateCategories := projects.Group("/:projectId/estimate-categories", authMiddleware)
			{
				estimateCategories.GET("", estimateH.ListCategories)
				estimateCategories.POST("", estimateH.CreateCategory)
				estimateCategories.POST("/defaults", estimateH.CreateDefaultCategories)
			}

			estimateTime := projects.Group("/:projectId/estimate-time", authMiddleware)
			{
				estimateTime.GET("", estimateH.ListTime)
				estimateTime.POST("", estimateH.CreateTime)
				estimateTime.POST("/defaults", estimateH.CreateDefaultTime)
			}

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
			issues.GET("/tree", issueH.Tree)                     // ?project_id=&search=&limit=&offset=
			issues.GET("/statistics", issueH.GetStatistics)       // ?project_id=
			issues.GET("/search", issueH.Search)                  // ?workspace_id=&query=
			issues.POST("/bulk/update", issueH.BulkUpdate)        // ?project_id=
			issues.POST("/bulk/delete", issueH.BulkDelete)
			issues.POST("/bulk/copy", issueH.BulkCopy)
			issues.POST("/bulk/move", issueH.BulkMove)
			issues.POST("/merge", issueH.MergeDuplicates)

			// Import
			issues.POST("/import/json", issueH.ImportJSON)   // ?project_id=&workspace_id=
			issues.POST("/import/csv", issueH.ImportCSV)     // ?project_id=&workspace_id=

			// Single issue
			issues.GET("/:issueId", issueH.Get)
			issues.PUT("/:issueId", issueH.Update)
			issues.DELETE("/:issueId", issueH.Delete)
			issues.POST("/:issueId/archive", issueH.Archive)
			issues.GET("/:issueId/children", issueH.Children)     // tree lazy-load children
			issues.POST("/:issueId/restore", issueH.Restore)
			issues.POST("/:issueId/convert-type", issueH.ConvertType)
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

	// Time Tracks
	issues.POST("/:issueId/time-tracks/start", timeTrackH.Start)
	issues.POST("/:issueId/time-tracks/stop", timeTrackH.Stop)
	issues.GET("/:issueId/time-tracks", timeTrackH.List)
	issues.GET("/:issueId/time-tracks/summary", timeTrackH.Summary)
	issues.DELETE("/:issueId/time-tracks/:id", timeTrackH.Delete)

	// Recurrence
	issues.POST("/:issueId/recurrence", recurrenceH.Create)
	issues.GET("/:issueId/recurrence", recurrenceH.Get)
	issues.PUT("/:issueId/recurrence", recurrenceH.Update)
	issues.DELETE("/:issueId/recurrence", recurrenceH.Delete)

			// Pages
			issues.GET("/:issueId/pages", issueH.ListPages)
			issues.POST("/:issueId/pages", issueH.AddPage)                // ?page_id=
			issues.DELETE("/:issueId/pages", issueH.RemovePage)           // ?page_id=

			// Attachments
			issues.GET("/:issueId/attachments", attachmentH.ListByIssue)
			issues.POST("/:issueId/attachments", attachmentH.Create)
			issues.GET("/:issueId/attachments/:attachmentId", attachmentH.Get)
			issues.DELETE("/:issueId/attachments/:attachmentId", attachmentH.Delete)
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
				issueTypes.PUT("/:typeId/fields/:fieldId", issueTypeH.UpdateField)
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

// ---- Conditional Fields (protected) ----
		conditionalFields := v1.Group("/conditional-fields", authMiddleware)
		{
			conditionalFields.POST("", conditionalFieldH.Create)                            // ?workspace_id=
			conditionalFields.GET("", conditionalFieldH.List)                               // ?workspace_id=&field_id=
			conditionalFields.GET("/:id", conditionalFieldH.Get)                           // ?workspace_id=
			conditionalFields.PUT("/:id", conditionalFieldH.Update)                         // ?workspace_id=
			conditionalFields.DELETE("/:id", conditionalFieldH.Delete)                     // ?workspace_id=
			conditionalFields.POST("/evaluate", conditionalFieldH.EvaluateFieldVisibility)  // ?workspace_id=
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
// ---- Relations (Workspace-level) ----
			relations := v1.Group("/relations", authMiddleware)
			{
				relations.POST("/types", relationH.CreateType)
				relations.GET("/types", relationH.ListTypes)
				relations.PUT("/types/:id", relationH.UpdateType)
				relations.DELETE("/types/:id", relationH.DeleteType)
			}
			// ---- Issue Relations ----
			issues.POST("/:issueId/relations", relationH.CreateRelation)
			issues.GET("/:issueId/relations", relationH.ListRelations)
			v1.DELETE("/relations/:relationId", authMiddleware, relationH.DeleteRelation)
// ---- Workflows (protected) ----
			workflows := v1.Group("/projects/:projectId/workflows", authMiddleware)
			{
				workflows.POST("", workflowH.CreateWorkflow)
				workflows.GET("", workflowH.ListWorkflows)
				workflows.GET("/:workflowId", workflowH.GetWorkflow)
				workflows.PUT("/:workflowId", workflowH.UpdateWorkflow)
				workflows.DELETE("/:workflowId", workflowH.DeleteWorkflow)
				workflows.POST("/:workflowId/transitions", workflowH.AddTransition)
				workflows.PUT("/:workflowId/transitions/:transitionId", workflowH.UpdateTransition)
				workflows.DELETE("/:workflowId/transitions/:transitionId", workflowH.DeleteTransition)
			}
			// ---- Comments (protected) ----
			comments := v1.Group("/comments", authMiddleware)
			{
				comments.POST("", commentH.Create)
				comments.GET("/issue/:issueId", commentH.ListByIssue)
				comments.GET("/:commentId", commentH.Get)
				comments.PATCH("/:commentId", commentH.Update)
				comments.DELETE("/:commentId", commentH.Delete)
				comments.POST("/:commentId/resolve", commentH.Resolve)
				comments.POST("/:commentId/unresolve", commentH.Unresolve)
			}
			// ---- RQL (protected) ----
			v1.POST("/pages/:pageId/ai", authMiddleware, aiH.PageAI)

	rqlHandler := rql.NewRQLHandler(db)
			rqlGroup := v1.Group("/rql", authMiddleware)
			{
				rqlGroup.POST("/search", rqlHandler.Search)
			}
			// ---- Automations (protected) ----
			automations := v1.Group("/projects/:projectId/automations", authMiddleware)
			{
				automations.POST("", workflowH.CreateAutomation)
				automations.GET("", workflowH.ListAutomations)
				automations.PUT("/:id", workflowH.UpdateAutomation)
				automations.DELETE("/:id", workflowH.DeleteAutomation)
			}
			// ---- AI (protected) ----
			aiGroup := projects.Group("/:projectId/ai", authMiddleware)
			{
				aiGroup.POST("/chat", aiH.Chat)
				aiGroup.POST("/search", aiH.Search)
				aiGroup.POST("/create", aiH.CreatePreview)
				aiGroup.POST("/analyze", aiH.Analyze)
			}

			// ---- Notifications (protected) ----
			notifications := v1.Group("/notifications", authMiddleware)
			{
				notifications.GET("", notificationH.List)
				notifications.GET("/summary", notificationH.GetSummary)
				notifications.GET("/:id", notificationH.Get)
				notifications.POST("", notificationH.Create)
				notifications.POST("/bulk", notificationH.CreateBulk)
				notifications.PATCH("/:id/read", notificationH.MarkRead)
				notifications.POST("/read-all", notificationH.MarkAllRead)
				notifications.DELETE("/:id", notificationH.Delete)
			}
	}
}
