package scene

// Composition represents an After Effects composition.
//
// TickRate is the ticks-per-second base used for ALL keyframe / marker
// times within this composition's layers. AE stores this per-comp in
// cdta and it varies with fps and AE version. Decoding logic:
//
//	scale := cdta_uint32_BE @ 0xA8
//	rate  := cdta_uint32_BE @ 0x08
//	if scale <= 1 {           // modern comps (any fps)
//	    TickRate = rate       // typically fps × 512 or fps × 1024
//	} else {                  // legacy NTSC comps (older AE versions)
//	    TickRate = rate * 1000 / scale  // typically 8000 for 29.97
//	}
//
// For older 29.97-fps comps this is 8000; for new 30-fps comps it's
// 30720; etc.
type Composition struct {
	ID        uint32  // AE internal item ID
	Name      string  // composition display name (writable via SetName)
	Width     uint16  // canvas width in pixels (writable via SetSize)
	Height    uint16  // canvas height in pixels (writable via SetSize)
	FrameRate float64 // frames per second
	Duration  float64 // seconds
	TickRate  float64 // keyframe/marker ticks per second (per-composition, not a global 8000)

	// PixelAspect is the comp's pixel aspect ratio (1.0 = square, 2.0
	// = anamorphic, 0.91 = NTSC D1, etc.). Decoded from cdta @0x90 /
	// @0x94 as an integer numerator/denominator pair.
	PixelAspect float64

	// ResolutionFactor is the comp's preview-resolution downsample, X
	// and Y. AE Scripting's CompItem.resolutionFactor: [1,1]=Full,
	// [2,2]=Half, [3,3]=Third, [4,4]=Quarter, custom pairs allowed.
	// Decoded from cdta @0x00 / @0x02 (two uint16 BE). Default [1,1].
	ResolutionFactor [2]uint16

	// Renderer is the comp's 3D rendering engine (AE Scripting's
	// CompItem.renderer). Stored as an internal match-name string in
	// the `prin` chunk of the comp's `PRin` LIST sibling. Examples:
	//
	//   "ADBE Escher"   — Advanced 3D (AE 2025 default; alias "ADBE Picasso")
	//   "ADBE Ernst"    — Cinema 4D Engine
	//   "ADBE Standard" — Classic 3D (legacy; AE 2025 may auto-promote to Advanced)
	//
	// Empty when the comp has no PRin LIST (rare; only seen on
	// programmatically built comps that skipped the AE serialization).
	Renderer string

	// BGColor is the composition background color (R, G, B), 0..255 per
	// channel. Decoded from cdta @0x34/0x35/0x36 (3 × uint8). Defaults to
	// {0,0,0} (AE's default black).
	BGColor [3]uint8

	// WorkAreaStart / WorkAreaEnd are the render-queue partial range in
	// seconds (AE's "Work Area" — N and B in the timeline). Decoded from
	// cdta dividend/divisor pairs at @0x1C/0x20 (start) and @0x24/0x28
	// (end). When AE leaves the work area "unset", it writes
	// 0xFFFFFFFF for the end dividend; we substitute Duration so callers
	// always see the effective range.
	WorkAreaStart float64
	WorkAreaEnd   float64

	// DisplayStartTime is the comp's display-time origin in seconds (AE
	// scripting's CompItem.displayStartTime / displayStartFrame). Decoded
	// from cdta @0xA4 / @0xA8 (uint32 BE dividend / divisor pair). AE
	// writes both 0 when displayStart is 0; non-zero pair encodes
	// start-time in seconds.
	DisplayStartTime float64

	// ShutterAngle is the motion-blur shutter angle in degrees (AE UI
	// range 0..720, default 180). Decoded from cdta @0xAE as uint16 BE.
	ShutterAngle uint16

	// ShutterPhase is the motion-blur shutter phase offset. Decoded from
	// cdta @0xB4 as int32 BE. Likely degrees in the AE UI's -360..+360
	// range (fixture value -90 plausibly matches AE's -90° UI display)
	// but the exact unit is not independently UI-verified — surface as
	// raw int32 and let callers interpret.
	ShutterPhase int32

	// MotionBlurAdaptiveSampleLimit / MotionBlurSamplesPerFrame are AE's
	// motion-blur quality knobs (Composition Settings → Advanced). Decoded
	// from cdta @0xC4 / @0xC8 as int32 BE. AE clamps these to positive
	// integers in the UI (defaults: 128 and 16).
	MotionBlurAdaptiveSampleLimit int32
	MotionBlurSamplesPerFrame     int32

	Layers []*Layer

	// Markers holds composition-level markers (AE's timeline-top chapter
	// markers, distinct from per-layer markers). Stored in the AEP as a
	// pseudo-layer (`LIST SecL` named "Markers") carrying an "ADBE Marker"
	// property — decoded with the same machinery as layer markers.
	Markers []*Marker

	// Guides holds the composition's ruler guides (AE's drag-from-ruler
	// alignment lines; UI-only, no render effect). Decoded from the comp's
	// Item-level LIST:Gide. See scene_guide.go.
	Guides []*Guide

	// MotionGraphicsTemplateName is the Essential Graphics / .mogrt template
	// name (AE default "Untitled" when the comp has no EG panel). Decoded
	// from the comp's Item-level LIST:CIF3. See scene_essential_graphics.go.
	MotionGraphicsTemplateName string

	// EssentialGraphicsControllers holds the comp's Essential Graphics panel
	// controllers (exposed properties / .mogrt controls), in panel order.
	// Empty when the comp has no EG panel.
	EssentialGraphicsControllers []*EssentialGraphicsController

	// proj is the owning project, set by parseProject after the comp is
	// appended. Used by Layer.SourceComposition() to walk the project's
	// other comps via Project.CompositionByID. Unexported to keep the API
	// minimal; nil for comps built outside the parser.
	proj *Project

	// Item-level metadata shared with Footage / Folder. Populated by
	// parseItem from the surrounding Item LIST (cmta child + idta byte).
	Comment string // Item.comment — AE's project-panel comment column
	Label   uint8  // Item.label — project-panel color index 0..16 (idta @0x3A)

	// back holds the underlying RIFX chunk refs that power length-preserving
	// writes. Nil for comps built outside the parser. See back_composition.go.
	back CompositionWriter
}

// CdtaRawBytes returns the comp's cdta chunk Data slice, or nil if the
// comp has no cdta. Read-only access for debugging / RE tools — the
// underlying byte slice is the live chunk data; do not mutate.
func (c *Composition) CdtaRawBytes() []byte {
	if c.back == nil {
		return nil
	}
	return c.back.CdtaData()
}

// PrdaRawBytes returns the comp's prda chunk Data slice (renderer-specific
// options), or nil if the comp has no PRin LIST. Read-only access for
// debugging / RE tools — the underlying byte slice is the live chunk data;
// do not mutate.
func (c *Composition) PrdaRawBytes() []byte {
	if c.back == nil {
		return nil
	}
	return c.back.PrdaData()
}

// LayerByID returns the first layer in this composition whose ID matches
// id, or nil if no such layer exists. ID 0 is treated as no-match (it's
// the sentinel used by Layer.ParentID to mean "no parent"); real AE
// projects start layer IDs at 1.
func (c *Composition) LayerByID(id uint32) *Layer {
	if id == 0 {
		return nil
	}
	for _, l := range c.Layers {
		if l.ID == id {
			return l
		}
	}
	return nil
}

// LayerByName returns the first layer in this composition whose Name
// matches the given string, or nil if none does. Layer names aren't
// guaranteed unique within a comp; use LayerByID for precise identity.
func (c *Composition) LayerByName(name string) *Layer {
	for _, l := range c.Layers {
		if l.Name == name {
			return l
		}
	}
	return nil
}

// ActiveCamera returns the topmost enabled (Visible == true) camera
// layer in the comp, or nil when no enabled camera is present. AE
// uses this layer's transform + camera-options for the comp's render.
// Multiple cameras may coexist; AE picks the topmost enabled one.
func (c *Composition) ActiveCamera() *Layer {
	for _, l := range c.Layers {
		if l.Type == LayerTypeCamera && l.Visible {
			return l
		}
	}
	return nil
}

// ItemID returns the composition's item id (mirrors the ID field).
// Provided so *Composition satisfies AVItem.
func (c *Composition) ItemID() uint32 { return c.ID }

// ItemName returns the composition's name. Provided so *Composition
// satisfies AVItem.
func (c *Composition) ItemName() string { return c.Name }
