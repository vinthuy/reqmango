package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/config"
	"github.com/reqmanpy/backend-go/internal/model"
	"github.com/reqmanpy/backend-go/internal/router"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg := config.Load()

	// Database connection
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	fmt.Println("Database connected")

	// Auto-migrate
	if err := db.AutoMigrate(
		&model.User{},
		&model.Workspace{},
		&model.WorkspaceMember{},
		&model.Project{},
		&model.ProjectMember{},
		&model.State{},
		&model.StateTransition{},
		&model.Label{},
		&model.Issue{},
		&model.IssueAssignee{},
		&model.IssueLabel{},
		&model.IssueCycle{},
		&model.IssueActivity{},
		&model.Cycle{},
		&model.Module{},
		&model.ModuleIssue{},
	); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}
	fmt.Println("Database migration completed")

	// Seed demo data
	seedDemoData(db)

	// Setup Gin
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	router.SetupRoutes(r, db, cfg)

	addr := fmt.Sprintf(":%s", cfg.Port)
	fmt.Printf("Server starting on %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func seedDemoData(db *gorm.DB) {
	// Check if demo users already exist
	var count int64
	db.Model(&model.User{}).Where("email LIKE 'demo%'").Count(&count)
	if count > 0 {
		fmt.Println("Demo data already exists, skipping seed")
		return
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	hash, _ := bcrypt.GenerateFromPassword([]byte("demo1234"), bcrypt.DefaultCost)

	// --- 20 users ---
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

	// --- Workspace ---
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

	// --- Project ---
	desc := "A demo project for testing"
	proj := model.Project{
		Name: "Demo Project", Identifier: "DEMO",
		Description: &desc, WorkspaceID: ws.ID,
	}
	if err := db.Create(&proj).Error; err != nil {
		fmt.Printf("Failed to create demo project: %v\n", err)
		return
	}

	// --- States ---
	stateDefs := []struct{ name, color string }{
		{"待处理 (Backlog)", "#6B7280"},
		{"待办 (Todo)", "#3B82F6"},
		{"进行中 (In Progress)", "#F59E0B"},
		{"评审中 (In Review)", "#8B5CF6"},
		{"已完成 (Done)", "#10B981"},
		{"已取消 (Cancelled)", "#EF4444"},
	}
	states := make([]model.State, len(stateDefs))
	for i, sd := range stateDefs {
		states[i] = model.State{
			Name: sd.name, Color: sd.color, Sequence: i + 1,
			ProjectID: proj.ID, WorkspaceID: ws.ID,
		}
		db.Create(&states[i])
	}
	fmt.Printf("Created %d states\n", len(states))

	// --- Cycles (4 sprints) ---
	cycleNames := []string{"Sprint 1", "Sprint 2", "Sprint 3", "Sprint 4"}
	cycles := make([]model.Cycle, len(cycleNames))
	baseDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	for i, name := range cycleNames {
		start := baseDate.AddDate(0, 0, i*14)
		end := start.AddDate(0, 0, 13)
		cycles[i] = model.Cycle{
			Name: name, StartDate: start, EndDate: &end,
			ProjectID: proj.ID, WorkspaceID: ws.ID,
		}
		db.Create(&cycles[i])
	}
	fmt.Printf("Created %d cycles\n", len(cycles))

	// --- Modules ---
	moduleNames := []string{"用户模块", "订单模块", "支付模块", "报表模块", "系统设置"}
	modules := make([]model.Module, len(moduleNames))
	for i, name := range moduleNames {
		modules[i] = model.Module{
			Name: name, ProjectID: proj.ID, WorkspaceID: ws.ID,
		}
		db.Create(&modules[i])
	}
	fmt.Printf("Created %d modules\n", len(modules))

	// --- 100 Issues ---
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
		"实现事件溯源机制", "添加数据校验规则", "优化文件存储方案", "修复异常处理逻辑",
		"实现弹性伸缩配置", "添加用户行为追踪", "优化网络请求策略", "修复SQL注入风险",
		"实现微服务拆分方案", "添加审计日志功能", "优化状态管理架构", "修复内存泄漏问题",
		"实现蓝绿部署策略", "添加数据脱敏功能", "优化路由加载速度", "修复XSS安全漏洞",
		"实现服务熔断机制", "添加版本控制功能", "优化事务处理性能", "修复死锁检测机制",
		"实现服务降级方案", "添加依赖注入框架", "优化异常捕获策略", "修复数据迁移脚本",
		"实现流量控制策略", "添加配置中心集成", "优化任务队列调度", "修复接口幂等性",
		"实现链路追踪系统", "添加灰度实验平台", "优化分布式锁机制", "修复数据同步延迟",
		"实现限流熔断器", "添加健康检查接口", "优化数据分片策略", "修复数据丢失问题",
	}

	for i := 0; i < 100; i++ {
		stateIdx := rng.Intn(len(states))
		cycleIdx := rng.Intn(len(cycles) + 1) // +1 for no cycle
		priorityIdx := rng.Intn(len(priorities))
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

		// Assign random cycle (70% chance)
		if cycleIdx < len(cycles) {
			db.Create(&model.IssueCycle{
				IssueID: issue.ID,
				CycleID: cycles[cycleIdx].ID,
			})
		}

		// Assign 1-3 random assignees
		numAssignees := 1 + rng.Intn(3)
		rng.Shuffle(len(users), func(a, b int) { users[a], users[b] = users[b], users[a] })
		for j := 0; j < numAssignees && j < len(users); j++ {
			db.Create(&model.IssueAssignee{
				IssueID: issue.ID,
				UserID:  users[j].ID,
			})
		}
	}
	fmt.Println("Created 100 issues with random states, cycles, and assignees")
}
