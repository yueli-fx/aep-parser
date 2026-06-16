package main

import (
	"fmt"
	"strings"
)

var validTiers = map[string]bool{
	"stable": true, "alpha": true, "planned": true, "missing": true, "negative": true,
}

var validVerify = map[string]bool{
	"none": true, "roundtrip": true, "ae-accept": true, "render-pixel": true,
}

// validDomains is the closed set of capability buckets (spec § 数据模型). "meta"
// is the lane for getters/readers/type aliases/enum consts — present in the API
// but not verified capabilities (no gate, verify none|roundtrip).
var validDomains = map[string]bool{
	"layer-create": true, "layer-set": true, "shape": true, "gradient": true,
	"keyframe": true, "effect": true, "text": true, "mask": true, "comp": true,
	"project": true, "render-queue": true, "eg": true, "expr": true, "io": true,
	"structural": true, "meta": true,
}

// parseCapTag scans a doc comment for an `aep:cap ...` directive (optionally
// continued across following non-blank lines) and parses it into a Cap.
// Returns (nil, false, nil) when no directive is present; (nil, true, err) on a
// malformed directive.
func parseCapTag(comment string) (*Cap, bool, error) {
	lines := strings.Split(comment, "\n")
	start := -1
	for i, ln := range lines {
		if strings.HasPrefix(strings.TrimSpace(ln), "aep:cap") {
			start = i
			break
		}
	}
	if start < 0 {
		return nil, false, nil
	}
	var b strings.Builder
	for i := start; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		if i > start && t == "" {
			break
		}
		if i > start {
			b.WriteByte(' ')
		}
		b.WriteString(t)
	}
	body := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(b.String()), "aep:cap"))
	kv, err := tokenizeKV(body)
	if err != nil {
		return nil, true, err
	}
	c := &Cap{MinVer: "2020"}
	for k, v := range kv {
		switch k {
		case "domain":
			c.Domain = v
		case "tier":
			c.Tier = v
		case "verify":
			c.Verify = v
		case "minver":
			c.MinVer = v
		case "boundary":
			c.Boundary = v
		case "gate":
			c.Gate = splitList(v)
		case "incident":
			c.Incident = splitList(v)
		case "alias":
			c.Alias = splitList(v)
		default:
			return nil, true, fmt.Errorf("unknown aep:cap key %q", k)
		}
	}
	return c, true, nil
}

// validateCap enforces the tier×verify consistency rules. It normalizes an empty
// MinVer to "2020".
func validateCap(c *Cap) error {
	if c.Domain == "" || c.Tier == "" || c.Verify == "" {
		return fmt.Errorf("aep:cap requires domain, tier, verify")
	}
	if !validDomains[c.Domain] {
		return fmt.Errorf("invalid domain %q", c.Domain)
	}
	if !validTiers[c.Tier] {
		return fmt.Errorf("invalid tier %q", c.Tier)
	}
	if !validVerify[c.Verify] {
		return fmt.Errorf("invalid verify %q", c.Verify)
	}
	// meta lane: getters/readers/aliases/enums — exist in the API but are not
	// verified capabilities. No gate; verify must be none|roundtrip.
	if c.Domain == "meta" {
		if c.Verify != "none" && c.Verify != "roundtrip" {
			return fmt.Errorf("domain=meta requires verify none|roundtrip, got %q", c.Verify)
		}
		if len(c.Gate) > 0 {
			return fmt.Errorf("domain=meta must not declare a gate")
		}
		if c.MinVer == "" {
			c.MinVer = "2020"
		}
		return nil
	}
	switch c.Tier {
	case "stable":
		if c.Verify != "ae-accept" && c.Verify != "render-pixel" {
			return fmt.Errorf("tier=stable requires verify ae-accept|render-pixel, got %q", c.Verify)
		}
	case "alpha":
		if c.Verify == "none" {
			return fmt.Errorf("tier=alpha requires verify roundtrip|ae-accept|render-pixel")
		}
	case "planned", "missing", "negative":
		if c.Verify != "none" || len(c.Gate) > 0 {
			return fmt.Errorf("tier=%s requires verify=none and no gate", c.Tier)
		}
	}
	if (c.Verify == "ae-accept" || c.Verify == "render-pixel") && len(c.Gate) == 0 {
		return fmt.Errorf("verify=%s requires at least one gate", c.Verify)
	}
	if c.MinVer == "" {
		c.MinVer = "2020"
	}
	return nil
}

// tokenizeKV parses space-separated key=value tokens; a value may be wrapped in
// double quotes to contain spaces.
func tokenizeKV(s string) (map[string]string, error) {
	out := map[string]string{}
	i, n := 0, len(s)
	for i < n {
		for i < n && s[i] == ' ' {
			i++
		}
		if i >= n {
			break
		}
		ks := i
		for i < n && s[i] != '=' && s[i] != ' ' {
			i++
		}
		if i >= n || s[i] != '=' {
			return nil, fmt.Errorf("expected key=value near %q", s[ks:])
		}
		key := s[ks:i]
		i++ // skip '='
		var val string
		if i < n && s[i] == '"' {
			i++
			vs := i
			for i < n && s[i] != '"' {
				i++
			}
			if i >= n {
				return nil, fmt.Errorf("unterminated quote for key %q", key)
			}
			val = s[vs:i]
			i++ // skip closing quote
		} else {
			vs := i
			for i < n && s[i] != ' ' {
				i++
			}
			val = s[vs:i]
		}
		out[key] = val
	}
	return out, nil
}

func splitList(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
