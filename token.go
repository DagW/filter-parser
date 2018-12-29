package scim_filtering

type Token int

const (
	UNKNOWN Token = iota

	WS   // whitespace (i.e. " ")
	ID   // identifier (i.e. "userName")
	EOF  // end-of-file
	V    // value (i.e. "john")

	// Attribute Operators
	EQ  // equal
	NE  // not equal
	CO  // contains
	SW  // starts with
	EW  // ends with
	PR  // present (has value)
	GT  // greater than
	GE  // greater than or equal to
	LT  // less than
	LE  // less than or equal to

	// Logical Operators
	AND  // logical "and
	OR   // logical "or"
	NOT  // "not" function
)

var tokens = [...]string{
	UNKNOWN: "unknown",

	WS:  " ",
	ID:  "id",
	EOF: "",
	V:   "value",

	EQ: "equal",
	NE: "not equal",
	CO: "contains",
	SW: "starts with",
	EW: "ends with",
	PR: "present",
	GT: ">",
	GE: ">=",
	LT: "<",
	LE: "<=",

	AND: "and",
	OR:  "or",
	NOT: "not",
}

func (token Token) IsOperator() bool {
	switch token {
	case EQ, NE, CO, SW, EW, PR, GT, GE, LT, LE:
		return true
	}
	return false
}

func (token Token) IsAssociative() bool {
	switch token {
	case AND, OR:
		return true
	}
	return false
}

const (
	HighestPrecedence = 2
	LowestPrecedence  = 0
)

func (token Token) Precedence() int {
	switch token {
	case AND:
		return 2
	case OR:
		return 1
	}
	return 0
}

func (token Token) String() string {
	if 0 <= token && token < Token(len(tokens)) {
		return tokens[token]
	}
	return tokens[0]
}
