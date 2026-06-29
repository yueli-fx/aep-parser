// Mask option materialization (synthesis-insert) for Mask Feather / Opacity /
// Expansion — the default-elided scalar leaves inside a mask atom's property
// group. AE elides them at default, so a from-scratch mask (AddMask) and most
// parsed masks carry no slot to overwrite. SetMaskOption clones the requested
// AE-native leaf from an embedded template (extracted from a re-saved
// mask-options fixture — masks are eagerly decoded so the bytes must be exact),
// splices it into the atom group in AE's canonical order (Feather, Opacity,
// Offset), and overwrites its cdat with the caller's value. When the leaf is
// already present (option authored or a prior splice) it just overwrites.
//
// Exposed through the scene Mask.SetFeather / SetOpacity / SetExpansion methods.
package serializer

import (
	"bytes"
	_ "embed"
	"fmt"
	"sync"

	"github.com/yueli-fx/aep-parser/internal/rifx"
)

//go:embed templates/options/mask_option_leaves.bin
var maskOptionLeavesBytes []byte

// maskOptionOrder is AE's canonical child order for the option leaves inside a
// mask atom (= extraction order in mask_option_leaves.bin). Mask Shape (ordinal
// −1 implicitly) always precedes them; Group End follows.
var maskOptionOrder = map[string]int{
	"ADBE Mask Feather": 0,
	"ADBE Mask Opacity": 1,
	"ADBE Mask Offset":  2,
}

var (
	maskOptionLeavesWrapper *rifx.Chunk
	maskOptionLeavesOnce    sync.Once
	maskOptionLeavesErr     error
)

// cloneMaskOptionLeaf returns fresh (tdmn, tdbs) chunks for matchName from the
// embedded mask-option template, or an error if matchName is unknown.
func cloneMaskOptionLeaf(matchName string) (tdmn, tdbs *rifx.Chunk, err error) {
	maskOptionLeavesOnce.Do(func() {
		maskOptionLeavesWrapper, maskOptionLeavesErr = rifx.ReadChunk(bytes.NewReader(maskOptionLeavesBytes))
	})
	if maskOptionLeavesErr != nil {
		return nil, nil, fmt.Errorf("parse mask option leaves: %w", maskOptionLeavesErr)
	}
	kids := maskOptionLeavesWrapper.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == matchName &&
			kids[i+1].IsList() && kids[i+1].FormType == rifx.IDTdbs {
			return cloneChunk(kids[i]), cloneChunk(kids[i+1]), nil
		}
	}
	return nil, nil, fmt.Errorf("mask option leaf %q not in template", matchName)
}

// encodeMaskOptionValue returns the cdat bytes for a mask option value: Feather
// is [2]float64 (X,Y px → 16 bytes), Opacity (0..1) and Offset/Expansion (px)
// are float64 (8 bytes).
func encodeMaskOptionValue(matchName string, value any) ([]byte, error) {
	switch matchName {
	case "ADBE Mask Feather":
		v, ok := value.([2]float64)
		if !ok {
			return nil, fmt.Errorf("SetMaskOption %q: want [2]float64, got %T", matchName, value)
		}
		return encodeF64sBE(v[0], v[1]), nil
	case "ADBE Mask Opacity", "ADBE Mask Offset":
		v, ok := value.(float64)
		if !ok {
			return nil, fmt.Errorf("SetMaskOption %q: want float64, got %T", matchName, value)
		}
		return encodeF64sBE(v), nil
	default:
		return nil, fmt.Errorf("SetMaskOption: %q is not a mask option", matchName)
	}
}

// SetMaskOption (MaskWriter) overwrites or splices a mask option leaf
// (Feather/Opacity/Offset) in the mask atom group and writes value's cdat.
func (b *maskBackrefs) SetMaskOption(matchName string, value any) error {
	if b.atomTdgp == nil {
		return fmt.Errorf("mask %q: no atom group chunk (built outside parser? round-trip through Reopen first)", b.maskName)
	}
	myOrd, ok := maskOptionOrder[matchName]
	if !ok {
		return fmt.Errorf("SetMaskOption: %q is not a mask option", matchName)
	}
	data, err := encodeMaskOptionValue(matchName, value)
	if err != nil {
		return err
	}

	// Already present → overwrite the cdat in place.
	if tdbs := groupLeafTdbs(b.atomTdgp, matchName); tdbs != nil {
		writeTdbsCdat(tdbs, data)
		return nil
	}

	// Splice the AE-native leaf in canonical order: before the first existing
	// option leaf with a higher ordinal, else before Group End.
	tdmnCh, tdbsCh, err := cloneMaskOptionLeaf(matchName)
	if err != nil {
		return err
	}
	writeTdbsCdat(tdbsCh, data)

	kids := b.atomTdgp.Children
	insertIdx := len(kids)
	for i := 0; i < len(kids); i++ {
		if kids[i].ID != rifx.IDTdmn {
			continue
		}
		name := trimChunkNUL(kids[i].Data)
		if name == "ADBE Group End" {
			insertIdx = i
			break
		}
		if ord, ok := maskOptionOrder[name]; ok && ord > myOrd {
			insertIdx = i
			break
		}
	}
	spliced := make([]*rifx.Chunk, 0, len(kids)+2)
	spliced = append(spliced, kids[:insertIdx]...)
	spliced = append(spliced, tdmnCh, tdbsCh)
	spliced = append(spliced, kids[insertIdx:]...)
	b.atomTdgp.Children = spliced
	return nil
}

// groupLeafTdbs returns the LIST(tdbs) following the tdmn matching matchName in
// group.Children, or nil.
func groupLeafTdbs(group *rifx.Chunk, matchName string) *rifx.Chunk {
	kids := group.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == matchName &&
			kids[i+1].IsList() && kids[i+1].FormType == rifx.IDTdbs {
			return kids[i+1]
		}
	}
	return nil
}

// writeTdbsCdat overwrites the first len(data) bytes of the cdat inside a tdbs.
func writeTdbsCdat(tdbs *rifx.Chunk, data []byte) {
	for _, ch := range tdbs.Children {
		if ch.ID == rifx.IDCdat && len(ch.Data) >= len(data) {
			copy(ch.Data[:len(data)], data)
			return
		}
	}
}
