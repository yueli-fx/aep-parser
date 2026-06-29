// internal/aep/new_project.go
package serializer

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/scene"
)

//go:embed templates/project/2020.aep
var embeddedTemplate2020 []byte

//go:embed templates/project/2022.aep
var embeddedTemplate2022 []byte

//go:embed templates/project/2025.aep
var embeddedTemplate2025 []byte

// NewProject returns a fresh empty Project parsed from the embedded
// AE skeleton matching the requested target.
// (Full contract + RE notes live on the aep.NewProject facade — docgen source.)
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

	// AE's factory default color depth is 8bpc. The embedded AE2020 skeleton was
	// incidentally authored at 32bpc (the 2022/2025 skeletons are already 8bpc);
	// normalize so every fresh project matches AE regardless of which skeleton
	// backs it. (Render-pixel ship-gates already force bitsPerChannel=8 in JSX to
	// work around the old discrepancy — this fixes it at the source.)
	if err := p.SetBitsPerChannel(BPC8); err != nil {
		panic(fmt.Sprintf("aep: failed to normalize fresh project to 8bpc (build bug): %v", err))
	}
	return p
}
