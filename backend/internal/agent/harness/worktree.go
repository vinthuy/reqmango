package harness

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// Worktree represents an isolated git worktree for agent execution.
type Worktree struct {
	Path   string
	Name   string
	Branch string
}

// WorktreePool manages a pool of git worktrees for parallel agent execution.
type WorktreePool struct {
	mu           sync.Mutex
	repoPath     string
	baseDir      string
	maxWorktrees int
	active       map[string]*Worktree
	sem          chan struct{}
}

// NewWorktreePool creates a new worktree pool.
func NewWorktreePool(repoPath string, maxWorktrees int) (*WorktreePool, error) {
	baseDir := filepath.Join(repoPath, ".claude", "worktrees")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create worktree base dir: %w", err)
	}

	if maxWorktrees <= 0 {
		maxWorktrees = 8
	}

	return &WorktreePool{
		repoPath:     repoPath,
		baseDir:      baseDir,
		maxWorktrees: maxWorktrees,
		active:       make(map[string]*Worktree),
		sem:          make(chan struct{}, maxWorktrees),
	}, nil
}

// Acquire creates a new git worktree for an agent. Blocks if pool is full.
func (p *WorktreePool) Acquire(name string) (*Worktree, error) {
	p.sem <- struct{}{} // blocks if pool full

	p.mu.Lock()
	defer p.mu.Unlock()

	branchName := fmt.Sprintf("agent-harness/%s", name)
	wtPath := filepath.Join(p.baseDir, name)

	// Check if already exists (from a previous run)
	if _, err := os.Stat(wtPath); err == nil {
		// Remove stale worktree first
		p.cleanupWorktree(name)
	}

	// Create the worktree
	cmd := exec.Command("git", "worktree", "add", "-b", branchName, wtPath, "HEAD")
	cmd.Dir = p.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		<-p.sem // release semaphore
		return nil, fmt.Errorf("git worktree add failed: %w\nOutput: %s", err, string(output))
	}

	wt := &Worktree{
		Path:   wtPath,
		Name:   name,
		Branch: branchName,
	}
	p.active[name] = wt
	return wt, nil
}

// Release removes a worktree from the pool.
func (p *WorktreePool) Release(wt *Worktree) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.active, wt.Name)
	<-p.sem // release semaphore

	return p.cleanupWorktree(wt.Name)
}

// cleanupWorktree removes a worktree directory and its git branch.
func (p *WorktreePool) cleanupWorktree(name string) error {
	wtPath := filepath.Join(p.baseDir, name)

	// Remove the worktree via git
	cmd := exec.Command("git", "worktree", "remove", "--force", wtPath)
	cmd.Dir = p.repoPath
	cmd.Run() // ignore errors — may already be cleaned

	// Clean up directory if still exists
	os.RemoveAll(wtPath)

	// Try to delete the branch
	delCmd := exec.Command("git", "branch", "-D", fmt.Sprintf("agent-harness/%s", name))
	delCmd.Dir = p.repoPath
	delCmd.Run() // ignore errors

	return nil
}

// Cleanup removes all worktrees in the pool.
func (p *WorktreePool) Cleanup() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for name := range p.active {
		p.cleanupWorktree(name)
	}
	p.active = make(map[string]*Worktree)
	return nil
}

// Active returns the number of active worktrees.
func (p *WorktreePool) Active() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.active)
}

// GetRepoPath returns the repository path.
func (p *WorktreePool) GetRepoPath() string {
	return p.repoPath
}

// IsAvailable checks if the pool has capacity.
func (p *WorktreePool) IsAvailable() bool {
	return len(p.sem) < p.maxWorktrees
}
