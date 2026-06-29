package serializer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

// aeLegacyTimeBase is the per-second keyframe tick rate used by older AE
// versions for NTSC comps. Modern AE writes a per-comp tick rate into
// cdta (decoded via deriveTickRate); this constant is the fallback when
// cdta is too small or absent.
const aeLegacyTimeBase = 8000.0

// parseCtx is the per-composition parsing context threaded through every
// parser sub-function. It carries comp-level state that keyframe/marker/
// mask decoders need without expanding every function signature each time
// we add a new field.
type parseCtx struct {
	// tickRate is the composition's keyframe-tick-per-second base (see
	// Composition.TickRate). Used by keyframe, marker, and mask-path-time
	// decoders. Defaults to aeLegacyTimeBase when zero.
	tickRate float64

	// compFps is the owning composition's user-facing frame rate
	// (Composition.FrameRate). Propagated into Keyframe.compFps and
	// Marker.compFps so seconds↔frames conversion (FrameTime,
	// SetFrameTime) doesn't need a back-pointer to the comp.
	compFps float64

	// compName is the owning composition's display name, prepended to
	// warnings so callers can locate which comp produced an anomaly.
	// Empty when the comp has no Utf8 name chunk.
	compName string

	// warnings, when non-nil, is appended to via warn() whenever a
	// sub-decoder hits a non-fatal anomaly. Shared across all comps of
	// one Project so Project.Warnings sees every comp's findings.
	warnings *[]string
}

// newParseCtx returns a context with tickRate falling back to the legacy
// constant when 0 is passed (e.g. cdta missing). warnings may be nil — the
// returned ctx's warn() is a no-op in that case.
func newParseCtx(tickRate float64, compName string, warnings *[]string) *parseCtx {
	if tickRate == 0 {
		tickRate = aeLegacyTimeBase
	}
	return &parseCtx{tickRate: tickRate, compName: compName, warnings: warnings}
}

// newParseCtxFPS extends newParseCtx with the comp's user-facing frame
// rate, used by frame-time accessors. Existing call sites still use
// newParseCtx (compFps stays 0; FrameTime methods then fall back to a
// default if needed).
func newParseCtxFPS(tickRate, fps float64, compName string, warnings *[]string) *parseCtx {
	ctx := newParseCtx(tickRate, compName, warnings)
	ctx.compFps = fps
	return ctx
}

// warn records a non-fatal parsing anomaly. Safe to call on a nil ctx or a
// ctx with a nil warnings slot — both are no-ops, so call sites don't need
// to gate themselves.
func (c *parseCtx) warn(format string, args ...any) {
	if c == nil || c.warnings == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if c.compName != "" {
		msg = "comp " + strconv.Quote(c.compName) + ": " + msg
	}
	*c.warnings = append(*c.warnings, msg)
}

// parseComposition reads a composition from an Item list.
//
// cdta layout (204 bytes in CC2020+ .aep files). Sources:
//   - "ref" = the reference parser's binary/composition_chunks.py CdtaChunk
//   - "fixture" = hex-dumped from test_data/re_tickrate.aep
//   - "local" = derived from observation only, no external source
//
// Offsets used here are big-endian unless otherwise noted.
//
//	0x08–0x0B : internal_timebase  (uint32) — modern: ticks/sec; legacy: see deriveTickRate           [ref + fixture]
//	0x1C–0x1F : work_area_start_dividend (uint32)                                                     [ref + fixture]
//	0x20–0x23 : work_area_start_divisor  (uint32)                                                     [ref + fixture]
//	0x24–0x27 : work_area_end_dividend   (uint32) — 0xFFFFFFFF = "extend to comp Duration"           [ref + fixture; sentinel = local]
//	0x28–0x2B : work_area_end_divisor    (uint32)                                                     [ref + fixture]
//	0x34      : bg_color_r (uint8)                                                                    [ref + fixture]
//	0x35      : bg_color_g (uint8)                                                                    [ref + fixture]
//	0x36      : bg_color_b (uint8)                                                                    [ref + fixture]
//	0x8C–0x8D : Width  (uint16)                                                                       [local + fixture]
//	0x8E–0x8F : Height (uint16)                                                                       [local + fixture]
//	0x9C–0x9D : FramerateWhole       (uint16)                                                          [local + fixture]
//	0x9E–0x9F : FramerateFractional  (uint16, units = 1/65536)                                         [local + fixture]
//	0xA8–0xAB : legacy_scale         (uint32) — see deriveTickRate                                     [local; deriveTickRate doc]
//	0xAE–0xAF : shutter_angle (uint16, degrees; AE UI default 180)                                    [ref + fixture]
//	0xB0–0xB3 : DurationFrames (uint32)                                                                [local + fixture]
//	0xB4–0xB7 : shutter_phase (int32 — likely degrees, exact units not UI-verified)                   [ref + fixture]
//	0xC4–0xC7 : motion_blur_adaptive_sample_limit (int32; AE default 128)                              [ref + fixture]
//	0xC8–0xCB : motion_blur_samples_per_frame (int32; AE default 16)                                   [ref + fixture]
func parseComposition(item *rifx.Chunk, id uint32, name string, warnings *[]string) (*Composition, error) {
	cb := &compositionBackrefs{compName: name, itemList: item}
	comp := &Composition{ID: id, Name: name}
	scene.SetCompositionBack(comp, cb)
	if utf8 := item.FindFirst(rifx.IDUtf8); utf8 != nil {
		cb.nameChunk = utf8
	}

	cdta := item.FindFirst(rifx.IDCdta)
	if cdta == nil {
		return comp, nil
	}
	cb.cdta = cdta
	d := cdta.Data

	if len(d) >= 0x04 {
		comp.ResolutionFactor = [2]uint16{
			binary.BigEndian.Uint16(d[0x00:0x02]),
			binary.BigEndian.Uint16(d[0x02:0x04]),
		}
	}

	if len(d) >= 0x37 {
		comp.BGColor = [3]uint8{d[0x34], d[0x35], d[0x36]}
	}

	if len(d) >= 0x90 {
		comp.Width = binary.BigEndian.Uint16(d[0x8C:0x8E])
		comp.Height = binary.BigEndian.Uint16(d[0x8E:0x90])
	}

	if len(d) >= 0xA0 {
		whole := binary.BigEndian.Uint16(d[0x9C:0x9E])
		frac := binary.BigEndian.Uint16(d[0x9E:0xA0])
		comp.FrameRate = float64(whole) + float64(frac)/65536.0
	}

	// Pixel aspect ratio: cdta @0x90 / @0x94 (uint32 BE numerator and
	// denominator). 1/1 = 1.0 square pixels (default).
	if len(d) >= 0x98 {
		num := binary.BigEndian.Uint32(d[0x90:0x94])
		den := binary.BigEndian.Uint32(d[0x94:0x98])
		if den != 0 {
			comp.PixelAspect = float64(num) / float64(den)
		}
	}

	if len(d) >= 0xB4 && comp.FrameRate > 0 {
		set := false
		if len(d) >= 0x30 {
			durTicks := binary.BigEndian.Uint32(d[0x2C:0x30])
			ticksPerFrame := uint32(binary.BigEndian.Uint16(d[0x06:0x08]))
			nominalTickRate := ticksPerFrame * uint32(comp.FrameRate+0.5)
			if durTicks > 0 && nominalTickRate > 0 {
				comp.Duration = float64(durTicks) / float64(nominalTickRate)
				set = true
			}
		}
		if !set {
			frames := binary.BigEndian.Uint32(d[0xB0:0xB4])
			comp.Duration = float64(frames) / comp.FrameRate
		}
	}

	// Work area uses its own dividend/divisor pair per side. The divisor
	// is per-cdta (e.g. 600 in the re_tickrate.aep fixture), NOT the
	// tick rate. End dividend == 0xFFFFFFFF is AE's sentinel for "use
	// full comp duration" — substitute Duration so callers get the
	// effective end without having to detect the sentinel themselves.
	if len(d) >= 0x2C {
		comp.WorkAreaStart = decodeFrac(d[0x1C:0x20], d[0x20:0x24])
		endDividend := binary.BigEndian.Uint32(d[0x24:0x28])
		if endDividend == 0xFFFFFFFF {
			comp.WorkAreaEnd = comp.Duration
		} else {
			comp.WorkAreaEnd = decodeFrac(d[0x24:0x28], d[0x28:0x2C])
		}
	}

	if len(d) >= 0xB0 {
		comp.ShutterAngle = binary.BigEndian.Uint16(d[0xAE:0xB0])
	}
	if len(d) >= 0xB8 {
		comp.ShutterPhase = int32(binary.BigEndian.Uint32(d[0xB4:0xB8]))
	}
	if len(d) >= 0xC8 {
		comp.MotionBlurAdaptiveSampleLimit = int32(binary.BigEndian.Uint32(d[0xC4:0xC8]))
	}
	if len(d) >= 0xCC {
		comp.MotionBlurSamplesPerFrame = int32(binary.BigEndian.Uint32(d[0xC8:0xCC]))
	}

	// Display start time (dividend/divisor pair at cdta @0xA4 / @0xA8).
	// AE writes both zero when displayStartFrame == 0. When set, the
	// pair encodes start-time in seconds with divisor matching TickRate.
	// Per-cdta divisor varies — derive seconds = dividend / divisor.
	if len(d) >= 0xAC {
		dsDividend := binary.BigEndian.Uint32(d[0xA4:0xA8])
		dsDivisor := binary.BigEndian.Uint32(d[0xA8:0xAC])
		if dsDivisor != 0 {
			comp.DisplayStartTime = float64(dsDividend) / float64(dsDivisor)
		}
	}

	comp.TickRate = deriveTickRate(d)
	cb.tickRate = comp.TickRate
	cb.FrameRateHz = comp.FrameRate

	// Renderer: PRin LIST → prin chunk, two NUL-separated ASCII strings
	// after a 4-byte prefix. First string = internal match-name (e.g.
	// "ADBE Escher" / "ADBE Ernst"); second = AE-localized display name
	// (e.g. "Advanced 3D" / "Cinema 4D"). We surface the match-name.
	if prinList := item.FindFirstList(rifx.IDPRin); prinList != nil {
		if prinChunk := prinList.FindFirst(rifx.IDPrin); prinChunk != nil && len(prinChunk.Data) > 4 {
			cb.prinChunk = prinChunk
			payload := prinChunk.Data[4:]
			if i := bytes.IndexByte(payload, 0); i >= 0 {
				comp.Renderer = string(payload[:i])
			} else {
				comp.Renderer = string(payload)
			}
		}
		cb.prdaChunk = prinList.FindFirst(rifx.IDPrda)
	}

	ctx := newParseCtxFPS(comp.TickRate, comp.FrameRate, comp.Name, warnings)

	for i, layrList := range item.FindAllList(rifx.IDLayr) {
		layer, err := parseLayer(layrList, i, ctx)
		if err != nil {
			continue // best-effort
		}
		scene.SetLayerComp(layer, comp) // wire back-pointer so Layer.Parent() works
		scene.AssignTransformDefaults(layer.Properties, comp, layer.Type)
		comp.Layers = append(comp.Layers, layer)
	}

	comp.Markers = findCompMarkers(item, ctx)
	bindMarkerListOwner(&comp.Markers)
	comp.Guides = parseGuides(item)
	comp.MotionGraphicsTemplateName, comp.EssentialGraphicsControllers = parseEssentialGraphics(item)

	return comp, nil
}

// findCompMarkers locates an Item's "comp-level markers" pseudo-layer
// (SecL LIST whose Utf8 child is "Markers") and decodes the embedded
// "ADBE Marker" mrst into a flat Marker slice. AE stores these
// alongside regular Layr LISTs but uses the SecL container as a flag
// that the contents are comp-scoped, not a real layer. Returns nil
// when the comp has no markers.
func findCompMarkers(item *rifx.Chunk, ctx *parseCtx) []*Marker {
	for _, secl := range item.FindAllList(rifx.IDSecL) {
		// Confirm this SecL is the Markers pseudo-layer (vs. AE's other
		// SecL uses; AE 2020 only emits "Markers" here, but guard anyway).
		if name := secl.FindFirst(rifx.IDUtf8); name == nil || name.Text() != "Markers" {
			continue
		}
		tdgp := secl.FindFirstList(rifx.IDTdgp)
		if tdgp == nil {
			continue
		}
		// Walk the tdmn/mrst pairs inside the Transform Group looking for
		// "ADBE Marker".
		kids := tdgp.Children
		for i := 0; i+1 < len(kids); i++ {
			if kids[i].ID != rifx.IDTdmn || trimNUL(kids[i].Data) != "ADBE Marker" {
				continue
			}
			mrst := kids[i+1]
			if !mrst.IsList() || mrst.FormType != rifx.IDMrst {
				continue
			}
			return parseMarkers(mrst, ctx)
		}
	}
	return nil
}

// decodeFrac reads a (dividend uint32 BE, divisor uint32 BE) pair and
// returns dividend/divisor as a float64. Returns 0 when divisor is 0.
// Both slices must be at least 4 bytes; callers gate by len(d).
func decodeFrac(dividend, divisor []byte) float64 {
	dvs := binary.BigEndian.Uint32(divisor)
	if dvs == 0 {
		return 0
	}
	return float64(binary.BigEndian.Uint32(dividend)) / float64(dvs)
}

// deriveTickRate computes a composition's keyframe-tick-per-second base
// from its cdta payload. See Composition.TickRate doc for the formula.
//
// cdta @0x08 holds the ticks-per-second base directly (= ticks_per_frame
// × fps, e.g. 800 × 29.97 = 23976 for NTSC, 1024 × 30 = 30720 for 30 fps).
// AE evaluates keyframe ticks against THIS value, verified end-to-end via
// AE's own valueAtTime / keyTime DOM (see incident
// ntsc-tickrate-derive-3x-off): RECT frame-8 kf = 8 × 800 = 6400 ticks →
// 6400 / 23976 = 0.2669 s, matching AE exactly.
//
// The earlier `rate × 1000 / cdta_0xA8` "legacy NTSC" correction (which
// turned 23976 into 8000 for 29.97) was a reference-parser-derived misreading never
// checked against AE's rendered keyframe times. It is latent under
// read-modify-write (keyframe ticks are preserved, so the wrong rate
// cancels) but corrupts any path that reads keyframe seconds out of one
// comp and writes them into another (from-scratch replication): the times
// land 3× too large. cdta @0xA8 is a display-time divisor, not a kf-rate
// scale.
func deriveTickRate(cdta []byte) float64 {
	if len(cdta) < 0x0C {
		return aeLegacyTimeBase
	}
	rate := binary.BigEndian.Uint32(cdta[0x08:0x0C])
	if rate == 0 {
		return aeLegacyTimeBase
	}
	return float64(rate)
}
