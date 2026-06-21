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
	for _, ln := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(ln)
		if strings.HasPrefix(trimmed, "@") {
			name, val := splitTag(trimmed)
			a.HasTags = true
			cur = ""
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
				return nil, fmt.Errorf("unknown @tag %q", name)
			}
			continue
		}
		if cur == "" || trimmed == "" {
			continue
		}
		switch cur {
		case "description":
			a.Description += "\n" + trimmed
		case "boundary":
			a.Boundary += "\n" + trimmed
		}
	}
	return a, nil
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
