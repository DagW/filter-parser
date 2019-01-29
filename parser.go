package scim_filter_parser

import (
	"fmt"
	"io"
)

// Parser is a parser.
type Parser struct {
	s      *Scanner
	prefix string
	buf    struct {
		token   Token  // last read token
		literal string // last read literal
		n       int    // buffer size (max = 1)
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// Parse returns an abstract syntax tree of the string in the scanner.
func (parser *Parser) Parse() (Expression, error) {
	return parser.expression(LowestPrecedence)
}

// expression is the implementation of the Pratt parser.
func (parser *Parser) expression(precedence int) (Expression, error) {
	var left interface{}
	token, literal := parser.scanIgnoreWhitespace()

	if parser.peek() == LeftBracket {
		parser.prefix = literal
		token, literal = parser.scanIgnoreWhitespace()
	}

	switch token {
	case UNKNOWN:
		return nil, fmt.Errorf("unknown token: %q", literal)
	case LeftParenthesis:
		expression, err := parser.expression(LowestPrecedence)
		if err != nil {
			return nil, err
		}
		parenthesis, parenthesisLiteral := parser.scanIgnoreWhitespace()
		if parenthesis != RightParenthesis {
			return nil, fmt.Errorf("found %q, expected right parenthesis", parenthesisLiteral)
		}

		left = expression
	case LeftBracket:
		expression, err := parser.expression(LowestPrecedence)
		if err != nil {
			return nil, err
		}
		parenthesis, parenthesisLiteral := parser.scanIgnoreWhitespace()
		if parenthesis != RightBracket {
			return nil, fmt.Errorf("found %q, expected right parenthesis", parenthesisLiteral)
		}

		parser.prefix = ""
		left = expression
	case IDENTIFIER:
		operator, operatorLiteral := parser.scanIgnoreWhitespace()
		if !operator.IsOperator() {
			return nil, fmt.Errorf("found %q, expected operator", operatorLiteral)
		}

		value, valueLiteral := parser.scanIgnoreWhitespace()
		if value != VALUE && valueLiteral != "" {
			return nil, fmt.Errorf("found %q, expected value", token)
		}

		if parser.prefix != "" {
			literal = parser.prefix + "." + literal
		}

		left = ValueExpression{
			Name:     literal,
			Operator: operator,
			Value:    valueLiteral,
		}
	case NOT:
		expression, err := parser.expression(HighestPrecedence)
		if err != nil {
			return nil, err
		}
		left = UnaryExpression{
			X:        expression,
			Operator: NOT,
		}
	}

	for precedence < parser.peek().Precedence() {
		token, _ := parser.scanIgnoreWhitespace()
		if token.IsAssociative() {
			expression, err := parser.expression(token.Precedence())
			if err != nil {
				return nil, err
			}
			left = BinaryExpression{
				X:        left,
				Operator: token,
				Y:        expression,
			}
		}
	}

	return left, nil
}

// scan returns the next token in the scanner.
func (parser *Parser) scan() (Token, string) {
	if parser.buf.n != 0 {
		parser.buf.n = 0
		return parser.buf.token, parser.buf.literal
	}

	token, literal := parser.s.Scan()
	parser.buf.token, parser.buf.literal = token, literal

	return token, literal
}

// unscan places the last read token back in the buffer.
func (parser *Parser) unscan() {
	parser.buf.n = 1
}

// peek returns the next token in the scanner that is not whitespace.
func (parser *Parser) peek() Token {
	token, _ := parser.scan()
	if token == WHITESPACE {
		token, _ = parser.scan()
		parser.unscan()
	}
	parser.unscan()
	return token
}

// scanIgnoreWhiteSpace scans the next token that is not whitespace.
func (parser *Parser) scanIgnoreWhitespace() (Token, string) {
	token, literal := parser.scan()
	if token == WHITESPACE {
		token, literal = parser.scan()
	}
	return token, literal
}