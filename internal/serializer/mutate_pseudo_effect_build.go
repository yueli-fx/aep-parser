package serializer

import (
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/rifx"
)

// PseudoControlKind is the AE control type of one pseudo-effect control.
type PseudoControlKind int

const (
	PseudoSlider   PseudoControlKind = iota // scalar slider
	PseudoColor                             // color swatch
	PseudoCheckbox                          // checkbox
	PseudoAngle                             // angle dial
	PseudoPoint                             // 2D point
	PseudoPoint3D                           // 3D point
)

// pardControlType maps a kind to the pard control_type byte (@0x0F), RE'd from a
// real Pseudo Effect Maker output (test_data/pseudo_rich_demo.aep): angle 0x03,
// checkbox 0x04, color 0x05, point 0x06, slider 0x0a, 3D point 0x12.
var pardControlType = map[PseudoControlKind]byte{
	PseudoAngle:    0x03,
	PseudoCheckbox: 0x04,
	PseudoColor:    0x05,
	PseudoPoint:    0x06,
	PseudoSlider:   0x0a,
	PseudoPoint3D:  0x12,
}

// PseudoControl is one control in a from-scratch pseudo effect.
type PseudoControl struct {
	Kind PseudoControlKind
	Name string // the control's label in AE's Effect Controls
}

// BuildPseudoEffect constructs a pseudo effect entirely in Go — no .ffx, no AE,
// and no cloned template bytes: every pard is synthesized field-by-field from
// the RE'd pard layout — and splices it into the layer's Effect Parade. uid is
// the per-effect unique id ("Pseudo/<uid>/<name>"); name is the match-name
// segment; displayName is the effect label; controls are its controls in order.
// Controls take their type defaults. Returns the parsed *Effect.
//
// (Full contract lives on the aep.BuildPseudoEffect facade.)
func BuildPseudoEffect(layer *Layer, uid, name, displayName string, controls []PseudoControl) (*Effect, error) {
	if layer == nil {
		return nil, fmt.Errorf("BuildPseudoEffect: layer is nil")
	}
	if uid == "" || name == "" {
		return nil, fmt.Errorf("BuildPseudoEffect: uid and name are required")
	}
	matchName := "Pseudo/" + uid + "/" + name
	label := displayName
	if label == "" {
		label = name
	}
	sspc, err := synthPseudoSspc(matchName, label, controls)
	if err != nil {
		return nil, err
	}
	return addEffectFromChunks(layer, "BuildPseudoEffect", matchName, makeTdmn(matchName), sspc)
}

// synthPseudoSspc assembles the in-parade sspc for a pseudo effect entirely from
// synthesized chunks: fnam(Utf8) + parT(header pard + control pards + built-in
// pard) + tdgp(scaffold + built-in value group, controls at defaults) + pgui.
func synthPseudoSspc(matchName, label string, controls []PseudoControl) (*rifx.Chunk, error) {
	parT := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDparT}
	parT.Children = append(parT.Children, makeParn(0)) // count patched below
	// -0000 effect header pard: control_type 0, struct-field (@0x30) = 2.
	parT.Children = append(parT.Children,
		makeTdmn(matchName+"-0000"), synthPard(0x00, "", func(d []byte) { bePutU32(d[0x30:], 2) }))
	for i, c := range controls {
		pard, err := synthControlPard(c)
		if err != nil {
			return nil, err
		}
		parT.Children = append(parT.Children, makeTdmn(fmt.Sprintf("%s-%04d", matchName, i+1)), pard)
		// A Checkbox carries its on/off label as a trailing pdnm string (AE:
		// "PF_ParamCheckbox must have nameptr set" without it).
		if c.Kind == PseudoCheckbox {
			parT.Children = append(parT.Children,
				&rifx.Chunk{ID: rifx.IDPdnm, Data: utf8StringData(c.Name)})
		}
	}
	// Built-in "Compositing Options" pard (control_type 0x09, otherwise empty).
	parT.Children = append(parT.Children,
		makeTdmn("ADBE Effect Built In Params"), synthPard(0x09, "", nil))
	if parn := childByID(parT, rifx.IDParn); parn != nil {
		var n uint32
		for _, ch := range parT.Children {
			if ch.ID == rifx.IDpard {
				n++
			}
		}
		bePutU32(parn.Data, n)
	}

	// tdgp value group: scaffold (tdsb + tdsn label) + built-in value group +
	// Group End — controls at pard defaults (the AE-accepted all-defaults form).
	biValTdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{
		makeTdsb(), makeTdsn(""), makeTdmn("ADBE Group End"),
	}}
	valTdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{
		makeTdsb(), makeTdsn(label),
		makeTdmn("ADBE Effect Built In Params"), biValTdgp,
		makeTdmn("ADBE Group End"),
	}}

	fnam := &rifx.Chunk{ID: rifx.IDFnam, Data: utf8StringData(label)}
	pgui := &rifx.Chunk{ID: rifx.IDPgui, Data: make([]byte, 16)}

	return &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDSspc, Children: []*rifx.Chunk{
		fnam, parT, valTdgp, pgui,
	}}, nil
}

// synthControlPard synthesizes one control's pard from the RE'd per-type layout.
func synthControlPard(c PseudoControl) (*rifx.Chunk, error) {
	ct, ok := pardControlType[c.Kind]
	if !ok {
		return nil, fmt.Errorf("BuildPseudoEffect: unsupported control kind %d", c.Kind)
	}
	return synthPard(ct, c.Name, func(d []byte) {
		switch c.Kind {
		case PseudoColor:
			// @0x40 ARGB last + @0x44 ARGB default — white by default.
			bePutU32(d[0x40:], 0xffffffff)
			bePutU32(d[0x44:], 0xffffffff)
		case PseudoPoint:
			// RE'd structural defaults (@0x3C, @0x48); actual position lives in
			// the value entry. AE rejects a point pard without these ("range has
			// no values").
			bePutU32(d[0x3C:], 0x00050000)
			bePutU32(d[0x48:], 0x00640000)
		case PseudoSlider:
			// @0x04 slider flag; @0x38 f8 default(0); @0x68/@0x6C f4 valid range;
			// @0x70/@0x74 f4 slider range; @0x78 f4 default.
			bePutU32(d[0x04:], 0x00000200)
			bePutF32(d[0x68:], -1000)
			bePutF32(d[0x6C:], 1000)
			bePutF32(d[0x70:], 0)
			bePutF32(d[0x74:], 100)
			bePutF32(d[0x78:], 0)
		case PseudoAngle, PseudoCheckbox, PseudoPoint3D:
			// default 0 / unchecked / origin — already zero.
		}
	}), nil
}

// synthPard builds a 148-byte pard: control_type at @0x0F, name at @0x10
// (32 bytes NUL-padded), then a type-specific writer over the zeroed body.
func synthPard(controlType byte, name string, body func(d []byte)) *rifx.Chunk {
	d := make([]byte, 148)
	d[0x0F] = controlType
	nb := []byte(name)
	if len(nb) > 31 {
		nb = nb[:31]
	}
	copy(d[0x10:0x30], nb)
	if body != nil {
		body(d)
	}
	return &rifx.Chunk{ID: rifx.IDpard, Data: d}
}

func makeParn(n uint32) *rifx.Chunk {
	d := make([]byte, 4)
	bePutU32(d, n)
	return &rifx.Chunk{ID: rifx.IDParn, Data: d}
}

func bePutU32(b []byte, v uint32) {
	b[0], b[1], b[2], b[3] = byte(v>>24), byte(v>>16), byte(v>>8), byte(v)
}

func bePutF32(b []byte, f float32) {
	bePutU32(b, math.Float32bits(f))
}
