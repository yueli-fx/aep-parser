package serializer

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/rifx"
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
	return ApplyPseudoEffectNamed(layer, ffxBytes, "")
}

// ApplyPseudoEffectNamed is ApplyPseudoEffect with a custom effect-instance
// display name (the label shown in AE's Effect Controls / timeline). displayName
// may be any UTF-8 string — including CJK like "伪效果" — because the name is
// written as AE's "Utf8" + byte-length + bytes sub-record (the length is in
// bytes, not runes, so multi-byte names round-trip cleanly). An empty
// displayName keeps the .ffx's own name. The match-name (the AE lookup key,
// "Pseudo/<uID>/<name>") is unaffected and stays ASCII.
func ApplyPseudoEffectNamed(layer *Layer, ffxBytes []byte, displayName string) (*Effect, error) {
	if layer == nil {
		return nil, fmt.Errorf("ApplyPseudoEffect: layer is nil")
	}
	matchName, tdmnCh, sspcCh, err := extractPseudoEffectUnit(ffxBytes)
	if err != nil {
		return nil, err
	}
	if err := transformPseudoSspcToInParade(sspcCh, displayName); err != nil {
		return nil, err
	}
	return addEffectFromChunks(layer, "ApplyPseudoEffect", matchName, tdmnCh, sspcCh)
}

// transformPseudoSspcToInParade rewrites a pseudo effect's sspc payload — read
// verbatim from a preset .ffx — into the form AE expects for an effect living
// inside a layer's Effect Parade. A .ffx is a not-yet-applied preset; every
// effect *instance* in a parade additionally carries the universal "ADBE Effect
// Built In Params" group (Compositing Options) that the preset lacks. AE adds it
// when it applies the preset; offline we splice the same constant blocks
// (sourced from any native effect template, which are AE-baked in-parade form)
// into the sspc's parT (param defs) and tdgp (param values), each just before
// their trailing Group End sentinel.
func transformPseudoSspcToInParade(sspc *rifx.Chunk, displayName string) error {
	parT := childByForm(sspc, rifx.IDparT)
	if parT == nil {
		return fmt.Errorf("ApplyPseudoEffect: pseudo sspc has no parT param-defs")
	}
	valTdgp := childByForm(sspc, rifx.IDTdgp)
	if valTdgp == nil {
		return fmt.Errorf("ApplyPseudoEffect: pseudo sspc has no tdgp values group")
	}

	// A .ffx (FaFX preset) stores name strings raw (fixed-width / length-by-size);
	// an Egg! project stores them Utf8-wrapped ("Utf8" + u32 len + bytes). Splicing
	// raw strings makes AE misread the next chunk's length → "file is damaged".
	// Re-encode fnam + every raw tdsn in the sspc tree to the Utf8 form. The fnam
	// (effect-instance display name) takes displayName when non-empty — any UTF-8,
	// incl. CJK (the length is in bytes, so multi-byte names are exact).
	if fnam := childByID(sspc, rifx.IDFnam); fnam != nil {
		name := displayName
		if name == "" {
			name = string(bytes.TrimRight(fnam.Data, "\x00"))
		}
		fnam.Data = utf8StringData(name)
	}
	reencodeRawTdsn(sspc)

	biParTPair, biTdgpPair, err := builtInParamsBlocks()
	if err != nil {
		return err
	}

	// parT: append (tdmn "ADBE Effect Built In Params", pard) at the end (the
	// .ffx parT has no Group End sentinel — it ends on the last param's pard),
	// then bump the parn count header to match the new pard total (AE reports
	// "missing data in file" when parn undercounts the pard entries present).
	parT.Children = append(parT.Children, biParTPair...)
	if parn := childByID(parT, rifx.IDParn); parn != nil && len(parn.Data) >= 4 {
		var pardCount uint32
		for _, ch := range parT.Children {
			if ch.ID == rifx.IDpard {
				pardCount++
			}
		}
		binary.BigEndian.PutUint32(parn.Data[0:4], pardCount)
	}

	// tdgp: AE elides param values left at their pard-defined default, so a
	// freshly-applied pseudo effect's value group carries only the structural
	// scaffold (tdsb + tdsn group name) + the built-in-params group + Group End;
	// the controls themselves come from the parT pard defs. We drop the .ffx's
	// per-param value entries — the effect applies at its pard defaults (AE 2020
	// + AE 2025 reject the spliced .ffx values as "missing data in file"; the
	// all-defaults form is ship-gate green on both). Authored .ffx values are
	// thus NOT yet preserved (a future step would materialize each non-default
	// value into a correctly-laid-out in-parade tdbs).
	scaffold := make([]*rifx.Chunk, 0, 5)
	for _, ch := range valTdgp.Children {
		if ch.ID == rifx.IDTdsb || ch.ID == rifx.IDTdsn {
			// AE shows the effect-instance label from this value-group tdsn
			// (NOT fnam), so the display-name override lands here.
			if ch.ID == rifx.IDTdsn && displayName != "" {
				ch.Data = utf8StringData(displayName)
			}
			scaffold = append(scaffold, ch)
		}
	}
	scaffold = append(scaffold, biTdgpPair...)
	scaffold = append(scaffold, makeTdmn("ADBE Group End"))
	valTdgp.Children = scaffold
	return nil
}

// builtInParamsBlocks returns deep clones of the universal "ADBE Effect Built In
// Params" entries — the parT pair (tdmn + pard) and the tdgp pair (tdmn + LIST
// value group) — extracted from a native effect template (canonical AE-baked
// in-parade bytes). These are identical across effects (Compositing Options).
func builtInParamsBlocks() (parTPair, tdgpPair []*rifx.Chunk, err error) {
	_, sspc, e := cloneEffectTemplate(EffectFill)
	if e != nil {
		return nil, nil, fmt.Errorf("ApplyPseudoEffect: load built-in-params reference template: %w", e)
	}
	parT := childByForm(sspc, rifx.IDparT)
	valTdgp := childByForm(sspc, rifx.IDTdgp)
	if parT == nil || valTdgp == nil {
		return nil, nil, fmt.Errorf("ApplyPseudoEffect: reference template missing parT/tdgp")
	}
	const biName = "ADBE Effect Built In Params"
	if parTPair = pairAfterTdmn(parT.Children, biName); parTPair == nil {
		return nil, nil, fmt.Errorf("ApplyPseudoEffect: reference template parT has no %q", biName)
	}
	if tdgpPair = pairAfterTdmn(valTdgp.Children, biName); tdgpPair == nil {
		return nil, nil, fmt.Errorf("ApplyPseudoEffect: reference template tdgp has no %q", biName)
	}
	return parTPair, tdgpPair, nil
}

// pairAfterTdmn finds the tdmn child whose match-name is name and returns deep
// clones of [that tdmn, the following chunk] (nil if not found / no follower).
func pairAfterTdmn(children []*rifx.Chunk, name string) []*rifx.Chunk {
	for i, ch := range children {
		if ch.ID == rifx.IDTdmn && string(bytes.TrimRight(ch.Data, "\x00")) == name && i+1 < len(children) {
			return []*rifx.Chunk{deepCloneChunk(ch), deepCloneChunk(children[i+1])}
		}
	}
	return nil
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

// utf8StringData encodes s as AE's in-project string sub-record: the 4-byte
// "Utf8" tag + a big-endian u32 length + the raw bytes.
func utf8StringData(s string) []byte {
	b := []byte(s)
	data := make([]byte, 8+len(b))
	copy(data[0:4], "Utf8")
	binary.BigEndian.PutUint32(data[4:8], uint32(len(b)))
	copy(data[8:], b)
	return data
}

// reencodeRawTdsn walks c and rewrites every tdsn whose payload is a raw string
// (not already a "Utf8" sub-record) into the Utf8 form. Built-in-params blocks
// spliced from native templates are already Utf8 and are left untouched.
func reencodeRawTdsn(c *rifx.Chunk) {
	if c.ID == rifx.IDTdsn && !bytes.HasPrefix(c.Data, []byte("Utf8")) {
		c.Data = utf8StringData(string(bytes.TrimRight(c.Data, "\x00")))
	}
	for _, ch := range c.Children {
		reencodeRawTdsn(ch)
	}
}

// childByID returns the first child of c with the given leaf chunk ID.
func childByID(c *rifx.Chunk, id rifx.ChunkID) *rifx.Chunk {
	for _, ch := range c.Children {
		if ch.ID == id {
			return ch
		}
	}
	return nil
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
