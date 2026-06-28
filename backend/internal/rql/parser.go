// backend/internal/rql/parser.go

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
		return node, &p.errors[0]
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
			Message:  "期望值",
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