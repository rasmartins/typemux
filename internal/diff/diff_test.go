package diff

import (
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
)

func TestDiffer_NoChanges(t *testing.T) {
	// Create identical schemas
	schema1 := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{
						Name:     "id",
						Type:     &ast.FieldType{Name: "string"},
						Required: true,
					},
				},
			},
		},
	}

	schema2 := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{
						Name:     "id",
						Type:     &ast.FieldType{Name: "string"},
						Required: true,
					},
				},
			},
		},
	}

	differ := NewDiffer(schema1, schema2)
	result := differ.Compare()

	if len(result.Changes) != 0 {
		t.Errorf("Expected no changes, got %d", len(result.Changes))
	}

	if result.BreakingCount != 0 {
		t.Errorf("Expected no breaking changes, got %d", result.BreakingCount)
	}

	if result.RecommendedSemverBump() != "patch" {
		t.Errorf("Expected patch bump, got %s", result.RecommendedSemverBump())
	}
}

func TestDiffer_TypeRemoved(t *testing.T) {
	// Base schema with User type
	base := &ast.Schema{
		Types: []*ast.Type{
			{Name: "User"},
			{Name: "Product"},
		},
	}

	// Head schema with Product type removed
	head := &ast.Schema{
		Types: []*ast.Type{
			{Name: "User"},
		},
	}

	differ := NewDiffer(base, head)
	result := differ.Compare()

	if result.BreakingCount == 0 {
		t.Error("Expected breaking changes for type removal")
	}

	if result.RecommendedSemverBump() != "major" {
		t.Errorf("Expected major bump for type removal, got %s", result.RecommendedSemverBump())
	}
}

func TestDiffer_FieldTypeChanged(t *testing.T) {
	base := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{
						Name:     "age",
						Type:     &ast.FieldType{Name: "int32"},
						Required: true,
					},
				},
			},
		},
	}

	head := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{
						Name:     "age",
						Type:     &ast.FieldType{Name: "int64"}, // Changed type
						Required: true,
					},
				},
			},
		},
	}

	differ := NewDiffer(base, head)
	result := differ.Compare()

	if result.BreakingCount == 0 {
		t.Error("Expected breaking changes for field type change")
	}

	// Find the specific change
	found := false
	for _, change := range result.Changes {
		if change.Type == ChangeTypeFieldTypeChanged {
			found = true
			if change.OldValue != "int32" || change.NewValue != "int64" {
				t.Errorf("Expected int32→int64, got %s→%s", change.OldValue, change.NewValue)
			}
		}
	}

	if !found {
		t.Error("Expected to find field type change")
	}
}

func TestDiffer_FieldMadeRequired(t *testing.T) {
	base := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{
						Name:     "email",
						Type:     &ast.FieldType{Name: "string"},
						Required: false,
					},
				},
			},
		},
	}

	head := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{
						Name:     "email",
						Type:     &ast.FieldType{Name: "string"},
						Required: true, // Made required
					},
				},
			},
		},
	}

	differ := NewDiffer(base, head)
	result := differ.Compare()

	if result.BreakingCount == 0 {
		t.Error("Expected breaking changes for field made required")
	}

	found := false
	for _, change := range result.Changes {
		if change.Type == ChangeTypeFieldMadeRequired {
			found = true
		}
	}

	if !found {
		t.Error("Expected to find field made required change")
	}
}

func TestDiffer_EnumValueRemoved(t *testing.T) {
	base := &ast.Schema{
		Enums: []*ast.Enum{
			{
				Name: "UserRole",
				Values: []*ast.EnumValue{
					{Name: "ADMIN"},
					{Name: "USER"},
					{Name: "GUEST"},
				},
			},
		},
	}

	head := &ast.Schema{
		Enums: []*ast.Enum{
			{
				Name: "UserRole",
				Values: []*ast.EnumValue{
					{Name: "ADMIN"},
					{Name: "USER"},
					// GUEST removed
				},
			},
		},
	}

	differ := NewDiffer(base, head)
	result := differ.Compare()

	if result.BreakingCount == 0 {
		t.Error("Expected breaking changes for enum value removal")
	}

	found := false
	for _, change := range result.Changes {
		if change.Type == ChangeTypeEnumValueRemoved && change.OldValue == "GUEST" {
			found = true
		}
	}

	if !found {
		t.Error("Expected to find enum value removal")
	}
}

func TestDiffer_NonBreakingAdditions(t *testing.T) {
	base := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string"}},
				},
			},
		},
	}

	head := &ast.Schema{
		Types: []*ast.Type{
			{
				Name: "User",
				Fields: []*ast.Field{
					{Name: "id", Type: &ast.FieldType{Name: "string"}},
					{Name: "email", Type: &ast.FieldType{Name: "string"}, Required: false}, // Optional field added
				},
			},
			{Name: "Product"}, // New type added
		},
	}

	differ := NewDiffer(base, head)
	result := differ.Compare()

	if result.BreakingCount != 0 {
		t.Errorf("Expected no breaking changes, got %d", result.BreakingCount)
	}

	if result.NonBreakingCount == 0 {
		t.Error("Expected non-breaking changes for additions")
	}

	if result.RecommendedSemverBump() != "minor" {
		t.Errorf("Expected minor bump for additions, got %s", result.RecommendedSemverBump())
	}
}

func TestDiffResult_GetChangesByProtocol(t *testing.T) {
	result := &Result{
		Changes: []*Change{
			{Protocol: ProtocolGraphQL},
			{Protocol: ProtocolProto},
			{Protocol: ProtocolGraphQL},
		},
	}

	graphqlChanges := result.GetChangesByProtocol(ProtocolGraphQL)
	if len(graphqlChanges) != 2 {
		t.Errorf("Expected 2 GraphQL changes, got %d", len(graphqlChanges))
	}

	protoChanges := result.GetChangesByProtocol(ProtocolProto)
	if len(protoChanges) != 1 {
		t.Errorf("Expected 1 Protobuf change, got %d", len(protoChanges))
	}
}

func TestDiffResult_GetChangesBySeverity(t *testing.T) {
	result := &Result{
		Changes: []*Change{
			{Severity: SeverityBreaking},
			{Severity: SeverityNonBreaking},
			{Severity: SeverityBreaking},
		},
	}

	breakingChanges := result.GetChangesBySeverity(SeverityBreaking)
	if len(breakingChanges) != 2 {
		t.Errorf("Expected 2 breaking changes, got %d", len(breakingChanges))
	}

	nonBreakingChanges := result.GetChangesBySeverity(SeverityNonBreaking)
	if len(nonBreakingChanges) != 1 {
		t.Errorf("Expected 1 non-breaking change, got %d", len(nonBreakingChanges))
	}
}
