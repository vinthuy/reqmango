// backend-go/internal/rql/builder_test.go

package rql

import (
	"testing"
)

func TestBuilder_SimpleComparison(t *testing.T) {
	lexer := NewLexer(`state = "待处理"`)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	ast, _ := parser.Parse()

	builder := NewQueryBuilder()
	where, args, err := builder.Build(ast)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if where != "state_id = ?" {
		t.Errorf("expected 'state_id = ?', got '%s'", where)
	}

	if len(args) != 1 || args[0] != "待处理" {
		t.Errorf("expected args ['待处理'], got %v", args)
	}
}

func TestBuilder_AndExpression(t *testing.T) {
	lexer := NewLexer(`state = "待处理" AND priority = "high"`)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	ast, _ := parser.Parse()

	builder := NewQueryBuilder()
	where, args, err := builder.Build(ast)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if where != "(state_id = ? AND priority = ?)" {
		t.Errorf("unexpected where clause: %s", where)
	}

	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

func TestBuilder_LikeExpression(t *testing.T) {
	lexer := NewLexer(`name LIKE "登录"`)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	ast, _ := parser.Parse()

	builder := NewQueryBuilder()
	where, args, err := builder.Build(ast)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if where != "name LIKE ?" {
		t.Errorf("expected 'name LIKE ?', got '%s'", where)
	}

	if len(args) != 1 || args[0] != "%登录%" {
		t.Errorf("expected args ['%%登录%%'], got %v", args)
	}
}

func TestBuilder_OrExpression(t *testing.T) {
	lexer := NewLexer(`assignee = "张三" OR assignee = "李四"`)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	ast, _ := parser.Parse()

	builder := NewQueryBuilder()
	where, _, err := builder.Build(ast)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if where != "(assignee_id = ? OR assignee_id = ?)" {
		t.Errorf("unexpected where clause: %s", where)
	}
}

func TestBuilder_NestedParentheses(t *testing.T) {
	lexer := NewLexer(`(state = "待处理" OR state = "进行中") AND priority = "high"`)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	ast, _ := parser.Parse()

	builder := NewQueryBuilder()
	where, _, err := builder.Build(ast)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should properly handle nested parentheses
	if where != "((state_id = ? OR state_id = ?) AND priority = ?)" {
		t.Errorf("unexpected where clause: %s", where)
	}
}

func TestBuilder_NilNode(t *testing.T) {
	builder := NewQueryBuilder()
	where, args, err := builder.Build(nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if where != "" {
		t.Errorf("expected empty where, got '%s'", where)
	}

	if len(args) != 0 {
		t.Errorf("expected empty args, got %v", args)
	}
}