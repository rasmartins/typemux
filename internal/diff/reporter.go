package diff

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// Reporter formats and outputs diff results
type Reporter struct {
	result *Result
	writer io.Writer
}

// write is a helper to write formatted output, ignoring errors for simplicity
func (r *Reporter) write(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(r.writer, format, args...) //nolint:errcheck // Output errors are non-critical
}

// NewReporter creates a new reporter
func NewReporter(result *Result, writer io.Writer) *Reporter {
	return &Reporter{
		result: result,
		writer: writer,
	}
}

// Report generates and outputs the full report
func (r *Reporter) Report() error {
	// Header
	r.write("╔═══════════════════════════════════════════════════════════════╗\n")
	r.write("║          TypeMUX Schema Breaking Change Analysis              ║\n")
	r.write("╚═══════════════════════════════════════════════════════════════╝\n\n")

	// Summary
	r.printSummary()

	// Breaking changes by protocol
	r.printProtocolChanges(ProtocolProto, "Protobuf")
	r.printProtocolChanges(ProtocolGraphQL, "GraphQL")
	r.printProtocolChanges(ProtocolOpenAPI, "OpenAPI")

	// Recommendation
	r.printRecommendation()

	return nil
}

func (r *Reporter) printSummary() {
	r.write("Summary:\n")
	r.write("  Total changes: %d\n", len(r.result.Changes))

	if r.result.BreakingCount > 0 {
		r.write("  ❌ Breaking:     %d\n", r.result.BreakingCount)
	} else {
		r.write("  ✅ Breaking:     %d\n", r.result.BreakingCount)
	}

	if r.result.DangerousCount > 0 {
		r.write("  ⚠️  Dangerous:    %d\n", r.result.DangerousCount)
	} else {
		r.write("  ✅ Dangerous:    %d\n", r.result.DangerousCount)
	}

	r.write("  ✨ Non-breaking: %d\n\n", r.result.NonBreakingCount)
}

func (r *Reporter) printProtocolChanges(protocol Protocol, displayName string) {
	changes := r.result.GetChangesByProtocol(protocol)
	if len(changes) == 0 {
		return
	}

	// Sort by severity (breaking first)
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Severity != changes[j].Severity {
			return severityOrder(changes[i].Severity) < severityOrder(changes[j].Severity)
		}
		return changes[i].Path < changes[j].Path
	})

	r.write("─────────────────────────────────────────────────────────────────\n")
	r.write("%s Changes (%d)\n", displayName, len(changes))
	r.write("─────────────────────────────────────────────────────────────────\n\n")

	for _, change := range changes {
		r.printChange(change)
	}
	r.write("\n")
}

func (r *Reporter) printChange(change *Change) {
	icon := r.getSeverityIcon(change.Severity)

	r.write("%s %s\n", icon, change.Description)
	r.write("   Location: %s\n", change.Path)

	if change.OldValue != "" && change.NewValue != "" {
		r.write("   Changed:  %s → %s\n", change.OldValue, change.NewValue)
	} else if change.OldValue != "" {
		r.write("   Removed:  %s\n", change.OldValue)
	} else if change.NewValue != "" {
		r.write("   Added:    %s\n", change.NewValue)
	}

	r.write("\n")
}

func (r *Reporter) printRecommendation() {
	r.write("═════════════════════════════════════════════════════════════════\n")
	r.write("Recommendation\n")
	r.write("═════════════════════════════════════════════════════════════════\n\n")

	bump := r.result.RecommendedSemverBump()

	switch bump {
	case "major":
		r.write("⚠️  MAJOR version bump required\n\n")
		r.write("Breaking changes detected that will impact existing clients.\n")
		r.write("Consider:\n")
		r.write("  • Deprecating instead of removing\n")
		r.write("  • Adding migration guides\n")
		r.write("  • Providing backwards compatibility\n")
	case "minor":
		r.write("✨ MINOR version bump recommended\n\n")
		r.write("New features or potentially risky changes detected.\n")
		r.write("Review dangerous changes carefully.\n")
	case "patch":
		r.write("✅ PATCH version bump (or no changes)\n\n")
		r.write("Only safe changes or bug fixes detected.\n")
	}

	r.write("\n")
}

func (r *Reporter) getSeverityIcon(severity Severity) string {
	switch severity {
	case SeverityBreaking:
		return "❌"
	case SeverityDangerous:
		return "⚠️ "
	case SeverityNonBreaking:
		return "✨"
	default:
		return "  "
	}
}

func severityOrder(s Severity) int {
	switch s {
	case SeverityBreaking:
		return 0
	case SeverityDangerous:
		return 1
	case SeverityNonBreaking:
		return 2
	default:
		return 3
	}
}

// CompactReport generates a compact one-line summary
func (r *Reporter) CompactReport() string {
	if r.result.BreakingCount > 0 {
		return fmt.Sprintf("❌ %d breaking, %d dangerous, %d safe changes | Recommended: %s version bump",
			r.result.BreakingCount,
			r.result.DangerousCount,
			r.result.NonBreakingCount,
			strings.ToUpper(r.result.RecommendedSemverBump()))
	}

	if r.result.DangerousCount > 0 {
		return fmt.Sprintf("⚠️  %d dangerous, %d safe changes | Recommended: %s version bump",
			r.result.DangerousCount,
			r.result.NonBreakingCount,
			strings.ToUpper(r.result.RecommendedSemverBump()))
	}

	if r.result.NonBreakingCount > 0 {
		return fmt.Sprintf("✅ %d safe changes | Recommended: %s version bump",
			r.result.NonBreakingCount,
			strings.ToUpper(r.result.RecommendedSemverBump()))
	}

	return "✅ No changes detected"
}

// JSONReport outputs the result as JSON
func (r *Reporter) JSONReport() error {
	// This would be implemented for machine-readable output
	return nil
}
