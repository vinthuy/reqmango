// backend/internal/rql/executor.go

package rql

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/reqmango/backend/internal/model"
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
	DB        *gorm.DB
	ProjectID uint64
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
func NewIssueQueryContext(db *gorm.DB, projectID uint64) *QueryContext {
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
		"type":          {ColumnName: "issues.issue_type_id", FieldType: "number", JoinTable: "issue_types", JoinKey: "name"},
		"assignee":      {ColumnName: "issues.id", FieldType: "user"},
		"assignee_id":   {ColumnName: "assignee_id", FieldType: "user", JoinTable: "issue_assignees"},
		"reporter":      {ColumnName: "issues.reporter_id", FieldType: "number"},
		"label":         {ColumnName: "issues.id", FieldType: "label"},
		"cycle":         {ColumnName: "issues.id", FieldType: "cycle"},
		"cycle_id":      {ColumnName: "cycle_id", FieldType: "cycle", JoinTable: "issue_cycles"},
		"module":        {ColumnName: "issues.id", FieldType: "module"},
		"module_id":     {ColumnName: "module_id", FieldType: "module", JoinTable: "module_issues"},
		"issue_type_id": {ColumnName: "issues.issue_type_id", FieldType: "number"},
			"project_id":    {ColumnName: "issues.project_id", FieldType: "number"},
			"workspace_id":  {ColumnName: "issues.workspace_id", FieldType: "number"},
			"created_at":    {ColumnName: "issues.created_at", FieldType: "date"},
			"updated_at":    {ColumnName: "issues.updated_at", FieldType: "date"},
			"start_date":    {ColumnName: "issues.start_date", FieldType: "date"},
			"target_date":   {ColumnName: "issues.target_date", FieldType: "date"},
		},
		DB:        db,
		ProjectID: projectID,
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

// normalizeCustomFieldPrefix converts "custom_field:N" to "cf_N" for uniform handling.
func normalizeCustomFieldPrefix(field string) string {
	if strings.HasPrefix(field, "custom_field:") {
		return "cf_" + strings.TrimPrefix(field, "custom_field:")
	}
	return field
}

// buildComparisonRaw 将 AST 节点转换为原始 SQL 条件
func (e *GORMExecutor) buildComparisonRaw(expr *Comparison, ctx *QueryContext) (*rawCondition, error) {
	field := normalizeCustomFieldPrefix(expr.Field)
	mapping, ok := ctx.FieldMap[field]
	if !ok {
		if strings.HasPrefix(field, "cf_") {
			return e.buildCustomFieldComparison(field, expr.Operator, expr.Value, ctx)
		}
		return &rawCondition{SQL: "1 = 1"}, nil
	}

	switch expr.Operator {
	case "=":
		return e.buildEqualRaw(expr.Value, mapping)
	case "!=":
		return e.buildNotEqualRaw(expr.Value, mapping)
	case ">":
		return e.buildComparisonOpRaw(expr.Value, mapping, ">")
	case "<":
		return e.buildComparisonOpRaw(expr.Value, mapping, "<")
	case ">=":
		return e.buildComparisonOpRaw(expr.Value, mapping, ">=")
	case "<=":
		return e.buildComparisonOpRaw(expr.Value, mapping, "<=")
	default:
		return nil, fmt.Errorf("unsupported operator: %s", expr.Operator)
	}
}

func (e *GORMExecutor) buildCustomFieldComparison(fieldName, operator, value interface{}, ctx *QueryContext) (*rawCondition, error) {
	fieldNameStr, ok := fieldName.(string)
	if !ok {
		return nil, fmt.Errorf("custom field name must be a string")
	}
	fieldIDStr := strings.TrimPrefix(fieldNameStr, "cf_")

	var fieldID uint64
	var fieldType string
	if _, err := strconv.ParseUint(fieldIDStr, 10, 64); err == nil {
		fieldID, _ = strconv.ParseUint(fieldIDStr, 10, 64)
		if ctx.DB != nil {
			var f model.CustomField
			ctx.DB.Select("field_type").First(&f, fieldID)
			fieldType = f.FieldType
		}
	} else {
		if ctx.DB == nil {
			return nil, fmt.Errorf("cannot resolve custom field name without DB context")
		}
		var projectID uint64
		if ctx.ProjectID > 0 {
			projectID = ctx.ProjectID
		}
		var fields []model.CustomField
		err := ctx.DB.Where("name = ? AND is_active = ?", fieldIDStr, true).
			Where("project_id = ? OR project_id IS NULL", projectID).
			Order(fmt.Sprintf("CASE WHEN project_id = %d THEN 0 ELSE 1 END", projectID)).
			Find(&fields).Error
		if err != nil {
			return nil, fmt.Errorf("custom field not found: %s", fieldIDStr)
		}
		if len(fields) == 0 {
			return nil, fmt.Errorf("custom field not found: %s", fieldIDStr)
		}
		fieldID = fields[0].ID
		fieldType = fields[0].FieldType
	}

	alias := fmt.Sprintf("icfv_%d", fieldID)
	join := fmt.Sprintf("JOIN issue_custom_field_values %s ON %s.issue_id = issues.id AND %s.field_id = %d", alias, alias, alias, fieldID)

	// For number fields, cast to numeric for proper numeric comparison
	valueExpr := fmt.Sprintf("%s.value", alias)
	if fieldType == "number" {
		valueExpr = fmt.Sprintf("CAST(%s.value AS NUMERIC)", alias)
	}

	var sql string
	switch operator {
	case "=":
		sql = valueExpr + " = ?"
	case "!=":
		sql = valueExpr + " != ?"
	case ">":
		sql = valueExpr + " > ?"
	case "<":
		sql = valueExpr + " < ?"
	case ">=":
		sql = valueExpr + " >= ?"
	case "<=":
		sql = valueExpr + " <= ?"
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
	case "string", "date":
		return &rawCondition{SQL: fmt.Sprintf("%s = ?", mapping.ColumnName), Args: []interface{}{value}}, nil

	case "number":
		// If JoinTable has JoinKey "name", lookup by name via subquery (e.g. state → states, type → issue_types)
		if mapping.JoinTable != "" && mapping.JoinKey == "name" {
			return &rawCondition{
				SQL:  fmt.Sprintf("%s IN (SELECT id FROM %s WHERE name = ?)", mapping.ColumnName, mapping.JoinTable),
				Args: []interface{}{value},
			}, nil
		}
		return &rawCondition{SQL: fmt.Sprintf("%s = ?", mapping.ColumnName), Args: []interface{}{value}}, nil

	case "state_group":
		return &rawCondition{
			SQL:   "states.group = ?",
			Args:  []interface{}{value},
			Joins: []string{"JOIN states ON states.id = issues.state_id"},
		}, nil

	case "user":
		// Look up assignee by display name through issue_assignees and users tables
		return &rawCondition{
			SQL:  "issues.id IN (SELECT ia.issue_id FROM issue_assignees ia JOIN users u ON ia.user_id = u.id WHERE COALESCE(NULLIF(TRIM(u.display_name), ''), u.username) = ?)",
			Args: []interface{}{value},
		}, nil

	case "label":
		// Look up label by name through issue_labels and labels tables
		return &rawCondition{
			SQL:  "issues.id IN (SELECT il.issue_id FROM issue_labels il JOIN labels l ON il.label_id = l.id WHERE l.name = ?)",
			Args: []interface{}{value},
		}, nil

	case "cycle":
		// Look up cycle by name through issue_cycles and cycles tables
		return &rawCondition{
			SQL:  "issues.id IN (SELECT ic.issue_id FROM issue_cycles ic JOIN cycles c ON ic.cycle_id = c.id WHERE c.name = ?)",
			Args: []interface{}{value},
		}, nil

	case "module":
		// Look up module by name through module_issues and modules tables
		return &rawCondition{
			SQL:  "issues.id IN (SELECT mi.issue_id FROM module_issues mi JOIN modules m ON mi.module_id = m.id WHERE m.name = ?)",
			Args: []interface{}{value},
		}, nil

	default:
		return &rawCondition{SQL: fmt.Sprintf("%s = ?", mapping.ColumnName), Args: []interface{}{value}}, nil
	}
}

func (e *GORMExecutor) buildNotEqualRaw(value interface{}, mapping FieldMapping) (*rawCondition, error) {
	// Handle special field types that require JOINs (consistent with buildEqualRaw)
	switch mapping.FieldType {
	case "number":
		// If JoinTable has JoinKey "name", match by name via subquery (e.g. state → states, type → issue_types)
		if mapping.JoinTable != "" && mapping.JoinKey == "name" {
			return &rawCondition{
				SQL:  fmt.Sprintf("%s NOT IN (SELECT id FROM %s WHERE name = ?)", mapping.ColumnName, mapping.JoinTable),
				Args: []interface{}{value},
			}, nil
		}
		return &rawCondition{SQL: fmt.Sprintf("%s != ?", mapping.ColumnName), Args: []interface{}{value}}, nil
	case "state_group":
		return &rawCondition{
			SQL:   "states.group != ?",
			Args:  []interface{}{value},
			Joins: []string{"JOIN states ON states.id = issues.state_id"},
		}, nil
	case "user":
		return &rawCondition{
			SQL:  "issues.id NOT IN (SELECT ia.issue_id FROM issue_assignees ia JOIN users u ON ia.user_id = u.id WHERE COALESCE(NULLIF(TRIM(u.display_name), ''), u.username) = ?)",
			Args: []interface{}{value},
		}, nil
	case "label":
		return &rawCondition{
			SQL:  "issues.id NOT IN (SELECT il.issue_id FROM issue_labels il JOIN labels l ON il.label_id = l.id WHERE l.name = ?)",
			Args: []interface{}{value},
		}, nil
	case "cycle":
		return &rawCondition{
			SQL:  "issues.id NOT IN (SELECT ic.issue_id FROM issue_cycles ic JOIN cycles c ON ic.cycle_id = c.id WHERE c.name = ?)",
			Args: []interface{}{value},
		}, nil
	case "module":
		return &rawCondition{
			SQL:  "issues.id NOT IN (SELECT mi.issue_id FROM module_issues mi JOIN modules m ON mi.module_id = m.id WHERE m.name = ?)",
			Args: []interface{}{value},
		}, nil
	}
	return &rawCondition{SQL: fmt.Sprintf("%s != ?", mapping.ColumnName), Args: []interface{}{value}}, nil
}

func (e *GORMExecutor) buildComparisonOpRaw(value interface{}, mapping FieldMapping, op string) (*rawCondition, error) {
	// For state field with join table, use subquery for comparison ops
	// Note: comparison ops on state names are lexicographic and rarely useful
	return &rawCondition{SQL: fmt.Sprintf("%s %s ?", mapping.ColumnName, op), Args: []interface{}{value}}, nil
}

func (e *GORMExecutor) buildLikeRaw(expr *LikeExpr, ctx *QueryContext) (*rawCondition, error) {
	field := normalizeCustomFieldPrefix(expr.Field)
	mapping, ok := ctx.FieldMap[field]
	if !ok {
		if strings.HasPrefix(field, "cf_") {
			return e.buildCustomFieldLike(field, expr.Operator, expr.Value, ctx)
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

	if field == "name" || field == "description" {
		return e.buildFullTextSearch(field, expr.Operator, expr.Value)
	}

	return &rawCondition{
		SQL:  fmt.Sprintf("%s %s ?", mapping.ColumnName, op),
		Args: []interface{}{value},
	}, nil
}

func (e *GORMExecutor) buildFullTextSearch(field, operator, value string) (*rawCondition, error) {
	// Use ILIKE for name/description search.
	// PostgreSQL full-text search (to_tsvector/plainto_tsquery) does not support
	// substring matching for Chinese/CJK text because those languages lack whitespace
	// tokenization, causing the entire string to be treated as a single token.
	// Search both name and description_stripped columns regardless of which field
	// was specified (same behavior as the original full-text search approach).
	op := "ILIKE"
	if operator == "NOT LIKE" {
		op = "NOT ILIKE"
	}

	// Ensure value has wildcards for substring matching
	pattern := value
	if !strings.ContainsAny(pattern, "%_") {
		pattern = "%" + pattern + "%"
	}

	return &rawCondition{
		SQL:  fmt.Sprintf("(COALESCE(name, '') %s ? OR COALESCE(description_stripped, '') %s ? OR issues.sequence_id::text %s ?)", op, op, op),
		Args: []interface{}{pattern, pattern, pattern},
	}, nil
}

func (e *GORMExecutor) buildCustomFieldLike(fieldName, operator, value string, ctx *QueryContext) (*rawCondition, error) {
	fieldIDStr := strings.TrimPrefix(fieldName, "cf_")

	var fieldID uint64
	if _, err := strconv.ParseUint(fieldIDStr, 10, 64); err == nil {
		fieldID, _ = strconv.ParseUint(fieldIDStr, 10, 64)
	} else {
		if ctx.DB == nil {
			return nil, fmt.Errorf("cannot resolve custom field name without DB context")
		}
		var projectID uint64
		if ctx.ProjectID > 0 {
			projectID = ctx.ProjectID
		}
		var fields []model.CustomField
		err := ctx.DB.Where("name = ? AND is_active = ?", fieldIDStr, true).
			Where("project_id = ? OR project_id IS NULL", projectID).
			Order(fmt.Sprintf("CASE WHEN project_id = %d THEN 0 ELSE 1 END", projectID)).
			Find(&fields).Error
		if err != nil {
			return nil, fmt.Errorf("custom field not found: %s", fieldIDStr)
		}
		if len(fields) == 0 {
			return nil, fmt.Errorf("custom field not found: %s", fieldIDStr)
		}
		fieldID = fields[0].ID
	}

	alias := fmt.Sprintf("icfv_%d", fieldID)
	join := fmt.Sprintf("JOIN issue_custom_field_values %s ON %s.issue_id = issues.id AND %s.field_id = %d", alias, alias, alias, fieldID)

	op := "ILIKE"
	if operator == "NOT LIKE" {
		op = "NOT ILIKE"
	}

	if !strings.ContainsAny(value, "%_") {
		value = fmt.Sprintf("%%%s%%", value)
	}

	return &rawCondition{
		SQL:   fmt.Sprintf("%s.value %s ?", alias, op),
		Args:  []interface{}{value},
		Joins: []string{join},
	}, nil
}

func (e *GORMExecutor) buildInRaw(expr *InExpr, ctx *QueryContext) (*rawCondition, error) {
	field := normalizeCustomFieldPrefix(expr.Field)
	mapping, ok := ctx.FieldMap[field]
	if !ok {
		if strings.HasPrefix(field, "cf_") {
			return e.buildCustomFieldIn(field, expr.Operator, expr.Values, ctx)
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

	// Handle special field types that require name-based lookups (consistent with buildEqualRaw)
	switch mapping.FieldType {
	case "number":
		// If JoinTable has JoinKey "name", use subquery to resolve names to IDs (e.g. state → states, type → issue_types)
		if mapping.JoinTable != "" && mapping.JoinKey == "name" {
			return &rawCondition{
				SQL:  fmt.Sprintf("%s %s (SELECT id FROM %s WHERE name IN (%s))", mapping.ColumnName, op, mapping.JoinTable, placeholderList),
				Args: args,
			}, nil
		}
	case "state_group":
		return &rawCondition{
			SQL:   fmt.Sprintf("states.group %s (%s)", op, placeholderList),
			Args:  args,
			Joins: []string{"JOIN states ON states.id = issues.state_id"},
		}, nil

	case "user":
		return &rawCondition{
			SQL:  fmt.Sprintf("issues.id %s (SELECT ia.issue_id FROM issue_assignees ia JOIN users u ON ia.user_id = u.id WHERE COALESCE(NULLIF(TRIM(u.display_name), ''), u.username) IN (%s))", op, placeholderList),
			Args: args,
		}, nil

	case "label":
		return &rawCondition{
			SQL:  fmt.Sprintf("issues.id %s (SELECT il.issue_id FROM issue_labels il JOIN labels l ON il.label_id = l.id WHERE l.name IN (%s))", op, placeholderList),
			Args: args,
		}, nil

	case "cycle":
		return &rawCondition{
			SQL:  fmt.Sprintf("issues.id %s (SELECT ic.issue_id FROM issue_cycles ic JOIN cycles c ON ic.cycle_id = c.id WHERE c.name IN (%s))", op, placeholderList),
			Args: args,
		}, nil

	case "module":
		return &rawCondition{
			SQL:  fmt.Sprintf("issues.id %s (SELECT mi.issue_id FROM module_issues mi JOIN modules m ON mi.module_id = m.id WHERE m.name IN (%s))", op, placeholderList),
			Args: args,
		}, nil
	}

	sql := fmt.Sprintf("%s %s (%s)", mapping.ColumnName, op, placeholderList)
	return &rawCondition{SQL: sql, Args: args}, nil
}

func (e *GORMExecutor) buildCustomFieldIn(fieldName, operator string, values []interface{}, ctx *QueryContext) (*rawCondition, error) {
	if len(values) == 0 {
		if operator == "NOT IN" {
			return &rawCondition{SQL: "1 = 1"}, nil
		}
		return &rawCondition{SQL: "1 = 0"}, nil
	}

	fieldIDStr := strings.TrimPrefix(fieldName, "cf_")

	var fieldID uint64
	if _, err := strconv.ParseUint(fieldIDStr, 10, 64); err == nil {
		fieldID, _ = strconv.ParseUint(fieldIDStr, 10, 64)
	} else {
		if ctx.DB == nil {
			return nil, fmt.Errorf("cannot resolve custom field name without DB context")
		}
		var projectID uint64
		if ctx.ProjectID > 0 {
			projectID = ctx.ProjectID
		}
		var fields []model.CustomField
		err := ctx.DB.Where("name = ? AND is_active = ?", fieldIDStr, true).
			Where("project_id = ? OR project_id IS NULL", projectID).
			Order(fmt.Sprintf("CASE WHEN project_id = %d THEN 0 ELSE 1 END", projectID)).
			Find(&fields).Error
		if err != nil {
			return nil, fmt.Errorf("custom field not found: %s", fieldIDStr)
		}
		if len(fields) == 0 {
			return nil, fmt.Errorf("custom field not found: %s", fieldIDStr)
		}
		fieldID = fields[0].ID
	}

	alias := fmt.Sprintf("icfv_%d", fieldID)
	join := fmt.Sprintf("JOIN issue_custom_field_values %s ON %s.issue_id = issues.id AND %s.field_id = %d", alias, alias, alias, fieldID)

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
		SQL:   fmt.Sprintf("%s.value %s (%s)", alias, op, placeholderList),
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
	field := normalizeCustomFieldPrefix(expr.Field)
	mapping, ok := ctx.FieldMap[field]
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
