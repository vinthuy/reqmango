# RQL 实施计划 (后端解析方案)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 Reqman 查询语言 (RQL)，前端提交 RQL 字符串，后端解析并执行查询

**Architecture:**
- 前端：Vue 3，只负责输入 RQL 并调用 API
- 后端：Go，实现 RQL Lexer、Parser、Query Builder，解析 RQL 并执行数据库查询

**Tech Stack:** Go (Gin), Vue 3, TypeScript, GORM

---

## 文件结构

### 后端 (Go)

```
backend-go/internal/
├── rql/
│   ├── lexer.go       # 词法分析器
│   ├── parser.go      # 语法分析器
│   ├── builder.go     # SQL Query Builder
│   ├── ast.go         # AST 类型定义
│   ├── errors.go      # 错误定义
│   └── handler.go     # HTTP Handler
├── dto/
│   └── request/
│       └── rql.go     # RQL 请求/响应 DTO
├── router/
│   └── router.go      # 添加 RQL 路由
```

### 前端 (Vue)

```
frontend/src/
├── utils/rql/
│   ├── types.ts       # RQL 相关类型（与后端一致）
│   └── index.ts       # 统一导出
├── api/
│   └── rql.ts         # RQL API 调用
├── components/RQL/
│   ├── RQLInput.vue   # RQL 输入框
│   ├── RQLHistory.vue # 查询历史
│   └── index.ts
└── composables/
    └── useRQL.ts
```

---

## API 设计

### POST /api/rql/search

**Request:**
```json
{
  "entity": "issue",           // issue | cycle | module
  "project_id": 1,
  "rql": "state = \"待处理\" AND priority = \"high\""
}
```

**Response (Success):**
```json
{
  "success": true,
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

**Response (Error):**
```json
{
  "success": false,
  "error": {
    "code": "RQL_PARSE_ERROR",
    "message": "语法错误 (位置 6): 期望操作符"
  }
}
```

---

## 实施任务

### Task 1: 后端 - AST 类型定义

**Files:**
- Create: `backend-go/internal/rql/ast.go`

- [ ] **Step 1: 创建 AST 类型定义**

```go
// backend-go/internal/rql/ast.go

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
	Field string
	Value string
}

func (l *LikeExpr) nodeType() string { return "LikeExpr" }

// InExpr IN 表达式
type InExpr struct {
	Field  string
	Values []interface{}
}

func (i *InExpr) nodeType() string { return "InExpr" }

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
```

- [ ] **Step 2: 提交**

```bash
git add backend-go/internal/rql/ast.go
git commit -m "feat(rql): add RQL AST types"
```

---

### Task 2: 后端 - 词法分析器

**Files:**
- Create: `backend-go/internal/rql/lexer.go`

- [ ] **Step 1: 创建词法分析器**

```go
// backend-go/internal/rql/lexer.go

package rql

import (
	"fmt"
	"strings"
	"unicode"
)

type Lexer struct {
	input    string
	position int
	errors   []ParseError
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:    strings.TrimSpace(input),
		position: 0,
		errors:   make([]ParseError, 0),
	}
}

func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token

	for l.position < len(l.input) {
		l.skipWhitespace()

		if l.position >= len(l.input) {
			break
		}

		char := rune(l.input[l.position])

		// 字符串
		if char == '"' || char == '\'' {
			token := l.readString(char)
			tokens = append(tokens, token)
			continue
		}

		// 括号
		if char == '(' {
			tokens = append(tokens, Token{Type: TOKEN_LPAREN, Value: "(", Position: l.position})
			l.position++
			continue
		}

		if char == ')' {
			tokens = append(tokens, Token{Type: TOKEN_RPAREN, Value: ")", Position: l.position})
			l.position++
			continue
		}

		// 逗号
		if char == ',' {
			tokens = append(tokens, Token{Type: TOKEN_COMMA, Value: ",", Position: l.position})
			l.position++
			continue
		}

		// 操作符
		if char == '=' || char == '!' || char == '>' || char == '<' {
			token := l.readOperator()
			tokens = append(tokens, token)
			continue
		}

		// 标识符
		if unicode.IsLetter(char) || char == '_' {
			token := l.readIdentifier()
			tokens = append(tokens, token)
			continue
		}

		// 数字/日期
		if unicode.IsDigit(char) {
			token := l.readNumberOrDate()
			tokens = append(tokens, token)
			continue
		}

		// 未知字符
		l.errors = append(l.errors, ParseError{
			Position: l.position,
			Message:  fmt.Sprintf("未知字符: %c", char),
		})
		l.position++
	}

	tokens = append(tokens, Token{Type: TOKEN_EOF, Value: "", Position: l.position})

	if len(l.errors) > 0 {
		return tokens, l.errors[0]
	}

	return tokens, nil
}

func (l *Lexer) skipWhitespace() {
	for l.position < len(l.input) && unicode.IsSpace(rune(l.input[l.position])) {
		l.position++
	}
}

func (l *Lexer) readString(quote rune) Token {
	start := l.position
	l.position++ // skip opening quote

	var value strings.Builder
	for l.position < len(l.input) {
		char := rune(l.input[l.position])
		if char == quote {
			l.position++ // skip closing quote
			return Token{Type: TOKEN_STRING, Value: value.String(), Position: start}
		}
		value.WriteRune(char)
		l.position++
	}

	l.errors = append(l.errors, ParseError{
		Position: start,
		Message:  "未终止的字符串",
	})

	return Token{Type: TOKEN_STRING, Value: value.String(), Position: start}
}

func (l *Lexer) readOperator() Token {
	start := l.position
	char := rune(l.input[l.position])
	l.position++

	// Two-character operators: >=, <=, !=
	if l.position < len(l.input) && rune(l.input[l.position]) == '=' {
		l.position++
		switch char {
		case '>':
			return Token{Type: TOKEN_OPERATOR, Value: ">=", Position: start}
		case '<':
			return Token{Type: TOKEN_OPERATOR, Value: "<=", Position: start}
		case '!':
			return Token{Type: TOKEN_OPERATOR, Value: "!=", Position: start}
		}
	}

	// Single-character operators
	switch char {
	case '=':
		return Token{Type: TOKEN_OPERATOR, Value: "=", Position: start}
	case '>':
		return Token{Type: TOKEN_OPERATOR, Value: ">", Position: start}
	case '<':
		return Token{Type: TOKEN_OPERATOR, Value: "<", Position: start}
	}

	return Token{Type: TOKEN_ILLEGAL, Value: string(char), Position: start}
}

func (l *Lexer) readIdentifier() Token {
	start := l.position
	var value strings.Builder

	for l.position < len(l.input) {
		char := rune(l.input[l.position])
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			break
		}
		value.WriteRune(char)
		l.position++
	}

	upper := strings.ToUpper(value.String())

	// Keywords
	switch upper {
	case "AND":
		return Token{Type: TOKEN_AND, Value: upper, Position: start}
	case "OR":
		return Token{Type: TOKEN_OR, Value: upper, Position: start}
	case "NOT":
		return Token{Type: TOKEN_NOT, Value: upper, Position: start}
	case "LIKE":
		return Token{Type: TOKEN_LIKE, Value: upper, Position: start}
	case "IN":
		return Token{Type: TOKEN_IN, Value: upper, Position: start}
	}

	return Token{Type: TOKEN_IDENTIFIER, Value: value.String(), Position: start}
}

func (l *Lexer) readNumberOrDate() Token {
	start := l.position
	var value strings.Builder

	for l.position < len(l.input) && (unicode.IsDigit(rune(l.input[l.position])) || l.input[l.position] == '-' || l.input[l.position] == 'T' || l.input[l.position] == ':' || l.input[l.position] == '.') {
		value.WriteByte(l.input[l.position])
		l.position++
	}

	// Check if it's a date
	if strings.Contains(value.String(), "-") {
		return Token{Type: TOKEN_DATE, Value: value.String(), Position: start}
	}

	return Token{Type: TOKEN_NUMBER, Value: value.String(), Position: start}
}

func (l *Lexer) HasErrors() bool {
	return len(l.errors) > 0
}

func (l *Lexer) GetErrors() []ParseError {
	return l.errors
}
```

- [ ] **Step 2: 编写测试**

```go
// backend-go/internal/rql/lexer_test.go

package rql

import (
	"testing"
)

func TestLexer_SimpleComparison(t *testing.T) {
	lexer := NewLexer(`state = "待处理"`)
	tokens, err := lexer.Tokenize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tokens) != 4 {
		t.Fatalf("expected 4 tokens, got %d", len(tokens))
	}

	if tokens[0].Type != TOKEN_IDENTIFIER || tokens[0].Value != "state" {
		t.Errorf("expected IDENTIFIER 'state', got %v '%s'", tokens[0].Type, tokens[0].Value)
	}

	if tokens[1].Type != TOKEN_OPERATOR || tokens[1].Value != "=" {
		t.Errorf("expected OPERATOR '=', got %v '%s'", tokens[1].Type, tokens[1].Value)
	}

	if tokens[2].Type != TOKEN_STRING || tokens[2].Value != "待处理" {
		t.Errorf("expected STRING '待处理', got %v '%s'", tokens[2].Type, tokens[2].Value)
	}

	if tokens[3].Type != TOKEN_EOF {
		t.Errorf("expected EOF, got %v", tokens[3].Type)
	}
}

func TestLexer_AndOr(t *testing.T) {
	lexer := NewLexer(`state = "待处理" AND priority = "high"`)
	tokens, err := lexer.Tokenize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[3].Type != TOKEN_AND {
		t.Errorf("expected AND, got %v", tokens[3].Type)
	}
}

func TestLexer_Like(t *testing.T) {
	lexer := NewLexer(`name LIKE "登录"`)
	tokens, err := lexer.Tokenize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[1].Type != TOKEN_LIKE {
		t.Errorf("expected LIKE, got %v", tokens[1].Type)
	}
}

func TestLexer_IN(t *testing.T) {
	lexer := NewLexer(`state IN ("待处理", "进行中")`)
	tokens, err := lexer.Tokenize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[1].Type != TOKEN_IN {
		t.Errorf("expected IN, got %v", tokens[1].Type)
	}
}

func TestLexer_Parentheses(t *testing.T) {
	lexer := NewLexer(`(state = "待处理")`)
	tokens, err := lexer.Tokenize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens[0].Type != TOKEN_LPAREN {
		t.Errorf("expected LPAREN, got %v", tokens[0].Type)
	}

	if tokens[3].Type != TOKEN_RPAREN {
		t.Errorf("expected RPAREN, got %v", tokens[3].Type)
	}
}
```

- [ ] **Step 3: 运行测试**

```bash
cd backend-go && go test ./internal/rql/... -v
```

- [ ] **Step 4: 提交**

```bash
git add backend-go/internal/rql/lexer.go backend-go/internal/rql/lexer_test.go
git commit -m "feat(rql): implement RQL lexer"
```

---

### Task 3: 后端 - 语法分析器

**Files:**
- Create: `backend-go/internal/rql/parser.go`

- [ ] **Step 1: 创建语法分析器**

```go
// backend-go/internal/rql/parser.go

package rql

type Parser struct {
	tokens   []Token
	position int
	errors   []ParseError
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens:   tokens,
		position: 0,
		errors:   make([]ParseError, 0),
	}
}

func (p *Parser) Parse() (Node, error) {
	if len(p.tokens) == 0 || p.current().Type == TOKEN_EOF {
		return nil, nil
	}

	node := p.parseOrExpr()

	if p.HasErrors() {
		return node, p.errors[0]
	}

	return node, nil
}

func (p *Parser) current() Token {
	if p.position >= len(p.tokens) {
		return Token{Type: TOKEN_EOF}
	}
	return p.tokens[p.position]
}

func (p *Parser) advance() Token {
	token := p.current()
	p.position++
	return token
}

func (p *Parser) peek() Token {
	if p.position+1 >= len(p.tokens) {
		return Token{Type: TOKEN_EOF}
	}
	return p.tokens[p.position+1]
}

func (p *Parser) match(tokenType TokenType) bool {
	if p.current().Type == tokenType {
		p.position++
		return true
	}
	return false
}

func (p *Parser) expect(tokenType TokenType, message string) Token {
	if p.current().Type == tokenType {
		return p.advance()
	}

	p.errors = append(p.errors, ParseError{
		Position: p.current().Position,
		Message:  message,
	})

	return Token{Type: tokenType}
}

func (p *Parser) parseOrExpr() Node {
	left := p.parseAndExpr()

	for p.match(TOKEN_OR) {
		right := p.parseAndExpr()
		left = &BinaryExpr{
			Left:     left,
			Operator: "OR",
			Right:    right,
		}
	}

	return left
}

func (p *Parser) parseAndExpr() Node {
	left := p.parseNotExpr()

	for p.match(TOKEN_AND) {
		right := p.parseNotExpr()
		left = &BinaryExpr{
			Left:     left,
			Operator: "AND",
			Right:    right,
		}
	}

	return left
}

func (p *Parser) parseNotExpr() Node {
	if p.match(TOKEN_NOT) {
		expr := p.parsePrimary()
		return &NotExpr{Expr: expr}
	}

	return p.parsePrimary()
}

func (p *Parser) parsePrimary() Node {
	// Parenthesized expression
	if p.match(TOKEN_LPAREN) {
		expr := p.parseOrExpr()
		p.expect(TOKEN_RPAREN, "期望 ')'")
		return expr
	}

	return p.parseComparison()
}

func (p *Parser) parseComparison() Node {
	// Get field identifier
	fieldToken := p.expect(TOKEN_IDENTIFIER, "期望字段名")
	field := fieldToken.Value

	token := p.current()

	// LIKE expression
	if token.Type == TOKEN_LIKE {
		p.advance()
		valueToken := p.expect(TOKEN_STRING, "期望字符串值")
		return &LikeExpr{
			Field: field,
			Value: valueToken.Value,
		}
	}

	// IN expression
	if token.Type == TOKEN_IN {
		p.advance()
		p.expect(TOKEN_LPAREN, "期望 '('")

		var values []interface{}
		values = append(values, p.expect(TOKEN_STRING, "期望字符串值").Value)

		for p.match(TOKEN_COMMA) {
			values = append(values, p.expect(TOKEN_STRING, "期望字符串值").Value)
		}

		p.expect(TOKEN_RPAREN, "期望 ')'")

		return &InExpr{
			Field:  field,
			Values: values,
		}
	}

	// Regular comparison
	opToken := p.expect(TOKEN_OPERATOR, "期望操作符")
	operator := opToken.Value

	var value interface{}
	valToken := p.current()

	switch valToken.Type {
	case TOKEN_STRING:
		value = p.advance().Value
	case TOKEN_NUMBER:
		value = p.advance().Value
	case TOKEN_DATE:
		value = p.advance().Value
	case TOKEN_IDENTIFIER:
		value = p.advance().Value
	default:
		p.errors = append(p.errors, ParseError{
			Position: p.current().Position,
			Message:   "期望值",
		})
	}

	return &Comparison{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
}

func (p *Parser) HasErrors() bool {
	return len(p.errors) > 0
}

func (p *Parser) GetErrors() []ParseError {
	return p.errors
}
```

- [ ] **Step 2: 编写测试**

```go
// backend-go/internal/rql/parser_test.go

package rql

import (
	"testing"
)

func TestParser_SimpleComparison(t *testing.T) {
	lexer := NewLexer(`state = "待处理"`)
	tokens, _ := lexer.Tokenize()

	parser := NewParser(tokens)
	node, err := parser.Parse()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp, ok := node.(*Comparison)
	if !ok {
		t.Fatalf("expected *Comparison, got %T", node)
	}

	if comp.Field != "state" {
		t.Errorf("expected field 'state', got '%s'", comp.Field)
	}

	if comp.Operator != "=" {
		t.Errorf("expected operator '=', got '%s'", comp.Operator)
	}

	if comp.Value != "待处理" {
		t.Errorf("expected value '待处理', got '%v'", comp.Value)
	}
}

func TestParser_AndExpression(t *testing.T) {
	lexer := NewLexer(`state = "待处理" AND priority = "high"`)
	tokens, _ := lexer.Tokenize()

	parser := NewParser(tokens)
	node, err := parser.Parse()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bin, ok := node.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected *BinaryExpr, got %T", node)
	}

	if bin.Operator != "AND" {
		t.Errorf("expected AND, got '%s'", bin.Operator)
	}

	left, ok := bin.Left.(*Comparison)
	if !ok {
		t.Fatalf("expected left *Comparison, got %T", bin.Left)
	}

	if left.Field != "state" {
		t.Errorf("expected left field 'state', got '%s'", left.Field)
	}
}

func TestParser_OrExpression(t *testing.T) {
	lexer := NewLexer(`assignee = "张三" OR assignee = "李四"`)
	tokens, _ := lexer.Tokenize()

	parser := NewParser(tokens)
	node, err := parser.Parse()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bin, ok := node.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected *BinaryExpr, got %T", node)
	}

	if bin.Operator != "OR" {
		t.Errorf("expected OR, got '%s'", bin.Operator)
	}
}

func TestParser_LikeExpression(t *testing.T) {
	lexer := NewLexer(`name LIKE "登录"`)
	tokens, _ := lexer.Tokenize()

	parser := NewParser(tokens)
	node, err := parser.Parse()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	like, ok := node.(*LikeExpr)
	if !ok {
		t.Fatalf("expected *LikeExpr, got %T", node)
	}

	if like.Field != "name" {
		t.Errorf("expected field 'name', got '%s'", like.Field)
	}

	if like.Value != "登录" {
		t.Errorf("expected value '登录', got '%s'", like.Value)
	}
}

func TestParser_InExpression(t *testing.T) {
	lexer := NewLexer(`state IN ("待处理", "进行中")`)
	tokens, _ := lexer.Tokenize()

	parser := NewParser(tokens)
	node, err := parser.Parse()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	in, ok := node.(*InExpr)
	if !ok {
		t.Fatalf("expected *InExpr, got %T", node)
	}

	if in.Field != "state" {
		t.Errorf("expected field 'state', got '%s'", in.Field)
	}

	if len(in.Values) != 2 {
		t.Errorf("expected 2 values, got %d", len(in.Values))
	}
}

func TestParser_NestedParentheses(t *testing.T) {
	lexer := NewLexer(`(state = "待处理" OR state = "进行中") AND priority = "high"`)
	tokens, _ := lexer.Tokenize()

	parser := NewParser(tokens)
	node, err := parser.Parse()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bin, ok := node.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected *BinaryExpr, got %T", node)
	}

	if bin.Operator != "AND" {
		t.Errorf("expected AND, got '%s'", bin.Operator)
	}

	inner, ok := bin.Left.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected left *BinaryExpr, got %T", bin.Left)
	}

	if inner.Operator != "OR" {
		t.Errorf("expected inner OR, got '%s'", inner.Operator)
	}
}
```

- [ ] **Step 3: 运行测试**

```bash
cd backend-go && go test ./internal/rql/... -v
```

- [ ] **Step 4: 提交**

```bash
git add backend-go/internal/rql/parser.go backend-go/internal/rql/parser_test.go
git commit -m "feat(rql): implement RQL parser"
```

---

### Task 4: 后端 - SQL Query Builder

**Files:**
- Create: `backend-go/internal/rql/builder.go`

- [ ] **Step 1: 创建 Query Builder**

```go
// backend-go/internal/rql/builder.go

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
```

- [ ] **Step 2: 提交**

```bash
git add backend-go/internal/rql/builder.go
git commit -m "feat(rql): implement RQL query builder"
```

---

### Task 5: 后端 - HTTP Handler

**Files:**
- Create: `backend-go/internal/rql/handler.go`
- Create: `backend-go/internal/dto/request/rql.go`

- [ ] **Step 1: 创建 DTO**

```go
// backend-go/internal/dto/request/rql.go

package request

type RQLSearchRequest struct {
	Entity   string `json:"entity" binding:"required,oneof=issue cycle module"`
	ProjectID uint64 `json:"project_id" binding:"required"`
	RQL      string `json:"rql" binding:"required"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type RQLSearchResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *RQLError   `json:"error,omitempty"`
}

type RQLError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
```

- [ ] **Step 2: 创建 Handler**

```go
// backend-go/internal/rql/handler.go

package rql

import (
	"net/http"

	"reqman/backend-go/internal/dto/request"

	"github.com/gin-gonic/gin"
)

type RQLHandler struct{}

func NewRQLHandler() *RQLHandler {
	return &RQLHandler{}
}

func (h *RQLHandler) Search(c *gin.Context) {
	var req request.RQLSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, request.RQLSearchResponse{
			Success: false,
			Error: &request.RQLError{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 词法分析
	lexer := NewLexer(req.RQL)
	tokens, err := lexer.Tokenize()
	if err != nil {
		c.JSON(http.StatusOK, request.RQLSearchResponse{
			Success: false,
			Error: &request.RQLError{
				Code:    "RQL_LEX_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// 语法分析
	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		c.JSON(http.StatusOK, request.RQLSearchResponse{
			Success: false,
			Error: &request.RQLError{
				Code:    "RQL_PARSE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// 构建查询
	builder := NewQueryBuilder()
	whereClause, args, err := builder.Build(ast)
	if err != nil {
		c.JSON(http.StatusOK, request.RQLSearchResponse{
			Success: false,
			Error: &request.RQLError{
				Code:    "RQL_BUILD_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// TODO: 根据 entity 类型执行不同的查询
	// 这里先返回模拟数据，实际需要接入数据库

	c.JSON(http.StatusOK, request.RQLSearchResponse{
		Success: true,
		Data: map[string]interface{}{
			"items":    []interface{}{},
			"total":    0,
			"page":     req.Page,
			"page_size": req.PageSize,
			"where":    whereClause,
			"args":     args,
		},
	})
}
```

- [ ] **Step 3: 更新路由**

在 `backend-go/internal/router/router.go` 中添加 RQL 路由：

```go
import "reqman/backend-go/internal/rql"

// 在 Setup 函数中添加
rqlHandler := rql.NewRQLHandler()
rqlGroup := v1.Group("/rql")
{
	rqlGroup.POST("/search", rqlHandler.Search)
}
```

- [ ] **Step 4: 提交**

```bash
git add backend-go/internal/rql/handler.go backend-go/internal/dto/request/rql.go backend-go/internal/router/router.go
git commit -m "feat(rql): add RQL HTTP handler and routes"
```

---

### Task 6: 前端 - RQL 类型和 API

**Files:**
- Create: `frontend/src/utils/rql/types.ts`
- Create: `frontend/src/api/rql.ts`

- [ ] **Step 1: 创建 RQL 类型定义**

```typescript
// frontend/src/utils/rql/types.ts

export interface RQLSearchRequest {
  entity: 'issue' | 'cycle' | 'module'
  project_id: number
  rql: string
  page?: number
  page_size?: number
}

export interface RQLSearchResponse {
  success: boolean
  data?: {
    items: any[]
    total: number
    page: number
    page_size: number
  }
  error?: {
    code: string
    message: string
  }
}

export interface RQLHistoryItem {
  id: string
  rql: string
  timestamp: number
  entityType: 'issue' | 'cycle' | 'module'
}
```

- [ ] **Step 2: 创建 RQL API**

```typescript
// frontend/src/api/rql.ts

import api from './index'
import type { RQLSearchRequest, RQLSearchResponse } from '../utils/rql/types'

export const rqlApi = {
  search: (data: RQLSearchRequest) => {
    return api.post<RQLSearchResponse>('/rql/search', data)
  }
}
```

- [ ] **Step 3: 更新 API 导出**

在 `frontend/src/api/index.ts` 中添加导出

- [ ] **Step 4: 提交**

```bash
git add frontend/src/utils/rql/types.ts frontend/src/api/rql.ts
git commit -m "feat(rql): add frontend RQL types and API client"
```

---

### Task 7: 前端 - useRQL Composable

**Files:**
- Create: `frontend/src/composables/useRQL.ts`

- [ ] **Step 1: 创建 useRQL Composable**

```typescript
// frontend/src/composables/useRQL.ts

import { ref } from 'vue'
import { rqlApi } from '../api/rql'
import type { RQLSearchRequest, RQLSearchResponse, RQLHistoryItem } from '../utils/rql/types'

const HISTORY_KEY = 'reqman:rql:history'
const MAX_HISTORY = 50

export function useRQL() {
  const rql = ref('')
  const loading = ref(false)
  const error = ref<string | null>(null)
  const results = ref<any[]>([])
  const total = ref(0)

  // 历史记录
  const getHistory = (): RQLHistoryItem[] => {
    try {
      const data = localStorage.getItem(HISTORY_KEY)
      return data ? JSON.parse(data) : []
    } catch {
      return []
    }
  }

  const addToHistory = (rqlText: string, entityType: 'issue' | 'cycle' | 'module' = 'issue') => {
    if (!rqlText.trim()) return

    const history = getHistory()
    const filtered = history.filter(h => h.rql !== rqlText)

    filtered.unshift({
      id: Date.now().toString(),
      rql: rqlText,
      timestamp: Date.now(),
      entityType
    })

    const trimmed = filtered.slice(0, MAX_HISTORY)
    localStorage.setItem(HISTORY_KEY, JSON.stringify(trimmed))
  }

  const clearHistory = () => {
    localStorage.removeItem(HISTORY_KEY)
  }

  // 执行搜索
  const search = async (projectId: number, entity: 'issue' | 'cycle' | 'module' = 'issue', page = 1, pageSize = 20) => {
    if (!rql.value.trim()) {
      error.value = null
      results.value = []
      total.value = 0
      return
    }

    loading.value = true
    error.value = null

    try {
      const request: RQLSearchRequest = {
        entity,
        project_id: projectId,
        rql: rql.value,
        page,
        page_size: pageSize
      }

      const response = await rqlApi.search(request)

      if (response.data.success) {
        results.value = response.data.data?.items || []
        total.value = response.data.data?.total || 0
        addToHistory(rql.value, entity)
      } else {
        error.value = response.data.error?.message || '查询失败'
      }
    } catch (err: any) {
      error.value = err.response?.data?.error?.message || err.message || '网络错误'
    } finally {
      loading.value = false
    }
  }

  return {
    rql,
    loading,
    error,
    results,
    total,
    search,
    getHistory,
    addToHistory,
    clearHistory
  }
}
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/composables/useRQL.ts
git commit -m "feat(rql): add useRQL composable"
```

---

### Task 8: 前端 - RQLInput 组件

**Files:**
- Create: `frontend/src/components/RQL/RQLInput.vue`
- Create: `frontend/src/components/RQL/RQLHistory.vue`
- Create: `frontend/src/components/RQL/index.ts`

- [ ] **Step 1: 创建 RQLInput 组件**

```vue
<!-- frontend/src/components/RQL/RQLInput.vue -->

<template>
  <div class="rql-input">
    <div class="relative">
      <input
        ref="inputRef"
        v-model="rql"
        type="text"
        :placeholder="placeholder"
        class="w-full pl-8 pr-20 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        :class="inputClass"
        @keydown.enter="handleSearch"
        @focus="showHistoryPanel = true"
        @blur="onBlur"
      />
      <svg class="w-4 h-4 text-gray-400 absolute left-2.5 top-1/2 -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
      <div class="absolute right-2 top-1/2 -translate-y-1/2 flex space-x-1">
        <button
          v-if="rql"
          @click="clearRQL"
          class="p-1 text-gray-400 hover:text-gray-600"
          title="清除"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
        <button
          v-if="showHistory"
          @click="toggleHistory"
          class="p-1 text-gray-400 hover:text-gray-600"
          title="历史"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </button>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="error" class="mt-1 text-xs text-red-500">
      ✗ {{ error }}
    </div>

    <!-- 提示 -->
    <div v-if="showHints && !rql" class="mt-1 text-xs text-gray-400">
      示例: <code class="bg-gray-100 px-1 rounded">state = "待处理" AND priority = "high"</code>
    </div>

    <!-- 历史记录面板 -->
    <RQLHistory
      v-if="showHistoryPanel"
      @select="onSelectHistory"
      @close="showHistoryPanel = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import RQLHistory from './RQLHistory.vue'

const props = withDefaults(defineProps<{
  modelValue?: string
  placeholder?: string
  showHistory?: boolean
  showHints?: boolean
  error?: string | null
}>(), {
  modelValue: '',
  placeholder: '输入 RQL 查询...',
  showHistory: true,
  showHints: true,
  error: null
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'search': [value: string]
}>()

const inputRef = ref<HTMLInputElement | null>(null)
const showHistoryPanel = ref(false)
const rql = ref(props.modelValue)

const inputClass = computed(() => {
  if (!rql.value) return 'border-gray-300'
  if (props.error) return 'border-red-300'
  return 'border-green-300'
})

watch(() => props.modelValue, (val) => {
  rql.value = val
})

watch(rql, (val) => {
  emit('update:modelValue', val)
})

const handleSearch = () => {
  emit('search', rql.value)
}

const clearRQL = () => {
  rql.value = ''
  emit('update:modelValue', '')
}

const toggleHistory = () => {
  showHistoryPanel.value = !showHistoryPanel.value
}

const onSelectHistory = (item: any) => {
  rql.value = item.rql
  showHistoryPanel.value = false
  emit('update:modelValue', item.rql)
  emit('search', item.rql)
}

const onBlur = () => {
  setTimeout(() => {
    showHistoryPanel.value = false
  }, 200)
}

const focus = () => {
  inputRef.value?.focus()
}

defineExpose({ focus })
</script>
```

- [ ] **Step 2: 创建 RQLHistory 组件**

```vue
<!-- frontend/src/components/RQL/RQLHistory.vue -->

<template>
  <div class="absolute left-0 top-full mt-1 w-80 bg-white border border-gray-200 rounded-lg shadow-lg z-50">
    <div class="px-3 py-2 border-b border-gray-100 flex items-center justify-between">
      <span class="text-sm font-medium text-gray-700">查询历史</span>
      <button @click="$emit('close')" class="text-gray-400 hover:text-gray-600">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
    <div class="max-h-64 overflow-y-auto">
      <div v-if="history.length === 0" class="px-3 py-4 text-sm text-gray-500 text-center">
        暂无历史记录
      </div>
      <div
        v-for="item in history"
        :key="item.id"
        @click="$emit('select', item)"
        class="px-3 py-2 hover:bg-gray-50 cursor-pointer border-b border-gray-50 last:border-b-0"
      >
        <div class="text-sm text-gray-800 truncate">{{ item.rql }}</div>
        <div class="text-xs text-gray-400 mt-0.5">{{ formatTime(item.timestamp) }}</div>
      </div>
    </div>
    <div v-if="history.length > 0" class="px-3 py-2 border-t border-gray-100">
      <button @click="clearHistory" class="text-xs text-red-500 hover:text-red-700">
        清除历史记录
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const emit = defineEmits<{
  'select': [item: any]
  'close': []
}>()

const HISTORY_KEY = 'reqman:rql:history'
const history = ref<any[]>([])

onMounted(() => {
  try {
    const data = localStorage.getItem(HISTORY_KEY)
    history.value = data ? JSON.parse(data) : []
  } catch {
    history.value = []
  }
})

const clearHistory = () => {
  localStorage.removeItem(HISTORY_KEY)
  history.value = []
}

const formatTime = (timestamp: number): string => {
  const diff = Date.now() - timestamp
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
  return `${Math.floor(diff / 86400000)} 天前`
}
</script>
```

- [ ] **Step 3: 创建组件导出**

```typescript
// frontend/src/components/RQL/index.ts

export { default as RQLInput } from './RQLInput.vue'
export { default as RQLHistory } from './RQLHistory.vue'
```

- [ ] **Step 4: 提交**

```bash
git add frontend/src/components/RQL/
git commit -m "feat(rql): add RQLInput and RQLHistory components"
```

---

### Task 9: 前端 - 集成到 IssueList

**Files:**
- Modify: `frontend/src/components/IssueList.vue`

- [ ] **Step 1: 集成 RQLInput 到 IssueList**

在 IssueList.vue 中替换现有搜索框：

```vue
<!-- 替换现有的搜索输入框 -->
<RQLInput
  v-model="rqlQuery"
  placeholder="输入 RQL 查询..."
  :error="rqlError"
  @search="onRQLSearch"
/>
```

添加逻辑：

```typescript
import { RQLInput } from './RQL'
import { useRQL } from '../composables/useRQL'

const {
  rql: rqlQuery,
  loading: rqlLoading,
  error: rqlError,
  search: doRQLSearch,
  results: rqlResults
} = useRQL()

const onRQLSearch = async () => {
  await doRQLSearch(projectId.value, 'issue')
  // 将结果设置到 issues
  if (!rqlError.value) {
    issues.value = rqlResults.value
  }
}
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/components/IssueList.vue
git commit -m "feat(rql): integrate RQLInput into IssueList"
```

---

### Task 10: 前端 - 集成到 IssueKanban

**Files:**
- Modify: `frontend/src/components/IssueKanban.vue`

- [ ] **Step 1: 集成 RQLInput 到 IssueKanban**

类似 IssueList 集成。

- [ ] **Step 2: 提交**

```bash
git add frontend/src/components/IssueKanban.vue
git commit -m "feat(rql): integrate RQLInput into IssueKanban"
```

---

## 实施总结

完成上述 10 个任务后，RQL MVP 功能将具备：

**后端 (Go):**
- ✅ RQL Lexer - 词法分析
- ✅ RQL Parser - 语法分析 → AST
- ✅ Query Builder - AST → SQL WHERE
- ✅ HTTP Handler - API 端点
- ✅ 路由注册

**前端 (Vue):**
- ✅ RQL 类型定义
- ✅ RQL API 客户端
- ✅ useRQL Composable
- ✅ RQLInput/RQLHistory 组件
- ✅ 集成到 IssueList/IssueKanban

**API:**
- POST `/api/rql/search` - 接收 RQL 字符串，返回查询结果
