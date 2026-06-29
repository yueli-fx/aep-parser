package serializer

import (
	"encoding/binary"

	"github.com/yueli-fx/aep-parser/internal/rifx"
)

// parse_essential_graphics.go — decode the Essential Graphics panel from a
// comp's Item-level LIST:CIF3. Mirrors the reference parser's parsers/essential_graphics.py.
//
//	LIST:CIF3
//	  LIST:CpS2 → Utf8           ← template name ("Untitled" default)
//	  LIST:CCtl (× N)            ← one per controller, panel order
//	    Utf8                     ← uuid (direct child, 36-char GUID)
//	    CTyp (u4)                ← controller type code
//	    LIST:CpS2 → Utf8         ← controller display name
//
// CIF3 is the most complete EG version AE writes (older CIF1/CIF2 ignored).

const defaultMogrtTemplateName = "Untitled"

// parseEssentialGraphics reads the EG template name + controllers from a
// comp's owning Item LIST. Returns the AE default name ("Untitled") and no
// controllers when the comp has no CIF3 (no EG panel).
func parseEssentialGraphics(item *rifx.Chunk) (string, []*EssentialGraphicsController) {
	cif3 := item.FindFirstList(rifx.IDCif3)
	if cif3 == nil {
		return defaultMogrtTemplateName, nil
	}

	name := defaultMogrtTemplateName
	if n := localizedString(cif3); n != "" {
		name = n
	}

	var controllers []*EssentialGraphicsController
	for _, cctl := range cif3.FindAllList(rifx.IDCctl) {
		ctrl := &EssentialGraphicsController{Name: localizedString(cctl)}
		if uuid := cctl.FindFirst(rifx.IDUtf8); uuid != nil {
			ctrl.UUID = uuid.Text()
		}
		if ctyp := cctl.FindFirst(rifx.IDCTyp); ctyp != nil && len(ctyp.Data) >= 4 {
			ctrl.Type = EGControllerType(binary.BigEndian.Uint32(ctyp.Data))
		}
		controllers = append(controllers, ctrl)
	}
	return name, controllers
}

// localizedString returns the first Utf8 value inside the chunk's first
// CpS2 (localized string) child — AE's pattern for names: a CpS2 holding the
// value Utf8 followed by a locale Utf8. Returns "" when absent.
func localizedString(parent *rifx.Chunk) string {
	cps2 := parent.FindFirstList(rifx.IDCpS2)
	if cps2 == nil {
		return ""
	}
	if u := cps2.FindFirst(rifx.IDUtf8); u != nil {
		return u.Text()
	}
	return ""
}
