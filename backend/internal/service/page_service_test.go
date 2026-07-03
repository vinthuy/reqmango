package service

import (
	"fmt"
	"testing"

	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NotNil(t, db)

	// Migrate tables
	err = db.AutoMigrate(
		&model.User{},
		&model.Workspace{},
		&model.Project{},
		&model.Page{},
		&model.PageVersion{},
		&model.PageTemplate{},
		&model.State{},
		&model.IssueType{},
		&model.Issue{},
	)
	require.NoError(t, err)

	return db
}

// seedTestData creates minimal test data (workspace, project, user).
func seedTestData(t *testing.T, db *gorm.DB) (workspaceID, projectID, userID uint64) {
	// User
	user := &model.User{DisplayName: "Test User", Email: "test@reqmango.com"}
	require.NoError(t, db.Create(user).Error)
	userID = user.ID

	// Workspace
	ws := &model.Workspace{Name: "Test WS", Slug: "test-ws"}
	require.NoError(t, db.Create(ws).Error)
	workspaceID = ws.ID

	// Project
	proj := &model.Project{Name: "Test Project", Identifier: "TEST", WorkspaceID: workspaceID}
	require.NoError(t, db.Create(proj).Error)
	projectID = proj.ID

	// Issue type
	it := &model.IssueType{Name: "Task", WorkspaceID: workspaceID}
	require.NoError(t, db.Create(it).Error)

	// Default state (required for convert-to-issue)
	state := &model.State{Name: "Backlog", Color: "#ccc", ProjectID: projectID, WorkspaceID: workspaceID}
	require.NoError(t, db.Create(state).Error)

	return
}

// ================ Page CRUD Tests ================

func TestPageService_CreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	wsID, projID, userID := seedTestData(t, db)
	svc := NewPageService(db)

	// Create
	req := &request.PageCreateRequest{Title: "Test Page", Content: "Hello world"}
	page, err := svc.Create(req, projID, wsID, userID)
	require.NoError(t, err)
	require.NotNil(t, page)
	assert.Equal(t, "Test Page", page.Title)
	assert.Equal(t, "Hello world", page.Content)
	assert.Equal(t, 0, page.Depth)
	assert.True(t, page.Published)

	// Get
	got, err := svc.Get(page.ID, projID)
	require.NoError(t, err)
	assert.Equal(t, page.ID, got.ID)
	assert.Equal(t, "Test Page", got.Title)

	t.Logf("✅ Create & Get: page ID=%d, title=%s", page.ID, page.Title)
}

func TestPageService_Update(t *testing.T) {
	db := setupTestDB(t)
	wsID, projID, userID := seedTestData(t, db)
	svc := NewPageService(db)

	// Create
	req := &request.PageCreateRequest{Title: "Original", Content: "v1"}
	page, err := svc.Create(req, projID, wsID, userID)
	require.NoError(t, err)

	// Update title + content
	newTitle := "Updated"
	newContent := "v2 updated"
	updated, err := svc.Update(page.ID, projID, userID, &request.PageUpdateRequest{
		Title:   &newTitle,
		Content: &newContent,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Title)
	assert.Equal(t, "v2 updated", updated.Content)

	// Check version was auto-created
	vs := NewPageVersionService(db)
	versions, err := vs.List(page.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(versions), 1, "Should auto-create a version snapshot on content change")
	t.Logf("✅ Update: auto-created %d versions", len(versions))
}

func TestPageService_Delete(t *testing.T) {
	db := setupTestDB(t)
	wsID, projID, userID := seedTestData(t, db)
	svc := NewPageService(db)

	// Create
	req := &request.PageCreateRequest{Title: "To Delete"}
	page, err := svc.Create(req, projID, wsID, userID)
	require.NoError(t, err)

	// Delete
	err = svc.Delete(page.ID, projID)
	require.NoError(t, err)

	// Should not be found
	_, err = svc.Get(page.ID, projID)
	assert.Error(t, err)

	t.Logf("✅ Delete: page %d soft-deleted", page.ID)
}

// ================ Tree & Hierarchy Tests ================

func TestPageService_Tree(t *testing.T) {
	db := setupTestDB(t)
	wsID, projID, userID := seedTestData(t, db)
	svc := NewPageService(db)

	// Create root + 2 children
	root, _ := svc.Create(&request.PageCreateRequest{Title: "Root"}, projID, wsID, userID)

	parentID := root.ID
	child1, _ := svc.Create(&request.PageCreateRequest{Title: "Child 1", ParentID: &parentID}, projID, wsID, userID)
	child2, _ := svc.Create(&request.PageCreateRequest{Title: "Child 2", ParentID: &parentID}, projID, wsID, userID)

	tree, err := svc.GetTree(projID)
	require.NoError(t, err)
	assert.Len(t, tree, 1)
	assert.Equal(t, "Root", tree[0].Title)
	assert.Len(t, tree[0].Children, 2)

	t.Logf("✅ Tree: root=%s, children=%d %s, %s",
		tree[0].Title, len(tree[0].Children),
		func() string {
			if len(tree[0].Children) > 0 {
				return tree[0].Children[0].Title
			}
			return ""
		}(),
		func() string {
			if len(tree[0].Children) > 1 {
				return tree[0].Children[1].Title
			}
			return ""
		}())

	// Prevent unused warnings
	_ = child1
	_ = child2
}

func TestPageService_Move(t *testing.T) {
	db := setupTestDB(t)
	wsID, projID, userID := seedTestData(t, db)
	svc := NewPageService(db)

	page, _ := svc.Create(&request.PageCreateRequest{Title: "Movable"}, projID, wsID, userID)
	newParent, _ := svc.Create(&request.PageCreateRequest{Title: "New Parent"}, projID, wsID, userID)

	parentID := newParent.ID
	moved, err := svc.Move(page.ID, projID, &request.PageMoveRequest{
		ParentID: &parentID,
		Sequence: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, &parentID, moved.ParentID)
	assert.Equal(t, 1, moved.Depth)

	t.Logf("✅ Move: page now under parent %d at depth %d", parentID, moved.Depth)
}

// ================ Archive & Restore ================

func TestPageService_ArchiveAndRestore(t *testing.T) {
	db := setupTestDB(t)
	wsID, projID, userID := seedTestData(t, db)
	svc := NewPageService(db)

	page, _ := svc.Create(&request.PageCreateRequest{Title: "Archive Me"}, projID, wsID, userID)

	// Archive
	err := svc.Archive(page.ID, projID)
	require.NoError(t, err)

	// Verify archived
	archived, _ := svc.Get(page.ID, projID)
	assert.NotNil(t, archived.ArchivedAt)

	// Verify not in tree
	tree, _ := svc.GetTree(projID)
	assert.Len(t, tree, 0, "Archived page should not appear in tree")

	// Restore
	err = svc.Restore(page.ID, projID)
	require.NoError(t, err)
	restored, _ := svc.Get(page.ID, projID)
	assert.Nil(t, restored.ArchivedAt)

	t.Logf("✅ Archive: page ID=%d archived and restored successfully", page.ID)
}

// ================ Locking Tests ================

func TestPageService_LockAndUnlock(t *testing.T) {
	db := setupTestDB(t)
	wsID, projID, userID := seedTestData(t, db)
	svc := NewPageService(db)

	page, _ := svc.Create(&request.PageCreateRequest{Title: "Lockable"}, projID, wsID, userID)

	// Lock
	locked, err := svc.Lock(page.ID, projID, userID)
	require.NoError(t, err)
	assert.Equal(t, &userID, locked.LockedByID)
	assert.NotNil(t, locked.LockedAt)

	// Unlock
	unlocked, err := svc.Unlock(page.ID, projID, userID)
	require.NoError(t, err)
	assert.Nil(t, unlocked.LockedByID)
	assert.Nil(t, unlocked.LockedAt)

	t.Logf("✅ Lock: page ID=%d locked and unlocked successfully", page.ID)
}

func TestPageService_LockedPagePreventUpdate(t *testing.T) {
	db := setupTestDB(t)
	wsID, projID, userID := seedTestData(t, db)
	svc := NewPageService(db)

	// Create second user
	otherUser := &model.User{DisplayName: "Other User", Email: "other@reqmango.com"}
	require.NoError(t, db.Create(otherUser).Error)

	page, _ := svc.Create(&request.PageCreateRequest{Title: "Locked Page"}, projID, wsID, userID)

	// User1 locks
	_, err := svc.Lock(page.ID, projID, userID)
	require.NoError(t, err)

	// User2 tries to update — should fail
	newTitle := "Hijacked"
	_, err = svc.Update(page.ID, projID, otherUser.ID, &request.PageUpdateRequest{Title: &newTitle})
	assert.Error(t, err, "Update by locked user should fail")

	t.Logf("✅ Lock Protection: page update blocked when locked by another user")
}

// ================ Version Tests ================

func TestPageVersionService_SaveAndRestore(t *testing.T) {
	db := setupTestDB(t)
	wsID, projID, userID := seedTestData(t, db)
	svc := NewPageService(db)
	vs := NewPageVersionService(db)

	page, _ := svc.Create(&request.PageCreateRequest{Title: "V1", Content: "original"}, projID, wsID, userID)

	// Update twice to create 2 versions
	c1 := "updated v1"
	svc.Update(page.ID, projID, userID, &request.PageUpdateRequest{Content: &c1})
	c2 := "updated v2"
	svc.Update(page.ID, projID, userID, &request.PageUpdateRequest{Content: &c2})

	// List versions
	versions, err := vs.List(page.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(versions), 2)
	t.Logf("Versions after 2 updates: %d", len(versions))

	// Get by version number
	v1, err := vs.GetByVersion(page.ID, versions[len(versions)-1].VersionNumber)
	require.NoError(t, err)
	assert.Equal(t, "original", v1.Content)
	t.Logf("✅ Version: v%d content=%s", v1.VersionNumber, v1.Content)
}

// ================ Template Tests ================

func TestPageTemplateService_CRUD(t *testing.T) {
	db := setupTestDB(t)
	wsID, projID, userID := seedTestData(t, db)
	ts := NewPageTemplateService(db)

	// Create
	req := &request.PageTemplateCreateRequest{
		Name:        "Meeting Notes",
		Description: "Standard meeting notes template",
		Content:     "# Meeting Notes\n\nDate: {{date}}\nAttendees: ",
		IsDefault:   true,
	}
	tmpl, err := ts.Create(req, wsID, userID)
	require.NoError(t, err)
	assert.True(t, tmpl.IsDefault)

	// List
	pID := projID
	templates, err := ts.List(wsID, &pID)
	require.NoError(t, err)
	assert.Len(t, templates, 1)

	// Get
	got, err := ts.Get(tmpl.ID)
	require.NoError(t, err)
	assert.Equal(t, "Meeting Notes", got.Name)

	// Update
	newName := "Weekly Standup Notes"
	_, err = ts.Update(tmpl.ID, userID, &request.PageTemplateUpdateRequest{Name: &newName})
	require.NoError(t, err)

	// Delete
	err = ts.Delete(tmpl.ID)
	require.NoError(t, err)
	_, err = ts.Get(tmpl.ID)
	assert.Error(t, err)

	t.Logf("✅ Template: created, updated, and deleted successfully")
}

// ================ Export Tests ================

func TestPageService_Export(t *testing.T) {
	db := setupTestDB(t)
	wsID, projID, userID := seedTestData(t, db)
	svc := NewPageService(db)

	page, _ := svc.Create(&request.PageCreateRequest{
		Title:   "Export Test",
		Content: "This is <b>test</b> content.",
	}, projID, wsID, userID)

	// Export as Markdown
	fn, content, err := svc.GetForExport(page.ID, projID, "md")
	require.NoError(t, err)
	assert.Contains(t, fn, ".md")
	assert.Contains(t, content, "# Export Test")
	assert.Contains(t, content, "test")
	t.Logf("✅ Export MD: %s (%d bytes)", fn, len(content))

	// Export as HTML
	fn2, content2, err := svc.GetForExport(page.ID, projID, "html")
	require.NoError(t, err)
	assert.Contains(t, fn2, ".html")
	assert.Contains(t, content2, "<!DOCTYPE html>")
	t.Logf("✅ Export HTML: %s (%d bytes)", fn2, len(content2))
}

// ================ Convert to Issue ================

func TestPageService_ConvertToIssue(t *testing.T) {
	db := setupTestDB(t)
	wsID, projID, userID := seedTestData(t, db)
	svc := NewPageService(db)

	page, _ := svc.Create(&request.PageCreateRequest{
		Title:   "Bug Report",
		Content: "<p>Something is broken</p>",
	}, projID, wsID, userID)

	issue, err := svc.ConvertToIssue(page.ID, projID, userID, nil)
	require.NoError(t, err)
	assert.Equal(t, "Bug Report", issue.Name)
	assert.Contains(t, issue.DescriptionHTML, "Something is broken")
	assert.Equal(t, projID, issue.ProjectID)

	t.Logf("✅ Convert: page → issue ID=%d, name=%s", issue.ID, issue.Name)
}

// ================ Benchmark ================

func BenchmarkPageService_GetTree(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	db.AutoMigrate(&model.User{}, &model.Workspace{}, &model.Project{}, &model.Page{})

	user := &model.User{DisplayName: "B", Email: "b@b.com"}
	db.Create(user)
	ws := &model.Workspace{Name: "BWS", Slug: "b-ws"}
	db.Create(ws)
	proj := &model.Project{Name: "BP", Identifier: "BP", WorkspaceID: ws.ID}
	db.Create(proj)

	svc := NewPageService(db)

	// Create 50 pages in 3-level hierarchy
	for i := 0; i < 10; i++ {
		root, _ := svc.Create(&request.PageCreateRequest{Title: fmt.Sprintf("Root %d", i)}, proj.ID, ws.ID, user.ID)
		for j := 0; j < 3; j++ {
			parentID := root.ID
			child, _ := svc.Create(&request.PageCreateRequest{Title: fmt.Sprintf("Child %d-%d", i, j), ParentID: &parentID}, proj.ID, ws.ID, user.ID)
			for k := 0; k < 2; k++ {
				gParentID := child.ID
				svc.Create(&request.PageCreateRequest{Title: fmt.Sprintf("Leaf %d-%d-%d", i, j, k), ParentID: &gParentID}, proj.ID, ws.ID, user.ID)
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GetTree(proj.ID)
	}
}
