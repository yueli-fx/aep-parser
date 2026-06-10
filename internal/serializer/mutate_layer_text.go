// Public-API entry for text layer creation.
//
// A text layer is source-less like Camera/Light — defined entirely by its ldta
// + the Text Properties group whose btds chunk carries the CoolType PostScript
// text document (btdk). We embed a whole AE-native text Layr (extracted from
// re_text.aep layer "baseline_A": point text "A", default styling) and splice
// it via the same newTemplatedLayer machinery. The btdk blob travels verbatim:
// its layout cache (/1/1[0]/1 — per-line/per-run char counts + glyph pixel
// metrics keyed to the rendered string) is what makes arbitrary-length text
// synthesis hard, and cloning sidesteps it entirely.
package serializer

import (
	_ "embed"

	"github.com/example/aep-parser/internal/scene"
)

//go:embed templates/layer_text_body.bin
var layerTextBodyBytes []byte

// NewTextLayer adds a new text layer to the composition.
// (Full contract lives on the aep.NewTextLayer facade — docgen source.)
func NewTextLayer(c *Composition, name string) (*Layer, error) {
	l, err := newTemplatedLayer(c, name, layerTextBodyBytes, LayerTypeText)
	if err != nil {
		return nil, err
	}
	// Wire the text-source back-ref the way parseLayer does, so the fresh
	// layer reads TextSource and supports the length-preserving SetText
	// immediately — no Reopen needed.
	if lb := layerBack(l); lb != nil {
		if btds := findTextSourceChunk(lb.layrList); btds != nil {
			lb.btdsChunk = btds
			l.TextSourceRaw = btds.Data
			if ts, _ := scene.DecodeTextSource(btds.Data); ts != nil {
				l.TextSource = ts
			}
		}
	}
	return l, nil
}
