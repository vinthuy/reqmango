package seed

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/reqmanpy/backend/internal/common"
	"github.com/reqmanpy/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func strPtr(s string) *string { return &s }

func SeedAll(db *gorm.DB) {
	fmt.Println("=== Starting data initialization ===")
	
	SeedRBACData(db)
	SeedDemoData(db)
	SeedConfigData(db)
	
	fmt.Println("=== Data initialization complete ===")
}

func SeedDemoData(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Where("email LIKE 'demo%'").Count(&count)
	if count > 0 {
		fmt.Println("Demo data already exists, skipping seed")
		return
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	hash, _ := bcrypt.GenerateFromPassword([]byte("demo1234"), bcrypt.DefaultCost)

	users := make([]model.User, 0, 20)
	users = append(users, model.User{
		Email: "demo@example.com", Username: "demo", DisplayName: "Demo User",
		PasswordHash: string(hash), IsActive: true,
	})
	for i := 1; i <= 19; i++ {
		email := fmt.Sprintf("demo%d@reqman.local", i)
		username := fmt.Sprintf("demo%d", i)
		displayName := fmt.Sprintf("Demo User %d", i)
		users = append(users, model.User{
			Email: email, Username: username, DisplayName: displayName,
			PasswordHash: string(hash), IsActive: true,
		})
	}
	for i := range users {
		db.Create(&users[i])
	}
	fmt.Printf("Created %d demo users\n", len(users))

	ws := model.Workspace{
		Name: "Demo Workspace", Slug: "demo",
		Timezone: "Asia/Shanghai", OwnerID: users[0].ID,
	}
	if err := db.Create(&ws).Error; err != nil {
		fmt.Printf("Failed to create demo workspace: %v\n", err)
		return
	}
	for _, u := range users {
		role := common.RoleMember
		if u.ID == users[0].ID {
			role = common.RoleAdmin
		}
		db.Create(&model.WorkspaceMember{
			WorkspaceID: ws.ID, UserID: u.ID, Role: role, IsActive: true,
		})
	}
	fmt.Println("Created demo workspace with all users as members")

	desc := "A comprehensive demo project for testing all features"
	proj := model.Project{
		Name: "Demo Project", Identifier: "DEMO",
		Description: &desc, WorkspaceID: ws.ID, Color: "#6366F1",
	}
	if err := db.Create(&proj).Error; err != nil {
		fmt.Printf("Failed to create demo project: %v\n", err)
		return
	}
	fmt.Println("Created demo project")

	stateDefs := []struct {
		name  string
		color string
		group string
	}{
		{"待处理 (Backlog)", "#6B7280", "backlog"},
		{"待办 (Todo)", "#3B82F6", "unstarted"},
		{"进行中 (In Progress)", "#F59E0B", "started"},
		{"评审中 (In Review)", "#8B5CF6", "started"},
		{"已完成 (Done)", "#10B981", "completed"},
		{"已取消 (Cancelled)", "#EF4444", "cancelled"},
	}
	states := make([]model.State, len(stateDefs))
	for i, sd := range stateDefs {
		states[i] = model.State{
			Name:        sd.name,
			Color:       sd.color,
			Group:       sd.group,
			Sequence:    i + 1,
			IsDefault:   i == 0,
			ProjectID:   proj.ID,
			WorkspaceID: ws.ID,
		}
		db.Create(&states[i])
	}
	fmt.Printf("Created %d states\n", len(states))

	cycleNames := []string{"Sprint 1", "Sprint 2", "Sprint 3", "Sprint 4", "Sprint 5", "Sprint 6"}
	cycles := make([]model.Cycle, len(cycleNames))
	baseDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	for i, name := range cycleNames {
		start := baseDate.AddDate(0, 0, i*14)
		end := start.AddDate(0, 0, 13)
		var completedAt *time.Time
		if i < 3 {
			completedAt = &end
		}
		cycles[i] = model.Cycle{
			Name:        name,
			StartDate:   start,
			EndDate:     &end,
			CompletedAt: completedAt,
			ProjectID:   proj.ID,
			WorkspaceID: ws.ID,
		}
		db.Create(&cycles[i])
	}
	fmt.Printf("Created %d cycles (sprints)\n", len(cycles))

	moduleNames := []string{"用户模块", "订单模块", "支付模块", "报表模块", "系统设置", "消息中心", "权限管理", "数据统计"}
	modules := make([]model.Module, len(moduleNames))
	for i, name := range moduleNames {
		modules[i] = model.Module{
			Name:        name,
			ProjectID:   proj.ID,
			WorkspaceID: ws.ID,
		}
		db.Create(&modules[i])
	}
	fmt.Printf("Created %d modules\n", len(modules))

	priorities := []string{"urgent", "high", "medium", "low", "none"}
	issueTitles := []string{
		"用户登录页面优化", "修复密码重置功能", "添加数据导出功能", "优化首页加载速度",
		"实现权限管理模块", "修复搜索功能BUG", "添加用户反馈入口", "优化移动端适配",
		"实现通知推送功能", "添加操作日志记录", "优化数据库查询性能", "修复文件上传兼容性",
		"实现批量操作功能", "添加数据看板图表", "优化API响应速度", "修复邮件发送延迟",
		"实现SSO单点登录", "添加多语言支持", "优化缓存策略", "修复日期选择器BUG",
		"实现工作流引擎", "添加自定义字段功能", "优化代码结构重构", "修复列表分页问题",
		"实现消息队列集成", "添加数据备份功能", "优化前端打包体积", "修复权限校验漏洞",
		"实现WebSocket实时更新", "添加Markdown编辑器", "优化图片加载性能", "修复表单验证逻辑",
		"实现插件系统架构", "添加数据导入功能", "优化内存使用效率", "修复并发竞争问题",
		"实现定时任务调度", "添加API文档生成", "优化日志输出格式", "修复会话管理问题",
		"实现灰度发布功能", "添加性能监控面板", "优化数据库索引策略", "修复第三方登录回调",
		"实现全文搜索功能", "添加访问统计报表", "优化CSS加载速度", "修复数据一致性检查",
		"实现自动化测试框架", "添加团队协作功能", "优化数据库连接池", "修复安全漏洞扫描",
		"实现容器化部署方案", "添加操作撤回功能", "优化接口响应时间", "修复跨域请求问题",
	}

	for i := 0; i < 150; i++ {
		stateIdx := rng.Intn(len(states))
		cycleIdx := rng.Intn(len(cycles) + 1)
		priorityIdx := rng.Intn(len(priorities))
		moduleIdx := rng.Intn(len(modules))
		titleIdx := i % len(issueTitles)

		suffix := ""
		if i >= len(issueTitles) {
			suffix = fmt.Sprintf(" - Part %d", (i/len(issueTitles))+1)
		}

		seqNum := i + 1

		issue := model.Issue{
			Name:        fmt.Sprintf("%s%s", issueTitles[titleIdx], suffix),
			Priority:    priorities[priorityIdx],
			SequenceID:  seqNum,
			StateID:     states[stateIdx].ID,
			ProjectID:   proj.ID,
			WorkspaceID: ws.ID,
		}

		if err := db.Create(&issue).Error; err != nil {
			fmt.Printf("Failed to create issue %d: %v\n", i, err)
			continue
		}

		if cycleIdx < len(cycles) {
			db.Create(&model.IssueCycle{
				IssueID: issue.ID,
				CycleID: cycles[cycleIdx].ID,
			})
		}

		db.Create(&model.ModuleIssue{
			IssueID:  issue.ID,
			ModuleID: modules[moduleIdx].ID,
		})

		numAssignees := 1 + rng.Intn(3)
		rng.Shuffle(len(users), func(a, b int) { users[a], users[b] = users[b], users[a] })
		for j := 0; j < numAssignees && j < len(users); j++ {
			db.Create(&model.IssueAssignee{
				IssueID: issue.ID,
				UserID:  users[j].ID,
			})
		}
	}
	fmt.Println("Created 150 issues with random states, cycles, modules, and assignees")
}

func SeedConfigData(db *gorm.DB) {
	fmt.Println("--- Seeding configuration data ---")

	var ws model.Workspace
	if db.Where("slug = ?", "demo").First(&ws).Error != nil {
		fmt.Println("No demo workspace found, skipping config seed")
		return
	}
	var proj model.Project
	if db.Where("workspace_id = ?", ws.ID).First(&proj).Error != nil {
		fmt.Println("No project found for workspace, skipping config seed")
		return
	}

	var typeCount int64
	db.Model(&model.IssueType{}).Where("workspace_id = ?", ws.ID).Count(&typeCount)
	if typeCount == 0 {
		epic := model.IssueType{Name: "Epic", Color: "#8B5CF6", Icon: "layers", Description: "顶层史诗级工作项", Level: 0, IsDefault: true, Sequence: 1, WorkspaceID: ws.ID}
		db.Create(&epic)
		feature := model.IssueType{Name: "Feature", Color: "#6366F1", Icon: "star", Description: "功能特性", Level: 1, ParentTypeID: &epic.ID, Sequence: 2, WorkspaceID: ws.ID}
		db.Create(&feature)
		bug := model.IssueType{Name: "Bug", Color: "#EF4444", Icon: "bug", Description: "缺陷/问题", Level: 2, ParentTypeID: &feature.ID, Sequence: 3, WorkspaceID: ws.ID}
		db.Create(&bug)
		task := model.IssueType{Name: "Task", Color: "#10B981", Icon: "check-circle", Description: "开发任务", Level: 2, ParentTypeID: &feature.ID, Sequence: 4, WorkspaceID: ws.ID}
		db.Create(&task)
		story := model.IssueType{Name: "Story", Color: "#F59E0B", Icon: "bookmark", Description: "用户故事", Level: 1, ParentTypeID: &epic.ID, Sequence: 5, WorkspaceID: ws.ID}
		db.Create(&story)
		spike := model.IssueType{Name: "Spike", Color: "#06B6D4", Icon: "zap", Description: "技术调研/探索", Level: 1, ParentTypeID: &epic.ID, Sequence: 6, WorkspaceID: ws.ID}
		db.Create(&spike)
		fmt.Printf("  Created 6 issue types with hierarchy\n")
	}

	var cfCount int64
	db.Model(&model.CustomField{}).Where("workspace_id = ?", ws.ID).Count(&cfCount)
	if cfCount == 0 {
		priority := model.CustomField{Name: "优先级", FieldType: "dropdown", IsRequired: true, WorkspaceID: ws.ID}
		db.Create(&priority)
		db.Create(&model.CustomFieldOption{FieldID: priority.ID, Value: "P0-紧急", Color: "#EF4444", Sequence: 1})
		db.Create(&model.CustomFieldOption{FieldID: priority.ID, Value: "P1-高", Color: "#F97316", Sequence: 2})
		db.Create(&model.CustomFieldOption{FieldID: priority.ID, Value: "P2-中", Color: "#F59E0B", Sequence: 3})
		db.Create(&model.CustomFieldOption{FieldID: priority.ID, Value: "P3-低", Color: "#10B981", Sequence: 4})

		deadline := model.CustomField{Name: "截止日期", FieldType: "date", WorkspaceID: ws.ID}
		db.Create(&deadline)

		version := model.CustomField{Name: "版本号", FieldType: "text", WorkspaceID: ws.ID}
		db.Create(&version)

		storyPoints := model.CustomField{Name: "故事点", FieldType: "number", WorkspaceID: ws.ID}
		db.Create(&storyPoints)

		reporter := model.CustomField{Name: "报告人", FieldType: "member", WorkspaceID: ws.ID}
		db.Create(&reporter)

		refURL := model.CustomField{Name: "参考链接", FieldType: "url", WorkspaceID: ws.ID}
		db.Create(&refURL)

		fmt.Printf("  Created 6 custom fields (dropdown/date/text/number/member/url)\n")
	}

	var bindCount int64
	db.Model(&model.IssueTypeField{}).Count(&bindCount)
	if bindCount == 0 {
		var types []model.IssueType
		db.Where("workspace_id = ?", ws.ID).Find(&types)
		var fields []model.CustomField
		db.Where("workspace_id = ?", ws.ID).Find(&fields)
		for _, t := range types {
			for _, f := range fields {
				isReq := f.Name == "优先级"
				db.Create(&model.IssueTypeField{TypeID: t.ID, FieldID: f.ID, IsRequired: isReq, Sequence: 1})
			}
		}
		fmt.Printf("  Bound %d fields x %d types\n", len(fields), len(types))
	}

	var lblCount int64
	db.Model(&model.Label{}).Where("project_id = ?", proj.ID).Count(&lblCount)
	if lblCount == 0 {
		labelDefs := []struct{ name, color string }{
			{"前端", "#3B82F6"}, {"后端", "#22C55E"}, {"文档", "#F59E0B"},
			{"DevOps", "#8B5CF6"}, {"安全", "#EF4444"}, {"性能优化", "#F97316"},
			{"Bug", "#DC2626"}, {"增强", "#06B6D4"}, {"测试", "#84CC16"},
			{"设计", "#A855F7"}, {"需求", "#EC4899"}, {"架构", "#0EA5E9"},
		}
		for _, l := range labelDefs {
			db.Create(&model.Label{Name: l.name, Color: l.color, ProjectID: proj.ID})
		}
		fmt.Printf("  Created %d labels\n", len(labelDefs))
	}

	var wfCount int64
	db.Model(&model.Workflow{}).Where("project_id = ?", proj.ID).Count(&wfCount)
	if wfCount == 0 {
		var stList []model.State
		db.Where("project_id = ? AND is_active = true", proj.ID).Order("sequence").Find(&stList)
		if len(stList) >= 5 {
			bid, tid, ipid, rid, dnid := stList[0].ID, stList[1].ID, stList[2].ID, stList[3].ID, stList[4].ID
			var cid uint64
			if len(stList) >= 6 {
				cid = stList[5].ID
			}
			wf := model.Workflow{Name: "Default Workflow", Description: "标准状态流转规则", ProjectID: proj.ID, IsActive: true}
			db.Create(&wf)
			trs := []model.StateTransition{
				{Name: "Backlog→Todo", WorkflowID: wf.ID, SourceStateID: bid, TargetStateID: tid, RuleType: "allow", ProjectID: proj.ID, WorkspaceID: ws.ID},
				{Name: "Todo→InProgress", WorkflowID: wf.ID, SourceStateID: tid, TargetStateID: ipid, RuleType: "allow", ProjectID: proj.ID, WorkspaceID: ws.ID},
				{Name: "InProgress→Review", WorkflowID: wf.ID, SourceStateID: ipid, TargetStateID: rid, RuleType: "allow", ProjectID: proj.ID, WorkspaceID: ws.ID},
				{Name: "Review→Done", WorkflowID: wf.ID, SourceStateID: rid, TargetStateID: dnid, RuleType: "approval", ProjectID: proj.ID, WorkspaceID: ws.ID},
				{Name: "Review→InProgress", WorkflowID: wf.ID, SourceStateID: rid, TargetStateID: ipid, RuleType: "allow", ProjectID: proj.ID, WorkspaceID: ws.ID},
				{Name: "InProgress→Todo", WorkflowID: wf.ID, SourceStateID: ipid, TargetStateID: tid, RuleType: "allow", ProjectID: proj.ID, WorkspaceID: ws.ID},
				{Name: "Todo→Backlog", WorkflowID: wf.ID, SourceStateID: tid, TargetStateID: bid, RuleType: "allow", ProjectID: proj.ID, WorkspaceID: ws.ID},
			}
			if len(stList) >= 6 {
				trs = append(trs,
					model.StateTransition{Name: "Backlog→Cancelled", WorkflowID: wf.ID, SourceStateID: bid, TargetStateID: cid, RuleType: "allow", ProjectID: proj.ID, WorkspaceID: ws.ID},
					model.StateTransition{Name: "Todo→Cancelled", WorkflowID: wf.ID, SourceStateID: tid, TargetStateID: cid, RuleType: "allow", ProjectID: proj.ID, WorkspaceID: ws.ID},
					model.StateTransition{Name: "InProgress→Cancelled", WorkflowID: wf.ID, SourceStateID: ipid, TargetStateID: cid, RuleType: "allow", ProjectID: proj.ID, WorkspaceID: ws.ID},
					model.StateTransition{Name: "Review→Cancelled", WorkflowID: wf.ID, SourceStateID: rid, TargetStateID: cid, RuleType: "allow", ProjectID: proj.ID, WorkspaceID: ws.ID},
					model.StateTransition{Name: "Done→Review", WorkflowID: wf.ID, SourceStateID: dnid, TargetStateID: rid, RuleType: "allow", ProjectID: proj.ID, WorkspaceID: ws.ID},
				)
			}
			for _, tr := range trs {
				db.Create(&tr)
			}
			fmt.Printf("  Created workflow with %d transitions\n", len(trs))
		} else {
			fmt.Printf("  Skipping workflow creation: only %d states found\n", len(stList))
		}
	}

	var autoCount int64
	db.Model(&model.AutomationRule{}).Where("project_id = ?", proj.ID).Count(&autoCount)
	if autoCount == 0 {
		db.Create(&model.AutomationRule{
			Name: "紧急任务自动分配", Description: "urgent优先级的任务创建时自动分配给管理员",
			ProjectID: proj.ID, TriggerType: "issue_created", IsEnabled: true, Sequence: 1,
			Conditions: `[{"field":"priority","operator":"equals","value":"urgent"}]`,
			Actions:    `[{"type":"assign","field":"assignee","value":"1"}]`,
		})
		db.Create(&model.AutomationRule{
			Name: "完成任务时记录时间", Description: "状态变为Done时自动记录完成时间",
			ProjectID: proj.ID, TriggerType: "state_changed", IsEnabled: true, Sequence: 2,
			Conditions: `[{"field":"state_group","operator":"equals","value":"completed"}]`,
			Actions:    `[{"type":"set_timestamp","field":"completed_at"}]`,
		})
		db.Create(&model.AutomationRule{
			Name: "自动设置截止日期", Description: "新建任务时自动设置截止日期为3天后",
			ProjectID: proj.ID, TriggerType: "issue_created", IsEnabled: true, Sequence: 3,
			Conditions: `[]`,
			Actions:    `[{"type":"set_date_offset","field":"target_date","offset_days":3}]`,
		})
		fmt.Println("  Created 3 automation rules")
	}

	var relCount int64
	db.Model(&model.RelationType{}).Where("workspace_id = ?", ws.ID).Count(&relCount)
	if relCount == 0 {
		relTypes := []model.RelationType{
			{Name: "阻塞", InwardName: "被阻塞于", OutwardName: "阻塞", WorkspaceID: ws.ID},
			{Name: "关联", InwardName: "关联到", OutwardName: "被关联于", WorkspaceID: ws.ID},
			{Name: "重复", InwardName: "重复于", OutwardName: "被重复于", WorkspaceID: ws.ID},
			{Name: "父子", InwardName: "子任务", OutwardName: "父任务", WorkspaceID: ws.ID},
			{Name: "依赖", InwardName: "依赖于", OutwardName: "被依赖于", WorkspaceID: ws.ID},
		}
		for _, r := range relTypes {
			db.Create(&r)
		}
		fmt.Printf("  Created %d relation types\n", len(relTypes))
	}

	var irCount int64
	db.Model(&model.IssueRelation{}).Count(&irCount)
	if irCount == 0 {
		var relBlocks, relRelated model.RelationType
		db.Where("workspace_id = ? AND name = ?", ws.ID, "阻塞").First(&relBlocks)
		db.Where("workspace_id = ? AND name = ?", ws.ID, "关联").First(&relRelated)
		var issues []model.Issue
		db.Where("project_id = ?", proj.ID).Limit(15).Find(&issues)
		if len(issues) >= 6 {
			if relBlocks.ID > 0 {
				db.Create(&model.IssueRelation{IssueID: issues[0].ID, RelatedIssueID: issues[1].ID, RelationTypeID: relBlocks.ID})
				db.Create(&model.IssueRelation{IssueID: issues[2].ID, RelatedIssueID: issues[3].ID, RelationTypeID: relBlocks.ID})
			}
			if relRelated.ID > 0 {
				db.Create(&model.IssueRelation{IssueID: issues[1].ID, RelatedIssueID: issues[4].ID, RelationTypeID: relRelated.ID})
				db.Create(&model.IssueRelation{IssueID: issues[4].ID, RelatedIssueID: issues[5].ID, RelationTypeID: relRelated.ID})
				db.Create(&model.IssueRelation{IssueID: issues[3].ID, RelatedIssueID: issues[0].ID, RelationTypeID: relRelated.ID})
			}
			fmt.Println("  Created sample issue relations")
		}
	}

	var commentCount int64
	db.Model(&model.Comment{}).Count(&commentCount)
	if commentCount == 0 {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		var issues []model.Issue
		db.Where("project_id = ?", proj.ID).Limit(20).Find(&issues)
		for _, issue := range issues {
			if rng.Intn(100) < 40 {
				authorID := uint64(1)
				db.Create(&model.Comment{
					IssueID:  issue.ID,
					AuthorID: &authorID,
					Body:     "这是一条测试评论，用于演示评论功能",
				})
			}
		}
		fmt.Println("  Created sample comments")
	}

	var notifCount int64
	db.Model(&model.Notification{}).Count(&notifCount)
	if notifCount == 0 {
		adminID := uint64(1)
		db.Create(&model.Notification{
			Title:       "Welcome to ReqManPy!",
			Message:     "Your project management platform is ready. Start by creating your first work item.",
			Type:        "info",
			Priority:    "medium",
			RecipientID: adminID,
		})
		db.Create(&model.Notification{
			Title:       "Project Created",
			Message:     "Demo project has been set up with sample data including issues, cycles, and modules.",
			Type:        "success",
			Priority:    "high",
			RecipientID: adminID,
			ProjectID:   &proj.ID,
		})
		fmt.Println("  Created sample notifications")
	}

	var agentCount int64
	db.Model(&model.Agent{}).Where("workspace_id = ?", ws.ID).Count(&agentCount)
	if agentCount == 0 {
		adminID := uint64(1)
		// Built-in Triage Agent
		db.Create(&model.Agent{
			Name:        "Triage Agent",
			Avatar:      "🏥",
			AgentType:   "builtin",
			Capabilities: []byte(`["analyze","search","comment","list"]`),
			Status:      "active",
			SystemPrompt: strPtr("You are a triage specialist. Analyze incoming issues and suggest the correct type, priority, labels, and assignee. Be concise and data-driven."),
			WorkspaceID: ws.ID,
			BaseModel:   model.BaseModel{CreatedByID: &adminID},
		})
		// Built-in Summary Agent
		db.Create(&model.Agent{
			Name:        "Summary Agent",
			Avatar:      "📋",
			AgentType:   "builtin",
			Capabilities: []byte(`["summarize","analyze","list"]`),
			Status:      "active",
			SystemPrompt: strPtr("You are a project analyst. Summarize sprint progress, project health, and team performance. Provide actionable insights."),
			WorkspaceID: ws.ID,
			BaseModel:   model.BaseModel{CreatedByID: &adminID},
		})
		// Built-in Assistant Agent
		db.Create(&model.Agent{
			Name:        "Assistant Agent",
			Avatar:      "🤖",
			AgentType:   "builtin",
			Capabilities: []byte(`["all"]`),
			Status:      "active",
			SystemPrompt: strPtr("You are a helpful project management assistant. Help users with any task — from searching issues to creating work items and analyzing data."),
			WorkspaceID: ws.ID,
			BaseModel:   model.BaseModel{CreatedByID: &adminID},
		})
		fmt.Println("  Created 3 built-in AI agents (Triage, Summary, Assistant)")
	}

	fmt.Println("--- Config seed complete ---")
}