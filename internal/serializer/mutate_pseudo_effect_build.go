package serializer

import (
	"encoding/hex"
	"fmt"
	"math"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/simplifiedchinese"

	"github.com/yueli-fx/aep-parser/internal/rifx"
)

// PseudoLabelCodepage selects the ANSI codepage a pseudo effect's control labels
// (pard @0x10 names) are encoded in. AE decodes that field in the viewing
// machine's system codepage (not UTF-8), so a CJK label only displays correctly
// on a matching-locale Windows — this picks which locale to target. ASCII labels
// are codepage-independent. See incidents/pseudo-control-label-ansi-codepage.md.
type PseudoLabelCodepage int

const (
	PseudoLabelGBK      PseudoLabelCodepage = iota // Simplified Chinese (GBK / cp936) — default; ASCII passes through
	PseudoLabelShiftJIS                            // Japanese (Shift-JIS / cp932)
)

func pseudoLabelEncoder(cp PseudoLabelCodepage) *encoding.Encoder {
	switch cp {
	case PseudoLabelShiftJIS:
		return japanese.ShiftJIS.NewEncoder()
	default:
		return simplifiedchinese.GBK.NewEncoder()
	}
}

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

// pardControlType maps a kind to the pard control_type byte (@0x0F), RE'd from a //nolint:jargon
// real Pseudo Effect Maker output (test_data/fixtures/pseudo_rich_demo.aep): angle 0x03,
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
	PseudoLabel:      0x0d, // same control type as group-start; emitted as a self-closing label pair
	PseudoGroupEnd:   0x0e,
	PseudoLayer:      0x00, // same control type as the effect header; distinguished by position (not first)
	PseudoPoint3D:    0x12,
}

// PseudoControl is one control in a from-scratch pseudo effect. The optional
// fields customize the control's pard defaults (RE'd from a real Pseudo Effect //nolint:jargon
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

	// Layer (PseudoLayer): the bound layer's internal ID. 0 binds the effect's
	// host layer (the PEM default) — AE hides a picker that resolves to no layer,
	// so "None" is not a renderable option. A non-zero ID binds that layer (but
	// binding a non-host layer does not render yet — see the render incident).
	LayerID uint32

	// Label (PseudoLabel): dim/gray the label text. Default false = normal
	// (bright); Dimmed true sets pard @0x04 bit 0x20.
	Dimmed bool
}

// BuildPseudoEffect constructs a pseudo effect entirely in Go — no .ffx, no AE,
// and no cloned template bytes: every pard is synthesized field-by-field from
// the RE'd pard layout — and splices it into the layer's Effect Parade. uid is
// the per-effect unique id ("Pseudo/<uid>/<name>"); name is the match-name
// segment; displayName is the effect label; controls are its controls in order.
// Controls take their type defaults. Returns the parsed *Effect.
//
// (Full contract lives on the aep.BuildPseudoEffect facade.)
func BuildPseudoEffect(layer *Layer, uid, name, displayName string, controls []PseudoControl, cp PseudoLabelCodepage) (*Effect, error) {
	if layer == nil {
		return nil, fmt.Errorf("BuildPseudoEffect: layer is nil")
	}
	if uid == "" || name == "" {
		return nil, fmt.Errorf("BuildPseudoEffect: uid and name are required")
	}
	// PEM's canonical pseudo match-name is "Pseudo/<uid>" — the effect name lives
	// in the display name (fnam / value-group tdsn), NOT the match-name. An extra
	// "/<name>" segment is tolerated by value params but makes AE fail to render a
	// layer-picker control (the one param it must recognize as a layer reference).
	matchName := "Pseudo/" + uid
	label := displayName
	if label == "" {
		label = name
	}
	sspc, err := synthPseudoSspc(matchName, label, controls, cp, layer.ID)
	if err != nil {
		return nil, err
	}
	return addEffectFromChunks(layer, "BuildPseudoEffect", matchName, makeTdmn(matchName), sspc)
}

// synthPseudoSspc assembles the in-parade sspc for a pseudo effect entirely from
// synthesized chunks: fnam(Utf8) + parT(header pard + control pards + built-in
// pard) + tdgp(scaffold + built-in value group, controls at defaults) + pgui.
func synthPseudoSspc(matchName, label string, controls []PseudoControl, cp PseudoLabelCodepage, hostLayerID uint32) (*rifx.Chunk, error) {
	parT := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDparT}
	parT.Children = append(parT.Children, makeParn(0)) // count patched below
	// -0000 effect header pard: control_type 0, struct-field (@0x30) = 2.
	parT.Children = append(parT.Children,
		makeTdmn(matchName+"-0000"), synthPard(0x00, "", cp, func(d []byte) { bePutU32(d[0x30:], 2) }))
	// The header's own value-group entry leads the value group (binds the effect
	// to its host layer); the property-tree scaffold AE's ECW walks starts here.
	valEntries := headerValueEntry(matchName+"-0000", hostLayerID)
	// Controls are numbered -0001.. by emitted pard, not by control: a Label
	// expands to a group-start/group-end pard pair, so it consumes two slots.
	idx := 0
	for _, c := range controls {
		entries, err := synthControlEntries(c, cp, hostLayerID)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			idx++
			mn := fmt.Sprintf("%s-%04d", matchName, idx)
			parT.Children = append(parT.Children, makeTdmn(mn), e.pard)
			parT.Children = append(parT.Children, e.trailing...)
			// Structural / value-bearing controls contribute a value-group entry
			// (the property-tree scaffold AE's Effect-Controls panel walks).
			if e.valueEntry != nil {
				valEntries = append(valEntries, e.valueEntry(mn)...)
			}
		}
	}
	// Built-in "Compositing Options" pard (control_type 0x09, otherwise empty).
	parT.Children = append(parT.Children,
		makeTdmn("ADBE Effect Built In Params"), synthPard(0x09, "", cp, nil))
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
// checkbox's on/off label or a dropdown's menu items) that follow it in parT,
// plus an optional value-group entry builder. The value entry is REQUIRED for
// structural controls (label/group markers) and value-bearing ones (point with
// coords, layer): AE builds the Effect-Controls property tree from the value
// group, and an omitted scaffold entry leaves the tree malformed so the panel
// fails to render later controls. nil = elided (AE reads the value from the pard
// — fine for slider/angle/color/checkbox/dropdown).
type pardEntry struct {
	pard       *rifx.Chunk
	trailing   []*rifx.Chunk
	valueEntry func(matchName string) []*rifx.Chunk
}

// synthControlEntries expands one PseudoControl into the parT pard entries it
// occupies. Most controls are a single pard; a Checkbox/Dropdown carries a
// trailing pdnm string; a Label is a group-start + group-end pair.
func synthControlEntries(c PseudoControl, cp PseudoLabelCodepage, hostLayerID uint32) ([]pardEntry, error) {
	pard, err := synthControlPard(c, cp)
	if err != nil {
		return nil, err
	}
	switch c.Kind {
	case PseudoCheckbox:
		// A Checkbox carries its on/off label as a trailing pdnm string (AE:
		// "PF_ParamCheckbox must have nameptr set" without it).
		return []pardEntry{{pard: pard, trailing: []*rifx.Chunk{pdnmChunk(c.Name)}}}, nil
	case PseudoDropdown:
		// The menu items travel in a trailing pdnm as "opt1|opt2|..".
		return []pardEntry{{pard: pard, trailing: []*rifx.Chunk{pdnmChunk(strings.Join(c.Options, "|"))}}}, nil
	case PseudoLabel:
		// A label is a 0x0d group-start immediately closed by a 0x0e group-end.
		// Both build from the pard alone — no value
		// entry. AE's own minimal form gives the value group only to the binding
		// controls (effect header + layer pickers); a synthesized scaffold entry
		// for a label/group corrupted AE's handling of a later binding control.
		return []pardEntry{{pard: pard}, {pard: synthGroupEndPard(cp)}}, nil
	case PseudoGroupStart, PseudoGroupEnd:
		return []pardEntry{{pard: pard}}, nil
	case PseudoPoint, PseudoPoint3D:
		// The default coords live in the pard (see synthControlPard); AE builds
		// the control from the pard alone — exactly like its own all-defaults
		// form, which displays a point/3D with an empty value group. Emitting a
		// synthesized value entry instead made AE hide the control.
		return []pardEntry{{pard: pard}}, nil
	case PseudoLayer:
		// A layer picker carries a value entry whose tdpi is the bound layer id.
		// Two RE'd musts (else AE hides it): (1) tdbs flag = 1 (plain property),
		// NOT 3 (the effect-header anchor flag); (2) tdpi must resolve to a layer
		// — a fresh PEM picker defaults to its own host layer. So an unset
		// LayerID binds the host; a caller value overrides the target.
		bound := c.LayerID
		if bound == 0 {
			bound = hostLayerID
		}
		name := c.Name
		return []pardEntry{{pard: pard, valueEntry: func(mn string) []*rifx.Chunk {
			return valueEntry(mn, 1, name, layerTdb4Hex, make([]byte, 40), true, bound)
		}}}, nil
	default:
		// Slider / Angle / Color: value elided (read from pard).
		return []pardEntry{{pard: pard}}, nil
	}
}

func pdnmChunk(s string) *rifx.Chunk {
	return &rifx.Chunk{ID: rifx.IDPdnm, Data: utf8StringData(s)}
}

// synthGroupEndPard builds the 0x0e group-end pard (@0x04=0x08, @0x30=2) that
// closes a Label's or a PseudoGroupStart's nesting.
func synthGroupEndPard(cp PseudoLabelCodepage) *rifx.Chunk {
	return synthPard(0x0e, "", cp, func(d []byte) {
		bePutU32(d[0x04:], 0x08)
		bePutU32(d[0x30:], 2)
	})
}

// value-entry tdb4 descriptors (124 bytes), verbatim from AE's own Pseudo Effect
// Maker output. Only the effect header and the layer-picker carry a value entry
// (everything else builds from its pard), and both are dimension-1 (scalar/
// reference) so the descriptor is comp-independent — copied as-is; the binding
// lives in the companion tdpi. tdb4 layout (RE'd): @0x02 dimension count, @0x0C
// the constant 0x5da8, @0x10.. a per-dimension matrix (trivial for dim-1).
const (
	layerTdb4Hex  = "db99000100010000000100ff00005da83f1a36e2eb1c432d3ff00000000000003ff00000000000003ff00000000000003ff00000000000000000000404000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	headerTdb4Hex = "db990001000100000001000000005da83f1a36e2eb1c432d3ff00000000000003ff00000000000003ff00000000000003ff00000000000000000000404000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
)

// headerValueEntry is the effect header's (-0000) value-group entry. AE writes
// one for every pseudo effect; its tdpi binds the effect to its host layer (the
// layer the effect is applied to). Omitting it leaves the value group one entry
// short of AE's canonical form.
func headerValueEntry(matchName string, hostLayerID uint32) []*rifx.Chunk {
	return valueEntry(matchName, 3, "", headerTdb4Hex, make([]byte, 40), true, hostLayerID)
}

// valueEntry assembles [tdmn, LIST tdbs{tdsb, tdsn, tdb4, cdat, [tdpi, tdps]}].
// tdsbFlag is the tdbs flag word (3 for the header/point/layer, 1 for label/group
// markers — AE-native). label is the display name (UTF-8 — value-entry tdsn is
// UTF-8, unlike the ANSI/GBK pard @0x10 name). withLayer appends tdpi/tdps.
func valueEntry(matchName string, tdsbFlag uint32, label, tdb4Hex string, cdat []byte, withLayer bool, layerID uint32) []*rifx.Chunk {
	tdb4 := mustHex124(tdb4Hex)
	tdbs := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdbs, Children: []*rifx.Chunk{
		makeTdsbFlags(tdsbFlag),
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
func synthControlPard(c PseudoControl, cp PseudoLabelCodepage) (*rifx.Chunk, error) {
	ct, ok := pardControlType[c.Kind]
	if !ok {
		return nil, fmt.Errorf("BuildPseudoEffect: unsupported control kind %d", c.Kind)
	}
	return synthPard(ct, c.Name, cp, func(d []byte) {
		switch c.Kind {
		case PseudoColor:
			// @0x38 ARGB last + @0x3C ARGB default (RE'd offsets — an earlier
			// draft wrote @0x40/@0x44, which AE silently ignored). nil → white.
			argb := colorToARGB(c.Color)
			bePutU32(d[0x38:], argb)
			bePutU32(d[0x3C:], argb)
		case PseudoPoint:
			// The point's default coords live IN the pard as 16.16-fixed
			// fractions of the layer/comp space — @0x38 x, @0x3C y — plus a
			// ×100 (percent) copy at @0x44/@0x48. AE hides a point whose pard
			// carries no value ("range has no values"); the value entry's cdat
			// holds the same coords for the current value. RE'd from a //nolint:jargon
			// confirmed-working AE sample (point=500,1080 → 0.2604,1.0).
			var px, py float64
			if len(c.PointDefault) >= 2 {
				px, py = c.PointDefault[0], c.PointDefault[1]
			}
			putFx1616(d[0x38:], px)
			putFx1616(d[0x3C:], py)
			putFx1616(d[0x44:], px*100)
			putFx1616(d[0x48:], py*100)
		case PseudoSlider:
			// @0x38 f8 default; @0x68/@0x6C f4 valid range; @0x70/@0x74 f4
			// visible range; @0x78 f4 default; @0x7C precision/display flags
			// (constant 0x00050003 in AE-authored sliders). NOTE: a confirmed-
			// working AE slider leaves @0x04 = 0 — an earlier RE wrote a 0x200
			// "flag" there that made AE hide the slider.
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
			bePutF64(d[0x38:], def)
			bePutF32(d[0x68:], float32(lo))
			bePutF32(d[0x6C:], float32(hi))
			bePutF32(d[0x70:], float32(lo))
			bePutF32(d[0x74:], float32(hi))
			bePutF32(d[0x78:], float32(def))
			// @0x7C display flags. 0x00020000 = AE's default-config slider
			// (verbatim from pseudo2.aep). The 0x00050003 form an earlier RE
			// copied turned on the percent display (that sample had it enabled).
			bePutU32(d[0x7C:], 0x00020000)
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
			// Like a 2D point but the default coords are f64 (not 16.16):
			// @0x38 x, @0x40 y, @0x48 z (fractions of the layer/comp space),
			// plus a ×100 (percent) copy at @0x50/@0x58/@0x60. AE hides a 3D
			// point whose pard has no value. RE'd from a confirmed-working AE //nolint:jargon
			// sample (100,200,300 → 0.0521,0.1852,0.2778).
			var qx, qy, qz float64
			if len(c.PointDefault) >= 3 {
				qx, qy, qz = c.PointDefault[0], c.PointDefault[1], c.PointDefault[2]
			}
			bePutF64(d[0x38:], qx)
			bePutF64(d[0x40:], qy)
			bePutF64(d[0x48:], qz)
			bePutF64(d[0x50:], qx*100)
			bePutF64(d[0x58:], qy*100)
			bePutF64(d[0x60:], qz*100)
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
			// 0x0d self-closing group = a label (a group-end is generated after
			// it). User panel verification: @0x04=0 renders as a normal bright
			// label; @0x04 bit 0x20 renders as gray/dim.
			if c.Dimmed {
				bePutU32(d[0x04:], 0x20)
			}
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
// The name is encoded with pardNameBytes in the cp codepage — AE reads pard
// @0x10 in the viewing machine's system ANSI codepage (NOT UTF-8), so a CJK
// label is encoded to match AE's own Pseudo Effect Maker output byte-for-byte
// and display correctly on a matching-locale system. ASCII passes through
// unchanged. See incidents/pseudo-control-label-ansi-codepage.md.
func synthPard(controlType byte, name string, cp PseudoLabelCodepage, body func(d []byte)) *rifx.Chunk {
	d := make([]byte, 148)
	d[0x0F] = controlType
	nb := pardNameBytes(name, cp)
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
func pardNameBytes(name string, cp PseudoLabelCodepage) []byte {
	if name == "" {
		return nil
	}
	if b, err := pseudoLabelEncoder(cp).Bytes([]byte(name)); err == nil {
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

// putFx1616 writes f as an unsigned 16.16 fixed-point u32 (AE's point-pard
// fraction encoding).
func putFx1616(b []byte, f float64) {
	bePutU32(b, uint32(int64(f*65536)))
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
