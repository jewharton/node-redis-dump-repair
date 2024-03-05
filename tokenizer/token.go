package tokenizer

import "fmt"

var (
	invalidToken = Token{kind: Invalid}
	newlineToken = Token{kind: Newline, value: "\n"}
	eofToken     = Token{kind: EOF}
)

// Kind represents a specific kind of token.
type Kind uint

const (
	// Invalid represents an invalid token.
	Invalid Kind = iota
	// String represents a string token.
	String
	// Newline represents a newline token.
	Newline
	// EOF represents an end-of-file token.
	EOF
)

// Token represents a lexical unit.
type Token struct {
	value string
	kind  Kind
}

// Value returns the string value of the token. It panics if the token's kind is Invalid or EOF.
func (t Token) Value() string {
	if t.kind == Invalid || t.kind == EOF {
		panic(fmt.Sprintf("cannot get value of token kind %d", t.kind))
	}
	return t.value
}

// Kind returns what kind of token this is.
func (t Token) Kind() Kind {
	return t.kind
}
