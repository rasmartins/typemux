package lexer

import (
	"testing"
)

func TestNew(t *testing.T) {
	input := "enum UserRole"
	l := New(input)

	if l.input != input {
		t.Errorf("Expected input to be %q, got %q", input, l.input)
	}
	if l.line != 1 {
		t.Errorf("Expected line to be 1, got %d", l.line)
	}
	if l.position != 0 {
		t.Errorf("Expected position to be 0, got %d", l.position)
	}
}

func TestNextToken_Keywords(t *testing.T) {
	tests := []struct {
		input    string
		expected []TokenType
	}{
		{
			input:    "enum",
			expected: []TokenType{TOKEN_ENUM},
		},
		{
			input:    "type",
			expected: []TokenType{TOKEN_TYPE},
		},
		{
			input:    "service",
			expected: []TokenType{TOKEN_SERVICE},
		},
		{
			input:    "rpc",
			expected: []TokenType{TOKEN_RPC},
		},
		{
			input:    "returns",
			expected: []TokenType{TOKEN_RETURNS},
		},
		{
			input:    "enum type service rpc returns",
			expected: []TokenType{TOKEN_ENUM, TOKEN_TYPE, TOKEN_SERVICE, TOKEN_RPC, TOKEN_RETURNS},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			for i, expectedType := range tt.expected {
				tok := l.NextToken()
				if tok.Type != expectedType {
					t.Errorf("Token %d: expected type %s, got %s", i, expectedType, tok.Type)
				}
			}
		})
	}
}

func TestNextToken_Identifiers(t *testing.T) {
	input := "UserRole myType _privateField field123"
	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_IDENT, "UserRole"},
		{TOKEN_IDENT, "myType"},
		{TOKEN_IDENT, "_privateField"},
		{TOKEN_IDENT, "field123"},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("Token %d: expected type %s, got %s", i, exp.tokenType, tok.Type)
		}
		if tok.Literal != exp.literal {
			t.Errorf("Token %d: expected literal %q, got %q", i, exp.literal, tok.Literal)
		}
	}
}

func TestNextToken_Numbers(t *testing.T) {
	input := "123 456 0 9999"
	expected := []string{"123", "456", "0", "9999"}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != TOKEN_NUMBER {
			t.Errorf("Token %d: expected type TOKEN_NUMBER, got %s", i, tok.Type)
		}
		if tok.Literal != exp {
			t.Errorf("Token %d: expected literal %q, got %q", i, exp, tok.Literal)
		}
	}
}

func TestNextToken_Symbols(t *testing.T) {
	input := "{ } ( ) [ ] , : @ < >"
	expected := []TokenType{
		TOKEN_LBRACE, TOKEN_RBRACE,
		TOKEN_LPAREN, TOKEN_RPAREN,
		TOKEN_LBRACKET, TOKEN_RBRACKET,
		TOKEN_COMMA, TOKEN_COLON, TOKEN_AT,
		TOKEN_LT, TOKEN_GT,
	}

	l := New(input)
	for i, expectedType := range expected {
		tok := l.NextToken()
		if tok.Type != expectedType {
			t.Errorf("Token %d: expected type %s, got %s", i, expectedType, tok.Type)
		}
	}
}

func TestNextToken_Comments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "single line comment",
			input:    "// This is a comment\nenum",
			expected: []TokenType{TOKEN_ENUM},
		},
		{
			name:     "comment between tokens",
			input:    "enum // comment\nUserRole",
			expected: []TokenType{TOKEN_ENUM, TOKEN_IDENT},
		},
		{
			name:     "multiple comments",
			input:    "// comment 1\n// comment 2\ntype",
			expected: []TokenType{TOKEN_TYPE},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, expectedType := range tt.expected {
				tok := l.NextToken()
				if tok.Type != expectedType {
					t.Errorf("Token %d: expected type %s, got %s", i, expectedType, tok.Type)
				}
			}
		})
	}
}

func TestNextToken_Whitespace(t *testing.T) {
	input := "enum  \t\n  UserRole"
	expected := []TokenType{TOKEN_ENUM, TOKEN_IDENT}

	l := New(input)
	for i, expectedType := range expected {
		tok := l.NextToken()
		if tok.Type != expectedType {
			t.Errorf("Token %d: expected type %s, got %s", i, expectedType, tok.Type)
		}
	}
}

func TestNextToken_ComplexInput(t *testing.T) {
	input := `
enum UserRole {
  ADMIN
  USER
}

type User {
  id: string @required
  age: int32
}
`
	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_ENUM, "enum"},
		{TOKEN_IDENT, "UserRole"},
		{TOKEN_LBRACE, "{"},
		{TOKEN_IDENT, "ADMIN"},
		{TOKEN_IDENT, "USER"},
		{TOKEN_RBRACE, "}"},
		{TOKEN_TYPE, "type"},
		{TOKEN_IDENT, "User"},
		{TOKEN_LBRACE, "{"},
		{TOKEN_IDENT, "id"},
		{TOKEN_COLON, ":"},
		{TOKEN_IDENT, "string"},
		{TOKEN_AT, "@"},
		{TOKEN_IDENT, "required"},
		{TOKEN_IDENT, "age"},
		{TOKEN_COLON, ":"},
		{TOKEN_IDENT, "int32"},
		{TOKEN_RBRACE, "}"},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("Token %d: expected type %s, got %s", i, exp.tokenType, tok.Type)
		}
		if tok.Literal != exp.literal {
			t.Errorf("Token %d: expected literal %q, got %q", i, exp.literal, tok.Literal)
		}
	}
}

func TestNextToken_ArrayType(t *testing.T) {
	input := "[]string"
	expected := []TokenType{TOKEN_LBRACKET, TOKEN_RBRACKET, TOKEN_IDENT}

	l := New(input)
	for i, expectedType := range expected {
		tok := l.NextToken()
		if tok.Type != expectedType {
			t.Errorf("Token %d: expected type %s, got %s", i, expectedType, tok.Type)
		}
	}
}

func TestNextToken_MapType(t *testing.T) {
	input := "map<string, int32>"
	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_IDENT, "map"},
		{TOKEN_LT, "<"},
		{TOKEN_IDENT, "string"},
		{TOKEN_COMMA, ","},
		{TOKEN_IDENT, "int32"},
		{TOKEN_GT, ">"},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("Token %d: expected type %s, got %s", i, exp.tokenType, tok.Type)
		}
		if tok.Literal != exp.literal {
			t.Errorf("Token %d: expected literal %q, got %q", i, exp.literal, tok.Literal)
		}
	}
}

func TestNextToken_ServiceDefinition(t *testing.T) {
	input := "service UserService { rpc GetUser(Request) returns (Response) }"
	expected := []TokenType{
		TOKEN_SERVICE, TOKEN_IDENT, TOKEN_LBRACE,
		TOKEN_RPC, TOKEN_IDENT,
		TOKEN_LPAREN, TOKEN_IDENT, TOKEN_RPAREN,
		TOKEN_RETURNS,
		TOKEN_LPAREN, TOKEN_IDENT, TOKEN_RPAREN,
		TOKEN_RBRACE,
	}

	l := New(input)
	for i, expectedType := range expected {
		tok := l.NextToken()
		if tok.Type != expectedType {
			t.Errorf("Token %d: expected type %s, got %s (literal: %q)", i, expectedType, tok.Type, tok.Literal)
		}
	}
}

func TestNextToken_EOF(t *testing.T) {
	input := "enum"
	l := New(input)

	tok := l.NextToken() // enum
	if tok.Type != TOKEN_ENUM {
		t.Errorf("Expected TOKEN_ENUM, got %s", tok.Type)
	}

	tok = l.NextToken() // EOF
	if tok.Type != TOKEN_EOF {
		t.Errorf("Expected TOKEN_EOF, got %s", tok.Type)
	}

	// Multiple calls should keep returning EOF
	tok = l.NextToken()
	if tok.Type != TOKEN_EOF {
		t.Errorf("Expected TOKEN_EOF, got %s", tok.Type)
	}
}

func TestNextToken_LineAndColumn(t *testing.T) {
	input := "enum\ntype"
	l := New(input)

	tok := l.NextToken()
	if tok.Line != 1 {
		t.Errorf("Expected line 1, got %d", tok.Line)
	}

	tok = l.NextToken()
	if tok.Line != 2 {
		t.Errorf("Expected line 2, got %d", tok.Line)
	}
}

func TestTokenType_String(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  string
	}{
		{TOKEN_EOF, "EOF"},
		{TOKEN_IDENT, "IDENT"},
		{TOKEN_ENUM, "ENUM"},
		{TOKEN_TYPE, "TYPE"},
		{TOKEN_SERVICE, "SERVICE"},
		{TOKEN_RPC, "RPC"},
		{TOKEN_RETURNS, "RETURNS"},
		{TOKEN_LBRACE, "{"},
		{TOKEN_RBRACE, "}"},
		{TOKEN_LPAREN, "("},
		{TOKEN_RPAREN, ")"},
		{TOKEN_LBRACKET, "["},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_COMMA, ","},
		{TOKEN_COLON, ":"},
		{TOKEN_AT, "@"},
		{TOKEN_LT, "<"},
		{TOKEN_GT, ">"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.tokenType.String(); got != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestLexerEmptyInput(t *testing.T) {
	l := New("")
	tok := l.NextToken()
	if tok.Type != TOKEN_EOF {
		t.Errorf("Expected TOKEN_EOF for empty input, got %s", tok.Type)
	}
}

func TestLexerOnlyWhitespace(t *testing.T) {
	l := New("   \t\n\r   ")
	tok := l.NextToken()
	if tok.Type != TOKEN_EOF {
		t.Errorf("Expected TOKEN_EOF for whitespace-only input, got %s", tok.Type)
	}
}

func TestLexerOnlyComments(t *testing.T) {
	l := New("// comment 1\n// comment 2\n")
	tok := l.NextToken()
	if tok.Type != TOKEN_EOF {
		t.Errorf("Expected TOKEN_EOF for comments-only input, got %s", tok.Type)
	}
}

func TestNextToken_DocComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			typ     TokenType
			literal string
		}
	}{
		{
			name:  "single doc comment",
			input: "/// This is a doc comment",
			expected: []struct {
				typ     TokenType
				literal string
			}{
				{TOKEN_DOC_COMMENT, "This is a doc comment"},
			},
		},
		{
			name:  "doc comment with extra spaces",
			input: "///   Extra spaces",
			expected: []struct {
				typ     TokenType
				literal string
			}{
				{TOKEN_DOC_COMMENT, "Extra spaces"},
			},
		},
		{
			name:  "multiple doc comments",
			input: "/// First line\n/// Second line",
			expected: []struct {
				typ     TokenType
				literal string
			}{
				{TOKEN_DOC_COMMENT, "First line"},
				{TOKEN_DOC_COMMENT, "Second line"},
			},
		},
		{
			name:  "doc comment before enum",
			input: "/// User roles\nenum UserRole",
			expected: []struct {
				typ     TokenType
				literal string
			}{
				{TOKEN_DOC_COMMENT, "User roles"},
				{TOKEN_ENUM, "enum"},
				{TOKEN_IDENT, "UserRole"},
			},
		},
		{
			name:  "language-specific doc comment",
			input: "/// @proto This is proto-specific",
			expected: []struct {
				typ     TokenType
				literal string
			}{
				{TOKEN_DOC_COMMENT, "@proto This is proto-specific"},
			},
		},
		{
			name:  "regular comment vs doc comment",
			input: "// Regular comment\n/// Doc comment\nenum",
			expected: []struct {
				typ     TokenType
				literal string
			}{
				{TOKEN_DOC_COMMENT, "Doc comment"},
				{TOKEN_ENUM, "enum"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, expected := range tt.expected {
				tok := l.NextToken()
				if tok.Type != expected.typ {
					t.Errorf("Token %d: expected type %s, got %s", i, expected.typ, tok.Type)
				}
				if tok.Literal != expected.literal {
					t.Errorf("Token %d: expected literal %q, got %q", i, expected.literal, tok.Literal)
				}
			}
		})
	}
}

func TestReadDocComment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple doc comment",
			input:    "/// Hello world",
			expected: "Hello world",
		},
		{
			name:     "doc comment with no space",
			input:    "///Hello",
			expected: "Hello",
		},
		{
			name:     "doc comment with trailing spaces",
			input:    "/// Text with trailing   ",
			expected: "Text with trailing   ",
		},
		{
			name:     "empty doc comment",
			input:    "///",
			expected: "",
		},
		{
			name:     "doc comment with special characters",
			input:    "/// @proto field description!",
			expected: "@proto field description!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()
			if tok.Type != TOKEN_DOC_COMMENT {
				t.Fatalf("Expected TOKEN_DOC_COMMENT, got %s", tok.Type)
			}
			if tok.Literal != tt.expected {
				t.Errorf("Expected literal %q, got %q", tt.expected, tok.Literal)
			}
		})
	}
}

func TestNextToken_Strings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			typ     TokenType
			literal string
		}
	}{
		{
			name:  "simple string",
			input: `"hello"`,
			expected: []struct {
				typ     TokenType
				literal string
			}{
				{TOKEN_STRING, "hello"},
			},
		},
		{
			name:  "string with path",
			input: `"/users/{id}"`,
			expected: []struct {
				typ     TokenType
				literal string
			}{
				{TOKEN_STRING, "/users/{id}"},
			},
		},
		{
			name:  "empty string",
			input: `""`,
			expected: []struct {
				typ     TokenType
				literal string
			}{
				{TOKEN_STRING, ""},
			},
		},
		{
			name:  "string with spaces",
			input: `"hello world"`,
			expected: []struct {
				typ     TokenType
				literal string
			}{
				{TOKEN_STRING, "hello world"},
			},
		},
		{
			name:  "string in annotation",
			input: `@path("/api/v1/users")`,
			expected: []struct {
				typ     TokenType
				literal string
			}{
				{TOKEN_AT, "@"},
				{TOKEN_IDENT, "path"},
				{TOKEN_LPAREN, "("},
				{TOKEN_STRING, "/api/v1/users"},
				{TOKEN_RPAREN, ")"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, expected := range tt.expected {
				tok := l.NextToken()
				if tok.Type != expected.typ {
					t.Errorf("Token %d: expected type %s, got %s", i, expected.typ, tok.Type)
				}
				if tok.Literal != expected.literal {
					t.Errorf("Token %d: expected literal %q, got %q", i, expected.literal, tok.Literal)
				}
			}
		})
	}
}
func TestNextToken_Namespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			typ     TokenType
			literal string
		}
	}{
		{
			name:  "simple namespace",
			input: "namespace api",
			expected: []struct {
				typ     TokenType
				literal string
			}{
				{TOKEN_NAMESPACE, "namespace"},
				{TOKEN_IDENT, "api"},
			},
		},
		{
			name:  "dotted namespace",
			input: "namespace com.example.api",
			expected: []struct {
				typ     TokenType
				literal string
			}{
				{TOKEN_NAMESPACE, "namespace"},
				{TOKEN_IDENT, "com"},
				{TOKEN_DOT, "."},
				{TOKEN_IDENT, "example"},
				{TOKEN_DOT, "."},
				{TOKEN_IDENT, "api"},
			},
		},
		{
			name:  "namespace with type definition",
			input: "namespace myapi\n\ntype User",
			expected: []struct {
				typ     TokenType
				literal string
			}{
				{TOKEN_NAMESPACE, "namespace"},
				{TOKEN_IDENT, "myapi"},
				{TOKEN_TYPE, "type"},
				{TOKEN_IDENT, "User"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, expected := range tt.expected {
				tok := l.NextToken()
				if tok.Type != expected.typ {
					t.Errorf("Token %d: expected type %s, got %s", i, expected.typ, tok.Type)
				}
				if tok.Literal != expected.literal {
					t.Errorf("Token %d: expected literal %q, got %q", i, expected.literal, tok.Literal)
				}
			}
		})
	}
}

func TestTokenizeOptionalSyntax(t *testing.T) {
	input := `type User {
		name: string?
	}`

	l := New(input)

	expectedTokens := []TokenType{
		TOKEN_TYPE,
		TOKEN_IDENT, // User
		TOKEN_LBRACE,
		TOKEN_IDENT, // name
		TOKEN_COLON,
		TOKEN_IDENT,    // string
		TOKEN_QUESTION, // ?
		TOKEN_RBRACE,
	}

	for i, expected := range expectedTokens {
		tok := l.NextToken()
		if tok.Type != expected {
			t.Errorf("Token %d: expected type %s, got %s (literal: %q)",
				i, expected, tok.Type, tok.Literal)
		}
	}
}

func TestTokenizeQuestionMark(t *testing.T) {
	input := "string?"
	l := New(input)

	tok := l.NextToken()
	if tok.Type != TOKEN_IDENT || tok.Literal != "string" {
		t.Errorf("Expected IDENT 'string', got %s '%s'", tok.Type, tok.Literal)
	}

	tok = l.NextToken()
	if tok.Type != TOKEN_QUESTION {
		t.Errorf("Expected TOKEN_QUESTION, got %s", tok.Type)
	}
	if tok.Literal != "?" {
		t.Errorf("Expected literal '?', got '%s'", tok.Literal)
	}
}

func TestTokenTypeQuestionString(t *testing.T) {
	if TOKEN_QUESTION.String() != "?" {
		t.Errorf("Expected TOKEN_QUESTION.String() to be '?', got '%s'", TOKEN_QUESTION.String())
	}
}
