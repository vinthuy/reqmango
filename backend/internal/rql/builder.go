// backend/internal/rql/builder.go

package rql

import (
	"fmt"
	"strings"
)

type QueryBuilder struct {
	conditions []string
	args      []interface{}
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		conditions: make([]string, 0),
		args:       make([]interface{}, 0),
	}
}

func (b *QueryBuilder) Build(node Node) (string, []interface{}, error) {
	if node == nil {
		return "", nil, nil
	}

	cond, args, err := b.buildNode(node)
	if err != nil {
		return "", nil, err
	}

	return cond, args, nil
}

func (b *QueryBuilder) buildNode(node Node) (string, []interface{}, error) {
	switch n := node.(type) {
	case *BinaryExpr:
		return b.buildBinaryExpr(n)
	case *Comparison:
		return b.buildComparison(n)
	case *LikeExpr:
		return b.buildLikeExpr(n)
	case *InExpr:
		return b.buildInExpr(n)
	case *NotExpr:
		return b.buildNotExpr(n)
	default:
		return "", nil, fmt.Errorf("unknown node type: %T", node)
	}
}

func (b *QueryBuilder) buildBinaryExpr(expr *BinaryExpr) (string, []interface{}, error) {
	leftCond, leftArgs, err := b.buildNode(expr.Left)
	if err != nil {
		return "", nil, err
	}

	rightCond, rightArgs, err := b.buildNode(expr.Right)
	if err != nil {
		return "", nil, err
	}

	operator := " AND "
	if expr.Operator == "OR" {
		operator = " OR "
	}

	return fmt.Sprintf("(%s%s%s)", leftCond, operator, rightCond), append(leftArgs, rightArgs...), nil
}

func (b *QueryBuilder) buildComparison(expr *Comparison) (string, []interface{}, error) {
	field := b.mapFieldName(expr.Field)
	operator := b.mapOperator(expr.Operator)

	if operator == "LIKE" || operator == "ILIKE" {
		return fmt.Sprintf("%s %s ?", field, operator), append(b.args, "%"+expr.Value.(string)+"%"), nil
	}

	return fmt.Sprintf("%s %s ?", field, operator), append(b.args, expr.Value), nil
}

func (b *QueryBuilder) buildLikeExpr(expr *LikeExpr) (string, []interface{}, error) {
	field := b.mapFieldName(expr.Field)
	return fmt.Sprintf("%s LIKE ?", field), append(b.args, "%"+expr.Value+"%"), nil
}

func (b *QueryBuilder) buildInExpr(expr *InExpr) (string, []interface{}, error) {
	field := b.mapFieldName(expr.Field)

	if len(expr.Values) == 0 {
		return "1=0", nil, nil
	}

	placeholders := make([]string, len(expr.Values))
	args := make([]interface{}, len(expr.Values))

	for i, v := range expr.Values {
		placeholders[i] = "?"
		args[i] = v
	}

	return fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", ")), args, nil
}

func (b *QueryBuilder) buildNotExpr(expr *NotExpr) (string, []interface{}, error) {
	cond, args, err := b.buildNode(expr.Expr)
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("NOT (%s)", cond), args, nil
}

func (b *QueryBuilder) mapFieldName(field string) string {
	mapping := map[string]string{
		"id":           "id",
		"sequence_id":  "sequence_id",
		"name":         "name",
		"description":  "description",
		"state":        "state_id",
		"priority":      "priority",
		"assignee":     "assignee_id",
		"reporter":     "reporter_id",
		"label":        "label_id",
		"cycle":        "cycle_id",
		"module":       "module_id",
		"created_at":   "created_at",
		"updated_at":   "updated_at",
		"due_date":     "due_date",
		"start_date":   "start_date",
		"end_date":     "end_date",
	}

	if mapped, ok := mapping[field]; ok {
		return mapped
	}

	return field
}

func (b *QueryBuilder) mapOperator(op string) string {
	mapping := map[string]string{
		"=":    "=",
		"!=":   "!=",
		">":    ">",
		"<":    "<",
		">=":   ">=",
		"<=":   "<=",
		"LIKE": "LIKE",
	}

	if mapped, ok := mapping[op]; ok {
		return mapped
	}

	return op
}

func (b *QueryBuilder) Reset() {
	b.conditions = b.conditions[:0]
	b.args = b.args[:0]
}