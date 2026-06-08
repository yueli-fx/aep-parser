package aep

import "github.com/example/aep-parser/internal/codec"

// Render queue runtime model (P3 §3A slice-1, read-only). Mirrors py-aep
// RenderQueue / RenderQueueItem / OutputModule, exposing only the fields that
// cross-validate byte-for-byte against py-aep golden JSON. Write paths, the
// full render-settings enum surface, format options and the 128-byte
// OutputModuleSettingsItem are deferred to later slices (see
// flightdeck/plans/2026-06-01-py-aep-p3-renderqueue-reader-plan.md).

// RenderQueue is the project's render queue (LIST:LRdr).
type RenderQueue struct {
	Items []*RenderQueueItem

	// back holds the LIST:LRdr chunk reference powering the structural
	// RemoveItem write (settings ldat + lhd3 count + Rout per-item block).
	// nil for queues built outside the parser. See back_render_queue.go.
	back *renderQueueBackrefs
}

// NumItems returns the number of render queue items.
func (rq *RenderQueue) NumItems() int {
	if rq == nil {
		return 0
	}
	return len(rq.Items)
}

// RenderQueueItem is one entry in the render queue.
type RenderQueueItem struct {
	// Comp is the composition this item renders, linked by comp_id from the
	// render-settings ldat. Nil if the referenced comp id was not found.
	Comp *Composition

	// Status is the raw render status code (RenderSettingsItem @0x0C). py-aep
	// maps this to its RQItemStatus enum; we expose the raw value for now.
	Status uint32

	// Name is the render-settings template name (RenderSettingsItem @0x5A,
	// windows-1252). Empty when the item uses custom (modified) settings.
	Name string

	// Comment is the item comment (RCom → Utf8), empty when no RCom present.
	Comment string

	// LogType is the raw log-type code (RenderSettingsItem @0x50). py-aep
	// maps this to its 3xxx-namespaced LogType enum; we expose raw.
	LogType uint16

	// QueueItemNotify mirrors the "notify on completion" flag (flag byte @0x07
	// bit 2).
	QueueItemNotify bool

	// ElapsedSeconds is render elapsed time (RenderSettingsItem @0x89A), 0 when
	// not yet rendered.
	ElapsedSeconds uint32

	// codec.RenderSettingsBlock holds the per-item render settings (the ExtendScript
	// get_settings() dict). Values follow py-aep NUMBER semantics: -1 means
	// "current settings" (binary 0xFFFF).
	RenderSettings RenderSettings

	// TimeSpanStart / TimeSpanDuration are resolved (seconds) per the item's
	// time_span_source: LENGTH_OF_COMP → (0, comp.Duration); WORK_AREA_ONLY →
	// (comp.WorkAreaStart, comp.WorkAreaEnd-Start); CUSTOM → ldat dividends.
	TimeSpanStart    float64
	TimeSpanDuration float64

	OutputModules []*OutputModule

	// settingsBlock aliases the 2246-byte slice of the render-settings ldat
	// for this item (Go subslices share the backing array, so patching it
	// mutates the chunk in place). nil for items built outside the parser.
	// Powers the length-preserving Set* methods in write_render_queue.go.
	settingsBlock []byte

	// back holds the RIFX chunk references that power the length-variable
	// SetComment write (RCom insert/replace). Nil outside the parser. See
	// back_render_queue.go.
	back *renderQueueItemBackrefs
}

// codec.RenderSettingsBlock is the per-item render settings (ExtendScript
// RenderQueueItem.getSettings). Enum-typed fields use py-aep NUMBER semantics:
// -1 = "current settings" (binary 0xFFFF). FieldRender/Pulldown/FrameRate have
// no current-settings sentinel.
type RenderSettings struct {
	Quality           int    // -1 current / 0 wireframe / 1 draft / 2 best
	ColorDepth        int    // -1 current / 0 8bpc / 1 16bpc / 2 32bpc
	Effects           int    // 0 all-off / 1 all-on / 2 current
	FieldRender       int    // 0 off / 1 upper-first / 2 lower-first
	Pulldown          int    // 0 off / 1..5 phase
	FrameBlending     int    // 0 off-all / 1 on-checked / 2 current
	MotionBlur        int    // 0 off-all / 1 on-checked / 2 current
	ProxyUse          int    // 0 none / 1 all / 2 current / 3 comp-only
	SoloSwitches      int    // 0 off / 2 current
	GuideLayers       int    // 0 off / 2 current
	DiskCache         int    // 0 read-only / 2 current
	FrameRate         int    // 0 use comp / 1 use this
	Resolution        [2]int // [x, y] divisors
	SkipExistingFiles bool
}

// NumOutputModules returns the number of output modules for this item.
func (it *RenderQueueItem) NumOutputModules() int {
	if it == nil {
		return 0
	}
	return len(it.OutputModules)
}

// OutputModule is one output module of a render queue item (a Roou-delimited
// group inside LIST:'LOm ').
type OutputModule struct {
	// Name is the output-module template name shown in the UI (first Utf8
	// after the Als2 LIST, e.g. "H.264 - Match Render Settings - 15 Mbps").
	Name string

	// FileTemplate is the raw file-name template (second Utf8 after Als2,
	// e.g. "[compName].[fileextension]"); template variables are not resolved.
	FileTemplate string

	// FullPath is the output folder/full path from the alas JSON "fullpath"
	// field inside the Als2 LIST. Empty when no alas/fullpath present.
	FullPath string

	// Settings holds the output-module settings (128B OutputModuleSettingsItem
	// + Roou). slice-3, read-only.
	Settings OutputModuleSettings

	// FormatOptions holds the format-specific render options (from the Ropt
	// chunk), or nil for XML-based formats (AVI/H264/QuickTime) which carry no
	// Ropt variant. slice-4, read-only.
	FormatOptions *FormatOptions

	// settingsBlock aliases this module's 128B OutputModuleSettingsItem; roouData
	// aliases the Roou chunk bytes. Both share the backing chunk array, powering
	// the length-preserving Set* methods in write_render_queue.go. nil outside
	// the parser.
	settingsBlock []byte
	roouData      []byte
}

// FormatOptions is a typed view of the Ropt chunk (format-specific render
// options). Defined in internal/codec; aliased here so callers importing aep
// continue to see aep.FormatOptions without qualification.
type FormatOptions = codec.FormatOptions

// OutputModuleSettings is the per-output-module settings (ExtendScript
// OutputModule.getSettings). Read-only (P3 §3A slice-3). Derived/mapped fields
// (Format enum, Output Audio derivation, resolved file path) are deferred.
type OutputModuleSettings struct {
	// from the 128B OutputModuleSettingsItem
	Channels            int // 0 RGB / 1 RGBA / 2 Alpha
	ResizeQuality       int
	Resize              bool
	LockAspectRatio     bool
	Crop                bool
	CropTop             int
	CropLeft            int
	CropBottom          int
	CropRight           int
	OutputAudio         int // raw (py-aep derives ON/OFF/AUTO)
	IncludeProjectLink  bool
	PostRenderAction    uint32 // raw
	ConvertToLinear     int
	UseCompFrameNumber  bool
	UseRegionOfInterest bool
	IncludeSourceXMP    bool
	PreserveRGB         bool

	// from the 154B Roou chunk
	VideoCodec      string
	FormatID        string // 4-char output format id (e.g. "H264", "TIF ")
	StartingNumber  uint32
	Width           int
	Height          int
	Depth           int // bits: 24/32/48/64/96/128
	VideoOutput     bool
	AudioSampleRate float64
	AudioBitDepth   int
	AudioChannels   int
	AudioEnabled    bool
}
