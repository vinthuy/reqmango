// backend/internal/rql/ast.go

package rql

// TokenType 词法标记类型
type TokenType int

const (
	TOKEN_ILLEGAL TokenType = iota
	TOKEN_EOF
	TOKEN_IDENTIFIER
	TOKEN_STRING
	TOKEN_NUMBER
	TOKEN_DATE
	TOKEN_OPERATOR
	TOKEN_LIKE
	TOKEN_IN
	TOKEN_AND
	TOKEN_OR
	TOKEN_NOT
	TOKEN_IS
	TOKEN_NULL
	TOKEN_LPAREN
	TOKEN_RPAREN
	TOKEN_COMMA
)

// Token 词法标记
type Token struct {
	Type     TokenType
	Value    string
	Position int
}

// AST 节点接口
type Node interface {
	nodeType() string
}

// BinaryExpr 二元表达式 (AND/OR)
type BinaryExpr struct {
	Left     Node
	Operator string // "AND" | "OR"
	Right    Node
}

func (b *BinaryExpr) nodeType() string { return "BinaryExpr" }

// Comparison 比较表达式
type Comparison struct {
	Field    string
	Operator string // "=", "!=", ">", "<", ">=", "<="
	Value    interface{}
}

func (c *Comparison) nodeType() string { return "Comparison" }

// LikeExpr 模糊匹配表达式
type LikeExpr struct {
	Field    string
	Value    string
	Operator string // "LIKE" | "NOT LIKE"
}

func (l *LikeExpr) nodeType() string { return "LikeExpr" }

// InExpr IN 表达式
type InExpr struct {
	Field    string
	Values   []interface{}
	Operator string // "IN" | "NOT IN"
}

func (i *InExpr) nodeType() string { return "InExpr" }

// NullCheck IS NULL / IS NOT NULL 表达式
type NullCheck struct {
	Field    string
	Operator string // "IS NULL" | "IS NOT NULL"
}

func (n *NullCheck) nodeType() string { return "NullCheck" }

// NotExpr NOT 表达式
type NotExpr struct {
	Expr Node
}

func (n *NotExpr) nodeType() string { return "NotExpr" }

// ParseError 解析错误
type ParseError struct {
	Position int
	Message  string
}

func (e *ParseError) Error() string {
	return e.Message
}

// RQLError RQL 错误
type RQLError struct {
	Code    string
	Message string
}

func (e *RQLError) Error() string {
	return e.Message
}
