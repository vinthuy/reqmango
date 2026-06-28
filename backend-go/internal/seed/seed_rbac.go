package seed

import (
	"fmt"

	"github.com/reqmanpy/backend-go/internal/model"
	"gorm.io/gorm"
)

func SeedRBACData(db *gorm.DB) {
	var permCount int64
	db.Model(&model.Permission{}).Count(&permCount)
	if permCount > 0 {
		fmt.Println("RBAC data already exists, skipping seed")
		return
	}

	fmt.Println("Seeding RBAC permissions and default roles...")

	// Define all permissions
	perms := []model.Permission{
		// Workspace-level
		{Code: "workspace:view", Name: "View Workspace", Resource: "workspace", Action: "view", Scope: "workspace"},
		{Code: "workspace:manage", Name: "Manage Workspace", Resource: "workspace", Action: "manage", Scope: "workspace"},
		{Code: "workspace:delete", Name: "Delete Workspace", Resource: "workspace", Action: "delete", Scope: "workspace"},
		{Code: "member:manage", Name: "Manage Members", Resource: "member", Action: "manage", Scope: "workspace"},
		{Code: "member:view", Name: "View Members", Resource: "member", Action: "view", Scope: "workspace"},
		{Code: "project:create", Name: "Create Project", Resource: "project", Action: "create", Scope: "workspace"},
		{Code: "project:view_all", Name: "View All Projects", Resource: "project", Action: "view_all", Scope: "workspace"},
		{Code: "initiative:view", Name: "View Initiatives", Resource: "initiative", Action: "view", Scope: "workspace"},
		{Code: "initiative:manage", Name: "Manage Initiatives", Resource: "initiative", Action: "manage", Scope: "workspace"},
		{Code: "role:manage", Name: "Manage Roles", Resource: "role", Action: "manage", Scope: "workspace"},

		// Project-level
		{Code: "project:view", Name: "View Project", Resource: "project", Action: "view", Scope: "project"},
		{Code: "project:manage", Name: "Manage Project", Resource: "project", Action: "manage", Scope: "project"},
		{Code: "project:delete", Name: "Delete Project", Resource: "project", Action: "delete", Scope: "project"},
		{Code: "issue:view", Name: "View Issues", Resource: "issue", Action: "view", Scope: "project"},
		{Code: "issue:create", Name: "Create Issues", Resource: "issue", Action: "create", Scope: "project"},
		{Code: "issue:edit", Name: "Edit Issues", Resource: "issue", Action: "edit", Scope: "project"},
		{Code: "issue:delete", Name: "Delete Issues", Resource: "issue", Action: "delete", Scope: "project"},
		{Code: "issue:import", Name: "Import Issues", Resource: "issue", Action: "import", Scope: "project"},
		{Code: "issue:export", Name: "Export Issues", Resource: "issue", Action: "export", Scope: "project"},
		{Code: "cycle:view", Name: "View Cycles", Resource: "cycle", Action: "view", Scope: "project"},
		{Code: "cycle:create", Name: "Create Cycles", Resource: "cycle", Action: "create", Scope: "project"},
		{Code: "cycle:edit", Name: "Edit Cycles", Resource: "cycle", Action: "edit", Scope: "project"},
		{Code: "cycle:delete", Name: "Delete Cycles", Resource: "cycle", Action: "delete", Scope: "project"},
		{Code: "module:view", Name: "View Modules", Resource: "module", Action: "view", Scope: "project"},
		{Code: "module:create", Name: "Create Modules", Resource: "module", Action: "create", Scope: "project"},
		{Code: "module:edit", Name: "Edit Modules", Resource: "module", Action: "edit", Scope: "project"},
		{Code: "module:delete", Name: "Delete Modules", Resource: "module", Action: "delete", Scope: "project"},
		{Code: "page:view", Name: "View Pages", Resource: "page", Action: "view", Scope: "project"},
		{Code: "page:create", Name: "Create Pages", Resource: "page", Action: "create", Scope: "project"},
		{Code: "page:edit", Name: "Edit Pages", Resource: "page", Action: "edit", Scope: "project"},
		{Code: "page:delete", Name: "Delete Pages", Resource: "page", Action: "delete", Scope: "project"},
		{Code: "comment:view", Name: "View Comments", Resource: "comment", Action: "view", Scope: "project"},
		{Code: "comment:create", Name: "Create Comments", Resource: "comment", Action: "create", Scope: "project"},
		{Code: "comment:edit", Name: "Edit Comments", Resource: "comment", Action: "edit", Scope: "project"},
		{Code: "comment:delete", Name: "Delete Comments", Resource: "comment", Action: "delete", Scope: "project"},
		{Code: "attachment:view", Name: "View Attachments", Resource: "attachment", Action: "view", Scope: "project"},
		{Code: "attachment:create", Name: "Create Attachments", Resource: "attachment", Action: "create", Scope: "project"},
		{Code: "attachment:delete", Name: "Delete Attachments", Resource: "attachment", Action: "delete", Scope: "project"},
		{Code: "workflow:view", Name: "View Workflows", Resource: "workflow", Action: "view", Scope: "project"},
		{Code: "workflow:manage", Name: "Manage Workflows", Resource: "workflow", Action: "manage", Scope: "project"},
		{Code: "settings:view", Name: "View Settings", Resource: "settings", Action: "view", Scope: "project"},
		{Code: "settings:manage", Name: "Manage Settings", Resource: "settings", Action: "manage", Scope: "project"},
		{Code: "automation:view", Name: "View Automations", Resource: "automation", Action: "view", Scope: "project"},
		{Code: "automation:manage", Name: "Manage Automations", Resource: "automation", Action: "manage", Scope: "project"},
		{Code: "report:view", Name: "View Reports", Resource: "report", Action: "view", Scope: "project"},
		{Code: "report:create", Name: "Create Reports", Resource: "report", Action: "create", Scope: "project"},
		{Code: "ai:use", Name: "Use AI Features", Resource: "ai", Action: "use", Scope: "project"},
		{Code: "agent:manage", Name: "Manage AI Agents", Resource: "agent", Action: "manage", Scope: "project"},
		{Code: "time_track:view", Name: "View Time Tracking", Resource: "time_track", Action: "view", Scope: "project"},
		{Code: "time_track:create", Name: "Log Time", Resource: "time_track", Action: "create", Scope: "project"},
		{Code: "time_track:delete", Name: "Delete Time Logs", Resource: "time_track", Action: "delete", Scope: "project"},
		{Code: "release:view", Name: "View Releases", Resource: "release", Action: "view", Scope: "project"},
		{Code: "release:manage", Name: "Manage Releases", Resource: "release", Action: "manage", Scope: "project"},
		{Code: "member:manage_project", Name: "Manage Project Members", Resource: "member", Action: "manage_project", Scope: "project"},
		{Code: "member:view_project", Name: "View Project Members", Resource: "member", Action: "view_project", Scope: "project"},
	}

	// Create all permissions
	for i := range perms {
		db.Create(&perms[i])
	}

	// Map permission codes to IDs
	permMap := make(map[string]uint64)
	for _, p := range perms {
		permMap[p.Code] = p.ID
	}

	// Admin permissions (everything)
	adminPerms := allPermCodes()

	// Member permissions
	memberPerms := []string{
		"project:view_all",
		"project:view",
		"issue:view", "issue:create", "issue:edit", "issue:import", "issue:export",
		"cycle:view", "cycle:create", "cycle:edit",
		"module:view", "module:create", "module:edit",
		"page:view", "page:create", "page:edit",
		"comment:view", "comment:create", "comment:edit",
		"attachment:view", "attachment:create",
		"workflow:view",
		"settings:view",
		"automation:view",
		"report:view", "report:create",
		"ai:use",
		"time_track:view", "time_track:create",
		"release:view",
		"member:view_project",
		"member:view",
		"initiative:view",
	}

	// Guest permissions (read-only)
	guestPerms := []string{
		"project:view_all",
		"project:view",
		"issue:view",
		"cycle:view",
		"module:view",
		"page:view",
		"comment:view",
		"attachment:view",
		"workflow:view",
		"report:view",
		"time_track:view",
		"release:view",
		"member:view_project",
		"member:view",
		"initiative:view",
	}

	// Create default roles
	defaultRoles := []struct {
		role model.Role
		perm []string
	}{
		{model.Role{Name: "Admin", Description: "Full access to all resources", Scope: "workspace", Level: 20, IsSystem: true, SortOrder: 1}, adminPerms},
		{model.Role{Name: "Member", Description: "Create and edit content", Scope: "workspace", Level: 15, IsSystem: true, SortOrder: 2}, memberPerms},
		{model.Role{Name: "Guest", Description: "Read-only access", Scope: "workspace", Level: 5, IsSystem: true, SortOrder: 3}, guestPerms},
	}

	for _, dr := range defaultRoles {
		db.Create(&dr.role)
		// Assign permissions
		var permsToAdd []model.Permission
		for _, code := range dr.perm {
			if pid, ok := permMap[code]; ok {
				permsToAdd = append(permsToAdd, model.Permission{BaseModel: model.BaseModel{ID: pid}})
			}
		}
		db.Model(&dr.role).Association("Permissions").Replace(permsToAdd)
	}

	fmt.Println("RBAC seed complete: 3 default roles with granular permissions")
}

func allPermCodes() []string {
	return []string{
		"workspace:view", "workspace:manage", "workspace:delete",
		"member:manage", "member:view",
		"project:create", "project:view_all",
		"initiative:view", "initiative:manage",
		"role:manage",
		"project:view", "project:manage", "project:delete",
		"issue:view", "issue:create", "issue:edit", "issue:delete", "issue:import", "issue:export",
		"cycle:view", "cycle:create", "cycle:edit", "cycle:delete",
		"module:view", "module:create", "module:edit", "module:delete",
		"page:view", "page:create", "page:edit", "page:delete",
		"comment:view", "comment:create", "comment:edit", "comment:delete",
		"attachment:view", "attachment:create", "attachment:delete",
		"workflow:view", "workflow:manage",
		"settings:view", "settings:manage",
		"automation:view", "automation:manage",
		"report:view", "report:create",
		"ai:use", "agent:manage",
		"time_track:view", "time_track:create", "time_track:delete",
		"release:view", "release:manage",
		"member:manage_project", "member:view_project",
	}
}
