package testutil

import (
	"time"

	"github.com/reqmango/backend/internal/model"
)

// NewTestUser returns a user fixture for tests.
func NewTestUser() *model.User {
	return &model.User{
		BaseModel: model.BaseModel{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Email:     "test@example.com",
		Username:  "testuser",
	}
}

// NewTestWorkspace returns a workspace fixture for tests.
func NewTestWorkspace() *model.Workspace {
	return &model.Workspace{
		BaseModel: model.BaseModel{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Name:      "Test Workspace",
		Slug:      "test-workspace",
		OwnerID:   1,
	}
}

// NewTestProject returns a project fixture for tests.
func NewTestProject(workspaceID uint64) *model.Project {
	return &model.Project{
		BaseModel:   model.BaseModel{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Name:        "Test Project",
		Identifier:  "TEST",
		WorkspaceID: workspaceID,
	}
}

// NewTestState returns a state fixture for tests.
func NewTestState(projectID, workspaceID uint64) *model.State {
	return &model.State{
		BaseModel:   model.BaseModel{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Name:        "Backlog",
		Group:       "backlog",
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
		IsDefault:   true,
	}
}
