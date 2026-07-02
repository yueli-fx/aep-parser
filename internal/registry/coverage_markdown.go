package registry

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type coverageOpenItem struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Scope  string `json:"scope"`
}

func RenderCoverageMarkdown(root, coveragePath string) (string, error) {
	var coverage coverageFile
	if err := readJSONPath(root, coveragePath, &coverage); err != nil {
		return "", err
	}

	var b strings.Builder
	line := func(format string, args ...any) {
		fmt.Fprintf(&b, format, args...)
		b.WriteByte('\n')
	}

	line("# Versioned AEP Migration Coverage")
	line("")
	line("Generated from `%s`.", filepath.ToSlash(coveragePath))
	line("")
	line("## Host Open Policy")
	line("")
	line("- Matrix command: %s", coverage.HostOpenPolicy.MatrixCommandStatus)
	line("- Default strategy: %s", coverage.HostOpenPolicy.DefaultStrategy)
	line("- Broad fanout: %s", coverage.HostOpenPolicy.BroadFanoutStatus)
	line("- Endpoint inference: direct=%s; inferred=%s; label=%s",
		strings.Join(coverage.HostOpenPolicy.EndpointInference.DirectHosts, ","),
		strings.Join(coverage.HostOpenPolicy.EndpointInference.InferredHosts, ","),
		coverage.HostOpenPolicy.EndpointInference.Label,
	)
	line("")
	line("## Recurring Gates")
	line("")
	line("| Gate | Status | Totals | Artifact |")
	line("| --- | --- | --- | --- |")
	for _, gate := range coverage.RecurringGates {
		line("| `%s` | %s | %s | `%s` |", gate.ID, gate.Status, formatCoverageTotals(gate.Totals), filepath.ToSlash(gate.Artifact))
	}
	line("")
	line("## Coverage")
	line("")
	line("| ID | Domain | Writer Status | Totals | Host Open | Artifact |")
	line("| --- | --- | --- | --- | --- | --- |")
	for _, record := range coverage.Coverage {
		line("| `%s` | %s | %s | %s | %s | `%s` |",
			record.ID,
			record.Domain,
			record.WriterStatus,
			formatCoverageTotals(record.Totals),
			record.HostOpenStatus,
			filepath.ToSlash(record.Artifact),
		)
	}
	line("")
	line("## Open Items")
	line("")
	for _, raw := range coverage.OpenItems {
		var item coverageOpenItem
		if err := json.Unmarshal(raw, &item); err != nil {
			return "", fmt.Errorf("open item: %w", err)
		}
		line("- `%s`: %s - %s", item.ID, item.Status, item.Scope)
	}
	return b.String(), nil
}

func formatCoverageTotals(t CoverageTotals) string {
	return fmt.Sprintf("%d/%d pass, blocked=%d, failed=%d, skipped=%d", t.Pass, t.Total, t.Blocked, t.Failed, t.Skipped)
}
