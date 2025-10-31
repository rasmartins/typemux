package diff

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rasmartins/typemux/internal/ast"
)

func TestReporter_Report(t *testing.T) {
	result := &Result{
		BaseSchema: &ast.Schema{},
		HeadSchema: &ast.Schema{},
		Changes: []*Change{
			{
				Type:        ChangeTypeTypeRemoved,
				Severity:    SeverityBreaking,
				Protocol:    ProtocolProto,
				Path:        "type.User",
				Description: "Type removed",
			},
			{
				Type:        ChangeTypeFieldAdded,
				Severity:    SeverityNonBreaking,
				Protocol:    ProtocolGraphQL,
				Path:        "type.Product.field.name",
				Description: "Field added",
			},
		},
		BreakingCount:    1,
		DangerousCount:   0,
		NonBreakingCount: 1,
	}

	var buf bytes.Buffer
	reporter := NewReporter(result, &buf)
	err := reporter.Report()

	if err != nil {
		t.Errorf("Report() returned error: %v", err)
	}

	output := buf.String()

	// Check that output contains some expected content
	if len(output) == 0 {
		t.Error("Expected non-empty output")
	}

	// Check that it mentions changes
	if !strings.Contains(output, "Type removed") {
		t.Error("Expected change description")
	}
}

func TestReporter_CompactReport(t *testing.T) {
	result := &Result{
		Changes: []*Change{
			{
				Type:     ChangeTypeTypeRemoved,
				Severity: SeverityBreaking,
			},
		},
		BreakingCount:    1,
		DangerousCount:   0,
		NonBreakingCount: 0,
	}

	var buf bytes.Buffer
	reporter := NewReporter(result, &buf)
	output := reporter.CompactReport() // Returns string, doesn't write to buffer

	// Just check we got some output
	if len(output) == 0 {
		t.Error("Expected non-empty compact report")
	}
}

func TestReporter_JSONReport(t *testing.T) {
	result := &Result{
		Changes: []*Change{
			{
				Type:        ChangeTypeTypeRemoved,
				Severity:    SeverityBreaking,
				Protocol:    ProtocolProto,
				Path:        "type.User",
				Description: "Type removed",
			},
		},
		BreakingCount:    1,
		DangerousCount:   0,
		NonBreakingCount: 0,
	}

	var buf bytes.Buffer
	reporter := NewReporter(result, &buf)
	err := reporter.JSONReport()

	// JSON report is not implemented yet, just check it doesn't crash
	if err != nil {
		t.Errorf("JSONReport returned error: %v", err)
	}
}

func TestReporter_NoChanges(t *testing.T) {
	result := &Result{
		BaseSchema:       &ast.Schema{},
		HeadSchema:       &ast.Schema{},
		Changes:          []*Change{},
		BreakingCount:    0,
		DangerousCount:   0,
		NonBreakingCount: 0,
	}

	var buf bytes.Buffer
	reporter := NewReporter(result, &buf)
	reporter.Report()

	output := buf.String()

	// Just check we got some output
	if len(output) == 0 {
		t.Error("Expected some output even for no changes")
	}
}

func TestReporter_MultipleProtocols(t *testing.T) {
	result := &Result{
		BaseSchema: &ast.Schema{},
		HeadSchema: &ast.Schema{},
		Changes: []*Change{
			{
				Type:        ChangeTypeTypeRemoved,
				Severity:    SeverityBreaking,
				Protocol:    ProtocolProto,
				Path:        "type.User",
				Description: "Type removed in Protobuf",
			},
			{
				Type:        ChangeTypeFieldAdded,
				Severity:    SeverityNonBreaking,
				Protocol:    ProtocolGraphQL,
				Path:        "type.Product.field.name",
				Description: "Field added in GraphQL",
			},
		},
		BreakingCount:    1,
		DangerousCount:   0,
		NonBreakingCount: 1,
	}

	var buf bytes.Buffer
	reporter := NewReporter(result, &buf)
	reporter.Report()

	output := buf.String()

	// Check that both protocols are mentioned
	if !strings.Contains(output, "Protobuf") {
		t.Error("Expected Protobuf protocol section")
	}

	if !strings.Contains(output, "GraphQL") {
		t.Error("Expected GraphQL protocol section")
	}
}

func TestReporter_DangerousChanges(t *testing.T) {
	result := &Result{
		BaseSchema: &ast.Schema{},
		HeadSchema: &ast.Schema{},
		Changes: []*Change{
			{
				Type:        ChangeTypeFieldMadeOptional,
				Severity:    SeverityDangerous,
				Protocol:    ProtocolProto,
				Path:        "type.User.field.email",
				Description: "Field made optional (dangerous)",
			},
		},
		BreakingCount:    0,
		DangerousCount:   1,
		NonBreakingCount: 0,
	}

	var buf bytes.Buffer
	reporter := NewReporter(result, &buf)
	reporter.Report()

	output := buf.String()

	// Just check we got some output
	if len(output) == 0 {
		t.Error("Expected non-empty output for dangerous changes")
	}
}

func TestGetSeverityIcon(t *testing.T) {
	tests := []struct {
		severity     Severity
		expectedIcon string
	}{
		{SeverityBreaking, "❌"},
		{SeverityDangerous, "⚠️ "},
		{SeverityNonBreaking, "✨"},
	}

	result := &Result{}
	reporter := NewReporter(result, nil)
	for _, tt := range tests {
		icon := reporter.getSeverityIcon(tt.severity)
		if icon != tt.expectedIcon {
			t.Errorf("Expected icon %s for severity %s, got %s", tt.expectedIcon, tt.severity, icon)
		}
	}
}

func TestSeverityOrder(t *testing.T) {
	// Breaking should be ordered first
	if severityOrder(SeverityBreaking) >= severityOrder(SeverityDangerous) {
		t.Error("Breaking should be ordered before Dangerous")
	}

	if severityOrder(SeverityDangerous) >= severityOrder(SeverityNonBreaking) {
		t.Error("Dangerous should be ordered before NonBreaking")
	}
}
