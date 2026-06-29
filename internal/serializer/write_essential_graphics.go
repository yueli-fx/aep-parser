package serializer

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/rifx"
)

// write_essential_graphics.go — Essential Graphics panel writes on the comp's
// Item-level CIF* lists. AE persists three byte-identical panel generations
// (CIFO / CIF2 / CIF3) per comp; every generation carries the template name
// twice: in the localized CpS2 (value Utf8 + locale Utf8) and in the CapS
// caption (CsCt + CapL + value Utf8). A rename must hit all six Utf8 slots or
// AE shows stale names depending on which generation a given AE version reads.

// egGenerations returns the comp Item's CIF* panel-generation lists, oldest
// first. Empty when the comp has no EG panel shell.
func egGenerations(item *rifx.Chunk) []*rifx.Chunk {
	var gens []*rifx.Chunk
	for _, ft := range []rifx.ChunkID{rifx.IDCifO, rifx.IDCif2, rifx.IDCif3} {
		if g := item.FindFirstList(ft); g != nil {
			gens = append(gens, g)
		}
	}
	return gens
}

func (b *compositionBackrefs) SetMotionGraphicsTemplateName(name string) error {
	if b == nil || b.itemList == nil {
		return fmt.Errorf("comp %q: no Item LIST reference (built outside parser?)", b.compName)
	}
	if name == "" {
		return fmt.Errorf("comp %q: Motion Graphics template name must be non-empty", b.compName)
	}
	gens := egGenerations(b.itemList)
	if len(gens) == 0 {
		return fmt.Errorf("comp %q: no Essential Graphics panel shell (CIFO/CIF2/CIF3) in Item LIST", b.compName)
	}
	// Resolve all six Utf8 slots before mutating any, so a malformed
	// generation cannot leave the panel half-renamed.
	var slots []*rifx.Chunk
	for _, gen := range gens {
		cps2 := gen.FindFirstList(rifx.IDCpS2)
		if cps2 == nil {
			return fmt.Errorf("comp %q: EG %s has no CpS2 template-name list", b.compName, gen.FormType)
		}
		val := cps2.FindFirst(rifx.IDUtf8)
		if val == nil {
			return fmt.Errorf("comp %q: EG %s CpS2 has no Utf8 value", b.compName, gen.FormType)
		}
		caps := gen.FindFirstList(rifx.IDCapS)
		if caps == nil {
			return fmt.Errorf("comp %q: EG %s has no CapS caption list", b.compName, gen.FormType)
		}
		capVal := caps.FindFirst(rifx.IDUtf8)
		if capVal == nil {
			return fmt.Errorf("comp %q: EG %s CapS has no Utf8 value", b.compName, gen.FormType)
		}
		slots = append(slots, val, capVal)
	}
	for _, s := range slots {
		s.Data = []byte(name)
	}
	return nil
}
