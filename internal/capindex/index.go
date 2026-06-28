package capindex

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
)

type Cap struct {
	Domain   string   `json:"domain"`
	Tier     string   `json:"tier"`
	Verify   string   `json:"verify"`
	MinVer   string   `json:"minver"`
	Gate     []string `json:"gate"`
	Boundary string   `json:"boundary"`
	Incident []string `json:"incident"`
	Alias    []string `json:"alias"`
}

type Entry struct {
	Symbol    string `json:"symbol"`
	Kind      string `json:"kind"`
	Recv      string `json:"recv,omitempty"`
	Signature string `json:"signature"`
	Summary   string `json:"summary"`
	Example   string `json:"example,omitempty"`
	HasCap    bool   `json:"has_cap"`
	Cap       Cap    `json:"cap"`
}

type Status string

const (
	StatusSupported   Status = "supported"
	StatusUnsupported Status = "unsupported"
	StatusUnknown     Status = "unknown"
)

type LookupResult struct {
	Query  string
	Status Status
	Entry  Entry
}

type Index struct {
	entries []Entry
}

func Load(path string) (*Index, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return New(entries), nil
}

func New(entries []Entry) *Index {
	cp := append([]Entry(nil), entries...)
	return &Index{entries: cp}
}

func (i *Index) Query(term string) []Entry {
	if i == nil {
		return nil
	}
	term = strings.ToLower(strings.TrimSpace(term))
	var out []Entry
	for _, e := range taggedSorted(i.entries) {
		if term == "" || matchEntry(e, term) {
			out = append(out, e)
		}
	}
	return out
}

func (i *Index) Lookup(query string) LookupResult {
	result := LookupResult{Query: query, Status: StatusUnknown}
	hits := i.Query(query)
	if len(hits) == 0 {
		return result
	}
	result.Entry = bestHit(hits, query)
	switch result.Entry.Cap.Tier {
	case "stable", "alpha":
		result.Status = StatusSupported
	case "planned", "missing", "negative":
		result.Status = StatusUnsupported
	default:
		result.Status = StatusUnknown
	}
	return result
}

func taggedSorted(entries []Entry) []Entry {
	var out []Entry
	for _, e := range entries {
		if e.HasCap {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Cap.Domain != out[j].Cap.Domain {
			return out[i].Cap.Domain < out[j].Cap.Domain
		}
		if symbolName(out[i]) != symbolName(out[j]) {
			return symbolName(out[i]) < symbolName(out[j])
		}
		return out[i].Summary < out[j].Summary
	})
	return out
}

func matchEntry(e Entry, term string) bool {
	if strings.Contains(strings.ToLower(symbolName(e)), term) ||
		strings.Contains(strings.ToLower(e.Symbol), term) ||
		strings.Contains(strings.ToLower(e.Cap.Domain), term) ||
		strings.Contains(strings.ToLower(e.Summary), term) {
		return true
	}
	for _, a := range e.Cap.Alias {
		if strings.Contains(strings.ToLower(a), term) {
			return true
		}
	}
	return false
}

func bestHit(hits []Entry, query string) Entry {
	term := strings.ToLower(strings.TrimSpace(query))
	for _, e := range hits {
		if strings.ToLower(symbolName(e)) == term || strings.ToLower(e.Symbol) == term {
			return e
		}
	}
	for _, e := range hits {
		for _, a := range e.Cap.Alias {
			if strings.ToLower(a) == term {
				return e
			}
		}
	}
	return hits[0]
}

func symbolName(e Entry) string {
	if e.Recv != "" {
		return strings.TrimPrefix(e.Recv, "*") + "." + e.Symbol
	}
	return e.Symbol
}
