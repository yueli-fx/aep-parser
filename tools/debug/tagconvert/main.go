// Command tagconvert rewrites a Go source file's exported-symbol aep:cap
// directives into @tag blocks with placeholders, for the manual conversion pass.
// One-shot, in place: `go run ./tools/debug/tagconvert internal/aep/facade.go`.
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: tagconvert <file.go>")
		os.Exit(2)
	}
	path := os.Args[1]
	src, err := os.ReadFile(path)
	if err != nil {
		fatal(err)
	}
	out, err := convertSource(src)
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("tagconvert: rewrote %s\n", path)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "tagconvert:", err)
	os.Exit(1)
}
