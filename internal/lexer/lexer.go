package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

// TokenType represents the type of a lexical token in the TypeMUX language.
type TokenType int

// Token types for the TypeMUX IDL lexer.
const (
	TOKEN_EOF TokenType = iota
	TOKEN_IDENT
	TOKEN_NAMESPACE
	TOKEN_IMPORT
	TOKEN_ENUM
	TOKEN_TYPE
	TOKEN_UNION
	TOKEN_SERVICE
	TOKEN_RPC
	TOKEN_RETURNS
	TOKEN_STREAM
	TOKEN_LBRACE
	TOKEN_RBRACE
	TOKEN_LPAREN
	TOKEN_RPAREN
	TOKEN_LBRACKET
	TOKEN_RBRACKET
	TOKEN_COMMA
	TOKEN_COLON
	TOKEN_AT
	TOKEN_DOT
	TOKEN_LT
	TOKEN_GT
	TOKEN_EQUALS
	TOKEN_STRING
	TOKEN_NUMBER
	TOKEN_DOC_COMMENT
	TOKEN_QUESTION
)

// Token represents a single lexical token with its type, value, and location.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// Lexer tokenizes TypeMUX IDL source code into a stream of tokens.
type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
}

// New creates a new lexer for the given input string.
func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
	if l.ch == '\n' {
		l.line++
		l.column = 0
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	if l.ch == '/' && l.peekChar() == '/' {
		// Don't skip if it's a doc comment (///)
		l.readChar() // consume first /
		if l.peekChar() == '/' {
			// This is a doc comment, put back the position
			l.position--
			l.readPosition--
			l.ch = '/'
			return
		}
		// Regular comment, skip to end of line
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
	}
}

func (l *Lexer) readDocComment() string {
	// Assume we're at the first /
	l.readChar() // skip first /
	l.readChar() // skip second /
	l.readChar() // skip third /

	// Skip any leading whitespace
	for l.ch == ' ' || l.ch == '\t' {
		l.readChar()
	}

	start := l.position
	// Read until end of line
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}

	return l.input[start:l.position]
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readString() string {
	// Skip opening quote
	l.readChar()
	position := l.position

	for l.ch != '"' && l.ch != 0 {
		// Handle escape sequences
		if l.ch == '\\' {
			l.readChar() // skip backslash
		}
		l.readChar()
	}

	str := l.input[position:l.position]
	// Skip closing quote if present
	if l.ch == '"' {
		l.readChar()
	}

	return str
}

// NextToken returns the next token from the input stream.
func (l *Lexer) NextToken() Token {
	var tok Token

	// Skip whitespace and comments (but not doc comments)
	for {
		l.skipWhitespace()
		if l.ch == '/' && l.peekChar() == '/' {
			// Check for doc comment (///)
			if l.position+2 < len(l.input) && l.input[l.position+2] == '/' {
				// This is a doc comment, don't skip it
				break
			}
			l.skipComment()
		} else {
			break
		}
	}

	tok.Line = l.line
	tok.Column = l.column

	// Check for doc comment
	if l.ch == '/' && l.peekChar() == '/' && l.position+2 < len(l.input) && l.input[l.position+2] == '/' {
		tok.Type = TOKEN_DOC_COMMENT
		tok.Literal = l.readDocComment()
		return tok
	}

	switch l.ch {
	case '{':
		tok = Token{Type: TOKEN_LBRACE, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '}':
		tok = Token{Type: TOKEN_RBRACE, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '(':
		tok = Token{Type: TOKEN_LPAREN, Literal: string(l.ch), Line: l.line, Column: l.column}
	case ')':
		tok = Token{Type: TOKEN_RPAREN, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '[':
		tok = Token{Type: TOKEN_LBRACKET, Literal: string(l.ch), Line: l.line, Column: l.column}
	case ']':
		tok = Token{Type: TOKEN_RBRACKET, Literal: string(l.ch), Line: l.line, Column: l.column}
	case ',':
		tok = Token{Type: TOKEN_COMMA, Literal: string(l.ch), Line: l.line, Column: l.column}
	case ':':
		tok = Token{Type: TOKEN_COLON, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '@':
		tok = Token{Type: TOKEN_AT, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '.':
		tok = Token{Type: TOKEN_DOT, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '<':
		tok = Token{Type: TOKEN_LT, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '>':
		tok = Token{Type: TOKEN_GT, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '=':
		tok = Token{Type: TOKEN_EQUALS, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '?':
		tok = Token{Type: TOKEN_QUESTION, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '"':
		tok.Type = TOKEN_STRING
		tok.Literal = l.readString()
		return tok
	case 0:
		tok.Type = TOKEN_EOF
		tok.Literal = ""
		return tok
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = lookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = TOKEN_NUMBER
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = Token{Type: TOKEN_EOF, Literal: string(l.ch), Line: l.line, Column: l.column}
		}
	}

	l.readChar()
	return tok
}

func lookupIdent(ident string) TokenType {
	keywords := map[string]TokenType{
		"namespace": TOKEN_NAMESPACE,
		"import":    TOKEN_IMPORT,
		"enum":      TOKEN_ENUM,
		"type":      TOKEN_TYPE,
		"union":     TOKEN_UNION,
		"service":   TOKEN_SERVICE,
		"rpc":       TOKEN_RPC,
		"returns":   TOKEN_RETURNS,
		"stream":    TOKEN_STREAM,
	}

	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TOKEN_IDENT
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (t TokenType) String() string {
	names := map[TokenType]string{
		TOKEN_EOF:         "EOF",
		TOKEN_IDENT:       "IDENT",
		TOKEN_NAMESPACE:   "NAMESPACE",
		TOKEN_IMPORT:      "IMPORT",
		TOKEN_ENUM:        "ENUM",
		TOKEN_TYPE:        "TYPE",
		TOKEN_UNION:       "UNION",
		TOKEN_SERVICE:     "SERVICE",
		TOKEN_RPC:         "RPC",
		TOKEN_RETURNS:     "RETURNS",
		TOKEN_LBRACE:      "{",
		TOKEN_RBRACE:      "}",
		TOKEN_LPAREN:      "(",
		TOKEN_RPAREN:      ")",
		TOKEN_LBRACKET:    "[",
		TOKEN_RBRACKET:    "]",
		TOKEN_COMMA:       ",",
		TOKEN_COLON:       ":",
		TOKEN_AT:          "@",
		TOKEN_DOT:         ".",
		TOKEN_LT:          "<",
		TOKEN_GT:          ">",
		TOKEN_EQUALS:      "=",
		TOKEN_STRING:      "STRING",
		TOKEN_NUMBER:      "NUMBER",
		TOKEN_DOC_COMMENT: "DOC_COMMENT",
		TOKEN_QUESTION:    "?",
	}
	if name, ok := names[t]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(%d)", t)
}

// TokensToString converts a slice of tokens into a human-readable string representation.
func TokensToString(tokens []Token) string {
	parts := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		parts = append(parts, fmt.Sprintf("%s(%s)", tok.Type, tok.Literal))
	}
	return strings.Join(parts, " ")
}
