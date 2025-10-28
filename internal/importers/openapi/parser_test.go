package openapi

import (
	"testing"
)

func TestParseBasicOpenAPI(t *testing.T) {
	input := `{
  "openapi": "3.0.0",
  "info": {
    "title": "Test API",
    "version": "1.0.0",
    "description": "A test API"
  },
  "paths": {}
}`

	parser := NewParser([]byte(input))
	spec, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.OpenAPI != "3.0.0" {
		t.Errorf("expected OpenAPI version %q, got %q", "3.0.0", spec.OpenAPI)
	}

	if spec.Info.Title != "Test API" {
		t.Errorf("expected title %q, got %q", "Test API", spec.Info.Title)
	}

	if spec.Info.Version != "1.0.0" {
		t.Errorf("expected version %q, got %q", "1.0.0", spec.Info.Version)
	}

	if spec.Info.Description != "A test API" {
		t.Errorf("expected description %q, got %q", "A test API", spec.Info.Description)
	}
}

func TestParseYAMLFormat(t *testing.T) {
	input := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}`

	parser := NewParser([]byte(input))
	spec, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.Info.Title != "Test API" {
		t.Errorf("expected title %q, got %q", "Test API", spec.Info.Title)
	}
}

func TestParseSchemaObject(t *testing.T) {
	input := `{
  "openapi": "3.0.0",
  "info": {
    "title": "Test",
    "version": "1.0.0"
  },
  "paths": {},
  "components": {
    "schemas": {
      "User": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string"
          },
          "name": {
            "type": "string"
          },
          "age": {
            "type": "integer"
          }
        },
        "required": ["id", "name"]
      }
    }
  }
}`

	parser := NewParser([]byte(input))
	spec, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.Components == nil {
		t.Fatal("expected components to be set")
	}

	if len(spec.Components.Schemas) != 1 {
		t.Fatalf("expected 1 schema, got %d", len(spec.Components.Schemas))
	}

	user, ok := spec.Components.Schemas["User"]
	if !ok {
		t.Fatal("expected User schema")
	}

	if user.Type != "object" {
		t.Errorf("expected type %q, got %q", "object", user.Type)
	}

	if len(user.Properties) != 3 {
		t.Fatalf("expected 3 properties, got %d", len(user.Properties))
	}

	if len(user.Required) != 2 {
		t.Fatalf("expected 2 required fields, got %d", len(user.Required))
	}
}

func TestParseEnum(t *testing.T) {
	input := `{
  "openapi": "3.0.0",
  "info": {
    "title": "Test",
    "version": "1.0.0"
  },
  "paths": {},
  "components": {
    "schemas": {
      "Status": {
        "type": "string",
        "enum": ["active", "inactive", "pending"]
      }
    }
  }
}`

	parser := NewParser([]byte(input))
	spec, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status, ok := spec.Components.Schemas["Status"]
	if !ok {
		t.Fatal("expected Status schema")
	}

	if len(status.Enum) != 3 {
		t.Fatalf("expected 3 enum values, got %d", len(status.Enum))
	}

	expectedValues := []string{"active", "inactive", "pending"}
	for i, expected := range expectedValues {
		if status.Enum[i] != expected {
			t.Errorf("expected enum value %q, got %q", expected, status.Enum[i])
		}
	}
}

func TestParseArray(t *testing.T) {
	input := `{
  "openapi": "3.0.0",
  "info": {
    "title": "Test",
    "version": "1.0.0"
  },
  "paths": {},
  "components": {
    "schemas": {
      "UserList": {
        "type": "array",
        "items": {
          "type": "string"
        }
      }
    }
  }
}`

	parser := NewParser([]byte(input))
	spec, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	userList, ok := spec.Components.Schemas["UserList"]
	if !ok {
		t.Fatal("expected UserList schema")
	}

	if userList.Type != "array" {
		t.Errorf("expected type %q, got %q", "array", userList.Type)
	}

	if userList.Items == nil {
		t.Fatal("expected items to be set")
	}

	if userList.Items.Type != "string" {
		t.Errorf("expected items type %q, got %q", "string", userList.Items.Type)
	}
}

func TestParsePaths(t *testing.T) {
	input := `{
  "openapi": "3.0.0",
  "info": {
    "title": "Test",
    "version": "1.0.0"
  },
  "paths": {
    "/users": {
      "get": {
        "operationId": "listUsers",
        "summary": "List all users",
        "responses": {
          "200": {
            "description": "Success"
          }
        }
      },
      "post": {
        "operationId": "createUser",
        "summary": "Create a user",
        "responses": {
          "201": {
            "description": "Created"
          }
        }
      }
    }
  }
}`

	parser := NewParser([]byte(input))
	spec, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(spec.Paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(spec.Paths))
	}

	usersPath, ok := spec.Paths["/users"]
	if !ok {
		t.Fatal("expected /users path")
	}

	if usersPath.Get == nil {
		t.Fatal("expected GET operation")
	}

	if usersPath.Get.OperationID != "listUsers" {
		t.Errorf("expected operation ID %q, got %q", "listUsers", usersPath.Get.OperationID)
	}

	if usersPath.Post == nil {
		t.Fatal("expected POST operation")
	}

	if usersPath.Post.OperationID != "createUser" {
		t.Errorf("expected operation ID %q, got %q", "createUser", usersPath.Post.OperationID)
	}
}

func TestParseOneOf(t *testing.T) {
	input := `{
  "openapi": "3.0.0",
  "info": {
    "title": "Test",
    "version": "1.0.0"
  },
  "components": {
    "schemas": {
      "Dog": {
        "type": "object",
        "properties": {
          "breed": {"type": "string"}
        }
      },
      "Cat": {
        "type": "object",
        "properties": {
          "meow": {"type": "boolean"}
        }
      },
      "Pet": {
        "oneOf": [
          {"$ref": "#/components/schemas/Dog"},
          {"$ref": "#/components/schemas/Cat"}
        ]
      }
    }
  }
}`

	parser := NewParser([]byte(input))
	spec, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.Components == nil {
		t.Fatal("expected components")
	}

	petSchema, ok := spec.Components.Schemas["Pet"]
	if !ok {
		t.Fatal("expected Pet schema")
	}

	if len(petSchema.OneOf) != 2 {
		t.Fatalf("expected 2 oneOf schemas, got %d", len(petSchema.OneOf))
	}

	// Check that oneOf references are present
	if petSchema.OneOf[0].Ref != "#/components/schemas/Dog" {
		t.Errorf("expected Dog ref, got %q", petSchema.OneOf[0].Ref)
	}

	if petSchema.OneOf[1].Ref != "#/components/schemas/Cat" {
		t.Errorf("expected Cat ref, got %q", petSchema.OneOf[1].Ref)
	}
}

func TestParseAnyOfAndAllOf(t *testing.T) {
	input := `{
  "openapi": "3.0.0",
  "info": {
    "title": "Test",
    "version": "1.0.0"
  },
  "components": {
    "schemas": {
      "FlexibleValue": {
        "anyOf": [
          {"type": "string"},
          {"type": "integer"}
        ]
      },
      "CombinedType": {
        "allOf": [
          {"$ref": "#/components/schemas/Base"},
          {"type": "object", "properties": {"extra": {"type": "string"}}}
        ]
      }
    }
  }
}`

	parser := NewParser([]byte(input))
	spec, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.Components == nil {
		t.Fatal("expected components")
	}

	// Check anyOf
	flexibleSchema, ok := spec.Components.Schemas["FlexibleValue"]
	if !ok {
		t.Fatal("expected FlexibleValue schema")
	}

	if len(flexibleSchema.AnyOf) != 2 {
		t.Fatalf("expected 2 anyOf schemas, got %d", len(flexibleSchema.AnyOf))
	}

	if flexibleSchema.AnyOf[0].Type != "string" {
		t.Errorf("expected string type, got %q", flexibleSchema.AnyOf[0].Type)
	}

	if flexibleSchema.AnyOf[1].Type != "integer" {
		t.Errorf("expected integer type, got %q", flexibleSchema.AnyOf[1].Type)
	}

	// Check allOf
	combinedSchema, ok := spec.Components.Schemas["CombinedType"]
	if !ok {
		t.Fatal("expected CombinedType schema")
	}

	if len(combinedSchema.AllOf) != 2 {
		t.Fatalf("expected 2 allOf schemas, got %d", len(combinedSchema.AllOf))
	}

	if combinedSchema.AllOf[0].Ref != "#/components/schemas/Base" {
		t.Errorf("expected Base ref, got %q", combinedSchema.AllOf[0].Ref)
	}

	if len(combinedSchema.AllOf[1].Properties) != 1 {
		t.Errorf("expected 1 property in second allOf schema, got %d", len(combinedSchema.AllOf[1].Properties))
	}
}

func TestParseParameters(t *testing.T) {
	input := `{
  "openapi": "3.0.0",
  "info": {
    "title": "Test",
    "version": "1.0.0"
  },
  "paths": {
    "/users/{id}": {
      "get": {
        "operationId": "getUser",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string"
            }
          },
          {
            "name": "limit",
            "in": "query",
            "required": false,
            "schema": {
              "type": "integer"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Success"
          }
        }
      }
    }
  }
}`

	parser := NewParser([]byte(input))
	spec, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	userPath, ok := spec.Paths["/users/{id}"]
	if !ok {
		t.Fatal("expected /users/{id} path")
	}

	if len(userPath.Get.Parameters) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(userPath.Get.Parameters))
	}

	idParam := userPath.Get.Parameters[0]
	if idParam.Name != "id" {
		t.Errorf("expected parameter name %q, got %q", "id", idParam.Name)
	}

	if idParam.In != "path" {
		t.Errorf("expected parameter in %q, got %q", "path", idParam.In)
	}

	if !idParam.Required {
		t.Error("expected id parameter to be required")
	}

	limitParam := userPath.Get.Parameters[1]
	if limitParam.Name != "limit" {
		t.Errorf("expected parameter name %q, got %q", "limit", limitParam.Name)
	}

	if limitParam.In != "query" {
		t.Errorf("expected parameter in %q, got %q", "query", limitParam.In)
	}

	if limitParam.Required {
		t.Error("expected limit parameter to be optional")
	}
}

func TestParseRequestBody(t *testing.T) {
	input := `{
  "openapi": "3.0.0",
  "info": {
    "title": "Test",
    "version": "1.0.0"
  },
  "paths": {
    "/users": {
      "post": {
        "operationId": "createUser",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "name": {
                    "type": "string"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Created"
          }
        }
      }
    }
  }
}`

	parser := NewParser([]byte(input))
	spec, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	usersPath := spec.Paths["/users"]
	if usersPath.Post.RequestBody == nil {
		t.Fatal("expected request body")
	}

	if !usersPath.Post.RequestBody.Required {
		t.Error("expected request body to be required")
	}

	if len(usersPath.Post.RequestBody.Content) == 0 {
		t.Fatal("expected request body content")
	}

	jsonContent, ok := usersPath.Post.RequestBody.Content["application/json"]
	if !ok {
		t.Fatal("expected application/json content")
	}

	if jsonContent.Schema == nil {
		t.Fatal("expected schema in content")
	}
}

func TestParseResponses(t *testing.T) {
	input := `{
  "openapi": "3.0.0",
  "info": {
    "title": "Test",
    "version": "1.0.0"
  },
  "paths": {
    "/users": {
      "get": {
        "operationId": "listUsers",
        "responses": {
          "200": {
            "description": "Success",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/User"
                  }
                }
              }
            }
          },
          "400": {
            "description": "Bad Request"
          }
        }
      }
    }
  }
}`

	parser := NewParser([]byte(input))
	spec, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	usersPath := spec.Paths["/users"]
	responses := usersPath.Get.Responses

	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}

	success, ok := responses["200"]
	if !ok {
		t.Fatal("expected 200 response")
	}

	if success.Description != "Success" {
		t.Errorf("expected description %q, got %q", "Success", success.Description)
	}

	if len(success.Content) == 0 {
		t.Fatal("expected response content")
	}
}
