// backend/internal/rql/lexer.go

package rql

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
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

func (l *Lexer) Tokenize() ([]Token, *ParseError) {
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
		return tokens, &l.errors[0]
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
		char, size := utf8.DecodeRuneInString(l.input[l.position:])
		if char == quote {
			l.position += size // skip closing quote
			return Token{Type: TOKEN_STRING, Value: value.String(), Position: start}
		}
		value.WriteRune(char)
		l.position += size
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
		char, size := utf8.DecodeRuneInString(l.input[l.position:])
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			break
		}
		value.WriteRune(char)
		l.position += size
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

	for l.position < len(l.input) {
		char, size := utf8.DecodeRuneInString(l.input[l.position:])
		if !unicode.IsDigit(char) && char != '-' && char != 'T' && char != ':' && char != '.' {
			break
		}
		value.WriteRune(char)
		l.position += size
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
