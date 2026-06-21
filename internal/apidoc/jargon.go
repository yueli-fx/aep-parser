package apidoc

import (
	"regexp"
	"strings"
)

// jargonPatterns is the auditable blocklist of project-internal codenames and
// process noise forbidden in comments (case-insensitive). This file is the one
// place the blocklist lives; extend it here.
var jargonPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)v2[._]?2`),      // V2.2 / V2_2 / V22
	regexp.MustCompile(`(?i)\bv3\b`),        // V3 milestone codename
	regexp.MustCompile(`(?i)\bm8\b`),        // M8 milestone codename
	regexp.MustCompile(`(?i)wave[ _-]?\d`),  // "wave 2"
	regexp.MustCompile(`(?i)phase[ _-]?\d`), // "Phase 0" as a codename
	regexp.MustCompile(`(?i)claude\.md`),
	regexp.MustCompile(`(?i)rules\.md`),
	regexp.MustCompile(`(?i)flightdeck`),
	regexp.MustCompile(`(?i)cockpit`),
	regexp.MustCompile(`(?i)py-?aep`),
	regexp.MustCompile(`(?i)\(probe\)`),
	regexp.MustCompile(`(?i)tmp_debug`),
	regexp.MustCompile(`(?i)re'?d from`),
}

// JargonViolation is one blocklist hit on a comment line.
type JargonViolation struct {
	Line    int
	Token   string
	Pattern string
}

// LintJargon scans comment text (lines joined by \n) for blocklisted tokens. A
// line ending with //nolint:jargon is exempt — the explicit escape for a token
// that genuinely must stay (e.g. a real external-version compatibility note).
func LintJargon(comment string) []JargonViolation {
	var out []JargonViolation
	for i, ln := range strings.Split(comment, "\n") {
		if strings.HasSuffix(strings.TrimSpace(ln), "//nolint:jargon") {
			continue
		}
		for _, re := range jargonPatterns {
			if m := re.FindString(ln); m != "" {
				out = append(out, JargonViolation{Line: i + 1, Token: m, Pattern: re.String()})
			}
		}
	}
	return out
}
