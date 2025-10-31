package annotations

// GetBuiltinAnnotations returns all built-in TypeMUX annotations
func GetBuiltinAnnotations() *AnnotationRegistry {
	registry := NewAnnotationRegistry()

	// Schema-level annotations
	registry.Register(&AnnotationMetadata{
		Name:        "@typemux",
		Scope:       []string{"schema"},
		Formats:     []string{"all"},
		Description: "Specifies the TypeMUX IDL format version",
		Parameters: []ParameterMetadata{
			{
				Name:        "version",
				Type:        "string",
				Required:    true,
				Description: "Version string (e.g., '1.0.0')",
			},
		},
		Examples: []string{`@typemux("1.0.0")`},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@version",
		Scope:       []string{"schema"},
		Formats:     []string{"all"},
		Description: "Specifies the schema/API version",
		Parameters: []ParameterMetadata{
			{
				Name:        "version",
				Type:        "string",
				Required:    true,
				Description: "Version string (e.g., '2.1.0')",
			},
		},
		Examples: []string{`@version("2.1.0")`},
	})

	// Namespace-level annotations
	registry.Register(&AnnotationMetadata{
		Name:        "@proto.option",
		Scope:       []string{"namespace", "type", "enum", "union"},
		Formats:     []string{"proto"},
		Description: "Adds Protobuf file-level or message-level options",
		Parameters: []ParameterMetadata{
			{
				Name:        "option",
				Type:        "string",
				Required:    true,
				Description: "Protobuf option declaration",
			},
		},
		Examples: []string{
			`@proto.option(go_package="github.com/example/api")`,
			`@proto.option([packed = false])`,
		},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@graphql.directive",
		Scope:       []string{"namespace", "type", "enum", "union", "field"},
		Formats:     []string{"graphql"},
		Description: "Adds GraphQL directives to schema elements",
		Parameters: []ParameterMetadata{
			{
				Name:        "directive",
				Type:        "string",
				Required:    true,
				Description: "GraphQL directive (e.g., @key, @external)",
			},
		},
		Examples: []string{
			`@graphql.directive(@key(fields: "id"))`,
			`@graphql.directive(@external)`,
		},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@go.package",
		Scope:       []string{"namespace"},
		Formats:     []string{"go"},
		Description: "Overrides the Go package name for generated code",
		Parameters: []ParameterMetadata{
			{
				Name:        "package",
				Type:        "string",
				Required:    true,
				Description: "Go package name",
			},
		},
		Examples: []string{`@go.package("mypackage")`},
	})

	// Type/Enum/Union-level annotations
	registry.Register(&AnnotationMetadata{
		Name:        "@proto.name",
		Scope:       []string{"type", "enum", "union", "field"},
		Formats:     []string{"proto"},
		Description: "Overrides the Protobuf name for the element",
		Parameters: []ParameterMetadata{
			{
				Name:        "name",
				Type:        "string",
				Required:    true,
				Description: "Protobuf name",
			},
		},
		Examples: []string{
			`@proto.name("UserV2")`,
			`@proto.name("user_id")`,
		},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@graphql.name",
		Scope:       []string{"type", "enum", "union", "field"},
		Formats:     []string{"graphql"},
		Description: "Overrides the GraphQL name for the element",
		Parameters: []ParameterMetadata{
			{
				Name:        "name",
				Type:        "string",
				Required:    true,
				Description: "GraphQL name",
			},
		},
		Examples: []string{
			`@graphql.name("UserAccount")`,
			`@graphql.name("userId")`,
		},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@openapi.name",
		Scope:       []string{"type", "enum", "union", "field"},
		Formats:     []string{"openapi"},
		Description: "Overrides the OpenAPI schema or property name",
		Parameters: []ParameterMetadata{
			{
				Name:        "name",
				Type:        "string",
				Required:    true,
				Description: "OpenAPI name",
			},
		},
		Examples: []string{
			`@openapi.name("UserProfile")`,
			`@openapi.name("user_id")`,
		},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@openapi.extension",
		Scope:       []string{"type", "enum", "union", "field"},
		Formats:     []string{"openapi"},
		Description: "Adds OpenAPI vendor extensions (x-* fields)",
		Parameters: []ParameterMetadata{
			{
				Name:        "extension",
				Type:        "object",
				Required:    true,
				Description: "JSON object with vendor extensions",
			},
		},
		Examples: []string{`@openapi.extension({"x-internal": true, "x-format": "currency"})`},
	})

	// Field-level annotations
	registry.Register(&AnnotationMetadata{
		Name:        "@required",
		Scope:       []string{"field"},
		Formats:     []string{"all"},
		Description: "Marks a field as required/non-nullable",
		Parameters:  []ParameterMetadata{},
		Examples:    []string{`id: string @required`},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@default",
		Scope:       []string{"field"},
		Formats:     []string{"all"},
		Description: "Sets a default value for the field",
		Parameters: []ParameterMetadata{
			{
				Name:        "value",
				Type:        "any",
				Required:    true,
				Description: "Default value (string, number, or boolean)",
			},
		},
		Examples: []string{
			`age: int32 @default(0)`,
			`active: bool @default(true)`,
			`status: string @default("pending")`,
		},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@exclude",
		Scope:       []string{"field"},
		Formats:     []string{"all"},
		Description: "Excludes field from specific output formats",
		Parameters: []ParameterMetadata{
			{
				Name:        "formats",
				Type:        "list",
				Required:    true,
				Description: "Comma-separated list of formats to exclude from",
			},
		},
		Examples: []string{
			`internal: string @exclude(graphql,openapi)`,
			`debug: string @exclude(proto)`,
		},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@only",
		Scope:       []string{"field"},
		Formats:     []string{"all"},
		Description: "Includes field only in specific output formats",
		Parameters: []ParameterMetadata{
			{
				Name:        "formats",
				Type:        "list",
				Required:    true,
				Description: "Comma-separated list of formats to include in",
			},
		},
		Examples: []string{
			`protoField: string @only(proto)`,
			`graphqlField: string @only(graphql)`,
		},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@deprecated",
		Scope:       []string{"field", "type", "enum", "method"},
		Formats:     []string{"all"},
		Description: "Marks element as deprecated with version information",
		Parameters: []ParameterMetadata{
			{
				Name:        "reason",
				Type:        "string",
				Required:    true,
				Description: "Reason for deprecation",
			},
			{
				Name:        "since",
				Type:        "string",
				Required:    false,
				Description: "Version when deprecated",
			},
			{
				Name:        "removed",
				Type:        "string",
				Required:    false,
				Description: "Version when it will be removed",
			},
		},
		Examples: []string{
			`@deprecated("Use fullName instead")`,
			`@deprecated("Use email field", since="2.0.0", removed="3.0.0")`,
		},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@since",
		Scope:       []string{"field", "type", "enum", "method"},
		Formats:     []string{"all"},
		Description: "Marks when an element was added to the schema",
		Parameters: []ParameterMetadata{
			{
				Name:        "version",
				Type:        "string",
				Required:    true,
				Description: "Version when element was added",
			},
		},
		Examples: []string{`@since("2.0.0")`},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@validate",
		Scope:       []string{"field"},
		Formats:     []string{"all"},
		Description: "Defines validation rules for the field",
		Parameters: []ParameterMetadata{
			{
				Name:        "format",
				Type:        "string",
				Required:    false,
				Description: "String format (email, uuid, uri, etc.)",
				ValidValues: []string{"email", "uuid", "uri", "date", "time", "datetime"},
			},
			{
				Name:        "pattern",
				Type:        "string",
				Required:    false,
				Description: "Regular expression pattern",
			},
			{
				Name:        "minLength",
				Type:        "number",
				Required:    false,
				Description: "Minimum string length",
			},
			{
				Name:        "maxLength",
				Type:        "number",
				Required:    false,
				Description: "Maximum string length",
			},
			{
				Name:        "min",
				Type:        "number",
				Required:    false,
				Description: "Minimum numeric value",
			},
			{
				Name:        "max",
				Type:        "number",
				Required:    false,
				Description: "Maximum numeric value",
			},
			{
				Name:        "exclusiveMin",
				Type:        "boolean",
				Required:    false,
				Description: "Whether min is exclusive",
			},
			{
				Name:        "exclusiveMax",
				Type:        "boolean",
				Required:    false,
				Description: "Whether max is exclusive",
			},
			{
				Name:        "multipleOf",
				Type:        "number",
				Required:    false,
				Description: "Number must be multiple of this value",
			},
			{
				Name:        "minItems",
				Type:        "number",
				Required:    false,
				Description: "Minimum array length",
			},
			{
				Name:        "maxItems",
				Type:        "number",
				Required:    false,
				Description: "Maximum array length",
			},
			{
				Name:        "uniqueItems",
				Type:        "boolean",
				Required:    false,
				Description: "Whether array items must be unique",
			},
			{
				Name:        "enum",
				Type:        "list",
				Required:    false,
				Description: "List of allowed values",
			},
		},
		Examples: []string{
			`@validate(format="email", maxLength=100)`,
			`@validate(min=0, max=150)`,
			`@validate(pattern="^[A-Z]{3}$")`,
		},
	})

	// Method-level annotations
	registry.Register(&AnnotationMetadata{
		Name:        "@http.method",
		Scope:       []string{"method"},
		Formats:     []string{"openapi"},
		Description: "Specifies the HTTP method for REST API mapping",
		Parameters: []ParameterMetadata{
			{
				Name:        "method",
				Type:        "string",
				Required:    true,
				Description: "HTTP method",
				ValidValues: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
			},
		},
		Examples: []string{
			`@http.method(GET)`,
			`@http.method(POST)`,
		},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@http.path",
		Scope:       []string{"method"},
		Formats:     []string{"openapi"},
		Description: "Specifies the URL path template for REST API mapping",
		Parameters: []ParameterMetadata{
			{
				Name:        "path",
				Type:        "string",
				Required:    true,
				Description: "URL path template with parameters in {braces}",
			},
		},
		Examples: []string{
			`@http.path("/api/v1/users")`,
			`@http.path("/api/v1/users/{id}")`,
		},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@http.success",
		Scope:       []string{"method"},
		Formats:     []string{"openapi"},
		Description: "Specifies additional success HTTP status codes beyond 200",
		Parameters: []ParameterMetadata{
			{
				Name:        "codes",
				Type:        "list",
				Required:    true,
				Description: "Comma-separated list of HTTP status codes",
			},
		},
		Examples: []string{
			`@http.success(201)`,
			`@http.success(201,204)`,
		},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@http.errors",
		Scope:       []string{"method"},
		Formats:     []string{"openapi"},
		Description: "Specifies expected error HTTP status codes",
		Parameters: []ParameterMetadata{
			{
				Name:        "codes",
				Type:        "list",
				Required:    true,
				Description: "Comma-separated list of HTTP status codes",
			},
		},
		Examples: []string{
			`@http.errors(404,500)`,
			`@http.errors(400,404,409,500)`,
		},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@graphql",
		Scope:       []string{"method"},
		Formats:     []string{"graphql"},
		Description: "Specifies the GraphQL operation type",
		Parameters: []ParameterMetadata{
			{
				Name:        "operation",
				Type:        "string",
				Required:    true,
				Description: "GraphQL operation type",
				ValidValues: []string{"query", "mutation", "subscription"},
			},
		},
		Examples: []string{
			`@graphql(query)`,
			`@graphql(mutation)`,
			`@graphql(subscription)`,
		},
	})

	// JSON serialization annotations
	registry.Register(&AnnotationMetadata{
		Name:        "@json.name",
		Scope:       []string{"field"},
		Formats:     []string{"all"},
		Description: "Overrides the JSON field name for serialization",
		Parameters: []ParameterMetadata{
			{
				Name:        "name",
				Type:        "string",
				Required:    true,
				Description: "JSON field name override",
			},
		},
		Examples: []string{
			`userId: string @json.name("user_id")`,
			`createdAt: timestamp @json.name("created_at")`,
		},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@json.nullable",
		Scope:       []string{"field"},
		Formats:     []string{"all"},
		Description: "Marks a field as explicitly nullable (can be null in JSON)",
		Parameters:  []ParameterMetadata{},
		Examples: []string{
			`middleName: string @json.nullable`,
			`phoneNumber: string @json.nullable`,
		},
	})

	registry.Register(&AnnotationMetadata{
		Name:        "@json.omitempty",
		Scope:       []string{"field"},
		Formats:     []string{"all"},
		Description: "Omits the field from JSON serialization if it has a zero/empty value",
		Parameters:  []ParameterMetadata{},
		Examples: []string{
			`description: string @json.omitempty`,
			`metadata: map<string, string> @json.omitempty`,
		},
	})

	return registry
}
