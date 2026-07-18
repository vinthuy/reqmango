package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/agent-service/config"
	"github.com/reqmango/agent-service/internal/client"
	"github.com/reqmango/agent-service/internal/handler"
	"github.com/reqmango/agent-service/internal/llm"
	"github.com/reqmango/agent-service/internal/middleware"
	"github.com/reqmango/agent-service/internal/model"
	"github.com/reqmango/agent-service/internal/registry"
	"github.com/reqmango/agent-service/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate agent-specific tables
	if err := db.AutoMigrate(
		&model.Loop{},
		&model.LoopRun{},
		&model.LoopIteration{},
		&model.AgentSession{},
		&model.Pipeline{},
		&model.PipelineRun{},
		&registry.AgentEntry{},
		// AI models (from Phase 1; tables already exist in shared DB)
		&model.Agent{},
		&model.AgentActivity{},
		&model.AIConfig{},
		&model.AIThread{},
		&model.AIMessage{},
	); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	// Seed built-in agents
	reg := registry.NewRegistry(db)
	if err := reg.SeedDefaults(nil); err != nil {
		log.Printf("Warning: failed to seed default agents: %v", err)
	}

	// Create HTTP clients to main backend
	agentClient := client.NewAgentClient(cfg.MainBackendURL)
	backendClient := client.NewBackendClient(cfg.MainBackendURL)

	// Initialize LLM client, AI service, and agent service
	llmClient := llm.NewLLMClient(cfg.AIAPIKey, cfg.AIModel, cfg.AIBaseURL, cfg.AIProvider)
	aiSvc := service.NewAIService(db, llmClient, backendClient)
	agentSvc := service.NewAgentService(db, llmClient, backendClient, aiSvc)

	// Initialize services
	loopSvc := service.NewLoopService(db, agentClient)

	// Seed Sprint Guardian loop for workspaces
	// (Will be seeded on-demand when a workspace first accesses the API)

	// Initialize handlers
	loopH := handler.NewAgentLoopHandler(loopSvc)
	sessionH := handler.NewAgentSessionHandler(db)
	pipelineH := handler.NewAgentPipelineHandler(db, reg)
	agentH := handler.NewAgentHandler(agentSvc)

	// Setup router
	r := gin.Default()
	auth := middleware.AuthMiddleware(cfg.SecretKey)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "agent-service"})
	})

	api := r.Group("/api/v1/workspaces/:wsParam", auth)
	{
		// Loop routes
		loops := api.Group("/loops")
		{
			loops.GET("", loopH.List)
			loops.POST("", loopH.Create)
			loops.GET("/:id", loopH.Get)
			loops.PUT("/:id", loopH.Update)
			loops.DELETE("/:id", loopH.Delete)
			loops.POST("/:id/start", loopH.Start)
			loops.GET("/:id/runs", loopH.GetRuns)
			loops.POST("/runs/:runId/stop", loopH.Stop)
			loops.GET("/runs/:runId", loopH.GetRun)
		}

		// Pipeline routes
		pipelines := api.Group("/pipelines")
		{
			pipelines.GET("", pipelineH.List)
			pipelines.POST("", pipelineH.Create)
			pipelines.GET("/:id", pipelineH.Get)
			pipelines.PUT("/:id", pipelineH.Update)
			pipelines.DELETE("/:id", pipelineH.Delete)
			pipelines.POST("/:id/run", pipelineH.Run)
			pipelines.GET("/:id/runs", pipelineH.GetRuns)
			pipelines.GET("/runs/:runId", pipelineH.GetRun)
		}

		// Registry routes
		registryGroup := api.Group("/agents/registry")
		{
			registryGroup.GET("", func(c *gin.Context) {
				wsID := getWorkspaceID(c)
				entries, _ := reg.ListByWorkspace(wsID)
				c.JSON(200, entries)
			})
		}

		// Session routes
		sessions := api.Group("/agent-sessions")
		{
			sessions.GET("", sessionH.List)
			sessions.GET("/:sessionId", sessionH.Get)
		}

		// Agent CRUD and dispatch routes
		agents := api.Group("/agents")
		{
			agents.GET("", agentH.List)
			agents.POST("", agentH.Create)
			agents.GET("/activity", agentH.ListWorkspaceActivity)
			agents.PATCH("/activity/:id/feedback", agentH.UpdateActivityFeedback)
			agents.GET("/:id", agentH.GetByID)
			agents.PUT("/:id", agentH.Update)
			agents.DELETE("/:id", agentH.Delete)
			agents.POST("/:id/dispatch", agentH.Dispatch)
			agents.GET("/:id/activity", agentH.GetActivity)
			agents.POST("/:id/auto-triage", agentH.AutoTriage)
			agents.POST("/:id/auto-assign", agentH.AutoAssign)
			agents.POST("/:id/mention", agentH.HandleMention)
		}
	}

	// Project-level agent routes (proxy preserved path)
	proj := r.Group("/api/v1/projects/:projectId", auth)
	{
		proj.POST("/agent/auto-triage", agentH.AutoTriageProject)
		proj.POST("/agent/auto-assign", agentH.AutoAssignProject)
	}

	// Issue-level agent mention
	issues := r.Group("/api/v1/issues/:issueId", auth)
	{
		issues.POST("/agents/:agentId/mention", agentH.HandleMention)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Agent Service starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getWorkspaceID(c *gin.Context) uint64 {
	wsParam := c.Param("wsParam")
	var id uint64
	fmt.Sscanf(wsParam, "%d", &id)
	return id
}
