// backend/internal/rql/parser_test.go

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