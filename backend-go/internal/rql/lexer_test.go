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

	if tokens[4].Type != TOKEN_RPAREN {
		t.Errorf("expected RPAREN, got %v", tokens[4].Type)
	}
}
