package diff

import "github.com/rasmartins/typemux/internal/ast"

// ChangeType represents the type of change detected
type ChangeType string

const (
	// ChangeTypeFieldRemoved indicates a field was removed from a type (breaking change)
	ChangeTypeFieldRemoved ChangeType = "field_removed"
	// ChangeTypeFieldTypeChanged indicates a field's type was changed (breaking change)
	ChangeTypeFieldTypeChanged ChangeType = "field_type_changed"
	// ChangeTypeFieldMadeRequired indicates an optional field became required (breaking change)
	ChangeTypeFieldMadeRequired ChangeType = "field_made_required"
	// ChangeTypeTypeRemoved indicates a type definition was removed (breaking change)
	ChangeTypeTypeRemoved ChangeType = "type_removed"
	// ChangeTypeEnumValueRemoved indicates an enum value was removed (breaking change)
	ChangeTypeEnumValueRemoved ChangeType = "enum_value_removed"
	// ChangeTypeMethodRemoved indicates a service method was removed (breaking change)
	ChangeTypeMethodRemoved ChangeType = "method_removed"
	// ChangeTypeMethodParamChanged indicates a method parameter was changed (breaking change)
	ChangeTypeMethodParamChanged ChangeType = "method_param_changed"
	// ChangeTypeMethodReturnChanged indicates a method return type was changed (breaking change)
	ChangeTypeMethodReturnChanged ChangeType = "method_return_changed"
	// ChangeTypeProtoFieldNumChanged indicates a Protobuf field number was changed (breaking change)
	ChangeTypeProtoFieldNumChanged ChangeType = "proto_field_num_changed"
	// ChangeTypeRequiredParamAdded indicates a required parameter was added (breaking change)
	ChangeTypeRequiredParamAdded ChangeType = "required_param_added"
	// ChangeTypeFieldArgRemoved indicates a field argument was removed (breaking change)
	ChangeTypeFieldArgRemoved ChangeType = "field_arg_removed"
	// ChangeTypeFieldArgTypeChanged indicates a field argument type was changed (breaking change)
	ChangeTypeFieldArgTypeChanged ChangeType = "field_arg_type_changed"
	// ChangeTypeFieldArgMadeRequired indicates a field argument became required (breaking change)
	ChangeTypeFieldArgMadeRequired ChangeType = "field_arg_made_required"
	// ChangeTypeRequiredFieldArgAdded indicates a required field argument was added (breaking change)
	ChangeTypeRequiredFieldArgAdded ChangeType = "required_field_arg_added"

	// ChangeTypeFieldRemovedNoReserve indicates a field was removed without reserving the field number (dangerous change)
	ChangeTypeFieldRemovedNoReserve ChangeType = "field_removed_no_reserve"
	// ChangeTypeFieldMadeOptional indicates a required field became optional (dangerous change)
	ChangeTypeFieldMadeOptional ChangeType = "field_made_optional"

	// ChangeTypeFieldAdded indicates a new field was added (non-breaking change)
	ChangeTypeFieldAdded ChangeType = "field_added"
	// ChangeTypeTypeAdded indicates a new type was added (non-breaking change)
	ChangeTypeTypeAdded ChangeType = "type_added"
	// ChangeTypeEnumValueAdded indicates a new enum value was added (non-breaking change)
	ChangeTypeEnumValueAdded ChangeType = "enum_value_added"
	// ChangeTypeMethodAdded indicates a new service method was added (non-breaking change)
	ChangeTypeMethodAdded ChangeType = "method_added"
	// ChangeTypeFieldDeprecated indicates a field was marked as deprecated (non-breaking change)
	ChangeTypeFieldDeprecated ChangeType = "field_deprecated"
	// ChangeTypeMethodDeprecated indicates a method was marked as deprecated (non-breaking change)
	ChangeTypeMethodDeprecated ChangeType = "method_deprecated"
	// ChangeTypeAnnotationAdded indicates an annotation was added (non-breaking change)
	ChangeTypeAnnotationAdded ChangeType = "annotation_added"
	// ChangeTypeAnnotationChanged indicates an annotation was modified (non-breaking change)
	ChangeTypeAnnotationChanged ChangeType = "annotation_changed"
	// ChangeTypeFieldArgAdded indicates a non-required field argument was added (non-breaking change)
	ChangeTypeFieldArgAdded ChangeType = "field_arg_added"
	// ChangeTypeFieldArgMadeOptional indicates a field argument became optional (non-breaking change)
	ChangeTypeFieldArgMadeOptional ChangeType = "field_arg_made_optional"
)

// Severity indicates how severe a change is
type Severity string

const (
	// SeverityBreaking indicates a change that will break existing clients
	SeverityBreaking Severity = "breaking"
	// SeverityDangerous indicates a change that might break clients in subtle ways
	SeverityDangerous Severity = "dangerous"
	// SeverityNonBreaking indicates a safe change that won't break clients
	SeverityNonBreaking Severity = "non-breaking"
)

// Protocol represents which protocol the change affects
type Protocol string

const (
	// ProtocolGraphQL indicates a change affecting GraphQL schema
	ProtocolGraphQL Protocol = "graphql"
	// ProtocolProto indicates a change affecting Protocol Buffers schema
	ProtocolProto Protocol = "proto"
	// ProtocolOpenAPI indicates a change affecting OpenAPI specification
	ProtocolOpenAPI Protocol = "openapi"
	// ProtocolGo indicates a change affecting Go code generation
	ProtocolGo Protocol = "go"
)

// Change represents a single detected change between schemas
type Change struct {
	Type        ChangeType
	Severity    Severity
	Protocol    Protocol
	Path        string // e.g., "User.email", "UserService.getUser"
	Description string
	OldValue    string
	NewValue    string
}

// Result contains all changes detected between two schemas
type Result struct {
	BaseSchema       *ast.Schema
	HeadSchema       *ast.Schema
	Changes          []*Change
	BreakingCount    int
	DangerousCount   int
	NonBreakingCount int
}

// GetChangesByProtocol returns changes for a specific protocol
func (d *Result) GetChangesByProtocol(protocol Protocol) []*Change {
	var changes []*Change
	for _, change := range d.Changes {
		if change.Protocol == protocol {
			changes = append(changes, change)
		}
	}
	return changes
}

// GetChangesBySeverity returns changes of a specific severity
func (d *Result) GetChangesBySeverity(severity Severity) []*Change {
	var changes []*Change
	for _, change := range d.Changes {
		if change.Severity == severity {
			changes = append(changes, change)
		}
	}
	return changes
}

// HasBreakingChanges returns true if there are any breaking changes
func (d *Result) HasBreakingChanges() bool {
	return d.BreakingCount > 0
}

// HasDangerousChanges returns true if there are any dangerous changes
func (d *Result) HasDangerousChanges() bool {
	return d.DangerousCount > 0
}

// RecommendedSemverBump returns the recommended semver bump based on changes
func (d *Result) RecommendedSemverBump() string {
	if d.BreakingCount > 0 {
		return "major"
	}
	if d.DangerousCount > 0 || d.NonBreakingCount > 0 {
		return "minor"
	}
	return "patch"
}
