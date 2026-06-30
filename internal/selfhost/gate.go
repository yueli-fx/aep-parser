package selfhost

import (
	"fmt"
	"strings"
)

type VerifyOptions struct {
	InputPath string
	OutRoot   string
	Limit     int
	Open      bool
}

func FormatVerifyDryRun(opts VerifyOptions) string {
	var b strings.Builder
	fmt.Fprintln(&b, "DRY RUN technique selfhost verify")
	fmt.Fprintln(&b, "gate engine: go")
	fmt.Fprintf(&b, "input_path: %s\n", valueOrDefault(opts.InputPath, "data/samples"))
	fmt.Fprintf(&b, "out_root: %s\n", opts.OutRoot)
	fmt.Fprintf(&b, "limit: %d\n", opts.Limit)
	fmt.Fprintf(&b, "open: %t\n", opts.Open)
	fmt.Fprintln(&b, "steps:")
	fmt.Fprintln(&b, "- run technique tests")
	fmt.Fprintln(&b, "- generate full technique report")
	fmt.Fprintln(&b, "- generate partial-error technique report")
	fmt.Fprintln(&b, "- compare reports")
	fmt.Fprintln(&b, "- smoke recipe drafts")
	fmt.Fprintln(&b, "- write acceptance, effectiveness, outcome, history, and index artifacts")
	return b.String()
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
