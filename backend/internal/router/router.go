package router

import (
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/client"
	"github.com/reqmango/backend/internal/config"
	"github.com/reqmango/backend/internal/handler"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/rql"
	"github.com/reqmango/backend/internal/service"
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
	automationSvc := service.NewAutomationService(db)
	slackSvc := service.NewSlackService(db)
	issueSvc := service.NewIssueService(db, notificationSvc, webhookSvc, automationSvc, slackSvc)
	cycleSvc := service.NewCycleService(db, notificationSvc)
	moduleSvc := service.NewModuleService(db)
	issueTypeSvc := service.NewIssueTypeService(db)
	customFieldSvc := service.NewCustomFieldService(db)
	templateSvc := service.NewProjectTemplateService(db)
	typeTemplateSvc := service.NewTypeTemplateService(db)
	relationSvc := service.NewRelationService(db)
	workflowSvc := service.NewWorkflowService(db)
	commentSvc := service.NewCommentService(db, notificationSvc)
	savedViewSvc := service.NewSavedViewService(db)
	searchTemplateSvc := service.NewSearchTemplateService(db)
	gitSvc := service.NewGitService(db)
	projectIssueTypeH := handler.NewProjectIssueTypeHandler(issueTypeSvc)
	pageSvc := service.NewPageService(db)
	pageVersionSvc := service.NewPageVersionService(db)
	pageTemplateSvc := service.NewPageTemplateService(db)
	witSvc := service.NewWorkItemTemplateService(db)
	conditionalFieldSvc := service.NewConditionalFieldService(db)
	releaseSvc := service.NewReleaseService(db)
	estimateSvc := service.NewEstimateService(db)
	attachmentSvc := service.NewAttachmentService(db)
	timeTrackSvc := service.NewTimeTrackService(db)
	recurrenceSvc := service.NewRecurrenceService(db)
	reportSvc := service.NewReportService(db)
	savedReportSvc := service.NewSavedReportService(db)
	metricSvc := service.NewMetricService(db)
	pageTabSvc := service.NewProjectPageTabService(db)
	intakeH := handler.NewIntakeHandler(db)
	reportH := handler.NewReportHandler(reportSvc)
	savedReportH := handler.NewSavedReportHandler(savedReportSvc)
	metricH := handler.NewMetricHandler(metricSvc)
	dashboardSvc := service.NewDashboardService(db)
	dashboardH := handler.NewDashboardHandler(dashboardSvc)
	agentClient := client.NewAgentClient(cfg.AgentServiceURL, cfg.SecretKey)
	automationSvc.SetAgentService(agentClient)        // break circular dependency: automation -> agent -> issue -> automation
	commentSvc.SetAgentService(agentClient)           // enable @agent-name mention handling in comments
	commentSvc.SetAutomationService(automationSvc) // enable comment_added automation trigger
	mcpSvc := service.NewMCPService(db)
	githubSvc := service.NewGitHubService(db)
	roleSvc := service.NewRoleService(db)
	fieldPermSvc := service.NewFieldPermissionService(db)
	pluginSvc := service.NewPluginService(db)

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
	searchTemplateH := handler.NewSearchTemplateHandler(searchTemplateSvc)
	pageH := handler.NewPageHandler(pageSvc)
	pageVersionH := handler.NewPageVersionHandler(pageVersionSvc)
	pageTemplateH := handler.NewPageTemplateHandler(pageTemplateSvc)
	conditionalFieldH := handler.NewConditionalFieldHandler(conditionalFieldSvc)
	witH := handler.NewWorkItemTemplateHandler(witSvc)
	releaseH := handler.NewReleaseHandler(releaseSvc)
	initiativeH := handler.NewInitiativeHandler(db)
	estimateH := handler.NewEstimateHandler(estimateSvc)
	attachmentH := handler.NewAttachmentHandler(attachmentSvc)
	webhookH := handler.NewWebhookHandler(webhookSvc)
	timeTrackH := handler.NewTimeTrackHandler(timeTrackSvc)
	recurrenceH := handler.NewRecurrenceHandler(recurrenceSvc)
	pageTabH := handler.NewProjectPageTabHandler(pageTabSvc)
	mcpH := handler.NewMCPHandler(mcpSvc)
	githubH := handler.NewGitHubHandler(githubSvc)
	slackH := handler.NewSlackHandler(slackSvc)
	roleH := handler.NewRoleHandler(roleSvc, db)
	fieldPermH := handler.NewFieldPermissionHandler(fieldPermSvc)
	pluginH := handler.NewPluginHandler(pluginSvc)
	automationH := handler.NewAutomationHandler(automationSvc, db)
	gitIntegrationH := handler.NewGitIntegrationHandler(gitSvc)
	gitWebhookH := handler.NewGitWebhookHandler(gitSvc)

	// JWT middleware
	authMiddleware := middleware.AuthMiddleware(db, cfg.SecretKey)

	// Language detection middleware
	r.Use(middleware.LanguageMiddleware())

	// Rate limiter: global middleware (applied before routes)
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitWindowSec)
	r.Use(rateLimiter.Middleware())

	// SSE endpoint for real-time notifications
	sseH := handler.NewSSEHandler()
	v1SSE := r.Group("/api/v1")
	v1SSE.GET("/sse", authMiddleware, sseH.Connect)
		// ---- Internal Data API (for agent-service) ----
		internalH := handler.NewInternalDataHandler(db)
		internal := r.Group("/api/internal", authMiddleware)
		{
			internal.GET("/issues/:id", internalH.GetIssue)
			internal.POST("/issues/search", internalH.SearchIssues)
			internal.GET("/projects/:id", internalH.GetProject)
			internal.GET("/users/:id", internalH.GetUser)
		}

	// Agent Service Reverse Proxy — forwards to agent-service:8001
	agentProxy := reverseProxy(cfg.AgentServiceURL)

	// ==================== API v1 ====================
	v1 := r.Group("/api/v1")
	{
		// ---- Intake (public) ----
		v1.POST("/intake/:projectId", intakeH.Submit)

		// ---- GitHub Webhook (public) ----
		v1.POST("/webhook/github/:id", githubH.Webhook)

		// ---- Git Webhook (public) ----
		v1.POST("/webhook/git/:projectId", gitWebhookH.GitHubWebhook)

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
			workspaces.GET("/:wsParam", workspaceH.Get)       // slug or numeric ID
			workspaces.PATCH("/:wsParam", workspaceH.Update)  // numeric ID or slug
			workspaces.DELETE("/:wsParam", workspaceH.Delete) // numeric ID or slug
			workspaces.GET("/:wsParam/members", workspaceH.ListMembers)
			workspaces.POST("/:wsParam/members", workspaceH.AddMember)
			workspaces.PATCH("/:wsParam/members/:userId", workspaceH.UpdateMember)
			workspaces.DELETE("/:wsParam/members/:userId", workspaceH.RemoveMember)
			workspaces.Any("/:wsParam/ai-config", agentProxy)

			// Initiatives
			workspaces.POST("/:wsParam/initiatives", initiativeH.Create)
			workspaces.GET("/:wsParam/initiatives", initiativeH.List)
			workspaces.GET("/:wsParam/initiatives/search", initiativeH.Search)

			// Agent routes -- proxied to agent-service:8001
			// Wildcard does not match the bare collection path, so register it explicitly.
			workspaces.Any("/:wsParam/agents", agentProxy)
			workspaces.Any("/:wsParam/agents/*path", agentProxy)

			// Agent Loops
			workspaces.Any("/:wsParam/loops", agentProxy)
			workspaces.Any("/:wsParam/loops/*path", agentProxy)
			workspaces.Any("/:wsParam/pipelines", agentProxy)
			workspaces.Any("/:wsParam/pipelines/*path", agentProxy)
			workspaces.Any("/:wsParam/agent-sessions", agentProxy)
			workspaces.Any("/:wsParam/agent-sessions/*path", agentProxy)

			// MCP Server
			workspaces.GET("/:wsParam/mcp", mcpH.List)
			workspaces.POST("/:wsParam/mcp", mcpH.Create)
			workspaces.GET("/:wsParam/mcp/:id", mcpH.Get)
			workspaces.PUT("/:wsParam/mcp/:id", mcpH.Update)
			workspaces.DELETE("/:wsParam/mcp/:id", mcpH.Delete)
			workspaces.POST("/:wsParam/mcp/:id/discover", mcpH.DiscoverTools)
			workspaces.GET("/:wsParam/mcp/:id/tools", mcpH.GetTools)
			workspaces.POST("/:wsParam/mcp/:id/execute", mcpH.ExecuteTool)

			// GitHub integration
			workspaces.GET("/:wsParam/github", githubH.List)
			workspaces.POST("/:wsParam/github", githubH.Create)
			workspaces.GET("/:wsParam/github/:id", githubH.Get)
			workspaces.PUT("/:wsParam/github/:id", githubH.Update)
			workspaces.DELETE("/:wsParam/github/:id", githubH.Delete)
			workspaces.POST("/:wsParam/github/:id/sync", githubH.SyncIssues)

			// Git integration
			workspaces.POST("/:wsParam/git-integration", gitIntegrationH.CreateIntegration)
			workspaces.GET("/:wsParam/git-integration", gitIntegrationH.GetIntegration)
			workspaces.PUT("/:wsParam/git-integration", gitIntegrationH.UpdateIntegration)
			workspaces.DELETE("/:wsParam/git-integration", gitIntegrationH.DeleteIntegration)
			workspaces.GET("/:wsParam/issues/:issueId/git-links", gitIntegrationH.GetIssueGitLinks)

			// Slack integration
			workspaces.GET("/:wsParam/slack", slackH.List)
			workspaces.POST("/:wsParam/slack", slackH.Create)
			workspaces.GET("/:wsParam/slack/:id", slackH.Get)
			workspaces.PUT("/:wsParam/slack/:id", slackH.Update)
			workspaces.DELETE("/:wsParam/slack/:id", slackH.Delete)
			workspaces.POST("/:wsParam/slack/:id/notify", slackH.SendNotification)
			workspaces.POST("/:wsParam/slack/:id/test", slackH.TestNotification)

			// RBAC Roles & Permissions
			workspaces.GET("/:wsParam/roles", roleH.ListRoles)
			workspaces.POST("/:wsParam/roles", roleH.CreateRole)
			workspaces.PUT("/:wsParam/roles/:id", roleH.UpdateRole)
			workspaces.DELETE("/:wsParam/roles/:id", roleH.DeleteRole)

			// Field Permissions
			workspaces.GET("/:wsParam/field-permissions", fieldPermH.List)
			workspaces.POST("/:wsParam/field-permissions", fieldPermH.Create)
			workspaces.PUT("/:wsParam/field-permissions/:id", fieldPermH.Update)
			workspaces.DELETE("/:wsParam/field-permissions/:id", fieldPermH.Delete)
			workspaces.GET("/:wsParam/field-permissions/check", fieldPermH.CheckAccess)

			// Plugin management
			workspaces.GET("/:wsParam/plugins/catalog", pluginH.ListCatalog)
			workspaces.GET("/:wsParam/plugins", pluginH.ListInstalled)
			workspaces.POST("/:wsParam/plugins", pluginH.Install)
			workspaces.GET("/:wsParam/plugins/:id", pluginH.Get)
			workspaces.PUT("/:wsParam/plugins/:id", pluginH.Update)
			workspaces.DELETE("/:wsParam/plugins/:id", pluginH.Uninstall)
			workspaces.POST("/:wsParam/plugins/:id/enable", pluginH.Enable)
			workspaces.POST("/:wsParam/plugins/:id/disable", pluginH.Disable)
			workspaces.GET("/:wsParam/plugins/:id/logs", pluginH.GetEventLogs)
			workspaces.POST("/:wsParam/plugins/:id/test", pluginH.TestExecute)

			// Workspace-level Automation Rules
			workspaces.GET("/:wsParam/automations", automationH.ListWorkspace)
			workspaces.POST("/:wsParam/automations", automationH.CreateWorkspace)
			workspaces.GET("/:wsParam/automations/:id", automationH.Get)
			workspaces.PUT("/:wsParam/automations/:id", automationH.UpdateWorkspace)
			workspaces.DELETE("/:wsParam/automations/:id", automationH.DeleteWorkspace)

			// Workspace-level Workflows
			workspaces.GET("/:wsParam/workflows", workflowH.ListWorkspaceWorkflows)
			workspaces.POST("/:wsParam/workflows", workflowH.CreateWorkspaceWorkflow)
			workspaces.GET("/:wsParam/workflows/:workflowId", workflowH.GetWorkflow)
			workspaces.PUT("/:wsParam/workflows/:workflowId", workflowH.UpdateWorkspaceWorkflow)
			workspaces.DELETE("/:wsParam/workflows/:workflowId", workflowH.DeleteWorkspaceWorkflow)
			workspaces.POST("/:wsParam/workflows/:workflowId/transitions", workflowH.AddTransition)
			workspaces.PUT("/:wsParam/workflows/:workflowId/transitions/:transitionId", workflowH.UpdateTransition)
			workspaces.DELETE("/:wsParam/workflows/:workflowId/transitions/:transitionId", workflowH.DeleteTransition)

			// Workspace-level Settings: States
			workspaces.GET("/:wsParam/settings/states", settingsH.ListWorkspaceStates)
			workspaces.POST("/:wsParam/settings/states", settingsH.CreateWorkspaceState)
			workspaces.GET("/:wsParam/settings/states/:stateId", settingsH.GetWorkspaceState)
			workspaces.PUT("/:wsParam/settings/states/:stateId", settingsH.UpdateWorkspaceState)
			workspaces.DELETE("/:wsParam/settings/states/:stateId", settingsH.DeleteWorkspaceState)

			// Workspace-level Settings: Labels
			workspaces.GET("/:wsParam/settings/labels", settingsH.ListWorkspaceLabels)
			workspaces.POST("/:wsParam/settings/labels", settingsH.CreateWorkspaceLabel)
			workspaces.GET("/:wsParam/settings/labels/:labelId", settingsH.GetWorkspaceLabel)
			workspaces.PUT("/:wsParam/settings/labels/:labelId", settingsH.UpdateWorkspaceLabel)
			workspaces.DELETE("/:wsParam/settings/labels/:labelId", settingsH.DeleteWorkspaceLabel)

			// Workspace-level Modules
			workspaces.GET("/:wsParam/modules", moduleH.ListWorkspaceModules)
			workspaces.POST("/:wsParam/modules", moduleH.CreateWorkspaceModule)
			workspaces.GET("/:wsParam/modules/:moduleId", moduleH.GetWorkspaceModule)
			workspaces.PUT("/:wsParam/modules/:moduleId", moduleH.UpdateWorkspaceModule)
			workspaces.DELETE("/:wsParam/modules/:moduleId", moduleH.DeleteWorkspaceModule)
		}

		// Permissions (global, read-only)
		v1.GET("/permissions", authMiddleware, roleH.ListPermissions)

		// Initiatives (top-level routes)
		initiatives := v1.Group("/initiatives", authMiddleware)
		{
			initiatives.GET("/:initiativeId", initiativeH.Get)
			initiatives.PUT("/:initiativeId", initiativeH.Update)
			initiatives.DELETE("/:initiativeId", initiativeH.Delete)
			initiatives.GET("/:initiativeId/progress", initiativeH.GetProgress)
		}

		// ---- Projects (protected) ----
		projects := v1.Group("/projects", authMiddleware)
		{
			projects.POST("", projectH.Create) // ?workspace_id=
			projects.GET("", projectH.List)    // ?workspace_id=
			projects.GET("/:projectId", projectH.Get)
			projects.PATCH("/:projectId", projectH.Update)
			projects.DELETE("/:projectId", projectH.Delete)
			projects.POST("/:projectId/archive", projectH.Archive)
			projects.POST("/:projectId/restore", projectH.Restore)
			projects.GET("/:projectId/members", projectH.ListMembers)            // ?only_active=
			projects.POST("/:projectId/members", projectH.AddMember)             // ?user_id=&role=
			projects.PATCH("/:projectId/members/:userId", projectH.UpdateMember) // ?role=
			projects.DELETE("/:projectId/members/:userId", projectH.RemoveMember)
			projects.GET("/:projectId/statistics", projectH.GetStatistics)
			projects.POST("/:projectId/reports", reportH.Generate)
			projects.POST("/:projectId/reports/v2", reportH.GenerateV2)
			projects.GET("/:projectId/saved-reports", savedReportH.List)
			projects.POST("/:projectId/saved-reports", savedReportH.Create)
			projects.PATCH("/:projectId/saved-reports/:id", savedReportH.Update)
			projects.DELETE("/:projectId/saved-reports/:id", savedReportH.Delete)
			// Metrics
			projects.GET("/:projectId/metrics/templates", metricH.ListTemplates)
			projects.GET("/:projectId/metrics/charts", metricH.ListCharts)
			projects.GET("/:projectId/metrics/charts/:chartId", metricH.GetChart)
			projects.POST("/:projectId/metrics/charts", metricH.CreateChart)
			projects.PUT("/:projectId/metrics/charts/:chartId", metricH.UpdateChart)
			projects.DELETE("/:projectId/metrics/charts/:chartId", metricH.DeleteChart)
			projects.POST("/:projectId/metrics/charts/:chartId/render", metricH.RenderChart)
			projects.POST("/:projectId/metrics/preview", metricH.PreviewChart)
			projects.POST("/:projectId/metrics/charts/reorder", metricH.ReorderCharts)
			projects.GET("/:projectId/metrics/filter-values", metricH.GetFilterValues)
			projects.GET("/:projectId/metrics/custom-fields", metricH.GetCustomFields)
			projects.POST("/:projectId/custom-fields/:fieldId/enroll", customFieldH.EnrollField)
			projects.POST("/:projectId/custom-fields/:fieldId/unenroll", customFieldH.UnenrollField)
			projects.GET("/:projectId/intake", intakeH.ListPending)
			projects.GET("/:projectId/webhooks", webhookH.List)
			projects.POST("/:projectId/webhooks", webhookH.Create)
			projects.PUT("/:projectId/webhooks/:id", webhookH.Update)
			projects.DELETE("/:projectId/webhooks/:id", webhookH.Delete)
			projects.POST("/:projectId/intake/:issueId/triage", intakeH.Triage)
			projects.Any("/:projectId/intake/:issueId/ai-analyze", agentProxy)
			projects.GET("/:projectId/issues-summary", projectH.GetIssuesSummary)
			projects.PATCH("/:projectId/lead", projectH.UpdateProjectLead)
			projects.GET("/:projectId/subscribers", projectH.ListSubscribers)
			projects.POST("/:projectId/subscribers", projectH.AddSubscriber)
			projects.DELETE("/:projectId/subscribers/:userId", projectH.RemoveSubscriber)
			// Project Updates
			projectUpdateH := handler.NewProjectUpdateHandler(db)
			projects.GET("/:projectId/updates", projectUpdateH.List)
			projects.POST("/:projectId/updates", projectUpdateH.Create)
			projects.GET("/:projectId/cycles", cycleH.List)
			projects.GET("/:projectId/cycles/search", cycleH.Search)
			projects.POST("/:projectId/cycles", cycleH.Create) // ?workspace_id=

			// ---- Pages ----
			pages := projects.Group("/:projectId/pages", authMiddleware)
			{
				pages.GET("", pageH.List)
				pages.GET("/search", pageH.Search) // ?q=
				pages.POST("", pageH.Create)
				pages.GET("/tree", pageH.GetTree)
				pages.GET("/:pageId", pageH.Get)
				pages.PUT("/:pageId", pageH.Update)
				pages.DELETE("/:pageId", pageH.Delete)
				pages.POST("/:pageId/archive", pageH.Archive)
				pages.POST("/:pageId/restore", pageH.Restore)
				pages.POST("/:pageId/move", pageH.Move)
				pages.GET("/:pageId/children", pageH.ListChildren)
				pages.POST("/:pageId/lock", pageH.Lock)
				pages.POST("/:pageId/unlock", pageH.Unlock)
				pages.GET("/:pageId/export", pageH.Export)
				pages.POST("/:pageId/convert-to-issue", pageH.ConvertToIssue)
				pages.GET("/:pageId/versions", pageVersionH.List)
				pages.GET("/:pageId/versions/:versionId", pageVersionH.Get)
				pages.POST("/:pageId/versions/:versionId/restore", pageVersionH.Restore)
			}

			// ---- Page Templates ----
			pageTemplates := projects.Group("/:projectId/page-templates", authMiddleware)
			{
				pageTemplates.GET("", pageTemplateH.List)
				pageTemplates.POST("", pageTemplateH.Create)
				pageTemplates.GET("/:templateId", pageTemplateH.Get)
				pageTemplates.PUT("/:templateId", pageTemplateH.Update)
				pageTemplates.DELETE("/:templateId", pageTemplateH.Delete)
			}

			// ---- Project Time Tracking Summary ----
			projects.GET("/:projectId/time-tracks/summary", timeTrackH.ProjectSummary)

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

			// ---- Dashboards ----
			dashboards := projects.Group("/:projectId/dashboards", authMiddleware)
			{
				dashboards.GET("", dashboardH.List)
				dashboards.POST("", dashboardH.Create)
				dashboards.GET("/:id", dashboardH.Get)
				dashboards.PUT("/:id", dashboardH.Update)
				dashboards.DELETE("/:id", dashboardH.Delete)
				dashboards.POST("/:id/set-default", dashboardH.SetDefault)
				dashboards.POST("/:id/duplicate", dashboardH.Duplicate)
				dashboards.GET("/:id/full", dashboardH.GetFull)

				// Widget CRUD
				dashboards.POST("/:id/widgets", dashboardH.AddWidget)
				dashboards.PUT("/:id/widgets/:widgetId", dashboardH.UpdateWidget)
				dashboards.DELETE("/:id/widgets/:widgetId", dashboardH.DeleteWidget)
				dashboards.PUT("/:id/widgets/reorder", dashboardH.ReorderWidgets)
			}

			// ---- Search Templates ----
			searchTemplates := projects.Group("/:projectId/search-templates", authMiddleware)
			{
				searchTemplates.GET("", searchTemplateH.List)
				searchTemplates.POST("", searchTemplateH.Create)
				searchTemplates.GET("/:templateId", searchTemplateH.Get)
				searchTemplates.DELETE("/:templateId", searchTemplateH.Delete)
				searchTemplates.POST("/:templateId/apply", searchTemplateH.Apply)
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
				releases.GET("/search", releaseH.Search)
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
				settings.POST("/states", settingsH.CreateState)                 // ?workspace_id=
				settings.GET("/states", settingsH.ListStates)                   // ?include_inactive=
				settings.POST("/states/default", settingsH.CreateDefaultStates) // ?workspace_id=
				settings.GET("/states/:stateId", settingsH.GetState)
				settings.PUT("/states/:stateId", settingsH.UpdateState)
				settings.DELETE("/states/:stateId", settingsH.DeleteState)

				// Labels
				settings.POST("/labels", settingsH.CreateLabel) // ?workspace_id=
				settings.GET("/labels", settingsH.ListLabels)
				settings.GET("/labels/search", settingsH.SearchLabels) // ?q=
				settings.GET("/labels/:labelId", settingsH.GetLabel)
				settings.PUT("/labels/:labelId", settingsH.UpdateLabel)
				settings.DELETE("/labels/:labelId", settingsH.DeleteLabel)
			}

			// ---- Project Page Tabs ----
			pageTabs := projects.Group("/:projectId/page-tabs", authMiddleware)
			{
				pageTabs.GET("", pageTabH.List)
				pageTabs.POST("", pageTabH.Create)
				pageTabs.PUT("/batch", pageTabH.BatchSave)
				pageTabs.PUT("/reorder", pageTabH.Reorder)
				pageTabs.PUT("/:tabId", pageTabH.Update)
				pageTabs.DELETE("/:tabId", pageTabH.Delete)
			}
		}

		// ---- Test route ----
		v1.GET("/test-issues", authMiddleware, issueH.List)

		// ---- Issues (protected) ----
		issues := v1.Group("/issues", authMiddleware)
		{
			// CRUD
			issues.POST("", issueH.Create)                                                                            // ?project_id=&workspace_id=
			issues.GET("", issueH.List)                                                                               // ?project_id=&filters...
			issues.GET("/tree", issueH.Tree)                                                                          // ?project_id=&search=&limit=&offset=
			issues.GET("/statistics", issueH.GetStatistics)                                                           // ?project_id=
			issues.GET("/flow-metrics", issueH.GetFlowMetrics)                                                        // ?project_id=
			issues.GET("/search", issueH.Search)                                                                      // ?workspace_id=&query=
			issues.GET("/suggest", issueH.Suggest)                                                                    // ?project_id=&query=&limit=
			issues.POST("/bulk/update", middleware.RequirePermission(db, "issue:edit", "project"), issueH.BulkUpdate) // ?project_id=
			issues.POST("/bulk/delete", middleware.RequirePermission(db, "issue:delete", "project"), issueH.BulkDelete)
			issues.POST("/bulk/copy", middleware.RequirePermission(db, "issue:edit", "project"), issueH.BulkCopy)
			issues.POST("/bulk/move", middleware.RequirePermission(db, "issue:edit", "project"), issueH.BulkMove)
			issues.POST("/merge", issueH.MergeDuplicates)

			// Import
			issues.POST("/import/json", issueH.ImportJSON) // ?project_id=&workspace_id=
			issues.POST("/import/csv", issueH.ImportCSV)   // ?project_id=&workspace_id=
			issues.GET("/import/template", issueH.ExportCSVTemplate)

			// Export
			issues.GET("/export", issueH.Export) // ?project_id=&format=csv|json

			// Single issue
			issues.GET("/:issueId", issueH.Get)
			issues.PUT("/:issueId", middleware.RequirePermission(db, "issue:edit", "project"), issueH.Update)
			issues.DELETE("/:issueId", middleware.RequirePermission(db, "issue:delete", "project"), issueH.Delete)
			issues.POST("/:issueId/archive", middleware.RequirePermission(db, "issue:edit", "project"), issueH.Archive)
			issues.GET("/:issueId/children", issueH.Children) // tree lazy-load children
			issues.POST("/:issueId/reorder-sub-issues", middleware.RequirePermission(db, "issue:edit", "project"), issueH.ReorderSubIssues)
			issues.POST("/:issueId/watch", issueH.AddWatcher)
			issues.DELETE("/:issueId/watch", issueH.RemoveWatcher)
			issues.GET("/:issueId/watchers", issueH.ListWatchers)
			issues.POST("/:issueId/restore", middleware.RequirePermission(db, "issue:edit", "project"), issueH.Restore)
			issues.POST("/:issueId/convert-type", middleware.RequirePermission(db, "issue:edit", "project"), issueH.ConvertType)
			issues.GET("/:issueId/activities", issueH.GetActivities) // ?limit=&offset=

			// Assignees
			issues.POST("/:issueId/assignees", middleware.RequirePermission(db, "issue:edit", "project"), issueH.AddAssignee) // ?user_id=
			issues.DELETE("/:issueId/assignees/:userId", middleware.RequirePermission(db, "issue:edit", "project"), issueH.RemoveAssignee)

			// Labels
			issues.POST("/:issueId/labels", middleware.RequirePermission(db, "issue:edit", "project"), issueH.AddLabel) // ?label_id=
			issues.DELETE("/:issueId/labels/:labelId", middleware.RequirePermission(db, "issue:edit", "project"), issueH.RemoveLabel)

			// Cycle
			issues.POST("/:issueId/cycle", middleware.RequirePermission(db, "issue:edit", "project"), issueH.SetCycle) // ?cycle_id=
			issues.DELETE("/:issueId/cycle", middleware.RequirePermission(db, "issue:edit", "project"), issueH.RemoveCycle)

			// Time Tracks
			issues.POST("/:issueId/time-tracks/start", timeTrackH.Start)
			issues.POST("/:issueId/time-tracks/stop", timeTrackH.Stop)
			issues.GET("/:issueId/time-tracks", timeTrackH.List)
			issues.GET("/:issueId/time-tracks/summary", timeTrackH.Summary)
			issues.DELETE("/:issueId/time-tracks/:id", timeTrackH.Delete)

			// Recurrence
			issues.POST("/:issueId/recurrence", recurrenceH.Create)
			issues.GET("/:issueId/recurrence", recurrenceH.Get)
			issues.Any("/:issueId/ai/comment", agentProxy)
			// Agent routes -- proxied to agent-service:8001
			issues.Any("/:issueId/agents/*path", agentProxy)
			issues.PUT("/:issueId/recurrence", recurrenceH.Update)
			issues.DELETE("/:issueId/recurrence", recurrenceH.Delete)

			// Pages
			issues.GET("/:issueId/pages", issueH.ListPages)
			issues.POST("/:issueId/pages", middleware.RequirePermission(db, "issue:edit", "project"), issueH.AddPage)      // ?page_id=
			issues.DELETE("/:issueId/pages", middleware.RequirePermission(db, "issue:edit", "project"), issueH.RemovePage) // ?page_id=

			// Attachments
			issues.GET("/:issueId/attachments", attachmentH.ListByIssue)
			issues.POST("/:issueId/attachments", middleware.RequirePermission(db, "issue:edit", "project"), attachmentH.Create)
			issues.GET("/:issueId/attachments/:attachmentId", attachmentH.Get)
			issues.GET("/:issueId/attachments/:attachmentId/download", attachmentH.Download)
			issues.DELETE("/:issueId/attachments/:attachmentId", middleware.RequirePermission(db, "issue:edit", "project"), attachmentH.Delete)
		}

		// ---- Modules (protected) ----
		modules := v1.Group("/modules", authMiddleware)
		{
			// CRUD
			modules.POST("", moduleH.Create)       // ?workspace_id=
			modules.GET("", moduleH.List)          // ?project_id=&workspace_id=&include_archived=
			modules.GET("/search", moduleH.Search) // ?project_id=&workspace_id=&q=
			modules.GET("/:moduleId", moduleH.Get)
			modules.PUT("/:moduleId", moduleH.Update)
			modules.DELETE("/:moduleId", moduleH.Delete)

			// Issues
			modules.POST("/:moduleId/issues", moduleH.AddIssue) // ?issue_id=
			modules.DELETE("/:moduleId/issues/:issueId", moduleH.RemoveIssue)
			modules.GET("/:moduleId/issues", moduleH.ListIssues)

			// Analysis
			modules.GET("/:moduleId/progress", moduleH.GetProgress)
			modules.GET("/:moduleId/statistics", moduleH.GetStatistics)

			// Tree
			modules.GET("/tree", moduleH.GetTree) // ?project_id=
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
			// Automation
			cycles.POST("/:cycleId/apply-auto-add", cycleH.ApplyAutoAddRules)
			cycles.POST("/:cycleId/apply-auto-close", cycleH.ApplyAutoCloseRules)
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
			issueTypes.PATCH("/reorder-workspace", issueTypeH.ReorderWorkspace) // ?workspace_id=
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
			customFields.GET("/workspace-fields", customFieldH.ListWorkspaceFieldsWithEnrollment)
		}

		// ---- Conditional Fields (protected) ----
		conditionalFields := v1.Group("/conditional-fields", authMiddleware)
		{
			conditionalFields.POST("", conditionalFieldH.Create)                           // ?workspace_id=
			conditionalFields.GET("", conditionalFieldH.List)                              // ?workspace_id=&field_id=
			conditionalFields.GET("/:id", conditionalFieldH.Get)                           // ?workspace_id=
			conditionalFields.PUT("/:id", conditionalFieldH.Update)                        // ?workspace_id=
			conditionalFields.DELETE("/:id", conditionalFieldH.Delete)                     // ?workspace_id=
			conditionalFields.POST("/evaluate", conditionalFieldH.EvaluateFieldVisibility) // ?workspace_id=
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
		// ---- Workflows (protected with RBAC) ----
		workflows := v1.Group("/projects/:projectId/workflows", authMiddleware)
		{
			workflows.GET("", workflowH.ListWorkflows)
			workflows.GET("/:workflowId", workflowH.GetWorkflow)
			workflows.POST("", middleware.RequirePermission(db, "workflow:manage", "project"), workflowH.CreateWorkflow)
			workflows.PUT("/:workflowId", middleware.RequirePermission(db, "workflow:manage", "project"), workflowH.UpdateWorkflow)
			workflows.DELETE("/:workflowId", middleware.RequirePermission(db, "workflow:manage", "project"), workflowH.DeleteWorkflow)
			workflows.POST("/:workflowId/transitions", middleware.RequirePermission(db, "workflow:manage", "project"), workflowH.AddTransition)
			workflows.PUT("/:workflowId/transitions/:transitionId", middleware.RequirePermission(db, "workflow:manage", "project"), workflowH.UpdateTransition)
			workflows.DELETE("/:workflowId/transitions/:transitionId", middleware.RequirePermission(db, "workflow:manage", "project"), workflowH.DeleteTransition)
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
		v1.Any("/pages/:pageId/ai", authMiddleware, agentProxy)

		rqlHandler := rql.NewRQLHandler(db)
		rqlGroup := v1.Group("/rql", authMiddleware)
		{
			rqlGroup.POST("/search", rqlHandler.Search)
		}
		// ---- AI (proxied to agent-service) ----
		projects.Any("/:projectId/ai/*path", agentProxy)

		// Agent routes -- proxied to agent-service:8001
		projects.Any("/:projectId/agent/*path", agentProxy)

		// Automation rules
		projects.GET("/:projectId/automations", automationH.List)
		projects.POST("/:projectId/automations", automationH.Create)
		projects.GET("/:projectId/automations/:id", automationH.Get)
		projects.PUT("/:projectId/automations/:id", automationH.Update)
		projects.DELETE("/:projectId/automations/:id", automationH.Delete)
		projects.POST("/:projectId/automations/:id/execute", automationH.Execute)

		// Execution history
		v1.GET("/issues/:issueId/automation-history", authMiddleware, automationH.GetExecutionHistory)
		v1.GET("/automations/:ruleId/execution-history", authMiddleware, automationH.GetRuleExecutionHistory)
		v1.GET("/projects/:projectId/automation-executions", authMiddleware, automationH.GetProjectExecutionHistory)

		// ---- Notifications (protected) ----
		notifications := v1.Group("/notifications", authMiddleware)
		{
			notifications.GET("", notificationH.List)
			notifications.GET("/summary", notificationH.GetSummary)
			notifications.GET("/:id", notificationH.Get)
			notifications.POST("", notificationH.Create)
			notifications.POST("/bulk", notificationH.CreateBulk)
			notifications.POST("/check-due-reminders", notificationH.CheckDueReminders)
			notifications.PATCH("/:id/read", notificationH.MarkRead)
			notifications.POST("/read-all", notificationH.MarkAllRead)
			notifications.DELETE("/:id", notificationH.Delete)
		}
	}
}

// reverseProxy creates a Gin handler that proxies requests to the agent service.
func reverseProxy(targetURL string) gin.HandlerFunc {
	target, _ := url.Parse(targetURL)
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(c *gin.Context) {
		// Preserve the original path including /api/v1/workspaces/:wsParam/... prefix
		c.Request.URL.Host = target.Host
		c.Request.URL.Scheme = target.Scheme
		c.Request.URL.Path = c.Request.URL.Path
		c.Request.Header.Set("X-Forwarded-Host", c.Request.Host)
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
