// backend-go/internal/rql/executor.go

package rql

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// GORMExecutor RQL AST 到 GORM 查询的转换器
type GORMExecutor struct{}

// NewGORMExecutor 创建新的 GORM Executor
func NewGORMExecutor() *GORMExecutor {
	return &GORMExecutor{}
}

// QueryContext 查询上下文，包含表名和字段映射
type QueryContext struct {
	TableName string
	FieldMap  map[string]FieldMapping
}

// FieldMapping 字段映射
type FieldMapping struct {
	ColumnName string   // 数据库列名
	FieldType  string   // 字段类型: string, number, date, user
	JoinTable  string   // 关联表名（如有）
	JoinKey    string   // 关联字段（如有）
}

// NewIssueQueryContext 创建 Issue 查询上下文
func NewIssueQueryContext() *QueryContext {
	return &QueryContext{
		TableName: "issues",
		FieldMap: map[string]FieldMapping{
			"id":           {ColumnName: "id", FieldType: "number"},
			"name":         {ColumnName: "name", FieldType: "string"},
			"description":  {ColumnName: "description_stripped", FieldType: "string"},
			"state":        {ColumnName: "state_id", FieldType: "number", JoinTable: "states", JoinKey: "name"},
			"priority":     {ColumnName: "priority", FieldType: "string"},
			"assignee":     {ColumnName: "assignee_id", FieldType: "user", JoinTable: "issue_assignees"},
			"reporter":     {ColumnName: "reporter_id", FieldType: "number"},
			"label":        {ColumnName: "label_id", FieldType: "label", JoinTable: "issue_labels"},
			"cycle":        {ColumnName: "cycle_id", FieldType: "cycle", JoinTable: "issue_cycles"},
			"module":       {ColumnName: "module_id", FieldType: "module", JoinTable: "module_issues"},
			"project_id":   {ColumnName: "project_id", FieldType: "number"},
			"workspace_id": {ColumnName: "workspace_id", FieldType: "number"},
			"created_at":   {ColumnName: "created_at", FieldType: "date"},
			"updated_at":   {ColumnName: "updated_at", FieldType: "date"},
			"start_date":   {ColumnName: "start_date", FieldType: "date"},
			"target_date":  {ColumnName: "target_date", FieldType: "date"},
		},
	}
}

// NewCycleQueryContext 创建 Cycle 查询上下文
func NewCycleQueryContext() *QueryContext {
	return &QueryContext{
		TableName: "cycles",
		FieldMap: map[string]FieldMapping{
			"id":           {ColumnName: "id", FieldType: "number"},
			"name":         {ColumnName: "name", FieldType: "string"},
			"description":  {ColumnName: "description", FieldType: "string"},
			"start_date":   {ColumnName: "start_date", FieldType: "date"},
			"end_date":     {ColumnName: "end_date", FieldType: "date"},
			"completed_at": {ColumnName: "completed_at", FieldType: "date"},
			"cancelled_at": {ColumnName: "cancelled_at", FieldType: "date"},
			"project_id":   {ColumnName: "project_id", FieldType: "number"},
			"workspace_id": {ColumnName: "workspace_id", FieldType: "number"},
			"created_at":   {ColumnName: "created_at", FieldType: "date"},
			"updated_at":   {ColumnName: "updated_at", FieldType: "date"},
		},
	}
}

// NewModuleQueryContext 创建 Module 查询上下文
func NewModuleQueryContext() *QueryContext {
	return &QueryContext{
		TableName: "modules",
		FieldMap: map[string]FieldMapping{
			"id":           {ColumnName: "id", FieldType: "number"},
			"name":         {ColumnName: "name", FieldType: "string"},
			"description":  {ColumnName: "description", FieldType: "string"},
			"parent_id":    {ColumnName: "parent_id", FieldType: "number"},
			"order":        {ColumnName: "order", FieldType: "number"},
			"is_archived":  {ColumnName: "is_archived", FieldType: "string"},
			"project_id":   {ColumnName: "project_id", FieldType: "number"},
			"workspace_id": {ColumnName: "workspace_id", FieldType: "number"},
			"created_at":   {ColumnName: "created_at", FieldType: "date"},
			"updated_at":   {ColumnName: "updated_at", FieldType: "date"},
		},
	}
}

// Execute 执行 RQL AST 并将结果应用到 GORM 查询
func (e *GORMExecutor) Execute(db *gorm.DB, node Node, ctx *QueryContext) (*gorm.DB, error) {
	if node == nil {
		return db, nil
	}

	return e.buildQuery(db, node, ctx)
}

func (e *GORMExecutor) buildQuery(db *gorm.DB, node Node, ctx *QueryContext) (*gorm.DB, error) {
	switch n := node.(type) {
	case *BinaryExpr:
		return e.buildBinaryExpr(db, n, ctx)
	case *Comparison:
		return e.buildComparison(db, n, ctx)
	case *LikeExpr:
		return e.buildLikeExpr(db, n, ctx)
	case *InExpr:
		return e.buildInExpr(db, n, ctx)
	case *NotExpr:
		return e.buildNotExpr(db, n, ctx)
	default:
		return db, fmt.Errorf("unknown node type: %T", node)
	}
}

func (e *GORMExecutor) buildBinaryExpr(db *gorm.DB, expr *BinaryExpr, ctx *QueryContext) (*gorm.DB, error) {
	var resultDB *gorm.DB

	if expr.Operator == "AND" {
		resultDB = db.Where(func(d *gorm.DB) *gorm.DB {
			leftDB, err := e.buildQuery(d, expr.Left, ctx)
			if err != nil {
				return d
			}
			return leftDB
		}(db))

		rightDB, err := e.buildQuery(resultDB, expr.Right, ctx)
		if err != nil {
			return db, err
		}
		return rightDB, nil
	}

	if expr.Operator == "OR" {
		// OR 需要分别查询然后合并
		leftDB, err := e.buildQuery(db.Session(&gorm.Session{}), expr.Left, ctx)
		if err != nil {
			return db, err
		}

		rightDB, err := e.buildQuery(db.Session(&gorm.Session{}), expr.Right, ctx)
		if err != nil {
			return db, err
		}

		// 获取两个查询的 WHERE 条件
		leftSQL := leftDB.Session(&gorm.Session{}).Session(&gorm.Session{}).Statement.SQL.String()
		rightSQL := rightDB.Session(&gorm.Session{}).Statement.SQL.String()

		// 使用 OR 合并
		if leftSQL != "" && rightSQL != "" {
			args := append(e.flattenArgs(leftDB.Statement.Vars...), e.flattenArgs(rightDB.Statement.Vars...)...)
			return db.Where("("+leftSQL+") OR ("+rightSQL+")", args...), nil
		}

		if leftSQL != "" {
			return db.Where(leftSQL, e.flattenArgs(leftDB.Statement.Vars...)...), nil
		}

		if rightSQL != "" {
			return db.Where(rightSQL, e.flattenArgs(rightDB.Statement.Vars...)...), nil
		}

		return db, nil
	}

	return db, nil
}

func (e *GORMExecutor) flattenArgs(args ...interface{}) []interface{} {
	var result []interface{}
	for _, arg := range args {
		if slice, ok := arg.([]interface{}); ok {
			result = append(result, slice...)
		} else {
			result = append(result, arg)
		}
	}
	return result
}

func (e *GORMExecutor) buildComparison(db *gorm.DB, expr *Comparison, ctx *QueryContext) (*gorm.DB, error) {
	mapping, ok := ctx.FieldMap[expr.Field]
	if !ok {
		// 未知字段，忽略
		return db, nil
	}

	switch expr.Operator {
	case "=":
		return e.buildEqual(db, expr.Field, expr.Value, mapping, ctx)
	case "!=":
		return e.buildNotEqual(db, expr.Field, expr.Value, mapping, ctx)
	case ">":
		return db.Where(fmt.Sprintf("%s > ?", mapping.ColumnName), expr.Value), nil
	case "<":
		return db.Where(fmt.Sprintf("%s < ?", mapping.ColumnName), expr.Value), nil
	case ">=":
		return db.Where(fmt.Sprintf("%s >= ?", mapping.ColumnName), expr.Value), nil
	case "<=":
		return db.Where(fmt.Sprintf("%s <= ?", mapping.ColumnName), expr.Value), nil
	default:
		return db, fmt.Errorf("unsupported operator: %s", expr.Operator)
	}
}

func (e *GORMExecutor) buildEqual(db *gorm.DB, field string, value interface{}, mapping FieldMapping, ctx *QueryContext) (*gorm.DB, error) {
	switch mapping.FieldType {
	case "string", "number", "date":
		return db.Where(fmt.Sprintf("%s = ?", mapping.ColumnName), value), nil

	case "user":
		// assignee 特殊处理，需要关联查询
		if field == "assignee" {
			return db.Joins("JOIN issue_assignees ON issue_assignees.issue_id = issues.id").
				Where("issue_assignees.user_id = ?", value), nil
		}
		return db.Where(fmt.Sprintf("%s = ?", mapping.ColumnName), value), nil

	case "label":
		return db.Joins("JOIN issue_labels ON issue_labels.issue_id = issues.id").
			Where("issue_labels.label_id = ?", value), nil

	case "cycle":
		return db.Joins("JOIN issue_cycles ON issue_cycles.issue_id = issues.id").
			Where("issue_cycles.cycle_id = ?", value), nil

	case "module":
		return db.Joins("JOIN module_issues ON module_issues.issue_id = issues.id").
			Where("module_issues.module_id = ?", value), nil

	default:
		return db.Where(fmt.Sprintf("%s = ?", mapping.ColumnName), value), nil
	}
}

func (e *GORMExecutor) buildNotEqual(db *gorm.DB, field string, value interface{}, mapping FieldMapping, ctx *QueryContext) (*gorm.DB, error) {
	return db.Where(fmt.Sprintf("%s != ?", mapping.ColumnName), value), nil
}

func (e *GORMExecutor) buildLikeExpr(db *gorm.DB, expr *LikeExpr, ctx *QueryContext) (*gorm.DB, error) {
	mapping, ok := ctx.FieldMap[expr.Field]
	if !ok {
		return db, nil
	}

	pattern := "%" + expr.Value + "%"

	switch mapping.FieldType {
	case "string":
		return db.Where(fmt.Sprintf("%s LIKE ?", mapping.ColumnName), pattern), nil
	default:
		return db.Where(fmt.Sprintf("%s LIKE ?", mapping.ColumnName), pattern), nil
	}
}

func (e *GORMExecutor) buildInExpr(db *gorm.DB, expr *InExpr, ctx *QueryContext) (*gorm.DB, error) {
	mapping, ok := ctx.FieldMap[expr.Field]
	if !ok {
		return db, nil
	}

	if len(expr.Values) == 0 {
		return db.Where("1 = 0"), nil
	}

	placeholders := make([]string, len(expr.Values))
	args := make([]interface{}, len(expr.Values))
	for i, v := range expr.Values {
		placeholders[i] = "?"
		args[i] = v
	}

	whereClause := fmt.Sprintf("%s IN (%s)", mapping.ColumnName, strings.Join(placeholders, ", "))
	return db.Where(whereClause, args...), nil
}

func (e *GORMExecutor) buildNotExpr(db *gorm.DB, expr *NotExpr, ctx *QueryContext) (*gorm.DB, error) {
	subDB, err := e.buildQuery(db, expr.Expr, ctx)
	if err != nil {
		return db, err
	}
	return db.Not(subDB), nil
}