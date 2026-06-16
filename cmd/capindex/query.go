package main

import "strings"

// query returns tagged entries matching term (case-insensitive substring) on
// symbol, domain, summary, or any alias. Empty term returns all tagged entries.
func query(entries []Entry, term string) []Entry {
	term = strings.ToLower(strings.TrimSpace(term))
	var out []Entry
	for _, e := range taggedSorted(entries) {
		if term == "" || matchEntry(e, term) {
			out = append(out, e)
		}
	}
	return out
}

func matchEntry(e Entry, term string) bool {
	if strings.Contains(strings.ToLower(e.Symbol), term) ||
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
