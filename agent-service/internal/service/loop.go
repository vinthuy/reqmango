package service

import (
	"github.com/reqmango/agent-service/internal/client"
	"gorm.io/gorm"
)

// LoopService handles business logic for agent loops.
type LoopService struct {
	db    *gorm.DB
	agent *client.AgentClient
}

// NewLoopService creates a new LoopService.
func NewLoopService(db *gorm.DB, agent *client.AgentClient) *LoopService {
	return &LoopService{db: db, agent: agent}
}
