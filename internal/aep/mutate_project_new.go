// internal/aep/new_project.go
package aep

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/example/aep-parser/internal/scene"
)

//go:embed templates/2020.aep
var embeddedTemplate2020 []byte

//go:embed templates/2022.aep
var embeddedTemplate2022 []byte

//go:embed templates/2025.aep
var embeddedTemplate2025 []byte

// NewProject returns a fresh empty Project parsed from the embedded
// AE skeleton matching the requested target.
//
// Optional target arg: zero args = TargetAE2020 (max compatibility). Pass
// at most one target. Subsequent NewComposition calls populate it.
//
// Never returns an error: the embedded templates are build-time trusted;
// parser bugs panic with a "build bug" message (not user-facing).
// Panics on: multiple target args, or unknown AETarget value (forward-incompat).
func NewProject(target ...AETarget) *Project {
	t := TargetAE2020 // default
	switch len(target) {
	case 0:
		// use default
	case 1:
		t = target[0]
	default:
		panic(fmt.Sprintf("aep: NewProject accepts at most one target, got %d", len(target)))
	}

	var tmpl []byte
	switch t {
	case TargetAE2020:
		tmpl = embeddedTemplate2020
	case TargetAE2022:
		tmpl = embeddedTemplate2022
	case TargetAE2025:
		tmpl = embeddedTemplate2025
	default:
		panic(fmt.Sprintf("aep: unknown AETarget %d (forward-incompat; upgrade library)", int(t)))
	}

	p, err := FromReader(bytes.NewReader(tmpl))
	if err != nil {
		panic(fmt.Sprintf("aep: corrupt embedded template for AE %d (build bug): %v", int(t), err))
	}
	scene.SetProjectTarget(p, t)
	return p
}
