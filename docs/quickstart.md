# Quick Start

Get started with TypeMUX in 5 minutes.

## Installation

### Prerequisites

- Go 1.21 or later

### Build from Source

```bash
git clone https://github.com/rasmartins/typemux.git
cd typemux
go build -o typemux
```

### Using Go Install

```bash
go install github.com/rasmartins/typemux@latest
```

## Your First Schema

Create a file named `blog.typemux`:

```typemux
/// A blog post with title, content, and author
type Post {
  id: string @required
  title: string @required
  content: string @required
  author: string @required
  published: bool @default("false")
  createdAt: timestamp @required
}

/// Blog post status
enum PostStatus {
  DRAFT
  PUBLISHED
  ARCHIVED
}

/// Request to get a post by ID
type GetPostRequest {
  id: string @required
}

/// Request to create a new post
type CreatePostRequest {
  title: string @required
  content: string @required
  author: string @required
}

/// Blog service for managing posts
service BlogService {
  /// Get a post by its ID
  rpc GetPost(GetPostRequest) returns (Post)
    @http(GET)
    @path("/api/v1/posts/{id}")
    @graphql(query)

  /// Create a new blog post
  rpc CreatePost(CreatePostRequest) returns (Post)
    @http(POST)
    @path("/api/v1/posts")
    @graphql(mutation)
}
```

## Generate Schemas

### Generate All Formats

```bash
./typemux -input blog.typemux -output ./generated
```

This creates:
- `generated/schema.graphql` - GraphQL schema
- `generated/schema.proto` - Protocol Buffers definition
- `generated/openapi.yaml` - OpenAPI specification

### Generate Specific Format

```bash
# Only GraphQL
./typemux -input blog.typemux -format graphql -output ./generated

# Only Protobuf
./typemux -input blog.typemux -format protobuf -output ./generated

# Only OpenAPI
./typemux -input blog.typemux -format openapi -output ./generated
```

## Inspect the Output

### GraphQL Schema

```bash
cat generated/schema.graphql
```

You'll see:

```graphql
"""
A blog post with title, content, and author
"""
type Post {
  id: String!
  title: String!
  content: String!
  author: String!
  published: Boolean
  createdAt: String!
}

"""
Blog post status
"""
enum PostStatus {
  DRAFT
  PUBLISHED
  ARCHIVED
}

type GetPostRequestInput {
  id: String!
}

type CreatePostRequestInput {
  title: String!
  content: String!
  author: String!
}

type Query {
  """
  Get a post by its ID
  """
  GetPost(input: GetPostRequestInput!): Post
}

type Mutation {
  """
  Create a new blog post
  """
  CreatePost(input: CreatePostRequestInput!): Post
}
```

### Protocol Buffers

```bash
cat generated/schema.proto
```

You'll see:

```protobuf
syntax = "proto3";

import "google/protobuf/timestamp.proto";

// A blog post with title, content, and author
message Post {
  string id = 1;
  string title = 2;
  string content = 3;
  string author = 4;
  bool published = 5;
  google.protobuf.Timestamp createdAt = 6;
}

// Blog post status
enum PostStatus {
  DRAFT = 0;
  PUBLISHED = 1;
  ARCHIVED = 2;
}

// Request to get a post by ID
message GetPostRequest {
  string id = 1;
}

// Request to create a new post
message CreatePostRequest {
  string title = 1;
  string content = 2;
  string author = 3;
}

// Blog service for managing posts
service BlogService {
  // Get a post by its ID
  rpc GetPost(GetPostRequest) returns (Post);
  // Create a new blog post
  rpc CreatePost(CreatePostRequest) returns (Post);
}
```

### OpenAPI Specification

```bash
cat generated/openapi.yaml
```

You'll see:

```yaml
openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths:
  /api/v1/posts/{id}:
    get:
      operationId: GetPost
      description: Get a post by its ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Post'
  /api/v1/posts:
    post:
      operationId: CreatePost
      description: Create a new blog post
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreatePostRequest'
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Post'
components:
  schemas:
    Post:
      type: object
      description: A blog post with title, content, and author
      required:
        - id
        - title
        - content
        - author
        - createdAt
      properties:
        id:
          type: string
        title:
          type: string
        content:
          type: string
        author:
          type: string
        published:
          type: boolean
          default: false
        createdAt:
          type: string
          format: date-time
    PostStatus:
      type: string
      description: Blog post status
      enum:
        - DRAFT
        - PUBLISHED
        - ARCHIVED
    GetPostRequest:
      type: object
      description: Request to get a post by ID
      required:
        - id
      properties:
        id:
          type: string
    CreatePostRequest:
      type: object
      description: Request to create a new post
      required:
        - title
        - content
        - author
      properties:
        title:
          type: string
        content:
          type: string
        author:
          type: string
```

## Next Steps

- [Learn more with the Tutorial](tutorial.md)
- [Explore language features in the Reference](reference.md)
- [Check out Examples](examples.md)
- [Configure advanced features](configuration.md)

## Common Issues

### Command not found

If you get "command not found", make sure the binary is in your PATH or use the full path:

```bash
# Add to PATH
export PATH=$PATH:/path/to/typemux

# Or use full path
/path/to/typemux -input schema.typemux -output ./generated
```

### Import errors in generated proto files

Make sure to include the Google protobuf imports when compiling:

```bash
protoc --proto_path=/usr/local/include \
       --proto_path=. \
       --go_out=. \
       schema.proto
```

### YAML annotation validation errors

If using external YAML annotations, ensure all referenced types and fields exist in your schema. See [Configuration Guide](configuration.md#yaml-annotations) for details.
