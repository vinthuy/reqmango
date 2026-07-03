// backend/internal/rql/executor.go

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
	ColumnName string // 数据库列名
	FieldType  string // 字段类型: string, number, date, user
	JoinTable  string // 关联表名（如有）
	JoinKey    string // 关联字段（如有）
}

// rawCondition 原始 SQL 条件片段
type rawCondition struct {
	SQL   string
	Args  []interface{}
	Joins []string
}

// NewIssueQueryContext 创建 Issue 查询上下文
func NewIssueQueryContext() *QueryContext {
	return &QueryContext{
		TableName: "issues",
		FieldMap: map[string]FieldMapping{
			"id":            {ColumnName: "issues.id", FieldType: "number"},
			"sequence_id":   {ColumnName: "issues.sequence_id", FieldType: "number"},
			"name":          {ColumnName: "issues.name", FieldType: "string"},
			"description":   {ColumnName: "issues.description_stripped", FieldType: "string"},
			"state":         {ColumnName: "issues.state_id", FieldType: "number", JoinTable: "states", JoinKey: "name"},
			"state_id":      {ColumnName: "issues.state_id", FieldType: "number"},
			"state_group":   {ColumnName: "state_group", FieldType: "state_group"},
			"priority":      {ColumnName: "issues.priority", FieldType: "string"},
			"assignee":      {ColumnName: "assignee_id", FieldType: "user", JoinTable: "issue_assignees"},
			"assignee_id":   {ColumnName: "assignee_id", FieldType: "user", JoinTable: "issue_assignees"},
			"reporter":      {ColumnName: "issues.reporter_id", FieldType: "number"},
			"label":         {ColumnName: "label_id", FieldType: "label", JoinTable: "issue_labels"},
			"cycle":         {ColumnName: "cycle_id", FieldType: "cycle", JoinTable: "issue_cycles"},
			"cycle_id":      {ColumnName: "cycle_id", FieldType: "cycle", JoinTable: "issue_cycles"},
			"module":        {ColumnName: "module_id", FieldType: "module", JoinTable: "module_issues"},
			"module_id":     {ColumnName: "module_id", FieldType: "module", JoinTable: "module_issues"},
			"issue_type_id": {ColumnName: "issues.issue_type_id", FieldType: "number"},
			"project_id":    {ColumnName: "issues.project_id", FieldType: "number"},
			"workspace_id":  {ColumnName: "issues.workspace_id", FieldType: "number"},
			"created_at":    {ColumnName: "issues.created_at", FieldType: "date"},
			"updated_at":    {ColumnName: "issues.updated_at", FieldType: "date"},
			"start_date":    {ColumnName: "issues.start_date", FieldType: "date"},
			"target_date":   {ColumnName: "issues.target_date", FieldType: "date"},
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

	// 使用 rawCondition 统一构建 WHERE 子句
	cond, err := e.buildRawCondition(node, ctx)
	if err != nil {
		return db, err
	}

	// 应用 JOIN
	for _, join := range cond.Joins {
		db = db.Joins(join)
	}

	// 应用 WHERE
	if cond.SQL != "" {
		db = db.Where(cond.SQL, cond.Args...)
	}

	return db, nil
}

// buildRawCondition 将 AST 节点转换为原始 SQL 条件
func (e *GORMExecutor) buildRawCondition(node Node, ctx *QueryContext) (*rawCondition, error) {
	switch n := node.(type) {
	case *BinaryExpr:
		return e.buildBinaryRaw(n, ctx)
	case *Comparison:
		return e.buildComparisonRaw(n, ctx)
	case *LikeExpr:
		return e.buildLikeRaw(n, ctx)
	case *InExpr:
		return e.buildInRaw(n, ctx)
	case *NotExpr:
		return e.buildNotRaw(n, ctx)
	case *NullCheck:
		return e.buildNullRaw(n, ctx)
	default:
		return nil, fmt.Errorf("unknown node type: %T", node)
	}
}

func (e *GORMExecutor) buildBinaryRaw(expr *BinaryExpr, ctx *QueryContext) (*rawCondition, error) {
	left, err := e.buildRawCondition(expr.Left, ctx)
	if err != nil {
		return nil, err
	}

	right, err := e.buildRawCondition(expr.Right, ctx)
	if err != nil {
		return nil, err
	}

	// 合并 JOIN
	joins := append(left.Joins, right.Joins...)

	// 合并参数
	args := append(left.Args, right.Args...)

	op := " AND "
	if expr.Operator == "OR" {
		op = " OR "
	}

	sql := "(" + left.SQL + op + right.SQL + ")"

	return &rawCondition{SQL: sql, Args: args, Joins: joins}, nil
}

func (e *GORMExecutor) buildComparisonRaw(expr *Comparison, ctx *QueryContext) (*rawCondition, error) {
	mapping, ok := ctx.FieldMap[expr.Field]
	if !ok {
		if strings.HasPrefix(expr.Field, "cf_") {
			return e.buildCustomFieldComparison(expr.Field, expr.Operator, expr.Value)
		}
		return &rawCondition{SQL: "1 = 1"}, nil
	}

	switch expr.Operator {
	case "=":
		return e.buildEqualRaw(expr.Value, mapping)
	case "!=":
		return e.buildNotEqualRaw(expr.Value, mapping)
	case ">":
		return &rawCondition{SQL: fmt.Sprintf("%s > ?", mapping.ColumnName), Args: []interface{}{expr.Value}}, nil
	case "<":
		return &rawCondition{SQL: fmt.Sprintf("%s < ?", mapping.ColumnName), Args: []interface{}{expr.Value}}, nil
	case ">=":
		return &rawCondition{SQL: fmt.Sprintf("%s >= ?", mapping.ColumnName), Args: []interface{}{expr.Value}}, nil
	case "<=":
		return &rawCondition{SQL: fmt.Sprintf("%s <= ?", mapping.ColumnName), Args: []interface{}{expr.Value}}, nil
	default:
		return nil, fmt.Errorf("unsupported operator: %s", expr.Operator)
	}
}

func (e *GORMExecutor) buildCustomFieldComparison(fieldName, operator, value interface{}) (*rawCondition, error) {
	fieldID := strings.TrimPrefix(fieldName.(string), "cf_")
	join := fmt.Sprintf("JOIN issue_custom_field_values icfv ON icfv.issue_id = issues.id AND icfv.field_id = %s", fieldID)

	var sql string
	switch operator {
	case "=":
		sql = "icfv.value = ?"
	case "!=":
		sql = "icfv.value != ?"
	case ">":
		sql = "icfv.value > ?"
	case "<":
		sql = "icfv.value < ?"
	case ">=":
		sql = "icfv.value >= ?"
	case "<=":
		sql = "icfv.value <= ?"
	default:
		return nil, fmt.Errorf("unsupported operator for custom field: %s", operator)
	}

	return &rawCondition{
		SQL:   sql,
		Args:  []interface{}{value},
		Joins: []string{join},
	}, nil
}

func (e *GORMExecutor) buildEqualRaw(value interface{}, mapping FieldMapping) (*rawCondition, error) {
	switch mapping.FieldType {
	case "string", "number", "date":
		return &rawCondition{SQL: fmt.Sprintf("%s = ?", mapping.ColumnName), Args: []interface{}{value}}, nil

	case "state_group":
		return &rawCondition{
			SQL:   "states.group = ?",
			Args:  []interface{}{value},
			Joins: []string{"JOIN states ON states.id = issues.state_id"},
		}, nil

	case "user":
		return &rawCondition{
			SQL:   "issue_assignees.user_id = ?",
			Args:  []interface{}{value},
			Joins: []string{"JOIN issue_assignees ON issue_assignees.issue_id = issues.id"},
		}, nil

	case "label":
		return &rawCondition{
			SQL:   "issue_labels.label_id = ?",
			Args:  []interface{}{value},
			Joins: []string{"JOIN issue_labels ON issue_labels.issue_id = issues.id"},
		}, nil

	case "cycle":
		return &rawCondition{
			SQL:   "issue_cycles.cycle_id = ?",
			Args:  []interface{}{value},
			Joins: []string{"JOIN issue_cycles ON issue_cycles.issue_id = issues.id"},
		}, nil

	case "module":
		return &rawCondition{
			SQL:   "module_issues.module_id = ?",
			Args:  []interface{}{value},
			Joins: []string{"JOIN module_issues ON module_issues.issue_id = issues.id"},
		}, nil

	default:
		return &rawCondition{SQL: fmt.Sprintf("%s = ?", mapping.ColumnName), Args: []interface{}{value}}, nil
	}
}

func (e *GORMExecutor) buildNotEqualRaw(value interface{}, mapping FieldMapping) (*rawCondition, error) {
	// Handle special field types that require JOINs (consistent with buildEqualRaw)
	switch mapping.FieldType {
	case "state_group":
		return &rawCondition{
			SQL:   "states.group != ?",
			Args:  []interface{}{value},
			Joins: []string{"JOIN states ON states.id = issues.state_id"},
		}, nil
	case "user":
		return &rawCondition{
			SQL:   "issue_assignees.user_id != ?",
			Args:  []interface{}{value},
			Joins: []string{"JOIN issue_assignees ON issue_assignees.issue_id = issues.id"},
		}, nil
	case "label":
		return &rawCondition{
			SQL:   "issue_labels.label_id != ?",
			Args:  []interface{}{value},
			Joins: []string{"JOIN issue_labels ON issue_labels.issue_id = issues.id"},
		}, nil
	case "cycle":
		return &rawCondition{
			SQL:   "issue_cycles.cycle_id != ?",
			Args:  []interface{}{value},
			Joins: []string{"JOIN issue_cycles ON issue_cycles.issue_id = issues.id"},
		}, nil
	case "module":
		return &rawCondition{
			SQL:   "module_issues.module_id != ?",
			Args:  []interface{}{value},
			Joins: []string{"JOIN module_issues ON module_issues.issue_id = issues.id"},
		}, nil
	}
	return &rawCondition{SQL: fmt.Sprintf("%s != ?", mapping.ColumnName), Args: []interface{}{value}}, nil
}

func (e *GORMExecutor) buildLikeRaw(expr *LikeExpr, ctx *QueryContext) (*rawCondition, error) {
	mapping, ok := ctx.FieldMap[expr.Field]
	if !ok {
		if strings.HasPrefix(expr.Field, "cf_") {
			return e.buildCustomFieldLike(expr.Field, expr.Operator, expr.Value)
		}
		return &rawCondition{SQL: "1 = 1"}, nil
	}

	op := "ILIKE"
	if expr.Operator == "NOT LIKE" {
		op = "NOT ILIKE"
	}

	value := expr.Value
	if !strings.ContainsAny(value, "%_") {
		value = fmt.Sprintf("%%%s%%", value)
	}

	if expr.Field == "name" || expr.Field == "description" {
		return e.buildFullTextSearch(expr.Field, expr.Operator, expr.Value)
	}

	return &rawCondition{
		SQL:  fmt.Sprintf("%s %s ?", mapping.ColumnName, op),
		Args: []interface{}{value},
	}, nil
}

func (e *GORMExecutor) buildFullTextSearch(field, operator, value string) (*rawCondition, error) {
	tsQuery := fmt.Sprintf("to_tsquery('english', '%s')", strings.ReplaceAll(value, " ", " & "))

	if operator == "NOT LIKE" {
		return &rawCondition{
			SQL: fmt.Sprintf("to_tsvector('english', COALESCE(name, '') || ' ' || COALESCE(description_stripped, '')) !@@ %s", tsQuery),
		}, nil
	}

	return &rawCondition{
		SQL: fmt.Sprintf("to_tsvector('english', COALESCE(name, '') || ' ' || COALESCE(description_stripped, '')) @@ %s", tsQuery),
	}, nil
}

func (e *GORMExecutor) buildCustomFieldLike(fieldName, operator, value string) (*rawCondition, error) {
	fieldID := strings.TrimPrefix(fieldName, "cf_")
	join := fmt.Sprintf("JOIN issue_custom_field_values icfv ON icfv.issue_id = issues.id AND icfv.field_id = %s", fieldID)

	op := "ILIKE"
	if operator == "NOT LIKE" {
		op = "NOT ILIKE"
	}

	if !strings.ContainsAny(value, "%_") {
		value = fmt.Sprintf("%%%s%%", value)
	}

	return &rawCondition{
		SQL:   fmt.Sprintf("icfv.value %s ?", op),
		Args:  []interface{}{value},
		Joins: []string{join},
	}, nil
}

func (e *GORMExecutor) buildInRaw(expr *InExpr, ctx *QueryContext) (*rawCondition, error) {
	mapping, ok := ctx.FieldMap[expr.Field]
	if !ok {
		if strings.HasPrefix(expr.Field, "cf_") {
			return e.buildCustomFieldIn(expr.Field, expr.Operator, expr.Values)
		}
		return &rawCondition{SQL: "1 = 1"}, nil
	}

	if len(expr.Values) == 0 {
		if expr.Operator == "NOT IN" {
			return &rawCondition{SQL: "1 = 1"}, nil
		}
		return &rawCondition{SQL: "1 = 0"}, nil
	}

	placeholders := make([]string, len(expr.Values))
	args := make([]interface{}, len(expr.Values))
	for i, v := range expr.Values {
		placeholders[i] = "?"
		args[i] = v
	}

	op := "IN"
	if expr.Operator == "NOT IN" {
		op = "NOT IN"
	}
	placeholderList := strings.Join(placeholders, ", ")

	// Handle special field types that require JOINs (consistent with buildEqualRaw)
	switch mapping.FieldType {
	case "state_group":
		return &rawCondition{
			SQL:   fmt.Sprintf("states.group %s (%s)", op, placeholderList),
			Args:  args,
			Joins: []string{"JOIN states ON states.id = issues.state_id"},
		}, nil

	case "user":
		return &rawCondition{
			SQL:   fmt.Sprintf("issue_assignees.user_id %s (%s)", op, placeholderList),
			Args:  args,
			Joins: []string{"JOIN issue_assignees ON issue_assignees.issue_id = issues.id"},
		}, nil

	case "label":
		return &rawCondition{
			SQL:   fmt.Sprintf("issue_labels.label_id %s (%s)", op, placeholderList),
			Args:  args,
			Joins: []string{"JOIN issue_labels ON issue_labels.issue_id = issues.id"},
		}, nil

	case "cycle":
		return &rawCondition{
			SQL:   fmt.Sprintf("issue_cycles.cycle_id %s (%s)", op, placeholderList),
			Args:  args,
			Joins: []string{"JOIN issue_cycles ON issue_cycles.issue_id = issues.id"},
		}, nil

	case "module":
		return &rawCondition{
			SQL:   fmt.Sprintf("module_issues.module_id %s (%s)", op, placeholderList),
			Args:  args,
			Joins: []string{"JOIN module_issues ON module_issues.issue_id = issues.id"},
		}, nil
	}

	sql := fmt.Sprintf("%s %s (%s)", mapping.ColumnName, op, placeholderList)
	return &rawCondition{SQL: sql, Args: args}, nil
}

func (e *GORMExecutor) buildCustomFieldIn(fieldName, operator string, values []interface{}) (*rawCondition, error) {
	if len(values) == 0 {
		if operator == "NOT IN" {
			return &rawCondition{SQL: "1 = 1"}, nil
		}
		return &rawCondition{SQL: "1 = 0"}, nil
	}

	fieldID := strings.TrimPrefix(fieldName, "cf_")
	join := fmt.Sprintf("JOIN issue_custom_field_values icfv ON icfv.issue_id = issues.id AND icfv.field_id = %s", fieldID)

	placeholders := make([]string, len(values))
	args := make([]interface{}, len(values))
	for i, v := range values {
		placeholders[i] = "?"
		args[i] = v
	}

	op := "IN"
	if operator == "NOT IN" {
		op = "NOT IN"
	}
	placeholderList := strings.Join(placeholders, ", ")

	return &rawCondition{
		SQL:   fmt.Sprintf("icfv.value %s (%s)", op, placeholderList),
		Args:  args,
		Joins: []string{join},
	}, nil
}

func (e *GORMExecutor) buildNotRaw(expr *NotExpr, ctx *QueryContext) (*rawCondition, error) {
	sub, err := e.buildRawCondition(expr.Expr, ctx)
	if err != nil {
		return nil, err
	}
	return &rawCondition{
		SQL:   "NOT (" + sub.SQL + ")",
		Args:  sub.Args,
		Joins: sub.Joins,
	}, nil
}

func (e *GORMExecutor) buildNullRaw(expr *NullCheck, ctx *QueryContext) (*rawCondition, error) {
	mapping, ok := ctx.FieldMap[expr.Field]
	if !ok {
		return &rawCondition{SQL: "1 = 1"}, nil
	}

	isNull := expr.Operator == "IS NULL"

	switch mapping.FieldType {
	case "user":
		if isNull {
			return &rawCondition{SQL: "NOT EXISTS (SELECT 1 FROM issue_assignees WHERE issue_assignees.issue_id = issues.id)"}, nil
		}
		return &rawCondition{SQL: "EXISTS (SELECT 1 FROM issue_assignees WHERE issue_assignees.issue_id = issues.id)"}, nil

	case "label":
		if isNull {
			return &rawCondition{SQL: "NOT EXISTS (SELECT 1 FROM issue_labels WHERE issue_labels.issue_id = issues.id)"}, nil
		}
		return &rawCondition{SQL: "EXISTS (SELECT 1 FROM issue_labels WHERE issue_labels.issue_id = issues.id)"}, nil

	case "cycle":
		if isNull {
			return &rawCondition{SQL: "NOT EXISTS (SELECT 1 FROM issue_cycles WHERE issue_cycles.issue_id = issues.id)"}, nil
		}
		return &rawCondition{SQL: "EXISTS (SELECT 1 FROM issue_cycles WHERE issue_cycles.issue_id = issues.id)"}, nil

	case "module":
		if isNull {
			return &rawCondition{SQL: "NOT EXISTS (SELECT 1 FROM module_issues WHERE module_issues.issue_id = issues.id)"}, nil
		}
		return &rawCondition{SQL: "EXISTS (SELECT 1 FROM module_issues WHERE module_issues.issue_id = issues.id)"}, nil

	default:
		if isNull {
			return &rawCondition{SQL: fmt.Sprintf("%s IS NULL", mapping.ColumnName)}, nil
		}
		return &rawCondition{SQL: fmt.Sprintf("%s IS NOT NULL", mapping.ColumnName)}, nil
	}
}
