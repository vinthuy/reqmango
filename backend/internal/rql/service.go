package rql

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// RQLService 处理 RQL 查询服务
type RQLService struct {
	executor *GORMExecutor
}

// NewRQLService 创建 RQL Service
func NewRQLService() *RQLService {
	return &RQLService{
		executor: NewGORMExecutor(),
	}
}

// ResolveTemplateVars 替换 RQL 中的模板变量
// 支持的变量：$CURRENT_USER, $TODAY, $END_OF_WEEK, $ONE_WEEK_AGO
func ResolveTemplateVars(rqlQuery string, currentUserID uint64) string {
	now := time.Now()
	today := now.Format("2006-01-02")

	// End of current week (Sunday)
	daysUntilSunday := (7 - int(now.Weekday())) % 7
	endOfWeek := now.AddDate(0, 0, daysUntilSunday).Format("2006-01-02")

	oneWeekAgo := now.AddDate(0, 0, -7).Format("2006-01-02")

	result := rqlQuery
	result = strings.ReplaceAll(result, "$CURRENT_USER", strconv.FormatUint(currentUserID, 10))
	// Date values: use raw YYYY-MM-DD format.
	// $TODAY (unquoted) → 2026-07-05 (parsed as TOKEN_DATE by the RQL lexer)
	// "$TODAY" (quoted) → "2026-07-05" (parsed as TOKEN_STRING by the RQL lexer)
	// Both forms work because the RQL parser handles both TOKEN_DATE and TOKEN_STRING.
	result = strings.ReplaceAll(result, "$TODAY", today)
	result = strings.ReplaceAll(result, "$END_OF_WEEK", endOfWeek)
	result = strings.ReplaceAll(result, "$ONE_WEEK_AGO", oneWeekAgo)
	return result
}

// SearchIssues 搜索工作项
func (s *RQLService) SearchIssues(db *gorm.DB, projectID uint64, rqlQuery string, page, pageSize int) ([]model.Issue, int64, error) {
	return s.SearchIssuesWithUser(db, projectID, rqlQuery, page, pageSize, 0)
}

// SearchIssuesWithUser 搜索工作项（带用户上下文，用于解析 $CURRENT_USER 等模板变量）
func (s *RQLService) SearchIssuesWithUser(db *gorm.DB, projectID uint64, rqlQuery string, page, pageSize int, currentUserID uint64) ([]model.Issue, int64, error) {
	// 解析模板变量
	if currentUserID > 0 {
		rqlQuery = ResolveTemplateVars(rqlQuery, currentUserID)
	}

	// Strip orderby clauses before parsing — sort is handled via sort_config/sort_by params.
	// Split by AND, filter out orderby clauses, then rejoin.
	if strings.Contains(strings.ToLower(rqlQuery), "orderby") {
		parts := strings.Split(rqlQuery, " AND ")
		var filtered []string
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if !strings.HasPrefix(strings.ToLower(trimmed), "orderby") {
				filtered = append(filtered, trimmed)
			}
		}
		rqlQuery = strings.Join(filtered, " AND ")
	}

	// 词法分析
	lexer := NewLexer(rqlQuery)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, 0, fmt.Errorf("lexer error: %w", err)
	}

	// 语法分析
	parser := NewParser(tokens)
	ast, parseErr := parser.Parse()
	if parseErr != nil {
		return nil, 0, parseErr
	}

	// 构建查询上下文
	ctx := NewIssueQueryContext(db, projectID)

	// 基础查询：只查询该项目的未归档工作项
	query := db.Model(&model.Issue{}).
		Where("issues.project_id = ?", projectID).
		Where("issues.archived_at IS NULL")

	// 应用 RQL 条件
	if ast != nil {
		var execErr error
		query, execErr = s.executor.Execute(query, ast, ctx)
		if execErr != nil {
			return nil, 0, execErr
		}
	}

	// 计算总数
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (page - 1) * pageSize
	var issues []model.Issue
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&issues).Error; err != nil {
		return nil, 0, err
	}

	return issues, total, nil
}

// SearchCycles 搜索周期
func (s *RQLService) SearchCycles(db *gorm.DB, projectID uint64, rqlQuery string, page, pageSize int) ([]model.Cycle, int64, error) {
	// 词法分析
	lexer := NewLexer(rqlQuery)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, 0, err
	}

	// 语法分析
	parser := NewParser(tokens)
	ast, parseErr := parser.Parse()
	if parseErr != nil {
		return nil, 0, parseErr
	}

	// 基础查询
	query := db.Model(&model.Cycle{}).
		Where("project_id = ?", projectID)

	// 应用 RQL 条件
	if ast != nil {
		var execErr error
		query, execErr = s.executor.Execute(query, ast, NewCycleQueryContext())
		if execErr != nil {
			return nil, 0, execErr
		}
	}

	// 计算总数
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (page - 1) * pageSize
	var cycles []model.Cycle
	if err := query.Offset(offset).Limit(pageSize).Order("start_date DESC").Find(&cycles).Error; err != nil {
		return nil, 0, err
	}

	return cycles, total, nil
}

// SearchModules 搜索模块
func (s *RQLService) SearchModules(db *gorm.DB, projectID uint64, rqlQuery string, page, pageSize int) ([]model.Module, int64, error) {
	// 词法分析
	lexer := NewLexer(rqlQuery)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, 0, err
	}

	// 语法分析
	parser := NewParser(tokens)
	ast, parseErr := parser.Parse()
	if parseErr != nil {
		return nil, 0, parseErr
	}

	// 基础查询
	query := db.Model(&model.Module{}).
		Where("project_id = ?", projectID).
		Where("is_archived = ?", false)

	// 应用 RQL 条件
	if ast != nil {
		var execErr error
		query, execErr = s.executor.Execute(query, ast, NewModuleQueryContext())
		if execErr != nil {
			return nil, 0, execErr
		}
	}

	// 计算总数
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (page - 1) * pageSize
	var modules []model.Module
	if err := query.Offset(offset).Limit(pageSize).Order("name ASC").Find(&modules).Error; err != nil {
		return nil, 0, err
	}

	return modules, total, nil
}
