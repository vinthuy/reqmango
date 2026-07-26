package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/reqmango/backend/internal/service"
)

// MemoryScheduler handles scheduled memory cleanup tasks
type MemoryScheduler struct {
	memSvc   *service.MemoryService
	running  bool
	cancelFn context.CancelFunc
	mu       sync.Mutex
}

// NewMemoryScheduler creates a new memory scheduler
func NewMemoryScheduler(memSvc *service.MemoryService) *MemoryScheduler {
	return &MemoryScheduler{
		memSvc:  memSvc,
		running: false,
	}
}

// Start starts the scheduler with daily cleanup at 2:00 AM
func (s *MemoryScheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	ctx, cancelFn := context.WithCancel(ctx)
	s.mu.Lock()
	s.cancelFn = cancelFn
	s.mu.Unlock()

	go s.runScheduler(ctx)
}

// Stop stops the scheduler
func (s *MemoryScheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	if s.cancelFn != nil {
		s.cancelFn()
	}
	s.running = false
	s.mu.Unlock()
}

// runScheduler runs the scheduled tasks
func (s *MemoryScheduler) runScheduler(ctx context.Context) {
	scheduleDailyCleanup(ctx, s.memSvc)
}

// scheduleDailyCleanup schedules cleanup to run at 2:00 AM daily
func scheduleDailyCleanup(ctx context.Context, memSvc *service.MemoryService) {
	for {
		now := time.Now()
		next := now.Truncate(24 * time.Hour).Add(2 * time.Hour)
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}

		slog.Info("Scheduling next memory cleanup", "next_run", next)

		timer := time.NewTimer(next.Sub(now))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			runDailyCleanup(ctx, memSvc)
		}
	}
}

// runDailyCleanup performs the daily memory cleanup
func runDailyCleanup(ctx context.Context, memSvc *service.MemoryService) {
	slog.Info("Starting daily memory cleanup")

	expiredCount, err := memSvc.PruneExpiredMemories(ctx)
	if err != nil {
		slog.Error("Failed to prune expired memories", "error", err)
	} else {
		slog.Info("Pruned expired short-term memories", "count", expiredCount)
	}

	lowRelevanceCount, err := memSvc.PruneLowRelevanceMemories(ctx, 30, 0.3)
	if err != nil {
		slog.Error("Failed to prune low relevance memories", "error", err)
	} else {
		slog.Info("Pruned low relevance memories (30+ days, score < 0.3)", "count", lowRelevanceCount)
	}

	slog.Info("Daily memory cleanup completed")
}