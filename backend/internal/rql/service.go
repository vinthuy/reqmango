package rql

import (
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

// SearchIssues 搜索工作项
func (s *RQLService) SearchIssues(db *gorm.DB, projectID uint64, rqlQuery string, page, pageSize int) ([]model.Issue, int64, error) {
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

	// 构建查询上下文
	ctx := NewIssueQueryContext()

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
