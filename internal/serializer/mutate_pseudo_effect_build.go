package serializer

import (
	"encoding/hex"
	"fmt"
	"math"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"

	"github.com/example/aep-parser/internal/rifx"
)

// PseudoControlKind is the AE control type of one pseudo-effect control.
type PseudoControlKind int

const (
	PseudoSlider     PseudoControlKind = iota // scalar slider
	PseudoColor                               // color swatch
	PseudoCheckbox                            // checkbox
	PseudoAngle                               // angle dial
	PseudoPoint                               // 2D point
	PseudoPoint3D                             // 3D point
	PseudoDropdown                            // dropdown menu (Options)
	PseudoGroupStart                          // group start (controls until PseudoGroupEnd nest under it)
	PseudoGroupEnd                            // group end
	PseudoLabel                               // static text label (no value)
	PseudoLayer                               // layer picker
)

// pardControlType maps a kind to the pard control_type byte (@0x0F), RE'd from a
// real Pseudo Effect Maker output (test_data/pseudo_rich_demo.aep): angle 0x03,
// checkbox 0x04, color 0x05, point 0x06, dropdown 0x07, slider 0x0a, group-start
// & label 0x0d, group-end 0x0e, 3D point 0x12.
var pardControlType = map[PseudoControlKind]byte{
	PseudoAngle:      0x03,
	PseudoCheckbox:   0x04,
	PseudoColor:      0x05,
	PseudoPoint:      0x06,
	PseudoDropdown:   0x07,
	PseudoSlider:     0x0a,
	PseudoGroupStart: 0x0d,
	PseudoLabel:      0x0d, // same control type as group-start; distinguished by the @0x04 flag
	PseudoGroupEnd:   0x0e,
	PseudoLayer:      0x00, // same control type as the effect header; distinguished by position (not first)
	PseudoPoint3D:    0x12,
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

	// Dropdown (PseudoDropdown): the menu items, in order. The pard records the
	// option count; the items are carried in a trailing pdnm ("a|b|c"). Default
	// (rounded, 1-based) selects the initial item; out-of-range clamps into 1..N.
	Options []string

	// Point (PseudoPoint) / Point3D (PseudoPoint3D) default position, as a
	// fraction of the host layer's coordinate space — AE stores effect point
	// params as value÷(layer source dim), or ÷(comp dim) for source-less layers
	// (shape/text), z÷height. {fx, fy} for 2D, {fx, fy, fz} for 3D. nil → origin
	// (no value entry synthesized — the plain type default).
	PointDefault []float64

	// Layer (PseudoLayer): the bound layer's internal ID, or 0 for "None". The
	// picker defaults to None unless a non-zero ID is given (AE validates that
	// the ID resolves to a layer in the comp when the project opens).
	LayerID uint32
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
	// Controls are numbered -0001.. by emitted pard, not by control: a Label
	// expands to a group-start/group-end pard pair, so it consumes two slots.
	idx := 0
	var valEntries []*rifx.Chunk
	for _, c := range controls {
		entries, err := synthControlEntries(c)
		if err != nil {
			return nil, err
		}
		firstIdx := idx + 1
		for _, e := range entries {
			idx++
			parT.Children = append(parT.Children, makeTdmn(fmt.Sprintf("%s-%04d", matchName, idx)), e.pard)
			parT.Children = append(parT.Children, e.trailing...)
		}
		// Controls whose value can't be elided into the pard (Point/3DPoint with
		// a non-origin default, Layer picker) carry a value-group entry keyed by
		// the control's first pard match-name.
		valEntries = append(valEntries, synthControlValueEntry(c, fmt.Sprintf("%s-%04d", matchName, firstIdx))...)
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
	}}
	valTdgp.Children = append(valTdgp.Children, valEntries...)
	valTdgp.Children = append(valTdgp.Children,
		makeTdmn("ADBE Effect Built In Params"), biValTdgp,
		makeTdmn("ADBE Group End"))

	fnam := &rifx.Chunk{ID: rifx.IDFnam, Data: utf8StringData(label)}
	pgui := &rifx.Chunk{ID: rifx.IDPgui, Data: make([]byte, 16)}

	return &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDSspc, Children: []*rifx.Chunk{
		fnam, parT, valTdgp, pgui,
	}}, nil
}

// pardEntry is one pard plus any trailing leaf chunks (e.g. the pdnm carrying a
// checkbox's on/off label or a dropdown's menu items) that follow it in parT.
type pardEntry struct {
	pard     *rifx.Chunk
	trailing []*rifx.Chunk
}

// synthControlEntries expands one PseudoControl into the parT pard entries it
// occupies. Most controls are a single pard; a Checkbox/Dropdown carries a
// trailing pdnm string; a Label is a group-start (label flag) + group-end pair.
func synthControlEntries(c PseudoControl) ([]pardEntry, error) {
	switch c.Kind {
	case PseudoCheckbox:
		// A Checkbox carries its on/off label as a trailing pdnm string (AE:
		// "PF_ParamCheckbox must have nameptr set" without it).
		pard, err := synthControlPard(c)
		if err != nil {
			return nil, err
		}
		return []pardEntry{{pard: pard, trailing: []*rifx.Chunk{pdnmChunk(c.Name)}}}, nil
	case PseudoDropdown:
		// The menu items travel in a trailing pdnm as "opt1|opt2|..".
		pard, err := synthControlPard(c)
		if err != nil {
			return nil, err
		}
		return []pardEntry{{pard: pard, trailing: []*rifx.Chunk{pdnmChunk(strings.Join(c.Options, "|"))}}}, nil
	case PseudoLabel:
		// AE's Pseudo Effect Maker emits a label as a 0x0d group-start carrying
		// the label flag (@0x04=0x20) immediately closed by a 0x0e group-end.
		start, err := synthControlPard(c)
		if err != nil {
			return nil, err
		}
		return []pardEntry{{pard: start}, {pard: synthGroupEndPard()}}, nil
	default:
		pard, err := synthControlPard(c)
		if err != nil {
			return nil, err
		}
		return []pardEntry{{pard: pard}}, nil
	}
}

func pdnmChunk(s string) *rifx.Chunk {
	return &rifx.Chunk{ID: rifx.IDPdnm, Data: utf8StringData(s)}
}

// synthGroupEndPard builds the 0x0e group-end pard (@0x04=0x08, @0x30=2) that
// closes a Label's or a PseudoGroupStart's nesting.
func synthGroupEndPard() *rifx.Chunk {
	return synthPard(0x0e, "", func(d []byte) {
		bePutU32(d[0x04:], 0x08)
		bePutU32(d[0x30:], 2)
	})
}

// Per-type value-entry tdb4 descriptors (124 bytes), verbatim from AE's own
// Pseudo Effect Maker output (test_data/pseudo_rich_demo.aep). The head encodes
// the dimension + per-type markers; @0x10.. is a per-dimension constant block
// (point and 3D-point share it, dim-1 differs) — not value-dependent, so it is
// copied as-is and the actual value lives in the companion cdat / tdpi.
const (
	pointTdb4Hex   = "db990002000f0003ffffff0400005da83d9b7cdfd9d7bdbc3ff00000000000003ff00000000000003ff00000000000003ff00000000000000000000406000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	point3DTdb4Hex = "db990003000f0003ffffff0400005da83d9b7cdfd9d7bdbc3ff00000000000003ff00000000000003ff00000000000003ff00000000000000000000809000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	layerTdb4Hex   = "db99000100010000000100ff00005da83f1a36e2eb1c432d3ff00000000000003ff00000000000003ff00000000000003ff00000000000000000000404000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
)

// synthControlValueEntry builds the value-group entry (tdmn + LIST tdbs) for one
// control whose value cannot be elided into the pard: a Point/Point3D with a
// non-origin default, or a Layer picker (carries its binding in tdpi). Returns
// nil when the control's value is elidable (AE reads it from the pard).
func synthControlValueEntry(c PseudoControl, matchName string) []*rifx.Chunk {
	switch c.Kind {
	case PseudoPoint:
		if len(c.PointDefault) < 2 {
			return nil // origin → elide
		}
		cdat := make([]byte, 48)
		bePutF64(cdat[0x00:], c.PointDefault[0])
		bePutF64(cdat[0x08:], c.PointDefault[1])
		return valueEntry(matchName, c.Name, pointTdb4Hex, cdat, false, 0)
	case PseudoPoint3D:
		if len(c.PointDefault) < 3 {
			return nil
		}
		cdat := make([]byte, 72)
		bePutF64(cdat[0x00:], c.PointDefault[0])
		bePutF64(cdat[0x08:], c.PointDefault[1])
		bePutF64(cdat[0x10:], c.PointDefault[2])
		return valueEntry(matchName, c.Name, point3DTdb4Hex, cdat, false, 0)
	case PseudoLayer:
		// A layer picker always carries a value entry (tdpi = bound layer id, 0
		// = None); the pard alone has no slot for the binding.
		return valueEntry(matchName, c.Name, layerTdb4Hex, make([]byte, 40), true, c.LayerID)
	}
	return nil
}

// valueEntry assembles [tdmn, LIST tdbs{tdsb, tdsn, tdb4, cdat, [tdpi, tdps]}].
// label is the control's display name (UTF-8 — value-entry tdsn is UTF-8, unlike
// the ANSI/GBK pard @0x10 name). withLayer appends the tdpi/tdps layer binding.
func valueEntry(matchName, label, tdb4Hex string, cdat []byte, withLayer bool, layerID uint32) []*rifx.Chunk {
	tdb4 := mustHex124(tdb4Hex)
	tdbs := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdbs, Children: []*rifx.Chunk{
		makeTdsbFlags(3), // value-entry tdsb = 0x00000003 (AE-native)
		makeTdsn(label),
		{ID: rifx.IDtdb4, Data: tdb4},
		{ID: rifx.IDCdat, Data: cdat},
	}}
	if withLayer {
		tdpi := make([]byte, 4)
		bePutU32(tdpi, layerID)
		tdbs.Children = append(tdbs.Children,
			&rifx.Chunk{ID: rifx.IDTdpi, Data: tdpi},
			&rifx.Chunk{ID: rifx.IDTdps, Data: make([]byte, 4)})
	}
	return []*rifx.Chunk{makeTdmn(matchName), tdbs}
}

func mustHex124(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 124 {
		panic(fmt.Sprintf("pseudo value-entry tdb4 must be 124 bytes: len=%d err=%v", len(b), err))
	}
	return b
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
		case PseudoDropdown:
			// @0x38 selected index (1-based); @0x3C hi16 = option count, lo16 =
			// selected index. The menu item strings live in the trailing pdnm.
			n := len(c.Options)
			if n < 1 {
				n = 1
			}
			sel := int(c.Default)
			if sel < 1 {
				sel = 1
			} else if sel > n {
				sel = n
			}
			bePutU32(d[0x38:], uint32(sel))
			bePutU32(d[0x3C:], uint32(n)<<16|uint32(sel))
		case PseudoGroupStart:
			// 0x0d, label flag (@0x04) left 0; @0x30 = 2 like all containers.
			bePutU32(d[0x30:], 2)
		case PseudoLabel:
			// 0x0d with the label flag set; closed by a generated group-end.
			bePutU32(d[0x04:], 0x20)
			bePutU32(d[0x30:], 2)
		case PseudoGroupEnd:
			// 0x0e group-end (when authored explicitly rather than via a Label).
			bePutU32(d[0x04:], 0x08)
			bePutU32(d[0x30:], 2)
		case PseudoLayer:
			// Layer picker: control_type 0x00 (like the header) with @0x30 = 2.
			// The bound layer (if any) lives in the value entry's tdpi.
			bePutU32(d[0x30:], 2)
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
