package filter

import (
	"fmt"
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/ast"
	"github.com/scim2/filter-parser/grammar"
	"github.com/scim2/filter-parser/types"
	"strconv"
	"strings"
)

// ParseAttrExp parses the given raw data as an AttributeExpression.
func ParseAttrExp(raw []byte) (AttributeExpression, error) {
	p, err := ast.New(raw)
	if err != nil {
		return AttributeExpression{}, err
	}
	node, err := grammar.AttrExp(p)
	if err != nil {
		return AttributeExpression{}, err
	}
	if _, err := p.Expect(parser.EOD); err != nil {
		return AttributeExpression{}, err
	}
	return parseAttrExp(node)
}

func parseAttrExp(node *ast.Node) (AttributeExpression, error) {
	if node.Type != typ.AttrExp {
		return AttributeExpression{}, invalidTypeError(typ.AttrExp, node.Type)
	}

	children := node.Children()
	if len(children) == 0 {
		return AttributeExpression{}, invalidLengthError(typ.AttrExp, 1, 0)
	}

	// AttrPath 'pr'
	attrPath, err := parseAttrPath(children[0])
	if err != nil {
		return AttributeExpression{}, err
	}

	if len(children) == 1 {
		return AttributeExpression{
			AttributePath: attrPath,
			Operator:      PR,
		}, nil
	}

	if l := len(children); l != 3 {
		return AttributeExpression{}, invalidLengthError(typ.AttrExp, 3, l)
	}

	var (
		compareOp    = CompareOperator(children[1].ValueString())
		compareValue interface{}
	)
	switch node := children[2]; node.Type {
	case typ.False:
		compareValue = false
	case typ.Null:
		compareValue = nil
	case typ.True:
		compareValue = true
	case typ.Number:
		value, err := parseNumber(node)
		if err != nil {
			return AttributeExpression{}, err
		}
		compareValue = value
	case typ.String:
		str := node.ValueString()
		str = strings.TrimPrefix(str, "\"")
		str = strings.TrimSuffix(str, "\"")
		compareValue = str
	default:
		return AttributeExpression{}, invalidChildTypeError(typ.AttrExp, node.Type)
	}

	return AttributeExpression{
		AttributePath: attrPath,
		Operator:      compareOp,
		CompareValue:  compareValue,
	}, nil
}

func parseNumber(node *ast.Node) (interface{}, error) {
	var frac, exp bool
	var nStr string
	for _, node := range node.Children() {
		switch t := node.Type; t {
		case typ.Minus:
			nStr = "-"
		case typ.Int:
			nStr += node.ValueString()
		case typ.Frac:
			frac = true
			children := node.Children()
			if l := len(children); l != 1 {
				return AttributeExpression{}, invalidLengthError(typ.Frac, 1, l)
			}
			nStr += fmt.Sprintf(".%s", children[0].ValueString())
		case typ.Exp:
			exp = true
			nStr += "e"
			for _, node := range node.Children() {
				switch t := node.Type; t {
				case typ.Sign, typ.Digits:
					nStr += node.ValueString()
				default:
					return AttributeExpression{}, invalidChildTypeError(typ.Number, node.Type)

				}
			}
		default:
			return AttributeExpression{}, invalidChildTypeError(typ.Number, node.Type)
		}
	}
	f, err := strconv.ParseFloat(nStr, 64)
	if err != nil {
		return AttributeExpression{}, &internalError{
			Message: err.Error(),
		}
	}

	// Integers can not contain fractional or exponent parts.
	// More info: https://tools.ietf.org/html/rfc7643#section-2.3.4
	if !frac && !exp {
		return int(f), nil
	}
	return f, err
}
