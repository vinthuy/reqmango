package service

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupRuntimeTestDB(t *testing.T) *gorm.DB {
	db := openTestDB(t)
	err := db.AutoMigrate(
		&model.User{},
		&model.Workspace{},
		&model.Runtime{},
	)
	require.NoError(t, err)
	return db
}

func makeRuntime(t *testing.T, db *gorm.DB, wid uint64, name, rtType string, capacity int) *model.Runtime {
	id := nextID()
	rt := &model.Runtime{
		Name:        fmt.Sprintf("%s_%d", name, id),
		RuntimeType: rtType,
		RuntimeMode: "pull",
		Status:      "offline",
		Capacity:    capacity,
		CurrentLoad: 0,
		WorkspaceID: wid,
	}
	require.NoError(t, db.Create(rt).Error)
	return rt
}

// ==================== Create ====================

func TestRuntimeService_Create(t *testing.T) {
	db := setupRuntimeTestDB(t)
	svc := NewRuntimeService(db)

	ws := &model.Workspace{Name: "ws_rt1", Slug: "ws_rt1", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	endpoint := "ws://localhost:8080/ws"
	metadata := json.RawMessage(`{"os": "linux"}`)
	req := request.RuntimeCreate{
		Name:        "Local Daemon",
		RuntimeType: "local_daemon",
		RuntimeMode: "pull",
		Endpoint:    &endpoint,
		Capacity:    4,
		Metadata:    metadata,
	}

	resp, err := svc.Create(ws.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "Local Daemon", resp.Name)
	assert.Equal(t, "local_daemon", resp.RuntimeType)
	assert.Equal(t, "pull", resp.RuntimeMode)
	assert.Equal(t, "offline", resp.Status)
	assert.Equal(t, 4, resp.Capacity)
	assert.Equal(t, 0, resp.CurrentLoad)
	assert.Equal(t, ws.ID, resp.WorkspaceID)
}

func TestRuntimeService_Create_MinimalFields(t *testing.T) {
	db := setupRuntimeTestDB(t)
	svc := NewRuntimeService(db)

	ws := &model.Workspace{Name: "ws_rt_min", Slug: "ws_rt_min", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	req := request.RuntimeCreate{
		Name:        "Minimal Runtime",
		RuntimeType: "cloud",
	}

	resp, err := svc.Create(ws.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "Minimal Runtime", resp.Name)
	assert.Equal(t, 1, resp.Capacity) // default
	assert.Equal(t, 0, resp.CurrentLoad)
}

// ==================== Get ====================

func TestRuntimeService_Get(t *testing.T) {
	db := setupRuntimeTestDB(t)
	svc := NewRuntimeService(db)

	ws := &model.Workspace{Name: "ws_rt_get", Slug: "ws_rt_get", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	rt := makeRuntime(t, db, ws.ID, "get_rt", "local_daemon", 2)

	resp, err := svc.Get(rt.ID)
	require.NoError(t, err)
	assert.Equal(t, rt.ID, resp.ID)
	assert.Equal(t, rt.Name, resp.Name)
	assert.Equal(t, ws.ID, resp.WorkspaceID)
}

func TestRuntimeService_Get_NotFound(t *testing.T) {
	db := setupRuntimeTestDB(t)
	svc := NewRuntimeService(db)

	_, err := svc.Get(99999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== List ====================

func TestRuntimeService_List(t *testing.T) {
	db := setupRuntimeTestDB(t)
	svc := NewRuntimeService(db)

	ws := &model.Workspace{Name: "ws_rt_list", Slug: "ws_rt_list", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	makeRuntime(t, db, ws.ID, "rt_a", "local_daemon", 2)
	makeRuntime(t, db, ws.ID, "rt_b", "cloud", 4)

	resp, err := svc.List(ws.ID)
	require.NoError(t, err)
	assert.Len(t, resp, 2)
}

func TestRuntimeService_List_Empty(t *testing.T) {
	db := setupRuntimeTestDB(t)
	svc := NewRuntimeService(db)

	resp, err := svc.List(99999)
	require.NoError(t, err)
	assert.Len(t, resp, 0)
}

func TestRuntimeService_List_OtherWorkspace(t *testing.T) {
	db := setupRuntimeTestDB(t)
	svc := NewRuntimeService(db)

	ws1 := &model.Workspace{Name: "ws_rt_l1", Slug: "ws_rt_l1", OwnerID: 1}
	ws2 := &model.Workspace{Name: "ws_rt_l2", Slug: "ws_rt_l2", OwnerID: 1}
	require.NoError(t, db.Create(ws1).Error)
	require.NoError(t, db.Create(ws2).Error)

	makeRuntime(t, db, ws1.ID, "ws1_rt", "local_daemon", 1)
	makeRuntime(t, db, ws2.ID, "ws2_rt", "cloud", 2)

	resp1, err := svc.List(ws1.ID)
	require.NoError(t, err)
	assert.Len(t, resp1, 1)

	resp2, err := svc.List(ws2.ID)
	require.NoError(t, err)
	assert.Len(t, resp2, 1)
}

// ==================== Update ====================

func TestRuntimeService_Update(t *testing.T) {
	db := setupRuntimeTestDB(t)
	svc := NewRuntimeService(db)

	ws := &model.Workspace{Name: "ws_rt_upd", Slug: "ws_rt_upd", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	rt := makeRuntime(t, db, ws.ID, "upd_rt", "local_daemon", 1)

	newName := "Updated Runtime"
	newType := "cloud"
	newCapacity := 8
	meta := json.RawMessage(`{"gpu": true}`)
	req := request.RuntimeUpdate{
		Name:        &newName,
		RuntimeType: &newType,
		Capacity:    &newCapacity,
		Metadata:    &meta,
	}

	resp, err := svc.Update(rt.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "Updated Runtime", resp.Name)
	assert.Equal(t, "cloud", resp.RuntimeType)
	assert.Equal(t, 8, resp.Capacity)
}

func TestRuntimeService_Update_NotFound(t *testing.T) {
	db := setupRuntimeTestDB(t)
	svc := NewRuntimeService(db)

	newName := "Ghost"
	req := request.RuntimeUpdate{Name: &newName}

	_, err := svc.Update(99999, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRuntimeService_Update_PartialFields(t *testing.T) {
	db := setupRuntimeTestDB(t)
	svc := NewRuntimeService(db)

	ws := &model.Workspace{Name: "ws_rt_part", Slug: "ws_rt_part", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	rt := makeRuntime(t, db, ws.ID, "part_rt", "local_daemon", 2)

	newCapacity := 16
	req := request.RuntimeUpdate{
		Capacity: &newCapacity,
	}

	resp, err := svc.Update(rt.ID, req)
	require.NoError(t, err)
	assert.Equal(t, 16, resp.Capacity)
	// Name should remain unchanged
	assert.Equal(t, rt.Name, resp.Name)
}

// ==================== Delete ====================

func TestRuntimeService_Delete(t *testing.T) {
	db := setupRuntimeTestDB(t)
	svc := NewRuntimeService(db)

	ws := &model.Workspace{Name: "ws_rt_del", Slug: "ws_rt_del", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	rt := makeRuntime(t, db, ws.ID, "del_rt", "local_daemon", 1)

	err := svc.Delete(rt.ID)
	require.NoError(t, err)

	_, err = svc.Get(rt.ID)
	assert.Error(t, err)
}

func TestRuntimeService_Delete_NotFound(t *testing.T) {
	db := setupRuntimeTestDB(t)
	svc := NewRuntimeService(db)

	err := svc.Delete(99999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ==================== Heartbeat ====================

func TestRuntimeService_Heartbeat(t *testing.T) {
	db := setupRuntimeTestDB(t)
	svc := NewRuntimeService(db)

	ws := &model.Workspace{Name: "ws_rt_hb", Slug: "ws_rt_hb", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	rt := makeRuntime(t, db, ws.ID, "hb_rt", "local_daemon", 4)
	assert.Equal(t, "offline", rt.Status)

	hostInfo := json.RawMessage(`{"cpu": 4, "mem": 8192}`)
	req := request.RuntimeHeartbeat{
		Version:     "1.2.0",
		HostInfo:    hostInfo,
		CurrentLoad: 2,
	}

	resp, err := svc.Heartbeat(rt.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "online", resp.Status)
	assert.Equal(t, "1.2.0", resp.Version)
	assert.Equal(t, 2, resp.CurrentLoad)
	assert.NotNil(t, resp.LastHeartbeat)
}

func TestRuntimeService_Heartbeat_NotFound(t *testing.T) {
	db := setupRuntimeTestDB(t)
	svc := NewRuntimeService(db)

	req := request.RuntimeHeartbeat{
		Version: "1.0",
	}

	_, err := svc.Heartbeat(99999, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRuntimeService_Heartbeat_UpdatesStatusToOnline(t *testing.T) {
	db := setupRuntimeTestDB(t)
	svc := NewRuntimeService(db)

	ws := &model.Workspace{Name: "ws_rt_hb2", Slug: "ws_rt_hb2", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	rt := makeRuntime(t, db, ws.ID, "hb2_rt", "cloud", 2)
	assert.Equal(t, "offline", rt.Status)

	req := request.RuntimeHeartbeat{
		Version: "2.0",
	}

	resp, err := svc.Heartbeat(rt.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "online", resp.Status)

	// Verify persisted
	got, err := svc.Get(rt.ID)
	require.NoError(t, err)
	assert.Equal(t, "online", got.Status)
}

// ==================== FindAvailable ====================

func TestRuntimeService_FindAvailable(t *testing.T) {
	db := setupRuntimeTestDB(t)
	svc := NewRuntimeService(db)

	ws := &model.Workspace{Name: "ws_rt_avail", Slug: "ws_rt_avail", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	// Create an offline runtime
	makeRuntime(t, db, ws.ID, "offline_rt", "local_daemon", 2)

	// Create an online runtime with capacity
	onlineRt := makeRuntime(t, db, ws.ID, "online_rt", "cloud", 4)
	db.Model(onlineRt).Updates(map[string]interface{}{
		"status": "online",
	})

	resp, err := svc.FindAvailable(ws.ID)
	require.NoError(t, err)
	assert.Equal(t, onlineRt.ID, resp.ID)
}

func TestRuntimeService_FindAvailable_None(t *testing.T) {
	db := setupRuntimeTestDB(t)
	svc := NewRuntimeService(db)

	ws := &model.Workspace{Name: "ws_rt_none", Slug: "ws_rt_none", OwnerID: 1}
	require.NoError(t, db.Create(ws).Error)

	// Only offline runtime
	makeRuntime(t, db, ws.ID, "off_rt", "local_daemon", 2)

	_, err := svc.FindAvailable(ws.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "No available runtime")
}
