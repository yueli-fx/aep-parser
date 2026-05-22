// internal/aep/new_project.go
package aep

import (
	_ "embed"
)

// AETarget selects which AE version's empty-project skeleton NewProject
// uses as the base. Output .aep files identify themselves as that AE
// version's format (via the `svap` chunk). Newer AE versions may silently
// upgrade them; older AE versions refuse to open files claiming a newer
// version.
//
// Underlying values are AE marketing years (2020/2022/2025/…) for
// debuggable panic messages and natural `target >= 2025` comparisons.
//
// AETarget values are NOT forward-compatible — unknown values panic.
// Upgrade the library when targeting a newer AE version.
type AETarget int

const (
	TargetAE2020 AETarget = 2020 // default; max compatibility (any AE 2020+ opens)
	TargetAE2022 AETarget = 2022 // 30 chunks; adds AE 24+ color-mgmt prefs
	TargetAE2025 AETarget = 2025 // 30 chunks; latest tested
)

//go:embed templates/2020.aep
var embeddedTemplate2020 []byte

//go:embed templates/2022.aep
var embeddedTemplate2022 []byte

//go:embed templates/2025.aep
var embeddedTemplate2025 []byte
