package diff

import (
	"github.com/rasmartins/typemux/internal/ast"
)

// Differ compares two schemas and detects changes
type Differ struct {
	baseSchema *ast.Schema
	headSchema *ast.Schema
	changes    []*Change
}

// NewDiffer creates a new schema differ
func NewDiffer(baseSchema, headSchema *ast.Schema) *Differ {
	return &Differ{
		baseSchema: baseSchema,
		headSchema: headSchema,
		changes:    make([]*Change, 0),
	}
}

// Compare performs the comparison and returns a diff result
func (d *Differ) Compare() *Result {
	// Compare types
	d.compareTypes()

	// Compare enums
	d.compareEnums()

	// Compare unions
	d.compareUnions()

	// Compare services
	d.compareServices()

	// Build result
	result := &Result{
		BaseSchema: d.baseSchema,
		HeadSchema: d.headSchema,
		Changes:    d.changes,
	}

	// Count by severity
	for _, change := range d.changes {
		switch change.Severity {
		case SeverityBreaking:
			result.BreakingCount++
		case SeverityDangerous:
			result.DangerousCount++
		case SeverityNonBreaking:
			result.NonBreakingCount++
		}
	}

	return result
}

// compareTypes compares type definitions
func (d *Differ) compareTypes() {
	baseTypes := make(map[string]*ast.Type)
	headTypes := make(map[string]*ast.Type)

	// Build maps
	for _, t := range d.baseSchema.Types {
		baseTypes[t.Name] = t
	}
	for _, t := range d.headSchema.Types {
		headTypes[t.Name] = t
	}

	// Check for removed types
	for name := range baseTypes {
		if _, exists := headTypes[name]; !exists {
			d.addChange(&Change{
				Type:        ChangeTypeTypeRemoved,
				Severity:    SeverityBreaking,
				Protocol:    ProtocolGraphQL, // Affects all protocols
				Path:        name,
				Description: "Type removed",
				OldValue:    name,
				NewValue:    "",
			})
		}
	}

	// Check for added types
	for name := range headTypes {
		if _, exists := baseTypes[name]; !exists {
			d.addChange(&Change{
				Type:        ChangeTypeTypeAdded,
				Severity:    SeverityNonBreaking,
				Protocol:    ProtocolGraphQL,
				Path:        name,
				Description: "Type added",
				OldValue:    "",
				NewValue:    name,
			})
		}
	}

	// Check for modified types
	for name, baseType := range baseTypes {
		if headType, exists := headTypes[name]; exists {
			d.compareTypeFields(baseType, headType)
		}
	}
}

// compareTypeFields compares fields within a type
func (d *Differ) compareTypeFields(baseType, headType *ast.Type) {
	baseFields := make(map[string]*ast.Field)
	headFields := make(map[string]*ast.Field)

	for _, f := range baseType.Fields {
		baseFields[f.Name] = f
	}
	for _, f := range headType.Fields {
		headFields[f.Name] = f
	}

	// Check for removed fields
	for fieldName, baseField := range baseFields {
		if _, exists := headFields[fieldName]; !exists {
			path := baseType.Name + "." + fieldName

			// Check if it's a Protobuf field number removal (more severe)
			if baseField.Number > 0 {
				d.addChange(&Change{
					Type:        ChangeTypeFieldRemovedNoReserve,
					Severity:    SeverityDangerous,
					Protocol:    ProtocolProto,
					Path:        path,
					Description: "Field removed without reserving field number (Protobuf)",
					OldValue:    fieldName,
					NewValue:    "",
				})
			}

			// Breaking for all protocols
			d.addChange(&Change{
				Type:        ChangeTypeFieldRemoved,
				Severity:    SeverityBreaking,
				Protocol:    ProtocolGraphQL,
				Path:        path,
				Description: "Field removed",
				OldValue:    fieldName,
				NewValue:    "",
			})
		}
	}

	// Check for added fields
	for fieldName, headField := range headFields {
		if _, exists := baseFields[fieldName]; !exists {
			path := headType.Name + "." + fieldName

			// Check if it's required (breaking for input types in GraphQL/OpenAPI)
			if headField.Required {
				d.addChange(&Change{
					Type:        ChangeTypeRequiredParamAdded,
					Severity:    SeverityBreaking,
					Protocol:    ProtocolOpenAPI,
					Path:        path,
					Description: "Required field added",
					OldValue:    "",
					NewValue:    fieldName,
				})
			} else {
				d.addChange(&Change{
					Type:        ChangeTypeFieldAdded,
					Severity:    SeverityNonBreaking,
					Protocol:    ProtocolGraphQL,
					Path:        path,
					Description: "Field added",
					OldValue:    "",
					NewValue:    fieldName,
				})
			}
		}
	}

	// Check for modified fields
	for fieldName, baseField := range baseFields {
		if headField, exists := headFields[fieldName]; exists {
			d.compareFieldChanges(baseType.Name, baseField, headField)
		}
	}
}

// compareFieldChanges compares individual field changes
func (d *Differ) compareFieldChanges(typeName string, baseField, headField *ast.Field) {
	path := typeName + "." + baseField.Name

	// Check for type changes
	if !fieldTypesEqual(baseField.Type, headField.Type) {
		d.addChange(&Change{
			Type:        ChangeTypeFieldTypeChanged,
			Severity:    SeverityBreaking,
			Protocol:    ProtocolGraphQL,
			Path:        path,
			Description: "Field type changed",
			OldValue:    formatFieldType(baseField.Type),
			NewValue:    formatFieldType(headField.Type),
		})
	}

	// Check for required changes
	if !baseField.Required && headField.Required {
		d.addChange(&Change{
			Type:        ChangeTypeFieldMadeRequired,
			Severity:    SeverityBreaking,
			Protocol:    ProtocolGraphQL,
			Path:        path,
			Description: "Field made required",
			OldValue:    "optional",
			NewValue:    "required",
		})
	} else if baseField.Required && !headField.Required {
		d.addChange(&Change{
			Type:        ChangeTypeFieldMadeOptional,
			Severity:    SeverityDangerous,
			Protocol:    ProtocolGraphQL,
			Path:        path,
			Description: "Field made optional (clients may stop sending it)",
			OldValue:    "required",
			NewValue:    "optional",
		})
	}

	// Check for Protobuf field number changes (critical!)
	if baseField.Number > 0 && headField.Number > 0 && baseField.Number != headField.Number {
		d.addChange(&Change{
			Type:        ChangeTypeProtoFieldNumChanged,
			Severity:    SeverityBreaking,
			Protocol:    ProtocolProto,
			Path:        path,
			Description: "Protobuf field number changed",
			OldValue:    string(rune(baseField.Number)),
			NewValue:    string(rune(headField.Number)),
		})
	}
}

// compareEnums compares enum definitions
func (d *Differ) compareEnums() {
	baseEnums := make(map[string]*ast.Enum)
	headEnums := make(map[string]*ast.Enum)

	for _, e := range d.baseSchema.Enums {
		baseEnums[e.Name] = e
	}
	for _, e := range d.headSchema.Enums {
		headEnums[e.Name] = e
	}

	// Check for removed enums
	for name := range baseEnums {
		if _, exists := headEnums[name]; !exists {
			d.addChange(&Change{
				Type:        ChangeTypeTypeRemoved,
				Severity:    SeverityBreaking,
				Protocol:    ProtocolGraphQL,
				Path:        name,
				Description: "Enum removed",
				OldValue:    name,
				NewValue:    "",
			})
		}
	}

	// Check for added enums
	for name := range headEnums {
		if _, exists := baseEnums[name]; !exists {
			d.addChange(&Change{
				Type:        ChangeTypeTypeAdded,
				Severity:    SeverityNonBreaking,
				Protocol:    ProtocolGraphQL,
				Path:        name,
				Description: "Enum added",
				OldValue:    "",
				NewValue:    name,
			})
		}
	}

	// Check for modified enums
	for name, baseEnum := range baseEnums {
		if headEnum, exists := headEnums[name]; exists {
			d.compareEnumValues(baseEnum, headEnum)
		}
	}
}

// compareEnumValues compares enum values
func (d *Differ) compareEnumValues(baseEnum, headEnum *ast.Enum) {
	baseValues := make(map[string]*ast.EnumValue)
	headValues := make(map[string]*ast.EnumValue)

	for _, v := range baseEnum.Values {
		baseValues[v.Name] = v
	}
	for _, v := range headEnum.Values {
		headValues[v.Name] = v
	}

	// Check for removed values (breaking!)
	for valueName := range baseValues {
		if _, exists := headValues[valueName]; !exists {
			path := baseEnum.Name + "." + valueName
			d.addChange(&Change{
				Type:        ChangeTypeEnumValueRemoved,
				Severity:    SeverityBreaking,
				Protocol:    ProtocolGraphQL,
				Path:        path,
				Description: "Enum value removed",
				OldValue:    valueName,
				NewValue:    "",
			})
		}
	}

	// Check for added values (safe)
	for valueName := range headValues {
		if _, exists := baseValues[valueName]; !exists {
			path := headEnum.Name + "." + valueName
			d.addChange(&Change{
				Type:        ChangeTypeEnumValueAdded,
				Severity:    SeverityNonBreaking,
				Protocol:    ProtocolGraphQL,
				Path:        path,
				Description: "Enum value added",
				OldValue:    "",
				NewValue:    valueName,
			})
		}
	}
}

// compareUnions compares union definitions
func (d *Differ) compareUnions() {
	baseUnions := make(map[string]*ast.Union)
	headUnions := make(map[string]*ast.Union)

	for _, u := range d.baseSchema.Unions {
		baseUnions[u.Name] = u
	}
	for _, u := range d.headSchema.Unions {
		headUnions[u.Name] = u
	}

	// Check for removed unions
	for name := range baseUnions {
		if _, exists := headUnions[name]; !exists {
			d.addChange(&Change{
				Type:        ChangeTypeTypeRemoved,
				Severity:    SeverityBreaking,
				Protocol:    ProtocolGraphQL,
				Path:        name,
				Description: "Union removed",
				OldValue:    name,
				NewValue:    "",
			})
		}
	}

	// Check for added unions
	for name := range headUnions {
		if _, exists := baseUnions[name]; !exists {
			d.addChange(&Change{
				Type:        ChangeTypeTypeAdded,
				Severity:    SeverityNonBreaking,
				Protocol:    ProtocolGraphQL,
				Path:        name,
				Description: "Union added",
				OldValue:    "",
				NewValue:    name,
			})
		}
	}
}

// compareServices compares service definitions
func (d *Differ) compareServices() {
	baseServices := make(map[string]*ast.Service)
	headServices := make(map[string]*ast.Service)

	for _, s := range d.baseSchema.Services {
		baseServices[s.Name] = s
	}
	for _, s := range d.headSchema.Services {
		headServices[s.Name] = s
	}

	// Check for removed services
	for name := range baseServices {
		if _, exists := headServices[name]; !exists {
			d.addChange(&Change{
				Type:        ChangeTypeTypeRemoved,
				Severity:    SeverityBreaking,
				Protocol:    ProtocolGraphQL,
				Path:        name,
				Description: "Service removed",
				OldValue:    name,
				NewValue:    "",
			})
		}
	}

	// Check for added services
	for name := range headServices {
		if _, exists := baseServices[name]; !exists {
			d.addChange(&Change{
				Type:        ChangeTypeTypeAdded,
				Severity:    SeverityNonBreaking,
				Protocol:    ProtocolGraphQL,
				Path:        name,
				Description: "Service added",
				OldValue:    "",
				NewValue:    name,
			})
		}
	}

	// Check for modified services
	for name, baseService := range baseServices {
		if headService, exists := headServices[name]; exists {
			d.compareServiceMethods(baseService, headService)
		}
	}
}

// compareServiceMethods compares methods within a service
func (d *Differ) compareServiceMethods(baseService, headService *ast.Service) {
	baseMethods := make(map[string]*ast.Method)
	headMethods := make(map[string]*ast.Method)

	for _, m := range baseService.Methods {
		baseMethods[m.Name] = m
	}
	for _, m := range headService.Methods {
		headMethods[m.Name] = m
	}

	// Check for removed methods
	for methodName := range baseMethods {
		if _, exists := headMethods[methodName]; !exists {
			path := baseService.Name + "." + methodName
			d.addChange(&Change{
				Type:        ChangeTypeMethodRemoved,
				Severity:    SeverityBreaking,
				Protocol:    ProtocolGraphQL,
				Path:        path,
				Description: "Method removed",
				OldValue:    methodName,
				NewValue:    "",
			})
		}
	}

	// Check for added methods
	for methodName := range headMethods {
		if _, exists := baseMethods[methodName]; !exists {
			path := headService.Name + "." + methodName
			d.addChange(&Change{
				Type:        ChangeTypeMethodAdded,
				Severity:    SeverityNonBreaking,
				Protocol:    ProtocolGraphQL,
				Path:        path,
				Description: "Method added",
				OldValue:    "",
				NewValue:    methodName,
			})
		}
	}

	// Check for modified methods
	for methodName, baseMethod := range baseMethods {
		if headMethod, exists := headMethods[methodName]; exists {
			d.compareMethodSignature(baseService.Name, baseMethod, headMethod)
		}
	}
}

// compareMethodSignature compares method signatures
func (d *Differ) compareMethodSignature(serviceName string, baseMethod, headMethod *ast.Method) {
	path := serviceName + "." + baseMethod.Name

	// Check for parameter type changes
	if baseMethod.InputType != headMethod.InputType {
		d.addChange(&Change{
			Type:        ChangeTypeMethodParamChanged,
			Severity:    SeverityBreaking,
			Protocol:    ProtocolGraphQL,
			Path:        path,
			Description: "Method parameter type changed",
			OldValue:    baseMethod.InputType,
			NewValue:    headMethod.InputType,
		})
	}

	// Check for return type changes
	if baseMethod.OutputType != headMethod.OutputType {
		d.addChange(&Change{
			Type:        ChangeTypeMethodReturnChanged,
			Severity:    SeverityBreaking,
			Protocol:    ProtocolGraphQL,
			Path:        path,
			Description: "Method return type changed",
			OldValue:    baseMethod.OutputType,
			NewValue:    headMethod.OutputType,
		})
	}
}

// Helper functions

func (d *Differ) addChange(change *Change) {
	d.changes = append(d.changes, change)
}

func fieldTypesEqual(t1, t2 *ast.FieldType) bool {
	if t1.Name != t2.Name {
		return false
	}
	if t1.IsArray != t2.IsArray {
		return false
	}
	if t1.IsMap != t2.IsMap {
		return false
	}
	return true
}

func formatFieldType(t *ast.FieldType) string {
	if t.IsArray {
		return "[]" + t.Name
	}
	if t.IsMap {
		if t.MapValueType != nil {
			return "map[" + t.MapKey + "]" + formatFieldType(t.MapValueType)
		}
		return "map[" + t.MapKey + "]" + t.MapValue
	}
	return t.Name
}
