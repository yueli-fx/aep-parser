// Package aep parses Adobe After Effects .aep project files and supports
// limited write-back of footage paths.
//
// .aep files use the RIFX (Big-Endian RIFF) binary format. This package
// reads the chunk tree via internal/rifx, extracts compositions, footage,
// folders and layers, and can serialize a (possibly mutated) project back
// to a valid .aep file.
//
// Typical use:
//
//	project, err := aep.Open("my-project.aep")
//	if err != nil { ... }
//	for _, c := range project.Compositions { ... }
//
//	// Rewrite a footage path and save:
//	project.Footage[0].SetPath(`D:\new\location\file.png`)
//	f, _ := os.Create("modified.aep")
//	defer f.Close()
//	project.WriteAEP(f)
//
// Concurrency: Project (and everything reachable from it) is NOT safe for
// concurrent use. Multiple goroutines may safely READ disjoint subtrees of
// a parsed Project, but any Set* call (Footage.SetPath, Property.SetStaticValue,
// Keyframe.SetTime, Keyframe.SetValue) mutates the underlying rifx.Chunk
// bytes in place and races with concurrent readers/writers of the same
// chunk. Wrap mutation paths in your own synchronization.
package scene

import (
	"fmt"
)

// BitsPerChannel represents color depth.
type BitsPerChannel uint8

const (
	BPC8  BitsPerChannel = 0x00
	BPC16 BitsPerChannel = 0x01
	BPC32 BitsPerChannel = 0x02
)

func (b BitsPerChannel) String() string {
	switch b {
	case BPC8:
		return "8bpc"
	case BPC16:
		return "16bpc"
	case BPC32:
		return "32bpc"
	default:
		return fmt.Sprintf("unknown(%d)", b)
	}
}

// FeetFramesFilmType represents the film type for feet+frames timecode display.
type FeetFramesFilmType uint8

const (
	FeetFramesFilmTypeMM35 FeetFramesFilmType = 0 // 35mm film
	FeetFramesFilmTypeMM16 FeetFramesFilmType = 1 // 16mm film
)

func (f FeetFramesFilmType) String() string {
	switch f {
	case FeetFramesFilmTypeMM35:
		return "MM35 (35mm)"
	case FeetFramesFilmTypeMM16:
		return "MM16 (16mm)"
	default:
		return fmt.Sprintf("unknown(%d)", f)
	}
}

// FootageTimecodeDisplayStartType represents how timecode is displayed for footage.
type FootageTimecodeDisplayStartType uint8

const (
	FootageTimecodeDisplayStartTypeStart0         FootageTimecodeDisplayStartType = 0 // Start at 0
	FootageTimecodeDisplayStartTypeUseSourceMedia FootageTimecodeDisplayStartType = 1 // Use source media
)

func (f FootageTimecodeDisplayStartType) String() string {
	switch f {
	case FootageTimecodeDisplayStartTypeStart0:
		return "Start at 0"
	case FootageTimecodeDisplayStartTypeUseSourceMedia:
		return "Use source media"
	default:
		return fmt.Sprintf("unknown(%d)", f)
	}
}

// FramesCountType represents how frames are counted in the project.
type FramesCountType uint8

const (
	FramesCountTypeStart0             FramesCountType = 0 // Start at 0
	FramesCountTypeStart1             FramesCountType = 1 // Start at 1
	FramesCountTypeTimecodeConversion FramesCountType = 2 // Timecode conversion
)

func (f FramesCountType) String() string {
	switch f {
	case FramesCountTypeStart0:
		return "Start at 0"
	case FramesCountTypeStart1:
		return "Start at 1"
	case FramesCountTypeTimecodeConversion:
		return "Timecode conversion"
	default:
		return fmt.Sprintf("unknown(%d)", f)
	}
}

// TimeDisplayType represents how time is displayed in the project.
type TimeDisplayType uint8

const (
	TimeDisplayTypeTimecode TimeDisplayType = 0 // Timecode
	TimeDisplayTypeFrames   TimeDisplayType = 1 // Frames
)

func (t TimeDisplayType) String() string {
	switch t {
	case TimeDisplayTypeTimecode:
		return "Timecode"
	case TimeDisplayTypeFrames:
		return "Frames"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

// ColorManagementSystem represents the color management system used by the project.
type ColorManagementSystem uint8

const (
	ColorManagementSystemAdobe ColorManagementSystem = 0 // Adobe color management
	ColorManagementSystemOCIO  ColorManagementSystem = 1 // OCIO color management
)

func (c ColorManagementSystem) String() string {
	switch c {
	case ColorManagementSystemAdobe:
		return "Adobe"
	case ColorManagementSystemOCIO:
		return "OCIO"
	default:
		return fmt.Sprintf("unknown(%d)", c)
	}
}

// LutInterpolationMethod represents the LUT interpolation method for the project.
type LutInterpolationMethod uint8

const (
	LutInterpolationMethodTrilinear   LutInterpolationMethod = 0 // Trilinear
	LutInterpolationMethodTetrahedral LutInterpolationMethod = 1 // Tetrahedral
)

func (l LutInterpolationMethod) String() string {
	switch l {
	case LutInterpolationMethodTrilinear:
		return "Trilinear"
	case LutInterpolationMethodTetrahedral:
		return "Tetrahedral"
	default:
		return fmt.Sprintf("unknown(%d)", l)
	}
}

// ItemType classifies a project item.
type ItemType string

const (
	ItemTypeComposition ItemType = "composition"
	ItemTypeFootage     ItemType = "footage"
	ItemTypeFolder      ItemType = "folder"
	ItemTypeUnknown     ItemType = "unknown"
)

// ProjectItem describes one entry in the After Effects project panel.
// Items are stored in project-panel traversal order. ParentID is zero for a
// root item, and Order is the item's zero-based position among its siblings.
// The type-specific payload remains in Compositions, Footage, or Folders.
type ProjectItem struct {
	ID       uint32
	Kind     ItemType
	ParentID uint32
	Order    int
}

// Project holds the fully parsed contents of an .aep file.
type Project struct {
	Items          []ProjectItem  // canonical project-panel hierarchy and mixed item order
	Compositions   []*Composition // all compositions in the project
	Footage        []*Footage     // all footage items (files / solids / placeholders)
	Folders        []*Folder      // project-panel folders
	BitsPerChannel BitsPerChannel // project color depth (8 / 16 / 32 bpc)

	// RenderQueue holds the parsed render queue (LIST:LRdr). Nil when the
	// project has no LRdr chunk; non-nil with empty Items for an empty queue.
	// Read-only (P3 §3A slice-1).
	RenderQueue *RenderQueue

	// Warnings collects non-fatal parsing anomalies — chunks whose header
	// looked sane enough to attempt decoding but whose payload did not
	// match the expected layout (length mismatch, impossible counts, etc).
	// Each entry is a human-readable string kept for compatibility and mutation
	// rollback checks; ParseWarnings carries the structured form. Empty (nil) on
	// a clean parse.
	Warnings []string

	// ParseWarnings is the structured companion to Warnings. New code should use
	// this when it needs machine-readable warning context; Warnings remains the
	// stable human-readable view and transaction signal.
	ParseWarnings []ParseWarning

	// back holds the underlying RIFX root + project-level single-field chunk
	// refs that power length-preserving writes. Nil for projects built outside
	// the parser. See back_project.go.
	back ProjectWriter

	// V2: derived state for structural mutation (NewComposition / 未来 NewFootage etc.)
	nextItemID uint32   // monotonic Item ID counter; never reused (see Invariants #9)
	target     AETarget // which AE-version template NewProject loaded; drives per-target builder chunk selection
}

// CompositionByID returns the first composition whose ID matches id, or
// nil if no such composition exists. ID 0 is treated as no-match (mirrors
// Composition.LayerByID; real AE projects start item IDs at 1).
// Footage and Folder items share the same ID namespace but are stored in
// separate slices, so an ID belonging to a non-composition item returns
// nil here.
func (p *Project) CompositionByID(id uint32) *Composition {
	if id == 0 {
		return nil
	}
	for _, c := range p.Compositions {
		if c.ID == id {
			return c
		}
	}
	return nil
}

// CompositionByName returns the first composition whose Name matches the
// given string, or nil if none does. Matching is exact (case-sensitive).
// Comp names aren't guaranteed unique in AE; use CompositionByID when
// you need precise identity.
func (p *Project) CompositionByName(name string) *Composition {
	for _, c := range p.Compositions {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// FootageByName returns the first footage item whose Name matches the
// given string, or nil if none does. Matching is exact (case-sensitive).
func (p *Project) FootageByName(name string) *Footage {
	for _, f := range p.Footage {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// AVItem is the polymorphic "AV item" set used by Media Replacement
// (Essential Properties → Property.setAlternateSource). AE allows either
// a Composition (precomp) or a Footage to be set as an alternate source;
// Folder items are not AV and are excluded.
//
// Implemented by *Composition and *Footage.
type AVItem interface {
	ItemID() uint32
	ItemName() string
}

// AVItemByID returns the *Composition or *Footage with the given id,
// or nil if no AV item in the project matches (or id == 0). Folders are
// not AV items and are not consulted.
func (p *Project) AVItemByID(id uint32) AVItem {
	if id == 0 {
		return nil
	}
	for _, c := range p.Compositions {
		if c.ID == id {
			return c
		}
	}
	for _, f := range p.Footage {
		if f.ID == id {
			return f
		}
	}
	return nil
}

// Footage represents a source media file, an AE solid, or a placeholder.
//
// Path is the one field with a length-variable writer (SetPath); it rewrites
// the path chunk wholesale rather than patching bytes in place. Solids and
// placeholders have no path.
type Footage struct {
	ID        uint32  // AE internal item ID
	Name      string  // display name; SetPath syncs this to the new path's basename
	Path      string  // source file path on disk; empty for solids/placeholders. Writable via SetPath
	Width     uint16  // pixel width
	Height    uint16  // pixel height
	FrameRate float64 // frames per second (video footage)
	Duration  float64 // duration in seconds; 0 for stills
	IsStill   bool    // AE flagged this as a still image
	IsSolid   bool    // AE solid (generated solid-color source); solids have no disk path
	// SolidColor is the solid's RGB color in 0..1 (only meaningful when
	// IsSolid; parsed from the opti "Soli" chunk, alpha is always 1.0).
	SolidColor [3]float64
	// IsPlaceholder is true when opti tag = "Plac" (AE's placeholder
	// footage — name + dimensions only, no source file). Mutually
	// exclusive with IsSolid and with having a non-empty Path.
	IsPlaceholder bool

	// Comment / Label — Item-level metadata (project-panel comment +
	// timeline color chip). Populated by parseItem from the cmta + idta
	// chunks under the Item LIST.
	Comment string
	Label   uint8

	// back holds the underlying RIFX chunk refs that power length-preserving
	// writes. Nil for footage items built outside the parser. See back_footage.go.
	back FootageWriter
}

// Folder is a project panel folder.
type Folder struct {
	ID   uint32 // AE internal item ID
	Name string // folder display name
}

// ItemID returns the footage's item id (mirrors the ID field). Provided
// so *Footage satisfies AVItem.
func (f *Footage) ItemID() uint32 { return f.ID }

// ItemName returns the footage's name. Provided so *Footage satisfies
// AVItem.
func (f *Footage) ItemName() string { return f.Name }
