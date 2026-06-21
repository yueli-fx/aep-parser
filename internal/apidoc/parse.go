package apidoc

import (
	"fmt"
	"strings"
)

// Param is one @param entry: a signature parameter name plus its English desc.
type Param struct {
	Name string
	Desc string
}

// Annotation is the parsed @tag block for one exported symbol. HasTags is false
// when the comment carries no @tag line at all (un-converted legacy prose).
type Annotation struct {
	Summary     string
	Description string
	Params      []Param
	Returns     string
	Domain      string
	Stability   string
	Verify      string
	Gate        []string
	Since       string
	Boundary    string
	Incident    []string
	Alias       []string
	HasTags     bool
}

// Parse reads a raw doc-comment body (each line already stripped of its leading
// "//" and one space) and extracts the @tag block. Lines before the first @tag
// are ignored (transitional prose). A line "@name value" opens a tag; subsequent
// lines that do NOT start with "@" are continuation lines appended to the current
// multi-line tag (@description / @boundary).
func Parse(raw string) (*Annotation, error) {
	a := &Annotation{}
	cur := "" // current multi-line tag ("" = none)
	var unknown []string
	for _, ln := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(ln)
		// A "@"-prefixed line is a tag candidate only when its name is alphabetic.
		// Legacy doc comments wrap chunk-offset refs (e.g. "@0x2D/@0x2E") onto
		// their own line; those start with a digit and are prose, not tags.
		if strings.HasPrefix(trimmed, "@") {
			if name, val := splitTag(trimmed); isTagName(name) {
				cur = ""
				known := true
				switch name {
				case "summary":
					a.Summary = val
				case "description":
					a.Description = val
					cur = "description"
				case "param":
					p, err := parseParam(val)
					if err != nil {
						return nil, err
					}
					a.Params = append(a.Params, p)
				case "returns":
					a.Returns = val
				case "domain":
					a.Domain = val
				case "stability":
					a.Stability = val
				case "verify":
					a.Verify = val
				case "gate":
					a.Gate = splitList(val)
				case "since":
					a.Since = val
				case "boundary":
					a.Boundary = val
					cur = "boundary"
				case "incident":
					a.Incident = splitList(val)
				case "alias":
					a.Alias = splitList(val)
				default:
					known = false
				}
				if known {
					a.HasTags = true
				} else {
					unknown = append(unknown, name) // possible typo; reported only if converted
				}
				continue
			}
			// non-alphabetic "@" token (chunk-offset ref) → fall through as prose.
		}
		if cur == "" {
			continue
		}
		// Continuation line. A blank line inside a multi-line field is kept as a
		// paragraph break; leading blanks (field still empty) are dropped.
		switch cur {
		case "description":
			a.Description = appendLine(a.Description, trimmed)
		case "boundary":
			a.Boundary = appendLine(a.Boundary, trimmed)
		}
	}
	a.Description = strings.TrimRight(a.Description, "\n")
	a.Boundary = strings.TrimRight(a.Boundary, "\n")
	// An unknown alphabetic @tag is a typo only inside a genuinely converted
	// block (≥ 1 known tag). In un-converted legacy prose it is just text.
	if a.HasTags && len(unknown) > 0 {
		return nil, fmt.Errorf("unknown @tag %q", unknown[0])
	}
	return a, nil
}

// isTagName reports whether s is a plausible @tag name: a non-empty run of
// lowercase letters and hyphens. This excludes hex chunk-offset refs (@0x2D)
// that appear in legacy doc prose.
func isTagName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && r != '-' {
			return false
		}
	}
	return true
}

// HasLegacyCap reports whether the raw comment carries a legacy aep:cap directive.
func HasLegacyCap(raw string) bool {
	for _, ln := range strings.Split(raw, "\n") {
		if strings.HasPrefix(strings.TrimSpace(ln), "aep:cap") {
			return true
		}
	}
	return false
}

// appendLine joins a continuation line onto a multi-line field, preserving an
// internal blank line as a paragraph break and dropping a leading blank.
func appendLine(field, line string) string {
	if field == "" {
		if line == "" {
			return field
		}
		return line
	}
	return field + "\n" + line
}

func splitTag(line string) (name, val string) {
	line = strings.TrimPrefix(line, "@")
	if i := strings.IndexAny(line, " \t"); i >= 0 {
		return line[:i], strings.TrimSpace(line[i:])
	}
	return line, ""
}

func parseParam(val string) (Param, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return Param{}, fmt.Errorf("@param requires a name")
	}
	name := strings.Fields(val)[0]
	desc := strings.TrimSpace(strings.TrimPrefix(val, name))
	return Param{Name: name, Desc: desc}, nil
}

func splitList(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
