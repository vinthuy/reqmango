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
	Description     string     `json:"description"`
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

func (AutomationRule) TableName() string { return "automation_rules" }

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

func (AutomationExecution) TableName() string { return "automation_executions" }

type Workspace struct {
	ID   uint64 `gorm:"primaryKey"`
	Name string
	Slug string
}

type Issue struct {
	ID          uint64 `gorm:"primaryKey"`
	ProjectID   uint64
	WorkspaceID uint64
	DeletedAt   *time.Time
}

func main() {
	dbURL := flag.String("db", "", "Database URL (default from DATABASE_URL env)")
	flag.Parse()

	if *dbURL == "" {
		*dbURL = os.Getenv("DATABASE_URL")
	}
	if *dbURL == "" {
		*dbURL = "postgres://postgres:postgres@localhost:5432/reqmango?sslmode=disable"
	}

	fmt.Println("================================================")
	fmt.Println(" Webhook 自动化规则种子工具")
	fmt.Println("================================================")
	fmt.Println()

	db, err := gorm.Open(postgres.Open(*dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Printf("[OK] 数据库连接成功\n")

	// 1. 获取所有工作区
	var workspaces []Workspace
	db.Find(&workspaces)
	if len(workspaces) == 0 {
		log.Fatalf("没有找到工作区，请先运行主程序初始化数据")
	}
	fmt.Printf("[OK] 找到 %d 个工作区\n", len(workspaces))

	now := time.Now()
	totalRulesCreated := 0
	totalLogsCreated := 0

	for _, ws := range workspaces {
		fmt.Printf("\n--- 工作区 #%d: %s ---\n", ws.ID, ws.Name)

		// 检查该工作区是否已有 webhook 规则
		var existingCount int64
		db.Model(&AutomationRule{}).Where("workspace_id = ? AND trigger_type = ? AND actions LIKE ?",
			ws.ID, "issue.created", "%call_webhook%").Count(&existingCount)
		if existingCount > 0 {
			fmt.Printf("  已有 webhook 规则, 跳过创建\n")
			// Still try to add execution logs for existing webhook rules
		} else {
			// 创建 Webhook 规则
			actions := []map[string]interface{}{
				{
					"type":  "call_webhook",
					"field": "https://httpbin.org/post",
					"value": map[string]interface{}{
						"method": "POST",
						"headers": map[string]string{
							"Content-Type": "application/json",
						},
						"body": `{"event":"issue_created","issue_id":{{issue_id}},"workspace_id":{{workspace_id}},"trigger":"{{trigger_type}}"}`,
					},
				},
			}
			actionsJSON, _ := json.Marshal(actions)

			rule := AutomationRule{
				WorkspaceID: ws.ID,
				ProjectID:   0,
				Name:        "Webhook通知示例",
				Description: "当创建新工作项时，通过Webhook通知外部系统",
				TriggerType: "issue.created",
				Conditions:  "[]",
				Actions:     string(actionsJSON),
				IsEnabled:   true,
				Scope:       "all",
				Sequence:    5,
			}
			if err := db.Create(&rule).Error; err != nil {
				fmt.Printf("  [失败] 创建 webhook 规则失败: %v\n", err)
				continue
			}
			fmt.Printf("  [创建] webhook 规则 #%d: '%s'\n", rule.ID, rule.Name)
			totalRulesCreated++
		}

		// 2. 获取该工作区下的 webhook 规则
		var webhookRules []AutomationRule
		db.Where("workspace_id = ? AND actions LIKE ?", ws.ID, "%call_webhook%").Find(&webhookRules)

		for _, rule := range webhookRules {
			// 查找一个 issue
			var issue Issue
			if err := db.Where("workspace_id = ? AND deleted_at IS NULL", ws.ID).First(&issue).Error; err != nil {
				fmt.Printf("  [警告] 工作区没有 Issue, 跳过运行日志: %v\n", err)
				continue
			}

			// 检查是否已有执行日志
			var logCount int64
			db.Model(&AutomationExecution{}).Where("rule_id = ?", rule.ID).Count(&logCount)
			if logCount > 0 {
				fmt.Printf("  规则 #%d 已有 %d 条运行日志, 跳过\n", rule.ID, logCount)
				continue
			}

			// 3. 生成执行日志
			// 成功记录
			ctxSuccess := map[string]interface{}{
				"event_type":   "issue.created",
				"trigger_type": "issue.created",
				"project_id":   issue.ProjectID,
				"workspace_id": ws.ID,
			}
			ctxJSON, _ := json.Marshal(ctxSuccess)

			actionsSuccess := []map[string]interface{}{
				{
					"type":          "call_webhook",
					"field":         "https://httpbin.org/post",
					"status":        "success",
					"response_code": 200,
				},
			}
			actionsJSON, _ := json.Marshal(actionsSuccess)

			db.Create(&AutomationExecution{
				RuleID:       rule.ID,
				IssueID:      issue.ID,
				TriggerType:  "issue.created",
				ContextJSON:  string(ctxJSON),
				ActionsTaken: string(actionsJSON),
				Status:       "success",
				Duration:     rand.Int63n(300) + 100, // 100-400ms
				ExecutedAt:   now.Add(-1 * time.Hour),
			})
			fmt.Printf("  [日志] 规则 #%d: success (HTTP 200, %dms)\n", rule.ID, rand.Int63n(300)+100)
			totalLogsCreated++

			// 失败记录 (超时)
			actionsFailed := []map[string]interface{}{
				{
					"type":   "call_webhook",
					"field":  "https://httpbin.org/post",
					"status": "failed",
				},
			}
			failActionsJSON, _ := json.Marshal(actionsFailed)

			db.Create(&AutomationExecution{
				RuleID:       rule.ID,
				IssueID:      issue.ID,
				TriggerType:  "issue.created",
				ContextJSON:  string(ctxJSON),
				ActionsTaken: string(failActionsJSON),
				Status:       "failed",
				Error:        "context deadline exceeded (Client.Timeout exceeded while awaiting headers)",
				Duration:     15000,
				ExecutedAt:   now.Add(-30 * time.Minute),
			})
			fmt.Printf("  [日志] 规则 #%d: failed (timeout, 15000ms)\n", rule.ID)
			totalLogsCreated++
		}
	}

	// 4. 统计
	fmt.Println()
	fmt.Println("================================================")
	fmt.Println(" 种子数据统计")
	fmt.Println("================================================")
	fmt.Printf("  新创建 webhook 规则:  %d 条\n", totalRulesCreated)
	fmt.Printf("  新创建执行日志:      %d 条\n", totalLogsCreated)

	// 显示所有 webhook 规则及执行日志数
	fmt.Println()
	fmt.Println("--- 所有 Webhook 规则 ---")
	var allWebhookRules []AutomationRule
	db.Where("actions LIKE ?", "%call_webhook%").Find(&allWebhookRules)

	if len(allWebhookRules) == 0 {
		fmt.Println("  (无)")
	} else {
		for _, r := range allWebhookRules {
			var cnt int64
			db.Model(&AutomationExecution{}).Where("rule_id = ?", r.ID).Count(&cnt)
			enabled := "禁用"
			if r.IsEnabled {
				enabled = "启用"
			}
			fmt.Printf("  #%-4d | %-30s | ws=%d | %s | 日志数=%d\n", r.ID, r.Name, r.WorkspaceID, enabled, cnt)
		}
	}

	fmt.Println()
	fmt.Println("现在可以到前端查看:")
	fmt.Println("  项目设置 > 自动化规则 > 查看 Webhook通知示例")
	fmt.Println("  点击运行日志图标查看执行历史")
}
