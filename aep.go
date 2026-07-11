// Package aep provides headless inspection and round-trip writing for Adobe
// After Effects project files without starting After Effects.
package aep

import (
	"encoding/json"
	"fmt"
	"io"

	internal "github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

const InspectionSchemaVersion = 1

// Limits bounds resources consumed while parsing one AEP. Zero fields use the
// corresponding default returned by DefaultLimits.
type Limits struct {
	MaxInputBytes     uint64
	MaxChunkBytes     uint64
	MaxAllocatedBytes uint64
	MaxNodes          int
	MaxDepth          int
}

// DefaultLimits returns budgets suitable for local project files. Services
// handling uploads should normally pass tighter limits to ParseWithLimits.
func DefaultLimits() Limits {
	limits := internal.DefaultParseLimits()
	return Limits{
		MaxInputBytes:     limits.MaxInputBytes,
		MaxChunkBytes:     limits.MaxChunkBytes,
		MaxAllocatedBytes: limits.MaxAllocatedBytes,
		MaxNodes:          limits.MaxNodes,
		MaxDepth:          limits.MaxDepth,
	}
}

func (l Limits) internal() internal.ParseLimits {
	return internal.ParseLimits{
		MaxInputBytes:     l.MaxInputBytes,
		MaxChunkBytes:     l.MaxChunkBytes,
		MaxAllocatedBytes: l.MaxAllocatedBytes,
		MaxNodes:          l.MaxNodes,
		MaxDepth:          l.MaxDepth,
	}
}

// Document owns one parsed project. A Document is not safe for concurrent use.
type Document struct {
	project *internal.Project
	source  string
}

// Inspection is the stable, lightweight project inventory returned by Inspect.
type Inspection struct {
	SchemaVersion    int                  `json:"schema_version"`
	Source           string               `json:"source,omitempty"`
	BitsPerChannel   string               `json:"bits_per_channel"`
	Compositions     []CompositionSummary `json:"compositions,omitempty"`
	CompositionCount int                  `json:"composition_count"`
	LayerCount       int                  `json:"layer_count"`
	FootageCount     int                  `json:"footage_count"`
	FolderCount      int                  `json:"folder_count"`
	Warnings         []string             `json:"warnings,omitempty"`
}

// CompositionSummary contains the fields needed to inventory a composition.
type CompositionSummary struct {
	ID        uint32  `json:"id"`
	Name      string  `json:"name"`
	Width     uint16  `json:"width"`
	Height    uint16  `json:"height"`
	FrameRate float64 `json:"frame_rate"`
	Duration  float64 `json:"duration_seconds"`
	Layers    int     `json:"layers"`
}

// Open reads and parses an AEP file from path using the library's bounded
// parser defaults.
func Open(path string) (*Document, error) {
	project, err := internal.Open(path)
	if err != nil {
		return nil, err
	}
	return &Document{project: project, source: path}, nil
}

// OpenWithLimits reads and parses an AEP path with explicit resource budgets.
func OpenWithLimits(path string, limits Limits) (*Document, error) {
	project, err := internal.OpenWithLimits(path, limits.internal())
	if err != nil {
		return nil, err
	}
	return &Document{project: project, source: path}, nil
}

// Parse reads an AEP project from a seekable input using bounded parser
// defaults.
func Parse(r io.ReadSeeker) (*Document, error) {
	project, err := internal.FromReader(r)
	if err != nil {
		return nil, err
	}
	return &Document{project: project}, nil
}

// ParseWithLimits parses a seekable AEP stream with explicit resource budgets.
func ParseWithLimits(r io.ReadSeeker, limits Limits) (*Document, error) {
	project, err := internal.FromReaderWithLimits(r, limits.internal())
	if err != nil {
		return nil, err
	}
	return &Document{project: project}, nil
}

// Inspect returns a detached project inventory that callers may modify freely.
func (d *Document) Inspect() Inspection {
	if d == nil || d.project == nil {
		return Inspection{SchemaVersion: InspectionSchemaVersion}
	}
	result := Inspection{
		SchemaVersion:    InspectionSchemaVersion,
		Source:           d.source,
		BitsPerChannel:   d.project.BitsPerChannel.String(),
		CompositionCount: len(d.project.Compositions),
		FootageCount:     len(d.project.Footage),
		FolderCount:      len(d.project.Folders),
		Warnings:         append([]string(nil), d.project.Warnings...),
		Compositions:     make([]CompositionSummary, 0, len(d.project.Compositions)),
	}
	for _, composition := range d.project.Compositions {
		if composition == nil {
			continue
		}
		result.LayerCount += len(composition.Layers)
		result.Compositions = append(result.Compositions, CompositionSummary{
			ID:        composition.ID,
			Name:      composition.Name,
			Width:     composition.Width,
			Height:    composition.Height,
			FrameRate: composition.FrameRate,
			Duration:  composition.Duration,
			Layers:    len(composition.Layers),
		})
	}
	return result
}

// ProfileJSON returns the normalized profile schema used by the diff and
// migration tools without exposing internal scene or serializer types.
func (d *Document) ProfileJSON() ([]byte, error) {
	if d == nil || d.project == nil {
		return nil, fmt.Errorf("aep: nil document")
	}
	result, err := profile.Build(d.project, profile.Options{Path: d.source})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// Write serializes the project while preserving unknown chunks that were not
// claimed by a mutation path.
func (d *Document) Write(w io.Writer) error {
	if d == nil || d.project == nil {
		return fmt.Errorf("aep: nil document")
	}
	if w == nil {
		return fmt.Errorf("aep: nil writer")
	}
	return d.project.WriteAEP(w)
}
