package annotations

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadYAMLAnnotations_Valid(t *testing.T) {
	// Create a temporary YAML file
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "annotations.yaml")

	yamlContent := `
types:
  User:
    proto:
      name: "UserV2"
    graphql:
      name: "UserAccount"
    openapi:
      name: "UserProfile"
    fields:
      email:
        required: true
        openapi:
          extension: '{"x-format": "email"}'

services:
  UserService:
    methods:
      GetUser:
        http: "GET"
        path: "/api/v1/users/{id}"
        graphql: "query"
        errors: [404, 500]
`

	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Load annotations
	annotations, err := LoadYAMLAnnotations(yamlFile)
	if err != nil {
		t.Fatalf("LoadYAMLAnnotations failed: %v", err)
	}

	// Verify type annotations
	if annotations.Types == nil {
		t.Fatal("Types map is nil")
	}

	userAnnotations, ok := annotations.Types["User"]
	if !ok {
		t.Fatal("User type not found in annotations")
	}

	if userAnnotations.Proto == nil || userAnnotations.Proto.Name != "UserV2" {
		t.Errorf("Expected Proto.Name 'UserV2', got '%v'", userAnnotations.Proto)
	}
	if userAnnotations.GraphQL == nil || userAnnotations.GraphQL.Name != "UserAccount" {
		t.Errorf("Expected GraphQL.Name 'UserAccount', got '%v'", userAnnotations.GraphQL)
	}
	if userAnnotations.OpenAPI == nil || userAnnotations.OpenAPI.Name != "UserProfile" {
		t.Errorf("Expected OpenAPI.Name 'UserProfile', got '%v'", userAnnotations.OpenAPI)
	}

	// Verify field annotations
	if userAnnotations.Fields == nil {
		t.Fatal("Fields map is nil")
	}

	emailAnnotations, ok := userAnnotations.Fields["email"]
	if !ok {
		t.Fatal("email field not found in annotations")
	}

	if !emailAnnotations.Required {
		t.Error("Expected email field to be required")
	}
	if emailAnnotations.OpenAPI == nil || emailAnnotations.OpenAPI.Extension != `{"x-format": "email"}` {
		t.Errorf("Expected OpenAPI.Extension, got '%v'", emailAnnotations.OpenAPI)
	}

	// Verify service annotations
	if annotations.Services == nil {
		t.Fatal("Services map is nil")
	}

	serviceAnnotations, ok := annotations.Services["UserService"]
	if !ok {
		t.Fatal("UserService not found in annotations")
	}

	methodAnnotations, ok := serviceAnnotations.Methods["GetUser"]
	if !ok {
		t.Fatal("GetUser method not found in annotations")
	}

	if methodAnnotations.HTTP != "GET" {
		t.Errorf("Expected HTTP method 'GET', got '%s'", methodAnnotations.HTTP)
	}
	if methodAnnotations.Path != "/api/v1/users/{id}" {
		t.Errorf("Expected path '/api/v1/users/{id}', got '%s'", methodAnnotations.Path)
	}
	if methodAnnotations.GraphQL != "query" {
		t.Errorf("Expected GraphQL type 'query', got '%s'", methodAnnotations.GraphQL)
	}
	if len(methodAnnotations.Errors) != 2 {
		t.Errorf("Expected 2 error codes, got %d", len(methodAnnotations.Errors))
	}
}

func TestLoadYAMLAnnotations_InvalidFile(t *testing.T) {
	_, err := LoadYAMLAnnotations("/nonexistent/file.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestLoadYAMLAnnotations_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "invalid.yaml")

	invalidYAML := `
types:
  User:
    proto.name: [this is invalid
`

	if err := os.WriteFile(yamlFile, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := LoadYAMLAnnotations(yamlFile)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestMergeYAMLAnnotations_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "annotations.yaml")

	yamlContent := `
types:
  User:
    proto:
      name: "UserV2"
`

	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	annotations, err := MergeYAMLAnnotations([]string{yamlFile})
	if err != nil {
		t.Fatalf("MergeYAMLAnnotations failed: %v", err)
	}

	if annotations.Types["User"].Proto == nil || annotations.Types["User"].Proto.Name != "UserV2" {
		t.Errorf("Expected Proto.Name 'UserV2', got '%v'", annotations.Types["User"].Proto)
	}
}

func TestMergeYAMLAnnotations_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create base file
	baseFile := filepath.Join(tmpDir, "base.yaml")
	baseContent := `
types:
  User:
    proto:
      name: "UserV1"
    graphql:
      name: "User"
  Product:
    proto:
      name: "ProductV1"
`
	if err := os.WriteFile(baseFile, []byte(baseContent), 0644); err != nil {
		t.Fatalf("Failed to create base file: %v", err)
	}

	// Create override file
	overrideFile := filepath.Join(tmpDir, "override.yaml")
	overrideContent := `
types:
  User:
    proto:
      name: "UserV2"
    openapi:
      name: "UserProfile"
`
	if err := os.WriteFile(overrideFile, []byte(overrideContent), 0644); err != nil {
		t.Fatalf("Failed to create override file: %v", err)
	}

	// Merge files
	annotations, err := MergeYAMLAnnotations([]string{baseFile, overrideFile})
	if err != nil {
		t.Fatalf("MergeYAMLAnnotations failed: %v", err)
	}

	// Verify User was overridden
	userAnnotations := annotations.Types["User"]
	if userAnnotations.Proto == nil || userAnnotations.Proto.Name != "UserV2" {
		t.Errorf("Expected Proto.Name 'UserV2' (overridden), got '%v'", userAnnotations.Proto)
	}
	if userAnnotations.GraphQL == nil || userAnnotations.GraphQL.Name != "User" {
		t.Errorf("Expected GraphQL.Name 'User' (from base), got '%v'", userAnnotations.GraphQL)
	}
	if userAnnotations.OpenAPI == nil || userAnnotations.OpenAPI.Name != "UserProfile" {
		t.Errorf("Expected OpenAPI.Name 'UserProfile' (from override), got '%v'", userAnnotations.OpenAPI)
	}

	// Verify Product was preserved
	productAnnotations := annotations.Types["Product"]
	if productAnnotations.Proto == nil || productAnnotations.Proto.Name != "ProductV1" {
		t.Errorf("Expected Proto.Name 'ProductV1', got '%v'", productAnnotations.Proto)
	}
}

func TestMergeYAMLAnnotations_FieldListMerging(t *testing.T) {
	tmpDir := t.TempDir()

	// Create base file with exclude list
	baseFile := filepath.Join(tmpDir, "base.yaml")
	baseContent := `
types:
  User:
    fields:
      email:
        exclude: ["proto"]
`
	if err := os.WriteFile(baseFile, []byte(baseContent), 0644); err != nil {
		t.Fatalf("Failed to create base file: %v", err)
	}

	// Create override file with additional exclude
	overrideFile := filepath.Join(tmpDir, "override.yaml")
	overrideContent := `
types:
  User:
    fields:
      email:
        exclude: ["graphql"]
`
	if err := os.WriteFile(overrideFile, []byte(overrideContent), 0644); err != nil {
		t.Fatalf("Failed to create override file: %v", err)
	}

	// Merge files
	annotations, err := MergeYAMLAnnotations([]string{baseFile, overrideFile})
	if err != nil {
		t.Fatalf("MergeYAMLAnnotations failed: %v", err)
	}

	// Verify exclude list was merged
	emailAnnotations := annotations.Types["User"].Fields["email"]
	if len(emailAnnotations.Exclude) != 2 {
		t.Errorf("Expected 2 exclude items, got %d", len(emailAnnotations.Exclude))
	}

	excludeMap := make(map[string]bool)
	for _, item := range emailAnnotations.Exclude {
		excludeMap[item] = true
	}

	if !excludeMap["proto"] {
		t.Error("Expected 'proto' in exclude list")
	}
	if !excludeMap["graphql"] {
		t.Error("Expected 'graphql' in exclude list")
	}
}
