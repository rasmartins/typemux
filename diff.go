package typemux

import (
	"fmt"

	"github.com/rasmartins/typemux/internal/diff"
)

// DiffResult contains the results of comparing two schemas for changes.
type DiffResult struct {
	// BaseSchema is the original schema
	BaseSchema *Schema

	// HeadSchema is the new schema being compared
	HeadSchema *Schema

	// Changes is the list of detected changes
	Changes []*Change

	// BreakingCount is the number of breaking changes
	BreakingCount int

	// DangerousCount is the number of potentially dangerous changes
	DangerousCount int

	// NonBreakingCount is the number of non-breaking changes
	NonBreakingCount int
}

// Change represents a single schema change detected by the differ.
type Change struct {
	// Type is the kind of change (e.g., "field_removed", "type_added")
	Type ChangeType

	// Severity indicates how impactful the change is
	Severity Severity

	// Protocol indicates which output format this change affects
	Protocol Protocol

	// Path is the location of the change (e.g., "type.User.field.email")
	Path string

	// Description is a human-readable description of the change
	Description string

	// OldValue is the previous value (if applicable)
	OldValue string

	// NewValue is the new value (if applicable)
	NewValue string
}

// ChangeType represents the type of schema change.
type ChangeType = diff.ChangeType

// Severity represents the impact level of a change.
type Severity = diff.Severity

// Protocol represents the output format affected by a change.
type Protocol = diff.Protocol

// Change type constants
const (
	ChangeTypeFieldAdded          = diff.ChangeTypeFieldAdded
	ChangeTypeFieldRemoved        = diff.ChangeTypeFieldRemoved
	ChangeTypeFieldTypeChanged    = diff.ChangeTypeFieldTypeChanged
	ChangeTypeFieldMadeRequired   = diff.ChangeTypeFieldMadeRequired
	ChangeTypeFieldMadeOptional   = diff.ChangeTypeFieldMadeOptional
	ChangeTypeTypeAdded           = diff.ChangeTypeTypeAdded
	ChangeTypeTypeRemoved         = diff.ChangeTypeTypeRemoved
	ChangeTypeEnumValueAdded      = diff.ChangeTypeEnumValueAdded
	ChangeTypeEnumValueRemoved    = diff.ChangeTypeEnumValueRemoved
	ChangeTypeMethodAdded         = diff.ChangeTypeMethodAdded
	ChangeTypeMethodRemoved       = diff.ChangeTypeMethodRemoved
	ChangeTypeMethodParamChanged  = diff.ChangeTypeMethodParamChanged
	ChangeTypeMethodReturnChanged = diff.ChangeTypeMethodReturnChanged
)

// Severity constants
const (
	SeverityBreaking    = diff.SeverityBreaking
	SeverityDangerous   = diff.SeverityDangerous
	SeverityNonBreaking = diff.SeverityNonBreaking
)

// Protocol constants
const (
	ProtocolGraphQL = diff.ProtocolGraphQL
	ProtocolProto   = diff.ProtocolProto
	ProtocolOpenAPI = diff.ProtocolOpenAPI
	ProtocolGo      = diff.ProtocolGo
)

// Diff compares two schemas and detects changes.
// It returns a DiffResult containing all detected changes categorized by severity.
//
// Example:
//
//	baseSchema, _ := typemux.ParseSchema(oldIDL)
//	headSchema, _ := typemux.ParseSchema(newIDL)
//	result, err := typemux.Diff(baseSchema, headSchema)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if result.HasBreakingChanges() {
//	    fmt.Println("WARNING: Breaking changes detected!")
//	    fmt.Println(result.Report())
//	}
func Diff(baseSchema, headSchema *Schema) (*DiffResult, error) {
	differ := diff.NewDiffer(baseSchema, headSchema)
	internalResult := differ.Compare()

	// Convert internal result to public API result
	result := &DiffResult{
		BaseSchema:       baseSchema,
		HeadSchema:       headSchema,
		Changes:          make([]*Change, len(internalResult.Changes)),
		BreakingCount:    internalResult.BreakingCount,
		DangerousCount:   internalResult.DangerousCount,
		NonBreakingCount: internalResult.NonBreakingCount,
	}

	// Convert changes
	for i, c := range internalResult.Changes {
		result.Changes[i] = &Change{
			Type:        c.Type,
			Severity:    c.Severity,
			Protocol:    c.Protocol,
			Path:        c.Path,
			Description: c.Description,
			OldValue:    c.OldValue,
			NewValue:    c.NewValue,
		}
	}

	return result, nil
}

// DiffOptions provides options for customizing schema comparison.
type DiffOptions struct {
	// IgnoreChanges specifies change types to ignore
	IgnoreChanges []ChangeType

	// Protocol filters changes to a specific protocol
	Protocol Protocol
}

// DiffWithOptions compares schemas with custom options.
//
// Example:
//
//	result, err := typemux.DiffWithOptions(base, head, typemux.DiffOptions{
//	    IgnoreChanges: []typemux.ChangeType{typemux.ChangeTypeFieldAdded},
//	    Protocol:      typemux.ProtocolGraphQL,
//	})
func DiffWithOptions(baseSchema, headSchema *Schema, opts DiffOptions) (*DiffResult, error) {
	result, err := Diff(baseSchema, headSchema)
	if err != nil {
		return nil, err
	}

	// Apply filters
	if len(opts.IgnoreChanges) > 0 || opts.Protocol != "" {
		filteredChanges := make([]*Change, 0, len(result.Changes))
		breakingCount := 0
		dangerousCount := 0
		nonBreakingCount := 0

		for _, change := range result.Changes {
			// Skip ignored change types
			ignored := false
			for _, ignoreType := range opts.IgnoreChanges {
				if change.Type == ignoreType {
					ignored = true
					break
				}
			}
			if ignored {
				continue
			}

			// Skip non-matching protocols
			if opts.Protocol != "" && change.Protocol != opts.Protocol {
				continue
			}

			filteredChanges = append(filteredChanges, change)

			// Update counts
			switch change.Severity {
			case SeverityBreaking:
				breakingCount++
			case SeverityDangerous:
				dangerousCount++
			case SeverityNonBreaking:
				nonBreakingCount++
			}
		}

		result.Changes = filteredChanges
		result.BreakingCount = breakingCount
		result.DangerousCount = dangerousCount
		result.NonBreakingCount = nonBreakingCount
	}

	return result, nil
}

// HasBreakingChanges returns true if the result contains breaking changes.
func (r *DiffResult) HasBreakingChanges() bool {
	return r.BreakingCount > 0
}

// HasChanges returns true if the result contains any changes.
func (r *DiffResult) HasChanges() bool {
	return len(r.Changes) > 0
}

// Report generates a human-readable report of the changes.
// The report is formatted with colors and icons for terminal display.
//
// Example:
//
//	fmt.Println(result.Report())
func (r *DiffResult) Report() string {
	// Convert to internal result format
	internalResult := &diff.Result{
		BaseSchema:       r.BaseSchema,
		HeadSchema:       r.HeadSchema,
		Changes:          make([]*diff.Change, len(r.Changes)),
		BreakingCount:    r.BreakingCount,
		DangerousCount:   r.DangerousCount,
		NonBreakingCount: r.NonBreakingCount,
	}

	for i, c := range r.Changes {
		internalResult.Changes[i] = &diff.Change{
			Type:        c.Type,
			Severity:    c.Severity,
			Protocol:    c.Protocol,
			Path:        c.Path,
			Description: c.Description,
			OldValue:    c.OldValue,
			NewValue:    c.NewValue,
		}
	}

	// Use internal reporter
	reporter := diff.NewReporter(internalResult, nil)
	return reporter.CompactReport() // Use compact for now; could add full report option
}

// CompactReport generates a one-line summary of the changes.
//
// Example:
//
//	fmt.Println(result.CompactReport())
//	// Output: 2 breaking, 1 dangerous, 3 non-breaking changes detected
func (r *DiffResult) CompactReport() string {
	internalResult := &diff.Result{
		BaseSchema:       r.BaseSchema,
		HeadSchema:       r.HeadSchema,
		Changes:          make([]*diff.Change, len(r.Changes)),
		BreakingCount:    r.BreakingCount,
		DangerousCount:   r.DangerousCount,
		NonBreakingCount: r.NonBreakingCount,
	}

	for i, c := range r.Changes {
		internalResult.Changes[i] = &diff.Change{
			Type:        c.Type,
			Severity:    c.Severity,
			Protocol:    c.Protocol,
			Path:        c.Path,
			Description: c.Description,
			OldValue:    c.OldValue,
			NewValue:    c.NewValue,
		}
	}

	reporter := diff.NewReporter(internalResult, nil)
	return reporter.CompactReport()
}

// JSONReport generates a JSON representation of the changes.
//
// Example:
//
//	jsonStr, err := result.JSONReport()
//	fmt.Println(jsonStr)
func (r *DiffResult) JSONReport() (string, error) {
	// TODO: Implement JSON serialization
	// For now, this is a placeholder
	return "", fmt.Errorf("JSON report not implemented yet")
}
