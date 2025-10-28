# Contributing to TypeMUX

Thank you for your interest in contributing to TypeMUX! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please be respectful and constructive in all interactions with the community.

## How to Contribute

### Reporting Bugs

If you find a bug, please create an issue with:

1. **Clear title** describing the bug
2. **Steps to reproduce** the issue
3. **Expected behavior** vs **actual behavior**
4. **Environment details** (OS, Go version, TypeMUX version)
5. **Schema example** (if applicable) - minimal reproducible example
6. **Generated output** (if relevant)

**Example bug report:**

```
Title: Union types not generating correctly in GraphQL

Description:
When defining a union type with three options, the generated GraphQL
schema is missing the @oneOf directive.

Steps to reproduce:
1. Create schema with union type
2. Run: typemux -input schema.typemux -format graphql
3. Check generated/schema.graphql

Expected: union Message @oneOf = TextMessage | ImageMessage | VideoMessage
Actual: union Message = TextMessage | ImageMessage | VideoMessage

Environment:
- OS: Ubuntu 22.04
- Go version: 1.21
- TypeMUX version: main branch (commit abc123)

Schema:
[Attach minimal .typemux file]
```

### Requesting Features

For feature requests, create an issue with:

1. **Clear use case** - why you need this feature
2. **Proposed syntax** (if applicable)
3. **Examples** of how it would work
4. **Alternatives considered**

### Contributing Code

#### Setup Development Environment

1. **Fork and clone:**
```bash
git clone https://github.com/rasmartins/typemux.git
cd typemux
```

2. **Install Go 1.21+:**
```bash
go version  # Should be 1.21 or higher
```

3. **Build:**
```bash
go build -o typemux
```

4. **Run tests:**
```bash
go test ./...
```

#### Making Changes

1. **Create a branch:**
```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/bug-description
```

2. **Make your changes:**
   - Follow existing code style
   - Add tests for new functionality
   - Update documentation if needed
   - Ensure tests pass

3. **Run tests with coverage:**
```bash
go test ./... -cover
```

**Important:** Test coverage must remain above 90% per `CLAUDE.md`.

4. **Test examples:**
```bash
make examples  # or test manually
```

#### Code Style

- Follow standard Go formatting: `go fmt ./...`
- Use meaningful variable names
- Add comments for complex logic
- Document exported functions and types

**Example:**
```go
// ParseSchema parses a TypeMUX schema file and returns an AST.
// Returns an error if the file cannot be read or parsed.
func ParseSchema(filename string) (*ast.Schema, error) {
    // Implementation
}
```

#### Writing Tests

All new functionality must include tests:

**Unit tests:**
```go
func TestGraphQLGenerator_Union(t *testing.T) {
    schema := &ast.Schema{
        Unions: []*ast.Union{
            {Name: "Message", Options: []string{"TextMessage", "ImageMessage"}},
        },
    }

    gen := &GraphQLGenerator{}
    output := gen.Generate(schema)

    if !strings.Contains(output, "@oneOf") {
        t.Error("Expected @oneOf directive in union")
    }
}
```

**Test file naming:**
- Unit tests: `*_test.go`
- Place in same package as code being tested

#### Commit Messages

Use clear, descriptive commit messages:

**Good:**
```
Add union type support for GraphQL generation

- Implement @oneOf directive for unions
- Add test cases for union generation
- Update documentation with union examples
```

**Bad:**
```
Fix stuff
WIP
Update
```

**Format:**
```
<type>: <short summary>

<optional longer description>

<optional footer>
```

**Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `test:` - Test additions/changes
- `refactor:` - Code refactoring
- `perf:` - Performance improvements
- `chore:` - Build/tooling changes

#### Pull Requests

1. **Push your branch:**
```bash
git push origin feature/your-feature-name
```

2. **Create pull request** on GitHub

3. **PR description should include:**
   - What changes were made
   - Why the changes were needed
   - How to test the changes
   - Related issues (if any)

**PR template:**
```markdown
## Description
Brief description of what this PR does.

## Motivation
Why is this change needed?

## Changes
- Change 1
- Change 2
- Change 3

## Testing
How to test these changes:
1. Step 1
2. Step 2
3. Expected result

## Checklist
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] Examples updated (if applicable)
- [ ] All tests pass
- [ ] Test coverage above 90%
```

4. **Respond to review feedback**
   - Address all comments
   - Push additional commits to your branch
   - Discuss any disagreements constructively

### Contributing Documentation

Documentation improvements are always welcome!

**To update docs:**

1. Edit files in `docs/github-site/`
2. Test locally with Jekyll (see docs/github-site/README.md)
3. Submit PR with documentation changes

**Documentation standards:**
- Clear, concise writing
- Code examples for all features
- Cross-references to related sections
- Tested code snippets

### Adding Examples

New examples help users understand features:

1. Create directory: `examples/your-example/`
2. Add schema file: `your-example.typemux`
3. Add README: `README.md` explaining the example
4. Generate output: `make examples`
5. Add to `examples.md` documentation

**Example structure:**
```
examples/your-example/
├── schema.typemux      # The schema
├── annotations.yaml    # Annotations (if needed)
├── README.md           # Explanation
└── generated/          # Generated output (for reference)
```

## Development Workflow

### Project Structure

```
typemux/
├── internal/
│   ├── ast/            # Abstract syntax tree definitions
│   ├── lexer/          # Tokenization
│   ├── parser/         # Parsing
│   ├── generator/      # Code generators (graphql, proto, openapi)
│   └── annotations/    # YAML annotation handling
├── examples/           # Usage examples
├── docs/               # Documentation
├── main.go             # CLI entry point
├── go.mod              # Go module definition
└── Makefile            # Build automation
```

### Adding a New Generator

To add support for a new output format:

1. **Create generator file:**
```go
// internal/generator/newformat.go
package generator

type NewFormatGenerator struct{}

func (g *NewFormatGenerator) Generate(schema *ast.Schema) string {
    // Implementation
}
```

2. **Add tests:**
```go
// internal/generator/newformat_test.go
package generator

func TestNewFormatGenerator_BasicType(t *testing.T) {
    // Test implementation
}
```

3. **Integrate with CLI:**
```go
// main.go
case "newformat":
    gen := &generator.NewFormatGenerator{}
    output := gen.Generate(schema)
    // Write to file
```

4. **Update documentation:**
   - Add to reference.md
   - Add example to examples.md
   - Update tutorial.md if needed

### Testing Checklist

Before submitting PR:

- [ ] All existing tests pass: `go test ./...`
- [ ] New tests added for new functionality
- [ ] Test coverage above 90%: `go test ./... -cover`
- [ ] Manual testing completed
- [ ] Examples still generate correctly: `make examples`
- [ ] Documentation updated
- [ ] Code formatted: `go fmt ./...`

### Release Process

(For maintainers)

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create git tag: `git tag v1.0.0`
4. Push tag: `git push origin v1.0.0`
5. Create GitHub release with notes
6. Update documentation website

## Getting Help

- **Questions:** Create a GitHub issue with the "question" label
- **Discussions:** Use GitHub Discussions (if enabled)
- **Bug reports:** Create an issue with detailed information

## Recognition

Contributors will be recognized in:
- CONTRIBUTORS.md file
- Release notes
- GitHub contributors page

Thank you for contributing to TypeMUX!
