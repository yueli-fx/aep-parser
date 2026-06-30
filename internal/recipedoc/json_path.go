package recipedoc

import "strings"

func jsonPathExists(value any, path string) bool {
	if path == "" {
		return true
	}
	parts := strings.SplitN(path, ".", 2)
	head := parts[0]
	tail := ""
	if len(parts) == 2 {
		tail = parts[1]
	}
	if strings.HasSuffix(head, "[]") {
		key := strings.TrimSuffix(head, "[]")
		obj, ok := value.(map[string]any)
		if !ok {
			return false
		}
		items, ok := obj[key].([]any)
		if !ok || len(items) == 0 {
			return false
		}
		if tail == "" {
			return true
		}
		for _, item := range items {
			if jsonPathExists(item, tail) {
				return true
			}
		}
		return false
	}
	obj, ok := value.(map[string]any)
	if !ok {
		return false
	}
	next, ok := obj[head]
	if !ok {
		return false
	}
	if tail == "" {
		return true
	}
	return jsonPathExists(next, tail)
}
