package serializer

import (
	"fmt"
	"math"

	"golang.org/x/text/encoding/simplifiedchinese"

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

// PseudoControl is one control in a from-scratch pseudo effect. The optional
// fields customize the control's pard defaults (RE'd from a real Pseudo Effect
// Maker output); their zero values reproduce AE's plain type defaults, so a
// bare {Kind, Name} keeps the previous behavior.
type PseudoControl struct {
	Kind PseudoControlKind
	Name string // the control's label in AE's Effect Controls

	// Slider (PseudoSlider): visible + valid range and initial value. When
	// Max <= Min the range falls back to 0..100. Default is clamped into range.
	Min, Max float64
	Default  float64 // Slider initial value; Angle initial value (degrees).

	// Checkbox (PseudoCheckbox): initial state.
	Checked bool

	// Color (PseudoColor): default as RGBA in 0..1. nil → white. Length-4
	// slices only; out-of-range components are clamped to [0,1].
	Color []float64
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
			// @0x38 ARGB last + @0x3C ARGB default (RE'd offsets — an earlier
			// draft wrote @0x40/@0x44, which AE silently ignored). nil → white.
			argb := colorToARGB(c.Color)
			bePutU32(d[0x38:], argb)
			bePutU32(d[0x3C:], argb)
		case PseudoPoint:
			// RE'd structural defaults (@0x3C, @0x48); actual position lives in
			// the value entry. AE rejects a point pard without these ("range has
			// no values").
			bePutU32(d[0x3C:], 0x00050000)
			bePutU32(d[0x48:], 0x00640000)
		case PseudoSlider:
			// @0x04 slider flag; @0x38 f8 default; @0x68/@0x6C f4 valid range;
			// @0x70/@0x74 f4 visible range; @0x78 f4 default; @0x7C precision/
			// display flags (constant 0x00050003 in AE-authored sliders).
			lo, hi := c.Min, c.Max
			if hi <= lo {
				lo, hi = 0, 100
			}
			def := c.Default
			if def < lo {
				def = lo
			} else if def > hi {
				def = hi
			}
			bePutU32(d[0x04:], 0x00000200)
			bePutF64(d[0x38:], def)
			bePutF32(d[0x68:], float32(lo))
			bePutF32(d[0x6C:], float32(hi))
			bePutF32(d[0x70:], float32(lo))
			bePutF32(d[0x74:], float32(hi))
			bePutF32(d[0x78:], float32(def))
			bePutU32(d[0x7C:], 0x00050003)
		case PseudoAngle:
			// @0x38 last + @0x3C default, both s4 degrees in 16.16 fixed point.
			fx := uint32(int32(c.Default * 65536))
			bePutU32(d[0x38:], fx)
			bePutU32(d[0x3C:], fx)
		case PseudoCheckbox:
			// @0x38 u32 last + @0x3C u8 default — 1 = checked.
			if c.Checked {
				bePutU32(d[0x38:], 1)
				d[0x3C] = 1
			}
		case PseudoPoint3D:
			// origin default — already zero.
		}
	}), nil
}

// synthPard builds a 148-byte pard: control_type at @0x0F, name at @0x10
// (32 bytes NUL-padded), then a type-specific writer over the zeroed body.
//
// The name is encoded with pardNameBytes — AE reads pard @0x10 in the viewing
// machine's system ANSI codepage (NOT UTF-8), so a CJK label is GBK-encoded to
// match AE's own Pseudo Effect Maker output byte-for-byte and display correctly
// on a simplified-Chinese (GBK) system. ASCII passes through unchanged. See
// incidents/pseudo-control-label-ansi-codepage.md.
func synthPard(controlType byte, name string, body func(d []byte)) *rifx.Chunk {
	d := make([]byte, 148)
	d[0x0F] = controlType
	nb := pardNameBytes(name)
	if len(nb) > 31 {
		nb = nb[:31]
	}
	copy(d[0x10:0x30], nb)
	if body != nil {
		body(d)
	}
	return &rifx.Chunk{ID: rifx.IDpard, Data: d}
}

// pardNameBytes encodes a control label for the pard @0x10 name field. AE
// decodes this field in the system ANSI codepage, so we GBK-encode (GBK is an
// ASCII superset — ASCII labels are byte-identical) to match AE's native output
// on simplified-Chinese systems. If a rune is not GBK-representable the raw
// UTF-8 bytes are kept (best-effort; it will mojibake, but nothing is dropped).
func pardNameBytes(name string) []byte {
	if name == "" {
		return nil
	}
	if b, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(name)); err == nil {
		return b
	}
	return []byte(name)
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

func bePutF64(b []byte, f float64) {
	v := math.Float64bits(f)
	for i := 0; i < 8; i++ {
		b[i] = byte(v >> (56 - 8*i))
	}
}

// colorToARGB packs an RGBA-in-0..1 slice into AE's 0xAARRGGBB pard color word.
// nil (or any malformed length) → opaque white, the AE default.
func colorToARGB(rgba []float64) uint32 {
	if len(rgba) != 4 {
		return 0xffffffff
	}
	clamp := func(f float64) uint32 {
		if f <= 0 {
			return 0
		}
		if f >= 1 {
			return 255
		}
		return uint32(f*255 + 0.5)
	}
	return clamp(rgba[3])<<24 | clamp(rgba[0])<<16 | clamp(rgba[1])<<8 | clamp(rgba[2])
}
