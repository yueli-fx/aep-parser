package serializer

import (
	"bytes"
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// ffxFormType is the RIFX form-type of an After Effects Animation Preset (.ffx)
// — "FaFX", distinct from a project's "Egg!". rifx.ReadChunk reads it without a
// form-type gate (unlike rifx.Parse, which requires Egg!).
var ffxFormType = rifx.ChunkID{'F', 'a', 'F', 'X'}

// ApplyPseudoEffect splices the pseudo effect carried by an Animation Preset
// (.ffx) into the layer's Effect Parade — the offline, pure-Go equivalent of
// AE's applyPreset, requiring no running AE. ffxBytes is the raw .ffx file
// content (RIFX "FaFX" form, e.g. a Pseudo Effect Maker export). Returns the
// parsed *Effect so the caller can tune its parameters.
//
// (Full contract + RE notes live on the aep.ApplyPseudoEffect facade.)
func ApplyPseudoEffect(layer *Layer, ffxBytes []byte) (*Effect, error) {
	if layer == nil {
		return nil, fmt.Errorf("ApplyPseudoEffect: layer is nil")
	}
	matchName, tdmnCh, sspcCh, err := extractPseudoEffectUnit(ffxBytes)
	if err != nil {
		return nil, err
	}
	return addEffectFromChunks(layer, "ApplyPseudoEffect", matchName, tdmnCh, sspcCh)
}

// extractPseudoEffectUnit parses a .ffx (RIFX "FaFX") and pulls out the
// effect-unit AE would splice into a layer's Effect Parade: the bare match-name
// (a fresh tdmn chunk) and a deep clone of the effect's LIST:sspc payload. The
// preset wrapper (head / besc descriptor / tdsp path steps / pgui) is dropped —
// only the (tdmn, sspc) pair is a parade member.
//
// .ffx layout (Scribe sample, RE 2026-06-20):
//
//	RIFX "FaFX"
//	├─ head
//	└─ LIST besc
//	   ├─ beso
//	   ├─ LIST tdsp  → tdsi "ADBE Effect Parade", tdsi "<matchName>"  ← bare match-name
//	   ├─ tdsn
//	   ├─ LIST tdsp  → tdsi "ADBE End of path sentinel"
//	   └─ LIST sspc  (fnam + parT param-defs + tdgp values)           ← the effect-unit
func extractPseudoEffectUnit(ffxBytes []byte) (matchName string, tdmn, sspc *rifx.Chunk, err error) {
	root, e := rifx.ReadChunk(bytes.NewReader(ffxBytes))
	if e != nil {
		return "", nil, nil, fmt.Errorf("ApplyPseudoEffect: parse .ffx: %w", e)
	}
	if root.ID != rifx.IDRifx || root.FormType != ffxFormType {
		return "", nil, nil, fmt.Errorf("ApplyPseudoEffect: not an Animation Preset (.ffx): want RIFX/FaFX, got %q/%q", root.ID, root.FormType)
	}

	besc := childByForm(root, rifx.IDBesc)
	if besc == nil {
		return "", nil, nil, fmt.Errorf("ApplyPseudoEffect: .ffx has no besc descriptor")
	}
	sspc = childByForm(besc, rifx.IDSspc)
	if sspc == nil {
		return "", nil, nil, fmt.Errorf("ApplyPseudoEffect: .ffx besc has no sspc effect-unit")
	}

	// The bare match-name is the non-boilerplate tdsi step in the first tdsp
	// path descriptor (the other steps are "ADBE Effect Parade" / sentinels).
	for _, ch := range besc.Children {
		if !ch.IsList() || ch.FormType != rifx.IDTdsp {
			continue
		}
		for _, step := range ch.Children {
			if !step.IsList() || step.FormType != rifx.IDTdsi {
				continue
			}
			mn := tdmnString(step)
			if mn != "" && mn != "ADBE Effect Parade" && mn != "ADBE End of path sentinel" {
				matchName = mn
			}
		}
		if matchName != "" {
			break
		}
	}
	if matchName == "" {
		return "", nil, nil, fmt.Errorf("ApplyPseudoEffect: could not find pseudo effect match-name in .ffx path descriptor")
	}

	return matchName, makeTdmn(matchName), deepCloneChunk(sspc), nil
}

// childByForm returns the first LIST child of c whose form-type equals form.
func childByForm(c *rifx.Chunk, form rifx.ChunkID) *rifx.Chunk {
	for _, ch := range c.Children {
		if ch.IsList() && ch.FormType == form {
			return ch
		}
	}
	return nil
}

// tdmnString returns the NUL-trimmed match-name of the first tdmn child of c
// (empty if none).
func tdmnString(c *rifx.Chunk) string {
	for _, ch := range c.Children {
		if ch.ID == rifx.IDTdmn {
			return string(bytes.TrimRight(ch.Data, "\x00"))
		}
	}
	return ""
}
