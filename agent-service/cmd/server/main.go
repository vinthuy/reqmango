package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/agent-service/config"
	"github.com/reqmango/agent-service/internal/client"
	"github.com/reqmango/agent-service/internal/handler"
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
	); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	// Seed built-in agents
	reg := registry.NewRegistry(db)
	if err := reg.SeedDefaults(nil); err != nil {
		log.Printf("Warning: failed to seed default agents: %v", err)
	}

	// Create HTTP client to main backend
	agentClient := client.NewAgentClient(cfg.MainBackendURL)

	// Initialize services
	loopSvc := service.NewLoopService(db, agentClient)

	// Seed Sprint Guardian loop for workspaces
	// (Will be seeded on-demand when a workspace first accesses the API)

	// Initialize handlers
	loopH := handler.NewAgentLoopHandler(loopSvc)
	sessionH := handler.NewAgentSessionHandler(db)
	pipelineH := handler.NewAgentPipelineHandler(db, reg)

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
