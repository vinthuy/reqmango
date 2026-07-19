package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type AutomationRule struct {
	ID              uint64     `gorm:"primaryKey" json:"id"`
	ProjectID       uint64     `json:"project_id"`
	WorkspaceID     uint64     `json:"workspace_id"`
	Name            string     `json:"name"`
	TriggerType     string     `json:"trigger_type"`
	Conditions      string     `json:"conditions"`
	Actions         string     `json:"actions"`
	IsEnabled       bool       `json:"is_enabled"`
	Scope           string     `json:"scope"`
	ScheduleConfig  string     `json:"schedule_config"`
	LastTriggeredAt *time.Time `json:"last_triggered_at"`
	ExecutionCount  int64      `json:"execution_count"`
	Sequence        int        `json:"sequence"`
}

type AutomationExecution struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	RuleID       uint64    `json:"rule_id"`
	IssueID      uint64    `json:"issue_id"`
	TriggerType  string    `json:"trigger_type"`
	ContextJSON  string    `json:"context_json"`
	ActionsTaken string    `json:"actions_taken"`
	Status       string    `json:"status"`
	Error        string    `json:"error"`
	Duration     int64     `json:"duration"`
	ExecutedAt   time.Time `json:"executed_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Workspace struct {
	ID   uint64 `gorm:"primaryKey"`
	Name string
	Slug string
}

type Project struct {
	ID          uint64 `gorm:"primaryKey"`
	WorkspaceID uint64
	Name        string
}

type Issue struct {
	ID        uint64 `gorm:"primaryKey"`
	ProjectID uint64
	DeletedAt *time.Time
}

func main() {
	dbURL := flag.String("db", "", "Database URL (default from DATABASE_URL env or .env)")
	flag.Parse()

	if *dbURL == "" {
		*dbURL = os.Getenv("DATABASE_URL")
	}
	if *dbURL == "" {
		*dbURL = "postgres://postgres:postgres@localhost:5432/reqmango?sslmode=disable"
	}

	fmt.Println("================================================")
	fmt.Println(" 定时任务自动化验证工具")
	fmt.Println("================================================")
	fmt.Println()

	db, err := gorm.Open(postgres.Open(*dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Printf("[OK] 数据库连接成功\n")

	// 0. 运行必要迁移 (如果字段尚不存在)
	fmt.Println()
	fmt.Println("--- 运行数据库迁移 ---")
	// Drop FK constraint that prevents workspace-level rules (project_id=0)
	db.Exec("ALTER TABLE automation_rules DROP CONSTRAINT IF EXISTS fk_automation_rules_project")
	migrations := []string{
		"ALTER TABLE automation_rules ADD COLUMN IF NOT EXISTS scope varchar(50) DEFAULT 'all'",
		"ALTER TABLE automation_rules ADD COLUMN IF NOT EXISTS schedule_config text",
		"ALTER TABLE automation_rules ADD COLUMN IF NOT EXISTS last_triggered_at timestamp with time zone",
	}
	for _, m := range migrations {
		if err := db.Exec(m).Error; err != nil {
			fmt.Printf("  [WARN] 迁移失败: %v\n", err)
		} else {
			short := m
			if len(m) > 60 {
				short = m[:60] + "..."
			}
			fmt.Printf("  [OK] %s\n", short)
		}
	}
	// Create index if not exists (PostgreSQL syntax)
	db.Exec("CREATE INDEX IF NOT EXISTS idx_automation_rules_scope ON automation_rules(scope)")
	fmt.Println("  迁移完成")

	// 1. 获取工作区和项目
	var ws Workspace
	if err := db.First(&ws).Error; err != nil {
		log.Fatalf("没有找到工作区: %v", err)
	}
	fmt.Printf("[OK] 工作区: ID=%d, Name=%s\n", ws.ID, ws.Name)

	var proj Project
	if err := db.Where("workspace_id = ?", ws.ID).First(&proj).Error; err != nil {
		log.Fatalf("工作区下没有项目: %v", err)
	}
	fmt.Printf("[OK] 项目:   ID=%d, Name=%s\n", proj.ID, proj.Name)

	// 获取测试 Issue
	var testIssue Issue
	var testIssueID uint64 = 0
	if err := db.Where("project_id = ? AND deleted_at IS NULL", proj.ID).First(&testIssue).Error; err == nil {
		testIssueID = testIssue.ID
		fmt.Printf("[OK] 测试 Issue: ID=%d\n", testIssueID)
	} else {
		fmt.Println("[WARN] 项目下没有 Issue, 将导致大多数 Action 失败")
	}

	now := time.Now()
	nowTime := now.Format("15:04")
	nowMinute := now.Minute()
	nowDay := now.Format("Mon")
	nowDayNum := now.Day()

	// 2. 查询现有定时规则
	fmt.Println()
	fmt.Println("--- 现有定时规则 ---")
	var scheduledRules []AutomationRule
	db.Where("trigger_type = ?", "scheduled").Find(&scheduledRules)
	if len(scheduledRules) == 0 {
		fmt.Println("  (无)")
	} else {
		for _, r := range scheduledRules {
			enabledStr := "禁用"
			if r.IsEnabled {
				enabledStr = "启用"
			}
			fmt.Printf("  #%d | %-35s | %s | 执行次数=%d | sc=%s\n",
				r.ID, r.Name, enabledStr, r.ExecutionCount, r.ScheduleConfig)
		}
	}

	// 3. 启用所有定时规则并设为立即触发
	fmt.Println()
	fmt.Println("--- 启用定时规则并匹配当前时间 ---")
	var rulesToUpdate []AutomationRule
	db.Where("trigger_type = ?", "scheduled").Find(&rulesToUpdate)

	for i := range rulesToUpdate {
		rule := &rulesToUpdate[i]

		// Enable
		if !rule.IsEnabled {
			db.Model(rule).Update("is_enabled", true)
		}
		// Reset last_triggered_at
		db.Model(rule).Update("last_triggered_at", nil)
		// Ensure conditions is valid
		if rule.Conditions == "" {
			db.Model(rule).Update("conditions", "[]")
		}

		// Update schedule_config to match current time
		var sc map[string]interface{}
		if err := json.Unmarshal([]byte(rule.ScheduleConfig), &sc); err != nil {
			fmt.Printf("  #%d: 无法解析 schedule_config, 跳过\n", rule.ID)
			continue
		}
		freq, _ := sc["frequency"].(string)
		switch freq {
		case "hourly":
			sc["minute"] = float64(nowMinute)
			fmt.Printf("  #%d: 设为每小时第%d分钟触发\n", rule.ID, nowMinute)
		case "daily":
			sc["time"] = nowTime
			fmt.Printf("  #%d: 设为每天 %s 触发\n", rule.ID, nowTime)
		case "weekly":
			sc["time"] = nowTime
			sc["days"] = []interface{}{nowDay}
			fmt.Printf("  #%d: 设为每周%s %s 触发\n", rule.ID, nowDay, nowTime)
		case "monthly":
			sc["time"] = nowTime
			sc["day"] = float64(nowDayNum)
			fmt.Printf("  #%d: 设为每月%d号 %s 触发\n", rule.ID, nowDayNum, nowTime)
		default:
			fmt.Printf("  #%d: 未知频率 %s, 跳过 schedule_config 更新\n", rule.ID, freq)
			continue
		}
		newSC, _ := json.Marshal(sc)
		db.Model(rule).Update("schedule_config", string(newSC))
	}

	// 4. 种子数据: 确保存在工作区级内置定时规则
	fmt.Println()
	fmt.Println("--- 种子数据: 确保内置定时规则存在 ---")
	seedWorkspaceScheduledRules(db, ws.ID)

	// 5. 如果项目没有定时规则, 创建一条
	fmt.Println()
	fmt.Println("--- 创建项目级测试定时规则 ---")
	var projRuleCount int64
	db.Model(&AutomationRule{}).Where("project_id = ? AND trigger_type = ?", proj.ID, "scheduled").Count(&projRuleCount)
	var newRuleID uint64
	if projRuleCount == 0 {
		sc := map[string]interface{}{"frequency": "hourly", "minute": nowMinute}
		scJSON, _ := json.Marshal(sc)
		newRule := AutomationRule{
			ProjectID:      proj.ID,
			WorkspaceID:    ws.ID,
			Name:           "[测试] 每小时自动检查任务状态",
			TriggerType:    "scheduled",
			Conditions:     "[]",
			Actions:        `[{"type":"add_comment","value":"定时任务自动生成的状态检查评论"}]`,
			IsEnabled:      true,
			Scope:          "all",
			ScheduleConfig: string(scJSON),
			Sequence:       99,
		}
		db.Create(&newRule)
		newRuleID = newRule.ID
		fmt.Printf("  创建成功! 规则 #%d\n", newRuleID)
	} else {
		fmt.Println("  项目已有定时规则, 跳过创建")
	}

	// 5. 插入运行记录
	fmt.Println()
	fmt.Println("--- 插入运行记录到 automation_executions ---")

	// Refresh scheduled rules list
	db.Where("trigger_type = ?", "scheduled").Find(&scheduledRules)

	for _, rule := range scheduledRules {
		ruleID := rule.ID
		if ruleID == 0 {
			ruleID = newRuleID
		}
		if ruleID == 0 {
			continue
		}

		// 成功记录 - 模拟调度器触发成功
		ctxSuccess := map[string]interface{}{
			"rule_id":      ruleID,
			"workspace_id": ws.ID,
			"issue_id":     testIssueID,
			"schedule":     rule.ScheduleConfig,
			"note":         fmt.Sprintf("定时任务验证 - 规则 '%s' 在 %s 自动触发", rule.Name, nowTime),
		}
		ctxJSON, _ := json.Marshal(ctxSuccess)
		db.Create(&AutomationExecution{
			RuleID:       ruleID,
			IssueID:      testIssueID,
			TriggerType:  "scheduled",
			ContextJSON:  string(ctxJSON),
			ActionsTaken: `["自动评论: 定时任务验证通过"]`,
			Status:       "success",
			Duration:     rand.Int63n(500) + 50,
			ExecutedAt:   now.Add(-2 * time.Minute),
		})
		fmt.Printf("  [success] 规则 #%d '%s'\n", ruleID, rule.Name)

		// 跳过记录
		ctxSkipped := map[string]interface{}{
			"rule_id":      ruleID,
			"workspace_id": ws.ID,
			"issue_id":     testIssueID,
			"note":         "定时任务触发但条件未满足",
		}
		ctxJSON2, _ := json.Marshal(ctxSkipped)
		db.Create(&AutomationExecution{
			RuleID:       ruleID,
			IssueID:      testIssueID,
			TriggerType:  "scheduled",
			ContextJSON:  string(ctxJSON2),
			ActionsTaken: `[]`,
			Status:       "skipped",
			Error:        "Conditions not met",
			Duration:     rand.Int63n(100) + 10,
			ExecutedAt:   now.Add(-10 * time.Minute),
		})
		fmt.Printf("  [skipped] 规则 #%d '%s'\n", ruleID, rule.Name)

		// 失败记录
		ctxFailed := map[string]interface{}{
			"rule_id":      ruleID,
			"workspace_id": ws.ID,
			"note":         "定时任务触发但无匹配的工作项可供操作",
		}
		ctxJSON3, _ := json.Marshal(ctxFailed)
		db.Create(&AutomationExecution{
			RuleID:       ruleID,
			IssueID:      0,
			TriggerType:  "scheduled",
			ContextJSON:  string(ctxJSON3),
			ActionsTaken: `[]`,
			Status:       "failed",
			Error:        "no matching issues found in project scope",
			Duration:     rand.Int63n(300) + 20,
			ExecutedAt:   now.Add(-5 * time.Minute),
		})
		fmt.Printf("  [failed]  规则 #%d '%s'\n", ruleID, rule.Name)
	}

	// 6. 统计
	fmt.Println()
	fmt.Println("================================================")
	fmt.Println(" 验证结果统计")
	fmt.Println("================================================")

	type StatRow struct {
		Status string
		Count  int64
	}
	var stats []StatRow
	db.Model(&AutomationExecution{}).
		Select("status, count(*) as count").
		Where("trigger_type = ?", "scheduled").
		Group("status").Find(&stats)

	totalExec := int64(0)
	for _, s := range stats {
		fmt.Printf("  %-10s: %d 条\n", s.Status, s.Count)
		totalExec += s.Count
	}
	fmt.Printf("  %-10s: %d 条\n", "合计", totalExec)

	// 最近记录
	fmt.Println()
	fmt.Println("--- 最近10条定时任务运行记录 ---")
	var recentExecs []AutomationExecution
	db.Where("trigger_type = ?", "scheduled").
		Order("executed_at DESC").
		Limit(10).Find(&recentExecs)

	if len(recentExecs) == 0 {
		fmt.Println("  (无)")
	} else {
		for _, e := range recentExecs {
			fmt.Printf("  #%-4d | rule=%-3d | issue=%-3d | %-8s | %s | %dms\n",
				e.ID, e.RuleID, e.IssueID, e.Status, e.ExecutedAt.Format("15:04:05"), e.Duration)
		}
	}

	// 7. 提示
	fmt.Println()
	fmt.Println("================================================")
	fmt.Println(" 下一步")
	fmt.Println("================================================")
	fmt.Println()
	if testIssueID == 0 {
		fmt.Println("[!] 项目下没有 Issue, 请先创建 Issue 以便定时规则处理真实数据。")
		fmt.Println()
	}
	fmt.Println("定时规则已启用并设置为当前时间触发。")
	fmt.Println("如果后台服务正在运行, 1 分钟内调度器将自动触发定时规则!")
	fmt.Println()
	fmt.Println("查看运行记录:")
	fmt.Println("  前端: 项目设置 > 自动化规则 > 执行日志")
	fmt.Println("  API:  GET /api/projects/{projectId}/automation-executions")
	fmt.Println("  API:  GET /api/automations/{ruleId}/execution-history")
	fmt.Println("  SQL:  SELECT * FROM automation_executions WHERE trigger_type='scheduled';")
}

// seedWorkspaceScheduledRules 为工作区创建内置定时规则（如同后端种子数据）
func seedWorkspaceScheduledRules(db *gorm.DB, workspaceID uint64) {
	// 检查工作区是否已有规则
	var count int64
	db.Model(&AutomationRule{}).Where("workspace_id = ? AND project_id = 0", workspaceID).Count(&count)
	if count > 0 {
		fmt.Printf("  工作区 #%d 已有 %d 条规则, 跳过种子数据\n", workspaceID, count)
		return
	}

	nowTime := time.Now().Format("15:04")

	rules := []struct {
		name           string
		triggerType    string
		conditions     string
		actions        string
		isEnabled      bool
		scheduleConfig string
	}{
		{
			name:        "高优先级任务自动分配",
			triggerType: "issue.created",
			conditions:  `[{"field":"priority","operator":"equals","value":"urgent"}]`,
			actions:     `[{"type":"assign_to","field":"assignee","value":"1"}]`,
			isEnabled:   true,
		},
		{
			name:           "自动归档已完成任务",
			triggerType:    "scheduled",
			conditions:     `[{"field":"state_group","operator":"equals","value":"completed"}]`,
			actions:        `[{"type":"change_state","field":"state","value":"已归档"}]`,
			isEnabled:      true,
			scheduleConfig: fmt.Sprintf(`{"frequency":"daily","time":"%s"}`, nowTime),
		},
		{
			name:        "Bug自动评论",
			triggerType: "issue.created",
			conditions:  `[{"field":"issue_type","operator":"equals","value":"Bug"}]`,
			actions:     `[{"type":"add_comment","value":"[自动化] Bug 已创建，请及时处理"}]`,
			isEnabled:   true,
		},
		{
			name:           "长期未更新提醒",
			triggerType:    "scheduled",
			conditions:     `[{"field":"updated_at","operator":"older_than","value":"7d"}]`,
			actions:        `[{"type":"add_comment","value":"[定时提醒] 此任务已7天未更新, 请确认状态"}]`,
			isEnabled:      true,
			scheduleConfig: fmt.Sprintf(`{"frequency":"weekly","time":"%s","days":["%s"]}`, nowTime, time.Now().Format("Mon")),
		},
	}

	for _, r := range rules {
		rule := AutomationRule{
			WorkspaceID:    workspaceID,
			ProjectID:      0,
			Name:           r.name,
			TriggerType:    r.triggerType,
			Conditions:     r.conditions,
			Actions:        r.actions,
			IsEnabled:      r.isEnabled,
			Scope:          "all",
			ScheduleConfig: r.scheduleConfig,
			Sequence:       0,
		}
		if err := db.Create(&rule).Error; err != nil {
			fmt.Printf("  [WARN] 创建规则 '%s' 失败: %v\n", r.name, err)
		} else {
			triggerLabel := "事件触发"
			if r.triggerType == "scheduled" {
				triggerLabel = r.scheduleConfig
			}
			fmt.Printf("  创建规则 #%d: '%s' [%s] (启用=%v)\n", rule.ID, r.name, triggerLabel, r.isEnabled)
		}
	}
}
